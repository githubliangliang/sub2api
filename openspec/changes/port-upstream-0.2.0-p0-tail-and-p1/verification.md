# 验收证据

执行日期：2026-09-02。分支 `sync/upstream-20260902-p1`，起点 `3da1c2dd0`，
代码尖 `0af248bc8`。

⚠️ **偏离**：第一档 [`port-upstream-0.2.0-p0-fixes`](../port-upstream-0.2.0-p0-fixes/)
仍未合回 `main`。本批从 `3da1c2dd0` 开工，WS/handler 上多出来的冲突按 hunk 手解。
第 9 项（dompurify `3.3.1` → `3.4.14`）按本目标 **未做**（不改 lockfile）。

## 0. 基线信号（改动前，起点 = `3da1c2dd0`，不是第一档合并提交）

| 门禁 | 命令 | 结果 |
|---|---|---|
| 后端编译 | `cd backend && go build ./...` | 未在开工前单独落盘。`3da1c2dd0` 是已合 0.1.183/0.1.184 的 `1.1.11` 尖；本批每一阶段 `go build` 均通过 |
| 后端 unit | `go test -tags=unit ./... -count=1` | 开工前全量未落盘（`baseline-build.log` 空）。第一档 verification 已核过的切片：`SchedulerCache\|TestProductionSQLUsesSQLiteDialect` ok；`AccountStatsCost\|FallbackPricing` ok |
| 迁移包 | `go test ./migrations/ -count=1` | 开工前最大号 `224_*`；本批新增 `225`/`226` 后仍绿（见 §1） |
| lint | `golangci-lint run ./...` | 开工前未落盘。第一轮全量在 `e79c15c6f` 后因 3 处 gofmt 失败，已在 `0af248bc8` 修掉 |
| 前端全量 | `cd frontend && pnpm run test:run` | 开工前未落盘。既有失败预期含 `useRoutePrefetch.spec.ts`（与 PORTING-0.1.179 §4.6 / 更早 change 一致） |
| 前端关键 | `make test-frontend-critical` | 开工前未落盘；`e79c15c6f` 后一次为 **14 files / 168 passed / 2 skipped**，`FRONT_CRIT_EXIT:0` |
| 迁移号 | `ls backend/migrations/ \| sort \| tail -1` | **`224_sqlite_redundant_index_cleanup.sql`** |
| 依赖审计（全量） | `cd frontend && pnpm audit --json` | 未跑（第 9 项未做）。基线事实：dompurify 直接 `^3.3.1` + mermaid 传递 `3.3.3`，全树 72 条 advisory 里 dompurify 占 18 |
| 依赖审计（门禁） | `pnpm audit --prod --audit-level=high` | 未跑（第 9 项未做）。预期 `xlsx` ×2 + `nanoid` ×1 |

已核过、可直接引用的基线事实（2026-09-02，`3da1c2dd0`）：

- `internal/service/billing_service.go` grep `fable` ⇒ **零命中**（阶段 4 是新增而非修正）
- `openAIWSRelayActiveTurnID` 全仓 ⇒ **零命中**（阶段 3 需补 11 行基座）
- `internal/repository/migrations_schema_integration_test.go` 首行 ⇒ `//go:build integration && postgres`（本批未改该文件）
- `channel_model_pricing` 等四张表存在；`cache_write_1h_price` 列不存在
- `internal/config/config.go:653` 有 `FallbackFile`，无 `OverrideFile`
- 推理档位基座齐：`internal/domain/reasoning_effort.go:5`、
  `internal/service/openai_reasoning_effort_policy.go`、`ent/schema/group.go`、
  `components/admin/group/ReasoningEffortPolicyFields.vue`、`views/admin/groupsReasoningEffort.ts`
- `backend/cmd/server/VERSION` = `1.1.11`

## 1. 交付后的全量信号

代码尖 `0af248bc8`。第一轮全量（`e79c15c6f`/`ff897d4d8`）`BUILD_EXIT:0`、`FRONT_CRIT_EXIT:0`，
但 `UNIT_EXIT:1`（`TestClassifyOpsLocalBusinessLimitErrorsExcludedFromSLA/group_reasoning_effort_over_limit_deny`
期望 `permission_error`、本 fork 分类为 `api_error`）、`LINT_EXIT:1`（3 处 gofmt）、
`FRONT_TYPE_LINT_EXIT:2`（`ChannelsView.vue:886` 缺 `cache_write_1h_price`）。
上述三项已在 `0af248bc8` 修掉。修后全量写入 `{SCRATCH}/full-gates.log`：

