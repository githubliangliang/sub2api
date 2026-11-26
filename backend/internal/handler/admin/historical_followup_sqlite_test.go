package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
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
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func followupSQLite(t *testing.T) (*sql.DB, *dbent.Client) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "followup.db")+"?_time_format=sqlite&_pragma=foreign_keys(1)")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	require.NoError(t, repository.ApplyMigrations(context.Background(), db))
	require.NoError(t, repository.EnsureSQLiteAuxTables(context.Background(), db))
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	return db, client
}

func followupHTTP(t *testing.T, server *httptest.Server, method, path, payload, apiKey string) (int, []byte) {
	t.Helper()
	req, err := http.NewRequest(method, server.URL+path, strings.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("x-api-key", apiKey)
	}
	res, err := server.Client().Do(req)
	require.NoError(t, err)
	defer func() { require.NoError(t, res.Body.Close()) }()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return res.StatusCode, body
}

func TestBackupRetainedSecretProtectedHTTPAndSQLite(t *testing.T) {
	_, client := followupSQLite(t)
	ctx := context.Background()
	_, err := client.User.Create().SetEmail("admin@example.invalid").SetPasswordHash("fixture").SetRole(service.RoleAdmin).Save(ctx)
	require.NoError(t, err)
	repo := repository.NewSettingRepository(client)
	require.NoError(t, repo.Set(ctx, service.SettingKeyAdminAPIKey, "fixture-admin-key"))
	cfg := &config.Config{Totp: config.TotpConfig{EncryptionKey: strings.Repeat("42", 32), EncryptionKeyConfigured: true}}
	enc, err := repository.NewAESEncryptor(cfg)
	require.NoError(t, err)
	// No store factory or dumper: saving configuration cannot start S3/backup work.
	backup := service.NewBackupService(repo, cfg, enc, nil, nil)
	t.Cleanup(backup.Stop)
	settings := service.NewSettingService(repo, cfg)
	users := service.NewUserService(repository.NewUserRepository(client, nil), repo, nil, nil)
	h := NewBackupHandler(backup, users, nil)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.HandlerFunc(middleware.NewAdminAuthMiddleware(nil, users, settings, nil)))
	router.GET("/api/v1/admin/backups/s3-config", h.GetS3Config)
	router.PUT("/api/v1/admin/backups/s3-config", gin.HandlerFunc(middleware.NewStepUpAuthMiddleware(nil, users, settings)), h.UpdateS3Config)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	const path = "/api/v1/admin/backups/s3-config"
	const secret = "synthetic-backup-secret"
	readRaw := func() string {
		raw, readErr := repo.GetValue(ctx, "backup_s3_config")
		require.NoError(t, readErr)
		return raw
	}
	status, _ := followupHTTP(t, server, http.MethodPut, path, `{"bucket":"unauthorized"}`, "")
	require.Equal(t, http.StatusUnauthorized, status)
	raw, err := repo.GetValue(ctx, "backup_s3_config")
	require.ErrorIs(t, err, service.ErrSettingNotFound)
	require.Empty(t, raw)
	for i, payload := range []string{
		`{"bucket":"first","access_key_id":"fixture","secret_access_key":"` + secret + `"}`,
		`{"bucket":"retained","access_key_id":"fixture"}`,
		`{"bucket":"retained-again","access_key_id":"fixture","secret_access_key":""}`,
		`{"bucket":"replacement","access_key_id":"fixture","secret_access_key":"synthetic-new-secret"}`,
	} {
		status, body := followupHTTP(t, server, http.MethodPut, path, payload, "fixture-admin-key")
		require.Equal(t, http.StatusOK, status, string(body))
		require.Empty(t, gjson.GetBytes(body, "data.secret_access_key").String())
		var stored service.BackupS3Config
		require.NoError(t, json.Unmarshal([]byte(readRaw()), &stored))
		plain, decryptErr := enc.Decrypt(stored.SecretAccessKey)
		require.NoError(t, decryptErr)
		expected := secret
		if i == 3 {
			expected = "synthetic-new-secret"
		}
		require.Equal(t, expected, plain)
		require.NotContains(t, readRaw(), expected)
		status, body = followupHTTP(t, server, http.MethodGet, path, "", "fixture-admin-key")
		require.Equal(t, http.StatusOK, status)
		require.Empty(t, gjson.GetBytes(body, "data.secret_access_key").String())
	}
	// The existing plaintext compatibility reader also crosses the encryption
	// boundary when an operator retains a legacy secret.
	require.NoError(t, repo.Set(ctx, "backup_s3_config", `{"bucket":"legacy","secret_access_key":"synthetic-legacy-secret"}`))
	status, body := followupHTTP(t, server, http.MethodPut, path, `{"bucket":"legacy-retained"}`, "fixture-admin-key")
	require.Equal(t, http.StatusOK, status, string(body))
	var legacy service.BackupS3Config
	require.NoError(t, json.Unmarshal([]byte(readRaw()), &legacy))
	plain, err := enc.Decrypt(legacy.SecretAccessKey)
	require.NoError(t, err)
	require.Equal(t, "synthetic-legacy-secret", plain)
	require.NotContains(t, readRaw(), plain)
	before := readRaw()
	require.NoError(t, repo.Set(ctx, service.SettingKeyStepUpEnabled, "true"))
	status, body = followupHTTP(t, server, http.MethodPut, path, `{"bucket":"denied"}`, "fixture-admin-key")
	require.Equal(t, http.StatusForbidden, status)
	require.Contains(t, string(body), "STEP_UP_ADMIN_API_KEY_FORBIDDEN")
	require.Equal(t, before, readRaw())
	require.NoError(t, repo.Set(ctx, service.SettingKeyStepUpEnabled, "false"))
	_, err = client.ExecContext(ctx, `CREATE TRIGGER fail_backup_config BEFORE UPDATE ON settings WHEN NEW.key='backup_s3_config' BEGIN SELECT RAISE(ABORT,'fixture write failure'); END`)
	require.NoError(t, err)
	status, _ = followupHTTP(t, server, http.MethodPut, path, `{"bucket":"write-failed"}`, "fixture-admin-key")
	require.Equal(t, http.StatusInternalServerError, status)
	require.Equal(t, before, readRaw())
}

