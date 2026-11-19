# 用户选定 GPT-6.2 Sol / Astra

2026-11-18 用户明确选择 [v0.2.23 历史评估](./PORTING-0.2.23.md) 的第三档 GPT-6.2 模型/价格/促销能力。fork 入口为 `38990cfa716e337740ae8aa3c316940368a3f9ce`，VERSION 1.1.26、Go 1.27.4、SQLite 229。详见 [OpenSpec](../../openspec/changes/port-upstream-0.2.23-gpt62/README.md) 与 [验收证据](./implementation-gpt62/README.md)。

| 模型 | 输入 | 输出 | 缓存写入 | 缓存读取 |
|---|---:|---:|---:|---:|
| gpt-6.2-sol | $2 | $10 | $2.50 | $0.10 |
| gpt-6.2-astra | $8 | $40 | $10 | $0.40 |

单位为 USD/百万 token，来源为 [PR8906 冻结 patch](./evidence-0.2.23/patches/PR8906.patch)。Fast 2 倍；Ultra Fast 在洛杉矶 2027-03-01 00:00 前 3 倍、到点起 6 倍，等于 UTC 08:00、北京时间 16:00。超过 272000 总输入时，整次输入及缓存 2 倍、输出 1.5 倍，沿用既有分组/账号/区间开关。

## 本地适配位置

- 模型身份与请求协议位于 `backend/internal/pkg/openai/constants.go`、`internal/pkg/apicompat/anthropic_to_responses.go` 和 `internal/service/openai_model_alias.go`、`openai_codex_transform.go`、`openai_compat_model.go`、`openai_compat_prompt_cache_key.go`、`openai_gateway_request_body.go`、两条 raw Chat fallback。
- 两份原始 Codex JSON 与 `openai_codex_models_service.go`、`openai_codex_model_metadata.go` 提供目录，账号声明优先，API Lite 转 full Responses。
- `backend/resources/model-pricing/model_prices_and_context_window.json`、`internal/service/pricing_service.go` 与 `billing_service.go` 提供一致价卡及按请求时刻的促销策略。
- `openai_gateway_usage.go`、`gateway_usage_billing.go` 与 `account_stats_pricing.go` 恢复并读取现有 PricingAt 上下文，复用统一费用明细，保持独立账号成本。
- `frontend/src/composables/useModelWhitelist.ts` 与 `components/keys/UseKeyModal.vue` 提供既有配置入口。

## 适配边界和核验

原始 PR 包含套餐补全，邻近 PR 包含强制分组与独立渠道倍率；这些继续独立按需，当前账号目录明确声明仍生效。本次不更改历史 intake/evidence、旧活动 change、VERSION、schema、依赖、Ent/Wire 或资源默认值。

官网访问为 web 工具 502 / pricing 页面 403，官方在线单价未独立核验。价格实施依据冻结上游价卡。实际本地 HTTP/WS/SQLite 使用合成 token、账号与定价时刻，和现网 OpenAI 请求区分记录；不外推生产性能或业务部署结果。
