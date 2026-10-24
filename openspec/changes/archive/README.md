# OpenSpec 归档清单

## 2026-10-24 批次

从 `main` / `0c62a424b`（v1.1.19）整理。使用 `@fission-ai/openspec@1.14.3 archive <change> --yes --json`，归档 15 个任务全部完成的变更，共 610 项完成任务；6 个仍有未完成任务的变更保留在活动目录，任务勾选状态保持原样。

25 个[主规格](../../specs/)汇总 183 个需求与原有场景。`port-upstream-0.1.183-p0-fixes` 的历史 `payment-balance-visibility` 规格完整保存在档案中，遵循个人用途边界，不纳入现行主规格；归档不构成新增或启用支付功能的授权。

## 已归档

| 变更 | 已完成任务 | 验收 |
|---|---:|---|
| [add-openai-compatible-prompt-audit](./2026-10-24-add-openai-compatible-prompt-audit/README.md) | 153/153 | [验收](./2026-10-24-add-openai-compatible-prompt-audit/verification.md) |
| [port-upstream-0.1.183-p0-fixes](./2026-10-24-port-upstream-0.1.183-p0-fixes/README.md) | 47/47 | [验收](./2026-10-24-port-upstream-0.1.183-p0-fixes/verification.md) |
| [port-upstream-0.2.11-tier1](./2026-10-24-port-upstream-0.2.11-tier1/README.md) | 18/18 | [验收](./2026-10-24-port-upstream-0.2.11-tier1/verification.md) |
| [port-upstream-0.2.11-tier2](./2026-10-24-port-upstream-0.2.11-tier2/README.md) | 24/24 | [验收](./2026-10-24-port-upstream-0.2.11-tier2/verification.md) |
| [port-upstream-0.2.12-tier1](./2026-10-24-port-upstream-0.2.12-tier1/README.md) | 15/15 | [验收](./2026-10-24-port-upstream-0.2.12-tier1/verification.md) |
| [port-upstream-0.2.12-tier2](./2026-10-24-port-upstream-0.2.12-tier2/README.md) | 15/15 | [验收](./2026-10-24-port-upstream-0.2.12-tier2/verification.md) |
| [port-upstream-0.2.16-tier1](./2026-10-24-port-upstream-0.2.16-tier1/README.md) | 48/48 | [验收](./2026-10-24-port-upstream-0.2.16-tier1/verification.md) |
| [port-upstream-0.2.16-tier2](./2026-10-24-port-upstream-0.2.16-tier2/README.md) | 54/54 | [验收](./2026-10-24-port-upstream-0.2.16-tier2/verification.md) |
| [port-upstream-0.2.18-tier1](./2026-10-24-port-upstream-0.2.18-tier1/README.md) | 22/22 | [验收](./2026-10-24-port-upstream-0.2.18-tier1/verification.md) |
| [port-upstream-0.2.18-tier2](./2026-10-24-port-upstream-0.2.18-tier2/README.md) | 34/34 | [验收](./2026-10-24-port-upstream-0.2.18-tier2/verification.md) |
| [port-upstream-0.2.4-p0-fixes](./2026-10-24-port-upstream-0.2.4-p0-fixes/README.md) | 18/18 | [验收](./2026-10-24-port-upstream-0.2.4-p0-fixes/verification.md) |
| [port-upstream-0.2.4-p1-stability](./2026-10-24-port-upstream-0.2.4-p1-stability/README.md) | 20/20 | [验收](./2026-10-24-port-upstream-0.2.4-p1-stability/verification.md) |
| [port-upstream-0.2.9-tier2](./2026-10-24-port-upstream-0.2.9-tier2/README.md) | 22/22 | [验收](./2026-10-24-port-upstream-0.2.9-tier2/verification.md) |
| [port-upstream-p1-tool-bridge-and-composite-dispatch](./2026-10-24-port-upstream-p1-tool-bridge-and-composite-dispatch/README.md) | 51/51 | [验收](./2026-10-24-port-upstream-p1-tool-bridge-and-composite-dispatch/verification.md) |
| [resolve-pending-decisions-and-p1-fixes](./2026-10-24-resolve-pending-decisions-and-p1-fixes/README.md) | 69/69 | [验收](./2026-10-24-resolve-pending-decisions-and-p1-fixes/verification.md) |

## 保留的活动变更

| 变更 | 任务状态 | 保留原因 |
|---|---:|---|
| [port-upstream-0.1.180-p0-fixes](../port-upstream-0.1.180-p0-fixes/tasks.md) | 46/47 | 上线观察期尚未核销。 |
| [port-upstream-0.2.0-p0-fixes](../port-upstream-0.2.0-p0-fixes/tasks.md) | 55/56 | 合回 main 后复验及 SHA 回填任务尚未核销。 |
| [port-upstream-0.2.0-p0-tail-and-p1](../port-upstream-0.2.0-p0-tail-and-p1/tasks.md) | 109/128 | 阶段 9 与合回/回填任务仍未勾选；DOMPurify 已由 0.2.16 第二档 S16 接替，不重复实施旧任务。 |
| [port-upstream-0.2.5-tier1](../port-upstream-0.2.5-tier1/tasks.md) | 28/34 | 初始记录与真实浏览器验收有缺口；旧续费 UI 项不属于当前个人用途范围。 |
| [port-upstream-0.2.5-tier2](../port-upstream-0.2.5-tier2/tasks.md) | 29/33 | 默认容量系数决策与迁移 227 相关任务未完成。 |
| [port-upstream-0.2.9-tier1](../port-upstream-0.2.9-tier1/tasks.md) | 0/21 | 整批任务未核销；部分前置由后续批次完成，仍需按最新 PORTING §5 逐项确认。 |

旧任务尚未核销不等于产品功能全部未实现。本次只依据现有任务与验收记录归档，没有补做实现或替历史验证打勾；已由后续批次接替的行为应结合[最新遗留盘点 §5](../../../docs/upstream-sync/PORTING-0.2.18.md)继续核实。

## 文档与冻结材料

- 所有 `source-baseline.md`、`source-feature-map.md`、`source-freeze/` 和原始证据文件原样迁移；原始命令、SHA、版本号、验收与部署结论保留历史值。
- 冻结文档中的相对路径仍以归档前目录解释；每个档案的 README 提供对应当前证据入口。历史 JSON、命令输出和哈希清单中的旧路径保持原样。
- 可维护的 Markdown 链接已迁至归档路径；AGENTS.md、上游同步索引与各轮 PORTING 的链接同步更新，评估内容及分档不变。
- 3 处 delta 规格的说明文字移入对应 design.md，避免被归档器丢弃；需求与场景原文保持不变。主规格补齐用途说明。

## 本次检查

最终检查记录见[verification.json](./verification.json)。检查覆盖全部活动变更与主规格的严格校验、15 个档案的任务完整性、需求与场景保留、冻结材料哈希、移动前后链接映射，以及文档差异格式。此次仅整理 OpenSpec 和相关文档。
