//go:build embed

package web

import (
	"context"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type nonceHTTPSettings struct {
	sync.Mutex
	value any
	err   error
}

func (s *nonceHTTPSettings) GetPublicSettingsForInjection(context.Context) (any, error) {
	s.Lock()
	defer s.Unlock()
	return s.value, s.err
}

func TestFrontendHTMLRealHTTPNonceSettingsAndAssets(t *testing.T) {
	provider := &nonceHTTPSettings{value: map[string]string{"site_name": "initial-023-title"}}
	frontend, err := NewFrontendServer(provider)
	require.NoError(t, err)
	router := gin.New()
	router.Use(middleware.SecurityHeaders(config.CSPConfig{Enabled: true}, nil), frontend.Middleware())
	server := httptest.NewServer(router)
	defer server.Close()
	read := func(path, etag string) (*http.Response, string) {
		t.Helper()
		req, err := http.NewRequest(http.MethodGet, server.URL+path, nil)
		require.NoError(t, err)
		if etag != "" {
			req.Header.Set("If-None-Match", etag)
		}
		resp, err := server.Client().Do(req)
		require.NoError(t, err)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
		return resp, string(body)
	}
	first, firstBody := read("/", "")
	require.Equal(t, http.StatusOK, first.StatusCode)
	assertNoStoreHTMLHeaders(t, first.Header)
	noncePattern := regexp.MustCompile(`'nonce-([^']+)'`)
	firstNonce := noncePattern.FindStringSubmatch(first.Header.Get("Content-Security-Policy"))
	require.Len(t, firstNonce, 2)
	require.Contains(t, firstBody, `nonce="`+firstNonce[1]+`"`)
	require.Contains(t, firstBody, "initial-023-title")
	second, secondBody := read("/", first.Header.Get("ETag"))
	require.Equal(t, http.StatusOK, second.StatusCode)
	assertNoStoreHTMLHeaders(t, second.Header)
	secondNonce := noncePattern.FindStringSubmatch(second.Header.Get("Content-Security-Policy"))
	require.Len(t, secondNonce, 2)
	require.NotEqual(t, firstNonce[1], secondNonce[1])
	require.Contains(t, secondBody, `nonce="`+secondNonce[1]+`"`)
	require.NotContains(t, secondBody, firstNonce[1])
	provider.Lock()
	provider.value = map[string]string{"site_name": "changed-023-title"}
	provider.Unlock()
	frontend.InvalidateCache()
	third, thirdBody := read("/", first.Header.Get("ETag"))
	require.Equal(t, http.StatusOK, third.StatusCode)
	require.NotEqual(t, first.Header.Get("ETag"), third.Header.Get("ETag"))
	require.Contains(t, thirdBody, "changed-023-title")
	require.NotContains(t, thirdBody, "initial-023-title")
	provider.Lock()
	provider.err = context.DeadlineExceeded
	provider.Unlock()
	frontend.InvalidateCache()
	fallback, fallbackBody := read("/", third.Header.Get("ETag"))
	require.Equal(t, http.StatusOK, fallback.StatusCode)
	assertNoStoreHTMLHeaders(t, fallback.Header)
	require.NotContains(t, fallbackBody, "changed-023-title")
	require.NotContains(t, fallbackBody, NonceHTMLPlaceholder)
	assets, err := fs.ReadDir(frontend.distFS, "assets")
	require.NoError(t, err)
	for _, asset := range assets {
		if asset.IsDir() || !strings.HasSuffix(asset.Name(), ".js") {
			continue
		}
		response, _ := read("/assets/"+asset.Name(), "")
		require.Equal(t, http.StatusOK, response.StatusCode)
		require.Contains(t, response.Header.Get("Cache-Control"), "immutable")
		require.NotContains(t, response.Header.Get("Cache-Control"), "no-store")
		return
	}
	t.Fatal("the embedded build contains no hashed JavaScript asset")
}
