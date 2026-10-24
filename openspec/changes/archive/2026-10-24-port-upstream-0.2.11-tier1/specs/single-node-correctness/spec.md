## ADDED Requirements

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
