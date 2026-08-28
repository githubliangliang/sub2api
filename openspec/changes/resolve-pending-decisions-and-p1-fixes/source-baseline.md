# 源基线固定

核查日期：2026-08-27。本仓库基线：`1b4837098`（0.1.180 P0）+ 工作区里 0.1.183 那 12 项。
`go build ./...` 已通过。**实施期间不得改动本文的 SHA。**

## 1. 固定点

| 点 | 值 |
|---|---|
| 上游仓库 | `https://github.com/Wei-Shaw/sub2api` |
| 上游最新正式版 | `v0.1.183` = `c21fd338`（2026-08-27 复查 tags，**仍是最新，无新版本**） |
| 本批 6 个 commit 所属 | `ed4207a16` `3fd66a33b` `39485f2e2` `f7145c750` `cbe258fd1` `73aabc861` 均在 `v0.1.180`；`4795650d` 在 v0.1.183 之后的 main |
| 本仓库版本号 | `backend/cmd/server/VERSION` = `1.1.8`（不同步上游） |
| 本仓库迁移号 | `224`（本批**无迁移**） |
| 本仓库 Go | `backend/go.mod:3` = `go 1.26.5` |
| 本仓库 golangci-lint | `.github/workflows/backend-ci.yml:80` = `version: v2.9` |
| 本仓库 x/net | `backend/go.mod:54` = `golang.org/x/net v0.56.0`（已是 9.2 相关行为涉及的版本） |

取 patch（部分裸克隆对这批较早提交逐 blob 拉取太慢，走 HTTPS）：

```bash
curl -sSL -o /tmp/pcnew/<sha>.patch "https://github.com/Wei-Shaw/sub2api/commit/<sha>.patch"
curl -sSL "https://raw.githubusercontent.com/Wei-Shaw/sub2api/v0.1.183/backend/<path>"
```

## 2. 逐条核查结论

### 2.1 依赖审计例外（本仓库自有，无上游 commit）

`.github/audit-exceptions.yml` 共 5 条，逐条算过期状态（今天 2026-08-27）：

| 包 | 公告 | expires_on | 状态 |
|---|---|---|---|
| `xlsx` | GHSA-4r6h-8v6p-xvw6 | 2026-10-06 | ok |
| `xlsx` | GHSA-5pgg-2g8v-p4x9 | 2026-10-06 | ok |
| `lodash` | GHSA-r5fr-rjxr-66jc | 2026-07-02 | **已过期 56 天** |
| `lodash-es` | GHSA-r5fr-rjxr-66jc | 2026-07-02 | **已过期 56 天** |
| `axios` | GHSA-3p68-rc4w-qgx5 | 2026-07-10 | **已过期 48 天** |

`tools/check_pnpm_audit_exceptions.py:216-240`：过期条目进 `expired_exceptions` → `errors` →
`return 1`。消费点是 `.github/workflows/security-scan.yml:56-59`。⇒ **门禁已红约 8 周。**

现状供续期时填 reason 用：`axios` 是直接依赖（`frontend/package.json:23` = `^1.18.0`，lock 解析到
`1.18.1`）；`lodash` / `lodash-es` 在 `frontend/src/` 里 **grep 不到直接 import**，是传递依赖
（lock 里 `lodash@4.17.21` / `lodash@4.18.1` / `lodash-es@4.18.1`）。

### 2.2 `ed4207a16` Grok 目录计费 —— **部分可移植，必须裁**

| 文件 | 三态 | 说明 |
|---|---|---|
| `backend/internal/pkg/xai/models.go` | 可移植 | 本仓库 `:54` `:58-62` `:89-90` `:116-117` 均是改前形态 |
| `backend/internal/service/billing_service.go` | 可移植 | 新增两张价卡 + `getFallbackPricing` 两个 case |
| `backend/internal/service/openai_gateway_grok.go` | ❌ **不可移植** | 该 hunk 调用 `grokSameAccountRetryMetadata`，本仓库 **grep 零命中**（属 0.1.180 §6.2(c) Grok 429 同号重试簇，不在本批）⇒ **整个文件的 hunk 丢弃** |

超收已重新验证（`1b4837098` 之后）：`grok-3-mini` 与 `grok-3-mini-fast` 在
`pkg/xai/models.go:89/90` 的 `defaultModels` 里；`billing_service.go` grep `grok-3` **零命中**；
`isGrokUnknownTextFamilyModel`（`billing_service.go:864`）对 `grok-` 后首字符为数字的 ID 返回 true；
`grokUnknownTextFamilyFallback`（`:857`）返回 `fallbackPrices["grok-4.5"]`（`:592` = $2 / $6）。

### 2.3 `3fd66a33b` 调度诊断 —— 依赖齐全

2 文件 +118-16。依赖符号全部存在：`openAIProfitControlVetoReason`
（`openai_profit_control.go`）、`isOpenAICompatibleAccountEligibleForRequestBeforeProfit` 与
`shouldAutoPauseOpenAIAccountByQuota`（均在 `openai_gateway_scheduling.go`）、
`reason.window`（`:341` `:352` 已在用）。

### 2.4 `4795650d` 真实上游端点 —— 需裁两处

上一轮（`PORTING-0.1.183.md` §4.2）已核实：本仓库无 `service.IsCNProvider`（无 CN 平台）
⇒ 删该分支；无 `shouldForwardOpenAIResponsesViaRawChatCompletions`（上游自己抽的门函数），
本仓库的门内联在 `openai_gateway_chat_completions.go:88/94/301` ⇒ 按内联条件写或顺手抽同名函数。
`SetActualOpenAIUpstreamEndpoint` 已存在（`openai_gateway_service.go:319`）。

