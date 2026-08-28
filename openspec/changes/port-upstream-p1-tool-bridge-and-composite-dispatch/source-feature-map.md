# 上游改动 → 本仓库落点

只列本批 6 条候选（第 5 条撤回）。逐条的上游 diff 摘要见 `docs/upstream-sync/PORTING-0.1.180.md` §6。

## 1. `cc894ef57` CC 流式 tool_call 身份字段

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `service/openai_gateway_cc_tool_call_identity.go` | 同名新建 | 照搬（`stripEmptyChatToolCallIdentityFromSSELine` + `stripEmptyChatToolCallIdentity`） |
| `service/openai_gateway_chat_completions_raw.go` | `streamRawChatCompletions` 的 SSE 行循环 | 在 `applyOllamaCloudRawChatCompletionsSSELine` 之后插一行 |

命中路径：所有走 raw CC 直转的账号（不限定某个上游）。热路径先 `bytes.Contains(payload, "tool_calls")`
快速失败，绝大多数 chunk 不进 JSON 解析。

## 2. `4d4a0be1a` Chat Completions file part

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `apicompat/types.go` | `ResponsesContentPart` +3 字段、`ChatContentPart` +1 字段、新增 `ChatFile` | 手写（上下文因本仓库 `x_search` 字段漂移） |
| `apicompat/chatcompletions_to_responses.go` | `convertChatContentPartsToResponses` 的 `switch p.Type` | 新增 `case "file"` |

## 3. `25da02ddd` 回放判定

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `service/openai_tool_continuation.go` | `AnalyzeToolCallOutputContextCoverageBytes` | 把 `ForEach` 闭包抽成 `analyzeItem`，并让 `input` 为**对象**时也走同一套分析 |
| `service/openai_ws_forwarder_ingress.go` | `ProxyResponsesWebSocketFromClient` 的 `needsBridgeReplay` | 从 `openAIWSRawPayloadHasToolCallOutput` 改成覆盖度判定 |

## 4. `66808413d` 孤儿 tool call

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `service/openai_ws_forwarder_payload.go` | 新增 `sanitizeOpenAIWSHistoricalReplayToolCalls`；`buildOpenAIWSReplayInputSequence` 里调用一次 | 照搬 |

## 5. `17c0ee385` Grok 422 —— **撤回，不交付**

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `service/openai_gateway_grok.go` | `forwardGrokResponses` 的重试守卫 | **不改**。补外层守卫是空转，内层 `isGrokInvalidEncryptedContentResponse` 仍硬门 400，见 `design.md` 决策 7 |

## 6. `68653fb2c` composite 分组 `/v1/messages`

| 上游文件 | 本仓库落点 | 动作 |
|---|---|---|
| `service/openai_messages_dispatch.go` | `sanitizeGroupMessagesDispatchFields` | composite 不再强制关 `AllowMessagesDispatch`（`DefaultMappedModel` / `MessagesDispatchModelConfig` 仍清空） |
| `handler/openai_gateway_handler.go` | `allowOpenAICompatibleMessagesDispatch` | composite 豁免收窄成「仅 grok 目标」；**上游的 `IsCNProvider` 两处分支整段丢弃** |
| `frontend/src/views/admin/groupsMessagesDispatch.ts` | 新增 `supportsMessagesDispatchPlatform` | 照搬 |
| `frontend/src/views/admin/GroupsView.vue` | 创建/编辑表单各一处 `v-if`、两处 watch、一处 import | 手写（本仓库该文件长期自有改动） |
| `handler/openai_gateway_cn_dispatch_test.go` | — | `NOFILE`，CN 专属，跳过 |
