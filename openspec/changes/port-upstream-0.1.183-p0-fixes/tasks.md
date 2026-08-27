## 1. 固定基线与实施边界

- [x] 1.1 按 `source-baseline.md` §1 重建上游部分裸克隆并 fetch 三个 tag，确认 `v0.1.183 = c21fd338`、`main = efb46db0` 与本文记录一致
- [x] 1.2 记录本仓库起始 commit 与 `git status --short`，确认工作区干净
- [x] 1.3 运行并保存移植前基线信号：`cd backend && go build ./... && go test -tags=unit ./... -count=1`
- [x] 1.4 运行并保存前端基线：`cd frontend && pnpm run typecheck && pnpm run lint:check`（已知 `src/composables/__tests__/useRoutePrefetch.spec.ts` 5 条失败与本次移植无关）
- [x] 1.5 确认 `backend/migrations/` 最大序号仍为 `224`，本 change 全程 MUST NOT 新增迁移文件
- [x] 1.6 确认 `backend/cmd/server/VERSION` 保持 `1.1.8`，MUST NOT 同步成 0.1.183
- [x] 1.7 按 `design.md` Migration Plan 把实现拆成 4 个可独立评审的提交/PR

## 2. 阶段 1 — 涉钱与涉缓存

- [x] 2.1 `8e60d574`：在 `internal/service/openai_gateway_scheduling.go:31` 的 `explicitOpenAIHeaderSessionNames` 首位加入 `session-id`，并同步更新 `GenerateSessionHash` 的优先级注释
- [x] 2.2 `8e60d574`：在 `internal/service/openai_ws_forwarder_logutil.go:68` 的 `resolveOpenAIWSSessionHeaders` 前置 `session-id` 分支，`SessionSource` 记 `header_session-id`，下划线分支保持 `header_session_id`
- [x] 2.3 `bc4a9ae4`：把 `internal/service/gateway_anthropic_passthrough.go:673/676` 的 `cc5m.Exists() && cc5m.Int() > 0` / `cc1h...` 改为只判 `Exists()`
- [x] 2.4 `bc4a9ae4`：把 `internal/service/gateway_upstream_response.go:1239/1243` 的 `parseSSEUsageInt(...); exists && v > 0` 改为只判 `exists`
- [x] 2.5 `bc4a9ae4`：在 `internal/service/billing_service.go` 新增 `normalizeCacheCreationBreakdown`（负值归零、聚合值非正或未超出时原样返回、超出时按原比例缩回并让两档之和精确等于聚合值），并给 import 加 `math`
- [x] 2.6 `bc4a9ae4`：`computeCacheCreationCost`（`billing_service.go:1223`）改为先归一化再计价；「明细全为 0 且聚合值为正」的 5m 单价回落分支 MUST 保留
- [x] 2.7 `e55727d4`：在 `selectAccountWithLoadAwareness`（`openai_gateway_scheduling.go:932`）开头引入 `stickySpillover := false`
- [x] 2.8 `e55727d4`：在 Layer 1 的 `waitingCount < cfg.StickySessionMaxWaiting` 分支（`:1029` 附近）置 `stickySpillover = true`
- [x] 2.9 `e55727d4`：给 `:1185` 与 `:1224` 两处 `setStickySessionAccountID` 的条件各加 `&& !stickySpillover`（两处都要，漏一处即失效）
- [x] 2.10 阶段 1 提交拆分：session-id / cache_creation / sticky spillover 各一个提交

## 3. 阶段 2 — 上游请求兼容性

- [x] 3.1 `19da0f24`：在 `internal/service/gemini_messages_compat_service.go:3532` 的 `cleanToolSchema` 剔除列表加 `deprecated`
- [x] 3.2 `19da0f24`：新增 `normalizeGeminiEnum`（字符串原样；数字/布尔/null 编码成字符串；对象/数组返回不可归一）并在 `cleanToolSchema` 里按结果保留或整体 `delete(cleaned, "enum")`；`encoding/json` 已在 import 中，MUST NOT 重复添加
- [x] 3.3 `1a9898a6`：在 `internal/service/antigravity_gateway_compat.go` 加 `antigravityCompatMaxTokens = 64000`，并在 `preserveChatCompletionTokenLimit`（`:152`）用 `min()` 封顶
- [x] 3.4 `329b92ef`：在 `internal/service/openai_images.go` 加 `openAIImagesVerbatimPromptInstructions` 常量
- [x] 3.5 `329b92ef`：在 `internal/service/openai_images_responses.go:357` 的 `buildOpenAIImagesResponsesRequest` 里 `sjson.SetBytes` 写入 `instructions`
- [x] 3.6 `e0e5e45c`：把 `internal/pkg/apicompat/responses_client_tools.go:220` 的 `dropInvalidLoweredFunctionItemID` 改名并改写为 `normalizeLoweredFunctionItemID`（能反向映射就换前缀，否则才删 ID），更新 4 处调用点
- [x] 3.7 `e0e5e45c`：新增 `responsesToolCallItemIDPrefixes` / `responsesToolCallItemIDPrefix` / `retypedResponsesToolCallItemID` / `retypeResponsesToolCallItemID`，并在 `restoreClientToolValue` 与 `restoreResponsesOutputClientTools` 的 `custom_tool_call` / `tool_search_call` 还原分支调用
- [x] 3.8 `e0e5e45c`：给 `responsesClientToolStreamCall` 增加 `clientItemID` 字段，`recordItem` 计算它，`Restore` 的 item added/done 与自定义工具输入 delta/done 事件改用它；内部 `r.calls` / `r.byOutput` 的键 MUST 仍用上游 ID 与 `call_id`
- [x] 3.9 `9fb26043`：`internal/pkg/xai/billing.go:22` 的 `CLIClientVersion` 改 `0.2.120`
- [x] 3.10 `9fb26043`：`internal/service/openai_gateway_chat_completions_raw.go:175` 与 `internal/service/grok_observed_models.go:95` 改用 `defaultGrokUpstreamUserAgent()`，并删除 `internal/service/grok_upstream_headers.go:15` 的 `grokUpstreamUserAgent` 常量
- [x] 3.11 `99ec347e`+`71aa6e35`：`internal/domain/constants.go:105` 的 `DefaultAntigravityModelMapping` 只改两个 key —— `claude-sonnet-4-5-thinking` 与 `claude-sonnet-4-5-20250929` 指向 `claude-sonnet-4-6`；`claude-sonnet-4-5` MUST 保持指向自身
- [x] 3.12 `99ec347e`：`internal/service/account_test_service.go` 加 `defaultAntigravityTestModel = "claude-sonnet-4-6"` 常量并抽出 `antigravityConnectionTestModel`，替换 `:2272` 的字面量
- [x] 3.13 同步上游对应的 7 个测试文件里的旧 Sonnet 期望值（`internal/domain/constants_test.go` 与 `internal/service/` 下 `account_test_service_antigravity_test.go`、`antigravity_credits_overages_test.go`、`antigravity_model_mapping_test.go`、`antigravity_rate_limit_test.go`、`antigravity_single_account_retry_test.go`、`model_rate_limit_test.go`）

