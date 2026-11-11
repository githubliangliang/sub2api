# Why

v0.2.21包含本地既有Claude Code、Codex及共享网关的协议、数据与资源问题。本批按本地消费者逐簇选择，Grok通过共享兼容、计价和会话路径受益。

# What Changes

- F01 Codex 元数据与 failover：Codex rate_limits/response.metadata SHALL 保持 attempt-local，真实模型输出前不得提交下游响应或结束首输出计时。
- F02 Responses 密封参数与 namespace：Responses 缓冲重建 SHALL 在函数调用中保留 namespace 和 encrypted_function_args 的缺失、null、空数组及非空值区别。
- F03 CC Switch reasoning 封装：Anthropic 转 Responses SHALL 忽略 ccswitch-openai-reasoning-v1 外来封装，并保留原生密文及工具历史。
- F04 Responses 工具回合会话种子：未提供显式会话ID的 Responses 请求 SHALL 仅使用稳定前缀和首个用户输入派生会话种子，后续 developer/system 提醒不能改变已有会话。
- F05 账号补全失败释放并发槽：Gateway 选中账号后补全失败时 SHALL 释放已取得的并发槽，并保留本地账号会话释放。
- F06 价格目录 hash 与正文一致：价格同步 SHALL 使用已下载正文的实际 hash；远端 hash 超前时不能把旧正文标记为最新。
- F07 Sonnet 5.5 缓存读兜底价：内置 Sonnet 5.5 缓存读取价格 SHALL 与 tag 的 0.10 USD/百万 token 一致，覆盖资源和两层 fallback。
- F08 Gemini 非流式正文排除思考摘要：Gemini 非流式兼容响应 SHALL 排除 thought=true 的文本，不将思考摘要混入正文，并保留正常文本、工具和用量。
- F09 Antigravity 并行工具图片顺序：非原生多模态工具输出的转换 SHALL 连续写出所有 functionResponse，再追加附带图片。
- F10 Codex 图片请求保留错误状态码：携带 X-Codex-Image-Turn-Id 的既有非流式图片入口 SHALL 延后提交响应，保留最终上游错误HTTP状态。
- F11 Responses UTF-8 输入边界：Responses 请求体 SHALL 在JSON转换和路由前拒绝非法UTF-8并返回明确的400错误。

# Impact

基线VERSION1.1.24 / Go1.27.3 / SQLite228 / auth snapshot21。依赖无需随上游升级。本批无拟议schema变更。

# Non-goals

不引入30m/12h TTL schema、新模型/平台、自适应预算新策略、支付充值购买退款、PG迁移、发布、推送或业务实例部署。第三/四档保持独立，不用本提案扩大授权。
