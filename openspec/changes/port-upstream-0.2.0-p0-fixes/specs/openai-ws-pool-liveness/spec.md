## ADDED Requirements

### Requirement: 不可保活的空闲连接必须在上游 keepalive 窗口过期前回收
WS 连接池的清理循环 SHALL 逐出同时满足以下全部条件的连接：未被租出、无等待者、
`supportsIdlePingWithoutReader()` 为 false、空闲时长达到回收阈值。阈值 SHALL 为 90 秒
（与既有的 `openAIWSConnHealthCheckIdle` 同值），本变更 MUST NOT 引入可配置项。

依据：coder/websocket 在没有 reader 的情况下消费不了 pong 帧，这类 socket 会一直留在池里直到上游
keepalive 窗口过期，取出来即是坏连接。

#### Scenario: 空闲且不可保活
- **WHEN** 一个未租出、无等待者、`supportsIdlePingWithoutReader()` 为 false 的连接空闲达到 90 秒
- **THEN** 该连接 MUST 从池中移除
- **THEN** 该连接 MUST 同时从 pinned 集合中移除（若在其中）
- **THEN** MUST 计入 scale-down 指标

#### Scenario: 支持无 reader ping 的连接
- **WHEN** 连接的 `supportsIdlePingWithoutReader()` 为 true
- **THEN** MUST NOT 因空闲时长被本规则逐出（仍受既有 `maxAge` 规则约束）

#### Scenario: 仍被租出或有等待者
- **WHEN** 连接正被租出，或 `waiters` 计数大于 0
- **THEN** MUST NOT 被逐出，无论空闲多久

#### Scenario: 被 pin 的连接
- **WHEN** 连接处于 pinned 状态且被 `isConnPinnedLocked` 判定为固定
- **THEN** MUST 沿用既有的跳过逻辑（本规则 MUST NOT 越过该守卫）

#### Scenario: 未达阈值
- **WHEN** 连接空闲时长小于 90 秒
- **THEN** MUST NOT 被本规则逐出
