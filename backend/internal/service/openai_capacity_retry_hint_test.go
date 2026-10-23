package service

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenAICapacityRetryHintsAreBoundedBeforeDurationConversion(t *testing.T) {
	for _, tc := range []struct {
		name, header, value, body string
		want                      time.Duration
	}{
		{"seconds", "Retry-After", "6", "", 6 * time.Second},
		{"fractional seconds", "Retry-After", "0.25", "", 250 * time.Millisecond},
		{"milliseconds", "Retry-After-Ms", "1500", "", 1500 * time.Millisecond},
		{"date", "Retry-After", time.Now().Add(time.Hour).UTC().Format(http.TimeFormat), "", 8 * time.Second},
		{"past date", "Retry-After", time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat), "", 0},
		{"seconds overflow", "Retry-After", "1e300", "", 8 * time.Second},
		{"milliseconds overflow", "Retry-After-Ms", "1e300", "", 8 * time.Second},
		{"negative", "Retry-After", "-9", "", 0},
		{"zero", "Retry-After", "0", "", 0},
		{"nan", "Retry-After", "NaN", "", 0},
		{"infinity", "Retry-After", "+Inf", "", 0},
		{"milliseconds infinity", "Retry-After-Ms", "+Inf", "", 0},
		{"invalid", "Retry-After", "not a delay", "", 0},
		{"json seconds", "", "", `{"error":{"headers":{"retry-after":"2"}}}`, 2 * time.Second},
		{"nested json milliseconds", "", "", `{"response":{"error":{"headers":{"retry-after-ms":1250}}}}`, 1250 * time.Millisecond},
		{"nested json seconds", "", "", `{"response":{"error":{"headers":{"retry-after":"3"}}}}`, 3 * time.Second},
		{"invalid header falls back", "Retry-After", "NaN", `{"error":{"headers":{"retry-after-ms":"500"}}}`, 500 * time.Millisecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			headers := make(http.Header)
			if tc.header != "" {
				headers.Set(tc.header, tc.value)
			}
			err := newOpenAIUpstreamFailoverError(http.StatusServiceUnavailable, headers, []byte(tc.body), "Our servers are currently overloaded. Please try again later.", false)
			require.True(t, err.RequestScopedTransient)
			require.Equal(t, tc.want, err.SameAccountRetryDelay)
		})
	}
	ordinary := newOpenAIUpstreamFailoverError(http.StatusServiceUnavailable, http.Header{"Retry-After": []string{"6"}}, nil, "ordinary server error", false)
	require.False(t, ordinary.RequestScopedTransient)
	require.Zero(t, ordinary.SameAccountRetryDelay)
}
