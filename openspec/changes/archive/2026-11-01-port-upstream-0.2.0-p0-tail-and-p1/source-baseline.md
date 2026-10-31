# 源基线固定

对照日期：2026-09-02。**实施期间不得改动本文的 SHA**；如需重新对照，新建下一轮 change。

## 1. 固定点

| 点 | 值 |
|---|---|
| 上游仓库 | `https://github.com/Wei-Shaw/sub2api` |
| 上游最新正式版 | `v0.2.0`（tag commit `aa2364883`，2026-09-02 11:09 UTC） |
| 中间版本 | `v0.1.185`（tag commit `2ac784c51`，2026-09-01 09:32 UTC） |
| 上游 main | `5097b3145`，v0.2.0 之后仅 `chore: sync VERSION to 0.2.0` |
| 本仓库基线 | **实际从 `3da1c2dd0` 开工**（第一档 `port-upstream-0.2.0-p0-fixes` 仍未合回 `main`；偏离见 `verification.md` 文首） |
| 本仓库版本号 | `backend/cmd/server/VERSION` = `1.1.11`（自有编号，**不同步成 0.2.0**） |
| 本仓库迁移号 | 起点 `224` → **本批顺延到 `226`** |
| 工作分支 | `sync/upstream-20260902-p1`（建议；实施时确认后回填） |
| 第 9 项的来源 | **不属于本轮**：上游 `4a1da2950`（v0.1.180 期），本仓库 0.1.180 §5.1 挂了一轮的安全项，2026-09-02 重量后并入本批 |

⚠️ **本批的四态证据是在 `3da1c2dd0`（第一档之前）上量的。** 第一档落地后，第 1 项的
`421a83282` 与第 6 项的两处 WS/handler 冲突预期会转成 `CLEAN`；实施时以当时的实际结果为准，
本文的数字只用于说明「冲突属于顺序问题而非难度」。

```bash
git fetch upstream --tags --prune
git show <sha> -- <file> | git apply --check -                    # 单 commit
git diff <merge>^1 <merge> -- <file> | git apply --check -        # PR 整体
```

## 2. 固定 SHA / PR merge 与四态证据

| # | 来源 | 版本 | 规模 | 四态 |
|---|---|---|---|---|
| 1a | `1be69e56a` | v0.1.185 | 2 文件 / +408 | 2/2 `CLEAN` |
| 1b | `421a83282` | v0.2.0 | 2 文件 / +225-3 | 1 `CLEAN` / 1 `CONFLICT`（**仅因 1a 未落**） |
| 2 | `504919a05` | v0.2.0 | 4 文件 / +159-4 | 3 `CLEAN` / 1 `CONFLICT`（`openai_gateway_chat_completions.go`） |
| 3 | `bfe0a5a87` | v0.2.0 | 2 文件 / +56-2 | 2 `CONFLICT` |
| 4 | `34b8bf1a6` 前半 | v0.2.0 | — | `domain/constants.go` `CLEAN`、`pkg/claude/constants.go` `CLEAN`、`billing_service.go` `CONFLICT`、`useModelWhitelist.ts` `CLEAN` |
| 5 | `593fc9365` | v0.1.185 | 4 文件 / +334 | 3 `CLEAN` / 1 `CONFLICT`（`pricing_service.go`） |
| 6a | PR#6447（merge `559960865`，分支尖 `aa7a811e6`） | v0.2.0 | 47 文件 / +746-100 | **38 `CLEAN`** / 9 `CONFLICT` |
| 6b | PR#6425（merge `3510aa22b`，分支尖 `7c01ec9be`） | v0.2.0 | 15 文件 / +995-100 | 6 `CLEAN` / 9 `CONFLICT`（**多数因 6a 未落**） |
| 7 | `34b8bf1a6` 后半 | v0.2.0 | — | `channel_repo_pricing.go` / `channel_repo_account_stats_pricing.go` / `channel_handler.go` / `available_channel_handler.go` / `channel{,_service}.go` / `model_pricing_resolver.go` `CONFLICT`；`migrations/232_channel_cache_write_1h_pricing.sql` `CLEAN`（**但是 PG 方言，必须重写**）；`channel_repo_pricing_time_test.go` `NOBASE` |
| 8 | `1a33dc8cc`（merge `aa2364883`） | v0.2.0 | 4 文件 / +27-10 | 1 `CLEAN` / 3 `CONFLICT` |
| 9 | `4a1da2950` | **v0.1.180** | 上游 2 文件 / +20-64 | **三态不适用**——不打 patch，见 §5 |

`34b8bf1a6` 整体是 44 文件 / +571-203，**一个 commit 两件事**，按 §4 / §7 拆开落，见 `design.md` 决策 4。

PR#6425 的分支里还有 `77729e272 merge: resolve conflicts with upstream/main` 与
`05ea883e2 fix(ent): correct group field indexes after merge` 两条上游解冲突提交，**都不合**；
用 `git diff 3510aa22b^1 3510aa22b` 取净效果。

