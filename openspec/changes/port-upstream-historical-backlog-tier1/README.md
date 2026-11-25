# 历史未合内容第一档：Token统计范围

状态：**已实施并完成验收**。真实起点为 v1.1.28 / `b9d9524c6cce8488edb2eaf899d771176f3b2414`。F01 Token 统计范围 共1簇，5/5任务完成；保持活动状态，纳入本次本地 commit。冻结评估及来源保留原文。

现有统计在平台过滤之外硬编码gpt前缀，真实SQLite探针证实遗漏OpenAI非gpt模型，并令Claude/Grok平台选择返回空结果。旧第三档据此明确调整为第一档。

[评估及patch sites](../../../docs/upstream-sync/UNMERGED-REVIEW-2026-11-25.md) · [本轮评估证据](../../../docs/upstream-sync/unmerged-review-2026-11-25/README.md) · [冻结来源](source-baseline.md) · [任务](tasks.md) · [实施验收](verification.md) · [规格](specs/historical-backlog-token-stats/spec.md)。评估探针与本次实施验收分开记录。

[实际实施证据](../../../docs/upstream-sync/implementation-historical-backlog/README.md) · [最终产品/测试哈希](../../../docs/upstream-sync/implementation-historical-backlog/implementation-files.sha256)。未 push、发布、归档或操作业务实例。
