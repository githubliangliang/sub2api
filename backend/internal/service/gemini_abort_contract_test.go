//go:build unit

package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type geminiAbortProtocol struct {
	name, path, body string
	forward          func(*GeminiMessagesCompatService, *gin.Context, string, []byte) (*ForwardResult, error)
}

var geminiAbortProtocols = []geminiAbortProtocol{
	{"native", "/v1beta/models/gemini-2.5-flash:streamGenerateContent", `{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`,
		func(s *GeminiMessagesCompatService, c *gin.Context, model string, body []byte) (*ForwardResult, error) {
			return s.ForwardNative(c.Request.Context(), c, geminiAbortAPIKeyTestAccount(), model, "streamGenerateContent", true, body)
		}},
	{"messages", "/v1/messages", `{"model":"gemini-2.5-flash","stream":true,"max_tokens":16,"messages":[{"role":"user","content":"hi"}]}`,
		func(s *GeminiMessagesCompatService, c *gin.Context, _ string, body []byte) (*ForwardResult, error) {
			return s.Forward(c.Request.Context(), c, geminiAbortAPIKeyTestAccount(), body)
		}},
	{"chat", "/v1/chat/completions", `{"model":"gemini-2.5-flash","stream":true,"messages":[{"role":"user","content":"hi"}]}`,
		func(s *GeminiMessagesCompatService, c *gin.Context, _ string, body []byte) (*ForwardResult, error) {
			return s.ForwardAsChatCompletions(c.Request.Context(), c, geminiAbortAPIKeyTestAccount(), body)
		}},
}

func TestGeminiAbortDoesNotMisclassifyLiveUpstreamCancellation(t *testing.T) {
	for _, p := range geminiAbortProtocols {
		t.Run(p.name, func(t *testing.T) {
			svc, c, _ := newGeminiAbortingStreamService(t, p.path, []byte(p.body), geminiAbortUsageChunk, false)
			upstream := svc.httpUpstream.(*geminiCompatHTTPUpstreamStub)
			upstream.response.Body.(*geminiAbortingStreamBody).abortErr = context.Canceled
			result, err := p.forward(svc, c, "gemini-2.5-flash", []byte(p.body))
			require.ErrorIs(t, err, context.Canceled)
			require.Nil(t, result, "an upstream cancellation is not proof that the client has left")
		})
	}
}

func TestGeminiAbortIgnoresEvenValidJSONOnErroredResidualLine(t *testing.T) {
	for _, p := range geminiAbortProtocols {
		t.Run(p.name, func(t *testing.T) {
			stream := geminiAbortUsageChunk + strings.TrimRight(geminiAbortFinalChunk, "\n")
			svc, c, _ := newGeminiAbortingStreamService(t, p.path, []byte(p.body), stream, true)
			result, err := p.forward(svc, c, "gemini-2.5-flash", []byte(p.body))
			require.NoError(t, err)
			require.True(t, result.ClientDisconnect)
			require.Equal(t, 10, result.Usage.InputTokens)
			require.Equal(t, 2, result.Usage.OutputTokens, "an unterminated read-error fragment is not an observed SSE frame")
		})
	}
}

type geminiAbortOnTextWriter struct {
	*httptest.ResponseRecorder
	failed                    bool
	writesAfter, flushesAfter int
}

func (w *geminiAbortOnTextWriter) Write(data []byte) (int, error) {
	if w.failed {
		w.writesAfter++
		return 0, io.ErrClosedPipe
	}
	if bytes.Contains(data, []byte("drawing")) {
		w.failed = true
		return 0, io.ErrClosedPipe
	}
	return w.ResponseRecorder.Write(data)
}
func (w *geminiAbortOnTextWriter) WriteString(data string) (int, error) { return w.Write([]byte(data)) }
func (w *geminiAbortOnTextWriter) Flush() {
	if w.failed {
		w.flushesAfter++
	}
	w.ResponseRecorder.Flush()
}

func TestGeminiAbortWriteFailureKeepsOnlyObservedFrameUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, p := range geminiAbortProtocols {
		t.Run(p.name, func(t *testing.T) {
			writer := &geminiAbortOnTextWriter{ResponseRecorder: httptest.NewRecorder()}
			c, _ := gin.CreateTestContext(writer)
			c.Request = httptest.NewRequest(http.MethodPost, p.path, strings.NewReader(p.body))
			svc, upstream := newGeminiAbortAPIKeyTestService()
			upstream.response = &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(geminiAbortUsageChunk + geminiAbortFinalChunk))}
			result, err := p.forward(svc, c, "gemini-2.5-flash", []byte(p.body))
			require.NoError(t, err)
			require.True(t, writer.failed)
			require.True(t, result.ClientDisconnect)
			require.Equal(t, 2, result.Usage.OutputTokens)
			require.Zero(t, writer.writesAfter)
			require.Zero(t, writer.flushesAfter)
		})
	}
}

func TestGeminiAbortImagesRequireActualObservedParts(t *testing.T) {
	for _, p := range geminiAbortProtocols {
		for _, tc := range []struct {
			name, parts string
			want        int
		}{
			{"finished text is not an image", `{"text":"done"}`, 0},
			{"observed file image", `{"fileData":{"mimeType":"image/png","fileUri":"gs://bucket/image.png"}}`, 1},
			{"empty file reference", `{"fileData":{"mimeType":"image/png","fileUri":""}}`, 0},
		} {
			t.Run(p.name+"/"+tc.name, func(t *testing.T) {
				model := "gemini-3-pro-image"
				body := strings.ReplaceAll(p.body, "gemini-2.5-flash", model)
				stream := `data: {"candidates":[{"content":{"parts":[` + tc.parts + `]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":5}}` + "\n\n"
				svc, c, _ := newGeminiAbortingStreamService(t, p.path, []byte(body), stream, true)
				result, err := p.forward(svc, c, model, []byte(body))
				require.NoError(t, err)
				require.True(t, result.ClientDisconnect)
				require.Equal(t, tc.want, result.ImageCount, "finishReason alone must never restore model-name image estimates")
			})
		}
	}
}

func TestGeminiAbortWriteFailureWithoutUsageIsNotSuccess(t *testing.T) {
	for _, p := range geminiAbortProtocols {
		t.Run(p.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(&geminiFailingResponseWriter{header: http.Header{}})
			c.Request = httptest.NewRequest(http.MethodPost, p.path, strings.NewReader(p.body))
			svc, upstream := newGeminiAbortAPIKeyTestService()
			upstream.response = &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(geminiAbortTextChunk))}
			result, err := p.forward(svc, c, "gemini-2.5-flash", []byte(p.body))
			require.ErrorIs(t, err, io.ErrClosedPipe)
			require.Nil(t, result)
		})
	}
}

func TestGeminiAbortObservedImageWithoutTokenUsageIsBillable(t *testing.T) {
	for _, p := range geminiAbortProtocols {
		t.Run(p.name, func(t *testing.T) {
			stream := `data: {"candidates":[{"content":{"parts":[{"inlineData":{"mimeType":"image/png","data":"aGVsbG8="}}]}}]}` + "\n\n"
			svc, c, _ := newGeminiAbortingStreamService(t, p.path, []byte(p.body), stream, true)
			result, err := p.forward(svc, c, "gemini-2.5-flash", []byte(p.body))
			require.NoError(t, err)
			require.True(t, result.ClientDisconnect)
			require.Equal(t, 1, result.ImageCount)
			require.Zero(t, result.Usage.InputTokens)
			require.Zero(t, result.Usage.OutputTokens)
		})
	}
}
