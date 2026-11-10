# v0.2.20 评估证据

2026-11-10。主报告：[PORTING-0.2.20](../PORTING-0.2.20.md)。基线Fast已提交326513cdc / VERSION1.1.23 / Go1.27.3 / SQLite228。此目录为评估与缺口复现，**不是移植实施验收**。

| 文件 | 含义 |
|---|---|
| [baseline.json](baseline.json)、[initial-facts.json](initial-facts.json)、[initial-status.txt](initial-status.txt) | 实际HEAD、工作区→并发Fast提交的内容等价、private index tree及tag对象/commit区别 |
| [fetch.log](fetch.log)、[releases.json](releases.json) | fetch及最新正式版发布API快照 |
| [commits.tsv](commits.tsv)、[candidates.tsv](candidates.tsv)、patches/ | 72提交、42整候选父SHA、完整diff及SHA256 |
| [files.tsv](files.tsv)、[candidate-stats.tsv](candidate-stats.tsv) | 562文件全部reverse后全部forward；A3/C330/X157/N72 |
| [collect-intake.py](collect-intake.py) | 冻结来源及临时index检查，不应用补丁；读取当时发布API临时文件 |
| [hunks.tsv](hunks.tsv)、[added-call-inventory.tsv](added-call-inventory.tsv)、[base-symbols.txt](base-symbols.txt) | hunk、宽松新增调用扫描和本地实际符号检索；不能当作新代码编译证明 |
| [source-test-inventory.tsv](source-test-inventory.tsv)、[integration-build-tags.tsv](integration-build-tags.tsv) | 来源179测试文件及本地integration首行；PG测试不计SQLite验收 |
| [assessment-decisions.json](assessment-decisions.json) | 24行为簇与42候选四档判断 |
| [baseline-test-list.log](baseline-test-list.log)、[baseline-targeted.log](baseline-targeted.log) | 10包2630个匹配顶层测试入口发现，定向运行通过 |
| [performance-neighbor-list.log](performance-neighbor-list.log)、[performance-neighbors.log](performance-neighbors.log) | 2包RPM/过滤/委派/快照/隐私邻近发现与回归，不是性能收益数据 |
| [baseline-frontend.log](baseline-frontend.log) | 4文件77测试通过 |
| [reproduce-probes.py](reproduce-probes.py)、*-probe.go.txt、*-probe.log | 临时overlay：14个顶层新行为断言失败，cyber无transcript和auto模式控制通过；没有改backend文件 |
| [ws-probe-initial-compile.log](ws-probe-initial-compile.log) | 来源测试缺本地计时接口的初次失败；随后改为本地observer两轮测试，非产品失败证据 |
| [frontend-production-audit.log](frontend-production-audit.log)、[dependency-summary.json](dependency-summary.json) | 12moderate/2low；当前生产扫描，不继承旧xlsx high、不代表全量审计 |
| [older-scope-search.txt](older-scope-search.txt)、[post-release-commits.txt](post-release-commits.txt) | 历史消费者复查与tag后的9提交；后者不纳入本轮建议 |
| [run-check.py](run-check.py)、各检查同名JSON | 完整命令、cwd、耗时、时间、退出码 |
| [protected-before.json](protected-before.json)、[document-validation.json](document-validation.json) | 6561已有文件哈希、产品/历史不变、链接、来源、未勾选任务、严格OpenSpec及index核对 |

评估中的正常基线PASS不证明新功能实现；缺陷探针FAIL不是正常测试门禁失败。Bedrock只使用16字节畸形帧复现越界，未执行大分配或真实上游攻击。不存在的网络接口、后续main代码和业务实例均没有被操作。

采集脚本最初解析历史154a迁移名时失败，修正为数字前缀后采集成功；没有改迁移。全部过程产物仅此目录及新报告/提案，旧证据内容哈希保持。
