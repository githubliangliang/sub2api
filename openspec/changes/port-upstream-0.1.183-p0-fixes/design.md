## Context

### 本仓库与上游的关系

本 fork 的 git 历史已重写，与上游没有共同祖先，**不能 `git merge upstream/main`**，只能按功能移植
（`CLAUDE.md`、`docs/upstream-sync/README.md` §2）。同时本 fork 只支持 SQLite、Redis 可关、
去掉了多实例与部分 SaaS 页面，所以「上游有这个 fix」不等于「本仓库有这个 bug」，也不等于
「这个 fix 在本仓库能编译」。

### 本批的性质

12 项全是**行为修正**，没有一项引入新配置、新表、新迁移或新对外接口。这决定了本 change 的形态：
不需要功能开关、不需要灰度、不需要数据迁移；风险集中在「改错」和「改漏」，而不是「上线顺序」。

### 判定过程

`v0.1.180..main` 44 个非 merge commit，逐条生成 patch → 按文件切开 → 单文件 `git apply --check`
→ 对新增函数调用额外 grep 依赖符号。三态与三种假信号见 `source-baseline.md` §3/§4。
44 条最终分成：12 项本批、2 项 P1（需手工改写，另开 change）、17 项缺基座或不成立、
2 项 CN 平台 N/A、6 项 chore、若干纯测试。

## Goals / Non-Goals

### Goals

- 让 12 个已确认存在的缺陷在本仓库消失，行为与上游 v0.1.183 一致。
- 每项都有可执行的验收证据（单测或明确的手工步骤），而不是「patch 应用成功」。
- 保持迁移号、VERSION、配置默认值、对外 API 契约不变。
- 把「为什么不做另外那些」固化下来，避免下一轮重新调查。

### Non-Goals

- 不跟进 Responses Lite 并行工具调用簇（5 条）。缺 0.1.180 的四个基座符号，做它等于先做
  0.1.180 §6.1 + §7.1。
- 不跟进 `f1aadd48`。本仓库该缺陷不成立（`source-baseline.md` §5 末段）。
- 不跟进 Codex routed model catalog 簇（11 条）。未进任何 tag，上游最新一条还在「harden」。
- 不做两项 P1：`49752060`+`b20f29d1`（监控 v2 composite 平台归属，上游是 PG 方言、本仓库是独立
  SQLite 重写，需重写语义）与 `4795650d`（真实上游端点，需裁掉 `IsCNProvider`）。
  它们值得做，但需要手工改写，不属于「移植」这个 change 的范围。
- 不同步 `backend/cmd/server/VERSION`，不合 sponsors / 素材类 chore。
- 不做 Kimi 相关两条（本仓库无 CN 平台）。

## Decisions

### 1. 以「行为契约」为交付物，patch 只是手段

12 项里 10 项可近乎逐字应用，但本 change 的验收标准是 `specs/*/spec.md` 里的 Requirement 成立，
不是 `git apply` 成功。理由：`e55727d4` 必须手改、`bc4a9ae4` 的 `billing_service.go` 有上下文差异、
`99ec347e`+`71aa6e35` 的净效果与单看任一条都不同。用行为而不是 diff 做验收，这三种情况才有统一口径。

### 2. `cache_creation` 明细修两层，不只修解析

第一层是把 `Exists() && v > 0` 改成 `Exists()`（4 处），让 `0` 能覆盖旧值——这是缺陷的直接成因。
第二层在 `computeCacheCreationCost` 前插入 `normalizeCacheCreationBreakdown`：当
`5m + 1h > cache_creation_input_tokens` 时按原比例缩回聚合值封顶。

第二层是防御性的，单独看是冗余的，仍然要合：上游用词是 "contradictory details"，说明报数矛盾这件事
不是一次性的解析 bug，而是上游持续可能发生的输入。有了封顶，后续任何新形态的矛盾报数都不会再超收。
封顶按原比例分配而不是简单截断，保留 5m/1h 的相对结构（两档单价不同，截断会引入新偏差）。

