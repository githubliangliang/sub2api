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
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// Exercise the real native and buffered response consumers.
func TestGrokNativeResponseModelAudit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, streaming := range []bool{false, true} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		body := []byte(`{"model":"grok","input":"hi","stream":false}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
		response := `{"id":"resp_audit","model":"grok-4.7","status":"completed","output":[],"usage":{"input_tokens":2,"output_tokens":1}}`
		contentType := "application/json"
		if streaming {
			response = "data: {\"type\":\"response.completed\",\"response\":" + response + "}\n\n"
			contentType = "text/event-stream"
		}
		upstream := &httpUpstreamRecorder{resp: &http.Response{
			StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{contentType}},
			Body: io.NopCloser(strings.NewReader(response)),
		}}
		account := &Account{ID: 19019, Platform: PlatformGrok, Type: AccountTypeAPIKey,
			Credentials: map[string]any{"api_key": "synthetic-key", "base_url": "https://api.x.ai/v1"}}
		result, err := (&OpenAIGatewayService{httpUpstream: upstream}).forwardGrokResponses(
			context.Background(), c, account, body, "grok", streaming, time.Now())
		if err != nil || result == nil {
			t.Fatalf("fixture failed: result=%v, err=%v", result, err)
		}
		observed := observedUpstreamResponseModel(c)
		if observed != "grok-4.7" || result.UpstreamResponseModel != "grok-4.7" {
			t.Fatalf("response audit mismatch: observed=%q, returned=%q", observed, result.UpstreamResponseModel)
		}
		if result.Model != "grok" || result.UpstreamModel != "grok-4.6" {
			t.Fatalf("request/mapping model changed: %+v", result)
		}
	}
}

func TestGrokBufferedPriorAuditIsBoundedAndOtherProvidersKeepTheirContract(t *testing.T) {
	prior := newOpenAICompatPriorModelAudit(&Account{Platform: PlatformGrok})
	for i := 0; i < 1000; i++ {
		prefix := "first-"
		if i > 0 {
			prefix = "other-"
		}
		prior.record(&apicompat.ResponsesStreamEvent{Type: "response.created", Response: &apicompat.ResponsesResponse{Model: prefix + strings.Repeat("界", 4000)}})
	}
	require.Equal(t, upstreamResponseModelMaxLength, utf8.RuneCountInString(prior.observer.first))
	require.Empty(t, prior.observer.terminal)
	require.True(t, prior.observer.conflict)
	var observer upstreamResponseModelObserver
	prior.apply(&observer)
	observer.Observe("Grok-Custom-Terminal", true)
	require.Equal(t, "Grok-Custom-Terminal", observer.Model())
	require.True(t, observer.Conflict())
	for _, platform := range []string{PlatformOpenAI, PlatformAnthropic, PlatformGemini} {
		other := newOpenAICompatPriorModelAudit(&Account{Platform: platform})
		other.record(&apicompat.ResponsesStreamEvent{Type: "response.created", Response: &apicompat.ResponsesResponse{Model: "prior"}})
		var observation upstreamResponseModelObserver
		other.apply(&observation)
		observation.Observe("terminal", true)
		require.False(t, observation.Conflict())
		require.Empty(t, other.observer.first)
	}
}

func TestGrokBufferedResponseModelConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, messages := range []bool{false, true} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
		stream := "data: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_audit\",\"model\":\"grok-4.6\"}}\n\n" +
			"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_audit\",\"model\":\"grok-4.7\",\"status\":\"completed\",\"output\":[{\"id\":\"msg_audit\",\"type\":\"message\",\"role\":\"assistant\",\"content\":[{\"type\":\"output_text\",\"text\":\"ok\"}]}],\"usage\":{\"input_tokens\":2,\"output_tokens\":1}}}\n\n"
		response := &http.Response{StatusCode: http.StatusOK,
			Header: http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:   io.NopCloser(strings.NewReader(stream))}
		account := &Account{ID: 19020, Platform: PlatformGrok}
		svc := &OpenAIGatewayService{}
		var result *OpenAIForwardResult
		var err error
		if messages {
			result, err = svc.handleAnthropicBufferedStreamingResponse(response, c, account, "grok", "grok-4.7", "grok-4.7", time.Now())
		} else {
			result, err = svc.handleChatBufferedStreamingResponse(response, c, account, "grok", "grok-4.7", "grok-4.7", time.Now())
		}
		if err != nil || result == nil {
			t.Fatalf("fixture failed: result=%v, err=%v", result, err)
		}
		if result.UpstreamResponseModel != "grok-4.7" || !result.UpstreamResponseModelConflict {
			t.Fatalf("response audit mismatch: model=%q, conflict=%v", result.UpstreamResponseModel, result.UpstreamResponseModelConflict)
		}
		if result.Model != "grok" || result.UpstreamModel != "grok-4.7" {
			t.Fatalf("request/mapping model changed: %+v", result)
		}
	}
}
