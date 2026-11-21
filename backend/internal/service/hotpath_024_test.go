package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"math"
	"runtime"
	"sort"
	"strings"
	"testing"
	"unsafe"
)

func legacy024CacheControlLimit(body []byte) []byte {
	if len(body) == 0 {
		return body
	}
	if bytes.Contains(body, []byte(`\u`)) {
		return normalizeCacheControlTTLOrder(enforceCacheControlBlockLimit(body))
	}
	key := []byte(`"cache_control"`)
	if !bytes.Contains(body, key) {
		return body
	}
	if bytes.Count(body, key) <= maxCacheControlBlocks && !bytes.Contains(body, []byte(`"thinking"`)) {
		if !bytes.Contains(body, []byte(`"1h"`)) {
			return body
		}
		return normalizeCacheControlTTLOrder(body)
	}
	return normalizeCacheControlTTLOrder(enforceCacheControlBlockLimit(body))
}

func legacy024StripNamespaces(body []byte, keepToolCallNamespaces bool) ([]byte, error) {
	if !bytes.Contains(body, []byte(`"namespace"`)) {
		return body, nil
	}
	input := gjson.GetBytes(body, "input")
	if !input.IsArray() {
		return body, nil
	}

	var rebuilt bytes.Buffer
	changed := false
	// Offsets below are relative to input.Raw, not the full request body.
	copyFrom := 0
	var stripErr error
	input.ForEach(func(_, item gjson.Result) bool {
		if !item.IsObject() || !item.Get("namespace").Exists() ||
			(keepToolCallNamespaces && isOpenAIResponsesToolCallItemType(item.Get("type").String())) {
			return true
		}
		if !changed && !gjson.ValidBytes(body) {
			stripErr = fmt.Errorf("invalid JSON request body")
			return false
		}
		itemBody, err := sjson.DeleteBytes([]byte(item.Raw), "namespace")
		if err != nil {
			stripErr = err
			return false
		}
		if !changed {
			// Allocate only on the first change; copy unchanged spans without
			// traversing their items again or converting them to byte slices.
			rebuilt.Grow(len(input.Raw))
			changed = true
		}
		itemStart := item.Index - input.Index
		_, _ = rebuilt.WriteString(input.Raw[copyFrom:itemStart])
		_, _ = rebuilt.Write(itemBody)
		copyFrom = itemStart + len(item.Raw)
		return true
	})
	if stripErr != nil {
		return body, fmt.Errorf("delete OpenAI input namespace: %w", stripErr)
	}
	if !changed {
		return body, nil
	}
	_, _ = rebuilt.WriteString(input.Raw[copyFrom:])
	stripped, err := sjson.SetRawBytes(body, "input", rebuilt.Bytes())
	if err != nil {
		return body, fmt.Errorf("replace OpenAI input after namespace deletion: %w", err)
	}
	return stripped, nil
}

func legacy024ImageIntent(endpoint string, requestedModel string, body []byte) bool {
	if IsImageGenerationEndpoint(endpoint) {
		return true
	}
	if isOpenAIImageGenerationModel(requestedModel) {
		return true
	}
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return false
	}

	var modelSeen, toolsSeen, inputSeen, toolChoiceSeen bool
	imageIntent := false
	parseRawJSONView(body).ForEach(func(key, value gjson.Result) bool {
		// GetBytes returns the first duplicate key; retain that behavior while walking the root once.
		switch key.Str {
		case "model":
			if !modelSeen {
				modelSeen = true
				imageIntent = isOpenAIImageGenerationModel(strings.TrimSpace(value.String()))
			}
		case "tools":
			if !toolsSeen {
				toolsSeen = true
				imageIntent = openAIJSONToolsContainImageGeneration(value)
			}
		case "input":
			if !inputSeen {
				inputSeen = true
				imageIntent = openAIJSONInputContainsImageGenTool(value)
			}
		case "tool_choice":
			if !toolChoiceSeen {
				toolChoiceSeen = true
				imageIntent = openAIJSONToolChoiceSelectsImageGeneration(value)
			}
		}
		return !imageIntent && (!modelSeen || !toolsSeen || !inputSeen || !toolChoiceSeen)
	})
	return imageIntent
}

