## 1. 固定基线与边界

- [x] 1.1 起点 `3da1c2dd0`，工作分支 `sync/upstream-20260902`；`cd backend && go build ./...` 通过
      （实际叠在 `sync/upstream-20260902-p1` @ `6527a33de`）
- [x] 1.2 记录基线信号（改动前）：`go test -tags=unit ./... -count=1` 与
      `golangci-lint run ./...` 的 EXIT 与既有失败清单，填 `verification.md` §0
- [x] 1.3 确认 `backend/migrations/` 最大序号仍为 `224`，全程 MUST NOT 新增迁移
      （叠在第二档上，最大号已是 `226`；本批 diff 无迁移文件）
- [x] 1.4 确认 `backend/cmd/server/VERSION` 保持 `1.1.11`，全程不改
- [x] 1.5 ⚠️ 全程 MUST NOT 动 `frontend/`（本批零前端改动）
- [x] 1.6 全程 MUST NOT 动 `wire.go` / `wire_gen.go` / `ent/`（本批不涉及 DI 与 schema）

## 2. 阶段 1 — `e93e6368f` 调度快照投影保留透传开关（**最高优先**）

- [x] 2.1 落地前先复现失效：写一个断言「透传账号经投影后仍支持白名单外模型」的用例，确认它**先失败**
- [x] 2.2 `internal/repository/scheduler_cache.go` 的 `filterSchedulerExtra`：在
      `"openai_responses_supported"` 之后插入 `"openai_passthrough"` / `"openai_oauth_passthrough"`
      两行 + 上游那段中文注释
- [x] 2.3 ⚠️ MUST NOT 顺手补齐 `codex_fingerprint_mode` / `codex_fingerprint_seed`（见 `design.md` 决策 1）
- [x] 2.4 ⚠️ 手写插入，MUST NOT 用模糊匹配硬塞 hunk（上下文因缺指纹两 key 漂移）
- [x] 2.5 落地上游用例 `scheduler_cache_unit_test.go`（+42），**必须含 JSON round-trip 那一跳**
      （见 `design.md` 决策 2）
- [x] 2.6 `go test -tags=unit ./internal/repository/ -run 'SchedulerCache' -count=1`
- [x] 2.7 独立提交

## 3. 阶段 2 — `200b1406d` Anthropic `fallbacks` 剥离

- [x] 3.1 `internal/pkg/claude/constants.go`：新增 server-side-fallback beta 标识常量（+11）
- [x] 3.2 `internal/service/gateway_request.go`：在既有 `context_management` sanitize 同处加
      `fallbacks` 剥离（+77-14）
- [x] 3.3 `internal/service/bedrock_request.go`：Bedrock 路径应用同一规则（+35）
- [x] 3.4 新建 `internal/service/gateway_fallbacks_sanitize_test.go`（照搬上游 +295）
- [x] 3.5 ⚠️ 验收必须双侧：**未带 beta ⇒ 剥离**、**带 beta ⇒ 原样保留**
- [x] 3.6 `go test -tags=unit ./internal/service/ -run 'Fallbacks' -count=1`
- [x] 3.7 独立提交

## 4. 阶段 3 — `1dc0a0900` ctx_pool WS ingress 容量降载改写

- [x] 4.1 `internal/service/openai_ws_forwarder_ingress.go`：`writeClientMessage(upstreamMessage)`
      处（本仓库 `:1039`）引入 `clientMessage := upstreamMessage`，`error` / `response.failed`
      两种事件类型时调 `sanitizeOpenAICapacityShedErrorCodeForClient`，改写后写 `clientMessage`
- [x] 4.2 ⚠️ MUST NOT 原地改 `upstreamMessage`（见 `design.md` 决策 3）
- [x] 4.3 ⚠️ 事件类型只收 `error` / `response.failed`，MUST NOT 扩大范围（决策 4）
- [x] 4.4 新建 `internal/service/openai_ws_ingress_capacity_shed_test.go`（照搬上游 +184）
- [x] 4.5 ⚠️ 验收必须双侧：容量类改写生效 **且** 非容量类原样下发
- [x] 4.6 `go test -tags=unit ./internal/service/ -run 'CapacityShed' -count=1`
- [x] 4.7 独立提交

## 5. 阶段 4 — `6d5f02784` WS 池空闲连接回收

