# v0.2.19 第一档：Claude/Codex及既有用量、错误与管理输入修复

归档于 2026-11-01，原变更 `port-upstream-0.2.19-tier1`；26/26 项任务完成。见[归档清单](../README.md)。

当前入口：[主规格](../../../specs/) / [实际实施证据](../../../../docs/upstream-sync/implementation-0.2.19-tier1/README.md) / [归档前 README](./README-before-archive.md)。

冻结来源、方案、设计、delta 规格、任务与验收保持原文；历史相对链接以原目录 `openspec/changes/port-upstream-0.2.19-tier1/` 解释，并按归档清单映射到现址。以下原记录中的版本、日期、延期及部署状态保留其历史时点，导航链接已迁移。

状态：2026-10-29 用户授权后已实现并完成验收，**26/26 tasks 完成**。实现已提交本地 main：`683354b2cdd2461dc581fa14158be09ef51de29b`；[验收](verification.md)记录各项修复前失败、修复后反例和最终门禁。change 保持活动状态等待后续归档。2026-10-28 的来源及评估证据保持冻结。

| 编号 | 行为 | 规格 |
|---|---|---|
| F01 | Claude Code UTF-16指纹 | [claude-code-fingerprint](specs/claude-code-fingerprint/spec.md) |
| F02 | Codex模型缓存flight内复查 | [model-list-refresh](specs/model-list-refresh/spec.md) |
| F03 | 共享缓存命中用量 | [gateway-usage-accounting](specs/gateway-usage-accounting/spec.md) |
| F04 | Gemini fileData图片统计 | [gateway-usage-accounting](specs/gateway-usage-accounting/spec.md) |
| F05 | Antigravity畸形工具终态 | [antigravity-error-handling](specs/antigravity-error-handling/spec.md) |
| F06 | 分页数值与偏移量边界 | [admin-input-correctness](specs/admin-input-correctness/spec.md) |
| F07 | 管理员保护返回400 | [admin-input-correctness](specs/admin-input-correctness/spec.md) |

- [proposal](proposal.md) / [design](design.md)：范围与适配决策。
- [source-baseline](source-baseline.md) / [source-feature-map](source-feature-map.md)：冻结SHA与行为来源。
- [tasks](tasks.md) / [verification](verification.md)：已完成的实施任务和实际验收证据。
- [评估与patch sites](../../../../docs/upstream-sync/PORTING-0.2.19.md) / [评估证据](../../../../docs/upstream-sync/evidence-0.2.19/MANIFEST.md)。

保留SQLite-only、miniredis、simple mode、成功usage dedup、固定quota窗口及迁移227。VERSION1.1.20/Go1.27.3/迁移227、依赖及生成代码未变；实际起点与实现哈希见verification，不改历史证据。Terra是用户选定的独立第三档提案；其它未选定新能力、支付、发布与部署不在此范围。
