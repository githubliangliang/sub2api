//go:build unit

package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/gemini"
	"github.com/Wei-Shaw/sub2api/internal/pkg/geminicli"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGeminiConnectionTestUsesSupportedDefaultAndExplicitMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct{ model, want string }{{"", "gemini-3.5-flash"}, {"legacy-alias", "gemini-2.5-flash"}} {
		t.Run(tc.want, func(t *testing.T) {
			upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"ok\"}]}}]}\n\n"))}}
			account := &Account{ID: 1, Platform: PlatformGemini, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "synthetic-key", "base_url": "https://models.example.test", "model_mapping": map[string]any{"legacy-alias": "gemini-2.5-flash"}}}
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/test", nil)
			err := (&AccountTestService{httpUpstream: upstream, cfg: &config.Config{}}).testGeminiAccountConnection(c, account, tc.model, "hi")
			require.NoError(t, err)
			require.NotNil(t, upstream.lastReq)
			require.Contains(t, upstream.lastReq.URL.Path, "/models/"+tc.want+":")
		})
	}
}

func TestGeminiCuratedListsAndConservativeGoogleOneDefault(t *testing.T) {
	var native, selectable []string
	for _, model := range gemini.DefaultModels() {
		native = append(native, strings.TrimPrefix(model.Name, "models/"))
	}
	for _, model := range geminicli.DefaultModels {
		selectable = append(selectable, model.ID)
	}
	require.ElementsMatch(t, native, selectable)
	for _, retired := range []string{"gemini-2.0-flash", "gemini-2.5-flash", "gemini-2.5-pro", "gemini-2.5-flash-image", "gemini-3-pro-preview"} {
		require.NotContains(t, native, retired)
	}
	legacy := &Account{Platform: PlatformGemini, Type: AccountTypeOAuth, Credentials: map[string]any{"oauth_type": "google_one"}}
	require.Equal(t, "gemini-2.5-flash", geminiDefaultTestModel(legacy))
	require.Equal(t, map[string]string{"gemini-2.5-flash": "gemini-2.5-flash", "gemini-2.5-pro": "gemini-2.5-pro", "gemini-2.0-flash": "gemini-2.0-flash"}, legacy.GetModelMapping())
}
