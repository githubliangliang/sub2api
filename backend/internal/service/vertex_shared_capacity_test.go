//go:build unit

package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestVertexSharedCapacityKeepsAccountSchedulable(t *testing.T) {
	body := []byte(`{"error":{"code":429,"status":"RESOURCE_EXHAUSTED","message":"Resource exhausted. Please try again later."}}`)
	repo := &rateLimit429AccountRepoStub{}
	svc := &GeminiMessagesCompatService{accountRepo: repo}
	account := &Account{ID: 604, Platform: PlatformGemini, Type: AccountTypeServiceAccount, Credentials: map[string]any{"project_id": "p"}}
	svc.handleGeminiUpstreamError(context.Background(), account, http.StatusTooManyRequests, http.Header{}, body)
	require.Zero(t, repo.rateLimitCalls)
}

func TestVertexAccountQuotaKeepsShortCooldown(t *testing.T) {
	for _, body := range []string{
		`{"error":{"code":429,"status":"RESOURCE_EXHAUSTED","message":"Quota exceeded. Please try again later."}}`,
		`{"error":{"code":429,"status":"RESOURCE_EXHAUSTED","message":"Resource exhausted. Please try again later.","details":[{"@type":"type.googleapis.com/google.rpc.QuotaFailure"}]}}`,
		`{"error":{"code":429,"status":"RESOURCE_EXHAUSTED","message":"Resource exhausted. Please try again later.","details":[{"@type":"type.googleapis.com/google.rpc.ErrorInfo"}]}}`,
	} {
		repo := &rateLimit429AccountRepoStub{}
		svc := &GeminiMessagesCompatService{accountRepo: repo}
		before := time.Now()
		svc.handleGeminiUpstreamError(context.Background(), &Account{ID: 605, Platform: PlatformGemini, Type: AccountTypeServiceAccount}, http.StatusTooManyRequests, http.Header{}, []byte(body))
		require.Equal(t, 1, repo.rateLimitCalls)
		require.WithinDuration(t, before.Add(time.Minute), repo.lastRateLimitReset, 2*time.Second)
	}
}