### 3. 粘性会话的两条修在同一文件，但拆成两个可独立回滚的改动

`8e60d574`（识别 `session-id`）与 `e55727d4`（溢出不改绑）都改
`internal/service/openai_gateway_scheduling.go`，但落在不同函数、互不依赖：

- 前者改 `explicitOpenAIHeaderSessionNames`（`:31`）与 `openai_ws_forwarder_logutil.go:68`。
- 后者改 `selectAccountWithLoadAwareness`（`:932` 起）内的三处：函数开头引入 `stickySpillover`、
  Layer 1 的 `waitingCount < cfg.StickySessionMaxWaiting` 分支置 true、`:1185` 与 `:1224` 两处
  `setStickySessionAccountID` 各加 `&& !stickySpillover`。

分开提交，因为两者对账号分布的影响方向相同（都让会话更粘），一旦出现意外集中要能分别定位。

⚠️ 这个文件还与 `docs/upstream-sync/PORTING-0.1.180.md` §6.3（`3fd66a33b` 调度诊断）和 §7.3
抢同一处。本 change 先落地，那两项后续 rebase 到本 change 之上。

### 4. Antigravity Sonnet 别名成对合，且以「显式 4.5 透传」为终态

`99ec347e` 把 `claude-sonnet-4-5` 系全部指向 4.6；`71aa6e35` 随后把 canonical 的
`claude-sonnet-4-5` 收回原样。**单合前者会吞掉客户端的显式选择**，因此两条必须一起。

对 `internal/domain/constants.go:105` 的 `DefaultAntigravityModelMapping` 净效果只有三行：

| key | 现在 | 终态 |
|---|---|---|
| `claude-sonnet-4-5` | `claude-sonnet-4-5` | `claude-sonnet-4-5`（不变） |
| `claude-sonnet-4-5-thinking` | `claude-sonnet-4-5-thinking` | `claude-sonnet-4-6` |
| `claude-sonnet-4-5-20250929` | `claude-sonnet-4-5` | `claude-sonnet-4-6` |

判定原则：**兼容别名可以迁，canonical 的显式选择不能改写**。这条原则也用来审查后续所有别名类移植。

### 5. Responses 工具 item ID 用「换前缀、保后缀」而不是删 ID

本仓库现状是降级时直接 `delete(item, "id")`、还原时不换回，于是客户端历史里存着 `fc_` 前缀的
`custom_tool_call`，下一轮 replay 到会校验 ID 的上游就报
`Invalid 'input[N].id' ... Expected an ID that begins with 'ctc'`。

上游改法保留后缀只换前缀（`ctc_` / `tsc_` / `fc_` 三者互转），使 ID 在双向映射下稳定且唯一；
流式路径额外用 `clientItemID` 区分「上游用的 ID」和「发给客户端的 ID」，这样后续上游事件仍能按
原 ID 匹配到同一次调用。不认识的前缀保持原样，不猜。

这与 `docs/upstream-sync/PORTING-0.1.180.md` §6.1 的工具桥接簇同族，但**不依赖那簇**，
产品文件（`internal/pkg/apicompat/`）独立可合。

### 6. 邮箱换绑：前置查重 + 锁内复查，两层都要

前置 `ensureEmailIdentityAvailableForUser` 在发码和提交两个入口做快速查重（字面地址 → provider
alias），当前用户自己的 alias 变体放行，便于用户更换自己的写法。但前置查重与写入之间仍有并发窗口，
所以写入侧新增 `UpdateEmailWithAliasGuard`：在调用方事务里按
`normalizedEmailUniquenessLockKey` + `emailAliasUniquenessLockKey` 加锁 → 复查归属 → 再写
`SetEmail` / `SetPasswordHash`。

**SQLite 说明**：不需要任何方言改写。本仓库 `lockRepositoryScopedKeys`
（`internal/repository/user_profile_identity_repo.go:208`）已经是纯进程内锁，PG advisory lock 那套
早就替掉了，单进程语义足够。这一条**不能**因为「我们跑 SQLite」就退化成只留前置查重——
那正是 `docs/upstream-sync/README.md` 硬约束第 9 条禁止的静默降级。

