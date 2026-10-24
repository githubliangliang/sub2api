# single-node-correctness Specification

## Purpose

保障个人单节点部署的认证、协议转换、用量结算、管理交互和异步生命周期正确，保留 SQLite、miniredis 与 simple mode 约束。

## Requirements

### Requirement: F01 多工具改名保持请求体完整
系统 MUST 满足：工具声明、强制工具选择和历史tool_use中的同名工具保持一致，改名不破坏JSON或无关内容。

#### Scenario: F01 不同长度与转义名称
- **WHEN** 同一请求含多个长度不同或有转义的工具名，且历史消息引用这些工具
- **THEN** 改写后JSON有效，所有引用对应同一映射，input字段不变

#### Scenario: F01 内置与未映射工具
- **WHEN** 请求含不应模拟的内置工具和未映射名称
- **THEN** 这些名称保持原样，最后工具缓存断点按既有规则处理

### Requirement: F02 白名单编辑保留显式模型映射
系统 MUST 满足：在新增白名单项会覆盖from到不同to的映射时提示并拒绝该新增，保留已有映射。

#### Scenario: F02 冲突添加
- **WHEN** 已有alias→real映射，用户尝试把alias加入identity白名单
- **THEN** 显示可读冲突提示，映射及已选白名单不变

#### Scenario: F02 合法添加
- **WHEN** 新增其它名称、自映射或空目标对应名称
- **THEN** 按既有规则添加，前后空白按规范化规则判断

### Requirement: F03 兼容入口遵守Claude Code降级分组
系统 MUST 满足：Claude Code限制分组配置fallback时，Chat Completions和Responses请求沿现有调度选择降级分组账号。

#### Scenario: F03 实际降级
- **WHEN** 两个兼容入口收到受限分组请求且fallback有效
- **THEN** 请求实际选中fallback账号并完成转发，不能仅把403改为另一个错误

#### Scenario: F03 拒绝边界
- **WHEN** 受限分组没有fallback或fallback链无效
- **THEN** 没有fallback仍403，无效链按现有错误契约拒绝且不转发原分组

### Requirement: F04 密钥使用指引匹配客户端限制
系统 MUST 满足：受Claude Code限制的分组只显示Claude客户端标签，分组限制改变时重置当前客户端选择。

#### Scenario: F04 分组切换
- **WHEN** 弹窗由普通OpenAI分组切换到Claude Code限制分组
- **THEN** 仅显示Claude标签并选中它，无过期Codex配置残留

#### Scenario: F04 普通分组
- **WHEN** 打开或切回不受限分组
- **THEN** 按平台恢复既有默认标签和客户端选项

### Requirement: F01 Grok CLI身份匹配官方交互版本

Grok CLI代理请求 SHALL 使用1.0.46默认身份与运行平台UA；包级最低版本1.0.13、最终transport最低preferred pin保持各自约束。额度探测与网关身份一致，回退官方API清除CLI专用头。

#### Scenario: 默认请求与旧覆盖值

- **WHEN** 调用CLI代理且未设版本或配置低于允许下限的值
- **THEN** 发出1.0.46、grok-pager、interactive及authenticate-response；包级和transport分别验证下限

#### Scenario: 有效覆盖与平台UA

- **WHEN** 使用符合相应层版本约束的覆盖值
- **THEN** UA与版本一致；amd64/arm64和darwin按Rust格式输出

#### Scenario: 官方API边界

- **WHEN** 独立访问api.x.ai或CLI访问拒绝后回退官方API
- **THEN** 独立API路径不被CLI-host helper覆盖；回退请求不携带新增CLI mode/authenticate头

### Requirement: F02 Antigravity客户端错误隐藏账号池身份

既有Antigravity Gemini出口 SHALL 将非failover上游错误规范化为JSON，仅保留code/status/message，清理项目引用、服务账号邮箱与敏感查询参数，details不返回客户端。

#### Scenario: JSON身份字段

- **WHEN** ForwardGemini上游错误正文包含projects引用、服务账号邮箱和details.metadata
- **THEN** 实际客户端响应中不含这些身份内容；HTTP状态保持，ops与调度错误处理保留

#### Scenario: 非JSON或空错误

