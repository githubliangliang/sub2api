## ADDED Requirements

### Requirement: 缓存创建明细必须以「字段存在即覆盖」语义解析
系统 SHALL 在解析上游 usage 的 `cache_creation.ephemeral_5m_input_tokens` 与
`cache_creation.ephemeral_1h_input_tokens` 时，只要字段存在就覆盖当前累积值，MUST NOT 以
「值必须大于 0」为覆盖前提。该语义 MUST 同时适用于 Anthropic 透传流式解析与统一上游响应的
SSE usage patch 两条路径。

#### Scenario: 后续事件把 5m 明细降为 0
- **WHEN** `message_start` 报 `ephemeral_5m_input_tokens=1000`、`ephemeral_1h_input_tokens` 缺失
- **AND** 后续 `message_delta` 报 `ephemeral_5m_input_tokens=0`、`ephemeral_1h_input_tokens=1000`
- **THEN** 最终 5m 明细 MUST 为 `0`
- **THEN** 最终 1h 明细 MUST 为 `1000`
- **THEN** `5m + 1h` MUST NOT 超过上游报告的 `cache_creation_input_tokens`

#### Scenario: 上游完全不报明细
- **WHEN** 上游只报 `cache_creation_input_tokens=N`，两个 ephemeral 字段都不出现
- **THEN** 两个明细 MUST 保持 `0`
- **THEN** 计费 MUST 回落到「全部按 5m 单价」的既有行为，费用 MUST 与移植前一致

#### Scenario: 明细字段存在且为正
- **WHEN** 上游报 `ephemeral_5m_input_tokens=300`、`ephemeral_1h_input_tokens=700`
- **THEN** 两个明细 MUST 分别为 `300` 与 `700`，与移植前行为一致

### Requirement: 明细与聚合值矛盾时必须按比例封顶到聚合值
系统 SHALL 在计算缓存创建费用前对明细做归一化：当聚合值 `cache_creation_input_tokens` 为正
且 `5m + 1h` 超过它时，MUST 按两档原有比例缩回，使归一化后的 `5m + 1h` 等于聚合值。
负值 MUST 归零。聚合值非正或明细未超出时 MUST 原样返回，MUST NOT 引入新的偏差。

#### Scenario: 明细之和超过聚合值
- **WHEN** 聚合值为 `1000`，明细为 `5m=1000`、`1h=1000`
- **THEN** 归一化后 `5m + 1h` MUST 等于 `1000`
- **THEN** 归一化后两档 MUST 大致保持原比例（各约 `500`）
- **THEN** 计费 MUST NOT 高于按聚合值全额计价的上限

#### Scenario: 明细之和不超过聚合值
- **WHEN** 聚合值为 `1000`，明细为 `5m=400`、`1h=600`
- **THEN** 明细 MUST 原样保留
- **THEN** 费用 MUST 与移植前逐分一致

#### Scenario: 聚合值缺失或为零
- **WHEN** 聚合值为 `0` 或负数，而明细为正
- **THEN** MUST NOT 触发封顶
- **THEN** MUST 保持明细原值，交由既有计费分支处理

#### Scenario: 明细出现负值
- **WHEN** 任一明细为负数
- **THEN** 该明细 MUST 被视为 `0`
- **THEN** MUST NOT 产生负费用

### Requirement: 归一化 MUST NOT 改变不支持明细拆分的计费路径
系统 MUST 只在 `SupportsCacheBreakdown` 且至少一档明细单价为正时应用明细与归一化逻辑。
其余模型 MUST 继续按聚合 token 数乘统一缓存创建单价计费。

#### Scenario: 模型不支持缓存明细
- **WHEN** 定价未标记支持缓存明细拆分
- **THEN** 费用 MUST 等于 `cache_creation_input_tokens × 统一单价 × 倍率`
- **THEN** 归一化 MUST NOT 被调用，结果 MUST 与移植前一致
