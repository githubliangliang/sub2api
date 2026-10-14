package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChannelMonitorV2HealthRequiresRequestSamples(t *testing.T) {
	for _, requestCount := range []int64{0, 8, 9, 49} {
		t.Run(fmt.Sprintf("requests_%d", requestCount), func(t *testing.T) {
			thresholds := DefaultChannelMonitorV2HealthThresholds()
			thresholds.WarningCacheRate = 0.85
			thresholds.CriticalCacheRate = 0.60
			metrics := ChannelMonitorV2Metric{SuccessRequests: requestCount, RequestCount: requestCount}
			if requestCount > 0 {
				metrics.CacheRateDenominator = 448
			}
			health := ChannelMonitorV2HealthForWithThresholds(metrics, thresholds)
			require.Equal(t, ChannelMonitorV2Health{
				Overall: "unknown", ErrorRate: "unknown", TTFT: "unknown", Cache: "unknown",
				MinimumSample: 50, Thresholds: NormalizeChannelMonitorV2HealthThresholds(thresholds),
			}, health)
		})
	}
}

func TestChannelMonitorV2HealthMinimumSampleBoundaries(t *testing.T) {
	for _, test := range []struct {
		name          string
		minimumSample int64
		wantMinimum   int64
	}{
		{name: "configured", minimumSample: 20, wantMinimum: 20},
		{name: "configured_one", minimumSample: 1, wantMinimum: 1},
		{name: "normalized_zero", minimumSample: 0, wantMinimum: 50},
		{name: "normalized_negative", minimumSample: -1, wantMinimum: 50},
		{name: "normalized_maximum", minimumSample: 10001, wantMinimum: 10000},
	} {
		t.Run(test.name, func(t *testing.T) {
			thresholds := DefaultChannelMonitorV2HealthThresholds()
			thresholds.MinimumSample = test.minimumSample
			thresholds.WarningCacheRate = 0.85
			thresholds.CriticalCacheRate = 0.60
			p50 := int64(1000)
			metrics := ChannelMonitorV2Metric{
				SuccessRequests: test.wantMinimum - 1, RequestCount: test.wantMinimum - 1,
				CacheRateDenominator: 20000,
				TTFT:                 ChannelMonitorV2Latency{SampleCount: 20000, P50Ms: &p50},
			}
			health := ChannelMonitorV2HealthForWithThresholds(metrics, thresholds)
			require.Equal(t, ChannelMonitorV2Health{
				Overall: "unknown", ErrorRate: "unknown", TTFT: "unknown", Cache: "unknown",
				MinimumSample: test.wantMinimum, Thresholds: NormalizeChannelMonitorV2HealthThresholds(thresholds),
			}, health)

			metrics.RequestCount = test.wantMinimum
			metrics.SuccessRequests = test.wantMinimum
			health = ChannelMonitorV2HealthForWithThresholds(metrics, thresholds)
			require.Equal(t, test.wantMinimum, health.MinimumSample)
			require.Equal(t, "healthy", health.ErrorRate)
			require.Equal(t, "healthy", health.TTFT)
			require.Equal(t, "critical", health.Cache)
			require.NotNil(t, health.ErrorRateScore)
			require.NotNil(t, health.TTFTScore)
			require.NotNil(t, health.CacheScore)
			require.NotNil(t, health.Score)
			require.InDelta(t, 80.0, *health.Score, 0.01)
			require.Equal(t, "healthy", health.Overall)
		})
	}
}

func TestChannelMonitorV2HealthSufficientHighErrorIsCritical(t *testing.T) {
	health := ChannelMonitorV2HealthFor(ChannelMonitorV2Metric{
		RequestCount: 50, ErrorRequests: 50, ErrorRate: 1,
	})
	require.Equal(t, "critical", health.ErrorRate)
	require.Equal(t, "critical", health.Overall)
	require.NotNil(t, health.Score)
	require.Zero(t, *health.Score)
}
