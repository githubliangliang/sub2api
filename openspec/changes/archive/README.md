# OpenSpec 归档清单

## 当前状态（2026-11-01）

已归档 **28 个变更**，活动变更为 **0**；74 个[主规格](../../specs/)汇总 356 个非支付需求。本次归档 9 个已完成变更、305 项任务，同步 105 个新增需求和 1 个既有需求修改。

首轮整理提交 `c36a19b7a` 归档 15 批、610 项任务。用户随后明确确认“0.1.180 P0，0.2.0 第一档，0.2.5 第一档，0.2.9 第一档，都已经完成”，2026-10-24 据此补归档四批、158 项任务，其中 29 项旧未勾选任务按用户确认核销；当时累计 768 项。本次九批已有完整任务与实施验收记录，累计 **1,073 项已关闭任务**。

2026-10-24 四批的 `tasks-before-archive.md` 与 `verification-before-archive.md` 逐字保留原始状态；各自的 `completion-record.md` 记录用户确认、旧状态与快照哈希。历史 PARTIAL、空白证据、版本号与实施 SHA 保留在快照中。本次九批的源文档、方案、设计、delta 规格、任务与验收逐字迁移，原 README 保存为 `README-before-archive.md`；当前 README 提供归档后的导航与实际实施证据入口。

## 2026-11-01 归档

| 变更 | 已关闭任务 | 实施提交与验收 |
|---|---:|---|
| [port-upstream-0.2.18-personal-features](./2026-11-01-port-upstream-0.2.18-personal-features/README.md) | 36/36 | `0e38d06d8` / [验收](./2026-11-01-port-upstream-0.2.18-personal-features/verification.md) |
| [port-upstream-0.2.0-p0-tail-and-p1](./2026-11-01-port-upstream-0.2.0-p0-tail-and-p1/README.md) | 128/128 | `5b3e1a716` / [收尾验收](./2026-11-01-port-upstream-0.2.0-p0-tail-and-p1/verification.md) |
| [port-upstream-0.2.5-tier2](./2026-11-01-port-upstream-0.2.5-tier2/README.md) | 33/33 | `b349c7408` / [余项验收](./2026-11-01-port-upstream-0.2.5-tier2/verification.md) |
| [port-upstream-0.2.18-personal-maintenance](./2026-11-01-port-upstream-0.2.18-personal-maintenance/README.md) | 16/16 | `b554bea55` / [验收](./2026-11-01-port-upstream-0.2.18-personal-maintenance/verification.md) |
| [port-upstream-0.2.19-tier1](./2026-11-01-port-upstream-0.2.19-tier1/README.md) | 26/26 | `683354b2c` / [验收](./2026-11-01-port-upstream-0.2.19-tier1/verification.md) |
| [port-upstream-0.2.19-tier2](./2026-11-01-port-upstream-0.2.19-tier2/README.md) | 23/23 | `cccf931ef` / [验收](./2026-11-01-port-upstream-0.2.19-tier2/verification.md) |
| [port-upstream-0.2.19-terra](./2026-11-01-port-upstream-0.2.19-terra/README.md) | 14/14 | `a541d0b61` / [验收](./2026-11-01-port-upstream-0.2.19-terra/verification.md) |
| [port-upstream-historical-client-compatibility](./2026-11-01-port-upstream-historical-client-compatibility/README.md) | 17/17 | `becd2d2a3` / [验收](./2026-11-01-port-upstream-historical-client-compatibility/verification.md) |
| [port-upstream-0.2.19-grok-model-diagnostics](./2026-11-01-port-upstream-0.2.19-grok-model-diagnostics/README.md) | 12/12 | `becd2d2a3` / [验收](./2026-11-01-port-upstream-0.2.19-grok-model-diagnostics/verification.md) |

## 2026-10-24 归档

