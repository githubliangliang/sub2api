## ADDED Requirements

### Requirement: S01 所有现有入口的重复字段拒绝与读取一致性

所有现有HTTP网关请求与每轮WS载荷 SHALL 在模型改写、调度和计费之前拒绝语义重复字段，并让内容识别、body重读与协议错误格式一致。

#### Scenario: 既有消费者满足目标行为

- **WHEN** 等值/异值/大小写/转义/Unicode键、嵌套/数组/freeform参数、BOM与伪multipart、preamble、标量重复/扩展name与多图合法列表；首帧/后续帧/session.update；无转发/无usage；实际HTTP/WS和1C1G并发预算。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 不引入完整group model_allowlist/glob/live、新平台或PG。只挂本地已存在路由；现有账号模型过滤、64MiB预算/40MiB SSE/8MiB暂存/120s idle不变。相同重复也必须拒绝，不能只拒不同值。

### Requirement: S02 OpenAI 不可逆 OAuth 轮换的 SQLite 保存

已完成的OpenAI轮换 SHALL 在有界脱离取消的上下文保存；CAS失败须读取当前持久化凭据且保留并发重新授权。

#### Scenario: 既有消费者满足目标行为

- **WHEN** 请求/后台/旧fallback路径，enrichment超时、重新授权/代理/非token编辑、CAS冲突/写失败/重读失败、缓存失效/outbox、锁租期覆盖保存清理、实际SQLite/miniredis与race。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 只对完整OpenAI轮换成功启用例外；Grok/Claude晚到规则、固定quota与取消边界保持。无新列，无旧token重试，不能使用普通Update覆盖并发授权。

### Requirement: S03 缓冲 Responses 的 hosted 输出重建

缓冲兼容响应 SHALL 从最终有效终态及已完成item恢复完整hosted内容，输出顺序、工具身份和流式语义一致。

#### Scenario: 既有消费者满足目标行为

- **WHEN** 空terminal.output、web/file/computer/code/search、custom工具、函数object input、独立content index、partial/done/terminal冲突、Grok server search不得伪装tool_use、真实HTTP/SSE。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 不以本簇启用未选ToolSearch新能力；不吞终态failed、不重复计费、不改变首内容failover/有界spool。

### Requirement: S04 Grok 成功搜索次数结算

Grok搜索结算 SHALL 在上游提供成功次数时以其为准，缺省时沿用已有事件计数与去重。

#### Scenario: 既有消费者满足目标行为

- **WHEN** web_search/x_search、显式0/正数/缺省/非法值、item.done+completed、HTTP/SSE/中断、用户与账号独立成本、SQLite settlement/dedup。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 不得把失败搜索按item数重复收费；普通函数/客户端工具不进入服务端搜索计费。

### Requirement: S05 OAuth 外族模型与 Messages 映射后调度

固定Codex上游的OAuth类账号 SHALL 跳过明确外族模型，且调度判断使用渠道映射后的实际模型。

#### Scenario: 既有消费者满足目标行为

- **WHEN** OAuth/setup-token/API Key、普通/透传、model_filter开关/白名单、claude别名→GPT、GPT别名→外族、未知别名、HTTP/WS候选池。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 保留用户已选#8748，不把API Key自定义base_url限制为OpenAI模型，也不误拦合法映射别名。

### Requirement: S06 批量透传编辑独立修改模型过滤

批量编辑 SHALL 只提交用户明确选择修改的透传模型过滤字段。

#### Scenario: 既有消费者满足目标行为

- **WHEN** OAuth/API Key只改透传、只改过滤、显式false、白名单独立修改、全选筛选目标、重新打开/目标变化、保存重开及移动端。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 未勾选修改不等于false；不覆盖既有账号的过滤或凭据。

### Requirement: S07 Claude/Gemini 用量快照刷新

Claude/Gemini快照变化 SHALL 使现有用量缓存失效并刷新，异步旧结果不得覆盖新快照。

#### Scenario: 既有消费者满足目标行为

- **WHEN** 5h/7d读数/窗口滚动、相同id更新、非可见行、批量队列、旧请求晚到、离页取消、错误保留旧读数、桌面/移动浏览器。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 保留本地已实施的不闪行与OpenAI unknown新鲜度语义；不引入费用估算/强查按钮或来源原生国内平台。

### Requirement: S08 Codex 原生音频输入声明