// legacy024ExplicitImageIntent 仅检测原生 image_generation 工具、图片模型和显式 tool_choice，
// 不检测被动的 image_gen namespace 声明。用于 capability 路由决策——被动 namespace 不应
// 强制要求原生 Responses 能力，否则 Chat Completions-only 账号会被误过滤（#4476）。
func legacy024ExplicitImageIntent(endpoint string, requestedModel string, body []byte) bool {
	if IsImageGenerationEndpoint(endpoint) || isOpenAIImageGenerationModel(requestedModel) {
		return true
	}
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return false
	}
	var modelSeen, toolsSeen, toolChoiceSeen bool
	imageIntent := false
	parseRawJSONView(body).ForEach(func(key, value gjson.Result) bool {
		switch key.Str {
		case "model":
			if !modelSeen {
				modelSeen = true
				imageIntent = isOpenAIImageGenerationModel(strings.TrimSpace(value.String()))
			}
		case "tools":
			if !toolsSeen {
				toolsSeen = true
				imageIntent = openAIJSONToolsContainNativeImageGeneration(value)
			}
		case "tool_choice":
			if !toolChoiceSeen {
				toolChoiceSeen = true
				imageIntent = openAIJSONToolChoiceSelectsExplicitImageGeneration(value)
			}
		}
		return !imageIntent && (!modelSeen || !toolsSeen || !toolChoiceSeen)
	})
	return imageIntent
}

// legacy024ImageIntentForPlatform applies platform-specific intent rules.
//
// Codex advertises the image_gen namespace on ordinary Responses requests so
// that it is available if the model needs it. Grok strips namespace and
// Responses Lite additional_tools declarations before forwarding, so those
// declarations alone must not turn every Codex request into an image request.
// Native image_generation tools, explicit image selection and image models
// remain image intent. Other platforms retain the original declaration rule.
func legacy024ImageIntentForPlatform(endpoint string, requestedModel string, body []byte, platform string) bool {
	if !strings.EqualFold(strings.TrimSpace(platform), PlatformGrok) {
		return legacy024ImageIntent(endpoint, requestedModel, body)
	}
	return legacy024GrokImageIntent(endpoint, requestedModel, body)
}

func legacy024GrokImageIntent(endpoint string, requestedModel string, body []byte) bool {
	if IsImageGenerationEndpoint(endpoint) || isOpenAIImageGenerationModel(requestedModel) {
		return true
	}
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return false
	}

	var modelSeen, toolsSeen, toolChoiceSeen bool
	imageIntent := false
	parseRawJSONView(body).ForEach(func(key, value gjson.Result) bool {
		switch key.Str {
		case "model":
			if !modelSeen {
				modelSeen = true
				imageIntent = isOpenAIImageGenerationModel(strings.TrimSpace(value.String()))
			}
		case "tools":
			if !toolsSeen {
				toolsSeen = true
				// Grok removes namespace catalogs before forwarding. Native
				// image_generation remains an explicit capability request.
				imageIntent = openAIJSONToolsContainNativeImageGeneration(value)
			}
		case "tool_choice":
			if !toolChoiceSeen {
				toolChoiceSeen = true
				imageIntent = openAIJSONToolChoiceSelectsExplicitImageGeneration(value)
			}
		}
		return !imageIntent && (!modelSeen || !toolsSeen || !toolChoiceSeen)
	})
	return imageIntent
}

