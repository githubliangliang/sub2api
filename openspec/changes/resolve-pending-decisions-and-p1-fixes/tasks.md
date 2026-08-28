## 1. 固定基线与边界

- [x] 1.1 确认起点：`1b4837098` + 工作区里 0.1.183 的 12 项；`go build ./...` 通过
- [x] 1.2 建议先把 0.1.183 那批提交掉，避免本批的改动和它混在同一个 diff 里
- [x] 1.3 保存基线信号：`cd backend && go test -tags=unit ./... -count=1` 与 `golangci-lint run ./...`（记录当前 v2.9 下的通过状态，作为阶段 5 的对照）
- [x] 1.4 确认 `backend/migrations/` 最大序号仍为 `224`，全程 MUST NOT 新增迁移
- [x] 1.5 确认 `backend/cmd/server/VERSION` 保持 `1.1.8`
- [x] 1.6 ⚠️ 全程 MUST NOT 改 `frontend/package.json` / `pnpm-lock.yaml` / `pnpm-workspace.yaml`（dompurify / nanoid 仍推迟，见 `design.md` 决策 1）
- [x] 1.7 按 `design.md` Migration Plan 拆成 5 个阶段的提交/PR

## 2. 阶段 1 — 恢复依赖审计门禁

- [x] 2.1 先跑一次门禁看全貌：`cd frontend && pnpm audit --prod --audit-level=high --json > /tmp/audit.json`，再 `python tools/check_pnpm_audit_exceptions.py --audit /tmp/audit.json --exceptions .github/audit-exceptions.yml`，记录**除三条过期外还有没有别的红**
- [x] 2.2 重新核实 `lodash` GHSA-r5fr-rjxr-66jc：确认 `frontend/src/` 仍无直接 `lodash` import（当前是传递依赖），把结论写进 `reason`
- [x] 2.3 重新核实 `lodash-es` 同一公告，同上
- [x] 2.4 重新核实 `axios` GHSA-3p68-rc4w-qgx5：它是直接依赖（`frontend/package.json:23` `^1.18.0` → lock `1.18.1`），确认本仓库用法不受影响并写进 `reason`
- [x] 2.5 三条各设新的 `expires_on`，**压在下一次计划性依赖维护之前**，不要给远期；`owner` 按实际填，不照抄 `security@your-domain` 占位值
- [x] 2.6 若任一条核实后发现理由已不成立 ⇒ **不要续期**，记为待处理项并在本 change 的完成说明里点出
- [x] 2.7 重跑 2.1 的门禁，确认通过；独立提交

## 3. 阶段 2 — Grok 目录、计费与默认模型（顺序固定）

### 3.1 `ed4207a16`（必须先）

- [x] 3.1.1 `pkg/xai/models.go`：新增 `DefaultImagineImage20Model = "grok-imagine-image-2.0"`，把 `DefaultImagineVideo15Model` 与 `DefaultImagineVideo15LegacyModel` 的取值互换回官方形态（`:58-62`）
- [x] 3.1.2 `pkg/xai/models.go:89-90`：从 `defaultModels` 移除 `grok-3-mini` 与 `grok-3-mini-fast`；`:97-101` 增加 image 2.0、修正 1.5 的 DisplayName
- [x] 3.1.3 `pkg/xai/models.go:112-113`：`"grok-4.5"` / `"grok-4.5-latest"` 从 `DefaultTextModel` 改成字面量 `"grok-4.5"`
- [x] 3.1.4 `pkg/xai/models.go:161`（`ModelMappingWithOptions`）与 `:272`（`ResolveGrokTextResponsesModelID`）：把无条件的 `canonical == DefaultTextModel` 收窄成 `(alias == "grok" || alias == "grok-latest" || alias == "grok-build-latest") && canonical == DefaultTextModel`
- [x] 3.1.5 `pkg/xai/models.go:181`：`mapping["grok-imagine-video-1.5"]` 改指 `DefaultImagineVideo15Model`
- [x] 3.1.6 `service/billing_service.go:704` 区域：新增 `grok-3-mini`（$0.30 / $0.50 / cache $0.075）与 `grok-3-mini-fast`（$0.60 / $4.00 / cache $0.15）两张价卡
- [x] 3.1.7 `service/billing_service.go:928` 区域：`getFallbackPricing` 增加 `grok-3-mini` / `grok-3-mini-fast` 两个 case
- [x] 3.1.8 ⚠️ **丢弃 `openai_gateway_grok.go` 的全部 hunk**：它调用本仓库不存在的 `grokSameAccountRetryMetadata`（属 §6.2(c)），合了不编译。见 `source-baseline.md` §2.2
- [x] 3.1.9 落地上游对应的测试改动（`pkg/xai/models_test.go`、`service/billing_service_test.go`）

### 3.2 `39485f2e2`（在 3.1 之后）

