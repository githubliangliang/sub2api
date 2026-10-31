# single-node-stability Specification

## Purpose

控制个人单节点部署中的请求回放、客户端取消、冷却同步、日志、代理、价格和 Redis 生命周期，避免资源泄漏及状态丢失。

## Requirements

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

### Requirement: S01 Responses 正文在仅 done/terminal 携带时也要恢复
桥接 SHALL 在增量缺失、正文仅出现在 `response.output_text.done` 或终态 response 的 message output 时恢复助手文本，并保持 Anthropic block 与 message 生命周期完整（来源 #6925）。

#### Scenario: 正文只在 done 或终态响应
- **WHEN** 上游没有文本增量，仅在 done 或支持的终态事件中给出 message 文本
- **THEN** 下游按顺序收到全部可恢复文本及完整的 start/delta/stop 事件；工具参数不得被当作助手正文。

#### Scenario: done 补齐增量尾部
- **WHEN** 已输出的文本是同一 output_index/content_index 的 done 文本前缀
- **THEN** 只补发尚未输出的后缀；done 文本与增量分歧时不覆盖或重发既有文本。

#### Scenario: 重复事件与索引差异
- **WHEN** 已通过增量或 done 输出正文，随后收到重复 done、索引不匹配的 done 或终态 output
- **THEN** 不重复输出，也不凭不匹配的索引追加无法确认的文本；message_stop 后不再追加内容。

#### Scenario: 工具与 thinking 共存
- **WHEN** 恢复正文时存在工具块或 thinking 签名，或本轮只有工具调用
- **THEN** 保留工具与签名语义；纯工具轮次不凭空生成文本，恢复文本在最终收尾后仍保留。

### Requirement: S01 Chat 桥接的 system/developer 角色规范化
Responses 转 Chat Completions SHALL 合并开头连续的 system/developer 为单条 leading system，并将中途 system/developer 改为 user，保留原位置与正文（来源 #7094）。

#### Scenario: instructions 与 leading developer 同时存在
- **WHEN** 转换后开头连续出现多条 system/developer
- **THEN** 按原顺序以空行连接非空文本，生成至多一条 leading system。

#### Scenario: 中途插入 developer 通知
- **WHEN** system/developer 出现在首条非指令消息之后
- **THEN** 改为 user，不上移，不丢正文，不改变其它消息和工具结果的相对顺序。

#### Scenario: 已符合角色规则
- **WHEN** 仅有一条前导 system，或没有 system/developer
- **THEN** 保留既有内容；单条前导消息不因合并而重新格式化。

### Requirement: S02 WS 连接池容量与系数一致
系统 SHALL 对旧池与 v2 ctx_pool 使用同一动态容量开关和账号类型系数；
开启动态容量时上限 SHALL 为 `min(max_conns_per_account, ceil(concurrency * factor))`（来源 #7064）。
连接数系数 MUST NOT 放大在飞请求的账号并发槽。

#### Scenario: 当前默认值与显式覆盖
- **WHEN** 并发为3、硬上限为20，分别使用当前默认系数1.0和显式系数1.5
- **THEN** 容量分别为3和5；OAuth/API Key分别使用自身系数，显式值不被默认值覆盖。

#### Scenario: 硬上限与动态开关
- **WHEN** 计算值超过硬上限，或关闭动态容量
- **THEN** 使用硬上限；v2中账号并发小于等于0时仍不可调度，不能借关闭动态容量放行。

### Requirement: S02 WS 默认系数调优与容量修复分开交付
系统 SHALL 保留 OAuth/API Key 默认系数 1.0，示例配置与默认值一致；#7043 的数值 5.0 方案未采用，不影响 #7064 容量公式的独立验收。显式系数 SHALL 按既有动态容量和硬上限规则生效，不放大在飞请求的账号并发槽。

#### Scenario: 默认系数与显式覆盖
- **WHEN** 并发为 3、硬上限为 20，分别使用默认系数和显式系数 1.5
- **THEN** OAuth/API Key 默认容量均为 3，显式 1.5 的容量为 5；示例配置保持默认 1.0。

#### Scenario: 未采用 5.0 的独立决策
- **WHEN** 复核默认系数方案与容量公式的完成状态
- **THEN** 记录默认 1.0 的决策已完成、数值 5.0 未采用；不将容量公式修复标记为失败，也不把未采集的业务会话和资源数据作为扩大默认连接预算的依据。

