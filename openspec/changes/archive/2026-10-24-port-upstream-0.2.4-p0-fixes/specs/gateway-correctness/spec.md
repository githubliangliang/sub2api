## ADDED Requirements

### Requirement: F01 仅依据真实耗尽信号或开启的 fallback 创建冷却
系统 SHALL 将未满窗口的 reset-after 视为窗口信息；fallback 关闭时 MUST NOT 因普通 429 建立默认冷却。

#### Scenario: 未耗尽且关闭 fallback
- **WHEN** 429 的 7d/5h 使用率都小于 100%，无明确正文配额 reset，且 fallback 关闭
- **THEN** 不按最长窗口停调，也不建立默认进程内冷却。

#### Scenario: 真实耗尽与 shadow
- **WHEN** 响应明确报告耗尽，或账号为 Spark shadow
- **THEN** 分别保留真实配额冷却与现有 shadow 排除规则。

### Requirement: F02 选号使用有效渠道映射模型
Responses、WS、Chat Completions、Embeddings SHALL 使用映射后的有效模型评估账号能力。

#### Scenario: 公共别名映射
- **WHEN** 客户端请求别名而账号仅支持映射目标
- **THEN** 账号可以被正确选中，响应与计费仍遵守各自已有模型命名规则。

#### Scenario: 空映射
- **WHEN** 映射未生效或目标为空白
- **THEN** 使用原请求模型。

### Requirement: F03 所有负载感知层执行逐账号渠道限制
系统 SHALL 在 routed、sticky 和普通候选层执行渠道限制，不允许缓存粘性绕过。

#### Scenario: 某账号映射到受限模型
- **WHEN** 该账号的真实上游模型被渠道限制
- **THEN** 三层均跳过它；全受限时返回适当的不可用错误。

### Requirement: F04 unavailable continuation 可以恢复
系统 SHALL 在已有恢复状态码边界内识别明确的 unavailable previous_response_id 错误。

#### Scenario: 用户侧 continuation 不可用
- **WHEN** 上游返回 previous_response_id is not available for this user
- **THEN** 进入既有历史恢复流程；无关错误不得触发恢复。

### Requirement: F05 透传保留显式 none
系统 SHALL 保留透传账号显式传入的 reasoning.effort=none。

#### Scenario: 透传与普通账号
- **WHEN** 两类账号分别传入 none
- **THEN** 透传保留，普通账号继续遵守其既有归一化规则。

### Requirement: F06 缓存提升生成有效 JSON
系统 SHALL 使用 JSON 字符串转义，保持原消息文本解码后的内容相同。

#### Scenario: 文本含控制字符
- **WHEN** 缓存提升处理控制字符、引号、反斜杠或普通文本
- **THEN** 转发正文仍为有效 JSON，解码后文字不变。

### Requirement: F07 billing 身份与最终出站 UA 一致
系统 SHALL 同步 billing 的 CLI 版本及依赖版本的 fingerprint 后缀。

#### Scenario: mimicry 覆盖缓存 UA
- **WHEN** 最终 UA 不同于缓存 UA，包括指纹统一关闭的情况
- **THEN** messages/count_tokens 的 billing 均与最终 UA 一致；无关 system 文本保持原样。

### Requirement: F08 合法 Claude 探测不限定 Haiku
系统 SHALL 在 UA 通过验证后识别任何模型的 max_tokens=1 探测，并在复用 ParsedRequest 时保留该值。

#### Scenario: 探测与普通请求
- **WHEN** 合法客户端探测非 Haiku 模型，或非法 UA/普通请求进入
- **THEN** 前者可用，后者仍接受原有严格校验。

### Requirement: F09 thinking binding 与 beta 对称
系统 SHALL 依据最终出站 beta header 处理 thinking.block_binding。

#### Scenario: 有或无 beta
- **WHEN** 最终 header 包含所需 token 或不包含
- **THEN** 分别保留或删除受保护字段；mimicry 使用相应 beta。

### Requirement: F10 credits_required 不扩大为整个账号限流
系统 SHALL 将已确认的 Fable credits_required 限定在 Fable 模型范围。

#### Scenario: 配额持久化失败
- **WHEN** 写入模型冷却失败
- **THEN** 不把该失败扩大为 Sonnet/Opus 等模型的账号级封锁；共享配额真实耗尽仍按既有规则处理。

### Requirement: F11 reasoning 无工具仍有 toolConfig
Antigravity Claude-to-Gemini 转换 SHALL 为 reasoning 无工具请求构造 toolConfig。

#### Scenario: 无工具与混合工具
- **WHEN** 请求无工具或同时有内置与函数工具
- **THEN** 前者含正常 toolConfig，后者保持已有 includeServerSideToolInvocations 行为。

### Requirement: F12 零指标与空指标可区分
系统 SHALL 将有效零值指标持久化为 0，仅缺失指针持久化为 NULL。

#### Scenario: 零延迟或零连接
- **WHEN** 指标指针值为 0 或 nil
- **THEN** SQLite 写入分别为 0 或 NULL；其它业务字段零值语义不变。

### Requirement: F13 API Key 不被注入 Codex 默认指令
系统 MUST NOT 为普通 API Key Responses 请求凭空合成 Codex instructions。

#### Scenario: API Key 与 OAuth
- **WHEN** API Key 请求省略 instructions，OAuth 缺默认指令，或客户端已显式提供指令
- **THEN** API Key 保持省略，OAuth 保持必要默认行为，显式指令不被覆盖；passthrough 和 metadata 清理继续可用。
