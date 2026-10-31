package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type grokDiagnosticsAdmin struct {
	*stubAdminService
	repo service.AccountRepository
}

func (s *grokDiagnosticsAdmin) GetAccount(ctx context.Context, id int64) (*service.Account, error) {
	return s.repo.GetByID(ctx, id)
}

type grokDiagnosticsHTTP struct{ calls int }

func (s *grokDiagnosticsHTTP) Do(*http.Request, string, int64, int) (*http.Response, error) {
	s.calls++
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"data":[{"id":"grok-4.6","reasoning":false,"context_window":32768,"input_modalities":["text"]},{"id":"grok-4.5","reasoning":false,"context_window":32768,"input_modalities":["text"]}]}`))}, nil
}

func (s *grokDiagnosticsHTTP) DoWithTLS(r *http.Request, proxy string, id int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return s.Do(r, proxy, id, concurrency)
}

func TestGrokModelDiagnosticsAdminSQLiteReadAndSync(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "diagnostics.db")+"?_time_format=sqlite&_pragma=foreign_keys(1)")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	require.NoError(t, repository.ApplyMigrations(ctx, db))
	require.NoError(t, repository.EnsureSQLiteAuxTables(ctx, db))
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	t.Cleanup(func() { _ = client.Close() })
	scope, err := service.NewGrokCatalogScope(101, service.AccountTypeAPIKey, "https://api.x.ai/v1/models")
	require.NoError(t, err)
	fetched := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339Nano)
	extra, err := json.Marshal(map[string]any{"grok_observed_models": service.GrokObservedModelsSnapshot{Models: []string{"grok-4.6"}, FetchedAt: fetched, Scope: scope}})
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO accounts(id,name,platform,type,credentials,extra) VALUES(101,'Grok','grok','apikey',?,?)`, `{"api_key":"fixture-secret","base_url":"https://api.x.ai/v1","model_mapping":{"friendly":"grok-4.6","missing":"grok-4.5"}}`, string(extra))
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO accounts(id,name,platform,type,credentials,extra) VALUES(102,'OpenAI','openai','apikey','{}','{}')`)
	require.NoError(t, err)
	repo := repository.NewAccountRepository(client, db, nil)
	adminSvc := &grokDiagnosticsAdmin{stubAdminService: newStubAdminService(), repo: repo}
	upstream := &grokDiagnosticsHTTP{}
	cfg := &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}}
	testSvc := service.NewAccountTestService(repo, nil, nil, nil, nil, upstream, cfg, nil)
	h := NewAccountHandler(adminSvc, nil, nil, nil, nil, nil, nil, nil, testSvc, nil, nil, nil, nil, nil)
	router := gin.New()
	router.GET("/api/v1/admin/accounts/:id", h.GetByID)
	router.POST("/api/v1/admin/accounts/:id/models/sync-upstream", h.SyncUpstreamModels)
	_, err = db.Exec(`CREATE TRIGGER diagnostics_no_writes BEFORE UPDATE ON accounts BEGIN SELECT RAISE(ABORT,'diagnostic read attempted write'); END`)
	require.NoError(t, err)
	var before string
	require.NoError(t, db.QueryRow(`SELECT extra FROM accounts WHERE id=101`).Scan(&before))
	detail := doJSON(t, router, http.MethodGet, "/api/v1/admin/accounts/101", nil)
	require.Equal(t, http.StatusOK, detail.Code)
	require.Equal(t, "fresh", gjson.GetBytes(detail.Body.Bytes(), "data.grok_model_diagnostics.state").String())
	require.Equal(t, "not_listed", gjson.GetBytes(detail.Body.Bytes(), `data.grok_model_diagnostics.mappings.#(requested=="missing").status`).String())
	var after string
	require.NoError(t, db.QueryRow(`SELECT extra FROM accounts WHERE id=101`).Scan(&after))
	require.Equal(t, before, after)
	require.Zero(t, upstream.calls, "reading diagnostics must not call upstream")
	other := doJSON(t, router, http.MethodGet, "/api/v1/admin/accounts/102", nil)
	require.Equal(t, http.StatusOK, other.Code)
	require.False(t, gjson.GetBytes(other.Body.Bytes(), "data.grok_model_diagnostics").Exists())
	account, err := repo.GetByID(ctx, 101)
	require.NoError(t, err)
	mapped, err := json.Marshal(dto.AccountFromServiceShallow(account))
	require.NoError(t, err)
	require.False(t, gjson.GetBytes(mapped, "grok_model_diagnostics").Exists(), "public/list mapper must omit diagnostics")
	_, err = db.Exec(`DROP TRIGGER diagnostics_no_writes`)
	require.NoError(t, err)
	synced := doJSON(t, router, http.MethodPost, "/api/v1/admin/accounts/101/models/sync-upstream", nil)
	require.Equal(t, http.StatusOK, synced.Code, "%s", synced.Body.String())
	require.Equal(t, "upstream", gjson.GetBytes(synced.Body.Bytes(), "data.discovery_source").String())
	require.Equal(t, "listed", gjson.GetBytes(synced.Body.Bytes(), `data.grok_model_diagnostics.mappings.#(requested=="missing").status`).String())
	require.NotEqual(t, fetched, gjson.GetBytes(synced.Body.Bytes(), "data.grok_model_diagnostics.fetched_at").String())
	require.Equal(t, 1, upstream.calls, fmt.Sprintf("unexpected enrichment or diagnostic HTTP calls: %d", upstream.calls))
}
