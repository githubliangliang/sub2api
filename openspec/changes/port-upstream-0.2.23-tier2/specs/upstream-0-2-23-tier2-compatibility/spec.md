## ADDED Requirements

### Requirement: S01 5m/1h 缓存明细与互斥计费

既有转换、读流及结算 SHALL 保留5m/1h明细和聚合presence，差额按5m计价，输入桶互斥。

#### Scenario: 选定既有路径

- **WHEN** 5m/1h 缓存明细与互斥计费对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证HTTP/普通与透传SSE/缓冲/WS、缓存字段单独更新、矛盾明细、1h费用、SQLite结算+去重。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 只用既有5m/1h列；不导入30m/12h字段、迁移或未选价格策略；自定义0/Fast保留。

### Requirement: S02 Claude 并发 token 轮换恢复

Claude OAuth 的旧token401 SHALL 仅在数据库证明有新有效token时同号重试一次。

#### Scenario: 选定既有路径

- **WHEN** Claude 并发 token 轮换恢复对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证barrier控制旧token在途/新token落库、cache旧值、一次重试、同token/错误类型/取消/预算。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 比对实际发送token；不在401路径刷新/持久化凭据，不绕过重试上限或正常401策略。

### Requirement: S03 SQLite 分组倍率/RPM 原子同步

单次分组倍率或RPM同步 SHALL 在一个事务内清空/删除/写入，失败保留旧配置。

#### Scenario: 选定既有路径

- **WHEN** SQLite 分组倍率/RPM 原子同步对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证事务失败/取消回滚、重复用户last-wins、另一列保留、bind-safe大批次、提交后cache。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 不得复制PG ANY/ALL/unnest/强制类型转换；不更改历史迁移或用户固定quota。

### Requirement: S04 首内容前换号与 keepalive 错误

首内容前仅keepalive输出 SHALL 不阻断安全换号；已提交SSE后的错误必须为SSE事件。

#### Scenario: 选定既有路径

- **WHEN** 首内容前换号与 keepalive 错误对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证created/item-added后失败→成功账号、keepalive前后耗尽、错误规则、取消/终态/槽位释放。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 前奏需沿用8MiB有界暂存，不能原样引入无界strings.Builder；私有账号头不得提前泄露。

### Requirement: S05 空闲超时与一次换号预算

OpenAI 首输出前空闲超时 SHALL 进入最多一次换号预算，计时按最后上游读取截止。

#### Scenario: 选定既有路径

- **WHEN** 空闲超时与一次换号预算对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证持续数据/注释、无内容超时、keepalive已写、两号都超时、已有内容与Grok负对照。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 保留配置的idle默认及既有模型冷却；不要把惩罚扩散到整个池。

### Requirement: S06 Codex additional_tools 声明 ID

Codex additional_tools 的声明ID SHALL 与引用一起保留，仅规范已识别UUID形态。

#### Scenario: 选定既有路径

- **WHEN** Codex additional_tools 声明 ID对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证UUID声明/引用、opaque ID、store=false、完整工具回放、普通callID/UTF8。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 不剥离声明ID、不改普通工具身份；不导入完整账号命名空间或ojson基座。

### Requirement: S07 Grok 服务端搜索调用

Grok xs_/as_ 服务端 custom 搜索 SHALL 不下发为客户端调用，并计入原生搜索次数。

#### Scenario: 选定既有路径

- **WHEN** Grok 服务端搜索调用对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证Grok Chat/Messages/Responses、终态与delta回退、搜索去重、普通custom工具和非Grok前缀。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 仅Grok启用prefix规则；不引入通用工具桥或新搜索平台。

### Requirement: S08 原文模型和真实端点观测

原生Anthropic和Chat/Responses桥 SHALL 把原文模型及冲突带到用量边界，并记录真实/messages端点。

#### Scenario: 选定既有路径

- **WHEN** 原文模型和真实端点观测对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证流/非流/中断、多模型冲突、一次attempt隔离、端点错误日志、SQLite计费回退。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 先观察原文再还原公开名；保留本地价格来源准入及已解析Fast tier，不为模型修复重做tier基座。

### Requirement: S09 Composite 公开模型名还原

Composite 响应 SHALL 还原客户端公开模型名，同时保留原文上游模型用于审计/计费。

#### Scenario: 选定既有路径

- **WHEN** Composite 公开模型名还原对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证原生/拟态/Chat/Responses的流与非流、仅message_start.model、模型冲突及SQLite费用。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 先观测后改名；不得替换正文/工具参数中的同名字符串，不影响非Composite响应。

### Requirement: S10 Anthropic 兼容入口 effort 策略

Anthropic组的Chat/Responses兼容入口 SHALL 执行现有group effort策略并一致返回错误。

#### Scenario: 选定既有路径

- **WHEN** Anthropic 兼容入口 effort 策略对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证deny/map/ceiling、显式非法值、禁用/省略、转换后最终effort和费用、count_tokens邻近。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 保留上一轮合成none/最终校验和GLM通用兼容；不复制缺失原生平台闸门。

### Requirement: S11 Bedrock 错误语义与部分用量

Bedrock SHALL 区分封号/缺用例/请求过滤/模型429，并在中断时返回已观测用量。

#### Scenario: 选定既有路径

- **WHEN** Bedrock 错误语义与部分用量对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证精确400/404、反射字段负对照、429多模型、渠道映射键、5xx过滤、截断SQLite结算。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 请求内容错误不换号；只模型冷却，沿用现有显式规则及SQLite单键更新。

### Requirement: S12 Gemini 并行与分片工具

Gemini流 SHALL 按调用身份而非仅名称划分工具块，保留字符串分片和终态快照。

#### Scenario: 选定既有路径

