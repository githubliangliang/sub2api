## ADDED Requirements

### Requirement: 只有上下文缺口才触发历史回放
HTTP bridge SHALL 仅在以下任一条件成立时回放历史：请求带 `previous_response_id`；
或请求含工具调用输出**且**当前上下文没有覆盖全部相关 `call_id`。
「payload 里存在工具调用输出」本身 MUST NOT 单独构成回放理由。

依据：客户端（如 Codex CLI）常常在每一轮都自带完整的工具历史。旧判据只看「有没有 tool call
output」，于是这段历史被网关再拼一次，同一批 item 在发给上游的 `input` 里出现两遍。

#### Scenario: 客户端自带完整工具历史
- **WHEN** 请求的 `input` 同时含 `custom_tool_call`（带 `call_id`）与其配对的 `custom_tool_call_output`，且不带 `previous_response_id`
- **THEN** MUST NOT 回放历史
- **THEN** 发给上游的 `input` MUST 与客户端提交的条目一一对应，不得出现重复条目

#### Scenario: 只有工具输出、缺少对应的调用上下文
- **WHEN** 请求只带 `custom_tool_call_output`，上下文里没有同 `call_id` 的调用条目
- **THEN** MUST 回放历史以补齐该 `call_id` 的调用条目
- **THEN** 回放后的 `input` MUST 让每个工具输出都能找到同 `call_id` 的调用条目

#### Scenario: 显式续接
- **WHEN** 请求带非空 `previous_response_id`
- **THEN** MUST 回放历史，无论工具上下文是否完整

### Requirement: 工具输出覆盖度分析必须同时支持数组与单对象 input
`input` 字段 SHALL 既可以是条目数组，也可以是单个条目对象；两种形态 MUST 走同一套覆盖度分析。
MUST NOT 因为 `input` 不是数组就直接返回零值覆盖度。

#### Scenario: input 是单个工具输出对象
- **WHEN** `input` 是 `{"type":"custom_tool_call_output","call_id":"call_1","output":"..."}`
- **THEN** 覆盖度分析 MUST 识别出存在工具调用输出
- **THEN** MUST 识别出该 `call_id` 未被上下文覆盖，从而触发回放

#### Scenario: 条目缺少 call_id
- **WHEN** 工具输出条目没有 `call_id` 或其值为空
- **THEN** 覆盖度分析 MUST 判定为「未覆盖」（保守回放），MUST NOT 判定为已覆盖

### Requirement: 回放历史中不得包含没有配对输出的工具调用
构造回放输入时，系统 SHALL 从**历史**条目中剔除那些 `call_id` 在历史与本轮的工具输出集合里
都找不到配对的工具调用上下文条目。配对集合 MUST 同时取自上一轮完整输入与本轮输入。
非工具调用类的历史条目 MUST 原样保留。

依据：上一轮产生了工具调用、但客户端这一轮换了话题不提交对应输出时，那条孤儿调用会被回放进
新请求，上游据此认为有一个悬空的工具调用。

#### Scenario: 上一轮的工具调用没有配对输出
- **WHEN** 上一轮输入含 `custom_tool_call`（`call_id=call_1`），本轮与上一轮都没有 `call_id=call_1` 的输出
- **THEN** 回放后的 `input` MUST NOT 包含该 `custom_tool_call`
- **THEN** 其余历史条目（如用户消息）MUST 保留

#### Scenario: 配对输出出现在本轮
- **WHEN** 上一轮输入含 `call_id=call_1` 的调用，本轮输入带 `call_id=call_1` 的输出
- **THEN** 该调用条目 MUST 保留在回放结果中

#### Scenario: 没有历史可回放
- **WHEN** 不存在上一轮完整输入
- **THEN** MUST 直接使用本轮条目，MUST NOT 因过滤逻辑改变本轮内容
