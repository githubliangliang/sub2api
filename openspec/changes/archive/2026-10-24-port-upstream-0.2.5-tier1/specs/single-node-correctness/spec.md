## ADDED Requirements

### Requirement: T01 Antigravity token 缓存按账号隔离
Antigravity OAuth token 缓存 SHALL 以账号 ID 作为唯一键；MUST NOT 使用 `project_id` 作为主键。
失效逻辑 SHALL 同时清理旧 `ag:<project_id>` 与新 `ag:account:<id>` 两种键。

#### Scenario: 两个账号共用同一 project_id
- **WHEN** 账号 A 与账号 B 的 `project_id` 相同，且 A 的 token 已缓存
- **THEN** B 取 token 时不得命中 A 的缓存条目，各自使用自己的凭据。

#### Scenario: 账号无 project_id
- **WHEN** 账号未配置 `project_id`
- **THEN** 仍按账号 ID 缓存，行为与既有回退分支一致。

#### Scenario: 升级后的旧缓存
- **WHEN** 升级前写入的 `ag:<project_id>` 条目仍存在
- **THEN** 失效动作清理该旧键，同时新读取路径不再访问旧 project 键；不要求全量扫描缓存。

### Requirement: T02 Antigravity Gemini SSE 单一空行分隔
透传 Gemini SSE 时，系统 SHALL 只在事件之间写出一个 `\n\n` 分隔；MUST NOT 额外透传上游的空行。

#### Scenario: 上游事件之间带空行
- **WHEN** 上游流在 `data:` 行之后另发一个空行
- **THEN** 下游收到的字节流中事件之间恰好一个空行，第二个及后续事件可被 `\n\n` 分帧器正常解析。

#### Scenario: 非空非 data 行
- **WHEN** 上游发送非空注释或其它非 data 行
- **THEN** 保持既有透传内容和顺序；data 行继续既有解包逻辑，不增加额外分隔。

### Requirement: T03 转发不产生重复 accept-encoding
转发请求头 SHALL 使用规范大小写写出 `Accept-Encoding`，使 `net/http` 能识别并去重。

#### Scenario: 客户端带 Accept-Encoding
- **WHEN** 客户端请求携带 `Accept-Encoding: gzip`
- **THEN** 出站请求只出现一个 `Accept-Encoding` 头。

#### Scenario: 其它被转发头
- **WHEN** 转发其它头部
- **THEN** 既有转发与剥离规则不变。

### Requirement: T04 批量生图账号优先级升序
批量生图选号 SHALL 按 `priority` 升序排序（小值优先），与本仓库其它调度路径一致。

#### Scenario: 多账号不同优先级
- **WHEN** 候选账号 priority 分别为 1、5、10 且均可用
- **THEN** 先选 priority=1 的账号。

#### Scenario: 优先级相同
- **WHEN** 多个账号 priority 相同
- **THEN** 保持 account ID 升序的次级排序规则；不可调度或不支持模型的账号仍被跳过。

### Requirement: T05 调度候选投影保留阈值与窗口元数据
`filterSchedulerCredentials` SHALL 保留 `account_scheduling_threshold`；
`filterSchedulerExtra` SHALL 保留 `session_window_utilization`、`passive_usage_7d_utilization`、
`passive_usage_7d_reset`、`passive_usage_7d_oi_utilization`、`passive_usage_7d_oi_reset`。

#### Scenario: 账号自带阈值覆盖
- **WHEN** 账号设置了区别于平台默认的 `account_scheduling_threshold`
- **THEN** 候选准入按账号阈值判定，不回退到平台阈值。

#### Scenario: Anthropic 共享窗口字段
- **WHEN** 账号带有 session window 或 passive usage 指标
- **THEN** 这些字段出现在投影中，供共享窗口判定使用。

#### Scenario: 未配置上述字段
- **WHEN** 账号没有这些键
- **THEN** 投影不新增空值键，既有其它键集合不变。

### Requirement: T06 Codex User-Agent 校验后才参与身份配对
系统 SHALL 在解析并转发 Codex User-Agent 前校验其为合法 header 字段值，
MUST NOT 允许 CR/LF 或不符合 header value 语法的控制字节进入出站头；校验 SHALL 在 trim 前执行。

