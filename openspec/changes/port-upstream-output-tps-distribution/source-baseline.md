# 冻结来源与实施起点

本文件在实施前冻结，不因验收结果回写。

- 来源：Wei-Shaw/sub2api PR #7911，完整 merge diff 为 `745ff1d3ca79a7f39520f627f6aa44d778a11580^1..745ff1d3ca79a7f39520f627f6aa44d778a11580`。
- 上游 merge 第一父：`8e2c4fad321fe8ce86d577dd7561c371bf602647`。
- fork 实施起点：`ed2340cc5`（完整 SHA 和补丁 SHA256 见 [baseline.json](../../../docs/upstream-sync/implementation-output-tps-distribution/baseline.json)）。
- VERSION1.1.22、Go1.27.3、最大 SQLite 迁移227；起点工作树无产品改动。
- 历史评估：[0.2.16](../../../docs/upstream-sync/PORTING-0.2.16.md)、[已选日志 TPS](../../../docs/upstream-sync/PORTING-0.2.18-personal-features.md)、[遗留复查](../../../docs/upstream-sync/UNMERGED-REVIEW-2026-10-31.md)。来源及原档位保留。
- 上游17文件/+402/-12包含已完成的日志展示；本次仅选择窗口运维统计消费者，不重做日志 TPS、不导入 PG 查询或 PG-only 夹具。
