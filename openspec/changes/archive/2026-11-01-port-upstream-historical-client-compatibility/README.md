# 历史兼容性尾项：Codex 工具与 Responses 事件

归档于 2026-11-01，原变更 `port-upstream-historical-client-compatibility`；17/17 项任务完成。见[归档清单](../README.md)。

当前入口：[主规格](../../../specs/) / [实际实施证据](../../../../docs/upstream-sync/implementation-historical-client-compatibility/README.md) / [归档前 README](./README-before-archive.md)。

冻结来源、方案、设计、delta 规格、任务与验收保持原文；历史相对链接以原目录 `openspec/changes/port-upstream-historical-client-compatibility/` 解释，并按归档清单映射到现址。以下原记录中的版本、日期、延期及部署状态保留其历史时点，导航链接已迁移。

状态：2026-10-31 用户授权后已实现并验收，17/17 tasks；本地main实现`becd2d2a31192e13bdc431bc43fbe44436e2c319`。活动目录等待归档，冻结来源和评估保持原文。

| 编号 | 行为 |
|---|---|
| C01 | 普通工具参数完整性、坏历史调用与匹配输出处理 |
| C02 | 无歧义自定义工具namespace别名还原与声明优先级 |
| C03 | 最终声明失去内建tool_search时清理孤儿defer_loading |
| C04 | Responses事件序号始终序列化且保留正确顺序 |

- [proposal](proposal.md) / [design](design.md)：范围与适配决策。
- [source-baseline](source-baseline.md) / [source-feature-map](source-feature-map.md)：冻结来源。
- [行为规格](specs/existing-gateway-correctness/spec.md)、[tasks](tasks.md)、[verification](verification.md)：未来实施与验收。
- [评估与patch sites](../../../../docs/upstream-sync/UNMERGED-REVIEW-2026-10-31.md)、[实际复查证据](../../../../docs/upstream-sync/unmerged-review-2026-10-31/README.md)。

保留SQLite/miniredis/simple mode、现有usage/dedup、固定quota窗口、VERSION1.1.21、Go1.27.3、迁移227和依赖/生成代码。Grok诊断、新模型、付款及其它未选能力独立。
