package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type lifecyclePricingClient struct {
	body            string
	hash            string
	started         chan struct{}
	release         chan struct{}
	once            sync.Once
	active          atomic.Int64
	downloads       atomic.Int64
	activeDownloads atomic.Int64
	peakDownloads   atomic.Int64
}

func (c *lifecyclePricingClient) FetchHashText(ctx context.Context, _ string) (string, error) {
	c.active.Add(1)
	defer c.active.Add(-1)
	if c.started != nil {
		c.once.Do(func() { close(c.started) })
	}
	if c.release != nil {
		select {
		case <-c.release:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return c.hash, nil
}

func (c *lifecyclePricingClient) FetchPricingJSON(ctx context.Context, _ string) ([]byte, error) {
	c.downloads.Add(1)
	active := c.activeDownloads.Add(1)
	defer c.activeDownloads.Add(-1)
	for peak := c.peakDownloads.Load(); active > peak && !c.peakDownloads.CompareAndSwap(peak, active); peak = c.peakDownloads.Load() {
	}
	select {
	case <-time.After(10 * time.Millisecond):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return []byte(c.body), nil
}

const lifecyclePricingBody = `{"fixture-model":{"input_cost_per_token":0.000001,"output_cost_per_token":0.000002}}`

func lifecyclePricingService(t *testing.T, remote PricingRemoteClient) *PricingService {
	t.Helper()
	svc := NewPricingService(&config.Config{Pricing: config.PricingConfig{DataDir: t.TempDir(), RemoteURL: "https://fixture.example/pricing.json", HashURL: "https://fixture.example/pricing.sha256", UpdateIntervalHours: 24}}, remote)
	require.NoError(t, os.WriteFile(svc.getPricingFilePath(), []byte(lifecyclePricingBody), 0600))
	return svc
}

func TestPricingLocalStartupAndStopOwnRemoteHash(t *testing.T) {
	remote := &lifecyclePricingClient{started: make(chan struct{}), release: make(chan struct{}), hash: fmt.Sprintf("%x", sha256.Sum256([]byte(lifecyclePricingBody)))}
	svc := lifecyclePricingService(t, remote)
	initDone := make(chan error, 1)
	go func() { initDone <- svc.Initialize() }()
	select {
	case err := <-initDone:
		require.NoError(t, err)
	case <-time.After(250 * time.Millisecond):
		close(remote.release)
		require.NoError(t, <-initDone)
		svc.Stop()
		t.Fatal("valid local prices must make Initialize return before the remote hash")
	}
	require.NotNil(t, svc.GetModelPricing("fixture-model"))
	select {
	case <-remote.started:
	case <-time.After(time.Second):
		t.Fatal("startup hash was not scheduled")
	}
	stopped := make(chan struct{})
	go func() { svc.Stop(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		close(remote.release)
		<-stopped
		t.Fatal("Stop must cancel the owned remote request")
	}
	require.Zero(t, remote.active.Load())
	require.Zero(t, remote.downloads.Load())
	require.NotPanics(t, svc.Stop)
}

func TestPricingStartupBodyHashAndOverrideSurviveSync(t *testing.T) {
	body := `{"fixture-model":{"input_cost_per_token":0.000003,"output_cost_per_token":0.000006}}`
	remote := &lifecyclePricingClient{body: body, hash: fmt.Sprintf("%x", sha256.Sum256([]byte(body)))}
	svc := lifecyclePricingService(t, remote)
	svc.cfg.Pricing.OverrideFile = svc.cfg.Pricing.DataDir + "/override.json"
	require.NoError(t, os.WriteFile(svc.cfg.Pricing.OverrideFile, []byte(`{"fixture-model":{"input_cost_per_token":0}}`), 0600))
	require.NoError(t, svc.Initialize())
	defer svc.Stop()
	require.Eventually(t, func() bool { svc.mu.RLock(); defer svc.mu.RUnlock(); return svc.localHash == remote.hash }, time.Second, 10*time.Millisecond)
	require.Zero(t, svc.GetModelPricing("fixture-model").InputCostPerToken)
	saved, err := os.ReadFile(svc.getPricingFilePath())
	require.NoError(t, err)
	require.Equal(t, body, string(saved))
	hash, err := os.ReadFile(svc.getHashFilePath())
	require.NoError(t, err)
	require.Equal(t, remote.hash+"\n", string(hash))
}

func TestPricingConcurrentDownloadsAreSerialized(t *testing.T) {
	remote := &lifecyclePricingClient{body: lifecyclePricingBody, hash: "stale-remote-hash"}
	svc := lifecyclePricingService(t, remote)
	var wg sync.WaitGroup
	errors := make(chan error, 6)
	for range 6 {
		wg.Add(1)
		go func() { defer wg.Done(); errors <- svc.downloadPricingData() }()
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
	require.EqualValues(t, 1, remote.peakDownloads.Load())
	want := fmt.Sprintf("%x", sha256.Sum256([]byte(lifecyclePricingBody)))
	svc.mu.RLock()
	require.Equal(t, want, svc.localHash)
	svc.mu.RUnlock()
	svc.Stop()
	require.ErrorIs(t, svc.downloadPricingData(), context.Canceled)
}
