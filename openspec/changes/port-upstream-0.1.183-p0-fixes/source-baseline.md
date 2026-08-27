# 源基线固定

对照日期：2026-08-26。**实施期间不得改动本文的 SHA**；如需重新对照，新建下一轮 change。

## 1. 上游固定点

| 点 | 值 |
|---|---|
| 上游仓库 | `https://github.com/Wei-Shaw/sub2api` |
| 上一轮对照基线 | `v0.1.180` = `c40edb4`（2026-08-24 11:58 UTC） |
| `v0.1.181` | `2b576826` |
| `v0.1.182` | `308a2d33` |
| `v0.1.183` | `c21fd338`（2026-08-25 13:53 UTC，本轮最新正式版） |
| 上游 `main` | `efb46db0`（2026-08-26，含 0.1.183 之后 6 条未发布提交） |
| `c40edb4..main` | 44 commits（非 merge） |
| 本仓库基线 | `77df363a0`（= v0.1.179 + 已移植项；`f8f257ad7..HEAD` 只有 docs / chore） |
| 本仓库版本号 | `backend/cmd/server/VERSION` = `1.1.8`（自有编号，**不同步成 0.1.183**） |
| 本仓库迁移号 | `224`（本轮上游无新迁移，**不顺延**） |

复现固定点：

```bash
git clone --bare --filter=blob:none --no-tags --single-branch --branch main \
  https://github.com/Wei-Shaw/sub2api.git /tmp/up183.git
git -C /tmp/up183.git fetch origin \
  refs/tags/v0.1.181:refs/tags/v0.1.181 \
  refs/tags/v0.1.182:refs/tags/v0.1.182 \
  refs/tags/v0.1.183:refs/tags/v0.1.183
```

## 2. 版本归属判定

上游 release bot **先打 tag，再补一条 `chore: sync VERSION to x`**，所以 `v0.1.181..v0.1.182`
区间里会出现名为「sync VERSION to 0.1.181」的提交。按 commit 日期或 VERSION 提交位置推断版本会错位。
本文与本 change 的版本号一律用祖先判定：

```bash
git -C /tmp/up183.git merge-base --is-ancestor <sha> <tag> && echo "in <tag>"
```

## 3. 本批 12 项的固定 SHA 与三态证据

`ok` / `CONFLICT` / `NOFILE` 是按**单文件** patch 逐个 `git apply --check` 的结果（整 commit
的结论会被测试文件淹没，见 §4）。

| # | commit | 发布版本 | 产品文件三态 | 测试文件 | 规模 |
|---|---|---|---|---|---|
| 1 | `8e60d574` | v0.1.183 | 2/2 ok | 2/2 ok | 4f +69-2 |
| 2 | `bc4a9ae4` | v0.1.182 | 2/3 ok（`billing_service.go` 仅尾部上下文差异） | 2/3 ok | 6f +168-26 |
| 3 | `19da0f24` | v0.1.181 | 1/1 ok | 1/1 ok | 2f +69-1 |
| 4 | `1a9898a6` | v0.1.183 | 1/1 ok | 1/1 ok | 2f +40-1 |
| 5 | `4ca86c52` | v0.1.183 | 2/2 ok | 1/1 ok | 3f +274-30 |
| 6 | `329b92ef` | v0.1.182 | 2/2 ok | 1/1 ok | 3f +23-6 |
| 7 | `e0e5e45c` | v0.1.183 | 1/1 ok | 2/3 ok | 4f +270-21 |
| 8 | `a6b11ccc` | v0.1.182 | 1/1 ok | 1/2 ok | 3f +144-3 |
| 9 | `9fb26043` | v0.1.181 | 5/5 ok | 3/4 ok | 9f +45-48 |
| 10 | `eb594eef` | v0.1.182 | 1/1 ok | 1/1 ok | 2f +57-3 |
| 11 | `e55727d4` | v0.1.183 | CONFLICT（同结构，行号偏移约 190） | 1/1 ok | 2f +51-2 |
| 12 | `99ec347e` + `71aa6e35` | v0.1.182 | 前者 ok；后者 CONFLICT（依赖前者） | — | 9f + 7f |

## 4. 三种假信号（判定时必须过的三道关）

- **`ok` 不等于该合**。`3e98a5a1`（composite 精确账号别名路由）6 文件全干净，但它给
  `CompositeRouteResolver` 加了 `CompositeModelOwnershipResolver` 回调并在 `NewGatewayService`
  里接线，是未发布的 routed-catalog 功能簇的一环，后面还挂着 5 条返工提交。**不在本批。**
- **`CONFLICT` 不等于不该合**。`bc4a9ae4` 的 `billing_service.go` 只是尾部上下文不同（上游
  `calculatePerRequestCost` 开头是 `units := input.UsageUnits`，本仓库是
  `count := input.RequestCount`），被改的 `computeCacheCreationCost` 函数本体字节一致。
- **`ok` 也不保证能编译**。`apply --check` 只比上下文，不看新代码引用的符号在不在。`d6012b0b`
  的目标文件在，但它调用的 `decodeOpenAIJSONUseNumber` 本仓库零命中。**凡新增函数调用都要额外
  `grep -rn "func .*<name>"`**。

## 5. 依赖符号核查结论（决定哪些不做）

已确认**存在**于本仓库，本批 12 项可直接落地：

| 符号 | 位置 |
|---|---|
| `defaultGrokUpstreamUserAgent` | `internal/service/grok_upstream_headers.go:26` |
| `lockRepositoryScopedKeys`（已是纯进程内锁，无 PG advisory lock） | `internal/repository/user_profile_identity_repo.go:208` |
| `normalizedEmailUniquenessLockKey` / `emailAliasUniquenessLockKey` | `internal/repository/user_repo.go:1273` / `:1325` |
| `txAwareSQLExecutor` | `internal/repository/user_profile_identity_repo.go:826` |
| `translatePersistenceError` | `internal/repository/error_translate.go:50` |
| `ParseCodexRateLimitHeaders` | `internal/service/openai_gateway_usage.go:824` |
| `dropInvalidLoweredFunctionItemID`（待改名） | `internal/pkg/apicompat/responses_client_tools.go:220` |
| `cleanToolSchema` + `encoding/json` import | `internal/service/gemini_messages_compat_service.go:3532` |

已确认**零命中**，对应上游修复不在本批：

| 缺失符号 | 拦下的上游 commit |
|---|---|
| `ensureOpenAIResponsesLiteParallelToolCalls` | `d6012b0b` `095b5253` |
| `decodeOpenAIJSONUseNumber` | `d6012b0b` |
| `normalizeOpenAIResponsesLitePayloadForAccount` | `d5e43ef7` |
| `normalizeOpenAIParallelToolCallsWithoutTools` | `1563db3f` |
| `shouldRetryOpenAIOAuth429OnSameAccount` / `ShouldRetryOpenAIOAuth429` / `newOpenAIAccountFailoverError` / `openAIOAuth429RetryWindowActive` | `f1aadd48` |
| `shouldForwardOpenAIResponsesViaRawChatCompletions` / `IsCNProvider` | `4795650d`（P1，需裁剪，不在本批） |

`f1aadd48` 值得单独记一笔：本仓库 `markOpenAIOAuth429RateLimited` 拿到 429 就直接
`BlockAccountScheduling(account, cooldownUntil, "429")`，语义上本来就是「一律按配额限流处理」，
不存在「配额耗尽的号被留在同号重试里」这个缺陷。**上游 fix 只对上游成立**，是本轮最典型的反例。