func TestUserAttributePayloadHTTPAndSQLite(t *testing.T) {
	_, client := followupSQLite(t)
	ctx := context.Background()
	u, err := client.User.Create().SetEmail("attributes@example.invalid").SetPasswordHash("fixture").Save(ctx)
	require.NoError(t, err)
	defs := repository.NewUserAttributeDefinitionRepository(client)
	values := repository.NewUserAttributeValueRepository(client)
	h := NewUserAttributeHandler(service.NewUserAttributeService(defs, values))
	router := gin.New()
	router.POST("/admin/user-attributes", h.CreateDefinition)
	router.PUT("/admin/user-attributes/:id", h.UpdateDefinition)
	router.PUT("/admin/users/:id/attributes", h.UpdateUserAttributes)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	status, body := followupHTTP(t, server, http.MethodPost, "/admin/user-attributes", `{"key":"integer","name":"Integer","type":"number","description":"old description","placeholder":"old placeholder","enabled":true,"validation":{"min":-5,"max":12}}`, "")
	require.Equal(t, http.StatusOK, status, string(body))
	id := gjson.GetBytes(body, "data.id").Int()
	path := fmt.Sprintf("/admin/users/%d/attributes", u.ID)
	for _, tc := range []struct {
		value  string
		status int
	}{
		{`"12"`, 200}, {`"0"`, 200}, {`"-3"`, 200}, {`""`, 200},
		{`12`, 400}, {`"12.5"`, 400}, {`"13"`, 400}, {`"-6"`, 400},
	} {
		t.Run(tc.value, func(t *testing.T) {
			before, getErr := values.GetByUserID(ctx, u.ID)
			require.NoError(t, getErr)
			status, body := followupHTTP(t, server, http.MethodPut, path, fmt.Sprintf(`{"values":{"%d":%s}}`, id, tc.value), "")
			require.Equal(t, tc.status, status, string(body))
			after, getErr := values.GetByUserID(ctx, u.ID)
			require.NoError(t, getErr)
			if status == http.StatusOK {
				require.Len(t, after, 1)
				require.Equal(t, strings.Trim(tc.value, `"`), after[0].Value)
			} else {
				require.Equal(t, before, after, "invalid values cannot write")
			}
		})
	}
	defPath := fmt.Sprintf("/admin/user-attributes/%d", id)
	for _, tc := range []struct{ payload, description, placeholder string }{
		{`{"name":"renamed"}`, "old description", "old placeholder"},
		{`{"description":"","placeholder":""}`, "", ""},
		{`{"description":"new text","placeholder":"new hint"}`, "new text", "new hint"},
	} {
		status, body := followupHTTP(t, server, http.MethodPut, defPath, tc.payload, "")
		require.Equal(t, http.StatusOK, status, string(body))
		stored, getErr := defs.GetByID(ctx, id)
		require.NoError(t, getErr)
		require.Equal(t, tc.description, stored.Description)
		require.Equal(t, tc.placeholder, stored.Placeholder)
	}
	status, body = followupHTTP(t, server, http.MethodPost, "/admin/user-attributes", `{"key":"text","name":"Text","type":"text","enabled":true}`, "")
	require.Equal(t, http.StatusOK, status, string(body))
	textID := gjson.GetBytes(body, "data.id").Int()
	status, body = followupHTTP(t, server, http.MethodPut, path, fmt.Sprintf(`{"values":{"%d":"ordinary text"}}`, textID), "")
	require.Equal(t, http.StatusOK, status, string(body))
	stored, err := values.GetByUserID(ctx, u.ID)
	require.NoError(t, err)
	require.Len(t, stored, 2)
}

