# 协议实施期间的编译夹具适配

以下为实施时实际工具输出的摘录与修正说明，并非重新执行或原始完整日志。`protocol-after.log` 等路径曾被下一次运行覆盖，最终完整成功输出与以下已观察到的中间失败严格区分。

- 来源gate helper在本地不存在：`internal/service/openai_gateway_request_body.go:913 undefined: JSONMayEscapePlainASCII`。用本地 `bytes.Contains(body, []byte(` + "`\u`" + `))` 必要条件替代，没有新增完整框架。
- 来源协议resolver在本地不存在：`internal/service/openai_codex_model_metadata.go:221:46: undefined: resolveUpstreamProtocol`、`undefined: APIProtocolResponses`；`openai_codex_models_service.go:2383`同样。改用本地已有的 `shouldForwardOpenAIResponsesViaRawChatCompletions`。
- 共享树并行修改被中间读取：`internal/service/upstream_024_antigravity_quota_test.go:48:47: too many arguments in call to NewAntigravityQuotaFetcher; have (nil, *config.Config); want (ProxyRepository)`。由负责S15代理适配后通过；没有改变其产品策略来绕过编译。
- 本地夹具删去未使用loopback构造器时短暂留有 `openai_buffered_024_test.go:101 undefined: server`，移除残留引用。真正loopback SSE测试仍通过本地HTTP读取器执行；stop forwarding测试使用现有HTTPUpstream recorder。
- S13把旧sanitizer函数搬到独立差分控制时遗漏测试import：`internal/service/hotpath_024_test.go:409:2 undefined: sort`，补测试import。
- 主代理默认lint发现 `openai_buffered_024_test.go:100 undefined forceChatResponsesFallbackAccount`，其helper仅unit标签可见；为文件补 `//go:build unit`，并不改产品代码。

以上均为构建/夹具适配，不计入修复前已编译的目标行为失败数量。真实行为失败完整日志在 `protocol-before.log`、`protocol-audio-before.log` 与 `protocol-buffered-before.log`；采纳前候选差分、采纳后回归、最终组合门禁分别记录。
