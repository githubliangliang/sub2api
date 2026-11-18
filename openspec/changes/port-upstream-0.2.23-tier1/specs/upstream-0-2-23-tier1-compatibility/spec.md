## ADDED Requirements

### Requirement: F01 Responses 密封调用内容

函数调用的 delta 回退重建 SHALL 保留 encrypted_content。

#### Scenario: 选定既有路径

- **WHEN** Responses 密封调用内容对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证terminal/item.done/delta 三路比对调用 ID、密封内容和原参数。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 保留上一轮已实现的 namespace、encrypted_function_args 与参数完成事件语义。

### Requirement: F02 对话中途 system 权限

Chat/Anthropic 中途 system 消息转 Responses SHALL 保持 developer 权限。

#### Scenario: 选定既有路径

- **WHEN** 对话中途 system 权限对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证首部系统提示及中途 hook 文本、工具历史的权限与顺序。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 不把用户文本或工具结果提升为系统权限。

### Requirement: F03 JSON Schema 默认 strict

Chat json_schema 未设置 strict 时转 Responses SHALL 显式写 false。

#### Scenario: 选定既有路径

- **WHEN** JSON Schema 默认 strict对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证缺省/true/false/null、schema 原值和普通 text/json_object 对照。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 不覆盖客户端显式值，不新增结构化输出能力。

### Requirement: F04 Codex HTTP 外部上下文

无对应调用的 agent browser_info/goal_objective/untrusted_context 输出 SHALL 转为不可信用户文本。

#### Scenario: 选定既有路径

- **WHEN** Codex HTTP 外部上下文对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证三种上下文及已有 page 快照、空输出、重复成员、真实调用、未知名称。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 保留 page 快照严格校验和 user 权限，不能吞正常 function_call_output。

### Requirement: F05 Claude per-turn beta

任一已支持的 per-turn/mid-conversation beta SHALL 保留消息级 output_config。

#### Scenario: 选定既有路径

- **WHEN** Claude per-turn beta对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证API Key/OAuth、空 system 控制消息、有正文消息和顶层 output_config。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 无相关 beta 仍执行既有净化，不能放开全部消息字段。

### Requirement: F06 count_tokens 请求头唯一性

Anthropic count_tokens 透传 SHALL 以原始大小写无关读取/写入避免重复默认头。

#### Scenario: 选定既有路径

- **WHEN** count_tokens 请求头唯一性对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证真实 HTTP 线格式大小写/多值、覆写和缺省头各只有预期值。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 保持账号鉴权、beta 净化与覆写的最终优先级。

### Requirement: F07 输出后的兼容流账号副作用

OpenAI Chat/Messages 已输出后的流错误 SHALL 执行既有账号/模型错误副作用。

#### Scenario: 选定既有路径

- **WHEN** 输出后的兼容流账号副作用对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证文本/工具输出后 401/429/overload、模型规则、普通请求错误与取消。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 已输出后不跨账号重放；沿用模型作用域，不扩大整号冷却。

### Requirement: F08 既有 Sol Ultra Fast 倍率

已支持的 GPT-6.1 Sol Ultra Fast SHALL 使用冻结 tag 的6倍模型倍率。

#### Scenario: 选定既有路径

- **WHEN** 既有 Sol Ultra Fast 倍率对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证Sol/Astra、别名、HTTP/WS结算、Fast2倍、Free Fast、自定义0和长期上下文。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 只摘既有型号的倍率修复；强制 Ultra Fast 分组、新渠道倍率和 GPT-6.2 不随行。

### Requirement: F09 登录身份回填空指针

登录邮箱身份回填失败返回 nil 时 SHALL 安全跳过默认设置判断。

#### Scenario: 选定既有路径

- **WHEN** 登录身份回填空指针对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证身份创建/查询失败、既有身份、回填/首次绑定来源与正常登录。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 不引入新的注册、赠额或支付策略。

### Requirement: F10 5h 滚动保留有效周用量

Claude 5h 窗口滚动 SHALL 仅清除已过期的周读数。

#### Scenario: 选定既有路径

- **WHEN** 5h 滚动保留有效周用量对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证未来/过期/缺 reset、仅7d响应、秒/毫秒重置和样本时间。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 只清5h和过期周值；Fable 模型/新额度接口仍独立按需。

### Requirement: F11 账号5h窗口上下界

主动读取更新5h窗口 end 时 SHALL 同步 start=end-5h。

#### Scenario: 选定既有路径

