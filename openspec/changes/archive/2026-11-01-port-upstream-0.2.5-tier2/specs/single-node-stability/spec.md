## ADDED Requirements

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
#7043 的默认系数5.0方案 SHALL 在后段独立评估，记录ctx_pool会话数、账号并发、硬上限和资源余量；
未采用时 SHALL 保留当前默认1.0并明确标注本项未实施，不阻塞#7064容量公式的独立验收。

#### Scenario: 有依据采用上游默认值
- **WHEN** 容量评估支持采用默认5.0，且实施该默认值调整
- **THEN** OAuth/API Key默认均为5.0，示例配置一致；并发3、硬上限20时容量为15，显式1.5仍为5。

#### Scenario: 暂不采用新默认
- **WHEN** 缺少容量依据或当前部署不需要扩大连接池上限
- **THEN** 保留默认1.0，记录#7043待实施；不将#7064修复标记为失败，也不宣称#7043已落地。

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

### Requirement: S12b 平台限额只保存已配置记录且需迁移批准
仅在用户明确批准迁移后，系统 SHALL 实施 #6954 的限额清理：至少一档日/周/月限额非 NULL 才算配置，数值 0 也算配置；无行等价于不限额。本地迁移预留 227，不继承上游 238。批准前本子项 SHALL 保持待实施，不得作为已通过验收。

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

#### Scenario: 获批后的清理迁移
- **WHEN** 用户批准后执行本地 227 清理迁移
- **THEN** 只删除三档 limit 都为 NULL 的历史行，保留任何非 NULL（含 0）行；重复执行无额外影响，不改 Ent schema 或用量账单口径。