### 2.5 `39485f2e2` + `f7145c750` Grok 4.6 —— 顺序对 `ed4207a16` 有硬依赖

⚠️ **`39485f2e2` 会改 `ed4207a16` 刚引入的行**，实测确认的重叠点：

- 别名收窄条件：`ed4207a16` 写成 `(alias == "grok" || alias == "grok-latest" || alias == "grok-build-latest")`，
  `39485f2e2` 再收窄成 `(alias == "grok" || alias == "grok-latest")`，并把
  `"grok-build-latest"` 的目标从 `DefaultTextModel` 改成 `"grok-build-0.1"`。
- `getFallbackPricing` 的 switch：`ed4207a16` 往里加两个 case，`39485f2e2` 重排整段并把
  `grok-4.20-*` 从 grok-4.3 拆到新的 `grok-4.20` 价卡。

⇒ **顺序固定 `ed4207a16` → `39485f2e2` → `f7145c750`**，反了要手工解冲突。

`f7145c750` 依赖符号全部存在：`SettingKeyGrokDefaultTextModel`（`setting_parse.go` /
`domain_constants.go`）、`codexRestrictionPolicyDBTimeout`（`setting_gateway_runtime.go`）、
`ErrSettingNotFound`。它对 `internal/service/wire.go` 的改动只是在
`ProvideSettingService` 体内加一次调用，**不改 provider 签名 ⇒ 不需要
`go generate ./cmd/server`**（README 硬约束第 7 条在这里不触发）。

### 2.6 `cbe258fd1` + `73aabc861` Go 1.27 —— 目标文件全部存在

本仓库需要动的点已逐一确认存在：

| 类别 | 位置 |
|---|---|
| go.mod | `backend/go.mod:3` |
| **版本断言（5 处，不是 3 处）** | `backend-ci.yml:37`、`backend-ci.yml:77`、`security-scan.yml:26`、`release.yml:99`、`release.yml:134` |
| golangci-lint pin | `backend-ci.yml:80` `version: v2.9` |
| golang 镜像（3 个） | `Dockerfile:11`、`backend/Dockerfile`、`deploy/Dockerfile`（均 `ARG GOLANG_IMAGE=golang:1.26.5-alpine`） |
| lint 配置 | `backend/.golangci.yml`（`version: "2"`） |
| ent 生成物 | `backend/ent/{group,group_create,group_update,mutation,usagecleanuptask,usagecleanuptask_create,usagecleanuptask_update}.go` |
| keepalive 断言 | `backend/internal/repository/http_upstream_http2_keepalive_test.go` |
| lint 修正目标 | `internal/pkg/servertiming/http.go`、`internal/service/{gateway_service,pricing_service}.go`、`internal/handler/admin/{grok_oauth_handler,setting_handler_update}.go`、`internal/handler/auth_oidc_oauth.go` |
| 文档版本字样 | `README.md:7/109/338`、`README_CN.md`、`README_JA.md`、`DEV_GUIDE.md:52`、`CLAUDE.md:21` |

⚠️ `DEV_GUIDE.md:52` 现在写的是「三个 workflow ... 两处」，实际断言有 **5 行**，改的时候一并纠正。

### 2.7 长上下文门控 —— 上游权威写法（不要自己发明）

本仓库 `billing_service.go:1103-1105` 现状：

```go
applyLongCtx := len(resolved.Intervals) == 0 && resolved.longContextPricingEnabled
if input.LongContextBillingEnabled != nil {
    applyLongCtx = applyLongCtx && *input.LongContextBillingEnabled
}
```

上游 `v0.1.183` 的 `billing_service.go:1232-1252`（已抓下来对照）：

```go
// 分组开关是统一入口；账号 API 开关保留为额外开启能力，但 false 不否决分组配置。
contextTierPricingEnabled := resolved.longContextPricingEnabled
if input.LongContextBillingEnabled != nil && *input.LongContextBillingEnabled {
    contextTierPricingEnabled = true
}
...
applyLongCtx := len(resolved.Intervals) == 0 && contextTierPricingEnabled
```

⚠️ **把现有那行的 `&&` 直接换成 `||` 是错的**：那会让账号开关绕过
`len(resolved.Intervals) == 0` 这个前置条件（区间定价已自含上下文分层，不能再叠长上下文倍率）。
必须按上游形状把「分组 OR 账号」和「无区间定价」两件事分开。

另有第二处 `billing_service.go:1045-1046`（`CalculateCostUnified` 的无 Resolver 回退路径）：
那里拿不到分组开关，账号开关是唯一信号，**保持原样**，不要一起改。

## 3. 与前两批的关系

- 本批的 `3fd66a33b` 与两批 P0 都改 `openai_gateway_scheduling.go`（不同函数）。P0 已落地，
  现在合冲突最小；再拖会和 0.1.180 §7.3（重置卡自动使用）抢同一文件。
- 本批**不包含** 0.1.180 §6.1（工具桥接 11 条）、§6.2(c)（Grok 429/Realtime 串）、§7.x，
  以及 0.1.183 §4.1（监控 v2 composite，需 SQLite 重写）、§5.x。
- 上一批推迟的 dompurify（`4a1da2950`）与 nanoid 例外（`b410c3913`）**仍然推迟**；
  本批第 1 项只续期已有例外，**不动 `package.json` / `pnpm-lock.yaml`**。