- **WHEN** 上游错误为普通文本、空正文或无法提取message
- **THEN** 客户端仍获得合法Gemini JSON及对应状态；已知身份形态清理，空内容使用通用错误

#### Scenario: 正常响应与failover

- **WHEN** 上游成功或错误符合既有failover条件
- **THEN** 保留成功用量及原有failover语义，不因脱敏增加重放

### Requirement: F03 自定义错误码说明符合执行语义

中英文说明 SHALL 描述自定义错误码筛选常规账号错误处理，不承诺未选中错误统一返回500，也不宣称其决定retry/failover。

#### Scenario: 查看设置说明

- **WHEN** 用户以中文或英文查看自定义错误码设置
- **THEN** 说明空列表不筛选、重试/切号由独立网关逻辑决定、最终状态取决于路径与透传规则

#### Scenario: 保持配置行为

- **WHEN** 修改此次说明并使用原有配置
- **THEN** 不改变错误码默认值、调度与重试实现

### Requirement: S01 验证码尝试上限与重置令牌单次消费

验证码校验 SHALL 在比较前原子预占最多5次尝试；重置令牌 SHALL 只缓存SHA-256摘要，并以原子比较删除保证单次消费。

#### Scenario: 并发错误猜测

- **WHEN** 同一码被20个并发请求猜测，覆盖普通邮箱和通知邮箱
- **THEN** 至多5次进入码值比较；超过限制拒绝且计数不丢失

#### Scenario: TTL及重新发码

- **WHEN** 验证码过期、被重新发送或缓存操作失败
- **THEN** 过期/错误拒绝，不延长旧码TTL；新码计数重新开始且辅助key不永久残留

#### Scenario: 哈希与重发

- **WHEN** 生成重置链接并在冷却允许后再次申请
- **THEN** 缓存无明文token，新链接替换旧链接；旧部署明文缓存链接拒绝并可重新申请

#### Scenario: 竞争消费

- **WHEN** 多个并发请求携带同一有效token，或有人提交错误token
- **THEN** 有效token只有一个消费者成功；错误token不能删除有效token；缓存错误不得放行

#### Scenario: 内置Redis与调用链

- **WHEN** redis.enabled=false并经实际密码重置服务消费token
- **THEN** Lua和TTL在miniredis通过；注册/绑定邮箱/通知邮箱及所有接口实现正常编译运行

### Requirement: S02 账号优先级快捷调整

账号列表 SHALL 提供优先级加减与直接输入，限制1–9999，合并连续操作且只更新目标账号priority。

#### Scenario: 连续按钮操作

- **WHEN** 用户在450ms窗口内连续按加减按钮
- **THEN** 更新草稿并合并一次保存；减小数字表示提高优先级；不能越界

#### Scenario: 输入与取消

- **WHEN** 点击数字编辑并使用Enter、blur、Esc或方向键
- **THEN** Enter/blur只提交一次有效值，Esc恢复已保存值，方向键调整；无效输入不产生非法请求

#### Scenario: 保存失败与重复操作

- **WHEN** 请求在途或失败
- **THEN** 在途不能重复提交；失败恢复服务器值并显示错误，其他账号字段保持

#### Scenario: 卸载与可访问性

- **WHEN** 存在待提交按钮修改时翻页/卸载，或通过键盘/触屏操作
- **THEN** 待保存修改只提交一次到原账号；晚到响应不更新错行；控件可聚焦且触屏可用

### Requirement: S03 API密钥按分组名稳定排序

密钥列表 SHALL 在数据库分页前按分组名排序，升降序均未分组最后，使用ID同向次排序并保留过滤和用户隔离。

#### Scenario: 名称与分页

- **WHEN** 分组ID顺序与名称顺序不同且密钥跨多个分页
- **THEN** 升降序按名称跨页正确，同名按ID同向稳定排序，未分组始终最后

#### Scenario: 过滤与隔离

- **WHEN** 同时使用搜索/状态/分组或未分组过滤，并存在其他用户/软删除记录
- **THEN** 只返回当前用户可见且匹配记录，总数准确，group预加载保持

#### Scenario: 前端交互

- **WHEN** 点击密钥页分组列表头
- **THEN** 发送group排序参数并切换升降序，展示服务端排序结果

### Requirement: S01 首次安装管理员校验统一

