## ADDED Requirements

### Requirement: F13 Bedrock EventStream 分配前边界

EventStream解码 SHALL 在分配和切片前验证headers长度、payload长度及总长关系；合法帧继续校验CRC。

#### Scenario: 具体边界与可观测结果

- **WHEN** CRC合法的EventStream prelude声称headers超出总帧或payload超出协议上限
- **THEN** 在分配和切片前返回错误且不panic；小型合法帧仍正常，CRC与截断错误保持

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** header越界、total<16、header>128KiB、payload>16MiB、truncated/CRC、合法frame；大长度仅喂prelude，证实分配前拒绝。
