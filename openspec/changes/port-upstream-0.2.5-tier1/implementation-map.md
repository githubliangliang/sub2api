# 现行实施映射与来源勘误

本文件为 T01–T23 的实施入口。T01–T16保持原编号，T17–T23承接原第二档七项；
第一档现行22个 PR，第二档现行21个 PR，两档合计仍为原43个。原始 source-* 文件保持冻结，其中 T08–T15 为旧编号；
以下表格按来源 PR/SHA 关联，不把历史编号直接作为当前任务号。

## 来源、产品位点与测试入口

路径均相对仓库根目录；表中短 SHA 对应 source-feature-map.md 中同 PR 的完整 SHA。
测试入口是本地现有文件或包，允许在该位置补充用例，不表示已经有覆盖或已经执行。

| ID | 冻结编号 | PR / SHA | 产品 patch site | 本地测试入口 / 验收方式 |
|---|---|---|---|---|
| T01 | T01 | #7082 / `1e1d15cca` | `backend/internal/service/antigravity_token_provider.go`、`backend/internal/service/token_cache_invalidator.go` | service 包：共享 project_id、独立 token、旧键失效 |
| T02 | T02 | #7162 / `b8275209e` | `backend/internal/service/antigravity_gateway_streaming.go` | service 包：有限 SSE 流输出字节及帧数 |
| T03 | T03 | #7022 / `5948988aa` | `backend/internal/service/header_util.go` | service 包：本地 httptest server 观察传输后请求头 |
| T04 | T04 | #7073 / `a4c517917` | `backend/internal/service/batch_image_public.go` | service 包：provider/账号候选 fixture |
| T05 | T05 | #6904 / `685e97a3a` | `backend/internal/repository/scheduler_cache.go` | `backend/internal/repository/scheduler_cache_test.go`；来源普通 unit 可适配 |
| T06 | T06 | #6924 / `2493b0dc3` | `backend/internal/pkg/openai/request.go`、`backend/internal/service/openai_codex_identity.go` | pkg/openai、service 包：配对及最终出站 UA |
| T07 | T07 后端 | #7052 / `ba57ea914` | `backend/internal/server/middleware/jwt_auth.go`、`backend/internal/server/middleware/admin_auth.go` | middleware 的 jwt_auth_test.go、admin_auth_test.go；api_key_auth_test.go 兼容性回归 |
| T08 | T07 前端 | #7052 / `ba57ea914` | `frontend/src/api/client.ts` | `frontend/src/api/__tests__/client.spec.ts`、`frontend/src/api/__tests__/tokenRefresh.spec.ts` |
| T09 | T08 | #6969 / `75b7dd1e0` | `backend/internal/service/admin_group.go` | `backend/internal/service/admin_service_group_test.go`、`backend/internal/service/group_peak_rate_test.go`；handler 响应验证 |
| T10 | T09 | #6843 / `8efe2fd8c` | `frontend/src/components/admin/usage/UsageTable.vue` | `frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts` |
| T11 | T10 | #6973 / `9c30951e4` | `frontend/src/views/user/PaymentView.vue` | `frontend/src/views/user/__tests__/PaymentView.spec.ts`；浏览器真实滚动 |
| T12 | T11 | #7012 / `e67ffda7a` | `backend/internal/service/payment_config_providers.go`、`frontend/src/components/payment/PaymentProviderDialog.vue`、`frontend/src/i18n/locales/en/admin/settings.ts`、`frontend/src/i18n/locales/zh/admin/settings.ts` | `backend/internal/service/payment_config_providers_test.go`、`frontend/src/components/payment/__tests__/PaymentProviderDialog.spec.ts` |
| T13 | T12 | #7050 / `67845665d` | `frontend/src/components/user/profile/ProfileBalanceNotifyCard.vue` | 在相邻 __tests__ 目录新增身份删除竞态用例 |
| T14 | T13 | #7053 / `99b93b298` | `frontend/src/views/admin/ProxiesView.vue` | `frontend/src/views/admin/__tests__/ProxiesView.list.spec.ts` |
| T15 | T14 | #7024 / `d2067668d` | `frontend/src/views/admin/SubscriptionsView.vue` | 在相邻 __tests__ 目录新增分配竞态用例，使用本地 fixture |
| T16 | T15 | #6912 / `67d3a896b` | `frontend/src/views/admin/RedeemView.vue` | 浏览器输入边界；`backend/internal/service/subscription_service.go` 的 MaxValidityDays 只读核对 |
| T17 | 原第二档 S11 | #7055 / `4ff3e6dfb` | `frontend/src/views/user/RedeemView.vue`、`frontend/src/i18n/locales/en/dashboard.ts`、`frontend/src/i18n/locales/zh/dashboard.ts` | 沿用原第二档行为测试；具体入口见 verification.md |
| T18 | 原第二档 S11 | #7054 / `f0dd49778` | `frontend/src/views/user/UsageView.vue` | 沿用原第二档行为测试；具体入口见 verification.md |
| T19 | 原第二档 S07 | #6964 / `310f8b7fa` | `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go` | 沿用原第二档行为测试；具体入口见 verification.md |
| T20 | 原第二档 S11 | #7026 / `0a378f343` | `frontend/src/views/user/KeysView.vue` | 沿用原第二档行为测试；具体入口见 verification.md |
| T21 | 原第二档 S11 | #7112 / `e9169901d` | `frontend/src/composables/useAutoRefresh.ts`、`frontend/src/views/user/ChannelStatusV1View.vue` | 沿用原第二档行为测试；具体入口见 verification.md |
| T22 | 原第二档 S11 | #7023 / `f2b51e3b2` | `frontend/src/components/user/profile/ProfileEditForm.vue`、`frontend/src/components/user/profile/ProfilePasswordForm.vue` | 沿用原第二档行为测试；具体入口见 verification.md |
| T23 | 原第二档 S11 | #7025 / `329641a86` | `frontend/src/components/admin/proxy/ImportDataModal.vue` | 沿用原第二档行为测试；具体入口见 verification.md |

