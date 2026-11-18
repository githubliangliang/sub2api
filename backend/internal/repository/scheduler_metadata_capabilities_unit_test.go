//go:build unit

package repository

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSchedulerMetadataPreservesExistingEndpointCapabilities(t *testing.T) {
	for _, scenario := range []string{"unconfigured", "empty", "chat_only", "embeddings_only", "map"} {
		t.Run(scenario, func(t *testing.T) {
			credentials := map[string]any{"api_key": "fixture", "base_url": "https://relay.example", "unlisted_secret": "drop"}
			wantChat, wantEmbeddings := true, true
			switch scenario {
			case "empty":
				credentials["openai_capabilities"] = []string{}
			case "chat_only":
				credentials["openai_capabilities"] = []string{"chat_completions"}
				wantEmbeddings = false
			case "embeddings_only":
				credentials["openai_capabilities"] = []string{"embeddings"}
				wantChat = false
			case "map":
				credentials["openai_capabilities"] = map[string]any{"chat_completions": true, "embeddings": false}
				wantEmbeddings = false
			}
			original := service.Account{ID: 1, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Credentials: credentials}
			projected := buildSchedulerMetadataAccount(original)
			require.Equal(t, wantChat, original.SupportsOpenAIEndpointCapability(service.OpenAIEndpointCapabilityChatCompletions))
			require.Equal(t, wantChat, projected.SupportsOpenAIEndpointCapability(service.OpenAIEndpointCapabilityChatCompletions))
			require.Equal(t, wantEmbeddings, projected.SupportsOpenAIEndpointCapability(service.OpenAIEndpointCapabilityEmbeddings))
			require.Equal(t, wantChat, projected.SupportsOpenAIEndpointCapability(service.OpenAIEndpointCapabilityResponses))
			require.NotContains(t, projected.Credentials, "unlisted_secret")
			if scenario == "unconfigured" {
				require.NotContains(t, projected.Credentials, "openai_capabilities")
			} else {
				require.Contains(t, projected.Credentials, "openai_capabilities")
			}
		})
	}
}

func TestSchedulerMetadataCapabilitiesSurviveRedisSnapshot(t *testing.T) {
	ctx := context.Background()
	cache := newSchedulerCacheUnit(t)
	bucket := service.SchedulerBucket{GroupID: 10, Platform: service.PlatformOpenAI, Mode: "mixed"}
	account := service.Account{ID: 1, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey,
		Status: service.StatusActive, Schedulable: true,
		Credentials: map[string]any{"api_key": "fixture", "openai_capabilities": map[string]any{"chat_completions": true, "embeddings": false}, "unlisted_secret": "full-account-only"},
	}
	token, err := cache.CaptureBucketWriteToken(ctx, bucket)
	require.NoError(t, err)
	require.NoError(t, cache.SetSnapshot(ctx, bucket, token, []service.Account{account}))
	snapshot, ready, err := cache.GetSnapshot(ctx, bucket)
	require.NoError(t, err)
	require.True(t, ready)
	require.Len(t, snapshot, 1)
	require.True(t, snapshot[0].SupportsOpenAIEndpointCapability(service.OpenAIEndpointCapabilityChatCompletions))
	require.False(t, snapshot[0].SupportsOpenAIEndpointCapability(service.OpenAIEndpointCapabilityEmbeddings))
	require.NotContains(t, snapshot[0].Credentials, "unlisted_secret")
	hydrated, err := cache.GetAccount(ctx, account.ID)
	require.NoError(t, err)
	require.NotNil(t, hydrated)
	require.False(t, hydrated.SupportsOpenAIEndpointCapability(service.OpenAIEndpointCapabilityEmbeddings))
	require.Equal(t, "full-account-only", hydrated.Credentials["unlisted_secret"])
}
