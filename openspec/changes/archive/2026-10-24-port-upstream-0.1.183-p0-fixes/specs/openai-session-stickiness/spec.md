## ADDED Requirements

### Requirement: 显式会话头必须同时识别连字符与下划线写法
系统 SHALL 把 `session-id` 与 `session_id` 都视为显式粘性会话头。Go 的 `http.Header` 会把两者
规范化成不同的键（`Session-Id` 与 `Session_id`），因此**只登记一种写法等于让另一种完全落空**。
识别顺序 MUST 为 `session-id` → `session_id` → `conversation_id` → OpenCode 亲和头 → CodeBuddy 头。

#### Scenario: Codex CLI 发送连字符写法
- **WHEN** 客户端请求携带 `session-id: abc123`
- **THEN** 该值 MUST 被用作显式会话标识
- **THEN** 同一 `session-id` 的后续请求 MUST 命中同一账号的粘性绑定
- **THEN** 会话来源 MUST 记为 `header_session-id`

#### Scenario: 旧客户端发送下划线写法
- **WHEN** 客户端请求携带 `session_id: abc123`
- **THEN** 行为 MUST 与移植前一致
- **THEN** 会话来源 MUST 记为 `header_session_id`

#### Scenario: 两种写法同时出现
- **WHEN** 请求同时携带 `session-id` 与 `session_id`
- **THEN** MUST 采用 `session-id`
- **THEN** MUST NOT 因两值不同而产生两个会话哈希

#### Scenario: WebSocket 转发日志的会话解析
- **WHEN** Responses WebSocket 请求携带 `session-id`
- **THEN** 会话解析 MUST 取到该值而不是回落到 `none`
- **THEN** 会话来源标记 MUST 与 HTTP 路径使用同一套取值优先级

### Requirement: 容量溢出 MUST NOT 被写回持久粘性绑定
当粘性账号自身健康、仅因有界等待队列已满而由后续层临时借用另一账号完成**本次**请求时，
系统 MUST NOT 用被借用的账号覆盖该会话的持久粘性绑定。持久绑定 MUST 继续指向原账号，
使突发结束后会话回到原账号，MUST NOT 因一次短促排队就把整段会话迁到 cache-cold 账号。

#### Scenario: 粘性账号排队已满时的一次性借号
- **WHEN** 会话已绑定账号 A，A 健康但等待队列达到 `StickySessionMaxWaiting`
- **AND** 调度在后续层选中账号 B 完成本次请求
- **THEN** 本次请求 MUST 由 B 成功服务
- **THEN** 会话的持久粘性绑定 MUST 仍指向 A
- **THEN** 突发结束后下一次请求 MUST 回到 A

#### Scenario: 粘性账号不可用时的正常改绑
- **WHEN** 会话已绑定账号 A，但 A 被排除、冷却或不可调度
- **AND** 调度改选账号 B
- **THEN** 持久粘性绑定 MUST 更新为 B
- **THEN** 行为 MUST 与移植前一致

#### Scenario: 无会话标识的请求
- **WHEN** 请求没有可用的会话哈希
- **THEN** MUST NOT 写入任何粘性绑定
- **THEN** 溢出判定 MUST NOT 影响账号选择结果