- **WHEN** Gemini 并行与分片工具对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证同名/同参并行、不同/相同ID、完整首片字符串、重传/快照、无ID歧义及结束事件。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 无ID的字符串快照与新相同调用不可区分时采用来源保守策略，不猜造调用。

### Requirement: S13 Gemini Chat 结构化响应格式

Gemini Chat兼容 SHALL 传递已声明json_object/json_schema格式，并保留schema字面量。

#### Scenario: 选定既有路径

- **WHEN** Gemini Chat 结构化响应格式对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证两个既有认证入口、属性顺序/大数、工具/图片互斥、文本/缺省/无效格式。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 采用responseJsonSchema，不扩展Gemini新模型或cache CRUD；保留资源限制。

### Requirement: S14 Antigravity 兼容思考/等待/空流

既有Antigravity兼容 SHALL 对齐思考配置、按响应头到达锚定等待，空流仅释放对应粘性会话。

#### Scenario: 选定既有路径

- **WHEN** Antigravity 兼容思考/等待/空流对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证slow headers/有效thought/连续ping/empty timeout、无内容已提交、绑定已改变、取消/用量。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 保留既有timeout默认，不导入Carbon等模型和完整新thinking策略；绑定检查需防并发误删。

### Requirement: S15 Gemini 原生稳定前缀会话

Gemini原生会话回退 SHALL 按API Key、模型、tools/system和首用户parts生成稳定哈希。

#### Scenario: 选定既有路径

- **WHEN** Gemini 原生稳定前缀会话对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证首用户/省略role、后续tool回合、模型/Key隔离、空parts、显式客户端ID优先。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 不按发布摘要删除ForceCacheBilling；两个微提交必须以最终PR状态解释。

### Requirement: S16 冷却账号快照成员

快照 SHALL 保留配置启用的完整账号池，并在每次读时判断临时冷却/expiry。

#### Scenario: 选定既有路径

- **WHEN** 冷却账号快照成员对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证快照构建时冷却→自然到期无outbox、真正禁用、simple/混排、取消和未就绪回源。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 不随行#8833的20s stale grace或合并回源；保留revision/retire/单节点缓存契约。

### Requirement: S17 OpenAI 用量观测新鲜度

额度查询 SHALL 区分真实上游观测、陈旧读数和本地未知统计，刷新失败不得伪造新时间。

#### Scenario: 选定既有路径

- **WHEN** OpenAI 用量观测新鲜度对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证主动失败/无quota/显式0、legacy无时间、被动缓存、unknown窗口、UI不误停账号。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 不引入新quota-summary或恢复监控v2整页；缺失Spark专用前置不作为依赖导入。

### Requirement: S18 实际扣除与账号成本统计

按account过滤的模型/端点统计 SHALL 分列actual_cost和account_cost，Gemini账号配额使用账号成本。

#### Scenario: 选定既有路径

- **WHEN** 实际扣除与账号成本统计对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证合成不同标准/实际/账号价、所有筛选/端点来源、管理员/用户redaction与Gemini批量一致。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 只增DTO字段，无新DB列；user API/UI不能泄露账号成本；保持model_source等维度。

### Requirement: S19 价格启动不阻塞远端哈希

已加载本地有效价格后 SHALL 不阻塞启动等待远端hash，并保持可停止的后台同步。

#### Scenario: 选定既有路径

- **WHEN** 价格启动不阻塞远端哈希对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证慢/失败hash不挡启动，下载hash一致、并发同步、Stop/取消、fallback/override。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 不能只加无管理goroutine；不改变已有目录实际正文hash和零覆盖优先级。

### Requirement: S20 上游 TCP 死连接检测

既有标准上游dialer SHALL 使用15s idle/5s interval/3 probes，Linux user-timeout设置失败仍拨号。

#### Scenario: 选定既有路径

- **WHEN** 上游 TCP 死连接检测对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证Linux实际socket选项、错误降级、非Linux编译、proxy/TLS配置和取消。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 30秒是来源参数推算，不是本次实测断线耗时；应用层120s idle保留。

### Requirement: S21 OpenAI 工具 schema 正则适配

现有OpenAI工具schema SHALL 窄处理来源不支持的regexp，并保留实例值与仍有效的schema约束。

#### Scenario: 选定既有路径

- **WHEN** OpenAI 工具 schema 正则适配对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证pattern/patternProperties、不支持转义/组、嵌套编辑、实例default/enum、有效Python形态及非OpenAI。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 不能把来源约899行parser整体覆盖本地窄sanitizer；需测实际适配规模和编辑重叠。

### Requirement: S22 既有流与请求热点窄优化

选定热点优化 SHALL 保持原字节、工具顺序、错误/用量语义和有界生命周期。

#### Scenario: 选定既有路径

- **WHEN** 既有流与请求热点窄优化对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证每个采用hunk做单P成本/字节对照，escaped type/model/usage、read前flush/EOF/断开、cache TTL次序。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 补转义回退；不导入DeepSeek缺失placeholder消费者、无管理的懒日志全body保活、#8816账务lane或完整COW。

### Requirement: S23 Go工具链与时钟测试维护

Go维护 SHALL 对齐1.27.4声明/CI/工具，并让已有峰谷价格测试使用固定PricingAt。

#### Scenario: 选定既有路径

- **WHEN** Go工具链与时钟测试维护对应的既有请求或操作发生
- **THEN** 系统满足选定行为，并验证全量unit/lint/race/embed、同版本lint重建、隔离Ent/Wire兼容及固定23:59边界。

#### Scenario: 保护相邻行为

- **WHEN** 请求或操作触及相邻或未选分支
- **THEN** 不升级库版本、上游PG依赖、fork VERSION或历史记录；codegen/CI尚不是本轮已验收项。