系统 SHALL 满足以下已有功能正确性契约：只有空库真正创建管理员时，缺省邮箱随机生成，显式邮箱可登录，密码8–72字节；CLI/Web/AUTO_SETUP 共用校验。

#### Scenario: 目标行为

- **WHEN** 使用首次安装管理员校验统一涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 只有空库真正创建管理员时，缺省邮箱随机生成，显式邮箱可登录，密码8–72字节；CLI/Web/AUTO_SETUP 共用校验。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 已有管理员或已有普通用户时先跳过创建，不用遗留无效配置阻断启动，不重置密码；保留安装互斥和锁文件。

### Requirement: S02 旧版函数调用与拒答转换完整性

系统 SHALL 满足以下已有功能正确性契约：旧函数调用按同名队列配对，稳定生成且避开已有 call ID；refusal 在 Chat message/delta/缓冲终态全部保留。

#### Scenario: 目标行为

- **WHEN** 使用旧版函数调用与拒答转换完整性涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 旧函数调用按同名队列配对，稳定生成且避开已有 call ID；refusal 在 Chat message/delta/缓冲终态全部保留。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 混合现代 tool_calls、同名多次调用、不配对结果和纯拒答均有测试；不双计文本或工具调用。

### Requirement: S03 显式工具变更 beta 保留及旧前置

系统 SHALL 满足以下已有功能正确性契约：只保留客户端显式请求的 structured-outputs、mid-conversation-tool-changes、inline-tools 兼容 token。

#### Scenario: 目标行为

- **WHEN** 使用显式工具变更 beta 保留及旧前置涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 只保留客户端显式请求的 structured-outputs、mid-conversation-tool-changes、inline-tools 兼容 token。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** effectiveDropSet 始终优先，未知 beta 不通过，不推断 companion token；count_tokens 原有不同规则保持。

### Requirement: S04 WebSocket 逐轮分组价格与图片用量

系统 SHALL 满足以下已有功能正确性契约：第二轮起获取同 Key 同组的新价格快照并绑定本轮；输入/输出图片 tokens 从 usage 或 tool_usage 回填并准确累计。

#### Scenario: 目标行为

- **WHEN** 使用WebSocket 逐轮分组价格与图片用量涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 第二轮起获取同 Key 同组的新价格快照并绑定本轮；输入/输出图片 tokens 从 usage 或 tool_usage 回填并准确累计。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 不更新连接用户/订阅/Key 限额身份；换组/换平台/查询失败回退建连快照；交错轮次不串价、不重复累计。

### Requirement: S05 受约束的上游兼容重试

系统 SHALL 满足以下已有功能正确性契约：特定 thinking_signature_invalid 三段诊断匹配既有一次性恢复；auto 且已知支持 Responses 的 API Key 对完整、无歧义 input 类型拒绝回退原始 Chat，并沿用起始时间。

#### Scenario: 目标行为

- **WHEN** 使用受约束的上游兼容重试涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 特定 thinking_signature_invalid 三段诊断匹配既有一次性恢复；auto 且已知支持 Responses 的 API Key 对完整、无歧义 input 类型拒绝回退原始 Chat，并沿用起始时间。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** Responses 形状入站、已输出/已取消、截断或读失败、重复字段、含用量/输出及其它错误不得触发危险重试；不降级强制模式。

### Requirement: S06 OAuth 搜索历史回放声明

系统 SHALL 满足以下已有功能正确性契约：OAuth 回放搜索历史时补 cached-only web_search，Lite 放 additional_tools；原本无工具时保持 tool_choice none，compaction_trigger 仍最后。

#### Scenario: 目标行为

- **WHEN** 使用OAuth 搜索历史回放声明涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** OAuth 回放搜索历史时补 cached-only web_search，Lite 放 additional_tools；原本无工具时保持 tool_choice none，compaction_trigger 仍最后。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 已有工具声明不重复；保留显式强制选择；API Key 及 /responses/compact 不套此重写；map/raw/WS 一致且保留大整数。

### Requirement: S07 组合路由公开模型目录

系统 SHALL 满足以下已有功能正确性契约：启用精确路由的公开别名加入对应 /models 与 Codex Responses 目录，去重并保持 endpoint 区分。

#### Scenario: 目标行为

