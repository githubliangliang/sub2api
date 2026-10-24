# 上游0.2.16 第一档

归档于 2026-10-24，原变更 `port-upstream-0.2.16-tier1`；[归档清单](../README.md)。以下版本、实施与部署状态保留为历史记录。

冻结的来源文档保持原文，其中历史相对路径以归档前目录为准；当前证据入口：[候选](../../../../docs/upstream-sync/evidence-0.2.16/candidates.tsv) / [四态](../../../../docs/upstream-sync/evidence-0.2.16/files.tsv) / [决定](../../../../docs/upstream-sync/evidence-0.2.16/decisions.tsv)。

状态：**已实现并验收**（2026-10-14，实现提交 `7c544c03f`，实施起点 `0fbcab935`）。14簇/40个本轮来源PR的选定子集。用户选择“暂不增加，优先现有功能修复”。

- [proposal](./proposal.md)：收益与范围
- [source-baseline](./source-baseline.md)：冻结SHA
- [source-feature-map](./source-feature-map.md)：来源与行为
- [design](./design.md)：适配与边界
- [spec](./specs/existing-gateway-correctness/spec.md)：行为契约
- [tasks](./tasks.md)：实施任务与完成状态
- [verification](./verification.md)：实施验收证据
- [PORTING](../../../../docs/upstream-sync/PORTING-0.2.16.md)：完整评估、patch sites与基线测试

SQLite/miniredis/simple mode和1.1.16版本边界保留。本change只完成第一档；第二档待实施。未发布/部署，来源基线及评估证据冻结。
