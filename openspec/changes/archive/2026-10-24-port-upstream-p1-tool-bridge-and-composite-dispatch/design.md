# 设计与决策

## 1. 移植策略

本批**只移植**，不重新设计。取舍标准是三条同时成立：

1. 本仓库确实患同一缺陷（grep 到改前形态，不是「上游改了所以我们也改」）；
2. 上游 patch site 与本仓库字节一致，或差异只在上下文（结构体/函数本体一致）；
3. 与本批其它条目、以及仍未排期的大簇之间**没有文件冲突**。

不满足第 3 条的条目一律留到整簇排期，即使它单独看起来很干净——0.1.180 §6.1 的 `31d5b67ba`
就是这样被排除的（5/6 干净，但第 6 个文件属于另一条还没做的成对项的文件族）。

## 2. 阶段划分

原计划五个阶段，阶段 4 在实施中撤回 ⇒ **实际交付四个**。每阶段独立提交、独立可回滚；
阶段之间没有依赖，唯一的顺序约束在阶段 3 内部。

| 阶段 | 内容 | 触碰文件 |
|---|---|---|
| 1 | `cc894ef57` CC 流式身份字段 | `service/` 2 产品 + 2 测试（1 个新建） |
| 2 | `4d4a0be1a` file part → `input_file` | `apicompat/` 2 产品 + 1 测试 |
| 3 | `25da02ddd` → `66808413d` 回放完整性（**顺序固定**） | `service/` 3 产品 + 3 测试 |
| ~~4~~ | ~~`17c0ee385` Grok 422~~ **撤回** | 无（见决策 7） |
| 5 | `68653fb2c` composite `/v1/messages` | `service/` + `handler/` + 前端 2 文件 |

## 3. 决策

### 决策 1：`types.go` 手写落地，不强行 apply

`4d4a0be1a` 的 `apicompat/types.go` hunk `CONFLICT`，但原因是本仓库在**相邻的** `ResponsesTool` /
`ChatTool` 上有自有的 `x_search` 字段，把上下文推移了。被改的 `ResponsesContentPart` 与
`ChatContentPart` 两个结构体本身与上游 pre-image 字节一致。

⇒ 按 hunk 手写三处（`ResponsesContentPart` 加 3 个 `input_file` 字段、`ChatContentPart` 加 `File`
字段、新增 `ChatFile` 结构体），**不要**用 `-3` / `--reject` 之类的模糊匹配把 hunk 硬塞进去，
那会在 `x_search` 字段之间留下错位。

### 决策 2：新增测试追加到文件末尾，不移动既有测试

阶段 3 的两条在 `openai_ws_http_bridge_test.go` 上 `CONFLICT`，原因是上游的插入锚点
`TestOpenAIWSHTTPBridgeAPIKeyReusesClientToolMappingWhenFollowupOmitsTools` 本仓库不存在。

⇒ 把上游新增的两个端到端测试**追加到本仓库该文件末尾**，并直接以 `66808413d` 修改后的形态写入
（即第一个测试用 4 个 upstream response、中间插一轮 `previous_response_id` 的孤儿场景），
不要先写 `25da02ddd` 的形态再改一遍。所需的 import 与 `httpUpstreamRecorder` 本仓库均已具备
（后者在同包的 `openai_oauth_passthrough_test.go`）。

### 决策 3：阶段 3 的顺序不可交换

`25da02ddd` 把 `needsBridgeReplay` 的判据从 `openAIWSRawPayloadHasToolCallOutput(payload)` 换成
`AnalyzeToolCallOutputContextCoverageBytes(payload)` 的两个字段；`66808413d` 的
`sanitizeOpenAIWSHistoricalReplayToolCalls` 是在**新判据已经生效**的前提下，进一步过滤回放历史里
没有配对输出的孤儿 context item。

先合 `66808413d` 会让它作用在旧判据上：旧判据只要 payload 里有任何 tool call output 就回放，
于是「本轮已自带完整历史」的请求仍然会被回放一次，只是回放内容少了孤儿项——错误的行为 + 更难
定位的症状。⇒ **必须 `25da02ddd` 在前**，且两条同一阶段落地。

