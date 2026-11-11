# 上游 v0.2.21 移植评估（v0.2.20 → v0.2.21）

评估日期：2026-11-11。GitHub releases API与git fetch确认本次最新正式版为v0.2.21，发布于2026-11-11 02:21:55 UTC / 10:21:55上海时间。范围按tag祖先固定，main后续两个直接提交单独记录。

建议**第一档11簇、第二档12簇**，优先Claude Code、Codex与共享网关正确性。当前为**评估完成、两批提案未实施**。这次同步产出完整来源清单、分档、隔离探针及OpenSpec；没有修改产品、VERSION、依赖或schema，没有commit/push/release或业务实例部署。

- [第一档方案](../../openspec/changes/port-upstream-0.2.21-tier1/README.md)
- [第二档方案](../../openspec/changes/port-upstream-0.2.21-tier2/README.md)
- [证据索引](./evidence-0.2.21/MANIFEST.md)

## 1. 版本、范围与规模

| 项目 | 冻结值 |
|---|---|
| fork HEAD | `4f4c80c3ce48ce4afbc9344536b9cf44ed3fd898` |
| VERSION / Go / SQLite / auth snapshot | `1.1.24 / 1.27.3 / 228 / 21` |
| private index tree | `4d58a09cfc75a321bd20f9a65f1cf156bdbc6441` |
| v0.2.20 commit | `dab829a4e9b7d1cc307189aea8b0319eb29e29d0` |
| v0.2.21 commit | `10ba8f9b4dc2e6023d4a9879ab77b21841babd10` |
| v0.2.21 annotated tag object | `43f15294c477bec427d586ddfc3f918ce1284c98` |
| upstream/main（观察） | `530c5e3a4f9c73ca2b0881191d0ef8e0d82c5484` |
| 非merge / 候选 | `44提交 / 34候选 = 33PR + 1直接提交` |
| 净diff | `200文件，+6699 / -541` |
| 候选×文件四态 | `214条：ALREADY0 / CLEAN121 / CONFLICT80 / NOBASE13` |

工作区开始时干净；Fast与v0.2.20两档均已在v1.1.24基线，不能继承旧评估“未实施”的历史表述。214是每候选文件之和，不是200净文件。第一档11PR、第二档12PR，合计23个不同PR；#8639为混合PR仅摘CLI身份维护。

## 2. 方法、渠道和边界

1. `git fetch upstream --tags --prune`，GitHub API取release元数据；[releases.json](./evidence-0.2.21/releases.json)、[baseline.json](./evidence-0.2.21/baseline.json)记录发布时间、tag对象和源码commit。
2. 全部PR以first parent→merge整体diff，覆盖44个非merge提交；[candidates.tsv](./evidence-0.2.21/candidates.tsv)保存父SHA、统计与SHA256，[commits.tsv](./evidence-0.2.21/commits.tsv)保存成员。不能按作者日期漏掉旧分支进入本tag的提交。
3. 同一个私有index对全部214文件先reverse、再forward，只有`--check`，不应用补丁。[files.tsv](./evidence-0.2.21/files.tsv)含逐条错误/退出码。ALREADY=0不是未合过任何相关行为；本地此前手工适配改变上下文。
4. [新增调用扫描](./evidence-0.2.21/added-call-inventory.tsv)1563条含方法/测试/内置函数，仅作候选搜索；[实际符号位置](./evidence-0.2.21/base-symbols.txt)与逐簇根因决定最小前置。CLEAN不能证明编译，NOBASE不能证明无缺陷。
5. 测试发现与基线、预期失败探针、未来实现验收分开。[80份来源测试文件](./evidence-0.2.21/source-test-inventory.tsv)、[76份本地integration首行](./evidence-0.2.21/integration-build-tags.tsv)保存build标签；不去掉postgres标签冒充SQLite执行。
6. 保留SQLite-only、miniredis、simple mode、固定quota窗口、usage dedup、完整Fast/Free Fast/Ultrafast与Terra。请求体预算64MiB、SSE行40MiB、idle读写120秒均为已验基线；本次不自动改成上游自适应最低1GiB。
7. 上游新增249覆盖索引、250 TTL migration；fork当前228。S08的覆盖索引如实施，新增本地229，按实际起点重查编号；TTL250属于未选新能力，不随行。PG `CONCURRENTLY`/`INCLUDE`不可复制，历史迁移checksum不可改。

### 全部候选分档

四态顺序为ALREADY/CLEAN/CONFLICT/NOBASE；完整候选含支付/文档/版本，不因排除就漏记。