- **WHEN** 使用组合路由公开模型目录涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 启用精确路由的公开别名加入对应 /models 与 Codex Responses 目录，去重并保持 endpoint 区分。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 禁用/glob 路由不凭空展开；读取失败回退现有目录；原自定义列表与组隔离不放宽。

### Requirement: S08 删除分组同步清理风控范围

系统 SHALL 满足以下已有功能正确性契约：组删除与审核范围更新在同一事务；并发配置保存不可重新写入已删除组。

#### Scenario: 目标行为

- **WHEN** 使用删除分组同步清理风控范围涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 组删除与审核范围更新在同一事务；并发配置保存不可重新写入已删除组。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 缺失/畸形设置不阻断删除也不覆盖原值；真正 DB 错误回滚；保留顺序与其它配置字段。

### Requirement: S09 分组倍率/RPM 编辑异步隔离

系统 SHALL 满足以下已有功能正确性契约：打开/切组/关闭/卸载使旧加载失效并清空旧行；RPM 为非负整数，倍率为有限正数。

#### Scenario: 目标行为

- **WHEN** 使用分组倍率/RPM 编辑异步隔离涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 打开/切组/关闭/卸载使旧加载失效并清空旧行；RPM 为非负整数，倍率为有限正数。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 迟到 success/error/finally 均不改当前组或 loading；不把不合法输入保存，不误合旧 F12 未完成的其它表单。

### Requirement: S10 异步页面、草稿、确认与弹窗隔离

系统 SHALL 满足以下已有功能正确性契约：每个视图按对象/请求代次接收响应，保留未提交草稿；同实例二次验证共享待定 Promise；公告仅成功标记后关闭。

#### Scenario: 目标行为

- **WHEN** 使用异步页面、草稿、确认与弹窗隔离涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 每个视图按对象/请求代次接收响应，保留未提交草稿；同实例二次验证共享待定 Promise；公告仅成功标记后关闭。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 错误/finally/nextTick/分页回包也受失效保护；换用户时清理敏感草稿；不启用仍被 SQLite 跳过的清理或定时测试 worker。

### Requirement: S11 GPT-6.1 Sol 上游订阅长上下文统计

系统 SHALL 满足以下已有功能正确性契约：实际凭据为 OpenAI OAuth-like 且最终上游模型为 GPT-6.1 Sol 时豁免模型价卡长上下文附加费，用户与账号成本口径一致。

#### Scenario: 目标行为

- **WHEN** 使用GPT-6.1 Sol 上游订阅长上下文统计涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 实际凭据为 OpenAI OAuth-like 且最终上游模型为 GPT-6.1 Sol 时豁免模型价卡长上下文附加费，用户与账号成本口径一致。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** API Key、其它模型仍原价卡；显式自定义区间价格不豁免；不引入付费订阅购买/续费或新价格体系。

### Requirement: S12 Gemini 既有路径的错误、额度和脱敏修复

系统 SHALL 满足以下已有功能正确性契约：第三方/Vertex 不套本地模拟日配额，第三方429仅显式重试时间或原全局fallback；缺失缓存在错误策略前识别并返回请求错误，不封号换号；本地Google错误正确归类；AQ凭据文本脱敏。

#### Scenario: 目标行为

- **WHEN** 使用Gemini 既有路径的错误、额度和脱敏修复涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 第三方/Vertex 不套本地模拟日配额，第三方429仅显式重试时间或原全局fallback；缺失缓存在错误策略前识别并返回请求错误，不封号换号；本地Google错误正确归类；AQ凭据文本脱敏。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 官方AI Studio/OAuth原额度保持；非缓存鉴权错误仍正常处理；不实现缓存创建/续期/删除。Flash促销价仅按来源2026-12-31有效窗口核对，过期需恢复标准兜底，不默许永久优惠价。

### Requirement: S13 监控配置补充已有平台入口

系统 SHALL 满足以下已有功能正确性契约：用户可加入一个当前缺失且本地后端支持的平台，初始无模型，沿原保存接口持久化。

#### Scenario: 目标行为

- **WHEN** 使用监控配置补充已有平台入口涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 用户可加入一个当前缺失且本地后端支持的平台，初始无模型，沿原保存接口持久化。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 不重复添加、不加入未实现平台，saving 时禁止修改；不是新增供应商平台。

