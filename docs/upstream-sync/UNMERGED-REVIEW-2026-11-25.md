# 历史未合内容重新评估（2026-11-25）

**建议先处理现有 Token 统计范围错误（第一档），再补 Gemini 原生 2xx 带内错误的运维归因（第二档）。** 两项均由当前代码和隔离探针重新确认，明确从旧第三档调整。其它新能力按使用收益排序：Claude Code 的原生 ToolSearch、搜索来源块、compaction 5m 策略；随后是独立最大上下文同步和监控 URL 前缀。不是重新推荐已完成的 Fast、TPS、Grok 目录诊断或历史工具桥。

本轮评估旧版本已经进入上游、尚未被本 fork 整体采用的内容；不是查询新 Release，也不实施功能。新建[第一档提案](../../openspec/changes/port-upstream-historical-backlog-tier1/README.md)5项任务与[第二档提案](../../openspec/changes/port-upstream-historical-backlog-tier2/README.md)6项任务，**全部未勾选、实施验收槽为空**。完整[134 PR / 44 类库存](./unmerged-review-2026-11-25/inventory.md)、[逐候选来源](./unmerged-review-2026-11-25/inventory.json)和[证据索引](./unmerged-review-2026-11-25/README.md)独立保存，旧评估、上一轮复核和原 Fast 诊断保持原文。

## 1. 基线、库存与完成状态

| 项目 | 本轮冻结值或结果 |
|---|---|
| fork HEAD / VERSION | `b9d9524c6cce8488edb2eaf899d771176f3b2414` / `1.1.28` |
| Go / SQLite 最大迁移 / auth snapshot | `1.27.4 / 230 / 21` |
| 来源范围 | 已评估的历史版本，终点 v0.2.25；不纳入新 tag 或未合上游分支 |
| upstream/main（来源观察） | `0f6671457132425d1b84f3d80536b53e02942424` |
| 完整整 PR / 功能分类 | 134 / 44；包含已选子集、按需余项及应关闭的旧标签 |
| 候选×文件 / A/C/X/N | 1,843 / 17 / 1,026 / 654 / 146，全部先 reverse 再 forward，仅 check |
| 推荐第一档 / 第二档 | 各1簇，5 / 6项未实施任务 |
| 其余库存 | 第三档38类/127 PR；第四档2类/3 PR；关闭旧标签2类/2 PR。支付、PG及其它缺平台类别另列§5，不计入这134 PR统计 |
| 启动既有工作区 | 2份已修改入口文档、23份未跟踪文件；含上轮复核及原Fast诊断，作为本轮输入保护 |
| 既有活动 OpenSpec | 15份、456已完成/0未完成；新建的2份提案另计，不将待归档视为待实现 |

已完成项的依据是实现提交、归档/verification和本轮相关回归：v0.2.20–25两档、选定GPT-6.2、requested/effective effort、完整Fast、窗口TPS、Grok诊断及历史四项客户端兼容尾项。旧报告中的“未实施”不会自动覆盖这些后续事实。

## 2. 方法与证据口径

1. 从各轮§5、2026-10-31旧复查和后续实施记录整理候选；初查84个PR后，补入v0.2.9全部数字编号第三档PR及遗漏的模型/区域/强制Ultra Fast子项，共134个。完整范围见[scope-expansion.json](./unmerged-review-2026-11-25/scope-expansion.json)。
2. 全部取 `merge^1..merge` 整PR净diff，不逐微提交试拼；补丁无损gzip、SHA256、逐文件reverse/forward退出码保留。只用私有index检查，从未apply到产品。
3. 收集16,111条新增调用索引和当前生产声明，结合[18类本地消费者查询](./unmerged-review-2026-11-25/base-review.json)审查。此索引含标准库/测试/同名方法，不是编译证明；整PR大小也不是剩余工作量。
4. 本轮对统计、Gemini原生响应、metadata边界、监控前缀、最大上下文、搜索来源、ToolSearch及缓存策略实际执行探针。其它旧小功能做来源/当前基座/用途复核，**没有逐PR运行全部行为测试**；表中第三档不冒充已确认缺陷。
5. 已列测试先执行 `go test -list` 再执行当前基线；集成文件首行另保存。不把postgres约束文件算作SQLite覆盖。新增测试仅通过overlay加载，产品目录没有新增测试文件。
6. 不读取业务配置推断使用频率，不更改实例。未来若选中需schema的项目，从实际最大迁移230之后分配，不复制来源239/246/248/249等编号。保持64MiB请求体预算、40MiB SSE行、8MiB暂存、120s idle与单节点SQLite/miniredis。

