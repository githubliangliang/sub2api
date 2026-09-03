# 源基线固定

对照日期：2026-09-02。**实施期间不得改动本文的 SHA**；如需重新对照，新建下一轮 change。

## 1. 固定点

| 点 | 值 |
|---|---|
| 上游仓库 | `https://github.com/Wei-Shaw/sub2api` |
| 上游最新正式版 | `v0.2.0`（tag commit `aa2364883`，2026-09-02 11:09 UTC） |
| 中间版本 | `v0.1.185`（tag commit `2ac784c51`，2026-09-01 09:32 UTC） |
| 上游 main | `5097b3145`，v0.2.0 之后仅 `chore: sync VERSION to 0.2.0` ⇒ 本批无需等新版本 |
| 本批 6 项的来源版本 | `e93e6368f` / `200b1406d` 属 v0.2.0；其余四条属 v0.1.185 |
| 本仓库基线 | `3da1c2dd0`（= 本仓库 `1.1.11`，0.1.183 全部 + 0.1.184 §3/§4/§5.1/§5.7 已合） |
| 本仓库版本号 | `backend/cmd/server/VERSION` = `1.1.11`（自有编号，**不同步成 0.2.0**） |
| 本仓库迁移号 | `224`（本批无新迁移，**不顺延**） |
| 工作分支 | `sync/upstream-20260902`（建议；实施时确认后回填） |

本轮直接用本仓库已有的 `upstream` remote，不再另开部分裸克隆：

```bash
git fetch upstream --tags --prune
git show <sha> -- <file> | git apply --check -            # 正向
git show <sha> -- <file> | git apply --check --reverse -   # 反向：ALREADY 判定
```

## 2. 本批 6 项的固定 SHA 与四态证据

四态是按**单文件** patch 逐个 `git apply --check` 的结果：`CLEAN`（正向可打）/
`ALREADY`（反向可打 = 已在库里）/ `CONFLICT` / `NOBASE`（本仓库没这个文件）。

| # | SHA | 标题 | 产品文件 | 测试文件 |
|---|---|---|---|---|
| 1 | `e93e6368f` | keep OpenAI passthrough flag in the scheduler snapshot projection | `scheduler_cache.go` **`CONFLICT`** | `scheduler_cache_unit_test.go` `CLEAN` |
| 2 | `200b1406d` | strip Anthropic fallbacks unless server-side-fallback beta is present | `pkg/claude/constants.go` `CLEAN`、`gateway_request.go` `CLEAN`、`bedrock_request.go` `CLEAN` | `gateway_fallbacks_sanitize_test.go` `CLEAN`（新建） |
| 3 | `1dc0a0900` | ctx_pool WS ingress 下发前改写上游容量降载错误码 | `openai_ws_forwarder_ingress.go` `CLEAN` | `openai_ws_ingress_capacity_shed_test.go` `CLEAN`（新建） |
| 4 | `6d5f02784` | recycle stale idle pool connections | `openai_ws_pool.go` `CLEAN` | `openai_ws_pool_test.go` `CLEAN` |
| 5a | `ba345f105` | ignore persistently disabled catalog accounts | `openai_codex_models_service.go` `CLEAN` | 4 个 `CLEAN` |
| 5b | `57c76584a` | advertise priority tier for fast models | `openai_codex_models_service.go` `CLEAN` | `openai_codex_models_service_test.go` `CLEAN` |
| 6 | `9eabd2a5b` + `e7c029875` | apply model pricing policies to account stats cost | `account_stats_pricing.go` **`CONFLICT`** | `account_stats_pricing_test.go` `CLEAN` |

两处 `CONFLICT` 的归属（**都不是「本仓库已经修过」**）：

- 第 1 项 `scheduler_cache.go`：上游 pre-image 的 `keys` 列表里 `"openai_responses_supported"`
  之后紧接 `"codex_fingerprint_mode"` / `"codex_fingerprint_seed"`，**本仓库这两个 key 不存在**，
  直接接 `"codex_5h_used_percent"` ⇒ 上下文漂移。被改的函数与语义完全对得上。指纹两个 key 是
  另一处分叉，见 `design.md` 决策 1。
- 第 6 项 `account_stats_pricing.go`：本仓库 `tryModelFilePricing` 与上游 pre-image **逐字节一致，
  只少条件里的 `normalizedTier == "fast"`**（上游 `9261dd77` 加的，本仓库没取）。上游这条把整个手算
  分支删掉 ⇒ 冲突随之消解。

