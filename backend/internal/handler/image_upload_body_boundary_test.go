//go:build unit

package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type canceledImageBody struct{}

func (canceledImageBody) Read([]byte) (int, error) { return 0, context.Canceled }
func (canceledImageBody) Close() error             { return nil }

func TestImageHandlersPreserveTotalBodyAndCancellationBoundaries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, grok := range []bool{false, true} {
		for _, canceled := range []bool{false, true} {
			h := &OpenAIGatewayHandler{gatewayService: &service.OpenAIGatewayService{}, billingCacheService: &service.BillingCacheService{}, apiKeyService: &service.APIKeyService{}, concurrencyHelper: &ConcurrencyHelper{concurrencyService: &service.ConcurrencyService{}}}
			router := gin.New()
			router.Use(middleware.RequestBodyLimit(1024))
			router.POST("/upload", func(c *gin.Context) {
				c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{ID: 1, UserID: 1})
				c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1})
				if grok {
					h.handleGrokMedia(c, service.GrokMediaEndpointImagesEdits, "")
				} else {
					h.Images(c)
				}
			})
			req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(strings.Repeat("x", 1025)))
			if canceled {
				req.Body = io.NopCloser(canceledImageBody{})
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if canceled {
				require.Equal(t, http.StatusBadRequest, rec.Code)
				require.Contains(t, rec.Body.String(), "Failed to read request body")
			} else {
				require.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
			}
		}
	}
}
