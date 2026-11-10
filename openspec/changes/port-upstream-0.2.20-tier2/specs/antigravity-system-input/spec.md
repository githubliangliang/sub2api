## ADDED Requirements

### Requirement: S06 Antigravity 系统身份保留任务

Antigravity路径 SHALL 仅规范化开头的已知客户端身份和attribution，保留附随任务及中间引用；原生Anthropic不受影响。

#### Scenario: 具体边界与可观测结果

- **WHEN** Antigravity的system/developer块以Codex或Claude身份开头且逗号后跟任务内容
- **THEN** 只规范化身份，任务和普通引用保留；已有Antigravity身份不阻止其它块清理，原生Anthropic路径不变

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** Claude/Codex多种前导、紧邻逗号/任务、引用/相似字符串、双身份、string/parts、native Anthropic对照；失败与成功请求body归属不混。