### Requirement: S14 Responses 无 namespace 时避免重建

系统 SHALL 满足以下已有功能正确性契约：无修改时返回原正文；需要剥离时仅首次分配并复制未变字节段，减少单节点长历史开销。

#### Scenario: 目标行为

- **WHEN** 使用Responses 无 namespace 时避免重建涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 无修改时返回原正文；需要剥离时仅首次分配并复制未变字节段，减少单节点长历史开销。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 保留工具调用 namespace 的既有规则不变；空白/数字/转义/非对象和错误回退保真，需来源 benchmark 或 AllocsPerRun 支撑收益。

### Requirement: S15 Chat 流心跳保活

系统 SHALL 满足以下已有功能正确性契约：仅 ping 到来时向 Chat 客户端写标准 SSE 注释保活；后续语义事件正常转换。

#### Scenario: 目标行为

- **WHEN** 使用Chat 流心跳保活涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 仅 ping 到来时向 Chat 客户端写标准 SSE 注释保活；后续语义事件正常转换。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 不泄露 event:ping，写失败停止；错误事件不得吞掉；心跳后失败按已经写出响应的规则处理，不重复响应或危险重放。

### Requirement: S16 前端依赖安全维护与旧待办归并

系统 SHALL 满足以下已有功能正确性契约：定向更新已有DOMPurify、axios、Vue及source-map-js受影响解析，保留pnpm锁文件一致性；sanitize调用、API错误处理与前端构建回归通过；审计重新记录目标包和剩余项。

#### Scenario: 目标行为

- **WHEN** 使用前端依赖安全维护与旧待办归并涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** 定向更新已有DOMPurify、axios、Vue及source-map-js受影响解析，保留pnpm锁文件一致性；sanitize调用、API错误处理与前端构建回归通过；审计重新记录目标包和剩余项。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 不宣称103条全部已解决；xlsx无npm修复路径及Vite/Vitest/tinypool工具链大版本另列旧维护范围。例外只能按实际不可利用条件和负责人复核，不自动续期、不扩大支付功能。

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

### Requirement: F01 调度 RPM 配置投影
系统 MUST 满足：账号 base_rpm、rpm_strategy、rpm_sticky_buffer 经过缓存投影后仍参与候选限流。

#### Scenario: F01 tiered配置经过缓存
- **WHEN** base_rpm=10、rpm_strategy=tiered、buffer=1的账号已用11 RPM，序列化后再解码
- **THEN** 三个配置字段保留，候选被拒绝；sticky_exempt同样负载仅允许已有粘性

### Requirement: F02 OAuth 重授权保留账号设置
系统 MUST 满足：更新认证凭据时保留原 model_mapping 等非认证设置，新认证字段覆盖旧值。

#### Scenario: F02 部分OAuth凭据
- **WHEN** 旧凭据含model_mapping及旧access_token，重授权只返回新access_token
- **THEN** model_mapping保持原值，access_token替换为新值，未返回的配置字段不丢失

### Requirement: F03 账号身份与错误状态
系统 MUST 满足：非法 UA 回退时保留已解析版本；读到用量快照不能抹掉 refresh token 拒绝错误。

#### Scenario: F03 非法UA与指定版本
- **WHEN** 配置已解析的CLI版本，UA含CR/LF或不能配对的指纹
- **THEN** 回退为合法规范UA，使用已解析版本而非丢弃该版本

#### Scenario: F03 快照不证明恢复
- **WHEN** 账号因refresh token拒绝而报错，读取缓存用量快照成功
- **THEN** 不清除账号错误；只有既有真实恢复流程可清除

### Requirement: F04 关闭 thinking 的桥接语义
系统 MUST 满足：Anthropic thinking.type=disabled 优先于 output_config.effort，两条 OpenAI 桥均输出 none。

#### Scenario: F04 显式禁用优先
- **WHEN** thinking.type=disabled且output_config.effort=high，分别走Responses和Chat桥
- **THEN** 两者reasoning effort为none，Responses不请求summary

#### Scenario: F04 未指定thinking
- **WHEN** 未给thinking及effort，或effort=max
- **THEN** 前者保持medium，后者映射xhigh

### Requirement: F05 显式 beta 头兼容
系统 MUST 满足：mimic 路径保留显式 structured-outputs beta；OpenAI 出站保留多智能体 beta 并剔除旧 Responses token。

