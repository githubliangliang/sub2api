"""Human-reviewed scope, behaviour, evidence and acceptance boundaries."""
import csv
import json
from pathlib import Path

OUT = Path(__file__).resolve().parent
review = []

def add(key, title, prs, behaviour, evidence, sites, acceptance, boundary, base):
    review.append(dict(id=key, tier=1 if key.startswith("F") else 2, title=title,
        candidates=prs.split(), behaviour=behaviour, evidence=evidence, patch_sites=sites,
        acceptance=acceptance, boundary=boundary, prerequisites=base))

add("F01", "Responses 密封调用内容", "PR8682",
    "函数调用的 delta 回退重建 SHALL 保留 encrypted_content。",
    "apicompat overlay 丢失 sealed-content；namespace/密封参数对照通过。",
    "backend/internal/pkg/apicompat/responses_to_chatcompletions.go: bufferedFuncCall、ProcessEvent、BuildOutput",
    "terminal/item.done/delta 三路比对调用 ID、密封内容和原参数。",
    "保留上一轮已实现的 namespace、encrypted_function_args 与参数完成事件语义。",
    "本地 accumulator 完整存在；只缺一个字段及两个赋值。")
add("F02", "对话中途 system 权限", "PR8703",
    "Chat/Anthropic 中途 system 消息转 Responses SHALL 保持 developer 权限。",
    "MidSystemAuthority overlay 实际 role=user。",
    "backend/internal/pkg/apicompat/anthropic_to_responses.go: anthropicMsgToResponsesItems",
    "首部系统提示及中途 hook 文本、工具历史的权限与顺序。",
    "不把用户文本或工具结果提升为系统权限。",
    "既有 user/assistant 转换存在；新增窄 system 分支。")
add("F03", "JSON Schema 默认 strict", "PR8761",
    "Chat json_schema 未设置 strict 时转 Responses SHALL 显式写 false。",
    "StrictDefault overlay 缺 strict；显式 true 对照通过。",
    "backend/internal/pkg/apicompat/response_format.go: chatResponseFormatToResponsesTextFormat",
    "缺省/true/false/null、schema 原值和普通 text/json_object 对照。",
    "不覆盖客户端显式值，不新增结构化输出能力。",
    "既有格式转换函数存在；五行默认值维护。")
add("F04", "Codex HTTP 外部上下文", "PR8769",
    "无对应调用的 agent browser_info/goal_objective/untrusted_context 输出 SHALL 转为不可信用户文本。",
    "handler overlay browser_info 未转换；真实 call_id 对照通过。",
    "backend/internal/handler/openai_page_context.go、openai_gateway_handler.go: normalizeCodexPageContext/Responses",
    "三种上下文及已有 page 快照、空输出、重复成员、真实调用、未知名称。",
    "保留 page 快照严格校验和 user 权限，不能吞正常 function_call_output。",
    "上一轮 page normalizer 已存在；扩大名称集合，不导入账号命名空间。")
add("F05", "Claude per-turn beta", "PR8742",
    "任一已支持的 per-turn/mid-conversation beta SHALL 保留消息级 output_config。",
    "PerTurnBeta overlay 被净化；原 beta 和无 beta 对照通过。",
    "backend/internal/pkg/claude/constants.go；backend/internal/service/gateway_request.go: stripAnthropicMessageOutputConfigUnlessBeta",
    "API Key/OAuth、空 system 控制消息、有正文消息和顶层 output_config。",
    "无相关 beta 仍执行既有净化，不能放开全部消息字段。",
    "现有 beta 净化函数存在；新增常量及一次精确 token 判断。")
add("F06", "count_tokens 请求头唯一性", "PR8788",
    "Anthropic count_tokens 透传 SHALL 以原始大小写无关读取/写入避免重复默认头。",
    "修正 DI 夹具后 header overlay 复现同名头计数为2。",
    "backend/internal/service/gateway_count_tokens.go: buildCountTokensRequestAnthropicAPIKeyPassthrough；gateway_service.go 调试头读取",
    "真实 HTTP 线格式大小写/多值、覆写和缺省头各只有预期值。",
    "保持账号鉴权、beta 净化与覆写的最终优先级。",
    "getHeaderRaw/setHeaderRaw 已存在；不能改回 net/http 的大小写假设。")