## 3. 第一档 F01：现有 Token 统计遗漏模型和平台（#7001）

**档位调整：旧第三档 → 第一档。** 这次不只依据“全平台统计”标题，而是用真正modernc SQLite调用当前 `GetOpenAITokenStats`：同一时间窗有OpenAI `gpt-6.1-sol`和`o3`，选择openai只返回前者；选择anthropic或grok返回0；不限定平台也只返回1个模型。现有接口、卡片和平台/group过滤已经存在，额外的 `ul.model LIKE 'gpt%'` 把查询结果再次截窄。

- **patch sites**：`backend/internal/repository/ops_repo_openai_token_stats.go`的公共WHERE；管理员handler注释、既有Ops卡片说明及中英文文案。复用 `buildUsageWhere`，不新增统计入口。
- **已测最小前置**：SQLite支持当前CTE/ROUND/CASE；4种过滤均执行到仓库结果断言，无PG语法/编译失败。来源7文件+169/-13，全部CLEAN，主要包含回归；不是迁移或新统计系统。
- **保持**：v0.2.24已完成的时长加权TPS、TopN/分页、时间半开区间、group优先/account回退平台归属，以及全窗口TPS分位数独立口径。
- **未来验收**：真实SQLite验证gpt/非gpt/Claude/Grok、平台与分组交叉、软删或无group归属、TopN/分页/总数/空窗；前端既有平台切换和分页重置。查询扩大后的大窗口成本另测，不宣称无成本。

它修复的是现有统计查询合同，不改用量账本、实际费用、定价或调用供应商；不引入支付。

## 4. 第二档 S01：Gemini原生HTTP200带内错误的Ops归因（#6929）

**档位调整：旧第三档 → 第二档。** JSON和SSE两条现有native writer中，`{"error":{"code":503,"status":"UNAVAILABLE",...}}`按HTTP200原样返回，却没有OpsStreamError或上游错误上下文。当前Ops中间件消费这些标记，无法仅凭200状态发现该错误。正常STOP/MAX_TOKENS对照通过。该结论是本地writer/上下文缺口，不声称已经连接外部供应商或完成Ops数据库落库验收。

- **patch sites**：`gemini_messages_compat_service.go`的native JSON、SSE及SSE聚合出口；新增最小信号解析；复用 `ops_upstream_context.go`与 `ops_error_logger.go`已有合同。
- **来源规模与真实工作**：9文件+1451/-34，4 CONFLICT/5 CLEAN；核心新helper420行，含错误信封、prompt block/内容过滤、空结果、数组/OAuth包装及有界SSE兜底。不能按一个标记函数估成小修。
- **后续已具备的前置**：当前已有 `MarkOpsStreamFailure`、`CountTowardsSLA`、请求级归因与 `skip_monitoring`；来源新增的能力可复用这些本地合同，**不需要照搬新的一套字段/状态系统**。
- **行为边界**：客户端原始状态、字节和已观察usage保持；不新增自动重试、账号处罚或配额策略。Google错误/真实空响应与请求级policy结果分别归因，STOP/MAX_TOKENS/普通工具停止不当上游失败；按现有优先级只登记一次。
- **未来验收**：实际本地HTTP服务、native API Key/OAuth包装、JSON/数组/SSE/聚合、首帧/尾帧/混合信号、内容过滤与空结果、取消/读写错误/大小界限；Ops实际落库、SLA/skip规则、独立成本/usage/dedup与固定配额控制。非data兜底必须有界，不替换现有有界reader或断连计量。

## 5. 仍未合内容的建议顺序与边界

### 5.1 优先选择的第三档

以下是使用收益和实施范围判断，不表示已授权实施。成本用“小/中/大”区分改动边界，不虚报完成天数。

