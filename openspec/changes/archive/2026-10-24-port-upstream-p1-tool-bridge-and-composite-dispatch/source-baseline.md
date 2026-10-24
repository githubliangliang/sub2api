# 源基线固定

对照日期：2026-08-28。**实施期间不得改动本文的 SHA**；如需重新对照，新建下一轮 change。

## 1. 固定点

| 点 | 值 |
|---|---|
| 上游仓库 | `https://github.com/Wei-Shaw/sub2api` |
| 上游最新正式版 | `v0.1.183`（2026-08-25 13:53 UTC）——2026-08-28 复查 releases，**仍是最新**，本批无需等新版本 |
| 本批 6 条的来源版本 | 全部属 `v0.1.180`（`c40edb4`）及其之前 |
| 本仓库基线 | `830b1097e`（= 三批 P0 + 决策批全部落地） |
| 本仓库版本号 | `backend/cmd/server/VERSION` = `1.1.9`（自有编号，**不同步成 0.1.183**） |
| 本仓库迁移号 | `224`（本批无新迁移，**不顺延**） |
| 工作分支 | `sync/upstream-20260828` |

取 patch 的通道（不动本仓库 remote 配置，且不受 API 每小时 60 次限制）：

```bash
curl -sSL -o /tmp/p61/<sha>.patch "https://github.com/Wei-Shaw/sub2api/commit/<sha>.patch"
```

## 2. 本批 6 条候选的固定 SHA 与三态证据（第 5 条撤回）

`ok` / `CONFLICT` / `NOFILE` 是按**单文件** patch 逐个 `git apply --check` 的结果（整 commit 的结论
会被测试文件淹没，见 §3）：

```bash
awk '/^diff --git /{n++; f=sprintf("%s/%03d.patch",out,n)} n{print > f}' out=/tmp/split61/<sha> /tmp/p61/<sha>.patch
for p in /tmp/split61/<sha>/*.patch; do git apply --check -p1 "$p"; done
```

| # | SHA | 标题 | 产品文件 | 测试文件 |
|---|---|---|---|---|
| 1 | `cc894ef57` | strip empty streamed tool-call id/name | `openai_gateway_cc_tool_call_identity.go` **新建** `ok`、`openai_gateway_chat_completions_raw.go` `ok` | 2 个 `ok` |
| 2 | `4d4a0be1a` | chat/completions file part → `input_file` | `chatcompletions_to_responses.go` `ok`、`types.go` **`CONFLICT`** | `chatcompletions_responses_test.go` `ok` |
| 3 | `25da02ddd` | avoid duplicate HTTP bridge replay | `openai_tool_continuation.go` `ok`、`openai_ws_forwarder_ingress.go` `ok` | `openai_tool_continuation_test.go` `ok`、`openai_ws_forwarder_ingress_test.go` `ok`、`openai_ws_http_bridge_test.go` **`CONFLICT`** |
| 4 | `66808413d` | drop orphan replay tool calls | `openai_ws_forwarder_payload.go` `ok` | `openai_ws_forwarder_ingress_test.go` `ok`、`openai_ws_http_bridge_test.go` **`CONFLICT`** |
| ~~5~~ | `17c0ee385` | Grok compaction 422 重试 —— **撤回** | `openai_gateway_grok.go` `ok`（1 行），但落地为 no-op ⇒ 不交付 | 无 |
| 6 | `68653fb2c` | allow messages dispatch for composite groups | `openai_messages_dispatch.go` `ok`、`handler/openai_gateway_handler.go` **`CONFLICT`**、`groupsMessagesDispatch.ts` `ok`、`GroupsView.vue` **`CONFLICT`** | `openai_messages_dispatch_test.go` `ok`、`groupsMessagesDispatch.spec.ts` `ok`、`openai_gateway_cn_dispatch_test.go` **`NOFILE`** |

三处 `CONFLICT` 的归属（**都不是「本仓库已经修过」**）：

- 第 2 项 `types.go`：本仓库在 `ResponsesTool` / `ChatTool` 上有自有的 `x_search` 字段，
  上下文偏移。被改的 `ResponsesContentPart` 与 `ChatContentPart` 两个结构体**本身与上游 pre-image 字节一致**。
- 第 3/4 项 `openai_ws_http_bridge_test.go`：上游的锚点测试
  `TestOpenAIWSHTTPBridgeAPIKeyReusesClientToolMappingWhenFollowupOmitsTools` 本仓库不存在，
  新增测试挂不上位置。产品代码不受影响。
- 第 6 项 `openai_gateway_handler.go`：上游 pre-image 里有 `service.IsCNProvider(...)` 分支，
  本仓库无 CN 平台 ⇒ 见 `design.md` 决策 5；`GroupsView.vue` 是本仓库长期自有改动导致的上下文漂移。

