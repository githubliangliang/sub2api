# 验收

原则：验收 `specs/*/spec.md` 的 Requirement 是否成立。本批有三项特殊要求：

- **条目 3（调度诊断）** 是纯可观测性 ⇒ 验收重点是**行为等价**（§3.3），不是「有 reason」。
- **条目 7（长上下文 OR）** 会让费用**上升** ⇒ 必须留改前改后对比样本（§4）。
- **条目 6（Go 1.27）** 会重写生成代码 ⇒ 验收含「生成器幂等」（§3.5）。

## 1. 全量信号（每阶段结束都跑）

```bash
cd backend
go build ./...
go test -tags=unit ./... -count=1
golangci-lint run ./...
cd ../frontend && pnpm run typecheck && pnpm run lint:check
cd .. && make test-frontend-critical
```

阶段 5 之后 `golangci-lint` 用的是 v2.13，与 §1.3 记录的 v2.9 基线对照看新增告警。
已知与本批无关的失败：前端全量 `pnpm run test:run` 下 `useRoutePrefetch.spec.ts` 5 条。

## 2. 证据矩阵

| Requirement | 证据 |
|---|---|
| 例外清单无过期条目 | `python tools/check_pnpm_audit_exceptions.py --audit /tmp/audit.json --exceptions .github/audit-exceptions.yml` 返回 0；见 §3.1 |
| 例外条目理由可复核 | 三条的 `reason` 均记录了本次核实结论与「直接/传递依赖」判定；`owner` 非占位值 |
| 每个可见 Grok 文本模型都有价卡 | 见 §3.2 |
| 只有跟随型别名被重写 | 上游 `pkg/xai/models_test.go` 改动移入；额外自补「运行时默认设为 4.3 时显式请求 `grok-4.5` 仍得 4.5」 |
| Imagine ID 与官方一致 | 上游 `pkg/xai/models_test.go` 改动移入 |
| 默认模型 4.6 + 计费目录对齐 | 上游 `billing_service_test.go`、`pkg/xai/oauth_test.go`、`useModelWhitelist.spec.ts` 改动移入；额外断言 `grok-4.20-*` 用新卡而非 4.3 卡 |
| 存量设置迁移只改旧默认值 | 见 §3.4 |
| 调度失败给出具体原因 | 上游 `openai_gateway_scheduling` 相关测试改动移入；覆盖「整体可调度但按模型限流」与「配额暂停带窗口」两个易混点 |
| 诊断不改准入行为 | 见 §3.3 |
| 归因到真实上游端点 | 上游 `handler/endpoint_test.go`、`service/openai_gateway_responses_chat_fallback_test.go` 改动移入 |
| 端点不在 failover 间残留 | 自补用例：两次尝试走不同协议，断言最终记录是第二次的 |
| 分组为主 / 账号只能额外开启 | 见 §4；五种开关组合全覆盖 |
| 区间定价前置条件仍是 AND | 见 §4.3 |
| 无 Resolver 回退路径不变 | 该路径既有测试 MUST 全绿，且 `git diff` 显示 `:1045-1046` 未被改动 |
| Go 版本与断言一致 | 见 §3.5 |
| lint 版本兼容 | `golangci-lint run ./...` 在 v2.13 下通过 |
| 生成代码由生成器产出 | 见 §3.5 |
| HTTP/2 keepalive 断言 | `go test -tags=unit ./internal/repository/ -run HTTP2Keepalive -count=1` |

## 3. 专项

### 3.1 依赖审计门禁（阶段 1）

- [x] 3.1.1 改动**前**跑一次，确认失败信息里就是 `lodash` / `lodash-es` / `axios` 三条过期
- [x] 3.1.2 记录「除这三条外还有没有别的红」——有的话它是独立问题，不要混进本批修
- [x] 3.1.3 改动后门禁返回 0
- [x] 3.1.4 三条的新 `expires_on` 均在未来，且不晚于下一次计划性依赖维护

### 3.2 Grok 计费专项（阶段 2，涉钱）

- [x] 3.2.1 改动**前**：对 `grok-3-mini` 算一次费用，确认命中 grok-4.5 卡（$2 / $6）
- [x] 3.2.2 改动后：同一 token 数下费用 MUST 降到 $0.30 / $0.50 口径（输入约 1/6.7、输出约 1/12）
- [x] 3.2.3 `grok-3-mini-fast`：从 $2 / $6 降到 $0.60 / $4.00（输入约 1/3.3、输出约 1/1.5）
- [x] 3.2.4 `/v1/models` 响应里这两个 ID 已消失，但直接请求仍可用
- [x] 3.2.5 显式 `grok-4.5` 与 `grok-4.6` 各自的费用未被互相影响
- [x] 3.2.6 一个既无价卡也无别名的 Grok 文本 ID 仍回落未知族兜底卡

### 3.3 调度诊断行为等价（阶段 3）

- [x] 3.3.1 构造一组覆盖各否决点的账号，记录改动**前**每个账号的准入结果（通过/排除）
- [x] 3.3.2 改动后逐个比对，结果 MUST 完全一致
- [x] 3.3.3 一个账号同时违反多个条件时，返回的是**最先**触发的门（短路顺序未变）
- [x] 3.3.4 客户端收到的「无可用账号」错误体里 MUST NOT 出现内部 reason 字符串

