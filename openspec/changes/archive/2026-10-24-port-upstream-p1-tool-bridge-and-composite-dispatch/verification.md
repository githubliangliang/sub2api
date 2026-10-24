# 验收证据

执行日期：2026-08-28。分支 `sync/upstream-20260828`，基线 `830b1097e`。

## 0. 基线信号（改动前，`830b1097e` 干净工作区）

| 项 | 结果 |
|---|---|
| `go build ./...` | 通过 |
| `go test -tags=unit ./... -count=1` | **EXIT=1**，1 个失败：`internal/repository` 的 `TestAliyunCaptchaVerifier_TransportError` |
| 该失败的性质 | **既有 flake，与本批无关**。单独跑 `go test -tags=unit ./internal/repository/ -count=1` 连续两次均 `ok`。该测试起一个 `httptest` server 后立刻 `Close()` 制造连接失败；全量并发下端口被其它测试的 server 复用，连接反而成功并返回 JSON 错误 ⇒ 被归一成 API error，断言失败 |
| `pnpm exec vitest run src/composables/__tests__/useRoutePrefetch.spec.ts` | **5 failed / 10 passed**，与 PORTING-0.1.179.md §4.6 记录一致 |

⚠️ 上面两项在**改动后必须仍是同一形态**（不能新增、也不能「恰好变绿」而掩盖别的问题）。

## 1. 交付后的全量信号

| 项 | 结果 | 与基线对比 |
|---|---|---|
| `go build ./...` | `BUILD_OK` | 一致 |
| `go test -tags=unit ./... -count=1` | **EXIT=0，0 个 FAIL** | 基线那条 flake 本次未复现（符合 flake 定性） |
| `golangci-lint run ./...` | **`0 issues.`**（v2.13） | 一致 |
| `pnpm run typecheck` | 通过，无输出 | 一致 |
| `pnpm run lint:check` | 通过，无输出 | 一致 |
| `make test-frontend-critical` | **14 files / 167 passed / 2 skipped** | 一致 |
| `pnpm run test:run`（全量） | 229 passed / **1 failed**（`useRoutePrefetch.spec.ts`，5 条） | **与基线逐条相同** |

`useRoutePrefetch.spec.ts` 的 5 条失败已复核为与本批无关：切回 `main`（`830b1097e`）单跑同一文件，
结果同样是 `5 failed | 10 passed`。

## 2. 逐项证据

### 2.1 `cc894ef57` CC 流式 tool_call 身份字段 — commit `2f091ed66`

| 检查 | 证据 |
|---|---|
| 落地方式 | 4/4 文件逐字 `git apply`，无手改 |
| 插入点正确 | `openai_gateway_chat_completions_raw.go` 的 SSE 行循环里，`applyOllamaCloudRawChatCompletionsSSELine` 之后、`writeLine(line)` 之前 |
| 单测 | `TestStripEmptyChatToolCallIdentity_*` 全部 PASS（首包不动 / 后续 delta 剥空 / 只空一个 / 空 arguments 保留 / 多 tool_calls / 6 个 passthrough 子用例 / SSE 行 4 个 passthrough 子用例） |
| 端到端 | `TestForwardAsRawChatCompletions_StripsEmptyToolCallIdentity` PASS |

### 2.2 `4d4a0be1a` file part → `input_file` — commit `ddc63c7ef`

| 检查 | 证据 |
|---|---|
| 落地方式 | `chatcompletions_to_responses.go` + 测试逐字 apply；`types.go` 三处**手写**（决策 1） |
| 手写正确性 | `go build` 通过；`ResponsesContentPart` / `ChatContentPart` / 新增 `ChatFile` 的字段与 tag 与上游一致；相邻的 `x_search` 字段未被扰动 |
| 单测 | `TestChatCompletionsToResponses_FilePartFileData` / `_FilePartFileID` / `_EmptyFilePartSkipped` 全部 PASS |
| 回归 | `go test -tags=unit ./internal/pkg/apicompat/ -count=1` 全绿（新字段均 `omitempty`，既有序列化路径无变化） |

### 2.3 `25da02ddd` + `66808413d` 回放完整性 — commit `a3f0c978b`

⚠️ **这两条属于「不抛错」的行为修复，按「代码改了」验收等于没验收。** 因此逐条做了**反向验证**：
把产品改动单独回退，确认新测试确实变红。

| 反向验证 | 命令 | 结果 |
|---|---|---|
| 单独回退 `25da02ddd` 的 `needsBridgeReplay` 判据 | `-run 'OpenAIWSHTTPBridgeFullCustomToolHistory\|OpenAIWSHTTPBridgeObjectToolOutput'` | **`--- FAIL: ...FullCustomToolHistoryWithoutPreviousResponseIDDoesNotReplay`** |
| 单独回退 `66808413d` 的 `sanitizeOpenAIWSHistoricalReplayToolCalls` 调用 | 同上 | **`--- FAIL: ...FullCustomToolHistoryWithoutPreviousResponseIDDoesNotReplay`** |
| 单独回退 `AnalyzeToolCallOutputContextCoverageBytes` 的 `IsObject` 放宽 | `-run 'TestAnalyzeToolCallOutputContextCoverageBytes'` | **`--- FAIL: .../object_tool_output_requires_context_replay`** |
| 三者全部恢复 | `go build ./...` + 上述两组 | 全部 `ok` |

