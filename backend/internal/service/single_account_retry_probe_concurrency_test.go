package service

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSingleAccountRetryProbeConcurrentConsumers(t *testing.T) {
	var calls atomic.Int64
	ctx := WithSingleAccountRetryProbe(context.Background(), func(context.Context) bool {
		calls.Add(1)
		return true
	})
	var wg sync.WaitGroup
	results := make(chan bool, 32)
	for range 32 {
		wg.Add(1)
		go func() { defer wg.Done(); results <- isSingleAccountRetry(ctx) }()
	}
	wg.Wait()
	close(results)
	for value := range results {
		require.True(t, value)
	}
	require.EqualValues(t, 1, calls.Load())
}

func TestSingleAccountRetryProbeCanceledRequestUsesBoundedContext(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	cancel()
	ctx := WithSingleAccountRetryProbe(parent, func(probeCtx context.Context) bool {
		require.NoError(t, probeCtx.Err())
		deadline, ok := probeCtx.Deadline()
		require.True(t, ok)
		require.LessOrEqual(t, time.Until(deadline), 5*time.Second)
		return true
	})
	require.True(t, isSingleAccountRetry(ctx))
}
