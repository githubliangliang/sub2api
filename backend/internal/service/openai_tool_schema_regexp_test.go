package service

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"testing"
)

func TestOpenAIToolSchemaRegexpScopeAndAdjacentEdits(t *testing.T) {
	body := []byte(`{"number":9007199254740993,"tools":[{"parameters":{"type":null,"required":null,"pattern":"(?=drop)","patternProperties":{"^\\p{L}+$":{"required":null,"pattern":"(?=nested)"},"^[a-z]+$":{"type":"string"}},"additionalProperties":{"required":null,"pattern":"(?=nested)"},"default":{"pattern":"(?=instance)","required":null},"enum":[{"pattern":"\\p{L}"}]}}]}`)
	for _, platform := range []string{PlatformAnthropic, PlatformGrok, PlatformGemini} {
		out, _, err := sanitizeOpenAIResponsesToolSchemasForPlatform(body, platform)
		require.NoError(t, err)
		require.Equal(t, "(?=drop)", gjson.GetBytes(out, "tools.0.parameters.pattern").String())
	}
	out, changed, err := sanitizeOpenAIResponsesToolSchemasForPlatform(body, PlatformOpenAI)
	require.NoError(t, err)
	require.True(t, changed)
	require.True(t, json.Valid(out))
	require.Equal(t, "9007199254740993", gjson.GetBytes(out, "number").Raw)
	require.Equal(t, "object", gjson.GetBytes(out, "tools.0.parameters.type").String())
	require.False(t, gjson.GetBytes(out, "tools.0.parameters.required").Exists())
	require.False(t, gjson.GetBytes(out, "tools.0.parameters.pattern").Exists())
	require.Equal(t, "true", gjson.GetBytes(out, "tools.0.parameters.additionalProperties").Raw)
	require.Equal(t, gjson.GetBytes(body, "tools.0.parameters.default").Raw, gjson.GetBytes(out, "tools.0.parameters.default").Raw)
	require.Equal(t, gjson.GetBytes(body, "tools.0.parameters.enum").Raw, gjson.GetBytes(out, "tools.0.parameters.enum").Raw)
}

func TestSanitizeOpenAIResponsesToolSchemaPatterns_DropsPythonRegexIncompatiblePatterns(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","name":"search","parameters":{"type":"object","properties":{
		"unicode":{"type":"string","pattern":"^\\p{L}+$"},
		"unknownEscape":{"type":"string","pattern":"^\\R+$"},
		"named":{"type":"string","pattern":"^(?<word>[a-z]+)$"},
		"namedSingle":{"type":"string","pattern":"^(?'word'[a-z]+)$"},
		"pythonNamed":{"type":"string","pattern":"^(?P<word>[a-z]+)$"},
		"escapedLiteral":{"type":"string","pattern":"^\\\\p{L}+$"},
		"digits":{"type":"string","pattern":"^\\d+$"}
	}}}]}`)

	sanitized, changed, err := sanitizeOpenAIResponsesToolSchemaPatterns(body)
	require.NoError(t, err)
	require.True(t, changed)
	for _, name := range []string{"unicode", "unknownEscape", "named", "namedSingle"} {
		require.False(t, gjson.GetBytes(sanitized, "tools.0.parameters.properties."+name+".pattern").Exists(), name)
	}
	require.Equal(t, `^(?P<word>[a-z]+)$`, gjson.GetBytes(sanitized, "tools.0.parameters.properties.pythonNamed.pattern").String())
	require.Equal(t, `^\\p{L}+$`, gjson.GetBytes(sanitized, "tools.0.parameters.properties.escapedLiteral.pattern").String())
	require.Equal(t, `^\d+$`, gjson.GetBytes(sanitized, "tools.0.parameters.properties.digits.pattern").String())
	require.True(t, json.Valid(sanitized))
}

func TestSanitizeOpenAIResponsesToolSchemaPatterns_DropsUnicodePropertyEscape(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","name":"search","parameters":{"type":"object","properties":{"word":{"type":"string","pattern":"^\\p{L}+$"}}}}]}`)

	sanitized, changed, err := sanitizeOpenAIResponsesToolSchemaPatterns(body)

	require.NoError(t, err)
	require.True(t, changed)
	require.False(t, gjson.GetBytes(sanitized, "tools.0.parameters.properties.word.pattern").Exists())
}

