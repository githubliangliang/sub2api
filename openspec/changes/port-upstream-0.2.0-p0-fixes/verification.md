# 验收证据

执行日期：2026-09-03。分支 `sync/upstream-20260902-p1`，起点 `6527a33de`
（第二档已在此分支落地；OpenSpec 写的 `3da1c2dd0` 未另开分支）。代码尖 `e9a1c427a`。

⚠️ 本批自身 **未新增迁移**。仓库最大号仍是第二档的 `226`，不是 OpenSpec 1.3 写的 `224`。
未合回 `main`（任务 8.7 按本目标 non-goals 不做）。

## 0. 基线信号（改动前，`6527a33de`）

| 门禁 | 命令 | 结果 |
|---|---|---|
| 后端编译 | `cd backend && go build ./...` | **BUILD_EXIT:0**（`{SCRATCH}/baseline-partial.log`） |
| 后端 unit | `go test -tags=unit ./... -count=1` | 开工前未再跑全量。上一批第二档在 `0af248bc8` 上 **UNIT_EXIT:0**；本批起点是其后的 docs commit `6527a33de` |
| lint | `golangci-lint run ./...` | **`0 issues.` LINT_EXIT:0** |
| SQLite 方言审计 | `go test -tags=unit ./internal/repository/ -run 'TestProductionSQLUsesSQLiteDialect' -count=1` | **ok**（与 SchedulerCache 同跑，`DIALECT_SCHED_EXIT:0`） |

已核过、可直接引用的基线事实：

- `backend/cmd/server/VERSION` = `1.1.11`
- 最大迁移号 = `226_channel_cache_write_1h_pricing.sql`（第二档已合；本批不加迁移）
- `filterSchedulerExtra` 在 `"openai_responses_supported"` 之后直接接 `"codex_5h_used_percent"`，无透传两 key
- `writeClientMessage(upstreamMessage)` 在 `openai_ws_forwarder_ingress.go:1039`（开工前）
- `tryModelFilePricing` 有手算分支，条件不含 `"fast"`

## 1. 交付后的全量信号

代码尖 `e9a1c427a`。全量写入 `{SCRATCH}/full-gates.log`。

| 门禁 | 命令 | 结果 |
|---|---|---|
| 后端编译 | `cd backend && go build ./...` | **BUILD_EXIT:0** |
| 后端 unit | `go test -tags=unit ./... -count=1` | **UNIT_EXIT:0**（无新增失败） |
| lint | `golangci-lint run ./...` | **`0 issues.` LINT_EXIT:0** |
| 迁移号未变 | `ls backend/migrations/*.sql \| sort \| tail -1` | **`226_channel_cache_write_1h_pricing.sql`**（本批 diff 无 `migrations/`） |
| 无越界改动 | `git diff --name-only 6527a33de..HEAD` | **17 个文件，全在 `backend/internal/{handler,pkg/claude,repository,service}`**。无 `frontend/` / `ent/` / `migrations/` / `VERSION` / `wire.go` / `wire_gen.go` |

## 2. 逐项证据

### 2.1 `e93e6368f` 调度快照投影 — commit `b04caf281`

**必须先复现失效再验修复**（这是本批唯一一条活缺陷）。

| 项 | 证据 |
|---|---|
| 失效复现 | 产品改动前跑 `TestBuildSchedulerMetadataAccount_KeepsOpenAIPassthroughForModelGate`：**FAIL**。`openai_passthrough` / `openai_oauth_passthrough` round-trip 后 Extra 为 nil（`{SCRATCH}/unit-snapshot-before.log`） |
| 修复 | 插入两 key 后同一测试 **ok 0.153s**（`{SCRATCH}/unit-snapshot-after.log`） |
| round-trip 覆盖 | 用例 `json.Marshal(meta)` → `json.Unmarshal` 到 `service.Account`，再断言 `IsModelSupported("gpt-5.6-sol")` |
| 兼容字段 | 子用例 `openai_oauth_passthrough` 与新字段同形 |
| 非透传不放宽 | `non_passthrough_still_uses_whitelist`：无开关 + 残留白名单 ⇒ `IsModelSupported` 仍 false |
| 白名单未越界 | `filterSchedulerExtra` 只 +2 行；`fingerprint_keys_still_dropped`：`codex_fingerprint_mode` / `codex_fingerprint_seed` 仍为 nil |

