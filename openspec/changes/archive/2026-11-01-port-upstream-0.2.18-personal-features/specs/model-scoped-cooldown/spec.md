# 临时规则作用范围

## ADDED Requirements

### Requirement: 临时不可调度规则明确支持账号级和模型级范围

系统 SHALL 在已有临时不可调度规则中支持 `scope=account|model`，创建/编辑可选择并回显。新规则默认model；缺失、空或非法scope按旧语义解析：非401且模型已知为model，其余保留账号级/认证处理。规则启用开关、错误码/关键词/时长和命中顺序保持原有行为。

#### Scenario: 显式账号级与模型级

- **WHEN** 相同非401错误分别命中scope为account和model的有效规则，且最终上游模型已知
- **THEN** 前者 SHALL 暂停该账号整体，后者只冷却该账号的目标模型
- **THEN** 状态诊断 SHALL 标明实际scope，模型级还标明目标模型

#### Scenario: 编辑旧规则保持行为

- **WHEN** 管理员打开并原样保存无scope的429规则和401规则
- **THEN** SHALL 分别回显等价的model/account设置，保留原有效规则顺序、启用和匹配条件
- **THEN** 旧非法scope值 SHALL 同样按缺省处理

#### Scenario: 多条规则同时匹配

- **WHEN** 两条不同scope规则同时匹配同一错误
- **THEN** 第一条匹配规则 SHALL 决定有效范围，不因后续规则或写失败改用另一范围

### Requirement: 模型隔离遵守真实上游目标且不扩大为账号阻塞

模型级规则 SHALL 使用既有实际上游模型/调度键归一逻辑。冷却期间只排除对应账号模型组合，其他账号/模型保持可用；TTL到期恢复。无可识别模型时 SHALL 回落到原账号级或认证策略，不写空模型键。

#### Scenario: 别名和独立模型

- **WHEN** 客户端别名A映射到上游模型M且触发model规则
- **THEN** 同账号映射到M的请求 SHALL 在冷却期排除，模型N与其他账号不受该规则阻塞
- **THEN** 普通、粘性及已有WS选择路径 SHALL 使用一致有效键

#### Scenario: 保留图片隔离

- **WHEN** Antigravity图片模型触发模型级规则
- **THEN** SHALL 沿用已有图片键，不把文本Gemini家族一并冷却

#### Scenario: 未知模型与到期

- **WHEN** 命中规则时无法确定目标模型，或者此前模型冷却已到期
- **THEN** 前者保持账号级/认证回退；后者按既有调度条件重新可用

#### Scenario: 模型状态写入失败

- **WHEN** 模型级规则命中但持久化失败
- **THEN** 当前请求 SHALL 按既有预算换号并记录失败
- **THEN** MUST NOT 写账号全局暂停或全局运行时熔断，也不伪称冷却已持久化

### Requirement: 模型401规则不得覆盖凭据失效处理

401仅在显式model、已知模型且不属于凭据失效时 SHALL 允许模型隔离。OAuth、SetupToken/影子账号及明确token_invalidated/token_revoked/Unauthorized等认证失效 MUST 保留原刷新、停号、换号及升级策略。规则本身不能降低真实账号级鉴权失败的处理范围。

#### Scenario: API Key 模型权限错误

- **WHEN** API Key收到非凭据级模型权限401，首个规则明确为model且模型已知
- **THEN** SHALL 仅冷却该模型，不进入账号重复401升级逻辑

#### Scenario: 凭据无效但配置model

- **WHEN** OAuth/SetupToken/影子凭据失效，或OpenAI响应明确token被撤销/失效/Unauthorized
- **THEN** SHALL 继续既有认证错误处理，不用model规则掩盖问题

### Requirement: 隔离沿用SQLite和单节点缓存语义

作用范围 SHALL 复用现有JSON和模型限流存储，在SQLite与可选miniredis下生效。已有池模式/错误码策略、默认限流、Grok团队模型键及真实配额窗口 MUST 保留；只在原本允许自定义规则的路径应用新增scope，不新增平台或后台服务。

#### Scenario: SQLite持久化与miniredis

- **WHEN** 使用真实SQLite账号、redis.enabled=false并分别触发account/model规则
- **THEN** SHALL 从持久状态与既有调度快照观察相同隔离结果，无PG语法或新迁移
- **THEN** 用量去重与固定配额窗口重置保持原有结果

#### Scenario: 未命中或既有策略跳过

- **WHEN** 规则未启用、错误码/关键词不匹配，或当前链路的既有策略明确跳过自定义临时规则
- **THEN** SHALL 保持原错误/重试策略，不仅因存在scope字段新增处罚