add("F07", "输出后的兼容流账号副作用", "PR8814",
    "OpenAI Chat/Messages 已输出后的流错误 SHALL 执行既有账号/模型错误副作用。",
    "静态：两 reader 在早期 failover 分支之后直接记录错误，缺副作用调用。",
    "backend/internal/service/openai_gateway_chat_completions.go: handleChatStreamingResponse；openai_gateway_messages.go: handleAnthropicStreamingResponse",
    "文本/工具输出后 401/429/overload、模型规则、普通请求错误与取消。",
    "已输出后不跨账号重放；沿用模型作用域，不扩大整号冷却。",
    "handleOpenAIStreamTerminalAccountSideEffects 和模型规则已有消费者。")
add("F08", "既有 Sol Ultra Fast 倍率", "PR8812",
    "已支持的 GPT-6.1 Sol Ultra Fast SHALL 使用冻结 tag 的6倍模型倍率。",
    "SolUltrafast overlay 实际缺模型倍率；Astra6倍对照通过。",
    "backend/internal/service/billing_service.go: applyModelSpecificPricingPolicy；openai_model_alias.go 既有型号判断",
    "Sol/Astra、别名、HTTP/WS结算、Fast2倍、Free Fast、自定义0和长期上下文。",
    "只摘既有型号的倍率修复；强制 Ultra Fast 分组、新渠道倍率和 GPT-6.2 不随行。",
    "本地完整 Fast 已实现；Sol 型号 helper 已有，源新增支持谓词可窄适配。")
add("F09", "登录身份回填空指针", "PR8684",
    "登录邮箱身份回填失败返回 nil 时 SHALL 安全跳过默认设置判断。",
    "LoginNilIdentity overlay 捕获 panic；历史回填来源对照通过。",
    "backend/internal/service/auth_service.go: shouldApplyEmailFirstBindDefaults",
    "身份创建/查询失败、既有身份、回填/首次绑定来源与正常登录。",
    "不引入新的注册、赠额或支付策略。",
    "函数在 nil 检查前读 Metadata；三行 guard 即可。")
add("F10", "5h 滚动保留有效周用量", "PR8723",
    "Claude 5h 窗口滚动 SHALL 仅清除已过期的周读数。",
    "WeeklySurvivesFiveHourRollover overlay 清空未到期7d读数。",
    "backend/internal/service/ratelimit_service.go: UpdateSessionWindow",
    "未来/过期/缺 reset、仅7d响应、秒/毫秒重置和样本时间。",
    "只清5h和过期周值；Fable 模型/新额度接口仍独立按需。",
    "既有通用7d被动采样存在；源 rollover helper30行可独立适配。")
add("F11", "账号5h窗口上下界", "PR8698",
    "主动读取更新5h窗口 end 时 SHALL 同步 start=end-5h。",
    "实际 modernc SQLite/Ent overlay 更新 end 后 start 仍在旧窗口。",
    "backend/internal/repository/account_repo.go: UpdateSessionWindowEnd",
    "正常滚动、毫秒格式读入、outbox传播和新窗口费用查询。",
    "账号窗口修复不改变用户平台固定 quota 窗口、软删除或重置策略。",
    "本地 Ent更新和 outbox 存在；无 schema 前置。")
add("F12", "首输出临时文件不可用", "PR8808",
    "首输出暂存无法创建临时文件时 SHALL 在同一容量上限内回退内存。",
    "SpoolUnavailable overlay 返回错误；stage上限拒绝对照通过。",
    "backend/internal/service/openai_first_output_timeout.go: openAIFirstOutputStage.prepareWrite",
    "无权限/目录不可用、commit/drop/取消、上限拒绝及清理。",
    "保留8MiB stage上限、无具名明文残留和预算释放。",
    "既有 unlink 失败回退已存在；创建失败缺同样处理。")
add("F13", "Claude 默认测试模型维护", "PR8757",
    "内置 Claude 默认测试模型 SHALL 与冻结 tag 的 claude-sonnet-4-6 一致。",
    "静态常量差异；基线相关测试通过，未调用真实 Anthropic 核实退役公告。",
    "backend/internal/pkg/claude/constants.go: DefaultTestModel",
    "默认账号测试使用已有4-6；显式模型和其它平台默认值不被覆盖。",
    "属于默认值维护，不新增模型目录/定价，也不宣称独立市场验证。",
    "本地4-6已有兼容和价格消费者；仅改常量及相关期望。")
add("F14", "Grok 媒体扁平错误", "PR8669",
    "Grok 既有媒体错误 SHALL 保留 xAI 扁平 error/code，保持脱敏与可配置错误规则。",
    "GrokFlatMediaMessage overlay 丢失字符串；嵌套/未知对象对照通过。",
    "backend/internal/service/grok_upstream_errors.go、grok_media.go、error_passthrough_runtime.go",
    "嵌套优先、扁平字符串/数字code、内容策略、自定义规则和未知对象。",
    "仅 Grok 扩展提取器；不改变其它平台的模型不存在/停号判断。",
    "源两个窄提取器和带提取器的规则入口需新增；既有媒体路由已存在。")
