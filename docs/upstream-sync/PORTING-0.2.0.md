# 移植清单：上游 v0.1.185 / v0.2.0

对照日期：2026-09-02。基线 `3da1c2dd0`（= 本仓库 `1.1.11`，含 0.1.183 全部 + 0.1.184
§3 的 14 项 P0、§4 的 11 项 P1、§5.1 Codex routed catalog 与 §5.7 Antigravity 混合工具两簇）。

上游 v0.1.185 是**定价目录重构 + 网关兼容性修复**版；v0.2.0 是上游第一个 0.2.x，但内容
不是架构变更，而是**分组级策略开关**（OpenAI Fast、推理档位超限处置）+ Kimi 原生 Responses
+ Fable 5.1。本文覆盖 `v0.1.184 (e98ef32eb) .. v0.2.0 (aa2364883)` 共 **60 个非 merge commit**
（v0.1.185 16 条 / v0.2.0 44 条）。`v0.2.0..upstream/main` 只剩一条 `chore: sync VERSION`。

⚠️ **上一轮的对照上游 main `200602b41` 已经在 v0.1.185 内部**（v0.1.185 = `200602b41` + 4 条）。
所以 v0.1.185 有一部分在上一轮就已可见，其中 `e2624fb65` 已随 §5.1「按终态取文件」一并落地，
本文按**已合**记（见 §6）。0.1.184 §1 那句「上游 main 之后 3 条本文不含」说的是那份文档的
覆盖范围，不是「本仓库没有」——这是本轮的第一条教训（§9.1）。

