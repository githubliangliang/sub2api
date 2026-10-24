# 实施验收证据

F01–F13 产品修复已完成；只记录本次执行的证据，不复用评估 PASS 或前任未保存的测试日志。

实际起点：`cf6b42752f88ee4eefe51e1c1c59ca1608ac20f5`，父提交 `c9d9bebe87926791bfdd716f59abcba413fee20a`。
该快照包含用户 metadata 清理和 intake 文档。接手时有 F04/F05/F06 产品代码以及
F01/F03–F12 未提交测试；全部保留并检查。仅在 `sync/upstream-024-p0-RLC1zS` 工作树实施。

产品完成 SHA：`0eb88e504244c429ba19b0486b06c8011d78da79`。后续仅本 change 的验收文档提交。
本分支的四个实施提交（按顺序）：

- `80bc5baf0a2f187f6700b5f331e5edf6e9d2e84d`：F04/F05/F06。
- `25daf1281cc860389b84fbc5cfdf9f3c6ae62ee4`：F01/F03/F10。
- `fd0080650fb6e4da30ec59ac4190d2bbb401aa68`：F07/F08/F09/F11/F12。
- `0eb88e504244c429ba19b0486b06c8011d78da79`：F02/F13。

执行日期：2026-09-12。

| 项目 | 修复前失败证据 | 修复后与反例证据 | 落地 SHA |
|---|---|---|---|
| F01 | 普通/37% 窗口/过期正文 reset 在关闭 fallback 时仍建 runtime block；非耗尽计算返回长 reset | 禁用 fallback 不建 block；明确耗尽、future body reset、既有 Spark 早退/重试规则 PASS | `25daf1281` |
| F02 | 三个 HTTP 入口 404 model_not_found；WS 关闭为 no available account | 四入口真实 handler/scheduler、上游正文、请求模型/计费 chain、WS 两轮响应别名 PASS；空/无映射 fallback PASS | `0eb88e504` |
| F03 | 五条负载感知回归选中受限账号 1 或未返回 unavailable | routed、routed sticky、sticky、普通候选、全受限不占槽、关闭限制反例 PASS | `25daf1281` |
| F04 | baseline overlay：unavailable 不恢复，仅转发一次，工具历史重放失败 | 恢复完整工具历史、重试仅一次、无关错误/403 不恢复 PASS | `80bc5baf0` |
| F05 | baseline overlay：API Key passthrough 的 none 被删除 | passthrough body 保留，已有 OAuth/普通账号与 Lite 规则 PASS | `80bc5baf0` |
| F06 | baseline overlay：5 个控制字符案例 JSON 解码失败 | 标准 JSON 解码、内容不变、cache 元数据与幂等 PASS | `80bc5baf0` |
| F07 | 8 个 wire UA/body 案例版本/后缀不一致；版本变化仍保留旧 suffix | 两个出口 × mimic/无 identity/关闭 FP/缓存 UA PASS；使用独立 SHA256 预期及手算固定 suffix | `fd0080650` |
| F08 | 三种模型 max_tokens=1 被拒；ParsedRequest 热路径丢 max_tokens | float64/int/int64 探测、热路径、严格 UA、普通/非法 max_tokens 反例 PASS | `fd0080650` |
| F09 | 无 beta 仍带 binding；mimic 最终 header 缺 token | messages/count_tokens 实际 wire body/header 三分支及纯函数/既有 beta 反例 PASS | `fd0080650` |
| F10 | Fable credits 建账号冷却；持久化失败场景未进入模型级调用 | 模型 scope、fallback、非 Fable、共享真实耗尽、写库失败不扩大账号冷却 PASS | `25daf1281` |
| F11 | reasoning 无工具序列化结果缺 toolConfig | 无工具/有工具 reasoning、普通模型、混合工具 includeServerSideToolInvocations PASS | `fd0080650` |
| F12 | SQL 参数零变 NULL；真实 SQLite 写入零延迟读回 NULL | 全部 19 个整型指标 0/7/nil 写入真实 SQLite；group_id=0 仍 NULL PASS | `fd0080650` |
| F13 | API Key omitted/empty 被注入 Codex prompt；raw 大整数回归多出 instructions | API Key omitted/explicit/empty/passthrough、OAuth default/explicit 与 metadata 清理 PASS | `0eb88e504` |

