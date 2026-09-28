## ADDED Requirements

### Requirement: S01 模型目录和计费别名
系统 MUST 满足：接入 GPT-6 Sol/Luna、Opus 5.5、Grok 4.7 的现有平台目录/能力/价格别名，并将 GPT≥5 按推理模型桥接。

#### Scenario: S01 新模型识别
- **WHEN** 选择GPT-6 Sol/Luna、Opus5.5及Grok4.7并通过已有平台映射转发
- **THEN** 目录/能力/计价使用各自固定v0.2.9定义，保留用户显式model_mapping

#### Scenario: S01 跨代及别名
- **WHEN** 桥接gpt-6-astra与gpt-6-sol；另传claude-opus-5.5或非数字gpt-image模型
- **THEN** 数字代际>=5走推理参数规则；Opus点号别名命中同族；图片等非数字系列不误判为推理文本模型

### Requirement: S02 工具 schema 兼容
系统 MUST 满足：处理根级 union、required:null、tuple/prefixItems，并保留 string const 与 enum 的交集。

#### Scenario: S02 root union
- **WHEN** 对象根anyOf两个分支分别required=[a]和required=[b]，另测试allOf相同分支
- **THEN** 展平为object；anyOf必填取交集，allOf取并集；嵌套属性约束保留可表达部分，明确这不是完整oneOf等价替换

#### Scenario: S02 null字段
- **WHEN** Responses parameters与Messages input_schema含required:null，分别经三种上游分流
- **THEN** 删除非法required成员；普通required数组原样保留，不修改非schema业务内容

#### Scenario: S02 tuple与const
- **WHEN** prefixItems存在且items=false；另给const=x和enum=[x,y]
- **THEN** 转换为Gemini可接受的items对象并去prefixItems，const/enum交集为[x]；冲突枚举不放宽为任意值

### Requirement: S03 工具参数和 PDF 保真
系统 MUST 满足：arguments.done 等于已发送 delta；仅 block_start 携带参数也保留；内嵌 PDF 转为 document/inlineData。

#### Scenario: S03 工具完成参数
- **WHEN** 工具先输出JSON参数delta，再发content_block_stop
- **THEN** arguments.done与累计delta逐字一致

#### Scenario: S03 inline参数
- **WHEN** tool_use仅在block_start携带非空input，无delta；另测试后续又有真实delta
- **THEN** 前者保留完整参数，后者只用delta，不拼接两份JSON

#### Scenario: S03 内嵌PDF
- **WHEN** Responses input_file含非空base64 PDF data URI，经Anthropic桥转Antigravity
- **THEN** 输出document再转Gemini inlineData，保留MIME和数据；file_id-only不自动下载

### Requirement: S04 流结束、取消与错误协议
系统 MUST 满足：terminal 完整发送后结束；先取消单次请求再关响应体；心跳不算语义输出；错误只发一次；Gemini 传输失败换号；客户端取消归类 499。

#### Scenario: S04 terminal不等EOF
- **WHEN** 上游发送完整completed事件及分隔空行后一直不关闭TCP
- **THEN** 客户端及时结束并记录一次完整用量；bare error后仍可读取后续completed

#### Scenario: S04 取消后关体
- **WHEN** 压缩或普通HTTP响应体正在Read，另一协程Close
- **THEN** 先取消本次attempt再关闭，Read被唤醒，无死锁；不取消重试/计费所用父上下文

#### Scenario: S04 心跳与失败
- **WHEN** 只收到keepalive或Anthropic ping，随后出现模型失败
- **THEN** 心跳不算语义TTFT/已输出正文；ping不泄漏到OpenAI协议；客户端已断开时不重放

#### Scenario: S04 单个协议错误
- **WHEN** 本地流扫描错误，或上游error.status=429/401
- **THEN** 输出正确Responses顶层code/message协议一次；按status分类，不误降级为通用502

#### Scenario: S04 Gemini传输故障
- **WHEN** 连接错误发生于普通Gemini请求或countTokens请求
- **THEN** 普通请求进入换号；countTokens保留本地估算；context取消不误冷却账号

#### Scenario: S04 客户端499
- **WHEN** 客户端在排队/选号/上游响应前取消，另测试真实上游错误先发生再取消
- **THEN** 纯取消归499且遵守IgnoreContextCanceled；已有真实上游失败仍保留错误归因，不能伪装200成功

### Requirement: S05 Antigravity 裸模型和空流
系统 MUST 满足：各入口将裸 Gemini 名按 thinking 配置映射；识别特定 SDK 的心跳兼容；MALFORMED_FUNCTION_CALL 空流触发 failover。