func legacy024ParseGatewayBody(parsed *ParsedRequest, protocol string) error {
	if parsed == nil || parsed.Body == nil {
		return fmt.Errorf("empty request body")
	}

	bodyBytes := parsed.Body.Bytes()
	if !gjson.ValidBytes(bodyBytes) {
		return DescribeInvalidJSON(bodyBytes)
	}

	// 只在当前函数内零拷贝读取 JSON 字段；ReplaceBody 后必须重新进入本函数刷新派生状态。
	jsonStr := *(*string)(unsafe.Pointer(&bodyBytes))
	clearGatewayRequestDerivedState(parsed)
	parsed.protocol = protocol

	modelResult := gjson.Get(jsonStr, "model")
	if modelResult.Exists() {
		if modelResult.Type != gjson.String {
			return fmt.Errorf("invalid model field type")
		}
		parsed.Model = modelResult.String()
		if protocol == domain.PlatformAnthropic {
			normalizedModel := normalizeClaudeCodeLongContextModel(parsed.Model)
			if normalizedModel != parsed.Model {
				normalizedBody, err := sjson.SetBytes(bodyBytes, "model", normalizedModel)
				if err != nil {
					return fmt.Errorf("normalize model field: %w", err)
				}
				parsed.Body.Replace(normalizedBody)
				bodyBytes = normalizedBody
				jsonStr = *(*string)(unsafe.Pointer(&bodyBytes))
				parsed.Model = normalizedModel
			}
		}
	}

	streamResult := gjson.Get(jsonStr, "stream")
	if streamResult.Exists() {
		if streamResult.Type != gjson.True && streamResult.Type != gjson.False {
			return fmt.Errorf("invalid stream field type")
		}
		parsed.Stream = streamResult.Bool()
	}

	parsed.MetadataUserID = gjson.Get(jsonStr, "metadata.user_id").String()

	thinkingType := gjson.Get(jsonStr, "thinking.type").String()
	parsed.ThinkingEnabled = thinkingType == "enabled" || thinkingType == "adaptive" || thinkingType == "between_tools" || (protocol == domain.PlatformAnthropic && isClaude55SignedThinkingModel(parsed.Model))

	parsed.OutputEffort = strings.TrimSpace(gjson.Get(jsonStr, "output_config.effort").String())

	maxTokensResult := gjson.Get(jsonStr, "max_tokens")
	if maxTokensResult.Exists() && maxTokensResult.Type == gjson.Number {
		f := maxTokensResult.Float()
		if !math.IsNaN(f) && !math.IsInf(f, 0) && f == math.Trunc(f) &&
			f <= float64(math.MaxInt) && f >= float64(math.MinInt) {
			parsed.MaxTokens = int(f)
		}
	}

	setGatewayRequestRanges(parsed, protocol, jsonStr)
	return nil
}