| 候选 | 非merge | 四态 | 判断 |
|---|---:|---|---|
| 6b4b37db4 | 1 | 0/0/1/0 | 第四档：上游VERSION同步，不改fork 1.1.24。 |
| PR8610 | 2 | 0/2/4/0 | 第三档：自适应预算是新策略；1/4内存且最低1GiB与本地1C1G/64MiB目标不符，保留已验默认。 |
| PR8607 | 2 | 0/0/1/1 | 第四档：文档包含未选原生2.6/async能力，不能原文同步；通用格式核验提示可未来随图片文档选用。 |
| PR8603 | 1 | 0/0/0/1 | 第四档/N/A：本地无openai_responses_orphan_precheck_test.go及其AllocsPerRun目标；不导入整orphan性能体系来修来源测试。 |
| PR8590 | 1 | 0/3/1/0 | 第三档：Antigravity用量强制查询按钮是新交互，既有缓存刷新维护S09不需要此能力。 |
| PR8606 | 2 | 0/24/32/6 | 第三档：30m/12h独立计价、价卡、usage字段和迁移250；免费输入Fast修复属于新增12h分支，本地现有Fast对照通过。 |
| PR8381 | 1 | 0/2/2/0 | 第一档 F07 Sonnet 5.5 缓存读兜底价 |
| PR8384 | 1 | 0/1/1/0 | 第一档 F06 价格目录 hash 与正文一致 |
| PR8634 | 1 | 0/3/2/0 | 第三档：Fable/Mythos5.2新模型价格/effort，与#8639按最终tag一起评估，未选定。 |
| PR8564 | 1 | 0/1/1/0 | 第二档 S01 原生 WS 后续轮次失败关闭 |
| PR8358 | 2 | 0/4/0/0 | 第二档 S02 OAuth WS 搜索激活与完整回放 |
| PR8565 | 1 | 0/1/3/0 | 第二档 S03 WS 与 SSE 首 token 口径一致 |
| PR8576 | 1 | 0/2/0/0 | 第一档 F01 Codex 元数据与 failover |
| PR8579 | 3 | 0/1/3/0 | 第二档 S04 显式模型冷却覆盖容量重试 |
| PR8585 | 1 | 0/3/0/0 | 第一档 F02 Responses 密封参数与 namespace |
| PR8386 | 1 | 0/2/0/0 | 第一档 F03 CC Switch reasoning 封装 |
| PR8589 | 1 | 0/2/0/0 | 第一档 F04 Responses 工具回合会话种子 |
| PR8385 | 1 | 0/1/2/0 | 第二档 S05 Anthropic 兼容入口账号错误策略 |
| PR8584 | 1 | 0/2/0/0 | 第一档 F08 Gemini 非流式正文排除思考摘要 |
| PR8621 | 1 | 0/4/0/0 | 第一档 F09 Antigravity 并行工具图片顺序 |
| PR8568 | 1 | 0/3/1/0 | 第三档（条件性）：Pro Argon独立quota；本地没有内置Argon目录及quota-summary，通用自定义映射可能到达，若实际使用需连同映射/套餐验证，不声称绝对无消费者。 |
| PR8566 | 1 | 0/5/2/0 | 第四档：EasyPay退款成功码，支付功能排除。 |
| PR8631 | 1 | 0/1/1/0 | 第一档 F05 账号补全失败释放并发槽 |
| PR8637 | 1 | 0/10/0/0 | 第二档 S07 Refresh Token 原子轮换 |
| PR8161 | 3 | 0/10/6/1 | 第二档 S06 合成 effort 的映射与最终校验 |
| PR8346 | 1 | 0/4/2/0 | 第二档 S08 Key 列表独立加载用量与 SQLite 覆盖索引 |
| PR8581 | 1 | 0/2/2/0 | 第二档 S09 账号刷新保留行与已知配额 |
| PR8642 | 1 | 0/4/2/2 | 第四档：充值倍率精度，支付功能排除。 |
| PR8367 | 1 | 0/1/0/0 | 第三档：设置页悬浮保存按钮，新布局按需。 |
| PR8622 | 1 | 0/3/0/0 | 第二档 S10 Codex Desktop 页面上下文兼容 |
| PR8379 | 2 | 0/4/1/0 | 第一档 F10 Codex 图片请求保留错误状态码 |
| PR8575 | 1 | 0/1/1/0 | 第一档 F11 Responses UTF-8 输入边界 |
| PR8357 | 1 | 0/5/0/0 | 第二档 S11 既有安全审计 worker 入队唤醒 |
| PR8639 | 2 | 0/10/9/2 | 第二档 S12 Claude Code 内置版本基线；Fable/Mythos5.2目录/计价/新协议/UI第三档，当前只选CLI版本维护。 |

## 3. 第一档：已确认的窄缺陷

### F01 Codex 元数据与 failover（PR8576）

- **本地证据/前置**：metadata overlay 已复现4个子例失败，普通 metadata 两个负对照通过。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：service/openai_gateway_passthrough.go:openAIStreamEventIsMetadata。
- **行为**：Codex rate_limits/response.metadata SHALL 保持 attempt-local，真实模型输出前不得提交下游响应或结束首输出计时。
- **验收**：两种元数据后跟 overload，在普通及透传 SSE 都返回可换号错误且下游未写出。
- **保护边界**：普通 response.metadata 对照通过；真实文本、工具输出及终态按既有语义发送，不把所有 codex.* 任意当作可丢弃内容。

### F02 Responses 密封参数与 namespace（PR8585）

