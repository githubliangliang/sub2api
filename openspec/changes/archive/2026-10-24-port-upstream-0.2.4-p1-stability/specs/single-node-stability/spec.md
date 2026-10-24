## ADDED Requirements

### Requirement: S01 回放共享遵守不可变所有权并记住被拒内容
系统 SHALL 共享不可变回放正文，避免随会话增长反复复制完整历史；被拒的 encrypted_content 按会话记录并有 TTL/容量边界。

#### Scenario: 长会话与失败重试
- **WHEN** 会话持续增加轮次且某段加密历史被拒绝
- **THEN** 后续恢复不反复加入相同拒绝内容，不污染已共享正文；分配证据显示避免原有重复整段复制。

#### Scenario: 隔离、到期与 metadata 清理
- **WHEN** 不同 group/session 同时运行，记录到期或请求经过 metadata 清理
- **THEN** 记录不跨会话泄漏，到期被清理，清理不修改其它回放视图或丢正常输入。

### Requirement: S02 客户端取消及时终止上游且正确归因
系统 SHALL 在流式关闭 Body 前取消上游，并将明确的 WS 客户端取消与账号故障区分。

#### Scenario: 客户端提前退出
- **WHEN** HTTP/WS 客户端在上游未结束时取消
- **THEN** 转发可及时返回、释放连接资源，不因关闭顺序挂起；已产生的用量仍按现有规则落库。

### Requirement: S03 持久冷却同步不删除并发新封锁
系统 SHALL 按持久冷却与本地 generation/deadline 协调账号封锁，并保持模型级限制。

#### Scenario: 持久冷却被清除
- **WHEN** 调度读取到持久字段均已无效且本地快照未被并发修改
- **THEN** 可清陈旧账号 block；本地快照较旧或 DB 写失败时的 fail-open 行为必须有专门证据。

#### Scenario: 更新的封锁到达
- **WHEN** peek 后出现新的 generation，即使 deadline 相同
- **THEN** 清理 MUST NOT 删除新封锁；有效持久冷却与模型级 block 继续生效。

### Requirement: S04 未服务的失败请求释放会话槽
系统 SHALL 对失败的 Anthropic 会话注册及时注销，并保持成功会话的空闲超时规则。

#### Scenario: failover、取消或换组
- **WHEN** 请求没有被相应账号服务便失败、退出或切换账号/分组
- **THEN** 已注册失败账号的 session 槽被释放；注销幂等且不依赖已经取消的请求上下文。

#### Scenario: 成功或已计量的中断
- **WHEN** 请求成功或流中断前已形成可记录的用量
- **THEN** 保持既有会话和计量语义，不把它当未服务请求删除。

### Requirement: S05 工具及 bootstrap 转换保持完整性
系统 SHALL 保留发现的可执行工具、done 事件中的完整参数、合法 heartbeat 与历史 delegation，并正确限制 OpenCode session 转发范围。

#### Scenario: discovery 或 done-only 参数
- **WHEN** Chat fallback 收到新发现工具，或参数只在 done 中完整提供
- **THEN** 工具可继续调用、参数不丢失；已发送的完整 delta 不重复追加，命名冲突仍拒绝。

#### Scenario: bootstrap 与历史
- **WHEN** 合法 heartbeat/delegation 与可明确配对历史共存
- **THEN** 适当归一化；含糊孤儿 output、非法 envelope、重复 JSON 字段仍受原校验边界约束。

### Requirement: S06 Astra 目录、能力、转发与计价一致
系统 SHALL 为本簇支持的 Astra 名称/别名提供一致的能力、推理模式、工具元数据、缓存、指令和定价行为。

#### Scenario: 同步或续聊
- **WHEN** OAuth/API Key 同步目录或恢复 Astra 对话，包括上游模型列表暂不可用
- **THEN** 保留已知有效能力与 continuation；API Key 不被错误赋予 OAuth 专属协议能力。

