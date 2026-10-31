# existing-gateway-correctness Specification

## Purpose

定义个人单节点网关现有协议、调度、用量统计、管理界面与工具 schema 的正确性边界，覆盖 Codex、Claude Code、Grok 和既有 Gemini 路径，并保留 SQLite、miniredis 与 simple mode 约束。

## Requirements

### Requirement: F01 Codex 工作区导入身份隔离
系统 SHALL 满足以下既有功能行为：以 workspace/account 与 user 的组合区分完整 OAuth 身份；同一工作区重复导入更新原账号，不覆盖其它工作区。

#### Scenario: 修复受影响路径
- **WHEN** 既有路径收到本簇涉及的请求、响应或配置组合
- **THEN** 两个工作区同一用户各自创建；再次导入只更新对应账号；只有 token、缺 user/account、共享 account 的成员身份分别覆盖。

#### Scenario: 保留个人版及相邻路径
- **WHEN** 请求不符合修复条件，或包含来源PR的其它新能力
- **THEN** 保留 accessToken-only 升级和 account 兜底；长度编码防拼接碰撞，不改变凭据存储格式。

### Requirement: F02 Codex 前导元数据不阻止重试
系统 SHALL 满足以下既有功能行为：排队和审查元数据保持首个语义输出之前的状态，后续容量错误仍可安全重试，首输出超时仍有效。

#### Scenario: 修复受影响路径
- **WHEN** 既有路径收到本簇涉及的请求、响应或配置组合
- **THEN** HTTP SSE、透传和 WS-HTTP 桥在 queued/metadata 后遇到 overload 可重试；已有有效输出后不重试。

#### Scenario: 保留个人版及相邻路径
- **WHEN** 请求不符合修复条件，或包含来源PR的其它新能力
- **THEN** 真实文本、工具参数和未知事件仍按现有输出语义；保留本地 SSE 注释写出后的安全换号规则。

### Requirement: F03 Grok Fast 用量成本价卡
系统 SHALL 满足以下既有功能行为：既有 Grok 模型映射/透传请求按 Fast 独立价卡计算；短上下文 input/cache/output 每百万为 4/1/12，达到 200k 后为 6/1.5/18（来源补丁值）。

#### Scenario: 修复受影响路径
- **WHEN** 既有路径收到本簇涉及的请求、响应或配置组合
- **THEN** fast/build-fast/provider 前缀、199999/200000 边界、缓存桶、动态资源与离线 fallback 一致；普通 4.7 不受影响。

#### Scenario: 保留个人版及相邻路径
- **WHEN** 请求不符合修复条件，或包含来源PR的其它新能力
- **THEN** 不扩展模型目录、平台或购买功能；普通 Grok 与显式渠道价格覆盖不变；历史账本不回写。

### Requirement: F04 Cloudflare 1010 不处罚账号
系统 SHALL 满足以下既有功能行为：识别明确的 Cloudflare 1010 或 browser_signature_banned，所有既有平台跳过账号处罚，保留请求错误和换号处理。

#### Scenario: 修复受影响路径
- **WHEN** 既有路径收到本簇涉及的请求、响应或配置组合
- **THEN** 五个已有平台的 HTML/JSON/嵌套 Cloudflare 样本不调用 SetError/限流计数；真实403继续原有策略。

#### Scenario: 保留个人版及相邻路径
- **WHEN** 请求不符合修复条件，或包含来源PR的其它新能力
- **THEN** 普通供应商 403、任意数字 1010、与 Cloudflare 无关业务错误不能被豁免；不引入 OpenCode/Command Code。

### Requirement: F05 Gemini 缓存命中用量非负
系统 SHALL 满足以下既有功能行为：仅缓存计数或缓存大于 prompt 时非缓存 input_tokens 下限为0，同时保留缓存、输出和图片用量。

#### Scenario: 修复受影响路径
- **WHEN** 既有路径收到本簇涉及的请求、响应或配置组合
- **THEN** 流式 message_start/更新与非流式在 prompt缺失、0、小于/等于/大于cached 时保持非负且桶语义一致。

#### Scenario: 保留个人版及相邻路径
- **WHEN** 请求不符合修复条件，或包含来源PR的其它新能力
- **THEN** 正常 prompt>=cached 的计数不变；不抹掉缓存命中，不引入显式缓存CRUD或存储计费。

