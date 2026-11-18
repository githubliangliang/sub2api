package service

import (
	"context"
	"testing"
	"time"
)

func TestAccountUsageService_GetOpenAIUsage_PreservesSnapshotObservationTime(t *testing.T) {
	t.Parallel()
	observedAt := time.Now().Add(-48 * time.Hour).UTC().Truncate(time.Second)
	for _, stamp := range []string{observedAt.Format(time.RFC3339), "", "invalid"} {
		t.Run(stamp, func(t *testing.T) {
			account := &Account{
				Platform: PlatformOpenAI,
				Type:     AccountTypeOAuth,
				Extra: map[string]any{
					"codex_5h_used_percent":  10.0,
					"codex_7d_used_percent":  100.0,
					"codex_usage_updated_at": stamp,
				},
			}
			usage, err := (&AccountUsageService{}).getOpenAIUsage(context.Background(), account, false)
			if err != nil {
				t.Fatal(err)
			}
			if stamp == observedAt.Format(time.RFC3339) {
				if usage.UpdatedAt == nil || !usage.UpdatedAt.Equal(observedAt) {
					t.Fatalf("cached snapshot must keep its observation time %v, got %v", observedAt, usage.UpdatedAt)
				}
			} else if usage.UpdatedAt != nil {
				t.Fatalf("unknown observation time must not become the request time: %v", usage.UpdatedAt)
			}
			if usage.Source != "passive" || usage.Error != "" {
				t.Fatalf("snapshot read should be passive, not a failed refresh: %#v", usage)
			}
		})
	}
}

func TestAccountUsageService_GetOpenAIUsage_RefreshFailureKeepsCachedDataAndReportsError(t *testing.T) {
	t.Parallel()
	observedAt := time.Now().Add(-24 * time.Hour).UTC().Truncate(time.Second)
	for _, shadow := range []bool{false, true} {
		account := &Account{
			Platform: PlatformOpenAI,
			Type:     AccountTypeOAuth,
			Extra: map[string]any{
				"codex_7d_used_percent":  100.0,
				"codex_usage_updated_at": observedAt.Format(time.RFC3339),
			},
		}
		if shadow {
			parentID := int64(10)
			account.ParentAccountID = &parentID
			account.QuotaDimension = QuotaDimensionSpark
		}
		// Ordinary accounts have no token; shadow accounts have no quota service.
		// Neither case may turn a failed query into a fresh, successful snapshot.
		usage, err := (&AccountUsageService{}).getOpenAIUsage(context.Background(), account, true)
		if err != nil {
			t.Fatal(err)
		}
		if usage.ErrorCode != "quota_refresh_failed" || usage.Error == "" {
			t.Fatalf("refresh failure was hidden (shadow=%v): %#v", shadow, usage)
		}
		if usage.SevenDay == nil || usage.SevenDay.Utilization != 100 {
			t.Fatalf("last known quota should remain available: %#v", usage.SevenDay)
		}
		if usage.UpdatedAt == nil || !usage.UpdatedAt.Equal(observedAt) || usage.Source != "passive" {
			t.Fatalf("failed refresh must not advance snapshot freshness: %#v", usage)
		}
		if account.RateLimitResetAt != nil || account.Extra["codex_usage_updated_at"] != observedAt.Format(time.RFC3339) {
			t.Fatal("failed refresh must not mutate the snapshot or scheduler state")
		}
	}
}

func TestAccountUsageService_GetOpenAIUsage_RefreshFailureWithoutSnapshot(t *testing.T) {
	t.Parallel()
	svc := &AccountUsageService{usageLogRepo: &usageBatchLogRepoStub{}}
	usage, err := svc.getOpenAIUsage(context.Background(), &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	if usage.ErrorCode != "quota_refresh_failed" || usage.UpdatedAt != nil {
		t.Fatalf("missing upstream quota must remain unknown after failure: %#v", usage)
	}
	for name, window := range map[string]*UsageProgress{"5h": usage.FiveHour, "7d": usage.SevenDay} {
		if window == nil || window.WindowStats == nil || !window.UtilizationUnknown {
			t.Errorf("%s should keep local stats while marking quota unknown: %#v", name, window)
		}
	}
}
