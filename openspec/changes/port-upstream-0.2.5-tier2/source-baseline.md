# 来源基线（实施期间不可改写）

评估日期：2026-09-19。第二档：收益明确、需要按行为移植的 12 簇 / 28 个候选。

| 基线 | 完整 SHA |
|---|---|
| 本 fork HEAD / v1.1.12 | fe6bc318800ca86d2b658f0e8f46e83d32626e8b |
| 上次上游终点 v0.2.4 | 5de5e2bed035d43591a2e10e51f420ef6a84eb98 |
| v0.2.5（本轮代码终态） | 86f93c28ee34cc74b629dafb748bd5ac5ca8c5ea |
| 拉取时 upstream/main，仅供观察 | efe9aab1e4ec89a42ba45e8dac20e882c5409a6a |

本仓库 VERSION=1.1.12，最大 SQLite 迁移为 226。
**除 S12（#6954）外本 change 不需要新迁移**；S12 单独等待用户确认，见 design.md 决策 1。
v0.2.5 之后 upstream/main 另有 28 条非 merge 提交，不计入本轮。

本仓库与上游无共同祖先。来源 PR 使用 `git diff <merge>^1 <merge>` 的整体 diff，不直接 merge 上游。
逐文件四态见 [files.tsv](../../../docs/upstream-sync/evidence-0.2.5/files.tsv)，
候选档位与理由见 [candidates.tsv](../../../docs/upstream-sync/evidence-0.2.5/candidates.tsv)。

四态固定在 HEAD 的独立临时 index（先全量反向 `git apply --cached --reverse --check`，再正向），
补丁不落工作树。工作树评估时除 `docs/upstream-sync/evidence-0.2.5/` 外无未跟踪改动。

实际实施起点、完成 SHA、测试与取舍填 verification.md；这里保持评估时的来源不变。