#### Scenario: UA 携带 CRLF
- **WHEN** User-Agent 含 `\r\n` 、NUL、DEL 等非法字节
- **THEN** 拒绝该身份配对，整体使用既有默认身份，不把非法输入值写入出站请求头。

#### Scenario: 合法 UA
- **WHEN** User-Agent 是正常的 `codex-tui/x.y.z (...)` 形态
- **THEN** 既有身份配对与 originator 推导行为不变。

### Requirement: T07 瞬时故障不得表现为身份不存在
认证中间件 SHALL 区分「用户确实不存在」与「查询失败」：
仅前者返回 401 `USER_NOT_FOUND`，后者 SHALL 返回 500。
JWT 与管理员 GetByID 查询 SHALL 使用 `errors.Is` 识别包装后的 `service.ErrUserNotFound`。
API Key 的既有 GetByKey 查询错误分类 SHALL 保留：非 not-found 错误为 500，过载为 503；
查询成功但关联 User 为 nil 时仍返回 401 `USER_NOT_FOUND`，不得把该分支当成查询异常修改。

#### Scenario: 查询期间数据库超时
- **WHEN** 用户查询因 DB 超时或连接失败返回非 not-found 错误
- **THEN** 返回 500，客户端会话不被清除。

#### Scenario: 用户确实不存在
- **WHEN** JWT 或管理员用户查询返回原始或包装后的 `service.ErrUserNotFound`
- **THEN** 保持既有 401 `USER_NOT_FOUND` 语义。

#### Scenario: API Key 既有错误分类
- **WHEN** GetByKey 返回普通查询错误、`ErrAPIKeyAuthOverloaded` 或 `ErrAPIKeyNotFound`
- **THEN** 分别保留 500 `INTERNAL_ERROR`、503 `API_KEY_AUTH_OVERLOADED`、401 `INVALID_API_KEY`。

#### Scenario: API Key 缺失关联用户
- **WHEN** GetByKey 成功返回但 `apiKey.User == nil`
- **THEN** 保持 401 `USER_NOT_FOUND`，不臆测为数据库查询异常。

### Requirement: T08 刷新令牌的瞬时失败保留会话
前端 SHALL 对 Axios 网络错误（无响应或状态 0）、429 与 5xx 保留会话并报告不可用；
401/403、其它非瞬时拒绝及现有刷新逻辑抛出的无效响应错误 SHALL 保持既有登出行为。
会话已被替换时 SHALL 优先保护新会话，不受旧刷新请求结果影响。

#### Scenario: 刷新请求遇到瞬时错误
- **WHEN** 刷新接口返回 429、500、503 或请求根本没有响应，且会话未替换
- **THEN** 保留 auth_token、refresh_token、auth_user、token_expires_at，不设置 auth_expired、不跳转 `/login`；返回 `TOKEN_REFRESH_UNAVAILABLE` 与原响应状态（无响应为 0），不自动增加重试。

#### Scenario: 刷新被明确拒绝
- **WHEN** 刷新接口返回 401/403
- **THEN** 保持既有登出与跳转行为。

#### Scenario: 刷新期间会话变化
- **WHEN** 旧请求刷新未完成时用户登出或切换到另一会话
- **THEN** 保留既有 sessionChanged 保护，旧请求不得删除新会话凭据。

#### Scenario: 刷新成功
- **WHEN** 刷新成功且会话仍有效
- **THEN** 使用新 access token 重放原请求，并保持现有同页/跨标签刷新协调。

### Requirement: T09 分组峰谷配置非法时返回 400
创建与更新分组时，峰谷配置校验失败 SHALL 以 400 及 `INVALID_PEAK_RATE_CONFIG` 返回，MUST NOT 返回 500。
既有归一化、校验规则和合法配置行为 SHALL 不变。

#### Scenario: 不支持的跨夜区间
- **WHEN** 为订阅分组启用峰谷并提交 22:00–02:00
- **THEN** 返回 400，错误信息指明具体字段，且分组未被写库。

#### Scenario: 合法峰谷配置
- **WHEN** 配置合法
- **THEN** 创建/更新照常成功。

