//go:build unit

package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/gemini"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type geminiDeclaredAccounts struct {
	service.AccountRepository
	accounts []service.Account
}

func (r *geminiDeclaredAccounts) ListSchedulableByGroupIDAndPlatforms(_ context.Context, _ int64, platforms []string) ([]service.Account, error) {
	var out []service.Account
	for _, account := range r.accounts {
		for _, platform := range platforms {
			if account.Platform == platform && account.Schedulable {
				out = append(out, account)
				break
			}
		}
	}
	return out, nil
}

func (r *geminiDeclaredAccounts) GetByID(_ context.Context, id int64) (*service.Account, error) {
	for _, account := range r.accounts {
		if account.ID == id {
			return &account, nil
		}
	}
	return nil, service.ErrAccountNotFound
}

type geminiDeclaredUpstream struct{ calls int }

func (u *geminiDeclaredUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	u.calls++
	body := `{"models":[{"name":"models/from-upstream"}]}`
	if !strings.HasSuffix(req.URL.Path, "/models") {
		body = `{"name":"models/from-upstream"}`
	}
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body))}, nil
}

func (u *geminiDeclaredUpstream) DoWithTLS(req *http.Request, proxy string, id int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxy, id, concurrency)
}

func geminiDeclaredFixture(t *testing.T, unmapped bool) (*GatewayHandler, *service.APIKey, *geminiDeclaredUpstream) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	credentials := func(mapping map[string]any) map[string]any {
		return map[string]any{"api_key": "synthetic-key", "base_url": "https://models.example.test", "model_mapping": mapping}
	}
	repo := &geminiDeclaredAccounts{accounts: []service.Account{
		{ID: 1, Platform: service.PlatformGemini, Type: service.AccountTypeAPIKey, Status: service.StatusActive, Schedulable: true, Credentials: credentials(map[string]any{"gemini-3.1-pro-preview": "gemini-3.1-pro-preview", "gemini-3.5-*": "gemini-3.5-flash"})},
		{ID: 2, Platform: service.PlatformGemini, Type: service.AccountTypeAPIKey, Status: service.StatusActive, Schedulable: true, Credentials: credentials(map[string]any{"gemini-3.1-pro-preview": "gemini-3.1-pro-preview"})},
	}}
	if unmapped {
		repo.accounts[1].Credentials = credentials(nil)
	}
	upstream := &geminiDeclaredUpstream{}
	svc := service.NewGeminiMessagesCompatService(repo, nil, nil, nil, nil, nil, upstream, nil, &config.Config{})
	group := &service.Group{ID: 1, Platform: service.PlatformGemini, Status: service.StatusActive}
	return &GatewayHandler{geminiCompatService: svc}, &service.APIKey{ID: 1, GroupID: &group.ID, Group: group}, upstream
}

func serveGeminiDeclared(t *testing.T, h *GatewayHandler, key *service.APIKey, model string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1beta/models/"+model, nil)
	c.Set(string(middleware.ContextKeyAPIKey), key)
	if model == "" {
		h.GeminiV1BetaListModels(c)
	} else {
		c.Params = gin.Params{{Key: "model", Value: model}}
		h.GeminiV1BetaGetModel(c)
	}
	return rec
}

func TestGeminiNativeDeclaredCatalogAndDetails(t *testing.T) {
	h, key, upstream := geminiDeclaredFixture(t, false)
	rec := serveGeminiDeclared(t, h, key, "")
	require.Equal(t, http.StatusOK, rec.Code)
	var list gemini.ModelsListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	var ids []string
	for _, model := range list.Models {
		ids = append(ids, model.Name)
	}
	require.Equal(t, []string{"models/gemini-3.1-pro-preview", "models/gemini-3.5-*"}, ids)
	for _, model := range []string{"gemini-3.5-flash", "gemini-3.1-pro-preview-customtools"} {
		rec = serveGeminiDeclared(t, h, key, model)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		require.Contains(t, rec.Body.String(), "models/"+model)
	}
	rec = serveGeminiDeclared(t, h, key, "gemini-unlisted")
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Zero(t, upstream.calls)
}

func TestGeminiNativeDeclaredDisplayListIsNotAuthorization(t *testing.T) {
	h, key, upstream := geminiDeclaredFixture(t, false)
	key.Group.ModelsListConfig = service.GroupModelsListConfig{Enabled: true, Models: []string{"gemini-3.5-flash"}}
	rec := serveGeminiDeclared(t, h, key, "")
	var list gemini.ModelsListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	require.Len(t, list.Models, 1)
	require.Equal(t, "models/gemini-3.5-flash", list.Models[0].Name)
	rec = serveGeminiDeclared(t, h, key, "gemini-3.1-pro-preview")
	require.Equal(t, http.StatusOK, rec.Code)
	require.Zero(t, upstream.calls)
}

func TestGeminiNativePartiallyUnmappedPreservesForwarding(t *testing.T) {
	h, key, upstream := geminiDeclaredFixture(t, true)
	for _, model := range []string{"", "gemini-unlisted"} {
		rec := serveGeminiDeclared(t, h, key, model)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		require.Contains(t, rec.Body.String(), "models/from-upstream")
	}
	require.Equal(t, 2, upstream.calls)
}

func TestGeminiNativeDeclaredKeepsPlatformAndPathGuards(t *testing.T) {
	h, key, upstream := geminiDeclaredFixture(t, false)
	rec := serveGeminiDeclared(t, h, key, "../invalid")
	require.Equal(t, http.StatusBadRequest, rec.Code)
	key.Group.Platform = service.PlatformOpenAI
	rec = serveGeminiDeclared(t, h, key, "")
	require.Equal(t, http.StatusBadRequest, rec.Code)
	rec = httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/antigravity/v1beta/models", nil)
	c.Set(string(middleware.ContextKeyAPIKey), key)
	c.Set(string(middleware.ContextKeyForcePlatform), service.PlatformAntigravity)
	h.GeminiV1BetaListModels(c)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "models/gemini-")
	require.Zero(t, upstream.calls)
}

func TestGeminiNativeDeclaredIncludesOnlyExistingMixedAntigravityModels(t *testing.T) {
	for _, mixed := range []bool{false, true} {
		h, key, upstream := geminiDeclaredFixture(t, false)
		repo := &geminiDeclaredAccounts{accounts: []service.Account{
			{ID: 1, Platform: service.PlatformGemini, Type: service.AccountTypeAPIKey, Schedulable: true, Credentials: map[string]any{"model_mapping": map[string]any{"gemini-3.5-flash": "gemini-3.5-flash"}}},
			{ID: 2, Platform: service.PlatformAntigravity, Type: service.AccountTypeOAuth, Schedulable: true, Extra: map[string]any{"mixed_scheduling": mixed}, Credentials: map[string]any{"model_mapping": map[string]any{"gemini-3-pro-high": "gemini-3-pro-high"}}},
		}}
		h.geminiCompatService = service.NewGeminiMessagesCompatService(repo, nil, nil, nil, nil, nil, upstream, nil, &config.Config{})
		rec := serveGeminiDeclared(t, h, key, "")
		if mixed {
			require.Contains(t, rec.Body.String(), "models/gemini-3-pro-high")
		} else {
			require.NotContains(t, rec.Body.String(), "models/gemini-3-pro-high")
		}
		rec = serveGeminiDeclared(t, h, key, "gemini-3-pro-high")
		if mixed {
			require.Equal(t, http.StatusOK, rec.Code)
		} else {
			require.Equal(t, http.StatusNotFound, rec.Code)
		}
		require.Zero(t, upstream.calls)
	}
}
