package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOpenAIResponses_RejectsInvalidUTF8BeforeRouting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, suffix := range []string{"", `,"stream":false`, `,"stream":true`} {
		for _, invalid := range [][]byte{{0xbb, 0xd8, 0xb8, 0xb4}, {0xe4, 0xbd}, {0xff}} {
			t.Run(fmt.Sprintf("%s/%x", suffix, invalid), func(t *testing.T) {
				body := append([]byte(`{"input":"`), invalid...)
				body = append(body, []byte(`"`+suffix+`}`)...)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
				groupID := int64(2)
				c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{ID: 101, GroupID: &groupID, User: &service.User{ID: 1}})
				c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1, Concurrency: 1})
				h := newOpenAIHandlerForPreviousResponseIDValidation(t, nil)
				h.Responses(c)
				require.Equal(t, http.StatusBadRequest, w.Code)
				require.Contains(t, w.Body.String(), "valid UTF-8 JSON")
				require.Contains(t, w.Body.String(), "invalid_request_error")
			})
		}
	}
}

func TestOpenAIResponses_ValidUnicodeReachesFieldValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(`{"input":"回复 ok 🌍"}`))
	groupID := int64(2)
	c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{ID: 101, GroupID: &groupID, User: &service.User{ID: 1}})
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1, Concurrency: 1})
	h := newOpenAIHandlerForPreviousResponseIDValidation(t, nil)
	h.Responses(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "model is required")
	require.NotContains(t, w.Body.String(), "UTF-8")
}

func TestOpenAIResponsesUTF8HTTPBoundaryAndBudget(t *testing.T) {
	gin.SetMode(gin.TestMode)
	budget := pkghttputil.NewBodyMemoryBudget(1 << 20)
	h := newOpenAIHandlerForPreviousResponseIDValidation(t, nil)
	router := gin.New()
	router.Use(middleware.RequestBodyLimit(4096, budget))
	router.Use(func(c *gin.Context) {
		if c.GetHeader("X-Fixture-Auth") == "yes" {
			groupID := int64(2)
			c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{ID: 101, GroupID: &groupID, User: &service.User{ID: 1}})
			c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1})
		}
	})
	router.POST("/v1/responses", h.Responses)
	router.POST("/v1/responses/compact", h.Responses)
	server := httptest.NewServer(router)
	defer server.Close()
	// The real handler has no forwarding dependencies: reaching scheduling would
	// fail differently. Each malformed request must exit at byte validation.
	for _, path := range []string{"/v1/responses", "/v1/responses/compact"} {
		for _, stream := range []bool{false, true} {
			valid := []byte(fmt.Sprintf(`{"model":"gpt-6.1-terra","input":"回复 🌍","stream":%t}`, stream))
			for _, position := range []int{0, len(valid) / 2, len(valid)} {
				for _, bad := range [][]byte{{0xff}, {0xc0, 0xaf}, {0xe4, 0xbd}} {
					body := append(append(append([]byte{}, valid[:position]...), bad...), valid[position:]...)
					req, err := http.NewRequest(http.MethodPost, server.URL+path, bytes.NewReader(body))
					require.NoError(t, err)
					req.Header.Set("X-Fixture-Auth", "yes")
					resp, err := server.Client().Do(req)
					require.NoError(t, err)
					raw, err := io.ReadAll(resp.Body)
					require.NoError(t, err)
					require.NoError(t, resp.Body.Close())
					require.Equal(t, http.StatusBadRequest, resp.StatusCode)
					require.Contains(t, string(raw), "valid UTF-8 JSON")
				}
			}
		}
	}
	req, err := http.NewRequest(http.MethodPost, server.URL+"/v1/responses", bytes.NewReader([]byte{0xff}))
	require.NoError(t, err)
	resp, err := server.Client().Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	// Ingress releases the full reservation even on early UTF-8 rejection.
	ctx, release := pkghttputil.WithRequestBodyMemoryBudget(t.Context(), budget)
	defer release()
	require.NoError(t, pkghttputil.ReserveRequestBodyMemory(ctx, 1<<20))
}
