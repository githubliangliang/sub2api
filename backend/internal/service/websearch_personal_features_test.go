//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/websearch"
	"github.com/stretchr/testify/require"
)

func TestShouldEmulateWebSearchFor_OpenAIEligibility(t *testing.T) {
	SetWebSearchManager(websearch.NewManager([]websearch.ProviderConfig{{Type: "brave", APIKey: "local"}}, nil))
	t.Cleanup(func() { SetWebSearchManager(nil); clearGlobalWebSearchConfig() })
	setGlobalWebSearchConfig(&WebSearchEmulationConfig{Enabled: true, Providers: []WebSearchProviderConfig{{Type: "brave", APIKey: "local"}}})
	settings := newSettingServiceForWebSearchTest(true)
	groupID := int64(42)
	channel := newChannelServiceWithCache(groupID, &Channel{ID: 10, Status: StatusActive,
		FeaturesConfig: map[string]any{featureKeyWebSearchEmulation: map[string]any{PlatformOpenAI: true}}})
	for _, tc := range []struct {
		name, platform, kind string
		mode                 any
		want                 bool
		shadow               bool
	}{
		{"enabled", PlatformOpenAI, AccountTypeAPIKey, "enabled", true, false},
		{"disabled", PlatformOpenAI, AccountTypeAPIKey, "disabled", false, false},
		{"channel", PlatformOpenAI, AccountTypeAPIKey, "default", true, false},
		{"old true", PlatformOpenAI, AccountTypeAPIKey, true, true, false},
		{"old false follows channel", PlatformOpenAI, AccountTypeAPIKey, false, true, false},
		{"oauth native", PlatformOpenAI, AccountTypeOAuth, "enabled", false, false},
		{"setup token native", PlatformOpenAI, AccountTypeSetupToken, "enabled", false, false},
		{"shadow native", PlatformOpenAI, AccountTypeAPIKey, "enabled", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := &Account{Platform: tc.platform, Type: tc.kind, Extra: map[string]any{featureKeyWebSearchEmulation: tc.mode}}
			if tc.shadow {
				parent := int64(7)
				a.ParentAccountID = &parent
			}
			require.Equal(t, tc.want, shouldEmulateWebSearchFor(context.Background(), settings, channel, a, &groupID, webSearchToolBody))
		})
	}
	a := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Extra: map[string]any{featureKeyWebSearchEmulation: "enabled"}}
	require.False(t, shouldEmulateWebSearchFor(context.Background(), nil, channel, a, &groupID, webSearchToolBody))
	require.False(t, shouldEmulateWebSearchFor(context.Background(), settings, channel, nil, &groupID, webSearchToolBody))
	require.False(t, shouldEmulateWebSearchFor(context.Background(), settings, channel, a, &groupID,
		[]byte(`{"tools":[{"type":"web_search"},{"name":"bash"}]}`)))
	setGlobalWebSearchConfig(&WebSearchEmulationConfig{Enabled: false})
	require.False(t, shouldEmulateWebSearchFor(context.Background(), settings, channel, a, &groupID, webSearchToolBody))
}

func TestSetWebSearchManager_TypedNilRemainsUnavailable(t *testing.T) {
	t.Cleanup(func() { SetWebSearchManager(nil) })
	var manager *websearch.Manager
	SetWebSearchManager(manager)
	require.Nil(t, getWebSearchExecutor())
	require.Nil(t, getWebSearchManager())
}