### Requirement: S02 WS 常驻读循环维持连接且隔离轮次
需要 reader loop 的连接 SHALL 在空闲期继续读取以响应 ping，按顺序向当前租约交付消息，并剔除已关闭或带残留数据的空闲连接（来源 #6965）。不需要 reader loop 的连接 SHALL 保留既有读与健康检查路径。

#### Scenario: 空闲 ping 与消息顺序
- **WHEN** 真实 WS 对端在空闲期发送 ping，随后租约内发送多个数据帧
- **THEN** 对端收到 pong，数据帧按原顺序交付；读循环不与另一个连接读取者争抢数据。

#### Scenario: 关闭与读超时
- **WHEN** 上游关闭连接或当前读操作超时
- **THEN** 等待读取者获知相应原因，连接退出可复用集合，已缓冲消息按来源语义先交付；本地主动关闭不误记为上游关闭。

#### Scenario: 空闲残留数据与重试
- **WHEN** 空闲连接收到数据，或轮次重试要求新连接
- **THEN** 后续租约不复用带残留数据的连接，重试取得新连接；满容量与排队交接也不得把旧轮次数据交给新请求。

#### Scenario: 后台探活与租用竞态
- **WHEN** 后台探活遇到慢 pong，或检查后连接已被租用
- **THEN** 按探活超时容忍慢 pong，不关闭已交给租约的连接；ping 不阻塞正常写入。

### Requirement: S03 内置工具与客户端工具混用时丢弃内置工具
Antigravity 的 Claude 转换、兼容请求与 Gemini 直入路径 SHALL 在存在可转发客户端 function tools 时移除内置搜索与代码执行工具，并停止注入旧 `IncludeServerSideToolInvocations` 策略（来源 #6689）。

#### Scenario: shell 与 web_search/code_execution 混用
- **WHEN** 请求同时携带客户端工具与内置工具
- **THEN** 保留客户端声明，去掉 `googleSearch`/`codeExecution`；不因已丢弃的 web_search 强制切换搜索模型或 request type；混合请求清理两种命名的旧 invocation 标志。

#### Scenario: 仅内置工具或无工具
- **WHEN** 没有可转发的客户端 function tools
- **THEN** 既有内置工具与搜索模型选择不变；Claude 转换的无工具请求仍保留上游要求的 toolConfig。

### Requirement: S04 部分刷新返回账号并展示 warning
Antigravity 部分刷新 SHALL 在响应中同时携带 account 与 warning，前端 SHALL 用返回账号更新列表并展示告警（来源 #7076）。

#### Scenario: missing_project_id_temporary
- **WHEN** 刷新部分成功且返回该 warning
- **THEN** 列表更新对应账号，warning 可见，不将其误报为无告警的完整成功。

#### Scenario: 完整成功与失败
- **WHEN** 刷新完整成功无 warning，或刷新接口失败
- **THEN** 分别保持既有成功更新或错误提示路径，失败时不伪造账号更新。

### Requirement: S05 DeepSeek 空映射按冻结名单校验并归一上下文后缀
DeepSeek 空映射 SHALL 按 #6977 在冻结 v0.2.5 的名单判定：`deepseek-flash`、`deepseek-v4-pro`、`deepseek-v4-flash`、`deepseek-v4-flash-vision-exp`、`deepseek-v4-pro-0813`。比较 SHALL 忽略大小写与首尾空白，并按既有规则去除 `[1m]` 后缀；出站也 SHALL 去除该客户端上下文后缀。本要求不新增模型目录或计费规则。

#### Scenario: 未知模型名
- **WHEN** 非透传 DeepSeek 账号无映射且请求模型不在名单内
- **THEN** 该账号在本地被排除，不把未知名发送到该账号上游，也不由此触发模型冷却。

#### Scenario: 合法名与上下文后缀
- **WHEN** 请求 `deepseek-flash[1m]` 或其它名单内模型的等价形式
- **THEN** 支持性校验通过，出站模型去掉 `[1m]`。

#### Scenario: 显式映射与原有例外
- **WHEN** 账号已配置映射、启用透传，或输入模型为空
- **THEN** 分别沿用映射判定、透传放行与上层必填校验；其它平台不受名单约束。

