package handler

import (
	"errors"
	"fmt"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/config"
	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
)

func extractMaxBytesError(err error) (*http.MaxBytesError, bool) {
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		return maxErr, true
	}
	return nil, false
}

func formatBodyLimit(limit int64) string {
	const mb = 1024 * 1024
	if limit >= mb {
		return fmt.Sprintf("%dMB", limit/mb)
	}
	return fmt.Sprintf("%dB", limit)
}

func buildBodyTooLargeMessage(limit int64) string {
	return fmt.Sprintf("Request body too large, limit is %s", formatBodyLimit(limit))
}

func readLenientJSONRequestBodyWithPrealloc(req *http.Request, cfg *config.Config) ([]byte, error) {
	return pkghttputil.ReadLenientJSONRequestBodyWithPrealloc(req, gatewayMaxBodySize(cfg))
}

func gatewayMaxBodySize(cfg *config.Config) int64 {
	if cfg == nil {
		return 0
	}
	return cfg.Gateway.MaxBodySize
}

func requestBodyReadStatus(c *gin.Context, err error) int {
	if errors.Is(err, pkghttputil.ErrRequestBodyMemoryBudget) {
		c.Header("Retry-After", "1")
		return http.StatusServiceUnavailable
	}
	failure := pkghttputil.DescribeRequestBodyReadFailure(c.Request, err)
	logger.FromContext(c.Request.Context()).Warn("read request body failed", failure.LogFields()...)
	return failure.Status
}
func requestBodyReadMessage(c *gin.Context, err error) string {
	if errors.Is(err, pkghttputil.ErrRequestBodyMemoryBudget) {
		return "Request body buffering capacity is exhausted; retry later"
	}
	return pkghttputil.DescribeRequestBodyReadFailure(c.Request, err).Message
}

func handleRequestBodyReadError(c *gin.Context, err error) bool {
	if handleRequestBodyMemoryOverload(c, err) {
		return true
	}
	failure := pkghttputil.DescribeRequestBodyReadFailure(c.Request, err)
	if failure.Status == 499 {
		logger.FromContext(c.Request.Context()).Warn("read request body failed", failure.LogFields()...)
		if !c.Writer.Written() {
			c.Status(499)
		}
		c.Abort()
		return true
	}
	return false
}

func handleRequestBodyMemoryOverload(c *gin.Context, err error) bool {
	if !errors.Is(err, pkghttputil.ErrRequestBodyMemoryBudget) {
		return false
	}
	c.Header("Retry-After", "1")
	if !c.Writer.Written() {
		c.AbortWithStatusJSON(503, gin.H{"error": gin.H{"type": "server_error", "message": "Request body buffering capacity is exhausted; retry later"}})
	} else {
		c.Abort()
	}
	return true
}
