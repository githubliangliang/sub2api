# 上游0.2.16 第二档

归档于 2026-10-24，原变更 `port-upstream-0.2.16-tier2`；[归档清单](../README.md)。以下版本、实施与部署状态保留为历史记录。

冻结的来源文档保持原文，其中历史相对路径以归档前目录为准；当前证据入口：[候选](../../../../docs/upstream-sync/evidence-0.2.16/candidates.tsv) / [四态](../../../../docs/upstream-sync/evidence-0.2.16/files.tsv) / [决定](../../../../docs/upstream-sync/evidence-0.2.16/decisions.tsv)。

状态：**已实现并验收**（2026-10-18；`847db09e1` + `5ff6e5bf9` 已快进合入 main，证据见 verification）。16簇/30个本轮来源PR的选定子集。用户选择“暂不增加，优先现有功能修复”。

- [proposal](./proposal.md)：收益与范围
- [source-baseline](./source-baseline.md)：冻结SHA
- [source-feature-map](./source-feature-map.md)：来源与行为
- [design](./design.md)：适配与边界
- [spec](./specs/single-node-correctness/spec.md)：行为契约
- [tasks](./tasks.md)：实施任务与逐簇进度
- [verification](./verification.md)：实施验收证据
- [PORTING](../../../../docs/upstream-sync/PORTING-0.2.16.md)：完整评估、patch sites与基线测试

SQLite/miniredis/simple mode 保留。实施起点为 `882fe9777` / VERSION 1.1.17；冻结来源与实施验收版本保留原值。用户随后要求发布 v1.1.18，VERSION 已更新为 1.1.18；[详细发布说明](../../../../docs/upstream-sync/implementation-0.2.16-tier2/release-v1.1.18.md)。逐簇验证与最终门禁分别记录，未部署业务实例。
