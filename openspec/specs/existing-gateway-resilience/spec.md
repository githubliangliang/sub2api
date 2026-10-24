# existing-gateway-resilience Specification

## Purpose

定义个人单节点网关在凭据路由、流式中断、用量结算、模型兼容、认证恢复和图片下载中的失败处理约束，确保既有客户端路径可靠运行并保留已观察用量。

## Requirements

### Requirement: S01 混合分组 OAuth 凭据协议隔离
系统 SHALL 满足以下既有功能行为：混合归属含只能走 OpenAI 网关的 OAuth 且模型兜底会选该族时拒绝歧义；Anthropic 请求构造在发送前阻止外平台 OAuth；count_tokens 给出可回退的404。

#### Scenario: 修复受影响路径
- **WHEN** 既有路径收到本簇涉及的请求、响应或配置组合
- **THEN** HTTP/WS歧义拒绝、明确路由、纯API Key歧义、冷却账号仍参与配置归属；对OpenAI/Grok OAuth构造请求时确认未发出HTTP，合法Anthropic/Vertex及独立Gemini路径保持工作。

#### Scenario: 保留个人版及相邻路径
- **WHEN** 请求不符合修复条件，或包含来源PR的其它新能力
- **THEN** API Key 多协议兼容保留；Antigravity/Gemini 独立转发在守卫前分流；配置归属不随短暂冷却变化；不记录token、不新增平台。

### Requirement: S02 Responses 时间戳与失败终态
系统 SHALL 满足以下既有功能行为：非透传 Responses JSON/SSE/WS桥缺失或null的created_at补齐；同一流共享首次确定时间；合成失败带created_at和sequence_number。

#### Scenario: 修复受影响路径
- **WHEN** 既有路径收到本簇涉及的请求、响应或配置组合
- **THEN** 缺失/null/已有时间、created与completed/failed一致、裸error替换序号、JSON和SSE转JSON、Grok桥/WS、透传未修改事件；严格JSON字段断言。

#### Scenario: 保留个人版及相邻路径
- **WHEN** 请求不符合修复条件，或包含来源PR的其它新能力
- **THEN** 已有值不覆盖；透传原上游帧保持字节语义，仅观察时间用于本地合成帧；保留上游usage、工具恢复与终态约束。

### Requirement: S03 Codex 已退役订阅模型调度一致性
系统 SHALL 满足以下既有功能行为：在实际映射目标上判定订阅账号退役模型；普通/粘性/compact调度和客户端目录一致过滤；不支持时返回model_not_found及迁移提示。

#### Scenario: 修复受影响路径
- **WHEN** 既有路径收到本簇涉及的请求、响应或配置组合
- **THEN** OAuth/SetupToken/shadow/API Key混合、显式别名双向映射、compact专用映射、冷却与previous_response_id、remote/file与空列表；用本地桩验证，不声称验证真实账号模型权限。

#### Scenario: 保留个人版及相邻路径
- **WHEN** 请求不符合修复条件，或包含来源PR的其它新能力
- **THEN** API Key上游仍可用；显式映射到可用替代模型允许，映射到退役模型拒绝；空目录有权威性；不全局替换用户模型、不删除历史价格、不引入pinned目录。

### Requirement: S04 Codex 容量错误恢复与出站身份
系统 SHALL 满足以下既有功能行为：TUI出站UA与originator成对归一为CLI并保留生效版本/系统指纹；空语义片段不阻止安全换号；容量错误尊重有效上游重试提示且最多8秒。

#### Scenario: 修复受影响路径
- **WHEN** 既有路径收到本簇涉及的请求、响应或配置组合
- **THEN** UA/originator/version配套，HTTP/WS与探测头一致；空/非空delta及done、秒/ms/HTTP-date提示、非法/负数/NaN/超大数防溢出、8秒上限和取消；保留已实施指数退避。

#### Scenario: 保留个人版及相邻路径
- **WHEN** 请求不符合修复条件，或包含来源PR的其它新能力
- **THEN** 已有真实输出不能重试；非容量错误不改变；合法非TUI身份和管理员版本策略保留；重试可取消，不引入Grok新配额/Realtime/截止时间机制。

### Requirement: S05 Anthropic 兼容流中断结算与终态
系统 SHALL 满足以下既有功能行为：客户端断开后停止下游写入、有界读完上游并收集usage；上游异常终止返回对应协议错误而非成功；有已观测用量的失败也记录一次，输出前可重试错误不重复结算。

#### Scenario: 修复受影响路径
- **WHEN** 既有路径收到本簇涉及的请求、响应或配置组合
- **THEN** 写失败后最终usage、客户端等待期间断开、缺stop/读错/超时/带内error、错误前后可见输出、流终态唯一、连接关闭与无goroutine泄漏；真实SQLite验证失败部分结算/删除Key/dedup。