- **本地证据/前置**：sealed overlay 24子例中delta fallback 8失败，其余16通过；namespace单独也丢失。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：pkg/apicompat/types.go:ResponsesOutput；responses_to_chatcompletions.go:bufferedFuncCall/ProcessEvent/BuildOutput。
- **行为**：Responses 缓冲重建 SHALL 在函数调用中保留 namespace 和 encrypted_function_args 的缺失、null、空数组及非空值区别。
- **验收**：普通/透传、terminal output、item.done、delta fallback 三形态逐字段比对，工具参数字节和 usage 保留。
- **保护边界**：现有 terminal 与 item.done 原文路径已经通过；只补 delta accumulator 缺口，不重复重构已实现的终态保真。

### F03 CC Switch reasoning 封装（PR8386）

- **本地证据/前置**：foreign reasoning overlay 失败，原生签名负对照通过。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：pkg/apicompat/anthropic_to_responses.go:anthropicAssistantToResponses。
- **行为**：Anthropic 转 Responses SHALL 忽略 ccswitch-openai-reasoning-v1 外来封装，并保留原生密文及工具历史。
- **验收**：合法/非法/带空白封装均不作为 encrypted_content 发送，普通密文继续往返。
- **保护边界**：文本、tool_use、tool_result、签名顺序不变；不把未知所有签名一概删除。

### F04 Responses 工具回合会话种子（PR8589）

- **本地证据/前置**：本地后续developer提示被拼入种子；SessionPrefix overlay 确认字符串变化。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：service/openai_content_session_seed.go:deriveOpenAIContentSessionSeed。
- **行为**：未提供显式会话ID的 Responses 请求 SHALL 仅使用稳定前缀和首个用户输入派生会话种子，后续 developer/system 提醒不能改变已有会话。
- **验收**：用户/assistant/function call/output/input_text 五种前缀边界；多回合工具消息仍命中原粘性账号。
- **保护边界**：首个系统/开发者提示、模型、工具、instructions、首用户差异仍隔离；显式 session-id/prompt_cache_key 优先。

### F05 账号补全失败释放并发槽（PR8631）

- **本地证据/前置**：失败overlay实际release=0，成功对照通过；本地签名适配仅在探针中。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：service/gateway_scheduling.go:newSelectionResult/hydrateSelectedAccount。
- **行为**：Gateway 选中账号后补全失败时 SHALL 释放已取得的并发槽，并保留本地账号会话释放。
- **验收**：取消、仓储错误、账号消失时release恰好一次；成功时交给调用者释放。
- **保护边界**：本地 newSelectionResult 多一个 sessionID 参数且已有 ReleaseAccountSession；不可用上游函数整体覆盖。

### F06 价格目录 hash 与正文一致（PR8384）

- **本地证据/前置**：PricingHash overlay下载次数实际1而期望2；remote stub不接触外部价格服务。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：service/pricing_service.go:downloadPricingData/syncWithRemote。
- **行为**：价格同步 SHALL 使用已下载正文的实际 hash；远端 hash 超前时不能把旧正文标记为最新。
- **验收**：hash接口先更新、正文后更新：下一轮必须再次下载并更新价格；随后内容一致可跳过。
- **保护边界**：远端hash缺失/大小写、无效JSON、网络失败、fallback/override及热重载并发保留，失败不发布半成品。

### F07 Sonnet 5.5 缓存读兜底价（PR8381）

- **本地证据/前置**：本地实际0.20/百万；SonnetPrice overlay失败。本结论是与冻结tag一致性，非独立市场价格审计。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：service/billing_service.go:initFallbackPricing；pricing_service.go:claudeSonnet55FallbackPricing；resources/model-pricing/model_prices_and_context_window.json。
- **行为**：内置 Sonnet 5.5 缓存读取价格 SHALL 与 tag 的 0.10 USD/百万 token 一致，覆盖资源和两层 fallback。
- **验收**：远端缺失、资源定价与 Billing fallback 三路、别名及Fast费用校验。
- **保护边界**：显式自定义0与覆盖价优先；其它模型和输出/缓存创建费率不改。

### F08 Gemini 非流式正文排除思考摘要（PR8584）

- **本地证据/前置**：GeminiThought overlay复现正文含private planning；前置已逐行量测。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：service/gemini_messages_compat_service.go:convertGeminiToClaudeMessage；gemini_chat_completions_compat_service.go消费者。
- **行为**：Gemini 非流式兼容响应 SHALL 排除 thought=true 的文本，不将思考摘要混入正文，并保留正常文本、工具和用量。
- **验收**：Messages和Chat两消费者、仅thought终态、混合工具与thoughtSignature、MAX_TOKENS及thoughts token统计。
- **保护边界**：本地缺 isGeminiThoughtPart，但来源helper仅4行布尔判断；只补这一最小前置，不导入320行完整thinking策略。

### F09 Antigravity 并行工具图片顺序（PR8621）

- **本地证据/前置**：ParallelToolImages sideband失败、native对照通过。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：pkg/antigravity/request_transformer.go:buildParts。
- **行为**：非原生多模态工具输出的转换 SHALL 连续写出所有 functionResponse，再追加附带图片。
- **验收**：至少两张图片加纯文本工具结果，Claude/Gemini既有模型和Responses兼容入口都保留ID、内容和顺序。
- **保护边界**：支持原生 functionResponse.parts 的模型继续嵌入媒体；单工具及无图片顺序不回退。

