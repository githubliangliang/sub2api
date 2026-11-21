package service

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAnthropicRegistryOfficialHostAndCustomEndpoint(t *testing.T) {
	a := &Account{Platform: PlatformAnthropic, Type: AccountTypeAPIKey, Credentials: map[string]any{}}
	require.Equal(t, "https://api.anthropic.com", upstreamModelRegistryBaseURL(a))
	registry := map[string]modelsDevProvider{"anthropic": {ID: "anthropic", Models: map[string]modelsDevModel{"fixture": {}}}}
	for _, endpoint := range []string{"https://api.anthropic.com", "https://API.Anthropic.Com/v1"} {
		p, ok := matchModelsDevProviderByKnownHost(registry, endpoint)
		require.True(t, ok)
		require.Equal(t, "anthropic", p.ID)
	}
	a.Credentials["base_url"] = "https://custom.example.invalid/v1"
	require.Equal(t, a.GetCredential("base_url"), upstreamModelRegistryBaseURL(a))
	_, ok := matchModelsDevProviderByKnownHost(registry, a.GetCredential("base_url"))
	require.False(t, ok)
}
