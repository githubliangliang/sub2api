## ADDED Requirements

### Requirement: F02 Codex 强制模式停止多余探测

系统 SHALL 在force_responses或force_chat_completions模式跳过能力探测；auto及无效模式按现有规则探测。

#### Scenario: 具体边界与可观测结果

- **WHEN** 账号模式分别为force_responses、force_chat_completions和auto
- **THEN** 前两种模式上游探测及结果写入次数均为0；auto仍按既有探测流程执行

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** 两个强制模式请求数/写Extra数均0；auto/非法/非字符串模式仍探测，凭据覆写和取消正常。

### Requirement: F03 WebSocket 每轮首 token 时间

系统 SHALL 为每个WS轮次独立记录首token耗时，无response.id的delta归属当前活跃轮次。

#### Scenario: 具体边界与可观测结果

- **WHEN** 同一WS连接两轮均在各自开始后700ms收到不带response.id的首个delta
- **THEN** 两轮均有各自的首token时间，第二轮不因连接已记录第一轮而变成nil；请求档位和计费隔离保持

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** 至少两轮无ID delta、带ID delta、负时差、重复terminal与取消；保留Fast每轮service_tier/模型冲突和usage隔离。

### Requirement: F04 Codex WS→HTTP 桥接指纹

系统 SHALL 在Codex WS HTTP桥接构建上游请求前设置账号级指纹，重试和切换账号时清除旧暂存值。

#### Scenario: 具体边界与可观测结果

- **WHEN** Codex请求通过WS→HTTP桥接转发，且下一次重试切换到另一个账号
- **THEN** 最终HTTP header及client_metadata使用当次账号指纹，旧账号暂存值被清除，非Codex路由保持原行为

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** 最终HTTP请求header/body同账号稳定、跨账号隔离；无prompt_cache_key/显式客户端key/API Key/Grok负对照；Fast策略不变。
