# v0.2.25 两档实施验收（2026-11-24）

用户选定的 F01 / S01 / S02 已在工作区实现。[第一档](../../../openspec/changes/port-upstream-0.2.25-tier1/verification.md)6/6、[第二档](../../../openspec/changes/port-upstream-0.2.25-tier2/verification.md)11/11。起点为 `39ca66d51d53b98f11349598bbd16cd8ecd0e2e0`、VERSION1.1.27、Go1.27.4、SQLite230/auth21；实施验收阶段没有提交、推送、发布、归档或业务实例操作。用户随后明确授权提交并推送origin/main，本次提交包含两档实现与对应证据，实际提交号和推送结果由会话回报；原先未跟踪的Fast诊断留在工作区。既有标签、VERSION和冻结验收记录保持。评估期的来源、探针和临时依赖可行性仍位于独立的 [evidence-0.2.25](../evidence-0.2.25/MANIFEST.md)，不作为本次实现验收。

## 实际改动

- F01：客户端错误文本净化现有 query 参数及账号 api_key/access_token/refresh_token，凭据按长度递减替换，避免共享前缀留下后缀；配置主机名在原始字节上大小写匹配。覆盖 Unicode、非法 UTF-8、IPv6、URL、正则字面量和无匹配；URL/模式不可解析时回退通用消息。管理员内部诊断、状态副作用和成功请求路径不改。
- S01：既有规则 runtime 返回完整匹配规则，最终透传错误按原始状态/body/账号平台执行规则。仅允许的确定性400或 passthrough_body 规则透出 type/code/param/message；标识符只接受字符串/数字，原值和脱敏后均限制256字节，拒绝对象/数组/布尔。缺失/null/空白 code 支持内嵌 JSON，显式畸形 code 不启用该回退。显式管理员文案保持原样。监控只借用原 body 匹配规则优先级，不在关闭详细日志时额外保存 body。
- compact 心跳提交后仅写一次 `response.failed`，沿用 `NextOpenAIResponsesSequence`，保留确定性 code 并将容量降载 code 转为既有可重试值；自定义2xx错误状态也不会被成功桥接。心跳前维持配置状态及 JSON 错误信封；失败不返回成功用量结果。详见 [S01记录](s01/README.md)。
- S02：只将 Gorilla 从v1.5.3定向更新至 `v1.5.4-0.20240701034025-d67f41855da4`，替换必要两条 go.sum；406模块的 MVS 只有该项不同。重新核对官方漏洞/修复/校验和及完整18文件差异。详见 [S02记录](s02/README.md)。

## 本轮实际命令与结果

所有 JSON 保存实际命令、cwd、退出码、时间和计数；相应 log 保存原始输出。顶层与含子例计数不混用，不累计相互重叠的定向检查。