## 4. 阶段 3 — 账号冷却与前端

- [x] 4.1 `a6b11ccc`：在 `internal/service/ratelimit_service.go` 加 `openCodeGoUsageLimitResetPattern` 与 `openCodeGoUsageLimitDurationPartPattern` 两个包级正则
- [x] 4.2 `a6b11ccc`：`parseOpenAIRateLimitResetTime` 接受 `GoUsageLimitError`，并在既有解析全部落空后从 message 解析时长
- [x] 4.3 `a6b11ccc`：新增 `parseOpenCodeGoUsageLimitResetDuration` 与 `openCodeGoUsageLimitDurationUnit`，支持 s/m/h/d/w 与多段拼接，带溢出与非正值保护
- [x] 4.4 `eb594eef`：`frontend/src/views/user/PaymentResultView.vue` 在确认订单已履约后触发余额刷新，并处理刷新失败
- [x] 4.5 `eb594eef`：同步上游 `frontend/src/views/user/__tests__/PaymentResultView.spec.ts` 的用例

## 5. 阶段 4 — 认证路径（隔离提交）

- [x] 5.1 `4ca86c52`：`internal/repository/user_repo.go` 把 `existsByEmailAliasWithClient` 改写为基于新的 `emailAliasOwnerIDWithClient`（查询改 `Select(ID, Email).All`，「其他用户」优先于当前用户返回）
- [x] 5.2 `4ca86c52`：新增 `userRepository.UpdateEmailWithAliasGuard`：要求事务上下文 → `lockRepositoryScopedKeys(normalizedEmailUniquenessLockKey, emailAliasUniquenessLockKey)` → 锁内复查归属 → `SetEmail` + `SetPasswordHash` → `translatePersistenceError`
- [x] 5.3 `4ca86c52`：`internal/service/auth_email_binding.go` 抽出 `ensureEmailIdentityAvailableForUser`（字面查重 → 当前用户 alias 放行 → `ExistsByEmailAlias`），并在 `BindEmailIdentity`（`:59`）与 `SendEmailIdentityBindCode` 两处替换原有的仅字面查重
- [x] 5.4 `4ca86c52`：新增 `emailIdentityAliasGuardRepository` 接口，`updateBoundEmailIdentityWithClient` 改为经该接口调用守卫方法；类型断言失败 MUST 返回服务不可用
- [x] 5.5 确认 `lockRepositoryScopedKeys` 仍是纯进程内锁实现且未被本次改动绕过（`design.md` 决策 6）

## 6. 门禁

- [x] 6.1 核对最终 diff 的文件清单，确认只包含 `source-baseline.md` §3 列出的 12 项涉及的文件，MUST NOT 出现 `composite_route_resolver.go` / `gateway_service.go` 等 routed-catalog 簇文件（防止顺手带入 `3e98a5a1`）
- [x] 6.2 确认没有新增 `backend/migrations/*.sql`，`VERSION` 未变
- [x] 6.3 确认没有新增配置项，`deploy/config.example.yaml` 与 `deploy/config.personal.sqlite.yaml` 未变
- [x] 6.4 确认未改动 `wire_gen.go`（本批无 wire 变更；若有则说明走偏）
- [x] 6.5 `grep -rn "sub2api-grok/1.0" backend/` 零命中
- [x] 6.6 `grep -rn "dropInvalidLoweredFunctionItemID" backend/` 零命中
- [x] 6.7 按 `verification.md` 逐条勾选证据后，把 `docs/upstream-sync/PORTING-0.1.183.md` 第 3 节对应条目的「状态」从**待合**改为**已合**