- [x] 3.2.1 `pkg/xai/models.go:54`：`DefaultTextModel` 改 `"grok-4.6"`
- [x] 3.2.2 `pkg/xai/models.go`：`"grok-build-latest"` 的目标从 `DefaultTextModel` 改成 `"grok-build-0.1"`；把 3.1.4 的收窄条件再去掉 `grok-build-latest` 一项（两处）
- [x] 3.2.3 `service/billing_service.go`：新增 `grok-4.20` 价卡（$1.25 / $2.50 / cache $0.20，200k 阈值 2× 倍率），把 `grok-4.20-*` 从 `grok-4.3` 的 case 拆出来指向新卡
- [x] 3.2.4 `service/billing_service.go`：`getFallbackPricing` 的 `grok` / `grok-latest` 归到 `grok-4.6`，`grok-4.5` / `grok-4.5-latest` 单独一个 case 保留 4.5 卡
- [x] 3.2.5 `service/setting_parse.go` / `service/setting_update.go`：默认模型可选值同步到 4.6
- [x] 3.2.6 `frontend/src/composables/useModelWhitelist.ts`：白名单同步
- [x] 3.2.7 `README.md` 里 Grok 媒体模型清单一行同步（加 image 2.0）
- [x] 3.2.8 落地上游对应的测试改动（`pkg/xai/models_test.go`、`pkg/xai/oauth_test.go`、`service/billing_service_test.go`、`frontend/.../useModelWhitelist.spec.ts`）

### 3.3 `f7145c750`（与 3.2 同批上线，不得分开）

- [x] 3.3.1 `service/setting_gateway_runtime.go`：新增 `SettingService.MigrateGrokDefaultTextModel`——读 `SettingKeyGrokDefaultTextModel`，`ErrSettingNotFound` ⇒ 返回 nil；`strings.TrimSpace(value) != "grok-4.5"` ⇒ 返回 nil；否则 `Set` 成 `"grok-4.6"`。超时用 `codexRestrictionPolicyDBTimeout` + `context.WithoutCancel`
- [x] 3.3.2 `service/wire.go:745` 的 `ProvideSettingService` 体内加一次调用，失败只记 Warning。**不改 provider 签名 ⇒ 不需要 `go generate ./cmd/server`**
- [x] 3.3.3 落地上游对应的测试改动（`server/api_contract_test.go`、`service/setting_service_codex_policy_test.go`）

## 4. 阶段 3 — 可观测性两项

- [x] 4.1 `3fd66a33b`：`service/openai_gateway_scheduling.go:360` 起，把 `isOpenAICompatibleAccountEligibleForRequest` 与 `...BeforeProfit` 改成 `... == ""` 的薄包装，新增两个返回 `string` 的 `openAICompatibleAccountEligibilityFailureReason*`
- [x] 4.2 `3fd66a33b`：逐个否决点给名字——`account_nil` / `platform_mismatch` / `model_rate_limited`（账号整体可调度但按模型限流）/ `not_schedulable` / `quota_auto_pause_<window>`（window 为空时退化成 `quota_auto_pause`）/ `model_not_supported` / `capability_mismatch` / `compact_unsupported`，利润控制直接用 `openAIProfitControlVetoReason` 返回的原因
- [x] 4.3 `3fd66a33b`：把 reason 接到 load-batch 的「无可用账号」服务端诊断上；MUST NOT 进客户端响应体
- [x] 4.4 `4795650d`：`service/openai_gateway_service.go` 新增 `ClearActualOpenAIUpstreamEndpoint`
- [x] 4.5 `4795650d`：`service/openai_gateway_cc_pipeline.go` 的 `sendCCUpstreamRequest` 每次发送都 `SetActualOpenAIUpstreamEndpoint(c, "/v1/chat/completions")`
- [x] 4.6 `4795650d`：`Forward` / `ForwardAsChatCompletions` / `ForwardAsAnthropic` 三个入口开头 `Clear...`，并在命中原生 CC 直转门时预置端点。⚠️ 本仓库没有 `shouldForwardOpenAIResponsesViaRawChatCompletions`，用 `openai_gateway_chat_completions.go:88/94/301` 的内联条件，或顺手抽同名函数
- [x] 4.7 `4795650d`：`handler/endpoint.go:312` 的 `GetUpstreamEndpoint` 对 `PlatformOpenAI` / `PlatformGrok` 优先取运行时端点。⚠️ **删掉上游的 `service.IsCNProvider(platform)` 分支**（本仓库无 CN 平台）
- [x] 4.8 落地上游对应的测试改动（`handler/endpoint_test.go`、`service/openai_gateway_responses_chat_fallback_test.go`）

## 5. 阶段 4 — 长上下文门控改 OR（单独提交）

