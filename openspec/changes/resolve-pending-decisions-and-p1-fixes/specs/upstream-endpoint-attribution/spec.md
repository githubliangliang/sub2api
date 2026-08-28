## ADDED Requirements

### Requirement: 错误与用量记录必须归因到真实上游端点
当请求实际发往的上游端点与入站端点不同时（例如账号被判定为强制走 Chat Completions，
而客户端调用的是 Responses），系统 SHALL 记录**真实**的上游端点。
在没有转发结果对象的失败路径（503、传输失败）上 MUST 同样成立，
MUST NOT 回落到入站端点——那会让排查指向错误的方向。

#### Scenario: 强制 Chat Completions 且上游 503
- **WHEN** 客户端调用 Responses，账号配置为强制走 Chat Completions
- **AND** 上游返回 503 或连接失败，没有产生转发结果
- **THEN** 错误记录中的上游端点 MUST 为 `/v1/chat/completions`
- **THEN** MUST NOT 记为入站的 `/v1/responses`

#### Scenario: 正常完成的请求
- **WHEN** 请求正常完成并产生了转发结果
- **THEN** 归因 MUST 与改动前一致

#### Scenario: 入站与上游端点相同
- **WHEN** 请求按原协议直转，未发生协议改写
- **THEN** 归因 MUST 为该端点，行为与改动前一致

### Requirement: 端点记录 MUST NOT 在故障转移尝试之间残留
同一请求的多次账号故障转移尝试共用同一个请求上下文，系统 SHALL 在每次转发尝试开始时清空
已记录的端点，并在实际发送时重新记录。MUST NOT 让上一次尝试的端点残留到下一次。

#### Scenario: 连续两次尝试走不同协议
- **WHEN** 第一次尝试走 Chat Completions 失败，第二次换账号后走 Responses
- **THEN** 最终记录的端点 MUST 是第二次尝试实际使用的端点

#### Scenario: 首次尝试即失败且未发出请求
- **WHEN** 转发在构造请求阶段就失败，没有实际发送
- **THEN** MUST NOT 残留上一个请求或上一次尝试的端点

#### Scenario: 非 OpenAI 兼容平台
- **WHEN** 请求走的是 OpenAI / Grok 之外的平台
- **THEN** 归因 MUST 沿用既有的入站推导逻辑，行为与改动前一致
