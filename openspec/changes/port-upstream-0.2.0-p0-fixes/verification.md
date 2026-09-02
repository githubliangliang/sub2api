# 验收证据

**状态：未开始。** 下面每一格在实施时填入真实命令输出摘要与提交 SHA；MUST NOT 预填。

## 0. 基线信号（改动前，`3da1c2dd0` 干净工作区）

| 门禁 | 命令 | 结果 |
|---|---|---|
| 后端编译 | `cd backend && go build ./...` | 待填 |
| 后端 unit | `go test -tags=unit ./... -count=1` | 待填（记录既有 flake / 既有失败清单） |
| lint | `golangci-lint run ./...` | 待填 |
| SQLite 方言审计 | `go test -tags=unit ./internal/repository/ -run 'TestProductionSQLUsesSQLiteDialect' -count=1` | 待填 |

已核过、可直接引用的基线事实（2026-09-02）：

- `go test -tags=unit ./internal/repository/ -run 'SchedulerCache|TestProductionSQLUsesSQLiteDialect' -count=1` ⇒ **ok**
- `go test -tags=unit ./internal/service/ -run 'AccountStatsCost|AccountStatsRule|CalculateStatsCost|FallbackPricing' -count=1` ⇒ **ok**
- `backend/migrations/` 最大序号 = `224`；`backend/cmd/server/VERSION` = `1.1.11`

## 1. 交付后的全量信号

| 门禁 | 命令 | 结果 |
|---|---|---|
| 后端编译 | `cd backend && go build ./...` | 待填 |
| 后端 unit | `go test -tags=unit ./... -count=1` | 待填（与 §0 逐条对比，MUST NOT 出现新增失败） |
| lint | `golangci-lint run ./...` | 待填 |
| 迁移号未变 | `ls backend/migrations/ \| sort \| tail -1` | 待填（MUST 是 `224_*`） |
| 无越界改动 | `git diff --stat <base>..HEAD` | 待填（MUST NOT 含 `frontend/` / `ent/` / `migrations/` / `VERSION` / `wire_gen.go`） |

## 2. 逐项证据

### 2.1 `e93e6368f` 调度快照投影 — commit 待填

**必须先复现失效再验修复**（这是本批唯一一条活缺陷）。

| 项 | 证据 |
|---|---|
| 失效复现 | 待填：改动前，透传账号经投影 round-trip 后被 `IsModelSupported` 判为不支持白名单外模型 |
| 修复 | 待填：改动后同一用例通过 |
| round-trip 覆盖 | 待填：用例确实经过 JSON 序列化/反序列化，而非只断言 `filterSchedulerExtra` 返回值 |
| 兼容字段 | 待填：只设 `openai_oauth_passthrough` 的账号同样放行 |
| 非透传不放宽 | 待填：未设开关的账号仍按 `model_mapping` 判定 |
| 白名单未越界 | 待填：`git diff` 显示 `keys` 列表只 +2 行，`codex_fingerprint_*` 仍不存在 |

### 2.2 `200b1406d` Anthropic `fallbacks` — commit 待填

| 项 | 证据 |
|---|---|
| 未带 beta ⇒ 剥离 | 待填 |
| 带 beta ⇒ 保留 | 待填 |
| OAuth mimic 路径 | 待填 |
| Bedrock 路径 | 待填 |
| 无 `fallbacks` 时请求体不变形 | 待填 |

### 2.3 `1dc0a0900` ctx_pool ingress — commit 待填

| 项 | 证据 |
|---|---|
| 容量类改写生效 | 待填（`server_is_overloaded` / `slow_down` 两种） |
| 非容量类原样下发 | 待填 |
| `response.failed` 形态 | 待填 |
| 账号状态判定用原始 payload | 待填：`git diff` 显示 `upstreamMessage` 未被原地改写 |
| 与另两条路径语义一致 | 待填：与 `openai_gateway_response_handling.go` / `openai_ws_http_bridge.go` 的调用形状对比 |

### 2.4 `6d5f02784` WS 池空闲回收 — commit 待填

| 项 | 证据 |
|---|---|
| 不可保活 + 空闲达阈值 ⇒ 逐出 | 待填 |
| 可保活连接不受影响 | 待填 |
| 已租出 / 有 waiter 不逐出 | 待填 |
| pinned 守卫仍在前 | 待填 |
| 指标计入 | 待填 |

### 2.5 `ba345f105` + `57c76584a` Codex 目录 — commit 待填 / 待填

| 项 | 证据 |
|---|---|
| 持久禁用账号被跳过 | 待填 |
| 临时不可调度不被误排除 | 待填 |
| fast 模型透出 priority tier | 待填 |
| 非 fast 条目不变 | 待填 |
| 未引入发送侧 `service_tier` 行为 | 待填：`git diff` 不含请求装配路径 |

### 2.6 `9eabd2a5b` + `e7c029875` 账号统计成本 — commit 待填

| 项 | 证据 |
|---|---|
| **标准档结果与改动前一致** | 待填（本条的兼容性底线，必须有数值对比） |
| `service_tier = "fast"` 不再按标准价 | 待填 |
| `priority` / `flex` 行为不变 | 待填 |
| 长上下文行为不变 | 待填 |
| 图片输出按 output 子集 | 待填 |
| 渠道定价未泄漏进优先级 3 | 待填 |
| 未误删 `GetModelPricing` / `shouldApplySessionLongContextPricing` | 待填：`grep -rn` 结果 |

## 3. 影响面复核

| 面 | 复核项 | 结果 |
|---|---|---|
| 调度 | 第 1 项会放大候选集：此前被误剔的透传账号开始参与选号。确认放大后的候选集里没有本该被其它规则挡住的账号 | 待填 |
| 计费 | 第 6 项只在此前被手算分支绕过的档位上产生差异，方向是修正为更高的真实成本 | 待填 |
| 客户端可见行为 | 第 3 项改的是发给客户端的 payload；第 2 项改的是发给上游的请求体。两者各自的「不该改的那一侧」有反例覆盖 | 待填 |
| 数据库 | 无迁移、无 schema 改动 | 待填 |
| 配置 | 无新增配置项、无默认值变化 | 待填 |
| 前端 | 零改动 | 待填 |

## 4. 上线后观察

- [ ] 观察 `model_not_supported` 剔除计数是否下降（第 1 项的直接指标）
- [ ] 观察是否还有 Codex 客户端报 "Selected model is at capacity" 后直接终止（第 3 项）
- [ ] 观察 WS 池 scale-down 计数与取连接失败率（第 4 项）
- [ ] 观察带 `fallbacks` 的 Anthropic 请求是否还有 400（第 2 项）