### 2.2 `200b1406d` Anthropic `fallbacks` — commit `e2d524d20`

| 项 | 证据 |
|---|---|
| 未带 beta ⇒ 剥离 | `TestSanitizeAnthropicBodyForBetaTokens_FallbacksStrippedWhenBetaMissing` / `_FallbacksStrippedWhenHeaderEmpty` |
| 带 beta ⇒ 保留 | `TestSanitizeAnthropicBodyForBetaTokens_FallbacksKeptWhenBetaPresent` |
| OAuth mimic 路径 | `TestBuildUpstreamRequest_OAuthMimicHaiku_StripsFallbacksEndToEnd` |
| Bedrock 路径 | `TestPrepareBedrockRequestBodyWithTokens_FallbacksRequireSupportedBeta`；`TestSanitizeBedrockCCFields_StripsFallbacksUnconditionally` |
| 无 `fallbacks` 时请求体不变形 | `TestSanitizeAnthropicBodyForBetaTokens_NoFallbackFieldsNoChange` / `_EmptyBodyUnchanged` |

`go test -tags=unit ./internal/service/ -run 'TestSanitizeAnthropicBodyForBetaTokens|...Fallbacks...'` ⇒ **ok 0.272s**。

### 2.3 `1dc0a0900` ctx_pool ingress — commit `8ada0bc7c`

| 项 | 证据 |
|---|---|
| 容量类改写生效 | `TestProxyResponsesWebSocketFromClient_RewritesCapacityShedCodeForClient/capacity_shed_error_and_failed_are_rewritten`：客户端帧含 `"code":"server_error"`，不含 `server_is_overloaded`。真实调用 `ProxyResponsesWebSocketFromClient` |
| 非容量类原样下发 | `non_capacity_error_code_is_passed_through` |
| `response.failed` 形态 | 同上 capacity 子用例第二条 upstream event |
| 账号状态判定用原始 payload | `openai_ws_forwarder_ingress.go`：`clientMessage := upstreamMessage`；sanitize 只写 `clientMessage`；`writeClientMessage(clientMessage)`。`upstreamMessage` 未原地改 |
| 与另两条路径语义一致 | HTTP/SSE `openai_gateway_response_handling.go:1595` 与 http_bridge `openai_ws_http_bridge.go:644` 同样调用 `sanitizeOpenAICapacityShedErrorCodeForClient` |

### 2.4 `6d5f02784` WS 池空闲回收 — commit `b9e5e6361`

| 项 | 证据 |
|---|---|
| 不可保活 + 空闲达阈值 ⇒ 逐出 | `TestOpenAIWSConnPool_RecyclesUnsupportedIdlePingConnection` PASS；`scaleDownTotal == 1`；`closedCh` 关闭 |
| 可保活连接不受影响 | `IdleRecycleKeeps.../pingable_idle_kept`：`openAIWSFakeConn` 未声明 unsafe ⇒ 视为可 ping，idle≥90s 仍留 |
| 已租出 / 有 waiter 不逐出 | `leased_unsupported_kept` / `waiter_unsupported_kept` |
| pinned 守卫仍在前 | `pinned_unsupported_kept`；产品码 `isConnPinnedLocked` 在 recycle 分支之前 |
| 指标计入 | 逐出用例断言 `scaleDownTotal`；keep 用例为 0 |
| `<90s` 不逐出 | `fresh_unsupported_kept` |

### 2.5 `ba345f105` + `57c76584a` Codex 目录 — commit `d3b90855d` / `623c555dd`

