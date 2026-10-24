## ADDED Requirements

### Requirement: 流式 tool_call 增量中的空身份字段不得下发
在 Chat Completions 流式直转路径上，网关 SHALL 从 `choices[*].delta.tool_calls[*]` 中删除
**存在且为空字符串**的 `id` 与 `function.name` 字段后再下发给客户端。删除是无状态的：
网关 MUST NOT 记忆首包身份，也 MUST NOT 补写任何值——「字段缺失」本身就是「不要覆盖」的信号。

依据：客户端普遍按 `field !== undefined` 合并流式增量，空串会被当作有效值覆盖首包的合法
`id` / `name`，最终以空名字派发工具（`unknown tool ""`）。该行为不限定某个上游厂商。

#### Scenario: 首包携带合法身份
- **WHEN** 某个 `tool_calls[i]` 的增量带非空 `id` 与非空 `function.name`
- **THEN** 该增量 MUST 原样下发，不得有任何字段被删除

#### Scenario: 后续参数增量携带空身份
- **WHEN** 后续增量带 `"id": ""` 与 `"function": {"name": ""}`，只追加 `arguments` 片段
- **THEN** `id` 与 `function.name` 两个字段 MUST 从下发的 payload 中删除
- **THEN** `arguments` 片段 MUST 原样保留
- **THEN** `index` 与 `type` MUST 原样保留

#### Scenario: 只有其中一个身份字段为空
- **WHEN** 增量带非空 `id` 但 `function.name` 为空串（或反之）
- **THEN** 只有为空的那一个字段 MUST 被删除
- **THEN** 非空的那一个 MUST 保留原值

#### Scenario: arguments 为空串
- **WHEN** 增量的 `function.arguments` 是空串
- **THEN** `arguments` MUST NOT 被删除（空 arguments 是合法的首包形态）

#### Scenario: 多 choice / 多 tool_calls 下标
- **WHEN** 一个 chunk 里有多个 `choices`，或一个 `delta` 里有多个 `tool_calls` 元素
- **THEN** 每一个元素 MUST 独立按上述规则处理

### Requirement: 剥离逻辑必须对非工具流量零影响且失败安全
该处理 SHALL 只作用于流式 chunk 的 `delta.tool_calls`；非流式响应的 `message.tool_calls`
不在范围内。处理 MUST 在以下任一情况下原样返回输入：payload 为空、不含 `tool_calls` 字面量、
不是合法 JSON、没有 `choices` 数组、没有 `delta` 对象、没有 `tool_calls` 数组、
或没有任何字段需要删除。任何一次 JSON 改写失败 MUST 导致整个 payload 原样返回（fail-closed）。

#### Scenario: 普通文本 chunk
- **WHEN** chunk 不含 `tool_calls`
- **THEN** MUST 在进入 JSON 解析前就短路返回原始行

#### Scenario: SSE 控制行
- **WHEN** 行是 `data: [DONE]`、空行或非 `data:` 前缀的行
- **THEN** MUST 原样返回

#### Scenario: 非法 JSON 载荷
- **WHEN** `data:` 之后不是合法 JSON
- **THEN** MUST 原样返回，MUST NOT 抛错、MUST NOT 中断流