### Requirement: T10 成本浮层保留 8 位小数
UsageTable 成本浮层 SHALL 对输入/输出、图像、缓存、原始/用户/账号费用使用 8 位小数；
缺省金额 SHALL 显示 `0.00000000`。表格行格式、其它单价格式及计费算法 SHALL 保持不变。

#### Scenario: 可表示的微小成本
- **WHEN** 浮层成本为 0.00000012
- **THEN** 显示 `$0.00000012`，不被 6 位小数舍入为零；小于 8 位精度的数仍按既有舍入规则显示。

#### Scenario: 缺省金额
- **WHEN** 浮层金额缺省走零值回退
- **THEN** 显示 `$0.00000000`。

### Requirement: T11 续费套餐弹窗可滚动
续费套餐选择弹窗 SHALL 在套餐数量超出视口时可纵向滚动，且标题保持可见。

#### Scenario: 套餐数量超出屏幕高度
- **WHEN** 可选套餐较多或视口较矮
- **THEN** 列表区域可滚动，最后一个套餐可达，弹窗不被裁切。

### Requirement: T12 EasyPay 自定义方式 code 允许点号
EasyPay 自定义支付方式的 `upstreamType` 校验 SHALL 接受包含点号的合法上游标识。

#### Scenario: 带点号的上游类型
- **WHEN** 配置 `upstreamType` 形如 `alipay.qr`
- **THEN** 校验通过并可保存。

#### Scenario: 非法字符
- **WHEN** 值含空格或其它未允许字符
- **THEN** 仍被拒绝。

### Requirement: T13 待验证邮箱按身份删除
删除待验证的余额通知邮箱 SHALL 按邮箱身份删除，MUST NOT 在异步完成时使用旧列表下标。

#### Scenario: 验证完成与列表变化竞态
- **WHEN** 邮箱 A 的验证请求未结束时邮箱 B 被删除或列表重排，随后 A 验证成功
- **THEN** 仅从 pending 列表移除 A，其它邮箱保留。

#### Scenario: 验证失败
- **WHEN** 验证请求失败
- **THEN** 保留对应 pending 邮箱且不误删其它条目。

### Requirement: T14 筛选变更重置分页
代理列表的筛选条件变更 SHALL 将分页重置到第 1 页再请求数据。

#### Scenario: 在靠后页码上收紧筛选
- **WHEN** 用户处于第 5 页并把筛选结果缩小到不足一页
- **THEN** 列表回到第 1 页并展示匹配结果，不出现空列表。

### Requirement: T15 订阅分配目标随搜索关键词失效
修改用户搜索关键词且 trim 后不等于已选邮箱时 SHALL 在防抖执行前失效 selectedUser 和 assignForm.user_id，
MUST NOT 保留与当前关键词不符的旧选择；trim 后仍相同的关键词 SHALL 保留有效选择。

#### Scenario: 改动关键词后立即提交
- **WHEN** 已选中某用户，随后修改关键词并在防抖搜索完成前提交
- **THEN** 不会把订阅分配给上一个已失效的用户。

#### Scenario: 同一邮箱的空白变化
- **WHEN** 关键词仅增加首尾空白，trim 后仍等于已选邮箱
- **THEN** 保留有效选择；搜索后显式选择另一用户时，只向新用户分配。

### Requirement: T16 兑换时长上限与后端一致
兑换码时长输入上限 SHALL 为后端 MaxValidityDays 的 36500，下限、步进与默认值 SHALL 保持不变。

#### Scenario: 超过一年的时长
- **WHEN** 管理员输入大于 365 且不超过后端上限的天数
- **THEN** 前端接受 366 与 36500 等输入，由后端统一裁决。

#### Scenario: 超过后端上限
- **WHEN** 输入 36501 天
- **THEN** 前端 max 约束仍拒绝该值，后端上限不变。

### Requirement: T17 兑换成功不被用户信息刷新失败覆盖
兑换接口已成功后，用户信息刷新失败 SHALL 仅展示独立 warning，不进入兑换失败分支，并继续后续订阅状态/兑换历史处理（来源 #7055）。

#### Scenario: 兑换成功但 refreshUser 失败
- **WHEN** 成功兑换后用户信息刷新请求失败
- **THEN** 保留兑换成功结果，提示信息刷新失败，不引导用户把已成功兑换视作失败重试。