| 顺序 | 项目 | 什么情况下值得选 | 当前增量与主要验收 |
|---|---|---|---|
| 1 | #8804 原生ToolSearch | Claude Code经Messages→Codex使用延迟工具发现，普通function桥不够 | 中：新增271行转换器及call/output/历史恢复，保持揭示前后顶层tools稳定；实际HTTP/流/非流、强制隐藏工具、混合命名与缓存前缀需整体测。现有普通ToolSearch和历史工具桥不重做。 |
| 2 | #8803 搜索来源块 | 希望Claude Code拿到宿主web_search实际检索来源 | 小到中：60行helper加include/类型/两个响应出口；当前只输出空结果块。异常URL/非字符串sources不能丢终态；不需新增Exa。 |
| 3 | #8794 compaction 5m | 已开启强制1h缓存重写，且频繁压缩后不再复用长缓存 | 小：约50行识别/重写，只有已有TTL设置分支生效；识别依赖固定Claude Code提示前缀，须测误识别/提示变化、工具循环及缓存排序。不推断官方价格或实际节省。 |
| 4 | #7057 独立最大上下文 | 自定义/第三方模型默认窗口与最大窗口不同，且需要客户端利用最大值 | 小到中：已有parser和Codex descriptor字段，通用持久化metadata却丢最大值；不需DDL/完整pinned系统。旧快照回退、混组最小值、上游显式较小窗口和可见额度要保守。 |
| 5 | #6917 监控URL前缀 | 端点确有 `/proxy`、`/anthropic`、`/v1` 前缀 | 小：5文件+111/-11；扩展原origin-only合同，同时处理完整路径段去重、编码路径、SSRF/DNS、私网和query/fragment限制。现有合法origin控制通过。 |
| 6 | #8976 / #8902 / #8367 | 常复制模型ID、手动刷新Token或长设置页操作 | 小：复制、风险提示、保存浮标分别独立；#8902仅14增/1删，#8976为46增/1删且当前四文件冲突。便利性不冒称修复。 |
| 7 | #9008 名称备注同步 / include=group | 经常维护多个第三方API Key账号，希望少手改名称 | 中：可先选只读group API/手动同步，再决定默认关闭的自动同步。现有GetByIDLite/Name/Description足够，不需整group i18n；SQLite事务/CAS/outbox、并发编辑字段归属需测。Gorilla子项已完成。 |
| 8 | #9052 请求头身份变量 | 上游确实按用户/Key/group归因或审计 | 中到大：认证scope、模板、request_id净化、后台共享模型缓存与WS身份分桶整体设计；静态头覆写仍可用，不是目前认证bug。 |
| 按权益 | #7684→#7726 Claude grant | 实际拥有重置grant且需要在网关管理 | 中：先只读查询，再选择兑换；profile权限、到期/阻塞/冷却、未知结果不可盲重试。不是普通5h/7d用量或支付功能。 |
| 按统计 | #7646 / #8900 | 经常看费用趋势或需要Claude窗口费用估算 | 小到中：metric缓存键、乱序、排名、估算/实际成本区分。窗口TPS和usage日志已完成，不能再次推荐其旧实现。 |

### 5.2 需要用途、schema或明确权衡的其余内容