#### Scenario: 保留个人版及相邻路径
- **WHEN** 请求不符合修复条件，或包含来源PR的其它新能力
- **THEN** 保留本地128c5f7bf固定窗口、删除Key结算、dedup与simple mode；ping/注释写出和真实语义输出分开；不能因排水无限占用连接或goroutine。

### Requirement: S06 既有 Antigravity 协议与别名修复
系统 SHALL 满足以下既有功能行为：仅修已有模型和路径：保留工具base64图片/文档、Gemini函数调用及签名；放行已解析到Antigravity的Gemini路由；已有Claude5.5档位别名用基准价格；明确的不同目标映射优先，自映射可选精确已有档位；账号测试按映射目标选协议并剔除通配符候选；兼容已有上游套餐细分ID。

#### Scenario: 修复受影响路径
- **WHEN** 既有路径收到本簇涉及的请求、响应或配置组合
- **THEN** 工具媒体顺序、文本兼容、functionCall/thoughtSignature回放、composite入口非Gemini仍拒绝、5.5别名动态/fallback价格、已有Gemini自映射/不同目标/通配符、映射测试协议、Pro/Ultra细分ID与未知档位。

#### Scenario: 保留个人版及相邻路径
- **WHEN** 请求不符合修复条件，或包含来源PR的其它新能力
- **THEN** 不增加Argon/Claude新档位目录，不引入元数据目录扩展、UA hub升级或生图测试新展示；工具媒体不远程下载；套餐ID只用于既有上游端点选择，不是支付功能。

### Requirement: S07 Gemini 错误策略单次裁决与 Vertex 认证
系统 SHALL 满足以下既有功能行为：同一上游响应只裁决一次错误策略；Vertex无quota/error明细的共享容量429只做请求内重试；确切SA不存在时标记认证失效并换号。

#### Scenario: 修复受影响路径
- **WHEN** 既有路径收到本簇涉及的请求、响应或配置组合
- **THEN** 三条既有Gemini入口各自验证首次/再次401；响应重试之间策略不串用；共享容量与QuotaFailure/RetryInfo/项目额度反例；SA不存在、瞬态token错误、取消与换号耗尽。

#### Scenario: 保留个人版及相邻路径
- **WHEN** 请求不符合修复条件，或包含来源PR的其它新能力
- **THEN** 真实项目配额/RetryInfo/QuotaFailure保留冷却；非account-not-found token错误不永久停号；只使用现有Vertex service_account，不引入Vertex API Key或缓存绑定。

### Requirement: S08 Gemini 客户端断开保留已观测用量
系统 SHALL 满足以下既有功能行为：确认由客户端断开且已观测到可计费用量时返回部分结果并结算；记录ClientDisconnect与是否收到finishReason；图片只按已观测结果处理。

#### Scenario: 修复受影响路径
- **WHEN** 既有路径收到本簇涉及的请求、响应或配置组合
- **THEN** 三协议的客户端取消/写失败、已完成后取消、无usage、图片已输出/未输出、截断残行和真正上游错误；usage队列在取消context下仍提交且SQLite去重。

#### Scenario: 保留个人版及相邻路径
- **WHEN** 请求不符合修复条件，或包含来源PR的其它新能力
- **THEN** 不估算未收到的token，不把真实上游错误或无用量中断伪造成功；残行读错误不算完整帧；不引入#8114新思考输出。

### Requirement: S09 现有图片转存下载边界
系统 SHALL 满足以下既有功能行为：默认下载客户端逐跳拒绝非公网目的地；直连在socket层校验实际IP；非图片响应失败，不转存为图片。

#### Scenario: 修复受影响路径
- **WHEN** 既有路径收到本簇涉及的请求、响应或配置组合
- **THEN** 直接/重定向/IPv4/IPv6/zone/private DNS/代理分支；JSON与HTML拒绝、合法图片成功、响应体关闭和大小限制；测试使用注入transport，不依赖外网。

#### Scenario: 保留个人版及相邻路径
- **WHEN** 请求不符合修复条件，或包含来源PR的其它新能力
- **THEN** 仅加固现有转存，不开启批量生图；正常公网图/b64/大小限制保留。代理模式只能验证本地DNS，无法证明远端代理解析结果，与来源限制一致。

### Requirement: S10 工具链与重定向依赖维护
系统 SHALL 满足以下既有功能行为：Go升级目标1.27.3并同步go.mod/CI/Docker版本断言；只升级follow-redirects到修复版本，双份pnpm overrides保持一致并提交锁文件。

#### Scenario: 修复受影响路径
- **WHEN** 既有路径收到本簇涉及的请求、响应或配置组合
- **THEN** 核验所有Go版本字串、后端build/unit/SQLite，前端typecheck/lint/build/关键vitest；重新跑audit及govulncheck，目标advisory消失且无无关依赖漂移。

#### Scenario: 保留个人版及相邻路径
- **WHEN** 请求不符合修复条件，或包含来源PR的其它新能力
- **THEN** 不复制上游PG Docker阶段、支付依赖或完整lockfile；不把audit变绿作为无条件目标，不延长旧例外；其它现存advisory单列风险。
