# 移植清单：上游 v0.1.181 / v0.1.182 / v0.1.183

对照日期：2026-08-26。基线 `77df363a0`（= v0.1.179 + 已移植项）。

⚠️ **本文第 1 节以下的「本仓库现状」是 2026-08-26 的快照。** 此后
[PORTING-0.1.180.md](./PORTING-0.1.180.md) 的清单已推进多批（写这句时它的 §5 P0 19 项、
§6.2 / §6.3 的决策与小项、§6.1 的 4 条工具桥接修复均已合），本文第 3 节 12 项 P0 与
§4.2 也已全部落地，只剩 §4.1 待合。读本文的行号与「零命中」结论前先按当前代码复核一遍。

上游 0.1.181–0.1.183 三个版本**全是 bugfix 版**（无新迁移、无新功能）。本文只覆盖
`v0.1.180 (c40edb4) .. main (efb46db0)` 这 44 个非 merge commit；0.1.180 自身的清单仍看
[PORTING-0.1.180.md](./PORTING-0.1.180.md)，两份可以并行推进——本文 P0 与那份的 P0 **没有文件冲突**。

移植前先读 README [第 4 节「硬约束」](./README.md#4-硬约束)；写 SQL 对照
[第 5 节转换速查](./README.md#5-pg--sqlite-转换速查)。

---

## 1. 版本对照

| 点 | 状态 |
|---|---|
| 上游最新正式版 | [`v0.1.183`](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.1.183)（tag commit `c21fd338`，2026-08-25 13:53 UTC） |
| 上游 main | `efb46db0`（2026-08-26，含 0.1.183 之后 6 条未发布提交） |
| 上一轮对照基线 | `v0.1.180`（tag commit `c40edb4`，2026-08-24 11:58 UTC） |
| `v0.1.180` → `main` | **44 commits（非 merge）**：v0.1.181 5 条 / v0.1.182 14 条 / v0.1.183 9 条 / 未发布 16 条 |
| 本仓库版本号 | `backend/cmd/server/VERSION` = `1.1.8`（自有编号，不要同步成 0.1.183） |
| 本仓库迁移号 | 仍是 `224`。**本轮上游没有新迁移**，无需顺延 |
| Go | 不变（本仓库 `1.26.5`，上游 `1.27.0` 的决定仍见 0.1.180 §9.2） |

对比链接：<https://github.com/Wei-Shaw/sub2api/compare/v0.1.180...v0.1.183>

⚠️ **版本归属按 tag 祖先判定，不要按 commit 日期或 `VERSION` 同步提交的位置猜**：上游 release bot
是「先打 tag，再补一条 `chore: sync VERSION to x` 」，所以 `v0.1.181..v0.1.182` 这个区间里会出现
名为「sync VERSION to 0.1.181」的提交。本文每条的版本号用
`git merge-base --is-ancestor <sha> <tag>` 逐条核过，与上游 release notes 对得上。

**结论：12 项 P0 可以直接吃下（患同一缺陷、patch site 对得上），2 项 P1 要手工改，
Responses Lite / OAuth 429 / Codex routed catalog 三簇本轮不做——前两簇缺 0.1.180 的基座，
第三簇上游自己还在返工。**

### 1.1 按簇统计

| 簇 | commits | 判断 |
|---|---|---|
| 可直接移植的独立 bugfix | 12（含一对成对项） | **P0 全收**（第 3 节） |
| 渠道监控 v2 composite 平台归属 | 2 | **P1**，需 SQLite 重写（4.1） |
| 错误日志记录真实上游端点 | 1 | **P1**，需裁剪（4.2） |
| Responses Lite 并行工具调用串 | 5 | 缺 0.1.180 基座 → 不做（5.1） |
| OpenAI OAuth 配额 429 分类 | 1 | 缺基座，且本仓库该缺陷**不成立** → 不做（5.2） |
| Codex routed model catalog | 11 | 未发布 + 上游仍在返工 → 不做（5.3） |
| 国产供应商（Kimi） | 2 | **N/A**（第 6 节） |
| chore（VERSION / sponsors） | 6 | 不合 |
| 纯测试补充 | 2 | 随对应功能项一起，本轮无对应项 |
| 其他（antigravity 测试迁移，已并入 P0 成对项） | 2 | — |

---

## 2. 本轮核查方法

沿用 0.1.180 §2 的两条通道（compare API 在这个体量上会 500）：

```bash
# 部分裸克隆：12 MB，之后 git show 按需惰性取 blob
git clone --bare --filter=blob:none --no-tags --single-branch --branch main \
  https://github.com/Wei-Shaw/sub2api.git /tmp/up183.git
git -C /tmp/up183.git fetch origin \
  refs/tags/v0.1.181:refs/tags/v0.1.181 \
  refs/tags/v0.1.182:refs/tags/v0.1.182 \
  refs/tags/v0.1.183:refs/tags/v0.1.183
git -C /tmp/up183.git log --reverse --format='%h %ad %s' --date=short --no-merges c40edb4..main
```

判定「本仓库有没有这个 bug」——本轮在 0.1.180 §2 的基础上多加一层**按文件切开**，
因为整 commit 的 `apply --check` 一旦在测试文件上失败就看不出产品代码到底对不对得上：

```bash
# 逐 commit 生成 patch
git -C /tmp/up183.git show --format='' --binary <sha> > /tmp/pc183/<sha>.patch
# 再按文件切成单文件 patch，逐个 apply --check
awk '/^diff --git /{n++; f=sprintf("%s/%03d.patch", out, n)} n{print > f}' \
  out=/tmp/split183/<sha> /tmp/pc183/<sha>.patch
```

产出三态而不是两态：`NOFILE`（本仓库没这个文件）/ `CONFLICT` / `ok`。
**大量 CONFLICT 只落在 `*_test.go` 上时，产品代码往往是干净的**，本轮 `9fb26043`
（9 文件里只有 1 个测试冲突）就是这种情况。

⚠️ 仍要注意 0.1.180 §2 提过的两种假信号，本轮又多一种：

- **`ok` 不等于该合**：`3e98a5a1`（composite 精确别名路由）6 文件全干净，但它是未发布的
  routed-catalog 功能簇的一环，后面挂着 5 条返工提交。
- **`CONFLICT` 不等于不该合**：`bc4a9ae4` 的 `billing_service.go` 只是**尾部上下文**不同
  （上游 `calculatePerRequestCost` 开头是 `units := input.UsageUnits`，本仓库是
  `count := input.RequestCount`），被改的函数本体字节一致。
- **`ok` 也不保证能编译**：`apply --check` 只比上下文，不看新代码引用的符号在不在。
  本轮 `d6012b0b` 的目标文件在，但它调用的 `decodeOpenAIJSONUseNumber` 本仓库根本没有。
  凡是新增函数调用，都要额外 `grep -rn "func .*<name>"` 确认基座存在。

---

## 3. P0 — 已核实本仓库有同一缺陷

按建议顺序排列（先小后大、先无冲突后需手改）。

本节 12 项已固化为 OpenSpec change
[`openspec/changes/port-upstream-0.1.183-p0-fixes/`](../../openspec/changes/port-upstream-0.1.183-p0-fixes/)：
`specs/*/spec.md` 是移植后必须成立的行为（17 条 Requirement / 50 个 Scenario），
`tasks.md` 是按文件的实施清单，`verification.md` 是验收证据矩阵。**逐条的 patch site 和上游 diff
说明只在本文，不要在那边重复。** 合完一项后把下面对应条目的状态改成「已合」。

### 3.1 `8e60d574` Codex `session-id` 头未被识别（v0.1.183）

4 文件 +69-2，**全部干净**。状态：**已合**

Codex CLI 发的是 `session-id`（连字符），本仓库
`openai_gateway_scheduling.go:31` 的 `explicitOpenAIHeaderSessionNames` 与
`openai_ws_forwarder_logutil.go:68` 都只读 `session_id`（下划线）。Go 的
`Header.Get` 会把两者规范化成不同的键（`Session-Id` vs `Session_id`），所以带连字符的头
**完全落空** → 粘性会话在客户端重连后漂移到别的账号，缓存命中率归零。

改动就是在两处各加一个分支，`SessionSource` 记 `header_session-id`。本轮性价比最高的一条。

### 3.2 `bc4a9ae4` Anthropic cache_creation 5m/1h 明细重复计费（v0.1.182）

6 文件 +168-26。产品代码 3 个文件里 2 个干净（见 §2 关于 `billing_service.go` 的说明）。
状态：**已合**

`gateway_anthropic_passthrough.go:673/676` 与 `gateway_upstream_response.go:1239/1243`
都是 `if ... .Exists() && v > 0 { 赋值 }`。流式场景里 `message_start` 先报
`ephemeral_5m=N`，后续 `message_delta` 报 `ephemeral_5m=0, ephemeral_1h=M` 时，
**0 被 `> 0` 挡住不覆盖**，旧的 N 留着，最终 `5m + 1h = N + M > cache_creation_input_tokens`
→ 缓存创建费**超收**。

修法两层：
1. 去掉 `&& v > 0`，让 0 也能覆盖（4 处）。
2. `billing_service.go` 新增 `normalizeCacheCreationBreakdown`，当明细之和超过聚合值时按
   原比例缩回聚合值封顶（需要给 import 加 `math`）。第二层是防御性的，即使上游又出新的
   矛盾报数也不会超收，建议一起合。

### 3.3 `19da0f24` Gemini 工具 schema 未清理的字段（v0.1.181）

2 文件 +69-1，**全部干净**。状态：**已合**

`gemini_messages_compat_service.go:3532` 的 `cleanToolSchema` 白名单里缺 `deprecated`，
且 `enum` 里出现非字符串值（数字 / 布尔 / null）时原样透传，Gemini 上游直接 400。
补 `deprecated` 到剔除列表 + 新增 `normalizeGeminiEnum`（能编码成字符串就编码，
遇到对象 / 数组就整个删掉 `enum`）。`encoding/json` 本仓库已在 import 里，无需动 import。

### 3.4 `1a9898a6` Antigravity 兼容模式 max_tokens 未封顶（v0.1.183）

2 文件 +40-1，**全部干净**。状态：**已合**

`antigravity_gateway_compat.go:152` 的 `preserveChatCompletionTokenLimit` 把客户端的
`max_tokens` / `max_completion_tokens` 原样塞进 `claudeRequest.MaxTokens`。上游兼容端点
上限是 64000，客户端要更大就整个请求失败。加 `antigravityCompatMaxTokens = 64000`
常量 + `min()` 封顶。

### 3.5 `4ca86c52` 邮箱换绑缺 alias 查重与并发守卫（v0.1.183）

3 文件 +274-30，**全部干净**。状态：**已合**

两个独立缺陷：

1. **只按字面地址查重**。`auth_email_binding.go:59` 的 `BindEmailIdentity` 与
   `SendEmailIdentityBindCode` 都只查 `GetByEmail(normalizedEmail)`，不查 provider alias
   （Gmail 点号 / `+suffix`）。别人已占 `a.b@gmail.com`，你能拿 `ab@gmail.com` 绑上去，
   两条记录指向**同一个收件箱**。上游抽出 `ensureEmailIdentityAvailableForUser`，
   在字面查重之后追加 `ExistsByEmailAlias`；当前用户自己的 alias 变体仍放行。
2. **前置查重与写入之间有并发窗口**。两个并发请求都能看到「未被占用」。上游新增
   `userRepository.UpdateEmailWithAliasGuard`：在调用方事务里先按
   `normalizedEmailUniquenessLockKey` + `emailAliasUniquenessLockKey` 加锁，复查归属，
   再写 `SetEmail` / `SetPasswordHash`。

**SQLite 说明**：这条不需要任何方言改写。本仓库
`repository/user_profile_identity_repo.go:208` 的 `lockRepositoryScopedKeys` 已经是
**纯进程内锁**（PG advisory lock 那套早就替掉了），单进程语义足够；
`normalizedEmailUniquenessLockKey` / `emailAliasUniquenessLockKey` /
`txAwareSQLExecutor` / `translatePersistenceError` 四个依赖全部已存在。

### 3.6 `329b92ef` OAuth 图片生成 prompt 被模型改写（v0.1.182）

3 文件 +23-6，**全部干净**。状态：**已合**

`openai_images_responses.go:357` 构造的 Responses 请求 `instructions` 是空串，模型会自行
润色 / 翻译 / 增删画面细节。新增 `openAIImagesVerbatimPromptInstructions` 常量并
`sjson.Set` 进去，要求逐字使用用户 prompt。纯提示词改动，零风险。

### 3.7 `e0e5e45c` 自定义工具 item ID 降级后类型不匹配（v0.1.183）

4 文件 +270-21，产品文件全部干净（`internal/pkg/apicompat/`）。状态：**已合**

本仓库 `responses_client_tools.go:220` 的 `dropInvalidLoweredFunctionItemID` 在把
`custom_tool_call` / `tool_search_call` 降级成 `function_call` 时**直接删掉** `ctc_*` /
`tsc_*` 的 item ID；还原时又不把上游回的 `fc_*` 换回 `ctc_*`。于是客户端历史里存着
`fc_` 前缀的 `custom_tool_call`，下一轮 replay 到会校验 ID 的上游就报
`Invalid 'input[N].id' ... Expected an ID that begins with 'ctc'`。

上游改成 `normalizeLoweredFunctionItemID` + `retypedResponsesToolCallItemID`：
保留后缀、只换前缀，双向映射；流式路径额外用 `clientItemID` 区分「上游用的 ID」与
「发给客户端的 ID」。这是 0.1.180 §6.1 工具桥接簇的同一族问题，但**不依赖那簇**，
产品文件独立可合。

### 3.8 `a6b11ccc` OpenCode Go 用量重置时长未解析（v0.1.182）

3 文件 +144-3，`ratelimit_service.go` 干净。状态：**已合**（用 OpenCode Go 订阅账号才有意义）

`parseOpenAIRateLimitResetTime` 只认 `usage_limit_reached` / `rate_limit_exceeded`
两种 `error.type`，OpenCode Go 订阅报的是 `GoUsageLimitError`，且重置时间只写在
人类可读的 message 里（`"Weekly usage limit reached. Resets in 2 days."`）。解析不到
→ 回落到默认冷却 → 账号被**提前**重新调度，撞回同一个 429。

新增两个正则 + `parseOpenCodeGoUsageLimitResetDuration`（支持 s/m/h/d/w 与
`"1h 30m"` 这种多段，带溢出保护）。自成一块，不碰调度。

### 3.9 `9fb26043` Grok 上游 User-Agent 用回官方 CLI（v0.1.181）

9 文件 +45-48，只有 `openai_gateway_grok_test.go` 冲突。状态：**已合**

两件事：
1. `openai_gateway_chat_completions_raw.go:175` 与 `grok_observed_models.go:95` 还在发
   占位 UA `sub2api-grok/1.0`（本仓库 `grok_upstream_headers.go` 里的
   `grokUpstreamUserAgent` 常量），上游把它们都换成 `defaultGrokUpstreamUserAgent()`
   并删掉那个常量。本仓库 `grok_upstream_headers.go:26` 已有该函数。
2. `internal/pkg/xai/billing.go:22` 的 `CLIClientVersion` `0.2.114` → `0.2.120`
   （对齐 <https://x.ai/cli/stable>）。

严格说第 2 点是维护性版本跟进而非缺陷，但和第 1 点在同一个 commit 里、且共用同一处常量，
建议整条合。

### 3.10 `eb594eef` 充值完成后余额不刷新（v0.1.182）

2 文件 +57-3，**全部干净**（纯前端 `views/user/PaymentResultView.vue`）。状态：**已合**

支付回调落地后页面不重新拉余额，用户要手动刷新才看到到账。simple mode 下支付页仍可达，
所以这条对本 fork 有效。

### 3.11 `e55727d4` 容量溢出被写成持久粘性绑定（v0.1.183）

2 文件 +51-2。`openai_gateway_scheduling.go` 需手改（本仓库同结构，行号偏移约 190）。
状态：**已合**

Layer 1 里粘性账号自身健康、只是等待队列满时，Layer 2 会**临时**借一个别的账号顶这一次
请求。但 Layer 2 成功后无条件 `setStickySessionAccountID(...)`，把这次一次性溢出**写成了
持久绑定** → 一阵短促突发就能把整段会话迁到 cache-cold 账号上，缓存命中率塌掉（真金白银）。

三处改动，对应本仓库：
- `openai_gateway_scheduling.go:932` 的 `selectAccountWithLoadAwareness` 开头加
  `stickySpillover := false`
- `:1029` 附近（Layer 1 的 `waitingCount < cfg.StickySessionMaxWaiting` 分支）置
  `stickySpillover = true`
- `:1185` 与 `:1224` 两处 `if sessionHash != "" && !gatewayProfitControlGateActive(ctx)`
  各加 `&& !stickySpillover`

⚠️ 与 3.1 的 `8e60d574` 同改 `openai_gateway_scheduling.go`（不同函数，不冲突），
也与 0.1.180 §6.3 的 `3fd66a33b`、§7.3 抢同一文件 —— 先后顺序要定。

### 3.12 `99ec347e` + `71aa6e35` Antigravity Sonnet 4.5 别名（v0.1.182，成对）

9+7 文件。`99ec347e` 全干净；`71aa6e35` 是对它的收窄，**两条必须一起**（单合前者会把
显式 `claude-sonnet-4-5` 也吞掉）。状态：**已合**

对本仓库 `domain/constants.go:105` 的 `DefaultAntigravityModelMapping` 净效果只有三行：

| key | 现在 | 合完 |
|---|---|---|
| `claude-sonnet-4-5` | `claude-sonnet-4-5` | `claude-sonnet-4-5`（不变，显式选择透传） |
| `claude-sonnet-4-5-thinking` | `claude-sonnet-4-5-thinking` | `claude-sonnet-4-6` |
| `claude-sonnet-4-5-20250929` | `claude-sonnet-4-5` | `claude-sonnet-4-6` |

外加 `account_test_service.go:2272` 的 Antigravity 连接测试默认模型
`claude-sonnet-4-5` → 常量 `defaultAntigravityTestModel = "claude-sonnet-4-6"`。

上游语义：兼容别名（`-thinking` / 日期后缀）迁到 4.6，**canonical 的显式 4.5 保持原样**。
其余全是测试文件跟着改。

---

## 4. P1 — 值得做，但要手工改

### 4.1 `49752060`（v0.1.182）+ `b20f29d1`（v0.1.183）渠道监控 v2 里 composite 分组的错误全部丢失

2 文件 +34-13 与 2 文件 +3-1。**两条必须合成一处改动**：`b20f29d1` 修的正是
`49752060` 写出的 `NULLIF(TRIM(a.platform))`（`NULLIF` 少了第二个参数，是硬 SQL 错误）。
状态：**待合**

缺陷：composite 只是路由层，`ops_error_logs.platform` 记的是 `composite`，而
`composite` 永远不是「已启用的 config platform」 ⇒ **composite 分组的错误被监控 v2 的
每一个查询过滤掉**，面板上看不到。用量侧早就用 `usageLogEffectivePlatformExpr` 解析到
真实账号平台了，错误侧没有。

修法：`channel_monitor_v2_classify_errors` 的取数 CTE 里 LEFT JOIN `groups` 与
`accounts`，`g.platform = 'composite'` 时改用 `a.platform`。

⚠️ **不能直接 apply**：上游那段是 PG 方言（`DISTINCT ON`、`date_trunc`、`jsonb_typeof`、
`id::text`），本仓库 `repository/channel_monitor_v2_aggregation.go:260` 的 `channelMonitorV2ClassifyErrorsSQL` 已经是**独立的
SQLite 重写**（`CREATE TEMP TABLE` + `ROW_NUMBER() OVER (PARTITION BY ...)` +
`strftime` + `json_valid`/`json_type`）。只搬语义：

```sql
    lower(CASE
      WHEN g.platform = 'composite' THEN COALESCE(
        NULLIF(TRIM(a.platform), ''),
        NULLIF(NULLIF(lower(TRIM(current_error.platform)), ''), 'composite'),
        'unknown')
      ELSE COALESCE(NULLIF(TRIM(current_error.platform), ''), 'unknown')
    END) AS platform,
...
  FROM ops_error_logs current_error
  LEFT JOIN groups g ON g.id = current_error.group_id
  LEFT JOIN accounts a ON a.id = current_error.account_id
```

只在用 composite 分组 **且** 开了监控 v2 时才有收益。

### 4.2 `4795650d` 错误日志把入站端点误报成上游端点（未发布）

8 文件 +46-0，3 个产品文件冲突。状态：**已合**（2026-08-27），见 [`resolve-pending-decisions-and-p1-fixes`](../../openspec/changes/resolve-pending-decisions-and-p1-fixes/)

`force_chat_completions` 生效时，请求实际发往 `/v1/chat/completions`，但 503 / 传输失败
这类**没有 `OpenAIForwardResult`** 的路径上，`handler.GetUpstreamEndpoint` 回落到入站端点，
错误日志和用量日志记成 `/v1/responses`。本仓库
`internal/pkg/openai_compat/upstream_capability.go:49` 有 `force_chat_completions`、
`openai_gateway_chat_completions.go:88/94/301` 有原生 CC 直转路径 ⇒ **同一缺陷成立**。

移植时两处要裁：

- 上游用了 `service.IsCNProvider(platform)`，本仓库无 CN 平台 ⇒ **删掉该分支**，
  只留 `PlatformOpenAI` / `PlatformGrok`。
- 上游用了 `shouldForwardOpenAIResponsesViaRawChatCompletions(account)`（它自己抽出的
  门函数），本仓库没有这个函数、门是内联在 `ForwardAsChatCompletions` 里的 ⇒
  按本仓库的内联条件写，或顺手抽同名函数。

其余照搬：新增 `ClearActualOpenAIUpstreamEndpoint`（failover 尝试之间清残留）、
`sendCCUpstreamRequest` 里每次发送都 `SetActualOpenAIUpstreamEndpoint(c, "/v1/chat/completions")`、
`GetUpstreamEndpoint` 优先取运行时端点。纯可观测性，行为不变。

---

## 5. 本轮不做

### 5.1 Responses Lite 并行工具调用簇（5 条）—— 缺 0.1.180 基座

`1563db3f`（v0.1.181）`53d76ad8` `d6012b0b` `d5e43ef7` `095b5253`（v0.1.182）。

上游这簇修的是「Lite 请求必须 `parallel_tool_calls: false`，否则 400
`unsupported_value`」的一串漏洞。**本仓库缺四个基座符号**，逐个 grep 零命中：

| 符号 | 用在 |
|---|---|
| `ensureOpenAIResponsesLiteParallelToolCalls` | `d6012b0b` / `095b5253` |
| `decodeOpenAIJSONUseNumber` | `d6012b0b`（0.1.180 §7.1 的 `openai_json_decode.go`） |
| `normalizeOpenAIResponsesLitePayloadForAccount` | `d5e43ef7` |
| `normalizeOpenAIParallelToolCallsWithoutTools` | `1563db3f` |

也就是说这簇挂在 0.1.180 §6.1（`7498d8fdc` Responses Lite 强制串行工具调用）与
§7.1（PR #5888 大礼包）上。**要做就先做 0.1.180 那两项**，届时这 5 条一起排期。
`d6012b0b` 是典型的「`apply --check` 过了但编译不过」——目标文件在，调用的函数不在。

### 5.2 `f1aadd48` OpenAI OAuth 配额耗尽 429 分类 —— 本仓库该缺陷不成立

4 文件 +118-14，全部冲突。上游修的是「5h/7d 配额打满的 429 被当成瞬时 429 继续同号重试」。

本仓库**没有那个同号重试窗口**：`shouldRetryOpenAIOAuth429OnSameAccount`、
`ShouldRetryOpenAIOAuth429`、`newOpenAIAccountFailoverError`、`openAIOAuth429RetryWindowActive`
逐个 grep **零命中**。`openai_account_runtime_block_fastpath.go` 里的
`markOpenAIOAuth429RateLimited` 是更早的形态：拿到 429 就直接
`BlockAccountScheduling(account, cooldownUntil, "429")`，reset 头也照样解析。

⇒ 语义上本仓库本来就是「一律按配额限流处理」，**不会**把配额耗尽的号留在同号重试里。
合这条毫无意义，而且要先把整套重试窗口搬过来。这条是「上游 fix 只对上游成立」的样本，
留在这里当反例。

### 5.3 Codex routed model catalog 簇（11 条）—— 未发布且仍在返工

`22e1b814` `e471be73` `3e98a5a1` `b16ed03c` `5a2f542a` `e39fce27` `fc589bce`
`db01fb98` `5934981e` `195b2197` `2abce650`（PR #5926，contrib）。

两个 feat 打头，后面挂 8 条 fix + 1 条 test，最新一条 `2abce650`「harden routed catalog
capability sync」还在 main 上没进任何 tag。`3e98a5a1`（composite 精确账号别名路由）
`apply --check` 6 文件全干净，很容易被当成独立 bugfix 顺手合 —— **别合**，它给
`CompositeRouteResolver` 加了 `CompositeModelOwnershipResolver` 回调并在
`NewGatewayService` 里接线，是这个功能簇的一部分。

**等上游打进正式 tag、返工停下来再整体评估。**

---

## 6. N/A

| 上游条目 | commit | 原因 |
|---|---|---|
| Kimi 并发 403 保持可恢复 | `3802268e` | 无 Kimi 平台。新增 `ratelimit_cn_providers.go` / `openai_gateway_cn_fixes_test.go` 本仓库不存在；与 0.1.179 §2 / 0.1.180 第 8 节结论一致 |
| composite 路由 Kimi Code K3 模型 ID | `4347e555` | 同上 |
| VERSION 同步 0.1.180 / 0.1.181 / 0.1.182 / 0.1.183 | `03e8ab41` `e2d9b823` `aa2c4e8d` `7634e3c2` | 本仓库用自有编号 `1.1.8` |
| sponsors 素材 | `3b7753a8` `66d664ff` `6ca1e15b` | 上游 README / 素材专属 |
| `d8694f03` WSv2 陈旧原生工具 ID 清理的回归测试 | `d8694f03` | 纯测试，对应的产品改动在 0.1.180 §7.1 里，本轮无处落地 |
| `5934981e` routed catalog 能力交集测试 | `5934981e` | 属 5.3 的簇 |

---

## 7. 建议顺序

```text
① 3.1  8e60d574          Codex session-id      —— 4 文件全干净，收益最大，先做
② 3.2  bc4a9ae4          cache 重复计费        —— 涉钱，第二个做
③ 3.3–3.10 一批干净项    gemini / antigravity clamp / 邮箱换绑 / 图片 prompt /
                         item ID / OpenCode / Grok UA / 支付余额
④ 3.11 e55727d4          粘性溢出（3 处手改，注意与 0.1.180 §6.3 抢同一文件）
⑤ 3.12 99ec347e+71aa6e35 Antigravity Sonnet 别名（成对）
⑥ 4.2  4795650d          真实上游端点（裁掉 IsCNProvider）
⑦ 4.1  49752060+b20f29d1 监控 v2 composite（SQLite 重写，用 composite 才做）
不做：Responses Lite 簇（等 0.1.180 §6.1/§7.1）、f1aadd48、Codex routed catalog 簇、Kimi
```

本文 P0 与 PORTING-0.1.180.md 的 P0 **无文件冲突**，可以并行。唯一要协调的是
`openai_gateway_scheduling.go`（本文 3.1 / 3.11 vs 0.1.180 §6.3 `3fd66a33b`、§7.3）。

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

- **计费**：`go test -tags=unit ./internal/service/ -run 'TestBilling|CacheCreation' -count=1`
  （3.2 改了 `computeCacheCreationCost` 与两条 SSE 用量解析路径）
- **调度 / 粘性会话**：`-run 'Sticky|SessionHash|LoadAwareness'`（3.1 + 3.11 同一文件）
- **登录 / 邮箱**：3.5 动了 `user_repo.go` 与 `auth_email_binding.go`，跑
  `-run 'EmailBind|EmailAlias'`；顺带确认登录与 `user_allowed_groups` 没被带坏（缺表会 503）
- **前端**：3.10 是 `PaymentResultView.vue`，`pnpm exec vitest run
  src/views/user/__tests__/PaymentResultView.spec.ts`
- 全量 `pnpm run test:run` 已知有 1 个**与移植无关**的失败文件
  （`src/composables/__tests__/useRoutePrefetch.spec.ts`，5 条），见 PORTING-0.1.179.md §4.6
