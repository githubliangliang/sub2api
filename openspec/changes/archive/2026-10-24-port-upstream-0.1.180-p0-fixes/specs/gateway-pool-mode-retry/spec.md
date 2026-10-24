## ADDED Requirements

### Requirement: 池模式同账号重试资格必须在所有转发协议上一致
系统 SHALL 在构造上游故障转移错误时统一填充「可在同账号上重试」标记，
判定条件为：账号处于池模式**且**上游状态码属于池模式可重试集合。该标记 MUST 同时出现在原生
Anthropic 转发、转 Chat Completions 转发与转 Responses 转发三条路径上，
MUST NOT 只在其中一条上生效。

#### Scenario: 池模式账号在转 Chat Completions 路径上遇到可重试状态码
- **WHEN** 池模式账号经 Chat Completions 转发得到属于可重试集合的上游错误
- **THEN** 故障转移错误 MUST 标记为可在同账号重试
- **THEN** 请求 MUST 先在同一账号上重试，而不是直接切换账号

#### Scenario: 池模式账号在转 Responses 路径上遇到可重试状态码
- **WHEN** 池模式账号经 Responses 转发得到属于可重试集合的上游错误
- **THEN** 故障转移错误 MUST 标记为可在同账号重试

#### Scenario: 非池模式账号
- **WHEN** 账号不处于池模式
- **THEN** MUST NOT 标记为可在同账号重试
- **THEN** 行为 MUST 与移植前一致

#### Scenario: 状态码不属于可重试集合
- **WHEN** 上游状态码不在池模式可重试集合内
- **THEN** MUST NOT 标记为可在同账号重试

### Requirement: 刚被判定要禁用的账号 MUST NOT 进入同账号重试
系统 SHALL 采信上游错误处理返回的「该账号应被禁用」判定：判定为真时，
即使账号处于池模式且状态码可重试，也 MUST NOT 标记为可在同账号重试。

#### Scenario: 上游错误处理判定禁用
- **WHEN** 上游错误处理对本次错误返回「应禁用该账号」
- **THEN** 故障转移错误 MUST NOT 标记为可在同账号重试
- **THEN** 请求 MUST 走正常的换账号故障转移

#### Scenario: 上游错误处理未判定禁用
- **WHEN** 上游错误处理返回「不需禁用」
- **THEN** 池模式与状态码条件成立时 MUST 标记为可在同账号重试
