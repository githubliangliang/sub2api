# 上游0.2.12 第一档

归档于 2026-10-24，原变更 `port-upstream-0.2.12-tier1`；[归档清单](../README.md)。以下版本、实施与部署状态保留为历史记录。

冻结的来源文档保持原文，其中历史相对路径以归档前目录为准；当前证据入口：[候选](../../../../docs/upstream-sync/evidence-0.2.12/candidates.tsv) / [四态](../../../../docs/upstream-sync/evidence-0.2.12/files.tsv) / [分档](../../../../docs/upstream-sync/evidence-0.2.12/decisions.tsv)。

状态：开发及验收完成（2026-10-02），实现提交 `405246789`，分支 `sync/upstream-0.2.12`。3簇/3来源PR。第二档S02/S03由用户确认需要后从第三档调整，本轮用户已授权实施两批。

- [proposal](./proposal.md)：收益与范围
- [source-baseline](./source-baseline.md)：冻结SHA
- [source-feature-map](./source-feature-map.md)：来源与行为
- [design](./design.md)：适配决策
- [spec](./specs/single-node-correctness/spec.md)：行为契约
- [tasks](./tasks.md)：已完成任务
- [verification](./verification.md)：逐场景验收证据
- [PORTING](../../../../docs/upstream-sync/PORTING-0.2.12.md)：完整分档与patch sites

保持SQLite/miniredis/simple mode及VERSION 1.1.15；不引入支付、TypeSafe、axios升级，不发布或部署。
