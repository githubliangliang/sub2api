# 设计与决策

## 范围与来源优先级

实施编号统一为 T01–T23（T01–T16不变，T17–T23从原第二档迁入），现行映射见 [implementation-map.md](./implementation-map.md)。
source-baseline.md 与 source-feature-map.md 保留评估原貌；出现矛盾时，以本设计、spec 和现行映射为实施契约。
本次仅完成文档，不把静态核对当作产品修复或测试通过。

移植选择：采用“整 PR 阅读、按行为手工移植”的方式。直接整 PR 应用会带入不适用的测试；
重写整个功能会扩大回归面。每项保留本地逻辑，实施可独立提交/回退，提交仅在另获用户指令后执行。

## T01–T06：数据与转发正确性

### T01 缓存隔离

删除 project_id 优先分支，统一使用 `ag:account:<id>`；无 project_id 时行为不变。
invalidator 清理账号键以及非空、trim 后的旧 `ag:<project_id>` 键。
新 reader 不再读取旧键，清理旧键是升级残留清理，不是修复后仍会命中旧键的前提。
不引入数据库迁移、缓存全量扫描或启动清空。两账号共 project_id 必须分别取得自己的 token。

### T02 SSE 分隔

已解包的 data 行仍写成 `data: ...\n\n`，随后跳过 trim 后为空的输入行。
非空、非 data 行保持既有透传；data 正文继续既有解包逻辑，不承诺与上游原始字节一致。
验证连续三个事件的完整输出和帧数，兼顾心跳、结束及非空注释行。

### T03 请求头大小写

仅把入站到出站头映射的值改为规范 `Accept-Encoding`；
不删除允许转发项，也不变更跳过集合。通过真实本地 HTTP transport 断言线上的请求头数量，
不能只比较 map 字符串，更不能把它描述为重复响应头。

### T04 账号优先级

priority 改升序；同 priority 保持 account ID 升序。
保留 provider 顺序、IsSchedulable、模型支持和 SupportsAccount 等过滤；低 priority 但不可用的账号不能被选中。

### T05 调度投影

credentials 保留 `account_scheduling_threshold`；extra 保留
`session_window_utilization`、`passive_usage_7d_utilization`、`passive_usage_7d_reset`、
`passive_usage_7d_oi_utilization`、`passive_usage_7d_oi_reset`。
完整账号 hydrate 前已经读取候选投影，因此两个白名单必须一起补齐。
普通 unit 覆盖配置值（含 0）、未配置不新增键、无关凭据仍过滤；不移植 postgres-gated 测试或修改其构建标签。

### T06 Codex UA

在 trim 前校验 header value，拒绝非法控制字节并显式拒绝 CR/LF。
配对函数与最终转发入口都必须覆盖；非法身份整体回退既有默认身份，不发送输入中的非法 UA。
正常 UA 保留版本、平台与 originator 推导。使用现有 `golang.org/x/net/http/httpguts`，
实施前确认依赖已满足，不升级 go.mod/go.sum；HTAB 按该库既有 header 语法处理，不声称禁止所有控制字符。

## T07–T09：错误语义

### T07 认证查询

JWT 与管理员的 GetByID：`errors.Is(err, service.ErrUserNotFound)` 返回 401 / USER_NOT_FOUND；
其它查询错误返回 500 / INTERNAL_ERROR，且不继续处理请求。包装后的 not-found 也必须识别。

API Key 认证调用 GetByKey，已经区分 ErrAPIKeyNotFound → 401 / INVALID_API_KEY、
ErrAPIKeyAuthOverloaded → 503 / API_KEY_AUTH_OVERLOADED、其它错误 → 500 / INTERNAL_ERROR。
后续 `apiKey.User == nil` 返回 401 / USER_NOT_FOUND；该分支没有可检查的 GetByID 错误。
因此本项修复两处产品查询，另对 API Key 既有分支做回归，不凭静态关键词匹配新增第三处补丁。

### T08 前端刷新

优先保留既有 sessionChanged 判断，旧请求不得销毁新会话。
在会话未替换时，Axios 错误无响应/状态 0、429、5xx 返回
`TOKEN_REFRESH_UNAVAILABLE`，保留 token、auth_user、过期时间和当前位置；
返回状态为原始状态或无响应时的 0，不设置 auth_expired，也不自动重复请求。

401/403、其它非瞬时拒绝，以及现有 refresh helper 抛出的无效响应错误，保留原有清理与跳转逻辑。
这是上游的错误分类边界，并非“只有 401/403 才能登出”；不在本批新增响应结构验证器。
refresh 成功继续重放原请求；同页 single-flight、跨标签轮换和旧会话保护不得改写。

