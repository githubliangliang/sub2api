## 1. 固定基线与边界

- [x] 1.1 起点 `830b1097e`，工作分支 `sync/upstream-20260828`；`cd backend && go build ./...` 通过
- [x] 1.2 记录基线信号：unit EXIT=1（`TestAliyunCaptchaVerifier_TransportError` 既有 flake，单跑该包两次均 ok）；
      前端 `useRoutePrefetch.spec.ts` 5 条既有失败。详见 `verification.md` §0
- [x] 1.3 确认 `backend/migrations/` 最大序号仍为 `224`，全程 MUST NOT 新增迁移
- [x] 1.4 确认 `backend/cmd/server/VERSION` 保持 `1.1.9`，全程不改
- [x] 1.5 ⚠️ 全程 MUST NOT 改 `frontend/package.json` / `pnpm-lock.yaml` / `pnpm-workspace.yaml`（dompurify 仍推迟）
- [x] 1.6 全程 MUST NOT 动 `wire.go` / `wire_gen.go`（本批不涉及 DI）

## 2. 阶段 1 — `cc894ef57` CC 流式 tool_call 身份字段

- [x] 2.1 新建 `backend/internal/service/openai_gateway_cc_tool_call_identity.go`：
      `stripEmptyChatToolCallIdentityFromSSELine` + `stripEmptyChatToolCallIdentity`
- [x] 2.2 `openai_gateway_chat_completions_raw.go` 的 `streamRawChatCompletions`：
      在 `applyOllamaCloudRawChatCompletionsSSELine(account, line)` 之后插入
      `line = stripEmptyChatToolCallIdentityFromSSELine(line)`
- [x] 2.3 落地上游单测 `openai_gateway_cc_tool_call_identity_test.go`（`//go:build unit`）
- [x] 2.4 落地上游端到端测试 `TestForwardAsRawChatCompletions_StripsEmptyToolCallIdentity`
- [x] 2.5 `go build ./... && go test -tags=unit ./internal/service/ -run 'ToolCallIdentity' -count=1`
- [x] 2.6 独立提交

## 3. 阶段 2 — `4d4a0be1a` file part → `input_file`

- [x] 3.1 `apicompat/types.go`：`ResponsesContentPart` 增加 `Filename` / `FileData` / `FileID`，
      注释里的 type 枚举补 `input_file`
- [x] 3.2 `apicompat/types.go`：`ChatContentPart` 增加 `File *ChatFile`，type 注释补 `file`
- [x] 3.3 `apicompat/types.go`：新增 `ChatFile` 结构体（`filename` / `file_data` / `file_id`）
- [x] 3.4 ⚠️ 三处均手写，MUST NOT 用模糊匹配硬塞 hunk（本仓库 `x_search` 字段会错位，见 `design.md` 决策 1）
- [x] 3.5 `apicompat/chatcompletions_to_responses.go`：`convertChatContentPartsToResponses` 增加
      `case "file"`，仅在 `FileData != "" || FileID != ""` 时产出 `input_file`
- [x] 3.6 落地上游三个测试（`FilePartFileData` / `FilePartFileID` / `EmptyFilePartSkipped`）
- [x] 3.7 `go test -tags=unit ./internal/pkg/apicompat/ -run 'FilePart|FileID|EmptyFile' -count=1`
- [x] 3.8 独立提交

## 4. 阶段 3 — `25da02ddd` → `66808413d` 回放完整性（**顺序固定**）

### 4.1 `25da02ddd`（必须先）

- [x] 4.1.1 `openai_tool_continuation.go` 的 `AnalyzeToolCallOutputContextCoverageBytes`：
      守卫放宽成 `!input.IsArray() && !input.IsObject()`
- [x] 4.1.2 同函数：把 `ForEach` 闭包抽成 `analyzeItem(item gjson.Result)`，
      内部的 `return true` 改成 `return`
- [x] 4.1.3 同函数：数组走 `ForEach`，对象直接 `analyzeItem(input)`
- [x] 4.1.4 `openai_ws_forwarder_ingress.go` 的 `needsBridgeReplay`：改成
      `previousResponseID != "" || (coverage.HasFunctionCallOutput && !coverage.ContextCoversAllCallIDs)`
- [x] 4.1.5 落地上游测试改动（`openai_tool_continuation_test.go`、`openai_ws_forwarder_ingress_test.go`）

### 4.2 `66808413d`（在 4.1 之后）

