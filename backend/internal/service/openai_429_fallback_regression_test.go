//go:build unit

package service

import (
	"context"
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestOpenAI429FastPath_DisabledFallbackHonorsOnlyExhaustion(t *testing.T) {
	for _, tc := range []struct {
		name        string
		used        string
		body        string
		wantBlocked bool
	}{
		{"ordinary", "", `{"detail":"Rate limit exceeded"}`, false},
		{"unexhausted_windows", "37", `{"detail":"Rate limit exceeded"}`, false},
		{"exhausted_window", "100", `{"detail":"Rate limit exceeded"}`, true},
		{"explicit_body_reset", "37", `{"error":{"type":"usage_limit_reached","resets_at":4102444800}}`, true},
		{"expired_body_reset", "37", `{"error":{"type":"usage_limit_reached","resets_at":1}}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &geminiErrorPolicyRepo{}
			settingRepo := newMockSettingRepo()
			settingRepo.data[SettingKeyRateLimit429CooldownSettings] = `{"enabled":false,"cooldown_seconds":12}`
			rateLimits := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
			rateLimits.SetSettingService(NewSettingService(settingRepo, &config.Config{}))
			svc := &OpenAIGatewayService{rateLimitService: rateLimits}
			rateLimits.SetAccountRuntimeBlocker(svc)
			account := &Account{ID: 426, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
			headers := http.Header{}
			if tc.used != "" {
				headers.Set("x-codex-primary-used-percent", tc.used)
				headers.Set("x-codex-primary-reset-after-seconds", "604800")
				headers.Set("x-codex-primary-window-minutes", "10080")
				headers.Set("x-codex-secondary-used-percent", "20")
				headers.Set("x-codex-secondary-reset-after-seconds", "3600")
				headers.Set("x-codex-secondary-window-minutes", "300")
			}
			svc.markOpenAIOAuth429RateLimited(context.Background(), account, headers, []byte(tc.body))
			require.Equal(t, tc.wantBlocked, svc.isOpenAIAccountRuntimeBlocked(account))
			if !tc.wantBlocked {
				require.Zero(t, repo.setRateLimitedCalls)
			}
		})
	}
}