### 决策 4：`AnalyzeToolCallOutputContextCoverageBytes` 同时接受数组与对象

上游把 `if !input.IsArray() { return coverage }` 放宽成 `!IsArray() && !IsObject()`，并把原来的
`ForEach` 闭包抽成 `analyzeItem`，对象形态直接分析自身。这不是重构而是修缺陷：客户端把单个
`custom_tool_call_output` 作为 `input` 直接发上来（非数组）时，旧代码整个函数直接返回零值覆盖度，
判定退化。⇒ 照搬，包括把 `return true`（ForEach 的继续信号）改成 `return`（普通函数的提前返回）。

### 决策 5：composite 闸门——丢掉 CN 分支，保留本仓库的 `ctx` 签名，并把豁免收窄

上游 `68653fb2c` 的 handler hunk 是在它自己的 CN 形态上改的，本仓库不能照搬：

| 上游 | 本仓库 | 处理 |
|---|---|---|
| `if service.IsCNProvider(apiKey.Group.Platform) { return true }` | 无 CN 平台，`IsCNProvider` 零命中 | **整段丢弃** |
| `allowOpenAICompatibleMessagesDispatch(c *gin.Context, ...)` | 签名是 `(ctx context.Context, ...)` | **保留本仓库签名**，不为了对齐上游而改调用方 |
| composite 分支里 `platform == Grok \|\| IsCNProvider(platform)` | 只保留 `platform == Grok` | 去掉 CN 项 |

本仓库现状的第三个分支是「**只要**解析到的目标平台是 grok 就放行」，没有要求分组本身是 composite。
今天这等价于 composite 专用（`ensureCompositeTargetPlatform` 只在 composite 分组上写这个 context
值；另一处写入点 `server/routes/gateway.go:649` 写的是 `gemini`），但语义上更宽。
⇒ 按上游形态补上 `apiKey.Group.Platform == service.PlatformComposite` 的前置条件，
行为不变、意图变明确，也挡住将来出现「非 composite 分组也解析出 grok 目标」时的意外放行。

**存量数据说明（重要）**：`sanitizeGroupMessagesDispatchFields` 此前在 `Create` 与 `Update`
两条路径上都会把 composite 分组的 `AllowMessagesDispatch` 抹成 `false` 再落库，因此
**升级后不会有任何 composite 分组自动获得放行**——所有存量 composite 分组的该字段都是 `false`，
需要管理员在分组编辑页显式打开。这是本项唯一的「行为变化方向」说明：**开关从死的变成活的，
默认仍是关的**。

### 决策 6：`d5824f6a5` 判 N/A（本批不做，且建议在 PORTING 文档里定性）

0.1.180 §6.3 把它列为待判。拉上游 patch 实测后结论是**对本仓库空转**：

- 它新增的 `supportsOpenAIReasoningEffortMax(model)` 在 `isOpenAIGPT56Model` 之外只放开
  `deepseek-v4` / `glm-` / `kimi-` / `moonshot-` / `k3` 五个前缀——**全部是本仓库没有的平台**。
- 剩下的改动是把 mappedModel 透进 `extractCCReasoningEffortFromBody` /
  `ExtractResponsesReasoningEffortFromBody`，而本仓库主路径
  `service/openai_gateway_request_body.go:825` 早就在传 `firstNonEmpty(modelCandidates...)`；
  被补的两条是 CC/Responses → **Anthropic 原生**上游的路径，那里 mappedModel 是 claude 模型，
  `supportsOpenAIReasoningEffortMax` 仍返回 false，归一化结果不变。

⇒ 不合，并把 PORTING-0.1.180 §6.3 该行从留白改成 N/A + 理由，避免下一轮重复评估。

### 决策 7：`17c0ee385` 撤回——`apply --check` 干净的 no-op

原计划作为阶段 4 落地，实施时发现它**在本仓库是空转**，已撤回。

`forwardGrokResponses` 的重试是两层门：

