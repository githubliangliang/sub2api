# 实施任务与当前收尾状态

2026-10-24 复核：**125/128 项已关闭**。本次将有实现、提交或后续验收依据的 16 项旧未勾选任务核销；`9.6`、`9.14`、`10.11` 保留未完成。

`[x]` 表示已落实，或已由明确的后续批次接管关闭；接管项逐条注明依据，不表示原步骤曾在 2026-09-02 执行。当前复核见 [verification.md](./verification.md)，原实施时间、版本和来源基线保持历史值。

## 0. 前置条件

- [x] 0.1 第一档（[`port-upstream-0.2.0-p0-fixes`](../archive/2026-10-24-port-upstream-0.2.0-p0-fixes)）的当前完成状态已确认：两档均已在 `c9d9bebe8` 合回 `main`，第一档已完成归档。
      **历史偏离保留**：本批开工时第一档尚未合回，实际起点仍为 `3da1c2dd0`；本次不改写执行顺序，见 `verification.md` 的历史验收部分。
- [x] 0.2 重新量一遍本批的四态（`source-baseline.md` §2 的数字是在第一档之前量的），偏差记入该文件 §2
- [x] 0.3 起点 = 第一档的合并提交（回填 SHA），工作分支 `sync/upstream-20260902-p1`
      （实际起点 `3da1c2dd0`，分支名符合）
- [x] 0.4 记录基线信号：后端 `go build` / `go test -tags=unit ./...` / `golangci-lint run ./...`，
      前端 `pnpm run test:run` / `make test-frontend-critical`，填 `verification.md` §0
- [x] 0.5 确认 `backend/migrations/` 最大序号是 `224`；本批将顺延到 `226`，**MUST NOT 用 `227` / `228`**
      （那是第三档 Fast 组策略的号）
- [x] 0.6 确认 `backend/cmd/server/VERSION` 保持 `1.1.11`，全程不改
- [x] 0.7 全程 MUST NOT 动 `wire.go` / `wire_gen.go`；MUST NOT 手改 `ent/` 生成物
- [x] 0.8 ⚠️ 阶段 1–8 全程 MUST NOT 改 `frontend/package.json` / `pnpm-lock.yaml` /
      `pnpm-workspace.yaml` —— 这三个文件**只在阶段 9** 动，且必须同一提交（决策 11）
- [x] 0.9 全程 MUST NOT 改 `.github/audit-exceptions.yml`（决策 11）

## 1. 阶段 1 — `1be69e56a` → `421a83282` bootstrap 缺 call id（**顺序不可换**）

### 1.1 `1be69e56a`（必须先）

- [x] 1.1.1 `internal/handler/openai_gateway_handler.go`：接受无 `call_id` 的 delegation bootstrap（+232）
- [x] 1.1.2 新建 `internal/handler/openai_delegation_bootstrap_test.go`（照搬 +176）
- [x] 1.1.3 ⚠️ 普通工具调用的 `call_id` 校验 MUST NOT 放宽——补一条反例用例
- [x] 1.1.4 `go test -tags=unit ./internal/handler/ -run 'DelegationBootstrap' -count=1`
- [x] 1.1.5 独立提交

### 1.2 `421a83282`（在 1.1 之后）

- [x] 1.2.1 同文件：接受无 `call_id` 的 scheduled-automation bootstrap（+95）
- [x] 1.2.2 新建 `internal/handler/openai_automation_bootstrap_test.go`（照搬 +133）
- [x] 1.2.3 `go test -tags=unit ./internal/handler/ -run 'AutomationBootstrap|DelegationBootstrap' -count=1`
- [x] 1.2.4 独立提交

## 2. 阶段 2 — `504919a05` API key 会话缓存身份

- [x] 2.1 `internal/service/openai_compat_prompt_cache_key.go`：照搬那 1 行
- [x] 2.2 `internal/service/openai_gateway_chat_completions.go`：手写落地 API key 分支（+26-2，`CONFLICT`）
- [x] 2.3 落地两个用例文件（`openai_agent_identity_compat_test.go` 新建 +46、
      `openai_gateway_chat_completions_test.go` +89）
