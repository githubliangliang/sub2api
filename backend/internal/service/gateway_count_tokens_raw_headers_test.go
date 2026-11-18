//go:build unit

package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func headerKeysFold(h http.Header, name string) []string {
	var keys []string
	for k := range h {
		if strings.EqualFold(k, name) {
			keys = append(keys, k)
		}
	}
	return keys
}

// The count_tokens passthrough builder copies client headers under their wire casing
// (lowercase content-type / anthropic-version). It must look those up under the same
// key before filling in defaults, so the upstream request carries exactly one of each.
func TestCountTokensPassthroughSendsOneContentTypeAndVersion(t *testing.T) {
	tests := []struct {
		name        string
		clientSends bool
		wantType    string
		wantVersion string
	}{
		{"client sent both", true, "application/json; charset=utf-8", "2024-01-01"},
		{"client sent neither", false, "application/json", "2023-06-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages/count_tokens", nil)
			if tt.clientSends {
				c.Request.Header.Set("Anthropic-Version", tt.wantVersion)
				c.Request.Header.Set("Content-Type", tt.wantType)
			}
			account := &Account{
				ID:          1,
				Platform:    PlatformAnthropic,
				Type:        AccountTypeAPIKey,
				Credentials: map[string]any{"api_key": "test-key"},
				Status:      StatusActive,
				Schedulable: true,
			}
			svc := &GatewayService{cfg: &config.Config{}}
			body := []byte(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"hi"}]}`)

			req, err := svc.buildCountTokensRequestAnthropicAPIKeyPassthrough(context.Background(), c, account, body, "test-key")
			require.NoError(t, err)

			require.Equal(t, []string{"content-type"}, headerKeysFold(req.Header, "content-type"))
			require.Equal(t, []string{tt.wantType}, req.Header["content-type"])
			require.Equal(t, []string{"anthropic-version"}, headerKeysFold(req.Header, "anthropic-version"))
			require.Equal(t, []string{tt.wantVersion}, req.Header["anthropic-version"])
		})
	}
}

// The Claude mimic debug line must report headers stored under their wire casing.
func TestBuildClaudeMimicDebugLineReadsRawCasedHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "https://upstream.example/v1/messages", nil)
	setHeaderRaw(req.Header, "anthropic-version", "2023-06-01")
	setHeaderRaw(req.Header, "anthropic-beta", "beta-a")
	setHeaderRaw(req.Header, "content-type", "application/json")
	setHeaderRaw(req.Header, "x-app", "cli")
	setHeaderRaw(req.Header, "X-Stainless-OS", "Linux")

	line := buildClaudeMimicDebugLine(req, []byte(`{}`), nil, "oauth", true)

	require.Contains(t, line, `anthropic-version="2023-06-01"`)
	require.Contains(t, line, `anthropic-beta="beta-a"`)
	require.Contains(t, line, `content-type="application/json"`)
	require.Contains(t, line, `x-app="cli"`)
	require.Contains(t, line, `x-stainless-os="Linux"`)
}

func TestCountTokensPassthroughHeadersOnHTTPWire(t *testing.T) {
	for _, mode := range []string{"defaults", "mixed_case_multivalue", "overrides"} {
		t.Run(mode, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages/count_tokens", nil)
			wantType := []string{"application/json"}
			wantVersion := []string{"2023-06-01"}
			wantBeta := []string{"per-turn-control-2026-07-01"}
			account := &Account{ID: 1, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Credentials: map[string]any{}}
			if mode != "defaults" {
				c.Request.Header = http.Header{
					"CONTENT-TYPE":      {"application/json; charset=utf-8", "application/custom-json"},
					"anthropic-version": {"2024-01-01", "2025-01-01"},
					"Anthropic-Beta":    wantBeta, "Authorization": {"Bearer client-secret"}, "Cookie": {"client=secret"},
				}
				wantType = c.Request.Header["CONTENT-TYPE"]
				wantVersion = c.Request.Header["anthropic-version"]
			} else {
				c.Request.Header.Set("Anthropic-Beta", wantBeta[0])
			}
			if mode == "overrides" {
				account.Credentials[credKeyHeaderOverrideEnabled] = true
				account.Credentials[credKeyHeaderOverrides] = map[string]any{"anthropic-version": "override-version", "anthropic-beta": "override-beta"}
				wantVersion = []string{"override-version"}
				wantBeta = []string{"override-beta"}
			}
			body := []byte(`{"model":"claude-sonnet-4-6","messages":[{"role":"system","content":[],"output_config":{"effort":"high"}}]}`)
			req, err := (&GatewayService{cfg: &config.Config{}}).buildCountTokensRequestAnthropicAPIKeyPassthrough(context.Background(), c, account, body, "account-key")
			require.NoError(t, err)
			type receivedRequest struct {
				headers http.Header
				body    []byte
			}
			receivedCh := make(chan receivedRequest, 1)
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedBody, _ := io.ReadAll(r.Body)
				receivedCh <- receivedRequest{r.Header.Clone(), receivedBody}
				w.WriteHeader(http.StatusOK)
			}))
			defer upstream.Close()
			local, err := http.NewRequestWithContext(req.Context(), req.Method, upstream.URL, req.Body)
			require.NoError(t, err)
			local.Header = req.Header
			resp, err := upstream.Client().Do(local)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			received := <-receivedCh
			require.Equal(t, wantType, received.headers.Values("Content-Type"))
			require.Equal(t, wantVersion, received.headers.Values("Anthropic-Version"))
			require.Equal(t, wantBeta, received.headers.Values("Anthropic-Beta"))
			require.Equal(t, []string{"account-key"}, received.headers.Values("X-Api-Key"))
			require.Empty(t, received.headers.Values("Authorization"))
			require.Empty(t, received.headers.Values("Cookie"))
			require.Equal(t, mode != "overrides", strings.Contains(string(received.body), "output_config"))
		})
	}
}
