# Why

v0.2.20修复了本地既有Claude Code、Codex、Grok及共享网关路径中的数据丢失、错误分类和资源边界问题。本批仅包含已核实本地消费者的11簇，详细缺口与复现见评估报告。

# What Changes

- S01 Codex Responses非流式与instructions回显：网关 SHALL 保留客户端原stream选择，将非流式SSE的completed/incomplete/cancelled转换成单JSON；缺Content-Type的OAuth SSE可识别。允许空instructions进入既有规范化，并将网关改写的instructions回显还原为客户端值。
- S02 Grok 委派消息最小适配：既有Grok Responses入口 SHALL 将agent_message的input_text及明文encrypted_content按原顺序转为user消息，空消息跳过，不尝试解密。
- S03 已使用 Terra 的目录与计费后续：已接入Terra SHALL 使用冻结tag中的Codex descriptor作为受限fallback，Codex默认low及500000窗口与max_context_window=872000区分，API兼容默认medium/none仍合法，minimal明确400。OAuth/setup-token长上下文豁免按最终发送模型决定，API Key保留272k严格边界与Fast结算。
- S04 请求体预算、读写超时与错误分类：所有既有网关入口 SHALL 共享实例级请求体内存预算，读取/解压/修补/异步移交/failover释放按实际所有权计数；超预算503+Retry-After，超时408、断开499、截断/解码失败保持安全分类。超时按单次读写停滞计，不限制正常SSE总时长。
- S05 上游流、drain与会话缓存有界：既有上游流 SHALL 限制SSE队列字节/单行/累计事件，非流式读取有界；客户端离开后的usage drain有总时限，缓存按数量和字节逐出并释放后台任务。
- S06 Antigravity 系统身份保留任务：Antigravity路径 SHALL 仅规范化开头的已知客户端身份和attribution，保留附随任务及中间引用；原生Anthropic不受影响。
- S07 图片 INTERNAL500 不牵连整账号：现有Antigravity图片模型重试耗尽INTERNAL500 SHALL 只冷却最终映射模型，并同步缓存；文本模型仍用既有账号惩罚。
- S08 Gemini 工具媒体原生承载：Gemini3及以后已使用模型的工具媒体 SHALL 保留在functionResponse.parts；旧模型按已有分离图片回退，关联工具名/顺序保持。
- S09 使用记录导出限流恢复：现有用户导出 SHALL 每页至多1000条，在429时有界等待并重试同一页；重试耗尽或取消不能下载部分结果。
- S10 Antigravity 测试提示词贯通：已有账号测试 SHALL 将管理员输入提示词传至Antigravity Claude/Gemini请求，并返回真实回复；空值沿用默认提示词。
- S11 RPM空覆盖缓存与既有JSON预检：鉴权快照 SHALL 区分已加载的nil RPM覆盖与加载失败，并按group隔离；既有thinking/委派预检仅在可证明无修改时跳过解析，不能改变输出语义。

# Impact

按行为手工适配，不合并upstream/main。无拟议新SQLite迁移；保持VERSION1.1.23、Go1.27.3及现有依赖。新增helper与消费者同批，具体改动范围和预算取值在实施时记录。

# Non-goals

不引入管理MCP/受限管理员密钥、注册黑名单、原生图片新链路、新平台/新模型、支付购买充值退款或PG迁移249。tag之后的修复另行评估。第三档、第四档不因本提案自动获准。
