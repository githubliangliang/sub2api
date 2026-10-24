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
