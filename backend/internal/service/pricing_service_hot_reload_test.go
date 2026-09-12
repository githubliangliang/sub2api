package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

const hotReloadCatalog = `{"catalog-model":{"input_cost_per_token":1,"output_cost_per_token":2,"litellm_provider":"test"}}`

func newScheduledPricingTestService(t *testing.T, interval int) *PricingService {
	t.Helper()
	dir := t.TempDir()
	svc := NewPricingService(&config.Config{Pricing: config.PricingConfig{
		DataDir: dir, FallbackFile: filepath.Join(dir, "fallback.json"), OverrideFile: filepath.Join(dir, "override.json"), HashCheckIntervalMinutes: interval,
	}}, nil)
	require.NoError(t, os.WriteFile(svc.getPricingFilePath(), []byte(hotReloadCatalog), 0600))
	require.NoError(t, os.WriteFile(svc.cfg.Pricing.FallbackFile, []byte(`{"local-model":{"input_cost_per_token":4,"output_cost_per_token":8}}`), 0600))
	require.NoError(t, os.WriteFile(svc.cfg.Pricing.OverrideFile, []byte(`{"catalog-model":{"input_cost_per_token":7,"output_cost_per_token":14}}`), 0600))
	require.NoError(t, svc.loadPricingData(svc.getPricingFilePath()))
	svc.startUpdateScheduler()
	t.Cleanup(svc.Stop)
	synctest.Wait()
	return svc
}

func TestPricingCustomFilesSchedulerUsesConfiguredInterval(t *testing.T) {
	for _, tc := range []struct {
		minutes  int
		interval time.Duration
	}{{1, time.Minute}, {3, 3 * time.Minute}, {0, 10 * time.Minute}} {
		t.Run(fmt.Sprint(tc.minutes), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				svc := newScheduledPricingTestService(t, tc.minutes)
				anchor, updated := svc.localHash, svc.lastUpdated
				require.NoError(t, os.WriteFile(svc.cfg.Pricing.FallbackFile, []byte(`{"local-model":{"input_cost_per_token":5,"output_cost_per_token":10}}`), 0600))
				time.Sleep(tc.interval - time.Second)
				synctest.Wait()
				require.Equal(t, float64(4), svc.GetModelPricing("local-model").InputCostPerToken)
				time.Sleep(time.Second)
				synctest.Wait()
				require.Equal(t, float64(5), svc.GetModelPricing("local-model").InputCostPerToken)
				require.Equal(t, anchor, svc.localHash)
				require.Equal(t, updated, svc.lastUpdated)
			})
		})
	}
}

func TestPricingCustomFilesSchedulerInvalidDeletedAndRecoveredLayers(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		svc := newScheduledPricingTestService(t, 1)
		tick := func() { time.Sleep(time.Minute); synctest.Wait() }
		// A bad override must prevent a valid fallback edit from partially publishing.
		for _, invalid := range []string{`{unfinished`, `[]`, `null`} {
			require.NoError(t, os.WriteFile(svc.cfg.Pricing.OverrideFile, []byte(invalid), 0600))
			require.NoError(t, os.WriteFile(svc.cfg.Pricing.FallbackFile, []byte(`{"local-model":{"input_cost_per_token":6,"output_cost_per_token":12}}`), 0600))
			tick()
			require.Equal(t, float64(7), svc.GetModelPricing("catalog-model").InputCostPerToken)
			require.Equal(t, float64(4), svc.GetModelPricing("local-model").InputCostPerToken)
		}
		require.NoError(t, os.WriteFile(svc.cfg.Pricing.OverrideFile, []byte(`{"catalog-model":{"input_cost_per_token":9}}`), 0600))
		tick()
		require.Equal(t, float64(9), svc.GetModelPricing("catalog-model").InputCostPerToken)
		require.Equal(t, float64(2), svc.GetModelPricing("catalog-model").OutputCostPerToken)
		require.Equal(t, float64(6), svc.GetModelPricing("local-model").InputCostPerToken)
		require.NoError(t, os.Remove(svc.cfg.Pricing.OverrideFile))
		tick()
		require.Equal(t, float64(1), svc.GetModelPricing("catalog-model").InputCostPerToken)
		require.NoError(t, os.Remove(svc.cfg.Pricing.FallbackFile))
		tick()
		require.Nil(t, svc.GetModelPricing("local-model"))
		require.NoError(t, os.WriteFile(svc.cfg.Pricing.FallbackFile, []byte(`{"local-model":{"input_cost_per_token":8,"output_cost_per_token":16}}`), 0600))
		tick()
		require.Equal(t, float64(8), svc.GetModelPricing("local-model").InputCostPerToken)
	})
}

func TestPricingCustomFilesConcurrentReadersSeeWholeSnapshots(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		svc := newScheduledPricingTestService(t, 1)
		original := svc.GetModelPricing("catalog-model")
		for version := 10; version < 30; version++ {
			body := fmt.Sprintf(`{"catalog-model":{"input_cost_per_token":%d,"output_cost_per_token":%d},"local-model":{"input_cost_per_token":%d,"output_cost_per_token":%d}}`, version, 2*version, version, 2*version)
			tmp := svc.cfg.Pricing.OverrideFile + ".tmp"
			require.NoError(t, os.WriteFile(tmp, []byte(body), 0600))
			require.NoError(t, os.Rename(tmp, svc.cfg.Pricing.OverrideFile))
			var readers sync.WaitGroup
			for i := 0; i < 8; i++ {
				readers.Go(func() {
					time.Sleep(time.Minute)
					// One acquired snapshot must not expose only part of an override.
					svc.mu.RLock()
					catalog, local := *svc.pricingData["catalog-model"], *svc.pricingData["local-model"]
					svc.mu.RUnlock()
					require.Equal(t, 2*catalog.InputCostPerToken, catalog.OutputCostPerToken)
					require.Equal(t, 2*local.InputCostPerToken, local.OutputCostPerToken)
					if catalog.InputCostPerToken >= 10 {
						require.Equal(t, catalog.InputCostPerToken, local.InputCostPerToken)
					}
				})
			}
			readers.Wait()
			synctest.Wait()
			require.Equal(t, float64(version), svc.GetModelPricing("catalog-model").InputCostPerToken)
			require.Equal(t, float64(7), original.InputCostPerToken, "published prices must remain immutable")
		}
	})
}