### Requirement: S06 Grok 媒体并发槽按请求释放
系统 SHALL 在取得媒体账号槽后即接管释放责任，所有返回、预检失败、取消、切换账号与转发结束路径 SHALL 至多释放一次已取得的槽（来源 #6661）。

#### Scenario: 取得槽后请求被拒绝
- **WHEN** 调度已取得槽，但后续资格预检失败、客户端取消或请求被拒绝
- **THEN** 槽被释放；重试下一账号前不继续持有上一账号的槽。

#### Scenario: 成功转发与未取得槽
- **WHEN** 正常转发结束，或选择结果没有取得槽
- **THEN** 前者释放一次，后者不错误释放别的请求的槽。

### Requirement: S06 视频查询固定使用任务所属账号
视频查询 SHALL 只准入已认证的任务所属账号，禁止通用粘性逃逸与回退改写任务归属，也不得将视频归属 TTL 替换成文本会话 TTL（来源 #6661）。

#### Scenario: 所属账号繁忙或健康指标触发逃逸
- **WHEN** 查询已有视频，所属账号没有空闲槽或命中普通粘性逃逸条件
- **THEN** 不查询其它账号；等待槽不刷新或缩短视频归属 TTL。

#### Scenario: 所属账号不可用
- **WHEN** 已绑定的任务所属账号无法被准入
- **THEN** 返回既有视频未找到语义，不回退到另一个账号；普通非任务查询的调度行为不变。

### Requirement: S08 客户端 cache_control 与 count_tokens 断点上限
Claude OAuth system 规范化 SHALL 保留客户端 cache_control；count_tokens mimic 出口 SHALL 在注入后使用既有收敛规则将全请求断点限制为最多 4 个（来源 #6943）。

#### Scenario: 注入关闭且客户端自带断点
- **WHEN** 客户端 system 包含 cache_control，本地 system 注入关闭且总断点数未超限
- **THEN** 断点及其属性保留，不因 system 规范化而删除。

#### Scenario: count_tokens 注入后超过四个断点
- **WHEN** system、messages 与 tools 注入合计产生 5 个或更多断点
- **THEN** 出站至多 4 个，按既有优先级收敛；不更改文本、模型映射或缓存计费口径。

#### Scenario: 已有注入锚点
- **WHEN** 本地 system 注入已启用且生成稳定缓存锚点
- **THEN** 锚点继续保留，仅在总断点超限时参与既有收敛。

### Requirement: S08 消息级 output_config 与最终 beta 一致
系统 SHALL 依据最终出站 `anthropic-beta` 是否含 `mid-conversation-output-config-2026-07-01` 决定保留 `messages[].output_config`；OAuth mimic beta 集合 SHALL 包含该 token（来源 #6995）。顶层 output_config/effort 不受本要求删除。

#### Scenario: 最终 beta 包含所需 token
- **WHEN** 最终出站 header 含该完整 token
- **THEN** 消息级 output_config 和相应控制消息原样保留。

#### Scenario: 缺少 beta 的空 system 控制消息
- **WHEN** 最终 header 不含 token，带 output_config 的 system 消息正文缺失、null、空字符串、空数组或只有空 text 块
- **THEN** 整条删除；有正文的 system、user/assistant 只移除消息级 output_config，保留其余字段与顺序。

#### Scenario: 顶层配置与未知内容块
- **WHEN** 只有顶层 output_config，或 system 含图片/未知非 text 块
- **THEN** 顶层配置不变；非 text 块视为有正文，不连带删除消息；无消息级字段时 body 不做字节改写。

### Requirement: S09 Codex 配额窗口与重置锚点一致
调度 SHALL 优先使用规范 5h/7d 用量，缺失窗口才按既有 Normalize 的窗口长度分类回退到 primary/secondary，并匹配同一窗口的重置数据。绝对 reset 时间优先；相对 reset 秒数 SHALL 锚定 `codex_usage_updated_at`（来源 #7074）。

#### Scenario: 规范字段与历史字段并存
- **WHEN** 规范字段存在但与 raw primary/secondary 不同，或 raw 两窗口顺序颠倒
- **THEN** 规范值优先，缺失项按窗口分类补齐，不固定将 primary 当 7d 或错配 reset。

