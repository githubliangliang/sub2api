# 移植评估：上游 v0.2.1 / v0.2.2 / v0.2.3 / v0.2.4

评估与实施日期：2026-09-12。**第一档 13 项与第二档 12 簇已落地 main，最终构建和总检查通过。**
本仓库基线为 `c9d9bebe87926791bfdd716f59abcba413fee20a`（`v1.1.11`）。
实际交付 **13 项第一档**（12 项本轮增量 + 1 项旧轮次重新判定）与 **12 簇第二档**。
第三、四档仅记录建议，未立实施 change；不执行整仓 merge。

最终结果、测试、运行检查和集成修正见 [implementation.md](./evidence-0.2.4/implementation.md)。
下文四态和工作量数字保留评估时口径；source-baseline 不改写。主线交付包含后续 API Key Responses namespace 修复，验证详见最终验收第 5 节。版本仍为 1.1.11，迁移仍为 226，本次未创建新 tag 或 Release。

## 1. 版本与覆盖范围

| 点 | 冻结值 / 说明 |
|---|---|
| 上次评估终点 | `v0.2.0` = `aa236488351eb71e120fc2b6fb32e36b0374c918` |
| [v0.2.1](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.1) | `578785ee7fb35030b094b69624efe25670a36f5f`；09-05 发布；相对前一 tag 48 条非 merge |
| v0.2.2 | `5485f368b29d05adb95a00f71801c7c23d8f48af`；存在 tag，当前 Releases 列表未列出对应 Release；57 条非 merge |
| [v0.2.3](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.3) | `8fa67d477d6651a744754392a8982ea589c26ae6`；09-08 发布；6 条非 merge |
| [v0.2.4](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.4) | `5de5e2bed035d43591a2e10e51f420ef6a84eb98`；09-09 发布；40 条非 merge |
| 拉取时 upstream/main | `4726bdd08b6201d426a80529b79be123a4008d20`；正式版之后另有 57 条非 merge，**不计入本轮建议批次** |
| 本仓库 | `VERSION=1.1.11`；SQLite-only；可使用进程内 miniredis；最大迁移 `226` |
| 工作树附加改动 | `openai_gateway_forward.go`、`openai_ws_http_bridge.go` 和两个未跟踪的 `openai_responses_input_metadata*.go`；属于用户已有改动 |

本轮以 **tag 的祖先关系**分版本，不以作者日期或发布说明出现的位置分版本。
151 条非 merge 提交归入 114 个第一父链候选（PR 整体或独立主线提交），恰好覆盖一次。
按候选计：第一档 12、第二档 36、第三档 53、第四档 13。第二档再按行为归并为 12 簇。
这些数字不是“可直接 cherry-pick 的提交数”，PR 中混入的无关改动仍要剔除。

### 1.1 四态统计与证据

共 1,044 条 **候选 × 文件**记录，重复文件在不同 PR 中分别计数：

| ALREADY | CLEAN | CONFLICT | NOBASE |
|---:|---:|---:|---:|
| 5 | 575 | 398 | 66 |

- [commits.tsv](./evidence-0.2.4/commits.tsv)：全部 151 条提交，完整 SHA。
- [candidates.tsv](./evidence-0.2.4/candidates.tsv)：114 个候选的发布归属、档位、批次、四态计数与理由。
- [files.tsv](./evidence-0.2.4/files.tsv)：逐文件四态，可回查任何候选。
- [baseline-tests.md](./evidence-0.2.4/baseline-tests.md)：当前基线的实际验证结果；不是移植完成证据。

## 2. 核查方法

### 2.1 来源与四态

执行 `git fetch upstream --tags --prune`；GitHub Releases API 核对正式发布列表，Git 对照代码。
每个 PR 使用 `git diff <merge>^1 <merge>`，不按功能内部微提交逐个下结论。

四态使用独立临时 index，先 `GIT_INDEX_FILE=<temp> git read-tree HEAD`，对**全部候选先做反向检查**，
再做正向检查；`git apply --cached --reverse --check` / `git apply --cached --check` 均不应用补丁。
旧文件不存在且不是新文件添加时记 NOBASE。新文件可创建时记 CLEAN。
因此四态针对上述**已提交基线**，不把用户工作树修复误算为已合入。

