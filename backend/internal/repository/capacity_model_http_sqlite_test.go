//go:build unit

package repository

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCapacityModelCooldownHTTPPersistsSQLiteAndSchedulerCache(t *testing.T) {
	ctx := context.Background()
	repo, client := newOllamaCloudUsageSQLiteRepository(t)
	_, err := repo.sql.ExecContext(ctx, `CREATE TABLE scheduler_outbox(id INTEGER PRIMARY KEY AUTOINCREMENT,event_type TEXT NOT NULL,account_id INTEGER,group_id INTEGER,payload TEXT,created_at DATETIME DEFAULT CURRENT_TIMESTAMP,dedup_key TEXT UNIQUE)`)
	require.NoError(t, err)
	disabled := false
	cfg := &config.Config{Redis: config.RedisConfig{Enabled: &disabled}, Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false, AllowInsecureHTTP: true}}}
	rdb := InitRedis(cfg)
	defer func() { _ = rdb.Close() }()
	scheduler := NewSchedulerCache(rdb)
	repo.schedulerCache = scheduler
	limits := service.NewRateLimitService(repo, nil, cfg, nil, NewTempUnschedCache(rdb))
	for i, transport := range []string{"http", "sse", "buffered"} {
		t.Run(transport, func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if transport == "http" {
					w.WriteHeader(503)
					_, _ = io.WriteString(w, `{"error":{"code":"server_is_overloaded","message":"capacity fixture"}}`)
					return
				}
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = io.WriteString(w, "data: {\"type\":\"response.failed\",\"response\":{\"error\":{\"code\":\"slow_down\",\"message\":\"capacity fixture\"}}}\n\n")
			}))
			defer upstream.Close()
			credentials := map[string]any{"api_key": "fixture", "base_url": upstream.URL, "model_mapping": map[string]any{"alias": "gpt-6.1-sol"}, "temp_unschedulable_enabled": true, "temp_unschedulable_rules": []any{map[string]any{"error_code": 503, "keywords": []any{"capacity fixture"}, "duration_minutes": 1, "scope": "model"}}}
			created, err := client.Account.Create().SetName(fmt.Sprintf("capacity-%d", i)).SetPlatform(service.PlatformOpenAI).SetType(service.AccountTypeAPIKey).SetCredentials(credentials).SetExtra(map[string]any{}).Save(ctx)
			require.NoError(t, err)
			account, err := repo.GetByID(ctx, created.ID)
			require.NoError(t, err)
			svc := service.NewOpenAIGatewayService(repo, nil, nil, nil, nil, nil, nil, cfg, nil, nil, nil, limits, nil, cacheHitHTTPUpstream{upstream.Client()}, service.NewDeferredService(nil, nil, time.Hour), nil, nil, nil, nil, nil, nil, nil)
			body := fmt.Sprintf(`{"model":"alias","input":"hello","stream":%t}`, transport != "buffered")
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("POST", "/v1/responses", strings.NewReader(body))
			_, err = svc.Forward(ctx, c, account, []byte(body))
			var failover *service.UpstreamFailoverError
			require.ErrorAs(t, err, &failover)
			require.False(t, failover.RetryableOnSameAccount)
			persisted, err := repo.GetByID(ctx, created.ID)
			require.NoError(t, err)
			cached, err := scheduler.GetAccount(ctx, created.ID)
			require.NoError(t, err)
			require.NotNil(t, cached)
			for _, state := range []*service.Account{persisted, cached} {
				require.False(t, state.IsSchedulableForModelWithContext(ctx, "alias"))
				require.True(t, state.IsSchedulableForModelWithContext(ctx, "unrelated"))
				require.Nil(t, state.TempUnschedulableUntil)
			}
		})
	}
}
