package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGatewayCodexGeneratedIdentityHTTP(t *testing.T) {
	for _, tc := range []struct{ public, target string }{
		{"gpt-5.5", "grok-4.5"},
		{"operator-alias", "gpt-5.5"},
		{"vendor/gpt-oss-120b", "vendor/gpt-oss-120b"},
	} {
		t.Run(tc.public, func(t *testing.T) {
			const groupID int64 = 707
			h := newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{byGroup: map[int64][]service.Account{groupID: {{ID: 1, Platform: service.PlatformOpenAI, Credentials: map[string]any{"model_mapping": map[string]any{tc.public: tc.target}}}}}})
			router := gin.New()
			router.GET("/models", func(c *gin.Context) {
				c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{Group: &service.Group{ID: groupID, Platform: service.PlatformComposite}})
				h.CodexModels(c)
			})
			server := httptest.NewServer(router)
			defer server.Close()
			res, err := server.Client().Get(server.URL + "/models?client_version=0.147.0")
			require.NoError(t, err)
			defer func() { require.NoError(t, res.Body.Close()) }()
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode, string(body))
			require.Equal(t, service.CodexModelsManifestETag(body), res.Header.Get("ETag"))
			var got struct {
				Models []struct {
					Slug          string `json:"slug"`
					ModelMessages struct {
						Instructions string `json:"instructions_template"`
					} `json:"model_messages"`
				} `json:"models"`
			}
			require.NoError(t, json.Unmarshal(body, &got))
			require.Len(t, got.Models, 1)
			require.Equal(t, tc.public, got.Models[0].Slug)
			base := openai.CodexBaseInstructionsForModel(tc.target)
			if tc.target == "gpt-5.5" {
				require.Equal(t, base, got.Models[0].ModelMessages.Instructions)
			} else {
				require.Equal(t, strings.Replace(base, " based on GPT-5", "", 1), got.Models[0].ModelMessages.Instructions)
			}
		})
	}
}
