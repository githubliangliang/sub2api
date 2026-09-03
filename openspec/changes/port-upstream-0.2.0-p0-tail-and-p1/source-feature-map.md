# 上游改动 → 本仓库落点

只列本批 8 项。逐条的上游 diff 摘要见 `docs/upstream-sync/PORTING-0.2.0.md` §3.8–§3.11 与 §4。

## 1. `1be69e56a` → `421a83282` bootstrap 缺 call id（顺序固定）

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `handler/openai_gateway_handler.go` | 请求准入 / `call_id` 校验处 | 照搬（`1be69e56a` +232，`421a83282` +95 落在前者之上） |
| `handler/openai_delegation_bootstrap_test.go` | 同名新建 | 照搬（+176） |
| `handler/openai_automation_bootstrap_test.go` | 同名新建 | 照搬（+133） |

## 2. `504919a05` API key 会话缓存身份

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `service/openai_compat_prompt_cache_key.go` | 同名（181 行） | 照搬（1 行） |
| `service/openai_gateway_chat_completions.go` | API key 分支 | 手写（`CONFLICT`，+26-2） |
| `service/openai_agent_identity_compat_test.go` / `openai_gateway_chat_completions_test.go` | 同名 | 照搬 |

## 3. `bfe0a5a87` 终止事件前 close（含 11 行基座）

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `service/openai_ws_v2/passthrough_relay.go` | 新增 `openAIWSRelayActiveTurnID`（上游 `:1026`，11 行纯函数） | **补基座**，随本项一起落 |
| 同上 | `runUpstreamToClient` 的 `ReadFrame` 错误分支（上游 `:533`） | 手写：`graceful := isDisconnectError(err)`；`graceful && 活跃 turn != ""` ⇒ 翻成失败 + 包前缀 |
| `service/openai_ws_v2/passthrough_relay_test.go` | 同名 | 照搬（+45） |

⚠️ 上游同一 helper 在 `:785` 还有第二个调用点（`observed.responseID = ...`），**属另一条改动，不落**。
落完 grep 一遍本仓库有没有别处在手写同样的遍历。

## 4. `34b8bf1a6` **前半** Fable 兜底定价 + Fable 5.1 模型面

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `service/billing_service.go` | `initFallbackPricing`：新增 `claude-fable-5` 与 `claude-fable-5-1` 两条 | 手写**插入**（本仓库 fable 零命中，是新增不是修正） |
| 同上 | `getFallbackPricing`：fable 判定置于 opus 判定**之前**，`fable-5-1` 先于 `fable-5` | 手写 |
| `domain/constants.go` | `DefaultAntigravityModelMapping` / `DefaultBedrockModelMapping` 各 +1 | 照搬 |
| `pkg/claude/constants.go` | `DefaultModels` +1 条（`claude-fable-5-1`） | 照搬 |
| `pkg/antigravity/{claude_types,request_transformer}.go` | 模型登记 | 照搬（`CLEAN`） |
| `frontend/src/composables/useModelWhitelist.ts` | `claudeModels` / `antigravityModels` + anthropic / antigravity / bedrock 三处预设 | 照搬 |
| `frontend/src/components/{channels/SupportedModelChip,keys/UseKeyModal}.vue`、`account/*.vue`、i18n | 模型面呈现 | 照搬（`CLEAN`） |

## 5. `593fc9365` `pricing.override_file`

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `config/config.go` | `PricingConfig` 新增 `OverrideFile`（紧邻既有 `FallbackFile`，`:653`） | 照搬 |
| `service/pricing_service.go` | `applyPricingOverrides` 挂在 `parsePricingData`（`:421`）入口；`mergeOverrideOnlyModels` 在回退合并后 | 手写（`CONFLICT`，本仓库该文件与上游长期分叉） |
| `service/pricing_service_override_test.go` | 同名新建 | 照搬（+201） |
| `deploy/config.example.yaml` | 新增示例（+8） | 照搬；**另需在 `deploy/config.personal.sqlite.yaml` 里说明**（本 fork 自有文件，上游没有） |

