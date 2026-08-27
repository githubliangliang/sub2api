## Why

`docs/upstream-sync/PORTING-0.1.180.md` §5 列出 21 项 P0（23 个 commit），判定依据是「本仓库已核实
有同一缺陷、patch 逐字对得上」。该清单至今**一条都没合**（`f8f257ad7..HEAD` 只有 docs / chore）。
本 change 复核了全部 23 个 commit（按文件 apply 三态 + 依赖符号 grep + 成对项顺序实测），
结论：**21 项全部仍然成立**，另外发现两处原文档漏掉的实施陷阱（`source-baseline.md` §4）。
其中 2 项前端依赖安全项**已决定推迟**（见「Deferred」一节），本 change 交付余下 **19 项**。

三项值得单独点名：

- **`b1e60ba45` 池模式同账号重试在两条 compat 路径上完全失效**。原生 Anthropic 路径
  （`gateway_forward.go:699/733`）已经在填 `RetryableOnSameAccount`，但
  `gateway_forward_as_chat_completions.go` 与 `gateway_forward_as_responses.go` 只返回
  `StatusCode` / `ResponseBody` ⇒ 池模式账号在 CC / Responses 上游报错时**永远不会同号重试**。
- **`cd05772e9` ops 面板混用 cgroup 与宿主机内存**。cgroup v2 未设内存上限时 `used` 是容器数、
  `total` 是宿主机数，面板显示 60 MB / 23 GB ≈ 0.3%，严重低估，等于内存监控失效。
- **`40c26f343` 空 `openai_capabilities` 把 OAuth 账号排除出文本调度**（#5530）。API 直写 / 导入 /
  历史数据留下的 `{}` 或 `[]` 被当成「已配置且不含任何能力」，账号被**静默**跳过。

其余 17 项覆盖流式输出保真（终端输出重建、空 tool name）、上游模型目录（OAuth 同步、Google One
收紧）、Ollama Cloud 兼容（`reasoning_content` 对齐、`max_tokens` clamp）、Antigravity 付费账号
端点、Composite 视频端点放行、日志量控制，以及五项管理台可用性。

## What Changes

- 前端 token 刷新去掉「未过期即接受对端结果」的分支，消除刷新锁空转。
- CC / Responses 两条 compat 转发路径在构造 `UpstreamFailoverError` 时填
  `RetryableOnSameAccount`，并采信 `HandleUpstreamError` 的 `shouldDisable` 返回值。
- ops 内存指标改为「整组 cgroup 或整组 host」，不再拼接两个来源。
- 空 `openai_capabilities` 容器与未配置等价（不限制能力）；非空但全 false 保持原语义。
- Chat 粘性种子只采信**前导**的 system / developer 前缀，后续动态注入的 system 消息不再让 hash 漂移。
- OpenAI OAuth 账号的模型同步走 ChatGPT Codex manifest，而不是 Platform API `/v1/models`。
- Google One OAuth 账号只暴露保守模型集（2.5 Flash / 2.5 Pro / 2.0 Flash）。
- 流式 Responses 终端输出按 `response.output_item.done` 上报的原始 item 重建。
- `ChatFunctionCall.Name` 改 `omitempty`，流式 arguments delta 不再用空 name 覆盖客户端已累积的工具名。
- Antigravity `plan_type ∈ {pro, ultra}` 的付费账号默认走 daily 端点，且 daily 端点纠正为
  `daily-cloudcode-pa.googleapis.com`（原来指向 `.sandbox.` 域）。
- Composite 分组放行视频生成任务创建（与已放行的状态/内容查询对齐）。
- Ollama Cloud raw CC 直转路径双向对齐 `reasoning_content`，并把 `max_tokens` /
  `max_completion_tokens` clamp 到 provider 上限（账号 extra 可覆盖，0/负数表示禁用）。
- `prompt_guard.config_loaded` 只在首次、版本变化或风控总闸翻转时记日志（原来每 5 秒一条）。
- 管理台：批量代理解析支持 `[IPv6]`、用户并发数 0 = 不限、账号优先级列默认展示、
  运维错误详情支持「返回列表」并保留筛选状态与诊断载荷分区展示。
- 修正一处文档自引用 URL。
- **不新增迁移、不改 schema、不改 wire 图、不新增配置项**（`ollama_max_tokens_cap` 是账号 `extra`
  里的可选键，不是全局配置）。

## Capabilities

### New Capabilities

