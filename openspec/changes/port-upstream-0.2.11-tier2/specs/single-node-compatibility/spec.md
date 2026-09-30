## ADDED Requirements

### Requirement: S01 Sonnet 5.5多入口兼容
系统 MUST 满足：Sonnet 5.5目录、定价、effort、thinking签名与工具beta在本地已有入口中一致。

#### Scenario: S01 between_tools与验证
- **WHEN** 映射目标为Sonnet 5.5，指定between_tools及low/medium/high effort
- **THEN** 保留模式；disabled/enabled、between_tools下xhigh/max或不支持的字段、强制工具及非默认采样设置明确拒绝

#### Scenario: S01 签名历史
- **WHEN** Sonnet或Opus 5.5多轮历史中有无效thinking签名
- **THEN** 清理不连续的thinking/redacted_thinking历史，保留普通正文与工具内容；完整签名历史保留

#### Scenario: S01 稳定工具beta
- **WHEN** Sonnet使用稳定computer/browser toolset，账号覆写还包含旧fine-grained beta
- **THEN** 最终出站移除该冲突token，保留其它beta；普通工具与其它模型不受影响

#### Scenario: S01 价格与目录
- **WHEN** 请求Sonnet 5.5别名且没有显式渠道价
- **THEN** 匹配该模型的普通及5m/1h缓存价格；显式渠道价仍优先，Codex/前端目录能表达模型

### Requirement: S02 GPT-6.1 Sol及账号套餐元数据
系统 MUST 满足：现有网关路径识别GPT-6.1 Sol并一致执行请求能力、模型目录与价格规则，账号套餐显示保留SKU身份。

#### Scenario: S02 映射后能力检查
- **WHEN** 公开alias映射到GPT-6.1 Sol，输入none/minimal effort或disabled thinking
- **THEN** 在有损转换前返回明确400；支持的effort不被静默改档，max保留

#### Scenario: S02 纯Chat账号工具
- **WHEN** GPT-6.1 Sol请求包含tools/functions，而账号只支持Chat Completions
- **THEN** 明确拒绝不支持的组合，不能丢弃工具后转发；无工具正常请求仍可用

#### Scenario: S02 官方目录完整性
- **WHEN** 生成GPT-6.1 Sol的Codex descriptor并映射公开slug
- **THEN** 保留官方未知字段和nested model_messages，路由slug由本地配置控制，显式账号元数据保持优先

#### Scenario: S02 价格隔离
- **WHEN** GPT-6.1 Sol走普通或priority计价且账号长上下文gate不同
- **THEN** 采用对应模型价卡并保留本地gate和显式渠道价；不改变Astra Ultrafast能力或价格

#### Scenario: S02 套餐SKU
- **WHEN** 两个不同canonical SKU使用相同展示标签，或账号返回未知SKU
- **THEN** 选项仍可区分并保留原值，未知值不被删除；其它平台套餐不套OpenAI命名

### Requirement: S03 流式用量输出与计费一致
系统 MUST 满足：现有Anthropic到Responses/Chat桥在收到真实usage时转发，并将权威prompt总量与缓存桶统一到计费口径。

#### Scenario: S03 显式零与缺失
- **WHEN** Chat客户端没有include_usage，上游分别发送显式零usage、缺失usage、null usage
- **THEN** 显式零对象仍转发；缺失/null不得因中间转换器合成数据而伪造收到usage

#### Scenario: S03 权威总量
- **WHEN** 上游给出prompt total或hit/miss及缓存信息
- **THEN** 归一为互斥input/cache桶，终态Responses/Chat usage与计费记录一致

#### Scenario: S03 后到缓存
- **WHEN** 早前事件有input，后续只有独立cache字段且没有权威总量
- **THEN** 不推断从早前input扣减缓存，避免重复扣减；使用现有消费者验证

#### Scenario: S03 终止与取消
- **WHEN** 上游完成或客户端断开后还有usage可读取
- **THEN** 遵守既有终态/drain策略，不因归一化重复落库或重放请求

### Requirement: S04 Antigravity首内容前保活边界
系统 MUST 满足：兼容流等待首内容时15秒发送注释保活，并在2分钟无语义内容时终止；已提交响应后的失败通过单次SSE错误报告。

#### Scenario: S04 慢首包
- **WHEN** 首内容晚于15秒但早于2分钟
- **THEN** 先收到注释ping，随后正常内容；首token耗时从真实内容计

#### Scenario: S04 已提交后失败
- **WHEN** ping已提交后上游空流、读取失败或等待超时
- **THEN** 收到一个协议正确的SSE错误，不尝试换号重放

#### Scenario: S04 硬截止
- **WHEN** 上游持续发送只有注释或signature的事件
- **THEN** 2分钟无语义内容仍终止，不能无限续期

#### Scenario: S04 未提交失败与取消
- **WHEN** ping之前空流，或客户端中途断开
- **THEN** 未提交时保留既有failover语义；断开后不持续写ping，reader和timer正常收尾

### Requirement: S05 Composite模型归属与多轮WS路由
系统 MUST 满足：OpenAI两种调度执行公开alias归属检查，WS按responses/any解析路由并隔离公开模型与出站映射。

#### Scenario: S05 账号归属
- **WHEN** account_model路由发布alias，有一个拥有显式映射的账号和一个不拥有的账号
- **THEN** advanced与legacy均只选owner，非owner不因优先级、粘性或传输条件入选

#### Scenario: S05 多轮WS
- **WHEN** 首帧公开alias路由到实际模型，后续重复或省略公开alias
- **THEN** 准入看到客户端原始候选，出站使用映射值，响应保持公开身份

#### Scenario: S05 换公开模型
- **WHEN** 连接中后续帧请求另一公开模型
- **THEN** 要求重连后重新解析和选账号，不复用旧路由/权限

#### Scenario: S05 拒绝与回退
- **WHEN** 路由未匹配、解析失败或指向不支持的目标
- **THEN** 按既有错误/关闭协议拒绝，不把公开别名直接交无归属账号

### Requirement: S06 Codex远程目录与文件兼容
系统 MUST 满足：提供Codex 0.156.0+远程目录和旧客户端本地文件选择，采用v0.2.11最终行为。

#### Scenario: S06 模式切换
- **WHEN** OpenAI HTTP/WS或既有路由分组切换remote/file
- **THEN** remote只在provider写model_catalog_url，file只写model_catalog_json，旧配置不残留

#### Scenario: S06 目录大小
- **WHEN** 取得UTF-8原始响应超过1MiB
- **THEN** 提示并切换文件模式，完整目录可下载且不截断；等于阈值仍可remote

#### Scenario: S06 URL认证与平台切换
- **WHEN** base URL为根、带/v1、带子路径或尾斜杠，期间切换分组
- **THEN** /models路径正确且无Key泄漏；迟到请求不能覆盖新分组，认证沿用现有配置

#### Scenario: S06 旧客户端与Windows
- **WHEN** 用户选择文件模式并复制Windows配置
- **THEN** 路径使用既有~/规则，旧客户端不需要远程目录支持，CC Switch行为保持
