# v0.2.25 评估证据（2026-11-24）

本目录是**来源、当前基线、期望失败探针和隔离依赖可行性**，不代表移植完成。主结论与 patch sites 见 [PORTING-0.2.25.md](../PORTING-0.2.25.md)。两份新 OpenSpec 保持未实施，6/11项任务未勾选，实施 verification 为空。

## 来源与冻结

- [initial-facts.json](initial-facts.json)：fork `39ca66d51d53b98f11349598bbd16cd8ecd0e2e0`，VERSION1.1.27/Go1.27.4/SQLite230/auth21。启动时已存在未跟踪 Fast 诊断，其哈希单独受保护。
- [protected-initial-hashes.json](protected-initial-hashes.json)：评估创建前8,653个 tracked 文件SHA256。只有 AGENTS.md 与同步README按流程新增入口信息。
- [latest-release.json](latest-release.json)、[官方列表](releases.json)、[HTTP头](latest-release.json.headers)、[获取说明](release-fetch.json)：官方最新正式版v0.2.25于2026-11-24 03:21:08Z发布。网页工具空返回，正文与git交叉核对，无可用网页citation ID。
- [完成前再次核对](latest-recheck.json)：2026-11-24 13:14:05 UTC请求官方latest端点，仍为同一v0.2.25及发布时间。
- [baseline.json](baseline.json)：fork/tree/private-index、旧新tag commit与annotated对象分列，main仅观察。新tag commit为348997267d756eab5729bb6b78a275a15d18796a，tag object为8fab111f92447a3edbcdf67a517e45b6a3cc4d31。
- [commits.tsv](commits.tsv)、[candidates.tsv](candidates.tsv)、[patches/](patches/)：12个非merge由3整PR+1 VERSION直接提交完整且不重叠覆盖，保存完整patch与SHA256。
- [files.tsv](files.tsv)、[candidate-stats.tsv](candidate-stats.tsv)：同一私有index，全部98条先reverse后forward，仅check；A0/C56/X32/N10，区别于91净变化文件。
- [post-release-commits.txt](post-release-commits.txt)：tag之后main仅VERSION同步；[依赖/迁移diff](dependency-and-migration-diff.txt)显示唯一Gorilla变更，无迁移/前端manifest/lock改动。

## 判断与可达性

- [classification.json](classification.json)：4完整候选的四档结论，混合PR按独立行为拆分。
- [review.json](review.json)：F01/S01/S02逐簇证据、最小前置、patch sites、行为/验收与边界。
- [added-call-inventory.tsv](added-call-inventory.tsv)：458条新增调用扫描，含标准库/第三方/测试；不证明编译。[local-base-symbols.txt](local-base-symbols.txt)保留生产声明与关键缺失项。
- [hunks.tsv](hunks.tsv)、[source-test-inventory.tsv](source-test-inventory.tsv)：来源hunk及24份来源测试文件首行、测试名和本地对应项。
- [integration-build-tags.json](integration-build-tags.json)：76份本地integration文件首行，53份含postgres约束，未记作SQLite覆盖。
- [older-active-changes.json](older-active-changes.json)：13份旧活动change，439已完成/0未完成；[older-base-review.json](older-base-review.json)重新扫描本地消费者和旧按需基座。
- [PR9052-production.txt](PR9052-production.txt)、[PR9008-repository.txt](PR9008-repository.txt)：所审关键生产完整diff摘录，完整范围仍以patches为准。

## 当前基线与缺陷探针

[probe-summary.json](probe-summary.json)区分顶层与含子例计数。

| 检查 | 结果与范围 |
|---|---|
| [baseline-test-list.json](baseline-test-list.json) / [log](baseline-test-list.log) | 六包真实go test -list完成；初始主定向middleware仅编译，随后单独执行auth。 |
| [baseline-consumers.json](baseline-consumers.json) / [log](baseline-consumers.log) | 主定向247顶层/495含子例通过，五包有匹配测试、六包编译。包含真实本地SQLite用量/配额/repository夹具。 |
| [baseline-auth.json](baseline-auth.json) / [log](baseline-auth.log) | middleware补充27顶层/53含子例通过；当前基线合计274顶层/548含子例通过。 |
| [baseline-frontend.json](baseline-frontend.json) / [log](baseline-frontend.log) | 4文件67测试通过；pnpm11位置提示保留，没有新增前端测试或顺带改配置/依赖。 |
| [behaviour-probes-final.json](behaviour-probes-final.json) / [log](behaviour-probes-final.log) | test-only overlay编译；5目标顶层失败、12含子例失败，4控制顶层通过，无panic。目标为配置泄露、确定性400字段、规则结果、compact code及仅原body监控匹配。 |
| [behaviour-probes.json](behaviour-probes.json) / [log](behaviour-probes.log) | 初组4目标失败/3控制通过；后续分开两种400消息并补原body/状态skip控制，原文本[service.initial.go.txt](probes/service.initial.go.txt)保留。 |
| [source-host-helper.json](source-host-helper.json) / [log](source-host-helper.log) | 最终来源helper和原测试在/tmp独立执行，1顶层/10含子例通过；没有接入生产writer。 |

[probes/service.go.txt](probes/service.go.txt)、[probe-replay.json](probe-replay.json)保存最终探针和相对映射。Go仅以虚拟test文件overlay增加诊断，没有替换产品实现。credential、域名和上游响应均合成；Forward经过实际服务构建/响应逻辑，httpUpstream使用既有测试recorder，不冒称外部供应商或完整生产实例。

