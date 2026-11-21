package middleware

import (
	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequestBodyLimit 使用 MaxBytesReader 限制请求体大小。
func RequestBodyLimit(maxBytes int64, budgets ...*pkghttputil.BodyMemoryBudget) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(budgets) > 0 {
			ctx, release := pkghttputil.WithRequestBodyMemoryBudget(c.Request.Context(), budgets[0])
			defer release()
			c.Request = c.Request.WithContext(ctx)
		}
		if body, ok := c.Request.Body.(*pkghttputil.PrereadBody); ok {
			if int64(len(body.Bytes())) > maxBytes {
				writeGatewayInvalidRequest(c, http.StatusRequestEntityTooLarge, pkghttputil.BodyTooLargeMessage(maxBytes))
				c.Abort()
				return
			}
			c.Next()
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
