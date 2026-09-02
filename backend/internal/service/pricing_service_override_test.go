package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

// gpt55OverrideCatalogJSON uses explicit long_context_* fields. This fork does
// not derive those from above_272k absolute-price keys (that conversion is the
// deferred data-driven ladder cluster).
const gpt55OverrideCatalogJSON = `{
	"gpt-5.5": {"litellm_provider": "openai", "mode": "chat",
		"input_cost_per_token": 5e-06, "input_cost_per_token_priority": 1.25e-05,
		"output_cost_per_token": 3e-05, "output_cost_per_token_priority": 7.5e-05,
		"cache_read_input_token_cost": 5e-07,
		"long_context_input_token_threshold": 272000,
		"long_context_input_cost_multiplier": 2.0,
		"long_context_output_cost_multiplier": 1.5},
	"gpt-5.4": {"litellm_provider": "openai", "mode": "chat",
		"input_cost_per_token": 2.5e-06, "output_cost_per_token": 1.5e-05,
		"cache_read_input_token_cost": 2.5e-07,
		"long_context_input_token_threshold": 272000,
		"long_context_input_cost_multiplier": 2.0,
		"long_context_output_cost_multiplier": 1.5}
}`

func newPricingServiceWithOverride(t *testing.T, overrideJSON string) *PricingService {
	t.Helper()
	path := filepath.Join(t.TempDir(), "overrides.json")
	require.NoError(t, os.WriteFile(path, []byte(overrideJSON), 0644))
	svc := &PricingService{cfg: &config.Config{}}
	svc.cfg.Pricing.OverrideFile = path
	return svc
}

// override 的旗舰用例：显式 threshold=0 压住 above 折算，把目录条目的阶梯关成标准价。
func TestPricingOverride_ExplicitZeroThresholdDisablesCatalogLadder(t *testing.T) {
	svc := newPricingServiceWithOverride(t, `{"gpt-5.5": {"long_context_input_token_threshold": 0}}`)
	data, err := svc.parsePricingData([]byte(gpt55OverrideCatalogJSON))
	require.NoError(t, err)

	patched := data["gpt-5.5"]
	require.NotNil(t, patched)
	require.Zero(t, patched.LongContextInputTokenThreshold)
	require.InDelta(t, 2.0, patched.LongContextInputCostMultiplier, 1e-9, "浅合并必须保留未覆盖字段")
	require.InDelta(t, 5e-6, patched.InputCostPerToken, 1e-12, "补丁不得影响基础价")
	require.InDelta(t, 3e-5, patched.OutputCostPerToken, 1e-12)
	require.Equal(t, 272000, data["gpt-5.4"].LongContextInputTokenThreshold, "未覆盖的模型保持目录阶梯")
}

func TestPricingOverride_ZeroThresholdDisablesLadderForNonPolicyModels(t *testing.T) {
	catalog := `{
		"custom-ladder-model": {"litellm_provider": "test", "mode": "chat",
			"input_cost_per_token": 5e-06, "output_cost_per_token": 3e-05,
			"cache_read_input_token_cost": 5e-07,
			"long_context_input_token_threshold": 272000,
			"long_context_input_cost_multiplier": 2.0,
			"long_context_output_cost_multiplier": 1.5}
	}`
	svc := newPricingServiceWithOverride(t, `{"custom-ladder-model": {"long_context_input_token_threshold": 0}}`)
	data, err := svc.parsePricingData([]byte(catalog))
	require.NoError(t, err)
	require.Zero(t, data["custom-ladder-model"].LongContextInputTokenThreshold)

	svc.pricingData = data
	billing := NewBillingService(&config.Config{}, svc)
	tokens := UsageTokens{InputTokens: 300000, OutputTokens: 1000, CacheReadTokens: 10000}
	cost, err := billing.CalculateCost("custom-ladder-model", tokens, 1)
	require.NoError(t, err)
	require.False(t, cost.LongContextBillingApplied)
	require.InDelta(t, 300000*5e-6, cost.InputCost, 1e-10)
	require.InDelta(t, 1000*3e-5, cost.OutputCost, 1e-10)
	require.InDelta(t, 10000*5e-7, cost.CacheReadCost, 1e-10)
}

func TestPricingOverride_FieldLevelMergeKeepsOtherFields(t *testing.T) {
	svc := newPricingServiceWithOverride(t, `{"gpt-5.4": {"input_cost_per_token": 3e-06}}`)
	data, err := svc.parsePricingData([]byte(gpt55OverrideCatalogJSON))
	require.NoError(t, err)

	patched := data["gpt-5.4"]
	require.InDelta(t, 3e-6, patched.InputCostPerToken, 1e-12)
	require.InDelta(t, 1.5e-5, patched.OutputCostPerToken, 1e-12, "未覆盖字段保持目录值")
	require.Equal(t, "openai", patched.LiteLLMProvider)
	require.Equal(t, 272000, patched.LongContextInputTokenThreshold, "未覆盖的阶梯字段保持目录值")
	require.InDelta(t, 2.0, patched.LongContextInputCostMultiplier, 1e-9)
}

