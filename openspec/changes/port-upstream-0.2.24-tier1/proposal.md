# Why

上游v0.2.24涉及现有网关解析、Claude/Codex/Grok协议与账号数据正确性，需要按SQLite个人部署进行选择性维护。

# What Changes

- F01 Codex 文献搜索 403 的账号副作用：alpha/search 工具级403 SHALL 允许既有换号流程且不冷却或永久置错健康账号；明确账号/工作区停用的403继续执行现有账号策略。
- F02 OpenAI 408 上游超时换号：上游 HTTP408 SHALL 进入既有有界换号流程。
- F03 Codex IDE 与独立上下文：合法独立 agent 上下文 SHALL 转为不可信 user 文本，保留各类型的空值规则。
- F04 Responses 每轮请求推理强度：请求推理强度日志 SHALL 优先记录本轮原始input最后一个有效configuration_update的effort，并与最终有效强度独立。
- F05 既有 Claude 账号测试 temperature：既有拒绝temperature的Claude家族在连接测试中 SHALL 省略该字段。
- F06 兼容转换保留 stop 与原始工具定义：既有兼容降级路径 SHALL 保留客户端stop序列与可转发的原始动态工具定义。
- F07 图片 SSE usage 单行载荷：转发图片的completed事件 SHALL 把usage压为单行JSON后写入SSE data行。
- F08 Gemini 最后工具轮的尾随内容：最后一个Gemini工具结果turn的尾随非functionResponse内容 SHALL 保序拆成后续user turn。
- F09 Antigravity 原生 SSE 注释帧：支持注释的客户端 SHALL 保留上游SSE注释及其分隔；不支持注释的既有SDK继续滤掉注释。
- F10 Grok 媒体输入错误的 400 分类：既有Grok媒体本地输入校验失败 SHALL 返回400并保留可用校验信息。

# Impact

不需要新增数据库列。保持VERSION1.1.26、Go1.27.4、SQLite230/auth21和依赖。新增共享入口检查需覆盖所有现有HTTP别名和WS轮帧；OAuth CAS适配SQLite，默认配额消费者实施时重新生成Wire。真实影响面由所选行为决定，不能以文本CLEAN替代编译。

# Non-goals

不导入支付或商业流程、缺失TypeSafe/DeepSeek native平台、新搜索供应商/来源块/ToolSearch能力、5图扩展、费用估算/刷新提示/复制ID、完整group allowlist、QuotaSummary、shadow、PGbatch/lane/COW/stale策略。发布、push、归档和业务实例操作另行处理。
