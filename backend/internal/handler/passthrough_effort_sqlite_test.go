//go:build unit

package handler

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestPassthroughFilterAndRequestedEffortOverHTTPWithSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "requests.db")+"?_time_format=sqlite")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, repository.ApplyMigrations(context.Background(), db))
	observed := make(chan []byte, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			t.Error(readErr)
			return
		}
		observed <- body
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", "fixture-response")
		_, _ = io.WriteString(w, `{"id":"fixture-response","model":"gpt-5.4","status":"completed","output":[],"usage":{"input_tokens":11,"output_tokens":7}}`)
	}))
	defer upstream.Close()
	cfg := &config.Config{RunMode: config.RunModeSimple}
	cfg.Security.URLAllowlist.AllowInsecureHTTP = true
	cfg.Default.RateMultiplier = 1
	accounts := []service.Account{}
	for i, model := range []string{"only-other", "gpt-5.4"} {
		accounts = append(accounts, service.Account{
			ID: int64(3 + i), Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey,
			Status: service.StatusActive, Schedulable: true, Priority: i + 1, Concurrency: 1,
			Credentials: map[string]any{"api_key": "fixture", "base_url": upstream.URL, "model_mapping": map[string]any{model: "unreachable-target"}},
			Extra:       map[string]any{"openai_passthrough": true, "openai_passthrough_model_filter": true},
		})
	}
	logs := repository.NewUsageLogRepository(nil, db)
	cache := service.NewBillingCacheService(nil, nil, nil, nil, nil, nil, cfg, nil)
	t.Cleanup(cache.Stop)
	gateway := service.NewOpenAIGatewayService(&openAIWSFailoverHandlerAccountRepoStub{accounts: accounts}, logs, nil, nil, nil, nil, nil, cfg, nil, nil, service.NewBillingService(cfg, nil), nil, cache, grokAuditHTTP{upstream.Client()}, service.NewDeferredService(nil, nil, time.Hour), nil, nil, nil, nil, nil, nil, nil)
	h := NewOpenAIGatewayHandler(gateway, service.NewConcurrencyService(nil), cache, service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, cfg), nil, nil, nil, nil, cfg)
	group := &service.Group{ID: 7, Platform: service.PlatformOpenAI, Status: service.StatusActive, MaxReasoningEffort: "high", RateMultiplier: 1}
	router := gin.New()
	router.POST("/v1/responses", func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{ID: 2, UserID: 1, User: &service.User{ID: 1, Status: service.StatusActive}, GroupID: &group.ID, Group: group})
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1})
		h.Responses(c)
	})
	server := httptest.NewServer(router)
	defer server.Close()
	resp, err := server.Client().Post(server.URL+"/v1/responses", "application/json", strings.NewReader(`{"model":"gpt-5.4","input":"test","reasoning":{"effort":"xhigh"},"stream":false}`))
	require.NoError(t, err)
	defer resp.Body.Close()
	response, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(response))
	wire := <-observed
	require.Equal(t, "gpt-5.4", gjson.GetBytes(wire, "model").String())
	require.Equal(t, "high", gjson.GetBytes(wire, "reasoning.effort").String())
	var id int64
	require.Eventually(t, func() bool { return db.QueryRow("SELECT id FROM usage_logs LIMIT 1").Scan(&id) == nil }, 3*time.Second, 10*time.Millisecond)
	row, err := logs.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.EqualValues(t, 4, row.AccountID)
	require.Equal(t, "xhigh", *row.RequestedReasoningEffort)
	require.Equal(t, "high", *row.ReasoningEffort)
	require.Equal(t, "xhigh", *dto.UsageLogFromService(row).ReasoningEffort)
	require.Equal(t, "high", *dto.UsageLogFromServiceAdmin(row).UpstreamReasoningEffort)
}
