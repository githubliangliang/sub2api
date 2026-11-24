package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

type opsWSUserRepo struct {
	service.UserRepository
	users map[int64]*service.User
}

func (r *opsWSUserRepo) GetByID(_ context.Context, id int64) (*service.User, error) {
	if user := r.users[id]; user != nil {
		return user, nil
	}
	return nil, service.ErrUserNotFound
}

func (r *opsWSUserRepo) GetUserAvatar(context.Context, int64) (*service.UserAvatar, error) {
	return nil, nil
}

func (r *opsWSUserRepo) UpdateUserLastActiveAt(context.Context, int64, time.Time) error {
	return nil
}

// Exercise the real admin middleware, upgrader and QPS connection lifecycle over
// TCP when changing Gorilla. User lookup is synthetic; no DB/Redis is required.
func TestOpsWSSecurityAndLifecycle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	oldConfig, oldLimits := opsWSProxyConfig, opsWSLimits
	opsWSProxyConfig = OpsWSProxyConfig{OriginPolicy: OriginPolicyStrict}
	opsWSLimits = opsWSRuntimeLimits{MaxConns: 1, MaxConnsPerIP: 1}
	t.Cleanup(func() {
		require.Eventually(t, func() bool { return wsConnCount.Load() == 0 }, time.Second, time.Millisecond)
		qpsWSCache.Stop()
		cancelQPSWSIdleStop()
		opsWSProxyConfig, opsWSLimits = oldConfig, oldLimits
	})

	users := map[int64]*service.User{
		1: {ID: 1, Email: "admin.synthetic@example.com", Role: service.RoleAdmin, Status: service.StatusActive, TokenVersion: 2},
		2: {ID: 2, Email: "user.synthetic@example.com", Role: service.RoleUser, Status: service.StatusActive, TokenVersion: 1},
		3: {ID: 3, Email: "inactive.synthetic@example.com", Role: service.RoleAdmin, Status: service.StatusDisabled, TokenVersion: 1},
	}
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "ops-ws-synthetic-test-secret", ExpireHour: 1}}
	auth := service.NewAuthService(nil, nil, nil, nil, cfg, nil, nil, nil, nil, nil, nil, nil, nil)
	userService := service.NewUserService(&opsWSUserRepo{users: users}, nil, nil, nil)
	router := gin.New()
	router.Use(gin.HandlerFunc(middleware.NewAdminAuthMiddleware(auth, userService, nil, nil)))
	router.GET("/qps", NewOpsHandler(&service.OpsService{}).QPSWSHandler)
	server := httptest.NewServer(router)
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/qps"
	tokenFor := func(user *service.User) string {
		token, err := auth.GenerateToken(context.Background(), user)
		require.NoError(t, err)
		return token
	}
	adminToken := tokenFor(users[1])
	dial := func(token, origin string) (*websocket.Conn, *http.Response, error) {
		d := websocket.Dialer{HandshakeTimeout: time.Second, Subprotocols: []string{"jwt." + token, "sub2api-admin"}}
		if token == "" {
			d.Subprotocols = []string{"sub2api-admin"}
		}
		h := http.Header{}
		if origin != "" {
			h.Set("Origin", origin)
		}
		return d.Dial(wsURL, h)
	}

	for _, tc := range []struct {
		name, token, origin string
		status              int
	}{
		{"missing-auth", "", server.URL, http.StatusUnauthorized},
		{"malformed-auth", "synthetic-invalid", server.URL, http.StatusUnauthorized},
		{"non-admin", tokenFor(users[2]), server.URL, http.StatusForbidden},
		{"inactive-admin", tokenFor(users[3]), server.URL, http.StatusUnauthorized},
		{"revoked-token", tokenFor(&service.User{ID: 1, Email: users[1].Email, Role: service.RoleAdmin, TokenVersion: 1}), server.URL, http.StatusUnauthorized},
		{"foreign-origin", adminToken, "https://foreign.synthetic.example", http.StatusForbidden},
		{"missing-strict-origin", adminToken, "", http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conn, response, err := dial(tc.token, tc.origin)
			require.Error(t, err)
			require.Nil(t, conn)
			require.NotNil(t, response)
			defer func() { _ = response.Body.Close() }()
			require.Equal(t, tc.status, response.StatusCode)
			require.Empty(t, response.Header.Get("Sec-WebSocket-Protocol"))
			require.Eventually(t, func() bool { return wsConnCount.Load() == 0 }, time.Second, time.Millisecond)
		})
	}

	t.Run("single-protocol-ping-pong-close-and-limits", func(t *testing.T) {
		conn, response, err := dial(adminToken, server.URL)
		require.NoError(t, err)
		defer func() { _ = conn.Close() }()
		require.Equal(t, "sub2api-admin", conn.Subprotocol())
		require.Equal(t, "sub2api-admin", response.Header.Get("Sec-WebSocket-Protocol"))
		require.NotContains(t, response.Header.Get("Sec-WebSocket-Protocol"), adminToken)

		second, denied, err := dial(adminToken, server.URL)
		require.Error(t, err)
		require.Nil(t, second)
		require.NotNil(t, denied)
		require.Equal(t, http.StatusServiceUnavailable, denied.StatusCode)
		require.NoError(t, denied.Body.Close())
		require.EqualValues(t, 1, wsConnCount.Load())

		pongReceived := make(chan string, 1)
		conn.SetPongHandler(func(payload string) error { pongReceived <- payload; return nil })
		readDone := make(chan error, 1)
		go func() {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					readDone <- err
					return
				}
			}
		}()
		require.NoError(t, conn.WriteControl(websocket.PingMessage, []byte("control"), time.Now().Add(time.Second)))
		select {
		case payload := <-pongReceived:
			require.Equal(t, "control", payload)
		case <-time.After(time.Second):
			t.Fatal("admin WS did not pong")
		}
		require.NoError(t, conn.WriteControl(websocket.PongMessage, nil, time.Now().Add(time.Second)))
		require.NoError(t, conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "done"), time.Now().Add(time.Second)))
		select {
		case err := <-readDone:
			require.True(t, websocket.IsCloseError(err, websocket.CloseNormalClosure), "%v", err)
		case <-time.After(time.Second):
			t.Fatal("admin WS did not acknowledge close")
		}
		require.Eventually(t, func() bool { return wsConnCount.Load() == 0 }, time.Second, time.Millisecond)
		wsConnCountByIPMu.Lock()
		remainingIPs := len(wsConnCountByIP)
		wsConnCountByIPMu.Unlock()
		require.Zero(t, remainingIPs)
	})

	t.Run("per-ip-limit-releases-reserved-total-slot", func(t *testing.T) {
		opsWSLimits.MaxConns = 2
		defer func() { opsWSLimits.MaxConns = 1 }()
		conn, _, err := dial(adminToken, server.URL)
		require.NoError(t, err)
		defer func() { _ = conn.Close() }()
		second, response, err := dial(adminToken, server.URL)
		require.Error(t, err)
		require.Nil(t, second)
		require.NotNil(t, response)
		require.Equal(t, http.StatusServiceUnavailable, response.StatusCode)
		require.NoError(t, response.Body.Close())
		require.Eventually(t, func() bool { return wsConnCount.Load() == 1 }, time.Second, time.Millisecond)
		require.NoError(t, conn.Close())
		require.Eventually(t, func() bool { return wsConnCount.Load() == 0 }, time.Second, time.Millisecond)
	})

	t.Run("oversized-frame-releases-slot", func(t *testing.T) {
		conn, _, err := dial(adminToken, server.URL)
		require.NoError(t, err)
		defer func() { _ = conn.Close() }()
		require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte(strings.Repeat("x", qpsWSMaxReadBytes+1))))
		require.NoError(t, conn.SetReadDeadline(time.Now().Add(time.Second)))
		_, _, err = conn.ReadMessage()
		require.True(t, websocket.IsCloseError(err, websocket.CloseMessageTooBig), "%v", err)
		require.Eventually(t, func() bool { return wsConnCount.Load() == 0 }, time.Second, time.Millisecond)
	})
}
