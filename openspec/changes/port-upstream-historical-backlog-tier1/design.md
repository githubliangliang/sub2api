# Context

真实实施起点为 v1.1.28 / `b9d9524c6cce8488edb2eaf899d771176f3b2414`。既有15份活动变更/456任务保持；本批已实施并完成验收，详见 verification。冻结来源及原评估状态不回写。

# Decisions

数据范围由现有时间/platform/group条件共同决定，模型名称不得附加无依据的gpt前缀限制。现有接口名可保持兼容；既有卡片的说明覆盖实际平台范围。复用后续已完成的时长加权TPS，禁止回退旧PR中的算术AVG。无迁移/依赖/Ent/Wire需求。

# Data and Resource Boundaries

保留SQLite-only、miniredis/simple与64MiB请求体/40MiB SSE行/8MiB暂存/120s idle。保留固定quota、usage dedup、独立用户费用/账号成本、Fast与requested/effective effort。

# Verification

评估的当前失败探针、基线通过、合成性能和历史验收分别记录。实施需新的真实起点、失败/通过日志、最终哈希及按变更范围执行的组合门禁；完成后才勾任务。
