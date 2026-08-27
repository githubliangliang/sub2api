## ADDED Requirements

### Requirement: 空的能力容器必须与「未配置」等价
账号的 OpenAI 端点能力配置为**空容器**（`{}` / `[]`，含对应的强类型空 map / slice）时，
系统 SHALL 视为「未配置」，即不对该账号施加任何能力限制。MUST NOT 视为「已配置且不含任何能力」——
后者会让 API 直写、批量导入或历史数据留下的空对象把账号**静默**排除出文本调度，账号既不报错也不可用。

#### Scenario: 能力配置是空对象
- **WHEN** 账号的能力配置为 `{}`
- **THEN** 能力判定 MUST 返回「未配置」
- **THEN** 该账号 MUST 可参与文本请求调度

#### Scenario: 能力配置是空数组
- **WHEN** 账号的能力配置为 `[]`
- **THEN** 能力判定 MUST 返回「未配置」
- **THEN** 该账号 MUST 可参与文本请求调度

#### Scenario: 非空但所有能力都为 false
- **WHEN** 账号的能力配置非空，但所有取值均为 false
- **THEN** MUST 仍视为「已配置且不含任何能力」
- **THEN** 行为 MUST 与移植前一致，MUST NOT 被本次修改放宽

#### Scenario: 正常配置了部分能力
- **WHEN** 账号显式声明了一部分能力
- **THEN** MUST 只放行声明过的能力
- **THEN** 行为 MUST 与移植前一致

### Requirement: Chat 会话种子只采信前导的 system/developer 前缀
系统 SHALL 在从 Chat Completions 请求推导粘性会话种子时，只把**位于首个 user 消息之前**的
system / developer 消息计入种子。出现 user 消息或其他角色后，后续注入的 system / developer 消息
MUST NOT 再进入种子。否则每轮动态注入（时间戳、上下文摘要等）都会改变种子，
粘性哈希逐轮漂移，缓存命中率归零。

#### Scenario: 对话中途注入动态 system 消息
- **WHEN** 同一会话的后续请求在历史消息之后追加了内容不同的 system 消息
- **THEN** 推导出的会话种子 MUST 与前一轮一致
- **THEN** 粘性哈希 MUST 保持稳定，请求 MUST 仍命中同一账号

#### Scenario: 前导 system 前缀发生变化
- **WHEN** 首个 user 消息之前的 system 前缀内容变化
- **THEN** 会话种子 MUST 随之变化
- **THEN** 这属于预期行为：系统提示词换了就是新会话

#### Scenario: 只有 user 消息
- **WHEN** 请求不含任何 system / developer 消息
- **THEN** 种子 MUST 仍可由模型、工具定义与首个 user 消息稳定推导

#### Scenario: Responses 协议的 instructions
- **WHEN** 请求使用 Responses 协议并携带 `instructions`
- **THEN** 该字段 MUST 继续计入种子，行为与移植前一致