| 门禁 | 命令 | 结果 |
|---|---|---|
| 后端编译 | `cd backend && go build ./...` | **BUILD_EXIT:0** |
| 后端 unit | `go test -tags=unit ./... -count=1` | **UNIT_EXIT:0**（含 `internal/handler` 与 `internal/service` 155s / `openai_ws_v2` 3.016s / `migrations`）。无新增失败 |
| 迁移包 | `go test ./migrations/ -count=1` | **MIG_EXIT:0**（ok 0.004s） |
| 方言审计 | `go test -tags=unit ./internal/repository/ -run 'TestProductionSQLUsesSQLiteDialect' -count=1` | **DIALECT_EXIT:0**（ok 0.922s） |
| lint | `golangci-lint run ./...` | **`0 issues.` LINT_EXIT:0** |
| 前端全量 | `cd frontend && pnpm run test:run` | **FRONT_FULL_EXIT:1**：234 files 里 **233 passed / 1 failed**；1656 tests 里 **1649 passed / 5 failed / 2 skipped**。失败全是既有 `useRoutePrefetch.spec.ts`（`/admin/dashboard` 映射，PORTING-0.1.179 §4.6），与本批无关 |
| 前端类型 / lint | `pnpm run typecheck && pnpm run lint:check` | **FRONT_TYPE_LINT_EXIT:0** |
| 前端关键 | `make test-frontend-critical` | **14 files / 168 passed / 2 skipped**，`FRONT_CRIT_EXIT:0` |
| 空库全量迁移 | 删 `*.db` 后完整启动 | 真 `cmd/server` 因定价下载 TLS timeout 被 `timeout 124` 杀掉（`{SCRATCH}/server-boot.log`），DB 已写出。验收改走 shipped runner：`TestMigrations225And226ApplyOnFreshSQLite/{db1,db2}` **PASS**（0.67s / 0.66s，`{SCRATCH}/migrate-1.log` / `migrate-2.log`）。`schema_migrations` 含 `225`/`226`；`groups.max_reasoning_effort_over_limit` 与四张渠道定价表的 `cache_write_1h_price` 都在 |
| dompurify 单副本 | `grep -c 'dompurify@' frontend/pnpm-lock.yaml` | **本批未做**（不改 lockfile） |
| frozen-lockfile | `cd frontend && pnpm install --frozen-lockfile` | **本批未做**（lockfile 未改） |
| 门禁未被改动 | `git diff -- .github/audit-exceptions.yml` | **空**（`3da1c2dd0..0af248bc8` 无该文件） |
| 无越界改动 | `git diff --name-only 3da1c2dd0..HEAD` | **无** `wire.go` / `wire_gen.go` / `package.json` / `pnpm-lock.yaml` / `pnpm-workspace.yaml` / `.github/audit-exceptions.yml` / postgres-gated `_integration_test.go`。`VERSION` 仍是 `1.1.11`。最大迁移号 `226_channel_cache_write_1h_pricing.sql` |

## 2. 逐项证据

### 2.1 `1be69e56a` → `421a83282` bootstrap — commit `f3df9e86b` / `1c38b6800`

| 项 | 证据 |
|---|---|
| 失效复现（delegation） | 基线无 `call_id` 的 `function_call_output` 走普通校验，返回 400 `function_call_output requires call_id`。新用例 `ordinary_tool-call_missing_call_id_is_still_rejected` 锁住这条 |
| 修复（delegation） | `TestCodexDelegationBootstrapAdmission/missing call_id is accepted and not given a synthetic call_id`：accepted、body 无合成 `call_id`。`{SCRATCH}/unit-new-summary.log` PASS |
| 修复（scheduled automation） | `TestCodexAutomationBootstrapAdmission/missing call_id is accepted and not given a synthetic call_id` PASS |
| 带 `call_id` 时原样透传 | 两个 Admission 测试的 `existing_call_id_is_passed_through_unchanged`：`call-1` 原样保留 |
| **普通工具调用仍校验** | `ordinary_tool-call_missing_call_id_is_still_rejected`（delegation + automation）+ `client-supplied_bootstrap_flag_does_not_relax_ordinary_tool-call_admission`：400，admission 看请求形态不是客户端 flag |
| 顺序 | `f3df9e86b`（delegation）在前，`1c38b6800`（scheduled-automation）在后，两条独立提交 |