符号、调用路径、SQLite 和前端联动再人工核查。第一档的“缺陷成立”是本地可达代码及触发条件成立，
**不声称已经在用户生产流量中观测到事故**。实施时需用失败测试锁住对应触发条件。
ALREADY 只证明该文件的反向补丁可应用；不证明整个功能已存在。CONFLICT 也不等于需要手工修复。

### 2.2 迁移编号

第一、二档建议范围**不需要新增迁移**，保留本仓库 `226`。上游本轮新增 7 个 SQL 文件：

| 上游迁移 | 所属功能 | 本轮判断 |
|---|---|---|
| `232_add_usage_log_upstream_request_id.sql` + `233_add_usage_log_upstream_request_id_index_notx.sql` | 上游请求 ID | 第三档；若以后做，自 `227` 起分配，索引不能照抄 PG CONCURRENTLY |
| `234_channel_max_reasoning_effort_multiplier.sql` | Anthropic effort 定价 | 第三档 |
| `234_group_codex_models_manifest_config.sql` | 固定账号模型清单 | 第三档；与上一项在上游重号，不继承其编号 |
| `235_group_model_allowlist.sql` | 分组模型访问限制 | 第三档；需 schema + Ent + 全入口授权核查 |
| `236_group_model_allowlist_repair.sql` | 旧结构自愈 | 第四档；本地没有该功能，PG `DO` 块不能执行 |
| `237_add_minimax_platform.sql` | MiniMax | 第四档；当前缺平台基座 |

以后选择上述功能时按**实际落地顺序**分配新号，不能把历史文档里的预留号当作已占用号。

## 3. 第一档：确认存在且改动较小

对应 change：[port-upstream-0.2.4-p0-fixes](../../openspec/changes/port-upstream-0.2.4-p0-fixes/)。
下表 C/X/N 分别为 CLEAN/CONFLICT/NOBASE，完整 ALREADY 数在证据表；本节本轮候选均无 ALREADY。
路径以 `backend/` 为根。每项实施时产品代码与相应回归测试一起移植。