| 项 | 证据 |
|---|---|
| 持久禁用账号被跳过 | `TestBuildCodexModelsManifestForGroupIgnoresPersistentlyDisabledMappedAccounts` / `TestBuildGroupConfiguredCodexModelsManifestIgnoresPersistentlyDisabledMappedAccounts` |
| 临时不可调度不被误排除 | `TestBuildCodexModelsManifestForGroupIntersectsTransientlyUnschedulableMappedAccounts` |
| fast 模型透出 priority tier | `TestBuildCodexModelsManifestAdvertisesPriorityServiceTierForFastGPTModels`：`gpt-5.4-mini` / `gpt-5.5` / `gpt-5.6-sol` 的 `service_tiers` 含 priority/Fast |
| 非 fast 条目不变 | `TestBuildCodexModelsManifestLeavesServiceTiersEmptyForOtherModels`：`gpt-4o` / claude / deepseek-v4-pro / 自定义模型 `service_tiers` 为空 |
| 未引入发送侧 `service_tier` 行为 | `git diff 6527a33de..HEAD` 不含请求装配路径；只改 catalog descriptor |
| 平台列表 | 本 fork 用 anthropic/openai/gemini/antigravity/grok/**composite**；**未引入** Kimi/Zhipu/Deepseek 常量 |

### 2.6 `9eabd2a5b` + `e7c029875` 账号统计成本 — commit `e9a1c427a`

| 项 | 证据 |
|---|---|
| **标准档结果与改动前一致** | `TestTryModelFilePricing_Success`：100×0.001 + 50×0.002 = **0.2**（手算口径） |
| `service_tier = "fast"` | `TestTryModelFilePricing_FastUsesSharedPipeline`：结果 == `CalculateCostWithServiceTier(..., "fast")`。本 fork 该函数尚未把 fast 归一成 priority，故目前与标准档同值；统计不再有独立手算钉死标准价 |
| `priority` / `flex` 行为不变 | `TestTryModelFilePricing_AppliesServiceTierPricing`：standard 0.265 / priority 0.53 / flex 0.1325 |
| 长上下文行为不变 | `TestTryModelFilePricing_AppliesLongContextPricing` 0.233；`_CombinesPriorityAndLongContextPricing` 0.534 |
| 图片输出按 output 子集 | `TestTryModelFilePricing_WithImageOutput` 期望 **0.28**（不再 0.3） |
| 渠道定价未泄漏进优先级 3 | `tryModelFilePricing` 调 `CalculateCostWithServiceTier` 无 channel 参数；`TestResolveAccountStatsCost_Priority3IgnoresChannelCustomPricing`：渠道 9/9 自定义价不改变 0.2 |
| 未误删 `GetModelPricing` / `shouldApplySessionLongContextPricing` | 两者仍在 `billing_service.go`（`:982` / `:1459`），其它调用点还在 |

`go test -tags=unit ./internal/service/ -run 'AccountStatsCost|...TryModelFilePricing...'` ⇒ **ok 0.110s**。

## 3. 影响面复核

| 面 | 复核项 | 结果 |
|---|---|---|
| 调度 | 第 1 项会放大候选集：此前被误剔的透传账号开始参与选号。确认放大后的候选集里没有本该被其它规则挡住的账号 | 投影仍走 `filterSchedulerExtra` 白名单；非透传仍按 `model_mapping`。其它调度字段未改。透传账号若本不该接流量，用 `schedulable`，不要靠这个 bug |
| 计费 | 第 6 项只在此前被手算分支绕过的档位上产生差异，方向是修正为更高的真实成本 | 标准档 0.2 不变。图片输出从重复计费改为子集（0.3→0.28，统计成本下降）。fast 目前与标准同值（管线未映射） |
| 客户端可见行为 | 第 3 项改的是发给客户端的 payload；第 2 项改的是发给上游的请求体。两者各自的「不该改的那一侧」有反例覆盖 | 容量类改写 vs 非容量原样；fallbacks 剥离 vs 带 beta 保留；无 fallbacks 字节不变 |
| 数据库 | 无迁移、无 schema 改动 | **确认** |
| 配置 | 无新增配置项、无默认值变化 | **确认**。空闲回收阈值硬编码 90s |
| 前端 | 零改动 | **确认** |

## 4. 上线后观察

- [ ] 观察 `model_not_supported` 剔除计数是否下降（第 1 项的直接指标）
- [ ] 观察是否还有 Codex 客户端报 "Selected model is at capacity" 后直接终止（第 3 项）
- [ ] 观察 WS 池 scale-down 计数与取连接失败率（第 4 项）
- [ ] 观察带 `fallbacks` 的 Anthropic 请求是否还有 400（第 2 项）