| 门禁 | 命令 / 结果 |
|---|---|
| 后端 build | `GOMAXPROCS=4 go build -p 2 ./...`，`0eb88e504`，exit 0 |
| 后端 unit | `GOMAXPROCS=4 go test -p 2 -tags=unit -count=1 -json ./...`，`0eb88e504`，exit 0；包括 service 156.790s。首轮中断不计 PASS |
| SQLite 方言 | `GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/repository -run '^TestProductionSQLUsesSQLiteDialect$' -count=1 -v`，`0eb88e504`，exit 0；另有 F12 真实 SQLite 写读 |
| 用户已有 metadata 修复 | `GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/service -run 'Test(StripOpenAIResponsesInputContentItemKinds\|PrepareOpenAIWSHTTPBridgeBodyStripsInputContentItemKinds\|OpenAIGatewayService_HTTPStripsInputContentItemKindsBeforeFirstForward)' -count=1 -v`，exit 0（表格中的反斜杠仅用于转义管道符） |
| 排除的无关 hunks | 按下面来源边界逐项记录；无 tier 2 代码、第三/四档、外网 token 测试调整 |
| diff 范围检查 | `git diff --check cf6b42752` exit 0；无 VERSION/迁移/Ent/Wire/依赖/前端变动，无 coordinator 所有文档变动 |

- [x] 13 项行为均有验证证据。
- [x] 来源冻结保持不变，实际落地与缺口如实记录。

F04–F06 green 命令（backend 目录；2026-09-12，exit 0）：
```sh
GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/service -run 'Test(AddMessageCacheBreakpoints|ForwardAsAnthropic_.*(PreviousResponse|Continuation)|OpenAICompatPreviousResponse|FilterOpenAIResponsesNoneReasoningEffortForAccount|NormalizeOpenAIParallelToolCallsWithoutTools)' -count=1
```
接手时三处产品代码已改变，red 使用 Go `-overlay` 将这三处映射到 `git show cf6b42752:<path>`
生成的原始文件，测试仍使用接手版本；未覆盖工作树产品文件。
overlay 的三个路径为 `backend/internal/service/` 下的 `gateway_messages_cache.go`、
`openai_gateway_request_body.go`、`openai_messages_continuation.go`。
其他产品文件当时仍为 `cf6b42752`。批量 red 原命令如下（exit 1，实际日志 `inherited-red.log`）：

```sh
GOMAXPROCS=4 go test -p 2 -tags=unit -overlay=../.tier1-evidence/baseline-overlay.json ./internal/service ./internal/handler ./internal/repository ./internal/pkg/antigravity -run 'Test(AddMessageCacheBreakpoints_JSONStringEscaping|ForwardAsAnthropic_.*(Unavailable|RetryFailure)|OpenAICompatPreviousResponseUnavailable|FilterOpenAIResponsesNoneReasoningEffortForAccount|CalculateOpenAI429ResetTime|OpenAI429FastPath|SelectAccountWithLoadAwareness_.*(Restriction|RestrictModels)|SyncBillingHeader|BuildOAuthRequest_BillingMatchesWireUserAgent|SanitizeAnthropicBodyForBetaTokens|ComputeFinalAnthropicBeta|ClaudeCodeValidator|.*MaxTokensOne|.*FableCredits|.*NonFableCredits|.*SharedWindowStillWins|ClaudeToGemini|InsertSystemMetrics)' -count=1
```
该批次未匹配 F08 handler 和 F11 新测试，所以另外运行下文对应 red；没有把 no tests to run 当作通过证据。

F01/F03/F10 green（exit 0）：
```sh
GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/service -run 'Test(CalculateOpenAI429ResetTime|OpenAI429FastPath|OpenAIOAuth429|ShouldStopOpenAIOAuth429|Handle429|HandleUpstreamError_Anthropic|SelectAccountWithLoadAwareness|.*Spark.*429|.*429.*Shadow)' -count=1
```
F10 额外 red：`GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/service -run 'TestHandleUpstreamError_AnthropicFableCredits(Persistence|Fallback)' -count=1`，exit 1；写库失败未走模型 scope（expected 1, actual 0）。禁用 fallback 为已通过反例。
F03 源补丁在本地 fresh DB recheck 上下文冲突，按行为手工合并且保留全部 fresh recheck。
F01 没有引入上游 quota classifier、retry-start map 或 retry-window 政策；仅改 reset 判据与 fallback gate。