| ID | 来源 / C-X-N | 本地缺陷、patch site 与符号核查 |
|---|---|---|
| F01 | #6372 `e2fd418a9`，1/3/0 | `service/ratelimit_service.go:1252` 的 `calculateOpenAI429ResetTime` 在两个窗口未满时仍取最大 reset；`openai_account_runtime_block_fastpath.go:142` 在 fallback 关闭时仍默认 Block。37% 用量 + 一周 reset 也会被停调。删除“未满取最大”分支，并让关闭 fallback 的情况不建默认冷却。`get429FallbackCooldown`、`BlockAccountScheduling` 在位；保留 fork 的 Spark shadow 早退。**不要引入未合的整套 quota classifier/retry-window 政策**。 |
| F02 | #6536 `83094abf2`，3/2/0 | `handler/openai_gateway_handler.go` 的 Responses/WS 与 `openai_chat_completions.go`、`openai_embeddings.go` 调度仍传请求模型，渠道映射后的真实模型未参与选号。`ChannelMappingResult`、`ResolveChannelMappingAndRestrict` 在位；PR 新增 `openAIChannelForwardModel` 自包含。只改选号模型，保留请求模型的响应/计费语义。剔除同 PR 的外网 token 测试调整。 |
| F03 | #6492 `d8b9a83f8`，1/1/0 | `service/gateway_scheduling.go` 的负载感知 routed/sticky/普通候选缺逐账号渠道限制。`gateway_forward.go:934/954` 已有 `isUpstreamModelRestrictedByChannel` 与 `needsUpstreamChannelRestrictionCheck`；补三层过滤，不能只改 fallback 选号。 |
| F04 | #6531 `8f1d6af3e`，2/0/0 | `service/openai_messages_continuation.go:107` 未识别 `previous_response_id is not available for this user`，现有恢复分支无法触发。仅补这个明确错误，既有状态码边界不放宽。 |
| F05 | #6529 `7b271bbee`，1/1/0 | `service/openai_gateway_request_body.go:54` 的 `shouldPreserveOpenAIResponsesNoneReasoningEffort` 未考虑透传，可能删去显式 none。`IsOpenAIPassthroughEnabled`、`IsOpenAIOAuthLike` 在位；新增透传早退，不改变非透传规则。 |
| F06 | #6702 `e274de45b`，2/0/0 | `service/gateway_messages_cache.go:159` 的 `mustJSONString` 用 Go `%q` 生成 JSON，控制字符可得到非法的 `\x7f` 等转义。改标准库 `json.Marshal(string)`，覆盖控制字符、引号和正常文本。 |
| F07 | #6640 `578785ee7`，4/0/0 | `gateway_billing_header.go`、`gateway_upstream_request.go`、`gateway_count_tokens.go` 按缓存 UA 写 billing，后续 mimicry 又覆盖出站 UA；版本和 fingerprint 后缀可能不一致。`computeClaudeCodeFingerprint`、`ExtractCLIVersion`、`claude.DefaultHeaders` 在位；按最终出站 UA 同步版本及后缀，覆盖 messages/count_tokens。 |
| F08 | #6838 `4e9b01fd5`，4/0/0 | `handler/gateway_helper.go` 复用 ParsedRequest 时没带 max_tokens；`service/claude_code_validator.go` 只对旧 Haiku 探测放行。补字段与 `isMaxTokensOneBody`；UA 验证仍在探测放行之前，普通请求仍严格校验。 |
| F09 | #6677 `2cc1e7ef9`，3/0/0 | `pkg/claude/constants.go` mimicry beta 缺 thinking-binding；`service/gateway_request.go` 没有相应 body sanitize。`stripAnthropicBodyFieldUnlessBeta` 已有且已用于 fallbacks。补 token，并按最终 header 决定保留或删除 `thinking.block_binding`。 |
| F10 | #6484 `c0420e2b8`，2/0/0 | `service/ratelimit_service.go` 已处理 Fable 7d_oi，但没有 credits_required 的模型级分支，可能误停 Sonnet/Opus。`isAnthropicFableModel`、`parseAnthropicResetTimestamp`、`firstRequestedModel`、`SetModelRateLimit` 在位；新增函数随补丁带入，写库失败也不能扩大为整个账号限流。 |
| F11 | #6674 `f7c48ba59`，2/0/0 | `pkg/antigravity/request_transformer.go` 在 reasoning 无工具时跳过 toolConfig。已有 `GeminiToolConfig` 和 `hasMixedToolInvocations`；总是构造 toolConfig，保留上一轮已落的混合工具开关。 |
| F12 | #6594 `570bd084c`，2/0/0 | `repository/ops_repo_metrics.go` 将有效的零延迟/零连接数送给把零视作空的 `opsNullInt`。新增两个 metric 专用 nullable helper，区分 nil 与 0；不全局更改 ID 等字段共用的 helper。SQLite 参数和列保持不变。 |
| F13（旧项重判） | `e21b849a926f5f30683ac9a421589f1cf37a4c51`，1/2/0（工作树检查） | 0.2.0 §5.4 原来“缺 helper，未核路径”。已追通：API Key + Responses 支持 + 关闭 passthrough 会到 `openai_gateway_forward.go:339`，无 instructions 时无条件合成 Codex 提示。透传分支已有 OAuth 守卫。上游 `UsesOpenAICodexProtocol` 仅 3 行、依赖已存在的 `IsOpenAIOAuthLike`，因此**明确提升为第一档**。只约束该合成点，不全仓替换账号谓词。 |

第一档顺序：F06/F04/F05 → F02/F03 → F01/F10 → F07/F08/F09/F11/F12 → F13。
F01 是当前最值得优先修的可用性问题；先合独立小项只是便于逐步验证。

## 4. 第二档：收益明确，需要按行为移植

对应 change：[port-upstream-0.2.4-p1-stability](../../openspec/changes/port-upstream-0.2.4-p1-stability/)。
每簇独立验收；列出的 PR 是来源边界，不授权整 PR 覆盖文件。