### 2.2 `504919a05` API key 会话缓存身份 — commit `3bc098ee2`

| 项 | 证据 |
|---|---|
| 同会话身份一致 | `TestForwardAsChatCompletions_APIKeyAutoDerivesStableIsolatedPromptCacheKey`：同一 api_key 追加轮次 `prompt_cache_key` 与 `session_id` 不变。`TestDeriveCompatPromptCacheKey_StableAcrossLaterTurns` |
| 不同会话不串号 | 同上：不同 api_key / 不同 messages 会话 ⇒ key 与 `session_id` 都不等。`TestDeriveCompatPromptCacheKey_DiffersAcrossSessions` |
| 无身份线索时退回原行为 | `TestForwardAsChatCompletions_ResponsesShapeDoesNotAutoDerivePromptCacheKey`：Responses 形态且未给 key ⇒ 不发明 `prompt_cache_key` / `session_id`；显式 key 仍按 `isolateOpenAISessionID` 投影 |

`go test -tags=unit ./internal/service/ -run 'PromptCacheKey\|AgentIdentity\|ChatCompletions'` ⇒ ok（`{SCRATCH}/unit-new-summary.log` stage2）。

### 2.3 `bfe0a5a87` + 基座 — commit `527986900`

| 项 | 证据 |
|---|---|
| 基座是纯函数（nil 安全） | `TestOpenAIWSRelayActiveTurnID`：nil / 空 state / 无 activeTurn ⇒ `""`；有 activeTurn ⇒ 对应 responseID |
| turn 活跃 + close 1000 ⇒ 失败 | `TestRelay_UpstreamNormalCloseBeforeTerminalIsFailure`：`Graceful=false`，错误含 `upstream websocket closed before terminal event`；追踪 `read_upstream_failed.Graceful` 与退出信号同值 |
| turn 活跃 + EOF ⇒ 失败 | `TestRelay_UpstreamEOFBeforeTerminalIsFailure`：同上前缀，非 graceful |
| **无活跃 turn + close ⇒ 仍 graceful** | `TestRelay_UpstreamDisconnect`：无 `response.created` 的上游 EOF 仍 graceful（`relayExit==nil`）。产品条件是 `graceful && openAIWSRelayActiveTurnID(state) != ""` |
| 非干净断开行为不变 | `TestRelay_DirtyUpstreamDisconnectUnchanged`：仍失败，错误是 `tls handshake timeout`，**不含** terminal-event 前缀 |
| 追踪事件与退出信号同值 | `TestRelay_UpstreamNormalCloseBeforeTerminalIsFailure` 断言 `relayExit.Graceful == readFailed.Graceful` |
| 未落上游第二个调用点 | `grep observed.responseID = openAIWSRelayActiveTurnID` **零命中**。helper 只在 `runUpstreamToClient` 的 ReadFrame 错误分支使用 |
| 无重复实现 | `turnTimingByID` 仅 `passthrough_relay.go` 的 map 维护 + helper 遍历；测试文件引用 helper，无第二套手写扫描 |

`go test -tags=unit ./internal/service/openai_ws_v2/ -count=1` ⇒ **ok 3.015s**。

### 2.4 `34b8bf1a6` 前半 Fable — commit `6bd7637c4`

| 项 | 证据 |
|---|---|
| 目录缺条目时落到兜底价（非 0、非别系列） | `TestGetFallbackPricing_FableCatalogMissVsHit`：`claude-fable-5` in $10/MTok，≠ 0，≠ opus fallback |
| 目录有条目时优先目录价 | 同上：注入 catalog 99e-6 / 88e-6 后 `GetModelPricing` 命中目录 |
| 5.1 缓存读取成本低于 5 | 同上 + `TestGetModelPricing_Fable51FallbackPricing`：5.1 cache read $0.25 vs 5 的 $1 per MTok |
| fable 判定先于 opus | `TestGetFallbackPricing_FamilyMatching` 含 `claude-fable-5-1` / `claude-fable-5` 子用例，expectedInput 10e-6 不是 opus 的 5e-6 |
| `fable-5-1` 四种别名形态先于 `fable-5` | FamilyMatching：`claude-fable-5-1` / `claude-fable-5.1` / `claude-fable5.1` / `claude-fable51` 都走 5.1 的 cache read 0.25e-6；`claude-fable-5` 走 1e-6 |
| **opus / sonnet / haiku / gemini 兜底命中与改动前一致** | `TestGetFallbackPricing_ExistingFamiliesUnchangedByFable`：`require.Same` 指回原来的 fallback 条目 |
| 前后端模型清单一致 | 后端 `DefaultModels` / Antigravity / Bedrock maps 加 `claude-fable-5-1`；前端 `useModelWhitelist.spec.ts` + `UseKeyModal.spec.ts` + `PlazaModelPricingTable.spec.ts` 同批 PASS（`{SCRATCH}/frontend-vitest.log` 64 passed） |
| **未触碰 `cache_write_1h` 相关文件** | `git show --name-only 6bd7637c4` 无 `226_*`、无 `cache_write_1h` 文件名 |

