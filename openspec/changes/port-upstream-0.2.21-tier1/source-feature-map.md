# Source Feature Map

实现范围以本表行为为限；源码位置仅在PORTING报告。

| 簇 | 来源 | 选定行为 | 边界 |
|---|---|---|---|
| F01 Codex 元数据与 failover | PR8576 | Codex rate_limits/response.metadata SHALL 保持 attempt-local，真实模型输出前不得提交下游响应或结束首输出计时。 | 普通 response.metadata 对照通过；真实文本、工具输出及终态按既有语义发送，不把所有 codex.* 任意当作可丢弃内容。 |
| F02 Responses 密封参数与 namespace | PR8585 | Responses 缓冲重建 SHALL 在函数调用中保留 namespace 和 encrypted_function_args 的缺失、null、空数组及非空值区别。 | 现有 terminal 与 item.done 原文路径已经通过；只补 delta accumulator 缺口，不重复重构已实现的终态保真。 |
| F03 CC Switch reasoning 封装 | PR8386 | Anthropic 转 Responses SHALL 忽略 ccswitch-openai-reasoning-v1 外来封装，并保留原生密文及工具历史。 | 文本、tool_use、tool_result、签名顺序不变；不把未知所有签名一概删除。 |
| F04 Responses 工具回合会话种子 | PR8589 | 未提供显式会话ID的 Responses 请求 SHALL 仅使用稳定前缀和首个用户输入派生会话种子，后续 developer/system 提醒不能改变已有会话。 | 首个系统/开发者提示、模型、工具、instructions、首用户差异仍隔离；显式 session-id/prompt_cache_key 优先。 |
| F05 账号补全失败释放并发槽 | PR8631 | Gateway 选中账号后补全失败时 SHALL 释放已取得的并发槽，并保留本地账号会话释放。 | 本地 newSelectionResult 多一个 sessionID 参数且已有 ReleaseAccountSession；不可用上游函数整体覆盖。 |
| F06 价格目录 hash 与正文一致 | PR8384 | 价格同步 SHALL 使用已下载正文的实际 hash；远端 hash 超前时不能把旧正文标记为最新。 | 远端hash缺失/大小写、无效JSON、网络失败、fallback/override及热重载并发保留，失败不发布半成品。 |
| F07 Sonnet 5.5 缓存读兜底价 | PR8381 | 内置 Sonnet 5.5 缓存读取价格 SHALL 与 tag 的 0.10 USD/百万 token 一致，覆盖资源和两层 fallback。 | 显式自定义0与覆盖价优先；其它模型和输出/缓存创建费率不改。 |
| F08 Gemini 非流式正文排除思考摘要 | PR8584 | Gemini 非流式兼容响应 SHALL 排除 thought=true 的文本，不将思考摘要混入正文，并保留正常文本、工具和用量。 | 本地缺 isGeminiThoughtPart，但来源helper仅4行布尔判断；只补这一最小前置，不导入320行完整thinking策略。 |
| F09 Antigravity 并行工具图片顺序 | PR8621 | 非原生多模态工具输出的转换 SHALL 连续写出所有 functionResponse，再追加附带图片。 | 支持原生 functionResponse.parts 的模型继续嵌入媒体；单工具及无图片顺序不回退。 |
| F10 Codex 图片请求保留错误状态码 | PR8379 | 携带 X-Codex-Image-Turn-Id 的既有非流式图片入口 SHALL 延后提交响应，保留最终上游错误HTTP状态。 | 其它客户端保留既有心跳，流式不改；该header不是认证。只维护现有图片入口，不引入2.6或新原生路由。 |
| F11 Responses UTF-8 输入边界 | PR8575 | Responses 请求体 SHALL 在JSON转换和路由前拒绝非法UTF-8并返回明确的400错误。 | 本次探针仅证明非法字节进入model字段检查，不证明曾产生真实上游503或PG故障；保留大小/预算释放和鉴权次序。 |
