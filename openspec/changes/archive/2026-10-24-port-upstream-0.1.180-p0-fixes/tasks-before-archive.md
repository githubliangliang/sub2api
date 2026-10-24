## 1. 固定基线与实施边界

- [x] 1.1 按 `source-baseline.md` §1 取 21 个 commit 的 patch，确认 `v0.1.180 = c40edb4`
- [x] 1.2 记录本仓库起始 commit 与 `git status --short`，确认工作区干净
- [x] 1.3 保存移植前基线：`cd backend && go build ./... && go test -tags=unit ./... -count=1`
- [x] 1.4 保存前端基线：`cd frontend && pnpm run typecheck && pnpm run lint:check`（已知 `useRoutePrefetch.spec.ts` 5 条失败与本次无关）
- [x] 1.5 确认 `backend/migrations/` 最大序号仍为 `224`，全程 MUST NOT 新增迁移
- [x] 1.6 确认 `backend/cmd/server/VERSION` 保持 `1.1.8`
- [x] 1.7 ⚠️ 确认 `frontend/package.json`、`frontend/pnpm-lock.yaml`、`.github/audit-exceptions.yml` **全程不动**（依赖安全两项已推迟，见 `design.md` 决策 1；只改 `package.json` 会让 `--frozen-lockfile` 的 CI 与镜像构建失败）
- [x] 1.8 按 `design.md` Migration Plan 拆成 5 个阶段的提交/PR

## 2. 阶段 1 — 静默失效三件套（各自独立提交）

- [x] 2.1 `b1e60ba45`：`service/gateway_forward_as_chat_completions.go:165` 处接住 `HandleUpstreamError` 的返回值到 `shouldDisable`，并给 `UpstreamFailoverError` 填 `RetryableOnSameAccount: !shouldDisable && account.IsPoolMode() && account.IsPoolModeRetryableStatus(resp.StatusCode)`
- [x] 2.2 `b1e60ba45`：`service/gateway_forward_as_responses.go:170` 同样处理（两处都要，漏一处即该协议仍然失效）
- [x] 2.3 `b1e60ba45`：落地上游新增的 `service/gateway_pool_mode_retry_test.go`
- [x] 2.4 `40c26f343`：`service/account.go` 的 `openAIEndpointCapabilitySet`（`:1861`）在 `[]any` / `[]string` / `map[string]any` / `map[string]bool` 四个分支各加空容器早退 `return nil, false`；非空全 false MUST 保持原语义
- [x] 2.5 `cd05772e9`：`service/ops_metrics_collector.go` 把 CPU 与内存的采集拆开；新增 `resolveMemoryStats(cgroupUsed, cgroupTotal, cgroupOK, host)` 返回自洽三元组，只在 `!cgroupOK || cgroupTotal == 0` 时才去读宿主机
- [x] 2.6 `cd05772e9`：落地上游新增的 `service/ops_metrics_collector_memory_test.go`
- [ ] 2.7 阶段 1 上线后进入观察期（`verification.md` §5）：2.4 会让此前被静默排除的 OAuth 账号重新进入调度

## 3. 阶段 2 — 流式与协议保真

- [x] 3.1 `243921dc0`：`service/openai_gateway_response_handling.go` 新增 `responsesStreamOutputItems`（按 `output_index` 逐字节归档 `response.output_item.done` 的 item），在流循环里 `Observe(dataBytes)`
- [x] 3.2 `243921dc0`：`normalizeResponsesStreamingTerminalOutput` 增加 `streamDoneItems` 参数并优先使用上报 item 重建，无上报时回落 delta 累积；`sort` 需加入 import
- [x] 3.3 `243921dc0`：落地上游新增的 `service/openai_responses_stream_output_items_test.go`
- [x] 3.4 `bafd2e293`：`pkg/apicompat/types.go:714` 的 `ChatFunctionCall.Name` 改 `json:"name,omitempty"`
- [x] 3.5 `bafd2e293`：落地上游新增的 `pkg/apicompat/responses_to_chatcompletions_tool_name_test.go`

## 4. 阶段 3 — 模型目录与账号端点

- [x] 4.1 ⚠️ **先补基座**：`service/openai_codex_identity.go` 新增 `func CodexCanonicalClientVersion() string { return resolveCodexOutboundIdentity("").version }`（与上游 `c40edb4` 的函数体一致）。不补这一步 `913ec5d74` 无法编译，见 `source-baseline.md` §4.2
- [x] 4.2 `913ec5d74`：`service/upstream_models.go` 的 `buildOpenAIUpstreamModelsRequest` 在开头对 `account.IsOpenAIOAuth()` 分流，新增 `buildOpenAIOAuthUpstreamModelsRequest`（走 Codex manifest + Agent 身份头）
- [x] 4.3 `f98a056f7`：**先做配置检查**（`design.md` 决策 6）——查现有 Google One OAuth 账号与分组白名单 / `model_pricing` 是否依赖将被移出清单的模型 ID
- [x] 4.4 `f98a056f7`：`pkg/geminicli/models.go` 新增 `GoogleOneModels` 与返回独立副本的 `GoogleOneModelMapping()`；`service/account.go:325` 新增 `IsGeminiGoogleOne()`；`handler/admin/account_handler.go:2614` 的 `GetAvailableModels` 在 OAuth 分支里优先返回 Google One 清单
- [x] 4.5 `21c07e835` + `e7a3c1202`：**同一个提交**。先把 `pkg/antigravity/oauth.go:56` 的 `antigravityDailyBaseURL` 从 `daily-cloudcode-pa.sandbox.googleapis.com` 改为 `daily-cloudcode-pa.googleapis.com`
- [x] 4.6 `e7a3c1202`：`service/antigravity_gateway_retry.go` 的 `resolveAntigravityForwardBaseURL` 改为接收 `*Account`，未显式配置且 `accountHasAntigravityPaidTier(account)`（`plan_type` ∈ `{pro, ultra}`，大小写/空白不敏感）时取 `baseURLs[1]`；更新 `:504` 的调用点
- [x] 4.7 `1e1798d90`：`server/routes/gateway.go:93` 的 `videoGenerationHandler` 把判定改为 `platform == PlatformGrok || platform == PlatformComposite`

