# OpenSpec 归档清单

## 当前状态（2026-10-24 补充归档）

已归档 **19 个变更**，活动目录保留 **2 个第二档变更**；42 个[主规格](../../specs/)汇总 251 个非支付需求。

首轮整理提交 `c36a19b7a` 归档 15 批、610 项任务。用户随后明确确认“0.1.180 P0，0.2.0 第一档，0.2.5 第一档，0.2.9 第一档，都已经完成”，本次据此补归档这四批，共 158 项任务，其中 29 项旧未勾选任务按用户确认核销。当前档案共有 768 项已关闭任务。

四批的 `tasks-before-archive.md` 与 `verification-before-archive.md` 逐字保留原始状态；各自的 `completion-record.md` 记录用户确认、旧状态与快照哈希。本次完成确认更新当前归档状态，历史 PARTIAL、空白证据、版本号与实施 SHA 保留在快照中。

## 已归档

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

## 保留的活动变更

| 变更 | 任务状态 | 保留原因 |
|---|---:|---|
| [port-upstream-0.2.0-p0-tail-and-p1](../port-upstream-0.2.0-p0-tail-and-p1/tasks.md) | 128/128 | 2026-10-27 三项收尾完成，实现提交 `5b3e1a716`；完整门禁与四处真实浏览器验收通过，保留活动目录等待后续归档。见[当前验收](../port-upstream-0.2.0-p0-tail-and-p1/verification.md)。 |
| [port-upstream-0.2.5-tier2](../port-upstream-0.2.5-tier2/tasks.md) | 29/33 | 默认容量系数决策与迁移 227 相关任务尚未关闭。未纳入本次四批完成确认。 |

## 文档与冻结材料

- `source-baseline.md`、`source-feature-map.md`、`source-freeze/` 和原始证据保持原文；本次新增的归档前快照同样逐字保存。历史路径以原变更目录解释，各档案 README 提供当前材料入口。
- 可维护 Markdown 的链接已迁至归档路径；AGENTS.md、上游同步索引与 PORTING 的当前状态补记同步更新，冻结评估内容与来源 SHA 不改写。
- `port-upstream-0.1.183-p0-fixes` 的历史 `payment-balance-visibility`、0.2.5 第一档的 T11 续费弹窗和 T12 EasyPay 需求完整保存在各自档案中，按个人用途边界不纳入现行主规格。
- 本次整理不修改产品实现、依赖、迁移或 VERSION，不追加实现提交、测试输出或部署结论。

## 检查记录

- [首轮验证快照](./verification.json)：保留 `c36a19b7a` 前首轮 15 批归档时的清单、哈希和检查结果。
- [补充归档验证](./verification-followup-2026-10-24.json)：本次四批归档及当前全量主规格、任务、冻结材料、快照、链接与格式检查。
