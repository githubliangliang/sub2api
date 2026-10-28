package admin

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type auditClearUserRepository struct {
	service.UserRepository
	user *service.User
}

func (r *auditClearUserRepository) GetByID(context.Context, int64) (*service.User, error) {
	return r.user, nil
}

func TestAuditLogClearHandlerSQLiteAndTotp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, body, auth       string
		enabled, authenticated bool
		attempts, status       int
	}{
		{"verified", "valid", "jwt", true, true, 0, http.StatusOK},
		{"invalid code", `{"totp_code":"abcdef"}`, "jwt", true, true, 0, http.StatusBadRequest},
		{"missing code", `{}`, "jwt", true, true, 0, http.StatusBadRequest},
		{"disabled totp", "valid", "jwt", false, true, 0, http.StatusBadRequest},
		{"machine credential", "valid", service.AuditAuthMethodAdminAPIKey, true, true, 0, http.StatusForbidden},
		{"unauthenticated", "valid", "jwt", true, false, 0, http.StatusUnauthorized},
		{"rate limited", "valid", "jwt", true, true, 10, http.StatusTooManyRequests},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "audit.db")+"?_time_format=sqlite")
			require.NoError(t, err)
			db.SetMaxOpenConns(1)
			t.Cleanup(func() { require.NoError(t, db.Close()) })
			require.NoError(t, repository.EnsureSQLiteAuxTables(ctx, db))
			repo := repository.NewAuditLogRepository(db)
			_, err = repo.BatchInsert(ctx, []*service.AuditLog{{Action: "before-1"}, {Action: "before-2"}})
			require.NoError(t, err)
			auditSvc := service.NewAuditLogService(repo, nil)
			t.Cleanup(auditSvc.Stop)
			mr := miniredis.RunT(t)
			rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			t.Cleanup(func() { require.NoError(t, rdb.Close()) })
			cache := repository.NewTotpCache(rdb)
			for range tc.attempts {
				_, err = cache.IncrementVerifyAttempts(ctx, 1)
				require.NoError(t, err)
			}
			cfg := &config.Config{}
			cfg.Totp.EncryptionKey = strings.Repeat("ab", 32)
			encryptor, err := repository.NewAESEncryptor(cfg)
			require.NoError(t, err)
			const secret = "JBSWY3DPEHPK3PXP"
			encrypted, err := encryptor.Encrypt(secret)
			require.NoError(t, err)
			user := &service.User{ID: 1, Role: service.RoleAdmin, TotpEnabled: tc.enabled, TotpSecretEncrypted: &encrypted}
			totpSvc := service.NewTotpService(&auditClearUserRepository{user: user}, encryptor, cache, nil, nil, nil)
			handler := NewAuditLogHandler(auditSvc, totpSvc)
			router := gin.New()
			router.POST("/api/v1/admin/audit-logs/clear", func(c *gin.Context) {
				c.Set("auth_method", tc.auth)
				if tc.authenticated {
					c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1})
					c.Set(string(middleware.ContextKeyUserRole), service.RoleAdmin)
					c.Set(middleware.ContextKeyAuthEmail, "admin@example.test")
				}
				handler.Clear(c)
			})
			body := tc.body
			if body == "valid" {
				code, err := totp.GenerateCode(secret, time.Now())
				require.NoError(t, err)
				encoded, err := json.Marshal(map[string]string{"totp_code": code})
				require.NoError(t, err)
				body = string(encoded)
			}
			req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/audit-logs/clear", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", "fixture-browser")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			require.Equal(t, tc.status, rec.Code, rec.Body.String())
			logs, err := repo.List(ctx, nil)
			require.NoError(t, err)
			if tc.status == http.StatusOK {
				var response struct {
					Data struct {
						Deleted int64 `json:"deleted"`
					} `json:"data"`
				}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				require.EqualValues(t, 2, response.Data.Deleted)
				require.Len(t, logs.Logs, 1)
				require.Equal(t, service.AuditActionAuditLogClear, logs.Logs[0].Action)
				require.Equal(t, "fixture-browser", logs.Logs[0].UserAgent)
				require.EqualValues(t, 1, *logs.Logs[0].ActorUserID)
			} else {
				require.Equal(t, 2, logs.Total)
			}
		})
	}
}