func TestSanitizeOpenAIResponsesToolSchemaPatterns_DropsUnsupportedPatternProperties(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","name":"search","parameters":{"type":"object","patternProperties":{
		"^\\p{L}+$":{"type":"string","pattern":"(?=nested)"},
		"^(?<word>[a-z]+)$":{"type":"number"},
		"^[a-z]+$":{"type":"string","pattern":"(?=nested)"},
		"^\\\\p{L}+$":{"type":"boolean"}
	}}}]}`)

	sanitized, changed, err := sanitizeOpenAIResponsesToolSchemaPatterns(body)

	require.NoError(t, err)
	require.True(t, changed)
	var doc struct {
		Tools []struct {
			Parameters struct {
				PatternProperties map[string]struct {
					Type    string  `json:"type"`
					Pattern *string `json:"pattern"`
				} `json:"patternProperties"`
			} `json:"parameters"`
		} `json:"tools"`
	}
	require.NoError(t, json.Unmarshal(sanitized, &doc))
	properties := doc.Tools[0].Parameters.PatternProperties
	require.NotContains(t, properties, `^\p{L}+$`)
	require.NotContains(t, properties, `^(?<word>[a-z]+)$`)
	valid := properties[`^[a-z]+$`]
	require.Equal(t, "string", valid.Type)
	require.Nil(t, valid.Pattern)
	escaped := properties[`^\\p{L}+$`]
	require.Equal(t, "boolean", escaped.Type)
	require.True(t, json.Valid(sanitized))
}

// patternProperties 与 additionalProperties 共同校验对象键。删除一个 Python re
// 不接受的 patternProperties 键后，原本匹配该模式的参数键会变成 additional
// property；对象若仍是 additionalProperties:false，合法调用会被上游拒绝。
func TestSanitizeOpenAIResponsesToolSchemaPatterns_ReopensClosedObjectAfterDroppedPattern(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		schemaPath string
		instance   string
	}{
		{
			name:       "additionalProperties after patternProperties",
			body:       `{"tools":[{"type":"function","name":"t","parameters":{"type":"object","patternProperties":{"^\\p{L}+$":{"type":"string"}},"additionalProperties":false}}]}`,
			schemaPath: "tools.0.parameters",
			instance:   `{"abc":"value"}`,
		},
		{
			name:       "additionalProperties before patternProperties",
			body:       `{"tools":[{"type":"function","name":"t","parameters":{"type":"object","additionalProperties":false,"patternProperties":{"^\\p{L}+$":{"type":"string"}}}}]}`,
			schemaPath: "tools.0.parameters",
			instance:   `{"abc":"value"}`,
		},
		{
			name:       "nested dictionary argument",
			body:       `{"tools":[{"type":"function","name":"t","parameters":{"type":"object","properties":{"labels":{"type":"object","patternProperties":{"^\\p{L}+$":{"type":"string"}},"additionalProperties":false}}}}]}`,
			schemaPath: "tools.0.parameters",
			instance:   `{"labels":{"abc":"value"}}`,
		},
		{
			name:       "additionalProperties schema",
			body:       `{"tools":[{"type":"function","name":"t","parameters":{"type":"object","patternProperties":{"^\\p{L}+$":{"type":"string"}},"additionalProperties":{"type":"integer"}}}]}`,
			schemaPath: "tools.0.parameters",
			instance:   `{"abc":"value","123":5}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized, changed, err := sanitizeOpenAIResponsesToolSchemaPatterns([]byte(tt.body))
			require.NoError(t, err)
			require.True(t, changed)
			require.NotContains(t, string(sanitized), `\p{L}`)
			require.NotEqual(t, "false", gjson.GetBytes(sanitized, tt.schemaPath+".additionalProperties").Raw)
			require.True(t, json.Valid(sanitized))
		})
	}
}

// 未删除 patternProperties 键时不得放宽封闭对象；只删除其值内部的 pattern
// 也不能改写 additionalProperties:false。
func TestSanitizeOpenAIResponsesToolSchemaPatterns_KeepsAdditionalPropertiesSchemaWhenNoPatternKeyDropped(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","name":"t","parameters":{"type":"object","patternProperties":{"^[a-z]+$":{"type":"string"}},"additionalProperties":{"type":"integer"}}}]}`)

	sanitized, changed, err := sanitizeOpenAIResponsesToolSchemaPatterns(body)

	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, "integer", gjson.GetBytes(sanitized, "tools.0.parameters.additionalProperties.type").String())
	require.True(t, json.Valid(sanitized))
}

