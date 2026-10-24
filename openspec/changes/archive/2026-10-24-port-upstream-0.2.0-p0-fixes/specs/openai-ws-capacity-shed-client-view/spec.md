## ADDED Requirements

### Requirement: ctx_pool WS ingress 下发前必须改写上游容量降载错误码
在 ctx_pool 的 WS ingress 直写路径上，事件类型为 `error` 或 `response.failed` 时，网关 SHALL 对
**发给客户端的副本**应用 `sanitizeOpenAICapacityShedErrorCodeForClient`。该处理 MUST 与 HTTP/SSE
路径、http_bridge 路径保持同一语义——ingress 是此前唯一漏掉的一条路径。

依据：Codex CLI 按闭集判定错误码，`server_is_overloaded` / `slow_down` 属致命集，客户端会打印
"Selected model is at capacity" 并终止会话、不退避重试。同一账号同一时刻切到 http_bridge 却能正常重试。

#### Scenario: 上游返回容量降载错误
- **WHEN** ingress 收到 `error` 事件，其错误码是 `server_is_overloaded` 或 `slow_down`
- **THEN** 下发给客户端的 payload MUST 已被改写为客户端会退避重试的形态

#### Scenario: `response.failed` 形态
- **WHEN** 降载以 `response.failed` 事件到达
- **THEN** MUST 应用同一改写

#### Scenario: 非容量类错误码
- **WHEN** ingress 收到 `error` 事件但错误码不属于容量降载集合
- **THEN** 下发给客户端的 payload MUST 原样不变

#### Scenario: 非终止类事件
- **WHEN** 事件类型不是 `error` 也不是 `response.failed`
- **THEN** MUST NOT 进入改写逻辑

### Requirement: 账号状态判定必须使用未改写的原始 payload
改写结果 SHALL 写入独立变量；上游原始消息 MUST 保持原值，供写出点之后的
`markOpenAIWSClientVisibleFailure` 与 `handleOpenAIWSTerminalTransientFailure` 使用。
网关 MUST NOT 原地改写上游消息。

依据：客户端可见视图与账号状态判定视图必须分离，否则真实的容量降载信号会被从摘号/冷却判定中抹掉。
这是 `sanitizeOpenAICapacityShedErrorCodeForClient` 注释里写明的前提，也是 http_bridge 的既有写法。

#### Scenario: 改写发生后的账号状态判定
- **WHEN** 某次降载事件的客户端副本已被改写
- **THEN** 账号状态判定 MUST 仍按原始错误码进行（该账号 MUST 被记为遇到上游容量降载）

#### Scenario: 改写未发生
- **WHEN** 错误码不属于容量降载集合，改写未发生
- **THEN** 客户端副本与账号状态判定 MUST 使用同一份未变的 payload