- **WHEN** 账号5h窗口上下界对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证正常滚动、毫秒格式读入、outbox传播和新窗口费用查询。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 账号窗口修复不改变用户平台固定 quota 窗口、软删除或重置策略。

### Requirement: F12 首输出临时文件不可用

首输出暂存无法创建临时文件时 SHALL 在同一容量上限内回退内存。

#### Scenario: 选定既有路径

- **WHEN** 首输出临时文件不可用对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证无权限/目录不可用、commit/drop/取消、上限拒绝及清理。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 保留8MiB stage上限、无具名明文残留和预算释放。

### Requirement: F13 Claude 默认测试模型维护

内置 Claude 默认测试模型 SHALL 与冻结 tag 的 claude-sonnet-4-6 一致。

#### Scenario: 选定既有路径

- **WHEN** Claude 默认测试模型维护对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证默认账号测试使用已有4-6；显式模型和其它平台默认值不被覆盖。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 属于默认值维护，不新增模型目录/定价，也不宣称独立市场验证。

### Requirement: F14 Grok 媒体扁平错误

Grok 既有媒体错误 SHALL 保留 xAI 扁平 error/code，保持脱敏与可配置错误规则。

#### Scenario: 选定既有路径

- **WHEN** Grok 媒体扁平错误对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证嵌套优先、扁平字符串/数字code、内容策略、自定义规则和未知对象。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 仅 Grok 扩展提取器；不改变其它平台的模型不存在/停号判断。

### Requirement: F15 Gemini 非数组 items

既有 Gemini schema 清理 SHALL 删除非 array 的 items，并保持合法数组 schema。

#### Scenario: 选定既有路径

- **WHEN** Gemini 非数组 items对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证嵌套对象/数组、非法items、工具schema及实例 default/enum 的保护。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 不把本地已经处理的 tuple 情形描述成新缺陷，不引入新 Gemini 模型。

### Requirement: F16 白名单精确模型输入

模型白名单添加操作 SHALL 拒绝含星号的条目并保留输入草稿。

#### Scenario: 选定既有路径

- **WHEN** 白名单精确模型输入对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证点击/Enter、首尾空白、纠正后添加、重复/映射冲突提示。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 不引入 glob 或透传白名单调度新开关。

### Requirement: F17 渠道 Bedrock 兼容开关类型

渠道编辑 SHALL 用 boolean 保存/清除 bedrock_cc_compat，并保留其它 features。

#### Scenario: 选定既有路径

- **WHEN** 渠道 Bedrock 兼容开关类型对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证创建/编辑/重开/显式false/移除Anthropic区段及其它features。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 保留本地 run_mode/菜单限制，不新增平台或商业配置。

### Requirement: F18 调度快照端点能力

调度快照 SHALL 保留已有 openai_capabilities 声明并据此过滤既有端点。

#### Scenario: 选定既有路径

- **WHEN** 调度快照端点能力对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证chat/embeddings/Responses生图意图、未配置、空配置、hydrate及WS读取。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 来源标题提及 Seedance，但本地现有能力消费者独立适用；不引入 Seedance。

### Requirement: F19 设置注入 HTML 与 CSP nonce

设置注入 HTML SHALL 禁止共享/浏览器缓存，携带请求 nonce 时不得按旧 ETag 返回304。

#### Scenario: 选定既有路径

- **WHEN** 设置注入 HTML 与 CSP nonce对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证两请求不同nonce、设置变更/回退、无nonce条件请求、静态hash资源。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 不承诺所有浏览器 bfcache 行为；仅约定响应头和 nonce/HTML 一致性。

### Requirement: F20 Chat 客户端 metadata.user_id

API Key 的 Chat转Anthropic SHALL 补回已识别格式的客户端 metadata.user_id。

#### Scenario: 选定既有路径

- **WHEN** Chat 客户端 metadata.user_id对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证合法JSON/legacy会话ID、空值/错误类型、两路径和OAuth负对照。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 只保留识别到的user_id字符串原字节，OAuth仍使用账号拟态身份。

### Requirement: F21 完成转发后的粘性绑定

完成转发后的粘性写入 SHALL 不随客户端取消，且受5秒截止限制。

#### Scenario: 选定既有路径

- **WHEN** 完成转发后的粘性绑定对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证完成后断开仍写入、超时返回、换号不覆盖原绑定和正常请求。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 保留本地 profit admission/既有绑定条件；只脱离这一收尾写入。