func TestSanitizeOpenAIResponsesToolSchemaPatterns_KeepsClosedObjectWhenNoPatternKeyDropped(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		changed bool
	}{
		{
			name:    "supported pattern key",
			body:    `{"tools":[{"type":"function","name":"t","parameters":{"type":"object","patternProperties":{"^[a-z]+$":{"type":"string"}},"additionalProperties":false}}]}`,
			changed: false,
		},
		{
			name:    "unsupported pattern inside pattern value",
			body:    `{"tools":[{"type":"function","name":"t","parameters":{"type":"object","patternProperties":{"^[a-z]+$":{"type":"string","pattern":"(?=x)"}},"additionalProperties":false}}]}`,
			changed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized, changed, err := sanitizeOpenAIResponsesToolSchemaPatterns([]byte(tt.body))
			require.NoError(t, err)
			require.Equal(t, tt.changed, changed)
			require.Equal(t, "false", gjson.GetBytes(sanitized, "tools.0.parameters.additionalProperties").Raw)
		})
	}
}

// 外层补丁（additionalProperties 整体改写为 true、整条删除 patternProperties
// 键）覆盖内层补丁时，内层补丁无意义，不得让整次清洗以 overlapping patches
// 失败（否则请求在本地就以 502 结束，连原样转发都做不到）。
func TestSanitizeOpenAIResponsesToolSchemaPatterns_EnclosingEditSupersedesNestedEdits(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		instance string
	}{
		{
			name:     "reopened additionalProperties schema contains unsupported pattern",
			body:     `{"tools":[{"type":"function","name":"t","parameters":{"type":"object","patternProperties":{"^\\p{L}+$":{"type":"string"}},"additionalProperties":{"type":"string","pattern":"^\\p{N}+$"}}}]}`,
			instance: `{"abc":"value","1":"x"}`,
		},
		{
			name:     "reopened additionalProperties schema contains its own reopen",
			body:     `{"tools":[{"type":"function","name":"t","parameters":{"type":"object","patternProperties":{"^\\p{L}+$":{"type":"string"}},"additionalProperties":{"type":"object","patternProperties":{"^\\p{N}+$":{"type":"string"}},"additionalProperties":false}}}]}`,
			instance: `{"abc":"value","1":{"2":"x"}}`,
		},
		{
			name:     "dropped patternProperties value contains a reopen",
			body:     `{"tools":[{"type":"function","name":"t","parameters":{"type":"object","patternProperties":{"^\\p{L}+$":{"type":"object","patternProperties":{"^\\p{N}+$":{"type":"string"}},"additionalProperties":false}}}}]}`,
			instance: `{"abc":{"1":"x"}}`,
		},
		{
			name:     "anthropic input_schema",
			body:     `{"tools":[{"name":"t","input_schema":{"type":"object","patternProperties":{"^\\p{L}+$":{"type":"string"}},"additionalProperties":{"type":"string","pattern":"(?<n>\\d+)"}}}]}`,
			instance: `{"abc":"value"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized, changed, err := sanitizeOpenAIResponsesToolSchemasForPlatform([]byte(tt.body), PlatformOpenAI)
			require.NoError(t, err)
			require.True(t, changed)
			require.True(t, json.Valid(sanitized))
			require.NotContains(t, string(sanitized), `\p{`)
			require.NotContains(t, string(sanitized), `(?<`)
			schema := gjson.GetBytes(sanitized, "tools.0.parameters")
			if !schema.Exists() {
				schema = gjson.GetBytes(sanitized, "tools.0.input_schema")
			}
			require.NotEqual(t, "false", schema.Get("additionalProperties").Raw)
		})
	}
}

// 包含关系以外的重叠仍须报错；零宽插入与同起点删除须同时生效。
// 评审复现路径：API Key 与 OAuth 账号都经过 WebSocket 兼容归一化。
func TestOpenAIResponsesToolSchemaPlatformGate_ReopensClosedObject(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","name":"t","parameters":{"type":"object","patternProperties":{"^\\p{L}+$":{"type":"string"}},"additionalProperties":false}}]}`)

	for _, accountType := range []string{AccountTypeAPIKey, AccountTypeOAuth} {
		t.Run(accountType, func(t *testing.T) {
			normalized, changed, err := normalizeOpenAIResponsesWebSocketCompatibilityBody(body, &Account{
				Platform: PlatformOpenAI,
				Type:     accountType,
			}, false)
			require.NoError(t, err)
			require.True(t, changed)
			require.NotEqual(t, "false", gjson.GetBytes(normalized, "tools.0.parameters.additionalProperties").Raw)
		})
	}
}