#### Scenario: Ultra 元数据与独立服务档位
- **WHEN** 路由暴露本簇支持的推理能力或映射别名
- **THEN** 上游模型、instructions、pricing 与目录一致，不要求启用未立项的 Ultrafast/Fast/pinned-account 功能。

### Requirement: S07 自定义定价热重载与 none 来源映射
系统 SHALL 在内容变化时一致地重载已配置价格文件，并允许既有推理映射把 none 作为来源。

#### Scenario: 定价文件改变、损坏或删除
- **WHEN** 文件更新、内容无效或暂时缺失
- **THEN** 有效变化可生效；失败不发布半更新快照或丢失可用价格，恢复后能重新加载。

#### Scenario: 显式 none
- **WHEN** 请求显式 none 且配置了映射
- **THEN** 应用既有映射机制；未配置与其它来源的行为不变，不引入 deny 新策略。

### Requirement: S08 代理编辑、备份和回退保持一致
系统 SHALL 区分省略、显式清空和具体值，允许合法的有向共享备份，并支持重复过期回退。

#### Scenario: 只改名称或明确清空
- **WHEN** 更新只改代理名称，或显式提交 null 清空可空字段
- **THEN** 前者保留有效期/备份/预警，后者仅清指定字段；按合并后的结果验证回退配置。

#### Scenario: 共享备份、重复回退和错误输入
- **WHEN** 多个代理共享备份，或代理再次过期，或收到非法日期/列表数据
- **THEN** 合法引用可用且原代理归属保留；自引用/环与无效日期被拒，异常列表不导致页面崩溃或清空已有结果。

### Requirement: S09 日志持久化有独立保留策略
系统 SHALL 默认减少每请求 access log 入库，以独立有效保留期清理 system logs，同时保留警告、错误和审计。

#### Scenario: 旧配置和 SQLite 清理
- **WHEN** 旧配置未含新字段，或 SQLite 定时清理执行
- **THEN** 使用兼容默认值，清理 worker 实际运行并沿用本地 SQL；保留期合法，sink 写失败退避不丢。

### Requirement: S10 精简列表不破坏编辑与管理流程
系统 SHALL 在账号列表使用紧凑数据，并为编辑/批量操作获取必要详情，保持 fork 的菜单与 simple-mode 规则。

#### Scenario: 全选、批量、详情编辑
- **WHEN** 用户从精简列表执行全选、批量编辑或打开详情
- **THEN** 缺少的字段不被写回空值，凭证不额外暴露，原选择语义不变。

#### Scenario: 既有界面缺陷
- **WHEN** 子路由活动时折叠菜单、移除停用分组、刷新令牌失败、创建 Antigravity OAuth、设置注册隐藏或查看分页密钥过滤器
- **THEN** 折叠可用、停用关系可移除、失败账号保留选择、plan 保留、注册入口遵循设置、过滤器包含后续页密钥；操作菜单不越界。

### Requirement: S11 Claude 版本覆盖、指纹下限与 billing 一致
系统 SHALL 接受合法且不低于内置下限的 CLI 版本覆盖，并在指纹新建、缓存正常和自愈路径保持版本下限。

#### Scenario: 旧指纹与新配置
- **WHEN** 缓存/客户端版本较旧，或运维配置有效覆盖版本
- **THEN** 更新版本而保留其它身份字段；header/body/billing 使用一致版本，较新客户端不被降级。

#### Scenario: 非法覆盖
- **WHEN** 覆盖版本格式非法、含不允许后缀或低于内置下限
- **THEN** 使用内置有效值并给出明确日志，不写入无效持久身份。

### Requirement: S12 Redis 客户端升级兼容单机模式
系统 SHALL 使用目标 go-redis 客户端并保持 miniredis 模式与索引回收行为。

#### Scenario: 缓存与过期索引
- **WHEN** 执行调度缓存、会话注销、队列过期扫描
- **THEN** ZRangeArgs 范围和数量语义与原行为一致，连接池和取消相关回归通过，SQLite 依赖保持可用。