F07/F08/F09/F11/F12 green（四包 exit 0）：
```sh
GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/service ./internal/handler ./internal/pkg/antigravity ./internal/repository -run 'Test(SyncBillingHeaderVersion|BuildOAuthRequest_BillingMatchesWireUserAgent|BuildAnthropicRequest_ThinkingBinding|SanitizeAnthropicBodyForBetaTokens|ComputeFinal.*AnthropicBeta|ClaudeCodeValidator|SetClaudeCodeClientContext|ToolConfigAlwaysPresent|TransformClaudeToGemini|GeminiToolConfig|InsertSystemMetrics)' -count=1
```
额外 red（各 exit 1，先确认行为失败再应用产品修复）：
```sh
GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/handler ./internal/pkg/antigravity -run 'Test(SetClaudeCodeClientContext_ParsedRequestProbeWithoutSystemPrompt|ToolConfigAlwaysPresent)' -count=1
GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/service -run 'Test(BuildAnthropicRequest_ThinkingBindingMatchesFinalBeta|SyncBillingHeaderVersion|BuildOAuthRequest_BillingMatchesWireUserAgent)' -count=1
GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/repository -run 'TestInsertSystemMetrics_SQLiteDistinguishesZeroFromMissing' -count=1
```
F12 测试读取既有 SQLite 033/042b 的 metrics DDL，没有新增/改写迁移或调用 PG integration。

F02 red/green：相同命令先 exit 1（三个 HTTP 404、WS no available account），修复后 exit 0：
```sh
GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/handler -run 'TestOpenAI(HTTP|ResponsesWebSocket)_ChannelMappedTargetSelectsAccountWithoutRequestedAlias' -count=1
```
green 同时加入 `|TestOpenAIChannelForwardModelForScheduler`，覆盖 trim、无映射、空目标。
fixture 初稿误用了 passthrough（按本 fork 契约跳过模型白名单）及旧 use_responses_api 字段，
并缺 availability 诊断方法；已改成非透传 + 正式 responses mode + 真实负载筛选 + 完整诊断 fixture。
这些测试准备错误不算 red，最终 red 是上述四个行为失败。

F13 red（exit 1）：
```sh
GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/service -run 'Test(OpenAIForward_InstructionsFollowAccountProtocol|OpenAIGatewayService_Forward_APIKeyMissingInstructionsKeepsLargeInputRaw)' -count=1
```
F13 green（exit 0）：
```sh
GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/service -run 'Test(OpenAIForward_InstructionsFollowAccountProtocol|OpenAIGatewayService_Forward_APIKeyMissingInstructionsKeepsLargeInputRaw|OpenAIGatewayService_.*(Instructions|Passthrough_PreservesBody)|StripOpenAIResponsesInputContentItemKinds|OpenAIGatewayService_HTTPStripsInputContentItemKinds|BuildOpenAIWSHTTPBridge.*ContentItem)' -count=1
```
F13 只添加 v0.2.4 三行谓词等价实现并守卫一处合成调用，未带 nativeDeepSeekResponses 或全仓谓词替换。
新增组合 fixture 曾把 OAuth 通用 map 解码的大整数精度问题纳入，发现为原有路径行为；未扩范围修复，
组合回归改用实际 metadata 字符串，API Key 原始大整数仍由独立 hotpath 回归覆盖。

来源复核使用 `GIT_INDEX_FILE=<独立临时 index> git read-tree cf6b42752`，
对 source-feature-map 每行的完整 `git diff <merge>^1 <merge>`，按文件先执行
`git apply --cached --reverse --check`，再执行 `git apply --cached --check`。
37 个文件记录为 ALREADY 0 / CLEAN 30 / CONFLICT 7 / NOBASE 0；这只说明可应用性，不代替编译和行为测试。