- [x] 4.2.1 `openai_ws_forwarder_payload.go`：新增 `sanitizeOpenAIWSHistoricalReplayToolCalls`
- [x] 4.2.2 `buildOpenAIWSReplayInputSequence`：在 `previousFullInputExists` 之后、
      `currentExists` 判断之前调用一次
- [x] 4.2.3 落地上游测试改动（`openai_ws_forwarder_ingress_test.go`）

### 4.3 两条共同的测试落地

- [x] 4.3.1 ⚠️ 上游锚点测试本仓库不存在 ⇒ 把两个端到端测试**追加到
      `openai_ws_http_bridge_test.go` 末尾**，直接写 `66808413d` 之后的形态（见 `design.md` 决策 2）
- [x] 4.3.2 `go test -tags=unit ./internal/service/ -run 'OpenAIWSHTTPBridge|ToolCallOutputContext' -count=1`
- [x] 4.3.3 两条一起提交（单独 revert 会留下中间态）

## 5. 阶段 4 — `17c0ee385` Grok compaction 422 —— **撤回，不交付**

- [x] 5.1 ~~改 `forwardGrokResponses` 的外层守卫~~ ⇒ **实测为 no-op，已回退**：
      内层 `isGrokInvalidEncryptedContentResponse`（`openai_gateway_grok.go:255`）仍硬门 400，
      422 进外层也会被内层否掉
- [x] 5.2 定位 widen 内层的上游 commit：§6.2(c) 的 `953028718`（含 compaction 错误码与
      `grokStructuredErrorMessageCandidates`，本仓库零命中）
- [x] 5.3 结论记入 `design.md` 决策 7，并随 §6.2(c) 整簇排期；`openai_gateway_grok.go` 保持不动

## 6. 阶段 5 — `68653fb2c` composite `/v1/messages`

- [x] 6.1 `service/openai_messages_dispatch.go` 的 `sanitizeGroupMessagesDispatchFields`：
      `AllowMessagesDispatch = false` 收窄成 `if g.Platform != PlatformComposite`
- [x] 6.2 `handler/openai_gateway_handler.go` 的 `allowOpenAICompatibleMessagesDispatch`：
      给「解析到 grok 目标」的豁免补上 `apiKey.Group.Platform == service.PlatformComposite` 前置条件，
      并更新注释说明「解析到 openai 目标由分组开关控制」
- [x] 6.3 ⚠️ MUST NOT 引入 `service.IsCNProvider` 分支（本仓库无 CN 平台，见 `design.md` 决策 5）
- [x] 6.4 ⚠️ MUST NOT 为对齐上游而把 `allowOpenAICompatibleMessagesDispatch` 的
      `ctx context.Context` 签名改成 `*gin.Context`
- [x] 6.5 `frontend/src/views/admin/groupsMessagesDispatch.ts`：新增
      `supportsMessagesDispatchPlatform(platform)`（openai / composite）
- [x] 6.6 `frontend/src/views/admin/GroupsView.vue`：创建与编辑表单各一处区块 `v-if` 改用该函数；
      逐模型映射区块的 `v-if` 追加 `platform === 'openai'`；两处平台 watch 改用该函数；补 import
- [x] 6.7 落地上游后端测试改动（`openai_messages_dispatch_test.go`）
- [x] 6.8 落地上游前端测试改动（`views/admin/__tests__/groupsMessagesDispatch.spec.ts`）
- [x] 6.9 ⚠️ `handler/openai_gateway_cn_dispatch_test.go` 是 CN 专属（`NOFILE`），跳过
- [x] 6.10 独立提交

## 7. 收尾

- [x] 7.1 `cd backend && go build ./... && go test -tags=unit ./... -count=1 && golangci-lint run ./...`
- [x] 7.2 `cd frontend && pnpm run typecheck && pnpm run lint:check`
- [x] 7.3 `make test-frontend-critical`
- [x] 7.4 填 `verification.md` 的证据矩阵（含三条「必须先复现失效再验修复」的项）
- [x] 7.5 `docs/upstream-sync/PORTING-0.1.180.md`：§6.1 四行 / §6.3 `68653fb2c` 标「已合」；
      `d5824f6a5` 标 N/A + 理由（决策 6）；§6.2(c) 补记 `17c0ee385` 不可独立移植（决策 7）
- [x] 7.6 `docs/upstream-sync/README.md` 顶部的「当前待移植清单」段落补一句本批的落点
- [x] 7.7 合回 `main` —— `d775cc12d`（`--no-ff`，2026-08-28）。合并后在 `main` 上复跑门禁：
      `go build` 通过 / `go test -tags=unit ./...` **EXIT=0** / `golangci-lint run ./...` **0 issues**