来源helper保存为[source](probes/source-host-helper.go.txt) / [test](probes/source-host-helper_test.go.txt)，[重放说明](source-helper-replay.json)仅证明小函数可行性。

## 安全依赖与临时模块

- [go-vuln-6278.json](go-vuln-6278.json)、[gorilla-fix.json](gorilla-fix.json)、[gorilla-pinned.info](gorilla-pinned.info)、[security-fetch.json](security-fetch.json)：官方漏洞范围、crypto/rand修复及模块代理元数据/获取时刻/响应头。
- [gorilla-versions.txt](gorilla-versions.txt) / [获取记录](gorilla-versions-fetch.json)：正式版本列表仍到v1.5.3；使用官方固定pseudo-version，不虚构正式1.5.4。
- [module-feasibility.json](module-feasibility.json)、[module-diff.json](module-diff.json)：临时modfile复制当前mod/sum，只切Gorilla到v1.5.4-0.20240701034025-d67f41855da4；MVS仅此模块变化，产品保持原哈希。
- [baseline-modules.log](baseline-modules.log)、[candidate-download.log](candidate-download.log)、[candidate-modules.log](candidate-modules.log)：原始模块选择和下载校验和。
- [gorilla-module-diff.patch](gorilla-module-diff.patch)：模块实际18文件差异，包含代理/握手行为；不是只有13行随机源提交。
- [candidate-ws-controls.json](candidate-ws-controls.json) / [log](candidate-ws-controls.log)：临时模块下三包应用WS定向376顶层/700含子例通过，仍是可行性。

漏洞涉及客户端mask，静态生产import仅ops服务端，其余在测试。没有执行全仓callgraph/govulncheck或漏洞利用，不将package命中写成确认生产利用。

## 库测试失败与对照

| 检查 | 原始结果 |
|---|---|
| [baseline-gorilla-tests.json](baseline-gorilla-tests.json) / [log](baseline-gorilla-tests.log) | 旧版55顶层通过/3失败，57含子例通过；失败ProxyDial、ProxyAuthorizationDial、PreparedMessage。 |
| [candidate-gorilla-tests.json](candidate-gorilla-tests.json) / [log](candidate-gorilla-tests.log) | 候选59顶层通过/2失败，61含子例通过；同两项代理unexpected EOF，PreparedMessage通过，不能报告库全量绿灯。 |
| [baseline-real-connect.json](baseline-real-connect.json) / [log](baseline-real-connect.log) | 真正本地CONNECT/Basic-auth/单管理员subprotocol，各2顶层/4含子例通过。 |
| [candidate-real-connect.json](candidate-real-connect.json) / [log](candidate-real-connect.log) | 同夹具、同2顶层/4含子例通过；发送、回显和隧道释放均检查。 |
| [baseline-proxy-framing-control.json](baseline-proxy-framing-control.json) / [candidate](candidate-proxy-framing-control.json) | Go禁止替换GOMODCACHE，overlay未执行；是工具限制。 |
| [baseline-proxy-framing-copy.json](baseline-proxy-framing-copy.json) / [candidate](candidate-proxy-framing-copy.json) | 复制模块到/tmp，仅两处历史fake CONNECT handler增加Content-Length:0，两版本仍各2失败；framing假设被否定。 |

真正隧道文本见[gorilla-connect_test.go.txt](probes/gorilla-connect_test.go.txt)和[重放说明](connect-replay.json)。历史fixture在CONNECT handler写200后直接返回，区别于真正隧道；EOF精确根因尚未完全确定，S02实施仍需处理/解释，不能用真实对照通过抹除原失败。[复制重放信息](proxy-framing-copy-replay.json)及两份fixture-only patch保留。旧PreparedMessage依赖全局rand.Seed，来源改用局部random reader；未据此宣称生产利用。

一次复制命令仍含路径占位符，[初始结果](baseline-proxy-framing-copy-initial.json) / [log](baseline-proxy-framing-copy-initial.log)保留，随后使用已创建的实际路径；该次未执行测试或修改产品。

## 收集与范围核验

- [collection-initial-failure.json](collection-initial-failure.json)：首次最大迁移解析未处理154a，在写入前退出；改为读取数字前缀，依赖命令找不到脚本的过程也记录。
- [collect-intake.py](collect-intake.py)：只读来源/私有index、完整PR覆盖、先反后正及调用/测试扫描。
- [run-check.py](run-check.py)：保存真实命令、环境、完整输出、退出码和时间。wrapper退出0不代表测试通过，应读JSON里的exit_code及原log。
- [render-report.py](render-report.py)：显式审查决定和两档未实施OpenSpec生成器；source-baseline/source-feature-map冻结，实施不得重跑改写。
- [validate-intake.py](validate-intake.py)、[document-validation.json](document-validation.json)：覆盖/patch hash、链接、严格OpenSpec、6/11未勾任务、空槽、HEAD/index及产品/历史/初始诊断哈希核验。
- [第一档严格校验](openspec-tier1.log)、[第二档严格校验](openspec-tier2.log)。

没有新的全仓漏洞扫描、全量backend/frontend/race/embed、真实浏览器、完整1C1G或业务实例/供应商操作。当前基线、已证实缺口、来源小函数、临时模块可行性和未来验收分别表述。
