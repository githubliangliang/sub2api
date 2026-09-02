# 设计与决策

## 1. 移植策略

按 `docs/upstream-sync/README.md` 第 3 节：**按功能手工移植，不 cherry-pick 整个 merge commit**。
本批 6 项彼此无文件冲突，可各自独立提交、独立 revert（第 5 项两条共享一个文件但改不同位置，
第 6 项两条是产品码 + 用例，各自成对）。

四态证据与依赖符号核查在 `source-baseline.md`；逐条 patch site 在
`docs/upstream-sync/PORTING-0.2.0.md` §3。本文件只记决策。

## 2. 阶段划分

| 阶段 | 内容 | 提交粒度 |
|---|---|---|
| 1 | `e93e6368f` 调度快照投影 | 单独 |
| 2 | `200b1406d` Anthropic `fallbacks` | 单独 |
| 3 | `1dc0a0900` ctx_pool ingress | 单独 |
| 4 | `6d5f02784` WS 池空闲回收 | 单独 |
| 5 | `ba345f105` → `57c76584a` Codex 目录 | 两条各自单独 |
| 6 | `9eabd2a5b` + `e7c029875` 账号统计成本 | 一起（产品码与用例互为验收） |

阶段 1 排最前的理由：它是唯一一条**当前部署下就在发生**的故障，且改动量最小（2 行）。

## 3. 决策

### 决策 1：第 1 项只加两个 key，不对齐整份白名单

本仓库 `filterSchedulerExtra` 的 `keys` 列表相对上游还缺 `codex_fingerprint_mode` /
`codex_fingerprint_seed`（本仓库是 `"openai_responses_supported"` 直接接 `"codex_5h_used_percent"`）。
这正是那处 `CONFLICT` 的来源。

**不顺手补齐**：那两个 key 属指纹收敛功能，投影里要不要带它们取决于选号阶段有没有依赖指纹字段的
判定——本批没有核过这条，补了等于在没有缺陷证据的情况下改调度输入。按 README「守卫类单行改动的
空转陷阱」的反面：**没核清就不改**。单独判、单独立项。

落地方式：手写插入两行 + 上游注释，`CONFLICT` 不用模糊匹配硬塞。

### 决策 2：第 1 项的用例必须走 JSON round-trip，不能只断言函数返回值

`filterSchedulerExtra` 的返回值在 `buildSchedulerMetadataAccount` 里被塞进 `service.Account`，
再序列化进 `sched:meta:<id>`。只断言「函数返回的 map 里有这个 key」会漏掉序列化那一跳
（字段 tag、`omitempty`、类型断言都可能在那里丢东西）。上游用例把投影 JSON round-trip 一遍
再断言模型门放行非白名单模型，**这个形状必须照抄**。

### 决策 3：第 3 项必须引入独立的 `clientMessage` 变量

改写结果写入独立变量、`upstreamMessage` 保持原值。理由不是风格：写出点之后的
`markOpenAIWSClientVisibleFailure` 与 `handleOpenAIWSTerminalTransientFailure` 仍要按**未改写的原始
payload** 判定账号状态（要不要摘号、要不要冷却）。原地改会让「客户端看到可重试」与「账号被判为
可重试」两件事被同一份数据决定，把上游真实的容量降载信号从账号状态判定里抹掉。

这也是本仓库 http_bridge 路径的既有写法，`sanitizeOpenAICapacityShedErrorCodeForClient` 的注释
里写明了这个前提。**第三条路径要和前两条一致，不是新发明。**

### 决策 4：第 3 项的改写范围只收 `error` / `response.failed` 两种事件类型

上游只对这两种事件调 sanitize。**不扩大到全部事件**：非终止事件里出现同名字段的语义不同，
扩大范围会改动正常流量的 payload。验收要求双侧证据——容量类改写生效 **且** 非容量类错误码原样下发
（见 `specs/openai-ws-capacity-shed-client-view/spec.md`）。

### 决策 5：第 4 项的空闲阈值照抄上游 90s，不调参

上游取 `openAIWSConnIdleRecycleAfter = 90 * time.Second`，与既有的
`openAIWSConnHealthCheckIdle = 90 * time.Second` 同值，理由是「在上游 keepalive 窗口过期之前回收」。
本批不引入可配置项、不改数值——调参属运维，且没有本仓库侧的测量支撑。

### 决策 6：第 6 项接受「删掉一个此前被引用的 helper 调用」，但不删 helper 本身

`9eabd2a5b` 删掉手算分支后，`tryModelFilePricing` 不再调用 `GetModelPricing` 与
`shouldApplySessionLongContextPricing`。**这两个函数还有别的调用点，不得顺手删除**；实施时
`grep -rn` 确认后再决定是否需要处理 lint 的未使用告警（预期不会有，两者都是方法）。

### 决策 7：第 6 项的 DeepSeek 峰谷价收益判 N/A，但结构收益成立

上游 commit message 的动机是「`b5827cfd` 引入的 DeepSeek 官方峰谷一直没有生效，高峰时段账号成本
最多低估一半」。本仓库无 DeepSeek 平台 ⇒ 那部分是 N/A。**但不因此丢掉这条**：本仓库的
`normalizeBillingServiceTier`（`internal/service/billing_service.go:122`）只做 lower+trim，
不把 `fast` 归一成 `priority`，所以 `service_tier=fast` 在账号统计里确实会落到手算分支按标准价计。
删掉第二份实现同时解决这个和「每加一个定价特性都要手工镜像一次」的结构问题。

判据记录在 `docs/upstream-sync/PORTING-0.2.0.md` §9.6。

## 4. 明确不做（本批边界）

- **不碰 PORTING §3.8–§3.11 与 §4** —— 那是第二档，见
  [`port-upstream-0.2.0-p0-tail-and-p1`](../port-upstream-0.2.0-p0-tail-and-p1/)。
- **不碰 PORTING §5 六簇** —— 第三档/第四档，**等用户确认后再决定是否立项**。
- **不新增迁移**（迁移号保持 `224`）、**不改 ent schema**、**不动 `wire.go` / `wire_gen.go`**、
  **不改 `VERSION`**、**不动 `frontend/`**（含 `package.json` / `pnpm-lock.yaml` / `pnpm-workspace.yaml`）。
- **不补齐 `filterSchedulerExtra` 的指纹两 key**（决策 1）。
- **不引入 `service.IsCNProvider` 或任何 CN 平台分支**（本仓库无该平台面）。

## 5. 回滚

6 项各自独立提交 ⇒ 单条 `git revert` 即可。风险最高的两条的回滚判据：

- 第 1 项：若透传账号开始承接流量后出现异常（此前被 bug 意外挡住），**先查那个账号**，
  revert 只是把 bug 请回来。真要临时挡住该账号，用 `schedulable` 开关，不要 revert。
- 第 3 项：若客户端出现「本该终止却在无限重试」，说明改写范围过宽 ⇒ 收窄事件类型，而不是整条 revert
  （revert 会让 Codex 重新就地终止会话）。