### T09 峰谷配置

CreateGroup/UpdateGroup 的 ValidatePeakRateConfig 错误包装为
`infraerrors.BadRequest("INVALID_PEAK_RATE_CONFIG", ...)`。
保留中文字段信息、归一化顺序和校验规则，不引入跨夜支持。
启用订阅峰谷时，以 22:00–02:00 为非法例、09:00–18:00 为合法例；非法配置不得调用写库操作。

## T10–T16：界面与配置

- **T10**：只把 UsageTable 成本浮层的输入/输出、图像、缓存、原始/用户/账号费用等金额改为 8 位，缺省金额显示 0.00000000。保留表格及其它单价格式和计费计算；用 0.00000012 验证，不声称任意正数都不会舍入为零。
- **T11**：续费弹窗外层受可用视口限制，标题不收缩，列表 min-height 为 0 且 overflow-y 为 auto。真实浏览器验证 360×640 及桌面视口；DOM 类名断言不能替代滚动验收。
- **T12**：仅在 upstreamType 的小写字母、数字、下划线、横线白名单中增加点号；前端说明与中英文 i18n 同步。名称/ID 等其它字段规则不扩大。
- **T13**：异步验证完成后按发起请求的邮箱身份删除 pending 项，不能使用 await 前保存的索引；手动删除入口同步核对。失败不得误删，重排/并发完成也不能删除其它邮箱。
- **T14**：代理筛选处理函数先设 page=1，再发起一次刷新；沿用既有请求竞态保护。分页按钮仍请求目标页，不重置 page_size。
- **T15**：搜索输入变化时在防抖之前比较 trim 后关键词与已选邮箱；不匹配立即清 selectedUser 和 assignForm.user_id。相同邮箱或仅首尾空格变化保留有效选择；表单不能提交空目标。
- **T16**：前端 max 从 365 对齐后端 MaxValidityDays=36500，保留既有下限/步进/默认值；36501 仍不允许，不修改后端边界。

## T17–T23：前移的既有修复

这些行为直接承接原第二档 spec，不新增验收能力；来源与产品路径见 implementation-map.md。

- **T17 #7055**：兑换已成功后，用户信息刷新失败只产生独立告警；沿用后续订阅/历史处理，不自动重试兑换。
- **T18 #7054**：导出开始时捕获筛选与日期，每页只改 page，文件名使用同一快照；不扩展成数据库一致性快照。
- **T19 #6964**：按片段顺序拼接 agent_message 的信封和自定义 provider 明文正文；空项跳过，不引入解密，保留其它工具能力。
- **T20 #7026**：await 前固定密钥对象，按返回状态更新；只有编辑器仍选中该对象时同步表单。
- **T21 #7112**：修复 tick 的一秒偏移和渠道页重置使用固定默认间隔的问题，不重写自动刷新机制。
- **T22 #7023**：资料与密码表单复用本地已有错误提取函数，不改变 API 错误格式。
- **T23 #7025**：记录本次打开期间的部分导入成功，关闭时通知一次；完整成功、全失败和重新打开的状态分别保持正确。

## 实施顺序与范围守恒

先处理串号、分帧、投影、认证、错误分配及 T17/T18/T19 的结果/正文正确性，再做其它局部修复；
EasyPay 等使用前提项在后段。具体顺序以 tasks.md 为准，编号不代表排序。
第二档 S01 #7094 后续也改 chat bridge，必须保留 T19；第二档 S11 #6916 必须保留 T15。
本批承接7个既有 PR，第一档22个、第二档21个，合计仍43个；不纳入第三/第四档。

## 验证、交付与回退

实施按 tasks.md 逐项推进，失败复现应证明行为问题，不能用缺 fixture/无法编译当作 RED。
纯输入约束与布局项可记录改前/改后的直接验证；不为单个类名增加镜像测试。
测试改用本地 fixture，尤其 T05 的 postgres 集成片段和 T15 缺失的上游测试文件。
后端/前端回归、边界与浏览器检查集中在 verification.md，不重复维护命令口径。

无数据库结构变更；产品回退按行为撤销对应补丁，T07/T08 建议同批交付并联合回归。
T01 回退到旧 reader 会恢复 project_id 串号风险，回退前需评估共享 project_id 账号并清理相关缓存；
其它回退会恢复各自原有错误表现。文档状态只有在执行证据充分时才能从“未实施”更新为“已验收”。
