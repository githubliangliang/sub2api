## ADDED Requirements

### Requirement: `pricing.override_file` 是价格数据的最高优先级来源
新增配置项 `pricing.override_file` SHALL 默认关闭（空值）。启用时，补丁文件按 JSON 字段**浅合并**
覆盖目录与回退数据，作为最高优先级来源。合并 SHALL 挂在目录解析入口，使目录下载、回退合并与灾备
三条路径的覆盖语义**完全一致**。

依据：个人部署需要在不自建价格镜像、也不修改出厂快照的前提下改个别官方条目
（例：显式 `long_context_input_token_threshold: 0` 关掉某模型的阶梯）。改出厂快照会与上游永久分叉。

#### Scenario: 覆盖已存在条目的单个字段
- **WHEN** 补丁给一个目录里已有的模型指定了某个价格字段
- **THEN** 该字段 MUST 被覆盖
- **THEN** 该条目的其余字段 MUST 保持目录值（浅合并，MUST NOT 整条替换）

#### Scenario: 字段值为 null
- **WHEN** 补丁把某字段的值写成 `null`
- **THEN** 该字段 MUST 从条目中删除

#### Scenario: 目录与回退都没有的模型
- **WHEN** 补丁描述了一个目录与回退里都不存在的模型
- **THEN** 该模型 MUST 在回退合并**之后**作为独立条目并入
- **THEN** 该条目 MUST 自带价格字段并通过有效性过滤，否则 MUST 被丢弃

#### Scenario: 纯补丁不得抢先建条目
- **WHEN** 某模型在回退数据里有完整条目，而补丁只给了它一个字段
- **THEN** MUST NOT 在主解析阶段先用补丁建一个残缺条目
- **THEN** 最终条目的其余分项价 MUST 来自回退数据，MUST NOT 静默变 0

#### Scenario: 三条路径语义一致
- **WHEN** 价格数据分别经目录下载、回退合并、灾备路径进入
- **THEN** 覆盖结果 MUST 一致

### Requirement: 覆盖文件必须失败安全并对无效条目告警
补丁文件缺失或损坏 SHALL 只跳过合并，MUST NOT 影响目录加载。最终未生效的补丁条目
（模型名拼错、落在不存在的模型上且不满足独立条目条件）SHALL 打 WARN 哨兵。

#### Scenario: 文件不存在
- **WHEN** `override_file` 指向的路径不存在
- **THEN** 目录 MUST 正常加载
- **THEN** MUST NOT 因此报错退出

#### Scenario: 文件不是合法 JSON
- **WHEN** 补丁文件内容不是合法 JSON
- **THEN** MUST 跳过合并并继续加载目录

#### Scenario: 模型名拼错
- **WHEN** 补丁里的模型名在目录与回退中都不存在且不满足独立条目条件
- **THEN** MUST 打 WARN 指明该条目未生效

#### Scenario: 生效时机
- **WHEN** 补丁文件在运行中被修改
- **THEN** 变更 MUST 在重启或下次目录下载时生效（与 `fallback_file` 一致），本变更 MUST NOT 引入热加载

### Requirement: `fallback_file` 的语义不得改变
`pricing.fallback_file` SHALL 继续只补缺失模型。本变更 MUST NOT 让它获得覆盖已存在条目的能力。

#### Scenario: 两个文件同时配置
- **WHEN** `fallback_file` 与 `override_file` 都配了同一个模型
- **THEN** `override_file` MUST 胜出
- **THEN** `fallback_file` 对目录里已存在的条目 MUST 仍然无覆盖效果
