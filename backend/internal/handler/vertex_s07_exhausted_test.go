//go:build unit

package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestVertexCredentialFailoverExhaustedUsesAuthenticationResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name   string
		handle func(*GatewayHandler, *gin.Context, *service.UpstreamFailoverError)
	}{
		{"native", func(h *GatewayHandler, c *gin.Context, failure *service.UpstreamFailoverError) {
			h.handleGeminiFailoverExhausted(c, failure)
		}},
		{"messages", func(h *GatewayHandler, c *gin.Context, failure *service.UpstreamFailoverError) {
			h.handleFailoverExhausted(c, failure, service.PlatformGemini, false)
		}},
		{"chat_completions", func(h *GatewayHandler, c *gin.Context, failure *service.UpstreamFailoverError) {
			h.handleCCFailoverExhausted(c, failure, false)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			failure := &service.UpstreamFailoverError{
				StatusCode:   http.StatusUnauthorized,
				Stage:        service.GatewayFailureStageAccountAuth,
				Scope:        service.GatewayFailureScopeAccount,
				Reason:       service.GatewayFailureReason("vertex_service_account_not_found"),
				ResponseBody: []byte(`{"error":{"code":401,"message":"Upstream authentication failed","status":"UNAUTHENTICATED"}}`),
			}

			tt.handle(&GatewayHandler{}, c, failure)

			require.Equal(t, http.StatusBadGateway, recorder.Code)
			require.Contains(t, recorder.Body.String(), "Upstream authentication failed, please contact administrator")
			require.NotContains(t, recorder.Body.String(), "Grok")
			require.NotContains(t, recorder.Body.String(), "account not found")
		})
	}
}
