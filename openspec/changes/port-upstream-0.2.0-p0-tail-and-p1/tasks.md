## 0. 前置条件

- [ ] 0.1 ⚠️ **第一档（[`port-upstream-0.2.0-p0-fixes`](../port-upstream-0.2.0-p0-fixes/)）必须已合回
      `main` 并通过门禁。** 本批多处冲突只因第一档未落，提前开工会把顺序误当难度。
      **偏离**：第一档仍未合；本批从 `3da1c2dd0` 开工，见 `verification.md` 文首。
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

## 9. 阶段 9 — `4a1da2950` dompurify `3.3.1` → `3.4.14`（并入项，**必须最后做**）

**本批未做**（目标 non-goals：不改 frontend lockfile）。0.1.180 §5.1 仍挂着。

⚠️ 这一项**不是打 patch**：上游那两个文件都不能照搬（上游 lockfile 是 pnpm 9 产物，本仓库本地是
pnpm 11，且本仓库多一个上游没有的 `pnpm-workspace.yaml`）。按结论手改，逐条事实见
`source-baseline.md` §5。

- [ ] 9.1 记录升级前基线：`cd frontend && pnpm audit --json`（记 dompurify advisory 数，预期 18）
      与 `pnpm audit --prod --audit-level=high`（记门禁集合，预期 `xlsx` ×2 + `nanoid` ×1）
- [ ] 9.2 `frontend/package.json`：`"dompurify": "^3.3.1"` → `"^3.4.14"`
- [ ] 9.3 同文件 `pnpm.overrides` 加 `"dompurify@<3.4.14": ">=3.4.14"`
- [ ] 9.4 `frontend/pnpm-workspace.yaml` 的 `overrides:` 块加 `dompurify@<3.4.14: '>=3.4.14'`
- [ ] 9.5 ⚠️ **两处必须同值**：本地 pnpm 11 只读 workspace 那份，CI / Docker（pnpm 9）只读
      `package.json` 那份。只改一处 = 一边去重、另一边没去重
- [ ] 9.6 同文件 `devDependencies` 删 `"@types/dompurify"`（官方标注的 stub，dompurify 自带类型）
- [ ] 9.7 `cd frontend && pnpm install --lockfile-only` 重新生成 `pnpm-lock.yaml`；
      ⚠️ **MUST NOT 从上游抄 lockfile**
- [ ] 9.8 核对 churn：预期 **22 行**，全在 dompurify 条目与 overrides 块；
      ⚠️ 若出现其它包的版本变化 ⇒ 停下来查，不要接受
- [ ] 9.9 核对去重：lockfile 里 `dompurify` MUST 只剩一个版本，且 mermaid 那条路径
      （`. > @lobehub/icons > @lobehub/ui > mermaid > dompurify`）解析到同一版本
- [ ] 9.10 `pnpm install --frozen-lockfile` MUST 通过（5 处 CI/Docker 都用它）
- [ ] 9.11 `pnpm audit --json` 复核：dompurify advisory MUST 归零
- [ ] 9.12 ⚠️ `pnpm audit --prod --audit-level=high` 的结果 MUST 与 9.1 **相同**——
      本项 MUST NOT 被当成让门禁转绿的手段；`.github/audit-exceptions.yml` MUST 逐字节不变
- [ ] 9.13 `pnpm run test:run && pnpm run typecheck && pnpm run lint:check`；`make test-frontend-critical`
- [ ] 9.14 ⚠️ 手工看四处渲染：`/legal/:documentId`（公开页）、`/custom/:id`（含 iframe 嵌入）、
      公告弹窗/铃铛、侧栏自定义 SVG 图标。MUST 与升级前一致
- [ ] 9.15 ⚠️ MUST NOT 改任何 `.ts` / `.vue`（7 个净化调用点一个都不动）
- [ ] 9.16 三个依赖文件**一起提交**（单独提 `package.json` 会让 5 处 `--frozen-lockfile` 立刻失败）

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
- [ ] 10.8 `docs/upstream-sync/PORTING-0.1.180.md`：§5.1 标头与 §12.6 标「已合」+ 本仓库短 SHA
- [x] 10.9 `docs/upstream-sync/PORTING-0.2.0.md`：§3.8–§3.11 与 §4.1–§4.4 逐条标「已合」+ 本仓库短 SHA；
      §2.3 的迁移号表回填实际号；§7 第二/三批打勾
- [x] 10.10 `docs/upstream-sync/README.md` 顶部「当前待移植清单」段落更新本批落点与剩余项（**两处**：0.2.0 那行与 0.1.180 那行）
- [ ] 10.11 合回 `main`（`--no-ff`），合并后在 `main` 上复跑 10.1–10.4 并回填 SHA
