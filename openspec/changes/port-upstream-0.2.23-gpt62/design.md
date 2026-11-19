# 设计

价格卡使用 USD/token，标准输入/输出/缓存读取分别为 Sol 2/10/0.10、Astra 8/40/0.40 USD/百万 token，缓存写入沿用该来源的 1.25 倍输入价格。Fast 与 Ultra Fast 在同一 token 费用明细上计算，长上下文再按输入及缓存 2 倍、输出 1.5 倍叠加。

促销终点为 `2027-03-01T08:00:00Z`，等于洛杉矶 2027-03-01 00:00 和北京时间 16:00。使用严格 Before，截止瞬间起为 6 倍，不依赖主机时区或运行时 tzdata。每次解析复制共享价格卡，截止前后的并发请求不污染资源、fallback 或预解析结果。

OpenAI RecordUsage 从 PricingAt 恢复已有共享网关定价上下文，用户价格、响应模型重算、Free Fast 对照及独立账号价格读取同一时刻。CostInput 允许显式 PricingAt，优先于上下文；缺省使用当前时间。保留既有长上下文配置开关、区间定价及用户/账号成本独立性。

Codex 原始描述保留 Sol 500k、Astra 272k 缺省窗口和共同 872k 上限，API 资源及 OpenCode 保留来源的 1050k/128k 能力。Ultra effort 为 Codex 多 agent 工作流，API effort 为 low/medium/high/xhigh/max，none/minimal 按既有校验方式返回 400。服务层级 ultrafast 与 effort 分别处理。

账号目录的显式 service_tiers 保持优先；官方 API 缺省补充 Ultra Fast。新套餐授予策略与强制分组/渠道字段保持独立范围。