func TestPricingOverride_NullFieldValueRemovesField(t *testing.T) {
	svc := newPricingServiceWithOverride(t, `{"gpt-5.5": {
		"long_context_input_token_threshold": null,
		"long_context_input_cost_multiplier": null,
		"long_context_output_cost_multiplier": null}}`)
	data, err := svc.parsePricingData([]byte(gpt55OverrideCatalogJSON))
	require.NoError(t, err)
	require.Zero(t, data["gpt-5.5"].LongContextInputTokenThreshold, "null 删除字段后不再带阶梯")
	require.Zero(t, data["gpt-5.5"].LongContextInputCostMultiplier)
	require.InDelta(t, 5e-6, data["gpt-5.5"].InputCostPerToken, 1e-12)
}

// 完整加载管线：纯补丁不得抢在回退合并前建条目（否则回退完整条目被跳过、
// 其余分项价变 0 少收）；目录/回退都没有的模型作为独立条目并入。
func TestPricingOverride_LoadPipelineAddsNewModelAndPatchesFallbackOnly(t *testing.T) {
	dir := t.TempDir()
	catalogPath := filepath.Join(dir, "catalog.json")
	require.NoError(t, os.WriteFile(catalogPath, []byte(`{
		"remote-model": {"litellm_provider": "test", "mode": "chat",
			"input_cost_per_token": 1e-06, "output_cost_per_token": 2e-06}
	}`), 0644))
	fallbackPath := filepath.Join(dir, "fallback.json")
	require.NoError(t, os.WriteFile(fallbackPath, []byte(`{
		"fallback-only-model": {"litellm_provider": "test", "mode": "chat",
			"input_cost_per_token": 4e-06, "output_cost_per_token": 8e-06,
			"cache_read_input_token_cost": 4e-07}
	}`), 0644))
	overridePath := filepath.Join(dir, "overrides.json")
	require.NoError(t, os.WriteFile(overridePath, []byte(`{
		"fallback-only-model": {"input_cost_per_token": 9e-06},
		"override-new-model": {"litellm_provider": "test", "mode": "chat",
			"input_cost_per_token": 5e-06, "output_cost_per_token": 1e-05}
	}`), 0644))

	svc := &PricingService{cfg: &config.Config{}}
	svc.cfg.Pricing.FallbackFile = fallbackPath
	svc.cfg.Pricing.OverrideFile = overridePath
	require.NoError(t, svc.loadPricingData(catalogPath))

	patched := svc.pricingData["fallback-only-model"]
	require.NotNil(t, patched)
	require.InDelta(t, 9e-6, patched.InputCostPerToken, 1e-12)
	require.InDelta(t, 8e-6, patched.OutputCostPerToken, 1e-12, "回退条目的其余字段必须保留")
	require.InDelta(t, 4e-7, patched.CacheReadInputTokenCost, 1e-12)

	added := svc.pricingData["override-new-model"]
	require.NotNil(t, added)
	require.InDelta(t, 5e-6, added.InputCostPerToken, 1e-12)
	require.InDelta(t, 1e-5, added.OutputCostPerToken, 1e-12)

	require.InDelta(t, 1e-6, svc.pricingData["remote-model"].InputCostPerToken, 1e-12)
}

// 拼错模型名（或纯补丁落在不存在的模型上）会被有效性过滤丢弃，必须有哨兵 WARN。
func TestPricingOverride_IneffectiveEntryWarns(t *testing.T) {
	logSink, restore := captureStructuredLog(t)
	defer restore()

	dir := t.TempDir()
	catalogPath := filepath.Join(dir, "catalog.json")
	require.NoError(t, os.WriteFile(catalogPath, []byte(`{
		"remote-model": {"litellm_provider": "test", "mode": "chat", "input_cost_per_token": 1e-06}
	}`), 0644))
	overridePath := filepath.Join(dir, "overrides.json")
	require.NoError(t, os.WriteFile(overridePath, []byte(`{
		"typo-model": {"long_context_input_token_threshold": 0}
	}`), 0644))

	svc := &PricingService{cfg: &config.Config{}}
	svc.cfg.Pricing.OverrideFile = overridePath
	require.NoError(t, svc.loadPricingData(catalogPath))

	require.NotContains(t, svc.pricingData, "typo-model")
	require.True(t, logSink.ContainsMessageAtLevel("override had no effect for 1 model(s): typo-model", "warn"))
}

