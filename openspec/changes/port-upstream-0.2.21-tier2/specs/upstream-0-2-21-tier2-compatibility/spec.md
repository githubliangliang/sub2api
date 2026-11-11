## ADDED Requirements

### Requirement: S01 原生 WS 后续轮次失败关闭

原生 Responses WS 后续轮次不能安全迁移完整会话时 SHALL 以可重试关闭结束，不能跨账号重放首帧。

#### Scenario: S01 正向行为

- **WHEN** 在既有入口执行本簇输入：第一轮成功、第二轮限流/过载：客户端1013关闭，第二账号未收到首轮任务；首轮failover仍允许。
- **THEN** 系统按上述SHALL行为完成，实际输出、状态与持久化证据可复核。

#### Scenario: S01 边界保护

- **WHEN** 请求命中边界、失败或相邻既有路径
- **THEN** 保留每轮Fast/模型/usage、lease与槽位释放；只有当前轮完整历史足够时才允许已有安全重试。

### Requirement: S02 OAuth WS 搜索激活与完整回放

OAuth 池化 WS 首次声明web_search时 SHALL 校验连接初始能力，必要时新建连接并携带完整历史。

#### Scenario: S02 正向行为

- **WHEN** 在既有入口执行本簇输入：无搜索prewarm→搜索、Lite additional_tools、工具续接保留reasoning/message/call/output；API Key及同名function不重连。
- **THEN** 系统按上述SHALL行为完成，实际输出、状态与持久化证据可复核。

#### Scenario: S02 边界保护

- **WHEN** 请求命中边界、失败或相邻既有路径
- **THEN** 未知previous_response_id或不完整工具历史要明确关闭，不能猜测回放；写失败不能标记初始化，状态受write锁保护，连接关闭/预算回收。

### Requirement: S03 WS 与 SSE 首 token 口径一致

WS 首token统计 SHALL 遵循管理员TTFT模式，并在每轮独立计算。

#### Scenario: S03 正向行为

- **WHEN** 在既有入口执行本簇输入：token/first event等所有已支持模式与SSE对照，至少两轮，空delta/工具输出/终态/codex元数据/取消。
- **THEN** 系统按上述SHALL行为完成，实际输出、状态与持久化证据可复核。

#### Scenario: S03 边界保护

- **WHEN** 请求命中边界、失败或相邻既有路径
- **THEN** 本地上一轮已经修每轮计时；保留本地起点接口，不引入来源FirstTurnStartedAt/TakeNextTurnStartedAt整基座只为编译。

### Requirement: S04 显式模型冷却覆盖容量重试

OpenAI 显式model范围冷却规则 SHALL 对 overload/slow_down及流内503生效，且只匹配实际错误字段。

#### Scenario: S04 正向行为

- **WHEN** 在既有入口执行本簇输入：HTTP、普通/透传SSE、两种缓冲路径及WS：匹配后停止同账号重试，作用于已映射模型，持久化/缓存一致。
- **THEN** 系统按上述SHALL行为完成，实际输出、状态与持久化证据可复核。

#### Scenario: S04 边界保护

- **WHEN** 请求命中边界、失败或相邻既有路径
- **THEN** 未配、account范围、无模型、排除码及其它平台保留请求级默认；output/metadata关键词不能触发；写失败不扩大整账号冻结。

### Requirement: S05 Anthropic 兼容入口账号错误策略

Anthropic账号经Chat或Responses兼容入口时 SHALL 执行已有自定义错误码及临时不可调度策略。

#### Scenario: S05 正向行为

- **WHEN** 在既有入口执行本簇输入：410自定义/临时规则在两入口、流/非流生效并留给handler换号；未配410及400保持原响应。
- **THEN** 系统按上述SHALL行为完成，实际输出、状态与持久化证据可复核。

#### Scenario: S05 边界保护

- **WHEN** 请求命中边界、失败或相邻既有路径
- **THEN** pool模式重试与shouldDisable结果一致；不要将全部客户端4xx升级为failover，不扩商业支付策略。

### Requirement: S06 合成 effort 的映射与最终校验

Messages转OpenAI SHALL 对显式disabled思考合成的none执行一次映射及上限策略，并在映射后校验最终effort。

#### Scenario: S06 正向行为

- **WHEN** 在既有入口执行本簇输入：disabled+none映射/deny/ceiling、显式非法值、普通省略default、Sol/Terra none/max、count_tokens和Chat fallback一致。
- **THEN** 系统按上述SHALL行为完成，实际输出、状态与持久化证据可复核。

#### Scenario: S06 边界保护

- **WHEN** 请求命中边界、失败或相邻既有路径
- **THEN** 普通省略不能被medium映射/上限意外改变；现有GLM通用兼容降级保留；不引入Kimi/Zhipu等原生平台或requested_effort schema。

