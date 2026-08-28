# 追踪表：条目 → Requirement → 目标代码 → 证据

7 个条目（4 项 P1 + 3 个决策，共 6 个上游 commit + 2 处本仓库自有改动）与 6 个 capability 的双向覆盖。

| # | 条目 | 来源 | capability / Requirement | 目标代码 | 证据 |
|---|---|---|---|---|---|
| 1 | 修过期依赖审计例外 | 本仓库自有 | `dependency-audit-hygiene` / 两条 Requirement | `.github/audit-exceptions.yml`（`lodash` / `lodash-es` / `axios` 三条） | §2 审计行 + §3.1 |
| 2 | Grok 目录计费与别名 | `ed4207a16`（0.1.180 §6.2a） | `grok-model-catalog-billing` / 前三条 Requirement | `pkg/xai/models.go:54,58-62,89-90,116-117,161,272`、`service/billing_service.go:704,928`；⚠️ **丢弃 `openai_gateway_grok.go` hunk** | §3.2 计费专项 |
| 3 | 调度 veto reason | `3fd66a33b`（0.1.180 §6.3） | `scheduling-veto-diagnostics` / 两条 Requirement | `service/openai_gateway_scheduling.go:360` 起 | §2 调度行 + §3.3 行为等价 |
| 4 | 真实上游端点 | `4795650d`（0.1.183 §4.2） | `upstream-endpoint-attribution` / 两条 Requirement | `handler/endpoint.go:312`、`service/openai_gateway_{service,cc_pipeline,chat_completions,forward,messages}.go` | §2 端点行 |
| 5 | **决策 9.1** Grok 默认 4.6 | `39485f2e2` + `f7145c750` | `grok-model-catalog-billing` / 后两条 Requirement | `pkg/xai/models.go:54,108-109,161,272`、`service/billing_service.go`、`service/setting_parse.go`、`service/setting_update.go`、`service/setting_gateway_runtime.go`（新增迁移）、`service/wire.go:745`、`frontend/src/composables/useModelWhitelist.ts` | §2 默认模型行 + §3.4 迁移专项 |
| 6 | **决策 9.2** Go 1.27 | `cbe258fd1` + `73aabc861` | `go-toolchain-127` / 四条 Requirement | `backend/go.mod:3`、5 处 workflow 断言、`backend-ci.yml:80`、3 个 Dockerfile、`backend/.golangci.yml`、`backend/ent/`（7 文件，生成）、`repository/http_upstream_http2_keepalive_test.go`、`pkg/servertiming/http.go`、`service/{gateway_service,pricing_service}.go`、3 个 handler、5 份文档 | §3.5 工具链专项 |
| 7 | **决策 9.3** 长上下文 OR | 本仓库自有（对齐上游 v0.1.179+） | `long-context-billing-gating` / 三条 Requirement | `service/billing_service.go:1103-1105`（改）、`:1045-1046`（**不改**） | §4 计费口径专项 |

## 反向核对：capability → 条目

| capability | Requirement 数 | 来源条目 |
|---|---|---|
| `dependency-audit-hygiene` | 2 | 1 |
| `grok-model-catalog-billing` | 5 | 2（前 3 条）+ 5（后 2 条） |
| `scheduling-veto-diagnostics` | 2 | 3 |
| `upstream-endpoint-attribution` | 2 | 4 |
| `go-toolchain-127` | 4 | 6 |
| `long-context-billing-gating` | 3 | 7 |

合计 18 条 Requirement / 60 个 Scenario，覆盖全部 7 个条目。

## 顺序约束（唯一的硬依赖链）

```text
条目 2 (ed4207a16)  →  条目 5 前半 (39485f2e2)  →  条目 5 后半 (f7145c750)
   改别名收窄条件         再收窄 + 重排价卡 switch        39485f2e2 的数据迁移
```

其余条目之间无依赖，可任意顺序；条目 6（Go 1.27）建议最后，理由见 `design.md` 决策 6。

## 明确不在本 change 内

| 内容 | 判定 |
|---|---|
| `4a1da2950` dompurify、`b410c3913` nanoid 例外 | 上一批已决定推迟，本批只**续期已有**例外，不动依赖 |
| 0.1.180 §6.1 工具桥接 11 条 | 下一批候选；`0.1.183` 的 Responses Lite 簇在等它 |
| 0.1.180 §6.2(c) Grok 429 / Realtime / 容量串 | `ed4207a16` 里引用 `grokSameAccountRetryMetadata` 的那个 hunk 属于这簇，本批丢弃 |
| 0.1.180 §6.3 除 `3fd66a33b` 外的 3 项（`68653fb2c` / `d5824f6a5` / `d493ce0bb`+`fa4587041`） | 按需 |
| 0.1.180 §7.x 全部 | P2 |
| 0.1.183 §4.1 监控 v2 composite | 需 SQLite 重写，另开 |
| 0.1.183 §5.x | 缺基座 / 不成立 / 上游返工中 |