前移来源的完整 SHA 保留在 [第二档冻结映射](../port-upstream-0.2.5-tier2/source-feature-map.md)；
原第二档 S07 退役，S11仅余#6654/#6916/#7111；现行契约不得在两档重复维护。

## 2026-09-19 勘误

核对基点为 `fe6bc318800ca86d2b658f0e8f46e83d32626e8b`，产品工作树无已跟踪修改。

1. **编号**：冻结来源表共15行；调整前tasks/spec共16项，前移后共23项。#7052的前后端分别使用T07/T08，
   #6969 起顺延一位，以本表为准。前移七项编号另用 T17–T23，不重编号原任务。
2. **认证误判**：api_key_auth.go 中 GetByKey 出错已区分 401/503/500；第 150 行是
   `apiKey.User == nil` 检查，没有 GetByID 错误。原 README/design 的“同源缺陷第三处”不成立。
   JWT/管理员的 not-found 保持 401，原 design 中的 404 是笔误。
3. **整 PR 复核**：对 15 个来源执行 `git diff <sha>^1 <sha>`，
   分别交给 `git apply --check` 和 `git apply --reverse --check`，均为只读检查。
   14 项正向成功；#7024 两向均未通过，因为修改的
   `frontend/src/views/admin/__tests__/SubscriptionsView.userUsageLink.spec.ts` 在本地不存在。
   按整 PR 记 CONFLICT；按该测试文件记 NOBASE。产品 SubscriptionsView.vue 单独正向检查成功，
   并非产品文件 debounce 上下文冲突。
4. **#7052 测试来源**：真实 diff 含 jwt_auth_test.go、client.spec.ts，
   不含 jwt_refresh_integration_test.go 或 admin_auth_test.go；原 proposal/source-baseline 的描述不能用于排除测试。
   admin_auth_test.go 在本地存在，可以自行补回归。当前整 PR 可应用不代表 fixture 必定兼容。
5. **冻结统计**：source-baseline 的 47 CLEAN / 3 CONFLICT 是历史每文件统计，不能用作本次整 PR 复核结果；
   原快照关于“全部测试非 postgres-gated”的结论与本地 T05 文件标签冲突。
6. **缺失总报告**：本地无 PORTING-0.2.5.md，现行 patch site 由本文件提供。
7. **范围措辞**：T03 是请求头；T10 来源只改成本浮层，不改表格行；T08 来源保留其它非瞬时失败的旧登出逻辑，
   并非严格仅 401/403 登出。spec/design 已按来源和本地行为收敛。
8. **证据文件格式复核**：evidence-0.2.5 下 TSV 使用真实 Tab 分隔，可按标准 TSV 读取。
   早先将工具输出中的转义显示误判为字面量 `\t`，现已更正；冻结证据本身无需修改。

## 排除与本地适配

- T05：丢弃 `backend/internal/repository/account_repo_integration_test.go` 的来源 hunk；
  保留本地 `integration && postgres` 标签，普通 unit 覆盖投影行为。
- T12：丢弃 `backend/internal/payment/provider/easypay_refund_test.go` 无关格式变化；
  支付配置提示和中英 i18n 属于本项交付内容，不能排除。
- T15：来源测试文件缺失，移植其中“关键词改变后不能给旧用户分配”的行为断言，
  不照搬其它 user usage link 测试与不可用 fixture。
- 其它来源测试逐一核查符号与 fixture；只有格式/上下文适用不能作为编译通过证据。

## 冻结文件完整性

文档补全前后的 SHA-256 应保持一致：

| 文件 | SHA-256 |
|---|---|
| source-baseline.md | `1929d39937a17b7467911049ab44318f29efc10afe4a2812a6f0e25f733011e5` |
| source-feature-map.md | `7163079b95c7927466902cdc6f1b2b62e3c922a03f653ab39a47ca8d4f495992` |