### 2.1 主要 `CONFLICT` 的归属（都不是「本仓库已经修过」）

| 文件 | 归属 |
|---|---|
| `openai_gateway_handler.go`（1b、6a） | 顺序问题：1a / 第一档的 `1dc0a0900` 未落 |
| `openai_ws_forwarder_ingress.go`（6a） | 同上（第一档 §3.3 未落） |
| `openai_ws_v2/passthrough_relay{,_test}.go`（3） | 缺基座 `openAIWSRelayActiveTurnID`，见 §3 |
| `repository/group_repo.go`（6a） | 本 fork 是**手写 SQLite SQL**，必然手改 |
| `ent/runtime/runtime.go`（6a） | 生成物 ⇒ 改 `ent/schema/group.go` 后 `go generate ./ent`，**不打补丁** |
| `handler/composite_platform{,_test}.go`（6a） | 本仓库 composite 有自有分叉 |
| `billing_service.go`（4） | 本仓库 fable 系列**整段不存在**（grep `fable` 零命中）⇒ 需插入而非替换 |
| `pricing_service.go`（5） | 本仓库该文件 1137 行、与上游长期分叉 |
| 渠道定价四个文件（7） | 本仓库渠道定价是手写 SQLite SQL + 自有 `image_input_price` 列 |
| 前端 `GroupsView.vue` / `types/index.ts` / `IntervalRow.vue` / `PricingEntryCard.vue`（6/7/8） | 本仓库长期自有改动 |

### 2.2 构建标签核查（本轮新增的一道关）

| 上游改到的文件 | 本仓库首行 | 处理 |
|---|---|---|
| `internal/repository/migrations_schema_integration_test.go` | `//go:build integration && postgres` | **永远不编译 ⇒ 整段 hunk 丢弃**，MUST NOT 改标签。改它的是**第三档**的 PR#6443 / PR#6444，本批不涉及；列在此处是为了让规则不被下一轮重新发现 |
| `internal/repository/api_key_repo_openai_fast_projection_integration_test.go` | 本仓库不存在（属第三档 PR#6443） | 与本批无关 |
| `internal/repository/channel_repo_pricing_time_test.go`（第 7 项） | 本仓库不存在（`NOBASE`） | 新建前先决定要不要标签；本批唯一需要看标签的文件 |

落地任何 `_integration_test.go` 之前先 `head -1`。

## 3. 依赖符号核查（`apply --check` 通过 ≠ 能编译）

| 新代码引用 | 本仓库 | 结论 |
|---|---|---|
| `openAIWSRelayActiveTurnID`（第 3 项） | **零命中** | ⇒ 需补基座，上游 11 行纯函数（遍历 `state.turnTimingByID` 找 `== state.activeTurn`），**随本项一起落** |
| `isDisconnectError`（第 3 项） | `internal/service/openai_ws_v2/passthrough_relay.go:985` | 可编译 |
| `internal/service/openai_compat_prompt_cache_key.go`（第 2 项基座） | 181 行，已存在 | 可编译 |
| `getFallbackPricing` / `initFallbackPricing`（第 4 项） | `internal/service/billing_service.go` 已存在 | 可编译；但 **fable 条目零命中 ⇒ 是新增不是修正** |
| `pricing.fallback_file` 配置（第 5 项前提） | `internal/config/config.go:653`，默认值 `:2318` | 语义可对齐 |
| `parsePricingData`（第 5 项挂载点） | `internal/service/pricing_service.go:421` | 挂载点在 |
| `ReasoningEffortMapping`（第 6 项） | `internal/domain/reasoning_effort.go:5`（别名 `internal/service/group.go:16`） | 基座在 |
| `NormalizeMaxReasoningEffort` / `NormalizeReasoningEffortMappings` / `ApplyOpenAIReasoningEffortPolicyFromContext`（第 6 项） | `internal/service/openai_reasoning_effort_policy.go`（233 行） | 基座在 |
| ent 字段 `max_reasoning_effort` / `reasoning_effort_mappings`（第 6 项） | `backend/ent/schema/group.go:270` / `:274` | 基座在 |
| 前端 `ReasoningEffortPolicyFields.vue` / `groupsReasoningEffort.ts`（第 6 项） | 均存在 | 基座在 |
| `CacheWrite1hPrice`（第 7 项） | `internal/service/channel_plaza.go:16`（服务层已有，读 LiteLLM `cache_creation_input_token_cost_above_1hr`） | 服务层在，**表列缺** |
| `channel_model_pricing` / `channel_pricing_intervals` / `channel_account_stats_model_pricing` / `channel_account_stats_pricing_intervals`（第 7 项四张表） | `migrations/081`/`082`/`085`/`086`/`101`/`106`/`178` 建立 | 四张表都在 |
| `service.IsCNProvider` / `nativeDeepSeekResponses` | **零命中**（本仓库无 CN 平台） | ⇒ 相关分支整段丢弃 |

