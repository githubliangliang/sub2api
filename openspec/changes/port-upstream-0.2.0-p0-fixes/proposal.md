## Why

上游 v0.1.185（2026-09-01）与 v0.2.0（2026-09-02）共 60 个非 merge commit，评估见
`docs/upstream-sync/PORTING-0.2.0.md`。本批取其中第一档 6 项——**缺陷在本仓库已核实、
上游 patch site 对得上、依赖符号全部在位、彼此之间没有文件冲突**。

其中一项是本轮唯一一条在本仓库运行路径上**逐跳追通**的活缺陷，也是本批优先级最高的：

- **透传账号在选号阶段被静默剔出候选集。** `filterSchedulerExtra`
  （`internal/repository/scheduler_cache.go:980`）是白名单，从来没有携带
  `openai_passthrough` / `openai_oauth_passthrough`，而同一份投影里
  `filterSchedulerCredentials` **完整保留 `model_mapping`**。`buildSchedulerMetadataAccount`
  （`:875`）用它构造写入 `sched:meta:<id>` 的投影，`SchedulerSnapshotService.ListSchedulableAccounts`
  读的正是这份投影 ⇒ `Account.IsOpenAIPassthroughEnabled()`（`internal/service/account.go:1742`）
  恒 false ⇒ `IsModelSupported`（`:816`）第一行那条 issue #4936 的短路**在快照路径上永不触发**，
  退回按残留白名单判定，账号被记为 `model_not_supported`。转发阶段另从完整账号 hydrate，
  所以症状是「单独测账号能通、走网关报 no available accounts / 404」。
  快照 worker 在 SQLite 上照常启动（`internal/service/wire.go:419` 明确不在
  `skipSQLiteBackgroundJobs` 之列），**这条在本 fork 的默认部署下就是活的**。

另外三项是外部可观察的硬故障：

- **带 `fallbacks` 的 Anthropic 请求被上游 400 掉。** 客户端现在会发 `body.fallbacks`，
  上游没有 `anthropic-beta: server-side-fallback-2026-07-01` 时直接
  `fallbacks: Extra inputs are not permitted`。OAuth mimic 与默认 API-key beta 两条路径都中，
  而以 Claude 为主用法正是本 fork 的主场景。
- **ctx_pool WS ingress 把上游容量降载码原样下发。** Codex CLI 按闭集判定，
  `server_is_overloaded` / `slow_down` 属致命集 ⇒ 打印 "Selected model is at capacity" 并
  **终止会话，不退避重试**。本仓库已有 `sanitizeOpenAICapacityShedErrorCodeForClient`
  （`internal/service/openai_gateway_passthrough.go:1131`），HTTP/SSE 与 http_bridge 两条路径早已调用，
  **ingress 直写是唯一漏掉的一条**——同一账号同一时刻切到 http_bridge 就能正常退避。
- **不支持无 reader ping 的空闲 WS 连接留在池里直到坏掉。** coder/websocket 没有 reader 时消费不了
  pong 帧，socket 会挂到上游 keepalive 窗口过期才被发现，取出来即是坏连接。

剩两项是准确性/可维护性：Codex 目录不该挑持久禁用的账号、fast 模型该透出 priority tier；
账号统计成本不该维护第二份「单价 × token」实现（本仓库正好卡在少一个 `"fast"` 字面量上，
`normalizeBillingServiceTier` 只做 lower+trim，不把 `fast` 归一成 `priority`）。

本变更**只做移植**，不跟进任何功能簇。不做项及理由见 `design.md` 第 4 节与 PORTING §5。

## What Changes

- `filterSchedulerExtra` 的白名单新增 `openai_passthrough` 与 `openai_oauth_passthrough` 两个 key，
  使调度快照投影上的模型门与完整账号一致。**不改该白名单的其余任何条目。**
- Anthropic 请求装配（原生与 Bedrock 两条）在未声明 server-side-fallback beta 时剥离
  `body.fallbacks`；声明了 beta 则原样透传。剥离方式对齐既有的 `context_management` sanitize。
- ctx_pool WS ingress 直写路径在**写出客户端副本前**改写容量降载错误码，改写结果写入独立变量；
  账号状态判定继续使用未改写的原始 payload。
- WS 连接池清理时，对「未租出 + 无 waiter + 不支持无 reader ping + 空闲超过阈值」的连接提前逐出。
- Codex 模型目录跳过持久禁用的账号；fast 模型在目录里透出 priority service tier。
- `tryModelFilePricing` 删除手算分支，统一走 `CalculateCostWithServiceTier`（`channelPricing` 仍为 nil，
  保持优先级 3「只取模型定价文件、不引入渠道自定义定价」的语义）。
- **不新增迁移、不改 ent schema、不动 wire 图、不新增配置项、不改 `VERSION`、不动前端。**

## Capabilities

### New Capabilities

- `scheduler-snapshot-projection-fidelity`: 调度快照投影必须保真到「模型门判定所依赖的字段」这一层。
- `anthropic-request-sanitize`: 未声明 beta 的 Anthropic 请求字段剥离契约。
- `openai-ws-capacity-shed-client-view`: 上游容量降载在「客户端可见视图」与「账号状态判定视图」上的分离。
- `openai-ws-pool-liveness`: WS 连接池对不可保活连接的回收契约。
- `codex-model-catalog-eligibility`: Codex 模型目录的账号准入与 service tier 透出。
- `account-stats-cost-pipeline`: 账号统计成本必须与用户计费共用同一条定价管线。

### Modified Capabilities

无。`openspec/` 下没有已发布的 capability 基线（只有 `changes/`），故本变更以 ADDED Requirements
形式固化「移植后应当成立的行为」。

## Impact

- **后端**：`internal/repository/scheduler_cache.go`；`internal/service/` 的 `gateway_request.go`、
  `bedrock_request.go`、`openai_ws_forwarder_ingress.go`、`openai_ws_pool.go`、
  `openai_codex_models_service.go`、`account_stats_pricing.go`；`internal/pkg/claude/constants.go`。
- **前端**：无。
- **数据库**：无。**迁移号仍是 `224`**，本批不得顺延。
- **配置**：无新增配置项，无默认值变化。
- **调度**：第 1 项会**放大候选集**——此前被误剔的透传账号开始参与选号。这是修正，不是放宽；
  但如果某个透传账号本身有问题，它此前被这个 bug 意外挡住，之后会开始承接流量。见 `verification.md` §3。
- **计费**：第 6 项让账号统计里 `service_tier` / 长上下文两类请求走统一管线。**低谷/标准档结果与原算式
  完全一致**（上游明确保证），差异只出现在此前被手算分支绕过的档位上，方向是**修正为更高的真实成本**。
- **兼容性**：无对外 breaking change。第 2 项只在**未声明 beta 时**删字段，声明了 beta 的请求不受影响。
- **风险面**：第 3 项动的是发给**客户端**的 payload（不是发给上游的），第 2 项动的是发给**上游**的
  请求体，两者都要求「改写范围」双侧验收——非容量类错误码必须原样下发、带 beta 的请求必须保留
  `fallbacks`。

## Execution References

- `source-baseline.md`：上游 commit 固定、按文件 apply 四态证据、依赖符号核查。
- `design.md`：移植策略、逐项决策、不做项的理由、阶段划分与回滚。
- `docs/upstream-sync/PORTING-0.2.0.md` §3：逐条 patch site 与上游 diff 摘要；§7 落地顺序。
- `docs/upstream-sync/README.md` §4 硬约束：本仓库的移植护栏。
