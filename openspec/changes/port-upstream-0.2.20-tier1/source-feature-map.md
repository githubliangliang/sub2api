# Source Feature Map

| 簇 | 整PR来源 | 纳入行为 | 验收边界 |
|---|---|---|---|
| F01 Claude setup-token 周用量 | PR8553 | Setup-token 查询 SHALL 使用已有被动采样，成功响应缺少5h头及429早退也保留7d/7d_oi；不要求profile权限或主动调用usage API。 | 无5h、有效/过期5h、显式0、无样本、429以及单查/force/passive/batch一致；SQLite Extra与固定quota窗口不回退。 |
| F02 Codex 强制模式停止多余探测 | PR8445 | 系统 SHALL 在force_responses或force_chat_completions模式跳过能力探测；auto及无效模式按现有规则探测。 | 两个强制模式请求数/写Extra数均0；auto/非法/非字符串模式仍探测，凭据覆写和取消正常。 |
| F03 WebSocket 每轮首 token 时间 | PR8520 | 系统 SHALL 为每个WS轮次独立记录首token耗时，无response.id的delta归属当前活跃轮次。 | 至少两轮无ID delta、带ID delta、负时差、重复terminal与取消；保留Fast每轮service_tier/模型冲突和usage隔离。 |
| F04 Codex WS→HTTP 桥接指纹 | PR8464 | 系统 SHALL 在Codex WS HTTP桥接构建上游请求前设置账号级指纹，重试和切换账号时清除旧暂存值。 | 最终HTTP请求header/body同账号稳定、跨账号隔离；无prompt_cache_key/显式客户端key/API Key/Grok负对照；Fast策略不变。 |
| F05 缓冲 Responses reasoning 保真 | PR8563 | Responses转Messages/Chat SHALL 优先保留非空summary，否则保留content.reasoning_text；encrypted_content签名独立保留。 | summary优先、不重复拼接、仅reasoning_text、仅签名、普通文本及两条真实缓冲网关路径；usage/dedup不重复。 |
| F06 Responses 完成消息字段 | PR8429 | 完成响应与流式item SHALL 使用一致message序列化，空content为数组，refusal、annotations及logprobs保持字段语义。 | 空message、output_text、refusal、annotation/logprobs、tool_search/function_call；JSON及SSE终态一致，序号0不被省略。 |
| F07 鉴权取消分类 | PR8471 | API Key查询因当前请求上下文取消而失败时 SHALL 按客户端离开结束；有效请求的真正仓储错误仍500，业务错误保持原分类，日志不含凭据。 | 取消/截止、真实DB失败、格式错误Key、503过载及401错误；不把所有context错误无条件吞成499，不记录Key。 |
| F08 413 透传规则优先级 | PR8437 | 耗尽failover后，已有管理员413透传规则 SHALL 先于内置body-too-large响应生效，账号切换行为保持。 | 有/无匹配规则、自定义状态/文案、skip-monitoring、已开始SSE及Retry-After；不提前终止账号failover。 |
| F09 Antigravity 直连 SSE 响应 | PR8441 | 流转换器 SHALL 按response信封是否存在选择解包形态，直接Gemini响应的文本、工具和用量不能静默丢失。 | 直接/包装响应、空/非法JSON、工具、usage和错误尾帧；普通流与缓冲兼容路径不产生假完成。 |
| F10 Antigravity JSON Schema 参数 | PR8409 | Antigravity转发 SHALL 将responseJsonSchema转换成上游支持的responseSchema，已有responseSchema优先。 | camel/snake命名、双字段优先、非object、undefined清理、原body不变正例；不修改原生Gemini协议。 |
| F11 GPT Image 账号测试参数 | PR8432 | 既有GPT Image API Key账号测试 SHALL 不发送被上游拒绝的response_format，并继续解析默认base64响应。 | 严格模拟上游检查无response_format、模型/prompt/n保留、有效图片解析、上游错误；不新增模型/尺寸路由。 |
| F12 Anthropic兼容 Cloudflare REST URL | PR8420 | 现有Anthropic自定义base URL消费者 SHALL 仅在精确Cloudflare账号REST messages路径省略beta查询，其余上游保持原行为。 | 精确HTTPS host/443/账号路径、相似域名/额外path/query/fragment/rawpath/userinfo负例、Cloudflare native gateway保留beta；三条调用链一致。 |
| F13 Bedrock EventStream 分配前边界 | PR8518 | EventStream解码 SHALL 在分配和切片前验证headers长度、payload长度及总长关系；合法帧继续校验CRC。 | header越界、total<16、header>128KiB、payload>16MiB、truncated/CRC、合法frame；大长度仅喂prelude，证实分配前拒绝。 |

来源SHA见source-baseline。PR8518混合资源修复、优化和新依赖，各簇只取明确列出的子集；没有授权整个PR。
