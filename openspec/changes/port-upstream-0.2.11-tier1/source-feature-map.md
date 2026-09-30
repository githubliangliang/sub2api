# 来源与行为映射

| 簇 | PR / merge SHA | 行为 | 范围边界 |
|---|---|---|---|
| F01 多工具改名保持请求体完整 | [#7701](https://github.com/Wei-Shaw/sub2api/pull/7701) `1dca86f287ad6b20ece6e329856d1d9914bf6434` | 工具声明、强制工具选择和历史tool_use中的同名工具保持一致，改名不破坏JSON或无关内容。 | 只改选中的工具名；内置工具、参数、缓存断点和未映射名称保留。排除图片keepalive测试的无关sleep。 |
| F02 白名单编辑保留显式模型映射 | [#7542](https://github.com/Wei-Shaw/sub2api/pull/7542) `f50b99b66cccfe0d91cfd553b27f78eda78355e9` | 在新增白名单项会覆盖from到不同to的映射时提示并拒绝该新增，保留已有映射。 | 覆盖创建、编辑、批量编辑的现有账号表单；不引入分组model_allowlist schema。 |
| F03 兼容入口遵守Claude Code降级分组 | [#7579](https://github.com/Wei-Shaw/sub2api/pull/7579) `a0f41f95a07ee6ca0b1300d3c96ce4b62e24724b` | Claude Code限制分组配置fallback时，Chat Completions和Responses请求沿现有调度选择降级分组账号。 | 无fallback仍拒绝；循环、缺失或不可用目标不得绕过现有检查。 |
| F04 密钥使用指引匹配客户端限制 | [#7678](https://github.com/Wei-Shaw/sub2api/pull/7678) `2f3fed2fdb0787141294cec81487a5df30426f7f` | 受Claude Code限制的分组只显示Claude客户端标签，分组限制改变时重置当前客户端选择。 | 前端不代替服务端授权；其它分组的现有客户端选项保持。 |