| ID | 来源 / C-X-N | 价值与实际工作 |
|---|---|---|
| S01 | #6397 `91cf66037`，6/4/0 | 10 文件 +1208/-77。长会话回放从反复复制改为共享不可变正文，并记住已被拒绝的 encrypted_content。对 1G 内存很有价值，但引入 `unsafe.String` 后必须证明底层字节不再原地改写。触及用户已有的 forward/http_bridge 改动，按语义保留 metadata 清理。 |
| S02 | #6434 `43569bb44` + #6754 `b63844523`，5/1/0 | WS 客户端取消的归因/关闭顺序 + Chat Completions 在关闭 Body 前取消上游。`detachUpstreamContext`、`forwardAsChatCompletions`、v2 relay 都存在；测试需确认取消不会挂住、不会把客户端退出计为账号故障，已产生用量仍落库。 |
| S03 | #6320 `6f0d0abab`，3/1/0 | 持久冷却已清除后，丢弃陈旧进程内 block。锁、generation、deadline、三个持久时间字段均在；新增 peek/CAS 随簇带入。**风险是 fail-open**：快照落后或写库失败时也可能清本地 block。放第二档，须验证本 fork 的 snapshot/hydrate 时序、并发新 block 和模型级冷却，不能简化成无条件 Delete。 |
| S04 | #6510 `9517f12cd`，7/0/0 | 转发失败即时释放 Anthropic 会话槽。`checkAndRegisterSession` 已运行；新增 `ReleaseAccountSession` 与 `SessionLimitCache.UnregisterSession`，后者实现是现有 Redis ZSET 的 ZREM，miniredis 同样适用。改 interface 必须覆盖所有 mock。成功/已计量的中断仍保留原会话语义。 |
| S05 | #6593/#6629/#6553/#6539/#6581，9/5/0 | 已发现工具不能在 Chat fallback 丢失；done 事件补全工具 arguments；heartbeat/历史 delegation 与 OpenCode session 兼容。`EffectiveResponsesTools`、`promoteResponsesToolSearchDiscoveries`、Responses accumulator 和旧 bootstrap parser 都在。新 promotion helper 由同簇提取。按发布终态整合，保留 namespace、自定义工具、重复字段拒绝边界。 |
| S06 | #6628/#6572/#6620/#6718/#6678/#6743/#6690，36/29/1 | Astra 别名、兜底定价、上下文/工具能力同步、prompt cache、Ultra 元数据、instructions 和 OAuth reasoning mode 应一起评估落地。仅加模型常量不够。已有 metadata snapshot/Extra 持久化基座，缺 `isOpenAIGPT6AstraModel`、`CodexToolCapabilities` 等由本簇引入；**不需要 Ent 迁移**。不顺带合固定账号清单、Ultrafast service_tier 或分组 allowlist。 |
| S07 | #6535 `78e1aaedb` + #6626 `63dc24b5e`，7/2/0 | 已有 `pricing.override_file` 加内容哈希热重载；现有 reasoning mapping 支持 none 来源。重载 PR 共 3 文件 +381/-21，保留上一次有效价格，处理文件删除/损坏和并发一致性。来源 none 与转发目标 deny 是两项不同功能，本批不带 deny 策略。 |
| S08 | #6811/#6816/#6815/#6810/#6836，29/3/0 | 代理局部更新不应清空有效期/备用代理；允许共享有向 backup 引用、重复过期回退；校验日期与列表响应。`dto.NullableInt64Field`、proxy 字段和 repository 已有，无新迁移。联合 handler/service/import/batch/UI 处理 omitted/null/value，不能只改表单。代理列表错误时保留现有数据。 |
| S09 | #6424 `95acbf1f0`，17/1/0 | 默认不将每条 http.access 写数据库，system log 按独立保留期清理，适合 SQLite 小机器。已有 OpsCleanup 与 sink 真的启动，不能以 SQLite 为由禁用服务。缺 `defaultOpsAdvancedSettingsForConfig`，上游仅 7 行，是 `defaultOpsAdvancedSettings` + cfg.Cleanup.Enabled 覆盖；随此簇补齐。保留本 fork 清理 SQL 和写失败退避。 |
| S10 | #6557/#6580/#6659/#6798/#6821/#6764/#6819/#6705，19/9/1 | 紧凑账号列表降低返回量，保留详情编辑；修复活动子路由折叠、移除停用分组、菜单越界、OAuth plan 丢失、刷新失败选择状态、注册入口显隐，以及用量过滤器漏加载分页密钥。紧凑列表 PR 8 文件 +607/-33，必须用真实组件测试核验编辑、批量操作和全选，不让缺字段写回空值。 |
| S11 | #6606 `9b4fcfb89` + #6481 `787a6a33d`，7/3/1 | Claude CLI 版本可配置且已有指纹能抬升，解决“只升级内置常量但老账号无效”。缺 `claude.CLIVersion()` 的文件由 #6606 新增；`extractProduct`、`isNewerVersion` 在位。还需按终态核查请求体所有版本写入点，避免 env/header/body 分裂，版本只能合法且不低于内置下限。 |
| S12 | #6814 `1923d1c27`，3/1/0 | go-redis `v9.17.2` → `v9.22.0` 及两个 ZRangeArgs 调整。**Redis optional 不代表无 Redis 客户端**，miniredis 模式也用 go-redis 连接池。只更新此依赖及必要传递依赖，保留 SQLite driver；以 cache/session/queue 用例验证，不能拷贝上游 go.mod。 |

