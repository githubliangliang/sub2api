## ADDED Requirements

### Requirement: S01 Codex Responses非流式与instructions回显

网关 SHALL 保留客户端原stream选择，将非流式SSE的completed/incomplete/cancelled转换成单JSON；缺Content-Type的OAuth SSE可识别。允许空instructions进入既有规范化，并将网关改写的instructions回显还原为客户端值。

#### Scenario: 具体边界与可观测结果

- **WHEN** 客户端要求stream=false而OAuth上游返回缺Content-Type的incomplete SSE，网关曾补写instructions
- **THEN** 客户端收到单JSON和正确Content-Type，保留incomplete状态/用量，instructions按原值还原；真正失败终态不伪造成成功

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** HTTP/透传/普通/compact，stream省略/false/true；incomplete/cancelled/canceled、失败终态仍失败；instructions缺失/null/空/多行与failover隔离；回显不泄漏注入内容，usage只落一次。
