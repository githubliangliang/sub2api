//go:build unit

package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func openAICapacityShedModelRuleAccount(accountType string) *Account {
	return &Account{
		ID: 101, Platform: PlatformOpenAI, Type: accountType, Status: StatusActive, Schedulable: true,
		Credentials: map[string]any{
			"temp_unschedulable_enabled": true,
			"temp_unschedulable_rules": []any{map[string]any{
				"error_code": http.StatusServiceUnavailable, "keywords": []any{"overloaded"},
				"duration_minutes": 5, "scope": TempUnschedScopeModel,
			}},
			"model_mapping": map[string]any{"alias": "gpt-6.1-sol", "gpt-6.1-sol": "must-not-remap"},
		},
	}
}

func TestOpenAICapacityShedModelRule_StreamReaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	readers := []struct {
		name string
		run  func(*OpenAIGatewayService, *gin.Context, *http.Response, *Account, []byte) error
	}{
		{"stream", func(s *OpenAIGatewayService, c *gin.Context, r *http.Response, a *Account, _ []byte) error {
			_, err := s.handleStreamingResponse(c.Request.Context(), r, c, a, time.Now(), "alias", "gpt-6.1-sol")
			return err
		}},
		{"passthrough", func(s *OpenAIGatewayService, c *gin.Context, r *http.Response, a *Account, _ []byte) error {
			_, err := s.handleStreamingResponsePassthrough(c.Request.Context(), r, c, a, time.Now(), "alias", "gpt-6.1-sol")
			return err
		}},
		{"buffered", func(s *OpenAIGatewayService, c *gin.Context, r *http.Response, a *Account, body []byte) error {
			_, err := s.handleSSEToJSON(r, c, a, body, "alias", "gpt-6.1-sol")
			return err
		}},
		{"buffered_passthrough", func(s *OpenAIGatewayService, c *gin.Context, r *http.Response, a *Account, body []byte) error {
			_, err := s.handlePassthroughSSEToJSON(r, c, a, body, "alias", "gpt-6.1-sol")
			return err
		}},
	}
	for _, accountType := range []string{AccountTypeOAuth, AccountTypeSetupToken, AccountTypeAPIKey} {
		for _, reader := range readers {
			for _, frame := range []struct{ name, payload string }{
				{"error", `{"type":"error","error":{"code":"server_is_overloaded","type":"service_unavailable_error","message":"Our servers are currently overloaded. Please try again later."}}`},
				{"response.failed", `{"type":"response.failed","response":{"error":{"code":"slow_down","message":"overloaded"}}}`},
				{"explicit503", `{"type":"response.failed","response":{"error":{"status_code":503,"message":"overloaded; please retry"}}}`},
				{"short_capacity", `{"type":"error","error":{"code":"slow_down","message":"overloaded"}}`},
			} {
				t.Run(accountType+"/"+reader.name+"/"+frame.name, func(t *testing.T) {
					repo := &modelNotFoundAccountRepoStub{}
					svc := &OpenAIGatewayService{rateLimitService: &RateLimitService{accountRepo: repo}}
					account := openAICapacityShedModelRuleAccount(accountType)
					rec := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(rec)
					c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
					body := "data: " + frame.payload + "\n\n"
					resp := &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/event-stream"}},
						Body: io.NopCloser(strings.NewReader(body))}

					err := reader.run(svc, c, resp, account, []byte(body))

					var failover *UpstreamFailoverError
					require.ErrorAs(t, err, &failover)
					require.Len(t, repo.modelRateLimitCalls, 1)
					require.False(t, failover.RetryableOnSameAccount, "a matching model cooldown must stop same-account replay")
					require.Equal(t, frame.name != "explicit503", failover.RequestScopedTransient, "keep capacity classification distinct from ordinary 503")
					require.Equal(t, http.StatusServiceUnavailable, failover.ClientStatusCode)
					call := repo.modelRateLimitCalls[0]
					require.Equal(t, "gpt-6.1-sol", call.scope, "use the already mapped scheduling model")
					require.WithinDuration(t, time.Now().Add(5*time.Minute), call.resetAt, time.Second)
					var state TempUnschedState
					require.NoError(t, json.Unmarshal([]byte(call.reason), &state))
					require.Equal(t, TempUnschedScopeModel, state.Scope)
					require.Equal(t, http.StatusServiceUnavailable, state.StatusCode)
					require.Zero(t, repo.tempCalls)
					require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
					require.False(t, c.Writer.Written())
					require.Empty(t, rec.Body.String())
				})
			}
		}
	}
}