add("F15", "Gemini 非数组 items", "PR8759",
    "既有 Gemini schema 清理 SHALL 删除非 array 的 items，并保持合法数组 schema。",
    "SchemaNonArrayItems overlay 失败；tuple/default及合法数组对照已经通过。",
    "backend/internal/pkg/antigravity/schema_cleaner.go: cleanJSONSchemaRecursive",
    "嵌套对象/数组、非法items、工具schema及实例 default/enum 的保护。",
    "不把本地已经处理的 tuple 情形描述成新缺陷，不引入新 Gemini 模型。",
    "清理器完整存在；窄类型条件修复。")
add("F16", "白名单精确模型输入", "PR8736",
    "模型白名单添加操作 SHALL 拒绝含星号的条目并保留输入草稿。",
    "临时前端4个来源新增测试失败；原11个控件测试通过。",
    "frontend/src/components/account/ModelWhitelistSelector.vue；zh/en admin/accounts 文案",
    "点击/Enter、首尾空白、纠正后添加、重复/映射冲突提示。",
    "不引入 glob 或透传白名单调度新开关。",
    "共享控件和提示API已有；来源测试需按本地位置插入。")
add("F17", "渠道 Bedrock 兼容开关类型", "PR8840",
    "渠道编辑 SHALL 用 boolean 保存/清除 bedrock_cc_compat，并保留其它 features。",
    "临时前端5个来源测试失败，实际保存的是 {anthropic:true/false}。",
    "frontend/src/views/admin/ChannelsView.vue: formToAPI",
    "创建/编辑/重开/显式false/移除Anthropic区段及其它features。",
    "保留本地 run_mode/菜单限制，不新增平台或商业配置。",
    "后端读 boolean、本地前端写对象；现有开关消费者已核实。")
add("F18", "调度快照端点能力", "PR8837",
    "调度快照 SHALL 保留已有 openai_capabilities 声明并据此过滤既有端点。",
    "repository overlay 丢键；实际 Account 能力判断由禁用 embeddings 变成允许。",
    "backend/internal/repository/scheduler_cache.go: filterSchedulerCredentials；既有 Account/调度器能力消费者",
    "chat/embeddings/Responses生图意图、未配置、空配置、hydrate及WS读取。",
    "来源标题提及 Seedance，但本地现有能力消费者独立适用；不引入 Seedance。",
    "键读取/普通端点过滤存在；只需补投影白名单。")
add("F19", "设置注入 HTML 与 CSP nonce", "PR8775",
    "设置注入 HTML SHALL 禁止共享/浏览器缓存，携带请求 nonce 时不得按旧 ETag 返回304。",
    "embed overlay 复现nonce仍304、缺no-store；无nonce304对照通过。",
    "backend/internal/web/embed_on.go: serveIndexHTML；新增窄HTML头 helper",
    "两请求不同nonce、设置变更/回退、无nonce条件请求、静态hash资源。",
    "不承诺所有浏览器 bfcache 行为；仅约定响应头和 nonce/HTML 一致性。",
    "现有HTML cache及注入存在；18行头helper可独立采用。")
add("F20", "Chat 客户端 metadata.user_id", "PR8786",
    "API Key 的 Chat转Anthropic SHALL 补回已识别格式的客户端 metadata.user_id。",
    "静态：DTO转换丢metadata，两个转发路径没有补回；尚无专用转发探针。",
    "backend/internal/service/gateway_forward_as_chat_completions.go、openai_gateway_chat_completions_anthropic_native.go；窄metadata helper",
    "合法JSON/legacy会话ID、空值/错误类型、两路径和OAuth负对照。",
    "只保留识别到的user_id字符串原字节，OAuth仍使用账号拟态身份。",
    "ParseMetadataUserID存在；源helper32行，不需账号命名空间基座。")
add("F21", "完成转发后的粘性绑定", "PR8726",
    "完成转发后的粘性写入 SHALL 不随客户端取消，且受5秒截止限制。",
    "静态：handler用请求ctx绑定；service原样交给cache，无内部脱离取消。",
    "backend/internal/handler/gateway_handler.go: Messages 完成后的绑定点",
    "完成后断开仍写入、超时返回、换号不覆盖原绑定和正常请求。",
    "保留本地 profit admission/既有绑定条件；只脱离这一收尾写入。",
    "WithoutCancel/WithTimeout和既有绑定helper均已有；无需新调度架构。")