`NOFILE` 的 `openai_gateway_cn_dispatch_test.go` 是 CN 专属测试，本仓库无该平台，跳过。

## 3. `apply --check` 的三种假信号（沿用 0.1.180 §2.1，本批再次命中两种）

1. **通过 ≠ 该合。** 本批被排除的 `3e98a5a1`（composite 精确别名路由）6 文件全干净，但它属于上游
   仍在返工的 routed-catalog 功能簇。
2. **失败 ≠ 不该合。** 本批 5 个 `CONFLICT` 里 4 个只是上下文漂移或测试锚点缺失，见 §2。
3. **通过 ≠ 能编译。** `apply --check` 只比上下文，不看新代码引用的符号。本批逐条 grep 过依赖符号：

| 新代码引用 | 本仓库 | 结论 |
|---|---|---|
| `extractOpenAISSEDataLine`（`cc894ef57`） | `openai_sse_json_documents.go:75` 等 4 处调用点，已存在 | 可编译 |
| `applyOllamaCloudRawChatCompletionsSSELine`（插入锚点） | `openai_gateway_chat_completions_raw.go` 已存在 | 锚点在 |
| `AnalyzeToolCallOutputContextCoverageBytes`（`25da02ddd`） | `openai_tool_continuation.go:235` 已存在，本批就地改造 | 可编译 |
| `isCodexToolCallOutputItemType` / `isCodexToolCallContextItemType`（`66808413d`） | 已存在 | 可编译 |
| `cloneOpenAIWSRawMessages`（`66808413d`） | 已存在 | 可编译 |
| `service.IsCNProvider`（`68653fb2c`） | **零命中** | ⇒ 该分支整段丢弃 |
| `grokSameAccountRetryMetadata`（0.1.180 §6.2(a) 曾踩过） | 零命中 | 与本批无关，`17c0ee385` 不引用它 |
| `grokStructuredErrorMessageCandidates`（`953028718` 引入） | **零命中** | ⇒ `17c0ee385` 撤回的直接证据，见下 |

**本批新发现的第四种假信号：通过 ≠ 有效。** `17c0ee385` 的 1 行 `apply --check` 干净、编译通过、
引用的符号全在，但它守卫的下游判定 `isGrokInvalidEncryptedContentResponse` 在本仓库仍硬门 400
⇒ 净行为为零。判定单行守卫类改动时必须连同下游判定一起看。复现命令：

```bash
# v0.1.180 区间里只有两个 commit 含 StatusUnprocessableEntity
for sha in 17c0ee385 953028718; do
  curl -sSL "https://github.com/Wei-Shaw/sub2api/commit/$sha.patch" | grep -c StatusUnprocessableEntity
done
# 上游 v0.1.180 的内层判定已是 400||422，而本仓库仍是 400
curl -sSL "https://raw.githubusercontent.com/Wei-Shaw/sub2api/c40edb4/backend/internal/service/openai_gateway_grok.go" \
  | grep -A2 'func isGrokInvalidEncryptedContentResponse'
```

## 4. 本批**不含**的相邻条目（避免下一轮重复判断）

| 上游 commit | 属簇 | 不做的理由 |
|---|---|---|
| `e2d9ce0ca` + `fbc9ee626` | 0.1.180 §6.1 | 成对且 4 处 `CONFLICT`（`chatcompletions_responses_bridge.go` 与 `openai_gateway_responses_chat_fallback.go` 本仓库都有自有改动），要逐 hunk 解，规模超出本批 |
| `31d5b67ba` | 0.1.180 §6.1 | 5/6 `ok`，但剩下那个是 `openai_gateway_responses_chat_fallback.go`，与上一条同一文件族，一起排期更省事 |
| `7a09a2eaf` | 0.1.180 §6.1 | 动本仓库自有改动过的 `openai_gateway_grok.go` 与 `responses_client_tools.go`（后者刚在 0.1.183 §3.7 改过） |
| `7498d8fdc` | 0.1.180 §6.1 | Responses Lite 强制串行；它是 0.1.183 §5.1 那 5 条的基座，要开就整簇开 |
| `49752060` + `b20f29d1` | 0.1.183 §4.1 | 需 SQLite 重写；且本机 `settings.channel_monitor_mode = 'v1'` ⇒ 监控 v2 未启用，当前零收益 |
| `d5824f6a5` | 0.1.180 §6.3 | 见 `design.md` 决策 6：核心新增只覆盖 CN 模型族，对本仓库空转 |
| 0.1.180 §6.2(c) 全簇（**含 `17c0ee385`**） | Grok 稳定性 | 含两组成对 revert 且与 §7.1 大礼包互相打补丁（`16b15e870`），必须整簇排期；`17c0ee385` 依赖同簇 `953028718` 才有效 |
