## ADDED Requirements

### Requirement: E1 捕获客户端声明的推理强度
网关 SHALL 在任何分组或模型改写前捕获客户端声明的推理强度，并区分显式声明、原模型后缀和未声明。

#### Scenario: 显式值优先于原模型后缀
- **WHEN** Responses、Chat 或 Messages 请求中协议对应的显式值为 `high`，客户端原模型以后缀 `-low` 结尾
- **THEN** 请求强度记录为 `high`；若省略显式值则记录已知后缀 `low`。

#### Scenario: 模型归一化与本地特殊值
- **WHEN** 客户端请求本地支持的 `max` 或 `none`，并经模型或分组的既有转换被接受
- **THEN** 请求值保留该声明，实际值取成功转发的结果；现有拒绝策略继续生效。

#### Scenario: 缺省或未知
- **WHEN** 客户端省略强度且原模型没有已知后缀，或提供未知的非空显式值
- **THEN** 请求值为空；兼容桥默认值和映射目标后缀均不能填充它，原有校验或转发结果不受影响。

### Requirement: E2 逐个请求或 WS 轮次记录最终结果
网关 SHALL 将同一请求或 WS 轮次的初始请求值和成功转发的实际值传到对应的用量记录。

#### Scenario: 分组降级、重试与流式结算
- **WHEN** 客户端请求 `xhigh`，组上限使成功发送的请求为 `high`，且经历换号或 SSE 结束后的异步写入
- **THEN** 该条记录对应 `xhigh → high`；若策略在转发前拒绝请求，则沿用既有拒绝和用量处理。

#### Scenario: 一个连接上的三轮请求
- **WHEN** 同一个 WS 连接依次收到显式 `xhigh`、显式 `low` 和未提供强度的三轮 create，各轮完成后延迟写入
- **THEN** 三条请求值依次为 `xhigh`、`low`、NULL；各自配对本轮的实际值，连接复用或写入顺序不会串轮。

### Requirement: E3 SQLite 兼容旧记录及既有结算
系统 SHALL 使用可空的请求强度列保存新增信息，保持既有实际强度、用量金额和去重语义。

#### Scenario: 从 SQLite 229 升级
- **WHEN** 包含历史 `reasoning_effort=high` 的 SQLite 229 数据库运行新迁移
- **THEN** 升级后该行的请求值为 NULL，已有强度、用量和金额可原样读取；新行可以保存两个不同值。

#### Scenario: 写入方式与重复请求
- **WHEN** 使用既有单条、批量或 best-effort 写入并通过列表和详情读取，或对同一成功请求重复结算
- **THEN** 两列正确对应原请求，其它列不移位；相同去重身份仍只结算一次，余额、账号成本及 quota 与既有行为一致。

### Requirement: E4 角色决定呈现的强度信息
用量 API、界面及导出 SHALL 根据现有角色权限呈现强度，并兼容仅有单值的记录。

#### Scenario: 两个值有差异
- **WHEN** 用户及管理员分别读取或导出一条 `xhigh → high` 的记录
- **THEN** 用户只获得 `reasoning_effort=xhigh`；管理员还能读取 `upstream_reasoning_effort=high`，页面及导出能区分请求与转发。

#### Scenario: 两个值相同或记录只有一个值
- **WHEN** 请求与实际均为 `high`、历史行仅有实际值 `high`，或两个值均为空
- **THEN** 前两种在列表显示一个 `high`，最后一种显示无值；API 省略管理员差异字段，用户导出没有上游强度列。
