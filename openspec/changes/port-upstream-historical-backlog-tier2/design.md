# Context

真实实施起点为 v1.1.28 / `b9d9524c6cce8488edb2eaf899d771176f3b2414`。既有15份活动变更/456任务保持；本批已实施并完成验收，详见 verification。冻结来源及原评估状态不回写。

# Decisions

复用MarkOpsStreamFailure、CountTowardsSLA、请求级归因及skip_monitoring；信号解析适配OAuth包装、数组、流优先级和有界非data兜底。错误信封/空结果与policy结果区分；STOP/MAX_TOKENS及其它普通生成停止不归因上游失败。仅登记最高优先级的一次结果，保持现有reader、断连、观察usage和取消所有权。来源新增完整体系不得替换本地现有Ops合同。无迁移/依赖/Ent/Wire或前端需求。

# Data and Resource Boundaries

保留SQLite-only、miniredis/simple与64MiB请求体/40MiB SSE行/8MiB暂存/120s idle。保留固定quota、usage dedup、独立用户费用/账号成本、Fast与requested/effective effort。

# Verification

评估的当前失败探针、基线通过、合成性能和历史验收分别记录。实施需新的真实起点、失败/通过日志、最终哈希及按变更范围执行的组合门禁；完成后才勾任务。
