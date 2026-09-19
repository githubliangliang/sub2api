# 来源与行为映射

每行 SHA 是 PR 的 merge commit；统一使用 `git diff <sha>^1 <sha>` 查看完整功能来源。
行为契约在 specs，逐条 patch site 只在 [PORTING-0.2.5.md](../../../docs/upstream-sync/PORTING-0.2.5.md)。
本表的文件清单取自 `git diff --stat <sha>^1 <sha>`，并已逐个 `ls` 核对本仓库实际路径。

| ID | 来源 SHA | 来源 PR | 本仓库改动文件（已核实路径） |
|---|---|---|---|
| T01 | `1e1d15cca527e31759ec1f1cc5b5876cc6f8377b` | #7082 antigravity-token-cache-isolation | `backend/internal/service/antigravity_token_provider.go`、`backend/internal/service/token_cache_invalidator.go` |
| T02 | `b8275209e8dc642ca7f974854084280022eba11e` | #7162 antigravity-gemini-sse-separator | `backend/internal/service/antigravity_gateway_streaming.go` |
| T03 | `5948988aae5ce873b17a2446c27e11f84e54d4a6` | #7022 issue-3623-accept-encoding-header | `backend/internal/service/header_util.go` |
| T04 | `a4c517917beb4c18bd36b584d7fbcdfa757804d1` | #7073 fix-batch-image-priority | `backend/internal/service/batch_image_public.go` |
| T05 | `685e97a3a626fc1c0a253def0c7e466b542e13d2` | #6904 fix-scheduling-threshold-snapshot | `backend/internal/repository/scheduler_cache.go` |
| T06 | `2493b0dc3112fc6d03e5429d1615542ad60513a3` | #6924 fix/codex-user-agent-validation | `backend/internal/pkg/openai/request.go`、`backend/internal/service/openai_codex_identity.go` |
| T07 | `ba57ea914b01701a8d40fb40793162622bc299de` | #7052 issue-2912-transient-auth-failures | `backend/internal/server/middleware/jwt_auth.go`、`backend/internal/server/middleware/admin_auth.go`、`frontend/src/api/client.ts` |
| T08 | `75b7dd1e0b019a6782526afea4a9114e1b1ad20d` | #6969 issue-5882-peak-rate-validation | `backend/internal/service/admin_group.go` |
| T09 | `8efe2fd8cf111fa24293be6ffd1076d22192e8f9` | #6843 issue-6841-usage-cost-precision | `frontend/src/components/admin/usage/UsageTable.vue` |
| T10 | `9c30951e4aca86c4590a361a06d599b29c06caad` | #6973 issue-4210-renewal-modal-scroll | `frontend/src/views/user/PaymentView.vue` |
| T11 | `e67ffda7aed038c1e678ee835d84cf94e9cc06a5` | #7012 issue-6982-easypay-upstream-type-dots | `backend/internal/service/payment_config_providers.go`、`frontend/src/components/payment/PaymentProviderDialog.vue`、两份 i18n |
| T12 | `67845665d6eedbaf8503c22bfc49a0803f013cef` | #7050 bug-balance-email-verification-race | `frontend/src/components/user/profile/ProfileBalanceNotifyCard.vue` |
| T13 | `99b93b298547b12ed303ec6e919e058037ba5cdc` | #7053 bug-proxy-filter-pagination | `frontend/src/views/admin/ProxiesView.vue` |
| T14 | `d2067668d37e9ea97fc852c5a74a1912c1ea59ac` | #7024 fix-subscription-assignment-selection | `frontend/src/views/admin/SubscriptionsView.vue` |
| T15 | `67d3a896bdd632eeee822fc7496cac74affc12ea` | #6912 issue-2106-redeem-subscription-duration | `frontend/src/views/admin/RedeemView.vue` |

## 本仓库缺陷位点（评估期核实，实施时以当前文件为准复核）

| ID | 位点与现状 |
|---|---|
| T01 | `antigravity_token_provider.go:233-239` `AntigravityTokenCacheKey` 先取 `project_id`，非空即返回 `"ag:"+projectID`。上游改法是**删除**该分支、恒按 account ID，并在 invalidator 同时清理旧 `ag:<projectID>` 键（净删除，不是加字段） |
| T02 | `antigravity_gateway_streaming.go:287` 写 `data: %s\n\n`，`:291` 再把原始空行按 `%s\n` 透传 → `\n\n\n` |
| T03 | `header_util.go:37` `"accept-encoding": "accept-encoding"`（值为小写），`:69` 亦在列表内 |
| T04 | `batch_image_public.go:949` `accounts[i].Priority > accounts[j].Priority` |
| T05 | `scheduler_cache.go:967` `filterSchedulerCredentials` 键表缺 `account_scheduling_threshold`；`:984` `filterSchedulerExtra` 键表缺 `session_window_utilization`、`passive_usage_7d_utilization`、`passive_usage_7d_reset`、`passive_usage_7d_oi_utilization`、`passive_usage_7d_oi_reset`。6 个键在该文件命中数均为 0 |
| T06 | `PairCodexClientIdentity` 位于 `backend/internal/pkg/openai/request.go`，未校验 CR/LF 与控制字节即参与身份配对 |
| T07 | `internal/server/middleware/jwt_auth.go:74` 与 `admin_auth.go:179` 对任何 `GetByID` 错误一律 `AbortWithError(c, 401, "USER_NOT_FOUND", ...)`；`frontend/src/api/client.ts` 刷新失败即清 token。`service.ErrUserNotFound` 已存在于 `internal/service/user_service.go:32` |
| T08 | `admin_group.go:384` 与 `:775` 均 `return nil, err` 裸返回；同文件已有 8 处 `infraerrors` 用法 |
| T09 | `UsageTable.vue:194`、`:378`、`:382` 为 `toFixed(6)`；`UsageView.vue:671-672` 已是 `toFixed(8)`，两处不一致 |
| T10 | `PaymentView.vue:234` 容器无 `flex/max-h-full`，`:239` 标题与套餐列表无 `shrink-0`/`min-h-0 overflow-y-auto` |
| T11 | `payment_config_providers.go:226` `easyPayCustomMethodCodePattern = ^[a-z0-9_-]+$`，`:256` 用它校验 `UpstreamType` |
| T12 | `ProfileBalanceNotifyCard.vue:308` `pendingEmails.value.splice(idx, 1)` 按下标删除（模板 `:110` 亦同） |
| T13 | `ProxiesView.vue:27`、`:35` 两个筛选 `@change="loadProxies"`，未重置 `pagination.page` |
| T14 | `SubscriptionsView.vue` 在 debounce 搜索前不失效 `selectedUser`/`assignForm.user_id` |
| T15 | `RedeemView.vue:353` `max="365"`；后端 `internal/service/subscription_service.go:26` `MaxValidityDays = 36500`，`subscription_handler.go:45/53` 亦为 36500 |

## 来源排除项（不随本 change 交付）

- T05 的 `backend/internal/repository/account_repo_integration_test.go` 上游为 `//go:build integration`，
  **本仓库同名文件是 `//go:build integration && postgres`**，在 SQLite-only fork 永不编译。该 hunk 直接丢弃，
  不得为让它可应用而改构建标签；T05 的回归改为新增 unit 用例覆盖两个键表。
- 上游 PR 内与本行为无关的测试/生成物不自动成为交付要求，逐项记入 verification.md。

完整候选档位与理由见 [candidates.tsv](../../../docs/upstream-sync/evidence-0.2.5/candidates.tsv)，
逐文件四态见 [files.tsv](../../../docs/upstream-sync/evidence-0.2.5/files.tsv)。