func TestHotpath024CandidateEquivalence(t *testing.T) {
	corpus := []string{
		`{"model":"alias","input":"text"}`,
		`{"model":"gpt-image-1","tools":[],"input":[]}`,
		`{"mo\u0064el":"GPT-IMAGE-1","to\u006fls":[{"ty\u0070e":"image_generation"}]}`,
		`{"tools":[{"type":"namespace","name":"files"}],"input":[{"type":"function_call","namespace":"files","arguments":"{}"},{"type":"message","content":{"namespace":"keep"}}]}`,
		`{"tools":[{"type":"function","parameters":{"type":null,"required":null,"properties":{"ID":{"type":"string"},"id":{"type":"string"}},"default":{"required":null,"pattern":"(?=keep)"}}}]}`,
		`{"tools":[{"type":"function","parameters":{"type":"object","pattern":"\u0028?=bad)","patternProperties":{"^\\p{L}":{"type":"string"}}}}]}`,
		`{"tools":[{"parameters":{"t\u0079pe":null}}]}`,
		`{"thinking":{"type":"enabled"},"messages":[{"role":"user","content":[{"type":"text","text":"hi","cache_control":{"type":"ephemeral","ttl":"5m"}}]}]}`,
		`{"messages":[{"content":[{"type":"text","cache_control":{"ttl":"5m"}},{"type":"text","cache_control":{"ttl":"1h"}}]}]}`,
		`{"messages":[{"content":[{"type":"text","cache_control":{"ttl":"1\u0068"}}]}]}`,
		`{"input":[{"namespace":"files"}]` + "\xff",
	}
	for _, text := range corpus {
		body := []byte(text)
		require.Equal(t, legacy024CacheControlLimit(body), enforceCacheControlLimit(body))
		require.Equal(t, legacy024ImageIntentForPlatform("/v1/responses", "alias", body, PlatformOpenAI), IsImageGenerationIntentForPlatform("/v1/responses", "alias", body, PlatformOpenAI), text)
		require.Equal(t, legacy024ExplicitImageIntent("/v1/responses", "alias", body), IsExplicitImageGenerationIntent("/v1/responses", "alias", body), text)
		require.Equal(t, legacy024ImageIntentForPlatform("/v1/responses", "alias", body, PlatformGrok), IsImageGenerationIntentForPlatform("/v1/responses", "alias", body, PlatformGrok), text)
		for _, keep := range []bool{true, false} {
			a, ae := legacy024StripNamespaces(body, keep)
			b, be := stripOpenAIResponsesInputNamespaces(body, keep)
			require.Equal(t, ae != nil, be != nil, text)
			require.Equal(t, a, b, text)
		}
		for _, patterns := range []bool{true, false} {
			a, ac, ae := legacy024SanitizeSchemas(body, patterns)
			b, bc, be := sanitizeOpenAIResponsesToolSchemas(body, patterns)
			require.Equal(t, ae != nil, be != nil, text)
			require.Equal(t, ac, bc, text)
			require.Equal(t, a, b, text)
		}
	}
}
func hotpath024Body() []byte {
	return []byte(`{"model":"alias","stream":true,"thinking":{"type":"enabled"},"tools":[{"type":"namespace","name":"files"}],"input":[{"type":"function_call","namespace":"files","name":"run","arguments":"{}"},{"role":"user","content":"` + strings.Repeat("abc text (ordinary) ", 1<<15) + `"}],"messages":[{"content":[{"type":"text","text":"hello","cache_control":{"type":"ephemeral"}}]}]}`)
}
func BenchmarkHotpath024Candidates(b *testing.B) {
	body := hotpath024Body()
	for _, tc := range []struct {
		name string
		fn   func()
	}{
		{"ttl_before", func() { runtime.KeepAlive(legacy024CacheControlLimit(body)) }},
		{"ttl_after", func() { runtime.KeepAlive(enforceCacheControlLimit(body)) }},
		{"namespace_before", func() { out, _ := legacy024StripNamespaces(body, false); runtime.KeepAlive(out) }},
		{"namespace_after", func() { out, _ := stripOpenAIResponsesInputNamespaces(body, false); runtime.KeepAlive(out) }},
		{"image_before", func() {
			runtime.KeepAlive(legacy024ImageIntentForPlatform("/v1/responses", "alias", body, PlatformOpenAI))
		}},
		{"image_after", func() {
			runtime.KeepAlive(IsImageGenerationIntentForPlatform("/v1/responses", "alias", body, PlatformOpenAI))
		}},
		{"schema_before", func() { out, _, _ := legacy024SanitizeSchemas(body, true); runtime.KeepAlive(out) }},
		{"schema_after", func() { out, _, _ := sanitizeOpenAIResponsesToolSchemas(body, true); runtime.KeepAlive(out) }},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(body)))
			for i := 0; i < b.N; i++ {
				tc.fn()
			}
		})
	}
}