#### Scenario: F05 结构化输出beta
- **WHEN** mimic请求显式携带structured-outputs-2025-11-13且未被策略drop
- **THEN** 该token保留；未知token不顺带放行；策略drop仍优先

#### Scenario: F05 Responses混合beta
- **WHEN** OpenAI请求携带旧Responses beta及需要保留的多智能体beta
- **THEN** 只删除旧Responses token；空beta不自动注入其它token

### Requirement: F06 Responses 消息与最终正文
系统 MUST 满足：角色消息带 type=message；终止事件空正文时回填已积累的文本。

#### Scenario: F06 角色项类型
- **WHEN** Chat system/user/assistant文本分别转换为Responses input
- **THEN** 每个角色项都含type=message及原角色

#### Scenario: F06 空终态回填
- **WHEN** 已积累delta文本hello，terminal output数组非空但message只有空白文本
- **THEN** 返回hello；若terminal已有world则保留world，工具output不被替换

### Requirement: F07 Alpha search 成功判定
系统 MUST 满足：仅在 response.completed 且状态成功后返回可计费用量。

#### Scenario: F07 成功才返回
- **WHEN** alpha search SSE含response.completed，response.status=completed
- **THEN** 返回解析结果并允许成功用量路径处理

#### Scenario: F07 失败或截断
- **WHEN** SSE只有文本delta/DONE/EOF，或含failed/incomplete/error，或completed缺response
- **THEN** 返回错误，不返回可计费成功结果

### Requirement: F08 未来重置时间前保持配额暂停
系统 MUST 满足：陈旧但有明确未来 reset 的额度快照继续用于暂停，达到 reset 后放行。

#### Scenario: F08 旧快照未来重置
- **WHEN** 额度已到暂停阈值，快照过期，但规范窗口reset仍在未来
- **THEN** 保持暂停并保留已知reset时间

#### Scenario: F08 失去暂停依据
- **WHEN** reset已经到达，或陈旧快照没有已知未来reset
- **THEN** 按既有fail-open规则不再依靠该旧百分比暂停

### Requirement: F09 TOTP 输入与计时器
系统 MUST 满足：验证码输入同步、错误文本可读，弹窗销毁后迟到响应不重启计时器。

#### Scenario: F09 输入与错误
- **WHEN** TOTP设置组件更换输入值并提交，API返回有可读消息的错误
- **THEN** 提交当前验证码，展示可读API错误而非对象字符串

#### Scenario: F09 销毁后迟到响应
- **WHEN** TOTP弹窗请求进行中关闭/卸载，随后返回成功并推进fake timer
- **THEN** 不重新建立倒计时；重新打开的独立会话正常工作

### Requirement: F10 公共表单交互
系统 MUST 满足：分页跳转、对话框标题 ID、复制失败、标签 Tab、IME、下拉键盘与搜索、日期关闭和跨午夜选择保持一致。

#### Scenario: F10 数字跳页
- **WHEN** 分页jumpPage由数值输入产生数字2，提交跳转
- **THEN** 不会调用数字.trim导致异常；跳到第2页

#### Scenario: F10 对话框唯一ID
- **WHEN** 同时挂载两个BaseDialog
- **THEN** 标题ID不同，各自aria-labelledby指向自身标题

#### Scenario: F10 复制失败
- **WHEN** 旧剪贴板fallback的execCommand抛错
- **THEN** 返回失败状态而不抛未处理异常

#### Scenario: F10 标签Tab
- **WHEN** 空标签输入按Tab；随后非空新标签按Tab
- **THEN** 空值允许焦点移动；非空值添加一次标签并阻止该次Tab默认动作

#### Scenario: F10 搜索IME
- **WHEN** 中文compositionstart到compositionend期间连续输入拼音
- **THEN** 组合中不发中间搜索，完成后按最终文本防抖搜索

#### Scenario: F10 Select筛选焦点
- **WHEN** 下拉已打开后搜索过滤选项，再按Enter；另用键盘打开无搜索下拉
- **THEN** 焦点移动到过滤后首个可用项并选中；无搜索下拉取得键盘焦点

