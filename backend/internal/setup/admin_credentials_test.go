package setup

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/require"
)

func TestAdminBootstrapCredentialsSQLite(t *testing.T) {
	for _, tc := range []struct {
		name, email, password, existingRole string
		invalid                             bool
	}{
		{name: "generated"},
		{name: "explicit", email: " owner@example.com ", password: "12345678"},
		{name: "bcrypt limit", email: "owner@example.com", password: strings.Repeat("a", 72)},
		{name: "unicode limit", email: "owner@example.com", password: strings.Repeat("密", 24)},
		{name: "short", email: "owner@example.com", password: "123", invalid: true},
		{name: "too long", email: "owner@example.com", password: strings.Repeat("a", 73), invalid: true},
		{name: "unicode overflow", email: "owner@example.com", password: strings.Repeat("密", 25), invalid: true},
		{name: "short domain", email: "a@b", password: "12345678", invalid: true},
		{name: "display name", email: "Owner <owner@example.com>", password: "12345678", invalid: true},
		{name: "existing admin", email: "invalid", password: "123", existingRole: "admin"},
		{name: "existing user", email: "invalid", password: "123", existingRole: "user"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &SetupConfig{Database: DatabaseConfig{Path: filepath.Join(t.TempDir(), "bootstrap.db")}, Admin: AdminConfig{Email: tc.email, Password: tc.password}}
			db, err := sql.Open("sqlite", buildSQLiteDSN(&cfg.Database))
			require.NoError(t, err)
			t.Cleanup(func() { _ = db.Close() })
			_, err = db.Exec(`CREATE TABLE users (email TEXT, password_hash TEXT, role TEXT, balance REAL, concurrency INTEGER, status TEXT, created_at DATETIME, updated_at DATETIME)`)
			require.NoError(t, err)
			if tc.existingRole != "" {
				_, err = db.Exec(`INSERT INTO users (email, password_hash, role) VALUES ('existing@example.com', 'existing-hash', ?)`, tc.existingRole)
				require.NoError(t, err)
			}
			created, _, err := createAdminUser(cfg)
			if tc.invalid {
				require.Error(t, err)
				var count int
				require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count))
				require.Zero(t, count)
				return
			}
			require.NoError(t, err)
			var email, hash string
			require.NoError(t, db.QueryRow("SELECT email, password_hash FROM users").Scan(&email, &hash))
			if tc.existingRole != "" {
				require.False(t, created)
				require.Equal(t, "existing@example.com", email)
				require.Equal(t, "existing-hash", hash)
				return
			}
			require.True(t, created)
			login := struct {
				Email string `binding:"required,email"`
			}{Email: email}
			require.NoError(t, binding.Validator.ValidateStruct(&login))
			require.True(t, (&service.User{PasswordHash: hash}).CheckPassword(cfg.Admin.Password))
			if tc.email == "" {
				require.Regexp(t, `^admin-[0-9a-f]{12}@sub2api\.local$`, email)
			} else {
				require.Equal(t, strings.TrimSpace(tc.email), email)
			}
		})
	}
}
