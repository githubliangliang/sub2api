package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type promptLocalHTTP struct {
	target *url.URL
	client *http.Client
}

func (p promptLocalHTTP) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	clone := req.Clone(req.Context())
	u := *clone.URL
	u.Scheme = p.target.Scheme
	u.Host = p.target.Host
	clone.URL = &u
	return p.client.Do(clone)
}
func (p promptLocalHTTP) DoWithTLS(req *http.Request, proxy string, id int64, n int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return p.Do(req, proxy, id, n)
}

func TestAntigravityAccountTestPromptReachesHTTP(t *testing.T) {
	for _, model := range []string{"gemini-2.5-flash", "claude-sonnet-4-5"} {
		for _, prompt := range []string{"", "你好\ninspect this task"} {
			t.Run(model+prompt, func(t *testing.T) {
				seen := make(chan []byte, 1)
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, err := io.ReadAll(r.Body)
					if err != nil {
						t.Error(err)
						w.WriteHeader(500)
						return
					}
					seen <- body
					w.Header().Set("Content-Type", "text/event-stream")
					_, _ = io.WriteString(w, "data: {\"response\":{\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"actual fixture reply\"}]},\"finishReason\":\"STOP\"}]}}\n\n")
				}))
				defer server.Close()
				target, err := url.Parse(server.URL)
				require.NoError(t, err)
				gateway := newAntigravityCompatService(config.GatewayConfig{}, promptLocalHTTP{target, server.Client()})
				account := &Account{ID: 91, Platform: PlatformAntigravity, Type: AccountTypeOAuth, Status: StatusActive, Credentials: map[string]any{"access_token": "synthetic", "project_id": "fixture-project", "model_mapping": map[string]any{model: model}}}
				svc := &AccountTestService{antigravityGatewayService: gateway}
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Request = httptest.NewRequest(http.MethodPost, "/test", nil).WithContext(context.Background())
				require.NoError(t, svc.testAntigravityAccountConnection(c, account, model, prompt))
				payload := <-seen
				want := prompt
				if strings.TrimSpace(want) == "" {
					want = "hi"
				}
				require.Equal(t, want, gjson.GetBytes(payload, "request.contents.0.parts.0.text").String())
				require.EqualValues(t, 256, gjson.GetBytes(payload, "request.generationConfig.maxOutputTokens").Int())
				require.Contains(t, rec.Body.String(), "actual fixture reply")
			})
		}
	}
}
