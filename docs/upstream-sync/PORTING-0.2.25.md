# 上游 v0.2.25 移植评估（v0.2.24 → v0.2.25）

评估日期 **2026-11-24**。官方 GitHub Releases API 与 fetch 后的 tag 交叉核对，最新正式版为 [v0.2.25](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.25)，发布于 **2026-11-24 03:21:08 UTC / 11:21:08 上海时间**。原始 [latest-release.json](./evidence-0.2.25/latest-release.json)、响应头及 [获取说明](./evidence-0.2.25/release-fetch.json) 已保存；网页工具未返回正文，不据此推断没有更新。

**建议第一档 1 簇、第二档 2 簇，选取 2 个混合 PR 的独立维护行为；评估完成，提案均未实施。** 先修现有透传错误泄露配置主机名/密钥，再补客户端恢复所需的标准错误字段、管理员规则和 compact 终态；Gorilla 安全依赖单独维护。新增请求头身份变量和上游分组名称同步保留第三档。

- [第一档：F01，6 项未勾选任务](../../openspec/changes/port-upstream-0.2.25-tier1/README.md)
- [第二档：S01–S02，11 项未勾选任务](../../openspec/changes/port-upstream-0.2.25-tier2/README.md)
- [来源、测试、失败过程与范围核验](./evidence-0.2.25/MANIFEST.md)

本轮产物是评估、原始证据与未实施提案。现有未跟踪的 Codex Fast 排查记录保留；产品代码、VERSION、依赖、schema、Ent/Wire 和冻结历史保持原值。没有 commit/push/release/归档或业务实例操作。

## 1. 版本、范围与规模

| 项目 | 冻结值 |
|---|---|
| fork HEAD / tree | `39ca66d51d53b98f11349598bbd16cd8ecd0e2e0` / `cd84cf715cc17aec530c995b2e8fff8b52e46db0` |
| VERSION / Go / SQLite / auth snapshot | `1.1.27 / 1.27.4 / 230 / 21` |
| 私有 index tree | `cd84cf715cc17aec530c995b2e8fff8b52e46db0` |
| v0.2.24 commit / tag object | `5c2982a81b46e2142d0143d2f83c1ce9c4da1d24` / `7515db6dc56c2a6120dfb7abed70ef089ebc1457` |
| v0.2.25 commit / tag object | `348997267d756eab5729bb6b78a275a15d18796a` / `8fab111f92447a3edbcdf67a517e45b6a3cc4d31` |
| upstream/main（观察） | `0f6671457132425d1b84f3d80536b53e02942424` |
| 非 merge / 整体候选 | 12 / 3 整 PR + 1 VERSION 直接提交 |
| 净 diff / 候选×文件状态 | 91 文件 +3486 -319 / 98 条 |
| ALREADY / CLEAN / CONFLICT / NOBASE | 0 / 56 / 32 / 10 |
| 既有活动 OpenSpec | 13 份，439 已完成 / 0 未完成任务，等待另行归档 |

当前 v1.1.27 已包含 v0.2.24 的 25 簇/89 任务，不能继续使用上一轮实施前的 d2488b110 / VERSION1.1.26 作为当前基线。启动时只有 `docs/upstream-sync/diagnostics/2026-11-24-codex-fast-service-tier.md` 未跟踪，其哈希受保护。评估创建前的 8,653 个 tracked 文件已冻结。

### 完整候选与逐行为分档

