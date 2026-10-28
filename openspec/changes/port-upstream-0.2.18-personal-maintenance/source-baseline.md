# Source Baseline（冻结）

冻结日期：2026-10-28。此文件不在实施中回填新的 HEAD、VERSION 或来源 SHA。

| 项目 | 冻结值 |
|---|---|
| fork HEAD | `5c01d92c7cad41d6605925d076ee571fb420ee41` |
| VERSION | `1.1.20` |
| Go | `1.27.3` |
| 最大迁移 | `227` |
| 上游 v0.2.18 commit | `ac1b8f1271be0d674fa6f074403c1410f30a3a81` |
| #8249 merge | `b6c366380bd2639c939871a2cc8663cb4f5db32e` |
| #8249 first parent | `b7512305123aff307d3ab92840f2077eabc2529f` |
| #8250 merge | `b3e66f79a81400c729dbea78183f150509478343` |
| #8250 first parent | `b6c366380bd2639c939871a2cc8663cb4f5db32e` |

F01/F02 没有新的上游 PR 来源，依据本 fork 在冻结 HEAD 的生产调用链修复。F03 对每个 PR 取完整 merge diff，不按微提交串行 cherry-pick。

独立临时 index 固定在上述 fork SHA；全体11个候选文件先反向后正向校验，结果 ALREADY 0 / CLEAN 9 / CONFLICT 2 / NOBASE 0。工作树中的新回归测试不影响冻结 index。逐文件记录见 [source-file-states.tsv](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/source-file-states.tsv)。新增 helper 与现有消费者需同时接入；文件态不代替编译或运行验收。

历史 PORTING、旧 OpenSpec 来源及原始证据不改写。#8251 的共享快照/COW 和 #8269 的 PG JSONB 实现继续排除。