func TestHotpath024RequestScanEquivalence(t *testing.T) {
	for _, body := range []string{
		`{"model":"alias","input":"text","stream":false}`,
		`{"model":"claude-sonnet-5-5[1m]","metadata":{"user_id":"u"},"thinking":{"type":"enabled"},"output_config":{"effort":"high"},"max_tokens":120,"system":[{"type":"text","text":"system"}],"messages":[{"role":"user","content":"hello"}]}`,
		`{"mo\u0064el":"alias","st\u0072eam":true,"metad\u0061ta":{"user_\u0069d":"u"},"input":[{"properties":{"ID":1,"id":2}}]}`,
		`{"model":"alias","metadata":"x","metadata":{"user_id":"u"},"input":[]}`,
		`{"model":"alias","metadata":{},"metadata":{"user_id":"u"},"thinking":{},"thinking":{"type":"enabled"},"output_config":{},"output_config":{"effort":"high"},"systemInstruction":{},"systemInstruction":{"parts":[{"text":"s"}]}}`,
		`{"model":"alias","metadata":null,"metadata":{"user_id":"u"},"thinking":[],"thinking":{"type":"enabled"},"output_config":null,"output_config":{"effort":"high"},"systemInstruction":[],"systemInstruction":{"parts":[{"text":"s"}]}}`,
		`{"model":"alias","metadata":{"user_id":null},"metadata":{"user_id":"u"},"thinking":{"type":""},"thinking":{"type":"enabled"},"output_config":{"effort":null},"output_config":{"effort":"high"},"systemInstruction":{"parts":null},"systemInstruction":{"parts":[{"text":"s"}]}}`,
		`{"model":"alias","contents":[{"parts":[{"text":"hi"}]}],"systemInstruction":{"parts":[{"text":"system"}]}}`,
		`{"model":"alias","stream":"true"}`,
		`{"model":false}`,
		`{"model":"alias","input":` + "\xff" + `}`,
	} {
		for _, protocol := range []string{domain.PlatformAnthropic, domain.PlatformGemini, "responses", "chat_completions"} {
			a := &ParsedRequest{Body: NewRequestBodyRef([]byte(body))}
			b := &ParsedRequest{Body: NewRequestBodyRef([]byte(body))}
			ae := legacy024ParseGatewayBody(a, protocol)
			be := parseGatewayRequestCurrentBody(b, protocol)
			require.Equal(t, ae != nil, be != nil, body)
			if ae == nil {
				require.Equal(t, a, b, body)
				validated, err := ParseGatewayRequestValidated(NewRequestBodyRef([]byte(body)), protocol)
				require.NoError(t, err)
				require.Equal(t, b, validated, body)
			}
		}
	}
}
func BenchmarkHotpath024RequestScan(b *testing.B) {
	body := hotpath024Body()
	for _, tc := range []struct {
		name string
		fn   func(*ParsedRequest, string) error
	}{{"before", legacy024ParseGatewayBody}, {"after", parseGatewayRequestCurrentBody}, {"validated", func(p *ParsedRequest, protocol string) error { return parseGatewayRequestBody(p, protocol, false) }}} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				p := &ParsedRequest{Body: NewRequestBodyRef(body)}
				if err := tc.fn(p, "responses"); err != nil {
					b.Fatal(err)
				}
				runtime.KeepAlive(p)
			}
		})
	}
}

