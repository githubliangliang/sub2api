package admin

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	_ "github.com/Wei-Shaw/sub2api/ent/runtime"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestUserHandler_AdminGuardsSQLite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, method, reason  string
		body                  map[string]any
		otherAdmin, failWrite bool
		status                int
	}{
		{"last admin demotion", http.MethodPut, "LAST_ADMIN_DEMOTE_FORBIDDEN", map[string]any{"role": "user", "username": "changed"}, false, false, http.StatusBadRequest},
		{"direct disable", http.MethodPut, "ADMIN_DISABLE_FORBIDDEN", map[string]any{"status": "disabled"}, false, false, http.StatusBadRequest},
		{"disable with another admin", http.MethodPut, "ADMIN_DISABLE_FORBIDDEN", map[string]any{"status": "disabled"}, true, false, http.StatusBadRequest},
		{"simultaneous demotion and disable", http.MethodPut, "ADMIN_DISABLE_FORBIDDEN", map[string]any{"role": "user", "status": "disabled"}, true, false, http.StatusBadRequest},
		{"direct delete", http.MethodDelete, "ADMIN_DELETE_FORBIDDEN", nil, false, false, http.StatusBadRequest},
		{"delete with another admin", http.MethodDelete, "ADMIN_DELETE_FORBIDDEN", nil, true, false, http.StatusBadRequest},
		{"demotion with another admin", http.MethodPut, "", map[string]any{"role": "user", "username": "changed"}, true, false, http.StatusOK},
		{"database write failure", http.MethodPut, "", map[string]any{"role": "user", "username": "changed"}, true, true, http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "guards.db")+"?_time_format=sqlite&_pragma=foreign_keys(1)")
			require.NoError(t, err)
			db.SetMaxOpenConns(1)
			require.NoError(t, repository.ApplyMigrations(ctx, db))
			require.NoError(t, repository.EnsureSQLiteAuxTables(ctx, db))
			client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.SQLite, db)))
			t.Cleanup(func() { require.NoError(t, client.Close()) })
			target, err := client.User.Create().SetEmail("admin@example.test").SetPasswordHash("hash").SetRole(service.RoleAdmin).SetStatus(service.StatusActive).SetUsername("original").Save(ctx)
			require.NoError(t, err)
			if tc.otherAdmin {
				_, err = client.User.Create().SetEmail("other@example.test").SetPasswordHash("hash").SetRole(service.RoleAdmin).SetStatus(service.StatusActive).Save(ctx)
				require.NoError(t, err)
			}
			if tc.failWrite {
				_, err = db.ExecContext(ctx, `CREATE TRIGGER fail_admin_role BEFORE UPDATE OF role ON users BEGIN SELECT unavailable_admin_write_function(); END`)
				require.NoError(t, err)
			}
			svc := service.NewAdminService(repository.NewUserRepository(client, db), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, client, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			h := NewUserHandler(svc, nil, nil, nil, nil, nil, nil)
			router := gin.New()
			router.PUT("/api/v1/admin/users/:id", h.Update)
			router.DELETE("/api/v1/admin/users/:id", h.Delete)
			rec := doJSON(t, router, tc.method, fmt.Sprintf("/api/v1/admin/users/%d", target.ID), tc.body)
			// Verify persistence even when the old HTTP classification is wrong.
			stored, err := client.User.Get(ctx, target.ID)
			require.NoError(t, err)
			require.Equal(t, service.StatusActive, stored.Status)
			if tc.status == http.StatusOK {
				require.Equal(t, service.RoleUser, stored.Role)
				require.Equal(t, "changed", stored.Username)
			} else {
				require.Equal(t, service.RoleAdmin, stored.Role)
				require.Equal(t, "original", stored.Username)
				require.True(t, target.UpdatedAt.Equal(stored.UpdatedAt), "rejected actions must leave the persisted timestamp unchanged")
			}
			require.Equal(t, tc.status, rec.Code)
			require.Equal(t, tc.reason, gjson.GetBytes(rec.Body.Bytes(), "reason").String())
		})
	}
}
