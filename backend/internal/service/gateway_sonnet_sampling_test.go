package service

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestClaudeOAuthSamplingDefaults(t *testing.T) {
	for _, tc := range []struct {
		name, model, sampling string
		wantTemperature       bool
		wantValue             float64
	}{
		{"sonnet top_p only", "claude-sonnet-5-5", `,"top_p":0.99`, false, 0},
		{"sonnet missing", "claude-sonnet-5-5", "", true, 1},
		{"sonnet explicit", "claude-sonnet-5-5", `,"temperature":0.5`, true, 0.5},
		{"opus missing", "claude-opus-5-5", "", false, 0},
		{"older model", "claude-sonnet-4-6", `,"top_p":0.99`, true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte(`{"model":"` + tc.model + `","messages":[]` + tc.sampling + `}`)
			out, _ := normalizeClaudeOAuthRequestBody(body, tc.model, claudeOAuthNormalizeOptions{})
			temperature := gjson.GetBytes(out, "temperature")
			require.Equal(t, tc.wantTemperature, temperature.Exists())
			if tc.wantTemperature {
				require.Equal(t, tc.wantValue, temperature.Float())
			}
			require.Equal(t, gjson.GetBytes(body, "top_p").Raw, gjson.GetBytes(out, "top_p").Raw)
		})
	}
}
