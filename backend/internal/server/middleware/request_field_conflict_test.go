package middleware

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRequestFieldConflictsFormatsAndReread(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, path := range []string{"/v1/responses", "/v1/messages", "/v1beta/models/gemini:generateContent", "/antigravity/v1beta/models/gemini:generateContent"} {
		t.Run(path, func(t *testing.T) {
			called := 0
			r := gin.New()
			r.Use(RequestBodyLimit(1<<20), RejectConflictingRequestFields())
			r.POST(path, func(c *gin.Context) {
				called++
				first, err := io.ReadAll(c.Request.Body)
				require.NoError(t, err)
				second, err := httputil.ReadRequestBodyWithPrealloc(c.Request)
				require.NoError(t, err)
				require.Equal(t, first, second)
				c.Data(200, "application/json", first)
			})
			for _, body := range []string{`{"model":"a","Model":"a"}`, `{"reasoning":{"effort":"none","\u0065ffort":"max"}}`} {
				w := httptest.NewRecorder()
				r.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader(body)))
				require.Equal(t, 400, w.Code)
				if strings.Contains(path, "v1beta") {
					require.Contains(t, w.Body.String(), "INVALID_ARGUMENT")
				} else {
					require.Contains(t, w.Body.String(), "invalid_request_error")
				}
			}
			require.Zero(t, called)
			body := `{"model":"fixture","input":[{"type":"function_call","arguments":{"id":1,"ID":2}}]}`
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader("\xef\xbb\xbf"+body)))
			require.Equal(t, 200, w.Code)
			require.Equal(t, body, w.Body.String())
			require.Equal(t, 1, called)
		})
	}
}

func TestConflictPrereadRetainsNarrowerTextLimitAndStorage(t *testing.T) {
	const body = `{"model":"fixture","input":"text"}`
	for _, limit := range []int64{8, 1024} {
		r := gin.New()
		var preread *httputil.PrereadBody
		r.Use(RequestBodyLimit(1024), RejectConflictingRequestFields(), func(c *gin.Context) {
			var ok bool
			preread, ok = c.Request.Body.(*httputil.PrereadBody)
			require.True(t, ok)
			c.Next()
		}, RequestBodyLimit(limit))
		r.POST("/responses", func(c *gin.Context) {
			require.Same(t, preread, c.Request.Body, "the route limit must preserve validated body storage")
			c.Status(204)
		})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("POST", "/responses", strings.NewReader(body)))
		if limit < int64(len(body)) {
			require.Equal(t, 413, w.Code)
		} else {
			require.Equal(t, 204, w.Code)
		}
	}
}

func TestConflictIngressConcurrentBudgetHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const budgetBytes = 64 << 20
	budget := httputil.NewBodyMemoryBudget(budgetBytes)
	r := gin.New()
	r.Use(RequestBodyLimit(8<<20, budget), RejectConflictingRequestFields())
	release := make(chan struct{})
	r.POST("/responses", func(c *gin.Context) { <-release; c.Status(204) })
	srv := httptest.NewServer(r)
	defer srv.Close()
	body := []byte(`{"model":"fixture","input":"` + strings.Repeat("x", 2<<20) + `"}`)
	var accepted, rejected atomic.Int32
	var wg sync.WaitGroup
	overloaded := make(chan struct{}, 64)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client := &http.Client{Timeout: 20 * time.Second}
	defer client.CloseIdleConnections()
	for range 64 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, err := http.NewRequestWithContext(ctx, "POST", srv.URL+"/responses", bytes.NewReader(body))
			if err != nil {
				t.Error(err)
				return
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Error(err)
				return
			}
			defer func() { require.NoError(t, resp.Body.Close()) }()
			switch resp.StatusCode {
			case 204:
				accepted.Add(1)
			case 503:
				rejected.Add(1)
				overloaded <- struct{}{}
			default:
				t.Errorf("unexpected status %d", resp.StatusCode)
			}
		}()
	}
	select {
	case <-overloaded:
	case <-time.After(10 * time.Second):
		t.Error("concurrent requests did not hit the shared budget")
	}
	close(release)
	wg.Wait()
	require.Positive(t, accepted.Load())
	require.Positive(t, rejected.Load())
	// Every request returned: a new reservation can claim the complete budget.
	checkCtx, free := httputil.WithRequestBodyMemoryBudget(context.Background(), budget)
	defer free()
	require.NoError(t, httputil.ReserveRequestBodyMemory(checkCtx, budgetBytes))
	t.Logf("actual loopback HTTP: accepted=%d overloaded=%d shared budget=%d bytes", accepted.Load(), rejected.Load(), budgetBytes)
}

func TestRequestFieldConflictsMultipartAndBudget(t *testing.T) {
	for _, duplicate := range []bool{false, true} {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		require.NoError(t, writer.WriteField("model", "fixture"))
		if duplicate {
			require.NoError(t, writer.WriteField("model", "fixture"))
		}
		require.NoError(t, writer.WriteField("image", "one"))
		require.NoError(t, writer.WriteField("image", "two"))
		require.NoError(t, writer.Close())
		r := gin.New()
		r.Use(RequestBodyLimit(1<<20, httputil.NewBodyMemoryBudget(1<<20)), RejectConflictingRequestFields())
		r.POST("/images/edits", func(c *gin.Context) { c.Status(204) })
		req := httptest.NewRequest(http.MethodPost, "/images/edits", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if duplicate {
			require.Equal(t, 400, w.Code)
		} else {
			require.Equal(t, 204, w.Code)
		}
	}
}