- [x] 2.4 ⚠️ 验收必须含「不同会话不串号」的反例
- [x] 2.5 `go test -tags=unit ./internal/service/ -run 'PromptCacheKey|AgentIdentity|ChatCompletions' -count=1`
- [x] 2.6 独立提交

## 3. 阶段 3 — `bfe0a5a87` 终止事件前 close（**基座与主体同批**）

- [x] 3.1 `internal/service/openai_ws_v2/passthrough_relay.go`：新增 11 行
      `openAIWSRelayActiveTurnID`（遍历 `state.turnTimingByID` 找 `== state.activeTurn`；
      state / activeTurn 为 nil ⇒ 返回 ""）
- [x] 3.2 同文件 `runUpstreamToClient` 的 `ReadFrame` 错误分支：`graceful := isDisconnectError(err)`；
      `graceful && openAIWSRelayActiveTurnID(state) != ""` ⇒ `graceful = false` 并包错误前缀；
      追踪事件与退出信号都用同一个 `graceful`
- [x] 3.3 ⚠️ MUST NOT 落上游 `:785` 那第二个调用点（见 `design.md` 决策 2）
- [x] 3.4 `grep -rn 'turnTimingByID' backend/internal/` 确认没有别处在手写同样的遍历
- [x] 3.5 落地上游用例 `passthrough_relay_test.go`（+45）
- [x] 3.6 ⚠️ 验收必须双侧：turn 活跃时 close ⇒ 失败；**无活跃 turn 时 close ⇒ 仍 graceful**
- [x] 3.7 `go test -tags=unit ./internal/service/openai_ws_v2/ -count=1`
- [x] 3.8 基座与主体**一起提交**（单独提基座是死代码）

## 4. 阶段 4 — `34b8bf1a6` **前半** Fable 兜底定价 + Fable 5.1 模型面

- [x] 4.1 `internal/service/billing_service.go` `initFallbackPricing`：**插入** `claude-fable-5` 与
      `claude-fable-5-1` 两条（in $10 / out $50 / cache write 5m $12.5 / 1h $20 per MTok；
      cache read 分别 $1 与 $0.25；两条都 `SupportsCacheBreakdown: true`）
- [x] 4.2 同文件 `getFallbackPricing`：fable 判定插在 **opus 判定之前**，`fable-5-1` 系列先于 `fable-5`
- [x] 4.3 ⚠️ 这是**新增**不是修正（本仓库 grep `fable` 零命中）⇒ 不要指望能 apply 上去
- [x] 4.4 `internal/domain/constants.go`：`DefaultAntigravityModelMapping` / `DefaultBedrockModelMapping`
      各加一行 `claude-fable-5-1`
- [x] 4.5 `internal/pkg/claude/constants.go`：`DefaultModels` 加 `claude-fable-5-1` 条目
- [x] 4.6 `internal/pkg/antigravity/{claude_types,request_transformer}.go`：照搬
- [x] 4.7 前端 `composables/useModelWhitelist.ts`：`claudeModels` / `antigravityModels` 与
      anthropic / antigravity / bedrock 三处预设各加一行
- [x] 4.8 前端其余照搬项（`SupportedModelChip.vue`、`UseKeyModal.vue`、`account/*.vue`、i18n）
- [x] 4.9 ⚠️ **不要**在本阶段碰任何 `cache_write_1h` 相关文件——那是阶段 7（`design.md` 决策 4）
- [x] 4.10 落地上游用例（`billing_service_test.go`、`antigravity_model_mapping_test.go`、
      `bedrock_request_test.go`、`constants_model_test.go`、`useModelWhitelist.spec.ts`、
      `PlazaModelPricingTable.spec.ts`、`UseKeyModal.spec.ts`）
- [x] 4.11 ⚠️ 必须验证「opus / sonnet / haiku / gemini 的兜底命中与改动前完全一致」
- [x] 4.12 `go test -tags=unit ./internal/service/ -run 'FallbackPricing' -count=1`；
      `cd frontend && pnpm run test:run`
