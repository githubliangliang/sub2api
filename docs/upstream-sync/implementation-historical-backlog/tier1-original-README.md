# 历史未合内容第一档：Token统计范围

状态：**仅评估提案，未实施**。2026-11-25基线v1.1.28 / `b9d9524c6cce8488edb2eaf899d771176f3b2414`。F01共1簇、5项未勾选任务。

现有统计在平台过滤之外硬编码gpt前缀，真实SQLite探针证实遗漏OpenAI非gpt模型，并令Claude/Grok平台选择返回空结果。旧第三档据此明确调整为第一档。

[评估及patch sites](../../../docs/upstream-sync/UNMERGED-REVIEW-2026-11-25.md) · [本轮评估证据](../../../docs/upstream-sync/unmerged-review-2026-11-25/README.md) · [冻结来源](source-baseline.md) · [任务](tasks.md) · [实施验收](verification.md) · [规格](specs/historical-backlog-token-stats/spec.md)。评估探针不是未来实施验收。
