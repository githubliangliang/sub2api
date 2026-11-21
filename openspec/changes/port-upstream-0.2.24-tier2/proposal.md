# Why

上游v0.2.24涉及现有网关解析、Claude/Codex/Grok协议与账号数据正确性，需要按SQLite个人部署进行选择性维护。

# What Changes

- S01 所有现有入口的重复字段拒绝与读取一致性：所有现有HTTP网关请求与每轮WS载荷 SHALL 在模型改写、调度和计费之前拒绝语义重复字段，并让内容识别、body重读与协议错误格式一致。
- S02 OpenAI 不可逆 OAuth 轮换的 SQLite 保存：已完成的OpenAI轮换 SHALL 在有界脱离取消的上下文保存；CAS失败须读取当前持久化凭据且保留并发重新授权。
- S03 缓冲 Responses 的 hosted 输出重建：缓冲兼容响应 SHALL 从最终有效终态及已完成item恢复完整hosted内容，输出顺序、工具身份和流式语义一致。
- S04 Grok 成功搜索次数结算：Grok搜索结算 SHALL 在上游提供成功次数时以其为准，缺省时沿用已有事件计数与去重。
- S05 OAuth 外族模型与 Messages 映射后调度：固定Codex上游的OAuth类账号 SHALL 跳过明确外族模型，且调度判断使用渠道映射后的实际模型。
- S06 批量透传编辑独立修改模型过滤：批量编辑 SHALL 只提交用户明确选择修改的透传模型过滤字段。
- S07 Claude/Gemini 用量快照刷新：Claude/Gemini快照变化 SHALL 使现有用量缓存失效并刷新，异步旧结果不得覆盖新快照。
- S08 Codex 原生音频输入声明：已有原生Responses音频输入能力 SHALL 在模型声明中正确保留，聚合目录只声明全部可调度候选支持的能力。
- S09 Anthropic metadata 长上下文目录：现有Anthropic目录 SHALL 按可服务账号的有效metadata展示长上下文变体，并正确识别默认官方registry host。
- S10 Composite 改写后的 Claude thinking：Composite将请求改写到既有Claude5.5时 SHALL 同步适配该型号接受的thinking模式并保留其它参数。
- S11 后台新建用户的默认平台配额：后台创建普通用户 SHALL 按当前默认平台配额初始化，显式0与未配置无限额保持区别。
- S12 异步用量与审计的请求体保活：离开请求生命周期的用量和审计数据 SHALL 拥有必要标量的独立存储，且不因字段地址或闭包滞留整个请求体。
- S13 已有热点的有界窄优化：采用的热点优化 SHALL 保持字节、策略和所有权语义，并通过单P前后成本与转义必要条件验证。
- S14 SQLite 模型平均速率按耗时加权：模型统计平均Tokens/秒 SHALL 使用有效请求output_tokens总和除以duration_ms总和。
- S15 Antigravity 裸 429 范围及额度端点：同一请求的无明确信号裸429 SHALL 最多冷却一个账号；既有额度查询 SHALL 使用该账号实际转发端点。

# Impact

不需要新增数据库列。保持VERSION1.1.26、Go1.27.4、SQLite230/auth21和依赖。新增共享入口检查需覆盖所有现有HTTP别名和WS轮帧；OAuth CAS适配SQLite，默认配额消费者实施时重新生成Wire。真实影响面由所选行为决定，不能以文本CLEAN替代编译。

# Non-goals

不导入支付或商业流程、缺失TypeSafe/DeepSeek native平台、新搜索供应商/来源块/ToolSearch能力、5图扩展、费用估算/刷新提示/复制ID、完整group allowlist、QuotaSummary、shadow、PGbatch/lane/COW/stale策略。发布、push、归档和业务实例操作另行处理。