add("S01", "5m/1h 缓存明细与互斥计费", "PR8174",
    "既有转换、读流及结算 SHALL 保留5m/1h明细和聚合presence，差额按5m计价，输入桶互斥。",
    "DTO往返/presence及缓存差额overlay失败；完整明细/显式0对照通过。",
    "backend/internal/pkg/apicompat/types.go及三协议桥；backend/internal/service/*native*、gateway_*usage*、billing_service.go、account_stats_pricing.go",
    "HTTP/普通与透传SSE/缓冲/WS、缓存字段单独更新、矛盾明细、1h费用、SQLite结算+去重。",
    "只用既有5m/1h列；不导入30m/12h字段、迁移或未选价格策略；自定义0/Fast保留。",
    "5m/1h UsageTokens/SQLite列已有；OpenAIUsage及Anthropic DTO/流合并需手工贯通。")
add("S02", "Claude 并发 token 轮换恢复", "PR8737",
    "Claude OAuth 的旧token401 SHALL 仅在数据库证明有新有效token时同号重试一次。",
    "静态追踪三转发路径；上一轮原子refresh仍存在，未跑新401并发传输探针。",
    "backend/internal/service/gateway_forward.go、gateway_forward_as_chat_completions.go、gateway_forward_as_responses.go",
    "barrier控制旧token在途/新token落库、cache旧值、一次重试、同token/错误类型/取消/预算。",
    "比对实际发送token；不在401路径刷新/持久化凭据，不绕过重试上限或正常401策略。",
    "仓储GetByID/token cache已有；31行恢复helper及三路径重发适配。")
add("S03", "SQLite 分组倍率/RPM 原子同步", "PR8659",
    "单次分组倍率或RPM同步 SHALL 在一个事务内清空/删除/写入，失败保留旧配置。",
    "两份真实modernc SQLite trigger故障探针复现旧行丢失；零倍率成功对照通过。",
    "backend/internal/repository/user_group_rate_repo.go: SyncGroupRateMultipliers/SyncGroupRPMOverrides",
    "事务失败/取消回滚、重复用户last-wins、另一列保留、bind-safe大批次、提交后cache。",
    "不得复制PG ANY/ALL/unnest/强制类型转换；不更改历史迁移或用户固定quota。",
    "本地已有SQLite SQL；增加事务与去重并保持本地SQL构造，无schema前置。")
add("S04", "首内容前换号与 keepalive 错误", "PR8811 PR8813 PR8724",
    "首内容前仅keepalive输出 SHALL 不阻断安全换号；已提交SSE后的错误必须为SSE事件。",
    "SSEErrorAfterPing overlay仍写裸JSON；源码确认Messages前奏立即写出和keepalive被stage扣住。",
    "backend/internal/service/openai_gateway_messages.go、openai_gateway_response_handling.go、gateway_upstream_response.go；backend/internal/handler/openai_gateway_handler.go",
    "created/item-added后失败→成功账号、keepalive前后耗尽、错误规则、取消/终态/槽位释放。",
    "前奏需沿用8MiB有界暂存，不能原样引入无界strings.Builder；私有账号头不得提前泄露。",
    "本地keepalive调整大小、首输出预算/规则/暂存已有；hold与耗尽规则需要适配。")
add("S05", "空闲超时与一次换号预算", "PR8577 PR8920",
    "OpenAI 首输出前空闲超时 SHALL 进入最多一次换号预算，计时按最后上游读取截止。",
    "两PR最终代码联合静态审查；未执行新超时/两账号HTTP探针。",
    "backend/internal/service/openai_gateway_response_handling.go: interval计时/timeout；既有handler首输出预算",
    "持续数据/注释、无内容超时、keepalive已写、两号都超时、已有内容与Grok负对照。",
    "保留配置的idle默认及既有模型冷却；不要把惩罚扩散到整个池。",
    "SafeToFailoverAfterWrite及一次预算已有；不可只取#8577漏#8920。")
add("S06", "Codex additional_tools 声明 ID", "PR8754",
    "Codex additional_tools 的声明ID SHALL 与引用一起保留，仅规范已识别UUID形态。",
    "AdditionalToolsDeclarationID overlay 实际id缺失。",
    "backend/internal/service/openai_codex_transform.go: filterCodexInputWithOptions；现有HTTP/透传/WS兼容调用点",
    "UUID声明/引用、opaque ID、store=false、完整工具回放、普通callID/UTF8。",
    "不剥离声明ID、不改普通工具身份；不导入完整账号命名空间或ojson基座。",
    "decoded过滤器已有；raw局部输入重建helper/实际调用点需补，不覆盖整来源文件。")
