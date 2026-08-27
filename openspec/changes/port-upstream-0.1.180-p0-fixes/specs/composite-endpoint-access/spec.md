## ADDED Requirements

### Requirement: Composite 分组的视频端点访问必须前后一致
Composite 分组的视频状态查询与内容获取已放行，系统 SHALL 让视频生成**任务创建**与之对齐：
当分组平台为 Composite 时同样进入对应的视频生成处理。MUST NOT 出现「能查状态、不能提交任务」
的不一致状态。

#### Scenario: Composite 分组提交视频生成任务
- **WHEN** 使用 Composite 分组的密钥请求视频生成
- **AND** 该分组可路由到支持视频的账号
- **THEN** 请求 MUST 进入视频生成处理
- **THEN** MUST NOT 因分组平台不是具体平台而被拒绝

#### Scenario: 直接使用具体平台分组
- **WHEN** 分组平台本身就是支持视频的具体平台
- **THEN** 行为 MUST 与移植前一致

#### Scenario: 不支持视频的分组平台
- **WHEN** 分组平台既非 Composite 也非支持视频的具体平台
- **THEN** MUST 保持既有的拒绝行为

#### Scenario: 状态查询与内容获取
- **WHEN** Composite 分组查询视频任务状态或获取内容
- **THEN** 行为 MUST 与移植前一致（本来就已放行）
