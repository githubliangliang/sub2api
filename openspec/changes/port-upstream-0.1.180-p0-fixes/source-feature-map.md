# 追踪表：上游 commit → Requirement → 目标代码 → 证据

19 项 P0（21 个 commit）与 11 个 capability 的双向覆盖。清单里另有 2 项前端依赖安全项
**已决定推迟**，列在末尾的「明确不在本 change 内」表中。

| # | 上游 commit | capability / Requirement | 目标代码 | 证据（`verification.md`） |
|---|---|---|---|---|
| 1 | `3445485eb` | `auth-token-refresh` / 不采信未过期快照 | `frontend/src/api/tokenRefresh.ts` | §2 token 刷新行 |
| 2 | `b1e60ba45` | `gateway-pool-mode-retry` / 两条 Requirement | `service/gateway_forward_as_chat_completions.go:165`、`service/gateway_forward_as_responses.go:170` | §3.1 |
| 3 | `cd05772e9` | `ops-resource-metrics` / 两条 Requirement | `service/ops_metrics_collector.go:601` 起 | §3.3 |
| 4 | `40c26f343` | `openai-scheduling-availability` / 空能力容器 | `service/account.go:1861` 的 `openAIEndpointCapabilitySet` | §3.2 |
| 5 | `e45490a36` | `openai-scheduling-availability` / 前导 system 前缀 | `service/openai_content_session_seed.go:118` | §2 粘性种子行 |
| 6 | `913ec5d74` | `upstream-model-catalog` / OAuth 走 Codex manifest | `service/upstream_models.go:343` 起；**需先补 `CodexCanonicalClientVersion()`** | §2 模型同步行 + §4.1 |
| 7 | `f98a056f7` | `upstream-model-catalog` / Google One 保守模型集 | `pkg/geminicli/models.go`、`service/account.go:325`、`handler/admin/account_handler.go:2614` | §2 Google One 行 + §4.2 配置检查 |
| 8 | `243921dc0` | `streaming-output-fidelity` / 按上报 item 重建 | `service/openai_gateway_response_handling.go:336/616/1992` | §2 流式重建行 |
| 9 | `bafd2e293` | `streaming-output-fidelity` / 空 tool name | `pkg/apicompat/types.go:714` | §2 tool name 行 |
| 10 | `b30651a0a` | `ollama-cloud-compat` / 思维字段双向对齐 | 新建 `service/openai_gateway_ollama_cloud_cc_reasoning.go`、`service/openai_gateway_chat_completions_raw.go` | §2 Ollama 思维行 |
| 11 | `86470628d` | `ollama-cloud-compat` / clamp 输出上限 | 新建 `service/openai_gateway_ollama_cloud_max_tokens.go` | §2 Ollama clamp 行 |
| 12 | `21c07e835` | `antigravity-forward-endpoint` / daily 指向官方域 | `pkg/antigravity/oauth.go:56` | §2 Antigravity 两行 |
| 13 | `e7a3c1202` | `antigravity-forward-endpoint` / 付费账号走 daily | `service/antigravity_gateway_retry.go:48/504` | 同上 + §4.3 顺序门禁 |
| 14 | `1e1798d90` | `composite-endpoint-access` / 视频端点一致 | `server/routes/gateway.go:93` | §2 Composite 行 |
| 15 | `f6aa9dc3c` | `log-volume-control` / 只在变化时记日志 | `securityaudit/prompt_config_store.go:136/504` | §2 日志量行 |
| 16 | `ee62dfbaf` | `admin-console-usability` / IPv6 代理解析 | `frontend/src/views/admin/ProxiesView.vue` | §2 IPv6 行 |
| 17 | `5dfad32b8` | `admin-console-usability` / 并发数 0 = 不限 | `components/admin/user/UserEditModal.vue`、`i18n/locales/{en,zh}/admin/overview.ts` | §2 并发数行 |
| 18 | `616df479e` | `admin-console-usability` / 优先级列默认展示 | `frontend/src/views/admin/AccountsView.vue` | §2 优先级列行 |
| 19 | `cfecc8d11` + `e4f869e0c` | `admin-console-usability` / 返回列表 + 诊断载荷分区 | `views/admin/ops/OpsDashboard.vue`、`ops/components/OpsErrorDetail*Modal.vue`、`i18n/locales/{en,zh}/admin/ops.ts` | §2 错误详情行 |
| — | `98c7b0e88` | **无 Requirement**（文档 URL 修正，见 `design.md` 决策 8） | `docs/ADMIN_PAYMENT_INTEGRATION_API.md` | `tasks.md` 5.10 |

## 反向核对：capability → commit

| capability | Requirement 数 | 来源 commit |
|---|---|---|
| `auth-token-refresh` | 1 | `3445485eb` |
| `gateway-pool-mode-retry` | 2 | `b1e60ba45` |
| `ops-resource-metrics` | 2 | `cd05772e9` |
| `openai-scheduling-availability` | 2 | `40c26f343`、`e45490a36` |
| `upstream-model-catalog` | 2 | `913ec5d74`、`f98a056f7` |
| `streaming-output-fidelity` | 2 | `243921dc0`、`bafd2e293` |
| `ollama-cloud-compat` | 2 | `b30651a0a`、`86470628d` |
| `antigravity-forward-endpoint` | 2 | `21c07e835`、`e7a3c1202` |
| `composite-endpoint-access` | 1 | `1e1798d90` |
| `log-volume-control` | 1 | `f6aa9dc3c` |
| `admin-console-usability` | 4 | `ee62dfbaf`、`5dfad32b8`、`616df479e`、`cfecc8d11`+`e4f869e0c` |

合计 21 条 Requirement / 77 个 Scenario，覆盖 19 项 P0（20 个 commit 有 Requirement，
`98c7b0e88` 为文档项）。

## 明确不在本 change 内

| 0.1.180 清单章节 | 内容 | 判定出处 |
|---|---|---|
| §5.1 | `4a1da2950` dompurify `3.3.1`→`3.4.14`（CVE-2026-65913） | **已决定推迟**，`design.md` 决策 1 |
| §5.4 | `b410c3913` nanoid GHSA-2v37-7h3g-55p8 审计例外 | **已决定推迟**，与上一条同批 |
| §6.1 | Responses / Chat 工具桥接 11 条 | P1，另开 change；`port-upstream-0.1.183-p0-fixes` 的 Responses Lite 簇在等它 |
| §6.2 | Grok 一套 30 条（含默认模型 4.5→4.6 决策） | P1 + §9.1 未决 |
| §6.3 | 调度诊断 `3fd66a33b` 等 4 项 | P1；与 `openai_gateway_scheduling.go` 抢文件 |
| §7 | PR #5888 大礼包、Fast `service_tier`、重置卡自动使用、模型列表读取上限、统一 token 计费路径 | P2 |
| §8 | 插件系统、模型广场分时价、CN 供应商、compose 四条、VERSION、sponsors | 不合 / N/A |
| §9 | Grok 默认模型、Go 1.27 工具链、长上下文计费门控 | 未决决策，不属移植 |
