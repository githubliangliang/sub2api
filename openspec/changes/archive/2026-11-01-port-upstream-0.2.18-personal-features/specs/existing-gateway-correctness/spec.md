# 既有Gemini图片隔离与新增规则UI兼容

## MODIFIED Requirements

### Requirement: F07 Gemini 工具图片与 Antigravity 图片冷却隔离

系统 SHALL 满足以下已有功能正确性契约：base64 工具结果图片转为 inlineData，工具响应顺序保持；图片模型冷却只写/查对应图片键。

#### Scenario: 目标行为

- **WHEN** 使用Gemini 工具图片与 Antigravity 图片冷却隔离涉及的既有入口，并触发对应的有效请求或配置变更
- **THEN** base64 工具结果图片转为 inlineData，工具响应顺序保持；图片模型冷却只写/查对应图片键。

#### Scenario: 反例与范围边界

- **WHEN** 输入不满足修复触发条件，或发生错误、切换、重复提交等边界情况
- **THEN** 普通文本工具结果及 Gemini 文本家族冷却不变；不下载远程图片。
- **THEN** 新增临时规则作用范围UI时 SHALL 遵守model-scoped-cooldown的默认、认证和隔离边界，继续保留图片/文本键隔离。