#### Scenario: 相对 reset 与无效快照时间
- **WHEN** 相同快照被重复评分，或相对 reset 没有合法采样时间
- **THEN** 前者重置点固定；后者不以 now 推算新的重置点；已过期/缺失用量沿用既有中性处理。

#### Scenario: 选号与诊断评分一致
- **WHEN** 存在未来的 Codex 5h reset 或仅存在未来 SessionWindowEnd
- **THEN** reset 因子优先用前者、回退后者，实际选号与诊断快照使用同一口径。

### Requirement: S09 调度耗时与粘性指标可观察
选号返回的 decision SHALL 包含本次测量的 LatencyMs；总体粘性命中 SHALL 按每次选号中 previous/session 任一命中计一次（来源 #7072），不改变该 PR 之外的选号策略。

#### Scenario: 可测量耗时返回调用方
- **WHEN** 用可控延迟超过 1ms 的依赖完成选号
- **THEN** 调用方与指标记录读到该次耗时，不因 defer 写入返回副本而丢失；不要求亚毫秒路径必为非零整数。

#### Scenario: 同次两类粘性命中
- **WHEN** 一次选号同时标记 previous 与 session 命中
- **THEN** 两个分项各记一次，总体只记一次，命中率不超过 100%；无命中不增加分子。

### Requirement: S09 automation 心跳接受完整信封
系统 SHALL 接受原有仅 automation_id 的合法心跳，以及成对带 current_time_iso 与 instructions 的完整信封；时间 SHALL 符合 RFC3339Nano，instructions SHALL 非空白（来源 #6869）。

#### Scenario: 完整与旧版信封
- **WHEN** automation_id 合法且可选字段成对合法，或两个可选字段均省略
- **THEN** 校验通过，保持原有心跳处理路径。

#### Scenario: 畸形信封
- **WHEN** 可选字段只出现一个、时间非法、instructions 为空白、子元素重复或未知，或违反既有 XML 层级/属性约束
- **THEN** 校验不通过，不将其作为合法 automation 心跳放行。

### Requirement: S10 大体积正文读取降低分配
系统 SHALL 在本仓库已有正文读取路径按实际到达数据分块读取并组装结果，避免依据不可信 Content-Length 预分配整个大缓冲区（来源 #6783）。不引入缺失的 raw-input/ingress-compat 基座，服务层仅适配已有调用路径。

#### Scenario: 大图与不可信长度
- **WHEN** 读取固定大图载荷，分别提供真实、未知及远大于实际值的 Content-Length
- **THEN** 返回正文一致，不按虚报长度分配全量空间；同环境同载荷的改前/改后基准显示分配字节数下降。

#### Scenario: 读取边界与既有转换
- **WHEN** 请求为空、小正文、分块到达、gzip 编码或读取失败
- **THEN** 保持既有返回数据、解压、大小限制与错误传播；reasoning-replay、call-id 归一和 Responses Lite 回归通过。

### Requirement: S10 隐私客户端指纹与质询识别
共享隐私客户端 SHALL 使用 Firefox 指纹，403/503 隐私设置响应 SHALL 优先依据 `cf-mitigated: challenge` 识别 Cloudflare 质询并回退到既有正文特征（来源 #7010）。本要求不承诺外部 Cloudflare 必然放行。

#### Scenario: 仅响应头标识质询
- **WHEN** 403/503 响应正文无关键词，但 cf-mitigated 为带空白或不同大小写的 challenge
- **THEN** 返回既有 CF blocked 状态；账号/订阅检查失败不被当作成功写入。

#### Scenario: 正常返回与其它客户端
- **WHEN** 隐私客户端得到正常账号/订阅数据，或 OAuth 服务创建自身客户端
- **THEN** 前者沿用现有解析更新；OAuth 服务独立的 Chrome 指纹不被全局替换。

### Requirement: S11 注册确认密码在提交前校验
注册表单 SHALL 提供独立确认密码输入，要求非空且与密码一致，并使用既有本地化错误提示（来源 #6654）。不改变后端注册契约。

#### Scenario: 确认密码缺失或不一致
- **WHEN** 提交注册时确认密码为空或与密码不同
- **THEN** 阻止进入注册提交路径并提示具体错误。

