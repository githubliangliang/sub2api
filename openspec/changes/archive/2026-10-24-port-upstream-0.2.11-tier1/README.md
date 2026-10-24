# 上游0.2.11 第一档

归档于 2026-10-24，原变更 `port-upstream-0.2.11-tier1`；[归档清单](../README.md)。以下版本、实施与部署状态保留为历史记录。

冻结的来源文档保持原文，其中历史相对路径以归档前目录为准；当前证据入口：[候选](../../../../docs/upstream-sync/evidence-0.2.11/candidates.tsv) / [四态](../../../../docs/upstream-sync/evidence-0.2.11/files.tsv) / [分档](../../../../docs/upstream-sync/evidence-0.2.11/decisions.tsv) / [跨轮次前置](../../../../docs/upstream-sync/evidence-0.2.11/prerequisites.md)。

状态：开发与验收完成，改动位于 `sync/upstream-0.2.11` 分支。4簇 / 4个来源PR。

- [proposal](./proposal.md)：收益与范围
- [source-baseline](./source-baseline.md)：冻结SHA
- [source-feature-map](./source-feature-map.md)：来源与行为
- [design](./design.md)：适配决策
- [spec](./specs/single-node-correctness/spec.md)：行为契约
- [tasks](./tasks.md)：已完成任务
- [verification](./verification.md)：逐场景实施证据
- [完整评估与patch sites](../../../../docs/upstream-sync/PORTING-0.2.11.md)

第三/四档不在实现范围；本次保持VERSION 1.1.14，已完成本地Docker及远端部署，记录见verification.md；未发布GitHub Release。
