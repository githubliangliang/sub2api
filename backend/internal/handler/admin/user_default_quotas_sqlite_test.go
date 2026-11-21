package admin

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestAdminUserCreationDefaultPlatformQuotaHTTPAndSQLite(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "new-user.db")+"?_time_format=sqlite&_pragma=foreign_keys(1)")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	require.NoError(t, repository.ApplyMigrations(ctx, db))
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	defer func() { require.NoError(t, client.Close()) }()
	settingsRepo := repository.NewSettingRepository(client)
	require.NoError(t, settingsRepo.Set(ctx, service.SettingKeyDefaultPlatformQuotas, `{"openai":{"daily":0},"anthropic":{"weekly":12},"gemini":{}}`))
	settings := service.NewSettingService(settingsRepo, &config.Config{})
	quotaRepo := repository.NewUserPlatformQuotaRepository(client)
	quotaPort := repository.NewUserPlatformQuotaServiceAdapter(quotaRepo)
	svc := service.NewAdminService(repository.NewUserRepository(client, db), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, client, settings, nil, nil, nil, nil, nil, nil, nil, nil, quotaPort)
	h := NewUserHandler(svc, nil, nil, nil, nil, nil, nil)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/users", h.Create)
	server := httptest.NewServer(router)
	defer server.Close()
	resp, err := http.Post(server.URL+"/users", "application/json", strings.NewReader(`{"email":"new-user@example.invalid","password":"fixture-password","concurrency":1}`))
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()
	require.Equal(t, 200, resp.StatusCode)
	var userID int64
	require.NoError(t, db.QueryRow("SELECT id FROM users WHERE email='new-user@example.invalid'").Scan(&userID))
	zero, err := quotaRepo.GetByUserPlatform(ctx, userID, service.PlatformOpenAI)
	require.NoError(t, err)
	require.NotNil(t, zero)
	require.NotNil(t, zero.DailyLimitUSD)
	require.Zero(t, *zero.DailyLimitUSD)
	weekly, err := quotaRepo.GetByUserPlatform(ctx, userID, service.PlatformAnthropic)
	require.NoError(t, err)
	require.NotNil(t, weekly)
	require.Equal(t, 12.0, *weekly.WeeklyLimitUSD)
	unlimited, err := quotaRepo.GetByUserPlatform(ctx, userID, service.PlatformGemini)
	require.NoError(t, err)
	require.Nil(t, unlimited)
	// Existing rows are never overwritten by an initial-default snapshot.
	limit := 99.0
	require.NoError(t, quotaRepo.BulkInsertInitial(ctx, []repository.UserPlatformQuotaRecord{{UserID: userID, Platform: service.PlatformOpenAI, DailyLimitUSD: &limit}}))
	zero, err = quotaRepo.GetByUserPlatform(ctx, userID, service.PlatformOpenAI)
	require.NoError(t, err)
	require.Zero(t, *zero.DailyLimitUSD)
	// A best-effort quota failure must still allow a second user to be created.
	_, err = db.Exec(`CREATE TRIGGER fail_initial_quota BEFORE INSERT ON user_platform_quotas BEGIN SELECT RAISE(ABORT, 'fixture quota failure'); END`)
	require.NoError(t, err)
	w := doJSON(t, router, http.MethodPost, "/users", map[string]any{"email": "fail-open@example.invalid", "password": "fixture-password", "concurrency": 1})
	require.Equal(t, 200, w.Code)
	require.Positive(t, gjson.GetBytes(w.Body.Bytes(), "data.id").Int())
}
