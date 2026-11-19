## ADDED Requirements

### Requirement: GPT-6.2 canonical model identity
系统 SHALL 在既有 OpenAI、Chat、Messages、Responses 与 WS 链路识别 gpt-6.2-sol 和 gpt-6.2-astra，保持账号映射和原始声明优先。

#### Scenario: Supported model and effort aliases
- **WHEN** 使用规范 ID、供应商前缀、大小写或既有 effort/compact 后缀
- **THEN** 解析到对应 6.2 模型，保持有效 effort；none/minimal 在兼容入口返回 400，工具请求遵循 Responses 能力。

#### Scenario: Unspecified or unknown model
- **WHEN** 未配置价格的 gpt-6.2 或未知 6.2 后缀被用于计费
- **THEN** 不借用 GPT-5.4 或另一模型的价格；运营者显式价格条目仍生效。

### Requirement: Official GPT-6.2 price cards
系统 SHALL 按冻结来源为 Sol 提供输入/输出/缓存写入/缓存读取 2/10/2.5/0.10、Astra 8/40/10/0.40 USD/百万 token，Fast 为标准价 2 倍，资源与双层 fallback 一致。

#### Scenario: Strict long context boundary
- **WHEN** 普通输入、缓存写入及缓存读取之和为 272000 或 272001，且既有长上下文计价开关允许
- **THEN** 272000 保持标准上下文，272001 的整次输入及缓存乘 2、整次输出乘 1.5，且可与 Fast/Ultra Fast 叠加。

#### Scenario: Explicit operator prices
- **WHEN** 分组、渠道或价格文件提供自定义价格或显式 0
- **THEN** 保持现有覆盖优先级及零价语义，用户价与独立账号成本分别计算。

### Requirement: Request-time Ultra Fast promotion
系统 SHALL 对两个模型按每次请求或 WS 每轮开始的 PricingAt，在 2027-03-01 00:00 America/Los_Angeles 前使用标准价 3 倍，从该瞬间起使用 6 倍。

#### Scenario: Delayed or concurrent settlement
- **WHEN** 截止前请求在截止后结算，或两侧时刻交错复用同一价格解析器
- **THEN** 保持各自 3/6 倍，输入、输出和缓存费用一致，共享价格与独立账号统计不受其它请求污染。

#### Scenario: Existing Fast and dedup controls
- **WHEN** Fast、Free Fast、Ultra Fast 及相同请求重复记录
- **THEN** Free Fast 仅减免既有 Fast 的用户费用，Ultra Fast 按选定倍率计费，SQLite 结算与 quota 去重仍只扣一次。

### Requirement: GPT-6.2 client catalog and configuration
系统 SHALL 保留来源 Codex 描述及账号声明优先，并在既有白名单、映射预设和 OpenCode 配置提供两个模型。

#### Scenario: Official API and account declarations
- **WHEN** 通过官方 API 或带显式目录的 OAuth/API 账号生成 Codex manifest
- **THEN** API 采用 full Responses，保留账号已声明的 service_tiers 与扩展字段；缺省模型窗口、effort、工具能力符合冻结描述。

#### Scenario: Existing configuration export
- **WHEN** 选择模型白名单、映射预设或导出 OpenCode 配置
- **THEN** 两个模型可往返保存，导出 1050000 context、128000 output、store=false 与 low/medium/high/xhigh/max 变体。