- [x] 4.13 独立提交

## 5. 阶段 5 — `593fc9365` `pricing.override_file`

- [x] 5.1 `internal/config/config.go`：`PricingConfig` 新增 `OverrideFile`（紧邻 `FallbackFile`，`:653`），
      默认空
- [x] 5.2 `internal/service/pricing_service.go`：`applyPricingOverrides` 挂在 `parsePricingData`（`:421`）
      入口；只修补已存在条目（同名字段覆盖，`null` 删字段）
- [x] 5.3 同文件：`mergeOverrideOnlyModels` 在**回退合并之后**并入目录与回退都没有的模型；
      ⚠️ MUST NOT 在主解析阶段抢先建条目（否则纯补丁会挡住回退的完整条目、其余分项价静默变 0）
- [x] 5.4 同文件：最终未生效的条目打 WARN；override 文件缺失/损坏只跳过合并
- [x] 5.5 新建 `internal/service/pricing_service_override_test.go`（照搬 +201）
- [x] 5.6 `deploy/config.example.yaml` 加示例（+8）
- [x] 5.7 ⚠️ 本 fork 自有：在 `deploy/config.personal.sqlite.yaml` 里也加一条注释说明（上游没这个文件）
- [x] 5.8 ⚠️ `fallback_file` 的「只补缺失模型」语义 MUST NOT 改变——补一条两文件同配的用例
- [x] 5.9 `go test -tags=unit ./internal/service/ -run 'PricingOverride|Pricing' -count=1`
- [x] 5.10 独立提交

## 6. 阶段 6 — PR#6447 → PR#6425 推理档位（**顺序不可换，+迁移 `225`**）

### 6.1 PR#6447 超限拒绝 / 降级（必须先）

- [x] 6.1.1 取净效果：`git diff 559960865^1 559960865`（**不要**逐条 cherry-pick 分支微提交）
- [x] 6.1.2 新建 `backend/migrations/225_group_reasoning_effort_over_limit.sql`：单列 `ADD COLUMN`，
      **去 `IF NOT EXISTS`**，无 `COMMENT ON`；⚠️ **`DEFAULT` 必须落在与现状等价的「降级」那一档**
      （`design.md` 决策 5）
- [x] 6.1.3 ⚠️ 迁移文件注释里 MUST NOT 出现 PG 语法字面量（0.1.184 §25.2）
- [x] 6.1.4 在 `backend/migrations/` 加 `TestMigration225...WithSQLiteSyntax` 用例
      （样板：`TestMigration151AddsAccountAutoPauseExpiryPartialIndex`）
- [x] 6.1.5 `backend/ent/schema/group.go` 加字段 → `cd backend && go generate ./ent`，**提交生成物**；
      ⚠️ MUST NOT 打 `ent/runtime/runtime.go` 等生成物的补丁
- [x] 6.1.6 `internal/service/openai_reasoning_effort_policy.go`：策略主体（+117-x）
- [x] 6.1.7 `internal/repository/group_repo.go`：手写 SQLite SQL 读写新列（**必然手改**）
- [x] 6.1.8 `internal/repository/api_key_repo.go`：认证投影带上新字段
- [x] 6.1.9 `internal/service/` 的 `admin_group.go` / `admin_group_duplicate.go` / `admin_service.go` /
      `api_key_auth_cache{,_impl}.go` / `group.go`（多为 `CLEAN`）
- [x] 6.1.10 四条发送路径：`openai_gateway_messages{,_chat_fallback}.go`、
      `openai_ws_forwarder{,_ingress}.go`、`openai_ws_v2_passthrough_adapter.go`（WS 三处手写）
- [x] 6.1.11 `internal/handler/` 的 `admin/group_handler.go` / `openai_chat_completions.go` /
      `openai_gateway_handler.go` / `ops_error_logger.go` / `composite_platform.go` / `dto/{types,mappers}.go`
