package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func s22Bodies() ([]byte, []byte) {
	messages := make([]map[string]string, 200)
	tools := make([]map[string]any, 200)
	for i := range messages {
		messages[i] = map[string]string{"role": "assistant", "content": strings.Repeat("x", 128), "reasoning_content": "thought <tag> 中文"}
		tools[i] = map[string]any{"type": "function", "name": fmt.Sprintf("custom_%d", i), "input_schema": map[string]string{"type": "object"}}
	}
	body, _ := json.Marshal(map[string]any{"model": "claude-sonnet-4-6", "tools": tools, "messages": messages})
	systems := make([]map[string]string, 8)
	for i := range systems {
		systems[i] = map[string]string{"type": "text", "text": "x-anthropic-billing-header: cc_version=2.1.1.aaaa;"}
	}
	billing, _ := json.Marshal(map[string]any{"system": systems, "messages": messages})
	return body, billing
}

func TestS22RequestCandidatesPreserveBytes(t *testing.T) {
	body, billing := s22Bodies()
	for _, input := range [][]byte{body, []byte(`{"messages":[{"role":"assistant","reasoning_content":"中文 <tag> \"quote\"\n","reasoning":null},{"role":"assistant","thinking":"keep","reasoning_content":"other"}]}`)} {
		require.Equal(t, legacyS22NormalizeOllamaCloudChatCompletionsRequest(input), normalizeOllamaCloudChatCompletionsRequest(input))
	}
	for _, input := range [][]byte{body, []byte(`{"tools":[{"name":"custom","defer_loading":true,"cache_control":{"type":"ephemeral","ttl":"5m"}}]}`), []byte(`{"tools":[]}`)} {
		before, candidate := legacyS22BuildToolNameRewriteFromBody(input), buildToolNameRewriteFromBody(input)
		if before == nil {
			require.Nil(t, candidate)
		} else {
			require.Equal(t, before.Forward, candidate.Forward)
			require.Equal(t, before.Reverse, candidate.Reverse)
			// Equal-length entries come from a map and have no ordering contract.
			require.ElementsMatch(t, before.ReverseOrdered, candidate.ReverseOrdered)
			for i := 1; i < len(candidate.ReverseOrdered); i++ {
				require.GreaterOrEqual(t, len(candidate.ReverseOrdered[i-1][0]), len(candidate.ReverseOrdered[i][0]))
			}
		}
		require.Equal(t, legacyS22ApplyToolsLastCacheBreakpoint(input), applyToolsLastCacheBreakpoint(input))
	}
	require.Equal(t, legacyS22SyncBillingHeaderVersion(billing, "claude-cli/2.1.321"), syncBillingHeaderVersion(billing, "claude-cli/2.1.321"))
	for _, input := range [][]byte{body, []byte(`{"system":[{"cache_control":{"type":"ephemeral","ttl":"5m"}},{"cache_control":{"type":"ephemeral","ttl":"1h"}}]}`), []byte(`{"system":[{"type":"thinking","cache_control":{"type":"ephemeral"}}]}`), []byte(`{"system":[{"cache_contro\u006c":{"type":"ephemeral","ttl":"\u0031h"}}]}`)} {
		require.Equal(t, legacyS22EnforceCacheControlLimit(input), enforceCacheControlLimit(input))
	}
}

var s22BenchmarkBytes []byte
var s22BenchmarkMapping *ToolNameRewrite
var s22BenchmarkString string

func BenchmarkS22Requests(b *testing.B) {
	body, billing := s22Bodies()
	ascii := bytes.ReplaceAll(body, []byte(`thought \u003ctag\u003e 中文`), []byte(`plain thought`))
	for _, candidate := range []struct {
		name     string
		input    []byte
		old, new func([]byte) []byte
	}{
		{"ollama", body, legacyS22NormalizeOllamaCloudChatCompletionsRequest, normalizeOllamaCloudChatCompletionsRequest},
		{"cache_limit_escaped", body, legacyS22EnforceCacheControlLimit, enforceCacheControlLimit},
		{"cache_limit_ascii", ascii, legacyS22EnforceCacheControlLimit, enforceCacheControlLimit},
		{"tool_cache", body, legacyS22ApplyToolsLastCacheBreakpoint, applyToolsLastCacheBreakpoint},
		{"billing", billing, func(v []byte) []byte { return legacyS22SyncBillingHeaderVersion(v, "claude-cli/2.1.321") }, func(v []byte) []byte { return syncBillingHeaderVersion(v, "claude-cli/2.1.321") }},
	} {
		for _, version := range []struct {
			name string
			fn   func([]byte) []byte
		}{{"before", candidate.old}, {"candidate", candidate.new}} {
			b.Run(candidate.name+"/"+version.name, func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					s22BenchmarkBytes = version.fn(candidate.input)
				}
			})
		}
	}
	for _, version := range []struct {
		name string
		fn   func([]byte) *ToolNameRewrite
	}{{"before", legacyS22BuildToolNameRewriteFromBody}, {"candidate", buildToolNameRewriteFromBody}} {
		b.Run("tool_map/"+version.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				s22BenchmarkMapping = version.fn(body)
			}
		})
	}
	chunk := strings.Repeat("plain text ", 10000)
	b.Run("reverse/before", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			s22BenchmarkString = string(reverseToolNamesIfPresent(nil, []byte(chunk)))
		}
	})
	b.Run("reverse/candidate", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			s22BenchmarkString = reverseToolNamesIfPresentString(nil, chunk)
		}
	})
}

func BenchmarkS22RawChatFlush(b *testing.B) {
	gin.SetMode(gin.TestMode)
	body := strings.Repeat(`data: {"model":"gpt-5.4","choices":[{"index":0,"delta":{"content":"x"}}]}`+"\n\n", 100) + `data: {"choices":[],"usage":{"prompt_tokens":3,"completion_tokens":100}}` + "\n\ndata: [DONE]\n\n"
	b.ReportAllocs()
	for b.Loop() {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
		resp := &http.Response{Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
		_, _ = (&OpenAIGatewayService{}).streamRawChatCompletions(c, resp, &Account{Platform: PlatformOpenAI}, "gpt-5.4", "gpt-5.4", "gpt-5.4", nil, nil, time.Now(), 100)
	}
}
