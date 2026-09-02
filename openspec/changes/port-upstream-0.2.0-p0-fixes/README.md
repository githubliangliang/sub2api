# port-upstream-0.2.0-p0-fixes

上游 v0.1.185 + v0.2.0 评估（[`docs/upstream-sync/PORTING-0.2.0.md`](../../../docs/upstream-sync/PORTING-0.2.0.md)）
划出的 **第一档**：6 项「缺陷已核实、基座符号在位、彼此无文件冲突」的小改动。
第二档见 [`port-upstream-0.2.0-p0-tail-and-p1`](../port-upstream-0.2.0-p0-tail-and-p1/)；
第三档（按需）与第四档（不合）**未立项，等用户确认**，理由见 PORTING §5。

| # | 内容 | 上游 | PORTING | 逐文件 apply 三态 |
|---|---|---|---|---|
| 1 | 调度快照投影裁掉透传开关 ⇒ 透传账号被误判 `model_not_supported` | `e93e6368f` | §3.1 | 1 `CONFLICT`（列表漂移）/ 1 `CLEAN` |
| 2 | Anthropic `fallbacks` 未带 beta 时不剥离 ⇒ 上游硬 400 | `200b1406d` | §3.2 | 4/4 `CLEAN` |
| 3 | ctx_pool WS ingress 直写路径漏掉容量降载改写 ⇒ Codex 就地终止会话 | `1dc0a0900` | §3.3 | 2/2 `CLEAN` |
| 4 | 不支持无 reader ping 的空闲 WS 连接不被回收 | `6d5f02784` | §3.4 | 2/2 `CLEAN` |
| 5 | Codex 目录挑到持久禁用账号 + fast 模型不透出 priority tier | `ba345f105` / `57c76584a` | §3.5 / §3.6 | 5/5 + 2/2 `CLEAN` |
| 6 | 账号统计成本维护第二份「单价 × token」实现 | `9eabd2a5b` + `e7c029875` | §3.7 | 1 `CONFLICT`（差一个字面量）/ 1 `CLEAN` |

⚠️ 四条必须先读的约束：

- **第 1 项只加两个 key，MUST NOT 顺手对齐整份白名单。** 本仓库 `filterSchedulerExtra` 还缺上游有的
  `codex_fingerprint_mode` / `codex_fingerprint_seed`，那是指纹收敛功能的**另一处分叉**，不在本批范围。
  见 `design.md` 决策 1。
- **第 1 项的用例必须覆盖 JSON round-trip**，否则改了列表却没覆盖序列化那一跳。见
  `specs/scheduler-snapshot-projection-fidelity/spec.md`。
- **第 3 项必须写进独立的 `clientMessage` 变量，不能原地改 `upstreamMessage`。** 写出点之后的
  `markOpenAIWSClientVisibleFailure` / `handleOpenAIWSTerminalTransientFailure` 仍要按未改写的原始
  payload 判定账号状态。见 `design.md` 决策 3。
- **第 6 项的 `CONFLICT` 不是「本仓库已修过」**，是本仓库条件里少一个 `"fast"` 字面量；上游这条把整个
  手算分支删掉，冲突随之消解。见 `source-baseline.md` §2。

阅读顺序：`proposal.md` → `source-baseline.md` → `source-feature-map.md` → `design.md` →
六个 `specs/*/spec.md` → `tasks.md` → `verification.md`。

逐条 patch site 与上游 diff 说明只在
[`docs/upstream-sync/PORTING-0.2.0.md`](../../../docs/upstream-sync/PORTING-0.2.0.md) §3，
本 change 不复制，只定义**移植后必须成立的行为**与验收证据。
