# v0.2.24 评估证据（2026-11-21）

本目录记录**评估、既有基线、期望失败探针与隔离来源可行性**，不是移植完成证据。产品补丁没有应用；两份新OpenSpec保持未实施，37/52项任务未勾选、verification空白。主结论与patch sites在[PORTING-0.2.24.md](../PORTING-0.2.24.md)。

## 冻结与来源

- [initial-facts.json](initial-facts.json)：干净启动的main `d2488b1105f091b6fce4445395fec1835306fcf8`，VERSION1.1.26、Go1.27.4、SQLite230、auth21。artifact目录创建前的status为空；创建后的自身untracked不属于产品变化。
- [releases.json](releases.json) / [release-fetch.json](release-fetch.json)：官方API元数据、原始发布正文、HTTP头与获取时间。最新正式v0.2.24于2026-11-20 09:19:50Z发布；网页工具空返回、gh未安装，使用官方REST及git交叉核对。
- [baseline.json](baseline.json)：fork/tag commit与annotated对象分列；v0.2.24 commit `5c2982a81b46e2142d0143d2f83c1ce9c4da1d24`，tag对象 `7515db6dc56c2a6120dfb7abed70ef089ebc1457`。最后merge含GHSA-p3r4-jcm3-2hph相关6个提交。
- [commits.tsv](commits.tsv) / [candidates.tsv](candidates.tsv)：65非merge被39整PR、1普通直接提交及1无PR合并簇完整且不重叠覆盖。[patches/](patches/)保留全部完整diff及SHA256。
- [files.tsv](files.tsv) / [candidate-stats.tsv](candidate-stats.tsv)：同一private index上全部217条先reverse后forward；A0/C135/X62/N20，净diff190文件。只check，不apply。
- [post-release-commits.txt](post-release-commits.txt)：main仅多出328db26f4的VERSION同步，不作为产品来源。
- [dependency-and-migration-diff.txt](dependency-and-migration-diff.txt)：tag范围无迁移/依赖manifest/lock变更；未作新的依赖漏洞审计。

## 逐项判断与测试可达性

- [classification.json](classification.json)：41候选逐项决定，含#8962错误分类/5图扩展的独立边界。
- [review.json](review.json)：第一档10簇、第二档15簇，共30PR+1安全合并簇；逐簇证据、最小前置、行为、验收与限制。
- [added-call-inventory.tsv](added-call-inventory.tsv)：2269条新增调用检索，含内置/方法/测试；不是编译保证。
- [local-base-symbols.txt](local-base-symbols.txt)：实际生产声明行号；缺失符号须与PORTING逐簇适配说明一起读。
- [hunks.tsv](hunks.tsv) / [source-test-inventory.tsv](source-test-inventory.tsv)：来源hunk及85份来源测试记录。
- [integration-build-tags.json](integration-build-tags.json)：76个本地integration文件首行，53个postgres专用；未冒充SQLite测试。
- [older-active-changes.json](older-active-changes.json)：11旧活动change，350完成/0未完成；本轮不归档或改写。
- [baseline-test-list.log](baseline-test-list.log)：九包实际go test -list；[命令与结果](baseline-test-list.json)。

## 当前基线与缺陷探针

[probe-summary.json](probe-summary.json)按顶层与全部子例分开统计。

| 证据 | 结果与范围 |
|---|---|
| [baseline-protocol.json](baseline-protocol.json) / [完整log](baseline-protocol.log) | apicompat/httputil/claude/middleware/routes五包：700顶层通过、1原跳过；含子例1220通过/1跳过。 |
| [baseline-consumers.json](baseline-consumers.json) / [完整log](baseline-consumers.log) | service/handler/repository/securityaudit四包定向：169顶层、321含子例通过；实际SQLite230等控制。 |
| [baseline-frontend.json](baseline-frontend.json) / [完整log](baseline-frontend.log) | 3文件69测试通过；保留已有负例日志及Browserslist提示。 |
| [behaviour-probes.json](behaviour-probes.json) / [完整log](behaviour-probes.log) | 五包test-only overlay全部编译；13顶层目标失败、10控制/观察通过，无panic。包含HTTP入口未拦截、gjson/json分叉、SQLite55 vs18.18、OAuth轮换丢弃等。 |
| [behaviour-probes-additional.json](behaviour-probes-additional.json) / [完整log](behaviour-probes-additional.log) | 两包追加4顶层目标失败：body alias、Gemini尾随part、buffered hosted item和Composite thinking；全部编译，无panic。 |
| [frontend-bulk-probes.json](frontend-bulk-probes.json) / [完整log](frontend-bulk-probes.log) | 2个目标失败（OAuth/API Key切透传时发送filter=false），1个无关priority控制通过。临时测试结束后删除。 |