已有原生Responses音频输入能力 SHALL 在模型声明中正确保留，聚合目录只声明全部可调度候选支持的能力。

#### Scenario: 既有消费者满足目标行为

- **WHEN** GPT6.2 Sol OAuth、API Key明确metadata、映射别名、混合账号交集、raw目录/缺metadata、仅Chat协议剔除音频与ETag稳定性。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 不引入Voice/Realtime/新的音频计费或模型；静态订阅目录不能替代API Key实际能力。

### Requirement: S09 Anthropic metadata 长上下文目录

现有Anthropic目录 SHALL 按可服务账号的有效metadata展示长上下文变体，并正确识别默认官方registry host。

#### Scenario: 既有消费者满足目标行为

- **WHEN** ContextWindow阈值900000、别名/显式mapping/路由账号范围、[1m]去重、大小写、自定义base_url、metadata缺失/失败与真实models HTTP。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 不凭目录增加新模型/价格，不导入未选group model_allowlist；现有[1m]请求规范化与计费保留。

### Requirement: S10 Composite 改写后的 Claude thinking

Composite将请求改写到既有Claude5.5时 SHALL 同步适配该型号接受的thinking模式并保留其它参数。

#### Scenario: 既有消费者满足目标行为

- **WHEN** enabled→adaptive并移除budget、Sonnet disabled→between_tools、Opus/普通Claude、nested model path、Bedrock及实际HTTP请求。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 不引入新型号或整套signed-thinking框架；保持用户原模型名的使用记录和最终映射策略。

### Requirement: S11 后台新建用户的默认平台配额

后台创建普通用户 SHALL 按当前默认平台配额初始化，显式0与未配置无限额保持区别。

#### Scenario: 既有消费者满足目标行为

- **WHEN** admin创建/注册两路径、普通用户/管理员、0/正数/缺省、写失败fail-open、已有事务隔离、BulkInsertInitial不覆盖已有值、实际SQLite及Wire生成。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 不导入支付、购买或新订阅策略；不改变固定窗口重置、已删Key结算与历史用户配额，不新增迁移。

### Requirement: S12 异步用量与审计的请求体保活

离开请求生命周期的用量和审计数据 SHALL 拥有必要标量的独立存储，且不因字段地址或闭包滞留整个请求体。

#### Scenario: 既有消费者满足目标行为

- **WHEN** 字段字节一致、alias地址控制、body释放/取消/队列积压、审计快照归属、延迟worker与heap/alloc测量，64MiB预算释放后资源边界。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 不全面复制所有body，不引入完整COW；对请求内零拷贝保留所有权，不把函数PoC宣传成生产内存上限。

### Requirement: S13 已有热点的有界窄优化

采用的热点优化 SHALL 保持字节、策略和所有权语义，并通过单P前后成本与转义必要条件验证。

#### Scenario: 既有消费者满足目标行为

- **WHEN** 每个hunk独立benchmark/alloc、JSON转义键和值/fuzz corpus、重复字段策略与S01组合、非法UTF8/JSON、body改写后range刷新、schema freeform不误改、快照成员及调用者图不变。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 可验证body完全相同才复用validated入口；不降低64MiB预算、不采用1GiB动态最低预算、stale grace/账务lane/完整COW；未达收益或等价性的hunk不落地。

### Requirement: S14 SQLite 模型平均速率按耗时加权

模型统计平均Tokens/秒 SHALL 使用有效请求output_tokens总和除以duration_ms总和。

#### Scenario: 既有消费者满足目标行为

- **WHEN** 不等耗时、无效/NULL/零token、空集、平台/分组/时间/分页与实际ops HTTP；验证输出TPS分布现有均值与分位数不变。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 与用户已选整窗口输出TPS统计口径不同：后者保留逐请求算术mean/P5/P10/P50，不能整站替换为加权值。

### Requirement: S15 Antigravity 裸 429 范围及额度端点

同一请求的无明确信号裸429 SHALL 最多冷却一个账号；既有额度查询 SHALL 使用该账号实际转发端点。

#### Scenario: 既有消费者满足目标行为

- **WHEN** 多账号同请求/独立请求、带details/reset的真实429、retry退出后外层、daily/prod/显式配置、代理/重定向/有界读取、取消和当前quota读数。
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** 不放宽明确限流或模型容量规则，不新增查询按钮/QuotaSummary平台/强查策略；保留本地端点安全与已实施的model cooldown。