⇒ 三处产品改动**各自都是 load-bearing**，不存在「测试恰好通过」的情况。

| 其它检查 | 证据 |
|---|---|
| 顺序约束 | `25da02ddd` 先落地，`66808413d` 后落地，同一 commit 交付 |
| 测试落点 | 两个端到端测试以 `66808413d` **之后**的形态追加到 `openai_ws_http_bridge_test.go` 末尾（上游锚点 `TestOpenAIWSHTTPBridgeAPIKeyReusesClientToolMappingWhenFollowupOmitsTools` 本仓库不存在） |
| 既有回放测试未回归 | `-run 'ToolCallOutputContext\|OrphanToolOutput'` 全绿，含 `TestBuildOpenAIWSCurrentTurnRetryPayloadRejectsOrphanToolOutput` |

### 2.4 `68653fb2c` composite `/v1/messages` — commit `b21a2df03`

| 检查 | 证据 |
|---|---|
| 反向验证 | 把 `sanitizeGroupMessagesDispatchFields` 的 composite 例外去掉后，`TestSanitizeGroupMessagesDispatchFields_PreservesCompositeDispatchToggle` **变红**；恢复后 PASS |
| 附属字段仍被清空 | 同一测试断言 `DefaultMappedModel` 为空、`MessagesDispatchModelConfig` 为零值 |
| 其它平台未被放宽 | `TestSanitizeGroupMessagesDispatchFields_ClearsNonOpenAIPlatform` 仍 PASS |
| CN 分支未被引入 | `grep -rn 'IsCNProvider' backend/` 仍**零命中** |
| 签名未被改动 | `allowOpenAICompatibleMessagesDispatch(ctx context.Context, ...)`，调用方未改 |
| 豁免已收窄 | 「解析到 grok 目标」的分支现在包在 `apiKey.Group.Platform == service.PlatformComposite` 内 |
| 收窄是行为中性的 | `ensureCompositeTargetPlatform`（`handler/composite_platform.go:13`）只在 composite 分组上写该 context 值；另一处写入点 `server/routes/gateway.go:649` 写的是 `gemini` ⇒ 今天不存在「非 composite 分组解析出 grok」的路径 |
| 前端 | `groupsMessagesDispatch.spec.ts` **5 passed**（含新增的 `supports OpenAI and composite groups`）；`typecheck` / `lint:check` 通过 |
| composite 相关既有测试 | `-run 'Composite'` 全绿（10 项） |

### 2.5 `17c0ee385` — **撤回，无交付物**

| 检查 | 证据 |
|---|---|
| 结论 | 单独合是 no-op，见 `design.md` 决策 7 |
| 证据 1 | 本仓库 `openai_gateway_grok.go:255` 的 `isGrokInvalidEncryptedContentResponse` 首行仍是 `if statusCode != http.StatusBadRequest { return false }` |
| 证据 2 | 上游 `c40edb4` 同一函数已是 `!= 400 && != 422`，且识别 `invalid_compaction` / `compaction_decode_error` 与「decode the compaction blob」 |
| 证据 3 | `grokStructuredErrorMessageCandidates` 本仓库 **零命中** |
| 证据 4 | `v0.1.180` 区间里仅 `17c0ee385` 与 `953028718` 两个 commit 含 `StatusUnprocessableEntity` ⇒ widen 内层的是后者（属 §6.2(c)） |
| 工作区状态 | `openai_gateway_grok.go` **未改动**（`git status` 干净，`grep StatusUnprocessableEntity` 零命中） |

## 3. 影响面复核

| 面 | 检查 | 结果 |
|---|---|---|
| 迁移 | `ls backend/migrations/ \| sort -V \| tail -1` | 仍是 `224`，未新增 |
| 版本号 | `cat backend/cmd/server/VERSION` | 仍是 `1.1.9`，未改 |
| wire | `git diff --name-only main...HEAD \| grep wire` | 无命中 |
| 前端依赖 | `git diff --name-only main...HEAD \| grep -E 'package.json\|pnpm-lock\|pnpm-workspace'` | 无命中（dompurify 推迟决定未被破坏） |
| 发给上游的请求体 | 阶段 3 会**减少**重复历史与孤儿 item；已由 2.3 的两个端到端测试逐条断言 `input` 的条目数与类型 |
| 计费 | 无直接改动；PDF 附件从此真的进 prompt ⇒ 该类请求 `prompt_tokens` 与费用**上升**（修正，非回退） |
| 准入 | composite 的 `/v1/messages`：grok 目标仍恒放行；openai 目标从「恒 403」变成「由分组开关决定」，存量落库值均为 `false` ⇒ 升级后无自动放宽 |

## 4. 上线后观察

- **PDF 场景**：确认带 `type:"file"` 的 `/v1/chat/completions` 请求 `prompt_tokens` 明显上升
  （此前是「附件被丢掉」的偏低值）。若某模型/账号对 `input_file` 报 400，是上游能力问题，
  不是本次转换错误——先看该账号是否支持文件输入。
- **WS bridge**：盯 Codex 场景的上游 `input` 长度与 4xx 比例。回放变少是预期，
  若出现「上游说找不到某个 `call_id`」，说明覆盖度判定过于乐观，回退阶段 3 那一个 commit 即可。
- **composite 分组**：开关现在是活的，但默认关。管理员打开后要确认 `/v1/messages` 解析到
  openai 目标的请求确实按预期放行。
