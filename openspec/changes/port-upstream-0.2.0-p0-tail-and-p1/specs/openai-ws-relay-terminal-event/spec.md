## ADDED Requirements

### Requirement: turn 活跃期间的上游 close 必须判为失败
WS v2 透传中继在读取上游帧失败时，若该错误本身是干净断开（`isDisconnectError` 为真）**且**当前存在
活跃 turn，则 SHALL 把结果判为**非 graceful**，并把错误信息包成「terminal event 之前关闭」的形态。
若没有活跃 turn，干净断开 MUST 继续判为 graceful。

依据：干净的 WebSocket close（1000 / EOF）只描述传输层握手。一旦上游开始了 Responses turn，
成功还要求一个终止协议事件；把早到的 close 当 graceful 会让适配层在 turn 仍活跃时报
`relay_completed`，把一次失败记成成功。

#### Scenario: turn 活跃时收到 1000
- **WHEN** 上游在某个 turn 仍活跃时以 close 1000 关闭连接
- **THEN** 中继退出信号 MUST 标记为非 graceful
- **THEN** 错误信息 MUST 指明是在终止事件之前关闭
- **THEN** 适配层 MUST NOT 报 `relay_completed`

#### Scenario: turn 活跃时收到 EOF
- **WHEN** 同上但错误是 EOF
- **THEN** 行为 MUST 与上一个 Scenario 一致

#### Scenario: 无活跃 turn 时收到 1000
- **WHEN** 上游在没有活跃 turn 时干净关闭
- **THEN** MUST 仍判为 graceful（本变更 MUST NOT 把正常收尾判成故障）

#### Scenario: 非干净断开
- **WHEN** 读帧失败的原因本来就不是干净断开
- **THEN** 行为 MUST 与本变更前一致（本条只改「干净断开」这一侧的判定）

#### Scenario: 追踪事件与退出信号一致
- **WHEN** 判定发生翻转
- **THEN** 中继追踪事件里的 graceful 标记与退出信号里的 MUST 是同一个值

### Requirement: 活跃 turn 的解析必须无副作用
用于判定「是否存在活跃 turn」的查询 SHALL 是纯函数：state 为空或没有活跃 turn 时返回空字符串，
MUST NOT 修改任何中继状态。

#### Scenario: state 为空
- **WHEN** 中继状态为 nil
- **THEN** MUST 返回空字符串，MUST NOT panic

#### Scenario: 没有活跃 turn
- **WHEN** 状态存在但没有活跃 turn
- **THEN** MUST 返回空字符串