### 2.5 `593fc9365` `override_file` — commit `5c7a9fe2b`

| 项 | 证据 |
|---|---|
| 浅合并覆盖单字段、其余字段保留 | `TestPricingOverride_FieldLevelMergeKeepsOtherFields`：只改 input，output / provider / 阶梯字段保留 |
| `null` 删字段 | `TestPricingOverride_NullFieldValueRemovesField`：`long_context_*` 置 null 后 threshold/multiplier 为 0，input 价仍在。用例改为显式 `long_context_*` 字段（本 fork 不从 `above_*` 折算 272k，见 plan Deviations） |
| 目录/回退都没有的模型作为独立条目并入 | `TestPricingOverride_LoadPipelineAddsNewModelAndPatchesFallbackOnly` |
| **纯补丁不抢先建条目**（其余分项价不变 0） | 同上：fallback-only 模型带着完整 cache_read，不被空 stub 挡住 |
| 三条路径语义一致 | `applyPricingOverrides` 挂在 `parsePricingData` 入口；`TestPricingOverride_PatchesDefaultCatalogFieldWithoutTouchingOthers` |
| 文件缺失 / 非法 JSON ⇒ 目录仍正常加载 | `TestPricingOverride_MissingOrInvalidFileIsIgnored/{missing_file,invalid_json}`，WARN 后继续 |
| 拼错模型名打 WARN | `TestPricingOverride_IneffectiveEntryWarns`：日志含 `typo-model` |
| **`fallback_file` 语义未变** | `TestPricingOverride_FallbackFileStillFillOnlyWhenBothConfigured` |
| `deploy/config.personal.sqlite.yaml` 已补说明 | 注释态 `#   override_file: "/opt/sub2api/data/model_pricing_overrides.json"`；`config.example.yaml` 同样是注释 |

第一轮 override 单跑曾因 `above_*`→272k 断言失败，适配后全量 unit 里 `internal/service` **ok 157s**（ops 失败在 handler 包，与 override 无关）。

### 2.6 PR#6447 → PR#6425 推理档位 — commit `45d8ffb0e` / `a3d5d36ba`

