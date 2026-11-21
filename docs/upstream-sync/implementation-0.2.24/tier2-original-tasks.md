# 实施任务（0/52，未实施）

- [ ] 核对真实实施起点、冻结来源、产品/历史保护哈希和全部最小前置。

## S01 所有现有入口的重复字段拒绝与读取一致性

- [ ] 在实际起点建立本簇复现与负对照，区分夹具/来源与产品缺陷。
- [ ] 实现选定行为及最小前置，保留边界：不引入完整group model_allowlist/glob/live、新平台或PG。只挂本地已存在路由；现有账号模型过滤、64MiB预算/40MiB SSE/8MiB暂存/120s idle不变。相同重复也必须拒绝，不能只拒不同值。
- [ ] 完成验收并保存证据：等值/异值/大小写/转义/Unicode键、嵌套/数组/freeform参数、BOM与伪multipart、preamble、标量重复/扩展name与多图合法列表；首帧/后续帧/session.update；无转发/无usage；实际HTTP/WS和1C1G并发预算。

## S02 OpenAI 不可逆 OAuth 轮换的 SQLite 保存

- [ ] 在实际起点建立本簇复现与负对照，区分夹具/来源与产品缺陷。
- [ ] 实现选定行为及最小前置，保留边界：只对完整OpenAI轮换成功启用例外；Grok/Claude晚到规则、固定quota与取消边界保持。无新列，无旧token重试，不能使用普通Update覆盖并发授权。
- [ ] 完成验收并保存证据：请求/后台/旧fallback路径，enrichment超时、重新授权/代理/非token编辑、CAS冲突/写失败/重读失败、缓存失效/outbox、锁租期覆盖保存清理、实际SQLite/miniredis与race。

## S03 缓冲 Responses 的 hosted 输出重建

- [ ] 在实际起点建立本簇复现与负对照，区分夹具/来源与产品缺陷。
- [ ] 实现选定行为及最小前置，保留边界：不以本簇启用未选ToolSearch新能力；不吞终态failed、不重复计费、不改变首内容failover/有界spool。
- [ ] 完成验收并保存证据：空terminal.output、web/file/computer/code/search、custom工具、函数object input、独立content index、partial/done/terminal冲突、Grok server search不得伪装tool_use、真实HTTP/SSE。

## S04 Grok 成功搜索次数结算

- [ ] 在实际起点建立本簇复现与负对照，区分夹具/来源与产品缺陷。
- [ ] 实现选定行为及最小前置，保留边界：不得把失败搜索按item数重复收费；普通函数/客户端工具不进入服务端搜索计费。
- [ ] 完成验收并保存证据：web_search/x_search、显式0/正数/缺省/非法值、item.done+completed、HTTP/SSE/中断、用户与账号独立成本、SQLite settlement/dedup。

## S05 OAuth 外族模型与 Messages 映射后调度

- [ ] 在实际起点建立本簇复现与负对照，区分夹具/来源与产品缺陷。
- [ ] 实现选定行为及最小前置，保留边界：保留用户已选#8748，不把API Key自定义base_url限制为OpenAI模型，也不误拦合法映射别名。
- [ ] 完成验收并保存证据：OAuth/setup-token/API Key、普通/透传、model_filter开关/白名单、claude别名→GPT、GPT别名→外族、未知别名、HTTP/WS候选池。

## S06 批量透传编辑独立修改模型过滤

- [ ] 在实际起点建立本簇复现与负对照，区分夹具/来源与产品缺陷。
- [ ] 实现选定行为及最小前置，保留边界：未勾选修改不等于false；不覆盖既有账号的过滤或凭据。
- [ ] 完成验收并保存证据：OAuth/API Key只改透传、只改过滤、显式false、白名单独立修改、全选筛选目标、重新打开/目标变化、保存重开及移动端。

## S07 Claude/Gemini 用量快照刷新

- [ ] 在实际起点建立本簇复现与负对照，区分夹具/来源与产品缺陷。
- [ ] 实现选定行为及最小前置，保留边界：保留本地已实施的不闪行与OpenAI unknown新鲜度语义；不引入费用估算/强查按钮或来源原生国内平台。
- [ ] 完成验收并保存证据：5h/7d读数/窗口滚动、相同id更新、非可见行、批量队列、旧请求晚到、离页取消、错误保留旧读数、桌面/移动浏览器。

## S08 Codex 原生音频输入声明

- [ ] 在实际起点建立本簇复现与负对照，区分夹具/来源与产品缺陷。
- [ ] 实现选定行为及最小前置，保留边界：不引入Voice/Realtime/新的音频计费或模型；静态订阅目录不能替代API Key实际能力。
- [ ] 完成验收并保存证据：GPT6.2 Sol OAuth、API Key明确metadata、映射别名、混合账号交集、raw目录/缺metadata、仅Chat协议剔除音频与ETag稳定性。

