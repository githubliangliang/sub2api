# v0.2.23 第一档实施验收

2026-11-18 按用户选择实现 F01–F21 的既有消费者。实际起点 `7a0b78fe511f17b0f7eb9d26fcf5e4ff083312c7` / tree `3882c3ce52fb24f6f8d89d782ffc7b48bb1e65e7`；开始时只有 intake 文档和两个提案，产品代码没有未提交改动。验收时实现保存在未提交工作区；本次提交保存实现，验收起点与产品哈希记录保持原值。[起点](start.json)、[完整产品差异](product.diff)、[31 个产品源码/24 个测试文件哈希](product-hashes.json)。

VERSION 1.1.25、Go 1.27.3、SQLite 229、auth snapshot 21 保留。依赖、Ent/Wire、迁移、配置与 64 MiB 请求体预算、40 MiB SSE 行、120 秒 read/write idle、8 MiB 首输出 stage 上限不改。第二档、支付、未选模型/平台、强制/渠道 Ultra Fast、新 TTL、发布、推送、归档和业务实例操作未随行。

## 验收结果

| 检查 | 结果 | 完整输出与命令 |
|---|---|---|
| 实际基线 tier1 overlay | 17 个目标顶层断言失败，11 个保护对照通过；只使用测试 overlay | [日志](baseline-tier1-probes.log) / [命令](baseline-tier1-probes.json) |
| 实施后定向 | 六包、149 个顶层测试通过，覆盖选定行为和相邻控制 | [日志](targeted-all-tier1.log) / [命令](targeted-all-tier1.json) |
| backend 全量 unit | 54 个有测试的包通过 | [日志](backend-unit.log) / [命令](backend-unit.json) |
| backend lint | 0 issues；初次两处新增测试写法修正后，对应测试重跑通过 | [最终](backend-lint-final.log) / [初次](backend-lint.log) / [测试重查](lint-test-adjustments.log) |
| 相关 race | service、handler、repository、apicompat、antigravity、web 六包通过，含固定 quota、成功去重、窗口、WS 能力和资源保护 | [日志](backend-targeted-race.log) / [命令](backend-targeted-race.json) |
| 真实 HTTP/WS/SQLite/CSP | Sol/Astra 16 种传输/阶层组合，生产 CSP 中间件与 HTML、设置变更/回退、静态资源通过 | [日志](real-http-ws-sqlite-csp-accepted.log) / [命令](real-http-ws-sqlite-csp-accepted.json) |
| 实际 miniredis | 声明能力经 metadata snapshot 保留，full account hydrate 保留；相邻 WS sticky 能力拒绝和绑定保护通过 | [定向](targeted-all-tier1.log) / [race](backend-targeted-race.log) |
| frontend typecheck/lint | 均通过 | [typecheck](frontend-typecheck.log) / [lint](frontend-lint.log) |
| frontend 全量测试 | 2206 通过，2 个既有 skip | [日志](frontend-full-test.log) / [命令](frontend-full-test.json) |
| frontend critical | 179 通过，2 个既有 skip | [日志](frontend-critical.log) / [命令](frontend-critical.json) |
| frontend / embed build | 前端通过，随后构建含新前端的 Go embed 二进制 | [前端](frontend-build.log) / [embed](backend-embed-build.log) / [产物哈希](artifact-hashes.json) |
| post-build HTML | 完整 embed web 包通过 | [日志](frontend-html-post-build.log) / [命令](frontend-html-post-build.json) |
| 真浏览器 | Chrome 桌面 1440×1000 / 移动视口 390×844，四组操作通过，122 次合成 API 请求、0 page errors | [日志](browser-final.log) / [结果](browser/results.json) / [脚本](browser-check.mjs) |
| 资源测量 | 临时 1 CPU / 1 GiB、无网络、只读容器通过 commit/drop/上限保护和 benchmark | [日志](resource-1c1g.log) / [命令](resource-1c1g.json) |
| 历史/静态保护 | 3557 个保护文件、669 个依赖/生成代码/迁移/配置文件原哈希保留 | [保护](protected-files.json) / [静态](static-start-hashes.json) / [最终范围](scope-final.json) |
| OpenSpec / 范围 | 严格校验与 70 项任务、21 个 requirement、产品范围核对 | [严格校验](openspec-tier1.log) / [范围](scope-final.json) |

