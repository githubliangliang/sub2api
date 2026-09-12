# 提案

本 fork 当前可能把未耗尽配额的 429 记成长期停调、按错误的模型选号、漏掉渠道限制，
以及在请求规范化时损坏控制字符、丢失 reasoning 或注入 API Key 未要求的 instructions。
Claude 和 Antigravity 请求还存在已经能定位到本地调用点的兼容性缺口。

本 change 按 F01-F13 修正这些既有行为。所有新测试须有明确触发条件和正常路径反例；
调度、Spark shadow、SQLite 用量写入和已有 namespace/Responses Lite 处理保持既有契约。

成功标准：13 项规格满足，相关回归与后端构建、unit、SQLite 方言审计通过，来源与落地 SHA 可追溯。
这里不包含第二档的 WS 状态重构、客户端依赖、模型大簇、UI，也不引入第三、四档功能。
