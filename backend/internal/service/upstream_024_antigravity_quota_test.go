//go:build unit

package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/antigravity"
	"github.com/stretchr/testify/require"
)

func TestUpstream024AntigravityQuotaForwardEndpoint(t *testing.T) {
	for _, tc := range []struct {
		name, plan, override string
		daily                bool
	}{
		{"paid", "Ultra", "", true}, {"free", "Free", "", false}, {"override", "Free", "daily", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(antigravityForwardBaseURLEnv, tc.override)
			var prodHits, dailyHits atomic.Int32
			newServer := func(counter *atomic.Int32, remaining string) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					if r.URL.Path == "/v1internal:fetchAvailableModels" {
						counter.Add(1)
						_, _ = w.Write([]byte(`{"models":{"gemini-3-flash":{"quotaInfo":{"remainingFraction":` + remaining + `}}}}`))
						return
					}
					_, _ = w.Write([]byte(`{"paidTier":{"id":"g1-ultra-tier"}}`))
				}))
			}
			prod := newServer(&prodHits, "0.75")
			defer prod.Close()
			daily := newServer(&dailyHits, "0.25")
			defer daily.Close()
			previous := antigravity.BaseURLs
			availability := antigravity.DefaultURLAvailability
			t.Cleanup(func() { antigravity.BaseURLs = previous; antigravity.DefaultURLAvailability = availability })
			antigravity.BaseURLs = []string{prod.URL, daily.URL}
			antigravity.DefaultURLAvailability = antigravity.NewURLAvailability(time.Minute)
			fetcher := NewAntigravityQuotaFetcher(nil)
			result, err := fetcher.FetchQuota(context.Background(), &Account{ID: 80, Platform: PlatformAntigravity, Type: AccountTypeOAuth, Credentials: map[string]any{"access_token": "synthetic-token", "project_id": "project", "plan_type": tc.plan}}, "")
			require.NoError(t, err)
			if tc.daily {
				require.Equal(t, int32(1), dailyHits.Load())
				require.Zero(t, prodHits.Load())
				require.Equal(t, 75, result.UsageInfo.AntigravityQuota["gemini-3-flash"].Utilization)
			} else {
				require.Equal(t, int32(1), prodHits.Load())
				require.Zero(t, dailyHits.Load())
				require.Equal(t, 25, result.UsageInfo.AntigravityQuota["gemini-3-flash"].Utilization)
			}
		})
	}
}

func TestUpstream024AntigravityQuotaDoesNotFallbackAcrossEndpoints(t *testing.T) {
	t.Setenv(antigravityForwardBaseURLEnv, "")
	var prodHits atomic.Int32
	prod := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prodHits.Add(1)
		_, _ = w.Write([]byte(`{"models":{}}`))
	}))
	defer prod.Close()
	daily := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "synthetic endpoint unavailable", http.StatusServiceUnavailable)
	}))
	defer daily.Close()
	previous := antigravity.BaseURLs
	t.Cleanup(func() { antigravity.BaseURLs = previous })
	antigravity.BaseURLs = []string{prod.URL, daily.URL}
	result, err := NewAntigravityQuotaFetcher(nil).FetchQuota(context.Background(), &Account{
		ID: 80, Platform: PlatformAntigravity, Type: AccountTypeOAuth,
		Credentials: map[string]any{"access_token": "synthetic-token", "plan_type": "Pro"},
	}, "")
	require.Error(t, err)
	require.Nil(t, result)
	require.Zero(t, prodHits.Load())
}
