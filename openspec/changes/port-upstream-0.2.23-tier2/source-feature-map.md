# 来源行为映射

| 簇 | 来源 | 所需行为 | 评估依据 |
|---|---|---|---|
| S01 5m/1h 缓存明细与互斥计费 | [#8174](https://github.com/Wei-Shaw/sub2api/pull/8174) | 既有转换、读流及结算 SHALL 保留5m/1h明细和聚合presence，差额按5m计价，输入桶互斥。 | DTO往返/presence及缓存差额overlay失败；完整明细/显式0对照通过。 |
| S02 Claude 并发 token 轮换恢复 | [#8737](https://github.com/Wei-Shaw/sub2api/pull/8737) | Claude OAuth 的旧token401 SHALL 仅在数据库证明有新有效token时同号重试一次。 | 静态追踪三转发路径；上一轮原子refresh仍存在，未跑新401并发传输探针。 |
| S03 SQLite 分组倍率/RPM 原子同步 | [#8659](https://github.com/Wei-Shaw/sub2api/pull/8659) | 单次分组倍率或RPM同步 SHALL 在一个事务内清空/删除/写入，失败保留旧配置。 | 两份真实modernc SQLite trigger故障探针复现旧行丢失；零倍率成功对照通过。 |
| S04 首内容前换号与 keepalive 错误 | [#8811](https://github.com/Wei-Shaw/sub2api/pull/8811)、[#8813](https://github.com/Wei-Shaw/sub2api/pull/8813)、[#8724](https://github.com/Wei-Shaw/sub2api/pull/8724) | 首内容前仅keepalive输出 SHALL 不阻断安全换号；已提交SSE后的错误必须为SSE事件。 | SSEErrorAfterPing overlay仍写裸JSON；源码确认Messages前奏立即写出和keepalive被stage扣住。 |
| S05 空闲超时与一次换号预算 | [#8577](https://github.com/Wei-Shaw/sub2api/pull/8577)、[#8920](https://github.com/Wei-Shaw/sub2api/pull/8920) | OpenAI 首输出前空闲超时 SHALL 进入最多一次换号预算，计时按最后上游读取截止。 | 两PR最终代码联合静态审查；未执行新超时/两账号HTTP探针。 |
| S06 Codex additional_tools 声明 ID | [#8754](https://github.com/Wei-Shaw/sub2api/pull/8754) | Codex additional_tools 的声明ID SHALL 与引用一起保留，仅规范已识别UUID形态。 | AdditionalToolsDeclarationID overlay 实际id缺失。 |
| S07 Grok 服务端搜索调用 | [#8834](https://github.com/Wei-Shaw/sub2api/pull/8834) | Grok xs_/as_ 服务端 custom 搜索 SHALL 不下发为客户端调用，并计入原生搜索次数。 | GrokServerSearchCount overlay 实际0期望2；普通custom/web_search对照通过。 |
| S08 原文模型和真实端点观测 | [#8727](https://github.com/Wei-Shaw/sub2api/pull/8727)、[#8728](https://github.com/Wei-Shaw/sub2api/pull/8728)、[#8745](https://github.com/Wei-Shaw/sub2api/pull/8745)、[#8741](https://github.com/Wei-Shaw/sub2api/pull/8741) | 原生Anthropic和Chat/Responses桥 SHALL 把原文模型及冲突带到用量边界，并记录真实/messages端点。 | ChatFallbackResponseModel overlay丢模型；现有observer/response_model计费消费者已定位。 |
| S09 Composite 公开模型名还原 | [#8796](https://github.com/Wei-Shaw/sub2api/pull/8796) | Composite 响应 SHALL 还原客户端公开模型名，同时保留原文上游模型用于审计/计费。 | 静态核对本地公开/已解析模型ctx；当前Anthropic响应仍使用转发模型，未跑新别名HTTP探针。 |
| S10 Anthropic 兼容入口 effort 策略 | [#8739](https://github.com/Wei-Shaw/sub2api/pull/8739) | Anthropic组的Chat/Responses兼容入口 SHALL 执行现有group effort策略并一致返回错误。 | 静态：通用OpenAI入口有policy，两个Anthropic兼容handler没有调用。 |
| S11 Bedrock 错误语义与部分用量 | [#8675](https://github.com/Wei-Shaw/sub2api/pull/8675) | Bedrock SHALL 区分封号/缺用例/请求过滤/模型429，并在中断时返回已观测用量。 | 静态核对既有Bedrock路由、部分流用量helper和模型限流仓储；未跑新完整错误探针。 |
| S12 Gemini 并行与分片工具 | [#8828](https://github.com/Wei-Shaw/sub2api/pull/8828) | Gemini流 SHALL 按调用身份而非仅名称划分工具块，保留字符串分片和终态快照。 | 静态：两个现有reader只以name判断边界；65行来源helper已量测。 |
| S13 Gemini Chat 结构化响应格式 | [#8662](https://github.com/Wei-Shaw/sub2api/pull/8662) | Gemini Chat兼容 SHALL 传递已声明json_object/json_schema格式，并保留schema字面量。 | 静态：当前Chat→Anthropic转换后丢response_format；真实消费者存在。 |
| S14 Antigravity 兼容思考/等待/空流 | [#8673](https://github.com/Wei-Shaw/sub2api/pull/8673)、[#8826](https://github.com/Wei-Shaw/sub2api/pull/8826) | 既有Antigravity兼容 SHALL 对齐思考配置、按响应头到达锚定等待，空流仅释放对应粘性会话。 | 静态核对现有adapter/keepalive；专用时间锚和sticky释放尚未执行。 |
| S15 Gemini 原生稳定前缀会话 | [#8822](https://github.com/Wei-Shaw/sub2api/pull/8822) | Gemini原生会话回退 SHALL 按API Key、模型、tools/system和首用户parts生成稳定哈希。 | 静态：本地仍通用ParseGatewayRequest回退；完整PR最后恢复ForceCacheBilling。 |
| S16 冷却账号快照成员 | [#8729](https://github.com/Wei-Shaw/sub2api/pull/8729) | 快照 SHALL 保留配置启用的完整账号池，并在每次读时判断临时冷却/expiry。 | 静态：本地loadAccountsFromDB仍用ListSchedulable*；完整候选仓储接口已存在。 |
| S17 OpenAI 用量观测新鲜度 | [#8747](https://github.com/Wei-Shaw/sub2api/pull/8747) | 额度查询 SHALL 区分真实上游观测、陈旧读数和本地未知统计，刷新失败不得伪造新时间。 | 静态：当前读操作生成UpdatedAt；现有用量/监控消费者都有可复用接口。 |
| S18 实际扣除与账号成本统计 | [#8806](https://github.com/Wei-Shaw/sub2api/pull/8806) | 按account过滤的模型/端点统计 SHALL 分列actual_cost和account_cost，Gemini账号配额使用账号成本。 | 静态核实本地条件覆盖actualCost；SQLite查询/字段和Gemini聚合均已有。 |
| S19 价格启动不阻塞远端哈希 | [#8798](https://github.com/Wei-Shaw/sub2api/pull/8798) | 已加载本地有效价格后 SHALL 不阻塞启动等待远端hash，并保持可停止的后台同步。 | 静态：Initialize同步fetchRemoteHash；本地正文实际hash修复已存在。 |
| S20 上游 TCP 死连接检测 | [#8800](https://github.com/Wei-Shaw/sub2api/pull/8800) | 既有标准上游dialer SHALL 使用15s idle/5s interval/3 probes，Linux user-timeout设置失败仍拨号。 | 静态socket实现审查；未做丢包/死peer真实计时实验。 |
| S21 OpenAI 工具 schema 正则适配 | [#8772](https://github.com/Wei-Shaw/sub2api/pull/8772) | 现有OpenAI工具schema SHALL 窄处理来源不支持的regexp，并保留实例值与仍有效的schema约束。 | 静态：本地只有null-type/required清理，没有来源完整regexp parser；未向远端发送新样例。 |
| S22 既有流与请求热点窄优化 | [#8653](https://github.com/Wei-Shaw/sub2api/pull/8653)、[#8667](https://github.com/Wei-Shaw/sub2api/pull/8667)、[#8903](https://github.com/Wei-Shaw/sub2api/pull/8903)、[#8904](https://github.com/Wei-Shaw/sub2api/pull/8904) | 选定热点优化 SHALL 保持原字节、工具顺序、错误/用量语义和有界生命周期。 | Ollama200消息PoC语义对照通过，7.92ms→0.34ms、12.56MB→0.26MB分配；来源SSE escaped-usage predicate失败。 |
| S23 Go工具链与时钟测试维护 | [#8767](https://github.com/Wei-Shaw/sub2api/pull/8767)、[#8743](https://github.com/Wei-Shaw/sub2api/pull/8743) | Go维护 SHALL 对齐1.27.4声明/CI/工具，并让已有峰谷价格测试使用固定PricingAt。 | 只读Go1.27.4 embed构建及config/apicompat/claude三包通过；首次下载网络超时另存。 |


补丁位置和最小前置只记录在评估报告。本表定义选取的行为；同PR未选分支不构成授权。完整候选和四态见冻结证据。
