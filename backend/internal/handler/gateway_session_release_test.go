//go:build unit

package handler

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	middleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type cancelAfterSessionRegistration struct {
	service.SessionLimitCache
	cancel context.CancelFunc
}

func (s cancelAfterSessionRegistration) RegisterSession(ctx context.Context, accountID int64, session string, max int, idle time.Duration) (bool, error) {
	allowed, err := s.SessionLimitCache.RegisterSession(ctx, accountID, session, max, idle)
	s.cancel()
	return allowed, err
}

type sessionReleaseHTTPUpstream struct{}

func (sessionReleaseHTTPUpstream) Do(*http.Request, string, int64, int) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"id":"msg-served","type":"message","role":"assistant","model":"claude-fable-5-1","content":[{"type":"text","text":"hello"}],"stop_reason":"end_turn","usage":{"input_tokens":17,"output_tokens":2}}`))}, nil
}

func (s sessionReleaseHTTPUpstream) DoWithTLS(req *http.Request, proxy string, accountID int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return s.Do(req, proxy, accountID, concurrency)
}

func TestGatewayFailedAnthropicRequestReleasesSessionSlot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, endpoint := range []string{"messages", "messages_canceled", "messages_success", "count_tokens"} {
		t.Run(endpoint, func(t *testing.T) {
			redisServer := miniredis.RunT(t)
			rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
			t.Cleanup(func() { _ = rdb.Close() })
			sessions := repository.NewSessionLimitCache(rdb, 5)
			requestCtx, cancel := context.WithCancel(context.Background())
			defer cancel()
			registrationCache := sessions
			if endpoint == "messages_canceled" {
				registrationCache = cancelAfterSessionRegistration{sessions, cancel}
			}
			groupID := int64(9100)
			group := &service.Group{ID: groupID, Hydrated: true, Platform: service.PlatformAnthropic, Status: service.StatusActive}
			account := &service.Account{
				ID: 9101, Platform: service.PlatformAnthropic, Type: service.AccountTypeOAuth,
				Status: service.StatusActive, Schedulable: true, Concurrency: 1,
				Extra:         map[string]any{"max_sessions": 1},
				AccountGroups: []service.AccountGroup{{AccountID: 9101, GroupID: groupID}},
			}
			if endpoint == "messages_success" {
				account.Credentials = map[string]any{"access_token": "test-token"}
			}
			snapshot := service.NewSchedulerSnapshotService(&fakeSchedulerCache{accounts: []*service.Account{account}}, nil, nil, nil, nil)
			cfg := &config.Config{RunMode: config.RunModeSimple}
			gateway := service.NewGatewayService(nil, &fakeGroupRepo{group: group}, nil, nil, nil, nil, nil, nil, cfg, snapshot, nil, service.NewBillingService(cfg, nil), nil, nil, nil, sessionReleaseHTTPUpstream{}, &service.DeferredService{}, nil, registrationCache, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			billing := service.NewBillingCacheService(nil, nil, nil, nil, nil, nil, cfg, nil)
			t.Cleanup(billing.Stop)
			h := &GatewayHandler{gatewayService: gateway, billingCacheService: billing, concurrencyHelper: NewConcurrencyHelper(service.NewConcurrencyService(&fakeConcurrencyCache{}), SSEPingFormatClaude, 0), cfg: cfg}
			key := &service.APIKey{ID: 9102, UserID: 9103, GroupID: &groupID, Group: group, Status: service.StatusActive, User: &service.User{ID: 9103, Concurrency: 10, Balance: 100}}
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			body := `{"model":"claude-fable-5-1","max_tokens":32,"messages":[{"role":"user","content":"hello"}],"metadata":{"user_id":"user_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_account_00000000-0000-0000-0000-000000000001_session_00000000-0000-0000-0000-000000000002"}}`
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewBufferString(body)).WithContext(context.WithValue(requestCtx, ctxkey.Group, group))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Set(string(middleware.ContextKeyAPIKey), key)
			c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: key.UserID, Concurrency: 10})
			if endpoint == "count_tokens" {
				allowed, err := sessions.RegisterSession(context.Background(), account.ID, "00000000-0000-0000-0000-000000000002", 1, 5*time.Minute)
				require.NoError(t, err)
				require.True(t, allowed)
				h.CountTokens(c)
			} else {
				h.Messages(c)
			}
			selected, exists := c.Get(opsAccountIDKey)
			if endpoint != "messages_canceled" {
				require.True(t, exists, recorder.Body.String())
				require.Equal(t, account.ID, selected)
			}
			if endpoint == "messages_success" {
				require.Equal(t, http.StatusOK, recorder.Code)
				require.Contains(t, recorder.Body.String(), "msg-served")
			}
			allowed, err := sessions.RegisterSession(context.Background(), account.ID, "next-session", 1, 5*time.Minute)
			require.NoError(t, err)
			if endpoint == "count_tokens" || endpoint == "messages_success" {
				require.False(t, allowed, "a served session must keep its slot until idle timeout")
			} else {
				require.True(t, allowed, "failed request must free the slot before another session arrives")
			}
		})
	}
}
