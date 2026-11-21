## ADDED Requirements

### Requirement: F01 Codex 文献搜索 403 的账号副作用

alpha/search 工具级403 SHALL 允许既有换号流程且不冷却或永久置错健康账号；明确账号/工作区停用的403继续执行现有账号策略。

#### Scenario: 既有消费者满足目标行为

- **WHEN** OAuth与PAT回退的 literature-search 403、空403、deactivated_workspace、401/404/405和429；记录仓储副作用次数。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 不能把所有403全局忽略；仅限定搜索工具路径，不放宽普通Responses鉴权与访问状态错误。

### Requirement: F02 OpenAI 408 上游超时换号

上游 HTTP408 SHALL 进入既有有界换号流程。

#### Scenario: 既有消费者满足目标行为

- **WHEN** Responses/图片408后下一账号成功、换号上限、父取消、普通400和已输出后不重放。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 不新增重试策略；保留一次idle换号预算及已输出边界。

### Requirement: F03 Codex IDE 与独立上下文

合法独立 agent 上下文 SHALL 转为不可信 user 文本，保留各类型的空值规则。

#### Scenario: 既有消费者满足目标行为

- **WHEN** IDE非空、untrusted_content允许空串、proposed_plan允许空白非空、null/数组拒绝、真实call_id保持、重复键保护。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 不提升到system/developer；未知工具、真实调用及page快照严格校验保持。

### Requirement: F04 Responses 每轮请求推理强度

请求推理强度日志 SHALL 优先记录本轮原始input最后一个有效configuration_update的effort，并与最终有效强度独立。

#### Scenario: 既有消费者满足目标行为

- **WHEN** HTTP/Chat/Messages/WS桥接与原生WS，多个update、空/无effort、JSON转义、none/max、映射/封顶、下一轮省略与SQLite230读写。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 保留d2488b110的原始载荷取值、output_config支持、缺省NULL及none/max；不重复加入迁移230或撤销本地已修的bridge采集。

### Requirement: F05 既有 Claude 账号测试 temperature

既有拒绝temperature的Claude家族在连接测试中 SHALL 省略该字段。

#### Scenario: 既有消费者满足目标行为

- **WHEN** Opus4.7/4.8/5.5及别名/Bedrock前缀，Sonnet4.6与显式测试模型对照；实际本地HTTP抓取请求。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 不扩充模型目录/价卡；未选Fable/Mythos helper不作为前置。

### Requirement: F06 兼容转换保留 stop 与原始工具定义

既有兼容降级路径 SHALL 保留客户端stop序列与可转发的原始动态工具定义。

#### Scenario: 既有消费者满足目标行为

- **WHEN** stop字符串/数组/省略；Chat→Responses→Chat、原生Responses/OAuth的字段白名单、API Key的动态工具、用户数值和工具顺序。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 只选已有通用API Key消费者；不引入原生Kimi平台或缺失的Anthropic-native适配器。

### Requirement: F07 图片 SSE usage 单行载荷

转发图片的completed事件 SHALL 把usage压为单行JSON后写入SSE data行。

#### Scenario: 既有消费者满足目标行为

- **WHEN** 多行/紧凑/空/非法usage，实际TCP SSE事件解析、图像结果与精确token值，断开后结算去重。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 不改变图片模型、价格、结果数量或SSE行上限。

### Requirement: F08 Gemini 最后工具轮的尾随内容

最后一个Gemini工具结果turn的尾随非functionResponse内容 SHALL 保序拆成后续user turn。

#### Scenario: 既有消费者满足目标行为

- **WHEN** 单/多个工具结果加文本/图片、纯工具结果、纯文本、历史非末尾turn与parallel工具结果媒体顺序。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 不导入新Gemini模型、完整thinking策略或改变已完成的工具媒体顺序规则。

### Requirement: F09 Antigravity 原生 SSE 注释帧

支持注释的客户端 SHALL 保留上游SSE注释及其分隔；不支持注释的既有SDK继续滤掉注释。

#### Scenario: 既有消费者满足目标行为

- **WHEN** 注释/data混排、CRLF、空行、重写data、keepalive、go-genai/python-genai及中断usage。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 不通过关闭默认ping回避错误；保持120s idle、40MiB行上限和计费drain。

### Requirement: F10 Grok 媒体输入错误的 400 分类

既有Grok媒体本地输入校验失败 SHALL 返回400并保留可用校验信息。

#### Scenario: 既有消费者满足目标行为

- **WHEN** JSON/multipart超出当前3图上限、合法输入、内容策略、脱敏及上游可配置错误策略。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 本簇保持grokMediaMaxEditSourceImages=3；扩为5张是独立第三档，未获选择不随修复进入。