- [x] 5.1 `internal/service/openai_ws_pool.go` 常量块：新增
      `openAIWSConnIdleRecycleAfter = 90 * time.Second` + 上游注释
- [x] 5.2 同文件 `cleanupAccountLocked`：在 `maxAge` 判定**之前**插入回收分支
      （未租出 + `waiters == 0` + `!supportsIdlePingWithoutReader()` + `idleDuration >= 阈值`）
- [x] 5.3 ⚠️ 保留既有 `isConnPinnedLocked` 守卫在前，新分支 MUST NOT 越过它
- [x] 5.4 ⚠️ 上游同时对齐了常量块缩进——那是格式，不要当成语义改动漏掉或误改
- [x] 5.5 落地上游用例 `openai_ws_pool_test.go`（+26）
- [x] 5.6 `go test -tags=unit ./internal/service/ -run 'WSPool|WSConn' -count=1`
- [x] 5.7 独立提交

## 6. 阶段 5 — Codex 目录两条（顺序 `ba345f105` → `57c76584a`）

- [x] 6.1 `ba345f105`：`internal/service/openai_codex_models_service.go` 跳过持久禁用账号（+27-x）
- [x] 6.2 `ba345f105` 的四个用例文件照搬（`handler/gateway_models_test.go`、
      `handler/openai_codex_models_handler_test.go`、`service/openai_codex_model_metadata_test.go`、
      `service/openai_codex_models_service_test.go`）
- [x] 6.3 独立提交
- [x] 6.4 `57c76584a`：同文件 fast 模型透出 priority service tier（+29）+ 用例（+56）
- [x] 6.5 ⚠️ MUST NOT 顺带引入任何发送侧 `service_tier` 行为（0.1.180 §7.2 仍未合）
- [x] 6.6 `go test -tags=unit ./internal/service/ ./internal/handler/ -run 'Codex.*Model|ModelsManifest' -count=1`
- [x] 6.7 独立提交

## 7. 阶段 6 — `9eabd2a5b` + `e7c029875` 账号统计成本

- [x] 7.1 `internal/service/account_stats_pricing.go` 的 `tryModelFilePricing`：整个函数体替换为
      `CalculateCostWithServiceTier(model, tokens, 1, normalizeBillingServiceTier(serviceTier))`，
      失败/空/非正 ⇒ 返回 nil；补上游那段中文注释
- [x] 7.2 ⚠️ `channelPricing` 仍为 nil，MUST NOT 让优先级 3 开始读渠道自定义定价
- [x] 7.3 ⚠️ 手写替换（本仓库条件少 `"fast"` 字面量，见 `source-baseline.md` §2）
- [x] 7.4 `grep -rn` 确认 `GetModelPricing` / `shouldApplySessionLongContextPricing` 仍有其它调用点，
      MUST NOT 顺手删除它们（`design.md` 决策 6）
- [x] 7.5 落地 `e7c029875` 的用例改动（`account_stats_pricing_test.go`，图片输出按 output 子集计价）
- [x] 7.6 ⚠️ 必须验证「标准档结果与改动前一致」这条兼容性底线
- [x] 7.7 `go test -tags=unit ./internal/service/ -run 'AccountStatsCost|AccountStatsRule|CalculateStatsCost' -count=1`
- [x] 7.8 产品码与用例一起提交

## 8. 收尾

- [x] 8.1 `cd backend && go build ./... && go test -tags=unit ./... -count=1 && golangci-lint run ./...`
- [x] 8.2 `go test -tags=unit ./internal/repository/ -run 'TestProductionSQLUsesSQLiteDialect' -count=1`
- [x] 8.3 确认 `git diff --stat` 里没有 `frontend/`、`ent/`、`migrations/`、`VERSION`、`wire_gen.go`
- [x] 8.4 填 `verification.md`（含 §0 基线、§2 逐项证据、§3 影响面复核）
- [x] 8.5 `docs/upstream-sync/PORTING-0.2.0.md`：§3.1–§3.7 逐条标「已合」+ 本仓库短 SHA；
      §7 第一批打勾
- [x] 8.6 `docs/upstream-sync/README.md` 顶部「当前待移植清单」段落补一句本批落点
- [ ] 8.7 合回 `main`（`--no-ff`），合并后在 `main` 上复跑 8.1 三道门禁并回填 SHA