### F10 Codex 图片请求保留错误状态码（PR8379）

- **本地证据/前置**：ImageStatus overlay期望400实际200；是现有心跳消费者缺陷，不依赖新模型。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：service/openai_images_json_keepalive.go:StartOpenAIImagesJSONKeepalive；既有handler图片入口。
- **行为**：携带 X-Codex-Image-Turn-Id 的既有非流式图片入口 SHALL 延后提交响应，保留最终上游错误HTTP状态。
- **验收**：真实本地HTTP generation/edit延迟错误与成功；超过原心跳间隔仍能返回400，并检查代理超时。
- **保护边界**：其它客户端保留既有心跳，流式不改；该header不是认证。只维护现有图片入口，不引入2.6或新原生路由。

### F11 Responses UTF-8 输入边界（PR8575）

- **本地证据/前置**：9个非法UTF8诊断子例失败，有效Unicode对照通过；较上一轮PG论据改为独立协议输入边界。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：handler/openai_gateway_handler.go:Responses。
- **行为**：Responses 请求体 SHALL 在JSON转换和路由前拒绝非法UTF-8并返回明确的400错误。
- **验收**：非法首/中/尾字节、截断多字节、流/非流、compact；有效中文/emoji继续进入字段检查；真实handler不得发起上游重试。
- **保护边界**：本次探针仅证明非法字节进入model字段检查，不证明曾产生真实上游503或PG故障；保留大小/预算释放和鉴权次序。


## 4. 第二档：跨路径适配、数据一致性及维护

### S01 原生 WS 后续轮次失败关闭（PR8564）

- **本地证据/前置**：源码复核：本地已有CurrentTurnFailover包装及handler处理，但主函数返回匿名error，缺后续轮次兜底；需适配返回路径。未在本次跑两轮WS新行为探针。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：service/openai_ws_forwarder_ingress.go:ProxyResponsesWebSocketFromClient；openai_ws_forwarder.go；handler/openai_gateway_handler.go。
- **行为**：原生 Responses WS 后续轮次不能安全迁移完整会话时 SHALL 以可重试关闭结束，不能跨账号重放首帧。
- **验收**：第一轮成功、第二轮限流/过载：客户端1013关闭，第二账号未收到首轮任务；首轮failover仍允许。
- **保护边界**：保留每轮Fast/模型/usage、lease与槽位释放；只有当前轮完整历史足够时才允许已有安全重试。

### S02 OAuth WS 搜索激活与完整回放（PR8358）

- **本地证据/前置**：AllItems、两个search检测器和当前轮回放builder本地已存在；新增来源helper60行及连接能力状态，未运行新能力专用WS探针。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：service/openai_ws_pool.go；openai_ws_forwarder_ingress.go；新窄search activation helper。
- **行为**：OAuth 池化 WS 首次声明web_search时 SHALL 校验连接初始能力，必要时新建连接并携带完整历史。
- **验收**：无搜索prewarm→搜索、Lite additional_tools、工具续接保留reasoning/message/call/output；API Key及同名function不重连。
- **保护边界**：未知previous_response_id或不完整工具历史要明确关闭，不能猜测回放；写失败不能标记初始化，状态受write锁保护，连接关闭/预算回收。

### S03 WS 与 SSE 首 token 口径一致（PR8565）

- **本地证据/前置**：本地仍硬编码isTokenEvent；openAIStreamDataStartsTTFT存在可复用。基线WS包通过，新增模式行为本次仅静态核实。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：service/openai_ws_v2/passthrough_relay.go:RelayOptions/observeUpstreamMessage；openai_ws_v2_passthrough_adapter.go。
- **行为**：WS 首token统计 SHALL 遵循管理员TTFT模式，并在每轮独立计算。
- **验收**：token/first event等所有已支持模式与SSE对照，至少两轮，空delta/工具输出/终态/codex元数据/取消。
- **保护边界**：本地上一轮已经修每轮计时；保留本地起点接口，不引入来源FirstTurnStartedAt/TakeNextTurnStartedAt整基座只为编译。

### S04 显式模型冷却覆盖容量重试（PR8579）

- **本地证据/前置**：HTTP及四reader×三凭据×两错误形态overlay失败。来源WithModel helper本地缺，不能直接复制测试；需贯通canonical model和terminal副作用，保留本地非流式failover修复。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：service/openai_account_runtime_block_fastpath.go；openai_gateway_passthrough.go；openai_gateway_upstream_errors.go；既有流/WS错误消费者。
- **行为**：OpenAI 显式model范围冷却规则 SHALL 对 overload/slow_down及流内503生效，且只匹配实际错误字段。
- **验收**：HTTP、普通/透传SSE、两种缓冲路径及WS：匹配后停止同账号重试，作用于已映射模型，持久化/缓存一致。
- **保护边界**：未配、account范围、无模型、排除码及其它平台保留请求级默认；output/metadata关键词不能触发；写失败不扩大整账号冻结。

### S05 Anthropic 兼容入口账号错误策略（PR8385）

