//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenAIRotationMergesConcurrentNonTokenEdits(t *testing.T) {
	for _, field := range []string{"proxy", "model mapping"} {
		t.Run(field, func(t *testing.T) {
			a := openAIRefreshTestAccount()
			a.Credentials["model_mapping"] = map[string]any{"old-alias": "gpt-6.2-sol"}
			repo := &refreshAPIAccountRepo{account: a}
			proxyID := int64(20)
			repo.beforeSuccessCAS = func(r *refreshAPIAccountRepo) {
				if r.openAISuccessCASCalls != 1 {
					return
				}
				if field == "proxy" {
					r.account.ProxyID = &proxyID
				} else {
					r.account.Credentials = shallowCopyMap(r.account.Credentials)
					r.account.Credentials["model_mapping"] = map[string]any{"new-alias": "gpt-6.2-astra"}
				}
			}
			executor := &refreshAPIExecutorStub{needsRefresh: true, credentials: MergeCredentials(a.Credentials, map[string]any{"access_token": "fixture-rotated", "refresh_token": "fixture-next"})}
			result, err := NewOAuthRefreshAPI(repo, nil).RefreshIfNeeded(context.Background(), a, executor, time.Hour)
			require.NoError(t, err)
			require.True(t, result.Refreshed)
			require.Equal(t, 2, repo.openAISuccessCASCalls)
			require.Equal(t, 1, executor.refreshCalls)
			require.Equal(t, "fixture-next", result.Account.GetOpenAIRefreshToken())
			if field == "proxy" {
				require.Equal(t, &proxyID, result.Account.ProxyID)
			} else {
				require.Equal(t, map[string]any{"new-alias": "gpt-6.2-astra"}, result.Account.Credentials["model_mapping"])
			}
		})
	}
}

func TestOpenAIRotationRepeatedCASConflictIsBounded(t *testing.T) {
	a := openAIRefreshTestAccount()
	repo := &refreshAPIAccountRepo{account: a}
	repo.beforeSuccessCAS = func(r *refreshAPIAccountRepo) {
		r.account.Credentials = shallowCopyMap(r.account.Credentials)
		r.account.Credentials["fixture_edit_version"] = r.openAISuccessCASCalls
	}
	executor := &refreshAPIExecutorStub{needsRefresh: true, credentials: MergeCredentials(a.Credentials, map[string]any{"access_token": "fixture-rotated", "refresh_token": "fixture-next"})}
	_, err := NewOAuthRefreshAPI(repo, nil).RefreshIfNeeded(context.Background(), a, executor, time.Hour)
	var containment *providerCycleContainmentRefreshError
	require.ErrorAs(t, err, &containment)
	require.Equal(t, 3, repo.openAISuccessCASCalls)
	require.Equal(t, 1, executor.refreshCalls)
	require.Zero(t, repo.updateCalls)
}
