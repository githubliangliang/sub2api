## ADDED Requirements

### Requirement: S07 图片 INTERNAL500 不牵连整账号

现有Antigravity图片模型重试耗尽INTERNAL500 SHALL 只冷却最终映射模型，并同步缓存；文本模型仍用既有账号惩罚。

#### Scenario: 具体边界与可观测结果

- **WHEN** 图片别名映射后的Antigravity模型耗尽INTERNAL500重试
- **THEN** 仅最终图片模型冷却并同步调度缓存；同账号文本模型仍可用，整账号惩罚计数不增加

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** alias映射图片/文本、相同模型计时、其他模型仍可调度、SetModelRateLimit失败、SQLite Extra/调度缓存一致；不增加账号处罚计数。
