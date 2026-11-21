package routes

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGatewayConflictingFieldsHTTPAliases(t *testing.T) {
	srv := httptest.NewServer(newGatewayRoutesTestRouter())
	defer srv.Close()
	for _, path := range []string{
		"/v1/messages", "/v1/messages/count_tokens", "/messages/count_tokens", "/v1/responses", "/responses", "/responses/compact", "/v1/responses/compact",
		"/backend-api/codex/responses", "/backend-api/codex/responses/compact", "/backend-api/codex/realtime/calls", "/backend-api/codex/alpha/search",
		"/chat/completions", "/v1/chat/completions", "/embeddings", "/v1/embeddings", "/images/generations", "/images/edits", "/v1/images/edits", "/v1/images/generations",
		"/images/generations/async", "/v1/images/generations/async", "/videos", "/v1/videos", "/videos/generations", "/v1/videos/generations", "/videos/edits", "/videos/extensions",
		"/tts", "/stt", "/custom-voices", "/web_search", "/x_search", "/v1/live", "/alpha/search", "/v1/alpha/search", "/antigravity/v1/messages",
	} {
		t.Run(path, func(t *testing.T) {
			for _, contentType := range []string{"application/json", "multipart/form-data; boundary=unused"} {
				body := "\xef\xbb\xbf" + `{"model":"gpt-6.2-sol","model":"gpt-6.2-sol","input":"fixture"}`
				resp, err := http.Post(srv.URL+path, contentType, strings.NewReader(body))
				require.NoError(t, err)
				data, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.NoError(t, resp.Body.Close())
				require.Equal(t, 400, resp.StatusCode, "%s: %s", path, data)
				require.Contains(t, string(data), "more than once")
			}
		})
	}
}