## 3. 依赖符号核查（`apply --check` 通过 ≠ 能编译）

对本批每条新增代码引用的符号逐个 grep：

| 新代码引用 | 本仓库 | 结论 |
|---|---|---|
| `filterSchedulerCredentials`（第 1 项同函数族） | `internal/repository/scheduler_cache.go` 已存在，保留 `model_mapping` | 缺陷前提成立 |
| `Account.IsOpenAIPassthroughEnabled`（第 1 项的受益方） | `internal/service/account.go:1742` | 缺陷链闭合 |
| `Account.IsModelSupported` 的透传短路（issue #4936） | `internal/service/account.go:816` 第一行 | 修复已在、只是输入被裁 |
| `SchedulerSnapshotService.ListSchedulableAccounts` | `internal/service/scheduler_snapshot_service.go:210` | 读的就是该投影 |
| 快照 worker 是否在 SQLite 上启动 | `internal/service/wire.go:419` 注释明确「必须照常」 | **不在 `skipSQLiteBackgroundJobs` 之列** |
| `sanitizeOpenAICapacityShedErrorCodeForClient`（第 3 项） | `internal/service/openai_gateway_passthrough.go:1131` | 可编译 |
| `supportsIdlePingWithoutReader`（第 4 项） | `internal/service/openai_ws_pool.go:466` | 可编译 |
| `idleDuration`（第 4 项） | `internal/service/openai_ws_pool.go:505` | 可编译 |
| `CalculateCostWithServiceTier`（第 6 项） | `internal/service/billing_service.go` 已存在 | 可编译 |
| `shouldApplySessionLongContextPricing`（第 6 项删掉的分支引用） | `internal/service/billing_service.go:1413` | 删除后应无残留引用，实施时复 grep |
| `normalizeBillingServiceTier`（第 6 项语义前提） | `internal/service/billing_service.go:122`，**只做 lower+trim，不把 `fast` 映射成 `priority`** | 缺陷前提成立 |
| `internal/service/account_stats_pricing.go` 的 `GetModelPricing` 调用 | 第 6 项删除后不再需要 | 实施时确认 lint 不报未使用 |

第 5a/5b 的基座是 0.1.184 §5.1 已合的 Codex routed catalog 整簇，`openai_codex_models_service.go`
在本仓库 2320 行，两条 patch 全 `CLEAN`。

## 4. 本批**不含**的相邻条目（避免下一轮重复判断）

| 上游 | 属簇 | 本批不含的理由 |
|---|---|---|
| `1be69e56a` + `421a83282` | PORTING §3.8 | 成链、+503 行，规模超出「小项」，放第二档 |
| `504919a05` | PORTING §3.9 | 1 处 `CONFLICT` 要手改，放第二档 |
| `bfe0a5a87` | PORTING §3.10 | 需补 11 行 `openAIWSRelayActiveTurnID` 基座，放第二档 |
| `34b8bf1a6` | PORTING §3.11 / §4.3 | 一个 PR 两件事，其中一件带迁移，放第二档 |
| `593fc9365` / PR#6447 / PR#6425 / `1a33dc8cc` | PORTING §4 | P1，放第二档 |
| PR#6443 / PR#6444 | PORTING §5.1 | 按需（是否真发 fast/priority），**等用户确认** |
| `530fb20f2` + `e2cfaa46e` | PORTING §5.2 | 基座 573 行，且不修任何可观察缺陷，**等用户确认** |
| PR#6463 | PORTING §5.3 | 缺 CN 平台整支 |
| `e21b849a9` | PORTING §5.4 | 缺 `UsesOpenAICodexProtocol`，且需先核清本 fork 的主调用路径 |
| `863667ce6` + `b6a7b8b7d` | PORTING §5.5 | PG-only（函数注释与 `lib/pq` 双证），永久不做 |
| `343858021` | PORTING §5.6 | 本仓库无 `classifySelectionFailureError` ⇒ **没有那个 bug**，合了是空转 |
| `e2624fb65` | PORTING §6 | **已合**（三文件 `ALREADY`，后像逐点核过） |
| `05ea883e2` / `77729e272` / sponsors / VERSION | PORTING §6 | 生成物与 chore |
