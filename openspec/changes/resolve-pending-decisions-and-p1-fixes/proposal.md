## Why

两批 P0 已落地（0.1.180 的 19 项 = `1b4837098`；0.1.183 的 12 项在工作区）。剩余 backlog 里，
四项 P1 的性价比清楚，三个决策拖着的成本在涨。这一批把它们一次做完。

**四项 P1：**

- **依赖审计门禁已经失灵**。`.github/audit-exceptions.yml` 里 `lodash` / `lodash-es`
  （到期 `2026-07-02`）与 `axios`（`2026-07-10`）三条例外**已过期**，而
  `tools/check_pnpm_audit_exceptions.py` 对过期条目同样 `return 1` ⇒ `security-scan.yml`
  大概已经红了 8 周。这条排最前，因为上一批**推迟 dompurify 的前提就是「靠审计信号盯着」**，
  而信号现在是常红的——等于收不到任何新公告通知。
- **`grok-3-mini` 系列在超收**（已在 `1b4837098` 之后重新验证）：两个模型都在
  `pkg/xai/models.go:89/90` 的 `defaultModels` 里对客户端可见，但 `billing_service.go` 没有它们的
  价卡；`isGrokUnknownTextFamilyModel("grok-3-mini")` 因「`grok-` 后首字符是数字」返回 true，
  兜到 `billing_service.go:861` 的 `fallbackPrices["grok-4.5"]`（$2 / $6）。官方价是
  `grok-3-mini` $0.30 / $0.50、`grok-3-mini-fast` $0.60 / $4.00
  ⇒ **前者输入超收 6.7×、输出超收 12×；后者输入 3.3×、输出 1.5×**。
  同一 commit 还修另一个真 bug：别名被无条件重映射，客户端**显式**要 `grok-4.5`
  也会被改写成运营方配的默认模型。
- **调度「无可用账号」只有 boolean**。以 `c35a482` 那次静默调度事故的经历，
  拿不到具体 veto reason 意味着下一次同类问题还得从零查。而且这一条改
  `openai_gateway_scheduling.go`——刚在两批 P0 里动过该文件，**现在合冲突最小**。
- **错误日志把入站端点误报成上游端点**。`force_chat_completions` 生效时请求实际发往
  `/v1/chat/completions`，但 503 / 传输失败这类没有 `OpenAIForwardResult` 的路径会记成
  `/v1/responses`，排查时指向错误的方向。

**三个决策：**

- **9.1 Grok 默认文本模型跟到 4.6**。0.1.176 清单当时明确要求保持 4.5，`pkg/xai/models.go:54`
  至今是 `const DefaultTextModel = "grok-4.5"`。跟上游后代码默认与存量 DB 设置必须一起改，
  因此 `39485f2e2` 与 `f7145c750` 成对。
- **9.2 Go 跟到 1.27.0**。不跟的复利成本是每轮移植都要在两套 ent 生成结果之间手工调和。
- **9.3 长上下文计费门控改成 OR**。对齐上游 v0.1.179 起的语义：分组开关是统一入口，
  账号 API 开关只能额外开启、不能否决分组配置。

## What Changes

- `.github/audit-exceptions.yml` 的三条过期例外重新核实并续期（或改为已解决）；清单中 MUST NOT
  再有过期条目。**不动 `package.json` / `pnpm-lock.yaml`**（依赖升级仍按上一批的推迟决定）。
- 新增 `grok-3-mini` / `grok-3-mini-fast` 官方价卡与 `getFallbackPricing` 分支；把这两个 ID 从
  `/v1/models` 的 `defaultModels` 摘掉，只保留别名与价卡。
- 别名重映射收窄：只有 `grok` / `grok-latest`（9.1 之后不含 `grok-build-latest`）跟随运行时默认，
  显式 `grok-4.5` / `grok-4.5-latest` 原样透传。
- 修正 Imagine 视频模型 ID 反了的问题（`grok-imagine-video-1.5` 与 `-preview` 互换），
  新增 `grok-imagine-image-2.0`。
- 调度准入的 boolean 门改成返回具体 veto reason（`platform_mismatch` / `model_rate_limited` /
  `not_schedulable` / `quota_auto_pause_<window>` / `model_not_supported` / `capability_mismatch` /
  `compact_unsupported` / 利润控制原因）。**准入行为不变，只加可观测性。**
- `GetUpstreamEndpoint` 优先采信 OpenAI 转发服务记录的运行时端点；新增
  `ClearActualOpenAIUpstreamEndpoint` 清理 failover 尝试之间的残留；CC 发送处每次都记录端点。
- `DefaultTextModel` 改 `grok-4.6`；补齐官方计费目录（新增 `grok-4.20` 独立价卡，
  从 grok-4.3 拆出）；`grok-build-latest` 不再跟随运行时默认而是固定 `grok-build-0.1`；
  `setting_parse.go` / `setting_update.go` / 前端模型白名单同步。
