## ADDED Requirements

### Requirement: S03 已使用 Terra 的目录与计费后续

已接入Terra SHALL 使用冻结tag中的Codex descriptor作为受限fallback，Codex默认low及500000窗口与max_context_window=872000区分，API兼容默认medium/none仍合法，minimal明确400。OAuth/setup-token长上下文豁免按最终发送模型决定，API Key保留272k严格边界与Fast结算。

#### Scenario: 具体边界与可观测结果

- **WHEN** 实际发送模型为Terra，分别使用OAuth、setup-token、API Key且输入为272000或272001 tokens
- **THEN** 前两种凭据不收模型长上下文附加费；API Key只在大于272000时应用既有2x输入/1.5x输出规则，并保留Fast/Free Fast与显式零价格语义

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** none/minimal/max/ultra的API与Codex不同标尺；别名/映射后校验，账号raw/metadata/空目录优先、ETag；271999/272000/272001×OAuth/setup-token/API Key×default/priority/ultrafast/Free Fast/显式零价与SQLite去重。
