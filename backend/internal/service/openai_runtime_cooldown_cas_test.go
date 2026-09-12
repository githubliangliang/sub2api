package service

import (
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestRuntimeBlockHonorsClearedPersistedCooldown(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := &Account{ID: 92, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true}
	svc.BlockAccountScheduling(account, time.Now().Add(30*time.Minute), "grok payment required")
	require.False(t, svc.isOpenAIAccountRequestRuntimeBlocked(account, "grok-3"))
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
}

func TestRuntimeBlockConditionalClearSkipsNewerGeneration(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := &Account{ID: 94, Platform: PlatformGrok, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true}
	firstUntil := time.Now().Add(10 * time.Minute)
	svc.BlockAccountScheduling(account, firstUntil, "stale")
	snapshot := svc.peekOpenAIAccountRuntimeBlock(account)
	require.True(t, snapshot.blocked)
	newerUntil := time.Now().Add(30 * time.Minute)
	svc.BlockAccountScheduling(account, newerUntil, "fresh")
	svc.clearOpenAIAccountRuntimeBlockIfUnchanged(account.ID, snapshot)
	require.True(t, svc.isOpenAIAccountRuntimeBlocked(account))
	require.False(t, svc.isOpenAIAccountRequestRuntimeBlocked(account, "grok-3"))
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
}

func TestRuntimeBlockKeepsActivePersistedCooldown(t *testing.T) {
	svc := &OpenAIGatewayService{}
	until := time.Now().Add(30 * time.Minute)
	account := &Account{
		ID:                     93,
		Platform:               PlatformGrok,
		Type:                   AccountTypeOAuth,
		Status:                 StatusActive,
		Schedulable:            true,
		TempUnschedulableUntil: &until,
	}
	svc.BlockAccountScheduling(account, until, "grok payment required")
	require.True(t, svc.isOpenAIAccountRequestRuntimeBlocked(account, "grok-3"))
	require.True(t, svc.isOpenAIAccountRuntimeBlocked(account))
}

func TestRuntimeCooldownCASPreservesConcurrentSameDeadlineGeneration(t *testing.T) {
	svc := &OpenAIGatewayService{}
	until := time.Now().Add(time.Minute)
	var wg sync.WaitGroup
	for worker := range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			account := &Account{ID: int64(worker + 1), Platform: PlatformOpenAI}
			for range 25 {
				svc.BlockAccountScheduling(account, until, "old")
				snapshot := svc.peekOpenAIAccountRuntimeBlock(account)
				svc.BlockAccountScheduling(account, until, "new")
				svc.clearOpenAIAccountRuntimeBlockIfUnchanged(account.ID, snapshot)
				if !svc.isOpenAIAccountRuntimeBlocked(account) {
					t.Error("CAS deleted a newer generation with the same deadline")
					return
				}
			}
		}()
	}
	wg.Wait()
}

func TestRuntimeCooldownReconciliationPreservesModelBlock(t *testing.T) {
	svc := &OpenAIGatewayService{}
	account := &Account{ID: 96, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	svc.BlockAccountScheduling(account, time.Now().Add(time.Minute), "stale")
	for range 3 {
		svc.recordOpenAIAccountModelTransientFailure(account, "gpt-5", time.Now())
	}
	require.True(t, svc.isOpenAIAccountRequestRuntimeBlocked(account, "gpt-5"))
	require.False(t, svc.isOpenAIAccountRuntimeBlocked(account))
	require.False(t, svc.isOpenAIAccountRequestRuntimeBlocked(account, "gpt-5.1"))
}