func legacy024SanitizeSchemas(body []byte, patterns bool) ([]byte, bool, error) {
	if len(body) == 0 {
		return body, false, nil
	}

	hits := make([]openAIResponsesToolSchemaNullType, 0, 2)
	collectOpenAIResponsesToolSchemaNullTypes(body, gjson.GetBytes(body, "tools"), 0, &hits, patterns)
	if input := gjson.GetBytes(body, "input"); input.IsArray() {
		input.ForEach(func(_, item gjson.Result) bool {
			if item.IsObject() {
				collectOpenAIResponsesToolSchemaNullTypes(body, item.Get("tools"), 0, &hits, patterns)
			}
			return true
		})
	}
	if len(hits) == 0 {
		return body, false, nil
	}

	// tools 与 input 在 body 里的先后顺序由客户端决定，收集顺序不保证单调。
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].offset != hits[j].offset {
			return hits[i].offset < hits[j].offset
		}
		return hits[i].length > hits[j].length
	})
	merged := hits[:0]
	for _, hit := range hits {
		if len(merged) > 0 {
			previous := &merged[len(merged)-1]
			if hit.offset < previous.offset+previous.length {
				if hit.offset+hit.length <= previous.offset+previous.length {
					continue
				}
				if hit.replacement != "" || previous.replacement != "" {
					return nil, false, fmt.Errorf("sanitize tool schemas: overlapping patches")
				}
				previous.length = hit.offset + hit.length - previous.offset
				continue
			}
		}
		merged = append(merged, hit)
	}
	hits = merged

	sanitized := make([]byte, 0, len(body)+len(hits)*len(openAIResponsesToolSchemaFallbackType))
	cursor := 0
	for _, hit := range hits {
		// 收集阶段已逐个校验过区间，这里再挡一次重叠，保证拼接严格单调向前。
		if hit.offset < cursor {
			continue
		}
		sanitized = append(sanitized, body[cursor:hit.offset]...)
		sanitized = append(sanitized, hit.replacement...)
		cursor = hit.offset + hit.length
	}
	sanitized = append(sanitized, body[cursor:]...)
	return sanitized, true, nil
}

func setGatewayRequestRanges(parsed *ParsedRequest, protocol string, jsonStr string) {
	if parsed == nil {
		return
	}
	switch protocol {
	case domain.PlatformGemini:
		if sysParts := gjson.Get(jsonStr, "systemInstruction.parts"); sysParts.Exists() && sysParts.IsArray() {
			parsed.systemRange = rangeFromResult(sysParts)
		}
		if contents := gjson.Get(jsonStr, "contents"); contents.Exists() && contents.IsArray() {
			parsed.messagesRange = rangeFromResult(contents)
		}
	default:
		if sys := gjson.Get(jsonStr, "system"); sys.Exists() {
			parsed.HasSystem = true
			parsed.systemRange = rangeFromResult(sys)
		}
		if msgs := gjson.Get(jsonStr, "messages"); msgs.Exists() && msgs.IsArray() {
			parsed.messagesRange = rangeFromResult(msgs)
		}
		if protocol == "responses" {
			if input := gjson.Get(jsonStr, "input"); input.Exists() {
				parsed.inputRange = rangeFromResult(input)
			}
		}
	}
}

func FuzzHotpath024Gates(f *testing.F) {
	for _, value := range []string{"plain", "image_generation", `(?=x)`, `\p{L}`, "none", "😀", "ID", "id"} {
		f.Add(value)
	}
	f.Fuzz(func(t *testing.T, value string) {
		// Marshal creates legal JSON so every random value tests a real string,
		// including surrogate/invalid UTF-8 replacement and escaped punctuation.
		literal, err := json.Marshal(value)
		require.NoError(t, err)
		body := []byte(`{"model":"alias","tools":[{"type":"function","parameters":{"type":"object","pattern":` + string(literal) + `,"default":{"pattern":` + string(literal) + `},"properties":{"ID":{"type":"string"},"id":{"type":"string"}}}}],"input":[{"role":"user","content":` + string(literal) + `}]}`)
		require.Equal(t, legacy024ImageIntentForPlatform("/v1/responses", "alias", body, PlatformOpenAI), IsImageGenerationIntentForPlatform("/v1/responses", "alias", body, PlatformOpenAI))
		for _, patterns := range []bool{true, false} {
			a, ac, ae := legacy024SanitizeSchemas(body, patterns)
			b, bc, be := sanitizeOpenAIResponsesToolSchemas(body, patterns)
			require.Equal(t, ae != nil, be != nil)
			require.Equal(t, ac, bc)
			require.Equal(t, a, b)
		}
	})
}