- **本地证据/前置**：compat-policy overlay复现选定规则被默认failover状态列表屏蔽，未配/默认500等负对照通过。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：service/gateway_forward_as_chat_completions.go；gateway_forward_as_responses.go错误处理。
- **行为**：Anthropic账号经Chat或Responses兼容入口时 SHALL 执行已有自定义错误码及临时不可调度策略。
- **验收**：410自定义/临时规则在两入口、流/非流生效并留给handler换号；未配410及400保持原响应。
- **保护边界**：pool模式重试与shouldDisable结果一致；不要将全部客户端4xx升级为failover，不扩商业支付策略。

### S06 合成 effort 的映射与最终校验（PR8161）

- **本地证据/前置**：effort-binding overlay两平台6子例失败，其余边界通过。需raw显式校验→转换→策略→最终校验，不能简单删除前置校验。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：handler/composite_platform.go；service/openai_reasoning_effort_policy.go/openai_gateway_messages*.go/openai_gateway_count_tokens.go；pkg/apicompat转换选项。
- **行为**：Messages转OpenAI SHALL 对显式disabled思考合成的none执行一次映射及上限策略，并在映射后校验最终effort。
- **验收**：disabled+none映射/deny/ceiling、显式非法值、普通省略default、Sol/Terra none/max、count_tokens和Chat fallback一致。
- **保护边界**：普通省略不能被medium映射/上限意外改变；现有GLM通用兼容降级保留；不引入Kimi/Zhipu等原生平台或requested_effort schema。

### S07 Refresh Token 原子轮换（PR8637）

- **本地证据/前置**：refresh-race overlay八请求实际全部成功；不需要上游支付或多实例平台前置。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：service/auth_service.go/refresh_token_cache.go；repository/refresh_token_cache.go；相关handler/service测试stub。
- **行为**：刷新令牌 SHALL 在校验后原子消费旧token，保证同一token并发只签发一份新令牌；缓存错误必须失败关闭。
- **验收**：同步barrier八请求仅一成功；实际miniredis原子GET+DEL、过期/重用/禁用用户/token version/家族吊销/缓存失败。
- **保护边界**：新增接口需更新所有stub；保持预校验顺序，失败不泄露token；单节点miniredis真实执行，不能仅用mock证明原子性。

### S08 Key 列表独立加载用量与 SQLite 覆盖索引（PR8346）

- **本地证据/前置**：本地await统计阻塞loading；真实modernc SQLite完整基线迁移（最大编号228）+20000合成记录，新增临时覆盖索引计划命中且仓储结果前后一致；非生产延迟收益证明。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：frontend/src/views/user/KeysView.vue；backend/internal/repository/usage_log_repo_stats.go；拟议backend/migrations/229_*.sql。
- **行为**：API Key列表 SHALL 在列表响应到达后展示，用量独立异步加载；统计查询可采用经验证的SQLite覆盖索引。
- **验收**：延迟/失败/取消用量不阻塞列表；快速筛选分页旧请求不能回写；空列表及原用量区间数值一致。
- **保护边界**：SQLite拟议229而非上游249；(api_key_id,created_at,actual_cost)先验证写入/大小代价与查询计划，禁止CONCURRENTLY/INCLUDE及修改旧迁移。

### S09 账号刷新保留行与已知配额（PR8581）

- **本地证据/前置**：本地loading整表骨架及quota分支仍存在；59项前端邻近基线通过，新交互本轮未跑浏览器。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：frontend/src/views/admin/AccountsView.vue；components/account/AccountUsageCell.vue。
- **行为**：账号列表和配额刷新 SHALL 保留当前已展示内容，并清晰标识刷新中的不可交互状态。
- **验收**：初次加载骨架、刷新保留DOM与额度、失败保留旧值、成功更新；分页/筛选/账号复用/批量选择/移动端。
- **保护边界**：只保留同账号数据；旧请求不可覆盖新账号；不依赖未选Antigravity新查询按钮或Ollama基座。

### S10 Codex Desktop 页面上下文兼容（PR8622）

- **本地证据/前置**：两项前置本地已有，来源新helper123行；本地仅automation/delegation正常化。静态确认缺转换，未宣称实际Desktop联机复现。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：handler/openai_gateway_handler.go:Responses；新增窄openai_page_context.go；现有mayContainCodexCallOutputCandidate/hasUniqueJSONMembers。
- **行为**：符合严格签名的Codex页面function_call_output SHALL 转为标记的非信任user文本，避免孤儿工具结果错误。
- **验收**：open_page及selected_page分片、历史重放、混合automation；真实HTTP捕获首轮消息，顺序和页面原文保留。
- **保护边界**：有真实call_id、错误namespace/name、重复JSON键、坏page_id/分片头均不转换；不提升system权限；日志不含页面内容。

### S11 既有安全审计 worker 入队唤醒（PR8357）