### 7. Grok UA：把版本号跟进和占位 UA 清理当成一条

严格说 `CLIClientVersion` `0.2.114` → `0.2.120` 是维护性跟进而非缺陷，但它和「两处仍在发
`sub2api-grok/1.0`」在同一 commit 且共用同一处常量。分开合会留下「常量已删、调用点还在」的中间态，
所以整条合，并在同一提交里删掉 `grok_upstream_headers.go:15` 的 `grokUpstreamUserAgent`。

### 8. 六个 capability 按「受影响的行为域」切，不按文件或 commit 切

`upstream-request-compatibility` 一个 spec 覆盖 Gemini / Antigravity / Grok / Responses 工具协议
五项修复，因为它们共享同一个不变量：**出站请求必须满足上游的硬约束，否则请求整体失败或被上游改写**。
计费、粘性、冷却、邮箱、支付各自独立成 spec，因为它们的失效模式和验收手段完全不同
（钱、缓存命中率、调度时机、并发安全、UI 可见性）。

### 9. 迁移号与 VERSION 明确不动

本轮上游没有新迁移，`backend/migrations/` 保持到 `224`。任何实现中出现新增 SQL 文件的冲动都说明
走偏了。`VERSION` 保持自有编号 `1.1.8`。

## Risks / Trade-offs

| 风险 | 处理 |
|---|---|
| `4ca86c52` 触及 `user_repo.go` 与认证路径，改错会影响登录 | 单独一个提交阶段；验收必须跑登录与 `user_allowed_groups` 相关测试（缺表会 503） |
| 封顶归一化让新旧账单在矛盾报数的请求上不可逐分比对 | 这是修正而非回退；在 `verification.md` 里要求记录一条封顶前后的对比样本，便于事后解释账单差异 |
| 粘性命中率上升导致账号分布更集中 | 两条粘性修改分开提交；观察期内盯账号级并发与 429 分布 |
| `e55727d4` 必须手改三处，容易漏一处 | spec 用「溢出请求完成后绑定仍指向原账号」而不是「代码里有 stickySpillover」做验收 |
| 12 项塞进一个 PR 难以评审 | 按 Migration Plan 的四个阶段拆提交/PR |
| 移植时顺手带入 `3e98a5a1` 之类「apply 干净但属于功能簇」的提交 | `source-baseline.md` §4 显式点名；tasks 里有一条门禁要求核对最终 diff 的文件清单 |

## Migration Plan

无数据迁移、无功能开关，四个阶段按「验收独立性」而不是依赖关系划分。

### 阶段 1：涉钱与涉缓存（先做，收益最大）

`8e60d574`（session-id）→ `bc4a9ae4`（cache_creation 两层）→ `e55727d4`（溢出不改绑）。
三条各自独立提交。阶段结束跑计费与调度相关单测。

### 阶段 2：上游请求兼容性（一批干净项）

`19da0f24`（Gemini schema）、`1a9898a6`（Antigravity clamp）、`329b92ef`（图片 prompt）、
`e0e5e45c`（item ID）、`9fb26043`（Grok UA）、`99ec347e`+`71aa6e35`（Sonnet 别名，成对）。
可合并成一个提交，但 Sonnet 别名那两条必须同时在内。

### 阶段 3：账号冷却与前端

`a6b11ccc`（OpenCode Go 重置时长）、`eb594eef`（充值余额刷新）。互不相干，一个提交即可。

### 阶段 4：认证路径（单独隔离）

`4ca86c52`（邮箱换绑 alias + 并发守卫）。单独一个提交/PR，验收要求见 `verification.md` §4。

### 回滚

每项都是纯代码修正，无 schema、无配置、无数据副作用 ⇒ `git revert` 对应提交即可，无回滚脚本。
唯一需要说明的是 `bc4a9ae4`：revert 后已按封顶值记账的历史记录不会回滚成超收值，也不需要回滚。
