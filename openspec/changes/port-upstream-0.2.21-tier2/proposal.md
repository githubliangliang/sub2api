# Why

v0.2.21包含本地既有Claude Code、Codex及共享网关的协议、数据与资源问题。本批按本地消费者逐簇选择，Grok通过共享兼容、计价和会话路径受益。

# What Changes

- S01 原生 WS 后续轮次失败关闭：原生 Responses WS 后续轮次不能安全迁移完整会话时 SHALL 以可重试关闭结束，不能跨账号重放首帧。
- S02 OAuth WS 搜索激活与完整回放：OAuth 池化 WS 首次声明web_search时 SHALL 校验连接初始能力，必要时新建连接并携带完整历史。
- S03 WS 与 SSE 首 token 口径一致：WS 首token统计 SHALL 遵循管理员TTFT模式，并在每轮独立计算。
- S04 显式模型冷却覆盖容量重试：OpenAI 显式model范围冷却规则 SHALL 对 overload/slow_down及流内503生效，且只匹配实际错误字段。
- S05 Anthropic 兼容入口账号错误策略：Anthropic账号经Chat或Responses兼容入口时 SHALL 执行已有自定义错误码及临时不可调度策略。
- S06 合成 effort 的映射与最终校验：Messages转OpenAI SHALL 对显式disabled思考合成的none执行一次映射及上限策略，并在映射后校验最终effort。
- S07 Refresh Token 原子轮换：刷新令牌 SHALL 在校验后原子消费旧token，保证同一token并发只签发一份新令牌；缓存错误必须失败关闭。
- S08 Key 列表独立加载用量与 SQLite 覆盖索引：API Key列表 SHALL 在列表响应到达后展示，用量独立异步加载；统计查询可采用经验证的SQLite覆盖索引。
- S09 账号刷新保留行与已知配额：账号列表和配额刷新 SHALL 保留当前已展示内容，并清晰标识刷新中的不可交互状态。
- S10 Codex Desktop 页面上下文兼容：符合严格签名的Codex页面function_call_output SHALL 转为标记的非信任user文本，避免孤儿工具结果错误。
- S11 既有安全审计 worker 入队唤醒：已有审计任务成功发布后 SHALL 非阻塞唤醒本进程worker，同时保留轮询拾取重试任务。
- S12 Claude Code 内置版本基线：既有Claude Code身份链 SHALL 统一使用2.1.321内置基线，并保留配置与有效环境覆盖规则。

# Impact

基线VERSION1.1.24 / Go1.27.3 / SQLite228 / auth snapshot21。依赖无需随上游升级。S08若实施索引须新增本地229迁移并实际验证；S07改变接口须改测试stub。无Ent字段变更；若Wire接线改变须生成而非手改。

# Non-goals

不引入30m/12h TTL schema、新模型/平台、自适应预算新策略、支付充值购买退款、PG迁移、发布、推送或业务实例部署。第三/四档保持独立，不用本提案扩大授权。