func TestUserEditorOriginAndOmittedFieldsHTTPAndSQLite(t *testing.T) {
	db, client := followupSQLite(t)
	ctx := context.Background()
	repo := repository.NewUserRepository(client, db)
	u := &service.User{Email: "origin@example.invalid", PasswordHash: "fixture", Role: service.RoleUser, Status: service.StatusActive, Balance: 17, Concurrency: 1}
	v := &service.User{Email: "other@example.invalid", PasswordHash: "fixture", Role: service.RoleUser, Status: service.StatusActive, Balance: 29, Concurrency: 1}
	require.NoError(t, repo.Create(ctx, u))
	require.NoError(t, repo.Create(ctx, v))
	g, err := client.Group.Create().SetName("exclusive").SetIsExclusive(true).Save(ctx)
	require.NoError(t, err)
	u.AllowedGroups = []int64{g.ID}
	require.NoError(t, repo.Update(ctx, u, service.UserUpdateFields{AllowedGroups: true}))
	adminSvc := service.NewAdminService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, client, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	h := NewUserHandler(adminSvc, nil, nil, nil, nil, nil, nil)
	defs := repository.NewUserAttributeDefinitionRepository(client)
	values := repository.NewUserAttributeValueRepository(client)
	attrSvc := service.NewUserAttributeService(defs, values)
	def, err := attrSvc.CreateDefinition(ctx, service.CreateAttributeDefinitionInput{Key: "note", Name: "Note", Type: service.AttributeTypeText, Enabled: true})
	require.NoError(t, err)
	for userID, value := range map[int64]string{u.ID: "origin", v.ID: "other"} {
		require.NoError(t, attrSvc.UpdateUserAttributes(ctx, userID, []service.UpdateUserAttributeInput{{AttributeID: def.ID, Value: value}}))
	}
	attrHandler := NewUserAttributeHandler(attrSvc)
	router := gin.New()
	router.PUT("/admin/users/:id", h.Update)
	router.GET("/admin/users/:id/attributes", attrHandler.GetUserAttributes)
	router.PUT("/admin/users/:id/attributes", attrHandler.UpdateUserAttributes)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	// A general save omits unavailable group/attribute data rather than clearing it.
	status, body := followupHTTP(t, server, http.MethodPut, fmt.Sprintf("/admin/users/%d", u.ID), `{"username":"changed"}`, "")
	require.Equal(t, http.StatusOK, status, string(body))
	loaded, err := repo.GetByID(ctx, u.ID)
	require.NoError(t, err)
	require.Equal(t, []int64{g.ID}, loaded.AllowedGroups)
	require.Equal(t, 17.0, loaded.Balance)
	_, err = db.Exec(`ALTER TABLE user_attribute_definitions RENAME TO fixture_unavailable_definitions`)
	require.NoError(t, err)
	status, _ = followupHTTP(t, server, http.MethodPut, fmt.Sprintf("/admin/users/%d/attributes", u.ID), fmt.Sprintf(`{"values":{"%d":"failed"}}`, def.ID), "")
	require.Equal(t, http.StatusInternalServerError, status)
	_, err = db.Exec(`ALTER TABLE fixture_unavailable_definitions RENAME TO user_attribute_definitions`)
	require.NoError(t, err)
	before, err := values.GetByUserID(ctx, u.ID)
	require.NoError(t, err)
	require.Equal(t, "origin", before[0].Value)
	// The originating user ID remains the target of an already captured save.
	status, body = followupHTTP(t, server, http.MethodPut, fmt.Sprintf("/admin/users/%d/attributes", u.ID), fmt.Sprintf(`{"values":{"%d":"origin snapshot"}}`, def.ID), "")
	require.Equal(t, http.StatusOK, status, string(body))
	other, err := values.GetByUserID(ctx, v.ID)
	require.NoError(t, err)
	require.Equal(t, "other", other[0].Value)
	for _, userID := range []int64{u.ID, v.ID} {
		status, body = followupHTTP(t, server, http.MethodGet, fmt.Sprintf("/admin/users/%d/attributes", userID), "", "")
		require.Equal(t, http.StatusOK, status, string(body))
	}
	// Explicit loaded empty selection retains the existing clear contract.
	status, body = followupHTTP(t, server, http.MethodPut, fmt.Sprintf("/admin/users/%d", u.ID), `{"allowed_groups":[]}`, "")
	require.Equal(t, http.StatusOK, status, string(body))
	loaded, err = repo.GetByID(ctx, u.ID)
	require.NoError(t, err)
	require.Empty(t, loaded.AllowedGroups)
	untouched, err := repo.GetByID(ctx, v.ID)
	require.NoError(t, err)
	require.Equal(t, 29.0, untouched.Balance)
}