共17个Go目标顶层失败、10控制/观察通过；它们是有意按新行为写的隔离诊断，**不表示产品已经修好，也不等于全量backend门禁失败**。部分静态适用簇尚未逐一动态复现，PORTING已区分。初始探针夹具错误不纳入该17项。

[probes/](probes/)保存实际测试文本，Go通过虚拟overlay增补，不写入产品树；[probe-replay.json](probe-replay.json)保存稳定相对路径映射，可重建绝对Replace后执行记录命令。前端临时文件必须通过finally删除。

## 来源可行性与有限测量

- [source-scanner.json](source-scanner.json) / [完整log](source-scanner.log)：冻结tag的472行scanner、原test和从来源单独提取的multipart helper，以显式/tmp文件列表编译；11顶层/85含子例通过。没有引入完整requestmodel/model_allowlist到产品，也没有实际挂HTTP或WS入口。
- [source-scanner-benchmark.json](source-scanner-benchmark.json) / [完整log](source-scanner-benchmark.log)：GOMAXPROCS=1、约225KB固定夹具，工作站95399ns/op、1192B/op、14allocs/op。不是1CPU cgroup隔离，不是1GiB实例/并发/生产安全接入验收。
- [scheduler-projection-poc.json](scheduler-projection-poc.json) / [完整log](scheduler-projection-poc.log)：本地现有marshaler加来源投影思路，合成40组full payload 260839→5060字节，成员/模型判断/调用者图控制通过。不是生产样本或完整cache hydrate性能验收。

S13其余性能hunk未作前后完整benchmark，必须在实施时逐个验证；不得把来源数据或部分已有raw-view优化宣称成本轮新收益。

## 过程失败与局限

- 初始Go探针有括号拼写错误；图片builder参数与SQLite指针断言也需对齐本地API。保留[初始日志](behaviour-probes-initial.log)、[初始结果](behaviour-probes-initial.json)及[初始service文本](probes/service.initial.go.txt)/[初始repository文本](probes/repository.initial.go.txt)。修正只在证据中的测试夹具，不改产品。
- HTTP入口探针使用实际Gin路由及最小本地handler夹具；500证明未在入口按400拦截并到达未装配handler，结合parser分叉确认根因，**不是完整带真实凭据的利用或结算攻击演示**。未使用业务实例、真实账号或外部上游模型。
- 未跑新的全量backend/frontend、race、真实浏览器或完整1C1G实例；这些是未来实施任务。基线/源码可行性/未来验收必须分别表述。

## 生成与范围核验

- [collect-intake.py](collect-intake.py)：只读完整来源、private index先反后正、声明/测试扫描。
- [build-report.py](build-report.py) / [render-report.py](render-report.py)：显式审查决定及文档生成；只产评估与未实施提案。
- [run-check.py](run-check.py)：保留原命令/cwd/完整输出/退出码/时间，不把预期失败抹平。
- [protected-initial-hashes.json](protected-initial-hashes.json)：初始8311 tracked文件SHA256。
- [validate-intake.py](validate-intake.py) / [document-validation.json](document-validation.json)：最终产品/历史不变、HEAD/index、候选覆盖与patch、链接、OpenSpec严格结果、37/52未勾任务及空验收槽。
- [第一档严格校验](openspec-tier1.log) / [第二档严格校验](openspec-tier2.log)。