#### Scenario: S05 裸名映射
- **WHEN** 裸Gemini模型仅配置-low/-high映射，请求thinkingLevel=low，经原生和兼容入口发送
- **THEN** 选择-low；显式裸名model_mapping存在时优先它；未配置对应变体按规定存在项回退

#### Scenario: S05 SDK心跳
- **WHEN** 客户端提示包含google-genai-sdk与gl-go或gl-python
- **THEN** 不发其不能接受的SSE注释心跳；其它客户端保留现有心跳

#### Scenario: S05 空流重试
- **WHEN** 流只有signature/stop，或MALFORMED_FUNCTION_CALL而无文本/思考/工具内容
- **THEN** 在未断开且未输出有效内容时进入空流failover；真实文本/工具不能重复重试

### Requirement: S06 调度路由和账号可用性
系统 MUST 满足：非高级调度遵守 previous_response 归属；热路径用轻量分组读；API Key 未知模型 401 不误禁账号；暂停 OAuth 仍刷新 token。

#### Scenario: S06 previous_response路由
- **WHEN** 关闭高级调度，previous_response绑定A且A仍满足分组/模型/隐私/传输限制
- **THEN** 选择A并记录正确决策标签；A不兼容时释放取得的槽位后重新选择

#### Scenario: S06 热路径轻量读
- **WHEN** 选号仅需分组privacy配置
- **THEN** 使用不聚合账号计数的轻量分组查询，保持权限/配置结果

#### Scenario: S06 模型401与认证401
- **WHEN** API Key兼容上游用401表示unknown model；另返回真实认证失败
- **THEN** 前者模型级冷却/换号而非永久禁账号，后者保持认证错误处理

#### Scenario: S06 暂停OAuth刷新
- **WHEN** 账号status=active且schedulable=false、refresh token仍有效
- **THEN** 保留在刷新候选；status=error及重试冷却候选仍按既有规则排除

### Requirement: S07 账号成本与图片价继承
系统 MUST 满足：账号成本是否收取长上下文溢价取决于账号 gate；渠道图片价未填继承目录，显式 0 才免费。

#### Scenario: S07 账号gate优先
- **WHEN** OpenAI请求token超过长上下文门槛，账号gate=false、分组售价gate=true
- **THEN** 模型文件账号成本不加长上下文溢价，用户售价仍按分组策略；账号gate=true则账号成本按阶梯

#### Scenario: S07 图片未填和零
- **WHEN** 渠道有价卡但ImageOutputPrice=nil，目录有图片单价；另显式设0
- **THEN** nil继承目录价，0明确免费；无目录图价时沿既有文本价回退，自定义账号价优先级不变

### Requirement: S08 Codex 和 CC Switch 配置
系统 MUST 满足：Codex config 请求正确 /v1；CC Switch保留配置端点（含显式/v1及子路径）、仅去尾斜杠且不自动追加/v1；usage路径恰好一个/v1；Windows catalog用~/。

#### Scenario: S08 CC Switch配置保真
- **WHEN** 输入https://host/x/及https://host/x/v1/生成CC Switch Codex provider
- **THEN** 分别得到https://host/x及https://host/x/v1；只去尾斜杠，不自动增删/v1

#### Scenario: S08 原生与usage路径
- **WHEN** 原生Codex配置使用root URL，usage脚本分别收到带/不带/v1的baseURL
- **THEN** 原生请求命中正确/v1/responses；usage最终均只有一个/v1/usage

#### Scenario: S08 Windows模型清单
- **WHEN** Windows下生成Codex/Grok/WS配置
- **THEN** model_catalog_json使用~/.codex/codex-models.json，不用%userprofile%；其它TOML转义仍合法

### Requirement: S09 代理回退一致性
系统 MUST 满足：代理恢复改变网络身份时失效探针；过期扫描写入前重新核对快照；禁用代理不可作为回退目标。

#### Scenario: S09 恢复代理失效探针
- **WHEN** API Key从回退proxy恢复原proxy且ID变化，extra中有upstream_billing_probe及其它字段
- **THEN** 仅删除probe、保留其它extra；proxy未变化不删probe

#### Scenario: S09 扫描竞争
- **WHEN** 扫描拿到过期快照后管理员续期、停用或修改回退配置，然后执行扫描写入
- **THEN** 条件写入不命中，账号/代理保持新状态；真实未变化过期快照正常处理

#### Scenario: S09 不可用目标
- **WHEN** 回退链中第一个未过期节点status=disabled，后续有active可用节点
- **THEN** 跳过disabled，不把账号指向它；循环/无目标遵守原有处理