| 项 | 证据 |
|---|---|
| 迁移 `225` 是 SQLite 方言（无 `IF NOT EXISTS` / 无 `COMMENT ON`） | `225_group_reasoning_effort_over_limit.sql`：裸 `ALTER TABLE groups ADD COLUMN max_reasoning_effort_over_limit VARCHAR(20) NOT NULL DEFAULT 'downgrade'` |
| 迁移注释无 PG 字面量（方言审计通过） | `TestProductionSQLUsesSQLiteDialect` ok；文件头是 `[sqlite-converted]` SQL 注释 |
| `TestMigration225...` 通过 | `TestMigration225AddsGroupReasoningEffortOverLimitWithSQLiteSyntax` |
| **存量分组行为不变**（默认落在降级档） | DEFAULT `'downgrade'`；`TestNormalizeMaxReasoningEffortOverLimit("")` ⇒ `downgrade`；`TestApplyOpenAIReasoningEffortPolicy` 默认 overLimit 空字符串走钳制 |
| 配置为拒绝 ⇒ 4xx 且错误信息含请求档位与上限 | `TestOpenAIReasoningEffortPolicyContext` deny：`ReasoningEffortOverLimitError` Requested=`max` Max=`medium`。ops 分类用例 `group_reasoning_effort_over_limit_deny`：403，本 fork `wantErrType: api_error`（不是上游 `permission_error`） |
| 配置为降级 ⇒ 与改动前一致 | `TestApplyOpenAIReasoningEffortPolicy` 大量 caps 子用例 + `TestOpenAIReasoningEffortPolicyFourPathsAgree` clamp 路径 |
| 未超限时两种配置行为相同 | `keeps lower value`：请求 `low`、上限 `high`，changed=false |
| 按模型限定映射命中 / 不命中 | `prefix mapping applies to matching model` / `prefix mapping skips non matching model` |
| **未限定模型的存量映射仍全模型生效** | 未带 MatchType/Model 的 mappings 子用例（`maps before cap` 等）仍对任意模型生效 |
| 多条映射同时可命中时优先级确定 | `exact mapping beats prefix`：exact `gpt-5.4` → medium，prefix gpt → low 被压过 |
| 四条发送路径结果一致 | `TestOpenAIReasoningEffortPolicyFourPathsAgree`：`ApplyOpenAIReasoningEffortPolicy` / WS ingress hooks / FromContext，clamp 与 deny 两侧 |
| api_key 认证投影带上新字段 | `api_key_repo` + auth cache 随 `45d8ffb0e` 改；admin/handler 单测 `ok`（reasoning-scope 切片） |
| 分组复制继承取值 | `admin_group_duplicate` 随 PR#6447 落地 |
| 非法取值被拒 | `TestNormalizeMaxReasoningEffortOverLimit("block")` 空；`normalizeMaxReasoningEffortOverLimitForPlatform(..., "block")` error `not supported` |
| ent 生成物走 `go generate` | `45d8ffb0e` 含 `ent/schema/group.go` + `ent/group*.go` / `ent/mutation.go` / `ent/runtime/runtime.go` / `ent/migrate/schema.go`，不是手打 runtime 补丁 |
| **未误动 postgres-gated 测试** | `git diff --name-only 3da1c2dd0..HEAD` 无 `migrations_schema_integration_test.go`；该文件首行仍是 `//go:build integration && postgres` |
| 前端 vitest 通过（不是只 typecheck） | `groupsReasoningEffort.spec.ts` 17 tests、`ReasoningEffortPolicyFields.spec.ts` 2 tests PASS（`{SCRATCH}/frontend-vitest.log`） |

### 2.7 `34b8bf1a6` 后半 渠道 1h 缓存价 — commit `e79c15c6f`（迁移断言 `ff897d4d8`，gofmt/表单 `0af248bc8`）

| 项 | 证据 |
|---|---|
| 迁移 `226` 四张表各加一列、SQLite 方言 | `226_channel_cache_write_1h_pricing.sql` 四条裸 `ALTER TABLE ... ADD COLUMN cache_write_1h_price NUMERIC(20,12)` |
| 迁移注释无 PG 字面量 | 头注 `[sqlite-converted]`；`NotContains IF NOT EXISTS` / `COMMENT ON`（`TestMigration226AddsChannelCacheWrite1hPriceWithSQLiteSyntax`） |
| `TestMigration226...` 通过 | 同上 + `TestMigrations225And226ApplyOnFreshSQLite` |
| 空库 + 已有库两种起点都能跑过 | 两次独立空库 `db1`/`db2` PASS；方言审计扫生产 SQL |
| 只配 5m ⇒ **1h 仍按 5m 值**（向后兼容底线） | `TestGetModelPricingWithChannel_CacheWriteTTLFourCombinations/only 5m set covers both TTLs`；区间侧 `TestIntervalToModelPricing_CacheWriteTTLFourCombinations` 同构 |
| 只配 1h ⇒ 1h 按新值、5m 回落官方价 | `only 1h set leaves 5m on official` |
| 两个都配 ⇒ 各按各的 | `both set are independent` |
| 都不配 ⇒ 与改动前一致 | `both nil keep official` |
| 区间定价生效 | `TestIntervalToModelPricing_CacheWriteTTLFourCombinations` 四组合 |
| 账号统计侧同规则 | `account_stats_pricing_test.go` 带 `CacheWrite1hPrice`；`account_stats_pricing.go` 走同一套 overlay |
| 表单留空落库为 NULL（不是 0） | `TestPricingRequestToService_NilPriceFields` 现断言 `CacheWrite1hPrice` 为 nil（`0af248bc8`）；前端 sync-models 对象显式 `cache_write_1h_price: null` |
| 模型广场展示该档 | `PlazaModelPricingTable.spec.ts` 16 tests PASS；`TestGetModelDefaultPricing_ReturnsFable51CacheTTLs` / `_OmitsUnsupportedCache1hPrice` |
| 前端 vitest 通过 | 触及的 plaza + types helper 已含新字段；全量见 §1 |

