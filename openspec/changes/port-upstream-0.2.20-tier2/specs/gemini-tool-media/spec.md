## ADDED Requirements

### Requirement: S08 Gemini 工具媒体原生承载

Gemini3及以后已使用模型的工具媒体 SHALL 保留在functionResponse.parts；旧模型按已有分离图片回退，关联工具名/顺序保持。

#### Scenario: 具体边界与可观测结果

- **WHEN** tool_result同时含文字与媒体，最终映射模型分别为Gemini3及Gemini2.5
- **THEN** Gemini3保留functionResponse.parts的工具关联，旧模型使用既有分离媒体回退；工具名/顺序及用量保持

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** Gemini2.5/3/4和别名、base64/URI/text混合、未知工具ID、重试去签名、原输入不变、tool_order及图片用量/dedup不重复。
