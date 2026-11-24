package service

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type upstream025RedactionHTTP struct {
	client *http.Client
	target *url.URL
}

func (u upstream025RedactionHTTP) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	local := req.Clone(req.Context())
	local.URL.Scheme, local.URL.Host = u.target.Scheme, u.target.Host
	local.Host = u.target.Host
	return u.client.Do(local)
}

func (u upstream025RedactionHTTP) DoWithTLS(req *http.Request, proxy string, accountID int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxy, accountID, concurrency)
}

func TestUpstream025AccountRedactionActualHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, status := range []int{http.StatusBadRequest, http.StatusForbidden} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			const message = "Ⱥ K Your input exceeds the context window at RELAY.SYNTHETIC.EXAMPLE for sk-synthetic-api access-synthetic refresh-synthetic?key=query-synthetic"
			var requests atomic.Int64
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				require.Equal(t, "Bearer sk-synthetic-api", r.Header.Get("Authorization"))
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Server", "synthetic-private-server")
				w.Header().Set("Set-Cookie", "synthetic-private-cookie")
				w.WriteHeader(status)
				responseMessage := message
				code := "context_length_exceeded"
				if status == http.StatusForbidden {
					responseMessage = "Private relay.synthetic.example sk-synthetic-api"
					code = "invalid_encrypted_content"
				}
				_ = json.NewEncoder(w).Encode(gin.H{"error": gin.H{"code": code, "message": responseMessage}})
			}))
			defer upstream.Close()
			target, err := url.Parse(upstream.URL)
			require.NoError(t, err)
			account := &Account{ID: 925, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true,
				Credentials: map[string]any{"api_key": "sk-synthetic-api", "access_token": "access-synthetic", "refresh_token": "refresh-synthetic", "base_url": "https://relay.synthetic.example/v1"},
				Extra:       map[string]any{"openai_passthrough": true}}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream025RedactionHTTP{client: upstream.Client(), target: target}}
			forwarded := make(chan bool, 1)
			gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() { forwarded <- true }()
				c, _ := gin.CreateTestContext(w)
				c.Request = r
				body, readErr := io.ReadAll(r.Body)
				require.NoError(t, readErr)
				result, forwardErr := svc.Forward(r.Context(), c, account, body)
				require.Error(t, forwardErr)
				require.Nil(t, result, "errors must not produce successful usage results")
				if status == http.StatusBadRequest {
					require.Contains(t, c.GetString(OpsUpstreamErrorMessageKey), "sk-synthetic-api", "internal diagnostic ownership is unchanged")
				}
			}))
			defer gateway.Close()
			client := &http.Client{Timeout: 5 * time.Second}
			response, err := client.Post(gateway.URL+"/v1/responses", "application/json", strings.NewReader(`{"model":"gpt-6-astra","input":"hi"}`))
			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()
			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.True(t, <-forwarded)
			require.EqualValues(t, 1, requests.Load())
			require.Empty(t, response.Header.Get("Server"))
			require.Empty(t, response.Header.Get("Set-Cookie"))
			var decoded map[string]any
			decoder := json.NewDecoder(strings.NewReader(string(body)))
			require.NoError(t, decoder.Decode(&decoded))
			require.ErrorIs(t, decoder.Decode(&decoded), io.EOF, "write exactly one error envelope")
			for _, sensitive := range []string{"relay.synthetic.example", "sk-synthetic-api", "access-synthetic", "refresh-synthetic", "query-synthetic", "synthetic-private"} {
				require.NotContains(t, strings.ToLower(string(body)), sensitive)
			}
			if status == http.StatusBadRequest {
				require.Equal(t, status, response.StatusCode)
				require.Contains(t, gjson.GetBytes(body, "error.message").String(), "Your input exceeds the context window")
			} else {
				require.Equal(t, http.StatusBadGateway, response.StatusCode)
				require.Equal(t, "Upstream access denied", gjson.GetBytes(body, "error.message").String())
			}
		})
	}
}
