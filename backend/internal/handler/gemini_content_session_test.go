package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGeminiNativeSessionStablePrefixAndExplicitIdentity(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-3.8-flash:generateContent", nil)
	body := []byte(`{"contents":[{"parts":[{"text":"first"}]}]}`)
	want := service.GenerateGeminiContentSessionHash(7, "gemini-3.8-flash", body)
	require.NotEmpty(t, want)
	require.Equal(t, want, geminiRequestSessionHash(c, 7, "gemini-3.8-flash", body))
	later := []byte(`{"contents":[{"role":"user","parts":[{"text":"first"}]},{"role":"model","parts":[{"functionCall":{"name":"lookup","args":{}}}]},{"parts":[{"functionResponse":{"name":"lookup","response":{}}}]}]}`)
	require.Equal(t, want, geminiRequestSessionHash(c, 7, "gemini-3.8-flash", later))
	c.Request.Header.Set("User-Agent", "different-client")
	require.Equal(t, want, geminiRequestSessionHash(c, 7, "gemini-3.8-flash", later))
	require.NotEqual(t, want, geminiRequestSessionHash(c, 8, "gemini-3.8-flash", later))
	explicit := []byte(`{"metadata":{"user_id":"user_a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2_account__session_123e4567-e89b-12d3-a456-426614174000"},"contents":[{"parts":[{"text":"first"}]}]}`)
	require.Equal(t, "123e4567-e89b-12d3-a456-426614174000", geminiRequestSessionHash(c, 7, "gemini-3.8-flash", explicit))
	require.Empty(t, geminiRequestSessionHash(c, 7, "gemini-3.8-flash", []byte(`{"contents":[{"parts":[]}]}`)))
}
