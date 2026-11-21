//go:build unit

package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestUpstream024ClaudeTestTemperature(t *testing.T) {
	for _, tc := range []struct {
		model string
		omit  bool
	}{
		{"claude-opus-4-7", true},
		{"claude-opus-4-8", true},
		{"anthropic/claude-opus-5.5", true},
		{"us.anthropic.claude-opus-4-8-20261001-v1:0", true},
		{"claude-sonnet-4-6", false},
		{"claude-sonnet-5-5", false},
		{"my-custom-model", false},
	} {
		t.Run(tc.model, func(t *testing.T) {
			payload, err := createTestPayload(tc.model)
			require.NoError(t, err)
			_, present := payload["temperature"]
			require.Equal(t, !tc.omit, present)
			require.Equal(t, tc.model, payload["model"])
			require.Equal(t, 1024, payload["max_tokens"])
			require.Equal(t, true, payload["stream"])
		})
	}
}

func TestUpstream024ClaudeTestSerializedPayload(t *testing.T) {
	for _, tc := range []struct {
		name    string
		model   string
		bedrock bool
		mapped  string
		omit    bool
	}{
		{name: "explicit opus", model: "claude-opus-4-7", omit: true},
		{name: "mapped alias", model: "client-alias", mapped: "claude-opus-4-8", omit: true},
		{name: "sonnet control", model: "claude-sonnet-4-6"},
		{name: "bedrock opus", model: "claude-opus-4-8", bedrock: true, omit: true},
		{name: "bedrock sonnet", model: "claude-sonnet-4-6", bedrock: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			account := &Account{ID: 100, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "synthetic-test-key"}}
			if tc.mapped != "" {
				account.Credentials["model_mapping"] = map[string]any{tc.model: tc.mapped}
			}
			response := cloudflareMessagesResponse()
			if tc.bedrock {
				account.Type = AccountTypeBedrock
				account.Credentials["auth_mode"] = "apikey"
				response = newJSONResponse(http.StatusOK, `{"content":[{"type":"text","text":"hello"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
			}
			upstream := &queuedHTTPUpstream{responses: []*http.Response{response}}
			svc := &AccountTestService{cfg: &config.Config{}, httpUpstream: upstream}
			c, _ := newTestContext()
			require.NoError(t, svc.testClaudeAccountConnection(c, account, tc.model))
			require.Len(t, upstream.requests, 1)
			body, err := io.ReadAll(upstream.requests[0].Body)
			require.NoError(t, err)
			require.True(t, json.Valid(body))
			require.Equal(t, !tc.omit, gjson.GetBytes(body, "temperature").Exists())
			if !tc.omit {
				require.Equal(t, int64(1), gjson.GetBytes(body, "temperature").Int())
			}
			if tc.mapped != "" {
				require.Equal(t, tc.mapped, gjson.GetBytes(body, "model").String())
			}
		})
	}
}

type upstream024LocalHTTP struct {
	HTTPUpstream
	client *http.Client
	target *url.URL
}

func (u upstream024LocalHTTP) DoWithTLS(req *http.Request, _ string, _ int64, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	local := req.Clone(req.Context())
	local.URL.Scheme, local.URL.Host = u.target.Scheme, u.target.Host
	local.Host = u.target.Host
	return u.client.Do(local)
}

func TestUpstream024ClaudeTemperatureActualHTTP(t *testing.T) {
	for _, bedrock := range []bool{false, true} {
		for _, model := range []string{"claude-opus-4-8", "claude-sonnet-4-6"} {
			t.Run(fmt.Sprintf("bedrock=%t/%s", bedrock, model), func(t *testing.T) {
				captured := make(chan []byte, 1)
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, _ := io.ReadAll(r.Body)
					captured <- body
					if bedrock {
						_, _ = io.WriteString(w, `{"content":[{"type":"text","text":"hello"}]}`)
						return
					}
					w.Header().Set("Content-Type", "text/event-stream")
					_, _ = io.WriteString(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
				}))
				defer server.Close()
				target, err := url.Parse(server.URL)
				require.NoError(t, err)
				account := &Account{ID: 100, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "synthetic-test-key"}}
				if bedrock {
					account.Type = AccountTypeBedrock
					account.Credentials["auth_mode"] = "apikey"
				}
				svc := &AccountTestService{cfg: &config.Config{}, httpUpstream: upstream024LocalHTTP{client: server.Client(), target: target}}
				c, _ := newTestContext()
				require.NoError(t, svc.testClaudeAccountConnection(c, account, model))
				body := <-captured
				require.Equal(t, model == "claude-sonnet-4-6", gjson.GetBytes(body, "temperature").Exists())
			})
		}
	}
}

func TestUpstream024ImagesUsageSingleLine(t *testing.T) {
	for _, raw := range []string{
		"{\n \"input_tokens\": 901,\n \"output_tokens\": 37,\n \"input_tokens_details\": {\"text_tokens\": 3}\n}",
		`{"input_tokens":901,"output_tokens":37,"input_tokens_details":{"text_tokens":3}}`,
		"",
		`{"input_tokens":`,
	} {
		t.Run(raw, func(t *testing.T) {
			payload := buildOpenAIImagesStreamCompletedPayload("image_generation.completed", openAIResponsesImageResult{Result: "aGVsbG8=", OutputFormat: "png"}, "b64_json", 1710000000, []byte(raw))
			require.True(t, json.Valid(payload))
			require.NotContains(t, string(payload), "\n")
			require.NotContains(t, string(payload), "\r")
			require.Equal(t, "aGVsbG8=", gjson.GetBytes(payload, "b64_json").String())
			if json.Valid([]byte(raw)) {
				require.Equal(t, int64(901), gjson.GetBytes(payload, "usage.input_tokens").Int())
				require.Equal(t, int64(37), gjson.GetBytes(payload, "usage.output_tokens").Int())
			} else {
				require.False(t, gjson.GetBytes(payload, "usage").Exists())
			}
		})
	}
}

func TestUpstream024ImagesUsageTCPEvent(t *testing.T) {
	payload := buildOpenAIImagesStreamCompletedPayload("image_generation.completed", openAIResponsesImageResult{Result: "aGVsbG8="}, "b64_json", 1710000000, []byte("{\n\"input_tokens\":901,\n\"output_tokens\":37\n}"))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprintf(w, "event: image_generation.completed\ndata: %s\n\n", payload)
	}))
	defer server.Close()
	resp, err := server.Client().Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	events := strings.Split(strings.TrimSuffix(string(body), "\n\n"), "\n\n")
	require.Len(t, events, 1)
	lines := strings.Split(events[0], "\n")
	require.Len(t, lines, 2)
	data := strings.TrimPrefix(lines[1], "data: ")
	require.True(t, json.Valid([]byte(data)))
	require.Equal(t, int64(901), gjson.Get(data, "usage.input_tokens").Int())
	require.Equal(t, int64(37), gjson.Get(data, "usage.output_tokens").Int())
	require.Equal(t, "aGVsbG8=", gjson.Get(data, "b64_json").String())
}

func TestUpstream024GeminiFinalToolTurn(t *testing.T) {
	tool := map[string]any{"type": "tool_result", "tool_use_id": "call_1", "content": "done"}
	text := map[string]any{"type": "text", "text": "next instruction"}
	image := map[string]any{"type": "image", "source": map[string]any{"type": "base64", "media_type": "image/png", "data": "AQID"}}
	for _, tc := range []struct {
		name     string
		messages []any
		turns    int
		parts    []int
	}{
		{"tool and text", []any{map[string]any{"role": "user", "content": []any{tool, text}}}, 2, []int{1, 1}},
		{"two tools text image", []any{map[string]any{"role": "user", "content": []any{tool, tool, text, image}}}, 2, []int{2, 2}},
		{"pure tools", []any{map[string]any{"role": "user", "content": []any{tool, tool}}}, 1, []int{2}},
		{"pure text", []any{map[string]any{"role": "user", "content": []any{text}}}, 1, []int{1}},
		{"historical tool turn", []any{map[string]any{"role": "user", "content": []any{tool, text}}, map[string]any{"role": "assistant", "content": "answer"}}, 2, []int{2, 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := convertClaudeMessagesToGeminiContents(tc.messages, map[string]string{"call_1": "inspect"}, false)
			require.NoError(t, err)
			require.Len(t, out, tc.turns)
			for i, n := range tc.parts {
				require.Len(t, out[i].(map[string]any)["parts"], n)
			}
			if tc.name == "two tools text image" {
				trailing := out[1].(map[string]any)
				require.Equal(t, "user", trailing["role"])
				parts := trailing["parts"].([]any)
				require.Equal(t, "next instruction", parts[0].(map[string]any)["text"])
				require.Contains(t, parts[1], "inlineData")
			}
		})
	}
}

func TestUpstream024AntigravityCommentFrames(t *testing.T) {
	data := `{"response":{"candidates":[{"content":{"parts":[{"text":"hello"}]}}],"usageMetadata":{"promptTokenCount":12,"candidatesTokenCount":3}}}`
	for _, tc := range []struct {
		name string
		ua   string
		sep  string
	}{
		{"ordinary LF", "curl/8.7.1", "\n"},
		{"ordinary CRLF", "curl/8.7.1", "\r\n"},
		{"go genai", "google-genai-sdk/1.71.0 gl-go/go1.28", "\n"},
		{"python genai", "google-genai-sdk/1.20.0 gl-python/3.12.4", "\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			svc := newAntigravityTestService(&config.Config{Gateway: config.GatewayConfig{MaxLineSize: defaultMaxLineSize}})
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			c.Request.Header.Set("User-Agent", tc.ua)
			body := strings.Join([]string{": upstream ping", "", "data: " + data, "", ": tail", "", "data: [DONE]", "", ""}, tc.sep)
			result, err := svc.handleGeminiStreamingResponse(c, &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, time.Now())
			require.NoError(t, err)
			require.Equal(t, 12, result.usage.InputTokens)
			require.Equal(t, 3, result.usage.OutputTokens)
			out := rec.Body.String()
			if strings.HasPrefix(tc.name, "ordinary") {
				require.Contains(t, out, ": upstream ping\n\n")
				require.Contains(t, out, ": tail\n\n")
			} else {
				require.NotContains(t, out, ": upstream ping")
				require.NotContains(t, out, ": tail")
			}
			require.Contains(t, out, "data: [DONE]\n\n")
			require.NotContains(t, out, "\n\n\n")
			require.False(t, bytes.Contains([]byte(out), []byte("\r")))
		})
	}
}

func TestUpstream024AntigravityCommentsDisconnectDrainsUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newAntigravityTestService(&config.Config{Gateway: config.GatewayConfig{MaxLineSize: defaultMaxLineSize}})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	c.Writer = &antigravityFailingWriter{ResponseWriter: c.Writer, failAfter: 0}
	body := ": comment\n\ndata: " + `{"response":{"usageMetadata":{"promptTokenCount":27,"candidatesTokenCount":9}}}` + "\n\n"
	result, err := svc.handleGeminiStreamingResponse(c, &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, time.Now())
	require.NoError(t, err)
	require.True(t, result.clientDisconnect)
	require.Equal(t, 27, result.usage.InputTokens)
	require.Equal(t, 9, result.usage.OutputTokens)
}
