# 上游改动 → 本仓库落点

只列本批 6 项。逐条的上游 diff 摘要见 `docs/upstream-sync/PORTING-0.2.0.md` §3。

## 1. `e93e6368f` 调度快照投影保留透传开关

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `repository/scheduler_cache.go` | `filterSchedulerExtra` 的 `keys` 列表，`"openai_responses_supported"` 之后 | 手写插入两行 + 上游那段注释（上下文因缺指纹两 key 漂移） |
| `repository/scheduler_cache_unit_test.go` | 同名文件 | 照搬（**必须含 JSON round-trip 那一跳**） |

命中路径：所有 OpenAI 透传账号的选号阶段（`sched:meta:<id>` 投影 → `ListSchedulableAccounts`
→ `IsModelSupported`）。转发阶段本来就对，不受影响。

## 2. `200b1406d` Anthropic `fallbacks` 剥离

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `pkg/claude/constants.go` | beta 标识常量区 | 照搬（+11） |
| `service/gateway_request.go` | 与既有 `context_management` sanitize 同一处 | 照搬（+77-14） |
| `service/bedrock_request.go` | Bedrock 请求装配 | 照搬（+35） |
| `service/gateway_fallbacks_sanitize_test.go` | 同名新建 | 照搬（+295） |

## 3. `1dc0a0900` ctx_pool WS ingress 容量降载改写

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `service/openai_ws_forwarder_ingress.go` | `ProxyResponsesWebSocketFromClient`（`:43`）里 `writeClientMessage(upstreamMessage)` 处，本仓库在 `:1039`（上游 `:1125`） | 照搬（引入独立 `clientMessage` 变量） |
| `service/openai_ws_ingress_capacity_shed_test.go` | 同名新建 | 照搬（+184） |

## 4. `6d5f02784` 空闲 WS 连接回收

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `service/openai_ws_pool.go` | 常量块新增 `openAIWSConnIdleRecycleAfter`；`cleanupAccountLocked` 里 `maxAge` 判定之前插一段 | 照搬。**常量块的缩进对齐是格式改动，不是语义** |
| `service/openai_ws_pool_test.go` | 同名文件 | 照搬（+26） |

## 5. `ba345f105` + `57c76584a` Codex 目录两条

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `service/openai_codex_models_service.go` | 目录账号筛选 / service tier 透出两处 | 照搬（两条都 `CLEAN`） |
| `handler/gateway_models_test.go`、`handler/openai_codex_models_handler_test.go`、`service/openai_codex_model_metadata_test.go`、`service/openai_codex_models_service_test.go` | 同名文件 | 照搬 |

顺序：`ba345f105` 先、`57c76584a` 后（上游顺序；两条改同一文件的不同位置，反序也能打，但按上游序省事）。

## 6. `9eabd2a5b` + `e7c029875` 账号统计成本

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `service/account_stats_pricing.go` | `tryModelFilePricing` 整个函数体 | 手写替换（本仓库条件少 `"fast"`，见 `source-baseline.md` §2） |
| `service/account_stats_pricing_test.go` | 同名文件 | 照搬（`e7c029875` 图片输出按 output 子集计价） |

⚠️ 上游 commit message 的动机是 DeepSeek 官方峰谷价（`b5827cfd` 引入），**本仓库无 DeepSeek 平台
⇒ 那部分收益 N/A**。本仓库拿到的是「删掉第二份实现」的结构收益 + `service_tier=fast` 不再落到手算分支。
