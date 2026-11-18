# Why

新的v0.2.22–v0.2.23合并内容涉及本地已有Claude Code、Codex、Grok及共享网关消费者。本批按实际存在的路径和最小前置选择，保留SQLite单节点用途。

# What Changes

- F01 Responses 密封调用内容：函数调用的 delta 回退重建 SHALL 保留 encrypted_content。
- F02 对话中途 system 权限：Chat/Anthropic 中途 system 消息转 Responses SHALL 保持 developer 权限。
- F03 JSON Schema 默认 strict：Chat json_schema 未设置 strict 时转 Responses SHALL 显式写 false。
- F04 Codex HTTP 外部上下文：无对应调用的 agent browser_info/goal_objective/untrusted_context 输出 SHALL 转为不可信用户文本。
- F05 Claude per-turn beta：任一已支持的 per-turn/mid-conversation beta SHALL 保留消息级 output_config。
- F06 count_tokens 请求头唯一性：Anthropic count_tokens 透传 SHALL 以原始大小写无关读取/写入避免重复默认头。
- F07 输出后的兼容流账号副作用：OpenAI Chat/Messages 已输出后的流错误 SHALL 执行既有账号/模型错误副作用。
- F08 既有 Sol Ultra Fast 倍率：已支持的 GPT-6.1 Sol Ultra Fast SHALL 使用冻结 tag 的6倍模型倍率。
- F09 登录身份回填空指针：登录邮箱身份回填失败返回 nil 时 SHALL 安全跳过默认设置判断。
- F10 5h 滚动保留有效周用量：Claude 5h 窗口滚动 SHALL 仅清除已过期的周读数。
- F11 账号5h窗口上下界：主动读取更新5h窗口 end 时 SHALL 同步 start=end-5h。
- F12 首输出临时文件不可用：首输出暂存无法创建临时文件时 SHALL 在同一容量上限内回退内存。
- F13 Claude 默认测试模型维护：内置 Claude 默认测试模型 SHALL 与冻结 tag 的 claude-sonnet-4-6 一致。
- F14 Grok 媒体扁平错误：Grok 既有媒体错误 SHALL 保留 xAI 扁平 error/code，保持脱敏与可配置错误规则。
- F15 Gemini 非数组 items：既有 Gemini schema 清理 SHALL 删除非 array 的 items，并保持合法数组 schema。
- F16 白名单精确模型输入：模型白名单添加操作 SHALL 拒绝含星号的条目并保留输入草稿。
- F17 渠道 Bedrock 兼容开关类型：渠道编辑 SHALL 用 boolean 保存/清除 bedrock_cc_compat，并保留其它 features。
- F18 调度快照端点能力：调度快照 SHALL 保留已有 openai_capabilities 声明并据此过滤既有端点。
- F19 设置注入 HTML 与 CSP nonce：设置注入 HTML SHALL 禁止共享/浏览器缓存，携带请求 nonce 时不得按旧 ETag 返回304。
- F20 Chat 客户端 metadata.user_id：API Key 的 Chat转Anthropic SHALL 补回已识别格式的客户端 metadata.user_id。
- F21 完成转发后的粘性绑定：完成转发后的粘性写入 SHALL 不随客户端取消，且受5秒截止限制。

# Impact

第一档共21簇。不新增产品schema；第二档的Go1.27.4维护需单独对齐工具和CI，库版本无需随上游升级。保持64MiB请求体预算、40MiB SSE行、120s读写idle和既有8MiB首输出暂存边界。真实实施起点和结果填verification。

# Non-goals

不引入支付、退款、充值/购买/续费及其依赖/迁移；不导入未选模型/平台、30m/12h桶、强制Ultra Fast、新渠道倍率、账务lane或完整COW。发布、推送、归档和业务实例操作不在本方案。
