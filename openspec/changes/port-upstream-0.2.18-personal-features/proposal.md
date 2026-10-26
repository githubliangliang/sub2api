# 个人版可观测性与网关选定能力

## Why

使用日志缺少单请求输出速度，账号列表刷新后丢失筛选，运维请求详情需要靠账号 ID 排查。既有模型冷却、搜索模拟和协议转换已有基础，用户现明确选择补齐六项能力并先形成可验收的规格。

## What Changes

- F01：管理员、用户使用记录的延迟区域展示 `output_tokens × 1000 / duration_ms`，单位 `tok/s`；明确适用记录和无效值处理。
- F02：仅记忆账号列表的平台与状态筛选，在第一次列表请求前恢复，浏览器存储不可用时仍可筛选。
- F03：运维请求详情桌面表格、移动卡片显示账号名；缺失时回落 `#ID`，名称加载失败不阻塞明细。
- F04：临时不可调度规则支持 `account` / `model`，兼容旧规则，并保留凭据级认证错误、模型映射和换号语义。
- F05：OpenAI API Key 的 `/v1/messages` 支持沿用现有搜索服务模拟纯搜索请求，提供账号三态与可用的渠道继承配置。
- F06：增加 `gateway.min_output_tokens`，默认 128、最低 16，贯穿既有协议转换并校验 HTTP/WS 透传边界。
- 对既有 Gemini 图片隔离规格作最小范围调整，使其与本次新增规则 UI 一致。

本提案先交付上述规格与任务。2026-10-24 用户后续明确授权两个agent实现，当前按既定范围实施并补充验收证据。

## Capabilities

### New Capabilities

- `usage-output-tps`：每条使用记录的输出速率、适用性和展示偏好。
- `account-filter-persistence`：账号列表平台、状态筛选的浏览器持久化。
- `ops-request-account-label`：运维请求详情的账号名称和 ID 回退。
- `model-scoped-cooldown`：可配置的模型级/账号级临时隔离及兼容、调度边界。
- `openai-apikey-websearch-emulation`：OpenAI API Key Messages 路径的搜索模拟和配置继承。
- `gateway-min-output-tokens`：可配置转换下限及原生 HTTP/WS 兼容边界。

### Modified Capabilities

- `existing-gateway-correctness`：仅调整“F07 Gemini 工具图片与 Antigravity 图片冷却隔离”的历史 UI 范围限制，保留已实现的工具图片与图片/文本冷却隔离行为。

## Impact

涉及使用记录、账号管理、运维明细、现有搜索配置和网关转换/错误策略。预计沿用现有 JSON 凭据、账号模型限流状态及 lite 账号接口，无新数据库列、迁移或依赖；配置示例需同步个人部署模板。

来源为官方 v0.2.15–v0.2.18 的六个完整 PR，采用行为适配而非整 PR 复制。源码定位和本地缺口仅在[补充评估](../../../docs/upstream-sync/PORTING-0.2.18-personal-features.md)维护。

保留 SQLite/miniredis/simple mode、用量去重、固定配额窗口及已有 Claude Code/Codex/Grok 行为。运维 TPS 分位数、新平台、新模型、生图范围扩展、性能缓存和支付均不在本次范围。
