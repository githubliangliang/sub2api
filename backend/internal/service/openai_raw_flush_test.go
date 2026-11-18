package service

import (
	"bytes"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

var s22TCPFlushCount atomic.Int64

type rawTCPFlushWriter struct{ http.ResponseWriter }

func (w rawTCPFlushWriter) Flush() {
	s22TCPFlushCount.Add(1)
	flusher, ok := w.ResponseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}

func BenchmarkS22RawChatFlushTCP(b *testing.B) {
	gin.SetMode(gin.TestMode)
	body := strings.Repeat(rawStreamFlushTestEvent("x"), 100) + "data: [DONE]\n\n"
	svc := &OpenAIGatewayService{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, _ := gin.CreateTestContext(rawTCPFlushWriter{w})
		c.Request = r
		resp := &http.Response{Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
		_, _ = svc.streamRawChatCompletions(c, resp, &Account{ID: 1, Platform: PlatformOpenAI}, "gpt-5.4", "gpt-5.4", "gpt-5.4", nil, nil, time.Now(), 100)
	}))
	b.Cleanup(server.Close)
	client := server.Client()
	s22TCPFlushCount.Store(0)
	b.ReportAllocs()
	for b.Loop() {
		resp, err := client.Get(server.URL)
		if err != nil {
			b.Fatal(err)
		}
		_, err = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(float64(s22TCPFlushCount.Load())/float64(b.N), "flushes/op")
}

// flushCountingRecorder 记录 Flush 次数，并在每次 Flush 时回调当时已写出的内容。
type flushCountingRecorder struct {
	*httptest.ResponseRecorder
	flushes int
	onFlush func(written string)
}

func (w *flushCountingRecorder) Flush() {
	w.flushes++
	w.ResponseRecorder.Flush()
	if w.onFlush != nil {
		w.onFlush(w.Body.String())
	}
}

func rawStreamFlushTestEvent(content string) string {
	return `data: {"id":"chatcmpl_f","object":"chat.completion.chunk","model":"gpt-5.4","choices":[{"index":0,"delta":{"content":"` + content + `"}}]}` + "\n\n"
}

// 一次读取里到达的多行合并成一次 flush，写给客户端的内容不变。
func TestStreamRawChatCompletions_CoalescesFlushesWithinOneUpstreamRead(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"gpt-5.4","messages":[{"role":"user","content":"hello"}],"stream":true}`)
	w := &flushCountingRecorder{ResponseRecorder: httptest.NewRecorder()}
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))

	upstreamBody := rawStreamFlushTestEvent("a") + rawStreamFlushTestEvent("b") + rawStreamFlushTestEvent("c") +
		`data: {"id":"chatcmpl_f","object":"chat.completion.chunk","model":"gpt-5.4","choices":[],"usage":{"prompt_tokens":3,"completion_tokens":3,"total_tokens":6}}` + "\n\n" +
		"data: [DONE]\n\n"
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(upstreamBody)),
	}}
	svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}

	result, err := svc.forwardAsRawChatCompletions(context.Background(), c, rawChatCompletionsTestAccount(), body, "")
	require.NoError(t, err)
	require.Equal(t, 3, result.Usage.OutputTokens)
	require.Equal(t, upstreamBody, w.Body.String())
	// 之前每行（含事件间空行）flush 一次，这里是 10 次。
	require.Equal(t, 1, w.flushes)
}

// 合并不能推迟已到达的数据：开始等待上游下一批数据之前，已到达的事件必须已经 flush 给客户端。
func TestStreamRawChatCompletions_FlushesBeforeWaitingForUpstream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"gpt-5.4","messages":[{"role":"user","content":"hello"}],"stream":true}`)
	firstEventFlushed := make(chan struct{})
	w := &flushCountingRecorder{ResponseRecorder: httptest.NewRecorder()}
	w.onFlush = func(written string) {
		if strings.Contains(written, `"content":"first"`) {
			select {
			case <-firstEventFlushed:
			default:
				close(firstEventFlushed)
			}
		}
	}
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))

	pr, pw := io.Pipe()
	writerErr := make(chan error, 1)
	go func() {
		defer func() { _ = pw.Close() }()
		if _, err := io.WriteString(pw, rawStreamFlushTestEvent("first")); err != nil {
			writerErr <- err
			return
		}
		// 上游在客户端收到第一个事件之前不再发送任何数据。
		select {
		case <-firstEventFlushed:
		case <-time.After(5 * time.Second):
			writerErr <- errors.New("first event was not flushed before waiting for more upstream data")
			return
		}
		_, err := io.WriteString(pw, rawStreamFlushTestEvent("second")+"data: [DONE]\n\n")
		writerErr <- err
	}()

	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       pr,
	}}
	svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}

	_, err := svc.forwardAsRawChatCompletions(context.Background(), c, rawChatCompletionsTestAccount(), body, "")
	require.NoError(t, err)
	require.NoError(t, <-writerErr)
	require.Contains(t, w.Body.String(), `"content":"second"`)
	require.Contains(t, w.Body.String(), "data: [DONE]")
}