## 5. 阶段 4 / 5 — Ollama Cloud、会话种子、日志、管理台

- [x] 5.1 `b30651a0a`：新增 `service/openai_gateway_ollama_cloud_cc_reasoning.go`（账号判定 + 请求/响应/SSE 三个方向的 `reasoning_content` 对齐），并在 `service/openai_gateway_chat_completions_raw.go` 接入
- [x] 5.2 `86470628d`：**在 5.1 之后**新增 `service/openai_gateway_ollama_cloud_max_tokens.go`（`OllamaCloudMaxTokensCapExtraKey` + 默认上限 + `clampOllamaCloudMaxTokens`），由 `applyOllamaCloudRawChatCompletionsRequest` 调用
- [x] 5.3 `e45490a36`：`service/openai_content_session_seed.go:118` 引入 `systemPrefixOpen`，只把首个 user 消息之前的 system/developer 计入种子；遇到 user 或其他角色即关闭前缀
- [x] 5.4 `f6aa9dc3c`：`securityaudit/prompt_config_store.go` 的 `clearLoadError` 改为返回 `bool`，新增 `shouldLogConfigLoaded(previous, storage, active)`，`Reload` 只在 `recovered || shouldLogConfigLoaded(...)` 时记 `EventConfigLoaded`
- [x] 5.5 `3445485eb`：`frontend/src/api/tokenRefresh.ts` 删掉 `TOKEN_REFRESH_BUFFER_MS` 常量与 `readPeerRefreshResult` 里「未过期即采信快照」的整个分支
- [x] 5.6 `ee62dfbaf`：`views/admin/ProxiesView.vue` 的代理 URL 正则支持 `[IPv6]`，解析后去掉方括号再提交
- [x] 5.7 `5dfad32b8`：`components/admin/user/UserEditModal.vue` 的并发数输入加 `min="0"` / `step="1"` / 占位符 / 提示，校验改为拒绝非整数与负数；`i18n/locales/{en,zh}/admin/overview.ts` 把 `concurrencyMin` 换成 `concurrencyNonNegative` 并新增 `concurrencyPlaceholder` / `concurrencyHint`
- [x] 5.8 `616df479e`：`views/admin/AccountsView.vue` 的 `DEFAULT_HIDDEN_COLUMNS` 去掉 `'priority'`
- [x] 5.9 `cfecc8d11` + `e4f869e0c`：ops 错误详情加 `backToList` prop 与 `back` 事件、`OpsDashboard.vue` 接线 `detailReturnTarget` / `handleBackToList`；补 `upstreamStatus` / `rootCause` / `diagnosticPayloads`（`client` / `upstream_message` / `upstream_detail` / `upstream_events`）中英文案与只展示非空分区的逻辑（两条顺序无关，见 `source-baseline.md` §4.5）
- [x] 5.10 `98c7b0e88`：修 `docs/ADMIN_PAYMENT_INTEGRATION_API.md` 的自引用 URL

## 6. 门禁

- [x] 6.1 核对最终 diff 文件清单，只包含 `source-feature-map.md` 列出的文件；MUST NOT 出现 0.1.180 §6 / §7 的文件（工具桥接、Grok 一套、#5888 大礼包）
- [x] 6.2 确认没有新增 `backend/migrations/*.sql`，`VERSION` 未变
- [x] 6.3 确认未改动 `wire_gen.go`
- [x] 6.4 确认没有新增**全局**配置项（`ollama_max_tokens_cap` 是账号 `extra` 的可选键，不是 config.yaml 项）
- [x] 6.5 ⚠️ `git diff --name-only` 中 MUST NOT 出现 `frontend/package.json`、`frontend/pnpm-lock.yaml`、`frontend/pnpm-workspace.yaml`、`.github/audit-exceptions.yml`（依赖安全两项已推迟）
- [x] 6.6 `grep -rn "daily-cloudcode-pa.sandbox" backend/` 零命中
- [x] 6.7 `grep -rn "TOKEN_REFRESH_BUFFER_MS" frontend/src/` 零命中
- [x] 6.8 `grep -rn "concurrencyMin" frontend/src/` 零命中
- [x] 6.9 `grep -rn "CodexCanonicalClientVersion" backend/` 有命中（4.1 确实补上了）
- [x] 6.10 逐条勾完 `verification.md` 后，把 `docs/upstream-sync/PORTING-0.1.180.md` §5 对应条目的状态从**待合**改为**已合**（§5.1 与 nanoid 那行保持「已决定推迟」）
