# v0.2.21 intake evidence

2026-11-11，fork `4f4c80c3ce48ce4afbc9344536b9cf44ed3fd898` / VERSION1.1.24 / Go1.27.3 / SQLite228 / auth snapshot21。**评估完成，产品未移植，OpenSpec仍为未实施提案。**

## 来源与范围

- [baseline.json](baseline.json)、[initial-facts.json](initial-facts.json)、[releases.json](releases.json)：tag对象/源码commit、release API元数据、fork和观察main。
- [commits.tsv](commits.tsv)、[candidates.tsv](candidates.tsv)、[patches/](patches/)、[collection.log](collection.log)：44非merge、33PR+1direct，完整PR diff及SHA256。
- [files.tsv](files.tsv)、[candidate-stats.tsv](candidate-stats.tsv)：214逐文件检查，同一私有index全部reverse后全部forward，A0/C121/X80/N13；没有应用patch。
- [hunks.tsv](hunks.tsv)、[added-call-inventory.tsv](added-call-inventory.tsv)、[base-symbols.txt](base-symbols.txt)、[minimal-prerequisites.md](minimal-prerequisites.md)：宽松调用扫描与人工核实的最小前置，不能当作编译成功。
- [source-test-inventory.tsv](source-test-inventory.tsv)、[integration-build-tags.tsv](integration-build-tags.tsv)：80份来源测试文件和76份本地integration文件首行；不自动去除postgres标签。
- [classification.json](classification.json)、[review.json](review.json)：全部34候选分档及23个拟议行为簇。
- [post-release-commits.txt](post-release-commits.txt)：tag之后的版本/赞助商提交，仅观察。

## 已执行评估

每条命令有同名JSON记录argv、cwd、完成UTC时间、耗时、退出码。实际执行通过`rtk proxy`保留原始输出；run-check设置已存在的Node24路径，不修改项目依赖。

| 检查 | 输出 | 结论 |
|---|---|---|
| Go测试发现 | [baseline-test-list.log](baseline-test-list.log) | 8包，exit0 |
| Go定向基线 | [baseline-targeted.log](baseline-targeted.log) | 8包，exit0；不是全量实施验收 |
| CLI身份基线 | [cli-identity-baseline.log](cli-identity-baseline.log) | Claude包/service相关测试通过 |
| 前端相关基线 | [baseline-frontend.log](baseline-frontend.log) | 5文件59测试通过；既有测试的预期错误日志不等于失败 |
| handler探针 | [internal-handler-probe.log](internal-handler-probe.log) | effort合成策略与UTF8早期边界预期失败；正常省略/显式策略/Unicode对照通过 |
| service探针 | [internal-service-probe.log](internal-service-probe.log) | metadata/sealed/seed/price/hash/slot/Gemini/image/custom policy/cooldown/refresh并发预期失败；现有Fast和成功补全对照通过 |
| apicompat探针 | [internal-pkg-apicompat-probe.log](internal-pkg-apicompat-probe.log) | foreign reasoning封装未过滤；原生签名通过 |
| antigravity探针 | [internal-pkg-antigravity-probe.log](internal-pkg-antigravity-probe.log) | sideband图片打断工具序列；原生媒体对照通过 |
| 探针摘要 | [probe-summary.json](probe-summary.json) | 4包均编译，16个顶层desired-behavior失败；这是缺口复现，不是产品已修 |
| SQLite索引可行性 | [sqlite-index-feasibility.log](sqlite-index-feasibility.log) | actual modernc，完整基线迁移最大228，20000合成记录，仓储结果一致；原INDEX→COVERING INDEX，非生产收益或写入成本结论 |
| 当前生产audit | [frontend-production-audit.log](frontend-production-audit.log)、[dependency-summary.json](dependency-summary.json) | 12moderate/2low/0high/0critical；exit1因advisory，未扫描Go/全开发依赖 |
| OpenSpec strict | [openspec-tier1.log](openspec-tier1.log)、[openspec-tier2.log](openspec-tier2.log) | 两提案验证通过，不代表实现 |
| 最终文档/范围 | [document-validation.json](document-validation.json) | 链接、任务/证据空槽、产品/历史哈希、补丁哈希、真实index |

## 复现与边界

- [reproduce-probes.py](reproduce-probes.py)使用临时Go overlay，把[probes/](probes/)中的测试映射为包内额外测试文件；不写入backend。当前旧基线测试预期退出1；脚本保留各包真实退出码，不把失败改写成PASS。 [probe-sources.json](probe-sources.json)记录来源与本地适配；service/apicompat/antigravity补充探针为本地编写，sqlite-index为本地可行性夹具。
- [reproduce-sqlite-index.py](reproduce-sqlite-index.py)只在临时SQLite中建索引；本地既有索引已支持Key/时间检索，本次验证第三列能覆盖actual_cost读取。
- [run-check.py](run-check.py)保存命令结果；[collect-intake.py](collect-intake.py)以完整PR收集和私有index检查。证据是冻结快照，重跑应复制到新的证据目录，避免覆盖历史。
- [internal-service-probe-initial-fixture.log](internal-service-probe-initial-fixture.log)：首次手写BillingService漏初始化map的测试夹具panic。修正夹具后才计入实际价格断言。
- [internal-service-probe-source-compile.log](internal-service-probe-source-compile.log)：来源测试调用缺失WithModel helper的编译失败。改用本地HTTP及四reader消费者重测；编译失败本身不算缺陷。
- Refresh并发探针使用真实service流程加锁保护的barrier stub，测得8次成功；这不替代未来真实miniredis原子消费验收。S01/S02/S03 WS新行为、Desktop新行为只做静态适用性核查，未来必须补实际本地HTTP/WS；没有对真实业务账号联机或运行新交互浏览器验收。
- UTF8探针使用缺model请求，证明非法字节进入后续字段检查；不据此声称发生真实远端重试或PG故障。F11未来需含model的真实handler零上游调用验证。
- [protected-files.json](protected-files.json)为开始时所有tracked常规文件内容hash；最终仅AGENTS.md与同步README的当前状态入口允许变化。产品、版本、依赖、schema及旧冻结证据全部核对。
- [SHA256SUMS](SHA256SUMS)记录本目录冻结内容（不含自身）。没有commit、push、发布或业务实例操作。
