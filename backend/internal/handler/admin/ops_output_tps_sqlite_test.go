package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestOpsOutputTPSHTTPWithSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "ops.db")+"?_pragma=foreign_keys(1)&_time_format=sqlite")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)
	ctx := context.Background()
	require.NoError(t, repository.ApplyMigrations(ctx, db))
	_, err = db.Exec(`
INSERT INTO users (id,email,password_hash) VALUES (9001,'tps@example.invalid','hash');
INSERT INTO groups (id,name,platform) VALUES (9001,'Output TPS','grok');
INSERT INTO accounts (id,name,platform,type,credentials) VALUES (9001,'TPS','anthropic','apikey','{}');
INSERT INTO api_keys (id,user_id,key,name) VALUES (9001,9001,'sk-output-tps-fixture','TPS');`)
	require.NoError(t, err)
	start := time.Date(2026, 1, 17, 0, 30, 0, 0, time.UTC)
	end := start.Add(6 * time.Hour)
	for i, output := range []int{1, 10, 100, 1000} {
		_, err = db.Exec(`INSERT INTO usage_logs
(user_id,api_key_id,account_id,group_id,model,output_tokens,duration_ms,request_type,created_at)
VALUES (9001,9001,9001,9001,'grok-test',$1,1000,$2,$3)`, output, service.RequestTypeSync, start.Add(time.Duration(i)*time.Hour))
		require.NoError(t, err)
	}
	repo := repository.NewOpsRepository(db)
	require.NoError(t, repo.UpsertHourlyMetrics(ctx, start.Truncate(time.Hour), end))
	svc := service.NewOpsService(repo, nil, &config.Config{Ops: config.OpsConfig{Enabled: true, UsePreaggregatedTables: true}}, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewOpsHandler(svc)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/overview", handler.GetDashboardOverview)
	router.GET("/snapshot", handler.GetDashboardSnapshotV2)
	for _, endpoint := range []string{"/overview", "/snapshot"} {
		for _, mode := range []string{"raw", "auto", "preagg"} {
			t.Run(endpoint+"/"+mode, func(t *testing.T) {
				params := url.Values{"start_time": {start.Format(time.RFC3339)}, "end_time": {end.Format(time.RFC3339)},
					"platform": {"grok"}, "group_id": {"9001"}, "mode": {mode}}
				w := httptest.NewRecorder()
				router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, endpoint+"?"+params.Encode(), nil))
				require.Equal(t, http.StatusOK, w.Code, w.Body.String())
				var envelope struct {
					Data json.RawMessage `json:"data"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
				if endpoint == "/snapshot" {
					var snapshot struct {
						Overview json.RawMessage `json:"overview"`
					}
					require.NoError(t, json.Unmarshal(envelope.Data, &snapshot))
					envelope.Data = snapshot.Overview
				}
				var result service.OpsDashboardOverview
				require.NoError(t, json.Unmarshal(envelope.Data, &result))
				require.NotNil(t, result.OutputTPS)
				require.EqualValues(t, 4, result.OutputTPS.SampleCount)
				require.InDelta(t, 2.35, *result.OutputTPS.P5, 1e-9)
				require.InDelta(t, 3.7, *result.OutputTPS.P10, 1e-9)
				require.Equal(t, 55.0, *result.OutputTPS.P50)
				require.Equal(t, 277.75, *result.OutputTPS.Avg)
			})
		}
	}
}
