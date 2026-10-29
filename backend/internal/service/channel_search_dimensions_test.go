//go:build unit

package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChannelUpdatePreservesSearchDimensions(t *testing.T) {
	for _, tc := range []struct {
		name     string
		incoming map[string]any
		search   any
	}{
		{"partial", map[string]any{featureKeyWebSearchEmulation: map[string]any{"anthropic": false}, "new": true}, map[string]any{"anthropic": false, "openai": true, "future": true}},
		{"missing", map[string]any{"new": true}, map[string]any{"anthropic": true, "openai": true, "future": true}},
		{"unknown", map[string]any{featureKeyWebSearchEmulation: map[string]any{"future": false}}, map[string]any{"anthropic": true, "openai": true, "future": false}},
		{"empty", map[string]any{featureKeyWebSearchEmulation: map[string]any{}}, map[string]any{}},
		{"null", map[string]any{featureKeyWebSearchEmulation: nil}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			existing := map[string]any{featureKeyWebSearchEmulation: map[string]any{"anthropic": true, "openai": true, "future": true}, "old": map[string]any{"keep": true}}
			original, err := json.Marshal(existing)
			require.NoError(t, err)
			input, err := json.Marshal(tc.incoming)
			require.NoError(t, err)
			channel := &Channel{FeaturesConfig: existing}
			err = (&ChannelService{}).applyUpdateInput(context.Background(), channel, &UpdateChannelInput{FeaturesConfig: tc.incoming})
			require.NoError(t, err)
			require.Equal(t, tc.search, channel.FeaturesConfig[featureKeyWebSearchEmulation])
			require.NotContains(t, channel.FeaturesConfig, "old", "other features retain whole-block replacement")
			after, err := json.Marshal(existing)
			require.NoError(t, err)
			require.Equal(t, original, after)
			after, err = json.Marshal(tc.incoming)
			require.NoError(t, err)
			require.Equal(t, input, after)
		})
	}
}