| 类别 | 当前建议 |
|---|---|
| #7085 整父metadata剥离 | **保持第三档语义选择**。4种现有HTTP出口证明仅删content_item_kinds、保留keep大整数；当前回归明确要求保留其它父字段。来源29行生产增量会删整个父字段，不能直接合或宣称“已经等价”。未连接真实Codex后端证明其它字段被拒，模拟接收成功也不代表供应商接受。 |
| #8606 30m/12h、#7412独立effort倍率、动态长上下文阶梯 | 新usage/价格/schema合同；272k模型专属门控、已有Fast及requested/effective日志不是整套阶梯/倍率。需要完整用户费用/账号成本、显式0/null、HTTP/WS时刻和SQLite去重验收。 |
| #8812强制UltraFast/渠道字段 | 既有Sol倍率修复、GPT-6.2促销时钟已经完成；新强制策略与字段保持第三档，不能用“Fast已合”推断默认权限或选中策略。 |
| 新模型/区域与权益 | Fable/Mythos、Haiku、Carbon/Hydrogen/Qwen、Argon、DeepSeek视觉、Vertex us/eu等按实际供应商/套餐核价；自定义映射可能可达，不能一律判缺平台，也不凭目录推断权限。 |
| 图片扩展 | #8323统一实际图数/分组开关、#8167跨平台生图开关需schema；#8962五图、#8348 50MiB、#8462尺寸路由、#8557新模型、#6890原生Codex Images独立按需。已有图数观察/断连计费/三图错误修复保留。 |
| Gemini/Antigravity quota与配置 | #8522新quota-summary、#8568 Argon套餐、#8590强查、#8114摘要、#8166 SA范围、#8287重试配置独立。普通读数/取消/刷新保留已知值已完成，默认5次总尝试无需为修复再配开关。 |
| 权限/准入/多用户 | #8319/#8476受限admin Key、#8389 MCP、#8504邮箱黑名单、#7679白名单、#7752 Key限额和#7681在途预占均需用途；预占来源本身在simple mode关闭。混合PR不引入支付操作。 |
| 其它管理/兼容/部署 | 批量Key/用户、渠道刷新/模板/详情、绑定来源、Baseten/DeepSeek/Kimi/GLM host限定子集、自动Claude版本、日志/S3/月备份、视频/TypeSafe等逐项见134项库存。未逐条行为复现的微修不列为已确认现存缺陷，不恢复已裁剪商业界面。 |
| 第四档 | 缺原生OpenCode/Command Code/Kimi/插件/Voice/cache-CRUD消费者的专属修复、PG rollup/锁/备份SQL、多实例Redis租约、transcript阻断和上游版本/赞助商不合。通用API Key兼容、OpenCode导出不证明原生平台存在。所有支付、购买/续费、充值、退款、商业返利及专属依赖迁移排除。 |

### 5.3 性能尾项的本轮重测

不能再只写“缺基座”。本轮以当前Go1.27.4实际运行单P、三次100ms合成夹具，以下为中位数；分配不是常驻内存或RSS。

| 场景 | 当前路径 | 隔离SQLite total-only窄原型 |
|---|---|---|
| extra无填充 | 25.873µs / 1,972 B/op | 39.812µs / 1,086 B/op，更慢 |
| extra额外4KiB | 52.040µs / 30,254 B/op | 53.640µs / 1,085 B/op，耗时没有改善 |
| extra额外32KiB | 216.525µs / 239,415 B/op | 152.827µs / 1,084 B/op，约29%耗时降低 |

这不是直接移植#8269的PG JSONB。原型只适用测试中的total-only形态；三次累加JSON等价、新日限额数据库守卫拒绝通过，仍未覆盖完整fixed/rolling窗口、脏类型、取消、outbox、去重及生产收益。当前 `updateFixedQuotaResetAt` 即使没有正限额也可能推进fixed锚点，不能因为“无daily limit”就盲目省掉该语义。**#8269继续第三档，小记录不能宣称提速。** #8816账务lane/PG batch同样保持第三档，需要全事务和SQLite串行写成本依据。

当前miniredis TCP `GetSnapshot`：1/10/100/500账号约46.849µs / 118.696µs / 836.947µs / 3.955ms，每次分配6,264 / 43,369 / 416,130 / 2,062,473字节。它证明当前路径随池大小的成本，不证明#8251/#8833缓存方案收益；额外母本、COW/revision/失效、200ms last_used和20s stale grace需要负载与语义选择。#8518已选资源治理与窄优化保留，完整JSON体系不随行；#8610的1GiB最低预算不替代本地64MiB。

### 5.4 旧安全维护的当前口径

本轮重新执行生产 `pnpm audit --prod --json`：12 moderate、2 low、0 high/critical，涉及yaml/uuid/mermaid/colord/katex；退出1表示仍有advisory，不是应用测试失败。当前DOMPurify/xlsx/Go/Gorilla已选维护不重复列为未合。其它依赖需要独立最小版本和实际使用面分析，不自动强制整个旧PR依赖图，也未执行Go/开发依赖全扫或声称可利用漏洞成立。[当前原始审计与摘要](./unmerged-review-2026-11-25/production-audit-summary.json)。

## 6. 应关闭的旧标签与历史边界

