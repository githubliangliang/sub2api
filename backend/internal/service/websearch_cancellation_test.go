//go:build unit

package service

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/websearch"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// Cancellation can become observable even when the writer reports a complete
// successful write. Check every event boundary, including events inside helpers.
type searchCancelBoundaryWriter struct {
	*httptest.ResponseRecorder
	cancel     context.CancelFunc
	onWrite    int
	onFlush    int
	writes     int
	flushes    int
	writeAfter int
	flushAfter int
	ctx        context.Context
}

func (w *searchCancelBoundaryWriter) Write(body []byte) (int, error) {
	if w.ctx.Err() != nil {
		w.writeAfter++
	}
	w.writes++
	n, err := w.ResponseRecorder.Write(body)
	if w.onWrite == w.writes {
		w.cancel()
	}
	return n, err
}

func (w *searchCancelBoundaryWriter) Flush() {
	if w.ctx.Err() != nil {
		w.flushAfter++
	}
	w.flushes++
	w.ResponseRecorder.Flush()
	if w.onFlush == w.flushes {
		w.cancel()
	}
}

func TestWebSearchStream_CancellationAtEveryEventBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, boundary := range []string{"write", "flush"} {
		for event := 1; event <= 10; event++ {
			t.Run(fmt.Sprintf("%s_event_%d", boundary, event), func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				t.Cleanup(cancel)
				rec := httptest.NewRecorder()
				writer := &searchCancelBoundaryWriter{ResponseRecorder: rec, cancel: cancel, ctx: ctx}
				if boundary == "write" {
					writer.onWrite = event
				} else {
					writer.onFlush = event
				}
				c, _ := gin.CreateTestContext(writer)
				c.Request = httptest.NewRequest("POST", "/v1/messages", nil).WithContext(ctx)
				result, err := writeWebSearchStreamResponse(c, "query", &websearch.SearchResponse{}, "model", time.Now())
				require.ErrorIs(t, err, context.Canceled)
				require.Nil(t, result)
				require.Zero(t, writer.writeAfter, "no event may follow an observed cancellation")
				require.Zero(t, writer.flushAfter, "no flush may follow an observed cancellation")
				if event < 10 {
					require.NotContains(t, rec.Body.String(), "event: message_stop\n")
				}
			})
		}
	}
}

func TestWebSearchJSON_CancellationDuringWriteDoesNotReportSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	writer := &searchCancelBoundaryWriter{ResponseRecorder: httptest.NewRecorder(), cancel: cancel, ctx: ctx, onWrite: 1}
	c, _ := gin.CreateTestContext(writer)
	c.Request = httptest.NewRequest("POST", "/v1/messages", nil).WithContext(ctx)
	result, err := writeWebSearchNonStreamResponse(c, "query", &websearch.SearchResponse{}, "model", time.Now())
	require.ErrorIs(t, err, context.Canceled)
	require.Nil(t, result)
	require.Equal(t, 1, strings.Count(writer.Body.String(), `"type":"message"`))
}