| 检查 | 结果与证据 |
|---|---|
| 实施起点 | [initial.json](initial.json)记录起始HEAD、工作区及8,774个现存文件SHA256，包含初始未跟踪Fast诊断与评估文件。 |
| backend全量unit | [backend-unit.json](backend-unit.json)：`go test -json -tags=unit -p 2 -count=1 ./...`，55测试包通过，10,793顶层/22,386含子例通过，16既有跳过。完整JSON流[无损gzip](backend-unit.log.gz)，[原文/压缩哈希](backend-unit.output.json)可核验。 |
| embed build | [backend-embed.json](backend-embed.json)：`CGO_ENABLED=0 go build -tags=embed -trimpath -o /tmp/sub2api-025-server ./cmd/server`通过。 |
| lint | [backend-lint-final.json](backend-lint-final.json) / [log](backend-lint-final.log)：Go1.27.4构建的golangci-lint2.14.0，`run ./...`为0 issues；[独立工具构建](lint-tool-build.json)不修改项目依赖。 |
| 错误/SQLite相关race | [backend-error-race.json](backend-error-race.json)：service/repository/middleware三包332顶层/669含子例通过。 |
| WS相关race | [s02/websocket-race-final.json](s02/websocket-race-final.json)：admin/handler/service/openai_ws_v2/middleware五包693顶层/1,351含子例通过。另有[S01两包race](s01/race.json)392/800通过，范围重叠不相加。 |
| 最终测试文件复验 | [final-targeted.json](final-targeted.json)：两包71顶层/214含子例通过，覆盖本轮测试专用修正、主要新增测试及冷缓存Fast WS夹具；取消控制另由全量unit/S01定向覆盖。 |
| 评估探针重放 | [intake-probes-after.json](intake-probes-after.json)：同一冻结test-only overlay的9顶层/16含子例全部通过，包含五类原缺口和四类对照。 |
| 实际HTTP/compact | [f01-http-corrected.json](f01-http-corrected.json)为上下游本地TCP的脱敏/普通403控制；[S01定向](s01/selected-final.json)包含12种心跳/状态/文案组合、本地序号、单终态及取消控制。不是外部供应商或业务实例验证。 |
| SQLite/计费保护 | [sqlite-billing-controls.json](sqlite-billing-controls.json)：四包69顶层/136含子例通过，含真实SQLite用量、结算去重、固定配额窗口、独立成本、Fast和requested/effective effort。全量unit另实际执行已有miniredis原子refresh/session控制。 |
| 管理员WS/CONNECT | [s02/admin-connect-final.json](s02/admin-connect-final.json)：2顶层/14含子例通过；鉴权/停用/Origin/单subprotocol/全局及IP限额/控制帧/关闭/释放，真正CONNECT及Basic-auth回显；旧依赖同夹具也通过。 |
| 最终产品/测试哈希 | [final-backend-hashes.json](final-backend-hashes.json)记录15文件，其中8份生产源码/模块内容与组合验收开始一致。全量unit启动后仅三份新测试的lint适配及一份旧测试桩变更，最终定向、WS race与lint分别复验。 |
| OpenSpec/范围保护 | [scope.json](scope.json)核验历史、初始诊断、冻结来源、前端、VERSION/Go/schema/生成代码及旧13活动change/439完成任务；两份严格校验见[第一档](openspec-tier1.json)/[第二档](openspec-tier2.json)。 |

## 失败记录与已定因局限

[f01-before.json](f01-before.json)原样保留真实配置泄漏的预期失败及普通403通过对照；[f01-http-before.json](f01-http-before.json)保留实际HTTP泄漏。[夹具说明](f01-http-fixture-note.md)记录首次403误用context code的纠正。S01记录保留五类原失败、错误优先级的额外回归失败，以及两处漏配API Key导致未到达上游的测试夹具失败。

[backend-lint.json](backend-lint.json) / [log](backend-lint.log)保留首次11项测试代码问题；只修正Close返回值、类型断言和HTTP状态常量，未以此修改产品。首次广义WS race的两项失败分别来自实施中的优先级回归和旧Fast设置stub冷缓存panic；旧Gorilla隔离复现后，仅为该stub补齐既有values读取，最终WS race通过。

[scope-initial-failure.json](scope-initial-failure.json)记录范围脚本首次把尚未生成的自身报告链接判为缺失；只调整生成期链接检查，产品不变，最终范围核验通过。

[commit-whitespace-artifacts.json](commit-whitespace-artifacts.json)记录暂存新增冻结patch、TSV及原始日志时的空白检查提示；这些证据保留原字节，产品代码和说明文档的暂存检查另行通过。

依赖库自身原测试仍有两项历史代理EOF：[原版本产品环境](s02/baseline-library-project.json)55顶层通过/3失败，[新版本产品环境](s02/candidate-library-project.json)59通过/2失败。Go1.27.4的非Hijack CONNECT handler写200后即关闭连接，原夹具错误地期待连接继续接收GET；双版本wire trace和临时副本只改两处为真隧道后各2/2通过共同确认原因。原版PreparedMessage还依赖旧rand.Seed行为，独立兼容控制确认；修复版该项通过。没有修改module cache/依赖源码，也不宣称依赖库原样全量绿灯、执行漏洞利用或全仓漏洞扫描。

## 数据与操作边界

保持VERSION1.1.27、Go1.27.4、SQLite230、auth21、其余库/前端依赖、Ent/Wire，以及64MiB body/40MiB SSE/8MiB staging/120s idle。没有前端改动，故未新增前端门禁或浏览器验收。没有全量integration、外部供应商/生产容量验收，亦未把历史资源测量当作本批结果。头身份变量、上游名称同步、group相关按需能力、新模型/平台和支付均未引入；冻结历史保持原措辞，两个新change维持active等待另行归档。