## S09 Anthropic metadata 长上下文目录

- [ ] 在实际起点建立本簇复现与负对照，区分夹具/来源与产品缺陷。
- [ ] 实现选定行为及最小前置，保留边界：不凭目录增加新模型/价格，不导入未选group model_allowlist；现有[1m]请求规范化与计费保留。
- [ ] 完成验收并保存证据：ContextWindow阈值900000、别名/显式mapping/路由账号范围、[1m]去重、大小写、自定义base_url、metadata缺失/失败与真实models HTTP。

## S10 Composite 改写后的 Claude thinking

- [ ] 在实际起点建立本簇复现与负对照，区分夹具/来源与产品缺陷。
- [ ] 实现选定行为及最小前置，保留边界：不引入新型号或整套signed-thinking框架；保持用户原模型名的使用记录和最终映射策略。
- [ ] 完成验收并保存证据：enabled→adaptive并移除budget、Sonnet disabled→between_tools、Opus/普通Claude、nested model path、Bedrock及实际HTTP请求。

## S11 后台新建用户的默认平台配额

- [ ] 在实际起点建立本簇复现与负对照，区分夹具/来源与产品缺陷。
- [ ] 实现选定行为及最小前置，保留边界：不导入支付、购买或新订阅策略；不改变固定窗口重置、已删Key结算与历史用户配额，不新增迁移。
- [ ] 完成验收并保存证据：admin创建/注册两路径、普通用户/管理员、0/正数/缺省、写失败fail-open、已有事务隔离、BulkInsertInitial不覆盖已有值、实际SQLite及Wire生成。

## S12 异步用量与审计的请求体保活

- [ ] 在实际起点建立本簇复现与负对照，区分夹具/来源与产品缺陷。
- [ ] 实现选定行为及最小前置，保留边界：不全面复制所有body，不引入完整COW；对请求内零拷贝保留所有权，不把函数PoC宣传成生产内存上限。
- [ ] 完成验收并保存证据：字段字节一致、alias地址控制、body释放/取消/队列积压、审计快照归属、延迟worker与heap/alloc测量，64MiB预算释放后资源边界。

## S13 已有热点的有界窄优化

- [ ] 在实际起点建立本簇复现与负对照，区分夹具/来源与产品缺陷。
- [ ] 实现选定行为及最小前置，保留边界：可验证body完全相同才复用validated入口；不降低64MiB预算、不采用1GiB动态最低预算、stale grace/账务lane/完整COW；未达收益或等价性的hunk不落地。
- [ ] 完成验收并保存证据：每个hunk独立benchmark/alloc、JSON转义键和值/fuzz corpus、重复字段策略与S01组合、非法UTF8/JSON、body改写后range刷新、schema freeform不误改、快照成员及调用者图不变。

## S14 SQLite 模型平均速率按耗时加权

- [ ] 在实际起点建立本簇复现与负对照，区分夹具/来源与产品缺陷。
- [ ] 实现选定行为及最小前置，保留边界：与用户已选整窗口输出TPS统计口径不同：后者保留逐请求算术mean/P5/P10/P50，不能整站替换为加权值。
- [ ] 完成验收并保存证据：不等耗时、无效/NULL/零token、空集、平台/分组/时间/分页与实际ops HTTP；验证输出TPS分布现有均值与分位数不变。

## S15 Antigravity 裸 429 范围及额度端点

- [ ] 在实际起点建立本簇复现与负对照，区分夹具/来源与产品缺陷。
- [ ] 实现选定行为及最小前置，保留边界：不放宽明确限流或模型容量规则，不新增查询按钮/QuotaSummary平台/强查策略；保留本地端点安全与已实施的model cooldown。
- [ ] 完成验收并保存证据：多账号同请求/独立请求、带details/reset的真实429、retry退出后外层、daily/prod/显式配置、代理/重定向/有界读取、取消和当前quota读数。

## 组合验收

- [ ] 全量backend unit、embed build和lint；变化涉及生成代码时真实生成并核对。
- [ ] 相关包race，覆盖并发CAS/缓存/取消/所有权及两档交叉路径。
- [ ] 实际临时HTTP/WS/SQLite/miniredis：迁移230基线、settlement/dedup、固定配额和独立成本。
- [ ] 前端受影响范围的测试/typecheck/lint/build；涉及交互时实际桌面/移动保存重开与旧请求验收。
- [ ] 逐hunk资源测量及1C1G相关边界，完成两档组合控制并记录局限。
- [ ] 记录真实实施SHA、测试命令/结果及最终哈希，核对产品范围和冻结历史，完成verification后再更新任务状态。
