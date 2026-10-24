## ADDED Requirements

### Requirement: Codex 模型目录不得使用持久禁用的账号
构建 Codex 模型目录（routed catalog / manifest）时，网关 SHALL 跳过处于持久禁用状态的账号。
目录请求 MUST NOT 因为挑到这类账号而失败或返回其陈旧快照。

#### Scenario: 组内存在持久禁用账号
- **WHEN** 分组里既有可用账号也有持久禁用账号
- **THEN** 目录 MUST 由可用账号构建

#### Scenario: 组内全部账号持久禁用
- **WHEN** 分组里所有候选账号都持久禁用
- **THEN** 目录构建 MUST 报出「无可用账号」而非静默使用被禁用账号

#### Scenario: 临时不可调度与持久禁用的区分
- **WHEN** 账号只是临时不可调度（冷却 / 限流窗口）而非持久禁用
- **THEN** 本规则 MUST NOT 把它排除（判据是持久状态，不是瞬时状态）

### Requirement: fast 模型必须在目录中透出 priority service tier
Codex 模型目录 SHALL 为 fast 类模型透出 priority service tier 描述。该改动只影响**目录里怎么描述
模型**，MUST NOT 引入任何发送侧的 `service_tier` 行为变化。

依据：本仓库尚未移植 Fast mode `service_tier` 整套发送逻辑（0.1.180 §7.2 未合）。这条与那簇无依赖关系。

#### Scenario: fast 模型条目
- **WHEN** 目录中包含一个 fast 类模型
- **THEN** 该条目 MUST 带 priority service tier 描述

#### Scenario: 非 fast 模型条目
- **WHEN** 目录中的模型不是 fast 类
- **THEN** 该条目 MUST 保持原有描述不变

#### Scenario: 不触发发送侧行为
- **WHEN** 客户端按目录描述发起请求
- **THEN** 网关发往上游的 `service_tier` 处理 MUST 与本变更前完全一致