- **本地证据/前置**：本地500ms ticker才claim，无wake接口；属于已有功能延迟优化，基线Worker通过，收益需实施时实测。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：backend/internal/securityaudit/prompt_enqueue.go/prompt_service.go/prompt_worker.go。
- **行为**：已有审计任务成功发布后 SHALL 非阻塞唤醒本进程worker，同时保留轮询拾取重试任务。
- **验收**：有任务立即claim、队列满/持久化失败不通知、连续任务合并唤醒、禁用/并发启停/Shutdown与到期重试。
- **保护边界**：保持SQLite claim原子性和1C1G资源界限；不新增审计规则、远端扫描策略或实例运营能力，不把来源平均250ms当本地实测。

### S12 Claude Code 内置版本基线（PR8639）

- **本地证据/前置**：本地2.1.258，CLIVersion覆盖基座完整；这是已有身份维护，列第二档，不声称复现版本导致拒绝。
- **patch sites**（backend/internal相对路径，另标frontend/resources除外）：pkg/claude/constants.go:CLICurrentVersion；cli_version.go与既有版本测试/setting同步测试。
- **行为**：既有Claude Code身份链 SHALL 统一使用2.1.321内置基线，并保留配置与有效环境覆盖规则。
- **验收**：CLIVersion、User-Agent、x-app/指纹相关已有消费者与版本同步服务一致；合法高版本覆盖保留，低版本/非法值回退。
- **保护边界**：仅constants的版本literal和必要期望测试；不引入Fable/Mythos5.2、目录/定价/白名单/UI等同PR新能力。

## 5. 暂不做与历史遗留

### 5.1 本轮第三档

| 能力 | 当前判断 |
|---|---|
| #8606 30m/12h TTL | 新增独立usage桶、用户/账号统计价卡、UI和schema，需要SQLite迁移及Ent生成；与既有Anthropic5m/1h不能混用。若用户选定，需要完整HTTP/SSE/WS计量→结算→去重→展示验收。通用写入价显式配置须保持平价，nullable与显式0不能混淆。 |
| #8634 + #8639 Fable/Mythos5.2 | 两个新模型及effort、Bedrock/白名单/OpenCode/价格链，未选定；S12只维护原有CLI身份，不让新模型借道进入。 |
| #8610 自适应预算 | 读取physical/cgroup v1/v2父层/GOMEMLIMIT最小值的helper104行可独立复用，但上游`max(memory/4,1GiB)`是增加大实例预算，不是保护1GiB小实例。当前已有64MiB测量；未发现必须更改默认的本地证据。将来须独立设计上限/下限、显式0、空环境变量与1C1G压力。 |
| #8590 Antigravity查询 | 新强制刷新按钮；与S09保留已知值的维护独立。不能让UI新增按钮自动激活未选quota-summary接口。 |
| #8568 Pro Argon quota | 当前无内置Argon目录与quota-summary，但通用自定义模型映射可能到达该模型。若用户实际配置Pro Argon，此项需提升为现有消费者维护并验证plan别名/Ultra/未知套餐、存量cooldown键；本次未读取业务账号，不假定已使用或不可达。 |
| #8367 保存浮标 | 设置页布局便利性，未选定；无必要基座关系。 |

### 5.2 第四档、N/A和混合PR拆分

- #8566 EasyPay退款与#8642充值精度全部排除。个人网关usage/cost统计保留，不引入支付provider、UI、依赖或迁移。
- #8603只隔离上游不存在于本地的orphan allocation测试；不能为了修测试导入未选的完整JSON/COW清理体系。
- #8607上游README描述原生2.6、fallback与异步图片链，原文不适用于本fork。通用MIME核验有价值，但不是导入那些入口的理由。
- 直接VERSION与tag之后赞助商/版本提交不采用；fork维持自己的1.1.x版本。
- #8161原生CN平台分支不引入；既有GLM通用兼容规则仍要作为负对照。#8379可在现有keepalive消费者修复，不需要原生新图片能力。
- #8606发布说明“普通输入免费时Fast倍率丢失”实际修的是新增12h价格乘数。该字段本地不存在；现有InputPrice=0、Terra输出Fast两倍的对照通过。不能按发布标题单造一个本地Fast bug。

### 5.3 旧轮次逐项当前口径

以下完成依据本地提交/归档记录，不宣称本次重跑旧轮全部验收。原报告、sources、tasks、verification保留原时点内容；本轮仅新增状态说明。

