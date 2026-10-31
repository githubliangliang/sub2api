# 用户选定：Grok 模型目录诊断

2026-10-31 用户选定实现Grok目录诊断，与历史工具/事件兼容性修复同时验收。该能力保持第三档，属于本次明确选择；评估来源与旧批次保持原文。

- [proposal](./proposal.md) / [design](./design.md)：范围和适配。
- [source-baseline](./source-baseline.md) / [source-feature-map](./source-feature-map.md)：冻结来源和行为。
- [规格](./specs/grok-model-catalog-diagnostics/spec.md)、[tasks](./tasks.md)、[verification](./verification.md)：实际实现与验收。
- [补充评估及patch sites](../../../docs/upstream-sync/PORTING-0.2.19-personal-compatibility.md)。

基于当前main `5250907f5b8294eabed3d3101d2c4bf79936e0f7`，保留模型参考价格提交、SQLite/miniredis/simple mode、VERSION1.1.21、Go1.27.3、迁移227和现有目录/CAS/ETag。仅在管理端账号详情和同步结果提供诊断，接入现有编辑页；其它未选能力独立。
