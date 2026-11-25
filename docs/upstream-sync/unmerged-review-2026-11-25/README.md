# 历史未合内容复评证据（2026-11-25）

对应[评估报告](../UNMERGED-REVIEW-2026-11-25.md)。这是评估、当前基线、隔离缺口探针和有限性能可行性证据；两个新OpenSpec批次尚未实施，验收槽为空。

| 资料 | 内容与边界 |
|---|---|
| [inventory.md](inventory.md) / [逐PR JSON](inventory.json) / [TSV](inventory.tsv) | 134个历史整PR，44类；含1第一档、1第二档、38第三档、2第四档和2关闭旧标签类别。不是134个新实施任务。 |
| [groups.json](groups.json) | 按独立行为分档、用途及缺基座说明；混合PR的已选子集/支付边界分别保留。 |
| [sources.json](sources.json) / patches目录 | 每PR merge/first-parent、完整来源大小及原文SHA256；整份补丁无损gzip。所有文件统一reverse后forward检查，未应用。 |
| [files.tsv](files.tsv) | 1,843条候选×文件状态A17/C1026/X654/N146和原始检查退出码/错误。 |
| [added-calls.tsv](added-calls.tsv) / [base-review.json](base-review.json) | 16,111新增调用索引和18类当前消费者查询。索引含第三方/标准库/测试/重名，不是编译或不存在行为的证明。 |
| [scope-expansion.json](scope-expansion.json) | 从84个重点PR扩至134，补v0.2.9数字编号第三档及模型/区域/强制策略项。 |
| [initial-hashes.json.gz](initial-hashes.json.gz) | 启动HEAD/index/refs、8,928份tracked工作区哈希与23份初始未跟踪文件。原上轮复核和Fast诊断受保护。 |
| [baseline-list.json](baseline-list.json) / [输出](baseline-list.stdout.log) | 四包测试目录先列出，确认实际可执行入口。 |
| [baseline.json](baseline.json) / [JSON流](baseline.stdout.log) | 当前四包定向98顶层/214含子例通过；含真实SQLite、miniredis、历史工具和共享序号控制。不是新的全仓验收。 |
| [integration-build-tags.json](integration-build-tags.json) | 当前integration文件首行，postgres约束不算SQLite覆盖。 |
| [frontend.json](frontend.json) / [计数](frontend-summary.json) / [输出](frontend.stdout.log) | 现有4文件66测试通过；stderr的401/失败刷新是预期测试日志，保留原文。 |
| [probes-final.json](probes-final.json) / [完整JSON流](probes-final.stdout.log) / [overlay](overlay.json) / probes目录 | 仅测试overlay，产品无改动；6目标顶层失败/14含子例失败，5观察或控制顶层通过/9含子例通过。两种已有维护缺口和四种可选能力分别分档。 |
| [初次探针](probes.json) / [原JSON流](probes.stdout.log) / [夹具说明](probe-fixture-note.json) | 初组还含`.test` DNS控制失败；保留初始源码。改为public literal IP的纯校验后控制通过，没有DNS或HTTP外部访问。 |
| [performance.json](performance.json) / [原输出](performance.stdout.log) / [摘要](performance-summary.json) | 单P三次100ms的本地SQLite total-only原型/当前路径及当前miniredis快照读取；不是完整优化候选/固定窗口/dedup/outbox或生产1C1G验收。 |
| [production-audit.json](production-audit.json) / [原JSON](production-audit.stdout.log) / [摘要](production-audit-summary.json) | 当前生产依赖12 moderate/2 low/0 high/critical，退出1有advisory；未做Go/开发全扫、利用性验证或改依赖。 |
| [collect.py](collect.py) / [run.py](run.py) | 整PR/私有index采集器与命令记录器；保留子命令真实退出码。 |
| [第一档strict](openspec-tier1.json) / [第二档strict](openspec-tier2.json) | 两份未实施提案的严格校验结果，不能当实现验收。 |
| [validation.json](validation.json) / [artifacts.json](artifacts.json) | 最终链接、来源哈希、未勾任务/空实施槽、旧15活动change/456完成任务、文件与HEAD/index/refs保护；本目录最终证据SHA256。 |

源码来源为仓库可读取的固定Git对象；网页工具未返回可读PR正文，不据此推断来源不存在。没有查询业务实例、提交/推送/发布/归档，也未运行新的全量unit/integration/race/build/浏览器或供应商验收。
