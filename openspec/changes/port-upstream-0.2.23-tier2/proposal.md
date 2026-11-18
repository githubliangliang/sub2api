# Why

新的v0.2.22–v0.2.23合并内容涉及本地已有Claude Code、Codex、Grok及共享网关消费者。本批按实际存在的路径和最小前置选择，保留SQLite单节点用途。

# What Changes

- S01 5m/1h 缓存明细与互斥计费：既有转换、读流及结算 SHALL 保留5m/1h明细和聚合presence，差额按5m计价，输入桶互斥。
- S02 Claude 并发 token 轮换恢复：Claude OAuth 的旧token401 SHALL 仅在数据库证明有新有效token时同号重试一次。
- S03 SQLite 分组倍率/RPM 原子同步：单次分组倍率或RPM同步 SHALL 在一个事务内清空/删除/写入，失败保留旧配置。
- S04 首内容前换号与 keepalive 错误：首内容前仅keepalive输出 SHALL 不阻断安全换号；已提交SSE后的错误必须为SSE事件。
- S05 空闲超时与一次换号预算：OpenAI 首输出前空闲超时 SHALL 进入最多一次换号预算，计时按最后上游读取截止。
- S06 Codex additional_tools 声明 ID：Codex additional_tools 的声明ID SHALL 与引用一起保留，仅规范已识别UUID形态。
- S07 Grok 服务端搜索调用：Grok xs_/as_ 服务端 custom 搜索 SHALL 不下发为客户端调用，并计入原生搜索次数。
- S08 原文模型和真实端点观测：原生Anthropic和Chat/Responses桥 SHALL 把原文模型及冲突带到用量边界，并记录真实/messages端点。
- S09 Composite 公开模型名还原：Composite 响应 SHALL 还原客户端公开模型名，同时保留原文上游模型用于审计/计费。
- S10 Anthropic 兼容入口 effort 策略：Anthropic组的Chat/Responses兼容入口 SHALL 执行现有group effort策略并一致返回错误。
- S11 Bedrock 错误语义与部分用量：Bedrock SHALL 区分封号/缺用例/请求过滤/模型429，并在中断时返回已观测用量。
- S12 Gemini 并行与分片工具：Gemini流 SHALL 按调用身份而非仅名称划分工具块，保留字符串分片和终态快照。
- S13 Gemini Chat 结构化响应格式：Gemini Chat兼容 SHALL 传递已声明json_object/json_schema格式，并保留schema字面量。
- S14 Antigravity 兼容思考/等待/空流：既有Antigravity兼容 SHALL 对齐思考配置、按响应头到达锚定等待，空流仅释放对应粘性会话。
- S15 Gemini 原生稳定前缀会话：Gemini原生会话回退 SHALL 按API Key、模型、tools/system和首用户parts生成稳定哈希。
- S16 冷却账号快照成员：快照 SHALL 保留配置启用的完整账号池，并在每次读时判断临时冷却/expiry。
- S17 OpenAI 用量观测新鲜度：额度查询 SHALL 区分真实上游观测、陈旧读数和本地未知统计，刷新失败不得伪造新时间。
- S18 实际扣除与账号成本统计：按account过滤的模型/端点统计 SHALL 分列actual_cost和account_cost，Gemini账号配额使用账号成本。
- S19 价格启动不阻塞远端哈希：已加载本地有效价格后 SHALL 不阻塞启动等待远端hash，并保持可停止的后台同步。
- S20 上游 TCP 死连接检测：既有标准上游dialer SHALL 使用15s idle/5s interval/3 probes，Linux user-timeout设置失败仍拨号。
- S21 OpenAI 工具 schema 正则适配：现有OpenAI工具schema SHALL 窄处理来源不支持的regexp，并保留实例值与仍有效的schema约束。
- S22 既有流与请求热点窄优化：选定热点优化 SHALL 保持原字节、工具顺序、错误/用量语义和有界生命周期。
- S23 Go工具链与时钟测试维护：Go维护 SHALL 对齐1.27.4声明/CI/工具，并让已有峰谷价格测试使用固定PricingAt。

# Impact

第二档共23簇。不新增产品schema；第二档的Go1.27.4维护需单独对齐工具和CI，库版本无需随上游升级。保持64MiB请求体预算、40MiB SSE行、120s读写idle和既有8MiB首输出暂存边界。真实实施起点和结果填verification。

# Non-goals

不引入支付、退款、充值/购买/续费及其依赖/迁移；不导入未选模型/平台、30m/12h桶、强制Ultra Fast、新渠道倍率、账务lane或完整COW。发布、推送、归档和业务实例操作不在本方案。