func TestPricingOverride_NonObjectEntryKeepsCatalogEntry(t *testing.T) {
	svc := newPricingServiceWithOverride(t, `{"gpt-5.5": "oops"}`)
	data, err := svc.parsePricingData([]byte(gpt55OverrideCatalogJSON))
	require.NoError(t, err)
	require.Equal(t, 272000, data["gpt-5.5"].LongContextInputTokenThreshold, "非法补丁忽略，目录条目原样保留")
	require.InDelta(t, 5e-6, data["gpt-5.5"].InputCostPerToken, 1e-12)
}

func TestPricingOverride_MissingOrInvalidFileIsIgnored(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		svc := &PricingService{cfg: &config.Config{}}
		svc.cfg.Pricing.OverrideFile = filepath.Join(t.TempDir(), "absent.json")
		data, err := svc.parsePricingData([]byte(gpt55OverrideCatalogJSON))
		require.NoError(t, err)
		require.Equal(t, 272000, data["gpt-5.5"].LongContextInputTokenThreshold)
	})

	t.Run("invalid json", func(t *testing.T) {
		svc := newPricingServiceWithOverride(t, `{invalid`)
		data, err := svc.parsePricingData([]byte(gpt55OverrideCatalogJSON))
		require.NoError(t, err)
		require.Equal(t, 272000, data["gpt-5.5"].LongContextInputTokenThreshold)
	})
}

// 出厂目录快照上的浅合并：覆盖单字段、其余模型不受影响。本 fork 的快照用
// above_272k 绝对价而不是 long_context_* 字段，阶梯折算属第三档，这里不测。
func TestPricingOverride_PatchesDefaultCatalogFieldWithoutTouchingOthers(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "..", "resources", "model-pricing", "model_prices_and_context_window.json"))
	require.NoError(t, err)

	svc := newPricingServiceWithOverride(t, `{"gpt-5.5": {"input_cost_per_token": 9e-06}}`)
	data, err := svc.parsePricingData(body)
	require.NoError(t, err)

	patched := data["gpt-5.5"]
	require.NotNil(t, patched)
	require.InDelta(t, 9e-6, patched.InputCostPerToken, 1e-12)
	require.InDelta(t, 3e-5, patched.OutputCostPerToken, 1e-12, "未覆盖字段保持目录值")

	untouched := data["gpt-5.4"]
	require.NotNil(t, untouched)
	require.InDelta(t, 2.5e-6, untouched.InputCostPerToken, 1e-12, "其他模型不受影响")
}

func TestPricingOverride_FallbackFileStillFillOnlyWhenBothConfigured(t *testing.T) {
	dir := t.TempDir()
	catalogPath := filepath.Join(dir, "catalog.json")
	require.NoError(t, os.WriteFile(catalogPath, []byte(`{
		"catalog-model": {"litellm_provider": "test", "mode": "chat",
			"input_cost_per_token": 1e-06, "output_cost_per_token": 2e-06}
	}`), 0644))
	fallbackPath := filepath.Join(dir, "fallback.json")
	require.NoError(t, os.WriteFile(fallbackPath, []byte(`{
		"catalog-model": {"litellm_provider": "test", "mode": "chat",
			"input_cost_per_token": 9e-06, "output_cost_per_token": 9e-06},
		"fallback-only-model": {"litellm_provider": "test", "mode": "chat",
			"input_cost_per_token": 4e-06, "output_cost_per_token": 8e-06}
	}`), 0644))
	overridePath := filepath.Join(dir, "overrides.json")
	require.NoError(t, os.WriteFile(overridePath, []byte(`{
		"catalog-model": {"input_cost_per_token": 3e-06}
	}`), 0644))

	svc := &PricingService{cfg: &config.Config{}}
	svc.cfg.Pricing.FallbackFile = fallbackPath
	svc.cfg.Pricing.OverrideFile = overridePath
	require.NoError(t, svc.loadPricingData(catalogPath))

	catalog := svc.pricingData["catalog-model"]
	require.NotNil(t, catalog)
	require.InDelta(t, 3e-6, catalog.InputCostPerToken, 1e-12, "override wins over catalog")
	require.InDelta(t, 2e-6, catalog.OutputCostPerToken, 1e-12, "fallback must not overwrite an existing catalog model")

	fallbackOnly := svc.pricingData["fallback-only-model"]
	require.NotNil(t, fallbackOnly)
	require.InDelta(t, 4e-6, fallbackOnly.InputCostPerToken, 1e-12)
}
