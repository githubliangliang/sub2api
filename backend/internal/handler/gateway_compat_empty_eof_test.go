//go:build unit

package handler

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGatewayCompatibilityHandlers_EmptyOrPingOnlyEOFFailsWithoutBilling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, endpoint := range []string{"chat/completions", "responses"} {
		for _, tc := range []struct {
			name   string
			stream string
		}{
			{"empty", ""},
			{"ping_only", "event: ping\ndata: {\"type\":\"ping\"}\n\n"},
		} {
			t.Run(endpoint+"/"+tc.name, func(t *testing.T) {
				run := runCompatStreamUsageHandler(t, endpoint, tc.stream, true)

				require.Equal(t, 1, run.upstream.calls)
				success, failure := compatStreamUsageMarkers(endpoint)
				body := run.rec.Body.String()
				require.Equal(t, 1, strings.Count(body, failure), "EOF without a terminal event must emit one protocol failure")
				require.Contains(t, body, "stream_incomplete")
				require.NotContains(t, body, success)
				require.NotContains(t, body, `data: {"type":"error"`, "handler must not append a second fallback error")
				run.billingRepo.mu.Lock()
				defer run.billingRepo.mu.Unlock()
				require.Empty(t, run.billingRepo.commands, "no upstream usage was observed")
				require.Empty(t, run.usageRepo.created)
			})
		}
	}
}