func TestOpenAICapacityShedModelRule_HTTP(t *testing.T) {
	account := openAICapacityShedModelRuleAccount(AccountTypeOAuth)
	repo := &modelNotFoundAccountRepoStub{}
	svc := &OpenAIGatewayService{rateLimitService: &RateLimitService{accountRepo: repo}}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	body := []byte(`{"error":{"code":"server_is_overloaded","message":"overloaded"}}`)
	err := svc.handleFailoverErrorResponsePassthrough(context.Background(),
		&http.Response{StatusCode: http.StatusServiceUnavailable, Header: http.Header{}}, c, account,
		[]byte(`{"model":"alias"}`), body)
	var failover *UpstreamFailoverError
	require.ErrorAs(t, err, &failover)
	require.False(t, failover.RetryableOnSameAccount)
	require.True(t, failover.RequestScopedTransient)
	require.Len(t, repo.modelRateLimitCalls, 1)
	require.Equal(t, "gpt-6.1-sol", repo.modelRateLimitCalls[0].scope)
	require.Zero(t, repo.tempCalls)
	require.False(t, c.Writer.Written())
}

func TestOpenAICapacityShedModelRuleBoundaries(t *testing.T) {
	for _, scenario := range []string{"unconfigured", "account scope", "unknown model", "excluded code", "other platform", "metadata only", "output only", "write failure", "mapped"} {
		t.Run(scenario, func(t *testing.T) {
			account := openAICapacityShedModelRuleAccount(AccountTypeAPIKey)
			repo := &modelNotFoundAccountRepoStub{}
			svc := &OpenAIGatewayService{rateLimitService: &RateLimitService{accountRepo: repo}}
			model := "gpt-6.1-sol"
			body := []byte(`{"error":{"code":"server_is_overloaded","message":"overloaded"}}`)
			switch scenario {
			case "unconfigured":
				account.Credentials["temp_unschedulable_enabled"] = false
			case "account scope":
				account.Credentials["temp_unschedulable_rules"].([]any)[0].(map[string]any)["scope"] = TempUnschedScopeAccount
			case "unknown model":
				model = ""
			case "excluded code":
				account.Credentials["custom_error_codes_enabled"] = true
				account.Credentials["custom_error_codes"] = []any{float64(410)}
			case "other platform":
				account.Platform = PlatformGrok
			case "metadata only":
				body = []byte(`{"error":{"code":"slow_down","message":"busy"},"metadata":{"note":"overloaded"}}`)
			case "output only":
				body = []byte(`{"response":{"error":{"code":"slow_down","message":"busy"},"output":[{"text":"overloaded"}]}}`)
			case "write failure":
				repo.modelRateLimitErr = errors.New("SQLite busy")
			}
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
			err := svc.newOpenAIStreamFailoverErrorWithModel(c, account, false, "fixture", body, extractOpenAISSEErrorMessage(body), model)
			attempted := scenario == "mapped" || scenario == "write failure"
			if attempted {
				require.Len(t, repo.modelRateLimitCalls, 1)
			} else {
				require.Empty(t, repo.modelRateLimitCalls)
			}
			require.Equal(t, !attempted, err.RetryableOnSameAccount)
			require.Zero(t, repo.tempCalls)
			require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
		})
	}
}

func TestOpenAICapacityModelRuleAfterCommittedOutputAndWS(t *testing.T) {
	for _, mode := range []string{"stream", "passthrough", "ws_error", "ws_terminal"} {
		t.Run(mode, func(t *testing.T) {
			account := openAICapacityShedModelRuleAccount(AccountTypeOAuth)
			repo := &modelNotFoundAccountRepoStub{}
			svc := &OpenAIGatewayService{rateLimitService: &RateLimitService{accountRepo: repo}}
			payload := []byte(`{"type":"response.failed","response":{"error":{"code":"slow_down","message":"overloaded"}}}`)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
			if mode == "ws_error" {
				payload = []byte(`{"type":"error","error":{"code":"server_is_overloaded","message":"overloaded"}}`)
				svc.handleOpenAIWSErrorEventTransientFailure(c.Request.Context(), account, "gpt-6.1-sol", http.Header{}, payload)
			} else if mode == "ws_terminal" {
				svc.handleOpenAIWSTerminalTransientFailure(c.Request.Context(), account, "gpt-6.1-sol", http.Header{}, payload)
			} else {
				body := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"answer\"}\n\ndata: " + string(payload) + "\n\n"
				response := &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: io.NopCloser(strings.NewReader(body))}
				if mode == "stream" {
					_, _ = svc.handleStreamingResponse(c.Request.Context(), response, c, account, time.Now(), "alias", "gpt-6.1-sol")
				} else {
					_, _ = svc.handleStreamingResponsePassthrough(c.Request.Context(), response, c, account, time.Now(), "alias", "gpt-6.1-sol")
				}
				require.True(t, c.Writer.Written())
			}
			require.Len(t, repo.modelRateLimitCalls, 1)
			require.Equal(t, "gpt-6.1-sol", repo.modelRateLimitCalls[0].scope)
			require.Zero(t, repo.tempCalls)
		})
	}
}
