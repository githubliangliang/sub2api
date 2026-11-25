# 实施验收（2026-11-26 本地 / 2026-11-25 UTC）

状态：**S01 已实施，6/6 任务完成**。本次本地 commit 包含实现与验收；保持活动状态。真实实施起点 `b9d9524c6cce8488edb2eaf899d771176f3b2414`；冻结来源和评估证据保持原文。

| 验收项 | 实际证据 |
|---|---|
| 真实实施起点与范围 | [初始状态](../../../docs/upstream-sync/implementation-historical-backlog/initial-state.json.gz)；[完整实施记录](../../../docs/upstream-sync/implementation-historical-backlog/README.md)。只接入既有native JSON/SSE/聚合和本地单标记Ops合同，未导入上游WS多标记体系、重试、账号处罚或新quota策略。 |
| 缺口复现与修复后行为 | [原始重放](../../../docs/upstream-sync/implementation-historical-backlog/baseline-replay.json) 复现两种200错误信封不归因，正常停止对照通过；[修复后重放](../../../docs/upstream-sync/implementation-historical-backlog/baseline-replay-fixed.json) 通过。新parser/native测试覆盖错误/策略/真实空结果、OAuth包装/数组、优先级、STOP/MAX_TOKENS/普通工具与未知停止排除。 |
| 实际 HTTP/SQLite/Ops 控制 | [本地HTTP 16场景](../../../docs/upstream-sync/implementation-historical-backlog/http-sqlite-first.json)；[关停修正后三轮 race](../../../docs/upstream-sync/implementation-historical-backlog/ops-worker-shutdown-fixed.json) 通过3顶层/51含子例。上下游均真实TCP，SQLite230/Ent/miniredis及实际Ops队列落库/SLA查询；错误信封/空结果按语义状态，policy为client/request且不计SLA；原字节、OAuth既有解包/聚合、单次调用与原始body skip匹配保持。记录两次相同usage验证独立用户倍率2/账号倍率0.5、余额/quota/dedup及simple仅记录。 |
| 相关 unit/race 及组合门禁 | [8KiB暂存修正后的55包全量 unit](../../../docs/upstream-sync/implementation-historical-backlog/backend-unit-final.json)：10821顶层/22495含子例通过，16既有skip；随后通道修正的 [完整handler unit](../../../docs/upstream-sync/implementation-historical-backlog/backend-handler-unit-final.json)：819/1973通过、1既有skip；[最终四包 race](../../../docs/upstream-sync/implementation-historical-backlog/backend-race-complete.json)：318/717通过，包含SQLite固定quota、dedup、既有读写中断与账务控制。 |
| build/lint/资源及数据保护 | [最终lint 0](../../../docs/upstream-sync/implementation-historical-backlog/backend-lint-publish.json) / [embed build通过](../../../docs/upstream-sync/implementation-historical-backlog/backend-build-publish.json)。保持既有有界reader、取消/断连与已观察usage所有权；新非data兜底至多1MiB，matcher前缀8KiB、详情/消息2KiB；转义JSON键正常检测。VERSION/Go/schema/auth/依赖/Ent/Wire和64MiB/40MiB/8MiB/120s默认保持。 |
| 最终哈希与失败/局限说明 | [最终18文件 SHA256](../../../docs/upstream-sync/implementation-historical-backlog/implementation-files.sha256)、[保护核验](../../../docs/upstream-sync/implementation-historical-backlog/validation.json)。[证据说明](../../../docs/upstream-sync/implementation-historical-backlog/README.md) 保留实施暂存边界失败、Ops通道关停race、初次CLI PATH失败及最终通过记录。未执行PostgreSQL/供应商/生产容量验收，未push/发布/归档/操作业务实例。 |
