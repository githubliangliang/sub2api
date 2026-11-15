package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	adminhandler "github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestUsageUserRankingRouteSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "ranking.db")+"?_time_format=sqlite")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)
	require.NoError(t, repository.ApplyMigrations(context.Background(), db))

	start := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	_, err = db.Exec(`
		INSERT INTO users (id, email, password_hash) VALUES
			(1, 'ranking-one@example.com', 'unused'),
			(2, 'ranking-two@example.com', 'unused');
		INSERT INTO usage_logs (
			user_id, api_key_id, account_id, group_id, request_id, model, requested_model,
			input_tokens, output_tokens, cache_creation_tokens, cache_read_tokens,
			total_cost, actual_cost, stream, request_type, billing_type, created_at
		) VALUES
			(1, 11, 21, 10, 'ranking-1', 'raw-alpha', 'alpha', 100, 20, 5, 5, 1, 1, TRUE, 0, 0, ?),
			(1, 11, 21, 10, 'ranking-2', 'raw-alpha', 'alpha', 50, 10, 0, 0, 2, 2, TRUE, ?, 0, ?),
			(2, 12, 22, 20, 'ranking-3', 'raw-beta', 'beta', 10, 20, 5, 5, 8, 8, FALSE, ?, 1, ?),
			(2, 12, 22, 20, 'ranking-before', 'raw-beta', 'beta', 1000, 0, 0, 0, 100, 100, FALSE, 0, 1, ?),
			(2, 12, 22, 20, 'ranking-after', 'raw-beta', 'beta', 1000, 0, 0, 0, 100, 100, FALSE, 0, 1, ?)`,
		start, int16(service.RequestTypeStream), start.Add(time.Hour),
		int16(service.RequestTypeSync), start.Add(23*time.Hour),
		start.Add(-time.Second), start.Add(24*time.Hour))
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	dashboard := service.NewDashboardService(repository.NewUsageLogRepository(nil, db), nil, nil, nil)
	handlers := &handler.Handlers{Admin: &handler.AdminHandlers{
		Dashboard: adminhandler.NewDashboardHandler(dashboard, nil),
	}}
	adminAuth := middleware.AdminAuthMiddleware(func(c *gin.Context) {
		switch c.GetHeader("Authorization") {
		case "Bearer admin-token":
			c.Next()
		case "":
			c.AbortWithStatus(http.StatusUnauthorized)
		default:
			c.AbortWithStatus(http.StatusForbidden)
		}
	})
	auditLog := middleware.AuditLogMiddleware(func(c *gin.Context) { c.Next() })
	stepUp := middleware.StepUpAuthMiddleware(func(c *gin.Context) { c.Next() })
	RegisterAdminRoutes(router.Group("/api/v1"), handlers, adminAuth, auditLog, stepUp, nil, nil)

	one := usagestats.UserBreakdownItem{
		UserID: 1, Email: "ranking-one@example.com", Requests: 2,
		InputTokens: 150, OutputTokens: 30, CacheTokens: 10, TotalTokens: 190,
		Cost: 3, ActualCost: 3, AccountCost: 3,
	}
	two := usagestats.UserBreakdownItem{
		UserID: 2, Email: "ranking-two@example.com", Requests: 1,
		InputTokens: 10, OutputTokens: 20, CacheTokens: 10, TotalTokens: 40,
		Cost: 8, ActualCost: 8, AccountCost: 8,
	}
	for _, tc := range []struct {
		name   string
		auth   string
		query  string
		status int
		users  []usagestats.UserBreakdownItem
	}{
		{name: "unauthenticated", status: http.StatusUnauthorized},
		{name: "non-admin", auth: "Bearer user-token", status: http.StatusForbidden},
		{name: "tokens and date boundaries", auth: "Bearer admin-token", query: "sort_by=total_tokens", status: http.StatusOK, users: []usagestats.UserBreakdownItem{one, two}},
		{name: "cost sorting", auth: "Bearer admin-token", query: "sort_by=actual_cost", status: http.StatusOK, users: []usagestats.UserBreakdownItem{two, one}},
		{name: "top limit", auth: "Bearer admin-token", query: "sort_by=total_tokens&limit=1", status: http.StatusOK, users: []usagestats.UserBreakdownItem{one}},
		{name: "group filter", auth: "Bearer admin-token", query: "group_id=20", status: http.StatusOK, users: []usagestats.UserBreakdownItem{two}},
		{name: "requested model filter", auth: "Bearer admin-token", query: "model=alpha", status: http.StatusOK, users: []usagestats.UserBreakdownItem{one}},
		{name: "shared filters and legacy stream", auth: "Bearer admin-token", query: "user_id=1&api_key_id=11&account_id=21&group_id=10&request_type=stream&billing_type=0", status: http.StatusOK, users: []usagestats.UserBreakdownItem{one}},
		{name: "empty ranking", auth: "Bearer admin-token", query: "group_id=99", status: http.StatusOK, users: []usagestats.UserBreakdownItem{}},
		{name: "invalid request type", auth: "Bearer admin-token", query: "request_type=invalid", status: http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet,
				"/api/v1/admin/dashboard/user-breakdown?start_date=2026-11-01&end_date=2026-11-01&timezone=UTC&"+tc.query, nil)
			request.Header.Set("Authorization", tc.auth)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			require.Equal(t, tc.status, recorder.Code, recorder.Body.String())
			if tc.status != http.StatusOK {
				return
			}
			var payload struct {
				Code int `json:"code"`
				Data struct {
					Users     []usagestats.UserBreakdownItem `json:"users"`
					StartDate string                         `json:"start_date"`
					EndDate   string                         `json:"end_date"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
			require.Zero(t, payload.Code)
			require.Equal(t, tc.users, payload.Data.Users)
			require.Equal(t, "2026-11-01", payload.Data.StartDate)
			require.Equal(t, "2026-11-01", payload.Data.EndDate)
		})
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard/stats", nil)
	request.Header.Set("Authorization", "Bearer admin-token")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusNotFound, recorder.Code, "restoring usage ranking must not register the full dashboard")
}
