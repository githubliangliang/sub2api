# 上游0.2.18 第二档

归档于 2026-10-24，原变更 `port-upstream-0.2.18-tier2`；[归档清单](../README.md)。以下版本、实施与部署状态保留为历史记录。

冻结的来源文档保持原文，其中历史相对路径以归档前目录为准；当前证据入口：[四态](../../../../docs/upstream-sync/evidence-0.2.18/files.tsv) / [逐来源决定](../../../../docs/upstream-sync/evidence-0.2.18/decisions.tsv)。

状态：**已实现并验收完成**（2026-10-24）。10簇；用户选择现有功能修复优先，新增能力暂缓。

- [proposal](./proposal.md)
- [source-baseline](./source-baseline.md)
- [source-feature-map](./source-feature-map.md)
- [design](./design.md)
- [tasks](./tasks.md)
- [verification](./verification.md)
- [spec](./specs/existing-gateway-resilience/spec.md)
- [完整评估及patch sites](../../../../docs/upstream-sync/PORTING-0.2.18.md)

SQLite/miniredis/simple mode及固定quota窗口修复保留。10簇均已完成；unit/build/race/真实SQLite及前端检查通过，依赖遗留风险单列。当前未提交、未发布、未部署；详见verification。

后续实现提交：`2b79ac320`（两档共同提交）。发布文件VERSION已升1.1.19；[发布说明](../../../../docs/releases/v1.1.19.md)。原验收中的1.1.18及未提交状态描述保留为验收时快照。
