package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOpsWeightedTokenRateHTTPWithSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "ops-rate.db")+"?_pragma=foreign_keys(1)&_time_format=sqlite")
	require.NoError(t, err)
	defer func() { require.NoError(t, db.Close()) }()
	db.SetMaxOpenConns(1)
	require.NoError(t, repository.ApplyMigrations(context.Background(), db))
	_, err = db.Exec(`
INSERT INTO users (id,email,password_hash) VALUES (9001,'rate@example.invalid','fixture');
INSERT INTO groups (id,name,platform) VALUES (9001,'Rate','openai');
INSERT INTO accounts (id,name,platform,type,credentials) VALUES (9001,'Rate','openai','apikey','{}');
INSERT INTO api_keys (id,user_id,key,name) VALUES (9001,9001,'local-rate-fixture','Rate');`)
	require.NoError(t, err)
	now := time.Now().UTC().Add(-time.Minute)
	for _, row := range []struct {
		model    string
		output   int
		duration any
	}{
		{"o3-weighted", 100, 1000}, {"o3-weighted", 100, 10000}, {"o3-weighted", 0, 9000}, {"o3-weighted", 10, 0}, {"o3-weighted", 7, nil}, {"o3-empty", 0, 1000},
	} {
		_, err = db.Exec(`INSERT INTO usage_logs(user_id,api_key_id,account_id,group_id,model,output_tokens,duration_ms,created_at) VALUES(9001,9001,9001,9001,$1,$2,$3,$4)`, row.model, row.output, row.duration, now)
		require.NoError(t, err)
	}
	svc := service.NewOpsService(repository.NewOpsRepository(db), nil, &config.Config{Ops: config.OpsConfig{Enabled: true}}, nil, nil, nil, nil, nil, nil, nil, nil)
	h := NewOpsHandler(svc)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/token-stats", h.GetDashboardOpenAITokenStats)
	server := httptest.NewServer(router)
	defer server.Close()
	for _, query := range []string{"time_range=1h&platform=openai&group_id=9001&top_n=10", "time_range=1h&platform=openai&group_id=9001&page=1&page_size=20"} {
		resp, err := http.Get(server.URL + "/token-stats?" + query)
		require.NoError(t, err)
		var envelope struct {
			Data service.OpsOpenAITokenStatsResponse `json:"data"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&envelope))
		require.NoError(t, resp.Body.Close())
		require.Equal(t, 200, resp.StatusCode)
		require.Len(t, envelope.Data.Items, 2)
		weighted := envelope.Data.Items[0]
		require.Equal(t, "o3-weighted", weighted.Model)
		require.InDelta(t, 18.18, *weighted.AvgTokensPerSec, 0.01)
		require.EqualValues(t, 5, weighted.RequestCount)
		require.EqualValues(t, 217, weighted.TotalOutputTokens)
		require.Nil(t, envelope.Data.Items[1].AvgTokensPerSec)
	}
}
