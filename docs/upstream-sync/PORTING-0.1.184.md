# 移植清单：上游 v0.1.184

对照日期：2026-08-31。基线 `b72e6eb73`（= 本仓库 `1.1.10`，含 0.1.183 P0 全 12 项 +
0.1.180 已推进的那几批）。

上游 v0.1.184 是**大混合版**：99 个非 merge commit、342 文件、+20205-1325，含 3 条新迁移
（全是 PG 方言，且上游自己撞号，三条都叫 `231_*`）。本文只覆盖
`v0.1.183 (e8cb019f) .. v0.1.184 (e98ef32e)`。

移植前先读 README [第 4 节「硬约束」](./README.md#4-硬约束)，尤其第 13 条；写 SQL 对照
[第 5 节转换速查](./README.md#5-pg--sqlite-转换速查)。

**结论：14 项 P0 可以吃下（缺陷已核实、基座符号在位），11 项 P1 按需，
Codex routed catalog / service_tier / Spark 429 / WS v2 / usage 两个新字段五簇本轮不做。**

---

## 1. 版本对照

| 点 | 状态 |
|---|---|
| 上游最新正式版 | [`v0.1.184`](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.1.184)（tag commit `e98ef32eb`，2026-08-31 16:43 +0800） |
| 上一轮对照基线 | `v0.1.183`（tag commit `e8cb019fa`，2026-08-25 21:32 +0800） |
| `v0.1.183` → `v0.1.184` | **99 commits（非 merge）** / 342 文件 / +20205-1325 |
| 上游 main | `200602b41`，v0.1.184 之后仅 3 条（`e2624fb65` Codex 图像能力保留 + sponsors + VERSION 同步），本文不含 |
| 本仓库基线 | `b72e6eb73`，`backend/cmd/server/VERSION` = `1.1.10`（自有编号，**不要**同步成 0.1.184） |
| 本仓库迁移号 | 仍是 `224`。上游本轮 3 条迁移**全部不合**（见 §5.5），故仍不顺延 |
| 本仓库平台面 | 只有 anthropic / openai / gemini / antigravity / grok 五个（`internal/domain/constants.go:21-25`）。Kimi / Zhipu / DeepSeek / Ollama **不是平台**，DeepSeek 仅作为 openai_compat 上游出现 |

对比链接：<https://github.com/Wei-Shaw/sub2api/compare/v0.1.183...v0.1.184>

### 1.1 按簇统计

| 簇 | commits | 判断 |
|---|---|---|
| 可移植的独立 bugfix | 14 | **P0 全收**（§3） |
| 值得但按需 / 成本中等 | 11 | **P1**（§4） |
| Codex routed model catalog | 11 | 已进正式 tag，但冲突面大 → 单独立项（§5.1） |
| Fast mode `service_tier` 一套 | 8 | 缺 0.1.180 §7.2 基座 → 不做（§5.2） |
| Spark / OpenAI 配额 429 | 4 | 本仓库该缺陷不成立，同 0.1.183 §5.2 → 不做（§5.3） |
| WS v2 passthrough | 5 | 缺 0.1.180 §7.1 基座 → 不做（§5.4） |
| usage 两个新字段（+2 迁移） | 3 | 功能非缺陷，要转 SQLite 迁移 → 推迟（§5.5） |
| 每用户限制公共分组（+1 迁移） | 1 | 要动 ent schema + 迁移 + aux 表，单人收益低 → 推迟（§5.6） |
| Antigravity 混合内置工具 | 1 | 缺 0.1.178 `hasMixedToolInvocations` 基座 → 可拆半条（§5.7） |
| Grok 无效工具 union / vision | 4 | 属 0.1.180 §6.2(c) 挂起的 Grok 整簇 → 不做（§5.8） |
| 国产 / Ollama / Zhipu | 2 | **N/A**（§6） |
| chore（sponsors / VERSION / 误放组件） | 5 | 不合（§6） |
| 已合 | 1 | `4795650d2` = 本仓库 `f6ed24ade`（§6） |

---

## 2. 本轮核查方法

本轮不再用部分裸克隆——`upstream` remote 已在本地，直接 `git fetch upstream --tags` 即可
（HTTPS 易 TLS 中断，用 SSH，见 README §1）。

```bash
git fetch upstream --tags
git rev-list --no-merges --reverse v0.1.183..v0.1.184 > /tmp/u184/shas.txt
```

三态判定沿用 0.1.183 §2 的**按文件切开**，但改成直接用本仓库工作树，不落中间 patch 文件：

```bash
while read sha; do
  for f in $(git show --pretty=format: --name-only $sha | grep -v '^$' | sort -u); do
    [ -e "$f" ] || { echo "NOFILE $f"; continue; }
    git diff $sha^ $sha -- "$f" | git apply --check - 2>/dev/null \
      && echo "ok   $f" || echo "CONF $f"
  done
done < /tmp/u184/shas.txt
```

再加一道「本仓库是否已引用过这个上游 SHA」的反查，避免重复移植：

```bash
git log --format='%H%n%B' main > /tmp/u184/ourlog.txt
cat docs/upstream-sync/*.md openspec/changes/*/*.md > /tmp/u184/ourdocs.txt
# 逐条 grep 短 SHA；本轮命中 16 条，其中 15 条是 §5.1 那簇的「已记录未合」，
# 1 条（4795650d2）确实已合
```

⚠️ 0.1.183 §2 列的三种假信号本轮全部复现，并且**第三种（`ok` 不保证能编译）在本轮抓到两条**，
都是靠「对新增代码里的每个函数调用逐个 grep 基座」抓出来的，不是靠 `apply --check`：

- `b1737cc84` 5/6 文件干净，但它改的 `hasMixedToolInvocations` 与
  `includeServerSideToolInvocations` 本仓库**零命中**（0.1.178 那条没移过）→ 降级到 §5.7。
- `b7ec3cdad` 目标文件在，但它重写的 return 里带 `UpstreamResponseServiceTier` /
  `resolvedOpenAIUpstreamServiceTier`，两个符号本仓库零命中（属 §5.2 未合的 service_tier
  基座）→ 保留在 P0，但必须按本地 return 形态改写，见 3.12。

反过来，**第二种假信号（`CONFLICT` 不等于不该合）**本轮的典型是 `c03776604`：产品文件
`UseKeyModal.vue` 干净，只有 spec 冲突。

---

## 3. P0 — 已核实本仓库有同一缺陷

按建议顺序排列（先小后大、先无冲突后需手改）。每条给出**上游 patch site → 本仓库对应位置**，
以及基座符号核查结论。状态一律先写「待合」，合完改成「已合」。

### 3.1 `8f5451587` Anthropic→Responses 流式转换：thinking 前不关 item、content_index 从不推进

2 文件 +234-2，**全部干净**（`backend/internal/pkg/apicompat/anthropic_to_responses_response.go`
+ 新增 stream 用例）。状态：**已合**

上游 PR #6414，是 v0.1.184 的最后一条提交。两个独立缺陷都在本仓库
`anthropic_to_responses_response.go` 里逐字成立：

1. **thinking 分支不关旧 item**（本仓库 `:279` `case "thinking":`，与上游 `:283` 同形）。
   同 switch 的 `tool_use` 分支本来就先调 `closeCurrentResponsesItem`，只有 thinking 漏了。
   一个 message item 在它的 text 块 `content_block_stop` 时是**刻意保持打开**的，所以
   thinking 块到来时它还开着 → 直接被 `CurrentItemType` / `CurrentItemID` 覆盖，
   累积在 `CurrentContent` 里的助手文本既不进 `state.Outputs` 也拿不到 `output_item.done`，
   `response.completed` 只带 reasoning。**客户端看到的是「HTTP 200 成功但零输出」。**
   交错思考（interleaved-thinking，本仓库在 anthropic-beta 透传里明确支持）稳定产生
   `text → thinking` 这个顺序。
2. **`ContentIndex` 只在开 item 时置 0，从不推进**。同一 message item 里第二个 text 块会
   再发一次 `content_part.added(content_index=0)`，与第一个 part 撞在同一下标；SDK 的累积式
   stream helper 按 `content[content_index]` 写入，后者覆盖前者 → **可见文本丢失**。
   反方向 `responses_to_anthropic.go` 有三处 `ContentBlockIndex++`，正方向一处都没有。

修法：thinking 与 text 两个分支各补一行 `events = append(events, closeCurrentResponsesItem(state)...)`；
`content_block_stop` 的 text 分支里先存 `contentIndex := state.ContentIndex`，两个终止事件
用它，之后 `state.ContentIndex++`。

**本 fork 优先级最高的一类缺陷**：Claude 号喂 Responses / Codex 客户端正是主用途。

### 3.2 `da10822d7` Anthropic `tool_use.input` 被 `{}` 占位符污染

4 文件 +171-1，产品文件 `gateway_forward_as_responses.go` 干净（3/4 ok，缺的是一个新测试文件）。
状态：**已合**（实施偏差见 §9.1）

Anthropic 在 `content_block_start` 里把 `tool_use.input` 初始化成 `{}`，真正的入参走
`input_json_delta` 增量。本仓库 `gateway_forward_as_responses.go:610` 的 `appendRawJSON`
只判 `len(existing) == 0`，于是拼出 `{}{"a":1}` —— **非法 JSON，工具调用直接坏掉**。

修法 6 行：`json.Unmarshal` 出 `map[string]json.RawMessage` 且长度为 0 时视作占位符，
与 `len(existing) == 0` 同等处理。

### 3.3 `9f1effd71` 分组部分更新会把日/周/月限额清成 0

4 文件 +81-7，产品文件 2/2 干净（3/4 ok）。状态：**已合**

两处缺陷本仓库逐字存在：

| 位置 | 现状 | 问题 |
|---|---|---|
| `handler/admin/group_handler.go:84` | `zero := 0.0; return &zero` | 字段缺省时回 0 |
| `service/admin_group.go:700-702` | `group.DailyLimitUSD = normalizeLimit(input.DailyLimitUSD)` 三行无条件覆盖 | nil 也照写 |

限额语义是「nil/负数 = 无限制，0 = **不允许任何用量**，正数 = 具体限额」。所以一次不带限额
字段的部分更新会把分组的三个限额从「无限制」写成「0 = 拒绝所有用量」——**分组直接不可用**。

修法：handler 侧 `zero := 0.0` 改成 `unlimited := -1.0`；service 侧三行各加 `!= nil` 守卫。
本仓库注释「前端始终发送这三个字段，无需 nil 守卫」要一并删掉，它正是这个 bug 的成因。

⚠️ **本仓库风险比上游大**：`skills/sub2api-admin`（`node scripts/sub2api-admin.js …`）就是拿来做
分组批量/部分更新的，CLAUDE.md 还明确推荐它优先于 ad-hoc curl。合这条之前，用该 skill 改过的
分组都值得回查一遍限额值。

### 3.4 `32ac921f2` Fable OAuth 请求被系统提示词形态判为 refusal（空响应）

2 文件 +38，**全部干净**。基座 `isAnthropicFableModel` 在位。状态：**已合（含原生路径扩写，见 §9.2）**

上游给 `applyClaudeCodeOAuthMimicryToBody`（本仓库 `gateway_claude_oauth_body.go:391`）加一行
`systemPromptBlocks = claudeOAuthSystemPromptBlocksForModel(model, systemPromptBlocks)`，
并新增常量 `claudeFableOAuthSystemPromptBlocks`——只保留 `{billing_header}` +
`{claude_code_system_prompt}` 两块，去掉通用 CLI 展开块。Fable 5 对那个展开块会回
`stop_reason=refusal` + 零 output token。

⚠️ **照抄会漏掉本 fork 最热的那条路。** `applyClaudeCodeOAuthMimicryToBody` **只被桥接路径调用**：

```
gateway_forward_as_chat_completions.go:102   ← Anthropic OAuth 号喂 Chat Completions 客户端
gateway_forward_as_responses.go:107          ← Anthropic OAuth 号喂 Responses / Codex 客户端
```

原生 `/v1/messages` 的注入点是**另一处**：`gateway_forward.go:195`（上游对应 `:206`），
上游那条 patch 没碰它——上游自己也是这样。本 fork 的主用法恰恰是
Claude Code → 原生 `/v1/messages` → Anthropic OAuth 号，照抄之后 Fable 仍会被注 CLI 展开块。

**移植时要在 `gateway_forward.go:197` 一并套上 `claudeOAuthSystemPromptBlocksForModel(reqModel, …)`**，
并补一条钉住原生路径的用例。这属 README §4 第 13 条那类「守卫单行改动空转」。

### 3.5 `ea4291a92` Fable 7d_oi 调度阈值会暂停整个 Anthropic 账号

5 文件 +203，**全部干净**。状态：**已合**

本仓库 `ratelimit_service.go:303-308` 已经按「7d_oi 是 Fable 专属窗口 → 只标记模型级限流」
处理**限流头**，但 `ApplyAccountSchedulingThreshold`（`ratelimit_service.go:149`）走的是另一套
**调度阈值**路径，那条仍然整号 `TempUnschedulable` —— 一个 Fable 窗口打到阈值，
账号对 Sonnet / Haiku 也不再被调度。

上游新增三块：`account_scheduling_threshold_eval.go` 里
`evaluateAnthropicFableSchedulingThreshold` + `anthropicFableThresholdCandidate`（读
`passive_usage_7d_oi_utilization` / `passive_usage_7d_oi_reset`）、`model_rate_limit.go` 里
`setAccountModelRateLimitSnapshot`、`ratelimit_service.go` 里
`applyAnthropicFableSchedulingThreshold`（在原判定 `!decision.ShouldPause` 分支里调用，
落 `SetModelRateLimit(anthropicFableRateLimitKey)` 而不是整号暂停）。

基座逐个 grep **全部在位**：`anthropicFableRateLimitKey` / `resolveEffectiveAccountSchedulingThreshold`
/ `candidateMatchesThreshold` / `parseSchedulingResetAt` / `utilizationAsPercent` /
`BuildDetailedAccountSchedulingThresholdReason` / `isRateLimitActiveForKey` / `SetModelRateLimit`
/ `modelRateLimitsKey` / 两个 `passive_usage_7d_oi_*` extra 键。

### 3.6 `901a77cfb` Anthropic→Chat 桥丢弃工具调用的 thinking，DeepSeek 多轮必现 400

3 文件 +266-3，产品文件干净（2/3 ok）。基座 `ChatMessage.ReasoningContent`
（`apicompat/types.go:671`）在位。状态：**已合**

`/v1/messages` 打到只支持 Chat Completions 的 openai_compat 上游时，本仓库
`chatcompletions_anthropic_bridge.go:254` 的 `anthropicAssistantToChatMessages` 把历史
assistant 消息里的 thinking 块整块丢弃。而这些 thinking 块本来就是**同一个文件的出站方向**
（`chatMessageToAnthropicBlocks`）用上游 `reasoning_content` 生成的——桥自己造出来的东西，
客户端原样回传时被自己丢掉。DeepSeek thinking mode 要求产生工具调用的 `reasoning_content`
随该 assistant 消息回传，缺失即 400：**单轮正常，一进多轮工具对话必现失败**。

修法：新增 `anthropicThinkingToReasoningContent(blocks, hasToolCalls)`，只在该消息带
`tool_calls` 时把 thinking 折回 `ReasoningContent`（多块用 `\n` 连接，`redacted_thinking`
与只有签名的占位块跳过）。作用域与兄弟路径 Responses→Chat 的 `pendingReasoning` 严格一致：
纯文本轮次维持现状。

本仓库虽然没有 DeepSeek **平台**，但 DeepSeek 是 `openai_compat` 上游的常见目标
（`pkg/openai_compat/upstream_capability.go` 里就有），这条成立。

### 3.7 `5688bcba9` `/v1/messages` 未按 Claude Code session 做粘性路由

3 文件 +71-4，**全部干净**。状态：**已合**

本仓库 `handler/openai_gateway_handler.go:1278` 的 `resolveOpenAIMessagesMetadataSession`
只看 `metadata.user_id` 与 body fallback，不看 `X-Claude-Code-Session-Id`
（`service/session_id.go:20` 的 `claudeCodeSessionHeader` 已存在，但只喂
`ExtractClientSessionID` 做用量关联，没进路由）。Claude Code 走 `/v1/messages` 时会话边界
拿不到最稳定的那个信号 → 账号粘性漂移、缓存命中率掉。

修法：`session_id.go` 导出 `ClaudeCodeSessionIDFromHeader(c)`；
`resolveOpenAIMessagesMetadataSession` 签名多收一个 `*gin.Context`，在
`promptCacheKey == ""` 时用 `DeriveSessionHashFromSeed(claudeSessionID)` 作 sessionHash。
**注意上游刻意不把它提升成 `prompt_cache_key` 或上游 `session_id`**，只做本地账号粘性，
否则会改变现有 Messages→Codex 缓存滚动语义——移植时不要"顺手优化"。

### 3.8 `c31fe2ed9` SMTP 连通性测试会把已保存的 TLS 配置覆盖掉

2 文件，产品文件 `handler/admin/setting_handler_email.go` 干净。状态：**已合**

本仓库 `setting_handler_email.go:20` / `:88` 两个请求结构体的 `SMTPUseTLS` 是值类型
`bool`，`:67` / `:145` 直接 `UseTLS: req.SMTPUseTLS`。前端「测试连接」/「发送测试邮件」
不带该字段时 JSON 解出 `false`，于是**把已保存的 TLS 开启态按 false 用**——测试失败得莫名，
更糟的是若该路径回写配置就直接破坏已存设置。

修法：两个字段改 `*bool`，新增 `resolveSMTPUseTLS(requested *bool, savedConfig *service.SMTPConfig) bool`
在 nil 时回落到 `savedConfig.UseTLS`。

### 3.9 `eb4237a2b` 带 effort 后缀的模型名命不中渠道定价、被官方兜底价覆盖

2 文件 +188-2，产品文件干净。基座 `normalizeKnownOpenAICodexModel` 在位（5 处）。
状态：**已合**

官方兜底价对 OpenAI / Codex 族会把 `gpt-5.6-luna-high` 这类变体名归一化到基名
（`billing_service.go` 的 `normalizeKnownOpenAICodexModel` 分支），而渠道定价此前只认字面名。
两者不对称 → 管理员只配基名、请求模型带 effort 后缀时，渠道定价未命中而官方兜底命中，
**计费候选循环首个成功即返回，渠道定价永远轮不到**（上游 issue #5256）。

修法：新增 `(*ModelPricingResolver).lookupChannelPricingNormalized`，字面名优先、未命中再用
归一化名查一次；`Resolve`（本仓库 `:71`）与 `applyChannelOverrides`（`:201`）两处调用点都换掉。
非 OpenAI 模型 `normalizeKnownOpenAICodexModel` 回空串，天然 no-op。

与本 fork「用量落库 / 计费准确」同向，属该收的一类。

### 3.10 `92a550973` OpenAI 账号重新授权没走 refresh token

2 文件，**全部干净**（纯前端 `components/admin/account/ReAuthAccountModal.vue`）。
状态：**已合**

重新授权弹窗此前只有 Antigravity 分支能用 refresh token 直接换凭据，OpenAI 号缺同一条路径。
上游补的是 `openaiOAuth.validateRefreshToken` → `adminAPI.accounts.applyOAuthCredentials`
这一串，错误回落顺序 `detail → message → error.message → i18n`。

### 3.11 `81ac8ccd6` 非流式路径把 HTTP 200 终止失败事件写成固定 502，不换号

4 文件 +327-1。基座**全部在位**：`openAIStreamFailedEventShouldFailover`(9)
/ `openAIStreamErrorEventShouldFailover`(4) / `writeOpenAINonStreamingProtocolError`(3)
/ `newOpenAIStreamFailoverError`(20) / `OpenAICompactKeepaliveAdjustedWrittenSize`(8)
/ `handleSSEToJSON`(6) / `handlePassthroughSSEToJSON`(3) / `IsResponseCommitted`(6)。
状态：**已合**（先按 `response.failed` 部分合，见 §9.3；裸 `error` 帧的基座随后补齐，见 §13）

`stream=false` 时上游仍可能回 SSE（其他 sub2api 实例、部分 openai_compat 上游），容量/限流
错误经 HTTP 200 的 `response.failed` / `error` 终止事件回传。`handleSSEToJSON` 与
`handlePassthroughSSEToJSON` 把所有终止事件塞进固定 502，而几百行外的**流式**读取器对同一帧
走 `openAIStreamFailedEventShouldFailover` / `openAIStreamErrorEventShouldFailover` 判定并返回
`UpstreamFailoverError`。**同一上游、同一事件，仅因请求上的 stream 标志而结果相反**：
流式换号，非流式把上游原文直接抛给客户端，池里还有可调度账号也不换。

修法：`openai_gateway_passthrough.go` 新增 `nonStreamingTerminalFailureFailover`，按
`terminalType` 分派（裸 `error` 帧走保守分类器，`response.failed` 走完整的），
`openai_gateway_response_handling.go` 加 3 行接上。响应体已被 `ReadUpstreamResponseBody`
完整缓冲，判定发生在写出任何语义字节之前；能否真正换号仍由 handler 的
`openAIForwardMayFailover` 仲裁，service 侧只额外拒绝 `IsResponseCommitted` 与 `account == nil`。

⚠️ **行为变更**：未被分类为不可重试的泛化 `response.failed` 在非流式路径上由「回写 502」
变为「换号」。上游现有用例 `TestHandleSSEToJSON_ResponseFailedReturnsProtocolError` 传的是
nil account，断言原样保留、仅更名，移植时照做。

### 3.12 `b7ec3cdad` raw CC 流在任何终止 chunk 之前被截断，却记成成功

4 文件 +2xx，含新文件 `service/openai_raw_stream_truncation.go`（161 行）。
状态：**已合（按本地 return 形态改写，见 §9.4）**

raw Chat Completions 直转路径（本仓库 `openai_gateway_chat_completions_raw.go:256`
`streamRawChatCompletions`）此前只要 HTTP 200 就按成功收尾，上游中途断流
（Cloudflare edge reset、后端 worker 掉线）被伪装成 **`HTTP 200 + usage 0/0`**：
客户端拿到半截回答，网关既不报错也不计入 SLA，Ops 侧完全不可见。

修法：新增 `openAIRawStreamTerminalState`，记录三种终止信号任一是否出现
（`[DONE]` / usage chunk / `finish_reason`；只认 `[DONE]` 会在 EOF-after-last-chunk 的兼容
上游上误判并丢掉真实交付的计费）。流结束时若非客户端取消且未见终止信号：
未写出任何字节 → `UpstreamFailoverError` 透明换号；已写出语义字节 → 带类型的
`openAIUpstreamStreamReadError`，handler 补发 SSE error 帧并计 SLA 失败。

⚠️ **必须改写而不是照抄**。上游把尾部 return 抽成 `resultWithUsage()` 闭包时，字段列表是
**上游的**形态；本仓库 `:384-397` 的 return 少一个字段、`ServiceTier` 表达式也不同：

| 上游 | 本仓库 |
|---|---|
| `UpstreamResponseServiceTier: observedUpstreamResponseServiceTier(c)` | **无此字段**（零命中，属 §5.2 未合的 service_tier 基座） |
| `ServiceTier: resolvedOpenAIUpstreamServiceTier(c, serviceTier)` | `ServiceTier: firstNonNilServiceTier(providerServiceTier, serviceTier)` |

抽闭包时保留本地两行，`ErrOpenAIUpstreamStreamTruncated` 在新文件里定义（本仓库零命中，
由这条自带）。其余 `newOpenAIUpstreamStreamReadError` / `newOpenAISilentRefusalFailoverError`
/ `newOpenAIChatSilentRefusalDetector` / `isOpenAIChatUsageOnlyStreamChunk` 都在位。

### 3.13 `897faea33` 重置配额没清 rate limit 冷却，账号仍不参与调度

11 文件 +179-13，9/11 ok。状态：**已合（Ent 形态手写；改名决定与文中相反，见 §9.5）**

管理端「重置配额」调 `admin_account.go:1546` → `accountRepo.ResetQuotaUsed`
（本仓库 `repository/account_repo.go:3459`）。本仓库这个实现**只动 extra 里的配额字段**，
不碰 `rate_limited_at` / `rate_limit_reset_at` —— 因配额打满被限流的号，重置配额后
**冷却仍在，账号照样调度不到**，管理员看不出为什么没生效。

⚠️ **上游那段 SQL 不能照抄**：它是 PG jsonb（`'{}'::jsonb`、`|| '{...}'::jsonb`、
`- 'quota_daily_start'`）。本仓库这个方法早就重写成 Ent 事务
（`withRepositoryTransaction` + `normalizeJSONMap` + `SetExtra`），移植时只需在同一事务里
补 `ClearRateLimitedAt()` / `ClearRateLimitResetAt()`，**不要引入任何 jsonb 语法**
（README §5）。上游那两个 `RowsAffected == 0 → ErrAccountNotFound` 的守卫本仓库已由
`client.Account.Get` + `translatePersistenceError` 覆盖，无需重复。

上游同时把方法**改名**为 `ResetQuotaUsedAndClearRateLimitCooldown`。要不要跟着改名自行决定，
但**若改名就要同步 `service/account_service.go:119` 的接口声明和所有测试 stub**
（本轮扫到 6 个文件里有该方法的 mock/stub，DEV_GUIDE 坑 6）。保留旧名、只改实现是更省的选项。

### 3.14 `44003d7f6` Anthropic / Bedrock 传输层错误不换号、不摘号

10 文件 +290-104，含新文件 `service/gateway_upstream_transport_error.go`（106 行）。7/10 ok。
状态：**已合**（改名 3 个标识符后其余全干净，见 §9.6）

上游 `Do` / `DoWithTLS` 返回**非 HTTP 错误**（代理 / DNS / TCP / TLS）时，Anthropic 侧五条
转发路径直接向客户端写 502 且不换号；单账号网络故障期间该账号**仍持续被调度**，请求全量失败。

本仓库 `gateway_forward.go:380-405` 就是修复前形态，逐字对得上：

```go
resp, err = s.httpUpstream.DoWithTLS(...)
if err != nil {
    ...
    c.JSON(http.StatusBadGateway, gin.H{...})          // ← 直接写 502
    return nil, fmt.Errorf("upstream request failed: %s", safeErr)  // ← 非 failover 错误
}
```

五条路径：`gateway_forward.go`、`gateway_anthropic_passthrough.go`、`gateway_bedrock.go`、
`gateway_forward_as_responses.go`、`gateway_forward_as_chat_completions.go`。

**OpenAI 侧本仓库已有对应实现**，所以这条本质是补齐平台间的不对称：

| 上游改名后 | 本仓库现名 | 位置 |
|---|---|---|
| `classifyUpstreamTransportError` | `classifyOpenAITransportError` | `openai_upstream_transport_error.go:68` |
| （沿用） | `handleOpenAIUpstreamTransportError` | 同文件 `:108`，20 处引用 |
| （沿用） | `tempUnscheduleOpenAITransportError` | 同文件 `:153`（已含「持久化 + 内存快路径」注释） |

移植要点：
1. 分类器改平台无关命名，两侧共用；本仓库改名要扫那 20 处引用。
2. 新增 `(s *GatewayService) handleUpstreamTransportError(ctx, c, account, err, OpsUpstreamErrorEvent{…})`，
   返回 `*UpstreamFailoverError`(502) 交给 handler 的 failover 循环，**service 不再写响应**。
3. 持久性故障（connection refused / no route / DNS not found / 代理认证失败）额外临时摘号 10 分钟。
   ⚠️ 上游注释写「Anthropic 调度以持久化状态为准，无内存快路径」——本仓库同样只需写库，
   但要确认 `SchedulerOutbox` 那条链路被触发（README §7 那次调度事故就是这一层出的问题）。
4. `context.Canceled` 保持原样返回：不换号也不摘号。
5. Ops 事件保留各路径原有字段（`UpstreamURL` / `DurationMs` / `Passthrough`）。
6. 本仓库 `gateway_forward.go` 那段里还有一行
   `scheduleOllamaCloudUsageActivity(s.deferredService, account)`（上游同形），
   上游删掉了它——本仓库无 Ollama 平台，跟着删或原样保留都行，**保留更保守**。

**为什么这条对单机 fork 最要紧**：账号数少（常常 1–2 个），一次代理抖动现在会让该账号所有请求
直接失败到客户端，而不是切到备用号；且故障期间它一直留在调度池里。

---

## 4. P1 — 值得做，按需排期

**4.1–4.11 已于 2026-08-31 全部合入**（实施记录见 §11）；下面「按需」那批仍未合。

本节原先只做了三态 + 基座抽查、**没有逐条读完整 diff**（§3 才是逐条核过的），
实施时逐条复核过，偏差见 §11。

| # | commit | 内容 | 三态 | 备注 |
|---|---|---|---|---|
| 4.1 | `a3bbf33c0` | 渠道监控 v2 按用户可见分组收窄（**权限收窄**） | 5/6 ok | 单人部署无暴露面；把 key 分给别人才成立。要动 `wire_gen.go`（走 `go generate ./cmd/server`，不要手改）与手写 SQLite 的 `channel_monitor_v2_repo.go` |
| 4.2 | `d881bfc0d` | pool 两跳重复计算 system 提示词 | 1/2 ok | 计费准确性，与本 fork「用量落库」同向 |
| 4.3 | `0756c9810` | 批量编辑无法关闭指纹收敛（显式提交 `codex_fingerprint_mode=off`） | 1/2 ok | 0.1.177 把默认改成 off 之后这个更刺眼 |
| 4.4 | `c66e700f0` | TTFT 指标模式挪到管理端设置 | 16/17 ok | 配置面小改，附带 `0fcec63b6` / `8553c91c1` 两条 fixture |
| 4.5 | `3f1581b2d` | 上游倍率探测导致账号列表整页刷新 | 7/8 ok | 纯前端 UX |
| 4.6 | 到期时间本地时区簇 | `5778739cd` `81e461f65` `ae1bcdc25` `263605779` `d66bc88e6` `8177f27aa` `b7aca87fd` | 混合 | ⚠️ **必须按最终形态整簇合**：`d66bc88e6` / `8177f27aa` 是同簇的路径修正（上游先改错位置再搬），单独 cherry-pick 会落到错文件 |
| 4.7 | `706b5676a` | 分组创建/更新失败时显示 API 报错 | CLEAN | 2 文件，纯 UX 收益 |
| 4.8 | `ed12ea716` | 前端 Codex API key 模式内联鉴权 | CLEAN | 2 文件 |
| 4.9 | `0aef702b6` | 隐藏的 cache tooltip 造成横向滚动 | CLEAN | 2 文件，纯样式 |
| 4.10 | `9e7aff59d` | 版本号带连字符后缀时 `parseVersion` 解析失败、误报有新版本 | CLEAN | 3 行。⚠️ 本 fork 用自有 `1.1.x`，更新检查本来就在和上游 `0.1.x` 比，这条只是让它不再乱报；真要治本是关掉更新检查 |
| 4.11 | `c03776604` | Claude attribution 头被前端丢掉 | 产品文件 ok / spec 冲突 | §2 第二种假信号的样本 |

按需（用到才合）：

- `32ad1dcdc` 订阅周/月窗口展示的重置时间与实际推进锚点对齐（3/4 ok）—— simple mode 弱化订阅，用订阅才合
- `b5827cfd5` DeepSeek 峰谷价对齐官方（3/5 ok）—— 只在真给 DeepSeek 模型计费时有意义
- `e6ea7b9af` + `d077002eb` + `6ff771d3d` OpenAI 图像工具冷却三条（3/4、0/3、4/9）—— 只在用图像生成时有意义；`6ff771d3d` 是 feat（冷却时长可配）
- `88cb79d8b` Grok 客户端 `prompt_cache_key` 优先于 `X-Grok-Conv-Id`（CLEAN）+ `00efee430` Grok 4.6 广告 xhigh 档位（2/4）—— 用 Grok 才合
- `50ba14629` 保留多模态客户端工具输出（1/3 ok）、`32064d39e` 跨供应商 reasoning replay 归一化（1/4 ok）、`3c5553e25` 网关合成 Responses 对象补 `created_at`（4/8 ok）、`60756c0ca` `/v1/responses` 透传路径预输出 SSE keepalive（1/3 ok）—— 都偏 §5.4 那簇的边缘，冲突面不小
- 支付三条 `02eee39dd`（充值汇率显示所选币种，CLEAN）/ `1e8745c88`（EasyPay 相对 payurl）/ `d522aed65`（OAuth 注册保留 promo code，4/5 ok）—— simple mode 下基本用不上

---

## 5. 本轮不做

### 5.1 Codex routed model catalog 簇（11 条）—— 条件解除但要单独立项

`22e1b8144` `e471be730` `3e98a5a1a` `b16ed03ca` `5a2f542ab` `e39fce270` `fc589bce1`
`db01fb98f` `5934981e2` `195b21970` `2abce6503`（PR #5926，contrib）。

PORTING-0.1.183 §5.3 当时的理由是「未发布 + 上游仍在返工」。**「未发布」这条现在解除了**——
整簇已进 v0.1.184 正式 tag。但冲突面很大，只能按功能手工移植：

| commit | 三态 |
|---|---|
| `e39fce270` sync routed capabilities | 10 ok / 11 conf / 2 miss |
| `2abce6503` harden capability sync | 2 ok / 11 conf |
| `195b21970` isolate API-key catalog cache | 0 ok / 7 conf / 2 miss |
| `e471be730` complete routed catalogs | 4 ok / 5 conf / 3 miss |
| `b16ed03ca` align with actual routes | 3 ok / 5 conf |

⚠️ **`3e98a5a1a`（composite 精确账号别名路由）6 文件全干净，仍然不能单独合。** 它给
`CompositeRouteResolver` 加了 `CompositeModelOwnershipResolver` 回调并在 `NewGatewayService`
里接线，是这个功能簇的一环。0.1.183 §5.3 已把它记成假信号样本，本轮**继续成立**。

📌 **2026-08-31 立项复核：上面「冲突面很大」这个判断被证伪了一半，见 §14。** 三态里的
CONFLICT 主要来自行号漂移，不是深度分歧——本仓库在这簇的核心文件上与上游**几乎没有分叉**。

上游 v0.1.184 之后还有一条 `e2624fb65`「fix(codex): preserve known image input capabilities」
仍在这簇上打补丁——立项时把它一起纳进来。

### 5.2 Fast mode `service_tier` 一套（8 条）—— 缺 0.1.180 §7.2 基座

`624e4eef6` `8b4b3f4a9` `2b8cb628b` `3a9070359` `50ad6e2e5` `82105f260` `f323d8464` `d39fc491e`。

三态以 NOFILE / CONFLICT 为主（`3a9070359` / `50ad6e2e5` 两条 2 文件全 NOFILE）。
`observedUpstreamResponseServiceTier` / `resolvedOpenAIUpstreamServiceTier` 本仓库零命中——
这两个符号也是 3.12 需要改写的原因。**要做先做 0.1.180 §7.2。**

### 5.3 Spark / OpenAI 配额 429 簇（4 条）—— 本仓库该缺陷不成立

`5d9c7abed`（3/10 ok）`3c22e78af`（0/3）`571d1e1d9`（0/3）`804679d99`（0/6）。

同 PORTING-0.1.183 §5.2 的结论：本仓库**没有那个同号重试窗口**
（`shouldRetryOpenAIOAuth429OnSameAccount` 等逐个零命中），语义上本来就是「拿到 429 直接
`BlockAccountScheduling`」。上游这簇是在那个窗口里做模型级收窄，**基座不存在，缺陷也不成立**。

### 5.4 WS v2 passthrough 簇（5 条）—— 缺 0.1.180 §7.1 基座

`d5a012463`（2 文件全 NOFILE）`f4e3eb1c5`（0/3 + 2 miss）`7c616db07`（2/4）
`c83dced4b`（0/1 + 1 miss）`d8694f03b`（纯测试，0/1）。

### 5.5 usage 两个新字段（3 条 + 2 迁移）—— 功能非缺陷，迁移要转 SQLite

`1cc6999ad` feat(usage) native compaction v2（42 文件，29 ok / 12 conf）、
`11ada80d5` feat(usage) 映射前 reasoning effort（40 文件，23 ok / 14 conf）、
`5705f4a4a` 对用户隐藏映射后 effort（0/8）、外加两条 test（`1a61eb715` `a8cfe746b`）。

收益只是用量记录多两列。代价：

- 上游迁移是 PG：`ALTER TABLE ... ADD COLUMN IF NOT EXISTS`（SQLite **不支持** `IF NOT EXISTS`
  在 ADD COLUMN 上）+ `COMMENT ON COLUMN`（SQLite 无此语法）。
- 上游三条新迁移**自己撞号**，都叫 `231_*`；本仓库最新是 `224`，要自己排号，
  **不要搬文件名**（README §5、CLAUDE.md SQLite 约束第 3/4 条）。
- 时间列若涉及 `*_at` 记得用 `DATETIME`（CLAUDE.md 约束第 2 条）。

### 5.6 `b56c61ecc` 每用户限制可绑定的公共分组（+1 迁移）—— 单人收益低

29 文件，25 ok / 2 conf / 2 miss，看着很干净，但要：改 `ent/schema/user.go` + `go generate ./ent`
+ 提交生成的 `ent/` + 新建 SQLite 迁移（`restrict_public_groups BOOLEAN NOT NULL DEFAULT false`）
+ 扩 `user_allowed_groups` 语义（从「只装独占分组」变成「也装被放行的公共分组」）。

⚠️ 撞 CLAUDE.md SQLite 约束第 5 条：`EnsureSQLiteAuxTables` 里 `user_allowed_groups` 的 DDL
必须与迁移**列对列**一致，两边都是 `IF NOT EXISTS`、谁先跑谁生效，不一致会让迁移文件变成
误导。而且登录路径会加载 allowed groups，**这张表出问题直接 503**。

### 5.7 `b1737cc84` Antigravity 混合内置工具 —— 缺 0.1.178 基座，可拆一半

6 文件 5 ok / 1 conf，但 §2 已说明这是「`ok` 不保证能编译」：它改的
`hasMixedToolInvocations` 与 `includeServerSideToolInvocations` 本仓库**零命中**
（上游 0.1.178 那条 Gemini `includeServerSideToolInvocations` 修复没移过），
唯一冲突的文件正是 `request_transformer.go`。

**可拆**：`apicompat/chatcompletions_to_responses.go` + `apicompat/types.go` 两个 hunk
（Chat 工具里 `web_search` / `code_execution` 类型透传到 Responses）**干净且不依赖 antigravity**，
可以单独作为 P1 合。剩下 antigravity 的两个文件等 0.1.178 那条基座。

### 5.8 Grok 无效工具 union / vision（4 条）—— 属挂起的 Grok 整簇

`de6ef7134`（1/4）`fd872550d`（0/2）`f4820c00d`（0/2）`b4b537164`（1/3，缺
`openai_gateway_grok_model_input.go`）。属 PORTING-0.1.180 §6.2(c) 那个整簇，
决策仍未拍板，本轮不单独动。

---

## 6. N/A / 已合

| 上游条目 | commit | 原因 |
|---|---|---|
| Ollama Cloud 用量窗口挂国产三家 | `30b29e51e` | 无 Ollama 平台（3/12 ok 但语义不成立） |
| Zhipu team GLM Coding Plan 用量查询 | `c4e46c3be` | 无 Zhipu 平台 |
| 监控配额抓取 singleflight 内重查缓存 | `e652f6e20` | 2 文件全 NOFILE |
| 前端用量窗口统一渲染 + 告警阈值 | `6532d5b61` | 2 ok / 4 conf / 4 miss，主要服务上面两条国产项 |
| VERSION 同步 0.1.183 | `7634e3c23` | 本仓库用自有编号 `1.1.10` |
| sponsors 素材 | `6ca1e15b0` `66d664ff0` | 上游 README / 素材专属 |
| 上游误放的组件副本清理 | `3673702af` `94edcd5d8` | 本仓库没有那两个误放文件 |
| rollup trigger 会话时区测试 | `c5ff640df` | 纯测试，对应产品改动在 0.1.177 日汇总表里 |
| 错误日志记录真实上游端点 | `4795650d2` | **已合**：本仓库 `f6ed24ade`（0.1.183 §4.2） |

---

## 7. 建议顺序

```text
① 3.1  8f5451587  Anthropic→Responses thinking/content_index  —— 2 文件全干净，症状最重（成功但零输出）
② 3.2  da10822d7  appendRawJSON 占位符                        —— 6 行
③ 3.3  9f1effd71  分组限额被清零                              —— 涉配置破坏，且被 admin skill 放大
④ 3.4  32ac921f2  Fable OAuth refusal（+扩写原生 /v1/messages）
⑤ 3.5  ea4291a92  Fable 调度阈值收窄
⑥ 3.6–3.10 一批：901a77cfb / 5688bcba9 / c31fe2ed9 / eb4237a2b / 92a550973
⑦ 3.11 81ac8ccd6  非流式终止事件换号（行为变更，单独一轮自测）
⑧ 3.12 b7ec3cdad  raw CC 流截断（按本地 return 形态改写）
⑨ 3.13 897faea33  重置配额清冷却（Ent 重写，注意是否改名）
⑩ 3.14 44003d7f6  Anthropic/Bedrock 传输层 failover  —— 最大手工量，收益最高，单独一轮
不做：§5 六簇。单独立项：§5.1 Codex routed catalog（含 upstream e2624fb65）
```

文件冲突协调：

- 3.4 与 3.14 都动 `gateway_forward.go`（3.4 在 `:195` 附近，3.14 在 `:380` 附近，不重叠但同文件）
- 3.2 与 3.14 都动 `gateway_forward_as_responses.go`（3.2 在 `:610`，3.14 在传输错误分支）
- 3.11 与 3.12 都在 OpenAI 网关，但分别是 `openai_gateway_passthrough.go` /
  `openai_gateway_chat_completions_raw.go`，不冲突
- 3.5 与 3.13 都碰账号限流/配额状态，但一个走 `SetModelRateLimit`、一个走 `ResetQuotaUsed`，
  语义不重叠——3.13 清的是**账号级**冷却，3.5 写的是**模型级**冷却，顺序无所谓

---

## 8. 自测

沿用 [README 第 3.4 节](./README.md#34-自测再合回-main)：

```bash
cd backend
go build ./...
go test -tags=unit ./... -count=1
golangci-lint run ./...
cd ../frontend && pnpm run typecheck && pnpm run lint:check
cd .. && make test-frontend-critical
```

按本轮改动重点盯：

- **apicompat 双向桥**（3.1 / 3.2 / 3.6）：
  `go test -tags=unit ./internal/pkg/apicompat/ -count=1`。3.1 那两条坏顺序用例
  （text→thinking、同 item 两个 text 块）与生命周期不变式用例（`output_item.added` /
  `.done` 必须配平）是这条的验收核心
- **failover / 传输层**（3.11 / 3.12 / 3.14）：
  `go test -tags=unit ./internal/service/ -run 'Failover|TransportError|Truncat|Terminal' -count=1`。
  3.14 额外手动验一次：把某账号的代理指向一个不存在的端口，确认请求**换到另一个账号**而不是回 502，
  且该账号被临时摘除 10 分钟
- **调度 / 限流**（3.5 / 3.13）：`-run 'SchedulingThreshold|RateLimit|ResetQuota'`。
  3.13 改完手动验：给账号打上限流 → 点「重置配额」→ 确认它立刻重回调度
- **计费**（3.9）：`-run 'Pricing|ChannelModelPricing'`
- **分组管理**（3.3）：`-run 'Group'`，并用 `skills/sub2api-admin` 实跑一次不带限额字段的分组更新，
  确认限额保持「无限制」而不是变 0
- **前端**（3.10 / §4 那批）：`pnpm exec vitest run src/components/admin/account/__tests__/`；
  全量 `pnpm run test:run` 已知有 1 个**与移植无关**的失败文件
  （`src/composables/__tests__/useRoutePrefetch.spec.ts`，5 条），见 PORTING-0.1.179.md §4.6
- **SQLite 专项**（3.13）：确认没有引入任何 jsonb / `::` 强转；跑
  `go test -tags=integration ./internal/repository/ -run 'Account' -count=1`

---

## 9. 实施记录（2026-08-31，分支 `sync/upstream-20260831`）

§3 的 14 项已全部落地（3.11 为部分合）。本节只记**与第 3 节预案不一致的地方**——预案里
说对了的部分不重复。

### 9.1 3.2 缺一个测试 fixture，就近补在本仓库有的那个文件里

`gateway_forward_as_responses_test.go` 新增的用例调 `toolAnthropicSSEStream()`，而这个
fixture 上游放在 `openai_gateway_anthropic_native_pump_test.go`——本仓库**没有**该文件
（属未移植的 native pump 一套）。逐字照搬那个文件不合适（它还带 pump 的超时用例），
所以把 fixture 就近定义在 `gateway_forward_as_responses_test.go` 里并注明出处。

⚠️ 三态是 3 ok / 1 nofile，看着"产品代码干净就行"，但**测试文件缺失同样会让包编译不过**。
`nofile` 落在测试文件上时，要检查它有没有被留下来的用例引用。

### 9.2 3.4 原生路径扩写落地，并踩到一个"用例永远为真"的坑

按 §3.4 的判断在 `gateway_forward.go` 的注入点补了
`claudeOAuthSystemPromptBlocksForModel(reqModel, systemPromptBlocks)`。用 `reqModel`
（客户端请求名）而不是映射后的名字，理由写在代码注释里：能走到该分支的都是 OAuth 账号，
其映射只有 `claude.NormalizeModelID` 这层前缀/长短 ID 归一化，`isAnthropicFableModel` 是
`strings.Contains(lower, "fable")`，归一化前后都保留该子串，判定不会翻转。

新增 `gateway_forward_fable_oauth_test.go`：复用 `gateway_forward_partial_usage_test.go` 的
`newForwardPartialUsageServiceForTest` + `anthropicHTTPUpstreamRecorder`，走**真实 Forward**
并断言**实际发往上游的 body**。`settingService` 为 nil 时
`claudeOAuthSystemPromptInjectionSettings` 返回 `(true, "", "")`，正好命中"注入开启 + 内置
默认 blocks"这一支。

⚠️ **本轮最容易骗过自己的一处**：照上游的写法
`require.NotContains(t, string(sent), claudeCodeSystemPromptExpansion)` 是**永远为真**的——
上游那条断言打在 `applyClaudeCodeOAuthMimicryToBody` 的返回值上没问题，而这里 body 是
JSON，换行已被转义成 `\n`，跟 Go 常量里的真换行永远对不上。也就是说即使展开块还在，
用例照样"通过"。改成对**解析后**的 block 文本逐块断言，并用一条反向用例
（非 Fable 模型必须仍带展开块）证明探针本身没失效。

### 9.3 3.11 只交付了 `response.failed`，裸 `error` 帧缺基座

`nonStreamingTerminalFailureFailover` 逐字移入，两处调用点接好，`response.failed` 的
换号行为已生效并有用例钉住。但**裸 `error` 帧走不到这个函数**：本仓库
`extractOpenAISSETerminalEvent` 走 `forEachOpenAISSEDataPayload`、按 JSON 的 `"type"` 字段
识别、取**第一个**终止帧，switch 里没有 `"error"`；上游那版早已改成 `forEachOpenAISSEFrame`
（按 SSE 的 `event:` 行识别、取**最后一个**终止帧）并把 `"error"` 加进 switch。

期间踩到第 13 条：先把调用点条件扩成 `|| terminalType == "error"`，跑完发现用例照旧红——
因为 `terminalOK` 根本不会为 true，那是**空转改动**，已撤回。上游用例
`TestNonStreamingSSEToJSON_BareErrorEventUsesConservativeClassifier` 打了 `t.Skip` 并写清
解锁条件；helper 里的 `error` 分派保留为与上游同形。

另外两处必须手改：

1. **要把 `account` 串下来**。上游 `handleNonStreamingResponsePassthrough` /
   `handlePassthroughSSEToJSON` 都带 `account *Account`，本仓库两个都没有。改了 2 处签名
   + 1 处产品调用点 + 3 处测试调用点（`nil` / `&Account{ID: 91}`，与上游 v0.1.184 一致）。
2. `TestHandleSSEToJSON_ResponseFailedWithoutAccountReturnsProtocolError` 上游把入参从
   `&Account{ID: 1, Type: AccountTypeOAuth}` 改成 `nil`，**这个改动不在 `81ac8ccd6` 里**
   （在本区间更靠后的某条提交里）。只合 `81ac8ccd6` 会得到"名字说 without account、实际传
   了 account"的自相矛盾用例并直接失败。按 v0.1.184 终态改成 `nil`。

⇒ 教训：**逐条 cherry-pick 会落在上游的中间态**。`nonStreamingTerminalFailureFailover` 的
签名也是例子：`81ac8ccd6` 里是 7 个参数，v0.1.184 终态多了 `mappedModel`。移植前先
`git show <tag>:<file>` 看终态，再决定按哪一版落。

### 9.4 3.12 按预案改写，无额外意外

`openai_raw_stream_truncation.go`（新文件）、`openai_stream_read_error.go`、测试文件全部
干净落地。`streamRawChatCompletions` 按 §3.12 的表改写：抽 `resultWithUsage()` 闭包时保留
本地的字段表（无 `UpstreamResponseServiceTier`、`ServiceTier` 仍是
`firstNonNilServiceTier(providerServiceTier, serviceTier)`）。上游 8 条新用例全过。

### 9.5 3.13 改名决定与 §3.13 相反：跟着上游改名更省

§3.13 建议"保留旧名以免动接口和 stub"。**实测反了**：上游那条提交里，
`service/account_service.go`（接口）、`admin_account.go` 与 5 个测试 stub/契约文件的改名
patch 在本仓库 `apply --check` **全部干净**，直接 apply 即可；保留旧名反而要手工改回
upstream 已经写好的 6 个文件。所以采纳上游命名
`ResetQuotaUsedAndClearRateLimitCooldown`，只有 `account_repo.go` 的实现按 Ent 形态手写
（`SetExtra(...).ClearRateLimitedAt().ClearRateLimitResetAt()`，不引入任何 jsonb 语法）。

两处测试要处理：

- **删掉**上游的 `TestAccountRepository_ResetQuotaUsedAndClearRateLimitCooldown_NoRowsAffectedReturnsNotFoundWithoutOutbox`。
  它拿 `newAccountRepositoryWithSQL(nil, exec, nil)` 断言"一条 UPDATE accounts +
  RowsAffected==0 → ErrAccountNotFound"，那是上游裸 SQL 实现的形状；本仓库是 Ent 事务，
  Ent client 为 nil 直接 panic。删除理由已写进该文件的注释。
- ⚠️ 上游补的那条**集成用例在本仓库永远不跑**：
  `account_repo_integration_test.go` 的 build tag 是 `integration && postgres`。真正的 SQLite
  覆盖改为扩本仓库自有的 `account_repo_sqlite_remaining_test.go`——先 `SetRateLimited`，
  再 `ResetQuotaUsedAndClearRateLimitCooldown`，断言 `RateLimitedAt` / `RateLimitResetAt`
  双双为 nil。**这条是本项唯一真正跑起来的回归**。

⇒ 教训：上游给的用例落在 `integration && postgres` 下时，等于零覆盖。移植带用例的项，
要顺手确认那些用例在本仓库的 tag 组合下会不会被执行。

### 9.6 3.14 改名 3 个标识符后其余全干净

只有 `openai_upstream_transport_error.go` 冲突，冲突全部来自平台无关化改名。手工 `sed`
三个标识符（`openAITransportErrorClass` → `upstreamTransportErrorClass`、
`openAIPersistentTransportErrorMarkers` → `persistentUpstreamTransportErrorMarkers`、
`classifyOpenAITransportError` → `classifyUpstreamTransportError`，本仓库仅 2 个文件 15 处
引用），其余 8 个文件（含新增 `gateway_upstream_transport_error.go` + 其用例，以及五条
Anthropic 路径）全部 `git apply` 干净。

§3.14 第 6 点提的 `scheduleOllamaCloudUsageActivity`：新 handler 在 `context.Canceled`
早退**之后**才调它，与旧内联代码里 `if !errors.Is(err, context.Canceled)` 的语义一致，
无需额外保留或删除。

---

## 10. 本轮验证结果

| 项 | 结果 |
|---|---|
| `go build ./...` | 通过 |
| `go vet -tags=unit ./...` | 通过 |
| `go test -tags=unit ./...`（全量） | 通过，无失败包 |
| 其中 `./internal/repository/` | 通过（含 §9.5 新增的 SQLite 清冷却断言，实跑） |
| 其中 `./internal/service/`、`./internal/pkg/apicompat/`、`./internal/handler/{,admin/}` | 通过 |
| `golangci-lint run ./...`（v2.13.0） | **0 issues** |
| `pnpm run typecheck` | 通过 |
| `pnpm run lint:check` | 通过 |
| `make test-frontend-critical` | 14 文件 / 167 passed / 2 skipped |
| `go test -tags=integration ./internal/repository/ -run AccountRepoSuite` | `no tests to run`（该 suite 是 `integration && postgres`，本 fork 用不上，见 §9.5） |

**未做的手动验证**（需要真实上游/账号，留给部署后）：

- 3.14：把某账号的代理指向不存在的端口，确认请求**换到另一个账号**而不是回 502，
  且该账号被临时摘除 10 分钟
- 3.13：给账号打上限流 → 管理端点「重置配额」→ 确认它立刻重回调度
- 3.3：用 `skills/sub2api-admin` 实跑一次不带限额字段的分组更新，确认限额保持「无限制」
- 3.4：真实 Fable OAuth 请求确认不再出现 `stop_reason=refusal` + 零 output

---

## 11. 实施记录：§4 的 4.1–4.11（2026-08-31，同一分支）

11 项全部合入。同样只记**与 §4 判断不一致的地方**。

干净落地、无偏差：4.5 `3f1581b2d`（8 文件全干净）、4.7 `706b5676a`、4.9 `0aef702b6`、
4.10 `9e7aff59d`。

### 11.1 4.8 `ed12ea716` 三态干净但直接 `is not defined` —— 前端也有这个坑

patch 6 个文件全干净，合上去 3 条 spec 立刻红：`escapeTomlBasicString is not defined`。
上游把这个 helper 放在 `22e1b8144`（Codex routed model catalog 簇，§5.1 未合），而
`ed12ea716` 调了它。

这是本轮**第三次**「`apply --check` 干净不等于能跑」，也是第一次出现在前端 —— §2 那条教训
不只适用于 Go：Vue SFC 里的 `<script setup>` 同样是"符号必须在当前文件/作用域里"。

处理：`escapeTomlBasicString` 本身与那个功能簇无关（TOML basic string 转义是独立正确的
行为），就地补 2 行并注明出处，不为它去拖整簇。

⚠️ 移植前端时**必须真跑 vitest**，`pnpm run typecheck` 抓不到这类问题：`vue-tsc` 对
`<script setup>` 里未定义的顶层标识符在本仓库配置下没报错，只有运行期才炸。

### 11.2 4.6 到期时间簇：`8177f27aa` 必须跳过，不是「顺序问题」

§4.6 只说"必须按最终形态整簇合"。实际形态更明确：

| commit | 时间 | 路径 |
|---|---|---|
| `8177f27aa` | 08-30 22:10 | `CreateAccountModal.vue` / `EditAccountModal.vue`（**仓库根目录**） |
| `d66bc88e6` | 08-30 22:11 | `frontend/src/components/account/…`（正确路径） |

`8177f27aa` 是上游误传到根目录的副本，1 分钟后由 `d66bc88e6` 在正确路径重做，那两个错位
文件再由 `3673702af` / `94edcd5d8` 删掉（本文 §6 已列为 N/A）。所以**跳过 `8177f27aa`、
只合 `d66bc88e6`**；`d66bc88e6` 在本仓库冲突（我们这两个 modal 的上下文与上游不同），
三处改动手工落：import 加 `getBrowserTimeZone`、`const browserTimeZone = getBrowserTimeZone()`、
hint 段落加 `expiresAtTimezoneHint`。

其余按时间顺序 `263605779` → `ae1bcdc25` → `81e461f65` → `b7aca87fd` → `5778739cd` 全干净。
`81e461f65`（`format.ts` 的 `getBrowserTimeZone` + 严格解析）必须在 `d66bc88e6` 之前或同批，
否则前者引用的 helper 还不存在。

### 11.3 4.4 `c66e700f0` 一处 hunk 手工落，并保留上游刻意的不对称

17 文件里 16 个干净，只有 `openai_gateway_response_handling.go` 冲突：上游的上下文里有
`pendingSSEEventType := ""` 这一行，本仓库没有（属另一处未移植改动）。

手工落这几处：`ttftMode := s.openAITTFTMode(ctx)`、`startsTTFTOutput := openAIStreamDataStartsTTFT(...)`、
`eventStartsVisibleOutput` → `eventStartsTTFTOutput`、`completedVisibleEvent` → `completedTTFTEvent`。

⚠️ **有一处看着该改、其实不能改**：本仓库 `:636`
`if firstTokenMs == nil && startsVisibleOutput { shouldFlush = true }`（为 TTFT 尽快 flush 的
启发式）在上游 v0.1.184 终态里**仍然用 `startsVisibleOutput`**，只有真正记 TTFT 的那处
（`:653`）切到 `startsTTFTOutput`。顺手一起改会让 flush 时机跟着 TTFT 模式变，属上游没打算
做的行为改动。判定方式还是 §9.3 那条：先 `git show <tag>:<file>` 看终态。

### 11.4 4.1 `a3bbf33c0` 三件事要额外处理

1. **`wire_gen.go` 不要相信 patch，自己重跑一遍。** 本机 `wire` 不在 PATH、`go run …@latest`
   拉不动（网络），但模块缓存里有 `wire@v0.7.0`，
   `go build -o /tmp/wirebin github.com/google/wire/cmd/wire` 可离线构建。跑
   `/tmp/wirebin gen ./cmd/server` 后与 patch 后的文件 `diff` —— **逐字节相同**，确认那处
   `NewChannelMonitorV2Handler(channelMonitorV2Service, apiKeyService)` 就是 wire 会生成的结果。
2. **SQL 方言核查通过**：这条 patch 对 `channel_monitor_v2_repo.go` 的改动全是 Go 层
   （空作用域短路 + 交集收窄），没有新增任何 SQL，`ANY(` / `::` / jsonb 逐个 grep 零命中。
3. **测试要改写，且不能引入 `lib/pq`。** 上游用 `require.Equal(t, pq.Array([]int64{4}), args[3])`
   断言 PG 的 `group_id = ANY($4)` 单数组参数；本仓库同一个 `channelMonitorV2Where` 走
   `sqlInt64In` 展开成 `group_id IN ($4,$5)` + 逐个标量参数。3 处断言按本仓库形态改写
   （`IN ($4)` + `int64(4)`），其余 5 条逐字照搬。`lib/pq` 虽然在 go.mod 里（几个
   `integration && postgres` 测试用），但**不该为一条新测试把 PG 驱动拉进 SQLite 路径**。

8 条新用例全过，其中 3 条 `EmptyRestrictedScope*` 用 sqlmock 且**不登记任何期望**——
`mock.ExpectationsWereMet()` 正是在钉"空作用域必须在发出任何 SQL 之前短路"。

### 11.5 4.2 / 4.3 / 4.11 手工落，原因都是本仓库上下文不同

- 4.2 `d881bfc0d`：本仓库 `openai_gateway_forward.go` 的 else 分支还是
  `applyCodexOAuthTransform(decoded, isCodexCLI, isCompactRequest)` 三参形态。改成
  `applyCodexOAuthTransformWithOptions` 并两个分支都带上
  `OmitPromotedSystemMessagesFromInput`。基座 `codexOAuthTransformOptions.OmitPromotedSystemMessagesFromInput`
  已存在（`openai_codex_transform.go:86`，Chat 兼容入口已在用），无需补。
- 4.3 `0756c9810`：`BulkEditAccountModal.vue` 冲突，两处手工落（checkbox 的 `id`
  测试钩子 + `extra.codex_fingerprint_mode` 改为无条件显式落键）。上游注释里那句关键
  理由照抄：批量接口只做顶层合并，删 payload 里的键清不掉账号上已有的
  device/session/full，而且只删不写会让 payload 退化成 `{extra:{}}` 被后端判空更新 400。
- 4.11 `c03776604`：产品文件 `UseKeyModal.vue` 干净（4 处 `CLAUDE_CODE_ATTRIBUTION_HEADER=0`
  全删）。spec 冲突（4.8 刚改过同一文件同一区域），只手工移植其中那条独立用例
  `omits the attribution override from every standard Claude Code setup form`；上游另一批
  断言插在它自己的 Grok 用例中间，本仓库那个用例上下文不同，未移植。

### 11.6 本批未合（§4「按需」那组，判断不变）

`32ad1dcdc`（订阅重置锚点）、`b5827cfd5`（DeepSeek 峰谷价）、
`e6ea7b9af` + `d077002eb` + `6ff771d3d`（图像工具冷却）、
`88cb79d8b` + `00efee430`（Grok）、
`50ba14629` + `32064d39e` + `3c5553e25` + `60756c0ca`（§5.4 边缘）、
`02eee39dd` + `1e8745c88` + `d522aed65`（支付）。

都不是"没核过"，是**取决于这套部署实际用不用那些功能**。用到哪块再按 §2 的三道关单独过。

---

## 12. §4 批次验证结果

| 项 | 结果 |
|---|---|
| `go build ./...` | 通过 |
| `go vet -tags=unit ./...` | 通过 |
| `go test -tags=unit ./...` | **53 个包全 ok**，无失败 |
| `golangci-lint run ./...`（v2.13.0） | **0 issues** |
| `wire gen ./cmd/server` 重跑后 `diff` | 与 patch 后的 `wire_gen.go` **逐字节相同** |
| `pnpm run typecheck` / `lint:check` | 通过 |
| `make test-frontend-critical` | 14 文件 / **168** passed / 2 skipped |
| 本批触及的 6 个前端 spec | 62 passed |
| `internal/repository -run ChannelMonitorV2` | 20 条全过（含 8 条新增） |

---

## 13. 补齐 3.11 的缺口：SSE frame 识别 + 裸 `error` 终止事件（2026-08-31）

§9.3 把 3.11 记成「部分合」，缺口是 `extractOpenAISSETerminalEvent` 的 frame 识别重构。
本节把它补上，3.11 转为**已合**，`t.Skip` 已去掉。

### 13.1 基座缺口比预估小得多

§9.3 当时只说「上游那版早已改成 `forEachOpenAISSEFrame`」，没量化代价。实测：

| 符号 | 本仓库 |
|---|---|
| `openAICompatSSEFrameParser` / `openAICompatSSEFrame` | **已存在**，且与上游**逐字同形**（`EventType` + `Data`、`AddLine` / `Finish` / `dispatch`） |
| `forEachOpenAISSEFrame` | 缺，30 行，只依赖上面那对 + 下面那个 |
| `effectiveOpenAISSEEventType` | 缺，5 行，无依赖 |

`openai_sse_data.go` 上下游的**唯一**差异就是少了 `forEachOpenAISSEFrame`。所以这不是
「要拖一整簇」，是 35 行的自包含移植。引入它的上游提交 `acce29af2`（「补齐 OpenAI 与 Grok
协议兼容处理」）是个 ≤ v0.1.183 的大杂烩，**不要整条 cherry-pick**，按符号取即可。

### 13.2 `extractOpenAISSETerminalEvent` 的三处语义变化

改成 `forEachOpenAISSEFrame` 之后，这个函数（3 个产品调用点）有三点变了，都与上游终态一致：

1. **读 `event:` 行**。只在 `event:` 给类型、data 里没有 `"type"` 的帧此前完全落空
   （`terminalOK=false`），非流式路径把它当普通 SSE 按 200 收尾。
2. **switch 里加 `"error"`**。裸 error 帧同属终止失败事件。
3. **取最后一个匹配帧**（旧实现取第一个）。一段体里可能先 `response.failed` 再补一个
   `error`，最后那个才是上游真正的收尾表态。

新增 `openai_sse_terminal_event_test.go` 5 条用例专门钉这三点 + 两条边界（只有增量事件与
`[DONE]` 时必须报 false、`response.completed` 仍是终止事件不能被 `"error"` 挤掉）。
这个函数改动的影响面比看起来大，单独立文件覆盖。

### 13.3 顺带发现：`openAIStreamErrorEventShouldFailover` 有个会挡住新形态的自校验

去掉 `t.Skip` 后 `non_transient` 子用例立刻绿，`transient_fails_over` 仍红。两个原因：

1. **自校验挡路**。本仓库这个函数开头有
   `if gjson.GetBytes(payload, "type").String() != "error" { return false }`，上游没有。
   frame 识别生效后，裸 error 帧的 `terminalType` 是 `"error"`（来自 `event:` 行）而
   payload 里没有 `"type"` —— 正好被这个自校验挡掉，等于新形态白改。
   三个调用点本来就都按事件类型分派过了（流式两处判 `eventType == "error"`，非流式走
   `nonStreamingTerminalFailureFailover` 的 `terminalType == "error"`），这个自校验是多余
   且有害的，按上游去掉。
2. **缺文本标记那一段**。上游末尾有
   `strings.Contains(combined, "temporary" / "try again" / "please retry")`，覆盖「上游只以
   文本形式表达瞬时故障、无 code 无 type」的情形。本仓库只有
   `isOpenAITransientProcessingError`，命不中这类文案。补上同一组标记。

⚠️ **仍与上游有差距，已在代码注释里标明**：上游这个函数还含 `detectOpenAICyberPolicy`
前置拒绝、`isOpenAIUpstreamAccessStateError`、以及按 `openAIStreamFailedEventSemanticStatus`
分派 403/401/429/529 的 switch。其中 `isOpenAIUpstreamAccessStateError` /
`openAIStream403AccountFailure` 本仓库**零命中**，属其它未移植簇，故本轮只补文本标记一段，
没有整体替换这个分类器。

⇒ 这条是 §2 那类假信号的又一个变体：**基座补齐了、调用点也改对了，但路径上还有一道自己
早年加的守卫会把新形态挡掉。** 补基座之后要把整条链路跑通一次，而不是只看新函数被调用到。

### 13.4 验证

| 项 | 结果 |
|---|---|
| `go build ./...` / `go vet -tags=unit ./...` | 通过 |
| `go test -tags=unit ./...` | **53 个包全 ok** |
| `golangci-lint run ./...`（v2.13.0） | **0 issues** |
| `-run BareErrorEvent` | 2 个子用例全过（`t.Skip` 已删） |
| `-run ExtractOpenAISSETerminalEvent` | 5 条新用例全过 |
| `./internal/service/` 整包 | 通过（165s） |

---

## 14. §5.1 Codex routed model catalog 立项（2026-08-31）

### 14.1 这个功能是什么

Codex CLI 的模型发现要求一份**顶层 models manifest**。上游此前：Composite 与其它非 OpenAI
分组要么落到 OpenAI 的 live-manifest handler、要么根本没有 Codex 专用响应；API key 用户也没有
受支持的方式去取那份 manifest 并在 `config.toml` 里引用它。

这簇做两件事：从每个分组的**有效模型列表**生成最小 manifest（同时保留官方 OpenAI live 路径
与普通 `/models` 响应）；给 Use Key 流程加「带鉴权下载 catalog + 配 `model_catalog_json`」，
且不把 API key 写进下载的文件。

**本仓库用得上**：composite 分组存在（`internal/service/composite_model_route.go`），
`/backend-api/codex/models` 端点也已注册（`internal/server/routes/gateway.go:378`）。
§11.1 补的 `escapeTomlBasicString` 就是这簇的产物——我们已经吃下了 Use Key 流程里
「API key 内联鉴权」那半，缺的正是 `model_catalog_json` 那半。

### 14.2 规模

12 条提交（含 v0.1.184 之后的 `e2624fb65`）、49 个文件、约 **+8200 行**。产品代码集中在两个
文件的近乎重写上：

| 文件 | 本簇增删 |
|---|---|
| `service/openai_codex_models_service.go` | **+1639 -97** |
| `service/upstream_models.go` | **+699 -21** |
| `service/openai_codex_model_metadata.go` | +321（新文件） |
| `handler/gateway_handler.go` | +97 -14 |
| `service/gateway_service.go` | +84 -2 |
| `pkg/claude/effort_catalog.go` | +69（新文件） |
| `service/openai_gateway_request_body.go` | +63 |
| 前端 `UseKeyModal.vue` / `api/codex.ts` / `utils/codexCatalogConfig.ts` | +328 / +56 / +63 |

测试约占 4400 行。8 个文件是纯新增（`effort_catalog.go`、`openai_codex_model_metadata.go`、
`api/codex.ts`、`utils/codexCatalogConfig.ts` 及各自用例）。

### 14.3 关键发现：本仓库在这簇的核心文件上几乎没有分叉

§5.1 原先按三态 CONFLICT 数判「冲突面很大、只能按功能手工移植」。**按正确的方法量一遍
（本仓库当前文件 vs 上游 `22e1b8144^` 即入簇前状态），结论相反**：

| 文件 | 与上游入簇前的差异行数 |
|---|---|
| `handler/openai_codex_models_handler.go` | **0**（逐字相同） |
| `service/openai_codex_models_service.go` | **22**（我们 761 行 / 上游 773 行） |
| `handler/gateway_handler.go` | 20 |
| `service/gateway_service.go` | 54 |
| `service/upstream_models.go` | 78 |
| `service/openai_gateway_request_body.go` | **409**（我们 1484 / 上游 1873，缺约 389 行） |

⇒ 六个核心文件里五个近乎一致，三态里的 CONFLICT 主要是**行号漂移**。唯一真正分叉的是
`openai_gateway_request_body.go`（我们少约 389 行，属其它未移植功能），而这簇对它只加 63 行，
需要单独确认落点。

**教训（第 5 条）：三态计数不能当分叉程度的度量。** CONFLICT 数高既可能是深度分歧，也可能
只是漂移。判分叉要拿「本仓库当前文件 vs 上游入簇前的同一文件」直接 diff，这个数才是手工量的
真实上限。§5.1 当初那句「冲突面很大」正是按错的指标下的判断。

### 14.4 依赖自包含（本轮最重要的一道关）

按 §2 的第三道关，对新增代码里的每个函数调用逐个 grep 基座。`openai_codex_model_metadata.go`
调了 6 个本仓库零命中的符号：

| 符号 | 定义在 |
|---|---|
| `codexExplicitModelMappingClaims` | `service/openai_codex_models_service.go`（**本簇内**） |
| `resolveCodexCompositeModelTarget` | 同上（本簇内） |
| `GetUpstreamModelMetadata` | `service/upstream_models.go`（本簇内） |
| `normalizeCodexInputModalities` | 同上（本簇内） |
| `normalizeReasoningLevel` | 同上（本簇内） |
| `normalizeReasoningLevels` | 同上（本簇内） |

**全部由本簇自己提供**，没有一个挂在别的未移植簇上。这与 §5.2（service_tier）、§5.4（WS v2）
那两簇有本质区别——它们缺的是**簇外**基座，做不了；这簇缺的都在簇内，可以做。

### 14.5 分层与建议

| 层 | commits | 性质 |
|---|---|---|
| 骨架 | `22e1b8144` `e471be730` | 两个 feat，manifest 生成 + 完整 routed catalog |
| 路由 | `3e98a5a1a` `b16ed03ca` `5a2f542ab` | composite 别名路由、按真实 route 对齐、配置优先于发现 |
| 能力同步 | `e39fce270` `2abce6503` `db01fb98f` `5934981e2` | 上游能力交集与不可调度账号下的稳定性 |
| 收尾 | `fc589bce1` `195b21970` `e2624fb65` | review 反馈、API-key catalog 缓存隔离、图像能力保留 |

**建议：整簇按「逐 commit 顺序重放」而不是取最终态。** 理由是这两个 feat 之后的 10 条全是对
同一批函数的反复返工，取最终态等于放弃了每一步的可读边界；而按 §14.3 的测量，逐条重放的
冲突主要是漂移，代价可控。中间态**不要求能编译**（§9.3 的教训：逐条 cherry-pick 必然落在中间
态），只要求最后一条落完后 `go build` + 全量用例通过。

**不建议**整体覆盖 `openai_gateway_request_body.go`——那 409 行分叉里含本仓库未移植的其它功能，
只落这簇的 63 行。

**排期**：这是一次独立的 change，规模与 0.1.180 §7.1 那档相当，不适合塞在 P0/P1 批次里。

### 14.6 试跑一遍重放：73/101 干净，但 §14.3 的结论要打个补丁

按 §14.5 的建议做了一次**探查性重放**（topo 顺序逐条、逐文件 apply，结果已撤销，工作树未留痕）：

- **101 个 file-patch 里 73 个干净落地**，8 个新文件全部到位。
- 28 处冲突，集中在两个文件：`openai_codex_models_service.go` **9 处**、
  `UseKeyModal.spec.ts` 5 处、`CreateAccountModal.vue`/`.spec.ts` 各 2 处、
  `upstream_models.go` 2 处，其余 8 个文件各 1 处。
- 核心文件的**第一条**（`22e1b8144`）是干净的；9 处冲突全来自后续返工——文件被前面几条
  改过之后，后面的上下文就对不上了。这符合「逐条重放会落在中间态」的预期。
- 重放后 `go build` 报的全是 `codexModelMetadataOverride` / `UpstreamModelMetadata` /
  `resolveCodexCompositeModelTarget` / `configuredCodexModelDescriptor` / `modelsDevProvider`
  未定义——都在那两个冲突文件里，即「冲突没解 → 符号没落地」，不是新的基座缺口。

⚠️ **§14.3 的「近乎没有分叉」需要打补丁：行数小 ≠ 语义可忽略。** 把那 22 行 / 78 行摊开看：

| 文件 | 本仓库与上游入簇前的真实关系 |
|---|---|
| `openai_codex_models_service.go` | **我们落后**：上游那 22 行是 Codex 出站身份收敛（`resolveCodexOutboundIdentity` / `NormalizeCodexClientVersion` / `CompareVersions` / `codexUpstreamMinVersion`，来自 0.1.175–0.1.177 那批），本仓库还是更早的 `openAICodexProbeVersion` + `codexCLIUserAgent` 形态 |
| `upstream_models.go` | **三种情况混在一起**：① 上游有 `IsCNProvider` 分支，本仓库**刻意没有**（无国产平台）；② 上游用协议感知的 `GetOpenAIProtocolAPIKey` / `GetOpenAIFormatBaseURL`，本仓库还是 `GetOpenAIApiKey` / `GetOpenAIBaseURL`；③ 本仓库有**上游没有的** `buildOpenAIOAuthUpstreamModelsRequest`（约 62 行，让 OAuth 账号用 Codex manifest 走管理端模型同步按钮） |

⇒ **不能拿上游最终态整文件覆盖这两个文件**：会把国产供应商路径合回来（违反硬约束第 1 条）、
引用可能不存在的协议感知 getter、并且**删掉本仓库独有的 OAuth 模型同步**。§14.5 原先只给
`openai_gateway_request_body.go` 打了这个记号，实际上两个核心文件同样适用。

**教训（第 6 条）：分叉的「行数」和「可否整文件覆盖」是两个问题。** 22 行里可能同时藏着
「我们落后于上游」「我们刻意不要」「我们独有」三种情形，每种的处理方式相反。判定必须把 diff
摊开逐段归类，不能只看数字。

### 14.7 修订后的建议

这簇**依然值得做**（§14.4 的依赖自包含结论不变，功能对本仓库也成立），但它是一次
**独立的 change**，不能塞进 P0/P1 批次尾巴，工作量集中在：

1. `openai_codex_models_service.go` 的 9 处冲突 —— 需要先决定是否顺带补齐 0.1.175–0.1.177 的
   Codex 出站身份收敛（补了，冲突大半自消；不补，每条返工都要手工对齐）。
2. `upstream_models.go` 的 2 处冲突 —— 逐段归类：CN 分支剔除、协议感知 getter 按本仓库形态
   落、`buildOpenAIOAuthUpstreamModelsRequest` 原样保留。
3. `openai_gateway_request_body.go` 只落本簇那 63 行（§14.5 已定）。
4. 前端 5 + 2 + 2 处冲突（`UseKeyModal.spec.ts` / `CreateAccountModal.vue` / `.spec.ts`），
   其中 `UseKeyModal.vue` 已被本轮 §11.1 / §11.5 改过两次，冲突里有一部分是我们自己造成的。

**入口条件**：先把「是否补 Codex 出站身份收敛」拍板，这是第 1 项的岔路口，也决定整簇的工作量
是 1 天还是 3 天量级。
