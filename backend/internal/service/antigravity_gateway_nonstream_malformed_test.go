package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"testing/iotest"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAntigravityMalformedFunctionCallNonStreaming(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, frames string
		malformed    bool
	}{
		{"signature only", `data: {"response":{"candidates":[{"finishReason":"MALFORMED_FUNCTION_CALL","content":{"parts":[{"thoughtSignature":"test-signature"}]}}]}}` + "\n\n", true},
		{"direct frame", `data: {"candidates":[{"finishReason":"MALFORMED_FUNCTION_CALL"}]}` + "\n\n", true},
		{"error after tool", `data: {"response":{"candidates":[{"content":{"parts":[{"functionCall":{"name":"read","args":{}}}]}}]}}` + "\n\n" + `data: {"response":{"candidates":[{"finishReason":"MALFORMED_FUNCTION_CALL"}]}}` + "\n\n", true},
		{"error after text and usage", `data: {"response":{"candidates":[{"content":{"parts":[{"text":"partial"}]}}],"usageMetadata":{"promptTokenCount":11,"candidatesTokenCount":5}}}` + "\n\n" + `data: {"response":{"candidates":[{"finishReason":"MALFORMED_FUNCTION_CALL"}]}}` + "\n\n", true},
		{"error before usage", `data: {"response":{"candidates":[{"finishReason":"MALFORMED_FUNCTION_CALL"}]}}` + "\n\n" + `data: {"response":{"usageMetadata":{"promptTokenCount":11}}}` + "\n\n", true},
		{"normal tool", `data: {"response":{"candidates":[{"finishReason":"STOP","content":{"parts":[{"functionCall":{"name":"read","args":{}}}]}}]}}` + "\n\n", false},
		{"normal text", `data: {"response":{"candidates":[{"finishReason":"STOP","content":{"parts":[{"text":"hello"}]}}],"usageMetadata":{"promptTokenCount":11,"candidatesTokenCount":5}}}` + "\n\n", false},
	} {
		for _, protocol := range []string{"responses", "chat", "messages"} {
			t.Run(protocol+"/"+tc.name, func(t *testing.T) {
				svc := newAntigravityCompatService(config.GatewayConfig{MaxLineSize: defaultMaxLineSize}, nil)
				c, rec := newAntigravityCompatContext(http.MethodPost, "/v1/"+protocol, nil)
				resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(tc.frames))}
				t.Cleanup(func() { require.NoError(t, resp.Body.Close()) })
				handle := svc.handleResponsesNonStreamingFromAntigravity
				switch protocol {
				case "chat":
					handle = svc.handleChatCompletionsNonStreamingFromAntigravity
				case "messages":
					handle = svc.handleClaudeStreamToNonStreaming
				}
				result, err := handle(c, resp, time.Now(), "test-model")
				if tc.malformed {
					require.Error(t, err)
					require.Nil(t, result, "failed requests must not return successful usage")
					require.Equal(t, http.StatusBadGateway, rec.Code)
					require.Contains(t, rec.Body.String(), `"type":"malformed_function_call"`)
					require.Contains(t, rec.Body.String(), "MALFORMED_FUNCTION_CALL")
					require.NotContains(t, rec.Body.String(), `"status":"completed"`)
					require.NotContains(t, rec.Body.String(), "[DONE]")
				} else {
					require.NoError(t, err)
					require.Equal(t, http.StatusOK, rec.Code)
					require.NotNil(t, result)
					if tc.name == "normal text" {
						require.Contains(t, rec.Body.String(), "hello")
						require.Equal(t, 11, result.usage.InputTokens)
						require.Equal(t, 5, result.usage.OutputTokens)
					} else {
						require.Contains(t, rec.Body.String(), "read")
					}
				}
			})
		}
	}
}