- [x] 6.1.12 ⚠️ 本阶段若碰到任何 `_integration_test.go`，落地前先 `head -1` 看构建标签；
      `//go:build integration && postgres` 的文件在本仓库永远不编译 ⇒ **整段 hunk 丢弃**，
      MUST NOT 改构建标签（决策 9）
- [x] 6.1.13 前端：`ReasoningEffortPolicyFields.vue`、`groupsReasoningEffort.ts`、`GroupsView.vue`、
      `types/index.ts`、zh/en i18n
- [x] 6.1.14 落地后端与前端用例（含 `api_contract_test.go`、`admin_service_*_test.go`、
      `groupsReasoningEffort.spec.ts`）
- [x] 6.1.15 ⚠️ **必须验证「迁移后未改配置的存量分组行为不变」**（决策 5 / spec 的存量分组 Scenario）
- [x] 6.1.16 `go test -tags=unit ./internal/... -run 'ReasoningEffort' -count=1`；
      `go test ./migrations/ -count=1`；`cd frontend && pnpm run test:run`
- [x] 6.1.17 独立提交

### 6.2 PR#6425 按模型限定映射（在 6.1 之后）

- [x] 6.2.1 取净效果：`git diff 3510aa22b^1 3510aa22b`；⚠️ **不合** `77729e272` 与 `05ea883e2`
      （上游解冲突提交与生成物）
- [x] 6.2.2 `internal/domain/reasoning_effort.go`：`ReasoningEffortMapping` 加按模型作用域字段
- [x] 6.2.3 `ent/schema/group.go` 若有形变 → `go generate ./ent`
- [x] 6.2.4 `internal/service/openai_reasoning_effort_policy.go`：+151，手写（与 6.1 同文件）
- [x] 6.2.5 `internal/handler/admin/group_handler.go`、`dto/types.go`
- [x] 6.2.6 前端：`ReasoningEffortPolicyFields.vue`（+226）、`groupsReasoningEffort.ts`（+188）、
      `types/index.ts`、zh/en i18n
- [x] 6.2.7 落地用例（`ReasoningEffortPolicyFields.spec.ts`、`groupsReasoningEffort.spec.ts`、
      `group_handler_reasoning_effort_test.go`、`admin_service_group_test.go`）
- [x] 6.2.8 ⚠️ 必须验证「未限定模型的存量映射仍对全部模型生效」
- [x] 6.2.9 同 6.1.16 的三组命令
- [x] 6.2.10 独立提交

## 7. 阶段 7 — `34b8bf1a6` **后半** 渠道 `cache_write_1h_price`（**+迁移 `226`**）

- [x] 7.1 新建 `backend/migrations/226_channel_cache_write_1h_pricing.sql`：4 条
      `ALTER TABLE ... ADD COLUMN cache_write_1h_price NUMERIC(20,12);`（`channel_model_pricing`、
      `channel_pricing_intervals`、`channel_account_stats_model_pricing`、
      `channel_account_stats_pricing_intervals`）
- [x] 7.2 ⚠️ **去掉 `IF NOT EXISTS`**、**整段删 4 条 `COMMENT ON COLUMN`**，说明改写成文件头 SQL 注释
      （照 `178_channel_image_input_price.sql` 的 `[sqlite-converted]` 头注）
- [x] 7.3 ⚠️ 注释里 MUST NOT 出现 PG 语法字面量
- [x] 7.4 加 `TestMigration226...` 用例
- [x] 7.5 `internal/service/billing_service.go` `applyChannelTokenPriceOverrides`：
      ⚠️ **`CacheWrite1hPrice == nil` 时 `cache_write_price` 继续覆盖两档**（决策 7，逐字保留上游那个
      `if` 分支，MUST NOT 简化成无条件赋值）
- [x] 7.6 `internal/repository/channel_repo_pricing.go` + `channel_repo_account_stats_pricing.go`：
      手写 SQLite SQL 增列读写
- [x] 7.7 `internal/service/` 的 `channel.go` / `channel_available.go` / `channel_service.go` /
      `model_pricing_resolver.go` / `account_stats_pricing.go`
