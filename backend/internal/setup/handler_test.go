package setup

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestInstallNotifiesCompletionAfterSuccess(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("DATA_DIR", dataDir)
	t.Setenv("SKIP_SETUP", "false")

	recorder := httptest.NewRecorder()
	notifications := 0
	router := gin.New()
	RegisterRoutes(router, func() {
		if recorder.Code != http.StatusOK {
			t.Errorf("completion notified before success response was written: status=%d", recorder.Code)
		}
		notifications++
	})

	body := fmt.Sprintf(`{
		"database":{"driver":"sqlite","path":%q},
		"redis":{"enabled":false,"host":"","port":6379,"username":"","password":"","db":0,"enable_tls":false},
		"admin":{"email":"admin@example.com","password":"password123"},
		"server":{"host":"0.0.0.0","port":8080,"mode":"release"}
	}`, filepath.Join(dataDir, "sub2api.db"))
	request := httptest.NewRequest(http.MethodPost, "/setup/install", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if notifications != 1 {
		t.Fatalf("completion notifications = %d, want 1", notifications)
	}
}

func TestInstallAdminCredentialsAndExistingUsers(t *testing.T) {
	for _, role := range []string{"", "admin", "user"} {
		t.Run("existing_"+role, func(t *testing.T) {
			dataDir := t.TempDir()
			t.Setenv("DATA_DIR", dataDir)
			t.Setenv("SKIP_SETUP", "false")
			cfg := &SetupConfig{Database: DatabaseConfig{Path: filepath.Join(dataDir, "sub2api.db")}}
			email, password := "", "password123"
			if role != "" {
				require.NoError(t, initializeDatabase(cfg))
				db, err := sql.Open("sqlite", buildSQLiteDSN(&cfg.Database))
				require.NoError(t, err)
				defer db.Close()
				_, err = db.Exec(`INSERT INTO users (email, password_hash, role) VALUES ('existing@example.com', 'original-hash', ?)`, role)
				require.NoError(t, err)
				email, password = "invalid", "short"
				defer func() {
					var hash string
					require.NoError(t, db.QueryRow("SELECT password_hash FROM users").Scan(&hash))
					require.Equal(t, "original-hash", hash)
				}()
			}
			notifications := 0
			router := gin.New()
			RegisterRoutes(router, func() { notifications++ })
			body := fmt.Sprintf(`{"database":{"path":%q},"redis":{"enabled":false},"admin":{"email":%q,"password":%q}}`, cfg.Database.Path, email, password)
			for _, wantStatus := range []int{http.StatusOK, http.StatusForbidden} {
				recorder := httptest.NewRecorder()
				request := httptest.NewRequest(http.MethodPost, "/setup/install", bytes.NewBufferString(body))
				request.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(recorder, request)
				require.Equal(t, wantStatus, recorder.Code, recorder.Body.String())
			}
			require.Equal(t, 1, notifications)
			require.FileExists(t, filepath.Join(dataDir, InstallLockFile))
		})
	}
}

func TestInstallDoesNotNotifyCompletionAfterValidationFailure(t *testing.T) {
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("SKIP_SETUP", "false")

	notifications := 0
	router := gin.New()
	RegisterRoutes(router, func() {
		notifications++
	})

	body := fmt.Sprintf(`{
		"database":{"driver":"sqlite","path":%q},
		"redis":{"enabled":false,"host":"","port":6379,"username":"","password":"","db":0,"enable_tls":false},
		"admin":{"email":"admin@example.com","password":"short"},
		"server":{"host":"0.0.0.0","port":8080,"mode":"release"}
	}`, filepath.Join(t.TempDir(), "sub2api.db"))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/setup/install", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	if notifications != 0 {
		t.Fatalf("completion notifications = %d, want 0", notifications)
	}
}