### 4.1 新调用与缺基座的具体处置

- S01 的 `openAIWSPayloadStringView`、`openAIWSIngressSessionHashContextKey`、lineage store 方法全部由 #6397 自带；`sanitizeEncryptedReasoningInputItem`、`decodeOpenAIJSONUseNumber`、`marshalOpenAIUpstreamJSON`、`ensureBindingCapacity` 已存在。不可只拷贝新增 lineage 文件。
- S06 的 NOBASE 包含 #6743 混入的 `redeem_admin_fulfillment_test.go`，它不属于 Astra，直接排除；#6743 的兑换测试变量改名和外簇断言不应带入。保留本地 routed catalog，按终态映射 Astra 改动。
- S10 的 `AccountsView.lite.spec.ts` 先在 #6557 添加，#6798 才修改；NOBASE 是依赖顺序，不是额外功能。前端采用本仓库现有 DTO、Icon 与组件接口。
- S11 的 `cli_version_test.go` 同样先由 #6606 添加，#6481 的 NOBASE 随正确顺序消失。新 `semver` 调用使用已在 go.mod 的 `golang.org/x/mod`。
- S03 是风险最高的调度变更，S01 是风险最高的内存所有权变更。两者均需 race/并发回归；CLEAN 数量不降低这个要求。

## 5. 第三、四档及旧轮次遗留

### 5.1 本轮按需项目

| 功能簇 | 建议与判断依据 |
|---|---|
| 固定账号 Codex 清单、混合账号目录、live 连接测试模型 | 第三档。#6602 → #6662 → #6688/#6772 是一条链，有分组配置迁移。后两者不能作为完全独立小修直接合。 |
| simple mode 基础分组、分组 model_allowlist | 第三档。前者改变本 fork 的分组入口约束，后者改全入口模型授权；需要单独确认个人使用方式。不能为合 v0.2.3 的 repair 先引入这套功能。 |
| Fast/Ultrafast、effort 定价/deny、区间定价 | 第三档。目录显示 priority 已有，发送侧 Fast 未合；不要把模型 Ultra reasoning 元数据与 Ultrafast service_tier 混为一谈。新增策略需先决定计费语义。 |
| 上游 request ID、详细代理归因、WS cyber-policy 日志 | 第三档。排障价值存在，但涉及两条迁移或大量当前没有的转发路径；按实际 ops 需求取子集。 |
| 图片 URL 回填、Image 2.5 / OAuth 生图 | 第三档。前者引入新的主动下载面，必须带上同 PR 的私网拒绝/嗅探修复；本 fork 不能因此被判已有该 SSRF。后者在未发布 main 上已有原生 Codex Images 后续重构，常用生图时单独冻结终态。 |
| Grok media、额度资格、external_web_access | 第三档，与旧 Grok 稳定性尾项一起按本地自有实现核查。不是看到一个字段删除就整文件覆盖。 |
| Ollama/GLM/DeepSeek、Gemini 自定义列表/监控、支付/订阅、展示小项 | 第三档。Ollama/GLM/DeepSeek 可经 OpenAI-compatible 路径使用，**没有专属平台不等于没有收益**；兼容路径可拆，CN 原生路径不可照搬。其它条目受功能是否使用约束，逐候选见证据表。 |
| 长流 HTTP/2 PING #6281 | 第三档，不能当成已接好的通用保活直接拿。本地 OpenAI H2 已有 15s/15s PING；上游新增 LongStream profile 并改为 10s/5s，但 `git grep HTTPUpstreamProfileLongStream v0.2.4 -- ':!*test.go'` 只有定义、getter 与消费者，**没有生产 setter**。若要全平台收益需补调用链与代理实测，调阈值本身也有误断风险。 |

