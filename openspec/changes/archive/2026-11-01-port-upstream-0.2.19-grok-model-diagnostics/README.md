# 用户选定：Grok 模型目录诊断

归档于 2026-11-01，原变更 `port-upstream-0.2.19-grok-model-diagnostics`；12/12 项任务完成。见[归档清单](../README.md)。

当前入口：[主规格](../../../specs/) / [实际实施证据](../../../../docs/upstream-sync/implementation-historical-client-compatibility/README.md) / [归档前 README](./README-before-archive.md)。

冻结来源、方案、设计、delta 规格、任务与验收保持原文；历史相对链接以原目录 `openspec/changes/port-upstream-0.2.19-grok-model-diagnostics/` 解释，并按归档清单映射到现址。以下原记录中的版本、日期、延期及部署状态保留其历史时点，导航链接已迁移。

2026-10-31 用户选定的Grok目录诊断已实现并验收，12/12 tasks，本地main实现`becd2d2a31192e13bdc431bc43fbe44436e2c319`。与历史工具/事件兼容性修复同时完成。该能力保持第三档，属于本次明确选择；评估来源与旧批次保持原文。

- [proposal](proposal.md) / [design](design.md)：范围和适配。
- [source-baseline](source-baseline.md) / [source-feature-map](source-feature-map.md)：冻结来源和行为。
- [规格](specs/grok-model-catalog-diagnostics/spec.md)、[tasks](tasks.md)、[verification](verification.md)：实际实现与验收。
- [补充评估及patch sites](../../../../docs/upstream-sync/PORTING-0.2.19-personal-compatibility.md)。

基于当前main `5250907f5b8294eabed3d3101d2c4bf79936e0f7`，保留模型参考价格提交、SQLite/miniredis/simple mode、VERSION1.1.21、Go1.27.3、迁移227和现有目录/CAS/ETag。仅在管理端账号详情和同步结果提供诊断，接入现有编辑页；其它未选能力独立。
