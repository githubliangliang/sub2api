## ADDED Requirements

### Requirement: API key 路径必须保持会话缓存身份
走 API key 的 Chat Completions 请求 SHALL 携带与 OAuth 路径一致的会话缓存身份，使上游 prompt cache
能够命中。同一逻辑会话的连续请求 MUST 解析出同一个缓存身份。

依据：API key 路径此前丢掉了该身份，导致本可命中的 prompt cache 落空——表现为费用偏高、TTFT 变长，
但没有任何报错。

#### Scenario: 同一会话的连续请求
- **WHEN** 同一 API key 会话连续发出两次请求
- **THEN** 两次请求解析出的缓存身份 MUST 相同

#### Scenario: 不同会话
- **WHEN** 两个不同会话各发一次请求
- **THEN** 两者的缓存身份 MUST 不同（MUST NOT 串号）

#### Scenario: 与 OAuth 路径一致
- **WHEN** 同一逻辑会话分别走 API key 与 OAuth 路径
- **THEN** 缓存身份的**推导规则** MUST 一致（本条不要求两条路径产出同一个值）

#### Scenario: 客户端未提供任何身份线索
- **WHEN** 请求里没有可用于推导身份的字段
- **THEN** MUST 退回本变更前的行为，MUST NOT 编造一个会在会话间碰撞的身份
