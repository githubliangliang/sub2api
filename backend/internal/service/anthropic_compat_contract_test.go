//go:build unit

package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAnthropicCompatErrorAfterOutputRetainsUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for name, handle := range compatStreamHandlers() {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			result, err := handle(&GatewayService{}, &http.Response{
				Header: http.Header{}, Body: io.NopCloser(strings.NewReader(anthropicStreamWithInBandError("Overloaded", ""))),
			}, c)
			require.Error(t, err)
			require.NotNil(t, result, "observed usage after visible output must remain billable on failure")
			require.Equal(t, 20, result.Usage.InputTokens)
			require.Equal(t, 1, result.Usage.OutputTokens, "do not estimate tokens absent from upstream usage")
			require.Equal(t, 1, strings.Count(rec.Body.String(), compatStreamFailureMarkers[name]))
			require.NotContains(t, rec.Body.String(), compatStreamTerminalMarkers[name])
		})
	}
}

func TestAnthropicCompatErrorBeforeSemanticOutputDiscardsPreamble(t *testing.T) {
	gin.SetMode(gin.TestMode)
	start := anthropicStreamWithMidStreamText()
	prefix := "event: ping\ndata: {\"type\":\"ping\"}\n\n" + start[:strings.Index(start, "event: content_block_start")]
	stream := prefix + "event: error\ndata: {\"type\":\"error\",\"error\":{\"type\":\"overloaded_error\",\"message\":\"Overloaded\"}}\n\n"
	for name, handle := range compatStreamHandlers() {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			result, err := handle(&GatewayService{}, &http.Response{Header: http.Header{}, Body: io.NopCloser(strings.NewReader(stream))}, c)
			var sseErr *sseStreamErrorEventError
			require.ErrorAs(t, err, &sseErr)
			require.Nil(t, result, "retryable pre-output attempts must not be settled twice")
			want := ""
			if name == "chat_completions" {
				want = ": ping\n\n"
			}
			require.Equal(t, want, rec.Body.String(), "transport comments may pass; response metadata must remain private until semantic output")
			require.False(t, IsResponseCommitted(c))
		})
	}
}

func TestAnthropicCompatActiveDrainHasTotalBound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	original := anthropicCompatDisconnectDrainInterval
	anthropicCompatDisconnectDrainInterval = 100 * time.Millisecond
	defer func() { anthropicCompatDisconnectDrainInterval = original }()
	for name, handle := range compatStreamHandlers() {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil).WithContext(ctx)
			reader, writer := io.Pipe()
			defer reader.Close()
			defer writer.Close()
			writerDone := make(chan struct{})
			go func() {
				defer close(writerDone)
				for {
					if _, err := io.WriteString(writer, "event: ping\ndata: {\"type\":\"ping\"}\n\n"); err != nil {
						return
					}
					time.Sleep(10 * time.Millisecond)
				}
			}()
			done := make(chan error, 1)
			go func() {
				_, err := handle(&GatewayService{}, &http.Response{Header: http.Header{}, Body: reader}, c)
				done <- err
			}()
			select {
			case <-done:
			case <-time.After(700 * time.Millisecond):
				_ = reader.CloseWithError(errors.New("test drain deadline"))
				<-done
				t.Error("active upstream traffic must not reset the total disconnect-drain bound")
			}
			select {
			case <-writerDone:
			case <-time.After(time.Second):
				t.Error("upstream reader/writer did not stop after drain bound")
			}
		})
	}
}

func TestAnthropicCompatStopReleasesBlockedReader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	stream := newAnthropicCompatStream(nil, c, &http.Response{Body: reader}, "test", "")
	stream.stop()
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	for {
		select {
		case _, ok := <-stream.pump.events:
			if !ok {
				return
			}
		case <-timer.C:
			t.Fatal("stream stop left an upstream reader blocked")
		}
	}
}
