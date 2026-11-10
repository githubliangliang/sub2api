## ADDED Requirements

### Requirement: F11 GPT Image 账号测试参数

既有GPT Image API Key账号测试 SHALL 不发送被上游拒绝的response_format，并继续解析默认base64响应。

#### Scenario: 具体边界与可观测结果

- **WHEN** 通过已有GPT Image API Key测试入口发送管理员prompt
- **THEN** 请求不含response_format，仍包含原模型/prompt/n并正确解析默认base64结果；不引入新图片模型

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** 严格模拟上游检查无response_format、模型/prompt/n保留、有效图片解析、上游错误；不新增模型/尺寸路由。