移植前先读 README [第 4 节「硬约束」](./README.md#4-硬约束)，尤其第 13 条；写 SQL 对照
[第 5 节转换速查](./README.md#5-pg--sqlite-转换速查)。

**结论：11 项 P0 可以吃下（缺陷已核实、基座符号在位，其中 §3.1 是本轮唯一一条我在本仓库
运行路径上完整追通的活缺陷），4 项 P1 需手工改；Fast 组策略 / 长上下文阶梯数据驱动 /
Kimi 原生 Responses / API key 合成 instructions / 数据库启动重试 / 404-vs-429 六簇本轮不做。**

**落地（2026-09-02）**：第二档 §3.8–§3.11 + §4.1–§4.4 已在分支 `sync/upstream-20260902-p1`
（起点 `3da1c2dd0`，代码尖 `0af248bc8`）合入。第一档 §3.1–§3.7 仍未合；第二档 OpenSpec 第 9 项
（0.1.180 §5.1 dompurify）本批未做。

---

## 1. 版本对照

| 点 | 状态 |
|---|---|
| 上游最新正式版 | [`v0.2.0`](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.0)（tag commit `aa2364883`，2026-09-02 11:09 UTC） |
| 中间版本 | [`v0.1.185`](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.1.185)（tag commit `2ac784c51`，2026-09-01 09:32 UTC） |
| 上一轮对照基线 | `v0.1.184`（tag commit `e98ef32eb`，2026-08-31 16:43 UTC） |
| `v0.1.184` → `v0.1.185` | **16 commits（非 merge）**，其中 3 条是 chore/VERSION |
| `v0.1.185` → `v0.2.0` | **44 commits（非 merge）**，13 个 PR merge；两个 Fast 功能各被拆成 13–18 条微提交 |
| 上游 main | `5097b3145`，v0.2.0 之后仅 `chore: sync VERSION to 0.2.0`，本文已含全部功能项 |
| 本仓库基线 | `3da1c2dd0`，`backend/cmd/server/VERSION` = `1.1.11`（自有编号，**不要**同步成 0.2.0） |
| 本仓库迁移号 | 起点 `224`，第二档落地后为 **`226`**（`225` 推理档位 / `226` 渠道 1h 缓存价）。上游本轮 4 条新迁移，**上游自己又撞号**（三条都叫 `232_*` + 一条 `233_*`）；本仓库按落地顺序自排，见 §2.3 |
| 本仓库平台面 | 仍只有 anthropic / openai / gemini / antigravity / grok / composite（`internal/domain/constants.go:21-26`）。Kimi / Zhipu / DeepSeek **不是平台**，这是 §5.3 的判据 |

对比链接：<https://github.com/Wei-Shaw/sub2api/compare/v0.1.184...v0.2.0>

### 1.1 按簇统计

| 簇 | commits / PR | 三态 | 判断 |
|---|---|---|---|
| 调度快照投影丢透传开关 | `e93e6368f` | 1 CONFLICT + 1 CLEAN | **P0 最高优先**（3.1），本仓库活缺陷 |
| Anthropic `fallbacks` 剥离 | `200b1406d` | 4 CLEAN | **P0**（3.2） |
| WS 两条自包含修复 | `1dc0a0900` / `6d5f02784` | 各 2 CLEAN | **P0**（3.3 / 3.4） |
| Codex 目录两条 | `ba345f105` / `57c76584a` | 5 + 2 CLEAN | **P0**（3.5 / 3.6） |
| 账号统计成本走统一管线 | `9eabd2a5b` + `e7c029875` | 1 CONFLICT + 1 CLEAN | **P0**（3.7） |
| bootstrap 缺 call id | `1be69e56a` + `421a83282` | 2 CLEAN + 链 | **P0**（3.8） |
| API key 会话缓存身份 | `504919a05` | 3 CLEAN + 1 CONFLICT | **P0**（3.9） |
| WS 终止事件前 close | `bfe0a5a87` | 2 CONFLICT + 11 行基座 | **P0**（3.10） |
| Fable fallback 定价 + Fable 5.1 | `34b8bf1a6` 前半 | 大部分 CLEAN | **P0**（3.11） |
| 推理档位两簇 | PR#6447 → PR#6425 | 38/47 + 6/15 CLEAN | **P1**，基座齐（4.1） |
| `pricing.override_file` | `593fc9365` | 3 CLEAN + 1 CONFLICT | **P1**，个人部署性价比最高（4.2） |
| 渠道 `cache_write_1h_price` | `34b8bf1a6` 后半 | 需 SQLite 重写 +1 迁移 | **P1**（4.3） |
| 分组定价弹窗布局 | `1a33dc8cc` | 1 CLEAN + 3 CONFLICT | **P1**（4.4） |
| Fast 组策略两簇 | PR#6443 → PR#6444 | 38/44 CLEAN（+2 迁移） | 按需 → 不做（5.1） |
| 长上下文阶梯数据驱动 | `530fb20f2` + `e2cfaa46e` | 5 NOBASE（基座 573 行） | 不做（5.2） |
| Kimi 原生 Responses | PR#6463 | 14 CONFLICT + 5 NOBASE | **缺 CN 平台整支** → 不做（5.3） |
| API key 不合成 instructions | `e21b849a9` | 缺 `UsesOpenAICodexProtocol` | 不做（5.4） |
| 数据库启动重试 | `863667ce6` + `b6a7b8b7d` | 1 CONFLICT | **PG-only** → 不做（5.5） |
| 404 不被 429 覆盖 | `343858021` | 1 CONFLICT | **本仓库该缺陷不成立** → 不做（5.6） |
| Codex 图像能力保留 | `e2624fb65` | 3 **ALREADY** | **已合**（第 6 节） |
| ent 索引修正 / sponsors / VERSION | `05ea883e2` 等 4 条 | — | N/A（第 6 节） |

---

## 2. 本轮核查方法

### 2.1 通道

本轮直接用本仓库已有的 `upstream` remote，不再另开部分裸克隆（0.1.183 §2 那套仍可用）：

```bash
git fetch upstream --tags --prune          # 本轮新到 v0.1.185 / v0.2.0
git log --oneline --no-merges v0.1.184..v0.2.0
git rev-list --count --no-merges v0.1.184..v0.1.185   # 16
git rev-list --count --no-merges v0.1.185..v0.2.0     # 44
```

版本归属仍按 tag 祖先判定（`git merge-base --is-ancestor <sha> <tag>`），本轮 14 条候选逐条核过。
上游 release bot 仍是「先打 tag 再补 `chore: sync VERSION`」，所以区间里会出现名为
「sync VERSION to 0.1.184」的提交落在 `v0.1.184..v0.1.185` 里。

### 2.2 三态判定（按文件切开）

沿用 0.1.183 §2 的按文件切分，但本轮把「PR 整体」也纳入：上游把两个 Fast 功能各拆成
13–18 条微提交（`Add the group Fast migration` / `Persist the group Fast setting` …），
逐条 `apply --check` 得到的是一串互相依赖的中间态，毫无意义。**对这种功能簇要按 PR merge
取整体 diff**：

```bash
# 单 commit
git show <sha> -- <file> | git apply --check -
# PR 整体（merge commit 的第一父 → merge commit）
git diff <merge>^1 <merge> -- <file> | git apply --check -
```

四态输出：`CLEAN`（正向可打）/ `ALREADY`（反向可打 = 已在库里）/ `CONFLICT` / `NOBASE`
（本仓库没这个文件）。**`ALREADY` 这一态本轮第一次派上用场**，`e2624fb65` 三文件全 ALREADY，
直接省掉一次重复移植。

⚠️ 本轮的假信号（README 已有的几种之外）：

- **链上的 CONFLICT 是顺序问题，不是难度**：PR#6444（free Fast）30 处 CONFLICT 里几乎全部
  只因 PR#6443（group Fast）未落；PR#6425 的 9 处同理，因为它和 PR#6447 改同一批文件。
  排序之后冲突面会塌掉一大半，**不要拿未排序的冲突数当工作量**。
- **`CONFLICT` 可能是「缺基座」的伪装**：`343858021` 在 `no_account_error.go` 上 CONFLICT，
  真实原因是本仓库根本没有 `classifySelectionFailureError`——也就是没有那个 bug（§5.6）。
  凡 CONFLICT 都要看一眼被改的函数在不在，而不是直接归为「手工改」。
- **`CLEAN` 仍不保证能编译**：`bfe0a5a87` 两文件 CONFLICT 但基座 `openAIWSRelayActiveTurnID`
  同样缺失；反过来 `6d5f02784` 全 CLEAN 且 `supportsIdlePingWithoutReader`
  （`openai_ws_pool.go:466`）/ `idleDuration`（`:505`）都在。逐个 grep 才能分开这两种。

### 2.3 迁移号

上游本轮 4 条新迁移，**上游自己第二次撞号**（0.1.184 撞过 `231_*`）：

| 上游文件 | 本文归属 | 本仓库号 | 转换要点 |
|---|---|---|---|
| `232_group_reasoning_effort_over_limit.sql` | §4.1 | `225` | 单列 `ADD COLUMN`，去 `IF NOT EXISTS` |
| `232_channel_cache_write_1h_pricing.sql` | §4.3 | `226` | 4 条 `ADD COLUMN` + 4 条 `COMMENT ON COLUMN`（后者整段删） |
| `232_group_force_openai_fast.sql` | §5.1 | 若做 → `227` | — |
| `233_group_free_openai_fast.sql` | §5.1 | 若做 → `228` | — |

SQLite 不支持 `ADD COLUMN ... IF NOT EXISTS`，也没有 `COMMENT ON`。照本仓库既有写法转，
参考 `migrations/178_channel_image_input_price.sql`（带 `[sqlite-converted]` 头注）与
`migrations/223_group_model_pricing.sql`（裸 `ALTER TABLE ... ADD COLUMN`）。

**实际落地号与上表一致**：`225_group_reasoning_effort_over_limit.sql`（`45d8ffb0e`）、
`226_channel_cache_write_1h_pricing.sql`（`e79c15c6f`）。`227` / `228` 未占用。

---

## 3. P0 — 已核实本仓库有同一缺陷

按建议顺序排列（先小后大、先无冲突后需手改）。合完一项把状态改成「已合」并补 fork 的短 SHA。

**本节已按档固化为两个 OpenSpec change**（行为契约与验收看 change，逐条 patch site 只看本文）：

- **第一档** = §3.1–§3.7 → [`port-upstream-0.2.0-p0-fixes`](../../openspec/changes/port-upstream-0.2.0-p0-fixes/)
  （6 项，6 个 capability；无迁移、无前端改动）
- **第二档** = §3.8–§3.11 + §4 → [`port-upstream-0.2.0-p0-tail-and-p1`](../../openspec/changes/port-upstream-0.2.0-p0-tail-and-p1/)
  （8 项，7 个 capability；+2 迁移、大面积前端）

⚠️ 第二档必须在第一档合完之后开工——它多处 `CONFLICT` 只因第一档未落（§9.3）。
§5 六簇（第三档/第四档）**未立项，等确认**。

### 3.1 `e93e6368f` 调度快照投影裁掉透传开关（v0.2.0）—— 本轮唯一一条完整追通的活缺陷 —— **已合** `b04caf281`

**上游**：2 文件 / +49（`scheduler_cache.go` +7、`scheduler_cache_unit_test.go` +42）。

**本仓库缺陷链**（逐跳都核过，不是推测）：

1. `internal/repository/scheduler_cache.go:980` `filterSchedulerExtra` 是**白名单**，那份 key
   列表里**没有** `openai_passthrough` / `openai_oauth_passthrough`。
2. 同文件 `buildSchedulerMetadataAccount`（`:875`）用 `Extra: filterSchedulerExtra(account.Extra)`
   构造投影写入 `sched:meta:<id>`（前缀常量 `:21`），而 `filterSchedulerCredentials`
   **保留 `model_mapping`**。
3. `internal/service/account.go:1742` `IsOpenAIPassthroughEnabled()` 只认 `Extra` 上那两个键，
   投影里被裁掉 → 恒 false。
4. `internal/service/account.go:816` `IsModelSupported` 的第一行短路（issue #4936 的修复，注释就写着
   「透传账号会被 model_mapping 白名单错误排除出候选集」）因此**在快照路径上永不触发**，
   退回按残留白名单判定 → 账号被记为 `model_not_supported` 剔出候选。
5. `SchedulerSnapshotService.ListSchedulableAccounts`（`internal/service/scheduler_snapshot_service.go:210`）
   读的正是这份投影；转发阶段另从完整账号 hydrate，所以**症状是「单独测账号能通、走网关报
   no available accounts / 404」**。

**它在 SQLite 上确实跑**：`internal/service/wire.go:419` 明确写了「与其它 skipSQLiteBackgroundJobs
服务不同，快照 worker 必须在 SQLite 上照常」启动，不在 `skipSQLiteBackgroundJobs` 之列。

**patch site**：`filterSchedulerExtra` 的 `keys` 列表，在 `"openai_responses_supported"` 之后
插两行 + 上游那段注释。

⚠️ **不要顺手对齐整份列表**：本仓库该列表还缺上游有的 `codex_fingerprint_mode` /
`codex_fingerprint_seed`（本仓库是 `"openai_responses_supported"` 直接接 `"codex_5h_used_percent"`）。
那是**另一处分叉**，属指纹收敛功能，不在本条范围内，单独判。

**验收**：上游用例把投影 JSON round-trip 一遍（真实 `sched:meta` 写读路径）再断言模型门放行
非白名单模型。这条必须照抄，否则改了列表却没覆盖序列化那一跳。

### 3.2 `200b1406d` Anthropic `fallbacks` 未带 beta 时不剥离（v0.2.0）—— **已合** `e2d524d20`

**上游**：4 文件 / +404-14，**全 CLEAN**。`pkg/claude/constants.go` +11（beta 常量）、
`service/gateway_request.go` +77-14（sanitize 主体）、`service/bedrock_request.go` +35、
`gateway_fallbacks_sanitize_test.go` +295（新用例）。

**缺陷**：客户端现在会发 `body.fallbacks`，上游没有 `anthropic-beta: server-side-fallback-2026-07-01`
时直接 400 `fallbacks: Extra inputs are not permitted`。OAuth mimic 与默认 API-key beta 两条路径
都受影响。上游做法是**镜像已有的 `context_management` sanitize**，所以本仓库已有同形状的落点。

**为什么是 P0**：这是硬 400，对以 Claude 为主用法的部署（本 fork 的主场景）一发就废。且
4 文件全干净，是本轮性价比最高的一条。

### 3.3 `1dc0a0900` ctx_pool WS ingress 漏掉容量降载改写（v0.1.185）—— **已合** `8ada0bc7c`

**上游**：2 文件 / +202-1，全 CLEAN。

**基座在位**：`sanitizeOpenAICapacityShedErrorCodeForClient` 在
`internal/service/openai_gateway_passthrough.go:1131`。

**缺陷**：上游 `server_is_overloaded` / `slow_down` 被 ctx_pool 的 ingress 直写路径原样转发。
Codex CLI 按闭集判定，这两个码属致命集 → 打印 "Selected model is at capacity" 并**终止会话，
不退避重试**。本仓库该 helper 只在 HTTP/SSE（`openai_gateway_response_handling.go`）与
http_bridge（`openai_ws_http_bridge.go`）两条路径被调用，ingress 直写是唯一漏掉的一条——
同一账号同一时刻切到 http_bridge 就能正常退避。

**patch site**：`internal/service/openai_ws_forwarder_ingress.go:1039` 的 `writeClientMessage(upstreamMessage)`（上游是 `:1125`，本仓库该文件更短）。

⚠️ **必须写进独立的 `clientMessage` 变量，不能原地改 `upstreamMessage`**：写出点之后的
`markOpenAIWSClientVisibleFailure` 与 `handleOpenAIWSTerminalTransientFailure` 仍要按**未改写的
原始 payload** 判定账号状态。这正是那个 helper 注释里写明的前提，也是 http_bridge 的既有写法。

### 3.4 `6d5f02784` 不支持无 reader ping 的空闲 WS 连接不被回收（v0.1.185）—— **已合** `b9e5e6361`

**上游**：2 文件 / +42-2，全 CLEAN。基座都在：`supportsIdlePingWithoutReader`
（`internal/service/openai_ws_pool.go:466`）、`idleDuration`（`:505`）。

**缺陷**：coder/websocket 没有 reader 时消费不了 pong 帧，空闲 socket 会挂到上游 keepalive
窗口过期才被发现，取出来就是坏连接。上游新增 `openAIWSConnIdleRecycleAfter = 90s`，在
`cleanupAccountLocked`（`:1356`）里对「未租出 + 无 waiter + 不支持无 reader ping + 空闲 ≥ 90s」
的连接提前逐出并计入 `scaleDownTotal`。

**patch site**：常量块 + `cleanupAccountLocked` 里 `maxAge` 判定之前插一段。注意上游同时对齐了
常量块的缩进，别把那部分当成语义改动。

### 3.5 `ba345f105` Codex 目录不跳过持久禁用的账号（v0.1.185）—— **已合** `d3b90855d`

**上游**：5 文件 / +146-30，**全 CLEAN**。§5.1 Codex routed catalog 整簇上一轮已合，基座齐。

**缺陷**：拉 Codex 模型目录时仍会挑到被持久禁用的账号，目录请求失败或拿到过期快照。

### 3.6 `57c76584a` Codex fast 模型不透出 priority service tier（v0.1.185）—— **已合** `623c555dd`

**上游**：2 文件 / +81-4，全 CLEAN。同 §3.5，基座在 `openai_codex_models_service.go`。

**注意归类**：这条只改**目录里怎么描述模型**，不引入 §5.1 那套 Fast mode `service_tier`
发送逻辑，所以和 0.1.180 §7.2 未合无关，可以单独吃。

### 3.7 `9eabd2a5b` + `e7c029875` 账号统计成本维护第二份定价实现（v0.1.185）—— **已合** `e9a1c427a`

**上游**：`9eabd2a5b` 1 文件 / +8-20（CONFLICT），`e7c029875` 1 测试文件 / +3-2（CLEAN）。

**冲突原因只有一个字面量**：本仓库 `internal/service/account_stats_pricing.go` 的
`tryModelFilePricing` 与上游 pre-image 逐字节一致，**只少条件里的 `normalizedTier == "fast"`**
（上游 9261dd77 加的，本仓库没取）。这条改动直接把整个手算分支删掉、改调
`CalculateCostWithServiceTier`，冲突随之消解，净 -12 行。

**本仓库的实际影响**：`normalizeBillingServiceTier`（`internal/service/billing_service.go:122`）
只做 `ToLower` + `TrimSpace`，**不把 `fast` 归一成 `priority`**。所以 `service_tier=fast`
在账号统计里确实会落到手算分支，按标准价计。上游那条 DeepSeek 峰谷价对本仓库 N/A
（没有 DeepSeek 平台），但「第二份实现每加一个定价特性都要手工镜像一次」这个结构问题成立。

**基座在位**：`CalculateCostWithServiceTier`、`shouldApplySessionLongContextPricing`
（`billing_service.go:1413`）都在。

### 3.8 `1be69e56a` + `421a83282` bootstrap 请求缺 call id 被拒（v0.1.185 + v0.2.0）—— **已合** `f3df9e86b` → `1c38b6800`

**上游**：`1be69e56a`（delegation bootstrap）2 文件 / +408，**全 CLEAN**；
`421a83282`（scheduled automation bootstrap）2 文件 / +225-3。

**成链，必须按序**：`421a83282` 在 `internal/handler/openai_gateway_handler.go` 上 CONFLICT
**只因为** `1be69e56a` 未落——后者新增的 +232 行正是前者的上下文。先落 `1be69e56a` 再落
`421a83282`，第二条就干净了。

### 3.9 `504919a05` API key 会话缓存身份丢失（v0.2.0）—— **已合** `3bc098ee2`

**上游**：4 文件 / +159-4，3 CLEAN（含 `openai_compat_prompt_cache_key.go` 的 1 行）+
1 CONFLICT（`openai_gateway_chat_completions.go`，+26-2）。

**基座在位**：`internal/service/openai_compat_prompt_cache_key.go`（181 行）。

### 3.10 `bfe0a5a87` 终止事件前收到 close 被判成功（v0.2.0）—— 附 11 行基座 —— **已合** `527986900`

**上游**：2 文件 / +56-2，两文件都 CONFLICT。

**缺陷**：干净的 WebSocket close 只描述传输层握手。上游一旦开了 Responses turn，成功还要求
一个终止协议事件；早到的 1000/EOF 现在被当成 graceful，适配层于是在 turn 仍活跃时报
`relay_completed`。上游改法是：`graceful && 有活跃 turn` → 翻成失败并包一层
`"upstream websocket closed before terminal event: "`。

**缺基座，但很便宜**（按 README「基座要顺着符号回溯一层先量大小」那条量过）：
`openAIWSRelayActiveTurnID` 本仓库没有，上游是 11 行纯函数——遍历 `state.turnTimingByID`
找出 `== state.activeTurn` 的那个 responseID。`isDisconnectError` 已在
`internal/service/openai_ws_v2/passthrough_relay.go:985`。**基座跟这条一起落**，
不要单独先合（README「基座要随用它的簇一起移植」）。

⚠️ 上游同一函数在 `:785` 还有第二个调用点（`observed.responseID = openAIWSRelayActiveTurnID(state)`），
属另一条改动。只落本条需要的那一处，落完 grep 一遍本仓库有没有别处在手写同样的遍历。

### 3.11 `34b8bf1a6` 前半：Fable 缺 fallback 定价 + Fable 5.1（v0.2.0）—— **已合** `6bd7637c4`

上游这个 PR 名字叫「support Claude Fable 5.1」，实际塞了**两件不相关的事**，本文按 §3.11 /
§4.3 拆开。前半是模型面 + 兜底定价：

**（a）本仓库的现存空洞**：`internal/service/billing_service.go` 里 grep `fable` **零命中**——
**完全没有 Fable 的 fallback 定价**。Fable 计费目前只靠 LiteLLM 目录，目录缺条目就静默落到
`getFallbackPricing` 的系列匹配（`opus` / `sonnet` …都不命中）。上游这条同时补齐
`claude-fable-5` 与 `claude-fable-5-1` 两档（in $10 / out $50 / cache write $12.5 / 1h $20
per MTok；5.1 把 cache read 从 $1 降到 $0.25），并在 `getFallbackPricing` 里**把 fable 判定
放在 opus 判定之前**（`fable-5-1` 要先于 `fable-5`，否则 5.1 被吃掉）。

**（b）Fable 5.1 模型面**（全 CLEAN，纯机械）：`pkg/claude/constants.go` `DefaultModels`、
`domain/constants.go` 的 `DefaultAntigravityModelMapping` / `DefaultBedrockModelMapping`、
前端 `composables/useModelWhitelist.ts` 的 `claudeModels` / `antigravityModels` 与三处预设映射
（anthropic / antigravity / bedrock）。本仓库 `claude-fable-5` 这些位置全都有，照着加一行即可。

**这一半不需要迁移。** 需要迁移的是后半，见 §4.3。

---

## 4. P1 — 值得做，需手工改

本节四项与 §3.8–§3.11 一起构成**第二档**，已固化为
[`port-upstream-0.2.0-p0-tail-and-p1`](../../openspec/changes/port-upstream-0.2.0-p0-tail-and-p1/)。

### 4.1 推理档位两簇：PR#6447 → PR#6425（v0.2.0，+1 迁移）—— **已合** `45d8ffb0e` → `a3d5d36ba`

**基座齐**，这是本轮唯一一个「大功能但不缺基座」的：

| 基座 | 位置 |
|---|---|
| `ReasoningEffortMapping` 类型 | `internal/domain/reasoning_effort.go:5`（别名 `internal/service/group.go:16`） |
| 策略实现（233 行） | `internal/service/openai_reasoning_effort_policy.go`，含 `NormalizeMaxReasoningEffort` / `NormalizeReasoningEffortMappings` / `ApplyOpenAIReasoningEffortPolicyFromContext` |
| ent 字段 | `ent/schema/group.go:270`（`max_reasoning_effort`）、`:274`（`reasoning_effort_mappings`） |
| 前端 | `components/admin/group/ReasoningEffortPolicyFields.vue`、`views/admin/groupsReasoningEffort.ts` |

**（a）PR#6447 `aa7a811e6` 超限「拒绝 or 降级」** —— 47 文件 / +746-100，**38 CLEAN**。
9 处 CONFLICT 里 `openai_gateway_handler.go` 与 `openai_ws_forwarder_ingress.go` **是因为
§3.8 / §3.3 未落**，先做 P0 能吃掉一部分。剩下的手改面：

- `internal/repository/group_repo.go` —— 本 fork 是手写 SQLite SQL，**必然手改**，只加两列读写
- `internal/handler/composite_platform.go` + 用例 —— 本仓库 composite 分叉过
- `internal/service/openai_ws_v2_passthrough_adapter.go`
- `internal/handler/ops_error_logger.go` 的用例
- `backend/ent/runtime/runtime.go` —— **不要打这个补丁**，改 `ent/schema/group.go` 后
  `go generate ./ent` 并提交生成物（README 第 3 节）

迁移：`232_group_reasoning_effort_over_limit.sql` → 本仓库 `225`，单列 `ADD COLUMN`。

**（b）PR#6425 `7c01ec9be` 按模型限定档位映射** —— 15 文件 / +995-100，前端占一半。
**必须在 (a) 之后**：它 9 处 CONFLICT 基本都是和 (a) 改同一批文件造成的。上游这个分支还带了
`77729e272 merge: resolve conflicts with upstream/main` 和 `05ea883e2 fix(ent): correct group
field indexes after merge` 两条解冲突提交，**都不要合**——用 `git diff 3510aa22b^1 3510aa22b`
取净效果。

⚠️ 前端占一半，按 0.1.184 §11.1 那条教训：**移植前端必须真跑 vitest**，
`pnpm run typecheck` 抓不到 `is not defined`。

### 4.2 `593fc9365` `pricing.override_file`（v0.1.185）—— **已合** `5c7a9fe2b`

**上游**：4 文件 / +334，3 CLEAN（`config/config.go`、新用例、`deploy/config.example.yaml`），
仅 `internal/service/pricing_service.go` CONFLICT。

**基座在位**：`pricing.fallback_file` 已在 `internal/config/config.go:653`，默认值在 `:2318`；
`parsePricingData` 在 `internal/service/pricing_service.go:421`（挂载点）。

**功能**：稀疏 JSON 补丁按字段浅合并覆盖目录/回退数据，作为最高优先级数据源。
`applyPricingOverrides` 挂在 `parsePricingData` 入口（目录、回退、灾备三条路径语义一致，
只修补已存在条目，值为 `null` 则删字段）；`mergeOverrideOnlyModels` 在回退合并后把目录和回退
都没有的模型作为独立条目并入；未生效条目（模型名拼错）打 WARN 哨兵；override 文件缺失/损坏
只跳过合并，不影响目录加载。

**为什么对本 fork 价值高**：个人部署想关掉某个模型的长上下文阶梯，只要写
`long_context_input_token_threshold: 0`，不用自建价格镜像、也不用改 `resources/` 里那份 199 KB
的出厂快照（改了就和上游 diff 永久分叉）。**它独立于 §5.2 那个大重构**，可以单独落。

### 4.3 `34b8bf1a6` 后半：渠道 `cache_write_1h_price`（v0.2.0，+1 迁移）—— **已合** `e79c15c6f`（断言 `ff897d4d8`，gofmt/表单 `0af248bc8`）

**上游**：`billing_service.go` 的 `applyChannelTokenPriceOverrides`（+8-1）、
`repository/channel_repo_pricing.go` / `channel_repo_account_stats_pricing.go`、
`handler/admin/channel_handler.go`、`available_channel_handler.go`、前端
`admin/channel/{IntervalRow,PricingEntryCard,types}.ts(vue)` / `ChannelsView.vue` /
`modelPlaza/PlazaModelPricingTable.vue` / `api/admin/channels.ts`。

**本仓库现状**：服务层已有 `CacheWrite1hPrice` 字段（`internal/service/channel_plaza.go:16`，
读 LiteLLM 的 `cache_creation_input_token_cost_above_1hr`），但**渠道自定义定价表没有这一列**
（`migrations/` 里 `channel_model_pricing` 相关的最后一条是 `178_channel_image_input_price.sql`）。

**迁移**：`232_channel_cache_write_1h_pricing.sql` → 本仓库 `226`。上游是 4 条
`ADD COLUMN IF NOT EXISTS ... NUMERIC(20,12)`（`channel_model_pricing`、
`channel_pricing_intervals`、`channel_account_stats_model_pricing`、
`channel_account_stats_pricing_intervals`）+ 4 条 `COMMENT ON COLUMN`。**去掉 `IF NOT EXISTS`、
整段删 `COMMENT ON`**，把说明写成 SQL 注释（照 `178_*` 的 `[sqlite-converted]` 头注格式）。

⚠️ 语义要保住：`cache_write_1h_price` 为 NULL 时，单独的 `cache_write_price` **继续同时覆盖
两个 TTL 档**（上游注释 "Preserve the pre-split behavior"）。照抄那个 `if channelPricing.CacheWrite1hPrice == nil`
分支，别简化成无条件赋值。

⚠️ 按 README §5 / 0.1.184 §25.2：**SQL 常量的注释里不要写 PG 语法字面量**，
`sqlite_dialect_audit_test` 连注释一起扫。写迁移注释时不要把 `COMMENT ON COLUMN` 抄进去当反例。

### 4.4 `1a33dc8cc` 分组模型定价弹窗布局（v0.2.0）—— **已合** `e79c15c6f`（与 §4.3 同提交）

4 文件 / +27-10，1 CLEAN + 3 CONFLICT（`IntervalRow.vue`、`PricingEntryCard.vue`、`GroupsView.vue`）。
纯样式，和 §4.3 改同一批前端文件，**建议跟 §4.3 一起落**否则会来回冲突两遍。同样必须真跑 vitest。

---

## 5. 本轮不做

本节六簇 = **第三档（按需）+ 第四档（不合）**，按 CLAUDE.md「Upstream release intake」第 3 步，
**未立 OpenSpec change，等用户确认**。第三档是 §5.1 与 §5.2（基座/工作量已量清，只差「要不要」）；
第四档是 §5.3–§5.6（缺整支平台 / 需先核调用点 / PG-only / 本仓库缺陷不成立）。

### 5.1 Fast 组策略两簇：PR#6443 → PR#6444（+2 迁移）—— 按需，不是缺基座

| PR | 规模 | 三态 |
|---|---|---|
| #6443 `groups.force_openai_fast` | 44 文件 / +868-19 | **38 CLEAN** / 6 CONFLICT |
| #6444 free Fast 按标准价计费 | 38 文件 / +481-25 | 3 CLEAN / 30 CONFLICT / 5 NOBASE |

**#6444 的 30 处冲突几乎全是因为 #6443 未落**（两条是同一功能的前后脚），排序后会塌掉。
#6443 本身 38/44 干净，portability 好得意外；6 处冲突是可预期的那几处：
`repository/group_repo.go`（手写 SQLite SQL）、`service/api_key_auth_cache_impl.go`、
`service/openai_gateway_chat_completions.go`、一个 profit 用例、前端 `types/index.ts` 与
`GroupsView.vue`。ent 生成物同 §4.1，走 `go generate`。

**不做的理由是「按需」而非「做不了」**：本仓库已有相当多 `service_tier` 管道
（71 个文件命中 `service_tier` / `ServiceTier`，客户端别名 `fast` → `priority` 的归一在
`internal/service/openai_gateway_request_body.go:937` / `:1095-1136`），但 0.1.180 §7.2
那整套 Fast mode 仍未合，这两簇是它的**上层分组开关 + UI**。只有真在发 fast / priority 请求、
且需要「整组强制 fast」+「免费 fast 按标准价计费」这两个开关时才划算。

迁移：`232_group_force_openai_fast.sql` / `233_group_free_openai_fast.sql` → 若做则本仓库 `227` / `228`。

### 5.2 `530fb20f2` + `e2cfaa46e` 长上下文阶梯改数据驱动 —— 基座 573 行

**上游**：21 文件 / +1242-638。5 处 NOBASE，基座是 0.1.180 §7 未合的 context-tier /
time-of-day 定价那一支：

| 缺的文件 | 上游行数（@v0.1.184） |
|---|---|
| `internal/service/billing_context_schedule.go` | 470 |
| `internal/service/billing_token_cost_request.go` | 103 |

按 README「基座要顺着符号回溯一层、先量基座本身有多大」那条量过：**573 行，不是几十行的便宜货**
（0.1.184 §17 那批「四处全部低估」的情况这次不成立）。而且这簇是**重构 + 契约哨兵，不修任何
本仓库能观察到的缺陷**——上游的动机是「阶梯计价数据驱动后，目录条目字段缺失会直接变成计费偏差」，
而本仓库还没有数据驱动的阶梯，风险不存在。

**可以单独摘出来的一小块**：`e2cfaa46e` 里 `backend/resources/model-pricing/model_prices_and_context_window.json`
那 26 行（6 个 Gemini pro 条目补 `cache_creation_input_token_cost`）**单独 apply 干净**。
我核过本仓库那份出厂快照，10 个 gemini pro 条目的该字段全是 `null`。**但影响接近零**：
上游同一条的用例自己断言 `extractGeminiUsage` 与 Antigravity 两条转换**从不产出 cache_creation
token**（Gemini usageMetadata 没有这个类别）。归为数据卫生，不值得单独开一轮。

### 5.3 PR#6463 Kimi 原生 Responses 转发 —— 缺 CN 平台整支

**上游**：20 文件 / +282-91，14 CONFLICT + 5 NOBASE，仅 1 CLEAN。

**判据是平台面，不是文件数**：`internal/domain/constants.go:21-26` 只有 anthropic / openai /
gemini / antigravity / grok / composite，**`PlatformKimi` / `PlatformZhipu` / `PlatformDeepSeek`
不存在**。本仓库 26 个文件出现 "kimi" 全是按模型名走 openai_compat 的顺带处理
（`thinking_protocol.go`、`openai_model_mapping.go`、`gateway_forward_as_chat_completions.go` …），
不是平台。5 处 NOBASE 正是那一支的骨架：`cn_providers_test.go`、`adaptive_api_protocol_test.go`、
`account_test_service_cn_adaptive.go(+test)`、前端 `credentialsBuilder.cnAdaptive.spec.ts`。

这是上游 0.1.178 / 0.1.179 两轮的整支（自适应 API 协议 + 三家国产供应商 + 配额监控），
和 0.1.184 §1「Kimi/Zhipu/DeepSeek 不是平台」的结论一致。**单条 fix 吃不下，要立项就立整支。**

### 5.4 `e21b849a9` API key 不合成 instructions —— 缺 `UsesOpenAICodexProtocol`

**上游**：3 文件 / +5-6，改动只有 1 行——把
`if instructionsEmpty && !compatMessagesBridge && !nativeDeepSeekResponses` 加上
`&& account.UsesOpenAICodexProtocol()`。

**两处对不上**：

1. `UsesOpenAICodexProtocol` 本仓库没有（上游 `internal/service/account.go:1304`）。
2. `nativeDeepSeekResponses` 属 §5.3 那支，本仓库该条件是
   `internal/service/openai_gateway_forward.go:333` 的 `instructionsEmpty && !compatMessagesBridge`。

谓词本身可以自己写等价的，但**先要确认本 fork 的 API-key 账号确实会走到这条合成路径**
（`defaultCodexSynthInstructions` 在本仓库有三个调用点：`openai_codex_transform.go:1326`、
`openai_gateway_passthrough.go:68`、`openai_gateway_forward.go:334`）。这正是 README 那条
「上游修复只覆盖了它自己的调用点，本 fork 的主路径可能是另一条」——**核查要问「这个函数被谁调用」**。
不核清就改，会得到一个「合了但改错了路径」的结果。

### 5.5 `863667ce6` + `b6a7b8b7d` 数据库启动重试 —— PG-only，永久不做

上游函数注释原文：「retries only errors that indicate **PostgreSQL** is temporarily unavailable
during startup」，且 `internal/repository/ent.go` 里 import `github.com/lib/pq`。配套改的是
`deploy/docker-compose.yml`（等 PG 起来）与 `deploy/DOCKER.md` / `README.md`。

本仓库 `InitEnt`（`internal/repository/ent.go:42`）是为 SQLite 重写过的
（`openEntDriver` → `applyDBPoolSettings` → `prepareSchema` → `ensureBootstrapSecrets` →
`ensureConfiguredAdmin`），只 import `modernc.org/sqlite`。**本地文件没有「数据库还没起来」
这个失败模式**，按 README §4「本 fork 架构上不需要」归类，不是等排期。

### 5.6 `343858021` 404 `model_not_found` 不被 429 覆盖 —— 本仓库该缺陷不成立

**这条是本轮抓到的空转陷阱**，而且是 CONFLICT 伪装成的（§2.2）。

上游修的是 `classifySelectionFailureError`：它按 `model_rate_limited=(\d+)` 无条件把选号失败
升级成 429，包括 fallback 已经是权威 404 的时候。**本仓库没有这个函数**——
`internal/handler/no_account_error.go`（147 行）只有 `classifyNoAccountError`（`:59`）与
`classifyNoAccountErrorFromGin`（`:97`），`selectionModelRateLimitedPattern` 零命中，
handler 层唯一提到 `model_rate_limited` 的地方是用例里一句
`require.NotContains(t, cls.Message, "model_rate_limited")`。

**我们从没做过那个 429 升级，所以没有那个 bug。** 合了是纯空转（README 第 13 条）。
上游的 `noAccountErrorClassification.ModelNotFound` 字段本仓库有，不要因为「字段在」就以为
基座在——那个字段是 404/503 分类用的，和选号失败升级无关。

---

## 6. N/A / 已合

| 上游 | 判断 |
|---|---|
| `e2624fb65` fix(codex): preserve known image input capabilities（v0.1.185） | **已合**。三文件全 `ALREADY`（反向可打），并逐点核过后像：`accountCodexModelSupportsImageInput` 里的 `GetUpstreamModelMetadata` 短路在 `openai_codex_models_service.go:1045`，`convertOpenAIModelListToCodexManifestForAccount` 在 `:1860`，`completeAPIKeyCodexModelsManifestMetadata(body, completeAll bool, account *Account)` 在 `:2046`。它随 0.1.184 §5.1「按终态取文件」一并落地 |
| `05ea883e2` fix(ent): correct group field indexes after merge（v0.2.0） | N/A。上游解冲突产生的生成物（`ent/mutation.go`、`ent/runtime/runtime.go`）。本仓库改 `ent/schema` 后 `go generate ./ent`，不打这个补丁 |
| `77729e272` merge: resolve conflicts with upstream/main（v0.2.0） | N/A。同上，上游分支的解冲突提交 |
| `9e2d97f25` / `52c7d8834` chore: merge upstream v0.1.185 | N/A |
| `cc6a8e517` chore: update sponsors（v0.1.185）、`cd04848b9`（v0.2.0） | 不合 |
| `52374af94` sync VERSION to 0.1.184、`a2fb09260` → 0.1.185、`5097b3145` → 0.2.0 | 不合。本仓库用自有 `1.1.x` 编号（README 与 CLAUDE.md「Releases / tags」） |

---

## 7. 建议顺序

冲突面高度依赖顺序（§2.2），按下面走能把手改量压到最小：

**第一批（P0 小项，全部低风险）** — 已合于 `sync/upstream-20260902-p1`（叠在第二档之后，起点 `6527a33de`）

- [x] 1. §3.1 `e93e6368f` 调度快照透传开关 —— 2 行 + 1 个 round-trip 用例。**最高优先，它是活缺陷** `b04caf281`
- [x] 2. §3.2 `200b1406d` Anthropic `fallbacks` 剥离 —— 4 文件全干净 `e2d524d20`
- [x] 3. §3.3 `1dc0a0900` ctx_pool ingress 容量降载改写 `8ada0bc7c`
- [x] 4. §3.4 `6d5f02784` 空闲 WS 连接回收 `b9e5e6361`
- [x] 5. §3.5 `ba345f105` + §3.6 `57c76584a` Codex 目录两条 `d3b90855d` / `623c555dd`
- [x] 6. §3.7 `9eabd2a5b` + `e7c029875` 账号统计成本（净 -12 行）`e9a1c427a`

**第二批（P0 余项）** — 已合于 `sync/upstream-20260902-p1`（第一档未合，从 `3da1c2dd0` 开工）

- [x] 7. §3.8 `1be69e56a` → `421a83282`（**必须按序**）`f3df9e86b` / `1c38b6800`
- [x] 8. §3.9 `504919a05` API key 会话缓存身份 `3bc098ee2`
- [x] 9. §3.10 `bfe0a5a87` + 11 行 `openAIWSRelayActiveTurnID` 基座 `527986900`
- [x] 10. §3.11 `34b8bf1a6` 前半：Fable fallback 定价 + Fable 5.1 模型面 `6bd7637c4`

**第三批（P1，带迁移与前端）** — 已合

- [x] 11. §4.2 `593fc9365` `pricing.override_file`（独立，可提前）`5c7a9fe2b`
- [x] 12. §4.1 PR#6447 → PR#6425（**必须按序**，迁移 `225`）`45d8ffb0e` / `a3d5d36ba`
- [x] 13. §4.3 + §4.4 渠道 `cache_write_1h_price`（迁移 `226`）与定价弹窗布局 —— **一起落**，同批前端文件 `e79c15c6f`（`ff897d4d8` / `0af248bc8`）

**推后**：§5.1 Fast 组策略（等真用到）、§5.2 长上下文阶梯（基座 573 行）、§5.3 Kimi 整支、
§5.4 需先核清调用点。

---

## 8. 自测

每批之后至少跑：

```bash
cd backend
go build -o bin/server ./cmd/server
go test -tags=unit ./internal/repository/ ./internal/service/ ./internal/handler/...
golangci-lint run ./...
```

按项加跑：

| 项 | 必跑 |
|---|---|
| §3.1 | `go test -tags=unit ./internal/repository/ -run 'SchedulerCache' -count=1`（现有 `TestSchedulerCacheSnapshotUsesSlimMetadataButKeepsFullAccount` 就是投影用例，新增的 round-trip 断言挂在它旁边） |
| §3.7 | `go test -tags=unit ./internal/service/ -run 'AccountStatsCost\|AccountStatsRule\|CalculateStatsCost' -count=1`（用例在 `account_stats_pricing_test.go`，命名是 `TestCalculateStatsCost_*` / `TestMatchAccountStatsRule_*`，没有 `AccountStatsPricing` 这个前缀） |
| §3.11(a) | `go test -tags=unit ./internal/service/ -run 'FallbackPricing' -count=1`（现有 3 个，含 `TestGetFallbackPricing_FamilyMatching` —— fable 判定加在 opus 之前必须让它继续通过） |
| §4.1 / §4.3 / §4.4 | `cd frontend && pnpm run test:run`（**不是只跑 typecheck**，见 0.1.184 §11.1）+ `make test-frontend-critical` |
| §4.1 / §4.3（迁移） | 照本仓库约定，在 `backend/migrations/` 里加一个 `TestMigration225...WithSQLiteSyntax` / `TestMigration226...` 用例（现有样板：`TestMigration151AddsAccountAutoPauseExpiryPartialIndex`、`TestMigration154aAddsSparkShadowIndexesWithSQLiteSyntax`），然后 `go test ./migrations/ -count=1` + `go test -tags=unit ./internal/repository/ -run 'Migration' -count=1`；最后删掉本地 `*.db` 完整起一次，确认 `225` / `226` 真跑过 |
| §4.3（迁移注释） | `go test -tags=unit ./internal/repository/ -run 'TestProductionSQLUsesSQLiteDialect' -count=1`（`sqlite_dialect_audit_test.go`，**注释里的 PG 字面量也会被扫**，0.1.184 §25.2） |

⚠️ **上游 PR#6443 / PR#6444（§5.1 Fast 组策略那簇）改了 `internal/repository/migrations_schema_integration_test.go`，
这个文件在本仓库是 `//go:build integration && postgres`——永远不编译。** 那部分 hunk 直接丢，
不要为了让它「能打上」去动构建标签。同理 `api_key_repo_openai_fast_projection_integration_test.go`
这类 `_integration_test.go` 落地前先 `head -1` 看标签。

`TestLatestMigrationBaseline`（`internal/repository/migrations_runner_extra_test.go:61`）用的是
合成 `fstest.MapFS`，**不硬编码 224**，加迁移不需要改它。

改了 `ent/schema/group.go`（§4.1）之后：

```bash
cd backend && go generate ./ent && go generate ./cmd/server   # 提交生成物
```

---

## 9. 本轮新增的通用教训

### 9.1 「某份文档不含」≠「本仓库没有」

0.1.184 §1 写着「上游 main `200602b41`，之后 3 条本文不含」。那句说的是**那份文档的覆盖范围**，
而 `200602b41` 本身在 v0.1.185 内部，且其中 `e2624fb65` 早就随 §5.1 的「按终态取文件」落进来了。
本轮靠**反向 `apply --check`（`ALREADY` 态）** 一次识别，省掉一轮重复移植。

**做法**：开新一轮时先把候选集整个跑一遍反向 apply，再看正向。三态改四态的成本几乎为零，
收益是直接消掉「上一轮顺带合了但没记」的那批。

### 9.2 上游把一个功能拆成十几条微提交时，逐条 `apply --check` 是噪声

本轮两个 Fast 功能各被拆成 13–18 条（`Add the ... migration` / `Persist the ... setting` /
`Return ... in admin responses` …）。逐条判定得到的是一串互相依赖的中间态，冲突数没有意义。
**按 PR merge 取整体 diff**（`git diff <merge>^1 <merge>`）才拿得到真实工作量——
PR#6443 逐条看是一片冲突，整体看是 38/44 干净。

### 9.3 链上的冲突数不是工作量，排序前不要报数

PR#6444 的 30 处 CONFLICT、PR#6425 的 9 处、`421a83282` 的 1 处，**全部只因为前一条未落**。
本轮如果按未排序的冲突数估工期，会把「顺序」误报成「难度」，量级差好几倍。

### 9.4 `CONFLICT` 也可能是「缺基座 / 无此缺陷」的伪装

README 已有「`CLEAN` 不保证能编译」和「`CONFLICT` 不等于不该合」，本轮补第三种：
`343858021` 在目标文件上 CONFLICT，真实原因是**被改的函数本仓库根本没有**——也就是没有那个 bug。
**凡 CONFLICT 都要先 grep 被改函数在不在，再判「手工改」还是「不成立」。** 只看三态会把一条
空转项排进 P1。

### 9.5 上游一个 PR 里可能塞两件不相关的事，按功能拆而不是按 PR 合

`34b8bf1a6`（44 文件）标题只说 Fable 5.1，实际一半是渠道 `cache_write_1h_price` 拆分（带迁移）。
按 PR 整合会把一条不需要迁移的模型面改动和一条需要 SQLite 重写的定价改动绑死。
**看到 PR 里出现迁移文件而标题没提，就先拆。**

### 9.6 「上游为 X 修的」不代表「本仓库因 X 受益」，但结构收益仍可能成立

`9eabd2a5b` 上游的动机是 DeepSeek 官方峰谷价（本仓库 N/A，没这个平台），但它删掉的那份
「单价 × token」二次实现是真实的结构问题——本仓库正好卡在少一个 `"fast"` 字面量上。
**平台 N/A 时不要直接丢，先看它顺手修掉的结构问题在不在。**

### 9.7 `_integration_test.go` 落地前先 `head -1` 看构建标签

本仓库有一批测试文件是 `//go:build integration && postgres`，在这个 SQLite-only fork 里
**永远不编译**——`internal/repository/migrations_schema_integration_test.go` 就是一个，而上游 PR#6443 / PR#6444（§5.1 Fast 组策略）改了它。这类 hunk 无论 `apply --check` 报 CLEAN 还是 CONFLICT 都该直接丢，
**不要为了让它能打上去动构建标签**。

这是「测试文件三态是假信号」的第三种形态（前两种见 README：`_test.go` 冲突不代表产品代码对不上；
用例 apply 干净不代表能编译）。判定成本极低：`head -1 <file>`。

### 9.8 复述旧 backlog 时，"安全项" 要单独拎出来问，而且要重量而不是重述

本轮按 CLAUDE.md「Upstream release intake」第 5 步复述旧轮次遗留项时，把 0.1.180 §5.1
（dompurify `3.3.1` → `3.4.14`）单独拎出来问了一句——它是整个 backlog 里唯一带安全属性的挂起项，
和其余「按需 / 架构上不需要」不是一类。确认后重量的结果记在
[PORTING-0.1.180.md §12](./PORTING-0.1.180.md#12-51-dompurify-重新量2026-09-02只读核查未改任何文件)，
三条可迁移的教训：

- **升级成本要在临时副本里真量一遍，不要凭 diff 估。** 先确认 `pnpm install --lockfile-only`
  在未改动输入上产出逐字节相同的 lockfile（本次 439ms no-op），之后的 churn 才能全部归因于改动。
  这次量出来的 22 行里含一条原判没估到的收益：两份副本（直接 `3.3.1` + mermaid 的 `3.3.3`）
  会合成一份。
- **安全判据要把「输入可信」和「受众范围」分开。** 原判「输入都是管理员自填 ⇒ 单管理员部署可达性低」
  把两件事混成一句。净化跑在**访问者**的浏览器里，本仓库 7 个 `DOMPurify.sanitize` 调用点里
  `/legal/:documentId` 是 `requiresAuth: false`——受众根本不需要「对外发 key」这个前提条件。
- **清单里引用的 CVE 未必被验证过。** 原判整条理由建在 `GHSA-cj63-jhhr-wcxv` 上，我用四种原型污染
  形态按原样调用都没能复现。这不证明 advisory 是假的，但说明**「已核实缺陷」这个标签在安全项上
  和在功能项上不是一个标准**：功能项能 grep 到 patch site 就算核实，安全项还需要一个能跑的 PoC。
  没有 PoC 时，理由要换成可验证的那些（advisory 数量、调用点数量、暴露面），而不是断言可利用。
- **并入时要问「它会不会让前面几个阶段的验证失效」。** 这一项按用户决定并入第二档，但排在
  **最后一个阶段**——它一改 lockfile，前面三个动前端的阶段就都是在旧依赖树上测的。
  依赖类改动的排序判据和代码类不同：**看它使谁的验证过期，而不是看它跟谁有文件冲突**（它跟谁都没有）。
