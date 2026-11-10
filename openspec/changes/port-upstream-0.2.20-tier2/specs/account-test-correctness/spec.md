## ADDED Requirements

### Requirement: S10 Antigravity 测试提示词贯通

已有账号测试 SHALL 将管理员输入提示词传至Antigravity Claude/Gemini请求，并返回真实回复；空值沿用默认提示词。

#### Scenario: 具体边界与可观测结果

- **WHEN** 管理员在既有Antigravity测试表单输入多行提示词并选择Claude或Gemini模型
- **THEN** 实际上游请求包含输入提示词和足够测试输出预算，返回真实内容；空提示词、旧定时测试调用和取消行为有明确兼容

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** 自定义/空/Unicode/多行prompt、Claude/Gemini模型、响应文本/错误/取消、定时测试旧调用兼容；真实本地HTTP捕获body，外部账号测试不自动执行。