add("S07", "Grok 服务端搜索调用", "PR8834",
    "Grok xs_/as_ 服务端 custom 搜索 SHALL 不下发为客户端调用，并计入原生搜索次数。",
    "GrokServerSearchCount overlay 实际0期望2；普通custom/web_search对照通过。",
    "backend/internal/pkg/apicompat/responses_to_anthropic.go、responses_to_chatcompletions.go；backend/internal/service/grok_search_count.go及流/缓冲消费者",
    "Grok Chat/Messages/Responses、终态与delta回退、搜索去重、普通custom工具和非Grok前缀。",
    "仅Grok启用prefix规则；不引入通用工具桥或新搜索平台。",
    "源24行predicate可复用；需贯通state/accumulator/provider标记，不只补计数。")
add("S08", "原文模型和真实端点观测", "PR8727 PR8728 PR8745 PR8741",
    "原生Anthropic和Chat/Responses桥 SHALL 把原文模型及冲突带到用量边界，并记录真实/messages端点。",
    "ChatFallbackResponseModel overlay丢模型；现有observer/response_model计费消费者已定位。",
    "backend/internal/service/openai_gateway_*anthropic_native.go、*chat_fallback.go、gateway_forward_as_*.go；既有upstream_response_model.go",
    "流/非流/中断、多模型冲突、一次attempt隔离、端点错误日志、SQLite计费回退。",
    "先观察原文再还原公开名；保留本地价格来源准入及已解析Fast tier，不为模型修复重做tier基座。",
    "本地observer已有约束；源部分ServiceTier观察接口缺失，按本地结果字段适配。")
add("S09", "Anthropic 兼容入口 effort 策略", "PR8739",
    "Anthropic组的Chat/Responses兼容入口 SHALL 执行现有group effort策略并一致返回错误。",
    "静态：通用OpenAI入口有policy，两个Anthropic兼容handler没有调用。",
    "backend/internal/handler/gateway_handler_chat_completions.go、gateway_handler_responses.go",
    "deny/map/ceiling、显式非法值、禁用/省略、转换后最终effort和费用、count_tokens邻近。",
    "保留上一轮合成none/最终校验和GLM通用兼容；不复制缺失原生平台闸门。",
    "applyOpenAIReasoningEffortPolicyForRequest已有；本地handler次序需手工适配。")
add("S10", "Bedrock 错误语义与部分用量", "PR8675",
    "Bedrock SHALL 区分封号/缺用例/请求过滤/模型429，并在中断时返回已观测用量。",
    "静态核对既有Bedrock路由、部分流用量helper和模型限流仓储；未跑新完整错误探针。",
    "backend/internal/service/gateway_bedrock.go、gateway_upstream_response.go、ratelimit_service.go；窄ratelimit_bedrock helper；Messages调度模型ctx",
    "精确400/404、反射字段负对照、429多模型、渠道映射键、5xx过滤、截断SQLite结算。",
    "请求内容错误不换号；只模型冷却，沿用现有显式规则及SQLite单键更新。",
    "110行语义helper及调度键ctx需适配；不导入新Bedrock平台/模型。")
add("S11", "Gemini 并行与分片工具", "PR8828",
    "Gemini流 SHALL 按调用身份而非仅名称划分工具块，保留字符串分片和终态快照。",
    "静态：两个现有reader只以name判断边界；65行来源helper已量测。",
    "backend/internal/service/gemini_messages_compat_service.go、gemini_chat_completions_compat_service.go",
    "同名/同参并行、不同/相同ID、完整首片字符串、重传/快照、无ID歧义及结束事件。",
    "无ID的字符串快照与新相同调用不可区分时采用来源保守策略，不猜造调用。",
    "现有delta累计器存在；两reader添加统一boundary，不需要新模型。")
add("S12", "Gemini Chat 结构化响应格式", "PR8662",
    "Gemini Chat兼容 SHALL 传递已声明json_object/json_schema格式，并保留schema字面量。",
    "静态：当前Chat→Anthropic转换后丢response_format；真实消费者存在。",
    "backend/internal/service/gemini_chat_completions_compat_service.go；窄gemini_compat_response_format helper",
    "两个既有认证入口、属性顺序/大数、工具/图片互斥、文本/缺省/无效格式。",
    "采用responseJsonSchema，不扩展Gemini新模型或cache CRUD；保留资源限制。",
    "源helper50行；CLI wrapper需原文嵌入而非整请求map重编码。")
