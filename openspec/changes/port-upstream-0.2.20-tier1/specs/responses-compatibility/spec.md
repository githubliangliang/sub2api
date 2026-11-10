## ADDED Requirements

### Requirement: F05 缓冲 Responses reasoning 保真

Responses转Messages/Chat SHALL 优先保留非空summary，否则保留content.reasoning_text；encrypted_content签名独立保留。

#### Scenario: 具体边界与可观测结果

- **WHEN** Responses reasoning没有summary但content包含reasoning_text，另有加密签名
- **THEN** Messages thinking与Chat reasoning_content保留可见文字，签名单独保留；summary非空时不重复拼入content

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** summary优先、不重复拼接、仅reasoning_text、仅签名、普通文本及两条真实缓冲网关路径；usage/dedup不重复。

### Requirement: F06 Responses 完成消息字段

完成响应与流式item SHALL 使用一致message序列化，空content为数组，refusal、annotations及logprobs保持字段语义。

#### Scenario: 具体边界与可观测结果

- **WHEN** 完成响应含空assistant message或refusal内容
- **THEN** 空content序列化为数组，refusal使用refusal字段，output_text保留annotations/logprobs，终态和流式item一致

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** 空message、output_text、refusal、annotation/logprobs、tool_search/function_call；JSON及SSE终态一致，序号0不被省略。
