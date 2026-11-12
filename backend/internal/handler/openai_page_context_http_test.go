//go:build unit

package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestCodexPageContextRealHTTPFirstRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	captured := make(chan []byte, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		captured <- body
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"resp_page","status":"completed","output":[],"usage":{"input_tokens":2,"output_tokens":0}}`)
	}))
	defer upstream.Close()
	cfg := &config.Config{RunMode: config.RunModeSimple, Default: config.DefaultConfig{RateMultiplier: 1}, Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false, AllowInsecureHTTP: true}}}
	repo := openAIImagesFailoverAccountRepo{accounts: []service.Account{{ID: 1, Name: "page-fixture", Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Status: service.StatusActive, Schedulable: true, Credentials: map[string]any{"api_key": "page-fixture", "base_url": upstream.URL}}}}
	deferred := service.NewDeferredService(repo, nil, time.Minute)
	gateway := service.NewOpenAIGatewayService(repo, nil, nil, nil, nil, nil, nil, cfg, nil, nil, service.NewBillingService(cfg, nil), nil, nil, grokAuditHTTP{upstream.Client()}, deferred, nil, nil, nil, nil, nil, nil, nil)
	billing := service.NewBillingCacheService(nil, nil, nil, nil, nil, nil, cfg, nil)
	defer billing.Stop()
	h := NewOpenAIGatewayHandler(gateway, service.NewConcurrencyService(nil), billing, service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, cfg), nil, nil, nil, nil, cfg)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		groupID := int64(1)
		c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{ID: 1, GroupID: &groupID, Group: &service.Group{ID: 1, Platform: service.PlatformOpenAI, Status: service.StatusActive}, User: &service.User{ID: 1, Status: service.StatusActive}})
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1})
	})
	router.POST("/v1/responses", h.Responses)
	server := httptest.NewServer(router)
	defer server.Close()
	body := pageContextRequest(t, pageContextItem("codex_apps_open_page", `{"page_id":"page-fixture","text":"untrusted page content"}`), pageContextItem("codex_apps_selected_page_part_1_of_1", "Companion Page snapshot part 1 of 1:\n"+`{"page_id":null}`), map[string]any{"role": "user", "content": "hello"})
	response, err := server.Client().Post(server.URL+"/v1/responses", "application/json", strings.NewReader(string(body)))
	require.NoError(t, err)
	raw, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.Equal(t, 200, response.StatusCode, string(raw))
	select {
	case forwarded := <-captured:
		items := gjson.GetBytes(forwarded, "input").Array()
		require.Len(t, items, 3)
		for _, item := range items[:2] {
			require.Equal(t, "user", item.Get("role").String())
			require.Equal(t, "message", item.Get("type").String())
			require.Contains(t, item.Get("content.0.text").String(), "<external_codex_apps_")
		}
		require.Contains(t, items[0].Get("content.0.text").String(), "untrusted page content")
		require.Equal(t, "hello", items[2].Get("content").String())
	case <-time.After(time.Second):
		t.Fatal("the first request never reached the HTTP upstream")
	}
}
