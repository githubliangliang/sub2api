# Why

v0.2.20修复了本地既有Claude Code、Codex、Grok及共享网关路径中的数据丢失、错误分类和资源边界问题。本批仅包含已核实本地消费者的13簇，详细缺口与复现见评估报告。

# What Changes

- F01 Claude setup-token 周用量：Setup-token 查询 SHALL 使用已有被动采样，成功响应缺少5h头及429早退也保留7d/7d_oi；不要求profile权限或主动调用usage API。
- F02 Codex 强制模式停止多余探测：系统 SHALL 在force_responses或force_chat_completions模式跳过能力探测；auto及无效模式按现有规则探测。
- F03 WebSocket 每轮首 token 时间：系统 SHALL 为每个WS轮次独立记录首token耗时，无response.id的delta归属当前活跃轮次。
- F04 Codex WS→HTTP 桥接指纹：系统 SHALL 在Codex WS HTTP桥接构建上游请求前设置账号级指纹，重试和切换账号时清除旧暂存值。
- F05 缓冲 Responses reasoning 保真：Responses转Messages/Chat SHALL 优先保留非空summary，否则保留content.reasoning_text；encrypted_content签名独立保留。
- F06 Responses 完成消息字段：完成响应与流式item SHALL 使用一致message序列化，空content为数组，refusal、annotations及logprobs保持字段语义。
- F07 鉴权取消分类：API Key查询因当前请求上下文取消而失败时 SHALL 按客户端离开结束；有效请求的真正仓储错误仍500，业务错误保持原分类，日志不含凭据。
- F08 413 透传规则优先级：耗尽failover后，已有管理员413透传规则 SHALL 先于内置body-too-large响应生效，账号切换行为保持。
- F09 Antigravity 直连 SSE 响应：流转换器 SHALL 按response信封是否存在选择解包形态，直接Gemini响应的文本、工具和用量不能静默丢失。
- F10 Antigravity JSON Schema 参数：Antigravity转发 SHALL 将responseJsonSchema转换成上游支持的responseSchema，已有responseSchema优先。
- F11 GPT Image 账号测试参数：既有GPT Image API Key账号测试 SHALL 不发送被上游拒绝的response_format，并继续解析默认base64响应。
- F12 Anthropic兼容 Cloudflare REST URL：现有Anthropic自定义base URL消费者 SHALL 仅在精确Cloudflare账号REST messages路径省略beta查询，其余上游保持原行为。
- F13 Bedrock EventStream 分配前边界：EventStream解码 SHALL 在分配和切片前验证headers长度、payload长度及总长关系；合法帧继续校验CRC。

# Impact

按行为手工适配，不合并upstream/main。无拟议新SQLite迁移；保持VERSION1.1.23、Go1.27.3及现有依赖。新增helper与消费者同批，具体改动范围和预算取值在实施时记录。

# Non-goals

不引入管理MCP/受限管理员密钥、注册黑名单、原生图片新链路、新平台/新模型、支付购买充值退款或PG迁移249。tag之后的修复另行评估。第三档、第四档不因本提案自动获准。
