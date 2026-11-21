package requestmodel

import (
	"fmt"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
)

// ConflictMessage fits the WebSocket close reason limit, including escaped
// user-provided names; ordinary protocol field names remain visible.
func ConflictMessage(field string) string {
	message := fmt.Sprintf("Request body declares field %.64q more than once", field)
	if len(message) > 120 {
		return "Request body declares a field more than once"
	}
	return message
}

func isMultipartContentType(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	return err == nil && strings.EqualFold(mediaType, "multipart/form-data")
}

func ResetRequestBody(req *http.Request, body []byte) {
	req.Body = httputil.NewPrereadBody(body)
	req.ContentLength = int64(len(body))
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))
}