完整检查记录见 [checks-summary.json](checks-summary.json)。各 JSON/log 保存 cwd、命令、时间、退出码和 stdout/stderr，失败记录未覆盖为成功。

## 逐簇行为

| 簇 | 实际实现与保护 | 验收证据 |
|---|---|---|
| F01 | 补 accumulator 的 encrypted_content，namespace、密封参数、arguments.done 保留；terminal/item.done/delta × 普通/透传/流式比较调用 ID 和原参数 | [前](protocol-f01-f03-before.log) / [后](protocol-f01-f03-after.log) |
| F02 | 中途 system 正文转 developer；tool_result 输出/图片保持工具/user 权限，首部提示和用户顺序保留 | [前](system-authority-fixture-corrected-before.log) / [后](protocol-f01-f03-after.log) |
| F03 | 缺省 strict 写 false；显式 true/false/null、原 schema、text/json_object 不覆盖 | [前](protocol-f01-f03-before.log) / [后](protocol-f01-f03-after.log) |
| F04 | 扩展三个 agent 上下文；旧 page 严格校验、重复成员拒绝、真实 call_id/未知调用原行为和真实 HTTP 通过 | [前](protocol-f04-f07-before.log) / [后](targeted-all-tier1.log) |
| F05 | 精确识别两个 beta；API Key/OAuth、最终 header policy、空/有正文消息和顶层 output_config 保护通过 | [前](protocol-f04-f07-before.log) / [后](protocol-f04-f07-after.log) |
| F06 | count_tokens 默认和调试头用 raw helper；真实 HTTP 多值、版本/beta 覆写优先、账号鉴权唯一、不带客户端 cookie | [前](compat-policy-and-wire-compiled-before.log) / [后](protocol-f04-f07-after.log) |
| F07 | 两 reader 输出后执行账号/模型副作用；Chat 补 bare error 和停止输出后 failover；容量只用既有明确模型规则，普通错误/取消不制造账号冻结 | [前](compat-policy-and-wire-compiled-before.log) / [后](protocol-f04-f07-after.log) / [race](backend-targeted-race.log) |
| F08 | 仅既有 Sol/Astra 的 Ultra Fast 6×；别名、272k、显式 0、自定义 Fast 0、实际 HTTP/WS 结算/去重通过。Free Fast 仍只免 priority/fast，Ultra 保留模型倍率 | [前](health-f08-f13-compiled-before.log) / [后](health-f08-f13-after.log) / [结算](real-http-ws-sqlite-csp-accepted.log) |
| F09 | nil 在 Metadata 访问前返回 false；实际 SQLite 身份归属冲突和既有登录/首次绑定控制通过 | [前](health-f08-f13-compiled-before.log) / [后](health-f08-f13-after.log) |
| F10 | 5h rollover 只清旧 5h 和过期/缺 reset 周桶；秒/毫秒、仅 7d、独立周桶/sample 时间保留，不添加模型/额度接口 | [前](health-f08-f13-compiled-before.log) / [后](health-f08-f13-after.log) |
| F11 | 一次 Ent UPDATE 同步 start=end−5h，保留 outbox；实际 SQLite 229 旧窗口费用排除、新费用/状态与毫秒精度 DATETIME 读取通过 | [前](health-f08-f13-before.log) / [后](health-f08-f13-after.log) |
| F12 | CreateTemp 失败永久回退有界内存；缺目录/权限、commit/drop/取消/清理、8 MiB 上限拒绝通过 | [前](health-f08-f13-compiled-before.log) / [后](health-f08-f13-after.log) / [1C1G](resource-1c1g.log) |
| F13 | 默认测试只改为已有 4-6，显式模型和其他默认值保留，不宣称供应商退役联机验证 | [前](health-f08-f13-compiled-before.log) / [后](health-f08-f13-after.log) |
| F14 | Grok 专用 message/code 和规则提取入口；嵌套优先、数字 code、脱敏、本地内容策略、可配改写/未知对象通过；共享模型不存在/停号判断不改 | [前](remaining-fixture-corrected-before.log) / [后](targeted-all-tier1.log) / [race](backend-targeted-race.log) |
| F15 | 非 array 删除 items，array 非 schema 值回退合法 schema；三协议工具及既有 tuple/default/enum 保护通过 | [前](remaining-f14-f21-before.log) / [定向](targeted-all-tier1.log) / [race](backend-targeted-race.log) |
| F16 | 含 * 的白名单输入拒绝并留草稿，点击/Enter/修正/重复/映射冲突、保存重开通过 | [前](frontend-f16-f17-before.log) / [后](frontend-f16-f17-after.log) / [桌面](browser/whitelist-desktop.png) / [移动](browser/whitelist-mobile.png) |
| F17 | 保存 boolean、false/移除区段清旧值、其他 features 保留；创建/编辑 true/false 和重开通过 | [前](frontend-f16-f17-before.log) / [后](frontend-f16-f17-after.log) / [桌面](browser/bedrock-desktop.png) / [移动](browser/bedrock-mobile.png) |
| F18 | 只补投影的 openai_capabilities；chat/embeddings/Responses 生图意图、空/未配置、full hydrate 和 WS sticky 保护通过，不添加 Seedance | [前](remaining-fixture-corrected-before.log) / [后](targeted-all-tier1.log) / [race](backend-targeted-race.log) |
| F19 | HTML 各出口统一 no-store；nonce 不按旧 ETag 304；设置/回退、无 nonce 304 和 hash assets 保留 | [前](remaining-f14-f21-before.log) / [HTTP](real-http-ws-sqlite-csp-accepted.log) / [构建后](frontend-html-post-build.log) |
| F20 | 既有 Anthropic 平台 API Key Chat 入口补原 user_id token，只保留该 key；JSON/legacy/空值/类型/转义/OAuth 控制通过。来源 native 分支本地 NOBASE，N/A | [前](remaining-fixture-corrected-before.log) / [后](targeted-all-tier1.log) / [事实](source-adaptation-facts.json) |
| F21 | 本地原有脱离取消的 3 秒收尾 helper；仅选中粘性写回对齐 5 秒。handler 完成后取消、超时、换号不覆盖/失败不绑定通过，profit admission 保留 | [基线](remaining-f14-f21-before.log) / [目标前](remaining-fixture-corrected-before.log) / [后](targeted-all-tier1.log) |

