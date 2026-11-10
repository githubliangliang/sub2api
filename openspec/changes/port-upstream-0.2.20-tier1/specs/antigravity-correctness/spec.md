## ADDED Requirements

### Requirement: F09 Antigravity 直连 SSE 响应

流转换器 SHALL 按response信封是否存在选择解包形态，直接Gemini响应的文本、工具和用量不能静默丢失。

#### Scenario: 具体边界与可观测结果

- **WHEN** SSE data是无response外层的直接Gemini candidates对象
- **THEN** 文本、工具和usage被正确转换；有response外层的原有格式仍正确且非法JSON不伪造成功

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** 直接/包装响应、空/非法JSON、工具、usage和错误尾帧；普通流与缓冲兼容路径不产生假完成。

### Requirement: F10 Antigravity JSON Schema 参数

Antigravity转发 SHALL 将responseJsonSchema转换成上游支持的responseSchema，已有responseSchema优先。

#### Scenario: 具体边界与可观测结果

- **WHEN** Antigravity请求只含generationConfig.responseJsonSchema，或同时已有responseSchema
- **THEN** 前者清理并映射到responseSchema，后者保留已有schema优先级；camel/snake形式一致

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** camel/snake命名、双字段优先、非object、undefined清理、原body不变正例；不修改原生Gemini协议。