### Requirement: S10 Responses 探测未知态
系统 MUST 满足：探测模型不存在时不写入endpoint能力标记（未知或已有值都保留），优先选普通GPT文本模型。

#### Scenario: S10 模型不可用不写能力
- **WHEN** 探测收到400/404且code为model_not_found，现有支持字段分别为未知/true/false
- **THEN** 不调用能力UpdateExtra，三个已有状态均原样保留；不强制清空为unknown

#### Scenario: S10 真正端点不支持
- **WHEN** 收到无模型不可用信号的404或405，另返回可成功探测的文本模型
- **THEN** 前者记录不支持；后者按正常探测结果处理，选模优先普通GPT文本而非图像/辅助模型

### Requirement: S11 公开响应模型别名
系统 MUST 满足：Responses 与 Chat 返回请求方的公开模型别名，同时保留真实上游模型用于记录。

#### Scenario: S11 别名回显
- **WHEN** 请求public-model映射到upstream-model，上游正文model字段返回upstream-model
- **THEN** 公开Responses/Chat模型字段回显public-model，用量记录仍保留upstream-model；其它工具内model字段不被误替换

### Requirement: S12 现有内容审计输入边界
系统 MUST 满足：关键词检查不丢弃客户端 reminder 内文本，尾随 system 不遮蔽当前用户输入。

#### Scenario: S12 reminder关键词
- **WHEN** 最新用户输入的system-reminder标签内含已配置禁词
- **THEN** 关键词检查仍命中，不因reminder过滤绕过；外部语义审核的既有过滤策略不扩大

#### Scenario: S12 尾随system
- **WHEN** Anthropic messages结尾为user禁词文本、system说明；另结尾为assistant/tool
- **THEN** 前者仍检查该user，后者按既有当前用户边界不重复审计历史内容

### Requirement: S13 Antigravity 系统身份兼容
系统 MUST 满足：仅在 Antigravity 转换中移除前导 attribution 并中和前导 SDK 身份。

#### Scenario: S13 限定前导身份
- **WHEN** Antigravity system以x-anthropic-billing-header或Claude Agent SDK身份句开头
- **THEN** 只处理前导元数据/身份句，保留后续用户指令；正文中普通提及Claude和原生Anthropic请求原样保持

### Requirement: S14 共享弹窗与标签收尾
系统 MUST 满足：嵌套弹窗关闭后滚动锁计数正确，模型标签 IME 不提前提交。

#### Scenario: S14 嵌套滚动锁
- **WHEN** 两个弹窗打开后先关闭一个，再关闭最后一个
- **THEN** 第一个关闭后body仍锁定，最后一个关闭后恢复原overflow

#### Scenario: S14 标签IME
- **WHEN** 模型标签输入中文组合中按Enter/Tab，随后完成组合
- **THEN** 组合中不添加半成品标签；结束后按正常规则提交，空Tab仍可移动焦点

### Requirement: S15 Codex WS 上下文切换
系统 MUST 满足：window_id 改变时删除旧 previous_response_id 并清空旧续聊推断锚点。

#### Scenario: S15 切换上下文窗口
- **WHEN** 成功轮次window_id=A，下一轮window_id=B却带旧previous_response_id
- **THEN** 删除旧ID并清空旧expectedPrev；成功处理后才更新last window

#### Scenario: S15 同窗口或缺标记
- **WHEN** 下一轮window_id=A或完全不含窗口标记
- **THEN** 不切断现有续聊链；失败轮次不提交新窗口状态

### Requirement: S16 Codex 请求边界兼容
系统 MUST 满足：映射到 GPT-5.5 时移除不兼容 Lite 标记但保留历史与工具；turn metadata 序列化保留非 ASCII 转义。

#### Scenario: S16 Lite映射
- **WHEN** OAuth类账号请求被映射成gpt-5.5且携带Lite header/WS metadata
- **THEN** 在最终出站请求移除不兼容Lite标记，保留additional_tools/namespace/历史及重放原始输入

#### Scenario: S16 非ASCII metadata
- **WHEN** turn metadata含中文和emoji后需要重写并放入HTTP头
- **THEN** 输出ASCII JSON，中文及UTF-16代理对转义正确，反序列化还原同一文本；纯ASCII不受影响

### Requirement: Preserve personal SQLite deployment
系统 MUST 保留SQLite唯一数据库、可关闭Redis、现有菜单和用量去重；不得引入支付或上游新迁移。

#### Scenario: Scope verification
- **WHEN** 审核相对实施起点的最终diff与真实SQLite运行结果
- **THEN** 无支付/插件/新平台扩展，无已应用迁移修改，成功用量只写入一次