add("S13", "Antigravity 兼容思考/等待/空流", "PR8673 PR8826",
    "既有Antigravity兼容 SHALL 对齐思考配置、按响应头到达锚定等待，空流仅释放对应粘性会话。",
    "静态核对现有adapter/keepalive；专用时间锚和sticky释放尚未执行。",
    "backend/internal/service/antigravity_gateway_compat*.go、antigravity_gemini_thinking_variant.go；两兼容handler错误收尾",
    "slow headers/有效thought/连续ping/empty timeout、无内容已提交、绑定已改变、取消/用量。",
    "保留既有timeout默认，不导入Carbon等模型和完整新thinking策略；绑定检查需防并发误删。",
    "adapter和thinking意图parser已有；源16行effort helper可摘，缺配置helper需窄补。")
add("S14", "Gemini 原生稳定前缀会话", "PR8822",
    "Gemini原生会话回退 SHALL 按API Key、模型、tools/system和首用户parts生成稳定哈希。",
    "静态：本地仍通用ParseGatewayRequest回退；完整PR最后恢复ForceCacheBilling。",
    "backend/internal/handler/gemini_v1beta_handler.go；backend/internal/service/gemini_session.go",
    "首用户/省略role、后续tool回合、模型/Key隔离、空parts、显式客户端ID优先。",
    "不按发布摘要删除ForceCacheBilling；两个微提交必须以最终PR状态解释。",
    "内容种子归一和xxhash已有相邻消费者；缺helper需适配，不能套来源未有的完整raw基座。")
add("S15", "冷却账号快照成员", "PR8729",
    "快照 SHALL 保留配置启用的完整账号池，并在每次读时判断临时冷却/expiry。",
    "静态：本地loadAccountsFromDB仍用ListSchedulable*；完整候选仓储接口已存在。",
    "backend/internal/service/scheduler_snapshot_service.go: loadAccountsFromDB/ListSchedulableAccounts",
    "快照构建时冷却→自然到期无outbox、真正禁用、simple/混排、取消和未就绪回源。",
    "不随行#8833的20s stale grace或合并回源；保留revision/retire/单节点缓存契约。",
    "ListModelAvailabilityCandidates已有；差异是缓存成员与读取过滤。")
add("S16", "OpenAI 用量观测新鲜度", "PR8747",
    "额度查询 SHALL 区分真实上游观测、陈旧读数和本地未知统计，刷新失败不得伪造新时间。",
    "静态：当前读操作生成UpdatedAt；现有用量/监控消费者都有可复用接口。",
    "backend/internal/service/account_usage_service.go、channel_monitor_quota_fetcher.go；frontend UsageProgressBar/AccountUsageCell/types",
    "主动失败/无quota/显式0、legacy无时间、被动缓存、unknown窗口、UI不误停账号。",
    "不引入新quota-summary或恢复监控v2整页；缺失Spark专用前置不作为依赖导入。",
    "源8行观测predicate及unknown字段需贯通已有UI；保留上一轮已实现的刷新留行。")
add("S17", "实际扣除与账号成本统计", "PR8806",
    "按account过滤的模型/端点统计 SHALL 分列actual_cost和account_cost，Gemini账号配额使用账号成本。",
    "静态核实本地条件覆盖actualCost；SQLite查询/字段和Gemini聚合均已有。",
    "backend/internal/repository/usage_log_repo_stats.go、usage_log_repo_trend.go；gemini_quota.go、usage_handler.go；EndpointDistributionChart/types/user UsageView",
    "合成不同标准/实际/账号价、所有筛选/端点来源、管理员/用户redaction与Gemini批量一致。",
    "只增DTO字段，无新DB列；user API/UI不能泄露账号成本；保持model_source等维度。",
    "源PG ARRAY查询不能覆盖本地SQLite SQL；需同时改查询、consumer和redaction。")
add("S18", "价格启动不阻塞远端哈希", "PR8798",
    "已加载本地有效价格后 SHALL 不阻塞启动等待远端hash，并保持可停止的后台同步。",
    "静态：Initialize同步fetchRemoteHash；本地正文实际hash修复已存在。",
    "backend/internal/service/pricing_service.go: Initialize/checkAndUpdatePricing/Stop",
    "慢/失败hash不挡启动，下载hash一致、并发同步、Stop/取消、fallback/override。",
    "不能只加无管理goroutine；不改变已有目录实际正文hash和零覆盖优先级。",
    "现有价格生命周期和远端client存在；需绑定worker所有权/截止/停止。")