### 5.2 不适用

第四档：PostgreSQL backup advisory lock、group schema repair、跨实例 channel invalidation、缺失的插件 ZIP 实现、MiniMax 整平台接线、上游 VERSION/sponsors。MiniMax 仅注册就涉及 109 文件及大量 CN 平台服务，当前不是独立小增量。

#6687 修的是不存在的 `openai_responses_input_compat.go` 规范化路径（上游终态 86 行）。
本地用户未提交的是 **input content item metadata 清理**，文件名和修复语义均不同，不将两者记为 ALREADY。
未来若确有 named standalone input 需求，应沿本地解析入口另行评估，不能只添加上游测试文件。

### 5.3 老轮次重新盘点

| 旧来源 | 当前剩余 / 本次复测 |
|---|---|
| 0.2.0 §5.4 | API Key instructions **提升 F13**：已确认可达路径，helper 3 行；不再沿用“缺基座”等待。 |
| 0.2.0 §5.1、0.1.180 §7.2、0.1.184 §5.2 | Fast 发送/分组/计费仍未合；本地有 priority 目录展示及 fast→priority 计费别名，不代表整个 Fast 功能存在。是否启用仍取决于使用方式。 |
| 0.2.0 §5.2、0.1.180 §7.5 | 动态长上下文阶梯仍未合。旧 `6466978d2` 整条本次量为 6 文件 +1200/-35、5 CLEAN/1 CONFLICT；历史“573 行”是当时核心基座口径，不能当整簇大小。当前上游 `billing_token_cost_request.go` 已拆成 62 行，但单拿它仍缺 schedule/pricing 调用链。 |
| 0.2.0 §5.3/§5.5/§5.6 | CN 平台、PG 启动重试、缺失 `classifySelectionFailureError` 的修复依然不直接合；兼容第三方模型时按本轮 §5.1 分开判断。 |
| 0.1.184 §5.5/§5.6 | usage compaction/requested effort 两字段、每用户公共分组限制仍需迁移/授权政策；单用户部署收益低。 |
| 0.1.184 §11.6 | 订阅重置锚点、DeepSeek 峰谷价、图像工具冷却、Grok/WS passthrough 边缘、支付仍按需；§3/§4 已落主体、Codex routed catalog、Antigravity 混合工具不重复列为待合。 |
| 0.1.180 §6.1 | `e2d9ce0ca` + `fbc9ee626` 非法工具参数、`31d5b67ba` 命名空间别名、`7a09a2eaf` deferred 标记仍未整体落地。本次四态分别 A/C/X/N = 0/3/3/0、1/0/4/0、0/5/1/0、1/1/4/0；文件都存在，不能再称缺整支。它们仍是旧工具桥接待办，与 S05 有同文件交叉，实施时决定先后并一并回归。 |
| 0.1.180 §6.1 的 `7498d8fdc` | 原补丁反向不匹配（0/0/4/0），但它的 Responses Lite 串行语义已由 0.1.184 §18 落地。**语义已合，不能凭 CONFLICT 再立一遍项**。 |
| 0.1.180 §6.2(c)/§7.1、0.1.184 §5.4/§5.8 | Grok 稳定性、WS passthrough、会话抢占仍未整体合；本轮 pending turn/HTTP bridge isolation/later-turn quota 等修复与其相关，继续按需，不静默引入多实例租约。 |
| 0.1.180 §7.3/§7.4 | 自动重置卡、模型列表读取上限配置仍按需；不是这轮小修的隐含前置条件。 |
| 0.1.180 §5.1 / 0.2.0 第二档阶段 9 | DOMPurify **仍未升级**：package.json 为 ^3.3.1，lock 中实际还有 3.3.1/3.3.3；xlsx 仍 0.18.5。延续原 OpenSpec 待办，不另建重复 change。旧记录的 advisory 数和“最新 3.4.14”是 09-02 的快照，本次不当作当前审计结果，也不声称 PoC 成立。实施前重新 pnpm audit，连同 lockfile 更新并回归净化调用点。 |
| 0.1.183 / 0.1.179 / 0.1.176 | 已落事项保持已落；不因源补丁上下文变化重新计入工作量。 |

