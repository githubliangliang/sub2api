# 上游 v0.2.18 移植评估（v0.2.16 → v0.2.18）

> 当前实施状态（2026-10-24）：第一档6簇与第二档10簇已实现；见[第一档验收](../../openspec/changes/port-upstream-0.2.18-tier1/verification.md)、[第二档验收](../../openspec/changes/port-upstream-0.2.18-tier2/verification.md)及§10–§11的实施记录。两档实现已提交`2b79ac320`，v1.1.19发布准备见§12；以下§1–§9保留评估时快照。

评估日期：2026-10-23。**仅评估，未实施：第一档6簇、第二档10簇，共覆盖18个PR及1个直接测试提交的选定行为。** 用户本轮明确选择“现有功能修复优先，新增能力暂缓”。支付始终第四档，不作为可选功能携带。

优先处理混合分组凭据误路由、Anthropic兼容流中断结算、Codex恢复/Responses字段与Grok成本；图片下载边界和依赖安全维护同样需要尽早处理。没有修改业务源码、VERSION、迁移或lockfile，没有提交/发布/部署。

## 1. 版本、范围与规模

| 基线 | 值 |
|---|---|
| fork main | `873962ce7f5c07a336479c823419a531529e6f99`，工作树起始干净，VERSION1.1.18 |
| 上次评估tag | v0.2.16 / `01842576b3d5c1e2545b5a99ad555d36e28ed306` |
| 抓取时upstream/main | `5c185c5c5d4f5e409e2139806e0964ae944e03e8`；tag后仅VERSION同步，不扩大范围 |
| 新区间 | 56非merge / 35PR / 8直接提交；净245文件，+14512/-1369 |
| 本地边界 | SQLite-only、最大迁移226、miniredis可选、simple mode；保留128c5f7bf固定quota窗口修复 |

| Release | 解引用commit SHA | 发布时间UTC | 非merge / PR / 直接提交 |
|---|---|---|---|
| [v0.2.17](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.17) | `060d0160417782cd962a0c7f460f6cd4565d4916` | 2026-10-16T04:41:43Z | 30 / 16 / 4 |
| [v0.2.18](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.18) | `ac1b8f1271be0d674fa6f074403c1410f30a3a81` | 2026-10-23T07:44:00Z | 26 / 19 / 4 |

