## ADDED Requirements

### Requirement: Ollama Cloud 的思维字段必须与 reasoning_content 双向对齐
Ollama Cloud 的 OpenAI 兼容 Chat Completions 把思维内容放在 `reasoning` / `thinking`，
而 DeepSeek / OpenAI 系客户端只认 `reasoning_content`。系统 SHALL 仅在 Ollama Cloud 的原生
Chat Completions 直转路径上做 wire 层字段对齐：出站把客户端历史里的 `reasoning_content` 映射为
上游字段，入站把上游字段映射回 `reasoning_content`。非 Ollama Cloud 账号 MUST NOT 受影响。

#### Scenario: 客户端历史携带 reasoning_content
- **WHEN** 请求消息中带有 `reasoning_content`，且账号为 Ollama Cloud
- **THEN** 出站请求 MUST 携带上游认识的思维字段
- **THEN** 已存在上游字段时 MUST NOT 重复覆盖

#### Scenario: 上游非流式响应返回 reasoning
- **WHEN** 上游响应的 choice 中带有思维字段
- **THEN** 返回给客户端的响应 MUST 带 `reasoning_content`
- **THEN** 已有 `reasoning_content` 时 MUST NOT 被覆盖

#### Scenario: 上游流式响应
- **WHEN** 上游以 SSE 增量返回思维内容
- **THEN** 每一行 MUST 被同样映射为 `reasoning_content`

#### Scenario: 非 Ollama Cloud 账号
- **WHEN** 账号不是 Ollama Cloud
- **THEN** 请求与响应 MUST 逐字节不变

### Requirement: Ollama Cloud 请求的输出上限必须 clamp 到 provider 上限
Ollama Cloud 对输出 token 数有 provider 级硬上限，超过会被上游直接 400 拒绝。系统 SHALL 把
`max_tokens` 与 `max_completion_tokens` clamp 到该上限。上限 MUST 可通过账号 `extra` 的可选键覆盖；
该键取 0 或负数时 MUST 视为显式禁用 clamp；类型异常时 MUST 回退默认上限。
该上限与模型无关，MUST NOT 做模型过滤。

#### Scenario: 客户端请求超过 provider 上限
- **WHEN** 请求的输出 token 上限超过默认 provider 上限
- **THEN** 出站值 MUST 被 clamp 到上限
- **THEN** 请求 MUST NOT 因该字段被上游 400 拒绝

#### Scenario: 账号自定义了更低上限
- **WHEN** 账号 `extra` 配置了正数上限
- **THEN** clamp MUST 采用该值

#### Scenario: 账号显式禁用 clamp
- **WHEN** 账号 `extra` 中该键为 0 或负数
- **THEN** MUST NOT 修改请求中的输出上限字段

#### Scenario: 请求未指定输出上限
- **WHEN** 两个字段都不存在
- **THEN** MUST NOT 写入任何上限字段

#### Scenario: clamp 的可观测性
- **WHEN** 任一字段实际被 clamp
- **THEN** MUST 记录一条调试级日志，便于事后解释「我明明要了更大的 max_tokens」
