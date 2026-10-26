//go:build unit

package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type failingAccountScopeRepo struct{ tempUnschedScopeRepoStub }

func (r *failingAccountScopeRepo) SetTempUnschedulable(context.Context, int64, time.Time, string) error {
	r.tempCalls++
	return errors.New("SQLite write failure")
}

func TestTempUnschedScope_FirstMatchWriteFailureDoesNotFallThrough(t *testing.T) {
	repo := &failingAccountScopeRepo{}
	account := newTempUnschedScopeAccount(TempUnschedScopeAccount, nil)
	rules := account.Credentials["temp_unschedulable_rules"].([]any)
	second := map[string]any{"error_code": 503, "keywords": []any{"unavailable"}, "duration_minutes": 20, "scope": "model"}
	account.Credentials["temp_unschedulable_rules"] = append(rules, second)
	svc := &RateLimitService{accountRepo: repo}
	require.False(t, svc.HandleTempUnschedulable(context.Background(), account, 503, tempUnschedScopeBody, "gpt-5.4"))
	require.Equal(t, 1, repo.tempCalls)
	require.Empty(t, repo.modelRateLimitCalls)
}

func TestTempUnschedScope_ExplicitAccountUsesRuntimeBlock(t *testing.T) {
	repo := &tempUnschedScopeRepoStub{}
	gateway := &OpenAIGatewayService{rateLimitService: &RateLimitService{accountRepo: repo, cfg: &config.Config{}}}
	account := newTempUnschedScopeAccount(TempUnschedScopeAccount, nil)
	require.True(t, gateway.handleOpenAIAccountUpstreamError(context.Background(), account, 503, http.Header{}, tempUnschedScopeBody, "gpt-5.4"))
	require.Equal(t, 1, repo.tempCalls)
	require.Empty(t, repo.modelRateLimitCalls)
	require.True(t, gateway.isOpenAIAccountRuntimeBlocked(account))
}

func TestTempUnschedScope_SetupToken401KeepsCredentialHandling(t *testing.T) {
	account := newTempUnschedScope401Account(PlatformOpenAI, TempUnschedScopeModel)
	account.Type = AccountTypeSetupToken
	require.False(t, tempUnschedModelRuleMatches(account, 401, tempUnschedScope401Body, "gpt-5.4"))
	repo := &tempUnschedScopeRepoStub{}
	gateway := &OpenAIGatewayService{rateLimitService: &RateLimitService{accountRepo: repo, cfg: &config.Config{}}}
	require.True(t, gateway.handleOpenAIAccountUpstreamError(context.Background(), account, 401, http.Header{}, tempUnschedScope401Body, "gpt-5.4"))
	require.Empty(t, repo.modelRateLimitCalls)
	require.Equal(t, 1, repo.errorCalls)
	require.True(t, gateway.isOpenAIAccountRuntimeBlocked(account))
}

func TestTempUnschedScope_CredentialMarkersAcrossExistingPlatforms(t *testing.T) {
	for _, platform := range []string{PlatformOpenAI, PlatformAnthropic, PlatformGrok, PlatformGemini} {
		for _, marker := range []string{"token_revoked", "token_invalidated", "invalid_api_key", "unauthorized", "authentication_error", "detail Unauthorized", "model permission"} {
			t.Run(platform+"/"+marker, func(t *testing.T) {
				account := newTempUnschedScope401Account(platform, TempUnschedScopeModel)
				repo := &tempUnschedScopeRepoStub{}
				svc := &RateLimitService{accountRepo: repo, cfg: &config.Config{}}
				body := []byte(`{"error":{"code":"` + marker + `","message":"unauthorized"}}`)
				if marker == "detail Unauthorized" {
					body = []byte(`{"detail":"Unauthorized"}`)
				} else if marker == "authentication_error" {
					body = []byte(`{"error":{"type":"authentication_error","message":"unauthorized"}}`)
				}
				require.True(t, svc.HandleUpstreamError(context.Background(), account, http.StatusUnauthorized, http.Header{}, body, "model-a"))
				if marker == "model permission" {
					require.Zero(t, repo.errorCalls)
					require.Len(t, repo.modelRateLimitCalls, 1, "unmarked per-model API key401 stays model-scoped")
				} else {
					require.Equal(t, 1, repo.errorCalls, "explicit credential failure must retain the platform's account authentication handling")
					require.Empty(t, repo.modelRateLimitCalls)
				}
				require.Zero(t, repo.tempCalls)
			})
		}
	}
}
