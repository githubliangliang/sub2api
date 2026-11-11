package handler

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAsyncImageBodyBudgetSurvivesIngressRelease(t *testing.T) {
	budget := httputil.NewBodyMemoryBudget(4096)
	ctx, release := httputil.WithRequestBodyMemoryBudget(context.Background(), budget)
	body := bytes.Repeat([]byte("x"), 2048)
	require.NoError(t, httputil.ReserveRequestBodyMemory(ctx, int64(cap(body))))
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations/async", nil).WithContext(ctx)
	task, _, cancel, err := newAsyncImageContext(c, body, time.Minute)
	require.NoError(t, err)
	defer cancel()
	release()
	got, err := httputil.ReadRequestBodyWithPrealloc(task.Request)
	require.NoError(t, err)
	require.Equal(t, body, got)
	probe, done := httputil.WithRequestBodyMemoryBudget(context.Background(), budget)
	defer done()
	require.ErrorIs(t, httputil.ReserveRequestBodyMemory(probe, 4096), httputil.ErrRequestBodyMemoryBudget)
	cancel()
	require.NoError(t, httputil.ReserveRequestBodyMemory(probe, 4096))
}

func TestRequestBodyOverloadHTTPReturns503AndRecovers(t *testing.T) {
	budget := httputil.NewBodyMemoryBudget(2048)
	ctx, release := httputil.WithRequestBodyMemoryBudget(context.Background(), budget)
	require.NoError(t, httputil.ReserveRequestBodyMemory(ctx, 2048))
	router := gin.New()
	router.Use(middleware.RequestBodyLimit(8192, budget))
	router.POST("/v1/responses", func(c *gin.Context) {
		body, err := httputil.ReadRequestBodyWithPrealloc(c.Request)
		if err != nil {
			if !handleRequestBodyReadError(c, err) {
				c.AbortWithStatus(requestBodyReadStatus(c, err))
			}
			return
		}
		c.Data(200, "application/json", body)
	})
	server := httptest.NewServer(router)
	defer server.Close()
	request := func() (*http.Response, []byte) {
		response, err := server.Client().Post(server.URL+"/v1/responses", "application/json", bytes.NewBufferString(`{"model":"fixture"}`))
		require.NoError(t, err)
		data, err := io.ReadAll(response.Body)
		require.NoError(t, err)
		_ = response.Body.Close()
		return response, data
	}
	response, _ := request()
	require.Equal(t, 503, response.StatusCode)
	require.Equal(t, "1", response.Header.Get("Retry-After"))
	release()
	response, body := request()
	require.Equal(t, 200, response.StatusCode)
	require.JSONEq(t, `{"model":"fixture"}`, string(body))
}