```go
// 外层：17c0ee385 改的就是这一行
if attempt > 0 || (resp.StatusCode != 400 && resp.StatusCode != 422) { break }
respBody := s.readUpstreamErrorBody(resp)
// 内层：真正决定要不要剥掉加密 reasoning 重试
if !isGrokInvalidEncryptedContentResponse(resp.StatusCode, respBody) { ...; break }
```

本仓库 `openai_gateway_grok.go:255` 的内层函数第一行仍是
`if statusCode != http.StatusBadRequest { return false }` ⇒ **422 进了外层也会在内层被否掉**，
净效果只是把 422 的响应体多读一遍再放回去，没有任何请求会因此重试。

widen 内层的是 §6.2(c) 的 **`953028718`**（实测：`17c0ee385` 与 `953028718` 是
`v0.1.180` 区间里仅有的两个含 `StatusUnprocessableEntity` 的 commit）。它同时引入了
compaction 专属错误码（`invalid_compaction` / `compaction_decode_error`）与
`grokStructuredErrorMessageCandidates`（本仓库零命中）——也就是说 `17c0ee385` 标题里的
「compaction 422」，其识别能力全在 `953028718` 里。

⇒ **这条实质是 §6.2(c) 的尾巴，不是独立 bugfix**。只补外层守卫会得到一个「文档上写着已支持
422 重试、实际一次都不会触发」的假状态，比不做更糟。随 §6.2(c) 整簇排期。

⚠️ 这是 0.1.180 §2 第三种假信号的一个新变种：不是「引用了不存在的符号所以编译不过」，
而是**符号存在、编译通过、行为为空**。今后判定单行守卫类改动时，要连同它守卫的下游判定一起看。

## 4. 明确不做（本批边界）

| 条目 | 理由 |
|---|---|
| 0.1.180 §6.1 的 `e2d9ce0ca`+`fbc9ee626` / `31d5b67ba` / `7a09a2eaf` / `7498d8fdc` | 都落在 `chatcompletions_responses_bridge.go`、`openai_gateway_responses_chat_fallback.go`、`responses_client_tools.go`、`openai_gateway_grok.go` 这几个本仓库有自有改动的文件上，需逐 hunk 解，且 `7498d8fdc` 是 0.1.183 §5.1 整簇的基座 |
| 0.1.180 §6.2(c) 整簇（**含本批撤回的 `17c0ee385`**） | 含两组成对 revert（`6c3edc095`↔`e62ec2c42`、`726de3010`↔`ab9cb69e7`），且 `16b15e870` 说明它与 §7.1 大礼包互相打补丁；`17c0ee385` 依赖同簇的 `953028718`，见决策 7 |
| 0.1.180 §7.1 / §7.2 / §7.3 / §7.4 / §7.5 | 维持原判：大礼包多天量级且冲突最重；`service_tier` 与重置卡自动化都是「没在用就别加」；§7.4 本仓库已硬编码 8 MB；§7.5 依赖不做的模型广场 |
| 0.1.183 §4.1 监控 v2 composite | 需 SQLite 重写，且 `settings.channel_monitor_mode = 'v1'` ⇒ 当前零收益 |
| 0.1.180 §5.1 dompurify | 维持推迟。nanoid 已改用本地审计例外（`59a62bb66`，全部例外续到 `2026-10-06`），原先「四个文件一个都不动」的绑定已解开，留待下次前端依赖升级批次 |

## 5. 回滚

四个阶段互不依赖，可单独 revert：

- 阶段 1 / 2：纯增量，revert 即回到改前行为。
- 阶段 3：**两条必须一起 revert**，只回退 `66808413d` 会留下「新判据 + 不过滤孤儿」的中间态。
- 阶段 5：revert 后 composite 分组的开关重新变成死的；**已经被管理员打开的
  `allow_messages_dispatch` 值仍留在库里**，只是不再被读取，因此 revert 无数据副作用。

无迁移、无 schema 变更、无 wire 变更 ⇒ 不存在需要单独回滚的数据面。