- [x] 7.8 `internal/handler/admin/channel_handler.go` + `available_channel_handler.go`
- [x] 7.9 前端：`admin/channel/{IntervalRow,PricingEntryCard,types}`、`ChannelsView.vue`、
      `modelPlaza/PlazaModelPricingTable.vue`、`api/{admin/,}channels.ts`、zh/en i18n
- [x] 7.10 ⚠️ `channel_repo_pricing_time_test.go` 是 `NOBASE`——按本仓库形态新建或跳过，
      落地前 `head -1` 看构建标签
- [x] 7.11 ⚠️ 验收必须覆盖四种组合：只配 5m / 只配 1h / 都配 / 都不配
- [x] 7.12 删掉本地 `*.db` 完整起一次，确认 `225` 与 `226` 都真跑过，四张表都有新列
- [x] 7.13 `go test ./migrations/ -count=1`；
      `go test -tags=unit ./internal/repository/ -run 'Migration|TestProductionSQLUsesSQLiteDialect' -count=1`；
      `go test -tags=unit ./internal/service/ -run 'ChannelPricing|CacheWrite' -count=1`；
      `cd frontend && pnpm run test:run`
- [x] 7.14 独立提交

## 8. 阶段 8 — `1a33dc8cc` 分组模型定价弹窗布局（紧随阶段 7）

- [x] 8.1 `frontend/src/components/admin/channel/{IntervalRow,PricingEntryCard}.vue`：手写（与阶段 7 同文件）
- [x] 8.2 `frontend/src/views/admin/GroupsView.vue`：手写
- [x] 8.3 落地 `views/admin/__tests__/groupsModelsListLayout.spec.ts`（照搬）
- [x] 8.4 `cd frontend && pnpm run test:run && pnpm run typecheck && pnpm run lint:check`
- [x] 8.5 独立提交

## 9. 阶段 9 — DOMPurify 维护（已由 0.2.16 第二档 S16 接管）

2026-09-02 本批确实未改依赖；该历史事实保持不变。旧 `3.4.14` 计划已由后续 S16 的 `3.4.17` 定向维护替代，提交为 `5ff6e5bf9`，已在当前 `main` 历史中。依据见 [S16 依赖报告](../../../docs/upstream-sync/implementation-0.2.16-tier2/dependency-review.md)及[验收](../archive/2026-10-24-port-upstream-0.2.16-tier2/verification.md)。