## 4. 本批**不含**的相邻条目（避免下一轮重复判断）

| 上游 | PORTING | 不含的理由 |
|---|---|---|
| PR#6443 + PR#6444（Fast 组策略，+2 迁移 `227`/`228`） | §5.1 | **第三档：按需**。基座不缺（PR#6443 38/44 `CLEAN`），但只有真在发 fast/priority 才划算 ⇒ **等用户确认** |
| `530fb20f2` + `e2cfaa46e`（长上下文阶梯数据驱动） | §5.2 | 基座 573 行（`billing_context_schedule.go` 470 + `billing_token_cost_request.go` 103），且不修任何可观察缺陷 ⇒ **等用户确认**。其中 `resources/model-pricing/*.json` 那 26 行可单独摘（影响接近零） |
| PR#6463（Kimi 原生 Responses） | §5.3 | 缺 CN 平台整支（`PlatformKimi` 等不存在），5 处 `NOBASE` 是那支的骨架 |
| `e21b849a9`（API key 不合成 instructions） | §5.4 | 缺 `UsesOpenAICodexProtocol`，且需先核清本 fork 的主调用路径（三个调用点） |
| `863667ce6` + `b6a7b8b7d`（数据库启动重试） | §5.5 | PG-only（函数注释与 `lib/pq` 双证），永久不做 |
| `343858021`（404 不被 429 覆盖） | §5.6 | 本仓库无 `classifySelectionFailureError` ⇒ **没有那个 bug**，合了是空转 |
| `e2624fb65` / `05ea883e2` / `77729e272` / sponsors / VERSION | §6 | 已合 / 生成物 / chore |

## 5. 第 9 项（dompurify）的固定事实

三态不适用：上游 `4a1da2950` 改的是 `package.json` + `pnpm-lock.yaml`，而**上游 lockfile 是 pnpm 9
产物、本仓库本地是 pnpm 11，且本仓库多一个上游没有的 `pnpm-workspace.yaml`** ⇒ 只能按结论手改。
下面每条都是 2026-09-02 在 `/tmp` 副本里实测的（仓库文件未动），完整记录见
[0.1.180 §12](../../../docs/upstream-sync/PORTING-0.1.180.md)：

| 事实 | 值 |
|---|---|
| 本仓库直接依赖 | `frontend/package.json:25` = `^3.3.1` → `3.3.1` |
| 本仓库传递依赖 | `3.3.3`，路径 `. > @lobehub/icons > @lobehub/ui > mermaid > dompurify` |
| 目标版本 | `3.4.14`，2026-09-02 查 npm 仍是 `latest`（`3.3.1..3.4.14` 之间 18 个发布版） |
| 测量前提 | `pnpm install --lockfile-only` 在未改动输入上产出**逐字节相同**的 lockfile（439ms no-op）⇒ churn 全部归因于改动 |
| lockfile churn | **22 行**，全在 dompurify 条目与 overrides 块；两份副本合成一份 |
| 新增 warning / peer 冲突 | **无**（把 no-op 那次与全量解析那次逐条比过，差异全是既有 deprecation） |
| 净化行为 | jsdom 下 `3.3.1` 与 `3.4.14` 对 `sanitize.ts:5` 的原样调用输出**逐字节相同** |
| advisory | `pnpm audit` 全树 72 条，dompurify **18 条**（单包最多），**全部 low/moderate，0 critical/high** |
| CI 门禁 | `pnpm audit --prod --audit-level=high` ⇒ dompurify 那 18 条**从未进门禁**；门禁现报 `xlsx` ×2 + `nanoid` ×1，均有例外条目 |
| 两处 overrides 都要改 | 本地 pnpm `11.21.0` 打印 `"pnpm" field ... ignored: "pnpm.overrides"`；CI 是 `pnpm/action-setup@v6` + `version: 9`（`backend-ci.yml:50-52`）、`Dockerfile:28` `pnpm@9`、`deploy/Dockerfile:25` `pnpm@9.15.9` |
| 必须同一提交 | 5 处 `--frozen-lockfile`：`security-scan.yml:49`、`backend-ci.yml:61`、`release.yml:67`、`Dockerfile:34`、`deploy/Dockerfile:29` |
| 暴露面 | **7 个** `DOMPurify.sanitize` 调用点（原清单只记 1 个）；`/legal/:documentId` 是 `requiresAuth: false` 公开页；`/custom/:id` 是 `requiresAdmin: false` 且放开 `iframe` + `src` |
| 原清单引用的 CVE | `GHSA-cj63-jhhr-wcxv` —— 四种原型污染形态均**未能复现**，不作为本项理由（见 0.1.180 §12.5） |
