## ADDED Requirements

### Requirement: composite 分组的 messages 调度开关必须可配置且可持久化
分组字段规范化 SHALL 只对**既不是 openai、也不是 composite** 的平台强制关闭
`allow_messages_dispatch`。composite 分组的该开关 MUST 可以被管理员打开并持久化，
创建与更新两条路径的行为 MUST 一致。

依据：此前规范化对一切非 openai 平台无条件置 `false`，管理台里改了也存不住，
开关在数据层就是死的。

#### Scenario: composite 分组打开开关
- **WHEN** 管理员为 composite 分组把 `allow_messages_dispatch` 设为 true 并保存
- **THEN** 该值 MUST 被持久化为 true
- **THEN** 重新读取该分组 MUST 仍为 true

#### Scenario: 其它非 openai 平台
- **WHEN** 分组平台是 anthropic / gemini / grok / antigravity 等
- **THEN** `allow_messages_dispatch` MUST 仍被强制置为 false

#### Scenario: 附属字段仍被清空
- **WHEN** 分组平台不是 openai（含 composite）
- **THEN** `default_mapped_model` MUST 被清空
- **THEN** `messages_dispatch_model_config` MUST 被重置为零值

#### Scenario: 存量 composite 分组
- **WHEN** 升级前已存在的 composite 分组（其落库值被旧规范化写成了 false）
- **THEN** 升级 MUST NOT 自动把它改成 true
- **THEN** 放行 MUST 需要管理员显式打开该开关

### Requirement: composite 的 /v1/messages 准入由解析到的目标平台决定
`/v1/messages` 的准入判定 SHALL 按以下优先级：分组平台本身是 grok ⇒ 放行；
分组平台是 composite **且**本次请求解析到的目标平台是 grok ⇒ 放行；
其余情况 ⇒ 由分组自己的 `allow_messages_dispatch` 决定。
豁免 MUST NOT 在「分组不是 composite」时依赖解析到的目标平台。

依据：grok 分组的 `/v1/messages` 就是其主要服务形态（原生直通 Claude Code），
不需要开关授权；而解析到 openai 目标时没有理由绕开分组配置。

#### Scenario: composite 解析到 grok 目标
- **WHEN** composite 分组的请求被解析到 grok 目标平台
- **THEN** MUST 放行，无论 `allow_messages_dispatch` 是否打开

#### Scenario: composite 解析到 openai 目标且开关关闭
- **WHEN** composite 分组的请求被解析到 openai 目标平台，且 `allow_messages_dispatch` 为 false
- **THEN** MUST NOT 放行

#### Scenario: composite 解析到 openai 目标且开关打开
- **WHEN** 同上但 `allow_messages_dispatch` 为 true
- **THEN** MUST 放行

#### Scenario: 独立 grok 分组
- **WHEN** 分组平台本身就是 grok
- **THEN** MUST 放行，行为与改动前一致

#### Scenario: 没有 API key 或分组上下文
- **WHEN** 请求上下文里没有 API key 或分组
- **THEN** MUST 沿用改动前的放行行为，本变更 MUST NOT 收紧该路径

### Requirement: 管理台必须为 composite 分组暴露该开关
分组创建与编辑表单 SHALL 在平台为 openai 或 composite 时显示 messages 调度配置区块与
`allow_messages_dispatch` 开关。**逐模型的调度映射配置**（`default_mapped_model` 与映射表）
MUST 仍只对 openai 平台显示，因为规范化对 composite 仍会清空这两个字段。
平台切换到不支持的值时，表单 MUST 重置这些字段。

#### Scenario: 平台选 composite
- **WHEN** 创建或编辑表单的平台是 composite
- **THEN** MUST 显示 `allow_messages_dispatch` 开关
- **THEN** MUST NOT 显示逐模型映射配置区块

#### Scenario: 平台选 openai
- **WHEN** 平台是 openai 且开关打开
- **THEN** MUST 同时显示开关与逐模型映射配置区块

#### Scenario: 平台切换到 anthropic
- **WHEN** 平台从 openai 或 composite 切换到不支持的平台
- **THEN** 表单 MUST 重置 messages 调度相关字段