- [x] 9.1 升级前后依赖审计由 S16 接管并留存原始 JSON：全树 103→70 条记录，DOMPurify 19→0。原 0.2.0 的预期 18 条及旧门禁集合仅是当时基线，不补造执行记录。
- [x] 9.2 `frontend/package.json` 的直接依赖已由 S16 升为 `"dompurify": "^3.4.17"`，满足原计划的最低版本要求。
- [x] 9.3 `package.json.pnpm.overrides` 已固定 `dompurify: 3.4.17`，统一覆盖直接与传递解析。
- [x] 9.4 `frontend/pnpm-workspace.yaml` 的 `overrides` 同样固定 `dompurify: 3.4.17`。
- [x] 9.5 两处 overrides 当前同值；pnpm 11 本地解析及 S16 的 pnpm 9 隔离验证均有记录。
- [ ] 9.6 移除 `devDependencies` 中的 `"@types/dompurify"` 类型桩依赖。当前仍声明 `^3.0.5`、安装为 `3.2.0`；此独立清理尚未执行，不影响已完成的 DOMPurify 单版本升级。
- [x] 9.7 S16 已使用本地 pnpm 11.21.0 生成锁文件，没有整取上游 lockfile；生成及安装记录见 S16 依赖报告。
- [x] 9.8 **由后续范围替代关闭**：原“仅 DOMPurify、22 行 churn”约束属于旧单包计划；S16 已明确接管 DOMPurify、axios、Vue、source-map-js 及其必要依赖闭包，并记录实际变更范围，不将其描述为原 22 行操作已执行。
- [x] 9.9 当前 lockfile 及 `pnpm why dompurify --json` 都只解析到 `3.4.17`；直接依赖、类型桩依赖与 mermaid 传递路径使用同一版本。
- [x] 9.10 S16 的隔离 pnpm 9.15.9 `install --lockfile-only --frozen-lockfile --ignore-scripts` 检查通过，锁文件 SHA256 不变；据此关闭锁文件与 manifest 一致性检查，不宣称本次重装了依赖。
- [x] 9.11 本次 `pnpm audit --json` 复核 DOMPurify 告警为 0；全树仍有其它包的 75 条记录、命令返回 1，不记为审计全绿。
- [x] 9.12 **由后续范围替代关闭**：旧固定门禁集合及“例外文件逐字节不变”约束由 S16 后续维护替代；六条过期例外已移除，未续期或新增豁免，生产 xlsx 剩余问题在 S16 报告单列。本次没有调整审计门禁或例外。
- [x] 9.13 S16 已完成前端全量 Vitest（2027 passed / 2 skipped）、typecheck、lint、build；本次定向 10 文件 / 98 测试通过，并补跑 `make test-frontend-critical`（14 文件 / 176 passed / 2 skipped）。后续验收覆盖与本次复跑分别记于 `verification.md`。
- [ ] 9.14 补齐四处真实浏览器人工验收记录：`/legal/:documentId`（公开页）、`/custom/:id`（含 iframe 嵌入）、公告弹窗/铃铛、侧栏自定义 SVG 图标。现有真实 DOMPurify 的 jsdom 回归通过，不能替代这四处渲染验收。
- [x] 9.15 `5ff6e5bf9` 仅修改三个依赖文件与 `.github/audit-exceptions.yml`；没有修改 `.ts` / `.vue` 净化调用点。
- [x] 9.16 `frontend/package.json`、`pnpm-lock.yaml`、`pnpm-workspace.yaml` 已在 `5ff6e5bf9` 同一提交内更新。

## 10. 收尾

- [x] 10.1 `cd backend && go build ./... && go test -tags=unit ./... -count=1 && golangci-lint run ./...`
- [x] 10.2 `go test ./migrations/ -count=1`
- [x] 10.3 `cd frontend && pnpm run test:run && pnpm run typecheck && pnpm run lint:check`
      （全量 vitest 仅既有 `useRoutePrefetch.spec.ts` 5 条失败）
- [x] 10.4 `make test-frontend-critical`
- [x] 10.5 确认 `backend/migrations/` 最大序号是 `226`，且 `VERSION` 仍是 `1.1.11`
- [x] 10.6 确认 `git diff --stat` 里没有 `wire_gen.go`、`.github/audit-exceptions.yml`；
      `package.json` / `pnpm-lock.yaml` / `pnpm-workspace.yaml` **应当有且仅有阶段 9 的那次改动**
- [x] 10.7 填 `verification.md`（含 §0 基线、§2 逐项证据、§3 影响面复核）
- [x] 10.8 `docs/upstream-sync/PORTING-0.1.180.md` 的 §5.1 与 §12.6 已补记“由 0.2.16 第二档 S16 / `5ff6e5bf9` 接管完成”；保留旧版本评估、未合状态与来源作为历史记录。
- [x] 10.9 `docs/upstream-sync/PORTING-0.2.0.md`：§3.8–§3.11 与 §4.1–§4.4 逐条标「已合」+ 本仓库短 SHA；
      §2.3 的迁移号表回填实际号；§7 第二/三批打勾
- [x] 10.10 `docs/upstream-sync/README.md` 顶部「当前待移植清单」段落更新本批落点与剩余项（**两处**：0.2.0 那行与 0.1.180 那行）
- [ ] 10.11 合回 `main` 部分已完成（双亲合并提交 `c9d9bebe8`）。仍待补齐本项要求的合并后 `main` 全量 10.1–10.4 门禁及 SHA 回填；本次仅定向后端、前端与关键套件复跑，不能记作全量复验完成。
