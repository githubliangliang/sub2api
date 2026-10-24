## Why

三批 P0 已全部落地（`1b4837098` / `92743740a` / `865683b12`+`f6ed24ade`+`5e8219643`+`a192a5a52`）。
剩余 backlog 里绝大部分是需要整簇排期的大功能。本批筛出 6 条候选（**缺陷已在本仓库核实、
上游 patch site 仍对得上、彼此之间没有文件冲突**），实施中其中 1 条被证实不可独立移植并撤回，
**实际交付 5 条**。其余大簇维持原判不动。

四项在实际使用中是「看不出来」的那一类，正是最值得优先合的：

- **`/v1/chat/completions` 带 PDF 的请求返回 200、模型照常回答，但 prompt 里没有那份文件。**
  `apicompat/types.go` 的 `ChatContentPart` 只有 `text` / `image_url` 两个分支，
  `type:"file"` 在转 Responses 时被静默丢弃，全仓 `input_file` 零命中。没有报错、没有日志，
  只有 `prompt_tokens` 少一截——客户端侧表现为「模型没读附件」。
- **CC 流式 tool_call 的后续 delta 带 `"id":""` / `"function":{"name":""}`。**
  客户端按 `!== undefined` 合并字段时，空串被当成有效值覆盖首包的合法 id/name，
  最终去调一个名为 `""` 的工具（`ToolNotFoundError: unknown tool ""`）。上游举的例子是
  DashScope/DeepSeek，但受害面是**任何**这样报数的 OpenAI 兼容上游。
- **HTTP bridge 在客户端已经自带完整工具历史时仍然整段回放一次。** 判定用的是
  「payload 里有没有 tool call output」这个粗粒度布尔，没有检查上下文是否已经覆盖所有 `call_id`，
  于是同一段历史被拼进请求两次；且 `input` 是单个对象（非数组）时判定直接落空。
- **composite 分组的 `/v1/messages` 开关是死的。** `sanitizeGroupMessagesDispatchFields`
  对一切非 openai 平台无条件 `AllowMessagesDispatch = false`，管理台里改了也存不住；
  handler 侧则用「解析到 grok/CN 目标就无条件豁免」来绕开，两边合起来是「开关无效 + 豁免过宽」。

第 6 个候选（`17c0ee385`，Grok compaction 422 同号重试）在实施阶段被撤回：它 `apply --check`
干净且只有 1 行，但本仓库的内层判定 `isGrokInvalidEncryptedContentResponse` 仍硬门 `400`，
补了外层守卫也不会有任何 422 真正走进重试。widen 内层判定的是 §6.2(c) 的 `953028718`
（同时引入 compaction 错误码与 `grokStructuredErrorMessageCandidates`，本仓库零命中）
⇒ 这条实质是 §6.2(c) 的尾巴，随该簇整体排期。见 `design.md` 决策 7。

本变更**只做移植**，不跟进任何功能簇。不做项及理由见 `design.md` 第 4 节。

## What Changes

- CC 流式直转路径（raw chat completions SSE）在下发前剔除 `choices[*].delta.tool_calls[*]` 上
  **存在但为空串**的 `id` 与 `function.name`；非空值、`arguments`（含空串）、`index`、`type` 一律不动。
- `/v1/chat/completions` 的 `type:"file"` content part 转换为 Responses 的 `input_file`
  （`filename` / `file_data` / `file_id` 透传）；既无 `file_data` 又无 `file_id` 的空 file part
  跳过，与现有空 image URL 的处理一致。
- HTTP bridge 的回放判定从「payload 里有 tool call output」改成
  「有 tool call output **且**上下文没有覆盖全部 `call_id`」；工具输出覆盖度分析同时支持
  `input` 为数组和为单个对象两种形态。
- 回放历史时丢弃**没有配对输出**的孤儿 tool call context item（前一轮与本轮的输出 `call_id`
  合并成配对集合后过滤）。
- `sanitizeGroupMessagesDispatchFields` 对 `composite` 平台不再强制关闭 `AllowMessagesDispatch`；
  handler 侧对 composite 的豁免收窄为「仅当解析到的目标平台是 grok 时」，解析到 openai 目标时
  改由分组自己的开关控制；管理台的分组编辑表单对 composite 平台显示该开关。
- **不新增迁移、不改 schema、不动 wire 图、不新增配置项、不改 `VERSION`。**

## Capabilities

### New Capabilities

- `cc-stream-tool-call-identity`: CC 流式 tool_call 增量里身份字段（`id` / `function.name`）的空串语义。
- `chat-completions-file-input`: Chat Completions → Responses 转换层对文件类 content part 的处理契约。
- `ws-bridge-replay-integrity`: HTTP bridge 决定是否回放历史、以及回放内容必须满足的配对不变量。
- `composite-messages-dispatch`: composite 分组对 `/v1/messages` 的准入由谁决定。

### Modified Capabilities

无。`openspec/` 下当前没有已发布的 capability 基线（仅有 `changes/`），因此本变更以 ADDED
Requirements 形式固化「移植后应当成立的行为」。

## Impact

- **后端**：新增 `internal/service/openai_gateway_cc_tool_call_identity.go`（+ 单测）；
  修改 `internal/service/` 的 `openai_gateway_chat_completions_raw.go`、`openai_tool_continuation.go`、
  `openai_ws_forwarder_ingress.go`、`openai_ws_forwarder_payload.go`、
  `openai_messages_dispatch.go`；`internal/handler/openai_gateway_handler.go`；
  `internal/pkg/apicompat/` 的 `types.go`、`chatcompletions_to_responses.go`。
- **前端**：`views/admin/groupsMessagesDispatch.ts` 新增 `supportsMessagesDispatchPlatform`，
  `views/admin/GroupsView.vue` 的创建/编辑表单两处按该函数放行 composite。
- **数据库**：无。**迁移号仍是 `224`**，本轮不得顺延。
- **配置**：无新增配置项，无默认值变化。
- **计费**：无直接改动。间接影响是 PDF 附件从此真的进入 prompt ⇒ 这类请求的 `prompt_tokens`
  与费用会**上升**（此前是「附件被丢掉的便宜价」）。这是修正，不是回退。
- **调度 / 准入**：composite 分组的 `/v1/messages` 准入语义变化——此前「解析到 grok 目标恒放行、
  解析到 openai 目标恒 403」，之后「grok 目标仍恒放行、openai 目标由分组开关决定」。
  **存量 composite 分组的 `allow_messages_dispatch` 落库值此前恒被写成 false**，
  因此升级后行为不会自动放宽，需要管理员显式打开，见 `design.md` 决策 5。
- **兼容性**：无对外 breaking change。剥空串 id/name 只删「存在且为空」的字段，
  合法值与缺失字段两种形态都不受影响。
- **风险面**：第 3/4 项动的是 WS↔HTTP bridge 的请求装配，是本批唯一可能改变**发给上游的请求体**
  的改动（少发重复历史 / 少发孤儿 item），验收要求见 `verification.md` §3。

## Execution References

- `source-baseline.md`：上游 commit 固定、按文件 apply 三态证据、依赖符号核查。
- `design.md`：移植策略、逐项决策、不做项的理由、阶段划分与回滚。
- `docs/upstream-sync/PORTING-0.1.180.md` §6：逐条 patch site 与上游 diff 摘要。
- `docs/upstream-sync/README.md` §4 硬约束：本仓库的移植护栏。
