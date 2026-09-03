## ADDED Requirements

### Requirement: 未声明 server-side-fallback beta 时必须剥离 `fallbacks`
向 Anthropic 上游（原生路径与 Bedrock 路径）发出的请求体中，若 `anthropic-beta` 请求头**不含**
server-side-fallback beta 标识，网关 SHALL 删除 `body.fallbacks` 后再发出。若请求头**含有**该标识，
`fallbacks` MUST 原样透传。

依据：上游在未声明该 beta 时对 `fallbacks` 返回 400 `fallbacks: Extra inputs are not permitted`。
客户端已开始默认发送该字段，OAuth mimic 与默认 API-key beta 两条路径都会中。剥离方式与既有的
`context_management` sanitize 对齐。

#### Scenario: 客户端发 fallbacks 但未声明 beta
- **WHEN** 请求体含 `fallbacks`，`anthropic-beta` 头不含 server-side-fallback 标识
- **THEN** 发往上游的请求体 MUST NOT 含 `fallbacks`
- **THEN** 请求体的其余字段 MUST 保持不变

#### Scenario: 客户端声明了 beta
- **WHEN** 请求体含 `fallbacks`，且 `anthropic-beta` 头含 server-side-fallback 标识
- **THEN** 发往上游的请求体 MUST 原样保留 `fallbacks`

#### Scenario: OAuth mimic 路径
- **WHEN** 请求走 OAuth mimic（网关自行组装 beta 头）且组装结果不含该标识
- **THEN** MUST 剥离 `fallbacks`

#### Scenario: Bedrock 路径
- **WHEN** 请求经 Bedrock 请求装配发出
- **THEN** MUST 应用同一条剥离规则

#### Scenario: 请求体不含 fallbacks
- **WHEN** 请求体没有 `fallbacks` 字段
- **THEN** 请求体 MUST 逐字节不变（MUST NOT 因为走过 sanitize 而被重新序列化改形）