### Requirement: F06 Gemini 工具 Schema 常量和数据边界
系统 SHALL 满足以下既有功能行为：字符串const转为单值enum，与已有enum求交；移除不支持的非字符串const和注解，区分Schema节点、属性名和数据值。

#### Scenario: 修复受影响路径
- **WHEN** 既有路径收到本簇涉及的请求、响应或配置组合
- **THEN** 字符串/数值const、enum交集、组合Schema、properties同名键、数据对象、exclusiveMinimum同时验收；原schema不被修改。

#### Scenario: 保留个人版及相邻路径
- **WHEN** 请求不符合修复条件，或包含来源PR的其它新能力
- **THEN** 保留上一轮 integer exclusiveMinimum 修复；不得误删名为const/examples/$schema的合法属性，也不得改写default等数据内部值。

### Requirement: F01 删除 Key 后完成请求仍结算

系统 SHALL 满足以下已有功能正确性契约：请求执行中 Key 被软删除，余额/账号配额及 dedup 按既有模式正常落库，已删除 Key 自身计数跳过。

#### Scenario: 目标行为

- **WHEN** 使用删除 Key 后完成请求仍结算涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 请求执行中 Key 被软删除，余额/账号配额及 dedup 按既有模式正常落库，已删除 Key 自身计数跳过。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 非 Key-not-found 的数据库错误仍失败回滚；重放相同请求不重复计数；simple mode 原本未启用的余额扣减不被开启。

### Requirement: F02 Chat 请求角色与具名工具选择

系统 SHALL 满足以下已有功能正确性契约：具名 function choice 转为 Responses 的顶层 name，developer 保持 developer。

#### Scenario: 目标行为

- **WHEN** 使用Chat 请求角色与具名工具选择涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 具名 function choice 转为 Responses 的顶层 name，developer 保持 developer。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** auto/none/required 及已经是 Responses 格式的 choice 不变；空名称和普通 system/user 消息保持原语义。

### Requirement: F03 Anthropic 缓冲索引与 thinking 输出

系统 SHALL 满足以下已有功能正确性契约：负数内容块索引被忽略；thinking 块即使内容/签名为空也有 thinking、signature 字段。

#### Scenario: 目标行为

- **WHEN** 使用Anthropic 缓冲索引与 thinking 输出涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 负数内容块索引被忽略；thinking 块即使内容/签名为空也有 thinking、signature 字段。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 普通文本/工具块不增加无关字段；正索引输出及后续 signature_delta 保真。

### Requirement: F04 Claude 拟态缓存 TTL 与 Sonnet 采样参数

系统 SHALL 满足以下已有功能正确性契约：缓存块数量限制后将较早 ephemeral 断点升至后续 1h；Sonnet 仅有 top_p 时不注入 temperature。

#### Scenario: 目标行为

- **WHEN** 使用Claude 拟态缓存 TTL 与 Sonnet 采样参数涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 缓存块数量限制后将较早 ephemeral 断点升至后续 1h；Sonnet 仅有 top_p 时不注入 temperature。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 已合法 TTL 顺序/无断点请求不变；Opus/其它模型原规则和四块上限保留，不顺带开放 Haiku 5.5。

### Requirement: F05 Grok 空完成事件与 SSO consent

系统 SHALL 满足以下已有功能正确性契约：Grok 空 response.completed 按现有空流失败规则处理；SSO 同一 flow 从合法 consent 页读取 hidden token 并提交 origin/referer。

#### Scenario: 目标行为

- **WHEN** 使用Grok 空完成事件与 SSO consent涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** Grok 空 response.completed 按现有空流失败规则处理；SSO 同一 flow 从合法 consent 页读取 hidden token 并提交 origin/referer。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 有输出完成事件保持成功；旧页面没有 consent 字段仍兼容；重定向不泄露首跳额外头。

### Requirement: F06 Codex 已有模型元数据与远程发现配置

系统 SHALL 满足以下已有功能正确性契约：remote 模式的生成配置包含发现开关；生成的 GPT-6.1 Sol 默认上下文为来源值 500000。

#### Scenario: 目标行为

- **WHEN** 使用Codex 已有模型元数据与远程发现配置涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** remote 模式的生成配置包含发现开关；生成的 GPT-6.1 Sol 默认上下文为来源值 500000。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** file 模式不加远程发现开关；显式上游 context_window 和用户目录值仍优先；不引入新模型/Ultrafast。

### Requirement: F07 Gemini 工具图片与 Antigravity 图片冷却隔离

