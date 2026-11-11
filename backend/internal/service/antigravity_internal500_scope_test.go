//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// --- mock: records model-level rate limits as well as account-level penalties ---

type internal500ScopeRepoStub struct {
	internal500AccountRepoStub
	modelLimitCalls []modelLimitCall
}

type modelLimitCall struct {
	accountID int64
	scope     string
	resetAt   time.Time
}

func (r *internal500ScopeRepoStub) SetModelRateLimit(_ context.Context, id int64, scope string, resetAt time.Time, _ ...string) error {
	r.modelLimitCalls = append(r.modelLimitCalls, modelLimitCall{accountID: id, scope: scope, resetAt: resetAt})
	return nil
}

// An INTERNAL 500 on an image model is the image endpoint failing, not the account: it must not
// take the account's text models (Argon, Flash) down with it.
func TestInternal500OnImageModelCoolsOnlyThatModel(t *testing.T) {
	repo := &internal500ScopeRepoStub{}
	cache := &mockInternal500Cache{incrementCount: 1}
	svc := &AntigravityGatewayService{accountRepo: repo, internal500Cache: cache}
	account := &Account{ID: 2, Name: "ag", Platform: PlatformAntigravity}

	before := time.Now()
	svc.handleInternal500ForModel(context.Background(), "[test]", account, "gemini-3.1-flash-image")

	require.Empty(t, cache.incrementCalls, "an image 500 must not count toward the account penalty")
	require.Empty(t, repo.tempUnschedCalls, "an image 500 must not make the account unschedulable")
	require.Empty(t, repo.setErrorCalls)
	require.Len(t, repo.modelLimitCalls, 1)
	call := repo.modelLimitCalls[0]
	require.Equal(t, "gemini-3.1-flash-image", call.scope, "only the image model is cooled, not antigravity:gemini")
	require.WithinDuration(t, before.Add((2 * time.Minute)), call.resetAt, 2*time.Second)
}

func TestInternal500OnTextModelKeepsTheAccountPenalty(t *testing.T) {
	repo := &internal500ScopeRepoStub{}
	cache := &mockInternal500Cache{incrementCount: 1}
	svc := &AntigravityGatewayService{accountRepo: repo, internal500Cache: cache}
	account := &Account{ID: 2, Name: "ag", Platform: PlatformAntigravity}

	svc.handleInternal500ForModel(context.Background(), "[test]", account, "gemini-4-argon-high")

	require.Len(t, cache.incrementCalls, 1)
	require.Len(t, repo.tempUnschedCalls, 1)
	require.Empty(t, repo.modelLimitCalls)
}