- [x] 5.1 `service/billing_service.go:1103-1105` 按上游形状改：先算 `contextTierPricingEnabled := resolved.longContextPricingEnabled`，账号开关**为真时**置 true，再 `applyLongCtx := len(resolved.Intervals) == 0 && contextTierPricingEnabled`
- [x] 5.2 ⚠️ MUST NOT 把原来那行的 `&&` 直接换成 `||`——那会让账号开关绕过 `len(resolved.Intervals) == 0` 前置条件。见 `source-baseline.md` §2.7
- [x] 5.3 ⚠️ `service/billing_service.go:1045-1046`（`CalculateCostUnified` 无 Resolver 回退路径）**保持原样**，不要一起改
- [x] 5.4 补测试：分组开×账号关 / 分组关×账号开 / 都开 / 都关 / 账号开关缺失 五种组合，外加「有区间定价时不叠倍率」
- [x] 5.5 按 `verification.md` §4 留一条改前改后的费用对比样本

## 6. 阶段 5 — Go 1.27 + golangci-lint v2.13（最后，单独提交）

- [x] 6.1 `backend/go.mod:3` 改 `go 1.27.0`
- [x] 6.2 5 处版本断言全部改：`backend-ci.yml:37`、`backend-ci.yml:77`、`security-scan.yml:26`、`release.yml:99`、`release.yml:134`
- [x] 6.3 `backend-ci.yml:80` 的 golangci-lint `version: v2.9` → `v2.13`
- [x] 6.4 3 个 Dockerfile 的 `ARG GOLANG_IMAGE`：`Dockerfile:11`、`backend/Dockerfile`、`deploy/Dockerfile`
- [x] 6.5 `backend/ent/` **`go generate ./ent` 重新生成并提交**；`group.model_pricing` 与 `usage_cleanup_task.filters` 在 jsonv2 下会变成 `jsontext.Value`。⚠️ MUST NOT 手改生成物
- [x] 6.6 `backend/.golangci.yml` 按上游 `73aabc861` 调整；新规则处理照搬上游：gosec G703/G704、`reflect.Ptr` → `reflect.Pointer`、SA4023 / SA1019 加定点 nolint
- [x] 6.7 lint 修正目标文件：`pkg/servertiming/http.go`、`service/gateway_service.go`、`service/pricing_service.go`、`handler/admin/grok_oauth_handler.go`、`handler/admin/setting_handler_update.go`、`handler/auth_oidc_oauth.go`
- [x] 6.8 `repository/http_upstream_http2_keepalive_test.go`：断言改成校验 `HTTP2Config.SendPingTimeout` / `PingTimeout`，不再断言 `TLSNextProto` 非空
- [x] 6.9 文档版本字样：`README.md:7`（徽章）/ `:109` / `:338`、`README_CN.md`、`README_JA.md`、`CLAUDE.md:21`
- [x] 6.10 `DEV_GUIDE.md:52`：改版本号，**并把「三个 workflow ... 两处」纠正成实际的 5 行断言**

## 7. 门禁

- [x] 7.1 `grep -rn "grokSameAccountRetryMetadata" backend/` 零命中（确认 3.1.8 没被违反）
- [x] 7.2 `grep -rn '"grok-3-mini"' backend/internal/service/billing_service.go` 有命中（价卡确实加了）
- [x] 7.3 `grep -n "grok-3-mini" backend/internal/pkg/xai/models.go` 只在别名表出现，`defaultModels` 里零命中
- [x] 7.4 `grep -n "DefaultTextModel = " backend/internal/pkg/xai/models.go` 为 `"grok-4.6"`
- [x] 7.5 别名收窄条件的最终形态是 `(alias == "grok" || alias == "grok-latest")`，**不含** `grok-build-latest`（3.1.4 → 3.2.2 两步的净结果）
- [x] 7.6 `grep -rn "IsCNProvider" backend/internal/handler/endpoint.go` 零命中
- [x] 7.7 `grep -n "applyLongCtx = applyLongCtx &&" backend/internal/service/billing_service.go` 零命中；`grep -n "contextTierPricingEnabled" ...` 有命中
- [x] 7.8 `grep -rn "go1.26.5\|1\.26\.5" .github/ Dockerfile* deploy/Dockerfile backend/go.mod README*.md DEV_GUIDE.md CLAUDE.md` 零命中
- [x] 7.9 `cd backend && go generate ./ent && git diff --stat backend/ent/` 无输出（生成物已是最新）
- [x] 7.10 确认未改动 `wire_gen.go`、未新增迁移、`VERSION` 未变
- [x] 7.11 确认未改动 `frontend/package.json` / `pnpm-lock.yaml` / `pnpm-workspace.yaml`
- [x] 7.12 逐条勾完 `verification.md` 后，把 `docs/upstream-sync/PORTING-0.1.180.md` §6.2(a) / §6.3 的 `3fd66a33b` / §9.1 / §9.2 / §9.3 与 `PORTING-0.1.183.md` §4.2 的状态改成「已合」/「已决策并落地」