系统 SHALL 满足以下已有功能正确性契约：base64 工具结果图片转为 inlineData，工具响应顺序保持；图片模型冷却只写/查对应图片键。

#### Scenario: 目标行为

- **WHEN** 使用Gemini 工具图片与 Antigravity 图片冷却隔离涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** base64 工具结果图片转为 inlineData，工具响应顺序保持；图片模型冷却只写/查对应图片键。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 普通文本工具结果及 Gemini 文本家族冷却不变；不下载远程图片。
- **THEN** 新增临时规则作用范围UI时 SHALL 遵守model-scoped-cooldown的默认、认证和隔离边界，继续保留图片/文本键隔离。

### Requirement: F08 价格、倍率与吞吐显示

系统 SHALL 满足以下已有功能正确性契约：科学计数价格与亚字节速率正确显示，零倍率保留，长上下文徽标不承诺统一倍数；只在空条目首次加入单模型时自动填价。

#### Scenario: 目标行为

- **WHEN** 使用价格、倍率与吞吐显示涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 科学计数价格与亚字节速率正确显示，零倍率保留，长上下文徽标不承诺统一倍数；只在空条目首次加入单模型时自动填价。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 已设价格、已有模型追加、一次多模型加入及缺失倍率默认1保持边界；不改变账本计算规则。

### Requirement: F09 共享弹窗和选择器输入边界

系统 SHALL 满足以下已有功能正确性契约：Esc 只关栈顶；disabled 收起并清理搜索；isComposing 时不处理导航键。

#### Scenario: 目标行为

- **WHEN** 使用共享弹窗和选择器输入边界涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** Esc 只关栈顶；disabled 收起并清理搜索；isComposing 时不处理导航键。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 保留原焦点归还/多弹窗 body 锁和非 IME 键盘操作；不重复移植旧 dialog ID 前置。

### Requirement: F10 管理端错误文本与开关响应

系统 SHALL 满足以下已有功能正确性契约：使用统一解析显示后端原因和合适 fallback；开关状态取服务端返回 enabled。

#### Scenario: 目标行为

- **WHEN** 使用管理端错误文本与开关响应涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 使用统一解析显示后端原因和合适 fallback；开关状态取服务端返回 enabled。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 不显示对象字符串或仅 Axios 通用文本；失败不伪造状态；不改接口和平台目录。

### Requirement: F11 计时器、访问重试与备份卸载

系统 SHALL 满足以下已有功能正确性契约：卸载后回包不重启计时器，回到可见标签立即更新倒计时；暂时查询失败后可再次检查访问权限。

#### Scenario: 目标行为

- **WHEN** 使用计时器、访问重试与备份卸载涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 卸载后回包不重启计时器，回到可见标签立即更新倒计时；暂时查询失败后可再次检查访问权限。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 成功权限结果仍缓存，重复轮询不增生；不启用后台服务、余额通知或批量生图功能。

### Requirement: F12 现有监控健康分与时间/设置边界

系统 SHALL 满足以下已有功能正确性契约：请求样本不足保持未知；规则分钟为整数；时间按本地初始化；自动刷新与手动刷新使用同一 detail 流程；未完成设置加载禁止保存。

#### Scenario: 目标行为

- **WHEN** 使用现有监控健康分与时间/设置边界涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 请求样本不足保持未知；规则分钟为整数；时间按本地初始化；自动刷新与手动刷新使用同一 detail 流程；未完成设置加载禁止保存。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 达样本阈值正常评分；刷新保持用户当前筛选；不启用隐藏 ops 菜单或增加 TPS 字段。

### Requirement: F13 账号测试诊断与分组名显示

系统 SHALL 满足以下已有功能正确性契约：测试列表列显式请求侧映射名并排序；HTTP/后台失败日志有账号、测试及请求关联且先脱敏后截断；账号表完整显示长组名。

#### Scenario: 目标行为

- **WHEN** 使用账号测试诊断与分组名显示涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 测试列表列显式请求侧映射名并排序；HTTP/后台失败日志有账号、测试及请求关联且先脱敏后截断；账号表完整显示长组名。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 无映射返回默认列表；普通 GroupBadge 默认仍截断；不通过日志泄露凭据，不更改测试执行计划。

### Requirement: F14 内嵌前端放行已有 API 别名

系统 SHALL 满足以下已有功能正确性契约：本地已注册 API 别名到达网关；测试使用当前构建实际静态资源。

