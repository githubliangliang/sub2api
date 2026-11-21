package service

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestCanonicalRequestedEffort024InputUpdates(t *testing.T) {
	for _, tc := range []struct{ body, want string }{
		{`{"reasoning":{"effort":"medium"},"input":[{"type":"configuration_update","reasoning":{"effort":"high"}}]}`, "high"},
		{`{"reasoning":{"effort":"low"},"input":[{"type":"configuration_update","reasoning":{"effort":"none"}},{"type":"configuration_update","reasoning":{"effort":"max"}}]}`, "max"},
		{`{"input":[{"type":"configuration_update","reasoning":{"effort":"none"}},{"type":"configuration_update","reasoning":{"effort":""}},{"type":"configuration_update"}]}`, "none"},
		{`{"input":[{"type":"configuration_update","reasoning":{"effort":"high"}},{"type":"configuration_update","reasoning":{"effort":null}}]}`, "high"},
		{`{"input":[{"type":"configuration_update","reasoning":{"effort":"high"}},{"type":"configuration_update","reasoning":{"effort":"invalid"}}]}`, "high"},
		{`{"reasoning":{"effort":"low"},"in\u0070ut":[{"type":"configuration_\u0075pdate","reasoning":{"effort":"hi\u0067h"}}]}`, "high"},
		{`{"output_config":{"effort":"max"},"input":[]}`, "max"},
		{`{"input":[{"type":"message","reasoning":{"effort":"high"}}]}`, ""},
		{`{"model":"gpt-6.2-sol","input":[]}`, ""},
	} {
		got := CanonicalRequestedReasoningEffort([]byte(tc.body))
		if tc.want == "" {
			require.Nil(t, got, tc.body)
		} else {
			require.NotNil(t, got, tc.body)
			require.Equal(t, tc.want, *got, tc.body)
		}
	}
}

func TestRequestedEffort024OwnsInputScalar(t *testing.T) {
	body := []byte(`{"input":[{"type":"configuration_update","reasoning":{"effort":"high"}}]}`)
	effort := CanonicalRequestedReasoningEffort(body)
	for i := range body {
		body[i] = 'x'
	}
	require.NotNil(t, effort)
	require.Equal(t, "high", *effort)
}

func TestRequestView024OwnsScalars(t *testing.T) {
	body := []byte(`{"model":"gpt-6.2-sol","prompt_cache_key":"cache","previous_response_id":"resp","service_tier":"fast","reasoning":{"effort":"high"},"input":"` + strings.Repeat("x", 1<<20) + `"}`)
	view := newOpenAIRequestView(body)
	for i := range body {
		body[i] = 'x'
	}
	require.Equal(t, "gpt-6.2-sol", view.Model)
	require.Equal(t, "cache", view.PromptCacheKey)
	require.Equal(t, "resp", view.PreviousResponseID)
	require.Equal(t, "fast", view.ServiceTier)
	require.Equal(t, "high", view.ReasoningEffort)
}