- **#7129的来源合同已满足**：完整PR只涉及共享事件类型、handler失败、compact和HTTP-WS桥等生产消费者。当前定向序号/历史工具/SQLite控制通过，已经保留零值和共享单调序号；不再把它宽泛称为“Grok全流序号未合”。来源本身没有修复任意上游Grok原样流缺序号，不能凭标题追加不存在的消费者，也不能拿固定0覆盖本地更严格顺序。
- **#7261旧升级意图不再排旧版本任务**：gRPC属于未选插件前置，其余x/*以之后实际Go维护为准；安全结果按本轮范围分别说明。
- **后续完成不能丢失**：#8748/#6188、Fast、窗口TPS、Grok目录诊断、历史参数/namespace/defer/sequence尾项、SQLite后台、quota配置清理、固定窗口和去重已完成。新ToolSearch、其它native平台或更大目录不是这些完成项的隐含尾工。
- **最老价格/额度项仍独立**：动态长上下文阶梯、独立effort倍率、usage compaction字段、自动重置卡、模型读取上限、峰谷价格/重置锚点及公共分组限制未因新模型默认采用。旧仅requested effort部分已经由SQLite230完成，不能再次申请同列。
- 支付与缺架构项的原历史PR不一一重跑；禁止边界保持，既有代码不做无关删除。旧微修未被实际验证的部分在库存明确保留第三档，不伪造“所有遗留均重新验收”。

## 7. 建议实施顺序

1. F01：先复现真实SQLite结果，再修现有查询和既有卡片文案；无schema，保留TPS和筛选/分页语义。
2. S01：先做信号/正常对照，再联动native三条出口与Ops实际落库/SLA/skip，保留原始响应、观察usage和资源边界。
3. 用户选择新增功能后，按Claude Code使用优先ToolSearch→来源块→可选5m策略；再选最大上下文/监控前缀/小UI。它们彼此不应被捆绑成强制前置。
4. 性能、TTL/effort schema、权限、新平台及权益管理独立选择；本轮不执行实施、归档、提交/推送/发布或业务实例操作。

## 8. 本轮实际自测

| 检查 | 实际结果与边界 |
|---|---|
| 来源/四态 | 134整PR，1,843文件状态A17/C1026/X654/N146；gzip/哈希、原始reverse/forward错误保存，无产品apply。 |
| 当前基线 | 四包98顶层/214含子例通过，无失败/跳过；先执行并保存测试目录。含真实SQLite用量/结算、miniredis与序号控制。 |
| 当前前端 | 4文件66测试通过：现有Ops统计卡片、模型表、账号用量、上游probe API。错误/401是夹具预期日志，退出0。 |
| 隔离探针 | 6个目标顶层失败/14含子例失败，5个观察或控制顶层通过/9含子例通过。两类现有维护与四类可选能力分别判断，**不以期望失败直接判第一档**。 |
| 初次夹具失败 | 保留`.test` DNS控制误判的原始日志/源码；修为public literal IP进行纯校验，不发网络请求，复查控制通过。 |
| 性能/可行性 | 本地modernc内存库与miniredis TCP、单P三次中位数；原型不加入产品，不是完整1C1G实例/供应商/生产性能验收。 |
| 安全 | 当前生产audit原文保存；未跑Go/开发全扫，未更新依赖或延长例外。 |
| 文档与保护 | 两份完整未实施OpenSpec严格校验；[最终核验](./unmerged-review-2026-11-25/validation.json)记录链接、空验收槽、原15变更/456完成任务、HEAD/index/产品及23份初始未跟踪文件保护。 |

未执行新的全量unit/integration/race、embed构建、真实浏览器或业务上游验收。这些属于选定实施后的门禁，当前探针和合成性能结果不作为实施完成证据。

## 9. 新的通用教训

- 旧第三档可以按实证调整，但须说明是已存在的接口/writer合同缺口；预期新增行为失败本身不足以升档。
- 档位按独立行为，整PR检查保留审查边界；CLEAN不代表有需求，CONFLICT也可能来自已经完成的更严格实现。
- 字段同名仍可能是不同合同：只删内部子字段和删整个metadata父字段不能互称等价；存raw manifest和存通用metadata也不同。
- SQLite省Go解析可以显著降低分配，也可能让小记录变慢；fixed锚点是语义，不能被“无额度”优化悄悄去掉。
- 待归档、已有序号合同、已选计费与新能力必须分别追踪；每轮复查应关闭过期标签，而不是永久沿抄。