### Requirement: S07 Refresh Token 原子轮换

刷新令牌 SHALL 在校验后原子消费旧token，保证同一token并发只签发一份新令牌；缓存错误必须失败关闭。

#### Scenario: S07 正向行为

- **WHEN** 在既有入口执行本簇输入：同步barrier八请求仅一成功；实际miniredis原子GET+DEL、过期/重用/禁用用户/token version/家族吊销/缓存失败。
- **THEN** 系统按上述SHALL行为完成，实际输出、状态与持久化证据可复核。

#### Scenario: S07 边界保护

- **WHEN** 请求命中边界、失败或相邻既有路径
- **THEN** 新增接口需更新所有stub；保持预校验顺序，失败不泄露token；单节点miniredis真实执行，不能仅用mock证明原子性。

### Requirement: S08 Key 列表独立加载用量与 SQLite 覆盖索引

API Key列表 SHALL 在列表响应到达后展示，用量独立异步加载；统计查询可采用经验证的SQLite覆盖索引。

#### Scenario: S08 正向行为

- **WHEN** 在既有入口执行本簇输入：延迟/失败/取消用量不阻塞列表；快速筛选分页旧请求不能回写；空列表及原用量区间数值一致。
- **THEN** 系统按上述SHALL行为完成，实际输出、状态与持久化证据可复核。

#### Scenario: S08 边界保护

- **WHEN** 请求命中边界、失败或相邻既有路径
- **THEN** SQLite拟议229而非上游249；(api_key_id,created_at,actual_cost)先验证写入/大小代价与查询计划，禁止CONCURRENTLY/INCLUDE及修改旧迁移。

### Requirement: S09 账号刷新保留行与已知配额

账号列表和配额刷新 SHALL 保留当前已展示内容，并清晰标识刷新中的不可交互状态。

#### Scenario: S09 正向行为

- **WHEN** 在既有入口执行本簇输入：初次加载骨架、刷新保留DOM与额度、失败保留旧值、成功更新；分页/筛选/账号复用/批量选择/移动端。
- **THEN** 系统按上述SHALL行为完成，实际输出、状态与持久化证据可复核。

#### Scenario: S09 边界保护

- **WHEN** 请求命中边界、失败或相邻既有路径
- **THEN** 只保留同账号数据；旧请求不可覆盖新账号；不依赖未选Antigravity新查询按钮或Ollama基座。

### Requirement: S10 Codex Desktop 页面上下文兼容

符合严格签名的Codex页面function_call_output SHALL 转为标记的非信任user文本，避免孤儿工具结果错误。

#### Scenario: S10 正向行为

- **WHEN** 在既有入口执行本簇输入：open_page及selected_page分片、历史重放、混合automation；真实HTTP捕获首轮消息，顺序和页面原文保留。
- **THEN** 系统按上述SHALL行为完成，实际输出、状态与持久化证据可复核。

#### Scenario: S10 边界保护

- **WHEN** 请求命中边界、失败或相邻既有路径
- **THEN** 有真实call_id、错误namespace/name、重复JSON键、坏page_id/分片头均不转换；不提升system权限；日志不含页面内容。

### Requirement: S11 既有安全审计 worker 入队唤醒

已有审计任务成功发布后 SHALL 非阻塞唤醒本进程worker，同时保留轮询拾取重试任务。

#### Scenario: S11 正向行为

- **WHEN** 在既有入口执行本簇输入：有任务立即claim、队列满/持久化失败不通知、连续任务合并唤醒、禁用/并发启停/Shutdown与到期重试。
- **THEN** 系统按上述SHALL行为完成，实际输出、状态与持久化证据可复核。

#### Scenario: S11 边界保护

- **WHEN** 请求命中边界、失败或相邻既有路径
- **THEN** 保持SQLite claim原子性和1C1G资源界限；不新增审计规则、远端扫描策略或实例运营能力，不把来源平均250ms当本地实测。

### Requirement: S12 Claude Code 内置版本基线

既有Claude Code身份链 SHALL 统一使用2.1.321内置基线，并保留配置与有效环境覆盖规则。

#### Scenario: S12 正向行为

- **WHEN** 在既有入口执行本簇输入：CLIVersion、User-Agent、x-app/指纹相关已有消费者与版本同步服务一致；合法高版本覆盖保留，低版本/非法值回退。
- **THEN** 系统按上述SHALL行为完成，实际输出、状态与持久化证据可复核。

#### Scenario: S12 边界保护

- **WHEN** 请求命中边界、失败或相邻既有路径
- **THEN** 仅constants的版本literal和必要期望测试；不引入Fable/Mythos5.2、目录/定价/白名单/UI等同PR新能力。

