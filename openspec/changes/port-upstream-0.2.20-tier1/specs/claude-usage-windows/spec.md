## ADDED Requirements

### Requirement: F01 Claude setup-token 周用量

Setup-token 查询 SHALL 使用已有被动采样，成功响应缺少5h头及429早退也保留7d/7d_oi；不要求profile权限或主动调用usage API。

#### Scenario: 具体边界与可观测结果

- **WHEN** Setup-token没有5h采样但持有42%的7d有效样本，或429响应带来最新7d值
- **THEN** 单查、强制刷新和批量查询均返回该周用量；不发主动usage API请求，不丢弃429样本

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** 无5h、有效/过期5h、显式0、无样本、429以及单查/force/passive/batch一致；SQLite Extra与固定quota窗口不回退。
