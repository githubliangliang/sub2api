## ADDED Requirements

### Requirement: S11 RPM空覆盖缓存与既有JSON预检

鉴权快照 SHALL 区分已加载的nil RPM覆盖与加载失败，并按group隔离；既有thinking/委派预检仅在可证明无修改时跳过解析，不能改变输出语义。

#### Scenario: 具体边界与可观测结果

- **WHEN** 鉴权快照已成功加载某group且RPM override为nil，之后相同Key重复请求或切换group
- **THEN** 同group命中不重复查覆盖值；查询失败或group不同仍查询，0值不当缺失；JSON预检输出与旧路径逐字节等价并经实测决定是否落地

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** nil/0/正值/错误/跨group、失效与并发、miniredis+SQLite查询次数；escaped/重复键/非法JSON/长输入byte等价，benchmark ns/op+B/op+allocs/op前后实测，无收益不落优化hunk。
