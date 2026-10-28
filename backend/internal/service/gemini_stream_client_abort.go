package service

import (
	"context"
	"errors"
	"io"

	"github.com/gin-gonic/gin"
)

func geminiStreamClientGone(c *gin.Context) bool {
	return c != nil && c.Request != nil && c.Request.Context().Err() != nil
}

func geminiStreamHasObservedUsage(c *gin.Context, usage *ClaudeUsage) bool {
	return usage.hasObservedTokens() || observedGeminiStreamImages(c) > 0
}

// An upstream context.Canceled alone is not evidence of a client disconnect.
func geminiStreamAbortedWithBillableOutput(c *gin.Context, err error, usage *ClaudeUsage) bool {
	return err != nil && !errors.Is(err, io.EOF) && geminiStreamClientGone(c) && geminiStreamHasObservedUsage(c, usage)
}

func geminiStreamAbortError(c *gin.Context, writeErr error) error {
	if writeErr != nil {
		return writeErr
	}
	if geminiStreamClientGone(c) {
		return c.Request.Context().Err()
	}
	return context.Canceled
}

// Preserve the first confirmed write error and stop all later writes/flushes.
type geminiStreamClientWriter struct {
	c   *gin.Context
	err error
}

func (w *geminiStreamClientWriter) Write(data []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	if geminiStreamClientGone(w.c) {
		w.err = w.c.Request.Context().Err()
		return 0, w.err
	}
	n, err := w.c.Writer.Write(data)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	w.err = err
	return n, err
}

func (w *geminiStreamClientWriter) Flush() {
	if w.err == nil && !geminiStreamClientGone(w.c) {
		w.c.Writer.Flush()
	}
}

// Keep interrupted-stream observation separate from the legacy completed
// request image fallback. Only actual inline images or file references count.
func observeGeminiStreamImages(c *gin.Context, payload []byte) {
	observeGeminiImageOutputs(c, payload)
	counter := geminiImageOutputCounterFromContext(c)
	if counter == nil {
		return
	}
	counter.streamCount = max(counter.streamCount, countGeminiImageOutputs(payload))
}

func observedGeminiStreamImages(c *gin.Context) int {
	if counter := geminiImageOutputCounterFromContext(c); counter != nil {
		return counter.streamCount
	}
	return 0
}