func TestAntigravityNonStreamingCollection_PreservesCancellation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, cancelErr := range []error{context.Canceled, context.DeadlineExceeded} {
		for _, protocol := range []string{"responses", "chat", "messages"} {
			t.Run(protocol+"/"+cancelErr.Error(), func(t *testing.T) {
				svc := newAntigravityCompatService(config.GatewayConfig{}, nil)
				c, rec := newAntigravityCompatContext(http.MethodPost, "/v1/"+protocol, nil)
				partial := `data: {"response":{"candidates":[{"content":{"parts":[{"text":"partial"}]}}]}}` + "\n\n"
				resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(io.MultiReader(strings.NewReader(partial), iotest.ErrReader(cancelErr)))}
				handle := svc.handleResponsesNonStreamingFromAntigravity
				switch protocol {
				case "chat":
					handle = svc.handleChatCompletionsNonStreamingFromAntigravity
				case "messages":
					handle = svc.handleClaudeStreamToNonStreaming
				}
				result, err := handle(c, resp, time.Now(), "test-model")
				require.ErrorIs(t, err, cancelErr)
				require.Nil(t, result)
				require.Empty(t, rec.Body.String())
				require.NoError(t, resp.Body.Close())
			})
		}
	}
}

type antigravityMalformedPipeBody struct {
	*io.PipeReader
	closed chan struct{}
	once   sync.Once
}

func (b *antigravityMalformedPipeBody) Close() error {
	b.once.Do(func() { close(b.closed) })
	return b.PipeReader.Close()
}

func TestAntigravityMalformedFunctionCallNonStreaming_ClosesLiveUpstream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, protocol := range []string{"chat", "responses"} {
		t.Run(protocol, func(t *testing.T) {
			reader, writer := io.Pipe()
			body := &antigravityMalformedPipeBody{PipeReader: reader, closed: make(chan struct{})}
			t.Cleanup(func() { _ = body.Close(); _ = writer.Close() })
			written := make(chan error, 1)
			go func() {
				_, err := fmt.Fprint(writer, `data: {"response":{"candidates":[{"finishReason":"MALFORMED_FUNCTION_CALL"}]}}`+"\n\n")
				written <- err
				// Keep the stream open after the error frame until its owner closes it.
			}()
			upstream := &queuedHTTPUpstreamStub{responses: []*http.Response{{StatusCode: http.StatusOK, Header: make(http.Header), Body: body}}}
			svc := newAntigravityCompatService(config.GatewayConfig{MaxLineSize: defaultMaxLineSize}, upstream)
			request := []byte(`{"model":"gemini-3.1-pro-high","input":"hello"}`)
			path := "/v1/responses"
			if protocol == "chat" {
				path = "/v1/chat/completions"
				request = []byte(`{"model":"gemini-3.1-pro-high","messages":[{"role":"user","content":"hello"}]}`)
			}
			c, rec := newAntigravityCompatContext(http.MethodPost, path, request)
			completed := make(chan error, 1)
			go func() {
				var result *ForwardResult
				var err error
				if protocol == "chat" {
					result, err = svc.ForwardAsChatCompletions(context.Background(), c, newAntigravityCompatAccount(AccountTypeOAuth), request, nil)
				} else {
					result, err = svc.ForwardAsResponses(context.Background(), c, newAntigravityCompatAccount(AccountTypeOAuth), request, nil)
				}
				if result != nil {
					err = fmt.Errorf("unexpected successful result: %+v", result)
				}
				completed <- err
			}()
			select {
			case err := <-completed:
				require.Error(t, err)
			case <-time.After(2 * time.Second):
				t.Fatal("malformed terminal must return without waiting for upstream EOF")
			}
			require.Equal(t, http.StatusBadGateway, rec.Code)
			require.Contains(t, rec.Body.String(), `"type":"malformed_function_call"`)
			select {
			case <-body.closed:
			default:
				t.Fatal("forwarding must close the upstream body after the malformed terminal")
			}
			require.NoError(t, <-written)
		})
	}
}