| 旧范围 | 2026-11-11当前状态及保留原因 |
|---|---|
| v0.2.20 第一/第二档 | 29681fdf5、a25ce7bed已实现，44/44及38/38，已发布v1.1.24；活动change待归档，不重开旧任务。 |
| Fast/Free Fast/Ultrafast | 326513cdc完成，SQLite228及完整HTTP/Chat/Messages/Responses/WS；本轮对照确认免费输入仍保留已有输出Fast倍率。 |
| v0.2.20 MCP/admin Key、邮箱黑名单 | #8389/#8504新管理授权/注册策略按需；#8476依赖未选#8319受限admin Key，修补不构成引入授权，支付操作排除。 |
| v0.2.20图片/导出/配额策略 | #8462尺寸路由、#8557 Image2.6、#8543/#8521冷却配置、#8522 quota-summary、#8380/#8495导出默认策略仍按需。S09不需要新查询按钮，F10不需要新原生模型。 |
| v0.2.20性能与缺平台项 | #8518完整ojson/COW/compact/orphan仍需收益及所有权测量；必要资源治理与RPM空值缓存已实现。#8503原生Kimi、#8549 transcript阻断算法、Voice/cache CRUD缺消费者，仍不取；显式session ID策略保留。 |
| v0.2.20发布后观察 | 原记录#8610/#8606这次已进入tag并重新完整分档，不继续标记为tag之外；其他main观察仅记录本轮tag之后提交。 |
| v0.2.19第一/第二档与Terra | 683354b2c、cccf931ef、a541d0b61完成并归档；Terra在使用范围，S06维护其none/max和映射。 |
| v0.2.19 Grok诊断/历史兼容 | becd2d2a3完成目录诊断、参数/别名/defer/序号尾项；本次F02是delta重建的后续缺口，terminal原文保留已通过，不能说之前整个修复无效。 |
| v0.2.19其它 | #8319受限admin、#8323真实图片张数、#8348 50MiB上传、#8287同号重试配置、#8395更广目录仍按实际用途；上传完整性/fileData/声明目录已完成。 |
| TPS/运维 | f333aa5f3完成整窗口P5/P10/P50/均值/样本数；usage TPS和完整ops入口已完成。新趋势图等按需。 |
| v0.2.18主体/个人能力 | 2b79ac320、0e38d06d8、b554bea55完成两档、日志/筛选/冷却/搜索/min tokens与SQLite后台恢复；1.1.23另修日志清空/写入顺序。不可重新关闭SQLite后台。 |
| v0.2.18性能/平台 | #8251完整共享解码COW、#8269 SQLite轻量quota查询仍需收益测量；Argon/思考摘要/SA quota范围/多平台生图开关/cache CRUD/Vertex API Key仍按需。F08仅4行bool前置，不把完整Gemini thinking体系带入。 |
| v0.2.16 | 7c544c03f、847db09e1、5ff6e5bf9主体完成；Haiku5.5/group i18n按需；模型冷却/TPS由后续完成。 |
| v0.2.12 / v0.2.11 | 405246789、96f32d018主体及快调/排序完成；TypeSafe缺平台，支付推广/查单排除。Claude原生重置管理、风控白名单、Key限制、在途预占按需；Ultrafast已完成。 |
| v0.2.9 | 第一档用户确认并归档、第二档159d04222；OpenCode/Command Code/Kimi原生、plugins、完整model_allowlist glob/pinned/live、视频路由、多实例备份仍未整体采用。 |
| v0.2.5 quota与WS | b349c7408完成33/33和SQLite227；默认WS倍率1.0为已批准选择，不采用5.0。#6960/#7049 exec_scope、抢占/queue-wake仍需单独并发设计，S01/S02不需要它们。 |
| v0.2.5模型/平台 | #6575/#6890/#6974/#6769 Gemini新能力/原生图片/峰谷/Ollama重置；#6747/#7091/#6691/#7153新平台/批量；#6989/#7057/#6968/#7011/#6945/#6988目录/compact/pinned仍按需。#6986支付排除。 |
| v0.2.5其它协议 | #7129 Grok全流序号、#7085 input stripper、#6929 Gemini带内错误、#6917监控BaseURL及大管理展示未整体采用；新窄修复不代表整簇已完成。 |
| v0.2.4/v0.2.0 | a26b6b55d/c9d9bebe8主体；5b3e1a716完成0.2.0尾项128/128（类型桩、四面浏览器、组合门禁）。 |
| v0.1.180/v0.1.184计价/usage | 5250907f5模型参考价及Fast已完成；动态长上下文阶梯、独立effort倍率、usage compaction/requested effort持久化、峰谷/重置锚点仍按需或需schema。新TTL是另一个独立schema选择。 |
| v0.1.180工具/Grok/WS及更早 | PDF/tool身份/HTTP回放/孤儿工具/composite闸门及历史四尾项已完成；Grok429/Realtime/抢占按需，跨节点Redis租约单节点不需要；#5925最小delay随容量消费者完成但非整簇。0.1.176/179/183/184主体和Responses Lite已验收；128c5f7bf固定窗口与227/228语义保留。 |

### 5.4 当前依赖与发布后main

本次tag没有要求升级本地依赖的必要前置。实际`pnpm audit --prod --json`仍为12 moderate / 2 low / 0 high / 0 critical；[摘要](./evidence-0.2.21/dependency-summary.json)和[原始结果](./evidence-0.2.21/frontend-production-audit.log)单列，不用此结论代替Go或开发依赖全量审计。退出1表示仍有advisory，不是测试失败；没有改依赖或自动延长例外。

[tag后列表](./evidence-0.2.21/post-release-commits.txt)只含VERSION同步和赞助商更新，不纳入实现建议；没有把当前main当tag来源。

## 6. ALREADY、N/A和最小前置结论