来源：[GitHub Releases](https://github.com/Wei-Shaw/sub2api/releases)、[release快照](./evidence-0.2.18/releases.json)、[56非merge列表](./evidence-0.2.18/commits.tsv)、[43完整候选](./evidence-0.2.18/candidates.tsv)。8个直接提交包括2次VERSION同步、4次赞助商更新、1个fixture修复、1个DOMPurify升级。

| 档位 | 簇数 / 来源数 | 决定 |
|---|---|---|
| 第一档 | 6簇 / 6PR | 本地已确认的小范围修复，生成OpenSpec，未实施 |
| 第二档 | 10簇 / 12PR + 1直接测试提交 | 手工适配、多路径/安全/依赖验收，生成OpenSpec，未实施 |
| 第三档 | 12PR（含混合子项） | 新能力及需量测的性能方案；用户暂缓 |
| 第四档 | 7PR + 7直接提交（含混合子项） | 支付、无对应消费者、PG实现、已满足/上游专用内容 |

#8093和#8269跨档，档位数量不能直接相加。[逐来源决定](./evidence-0.2.18/decisions.tsv)完整覆盖43候选。

## 2. 方法与适配边界

先执行`git fetch upstream --tags --prune`；首次SSH抓取耗时期间发起HTTPS回退，两个抓取重叠造成ref锁竞争；HTTPS已取得目标main/tags，待所有旧抓取退出后串行复核（见fetch-final.txt），冻结SHA未变，未改remote配置。按first-parent完整PR diff（`git diff <merge>^1 <merge>`）取证，候选覆盖的非merge集合与56条区间严格相等。annotated tag用`^{commit}`解引用，作者时间不作为范围条件。

独立临时`GIT_INDEX_FILE`读取冻结fork SHA，**全候选文件先反向check，再正向check**；未向工作树应用补丁。312条逐候选文件证据：**ALREADY4 / CLEAN170 / CONFLICT117 / NOBASE21**。[files.tsv](./evidence-0.2.18/files.tsv)保留双向退出码/诊断，[candidate-stats.tsv](./evidence-0.2.18/candidate-stats.tsv)汇总。CLEAN不证明可编译，ALREADY不证明整PR已合。

[新增调用清单](./evidence-0.2.18/added-call-inventory.tsv)按新增Go行检索当前func声明及同PR新增定义（包含标准库、局部closure和词法误命中，不能据同名视作已接线）；[关键符号](./evidence-0.2.18/base-symbols.txt)、[调用复核](./evidence-0.2.18/call-review.md)、[hunks](./evidence-0.2.18/hunks.tsv)记录本地替代和缺前置。未使用子代理。

新增上游SQL只有246生图开关迁移，属第三档未采用。选定16簇**不需要新增迁移**；226及以前checksum不动，旧预留227未占用。未来新增schema必须从实施时实际最大号之后分配SQLite迁移。上游#8269的PG JSONB快路径不能套到本地SQLite事务。

## 3. 第一档：现有缺陷的小范围修复

A/C/X/N表示整来源候选文件状态，包括不取的hunk与测试，重复来源不能累计为独立工作量。

| 簇 | 来源 | A/C/X/N | 本地缺口 |
|---|---|---|---|
| F01 Codex 工作区导入身份隔离 | [#8088](https://github.com/Wei-Shaw/sub2api/pull/8088) | 0/2/0/0 | buildCodexStoredIdentityKeys 在有 account_id 时仍把 user_id 作为首选键；两个工作区同一用户生成相同 user:same-user，探针已复现。 |
| F02 Codex 前导元数据不阻止重试 | [#8139](https://github.com/Wei-Shaw/sub2api/pull/8139) | 0/2/1/0 | openAIStreamEventIsMetadata 不含 response.queued/response.metadata；探针确认 queued 被当成可见模型输出。 |
| F03 Grok Fast 用量成本价卡 | [#8170](https://github.com/Wei-Shaw/sub2api/pull/8170) | 0/1/3/0 | 本地 grok-4.7-fast/build-fast 未有专属 fallback；探针确认 fast 与普通 4.7 费用参数相同。本地未知 Grok 实际兜底键为 grok-4.6，其当前价卡与普通 4.7 等价。 |
| F04 Cloudflare 1010 不处罚账号 | [#8255](https://github.com/Wei-Shaw/sub2api/pull/8255) | 0/0/2/0 | handle403 全平台入口无 1010 判断；现有 OpenAI 只豁免无结构错误的 HTML，结构化 Cloudflare JSON 仍会增加处罚。 |
| F05 Gemini 缓存命中用量非负 | [#8177](https://github.com/Wei-Shaw/sub2api/pull/8177) | 0/4/0/2 | extractGeminiUsage 和 Antigravity 两个响应转换器直接 prompt-cached；仅 cached=8 的探针返回 input_tokens=-8。 |
| F06 Gemini 工具 Schema 常量和数据边界 | [#8188](https://github.com/Wei-Shaw/sub2api/pull/8188) | 0/2/0/0 | cleanToolSchema 保留不支持的 const/examples，并对任意嵌套 map 按Schema清理。探针确认 const/examples 仍出站。 |

[第一档OpenSpec](../../openspec/changes/port-upstream-0.2.18-tier1/README.md)。

### F01 patch sites / 基座

backend/internal/handler/admin/account_codex_import.go:913，buildCodexStoredIdentityKeys/buildCodexAccountIndex；新增 codexCompositeIdentityKey 仅依赖现有 fmt。

应达到：以 workspace/account 与 user 的组合区分完整 OAuth 身份；同一工作区重复导入更新原账号，不覆盖其它工作区。

边界：保留 accessToken-only 升级和 account 兜底；长度编码防拼接碰撞，不改变凭据存储格式。

验收：两个工作区同一用户各自创建；再次导入只更新对应账号；只有 token、缺 user/account、共享 account 的成员身份分别覆盖。

### F02 patch sites / 基座

backend/internal/service/openai_gateway_passthrough.go:876，openAIStreamEventIsMetadata；消费者 openAIStreamDataStartsClientOutput 与 openai_ws_http_bridge.go 已存在。

应达到：排队和审查元数据保持首个语义输出之前的状态，后续容量错误仍可安全重试，首输出超时仍有效。

边界：真实文本、工具参数和未知事件仍按现有输出语义；保留本地 SSE 注释写出后的安全换号规则。

验收：HTTP SSE、透传和 WS-HTTP 桥在 queued/metadata 后遇到 overload 可重试；已有有效输出后不重试。

### F03 patch sites / 基座

backend/internal/service/billing_service.go 的 initFallbackPricing/getFallbackPricing，backend/resources/model-pricing/model_prices_and_context_window.json；本地已有 LongContextThresholdInclusive 与动态价卡解析。只摘 Fast 行，不带来源上下文中的 Argon。

应达到：既有 Grok 模型映射/透传请求按 Fast 独立价卡计算；短上下文 input/cache/output 每百万为 4/1/12，达到 200k 后为 6/1.5/18（来源补丁值）。

边界：不扩展模型目录、平台或购买功能；普通 Grok 与显式渠道价格覆盖不变；历史账本不回写。

验收：fast/build-fast/provider 前缀、199999/200000 边界、缓存桶、动态资源与离线 fallback 一致；普通 4.7 不受影响。

### F04 patch sites / 基座

backend/internal/service/ratelimit_service.go:944 handle403；本地 isCloudflareBotBlockResponse 缺失，补完整小 helper（旧 HTML 判断4行 + 本PR JSON递归辅助），无需旧平台消费者；保留 handleOpenAI403 的已有 HTML 防护。

应达到：识别明确的 Cloudflare 1010 或 browser_signature_banned，所有既有平台跳过账号处罚，保留请求错误和换号处理。

边界：普通供应商 403、任意数字 1010、与 Cloudflare 无关业务错误不能被豁免；不引入 OpenCode/Command Code。

验收：五个已有平台的 HTML/JSON/嵌套 Cloudflare 样本不调用 SetError/限流计数；真实403继续原有策略。

### F05 patch sites / 基座

backend/internal/service/gemini_messages_compat_service.go:2884 extractGeminiUsage；backend/internal/pkg/antigravity/{response_transformer.go,stream_transformer.go} 三处相减；max 是当前 Go 内建。

应达到：仅缓存计数或缓存大于 prompt 时非缓存 input_tokens 下限为0，同时保留缓存、输出和图片用量。

边界：正常 prompt>=cached 的计数不变；不抹掉缓存命中，不引入显式缓存CRUD或存储计费。

验收：流式 message_start/更新与非流式在 prompt缺失、0、小于/等于/大于cached 时保持非负且桶语义一致。

### F06 patch sites / 基座

backend/internal/service/gemini_messages_compat_service.go:3570 cleanToolSchema；新增 cleanToolSchemaNode 递归辅助，不新增依赖；normalizeGeminiEnum/incrementIntegralSchemaBound 已存在。

应达到：字符串const转为单值enum，与已有enum求交；移除不支持的非字符串const和注解，区分Schema节点、属性名和数据值。

边界：保留上一轮 integer exclusiveMinimum 修复；不得误删名为const/examples/$schema的合法属性，也不得改写default等数据内部值。

验收：字符串/数值const、enum交集、组合Schema、properties同名键、数据对象、exclusiveMinimum同时验收；原schema不被修改。

## 4. 第二档：值得做，但需要适配与跨路径验收

| 簇 | 来源 | A/C/X/N | 本地缺口 |
|---|---|---|---|
| S01 混合分组 OAuth 凭据协议隔离 | [#8106](https://github.com/Wei-Shaw/sub2api/pull/8106) | 0/11/2/0 | 本地有 composite 归属解析，跨平台歧义仅返回未匹配；buildUpstreamRequest 缺凭据平台守卫。探针用合成 OpenAI OAuth 账号构造出了 api.anthropic.com 请求，未向外发送。 |
| S02 Responses 时间戳与失败终态 | [#8140](https://github.com/Wei-Shaw/sub2api/pull/8140)、[`f2e5b55f3`](https://github.com/Wei-Shaw/sub2api/commit/f2e5b55f3e8069df1c02a9a27af537b5fdefa64e) | 0/4/3/0 | 本地合成 response.failed 不带 created_at，HTTP/WS共享响应也未补缺失字段；探针确认合成帧缺字段。 |
| S03 Codex 已退役订阅模型调度一致性 | [#8121](https://github.com/Wei-Shaw/sub2api/pull/8121) | 0/9/7/3 | 本地 OAuth IsModelSupported(gpt-5.5) 仍返回true，默认模型/compact fallback仍有5.5；上游按ChatGPT订阅退役规则修正。 |
| S04 Codex 容量错误恢复与出站身份 | [#8305](https://github.com/Wei-Shaw/sub2api/pull/8305) | 0/12/3/1 | 默认身份仍为codex-tui；空字符串delta/done被当成输出；请求级重试不读取上游Retry-After，且缺 SameAccountRetryDelay 字段。 |
| S05 Anthropic 兼容流中断结算与终态 | [#8210](https://github.com/Wei-Shaw/sub2api/pull/8210) | 0/4/5/4 | 两条既有Anthropic→Chat/Responses转换在写失败时提前返回，读错误或缺message_stop仍可能成功；探针复现仅message_start后EOF返回nil错误；handler错误分支丢弃部分usage。 |
| S06 既有 Antigravity 协议与别名修复 | [#8093](https://github.com/Wei-Shaw/sub2api/pull/8093) | 0/20/20/0 | tool_result媒体只变文本；原生Gemini缓冲聚合只累加图片/文字，丢functionCall和签名；composite→Antigravity的/v1beta被平台守卫拒绝；5.5档位别名会走泛家族价；已有裸名自映射阻止精确档位选择。 |
| S07 Gemini 错误策略单次裁决与 Vertex 认证 | [#8060](https://github.com/Wei-Shaw/sub2api/pull/8060)、[#8076](https://github.com/Wei-Shaw/sub2api/pull/8076) | 0/4/4/1 | 本地checkErrorPolicyInLoop与循环外分别CheckErrorPolicy，同一401可被处理两次；Vertex共享容量429进入账号冷却；SA token account-not-found只返回普通构造错误。 |
| S08 Gemini 客户端断开保留已观测用量 | [#8168](https://github.com/Wei-Shaw/sub2api/pull/8168) | 0/2/2/0 | messages/chat/native在读取消后返回nil结果，chat成功结果还硬写ClientDisconnect=false，已有token/图片用量会丢失。 |
| S09 现有图片转存下载边界 | [#8223](https://github.com/Wei-Shaw/sub2api/pull/8223) | 0/2/0/0 | 现有ImageResultUploader用普通http.Client，接受内网URL和非图片内容；仅本机httptest探针复现从loopback下载JSON成功，未访问任何真实内网服务。 |
| S10 工具链与重定向依赖维护 | [#8267](https://github.com/Wei-Shaw/sub2api/pull/8267)、[#8275](https://github.com/Wei-Shaw/sub2api/pull/8275) | 0/0/7/0 | 当前Go1.27.0；govulncheck package扫描命中标准库修复版本1.27.2/1.27.3。pnpm audit命中follow-redirects1.16.0两条high，要求>=1.16.1；DOMPurify3.4.17已在本地。 |

[第二档OpenSpec](../../openspec/changes/port-upstream-0.2.18-tier2/README.md)。

### S01 patch sites / 基座

composite_model_route.go/composite_route_resolver.go、gateway_service.go:1447 resolveCompositeModelOwnership、gateway_upstream_request.go、gateway_count_tokens.go、server/routes/gateway.go 与 handler/openai_gateway_handler.go。ListModelAvailabilityCandidates 已存在；domain.LookupPlatform/UsesOpenAIGateway/ConcretePlatformIDs 缺失，按本地五平台常量和实际路由表达，不导入平台catalog。

应达到：混合归属含只能走 OpenAI 网关的 OAuth 且模型兜底会选该族时拒绝歧义；Anthropic 请求构造在发送前阻止外平台 OAuth；count_tokens 给出可回退的404。

边界：API Key 多协议兼容保留；Antigravity/Gemini 独立转发在守卫前分流；配置归属不随短暂冷却变化；不记录token、不新增平台。

验收：HTTP/WS歧义拒绝、明确路由、纯API Key歧义、冷却账号仍参与配置归属；对OpenAI/Grok OAuth构造请求时确认未发出HTTP，合法Anthropic/Vertex及独立Gemini路径保持工作。

### S02 patch sites / 基座

来源主改 openai_gateway_response_handling.go，但本地 buildOpenAIResponseFailedSSE 已移至 openai_bare_error_stream.go:63；同时适配 openai_gateway_passthrough.go/openai_ws_http_bridge.go。新增94行 stamper文件只用time/strings/gjson/sjson；本地局部 bareError state与来源不同，禁止整文件覆盖。直接测试提交仅随本簇对齐旧fixture。

应达到：非透传 Responses JSON/SSE/WS桥缺失或null的created_at补齐；同一流共享首次确定时间；合成失败带created_at和sequence_number。

边界：已有值不覆盖；透传原上游帧保持字节语义，仅观察时间用于本地合成帧；保留上游usage、工具恢复与终态约束。

验收：缺失/null/已有时间、created与completed/failed一致、裸error替换序号、JSON和SSE转JSON、Grok桥/WS、透传未修改事件；严格JSON字段断言。

### S03 patch sites / 基座

account.go:IsModelSupported、openai_account_scheduler.go、openai_gateway_scheduling.go、openai_compact_fallback.go、openai_ws_forwarder_support.go、gateway/openai_gateway_model_availability.go、handler/gateway_handler.go/no_account_error.go 与现有 openai_codex_models_service.go。来源 openai_models_list.go 本地整文件缺失，但 modelCatalogEntries 只有17行标准库JSON解析；只随消费者补等价解析，remote/file目录按本地既有别名投影链过滤后重算ETag。

应达到：在实际映射目标上判定订阅账号退役模型；普通/粘性/compact调度和客户端目录一致过滤；不支持时返回model_not_found及迁移提示。

边界：API Key上游仍可用；显式映射到可用替代模型允许，映射到退役模型拒绝；空目录有权威性；不全局替换用户模型、不删除历史价格、不引入pinned目录。

验收：OAuth/SetupToken/shadow/API Key混合、显式别名双向映射、compact专用映射、冷却与previous_response_id、remote/file与空列表；用本地桩验证，不声称验证真实账号模型权限。

### S04 patch sites / 基座

pkg/openai/request.go、service/openai_codex_identity.go、openai_gateway_passthrough.go/openai_gateway_upstream_errors.go 与现有设置文案。#5925完整PR（4033387fd，内部953028718）仅需一个time.Duration字段和handler/failover_loop.go约6行读取逻辑，不取SameAccountRetryDeadline或Grok其余改动；来源调用点与本地都存在。

应达到：TUI出站UA与originator成对归一为CLI并保留生效版本/系统指纹；空语义片段不阻止安全换号；容量错误尊重有效上游重试提示且最多8秒。

边界：已有真实输出不能重试；非容量错误不改变；合法非TUI身份和管理员版本策略保留；重试可取消，不引入Grok新配额/Realtime/截止时间机制。

验收：UA/originator/version配套，HTTP/WS与探测头一致；空/非空delta及done、秒/ms/HTTP-date提示、非法/负数/NaN/超大数防溢出、8秒上限和取消；保留已实施指数退避。

### S05 patch sites / 基座

gateway_forward_as_chat_completions.go:357、gateway_forward_as_responses.go:509、gateway_forward.go 和两个handler。新增 gateway_forward_compat_stream.go 291行；本地没有来源 native line pump，需以v0.2.18终态补162行纯流读取helper（包含config、定时器/取消），不引入两条缺失CN native消费者。opsUpstreamProxyID/Name、RequestedReasoningEffortFromContext等来源辅助缺失，使用本地既有Proxy归属和usage字段，不为复制测试拉大功能。

应达到：客户端断开后停止下游写入、有界读完上游并收集usage；上游异常终止返回对应协议错误而非成功；有已观测用量的失败也记录一次，输出前可重试错误不重复结算。

边界：保留本地128c5f7bf固定窗口、删除Key结算、dedup与simple mode；ping/注释写出和真实语义输出分开；不能因排水无限占用连接或goroutine。

验收：写失败后最终usage、客户端等待期间断开、缺stop/读错/超时/带内error、错误前后可见输出、流终态唯一、连接关闭与无goroutine泄漏；真实SQLite验证失败部分结算/删除Key/dedup。

### S06 patch sites / 基座

pkg/antigravity/request_transformer.go:buildParts；antigravity_gateway_gemini.go 的 ensureGeminiFunctionCallThoughtSignatures 已存在可复用；antigravity_gateway_streaming.go:mergeCollectedPartsToResponse；handler/gemini_v1beta_handler.go；pricing_service.go/billing_service.go；antigravity_gemini_thinking_variant.go；account_handler.go:antigravityAccountTestModels、antigravity_gateway_service.go:TestConnection；pkg/antigravity/client.go:TierIDToPlanType 和 gateway_retry.go:accountHasAntigravityPaidTier。测试请求只修按mappedModel判定，保留旧TestConnection接口；不复制40文件整PR。

应达到：仅修已有模型和路径：保留工具base64图片/文档、Gemini函数调用及签名；放行已解析到Antigravity的Gemini路由；已有Claude5.5档位别名用基准价格；明确的不同目标映射优先，自映射可选精确已有档位；账号测试按映射目标选协议并剔除通配符候选；兼容已有上游套餐细分ID。

边界：不增加Argon/Claude新档位目录，不引入元数据目录扩展、UA hub升级或生图测试新展示；工具媒体不远程下载；套餐ID只用于既有上游端点选择，不是支付功能。

验收：工具媒体顺序、文本兼容、functionCall/thoughtSignature回放、composite入口非Gemini仍拒绝、5.5别名动态/fallback价格、已有Gemini自映射/不同目标/通配符、映射测试协议、Pro/Ultra细分ID与未知档位。

### S07 patch sites / 基座

gemini_messages_compat_service.go 的 Forward/ForwardNative/checkErrorPolicyInLoop、gemini_chat_completions_compat_service.go、vertex_service_account.go。新增约24行429判定和typed token error，使用本地logger/Proxy归属；来源opsUpstreamProxyAttribution缺失，不复制隐含ops重构。

应达到：同一上游响应只裁决一次错误策略；Vertex无quota/error明细的共享容量429只做请求内重试；确切SA不存在时标记认证失效并换号。

边界：真实项目配额/RetryInfo/QuotaFailure保留冷却；非account-not-found token错误不永久停号；只使用现有Vertex service_account，不引入Vertex API Key或缓存绑定。

验收：三条既有Gemini入口各自验证首次/再次401；响应重试之间策略不串用；共享容量与QuotaFailure/RetryInfo/项目额度反例；SA不存在、瞬态token错误、取消与换号耗尽。

### S08 patch sites / 基座

gemini_messages_compat_service.go 的 handleStreamingResponse/handleNativeStreamingResponse，gemini_chat_completions_compat_service.go；新helper gemini_stream_client_abort.go 25行，hasObservedTokens/observedGeminiImageOutputs本地已有。与本地图片计数修复合并，不覆盖整文件。

应达到：确认由客户端断开且已观测到可计费用量时返回部分结果并结算；记录ClientDisconnect与是否收到finishReason；图片只按已观测结果处理。

边界：不估算未收到的token，不把真实上游错误或无用量中断伪造成功；残行读错误不算完整帧；不引入#8114新思考输出。

验收：三协议的客户端取消/写失败、已完成后取消、无usage、图片已输出/未输出、截断残行和真正上游错误；usage队列在取消context下仍提交且SQLite去重。

### S09 patch sites / 基座

backend/internal/service/image_storage.go:defaultImageDownloadHTTPClient/download；urlvalidator.IsBlockedHost、isBlockedHostname/isPrivateIP/isPrivateOrLoopbackHost与detectedImageContentType均存在。新增transport共约115行，安全边界与redirect需专测，所以即使2文件CLEAN也列第二档。

应达到：默认下载客户端逐跳拒绝非公网目的地；直连在socket层校验实际IP；非图片响应失败，不转存为图片。

边界：仅加固现有转存，不开启批量生图；正常公网图/b64/大小限制保留。代理模式只能验证本地DNS，无法证明远端代理解析结果，与来源限制一致。

验收：直接/重定向/IPv4/IPv6/zone/private DNS/代理分支；JSON与HTML拒绝、合法图片成功、响应体关闭和大小限制；测试使用注入transport，不依赖外网。

### S10 patch sites / 基座

backend/go.mod、.github/workflows/{backend-ci,release,security-scan}.yml、Dockerfile；frontend/{package.json,pnpm-workspace.yaml,pnpm-lock.yaml}。本地基线Go1.27.0与上游补丁1.27.2上下文不同，必须成组更新；pnpm11从workspace读取overrides。

应达到：Go升级目标1.27.3并同步go.mod/CI/Docker版本断言；只升级follow-redirects到修复版本，双份pnpm overrides保持一致并提交锁文件。

边界：不复制上游PG Docker阶段、支付依赖或完整lockfile；不把audit变绿作为无条件目标，不延长旧例外；其它现存advisory单列风险。

验收：核验所有Go版本字串、后端build/unit/SQLite，前端typecheck/lint/build/关键vitest；重新跑audit及govulncheck，目标advisory消失且无无关依赖漂移。

## 5. 暂缓、不做与历史遗留

### 5.1 本轮第三/四档

第三档均经用户确认暂缓，不生成实施任务。

- [#8078](https://github.com/Wei-Shaw/sub2api/pull/8078)：Gemini4 Argon价卡/缓存存储价格新模型能力；用户暂缓。
- [#8093](https://github.com/Wei-Shaw/sub2api/pull/8093)：Argon/新增档位目录、Antigravity模型能力目录扩展、hub UA和图片测试新展示暂缓；现有协议修复独立为S06。
- [#8094](https://github.com/Wei-Shaw/sub2api/pull/8094)：新增min_output_tokens配置与跨协议下限控制；本地已有固定floor，不构成必须新增设置的现存缺陷。
- [#8114](https://github.com/Wei-Shaw/sub2api/pull/8114)：Gemini兼容入口思考档位/流式摘要新能力；不为S08引入该前置。
- [#8080](https://github.com/Wei-Shaw/sub2api/pull/8080)：OpenAI API Key的Web Search模拟新增路径；用户暂缓。
- [#8142](https://github.com/Wei-Shaw/sub2api/pull/8142)：平台/状态筛选持久化属于新增便利交互；用户修复优先。
- [#8166](https://github.com/Wei-Shaw/sub2api/pull/8166)：Vertex service_account配额上限扩展；保留已有metadata，但不扩大quota类型范围。
- [#8167](https://github.com/Wei-Shaw/sub2api/pull/8167)：扩大生图开关到Gemini/Antigravity/composite，改变旧分组行为并带迁移246；用户明确暂缓。
- [#8249](https://github.com/Wei-Shaw/sub2api/pull/8249)：延迟单账号探测为性能优化，新增请求内sync.Once探测；有基座但未量测个人规模瓶颈，暂缓。
- [#8250](https://github.com/Wei-Shaw/sub2api/pull/8250)：modelMappingSignature零分配为性能优化；现有计算正确，暂缓，可单独benchmark。
- [#8251](https://github.com/Wei-Shaw/sub2api/pull/8251)：新增进程内解码快照缓存及元数据revision；额外常驻内存/失效协议待量测。COW是该共享契约的必要前置，本地当前每次JSON解码未证实同类跨请求竞态。
- [#8269](https://github.com/Wei-Shaw/sub2api/pull/8269)：轻量quota SQL思路可将来单独量测SQLite；不得照抄PG实现或带入Vertex配额扩展。

第四档不合；已满足项另在§6解释。

- [#8087](https://github.com/Wei-Shaw/sub2api/pull/8087)：智谱CN Coding Plan消费者文件缺失，个人版没有对应窗口解析缺陷。
- [#8079](https://github.com/Wei-Shaw/sub2api/pull/8079)：Adaptive多协议GetAnthropicProtocolBaseURL基座缺失；本地已有账号只走GetBaseURL。10行wrapper不能凭空证明本地存在协议URL配置丢失，勿新增CN平台作前置。
- [#8120](https://github.com/Wei-Shaw/sub2api/pull/8120)：本地从未引入#7853折叠筛选，本来就是平铺；4文件ALREADY，另1文件CONFLICT不代表待修。
- [#8122](https://github.com/Wei-Shaw/sub2api/pull/8122)：订单类型/退款hold/支付宝退款/Stripe client_secret均为payment-only，个人范围排除。
- [#8133](https://github.com/Wei-Shaw/sub2api/pull/8133)：EasyPay查单/回调入账payment-only，个人范围排除。
- [#8224](https://github.com/Wei-Shaw/sub2api/pull/8224)：本地无DeepSeek图片双字段alias/normalize链及测试文件，不引入缺失平台以删除其额外url字段。
- [#8269](https://github.com/Wei-Shaw/sub2api/pull/8269)：上游具体实现使用jsonb_typeof/::numeric/JSONB拼接；本地SQLite既有实现不同，PG补丁不合；可选SQLite性能探索另列第三档。
- [`4d9ac3e48`](https://github.com/Wei-Shaw/sub2api/commit/4d9ac3e482106d2da3a39fe3e546d8d6801f1cc1)：上游VERSION同步，不改fork1.1.18。
- [`b0a02a330`](https://github.com/Wei-Shaw/sub2api/commit/b0a02a3304a0a78cdd003ddda1dd23ba02c03084)：上游VERSION同步，不改fork1.1.18。
- [`9c2cca9e6`](https://github.com/Wei-Shaw/sub2api/commit/9c2cca9e627291bc719bada3d57de23b625a4180)：上游赞助商资源/宣传，不复制。
- [`92c9433d4`](https://github.com/Wei-Shaw/sub2api/commit/92c9433d40db69b00f8e158fc573e2b796c4a678)：上游赞助商资源/宣传，不复制。
- [`7da09cab4`](https://github.com/Wei-Shaw/sub2api/commit/7da09cab4e875463baa79fc150adde700c5438df)：上游赞助商资源/宣传，不复制。
- [`ac1b8f127`](https://github.com/Wei-Shaw/sub2api/commit/ac1b8f1271be0d674fa6f074403c1410f30a3a81)：上游赞助商资源/宣传，不复制。
- [`828626804`](https://github.com/Wei-Shaw/sub2api/commit/8286268047cbbca4e7f5fbb63ff0890c1ea37cf5)：DOMPurify3.4.17语义已合；本地package/workspace/lock均已固定，不因原始文件CONFLICT重复维护。

### 5.2 旧轮次遗留（当前状态，不改写冻结报告）

| 轮次 | 尚未整体完成/按需/排除边界 |
|---|---|
| 0.2.16 | 第一档14簇已在7c544c03f完成；第二档16簇已在847db09e1/5ff6e5bf9完成。本轮不重复立项删除Key结算、DOMPurify/axios/Vue/source-map-js、旧协议/前端修复。第三档Haiku5.5、模型scope冷却、分组i18n、TPS、监控详情列、Gemini显式缓存CRUD/Vertex API Key仍暂缓；支付/缺平台继续排除。旧紧凑筛选上游又撤回，本地从未采用。 |
| 0.2.12 | 第一/二档主体已在405246789完成。TypeSafe无平台，支付推广/查单排除；旧axios维护已在0.2.16完成。本轮不把follow-redirects等同axios旧任务未做。 |
| 0.2.11 | 第一/二档主体已在96f32d018完成；Claude原生额度查询/重置、风控白名单、趋势切换、Key创建限制、余额在途预占、Astra Ultrafast按用途暂缓。上游账号订阅metadata不等于支付。 |
| 0.2.9 第一档F01–F05 | RPM配置投影、OAuth重授权合并、账号UA/错误状态等仍未整批处理；F04 thinking-disabled已于0.2.11完成；F05的#7638结构输出beta已随0.2.16 S03完成，#7617其它beta子项不算完成。本轮S04只处理当前身份/空输出/Retry-After，不宣称旧簇全合。 |
| 0.2.9 第一档F06–F08 | Responses消息type/终态正文、Alpha search成功判定、未来reset前保持暂停仍需逐项移植；本轮created_at与这些不同。当前已有输出累积器不等于#7569所有终态回填边界已验证。 |
| 0.2.9 第一档F09–F15 | TOTP/公共表单/代理/配置输入/异步弹窗/价格显示/请求ID与Grok配额诊断仍未整批完成；共享dialog ID/空Tab、以及0.2.16的Esc/disabled/IME/部分异步状态已完成，不将旧整簇勾选。 |
| 0.2.9 第二档 | 48PR/16簇在159d04222完成；当前目录/工具Schema/PDF/流取消/图片统计/Codex配置作为基座保留。已有Gemini thinking-variant helper已存在，本轮修自映射分支，不能沿用历史“缺整个文件”的判断。 |
| 0.2.9 按需/排除 | OpenCode/Command Code/Kimi原生、plugins、model_allowlist glob/pinned/live、Free Fast/effort倍率/自动用卡、视频图像新路由、TypeSafe/月备份/部署运营增强未合。支付/返利始终排除。remote/file配置不是完整pinned目录。 |
| 0.2.5 主体/调度 | 主体已在bfbcd79完成；WS系数#7043仍1.0，#6954/S12b平台额度清理及旧预留227未实施；#6960/#7049完整exec_scope/抢占/queue-wake未整体采用。按需或需schema设计，非本轮前置。 |
| 0.2.5 平台/模型 | #6575/#6890/#6974/#6769新Gemini、原生Codex Images、DeepSeek峰谷价、Ollama异步重置；#6747/#7091/#6691/#7153平台/订阅/Key批量能力；#6989/#7057/#6968/#7011/#6945/#6988目录/上下文/视觉/compact/pinned按用途暂缓。收费开关#6986排除。 |
| 0.2.5 其它 | #7129 Grok序号、#7085 input stripper、#6929 Gemini带内错误、#6917监控BaseURL，及#6847/#7001/#6845/#6913/#6915管理展示未整体采用。S02只补本地合成失败序号，不等于完成Grok原生全流序号方案。#7126/#7110/#7148的已有namespace语义/撤回/PG不重做。 |
| 0.2.4/0.2.0 | 主体已落a26b6b55d/c9d9bebe8。旧DOMPurify维护现在已由0.2.16完成，历史未勾checkbox不改写。新advisory以本次audit为准。 |
| 0.1.180/0.1.184价格 | 动态长上下文阶梯、effort、Fast/Free Fast/Ultrafast完整机制未合；F03是已有Grok透传模型成本修复，不推广旧新计费机制。 |
| 0.1.180工具桥/Grok/WS | 非法参数、namespace别名、deferred尾项未整体合；Grok冷却/Realtime/抢占、跨节点Redis租约按用途或单节点架构暂缓/排除。#5925原整体暂缓，本轮重新量测只需delay字段+读取分支给S04，不带deadline或Grok其它功能。 |
| 0.1.184其它 | usage compaction/requested effort迁移、公共分组限制、订阅重置锚点、峰谷价/图像冷却、自动重置卡/模型读取上限仍独立待定。固定账号quota窗口推进是128c5f7bf已完成修复，不能被新PG quota优化覆盖。 |
| 更早主体 | 0.1.176/179/183与0.1.184已验收主体、API Key instructions/routed catalog/Responses Lite保留；695ebede7的14行通用usage字段已于0.2.11摘入，不为旧需求重新引入支付或大schema。 |
| SQLite后台 | UsageCleanupService、AccountExpiryService、ScheduledTestRunnerService仍被skipSQLiteBackgroundJobs跳过；需逐SQL核查与单独验收后才启用，不算本轮流式修复的隐含任务。 |

本轮对直接相关历史基座做了符号/消费者重测，其余依据各轮落地commit及冻结报告复核，**没有逐条重跑所有历史PR**。完整旧第一档保留为独立待办，本轮不因它们时间久而默认为已合。

### 5.3 重测缺前置与依赖遗留

- S04只补#5925一个time.Duration字段和约6行handler读取分支，纳入第二档的理由是当前Codex消费者需要；旧Grok整簇档位不变。输入必须先校验有限正数再限幅，避免直接time.Duration转换溢出。
- S03缺的modelCatalogEntries约17行、仅encoding/json/bytes/fmt，可随退役过滤补等价解析；不导入缺失的完整标准模型/pinned投影系统。
- S05缺的native line pump在v0.2.18终态162行；连同291行compat状态机需专门生命周期验收，所以不是低风险小补丁。旧CN平台仍不引入。
- #8079 wrapper只有10行，但真正的GetAnthropicProtocolBaseURL及多协议配置不存在；小wrapper不使上游缺陷在本地成立，维持第四档。
- #8251的COW需以共享所有权消费者为依据。本地GetSnapshot逐次decodeCachedAccount，尚未证明该并发崩溃在现架构可达；共享缓存/COW一并第三档暂缓，不能凭两段相同代码夸大成本或漏洞。
- 本次前端audit：全量78记录（3 critical/36 high/35 moderate/4 low），生产19记录（0 critical/4 high/13 moderate/2 low）。生产high为xlsx两条及follow-redirects两条；S10只处理本轮明确目标。开发工具链critical与xlsx专项仍待维护，旧例外不续期。
- **旧轮次工具链按需项中的Go安全更新，经本次标准库finding证据明确调整为S10第二档**；其它工具链项不随之自动升级。
- Go扫描使用package级别，命中不等于符号可达利用证明。标准库更新纳入S10；x/net、x/crypto、x/image、x/text、x/mod、compress、otel等其它记录单列[依赖复核](./evidence-0.2.18/dependency-review.md)，不顺带扩大上游依赖集。

## 6. ALREADY、N/A与本地替代实现

- #8120四个文件反向可应用，本地原本平铺，无需先引入折叠再撤回。第五个CONFLICT为fork的不同UI实现。
- 828626804 DOMPurify升级整文件CONFLICT，但package/workspace/lock已是3.4.17且本轮audit无DOMPurify记录，语义已满足。
- #8087智谱、#8079自适应协议URL、#8224 DeepSeek双图片字段链都缺生产消费者。缺失新测试文件不能当作缺功能的唯一证据。
- #8140来源合成错误helper位于response_handling，本地在openai_bare_error_stream；必须沿当前调用链适配，不能被文件态误导。
- #8269是PG性能变体，不能替换本地SQLite事务实现；删除Key正常结算和固定窗口推进均是已合基座。

## 7. 建议落地顺序

1. S01凭据隔离、S09图片下载边界；F04 Cloudflare处罚修复和S10安全维护可独立批次评审。
2. F01/F02/F03，再S02/S03/S04，覆盖Codex导入、发现/调度、Responses字段、容量恢复以及Grok成本。先补最小前置再接消费者。
3. S05 Anthropic兼容流生命周期与用量结算。保留已完成的Claude拟态/工具/缓存修复。
4. F05/F06、S06/S07/S08，覆盖既有Gemini/Antigravity媒体、Schema、认证、断流计数。共享文件串行适配并验证最终组合。
5. 完成下节实施门槛后再提交实现；用户另行决定发布/部署。第三/四档不夹带。

## 8. 基线自测与实施门槛

[详细基线记录](./evidence-0.2.18/baseline-tests.md)。已完成go test -list、协议包单测、定向service/handler/admin/repository回归、相邻Gemini/Antigravity/Codex测试、前端11测试及最小overlay探针。探针断言的是旧缺陷存在，PASS不是修复已完成。第一版探针对const同名属性的假设被实测否定，修正为实际的const/examples未清理；初次失败记录保留。

[来源测试目录](./evidence-0.2.18/source-test-inventory.tsv)区分当前存在/待移植；[integration首行](./evidence-0.2.18/integration-build-tags.txt)保留build tag。usage_billing_repo_integration_test.go等为integration && postgres，不作为本fork可执行验收。真实SQLite验收使用当前UsageBillingDeletedKeySQLite、UsageBillingRepositoryApplySQLiteQuotaWindows等；新增断流结算测试需适配真实SQLite。

本轮未修改产品，因此没有伪报整批实现unit/build/race完成。实施必须：新增失败回归→修复→定向通过→后端unit/build、流/共享状态race、SQLite结算/dedup/固定窗口；依赖变更另做前端typecheck/lint/build/关键vitest和audit/Go扫描。若真实上游模型/风控行为需账号验证，应单独记录环境与证据，不能以本地桩代替。

## 9. 新增通用教训

- 新缓存附带的COW修复先追所有权；旧代码写map不自动证明跨请求共享。性能优化的发布说明不能代替本地可达性分析。
- 混合PR按消费者拆分：Argon新能力可暂缓，工具媒体丢失与既有5.5别名价格仍可独立修。
- 相同功能的helper可能已被fork移文件或换状态机。按函数及调用链适配，不按补丁文件名覆盖。
- 历史“缺基座”重新量大小：17行目录解析、一个delay字段可以随当前消费者取；10行wrapper若消费者不存在，仍无采用理由。
- 流生命周期必须同时验证终态、用量和资源释放；返回部分result不代表handler会记录，缺stop不能补成功终态。
- 审计数字是当前快照，已修DOMPurify不能抵消后来出现的follow-redirects/工具链记录；package命中也不能写成已证实可利用漏洞。
- 探针假设被否定应记录并收窄结论；不能修改生产代码来让“基线验证”通过。

## 10. 第一档实施补记（2026-10-23）

以上§1–§9为冻结的评估记录。用户随后授权实施第一档；从`498432759`干净工作树手工移植F01–F06，实际状态及门禁见[第一档验收](../../openspec/changes/port-upstream-0.2.18-tier1/verification.md)和[实施证据](./implementation-0.2.18-tier1/)。第二档未实施；新增能力、支付和其它第三/四档不扩围。VERSION 1.1.18、迁移226、依赖保持不变。

实施量测补充的patch sites：

- F03：billing_service.go的HasIdentifiedTokenPricing和动态ModelPricing转换需要同时识别Fast/build-fast及provider前缀，否则离线别名可能失去确定性价格身份；pricing_service.go的buildModelLookupCandidates保留显式条目优先，再回退共用Fast卡。LiteLLMRawEntry/parsePricingData需把来源input/output_cost_per_token_above_200k_tokens换算为本地threshold/multiplier，**仅适配Fast**并保留显式本地字段。来源资源不添加本地扩展字段，避免用改写fixture掩盖来源格式不兼容；普通Grok策略不变。
- F04：ratelimit_service.go补完整JSON递归helper并在handle403接线；HTML/文本同时要求Cloudflare标记和完整1010代码，JSON要求明确cloudflare_error与同节点code/error_code或browser_signature_banned。比来源的纯文本子串判断更严格，实测业务1010和10101仍处罚；保留本地OpenAI通用HTML豁免。
- F02：openAIStreamEventIsMetadata是三条流路径的共享分类点。首输出超时验收针对本地已存在守卫的native SSE路径；透传/WS-HTTP桥验证queued/metadata后的容量重试和真实输出后不重试，不引入此前不存在的计时守卫。

回归先失败再修复，未把评估probe或初始测试脚手架编译错误当作修复证据。没有改source-baseline、评估证据、SQL/schema或支付代码；未提交、发布、部署。

## 11. 第二档实施补记（2026-10-24）

用户随后授权第二档实施，并指定三个agent协作；S01–S10已实现并完成定向验收，最终组合结果以[第二档验收](../../openspec/changes/port-upstream-0.2.18-tier2/verification.md)和[实施证据](./implementation-0.2.18-tier2/)为准。实施起点HEAD仍为`498432759`，叠加本会话第一档未提交实现；[start.json](./implementation-0.2.18-tier2/start.json)冻结其52个既有文件。没有使用git merge上游，没有扩大第三/四档，历史§1–§9及两份source-baseline保持冻结。

本地量测补充的patch sites（来源站点仍见§4）：

- S01：composite_route_resolver/gateway routes和OpenAI WS入口读取配置候选归属，OAuth歧义只在兜底实际落到OpenAI/Grok族时拒绝。gateway_upstream_request的Anthropic构造器只接本地五平台守卫；gateway_count_tokens在取外平台OAuth token前返回404，独立Gemini/Antigravity提前分流保留。
- S02：openai_responses_created_at共用stamper接原生JSON/SSE/WS与SSE转JSON；透传仅Observe。本地合成失败helper仍在openai_bare_error_stream，不整文件覆盖来源。apicompat/types原缺CreatedAt，anthropic_to_responses_response与chatcompletions_responses_bridge必须同时补字段及共享时间；真实Chat创建时间优先，失败帧也使用该状态时间。
- S03：openai_model_retirement接account.IsModelSupported、三种实际调度/previous_response_id和目录过滤。本地缺来源全局compact fallback，只用compact_model_mapping与最新DB复检。目录解析最小17行基座随真实消费者摘取；remote/file合并后的空目录保留权威性并最终重算ETag。配置候选决定永久退休结论，不能因全OAuth冷却而重新退回默认退休目录；冷却API Key仍影响永久可用性结论。
- S04：pkg/openai与openai_codex_identity共用TUI身份归一，设置文案同步。openai_first_output metadata分类覆盖空text/refusal/done；UpstreamFailoverError只增SameAccountRetryDelay，不引入新deadline。openai_gateway_upstream_errors解析秒/ms/date，验证有限值并在Duration转换前封顶；failover_loop保留无提示指数退避及可取消等待。
- S05：openai_gateway_anthropic_native_pump和gateway_forward_compat_stream接本地Chat/Responses转换器；沿用既有bounded first-output spool保留签名及语义前重试，排水同时限制idle与total，停止关闭body。native Messages旧403规则未改。Chat handler补失败部分result结算，Responses原已有；二者使用mandatory队列并提前快照Gin归属字段。gateway_compat_stream_failure处理仅注释后换号耗尽；严格终态校验还覆盖空流/仅ping EOF。真实SQLite夹具调用实际handler→RecordUsage→Apply，保留已删除Key、取消及去重结算。
- S06：antigravity/request_transformer仅追加tool_result base64媒体；antigravity_gateway_gemini与collectGeminiSSE保留函数调用/签名并接回放。Gemini handler只放行已解析到Antigravity的既有Gemini路由。pricing/billing按既有Claude5.5基准归一，显式条目优先；变体选择保留不同目标映射，只对自映射选精确已有变体。账号TestConnection按映射目标选既有协议，admin候选跳过通配符。antigravity/client只摘上游套餐细分ID及已有端点判断，未取新目录/Argon/3.8低档。
- S07：Gemini三个入口复用每个响应一次的ErrorPolicyResult，重试下一响应重置。gemini_messages_compat_service的handleGeminiUpstreamError/isVertexSharedCapacity429只对现有SA的无配额明细共享429免冷却；本地缺短冷却前置，补1分钟fallback而不引入账户/请求新scope枚举。vertex_service_account保留typed HTTP错误和固定可信Google端点；gemini_vertex_token_error只识别确切400 invalid_grant/account-not-found，SetError后换号。Ops沿用既有Stage/Scope/Reason，不新增代理字段；credentialFailoverClientResponse为Chat补Vertex认证reason，native/messages沿用401映射，三协议均验证耗尽安全提示，敏感查询值在账号错误及日志均脱敏。
- S08：gemini_stream_client_abort为三协议共用原请求取消、首个写错误和实际图片观察；Gemini兼容两文件在解析残行前判读错，ClientDisconnect及finishReason状态保留，写错后不再写/flush、不读取未来usage。gemini_image_output_accounting另存断流实际inline/file计数，正常完成旧fallback保留；断流即使已finish也不估图。Messages/native成功结果结算路径与Chat对齐mandatory及Gin快照，覆盖队列drop策略和请求复用。
- S09：image_storage共享默认连接池，逐跳URL和直连socket IP双层校验，代理分支只证明本地DNS；urlvalidator仅补私有既有函数的五行导出包装。组播/广播补强仅限图片下载，不顺带改变monitor策略。内容类型/嗅探拒绝JSON与HTML，保留base64、大小限制和body关闭。
- S10：go.mod、三份workflow、根及deploy Dockerfile同步Go1.27.3；两份pnpm override与九行lockfile定向更新follow-redirects1.16.1。Go1.27.3实测暴露Ent loader的export V5不兼容，经版本矩阵测定x/tools0.50.0为最小可用前置，只取MVS强制传递依赖、无tidy。详见[依赖复核](./implementation-0.2.18-tier2/dependency-review.md)，不把该必要前置冒充上游评估时已有结论。门禁额外修稳既有gemini_sse_comment_compat_test的固定时长关闭夹具及openai_responses_tool_schema_test的进程全局分配计数：前者等待实际心跳，后者隔离进程保留原阈值并通过线性退化负对照；没有修改对应生产逻辑。

范围仍为SQLite/miniredis/simple mode、VERSION1.1.18和迁移226。未新增schema、支付、新平台或模型目录；旧按需/架构不适用项继续见§5。前端全量测试2027通过/2既有跳过，lint/typecheck/build通过；目标follow-redirects两条告警消失，生产high仍为xlsx两条；Go包级扫描stdlib为0但仍有13个不同OSV。真实上游账号权限、Google凭据及远端代理DNS没有被本地桩证明。未发布或部署，最终提交状态见verification。

## 12. v1.1.19 提交与发布准备（2026-10-24）

用户明确要求commit、push main，再创建详细中文annotated tag v1.1.19并push tag。两档现有功能修复、定向依赖维护及实施证据共同提交为`2b79ac32078e4ac01e037c7ae1fa432d0d3ce1e0`，承接评估文档提交`498432759`；相对上一个tag v1.1.18（`873962ce7`）新增F01–F06和S01–S10，选择18个PR和1个直接fixture提交的行为。

VERSION在独立发布准备提交中从1.1.18改为1.1.19，必须先commit并push main再创建annotated tag。两份source-baseline、§1–§11、评估证据和原始实施验收所记录的VERSION1.1.18保持历史原值，不把发布准备冒充移植时基线。SQLite/miniredis/simple mode、迁移226和固定窗口保留，未部署运行中的业务实例。

详细中文tag正文见[发布说明](../releases/v1.1.19.md)，发布流程前置及当前检查见[readiness](../releases/evidence-v1.1.19/readiness.md)。完整GitHub Actions及上传结果以tag触发的实际运行记录为准，不能以本地检查代替。
