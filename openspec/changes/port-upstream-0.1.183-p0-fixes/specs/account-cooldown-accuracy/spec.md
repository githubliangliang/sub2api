## ADDED Requirements

### Requirement: 账号限流解析必须覆盖 OpenCode Go 的用量限制类型
系统 SHALL 在解析上游 429 载荷推导重置时间时，除 `usage_limit_reached` 与 `rate_limit_exceeded`
之外，同样接受 `GoUsageLimitError`。该类型的重置时间只出现在人类可读的 message 中，
系统 MUST 从 message 解析出重置时长并换算为绝对重置时间。解析失败 MUST 回落到既有默认冷却，
MUST NOT 因解析失败而放弃冷却。

#### Scenario: 周用量耗尽且只在文案里给出重置时长
- **WHEN** 上游返回 `error.type = "GoUsageLimitError"`，message 为 `Weekly usage limit reached. Resets in 2 days.`
- **THEN** 推导出的重置时间 MUST 约为当前时间加 2 天
- **THEN** 账号冷却截止时间 MUST 采用该重置时间，而不是默认回落冷却
- **THEN** 账号 MUST NOT 在重置时间之前被重新调度

#### Scenario: 多段时长
- **WHEN** message 中的时长为 `Resets in 1h 30m`
- **THEN** 解析结果 MUST 为 90 分钟
- **THEN** MUST 支持 `s` / `m` / `h` / `d` / `w` 及其常见全称与复数写法

#### Scenario: 无法解析的文案
- **WHEN** `error.type = "GoUsageLimitError"` 但 message 中没有可识别的时长
- **THEN** MUST 返回「无重置时间」
- **THEN** 调用方 MUST 回落到既有默认冷却逻辑

#### Scenario: 时长溢出或非正值
- **WHEN** message 中的数值会导致时长溢出，或为 0 / 负数
- **THEN** MUST 返回「无重置时间」
- **THEN** MUST NOT 产生负的或环绕后的冷却截止时间

#### Scenario: 既有类型行为不变
- **WHEN** 上游返回 `usage_limit_reached` 或 `rate_limit_exceeded`
- **THEN** 解析结果 MUST 与移植前完全一致