- 无ALREADY文件不意味着所有语义缺失：sealed terminal/item.done已有完整保真；只有delta fallback缺标记/namespace。
- Gemini helper缺文件但只有4行，明确作为F08最小前置；320行thinking策略不需导入。
- 冷却WithModel helper本地缺失却有现有模型规则消费者；S04需改造真实错误链，不能按NOBASE排除，也不能声称小hunk即可完成。
- WS回放AllItems及search检测已存在；S02不需要重写完整WS系统。S01匿名error返回与来源named return差异需处理。
- S12只摘CLI身份版本，不以维护名义加入5.2模型；Fast既有零输入对照通过，新TTL缺口保持第三档。

## 7. 建议实施顺序

1. F05释放槽位与S07原子刷新先处理资源和认证正确性；F01/F02/F03/F04及S01紧接，优先Claude/Codex实际故障。
2. S02/S03共用WS文件按顺序适配，验证每轮上下文、Fast、取消、usage；S04/S05错误策略分别覆盖HTTP、流和缓冲，保留模型作用域。
3. S06 effort与S10页面上下文，跑Sol/Terra与普通工具调用负对照；S12 CLI身份独立摘取和验证。
4. F06/F07价格、F08/F09媒体/正文、F10/F11输入输出边界可逐簇落地。
5. S08/S09列表交互及SQLite索引、S11审计唤醒随后；索引须新增迁移并测写入成本，不能只展示covering计划。
6. 两批全部完成后填写实际verification，归档/发布/推送/部署另行处理。

## 8. 本次自测与实施验收门槛

| 已执行 | 结果与限制 |
|---|---|
| 整PR/四态/符号/测试目录 | 34候选覆盖44非merge；214文件全部先反后正；未应用产品补丁 |
| 后端8包发现与定向执行 | 两命令退出0；service、WS v2、handler、repository、apicompat、antigravity、config、securityaudit；[发现](./evidence-0.2.21/baseline-test-list.log)/[执行](./evidence-0.2.21/baseline-targeted.log) |
| Claude身份邻近基线 | `pkg/claude`与service版本相关测试通过；[输出](./evidence-0.2.21/cli-identity-baseline.log) |
| 前端5文件 | 59测试通过；KeysView、AccountUsageCell、usage refresh、Accounts lite/select-all；[输出](./evidence-0.2.21/baseline-frontend.log) |
| Go隔离overlay | 4包全部实际编译；最终16个顶层新行为测试按预期失败，既有对照通过；[摘要](./evidence-0.2.21/probe-summary.json)/[复现](./evidence-0.2.21/reproduce-probes.py) |
| Refresh race probe | barrier控制8并发均读到旧token，实际8成功，期望仅1；service层真实逻辑+锁保护stub，不冒充真实Redis原子实现验收 |
| SQLite覆盖索引可行性 | 实际modernc SQLite、完整基线迁移（最大编号228）、20000合成记录，仓储结果一致且使用covering index；[输出](./evidence-0.2.21/sqlite-index-feasibility.log)，未修改产品schema，不声称生产延迟收益 |
| 当前生产依赖审计 | 12 moderate/2 low；没有Go/全开发依赖全扫 |
| 文档/保护范围 | [校验](./evidence-0.2.21/document-validation.json)：两change严格校验、所有实施任务未勾选、链接及历史/产品哈希保护 |

探针的已知过程失败单独保存：首次手写BillingService夹具未初始化fallback map造成panic，修正夹具后才计入产品价格断言；直接来源冷却测试引用本地不存在的WithModel方法编译失败，改为本地HTTP和四reader实际入口重测。两者都不是产品缺陷证据。最终失败是期望新行为在旧基线不满足，不是实现验收失败。UTF8探针仅证实早期校验缺口，没有真实远端重试结论。WS搜索/后续轮次与Desktop仍为静态适用性核对，实施必须补真实本地HTTP/WS复现；本轮没有真实业务账号联机或浏览器验收。

未来实现必须执行后端全量`go test -tags=unit ./...`、embed build、lint、相关race；真实SQLite/miniredis覆盖结算、成功去重、quota固定窗口和迁移。S07需真实miniredis并发原子测试；S01/S02/S03/S04需实际WS/SSE多回合、取消与lease释放；前端需typecheck/lint/test/build/critical和本地浏览器快速切换/失败/移动端。S08迁移先临时库，测索引大小、写入与查询成本；不把评估的PASS回填实施槽位。

## 9. 本轮通用教训

- 发布摘要中的“修复”可能只修新增能力里的分支：先追踪字段存在性，再复现；不要把12h Fast倍率缺陷套给旧通用Fast。
- 动态预算的下限比“自适应”标题更重要：最低1GiB不能保护1GiB实例，保留已测本地默认。
- 同一个协议字段在terminal、item.done、delta fallback可能走不同所有权路径；完整路径对照比整PR状态更可靠。
- 原子令牌轮换必须用可控并发窗口复现；GET+DEL即使都返回成功也不构成单次消费。
- 槽位释放需保留fork已有会话租约释放，不能整个函数照搬覆盖本地额外责任。
- 小helper可成为最小前置；缺失完整错误消费者则需列第二档，不能从补丁行数推断风险。
- SQLite现有索引已能按Key/时间检索；覆盖索引能用第三列替代PG INCLUDE，但查询计划只是可行性，不是迁移写入成本和生产收益的证明。