| 候选 | 非 merge | A/C/X/N | 完整来源规模 | 判断 |
|---|---:|---|---|---|
| [328db26f4](https://github.com/Wei-Shaw/sub2api/commit/328db26f450edbdf9f8d4344a3ea903ac6396842) | 1 | 0/0/1/0 | 1 文件 +1 -1 | 第四档：上游 VERSION 同步；fork 使用 1.1.x。 |
| [#9008](https://github.com/Wei-Shaw/sub2api/pull/9008) | 5 | 0/20/15/2 | 37 文件 +1032 -192 | 第二档 S02 只选独立 Gorilla 维护；include=group、手动/自动名称备注同步及文字校验为第三档。 |
| [#9045](https://github.com/Wei-Shaw/sub2api/pull/9045) | 4 | 0/2/2/0 | 4 文件 +873 -12 | 第一档 F01 + 第二档 S01：现有透传错误处理。最终整 PR 包含 Unicode 安全实现与 compact 配套，不能取中间有缺陷版本。 |
| [#9052](https://github.com/Wei-Shaw/sub2api/pull/9052) | 2 | 0/34/14/8 | 56 文件 +1580 -114 | 第三档：身份上报变量、编辑器、错误详情和 WS 身份分桶；新增能力的审计修复随能力留待选择。 |

12 个非 merge 由以上四个候选完整且不重叠覆盖。两个分支各有同题 Gorilla 升级提交，但 first-parent 整 PR 的净依赖变更只出现在 #9008；不会重复计算或重复应用。98 是候选×文件数，不是 91 个净变化文件。

## 2. 方法、最小前置与迁移编号

1. 执行 `git fetch upstream --tags --prune`，冻结正式 Release 原文、tag object、源码 commit、main 与 fork。tag 后 main 只有 `0f6671457` 的 VERSION 同步，[观察记录](./evidence-0.2.25/post-release-commits.txt)不扩展产品来源。
2. 在同一私有 index 上，对全部 98 条文件先统一 reverse check，再统一 forward check；没有 apply。保留完整补丁/哈希、每次退出码和错误文本，见 [files.tsv](./evidence-0.2.25/files.tsv)、[candidates.tsv](./evidence-0.2.25/candidates.tsv)。
3. 对整 PR 进行 [458 条新增调用扫描](./evidence-0.2.25/added-call-inventory.tsv)、[本地声明复核](./evidence-0.2.25/local-base-symbols.txt)，并查看 24 份来源测试文件。标准库/第三方方法与本地业务声明分开；CLEAN 不证明编译，CONFLICT/NOBASE 不证明本地有同一缺陷。
4. 实际运行六包 `go test -list`、当前基线测试、test-only overlay 和前端现有测试。本地 76 份 `_integration_test.go` 中 53 份含 postgres 约束，不能作为 SQLite 覆盖证据。本轮相关真实 SQLite 测试来自当前可执行的 unit 夹具。
5. tag 范围无迁移、Ent/Wire、前端依赖变更；唯一依赖增量是 Gorilla。两档提案无需新 schema；若以后单独选择新增持久化字段，必须从当时实际最大迁移 230 之后分配，不能复用 230。
6. 保留 SQLite-only/miniredis/simple、固定配额、usage dedup、独立账号成本与用户费用、Fast/Ultra Fast、GPT-6.2 和 requested/effective effort；资源默认仍为请求体64MiB、SSE行40MiB、暂存8MiB、读写idle120s。未读取业务实例推断功能使用情况。

## 3. 第一档：已确认的窄修复

### F01 透传上下文错误的账号凭据与主机名脱敏（#9045 子集）

- **本地依据**：`handleErrorResponsePassthrough` 对上下文超限保留上游文案，`sanitizeUpstreamErrorMessage` 只替换敏感 query 参数。实际 `Forward` 隔离探针中，配置主机名和独立账号密钥出现在下游；query 密钥已遮蔽。ASCII、Ⱥ、Kelvin 前缀均复现泄露，无 panic。
- **patch sites**：`backend/internal/service/openai_gateway_passthrough.go` 的 `handleErrorResponsePassthrough`、客户端错误 writer；复用现有 `Account.GetCredential` 与 query 净化，新增账号感知的脱敏 helper。
- **最小前置**：`net/url`、`regexp` 标准库与现有信封；不需要新增平台、动态身份头或依赖。按最终 tag 采用在原文本上进行安全大小写匹配的思路。
- **行为**：已有 OpenAI 透传错误向客户端保留上游消息时，应隐去配置的凭据与主机名，保留可操作的非敏感语义。
- **验收**：实际本地 HTTP，上下文超限、配置 api_key/access_token/refresh_token、host 大小写/IPv6/缺失、Unicode/非法 UTF-8/正则元字符、无匹配、query 与普通403控制；一次写回、取消及原账号副作用。
- **适用性区别**：上游新 `replaceCaseInsensitive` 的旧实现曾因 `ToLower` 改变字节长度而 panic；该函数在本地完全不存在。**本地确认的是已有文案泄露，不是已有 Unicode panic。** 已隔离运行最终来源 helper 的 10 项含子例，不能代替生产接入验收。

## 4. 第二档：需要联动适配

### S01 透传结构化错误、规则与 compact 终态一致性（#9045）

- **本地依据**：400 的 `invalid_encrypted_content` / `context_length_exceeded` 被重建为 `upstream_error`，code/param 丢失；已有规则指定422/自定义文案未影响最终写回。compact 心跳后返回 `response.failed`，但确定性 code 变成 `upstream_error`。关闭错误详情日志时，仅在原始 code 中匹配的 `skip_monitoring` 规则也失效。
- **patch sites**：`backend/internal/service/error_passthrough_runtime.go` 的匹配结果返回；`openai_gateway_passthrough.go` 的最终错误分支、有限标准字段信封；`openai_compact_stream_bridge.go` 的 terminal writer。`ops_upstream_context.go` 只作为已有副作用控制核验。
- **最小前置**：规则服务/缓存/平台匹配、compact keepalive/bridge、容量可重试码已在。新增 `applyErrorPassthroughRuleWithRule`/Detailed 可复用旧 wrapper；保留本地 wrapper 对 Grok 自定义消息提取器的支持，不能全文件替换。
- **边界**：先完成 F01。仅保留脱敏的 type/code/param/message 标准标量，标识符限256字节，拒绝对象/数组/布尔，不放行 debug/private 字段或敏感响应头。null/空白 code 继续与既有内嵌 JSON 错误分类保持一致。管理员显式自定义文案保持其配置语义。
- **compact 适配**：心跳提交后只能写一次有序 failed 事件，保留确定性 code；新增结构化路径若暴露 `server_is_overloaded` / `slow_down`，需改为既有 `server_error`。来源 writer 的 `sequence_number=0` 与本地已完成的共享序号不一致，应保留 `NextOpenAIResponsesSequence`、停止心跳、锁/取消和响应所有权。
- **已有控制不能重开**：状态码匹配的 skip_monitoring 已由 `appendOpsUpstreamError→checkSkipMonitoringForUpstreamEvent` 生效；缺口是仅原始 body 命中的规则。当前 compact 通用信封没有保留致命容量码，原本已输出 `upstream_error`；容量改码是扩展结构化字段后的必要配套，不单独宣称当前已发生该致命码终止。
- **验收**：本地 HTTP/compact 心跳前后、标准400/非400/无规则、平台/优先级/原 body/原状态、自定义文案与状态、skip marker、标量类型/限长/嵌套 code、序号、停止/取消/单次终态、原始账号/监控分类；失败不记成功用量，SQLite 账本/独立成本/固定配额与 dedup 控制。

### S02 Gorilla WebSocket 定向安全维护（#9008 独立依赖子项）

- **官方依据**：[GO-2026-6278 原文](./evidence-0.2.25/go-vuln-6278.json)将 v1.5.3 列入弱 mask 随机源影响范围，修复点为 `v1.5.4-0.20240701034025-d67f41855da4`；[官方提交](./evidence-0.2.25/gorilla-fix.json)将 `newMaskKey` 从 math/rand 改为 crypto/rand。[模块代理元数据](./evidence-0.2.25/gorilla-pinned.info)确认对应 commit/时间。当前代理列出的正式版本仍到 v1.5.3，不能虚构正式 v1.5.4。
- **本地实际路径**：唯一生产 Gorilla import 在 `backend/internal/handler/admin/ops_ws_handler.go`，用于服务端 Upgrader 和 WS 推送；其余 import 在测试。客户端 mask 弱随机机制没有确认的生产调用。Codex 原生 WS 使用另一实现，不能把 package 命中等同于已复现的 Codex 可利用漏洞。
- **patch sites**：仅 `backend/go.mod`、`backend/go.sum`；ops WS 和既有网关消费者用于验收。与 #9008 的名称同步、group API 和前端完全拆分。
- **已测最小前置**：临时 `-modfile` 的 MVS 前后仅此一个模块变化，Go1.27.4 和当前 x/net0.62.0 足够，不需批量升级。下载校验和与来源一致。
- **为何第二档**：这不是只有13行随机源变更。实际 v1.5.3→固定版本有18文件差异，含代理重构、Upgrader 借助 ResponseController、subprotocol 优先顺序以及坏握手状态400→426。本地仅声明一个 `sub2api-admin`，已验证不会因优先顺序改变而反射 JWT subprotocol。
- **实际结果与局限**：临时候选模块下三包 WS 定向376顶层/700含子例通过；旧/候选两版本的真实 CONNECT、Basic-auth、单一管理员 subprotocol 控制各2顶层/4含子例通过。依赖自身全量测试仍有两项代理 EOF，旧版本同样失败；旧版另有 PreparedMessage 测试失败。原始失败和一次被否定的夹具 framing 假设均保留，**不能报告依赖全量通过**。这些是假模块/本地合成网络可行性，不是本仓库已经完成升级或业务实例验收。
- **后续验收**：处理历史代理夹具原因，复核精确模块图/完整 diff，实际管理员鉴权/Origin/连接限额/ping/pong/关闭和 targeted race/build。保持已有 SQLite 和资源默认，不将可达性有限误写为“漏洞已不存在”。

## 5. 按需、排除与历史遗留

### 5.1 本轮第三/第四档

| 能力 | 判断与已量最小基座 |
|---|---|
| #9008 `include=group` | 新增名称/描述 API，可独立于同步实现。来源 getter 约16行，`GetByIDLite` 已在，但 `Group.I18n` 不在本地；未来可复用现有 Name/Description，无须引入整个 group i18n。小基座仍不自动使新增能力升档。 |
| #9008 手动/自动名称备注同步 | 默认关闭的账号新能力：预览端点、65行按钮、创建/编辑/列表入口、probe flag 与账户持久化接口联动；repository 约117行变更包含 PG 锁与 jsonb。未来若选定，必须独立适配 SQLite 事务/CAS/outbox、同时编辑的字段归属和探测开关，不能复制 `FOR NO KEY UPDATE` / jsonb。无新列不等于无 SQLite 工作。 |
| #9008 远端分组文字校验 | 只服务新增 name/notes 写入；本地当前 parser 忽略 group_name/group_description，控制探针证明带 NUL 的远端组名不会阻断现有倍率快照。不能将上游 PG text 拒绝改判为本地现存倍率故障。未来能力仍要验证 UTF-8、控制字符、trim 后校验和100字节名称约束。 |
| #9052 头覆写身份变量 | 334行新变量实现、认证scope、40余生产调用点/前端编辑器及错误详情，且 WS 池需按实际上报身份分桶。它是对静态头覆写的新增语义；缺scope/类型限制/request_id净化、后台模型共享缓存与连接身份隔离必须整体考虑，不能仅取插值函数。未获得该新增能力选择。 |
| #9052 审计修复 | 固定文本密钥不进日志、request_id分隔符伪造、WS request_id不随连接复用、跨用户共享模型请求不带身份，均与新增变量消费者一起保留第三档；本地没有该动态身份基座，不制造现存漏洞工单。 |
| VERSION / 支付 | 上游 VERSION 同步第四档。支付、充值、退款、付费购买/续费及其依赖迁移按个人使用边界排除；本轮没有新增支付专属候选，不删除既有支付代码。 |

### 5.2 历史遗留的当前状态

[旧基座重新扫描](./evidence-0.2.25/older-base-review.json)与[活动任务清单](./evidence-0.2.25/older-active-changes.json)按当前 HEAD 采集。此前历史报告不回写。本轮需摘取的旧必要基座已经存在：query净化、规则服务、compact生命周期、容量码与共享序号；不引入其它旧按需系统。

| 历史范围 | 2026-11-24 复核状态与保留原因 |
|---|---|
| v0.2.24 两档 | `c2075a8e9` 完成25簇/89任务，已在 v1.1.27；`39ca66d51` 仅对齐旧集成断言。不重开重复键、OAuth保存、过滤编辑/effort等已完成任务。 |
| v0.2.24 按需尾项 | #8794 compaction5m缓存策略、#8801 Exa、#8803来源块、#8804 ToolSearch、#8900费用估算、#8902刷新提示、#8976模型复制继续按用途；Grok3→5图也未选。新错误处理不依赖这些能力。 |
| v0.2.23 两档 / GPT-6.2 | `d8d126da9` / `3652ddb6a` 的70/76任务、`ad0967903` 的12任务完成；Pro200/500放行、强制/渠道 Ultra Fast 控制仍按需。保留冻结价卡与时钟证据，不把本轮当价格更新。 |
| #8748 / #6188 | `d2488b110` 完成过滤与requested/effective日志12任务，SQLite230；本轮错误code不是新增用量列，也不重复做迁移。 |
| v0.2.22–23 新模型/策略 | Fable/Mythos完整目录价卡、Carbon/Hydrogen/Qwen、Vertex us/eu、ping=0、30m/12h独立缓存桶仍按实际用途。TTL独立桶需schema；现有Fast与5m/1h完整性已经维护。 |
| v0.2.22–23 性能 | #8816账务lane/PGbatch、#8833合并加载/last_used/20s stale grace仍需SQLite事务、dedup、内存和负载收益测量；#8693 PG rollup无本地消费者。已实施窄热点不代表采用整套架构。 |
| v0.2.21 / v0.2.20 两档 | `63c764704` / `83729b130`（39/42任务）、`29681fdf5` / `a25ce7bed`（44/38任务）均完成；SQLite229覆盖索引、auth21和资源治理不重开。 |
| Fast / TPS / Grok诊断 / Terra | `326513cdc`、`f333aa5f3`、`becd2d2a3`、`a541d0b61` 已完成；保留独立账号成本、整窗口分位数与算术均值、none/max与用户所选档位。本轮没有 service_tier 计价变更。 |
| v0.2.20 管理/注册能力 | #8389 MCP、#8319受限admin Key/#8476、#8504邮箱黑名单继续按需；支付操作排除，安全标题不是授权前置。 |
| v0.2.20 图片/导出/配额 | #8462尺寸路由、#8557新图片模型、#8543/#8521冷却策略、#8522 quota-summary、#8380/#8495导出默认、#8590查询按钮、#8367浮标按用途。当前无 QuotaSummary 消费者，不为新组名API随带。 |
| v0.2.20 性能/缺平台 | #8518完整ojson/COW/compact/orphan仍需收益与所有权测量；#8503原生Kimi、#8549 transcript阻断、Voice/cache CRUD缺相应消费者。本地明确session策略与已选资源治理保持。旧main观察#8610/#8606已在后续tag评估，64MiB预算保留。 |
| v0.2.19 主体与其它 | 两档/历史四兼容尾项已完成归档；#8323实际图数、#8348 50MiB、#8287同号重试配置、#8395更广目录仍按用途。完整性/声明目录已完成，不重复列未合。 |
| v0.2.18 主体/个人能力 | `2b79ac320` / `0e38d06d8` / `b554bea55`完成两档、日志筛选、模型冷却、搜索/min tokens、SQLite后台；v1.1.23另修日志清空顺序。当前维护不得重关SQLite后台。 |
| v0.2.18 其它 | #8251完整共享解码COW、#8269 SQLite轻量quota仍需负载收益证明；不是简单“缺基座”。Argon完整quota/思考摘要/SA quota、多平台图片开关、cache CRUD、Vertex API Key按需。 |
| v0.2.16 / v0.2.12 / v0.2.11 | 主体、依赖维护、快调/排序完成；Haiku5.5/group i18n、Claude原生重置、风控白名单/Key限制/在途预占按需。TypeSafe/SystemOne与特定 DeepSeek Anthropic-native 文件仍缺消费者；普通API Key映射不等于这些原生路径。 |
| v0.2.9 | 两档已完成归档；OpenCode/Command Code/Kimi原生、plugins、完整model_allowlist glob/pinned/live、视频路由、多实例备份未整体采用；既有模型白名单UI不等于完整共享allowlist系统。 |
| v0.2.5 quota/WS | `b349c7408`完成33任务，SQLite227；WS倍率1.0为已批准选择。#6960/#7049 exec_scope/抢占/queue-wake仍需单独并发设计，与本轮无必要依赖。 |
| v0.2.5 模型/平台/协议 | #6575/#6890/#6974/#6769、#6747/#7091/#6691/#7153及#6989/#7057/#6968/#7011/#6945/#6988按用途；#7129完整Grok序号、#7085输入剥离、#6929 Gemini带内错误、#6917监控BaseURL及大管理展示未整体采用。#6986支付排除。 |
| v0.2.4 / v0.2.0 与更早 | `a26b6b55d` / `c9d9bebe8`主体、`5b3e1a716`尾项128任务完成；DOMPurify已由后续维护接管。动态长上下文阶梯、独立effort倍率、usage compaction、峰谷/重置锚点仍按需或需schema。 |
| v0.1.180 工具/Grok/WS与更早 | PDF/tool身份/HTTP回放/孤儿工具/composite闸门和兼容尾项已完成；Grok429/Realtime/抢占仍按用途；单节点不需要跨节点Redis租约。#5925仅必要delay已随消费者完成，0.1.176/179/183/184主体及Responses Lite已验收。 |

### 5.3 依赖与资源成本

本轮没有运行新的全量 pnpm audit 或全仓 govulncheck，历史12 moderate/2 low不得复用为当前结论。仅对 GO-2026-6278 核对官方数据库、修复提交、模块代理、精确MVS与包/消费者测试。临时modfile未改变产品依赖；前端现有测试保留 pnpm11 `package.json.pnpm` 位置提示，未顺带改配置。

F01/S01 主要是错误路径；实施应复用有界错误体、限制标准字段，避免增加成功路径拷贝。未作完整1C1G资源压测或宣称收益。第三档身份scope/WS身份分桶会影响请求分配与连接池基数，若日后选定，应单独测缓存共享、池上限与资源回收；名称同步也应测SQLite事务/重试，不能由来源几行SQL推断开销。

## 6. 已具备与缺基座的精确结论

| 核验项 | 结论 |
|---|---|
| ALREADY=0 | 只表示完整候选文本没有反向整文件匹配；手工等价行为仍逐项核对。 |
| 状态码规则 skip_monitoring | 已有 upstream event 路径生效；S01 补原始 body/code 仅匹配分支，不重写监控系统。 |
| compact 容量错误 | 当前通用code为upstream_error；源码新增保留code后必须配套容量改写，避免引入回归。 |
| compact 序号 | 本地已有NextOpenAIResponsesSequence，来源固定0不能覆盖。 |
| Unicode 脱敏 panic | 来源新函数的中间问题，不是本地原有函数；本地独立确认的是上下文错误配置泄露。 |
| 分组文字写入失败 | 本地没有新名称同步，远端文字不进入现有 snapshot；不能继承PG故障结论。 |
| 头身份注入 | 本地仅静态账号覆写；新认证scope/模板/WS分桶整体按需，不能当现有认证修复。 |
| 本轮模型/档位/迁移 | 无新模型或Fast计价改动，两档不需要迁移/Ent/Wire。 |

## 7. 建议落地顺序

1. F01：先消除现有上下文错误文本泄露；保留原状态、响应头与账号副作用。
2. S01：在安全文本前置上一次性组合标准字段、规则、compact 生命周期和共享序号；同文件相关hunk不能互相覆盖。
3. S02：独立精确模块维护，显式处理依赖历史失败并通过管理员/网关WS验收。
4. 完成两档组合unit/build/lint/race、SQLite账本/固定配额/dedup及最终哈希后才回填任务。当前仅评估不执行落地。

第三档若以后被选，应独立定义能力与成本验收；不因本轮文档存在而视作已授权。支付保持排除。

## 8. 本轮实际自测

| 检查 | 实际结果与边界 |
|---|---|
| 来源/四态 | 12非merge、4完整候选覆盖；98文件先反后正，A0/C56/X32/N10，补丁和哈希保留。 |
| 当前 backend 基线 | 六包编译，主定向247顶层/495含子例通过；补充auth27顶层/53含子例通过。合计274/548通过，无既有测试失败。 |
| 当前 frontend 基线 | 4文件67测试通过：头覆写、倍率探测API、错误详情解析/组件；未新增产品测试或改依赖。 |
| 缺陷探针 | test-only overlay编译成功，5个目标顶层失败/12含子例失败，4个控制顶层通过，无panic。是期望新行为的当前缺口，不是已经修复。初组4失败/3控制与后续细化均保留。 |
| 来源脱敏 helper | 在/tmp显式文件列表执行最终源码helper及原测试：1顶层/10含子例通过；没有挂入生产路径。 |
| 临时模块MVS | 唯一选择变化为Gorilla v1.5.3→固定修复pseudo-version，未升级Go或其它模块。 |
| 临时候选应用WS | 三包376顶层/700含子例通过，含本地HTTP/WS/SQLite夹具；不当成产品依赖已更新。 |
| 实际CONNECT对照 | 旧/候选均通过真实本地隧道、Basic-auth与单subprotocol：各2顶层/4含子例。 |
| 依赖自身测试 | 旧版55顶层通过/3失败，候选59通过/2失败；两版本相同代理EOF，候选PreparedMessage通过。完整日志不隐藏。 |
| 库夹具调查 | GOMODCACHE不支持overlay，保留工具限制；隔离拷贝仅加Content-Length:0仍两项失败，否定该framing假设，不能当修复或验收。另保留一次路径占位符错误。 |
| 文档与范围 | OpenSpec严格校验、链接、补丁覆盖、6/11未勾任务、空验收槽、HEAD/index/产品/历史与初始诊断哈希见[document-validation.json](./evidence-0.2.25/document-validation.json)。 |

本轮没有新的全量backend/frontend、race、embed构建、实际浏览器、完整1C1G实例或业务上游调用。这些属于后续实施门禁；现有unit中的真实SQLite/网络夹具与外部供应商验收分别表述。

## 9. 本轮通用教训

- 同一“安全修复”标题可能修的是新增函数；先证明本地消费者与数据流，再区分已有泄露、来源新panic与未来安全前置。
- 规则系统已经存在，不代表所有响应分支使用规则结果；监控事件链和最终客户端writer也可能只完成一部分。
- 保留结构化code会扩大暴露面和客户端状态机输入，必须与字段类型/限长、脱敏、compact容量改码和本地序号组合验收。
- 混合PR里的依赖更新可以独立摘取；但单模块精确升级仍可能包含代理/握手行为差异，不能把go.mod两行当成完整改动规模。
- 漏洞package命中、相关机制生产可达、可利用证明分别记录；服务端Upgrader引用不等于客户端mask机制在生产执行。
- 旧/候选均失败的库测试和真实隧道对照应同时保留；被否定的夹具假设、工具限制不得擦成“全通过”。
- PG非法文本故障不会自动转移到SQLite，更不能在本地尚无新name/notes消费者时列为既有倍率故障。小型新增API也仍需按第三档选择。
