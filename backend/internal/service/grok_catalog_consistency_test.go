//go:build unit

package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/stretchr/testify/require"
)

func TestGrokCatalogRejectsMissingProtocolID(t *testing.T) {
	for _, body := range []string{`{"error":{"message":"upstream failed"}}`, `{"data":[{"name":"Display Label"}]}`, `{"data":[{"id":"valid"},{}]}`, `{"data":[{"meta":{"name":"Display Label"}}]}`} {
		t.Run(body, func(t *testing.T) {
			ids, err := extractGrokUpstreamModelIDs([]byte(body))
			require.Error(t, err)
			require.Empty(t, ids)
		})
	}
}

func TestGrokCatalogConsumersUseSameIDs(t *testing.T) {
	body := []byte(`{"data":[{"model":"Custom-ID"},{"id":"grok-4.7"},{"id":"Custom-ID"},{"id":"custom-id"}]}`)
	manual, err := extractGrokUpstreamModelIDs(body)
	require.NoError(t, err)
	observed, err := extractGrokModelIDsFromModelsBody(body)
	require.NoError(t, err)
	require.Equal(t, []string{"Custom-ID", "custom-id", "grok-4.7"}, manual)
	require.Equal(t, manual, observed)
}

func TestGrokCatalogDistinguishesEmptyAndInvalid(t *testing.T) {
	for _, body := range []string{`[]`, `{"data":[]}`, `{"models":[]}`} {
		_, err := parseGrokCatalogModelIDs([]byte(body))
		require.ErrorIs(t, err, errGrokCatalogEmpty)
	}
	for _, body := range []string{`null`, `{"data":null}`, `{"models":{}}`, `{"data":[null]}`, `{`, `{"data":[123]}`} {
		_, err := parseGrokCatalogModelIDs([]byte(body))
		require.ErrorIs(t, err, errGrokCatalogInvalidResponse)
	}
	ids, err := parseGrokCatalogModelIDs([]byte(`["Custom-ID", {"model_id":"models/grok-4.7"}, "Custom-ID", "custom-id"]`))
	require.NoError(t, err)
	require.Equal(t, []string{"Custom-ID", "custom-id", "grok-4.7"}, ids)
}

type grokCatalogHTTPFixture struct {
	httpUpstreamRecorder
	body   string
	status int
	err    error
	calls  int
}

func (f *grokCatalogHTTPFixture) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	f.calls++
	if err := req.Context().Err(); err != nil {
		return nil, err
	}
	if f.err != nil {
		return nil, f.err
	}
	return &http.Response{StatusCode: f.status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(f.body))}, nil
}

func (f *grokCatalogHTTPFixture) DoWithTLS(req *http.Request, proxy string, id int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return f.Do(req, proxy, id, concurrency)
}

func TestGrokCatalogBothSyncConsumersPreserveSuccessOnFailures(t *testing.T) {
	for _, background := range []bool{false, true} {
		for _, tc := range []struct {
			name, body, code string
			status           int
			err              error
		}{
			{"success", `{"data":[{"model":"Custom-ID"},{"id":"custom-id"}]}`, "", 200, nil},
			{"invalid", `{"error":{"message":"synthetic-secret"}}`, "invalid_response", 200, nil},
			{"missing-id", `{"data":[{}]}`, "invalid_response", 200, nil},
			{"empty", `{"data":[]}`, "empty_catalog", 200, nil},
			{"http", `{}`, "http_error", 503, nil},
			{"transport", "", "transport", 0, errors.New("synthetic-secret")},
			{"too-large", strings.Repeat(" ", int(upstreamModelsBodyLimit)+1), "too_large", 200, nil},
		} {
			t.Run(tc.name+map[bool]string{true: "/background", false: "/manual"}[background], func(t *testing.T) {
				account := grokCatalogPersistAccount()
				account.Type = AccountTypeOAuth
				account.Status = StatusActive
				account.Credentials["refresh_token"] = "synthetic-refresh"
				account.Credentials["expires_at"] = time.Now().Add(time.Hour).Format(time.RFC3339)
				repo := newGrokCatalogMemoryRepo(account)
				upstream := &grokCatalogHTTPFixture{body: tc.body, status: tc.status, err: tc.err}
				var err error
				if background {
					err = (&GrokQuotaService{accountRepo: repo, httpUpstream: upstream, cfg: &config.Config{}}).syncGrokObservedModels(context.Background(), account)
				} else {
					_, err = (&AccountTestService{accountRepo: repo, httpUpstream: upstream, cfg: &config.Config{}, grokTokenProvider: NewGrokTokenProvider(nil, nil)}).SyncUpstreamModelCatalog(context.Background(), account)
				}
				stored, getErr := repo.GetByID(context.Background(), account.ID)
				require.NoError(t, getErr)
				snap := parseGrokObservedModels(stored.Extra)
				require.NotNil(t, snap)
				if tc.code == "" {
					require.NoError(t, err)
					require.Equal(t, []string{"Custom-ID", "custom-id"}, snap.Models)
					require.NotNil(t, snap.Scope)
					require.Equal(t, "success", snap.LastAttempt.Status)
				} else {
					require.Error(t, err)
					require.Equal(t, []string{"kept"}, snap.Models)
					require.Equal(t, "2026-10-25T01:00:01Z", snap.FetchedAt)
					require.Equal(t, tc.code, snap.LastAttempt.ErrorCode)
					require.Equal(t, "failed", snap.LastAttempt.Status)
				}
				require.Equal(t, "keep", stored.Extra["unrelated"])
				require.NotContains(t, snap.LastAttempt.ErrorCode, "synthetic-secret")
			})
		}
	}
}

func TestGrokCatalogFreshnessRequiresValidTimeAndCurrentScope(t *testing.T) {
	now := time.Now().UTC()
	scope := grokCatalogPersistScope(t, 25)
	snap := &GrokObservedModelsSnapshot{Models: []string{"kept"}, FetchedAt: now.Add(-time.Minute).Format(time.RFC3339Nano), Scope: scope}
	require.True(t, grokObservedSnapshotFresh(snap, scope, now))
	for _, fetched := range []string{"", "tomorrow", now.Add(time.Second).Format(time.RFC3339Nano), now.Add(-grokObservedModelsTTL).Format(time.RFC3339Nano)} {
		invalid := *snap
		invalid.FetchedAt = fetched
		require.False(t, grokObservedSnapshotFresh(&invalid, scope, now))
	}
	legacy := *snap
	legacy.Scope = nil
	require.False(t, grokObservedSnapshotFresh(&legacy, scope, now))
	other, err := NewGrokCatalogScope(25, AccountTypeOAuth, "https://api.x.ai/v1/models?scope=other")
	require.NoError(t, err)
	require.False(t, grokObservedSnapshotFresh(snap, other, now))
	other = cloneGrokCatalogScope(scope)
	other.AccountID++
	require.False(t, grokObservedSnapshotFresh(snap, other, now))
}