- 新增 `SettingService.MigrateGrokDefaultTextModel`：启动时把存量 DB 里显式存着 `grok-4.5` 的
  `grok_default_text_model` 重写成 `grok-4.6`；未写过（`ErrSettingNotFound`）为 no-op，
  其他值视为运营方显式选择、不动。
- Go `1.26.5` → `1.27.0`：`backend/go.mod`、**5 处** workflow 版本断言、3 个 Dockerfile 的
  golang 镜像、README/README_CN/README_JA/DEV_GUIDE/CLAUDE.md 的版本字样；
  golangci-lint `v2.9` → `v2.13` 并处理新规则（gosec G703/G704、`reflect.Ptr` → `reflect.Pointer`、
  SA4023 / SA1019 加 nolint）；**`backend/ent/` 在 jsonv2 默认引擎下重新生成**。
- 长上下文门控改成：分组开关为主，账号开关只能额外开启。**区间定价的
  `len(resolved.Intervals) == 0` 前置条件仍是 AND**，不能一起改成 OR。

## Capabilities

### New Capabilities

- `dependency-audit-hygiene`: 依赖审计例外清单不得含过期条目，以及门禁必须真实反映依赖风险。
- `grok-model-catalog-billing`: Grok 文本/媒体模型目录、别名跟随规则与价卡的对应关系，含默认模型迁移。
- `scheduling-veto-diagnostics`: 调度准入失败必须给出具体原因，且准入行为不得改变。
- `upstream-endpoint-attribution`: 错误与用量记录必须归因到真实上游端点。
- `go-toolchain-127`: Go 与 lint 工具链版本的一致性要求，含生成代码与 keepalive 断言。
- `long-context-billing-gating`: 长上下文阶梯定价的开关组合语义。

### Modified Capabilities

无。`openspec/` 下没有已发布的 capability 基线。

## Impact

- **后端**：`internal/pkg/xai/models.go`、`internal/service/{billing_service,openai_gateway_scheduling,
  setting_gateway_runtime,setting_parse,setting_update,gateway_service,pricing_service,
  openai_gateway_cc_pipeline,openai_gateway_chat_completions,openai_gateway_forward,
  openai_gateway_messages,openai_gateway_service,wire}.go`、`internal/handler/endpoint.go`、
  `internal/pkg/servertiming/http.go`、3 个 handler 文件（lint 修正）。
- **生成代码**：`backend/ent/`（group、usagecleanuptask、mutation）必须 `go generate ./ent` 重新生成。
  ⚠️ **不要手改**。
- **前端**：`src/composables/useModelWhitelist.ts`。
- **CI / 构建**：3 个 workflow、`Dockerfile`、`backend/Dockerfile`、`deploy/Dockerfile`、
  `backend/.golangci.yml`、`.github/audit-exceptions.yml`。
- **文档**：`README.md` / `README_CN.md` / `README_JA.md` / `DEV_GUIDE.md` / `CLAUDE.md` 的版本字样。
- **数据库**：无迁移。迁移号保持 `224`。`f7145c750` 只改 `settings` 表里一个键的值，
  且仅当它显式等于 `grok-4.5`。
- **wire**：`internal/service/wire.go` 只在 `ProvideSettingService` 体内多一次调用，**不改 provider
  签名 ⇒ 不需要 `go generate ./cmd/server`**。
- **⚠️ 计费（往上）**：9.3 改成 OR 后，超过长上下文阈值的请求只要**分组**开关开着就会被收
  2× 输入 / 1.5× 输出（此前还需账号开关也开着）。分组开关默认开、账号开关默认关 ⇒
  **既有部署的这类请求费用会上升**。想保持旧行为的分组，把它的
  `long_context_pricing_enabled` 关掉。见 `design.md` 决策 5。
- **计费（往下）**：`grok-3-mini` / `grok-3-mini-fast` 从超收回到官方价。
- **对客户端可见**：`grok-3-mini` / `grok-3-mini-fast` 不再出现在 `/v1/models`（仍可请求，
  按别名 + 官方价卡计费）；默认模型请求（`grok` / `grok-latest` / 空值）从 4.5 改到 4.6。
- **兼容性**：显式请求 `grok-4.5` 的客户端行为**不变**（这正是 `ed4207a16` 修的 bug）。
  `grok-build-latest` 从「跟随运行时默认」改为固定 `grok-build-0.1`。

## Execution References

- `source-baseline.md`：6 个上游 commit 的固定 SHA、按文件三态、依赖符号核查、顺序实测。
- `source-feature-map.md`：条目 → Requirement → 目标代码 → 证据的双向追踪。
- `design.md`：七项的逐条决策、顺序约束、风险与阶段划分。
- `docs/upstream-sync/PORTING-0.1.180.md` §6.2 / §6.3 / §9，`PORTING-0.1.183.md` §4.2。
