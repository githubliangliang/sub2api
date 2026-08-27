## ADDED Requirements

本 capability 覆盖五组出站请求约束。共同不变量：**出站请求必须满足上游的硬约束，否则请求整体
失败、被上游改写，或在后续轮次被拒绝**。

### Requirement: Gemini 工具 schema 必须剔除上游不接受的字段
系统 SHALL 在把工具 schema 转成 Gemini 形态时，除既有剔除列表外同样剔除 `deprecated`。
剔除 MUST 递归作用于嵌套对象。

#### Scenario: schema 携带 deprecated
- **WHEN** 客户端工具 schema 的任一层级包含 `deprecated`
- **THEN** 发往 Gemini 的 schema MUST NOT 包含该字段
- **THEN** 其余字段与嵌套结构 MUST 保持不变

### Requirement: Gemini 工具 schema 的 enum 值必须全部为字符串
系统 SHALL 把 `enum` 中的非字符串标量（数字、布尔、null）编码成字符串；当任一元素是对象或数组
等无法编码为标量的值时，MUST 整体删除该 `enum` 而不是透传。

#### Scenario: enum 含数字与布尔
- **WHEN** `enum` 为 `[1, 2, true]`
- **THEN** 发往上游的 `enum` MUST 为对应的字符串序列
- **THEN** 上游 MUST NOT 因 enum 类型而返回 400

#### Scenario: enum 含对象
- **WHEN** `enum` 中出现对象或数组元素
- **THEN** 该 `enum` MUST 从 schema 中整体删除
- **THEN** 其所在属性的其余约束 MUST 保留

#### Scenario: enum 全为字符串
- **WHEN** `enum` 已全部是字符串
- **THEN** MUST 原样保留，行为与移植前一致

### Requirement: Antigravity 兼容模式必须把 token 上限封顶到上游上限
系统 SHALL 在把 Chat Completions 的 `max_tokens` / `max_completion_tokens` 映射为上游请求的
最大输出 token 时，封顶到 Antigravity 兼容端点上限 `64000`。客户端请求更大值 MUST 被封顶，
MUST NOT 使整个请求失败。

#### Scenario: 客户端请求超过上限
- **WHEN** 客户端传入 `max_tokens = 200000`
- **THEN** 出站请求的最大输出 token MUST 为 `64000`
- **THEN** 请求 MUST 成功进入上游而不是整体失败

#### Scenario: 客户端请求在上限内
- **WHEN** 客户端传入 `max_completion_tokens = 8192`
- **THEN** 出站值 MUST 为 `8192`

#### Scenario: 客户端未指定或为非正值
- **WHEN** 两个字段都缺失，或值为 0 / 负数
- **THEN** MUST 保持既有默认行为，MUST NOT 写入封顶值

### Requirement: Antigravity 旧 Sonnet 兼容别名迁移 MUST NOT 改写显式 canonical 选择
系统 SHALL 把 Antigravity 的旧 Sonnet 兼容别名（`-thinking` 后缀、日期后缀）映射到
`claude-sonnet-4-6`，同时 MUST 保持 canonical 的 `claude-sonnet-4-5` 原样透传。
Antigravity 连接测试的默认模型 MUST 使用具名常量而非字面量。

#### Scenario: 客户端显式请求 canonical 4.5
- **WHEN** 客户端请求 `claude-sonnet-4-5`
- **THEN** 上游模型 MUST 仍为 `claude-sonnet-4-5`
- **THEN** MUST NOT 被改写为 4.6

#### Scenario: 客户端请求旧兼容别名
- **WHEN** 客户端请求 `claude-sonnet-4-5-thinking` 或 `claude-sonnet-4-5-20250929`
- **THEN** 上游模型 MUST 为 `claude-sonnet-4-6`

#### Scenario: 账号连接测试未指定模型
- **WHEN** 管理端对 Antigravity 账号发起连接测试且未指定模型
- **THEN** MUST 使用 `claude-sonnet-4-6`

### Requirement: OAuth 图片生成必须逐字使用用户 prompt
系统 SHALL 在构造图片生成的 Responses 请求时下发明确的 `instructions`，要求模型逐字使用用户
prompt，不得润色、翻译、增删画面细节或规范化标点。`instructions` MUST NOT 为空串。

#### Scenario: 非英文且含特定标点的 prompt
- **WHEN** 用户提交带引号与中文的图片 prompt
- **THEN** 出站请求的 `instructions` MUST 为逐字使用 prompt 的指令
- **THEN** 用户 prompt 本身 MUST 原样进入 input 文本

### Requirement: Grok 出站请求必须使用官方 CLI User-Agent
系统 SHALL 在所有 Grok 出站路径使用统一的官方 CLI User-Agent 来源，MUST NOT 保留占位
User-Agent `sub2api-grok/1.0`，且该占位常量 MUST 从代码中删除。账号显式配置的自定义
User-Agent MUST 继续优先。

#### Scenario: Grok OAuth 账号且未配置自定义 UA
- **WHEN** 通过 Grok OAuth 账号发起原生 Chat Completions 直转或观测模型同步
- **THEN** 请求 User-Agent MUST 来自统一的官方 CLI 身份
- **THEN** MUST NOT 为 `sub2api-grok/1.0`

#### Scenario: 账号配置了自定义 UA
- **WHEN** 账号显式设置了 User-Agent
- **THEN** MUST 使用账号配置值

### Requirement: 降级与还原之间必须保持工具调用 item ID 的类型一致
系统 SHALL 在把 `custom_tool_call` / `tool_search_call` 降级为 `function_call` 协议时，
把 item ID 的前缀换成目标类型要求的前缀并**保留后缀**；还原为原类型时反向换回。
无法识别前缀的 ID MUST 保持原样，MUST NOT 猜测。流式路径 MUST 区分「上游使用的 ID」与
「发给客户端的 ID」，使后续上游事件仍能匹配到同一次调用。

#### Scenario: 客户端历史回放到会校验 ID 的上游
- **WHEN** 客户端历史中的 `custom_tool_call` 携带上游返回的 `fc_` 前缀 ID
- **THEN** 还原给客户端的 item ID MUST 为 `ctc_` 前缀
- **THEN** 下一轮把该历史回放到上游 MUST NOT 触发
  `Invalid 'input[N].id' ... Expected an ID that begins with 'ctc'`

#### Scenario: 降级时的 ID 处理
- **WHEN** 携带 `ctc_` 或 `tsc_` 前缀 ID 的项被降级为 `function_call` 协议
- **THEN** 出站 item ID MUST 为同后缀的 `fc_` 前缀 ID
- **THEN** `call_id` MUST 保持不变，仍作为调用与输出的配对键

#### Scenario: tool_search 调用的还原
- **WHEN** 上游以 `function_call` 形式回一个 tool-search 代理调用
- **THEN** 还原后的类型 MUST 为 `tool_search_call`
- **THEN** item ID MUST 为 `tsc_` 前缀

#### Scenario: 流式事件的 ID 一致性
- **WHEN** 流式还原发出自定义工具输入的 delta 与 done 事件
- **THEN** 事件中的 item ID MUST 是发给客户端的那个 ID
- **THEN** 内部按上游 ID 与 `call_id` 的匹配 MUST 仍然成立

#### Scenario: 未知前缀的 ID
- **WHEN** item ID 既不是 `fc_` / `ctc_` / `tsc_` 任一前缀
- **THEN** MUST 保持原值不变