### 2.8 `1a33dc8cc` 弹窗布局 — commit `e79c15c6f`（与 2.7 同提交）

| 项 | 证据 |
|---|---|
| `groupsModelsListLayout.spec.ts` 通过 | **2 tests PASS**（`{SCRATCH}/frontend-vitest.log`）。本 fork 用 `width="wide"` + `shrink-0 whitespace-nowrap` + `flex flex-wrap`，不是上游 `btn btn-secondary` |
| 不引入横向滚动 | 规格断言 wide dialog + wrap/shrink-0；无 `sticky top-0` 把工具条钉进滚动区 |
| 与阶段 7 相邻实施（同文件只解一次冲突） | 同一 commit `e79c15c6f` 改 `IntervalRow.vue` / `PricingEntryCard.vue` / `GroupsView.vue` |

### 2.9 `4a1da2950` dompurify 升级 — **本批未做**

按本目标 non-goals：不改 `frontend/package.json` / `pnpm-lock.yaml` / `pnpm-workspace.yaml`。
0.1.180 §5.1 仍挂着；intake 重量记录在 `docs/upstream-sync/PORTING-0.1.180.md` §12，未标「已合」。

## 3. 影响面复核

| 面 | 复核项 | 结果 |
|---|---|---|
| 数据库 | 迁移号 `224` → `226`；两条都不可再改（checksum 不可变）⇒ 落库前已 review 完 | **是**。`225` DEFAULT=`downgrade`；`226` 四列可 NULL。空库 runner 两次都执行过 |
| 计费（a） | Fable 请求在目录缺条目时从「无兜底」变成有兜底价 ⇒ 这类请求的记账金额会变化 | **是**，这是修正。目录命中仍赢 |
| 计费（b） | 配了 `cache_write_1h_price` 的渠道，1h 缓存写入按新列计价 | **是**。NULL 时 5m 列继续覆盖两档 |
| 计费（c） | `override_file` 启用即为最高优先级 ⇒ 配错直接改变计费。确认默认关闭且示例是注释态 | **默认空**；两份 yaml 都是注释示例 |
| 准入 | 推理档位超限从「静默钳制」变为「按分组配置」。确认存量分组默认等价 | DEFAULT `downgrade` = 改动前静默钳制。deny 才 4xx |
| 兼容性 | `cache_write_1h_price` 为 NULL 时的向后兼容条款逐字保留 | `applyChannelCacheWriteTTLPrices` 保留 `CacheWrite1hPrice == nil` 分支 |
| 前端 | 三个阶段都动前端，vitest 全跑过（不是只 typecheck） | 触及文件 64 passed；critical 168 passed / 2 skipped。全量 233 files / 1649 tests passed，唯一失败是既有 `useRoutePrefetch.spec.ts` |
| DI | `wire.go` / `wire_gen.go` 零改动 | **确认** |
| 前端依赖 | dompurify 两份副本合一、无其它包版本变化 | **本批未改依赖**。阶段 9 未做 |
| 安全门禁 | 升级前后 `--prod --audit-level=high` 结果相同；例外清单未动 | 例外清单未动。审计命令因阶段 9 未做而未跑 |
| 版本 | `VERSION` 仍是 `1.1.11` | **是** |

## 4. 上线后观察

- [ ] 观察 Fable 请求的记账金额是否出现预期内的变化（阶段 4）
- [ ] 观察是否有分组因超限处置被配成「拒绝」而出现意外 4xx（阶段 6）
- [ ] 观察 1h 缓存写入 token 的记账是否与渠道配置一致（阶段 7）
- [ ] 若启用了 `override_file`，观察启动日志里的 WARN 哨兵是否为空（阶段 5）
- [ ] 观察 WS v2 透传的 `relay_completed` 与失败计数比例变化（阶段 3）
- [ ] ~~观察 `/legal/*`、`/custom/*`、公告、侧栏 SVG 图标四处渲染是否有回归（阶段 9）~~ 阶段 9 未做
- [ ] 下次依赖维护时重新看 `xlsx`（运行时唯一的 high，无干净升级路径，例外 `expires_on: 2026-10-06`）
- [ ] 第一档 `port-upstream-0.2.0-p0-fixes` 仍未合；合它时本批已手解的 WS/handler 冲突面需要再量一次
