package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/requestmodel"
	"github.com/gin-gonic/gin"
)

// RejectConflictingRequestFields runs after authentication and the body budget,
// before composite model rewrites, scheduling and billing.
func RejectConflictingRequestFields() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil || c.Request.Body == http.NoBody ||
			(c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut && c.Request.Method != http.MethodPatch) {
			c.Next()
			return
		}
		body, err := httputil.ReadRequestBodyWithPrealloc(c.Request)
		if err != nil {
			if errors.Is(err, httputil.ErrRequestBodyMemoryBudget) {
				c.Header("Retry-After", "1")
				writeGatewayInvalidRequest(c, http.StatusServiceUnavailable, "Request body buffering capacity is exhausted; retry later")
			} else {
				failure := httputil.DescribeRequestBodyReadFailure(c.Request, err)
				logger.FromContext(c.Request.Context()).Warn("read request body failed", failure.LogFields()...)
				if failure.Status == httputil.StatusClientClosedRequest {
					c.Status(failure.Status)
				} else {
					writeGatewayInvalidRequest(c, failure.Status, failure.ClientMessage())
				}
			}
			c.Abort()
			return
		}
		if doc := requestmodel.TrimUTF8BOM(body); requestmodel.LooksLikeJSON(doc) {
			body = doc
		}
		requestmodel.ResetRequestBody(c.Request, body)
		if field, conflict := requestmodel.ConflictingField(c.GetHeader("Content-Type"), body); conflict {
			MarkIngressRejected(c, IngressRejectConflictingRequestFields)
			writeGatewayInvalidRequest(c, http.StatusBadRequest, requestmodel.ConflictMessage(field))
			c.Abort()
			return
		}
		c.Next()
	}
}

func writeGatewayInvalidRequest(c *gin.Context, status int, message string) {
	path := c.Request.URL.Path
	switch {
	case strings.HasPrefix(path, "/v1beta") || strings.HasPrefix(path, "/antigravity/v1beta"):
		GoogleErrorWriter(c, status, message)
	case strings.Contains(path, "/messages"):
		c.JSON(status, gin.H{"type": "error", "error": gin.H{"type": "invalid_request_error", "message": message}})
	default:
		c.JSON(status, gin.H{"error": gin.H{"type": "invalid_request_error", "message": message}})
	}
}