add("S19", "上游 TCP 死连接检测", "PR8800",
    "既有标准上游dialer SHALL 使用15s idle/5s interval/3 probes，Linux user-timeout设置失败仍拨号。",
    "静态socket实现审查；未做丢包/死peer真实计时实验。",
    "backend/internal/repository/http_upstream.go: newUpstreamDialer；平台专用keepalive helper",
    "Linux实际socket选项、错误降级、非Linux编译、proxy/TLS配置和取消。",
    "30秒是来源参数推算，不是本次实测断线耗时；应用层120s idle保留。",
    "Go1.27.3支持该标准API；不依赖Go1.27.4或新x/sys版本。")
add("S20", "OpenAI 工具 schema 正则适配", "PR8772",
    "现有OpenAI工具schema SHALL 窄处理来源不支持的regexp，并保留实例值与仍有效的schema约束。",
    "静态：本地只有null-type/required清理，没有来源完整regexp parser；未向远端发送新样例。",
    "backend/internal/service/openai_responses_tool_schema.go；HTTP/透传/WS实际schema规范化点",
    "pattern/patternProperties、不支持转义/组、嵌套编辑、实例default/enum、有效Python形态及非OpenAI。",
    "不能把来源约899行parser整体覆盖本地窄sanitizer；需测实际适配规模和编辑重叠。",
    "基础工具消费者在此；新正则能力不能凭CLEAN判编译，jsonschema-go0.4.2已在本地图中。")
add("S21", "既有流与请求热点窄优化", "PR8653 PR8667 PR8903 PR8904",
    "选定热点优化 SHALL 保持原字节、工具顺序、错误/用量语义和有界生命周期。",
    "Ollama200消息PoC语义对照通过，7.92ms→0.34ms、12.56MB→0.26MB分配；来源SSE escaped-usage predicate失败。",
    "backend/internal/service/gateway_tool_rewrite.go、gateway_upstream_response.go、gateway_claude_oauth_body.go、gateway_billing_header.go、openai_gateway_chat_completions_raw.go、openai_gateway_ollama_cloud_cc_reasoning.go",
    "每个采用hunk做单P成本/字节对照，escaped type/model/usage、read前flush/EOF/断开、cache TTL次序。",
    "补转义回退；不导入DeepSeek缺失placeholder消费者、无管理的懒日志全body保活、#8816账务lane或完整COW。",
    "raw view已存在；字符串reverse/helper88行/flush-reader16行可窄适配，完整根文件不能覆盖。")
add("S23", "Go工具链与时钟测试维护", "PR8767 PR8743",
    "Go维护 SHALL 对齐1.27.4声明/CI/工具，并让已有峰谷价格测试使用固定PricingAt。",
    "只读Go1.27.4 embed构建及config/apicompat/claude三包通过；首次下载网络超时另存。",
    "backend/go.mod、本地CI/Docker/开发版本断言；两个现有*_record_usage_test.go",
    "全量unit/lint/race/embed、同版本lint重建、隔离Ent/Wire兼容及固定23:59边界。",
    "不升级库版本、上游PG依赖、fork VERSION或历史记录；codegen/CI尚不是本轮已验收项。",
    "范围内go.sum/前端依赖无变；jsonschema-go只是同版本direct标记，非新依赖前置。")

add("S24", "Composite 公开模型名还原", "PR8796",
    "Composite 响应 SHALL 还原客户端公开模型名，同时保留原文上游模型用于审计/计费。",
    "静态核对本地公开/已解析模型ctx；当前Anthropic响应仍使用转发模型，未跑新别名HTTP探针。",
    "backend/internal/service/composite_platform.go、gateway_anthropic_passthrough.go、gateway_upstream_response.go、openai_gateway_messages*.go",
    "原生/拟态/Chat/Responses的流与非流、仅message_start.model、模型冲突及SQLite费用。",
    "先观测后改名；不得替换正文/工具参数中的同名字符串，不影响非Composite响应。",
    "本地RequestedPublicModel/ResolvedUpstreamModel上下文已有；源窄helper需适配。")
for r in review:
    if r["id"] == "S24":
        r["id"] = "S09"
    elif r["tier"] == 2 and 9 <= int(r["id"][1:]) <= 21:
        r["id"] = f"S{int(r['id'][1:])+1:02}"

