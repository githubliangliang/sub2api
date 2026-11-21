package handler

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"testing"
)

func TestCodex024StandaloneContexts(t *testing.T) {
	for _, tc := range []struct {
		name, output string
		changed      bool
	}{
		{"ide_context", "# Active file: main.go", true},
		{"ide_context", " ", false},
		{"untrusted_content", "", true},
		{"untrusted_content", " ", true},
		{"proposed_plan", " ", true},
		{"proposed_plan", "", false},
		{"unknown", "context", false},
	} {
		t.Run(tc.name+tc.output, func(t *testing.T) {
			body := pageContextRequest(t, pageContextItem(tc.name, tc.output))
			got, changed := normalizeCodexPageContext(body)
			require.Equal(t, tc.changed, changed)
			if changed {
				require.Equal(t, "user", gjson.GetBytes(got, "input.0.role").String())
				require.Equal(t, "<external_"+tc.name+">"+tc.output+"</external_"+tc.name+">", gjson.GetBytes(got, "input.0.content.0.text").String())
			} else {
				require.Equal(t, body, got)
			}
		})
	}
	for _, name := range []string{"ide_context", "untrusted_content", "proposed_plan"} {
		for _, override := range []map[string]any{{"output": nil}, {"output": []any{}}, {"call_id": "real"}, {"call_id": nil}, {"namespace": "other"}} {
			item := pageContextItem(name, "context")
			for k, v := range override {
				item[k] = v
			}
			body := pageContextRequest(t, item)
			got, changed := normalizeCodexPageContext(body)
			require.False(t, changed, "%s %v", name, override)
			require.Equal(t, body, got)
		}
		raw, _ := json.Marshal(name)
		body := []byte(`{"input":[{"type":"function_call_output","namespace":"agent","name":` + string(raw) + `,"output":"a","output":"b"}]}`)
		_, changed := normalizeCodexPageContext(body)
		require.False(t, changed)
	}
}