| ID | CLEAN | CONFLICT | 适配 / 排除 |
|---|---:|---:|---|
| F01 | 2 | 2 | 排除 absent quota classifier/retry-window 栈与依赖该栈的上游 fixture；保留 Spark 早退及既有 retry/storm 路径 |
| F02 | 3 | 2 | 排除 `openai_gateway_count_tokens_test.go` 的外网 API 超时 skip 调整；不引入 S06 forward-model context 基座 |
| F03 | 1 | 1 | 保留本 fork fresh DB recheck；按相同逐账号规则接入所有负载层 |
| F04 | 2 | 0 | 全行为移植；无额外排除 |
| F05 | 1 | 1 | 使用单独 passthrough 回归文件，保留既有 reasoning/Lite 测试 |
| F06 | 2 | 0 | 全行为移植；无额外排除 |
| F07 | 4 | 0 | 全产品行为；测试预期改为固定 suffix / 独立 SHA256，不复用被测 fingerprint helper |
| F08 | 4 | 0 | 全行为，追加 int64 和非探测值反例；不放宽 UA |
| F09 | 3 | 0 | 全行为，追加两个出口最终 header/body 的联动测试 |
| F10 | 2 | 0 | 全行为，追加写库失败与禁用 fallback 反例 |
| F11 | 2 | 0 | 全行为，删除来源补丁遗留的“过滤空 ToolConfig”矛盾注释 |
| F12 | 2 | 0 | 保留本 fork SQLite DDL/参数；追加真实 SQLite 19 指标回归，不改变通用 ID nullable helper |
| F13 | 2 | 1 | 只取合成谓词及相关回归，不引入 nativeDeepSeekResponses 或全仓账号谓词替换；API Key passthrough 由新增组合测试覆盖 |

冻结文件与实际起点逐字节相同，SHA256：

- source-baseline.md：`565ad1f41f259c603261fe4941b5338d044b31ba86b3b22c5a0e042a007794f3`
- source-feature-map.md：`5be9f2c1689835f0e821c0642562fffdd36d4f500137c9dc3e29ea6fa2a0bdf7`

用户的 `openai_responses_input_metadata.go`、对应测试及 `openai_ws_http_bridge.go` 相对
`cf6b42752` 没有 diff；`openai_gateway_forward.go` 只有 F13 guard 一行变化。
全部新调用的项目符号都有本地定义，fixture 在编译和实际行为调用中得到验证；未新增仅供测试调用的产品方法。
没有迁移、schema、Ent、Wire、依赖、frontend、VERSION 改动；SQLite/miniredis 路径保持原实现。
没有执行 PG integration、生产流量测试、push、发布或上游整仓 merge。

交给 coordinator 的集成要点：

- F01/S03 共用 `openai_account_runtime_block_fastpath.go`：保留 F01 的 disabled-fallback 早退及两个 Spark 早退；S03 的陈旧 block 同步不得恢复默认冷却或更改本 fork retry 政策。
- F02/S06 共用三个 handler：本次新增 `forwardModel`/`wsForwardModel` 只供选号；S06 若补同名变量或 context setter，应复用变量，同时保留 `reqModel` 的客户端/计费归属。复跑四入口映射测试。
- F05/S06 共用 request body：保留 `IsOpenAIPassthroughEnabled` 的 none-preservation 早退。
- F13/S01/S06 共用 Forward：保留 `UsesOpenAICodexProtocol` 的合成 guard、用户 content_item_kinds 清理，以及原 namespaces/Lite 路径；不得让 API Key 再获得隐式 Codex prompt。
- F07/S11 共用 Claude 身份链：合并配置版本提升后，`effectiveBillingUserAgent` 必须仍与实际 messages/count_tokens 的最终 UA 一致；复跑 `TestBuildOAuthRequest_BillingMatchesWireUserAgent` 和 S11 版本覆盖测试。

coordinator 已确认集成至 `0eb88e504`；root CLAUDE.md、upstream-sync README/PORTING 的总状态由 coordinator 更新。

完整 unit 重跑于 2026-09-12 14:14:05–14:17:39（Asia/Shanghai）完成，进程 exit 0：
54 个有测试包 PASS，52 个包无测试；包含子测试共 16,508 PASS、15 个既有条件跳过、0 FAIL。
`-count=1` 禁用结果缓存。首轮 `unit.log` 在 provider HTTP 503/工具中断后没有 service 结果，明确标为 interrupted，
不与本轮通过统计合并。没有因为中断更改产品或测试契约。

原始日志、源完整 diff、独立 index 四态 TSV 和门禁汇总保存在
`/tmp/sub2api-port-024-RLC1zS/tier1-evidence/`；`gate-results.json` 记录完整命令、退出码及起止时间，
`unit-resumed.jsonl` 和 `unit-resumed.exit` 是完整重跑的终态证据。
最初 red 的 overlay 绝对路径指向运行时的 p0/.tier1-evidence；归档后若重放，需要按新位置重新生成 overlay。
本 change 提交验收摘要，临时原始日志放在工作树外；最终分支只含四个产品提交与本地验收文档提交。
