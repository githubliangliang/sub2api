package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// A transport comment may have committed HTTP 200 while no semantic output was
// delivered. Retry exhaustion must finish that SSE stream in its own protocol.
func writeCompatibilityFailoverAfterComment(c *gin.Context, failure *service.UpstreamFailoverError, responses bool) bool {
	if failure == nil || !failure.SafeToFailoverAfterWrite || !c.Writer.Written() {
		return false
	}
	message := "All available accounts exhausted"
	payload := gin.H{"error": gin.H{"type": "upstream_error", "message": message}}
	prefix := ""
	if responses {
		prefix = "event: response.failed\n"
		payload = gin.H{
			"type": "response.failed", "sequence_number": 0,
			"response": gin.H{
				"id": "resp_" + strings.ReplaceAll(uuid.NewString(), "-", ""), "object": "response", "created_at": time.Now().Unix(),
				"status": "failed", "output": []any{}, "error": gin.H{"code": "upstream_error", "message": message},
			},
		}
	}
	encoded, _ := json.Marshal(payload)
	service.MarkOpsStreamError(c, "upstream_error", message, http.StatusBadGateway)
	service.MarkResponseCommitted(c)
	if _, err := io.WriteString(c.Writer, prefix+"data: "+string(encoded)+"\n\n"); err == nil {
		c.Writer.Flush()
	}
	return true
}