## 6. 已有能力与“已合”的边界

- 5 个文件级 ALREADY：#6658 的四个既有测试文件与 #6235 的 go.sum。**没有任何本轮功能 PR 全文件 ALREADY**。
- 0.2.0 第一、二档主体已在 `c9d9bebe8` 合入 main；老 change 中“未合回 main”的文字是当时的实施记录，本轮不重写其冻结基线。
- 本地已有 OpenAI H2 PING、Responses Lite 串行、Fable 5.1、动态 override 文件读取、渠道 1h cache 价；本轮要补的是各自后续行为，不能重复报告为“从零新增”。
- 用户 metadata 修复不在 HEAD，且本轮未改变它；后续 S01/F13 进入对应文件时必须合并保留。

## 7. 建议落地顺序

- [x] 第一档 change：先独立小修，完成 F01/F02/F03 的可用性修复后再收尾 F13。
- [x] S11 Claude CLI 身份链；与 F07 联合验收 header/body/缓存版本。
- [x] S05 工具与 bootstrap → S01 回放；保留用户工作树 metadata 修复及旧工具桥接约束。
- [x] S02 取消与 S04 会话槽 → S03 冷却同步；S03 单独压测和审查 fail-open 边界。
- [x] S06 Astra 终态；复跑目录、能力同步、推理、续聊、计费和前端模型选择。
- [x] S07 定价重载/none 来源 → S09 日志 → S08 代理 → S10 账号界面 → S12 客户端依赖。
- [ ] 单独完成旧 DOMPurify 待办；第三、四档经功能决策后才另立 change。

上述第一、二档队列已完成，详见两份 change 的 tasks/verification 及最终验收记录；旧 DOMPurify 是单独待办，本次未实施。

## 8. 自测与验收

本轮评估阶段已核实测试存在、能编译并运行，详见 [基线证据](./evidence-0.2.4/baseline-tests.md)。
当前测试通过只证明可作为起点；例如 `TestCalculateOpenAI429ResetTime_NeitherExhausted_UsesMax` 还在验证旧行为，F01 实施时必须改为新契约并先看到失败。

移植阶段最低门禁：目标失败场景的前后对照、相关 package unit、后端 build、SQLite 方言审计；
触及前端的簇跑真实组件 vitest/typecheck，依赖变动保留 pnpm lock；S01/S03/S04 跑相关 race 测试。
`backend/internal/repository/migrations_schema_integration_test.go` 首行是 `//go:build integration && postgres`，
**不能引用为 SQLite 验收证据**。本轮也没有运行 PostgreSQL integration 或生产流量测试。

## 9. 本轮新增教训

1. **缺 Release 不等于缺 tag。** v0.2.2 的 57 条增量必须纳入，发布归属按祖先而非作者日期。
2. **新增 profile 要找 setter。** #6281 的新枚举、读取与 transport 分支都存在，仍没有生产接入点；现有 OpenAI PING 也不能报成首次支持。
3. **临时 index 冻结四态，工作树另记。** 避免把同事未提交的修复当成 HEAD 已吸收，更不能拿它抵消整 PR。
4. **“缺 helper”要量到函数体。** F13 是 3 行，S09 是 7 行；但小 helper 不构成扩大功能范围的理由。
5. **profile/账号字段不等于功能开关已接线。** miniredis 仍经过 go-redis；Astra Ultra reasoning 与 Ultrafast service_tier 是不同链路。
6. **共享字节和 fail-open 属于行为风险。** 四态干净不能替代内存所有权、CAS、数据库与快照时序的测试。
7. **PR 内混入其它测试也要拆。** #6743 的兑换测试、#6536 的外网测试调整不是对应功能依赖。
8. **旧安全审计日期必须保留。** 不把之前的 advisory 数和目标版本说成今日核验，也不把新增可选功能的安全修复说成既有漏洞。
