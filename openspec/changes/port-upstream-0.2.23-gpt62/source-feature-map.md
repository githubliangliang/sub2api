# 来源行为映射

| 选定行为 | 来源 | 本地适配 |
|---|---|---|
| 模型身份、effort、Responses 工具、采样校验 | PR8906 constants / compat / request body | 复用现有统一模型归一化和最终请求校验 |
| Codex 描述与 API/OAuth 协议能力 | PR8906 两份 descriptor / manifest | 原文保留；账号声明覆盖缺省；API 使用 full Responses |
| 标准/Fast 价格与 272k 整次计价 | PR8906 resource / billing / pricing | 两层 fallback、显式零价与配置优先级一起验证 |
| Ultra Fast 3→6 促销时钟 | PR8906 billing / usage | 用既有 PricingAt 和上下文贯通 resolver、延迟结算及独立账号成本 |
| 既有客户端配置入口 | PR8906 whitelist / UseKeyModal | 加入两个模型与五个 API effort |

原先第三档分类保留在 [历史评估](../../../docs/upstream-sync/PORTING-0.2.23.md)。本次是用户明确选择后的独立实施，不重写此前未选定的历史措辞。
