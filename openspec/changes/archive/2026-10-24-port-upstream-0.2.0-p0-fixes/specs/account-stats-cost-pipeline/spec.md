## ADDED Requirements

### Requirement: 账号统计成本必须与用户计费共用同一条定价管线
账号统计的模型定价文件分支（优先级 3）SHALL 通过 `CalculateCostWithServiceTier` 计算成本，
MUST NOT 自行实现「单价 × token 数」求和。该分支 SHALL 继续以 `channelPricing = nil` 调用，
以保持优先级 3 的语义：只取模型定价文件，不引入渠道自定义定价。

依据：第二份实现意味着每加一个定价特性都要手工镜像一次。本仓库的
`normalizeBillingServiceTier` 只做 `ToLower` + `TrimSpace`、**不把 `fast` 归一成 `priority`**，
所以 `service_tier = "fast"` 此前会落到手算分支按标准价计。

#### Scenario: 标准档（无 service tier、未触发长上下文）
- **WHEN** 请求没有 service tier 且 token 数不触发长上下文阶梯
- **THEN** 计算结果 MUST 与本变更前的手算结果一致（这是本条的兼容性底线）

#### Scenario: `service_tier = "fast"`
- **WHEN** 用量记录的 service tier 是 `fast`
- **THEN** MUST 走统一管线，MUST NOT 按标准价计

#### Scenario: `service_tier = "priority"` / `"flex"`
- **WHEN** service tier 是 `priority` 或 `flex`
- **THEN** 行为 MUST 与本变更前一致（此前已走统一管线）

#### Scenario: 触发长上下文阶梯
- **WHEN** token 数触发长上下文阶梯
- **THEN** 行为 MUST 与本变更前一致（此前已走统一管线）

#### Scenario: 图片输出 token
- **WHEN** 用量含图片输出 token
- **THEN** MUST 按 output 子集计价

#### Scenario: 定价缺失或结果非正
- **WHEN** 统一管线报错、返回空、或总成本不大于 0
- **THEN** 本分支 MUST 返回「无结果」，交由后续优先级处理（与本变更前的失败语义一致）

### Requirement: 优先级顺序与渠道定价隔离不得改变
本变更 SHALL NOT 改动 `resolveAccountStatsCost` 的优先级顺序，也 MUST NOT 让优先级 3 开始读取
渠道自定义定价。

#### Scenario: 自定义规则优先
- **WHEN** 存在命中的账号统计自定义规则
- **THEN** MUST 仍按规则结果计价，不进入优先级 3

#### Scenario: 渠道定价不泄漏进优先级 3
- **WHEN** 该模型在某渠道上配了自定义定价
- **THEN** 优先级 3 的结果 MUST NOT 受该渠道定价影响