- `auth-token-refresh`: 跨标签页 token 刷新的锁与对端结果采信条件。
- `gateway-pool-mode-retry`: 池模式账号在各转发协议上同号重试资格的一致性。
- `ops-resource-metrics`: 容器与宿主机资源指标不得混用的自洽性要求。
- `openai-scheduling-availability`: 账号能力配置的空值语义与会话种子的稳定性。
- `upstream-model-catalog`: 按账号类型选择正确的上游模型清单来源与暴露范围。
- `streaming-output-fidelity`: 流式响应终端输出与工具调用标识的保真要求。
- `ollama-cloud-compat`: Ollama Cloud OpenAI 兼容路径的思维字段与输出上限适配。
- `antigravity-forward-endpoint`: Antigravity 转发 base URL 按订阅层级的选择规则。
- `composite-endpoint-access`: Composite 分组可访问端点集合的一致性。
- `admin-console-usability`: 管理台输入校验、默认列与错误详情导航的可用性要求。
- `log-volume-control`: 周期性重载的日志只在状态变化时输出。

### Modified Capabilities

无。`openspec/` 下没有已发布的 capability 基线，本变更以 ADDED Requirements 固化「移植后应当成立
的行为」。

## Deferred

清单 §5 的 2 项前端依赖安全项**已决定推迟**，不在本 change 交付范围内。决定依据与重新评估触发条件
见 `design.md` 决策 1；`docs/upstream-sync/PORTING-0.1.180.md` §5.1 的状态已同步改为「已决定推迟」。

| 上游 commit | 内容 | 推迟理由摘要 |
|---|---|---|
| `4a1da2950` | dompurify `3.3.1` → `3.4.14`（CVE-2026-65913 / GHSA-cj63-jhhr-wcxv） | 全部 `sanitizeSvg` 输入都是管理员自填，且该绕过还需页面里另有 prototype pollution 原语；单管理员部署下实际可达性很低 |
| `b410c3913` | nanoid GHSA-2v37-7h3g-55p8 审计例外条目 | 与上一条同批推迟 |

⚠️ 推迟意味着**这两个文件一个都不动**。特别是 `frontend/package.json`：三个 workflow
（`security-scan.yml` / `backend-ci.yml` / `release.yml`）与 `Dockerfile` 都用
`pnpm install --frozen-lockfile`，只改 `package.json` 不同步 lockfile 会让 CI 与镜像构建**直接失败**。
没有「只改一半」的中间状态。

## Impact

- **后端**：`internal/service/`（9 个文件 + 3 个新文件）、`internal/handler/admin/account_handler.go`、
  `internal/pkg/{apicompat,geminicli,antigravity}/`、`internal/securityaudit/prompt_config_store.go`、
  `internal/server/routes/gateway.go`。
- **前端**：`api/tokenRefresh.ts`、`views/admin/ProxiesView.vue`、
  `views/admin/AccountsView.vue`、`components/admin/user/UserEditModal.vue`、
  `views/admin/ops/`（4 个组件 + Dashboard）、`i18n/locales/{en,zh}/admin/{overview,ops}.ts`。
- **其他**：`docs/ADMIN_PAYMENT_INTEGRATION_API.md`。
- **数据库**：无。迁移号保持 `224`。
- **调度**：`40c26f343` 会让此前被静默排除的 OAuth 账号重新进入文本调度 ⇒ 账号池实际可用容量上升，
  流量分布会变。`e45490a36` 会提高 chat 粘性命中率。
- **计费**：无变化。
- **安全**：本 change 不含依赖安全项（已推迟，见 Deferred）。19 项均为行为修正。
- **兼容性**：`f98a056f7` 会**收窄** Google One OAuth 账号在管理台可见的模型清单；
  原先展示的 3.x / image 模型该渠道本就无法服务，但如果有分组白名单依赖旧清单，需要复核。
- **前端依赖**：本 change **不改动** `package.json` / `pnpm-lock.yaml`，因此不涉及依赖解析变化。

## Execution References

- `source-baseline.md`：23 个 commit 的固定 SHA、按文件三态证据、**两处原文档漏项**、成对项顺序实测。
- `source-feature-map.md`：commit → Requirement → 目标代码 → 证据的双向追踪。
- `design.md`：移植策略、逐项决策、阶段划分与回滚。
- `docs/upstream-sync/PORTING-0.1.180.md` §5：逐条 patch site 与上游 diff 摘要；§4 硬约束见
  `docs/upstream-sync/README.md`。