## 6. PR#6447 → PR#6425 推理档位（顺序固定，+迁移 `225`）

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `migrations/232_group_reasoning_effort_over_limit.sql` | **新建 `migrations/225_group_reasoning_effort_over_limit.sql`** | PG → SQLite 重写：单列 `ADD COLUMN`，去 `IF NOT EXISTS` |
| `ent/schema/group.go` | 新增字段 | 手写 schema，然后 `go generate ./ent` 并提交生成物 |
| `ent/{group,mutation,runtime,migrate}*` | — | **生成物，不打补丁** |
| `internal/domain/reasoning_effort.go` | `ReasoningEffortMapping` 加按模型作用域字段（PR#6425） | 照搬 |
| `service/openai_reasoning_effort_policy.go` | 策略主体（两 PR 共 +268） | 手写（两 PR 都改这个文件） |
| `repository/group_repo.go` | 手写 SQLite SQL 读写新列 | **必然手改** |
| `repository/api_key_repo.go` | 认证投影带上新字段 | 照搬 |
| `service/{admin_group,admin_group_duplicate,admin_service,api_key_auth_cache,api_key_auth_cache_impl,group}.go` | 分组生命周期 / 复制 / 认证缓存 | 照搬（`CLEAN`） |
| `service/openai_gateway_messages{,_chat_fallback}.go`、`openai_ws_forwarder{,_ingress}.go`、`openai_ws_v2_passthrough_adapter.go` | 四条发送路径应用策略 | 部分手写（WS 三处 `CONFLICT`） |
| `handler/{admin/group_handler,openai_chat_completions,openai_gateway_handler,ops_error_logger,composite_platform}.go`、`dto/{types,mappers}.go` | API 面 | 部分手写 |
| 前端 `ReasoningEffortPolicyFields.vue`、`groupsReasoningEffort.ts`、`GroupsView.vue`、`types/index.ts`、i18n | 表单与文案 | PR#6447 多为 `CLEAN`，PR#6425 多为 `CONFLICT` ⇒ 手写 |

## 7. `34b8bf1a6` **后半** 渠道 `cache_write_1h_price`（+迁移 `226`）

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `migrations/232_channel_cache_write_1h_pricing.sql` | **新建 `migrations/226_channel_cache_write_1h_pricing.sql`** | PG → SQLite 重写：4 条 `ADD COLUMN`（去 `IF NOT EXISTS`），**整段删 4 条 `COMMENT ON COLUMN`**，说明改写成 SQL 注释（照 `178_channel_image_input_price.sql` 的 `[sqlite-converted]` 头注） |
| `service/billing_service.go` | `applyChannelTokenPriceOverrides` | 手写：`CacheWrite1hPrice == nil` 时 `cache_write_price` **继续覆盖两档**（向后兼容条款） |
| `repository/channel_repo_pricing.go` / `channel_repo_account_stats_pricing.go` | 手写 SQLite SQL 增列读写 | **必然手改** |
| `service/{channel,channel_available,channel_service,model_pricing_resolver}.go` | 结构体与解析 | 部分手写 |
| `handler/admin/channel_handler.go`、`available_channel_handler.go` | API 面 | 手写 |
| `service/account_stats_pricing.go` | 账号统计侧同名列 | 照搬（`CLEAN`） |
| 前端 `admin/channel/{IntervalRow,PricingEntryCard,types}`、`ChannelsView.vue`、`modelPlaza/PlazaModelPricingTable.vue`、`api/{admin/,}channels.ts`、i18n | 表单与展示 | 部分手写 |
| `repository/channel_repo_pricing_time_test.go` | `NOBASE` | 本仓库没这个文件 ⇒ 跳过或按本仓库形态新建 |

## 8. `1a33dc8cc` 分组模型定价弹窗布局

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `frontend/src/components/admin/channel/{IntervalRow,PricingEntryCard}.vue` | 同名 | 手写（与第 7 项同文件 ⇒ **相邻实施**） |
| `frontend/src/views/admin/GroupsView.vue` | 同名 | 手写 |
| `frontend/src/views/admin/__tests__/groupsModelsListLayout.spec.ts` | 同名 | 照搬（`CLEAN`） |

## 9. `4a1da2950` dompurify `3.3.1` → `3.4.14`（并入项，**不打 patch**）

上游那两个文件（`package.json` + `pnpm-lock.yaml`）**都不能照搬**：上游 lockfile 是 pnpm 9 产物，
本仓库本地是 pnpm 11，且本仓库多一个上游没有的 `pnpm-workspace.yaml`。所以这一项是「按结论手改」，
不是「按 diff 移植」。

| 本仓库文件 | 动作 |
|---|---|
| `frontend/package.json` | `"dompurify": "^3.3.1"` → `"^3.4.14"`；`pnpm.overrides` 加 `"dompurify@<3.4.14": ">=3.4.14"`；`devDependencies` 删 `"@types/dompurify"` |
| `frontend/pnpm-workspace.yaml` | `overrides:` 块加 `dompurify@<3.4.14: '>=3.4.14'`（**与上一项同值**，本地 pnpm 11 只读这份） |
| `frontend/pnpm-lock.yaml` | `pnpm install --lockfile-only` 重新生成（预期 churn 22 行，全在 dompurify 与 overrides 块） |
| `.github/audit-exceptions.yml` | **不动**（dompurify 从未进门禁，见 `specs/frontend-sanitizer-dependency/spec.md` 最后一条） |
| 任何 `.ts` / `.vue` | **不动**（7 个调用点一个都不改，净化行为零变化） |

预期 lockfile 结果：直接依赖的 `3.3.1` 与 `mermaid` 传递的 `3.3.3` **合成单个 `3.4.14`**。