deferred = {
    "6dd5a045b": (4, "上游VERSION同步；fork保持自己的1.1.x版本。"),
    "530c5e3a4": (4, "赞助商/品牌文档，无个人网关移植内容；已在上一轮main观察，仍完整记账。"),
    "9cd1a1eb2": (4, "赞助商/品牌文档，不同步fork产品说明。"),
    "15871f904": (4, "上游VERSION同步，不改变fork版本。"),
    "PR8644": (3, "Fable limits[]周额度；本地有被动读数分支但未选Fable目录/价格，实际使用时连同已有消费者重评。"),
    "PR8649": (4, "支付宝退款request-id；支付功能个人用途排除，不能因修复小而采用。"),
    "PR8651": (4, "语义已具备：a25ce7bed按groupID隔离RPM，跨组/本组overlay对照通过；保留legacy groupID=0适配。"),
    "PR8664": (4, "N/A：openai_codex_account_identity.go及该namespace改写本地不存在，不引入整套改写来修父缓存key。"),
    "PR8678": (3, "concurrency.ping_interval=0是新控制能力；SSE错误修复不依赖关闭ping，默认不改。"),
    "PR8693": (4, "PG分组rollup新一代build/锁/迁移251；本地两核心helper文件缺失，SQLite现有汇总不走该维护链。"),
    "PR8735": (4, "N/A：完整raw OAuth输入metadata清理消费者本地不存在；parseRawJSONView存在不代表该缺陷存在。"),
    "PR8738": (4, "国内原生厂商quota cell/查询平台缺失，不为前端刷新修复导入平台。"),
    "PR8740": (4, "语义已具备：本地Gemini分支已有loading && !usageInfo；上一轮共享刷新留行维护保留。"),
    "PR8763": (4, "原生Kimi月度quota parser平台缺失；普通OpenAI API Key转发不消费此模块。"),
    "PR8779": (3, "Fable7d_oi恢复；现有model冷却分支可被自定义映射到达，但未确认使用，不能自动引入Fable新能力。"),
    "PR8790": (4, "只扩大到Zhipu/CommandCode/OpenCodeGo/MiniMax原生平台；现有Anthropic/OpenAI模式已生效，缺平台分支不导入。"),
    "PR8825": (4, "N/A/对照通过：本地没有强制DeepSeek峰谷覆盖策略；现有override_file的0/正数/缓存读价均保留，不补源强制策略。"),
    "PR8746": (3, "Gemini4 Hydrogen兜底价是新模型覆盖；通用映射可能到达，按实际使用选择并核价。"),
    "PR8827": (3, "Vertex us/eu多区域和1.1倍率是新端点/计价策略，需按实际账户与区域验证。"),
    "PR8816": (3, "按用户lane批量账务与广泛零拷贝；源新增544行PGbatch repo+311行combiner，须SQLite事务/dedup/取消/1C1G独立测量。"),
    "PR8833": (3, "大bucket合并加载+200ms last_used缓存+20s stale grace为新性能策略；S15快照正确性不需要它，延迟/内存先量测。"),
    "PR8661": (3, "Gemini4 Carbon及完整思考策略新能力；现有Gemini修复不依赖整批目录/价格。"),
    "PR8748": (3, "透传模型白名单过滤是新开关/schema外配置；F16输入校验不自动激活此策略。"),
    "PR7540": (3, "Qwen3 embedding/reranker新兜底目录；包含CNY折算假设，按实际模型/站点核价后选择。"),
    "PR8906": (3, "GPT-6.2 Sol/Astra、272k计价、Pro套餐/Ultrafast促销时钟为新能力，旧Sol倍率修复独立。"),
    "PR8598": (4, "N/A：billing_inflight_reservation_test.go及其在途预占基座本地不存在，不为heap测试引入新能力。"),
    "PR8914": (3, "新增渠道Ultra Fast倍率及PG252迁移，是新配置能力；现有Fast倍率不受本评估扩展。"),
}

by_candidate = {}
for r in review:
    for name in r["candidates"]:
        assert name not in by_candidate, name
        by_candidate[name] = dict(tier=r["tier"], cluster=r["id"], reason=r["title"])
for name, (tier, reason) in deferred.items():
    assert name not in by_candidate, name
    by_candidate[name] = dict(tier=tier, cluster="", reason=reason)
cs = list(csv.DictReader((OUT / "candidates.tsv").open(), delimiter="\t"))
assert set(by_candidate) == {c["candidate"] for c in cs}
assert sum(r["tier"] == 1 for r in review) == 21
assert sum(r["tier"] == 2 for r in review) == 23
(OUT / "review.json").write_text(json.dumps(sorted(review, key=lambda r:(r["tier"],int(r["id"][1:]))), ensure_ascii=False, indent=2)+"\n")
(OUT / "classification.json").write_text(json.dumps(by_candidate, ensure_ascii=False, indent=2)+"\n")
print(json.dumps({"first_tier_clusters":21,"second_tier_clusters":23,
    "selected_prs":sum(bool(v["cluster"]) for v in by_candidate.values()),
    "candidates":len(by_candidate)},ensure_ascii=False))
