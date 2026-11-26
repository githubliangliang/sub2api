# 历史未合第二轮评估证据（2026-11-26）

对应[报告](../UNMERGED-REVIEW-2026-11-26.md)。上批11任务已经完成并提交89060f532；本轮只评估下一批，第一/第二档3+3簇、13/14任务未实施，不能将失败探针写成实现验收。

| 证据 | 内容与限制 |
|---|---|
| [初始哈希 gzip](initial-hashes.json.gz) | 起始HEAD/index/refs、9,286份tracked和原Fast诊断哈希；仅原诊断未跟踪。 |
| [completion-hash-check.json](completion-hash-check.json) | 上一批18份最终实现源码/测试哈希匹配，复用旧验收不冒充本轮全量测试。 |
| [sources.json](sources.json) / patches目录 / [files.tsv](files.tsv) | 上轮134PR中的24个整PR；merge/parent、补丁SHA256、61条A0/C46/X13/N2原始检查。来源原文哈希与上轮一致，私有index先反后正，仅check。 |
| [decisions.md](decisions.md) / [JSON](decisions.json) | 24PR完整判断：第一档4PR、第二档4PR、第三档10PR、第四档6PR；按独立消费者说明旧档位调整。 |
| [added-calls.tsv](added-calls.tsv) / [consumer-review.json](consumer-review.json) | 549条新增调用索引及实际清单、picker、S3、用户/模板路由、原生消费者查询；CLEAN/缺同名函数均不单独证明行为。 |
| [backend-list.json](backend-list.json) / [目录原文](backend-list.stdout.log) | 实际service/repository/admin入口目录。 |
| [backend-baseline.json](backend-baseline.json) / [JSON流](backend-baseline.stdout.log) | 当前三包75顶层/121含子例通过，既有备份/加密/Codex清单/管理员模型列表，不是新全量验收。 |
| [backend-probes.json](backend-probes.json) / [JSON流](backend-probes.stdout.log) / probes目录 / [overlay](overlay.json) | 2个目标顶层失败/6含父子例失败，2个正常控制顶层通过。现有manifest生成与真实AES-GCM/Ent设置仓库/SQLite。SQLite只创建匹配settings需要的最小表，不冒称完整迁移/生产实例。没有S3调用或备份操作。 |
| [wire-controls-final.json](wire-controls-final.json) / [原文](wire-controls-final.stdout.log) | 实际属性handler经HTTP recorder：number整数12为400且不写，string整数12为200；窄service仓库stub，不当真实用户数据库验收。 |
| [frontend-baseline-summary.json](frontend-baseline-summary.json) / [执行记录](frontend-baseline.json) / [JSON](frontend-baseline.stdout.log) | 既有四文件15项通过。 |
| [frontend-probes-final-behavior-summary.json](frontend-probes-final-behavior-summary.json) / [执行记录](frontend-probes-final-behavior.json) / [JSON](frontend-probes-final-behavior.stdout.log) | 真实Vue组件11项：7个目标缺口失败、4个正常控制/观察通过；可控Promise/FileReader和API spy，不是浏览器或真实业务写入。卸载后旧属性响应不emit的已有保护通过。 |
| [fixture-failures.json](fixture-failures.json) / [首轮前端](frontend-probes.json) / [次轮](frontend-probes-final.json) / [选择器更正前](frontend-probes-corrected.json) / [原HTTP控制](wire-controls.json) | 临时TS缺括号、缺user.email、DOM选择器和小数不符合已有整数验证的失败逐条保留；保留各阶段源码/原始输出。只改夹具，不改产品。 |
| [frontend-probe-cleanup.json](frontend-probe-cleanup.json) / [run-frontend.py](run-frontend.py) | 原始测试在现有Vitest中临时执行，finally删除产品目录中的测试。 |
| [collect.py](collect.py) / [run.py](run.py) | 整PR/私有index采集及实际命令/退出码/UTC时间记录。 |
| [integration-build-tags.json](integration-build-tags.json) | 现有integration文件首行；postgres约束不当SQLite覆盖。本轮无新integration执行。 |
| [tier1严格校验](openspec-tier1.json) / [tier2严格校验](openspec-tier2.json) | 两份完整未实施提案检查。 |
| [validation.json](validation.json) / [artifacts.json](artifacts.json) | 最终链接、源补丁、哈希/范围、旧17活动change/467完成任务和新13/14未勾任务/空实施槽；本目录证据SHA256。 |

本轮未取最新Release、未重跑漏洞审计/全量unit/integration/race/build/浏览器/供应商或容量测试。网页工具没有返回可读PR正文，判断使用固定Git对象和仓库实际代码。没有commit/push/release/archive或业务实例操作。
