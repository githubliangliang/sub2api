//go:build unit

package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type compatSideEffectCancelWriter struct {
	*httptest.ResponseRecorder
	cancel context.CancelFunc
}

func (w *compatSideEffectCancelWriter) Write(p []byte) (int, error) {
	n, err := w.ResponseRecorder.Write(p)
	if bytes.Contains(p, []byte("partial")) {
		w.cancel()
	}
	return n, err
}

type compatSideEffectCancelBody struct {
	io.Reader
	ctx context.Context
}

func (b *compatSideEffectCancelBody) Read(p []byte) (int, error) {
	n, err := b.Reader.Read(p)
	if err == io.EOF {
		<-b.ctx.Done()
		return 0, b.ctx.Err()
	}
	return n, err
}

func (*compatSideEffectCancelBody) Close() error { return nil }

func TestOpenAICompatClientCancellationHasNoAccountSideEffects(t *testing.T) {
	for _, protocol := range []string{"chat", "messages"} {
		t.Run(protocol, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(&compatSideEffectCancelWriter{rec, cancel})
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/"+protocol, nil).WithContext(ctx)
			prefix := "data: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_cancel\"}}\n\ndata: {\"type\":\"response.output_text.delta\",\"delta\":\"partial\"}\n\n"
			svc, repo, account := newCompatStreamSideEffectsService("")
			response := &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}}, Body: &compatSideEffectCancelBody{strings.NewReader(prefix), ctx}}
			var err error
			if protocol == "chat" {
				_, err = svc.handleChatStreamingResponse(response, c, account, "model", "model", "model", time.Now(), 0)
			} else {
				_, err = svc.handleAnthropicStreamingResponse(response, c, account, "model", "model", "model", time.Now())
			}
			require.Error(t, err)
			var failover *UpstreamFailoverError
			require.False(t, errors.As(err, &failover))
			require.Contains(t, rec.Body.String(), "partial")
			require.Empty(t, repo.errorCalls)
			require.Empty(t, repo.rateLimitCalls)
			require.Empty(t, repo.modelScopes)
			require.Zero(t, repo.tempCalls)
		})
	}
}

func TestFirstOutputMemoryFallbackEightMiBLifecycle(t *testing.T) {
	for _, mode := range []string{"commit", "drop"} {
		t.Run(mode, func(t *testing.T) {
			stage := newDefaultOpenAIFirstOutputStage()
			creates := 0
			stage.createTemp = func() (*os.File, error) { creates++; return nil, os.ErrPermission }
			chunk := bytes.Repeat([]byte("x"), 32*1024)
			for i := 0; i < openAIFirstOutputStageMaxBytes/len(chunk); i++ {
				_, err := stage.Write(chunk)
				require.NoError(t, err)
			}
			require.EqualValues(t, openAIFirstOutputStageMaxBytes, stage.Buffered())
			require.Equal(t, 1, creates)
			_, err := stage.WriteString("overflow")
			require.ErrorIs(t, err, errOpenAIFirstOutputStageLimit)
			if mode == "commit" {
				require.NoError(t, stage.CommitTo(io.Discard))
			}
			require.NoError(t, stage.Close())
			require.Zero(t, stage.Buffered())
			require.NoError(t, stage.Close())
		})
	}
}

func BenchmarkFirstOutputMemoryFallbackEightMiB(b *testing.B) {
	chunk := bytes.Repeat([]byte("x"), 32*1024)
	b.SetBytes(openAIFirstOutputStageMaxBytes)
	b.ReportAllocs()
	for b.Loop() {
		stage := newDefaultOpenAIFirstOutputStage()
		stage.createTemp = func() (*os.File, error) { return nil, os.ErrPermission }
		for i := 0; i < openAIFirstOutputStageMaxBytes/len(chunk); i++ {
			if _, err := stage.Write(chunk); err != nil {
				b.Fatal(err)
			}
		}
		if err := stage.CommitTo(io.Discard); err != nil {
			b.Fatal(err)
		}
		if err := stage.Close(); err != nil {
			b.Fatal(err)
		}
	}
}
