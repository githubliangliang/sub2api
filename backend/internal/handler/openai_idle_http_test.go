//go:build unit

package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOpenAIIdleHTTPBudgetAndKeepalive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, scenario := range []string{"first-idle-then-success", "both-idle", "content-then-idle", "continuous-comments", "canceled"} {
		t.Run(scenario, func(t *testing.T) {
			var hits, acquired, released atomic.Int32
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				call := hits.Add(1)
				_, _ = io.Copy(io.Discard, r.Body)
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("X-Request-Id", fmt.Sprintf("account-%d-private", call))
				if scenario == "first-idle-then-success" && call == 2 {
					_, _ = io.WriteString(w, preContentHandlerSuccess)
					return
				}
				_, _ = io.WriteString(w, preContentHandlerCreated+"event: keepalive\ndata: {\"type\":\"keepalive\"}\n\n")
				w.(http.Flusher).Flush()
				if scenario == "content-then-idle" {
					_, _ = io.WriteString(w, `data: {"type":"response.output_text.delta","delta":"first account content"}`+"\n\n")
					w.(http.Flusher).Flush()
				}
				if scenario == "continuous-comments" {
					for range 5 {
						select {
						case <-time.After(300 * time.Millisecond):
						case <-r.Context().Done():
							return
						}
						_, _ = io.WriteString(w, ": keepalive\n\n")
						w.(http.Flusher).Flush()
					}
					_, _ = io.WriteString(w, preContentHandlerSuccess)
					return
				}
				<-r.Context().Done()
			}))
			t.Cleanup(upstream.Close)
			accounts := preContentAccounts(3)
			for i := range accounts {
				accounts[i].Credentials["base_url"] = upstream.URL
			}
			repo := &grokCredentialHandlerRepo{accounts: accounts, missingOnGet: map[int64]bool{}}
			cfg := &config.Config{RunMode: config.RunModeSimple, Gateway: config.GatewayConfig{MaxAccountSwitches: 3, StreamKeepaliveInterval: 1, StreamDataIntervalTimeout: 1}, Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{AllowInsecureHTTP: true}}}
			billing := service.NewBillingCacheService(nil, nil, nil, nil, nil, nil, cfg, nil)
			t.Cleanup(billing.Stop)
			cache := &concurrencyCacheMock{
				acquireUserSlotFn:    func(context.Context, int64, int, string) (bool, error) { return true, nil },
				acquireAccountSlotFn: func(context.Context, int64, int, string) (bool, error) { acquired.Add(1); return true, nil },
			}
			// The cache's ReleaseAccountSlot hook counts the real handler lease releases.
			leases := &tier223LeaseCache{concurrencyCacheMock: cache, released: &released}
			concurrency := service.NewConcurrencyService(leases)
			gateway := service.NewOpenAIGatewayService(repo, nil, nil, nil, nil, nil, nil, cfg, nil, concurrency, service.NewBillingService(cfg, nil), nil, billing, grokAuditHTTP{upstream.Client()}, &service.DeferredService{}, nil, nil, nil, nil, nil, nil, nil)
			h := NewOpenAIGatewayHandler(gateway, concurrency, billing, &service.APIKeyService{}, nil, nil, nil, nil, cfg)
			groupID := int64(9101)
			key := &service.APIKey{ID: 9102, GroupID: &groupID, User: &service.User{ID: 9103, Status: service.StatusActive}, Group: &service.Group{ID: groupID, Platform: service.PlatformOpenAI, Status: service.StatusActive}}
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set(string(middleware.ContextKeyAPIKey), key)
				c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: key.User.ID, Concurrency: 1})
				c.Next()
			})
			router.POST("/v1/responses", h.Responses)
			server := httptest.NewServer(router)
			t.Cleanup(server.Close)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/v1/responses", strings.NewReader(`{"model":"gpt-5.4","stream":true,"input":"hello"}`))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			response, err := server.Client().Do(req)
			require.NoError(t, err)
			var payload []byte
			if scenario == "canceled" {
				small := make([]byte, 16)
				_, err = response.Body.Read(small)
				require.NoError(t, err)
				cancel()
				_, _ = io.Copy(io.Discard, response.Body)
			} else {
				payload, err = io.ReadAll(response.Body)
				require.NoError(t, err)
			}
			require.NoError(t, response.Body.Close())
			wantHits := int32(1)
			if scenario == "first-idle-then-success" || scenario == "both-idle" {
				wantHits = 2
			}
			require.Equal(t, wantHits, hits.Load(), "the third account must not be consumed by two idle attempts")
			require.Eventually(t, func() bool { return released.Load() == acquired.Load() }, time.Second, 10*time.Millisecond)
			require.Equal(t, wantHits, acquired.Load())
			body := string(payload)
			if scenario == "first-idle-then-success" || scenario == "continuous-comments" {
				require.Contains(t, body, "answer from second account")
				require.Equal(t, 1, strings.Count(body, `"type":"response.completed"`))
			}
			if scenario == "both-idle" {
				require.Contains(t, body, "event: response.failed")
			}
			if scenario == "content-then-idle" {
				require.Contains(t, body, "event: error")
			}
			for _, line := range strings.Split(body, "\n") {
				require.False(t, strings.HasPrefix(line, "{"), "bare JSON in committed SSE")
			}
			if scenario == "content-then-idle" {
				require.Contains(t, body, "first account content")
			}
			if scenario == "first-idle-then-success" {
				require.Empty(t, response.Header.Get("X-Request-Id"), "account headers remain withheld after an early keepalive committed HTTP headers")
			}
			repo.mu.Lock()
			require.Empty(t, repo.setErrorIDs)
			require.Empty(t, repo.setTempIDs)
			require.Empty(t, repo.rateLimitIDs)
			repo.mu.Unlock()
		})
	}
}

type tier223LeaseCache struct {
	*concurrencyCacheMock
	released *atomic.Int32
}

func (c *tier223LeaseCache) ReleaseAccountSlot(context.Context, int64, string) error {
	c.released.Add(1)
	return nil
}