#### Scenario: F10 取消日期草稿
- **WHEN** 改日期范围但按Escape/点击外部关闭，随后重开
- **THEN** 恢复父组件已应用范围，不能复用未应用草稿

#### Scenario: F10 跨午夜日期
- **WHEN** 组件在本地午夜前挂载，午夜后点击今天或最近范围预设
- **THEN** 按点击时日期计算，结束日期/输入max不缓存前一天

### Requirement: F11 代理管理小修
系统 MUST 满足：同代理重复测试共享进行中的结果，过期边界展示与后台一致。

#### Scenario: F11 代理重复测试
- **WHEN** 同一代理已有进行中测试，再发一次；同时测试不同代理
- **THEN** 同一代理不重复请求，不同代理结果彼此隔离

#### Scenario: F11 到期边界
- **WHEN** 代理expires_at等于当前时间
- **THEN** 按已到期展示；没有expires_at的代理不显示虚假过期

### Requirement: F12 配置表单错误与输入边界
系统 MUST 满足：设置和 Antigravity 映射加载失败可重试，替换分组显示错误，倍率必须正数，RPM override 只接受非负整数。

#### Scenario: F12 可重试加载
- **WHEN** 第一次Settings或Antigravity映射加载失败，第二次成功
- **THEN** 失败不写loaded=true或永久缓存空数组，第二次会重新请求

#### Scenario: F12 替换分组失败
- **WHEN** GroupReplace API返回有消息的错误
- **THEN** 用户看到该消息，不能显示成功状态

#### Scenario: F12 倍率与RPM校验
- **WHEN** 分别输入倍率0/-1/0.5，以及RPM空值/0/1.5/2
- **THEN** 倍率只接受0.5；RPM接受0和2，空值/1.5拒绝且不提交

### Requirement: F13 异步弹窗状态隔离
系统 MUST 满足：错误详情、用户 API Key、临时不可调度状态仅接受当前对象请求，分组弹窗释放监听和待执行搜索。

#### Scenario: F13 A到B乱序返回
- **WHEN** 错误详情、UserApiKeys和TempUnsched弹窗从对象A切到B，B先返回A后返回
- **THEN** 最终只显示B的数据/错误，A返回不修改B的loading状态

#### Scenario: F13 卸载清理
- **WHEN** 分组RPM/倍率弹窗已注册document点击监听并有待执行搜索，随后卸载
- **THEN** 清除监听与定时搜索，不再发送卸载后搜索

### Requirement: F14 用量与价格显示
系统 MUST 满足：CSV 保留缺失推理力度的单独减号占位符；区间输入正确解析科学计数法；空闲窗口有 reset 时显示倒计时。

#### Scenario: F14 CSV占位符
- **WHEN** 导出缺推理力度的字段值为单独字符-，并另导出以=或其他公式前缀开头的内容
- **THEN** 单独-原样保留，其它危险公式前缀继续按已有规则转义

#### Scenario: F14 科学计数法
- **WHEN** IntervalRow token边界输入1e5
- **THEN** 解析为100000而不是parseInt得到的1

#### Scenario: F14 空闲倒计时
- **WHEN** utilization=0、showNowWhenIdle=true且resetsAt有未来时间
- **THEN** 显示真实倒计时；未给reset时才显示立即可用

### Requirement: F15 请求 ID 和配额诊断
系统 MUST 满足：已识别的 Responses input item ID 超过64字节时移除；Grok 冷却期仍能查询配额。

#### Scenario: F15 过长ID
- **WHEN** 已识别message/reasoning/tool-call项携带长度65且前缀合法的ID
- **THEN** 移除ID；长度64的合法ID保留，未知类型遵守本地原逻辑

#### Scenario: F15 冷却查询配额
- **WHEN** 有效Grok认证账号处于调度冷却中，管理员发起配额查询
- **THEN** 使用手动测试认证通道完成查询，不受模型调度门阻断；无效凭据仍失败

### Requirement: Preserve personal SQLite deployment
系统 MUST 保留SQLite唯一数据库、可关闭Redis、现有菜单和用量去重；不得引入支付或上游新迁移。

#### Scenario: Scope verification
- **WHEN** 审核相对实施起点的最终diff与真实SQLite运行结果
- **THEN** 无支付/插件/新平台扩展，无已应用迁移修改，成功用量只写入一次
