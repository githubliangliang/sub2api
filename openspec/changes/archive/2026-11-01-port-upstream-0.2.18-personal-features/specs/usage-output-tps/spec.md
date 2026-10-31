# 使用日志输出 TPS

## ADDED Requirements

### Requirement: 使用记录展示单请求输出速率

管理员和用户使用记录 SHALL 在既有延迟区域展示“输出 TPS”，按 `output_tokens × 1000 / duration_ms` 计算，保留一位小数并标注 `tok/s`。说明 MUST 明确包含首字等待，输出 Token 已包含的推理用量不得重复相加。

#### Scenario: 使用完整请求耗时

- **WHEN** 一条文本记录输出100 Token、总耗时2000ms、首字等待1500ms
- **THEN** 两种使用记录页面均显示 `50.0 tok/s`
- **THEN** MUST NOT 扣除1500ms首字等待或使用输入/缓存 Token 计算

#### Scenario: 极小但有效的速率

- **WHEN** 合法正数计算结果小于0.05 tok/s
- **THEN** SHALL 按一位小数显示 `0.0 tok/s`，与不可用标记区分

### Requirement: 不适用或无效记录明确显示不可用

系统 SHALL 对缺失、非有限或非正的输出数/耗时显示 `—`。图片输出记录、图片计费记录以及显式的 Live/unknown/其他不支持类型 MUST 显示 `—`；支持 sync、stream、ws_v2、cyber，历史缺失类型时按其余字段判断。

#### Scenario: 缺少可信数值

- **WHEN** 输出数或耗时为 null、缺失、NaN、Infinity、0或负数
- **THEN** SHALL 显示 `—`，不出现 NaN、Infinity、异常或除零结果

#### Scenario: 图片与 Live

- **WHEN** image_count或image_output_tokens大于0、billing_mode为image，或request_type为live/unknown
- **THEN** SHALL 显示 `—`

#### Scenario: 图片输入后的文本输出和历史记录

- **WHEN** 记录只有图片输入、没有图片输出/图片计费，且文本输出/耗时有效；或者记录缺少request_type但满足其他有效条件
- **THEN** SHALL 计算输出 TPS

### Requirement: 展示沿用已有权限、列偏好和使用量数据

新增展示 SHALL 随管理员、用户各自的延迟列显隐，保留权限、分页和数据刷新行为。系统 MUST 从已有记录派生速率，不回写历史使用量、不增加数据库或运维overview的TPS字段。

#### Scenario: 默认显示与用户隐藏

- **WHEN** 用户没有延迟列偏好，或已显式隐藏延迟列
- **THEN** 前者默认可见输出 TPS，后者继续隐藏该区域
- **THEN** 中英文均有标签和计算说明

#### Scenario: 分页或刷新

- **WHEN** 表格切到另一页或记录刷新
- **THEN** 每行 SHALL 使用该行最新的输出数和耗时，不复用上一行/上一页的速率