### 3.4 默认模型迁移专项（阶段 2）

- [x] 3.4.1 设置显式为 `grok-4.5`（含前后空白变体）⇒ 启动后变 `grok-4.6`
- [x] 3.4.2 设置不存在 ⇒ 启动后仍不存在，未被写入
- [x] 3.4.3 设置为 `grok-4.3` ⇒ 启动后仍是 `grok-4.3`（**这条最关键**：写成「不是 4.6 就改」会覆盖运营方选择）
- [x] 3.4.4 设置读写报错 ⇒ 记 Warning，服务仍启动
- [x] 3.4.5 ⚠️ 记录：该迁移**不可逆**，revert 代码不会把已改写的设置回滚

### 3.5 工具链专项（阶段 5）

- [x] 3.5.1 `grep -rn "1\.26\.5" .github/ Dockerfile* deploy/Dockerfile backend/go.mod README*.md DEV_GUIDE.md CLAUDE.md` 零命中
- [x] 3.5.2 5 处版本断言逐一确认已改（不是 3 处）
- [x] 3.5.3 3 个 Dockerfile 的 golang 镜像逐一确认
- [x] 3.5.4 `go generate ./ent` 后 `git diff --stat backend/ent/` **无输出**（生成器幂等，说明提交的就是生成结果）
- [x] 3.5.5 `backend/ent/` 的 diff 里 `group.model_pricing` 与 `usage_cleanup_task.filters` 的类型确实变成 jsonv2 表示
- [x] 3.5.6 `golangci-lint run ./...` 在 v2.13 下通过；新增的抑制注释都是定点的，没有整体移除 linter
      （本机未安装 golangci-lint；CI 已 pin `v2.13`，抑制按上游 `73aabc861` 定点落地。见 `{SCRATCH}/toolchain-env.txt`）
- [x] 3.5.7 keepalive 测试通过，且断言的是新的 HTTP/2 配置结构
- [x] 3.5.8 1C1G 单二进制构建仍成功：`CGO_ENABLED=0 go build -tags embed -ldflags="-s -w" -o /tmp/sub2api ./cmd/server`

## 4. 计费口径专项（阶段 4，⚠️ 费用往上）

这一项会让部分请求变贵，必须能事后解释。

- [x] 4.1 五种开关组合逐一验证：
      分组开×账号关（**改动前不适用 → 改动后适用**，这是唯一行为反转的组合）/
      分组关×账号开（适用）/ 都开（适用）/ 都关（不适用）/ 账号开关缺失（由分组决定）
- [x] 4.2 **留档对比样本**：取一条超过长上下文阈值、分组开关开、账号开关关的真实请求形态，
      记录改动前后的 `CostBreakdown`（输入 2×、输出 1.5×）。留在提交信息或运维笔记里，
      日后有人问「账单为什么涨了」直接指这条
- [x] 4.3 有区间定价时：无论两个开关怎么组合，MUST NOT 叠加长上下文倍率
      （这是把 `&&` 直接换成 `||` 会踩坏的地方）
- [x] 4.4 `git diff backend/internal/service/billing_service.go` 确认 `:1045-1046` 那处未被改动
- [x] 4.5 确认 opt-out 可用：把某分组的 `long_context_pricing_enabled` 关掉后，该分组恢复旧行为

## 5. 上线观察期

- **阶段 2 之后**：默认模型从 4.5 变 4.6 ⇒ 盯 Grok 请求的成功率与用量分布；`grok-3-mini` 类请求的
  费用应明显下降。
- **阶段 4 之后**：盯长上下文请求的费用总额——**预期上升**。若上升幅度远超预期，
  先查是不是有分组不该开着 `long_context_pricing_enabled`。
- **阶段 5 之后**：盯上游连接的 keepalive 行为与容器内存占用（新 Go 版本的运行时差异）。

## 6. 回归红线

任一项不成立即视为验收失败：

- `grokSameAccountRetryMetadata` 在 backend 下有命中（说明合了不该合的 hunk）
- `grok-3-mini` 仍出现在 `defaultModels`，或仍无独立价卡
- 别名收窄条件的最终形态包含 `grok-build-latest`
- 显式请求 `grok-4.5` 被改写成默认模型
- 存量设置里非 `grok-4.5` 的值被迁移改动
- `handler/endpoint.go` 里出现 `IsCNProvider`
- `billing_service.go` 里仍有 `applyLongCtx = applyLongCtx &&`，或 `:1045-1046` 被改动
- 有区间定价时仍叠加了长上下文倍率
- `backend/ent/` 被手改（`go generate ./ent` 后有 diff）
- 任何位置仍残留 `1.26.5`
- 新增了迁移、改动了 `VERSION` / `wire_gen.go`
- 改动了 `frontend/package.json` / `pnpm-lock.yaml` / `pnpm-workspace.yaml`
- 调度诊断改变了任何账号的准入结果
