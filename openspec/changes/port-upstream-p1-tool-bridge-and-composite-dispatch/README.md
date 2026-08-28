# port-upstream-p1-tool-bridge-and-composite-dispatch

三批 P0（[0.1.180](../port-upstream-0.1.180-p0-fixes/) / [0.1.183](../port-upstream-0.1.183-p0-fixes/)）
与 [`resolve-pending-decisions-and-p1-fixes`](../resolve-pending-decisions-and-p1-fixes/) 之后的下一批：
**从剩余 P1/P2 backlog 里挑出 6 条候选，实际交付 5 条**（第 5 条实施中被证实不可独立移植，见下），
其余大簇（0.1.180 §6.2(c) Grok 稳定性、§7.1 大礼包、§7.2/§7.3、0.1.183 §4.1 监控 v2）**本批不做**。

| # | 内容 | 来源 | 逐文件 apply 三态 |
|---|---|---|---|
| 1 | CC 流式 tool_call 的空 `id` / `function.name` 被剥掉 | 上游 `cc894ef57`，0.1.180 §6.1 | 4/4 `ok` |
| 2 | `/v1/chat/completions` 的 `type:"file"`（PDF）不再被静默丢弃 | 上游 `4d4a0be1a`，0.1.180 §6.1 | 2 `ok` / 1 `CONFLICT` |
| 3 | HTTP bridge 不再重复回放整段工具历史 | 上游 `25da02ddd`，0.1.180 §6.1 | 产品码 4/4 `ok` |
| 4 | 回放时丢弃没有配对输出的孤儿 tool call | 上游 `66808413d`，0.1.180 §6.1 | 产品码 2/2 `ok` |
| ~~5~~ | ~~Grok compaction **422** 同号重试~~ **撤回，不交付** | 上游 `17c0ee385`，0.1.180 §6.2(c) | 1/1 `ok`，但**是 no-op**，见 `design.md` 决策 7 |
| 6 | composite 分组的 `/v1/messages` 闸门尊重分组自己的开关 | 上游 `68653fb2c`，0.1.180 §6.3 | 4 `ok` / 2 `CONFLICT` / 1 `NOFILE` |

⚠️ 四条必须先读的约束：

- **顺序固定 `25da02ddd` → `66808413d`**：后者在前者写出的判定之上再过滤孤儿 context item，
  单合后者会作用在旧的 `openAIWSRawPayloadHasToolCallOutput` 判定上，语义不同。见 `design.md` 决策 3。
- **第 6 项会放开一个此前恒关的开关**。`sanitizeGroupMessagesDispatchFields` 现在对一切非 openai
  平台强制 `AllowMessagesDispatch=false`；改完之后 composite 分组的这个开关**开始生效**，
  而 handler 里原先对 composite 的无条件豁免要同步收回，否则语义反而更松。见 `design.md` 决策 5。
- **第 6 项的上游 hunk 带 `IsCNProvider`**，本仓库无 CN 平台 ⇒ 该分支整段丢弃，不要照搬。
- **第 5 项已撤回。** 它 `apply --check` 干净、只有 1 行，但落地后是空转：真正判定「要不要剥掉
  加密 reasoning 重试」的 `isGrokInvalidEncryptedContentResponse` 在本仓库仍硬门 `400`，
  422 进了外层守卫也会在内层被否掉。widen 内层的是 §6.2(c) 的 `953028718`。见 `design.md` 决策 7。

阅读顺序：`proposal.md` → `source-baseline.md` → `source-feature-map.md` → `design.md` →
四个 `specs/*/spec.md` → `tasks.md` → `verification.md`。

逐条的 patch site 与上游 diff 说明在
[`docs/upstream-sync/PORTING-0.1.180.md`](../../../docs/upstream-sync/PORTING-0.1.180.md) §6，
本 change 不复制那份内容，只定义**移植后必须成立的行为**与验收证据。