#### Scenario: 目标行为

- **WHEN** 使用内嵌前端放行已有 API 别名涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 本地已注册 API 别名到达网关；测试使用当前构建实际静态资源。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 正常 SPA 深链接和静态资源保持；不注册上游缺失的视频/语音/Realtime API。

### Requirement: 普通工具参数完成与历史回放必须保持JSON完整性
系统 SHALL 在现有Chat兼容转换中阻止非法JSON的普通function-call被标记为完成；回放中的坏普通调用及其匹配输出 SHALL 配套处理。系统 MUST 保留合法参数、既有空参数默认和custom自由文本语义。

#### Scenario: 上游返回截断普通工具参数
- **WHEN** 非流响应或流终态的普通function-call参数是截断JSON
- **THEN** 客户端 MUST NOT 收到该工具的成功完成项
- **THEN** 已观测usage SHALL 按既有写入政策保留并去重

#### Scenario: 历史包含坏调用及对应输出
- **WHEN** 客户端回放非法普通function-call及匹配call_id输出
- **THEN** 出站Chat请求 MUST NOT 继续转发这对坏历史项
- **THEN** 其它合法对话与工具项 SHALL 保留

#### Scenario: 合法参数与custom输入
- **WHEN** 普通参数是有效JSON但finish_reason为length，或工具属于custom/freeform
- **THEN** 系统 MUST NOT 仅以length或custom输入不是JSON而拒绝该工具
- **THEN** 整体响应的既有incomplete和工具类型语义 SHALL 保留

### Requirement: 自定义工具别名还原必须尊重显式声明与命名空间归属
系统 SHALL 在既有Chat到Responses的buffer与SSE转换中还原无歧义namespace前缀custom别名；显式function与真实namespace子工具 MUST 优先，歧义名称 MUST NOT 通过猜测重写。

#### Scenario: 唯一custom前缀别名
- **WHEN** 已声明custom exec、namespace functions下的wait，且上游返回functions__exec
- **THEN** 输出 SHALL 为custom_tool_call、名称exec及原始custom input
- **THEN** 流added/done与最终output SHALL 使用一致的身份

#### Scenario: 别名与真实声明冲突
- **WHEN** 返回名称对应显式function、真实namespace子工具或多个可能custom归属
- **THEN** 系统 MUST 保留确定的显式/namespace归属或采用既有明确拒绝策略
- **THEN** 系统 MUST NOT 将其擅自还原为另一个custom工具

### Requirement: 最终工具声明必须清理失去前置的defer标记
系统 SHALL 依据最终声明列表中的真实内建tool_search清理孤儿defer_loading；同名普通function MUST NOT 被视为内建前置。该行为 SHALL 覆盖现有通用adapter与Grok请求出口。

#### Scenario: 内建tool_search已降级或移除
- **WHEN** 最终列表只剩普通function等工具且仍有defer_loading字段
- **THEN** 出站工具 SHALL 不再含孤儿defer_loading
- **THEN** 工具schema、选择与历史的其余既有语义 SHALL 保留

#### Scenario: 合法内建tool_search仍存在
- **WHEN** 最终声明仍包含真实内建tool_search
- **THEN** 合法defer_loading SHALL 保留
- **THEN** 系统 MUST NOT 改写共享账号元数据或目录缓存

### Requirement: Responses事件必须序列化零序号并保留事件顺序
系统 SHALL 在现有Responses序列化及本地合成事件中始终写出sequence_number，包括0。已知上游序号 SHALL 保留，本地有序事件 SHALL 递增。

#### Scenario: 第一帧或仅有heartbeat之后的失败
- **WHEN** 第一条数字事件的序号为0，或仅有heartbeat后合成失败
- **THEN** 事件 SHALL 显式包含sequence_number为0
- **THEN** 既有合法终止类型与错误字段 SHALL 保留

#### Scenario: compact和桥接合成事件
- **WHEN** compact生成item/terminal事件，或handler/WS桥合成错误或失败
- **THEN** 所有合成事件 SHALL 携带序号
- **THEN** 已编号流后续合成事件 SHALL 从实际writer/state继续，不统一重置为0

#### Scenario: 既有正值序号
- **WHEN** 事件已有正值序号或流已写出有序事件
- **THEN** 序列化 SHALL 保留现有值和顺序
- **THEN** 未改变的native透传语义与未知字段 SHALL 保留