| 变更 | 已关闭任务 | 完成依据 |
|---|---:|---|
| [add-openai-compatible-prompt-audit](./2026-10-24-add-openai-compatible-prompt-audit/README.md) | 153/153 | [原验收](./2026-10-24-add-openai-compatible-prompt-audit/verification.md) |
| [port-upstream-0.1.180-p0-fixes](./2026-10-24-port-upstream-0.1.180-p0-fixes/README.md) | 47/47 | [用户完成确认](./2026-10-24-port-upstream-0.1.180-p0-fixes/completion-record.md) |
| [port-upstream-0.1.183-p0-fixes](./2026-10-24-port-upstream-0.1.183-p0-fixes/README.md) | 47/47 | [原验收](./2026-10-24-port-upstream-0.1.183-p0-fixes/verification.md) |
| [port-upstream-0.2.0-p0-fixes](./2026-10-24-port-upstream-0.2.0-p0-fixes/README.md) | 56/56 | [用户完成确认](./2026-10-24-port-upstream-0.2.0-p0-fixes/completion-record.md) |
| [port-upstream-0.2.11-tier1](./2026-10-24-port-upstream-0.2.11-tier1/README.md) | 18/18 | [原验收](./2026-10-24-port-upstream-0.2.11-tier1/verification.md) |
| [port-upstream-0.2.11-tier2](./2026-10-24-port-upstream-0.2.11-tier2/README.md) | 24/24 | [原验收](./2026-10-24-port-upstream-0.2.11-tier2/verification.md) |
| [port-upstream-0.2.12-tier1](./2026-10-24-port-upstream-0.2.12-tier1/README.md) | 15/15 | [原验收](./2026-10-24-port-upstream-0.2.12-tier1/verification.md) |
| [port-upstream-0.2.12-tier2](./2026-10-24-port-upstream-0.2.12-tier2/README.md) | 15/15 | [原验收](./2026-10-24-port-upstream-0.2.12-tier2/verification.md) |
| [port-upstream-0.2.16-tier1](./2026-10-24-port-upstream-0.2.16-tier1/README.md) | 48/48 | [原验收](./2026-10-24-port-upstream-0.2.16-tier1/verification.md) |
| [port-upstream-0.2.16-tier2](./2026-10-24-port-upstream-0.2.16-tier2/README.md) | 54/54 | [原验收](./2026-10-24-port-upstream-0.2.16-tier2/verification.md) |
| [port-upstream-0.2.18-tier1](./2026-10-24-port-upstream-0.2.18-tier1/README.md) | 22/22 | [原验收](./2026-10-24-port-upstream-0.2.18-tier1/verification.md) |
| [port-upstream-0.2.18-tier2](./2026-10-24-port-upstream-0.2.18-tier2/README.md) | 34/34 | [原验收](./2026-10-24-port-upstream-0.2.18-tier2/verification.md) |
| [port-upstream-0.2.4-p0-fixes](./2026-10-24-port-upstream-0.2.4-p0-fixes/README.md) | 18/18 | [原验收](./2026-10-24-port-upstream-0.2.4-p0-fixes/verification.md) |
| [port-upstream-0.2.4-p1-stability](./2026-10-24-port-upstream-0.2.4-p1-stability/README.md) | 20/20 | [原验收](./2026-10-24-port-upstream-0.2.4-p1-stability/verification.md) |
| [port-upstream-0.2.5-tier1](./2026-10-24-port-upstream-0.2.5-tier1/README.md) | 34/34 | [用户完成确认](./2026-10-24-port-upstream-0.2.5-tier1/completion-record.md) |
| [port-upstream-0.2.9-tier1](./2026-10-24-port-upstream-0.2.9-tier1/README.md) | 21/21 | [用户完成确认](./2026-10-24-port-upstream-0.2.9-tier1/completion-record.md) |
| [port-upstream-0.2.9-tier2](./2026-10-24-port-upstream-0.2.9-tier2/README.md) | 22/22 | [原验收](./2026-10-24-port-upstream-0.2.9-tier2/verification.md) |
| [port-upstream-p1-tool-bridge-and-composite-dispatch](./2026-10-24-port-upstream-p1-tool-bridge-and-composite-dispatch/README.md) | 51/51 | [原验收](./2026-10-24-port-upstream-p1-tool-bridge-and-composite-dispatch/verification.md) |
| [resolve-pending-decisions-and-p1-fixes](./2026-10-24-resolve-pending-decisions-and-p1-fixes/README.md) | 69/69 | [原验收](./2026-10-24-resolve-pending-decisions-and-p1-fixes/verification.md) |

## 活动变更

无。其它未选定的按需能力仍按各轮 PORTING 记录处理，本次归档未将其纳入实现范围。

## 文档与冻结材料

- `source-baseline.md`、`source-feature-map.md`、`source-freeze/`、原方案/设计/delta 规格/任务/验收及原始证据保持原文；九份归档前 README 快照逐字保存。历史路径以原变更目录解释，各档案 README 提供当前材料入口；本次九批的原目录与现址映射见下方归档核对记录。
- 当前 README、AGENTS.md 与上游同步索引的入口已迁至归档路径。历史 PORTING、实施报告、旧档案和原始校验记录的链接与来源 SHA 保留原文，按原目录及归档映射解释。
- 主规格按实际完成状态同步：0.2.5 的 WS 默认保留 1.0，配额清理使用 SQLite 227；净化器约束纳入已批准的 S16 依赖维护与过期例外清理，不要求恢复旧告警或旧例外。原方案里的延期条件和安全审计样本仍在档案中。
- `port-upstream-0.1.183-p0-fixes` 的历史 `payment-balance-visibility`、0.2.5 第一档的 T11 续费弹窗和 T12 EasyPay 需求完整保存在各自档案中，按个人用途边界不纳入现行主规格。
- 本次整理只更新归档、主规格和文档入口；产品实现、依赖、迁移和 VERSION 不变。归档核对使用既有实施证据及本次文档校验，未重跑产品测试或追加部署结论。

## 检查记录

- [首轮验证快照](./verification.json)：保留 `c36a19b7a` 前首轮 15 批归档时的清单、哈希和检查结果。
- [2026-10-24 补充归档验证](./verification-followup-2026-10-24.json)：当时四批归档及主规格、任务、冻结材料、快照、链接与格式检查，原始值保留。
- [2026-11-01 归档核对](./verification-followup-2026-11-01.json)：九批归档、实施 SHA、主规格合并、任务、原文哈希、README 快照、历史链接映射与本次严格校验。