## 来源适配、失败和测量边界

F07 的报告符号在实际基线缺失，Chat 的 bare error/输出后 failover guard 也需要最小前置；新 helper 复用账号错误处理与模型冷却，不修改 pre-output 请求级 failover 的账号状态约定。F20 的 native Anthropic 文件在冻结逐文件表中已是 NOBASE，不为 metadata 补丁引入缺失消费者。F21 实际已有取消保护，本次仅对齐选定写入的 5 秒截止。与 intake 静态描述的差异单独记录，不改冻结报告。

保留夹具修正：F02 原测试误用了 instructions/索引；来源服务测试缺 Config、F12 缺 fmt、一次调用签名和不存在的导出断言需要适配；OAuth 使用本地 event+data SSE；空能力容器保留既有不限制语义；Grok 内容控制使用本地已支持的拒绝样本，未扩大 permission-denied 判断。首次结算夹具误取 Sol cache-read 价格并对未启动 deferred 调用 Stop，按冻结价格/生命周期修正；Free Fast 只覆盖原 priority/fast，不为测试免除 Ultra。失败命令/输出完整保留。

浏览器初次需切换 Anthropic tab；同组不能绑定两个渠道，后续创建用独立临时组。保留 [首次](browser-first.log)、[tab 修正](browser-accordion-fixed.log) 和阶段结果。视觉记录等待既有 CSS 过渡结束，没有修改 toast/dialog/组冲突产品规则。

Chrome 使用本地构建与合成 API fixtures；Go 使用实际本地 HTTP/WS、modernc SQLite、Ent 和 miniredis。这不是真实供应商账号、Codex Desktop 或业务库/实例验收。1 CPU/1 GiB 单独 test binary 的 30 次 8 MiB 回退为 8.320 ms/op、16,744,496 B/op、10 allocs/op，memory.peak=70,287,360 B（约 67 MiB），不外推整个生产实例或稳定性能收益。HTML 只约定响应头和 nonce/HTML 一致性，不保证各浏览器 bfcache 策略。

第一档既有消费者无产品遗留；唯一来源 N/A 为 F20 native 分支。第二档、独立按需能力和历史待归档状态保留。实现随本次提交保存，仍为活动变更，等待归档。[OpenSpec 验收](../../../openspec/changes/port-upstream-0.2.23-tier1/verification.md)。
