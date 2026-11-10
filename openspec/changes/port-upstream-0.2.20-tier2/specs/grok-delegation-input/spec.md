## ADDED Requirements

### Requirement: S02 Grok 委派消息最小适配

既有Grok Responses入口 SHALL 将agent_message的input_text及明文encrypted_content按原顺序转为user消息，空消息跳过，不尝试解密。

#### Scenario: 具体边界与可观测结果

- **WHEN** Grok输入含agent_message，其input_text是任务封套，encrypted_content字段是明文任务正文
- **THEN** 只将该item转为user消息，原顺序拼接两段；空委派跳过，其它item、工具声明、大整数和现有映射不被泛化改写

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** 封套+正文顺序、空/非文本片段、字符串content、非委派item字节/大整数保真；additional_tools原提升规则、Chat对照、HTTP/Grok usage和模型审计保持。