#### Scenario: 确认一致与表单禁用
- **WHEN** 两次密码一致且其它校验通过，或注册动作被禁用
- **THEN** 前者沿用现有注册流程，不发送额外确认密码字段；后者确认输入与显隐按钮随表单禁用。

### Requirement: S11 订阅分配搜索排除已删除用户
订阅分配的目标搜索 SHALL 使用管理端用户列表的有效用户语义，不使用包含历史已删除用户的用量搜索（来源 #6916）。

#### Scenario: 同关键词命中活跃与已删除用户
- **WHEN** 在订阅分配弹窗搜索用户
- **THEN** 只提供可分配用户；用量历史筛选继续保持自身原有语义。

#### Scenario: 更改关键词
- **WHEN** 选中用户后修改关键词
- **THEN** 保留第一档已建立的旧选择失效规则，不能把订阅提交给失效选择。

### Requirement: S11 代理凭据区分省略与显式清空
代理更新 SHALL 在前后端区分未提供字段与空字符串：省略/null 保持旧值，显式空串清空；非空值按既有 trim 规则保存（来源 #7111）。现有账号/代理导入调用 SHALL 适配相同更新语义。

#### Scenario: 显式清空用户名或密码
- **WHEN** 用户将凭据清空并提交，密码字段确已被编辑
- **THEN** 请求发送空字符串，后端持久化清空，不因空值被忽略。

#### Scenario: 未编辑密码与独立字段更新
- **WHEN** 密码未编辑而省略，或只更新其中一个凭据字段
- **THEN** 未提供的字段保持旧值，既有导入更新不意外删除凭据。

### Requirement: S12a TTFT 详情使用首 token 延迟
运维请求详情 SHALL 暴露可空 first_token_ms，支持 ttft_desc 排序；TTFT 卡片详情 SHALL 默认查询成功请求并展示首 token 延迟，而非总耗时（来源 #6971）。

#### Scenario: TTFT 与总耗时排序不同
- **WHEN** 两条请求的首 token 延迟排序与总耗时排序相反
- **THEN** TTFT 入口使用 kind=success、sort=ttft_desc，按首 token 延迟降序；相同值按 created_at 降序，NULL 排最后，使用真实 SQLite 验证。

#### Scenario: 缺失值与其它详情入口
- **WHEN** 首 token 延迟为 0、缺失，或从普通总耗时详情进入
- **THEN** TTFT 分别显示 `0 ms` 与 `-`；普通详情继续展示 duration_ms；移动端与桌面语义一致。

### Requirement: S12b 平台限额只保存已配置记录
系统 SHALL 仅保存至少一档日/周/月限额非 NULL 的配置，数值 0 也算配置；无行等价于不限额。历史空配置 SHALL 由已批准并实现的本地 SQLite 迁移 227 清理，不继承上游 238，不改 Ent schema 或用量账单口径。

#### Scenario: 初始化与清空限额
- **WHEN** 默认配置三档全空，或管理员将既有平台三档全部清空
- **THEN** 初始化不建空行，清空使既有活跃行软删除；至少一档为 0 的记录保留；之后重新配置新建行并从新窗口起算。

#### Scenario: 初始化补齐与正常限额更新
- **WHEN** 已有记录再次应用默认值，或修改仍有配置的平台限额
- **THEN** 初始化只补 NULL 档位，不覆盖已配置值；修改活跃行限额保留该行用量与窗口。

#### Scenario: 用量写入不重建空行
- **WHEN** 增量累计或批量快照遇到不存在/已软删的配额行
- **THEN** 跳过该行，不插入或复活；既有活跃行正常更新，批量写入使用本地 SQLite 兼容实现。

#### Scenario: 仪表盘与重置操作
- **WHEN** 平台仅有用量、仅有限额配置或两者皆无
- **THEN** 卡片集合为有用量与有配置平台的并集，计数不含“其他”差额卡；无已保存配置的平台不可重置用量窗口。

#### Scenario: SQLite 227 清理历史空配置
- **WHEN** 在临时验证库或按既有迁移流程执行本地 227 清理迁移
- **THEN** 只删除三档 limit 都为 NULL 的历史行，保留任何非 NULL（含 0）行；重复执行无额外影响，不改 Ent schema 或用量账单口径。