#### Scenario: 兑换本身失败
- **WHEN** 兑换接口直接返回错误
- **THEN** 继续使用原有兑换失败提示，不显示成功结果。

### Requirement: T18 CSV 导出冻结筛选参数
用量 CSV 导出 SHALL 在开始时冻结筛选与日期，所有页只变更页码，文件名 SHALL 使用同一日期快照（来源 #7054）。

#### Scenario: 导出期间修改筛选
- **WHEN** 跨页导出时用户更改日期或其它筛选
- **THEN** 后续页仍使用开始时的参数，文件名日期也不随界面变化。

#### Scenario: 单页导出
- **WHEN** 结果只有一页且筛选未变
- **THEN** 数据内容与原有列格式保持一致；本项不新增分页重置功能。

### Requirement: T19 chat 桥接保留 agent_message 正文
桥接 SHALL 将 `agent_message` 中字符串正文，或按顺序拼接的 `input_text`/`text` 与 `encrypted_content` 文本片段转换为原位置的一条 user 消息（来源 #6964）。这里只转送自定义 provider 的明文内容，不引入解密能力。

#### Scenario: 子代理信封与任务正文分开存放
- **WHEN** 信封在 input_text、任务正文在 encrypted_content 片段
- **THEN** 两部分按原顺序送达 chat 上游，并清理跨越该消息的待附加 reasoning 状态。

#### Scenario: 空内容与既有工具能力
- **WHEN** agent_message 没有可提取文本，或同请求含既有 additional_tools、namespace 与工具结果
- **THEN** 空项不生成空 user 消息，其它既有能力不被覆盖。

### Requirement: T20 密钥配额重置同步目标状态
配额重置 SHALL 使用接口返回的 quota_used/status 更新操作开始时的密钥；仅当前仍选中同一密钥时同步编辑表单（来源 #7026）。

#### Scenario: 配额重置恢复可用
- **WHEN** 返回账号密钥 status 从停用变为 active
- **THEN** 列表与仍在编辑的该密钥状态同步，不依赖刷新；不硬编码假设每次都会变为 active。

#### Scenario: 等待期间切换密钥
- **WHEN** A 的重置请求未完成时用户选中 B
- **THEN** 返回后更新 A，不覆写 B 的配额或表单状态。

### Requirement: T21 自动刷新遵循所选间隔
自动刷新 SHALL 在所选秒数到达时触发，不额外多等一秒；渠道状态页加载完成后 SHALL 按当前所选间隔重置倒计时（来源 #7112）。

#### Scenario: 连续自动刷新与手动刷新
- **WHEN** 用户选择非默认间隔并触发自动或手动刷新
- **THEN** 后续倒计时仍使用所选间隔，每周期不多等待一秒。

#### Scenario: 自动刷新关闭
- **WHEN** 用户关闭自动刷新
- **THEN** 保持既有关闭行为，不因加载完成重新开启计时。

### Requirement: T22 资料与密码错误使用统一提取规则
个人资料与密码更新失败 SHALL 使用既有 extractApiErrorMessage 规则展示后端错误，并在不可提取时使用各自本地化兜底（来源 #7023）。

#### Scenario: 标准 API 错误与未知异常
- **WHEN** 后端返回标准错误消息，或异常没有可提取消息
- **THEN** 分别展示统一提取结果或对应兜底，避免只读 detail 导致有效错误丢失。

#### Scenario: 更新成功
- **WHEN** 资料或密码更新成功
- **THEN** 保留既有用户状态更新、成功提示与密码表单清空行为。

### Requirement: T23 代理部分导入后关闭弹窗刷新列表
代理导入弹窗 SHALL 记住本次打开期间是否发生过成功创建/复用；部分失败后关闭 SHALL 在存在成功数据时通知父列表刷新一次（来源 #7025）。

#### Scenario: 部分成功后关闭
- **WHEN** 导入结果同时包含失败与成功创建/复用，用户随后关闭弹窗
- **THEN** 保留失败详情并在关闭时发出一次 imported 通知，父列表可见已导入数据。

#### Scenario: 全失败、完整成功与重新打开
- **WHEN** 导入全失败、完整成功或弹窗重新打开
- **THEN** 全失败不触发虚假刷新；完整成功按既有路径通知且关闭不重复；重新打开清空上次待刷新状态。
