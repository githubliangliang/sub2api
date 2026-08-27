## Why

上游 v0.1.181 / v0.1.182 / v0.1.183 三个版本全是 bugfix 版（无新迁移、无新功能）。对
`v0.1.180 (c40edb4) .. main (efb46db0)` 的 44 个非 merge commit 逐条按文件试应用并 grep 依赖符号后，
确认 **12 项修复对应的缺陷在本仓库同样存在，且 patch site 与上游字节一致或仅有尾部上下文差异**。

其中两项在持续造成实际损失，不是理论缺陷：

- `gateway_anthropic_passthrough.go` / `gateway_upstream_response.go` 用 `Exists() && v > 0` 写
  `cache_creation` 的 5m/1h 明细。流式里后续事件报 `ephemeral_5m=0` 时 **0 被 `> 0` 挡住不覆盖**，
  旧值残留，最终 `5m + 1h > cache_creation_input_tokens` ⇒ 缓存创建费**超收**。
- Codex CLI 发送的粘性会话头是 `session-id`（连字符），本仓库只读 `session_id`（下划线）。Go 的
  `Header.Get` 把两者规范化成不同的键，该头**完全落空** ⇒ 客户端重连后会话漂移到别的账号；
  同一路径上另有「一次性容量溢出被写成持久粘性绑定」，一阵突发就能把整段会话迁到 cache-cold 账号。
  两条叠加直接吃掉 prompt cache 命中率。

余下 10 项是上游请求兼容性（Gemini 工具 schema、Antigravity token 上限与模型别名、Grok 上游 UA、
Responses 自定义工具 item ID）、账号冷却精度（OpenCode Go 重置时长）、邮箱换绑的 alias 查重与并发
守卫，以及充值完成后前端余额不刷新。

本变更只做移植，**不跟进上游的功能簇**：Responses Lite 并行工具调用簇缺 0.1.180 的四个基座符号、
`f1aadd48`（OAuth 配额 429 分类）修的缺陷在本仓库语义下不成立、Codex routed model catalog 簇尚未
进入任何 tag 且上游仍在返工。判定依据见 `source-baseline.md` 与
`docs/upstream-sync/PORTING-0.1.183.md` 第 5 节。

## What Changes

- Anthropic `cache_creation` 明细解析改为「字段存在即覆盖」，并在计费层新增聚合值封顶归一化，
  使 `5m + 1h` 永不超过 `cache_creation_input_tokens`。
- OpenAI/Codex 粘性会话识别 `session-id` 与 `session_id` 两种头名，`SessionSource` 区分来源。
- 粘性账号因等待队列打满而发生的一次性容量溢出，不再被写回持久粘性绑定。
- Gemini 工具 schema 清理补上 `deprecated`，并把非字符串 `enum` 值归一为字符串、无法归一时整体删除 `enum`。
- Antigravity 兼容模式把客户端 `max_tokens` / `max_completion_tokens` 封顶到上游上限 64000。
- Antigravity 旧 Sonnet 兼容别名（`-thinking`、日期后缀）迁到 4.6，**显式 canonical `claude-sonnet-4-5`
  保持透传**；连接测试默认模型改为常量化的 4.6。
- OAuth 图片生成请求下发「逐字使用用户 prompt」的 `instructions`，模型不再改写画面描述。
- Responses 自定义工具 / tool-search item ID 在协议降级与还原之间**保留后缀、只换前缀**双向映射，
  流式路径区分「上游 ID」与「发给客户端的 ID」。
- 账号限流解析识别 OpenCode Go 的 `GoUsageLimitError`，并从人类可读 message 里解析重置时长。
- Grok 上游请求统一使用官方 CLI User-Agent，删除占位常量 `sub2api-grok/1.0`，`CLIClientVersion`
  跟到 `0.2.120`。
- 邮箱换绑在字面地址查重之外增加 provider alias 查重，并把查重与写入收进同一把 key 锁内的事务。
- 充值成功回调落地后前端主动刷新余额。
- 不新增迁移、不改 schema、不动 wire 图、不改任何默认配置项、不改对外 API 契约。

## Capabilities

### New Capabilities

- `anthropic-cache-billing-integrity`: 流式与非流式下 `cache_creation` 明细的覆盖语义，以及明细与聚合值矛盾时的封顶归一化不变量。
- `openai-session-stickiness`: 显式会话头的识别集合、来源标记，以及容量溢出不得污染持久粘性绑定。
- `upstream-request-compatibility`: 出站请求在 Gemini / Antigravity / Grok / Responses 工具协议上必须满足的上游约束。
- `account-cooldown-accuracy`: 从上游 429 载荷推导账号冷却截止时间的解析覆盖面与溢出安全。
- `email-identity-binding-integrity`: 主邮箱换绑的收件箱唯一性判定与并发窗口关闭。
- `payment-balance-visibility`: 充值完成后用户可见余额的刷新时机。

### Modified Capabilities

无。`openspec/` 下当前没有已发布的 capability 基线（仅有 `changes/`），因此本变更以 ADDED
Requirements 形式固化「移植后应当成立的行为」，不声明对既有正式需求的修改。

## Impact

- **后端**：`internal/service/`（8 个文件）、`internal/repository/user_repo.go`、
  `internal/pkg/apicompat/responses_client_tools.go`、`internal/pkg/xai/billing.go`、
  `internal/domain/constants.go`。新增 `math` import 一处，删除 `grokUpstreamUserAgent` 常量一处。
- **前端**：`frontend/src/views/user/PaymentResultView.vue` 及其 spec。
- **数据库**：无。**迁移号仍是 `224`**，本轮上游没有新迁移，不得顺延。
- **配置**：无新增配置项，无默认值变化。
- **计费**：`computeCacheCreationCost` 的结果对「明细之和 > 聚合值」的历史矛盾报数会**变小**
  （从超收回到封顶值）。这是修正而非行为回退，但会让新旧账单在这类请求上不可逐分比对。
- **调度**：粘性会话命中率会**上升**（多识别一种头名、溢出不再改绑），账号分布随之更集中；
  队列打满时的临时借号行为本身不变。
- **兼容性**：无对外 breaking change。客户端显式请求 `claude-sonnet-4-5` 的行为不变；
  只有请求 `claude-sonnet-4-5-thinking` / `claude-sonnet-4-5-20250929` 的调用会被路由到 4.6。
- **风险面**：`4ca86c52` 触及登录/身份路径（`user_repo.go` + `auth_email_binding.go`），
  是本批唯一动认证相关代码的一项，验收要求见 `verification.md` §4。

## Execution References

- `source-baseline.md`：上游 commit 固定、按文件 apply 三态证据、`apply --check` 的三种假信号。
- `design.md`：移植策略、逐项决策、不做项的理由、阶段划分与回滚。
- `docs/upstream-sync/PORTING-0.1.183.md` §3：逐条 patch site、本仓库行号与上游 diff 摘要。
- `docs/upstream-sync/README.md` §4 硬约束 / §5 PG→SQLite 转换速查：本仓库的移植护栏。
