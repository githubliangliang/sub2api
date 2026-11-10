## ADDED Requirements

### Requirement: F07 鉴权取消分类

API Key查询因当前请求上下文取消而失败时 SHALL 按客户端离开结束；有效请求的真正仓储错误仍500，业务错误保持原分类，日志不含凭据。

#### Scenario: 具体边界与可观测结果

- **WHEN** 请求上下文已取消且API Key查询返回上下文错误
- **THEN** 普通和Google鉴权按客户端离开结束；未取消请求的真实仓储失败仍500，业务401/503分类保留

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** 取消/截止、真实DB失败、格式错误Key、503过载及401错误；不把所有context错误无条件吞成499，不记录Key。

### Requirement: F08 413 透传规则优先级

耗尽failover后，已有管理员413透传规则 SHALL 先于内置body-too-large响应生效，账号切换行为保持。

#### Scenario: 具体边界与可观测结果

- **WHEN** 上游413已耗尽账号切换且存在匹配的管理员透传规则
- **THEN** 执行该规则的状态/文案/监控策略；不存在规则时保留内置413，已开始SSE使用既有流式错误格式

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** 有/无匹配规则、自定义状态/文案、skip-monitoring、已开始SSE及Retry-After；不提前终止账号failover。
