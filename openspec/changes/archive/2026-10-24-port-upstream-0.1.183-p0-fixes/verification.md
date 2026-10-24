# 验收

原则：验收 `specs/*/spec.md` 的 Requirement 是否成立，**不以 `git apply` 成功为通过条件**
（`design.md` 决策 1）。上游对应的测试文件能移就一起移，移不动的按下表补本仓库自己的用例。

## 1. 全量信号（每个阶段结束都跑）

```bash
cd backend
go build ./...
go test -tags=unit ./... -count=1
golangci-lint run ./...
cd ../frontend && pnpm run typecheck && pnpm run lint:check
cd .. && make test-frontend-critical
```

已知与本次移植无关的失败：前端全量 `pnpm run test:run` 下
`src/composables/__tests__/useRoutePrefetch.spec.ts` 5 条（见 `PORTING-0.1.179.md` §4.6）。

## 2. 证据矩阵

| capability / Requirement | 证据 |
|---|---|
| 缓存创建明细「存在即覆盖」 | `go test -tags=unit ./internal/service/ -run 'SSEUsage|Passthrough.*Cache|CacheCreation' -count=1`；上游 `gateway_streaming_test.go` 与 `gateway_anthropic_apikey_passthrough_test.go` 的改动一并移入 |
| 明细矛盾按比例封顶 | 上游 `billing_service_test.go` +116 行整段移入；补一条「聚合 1000 / 明细 1000+1000 ⇒ 归一化后和为 1000」的断言 |
| 不支持明细的模型行为不变 | 同上测试文件中的既有用例 MUST 全绿，费用与移植前逐分一致 |
| `session-id` / `session_id` 双写法 | 上游 `openai_gateway_service_test.go` 与 `openai_ws_forwarder_logutil_test.go` 的改动移入；断言 `SessionSource` 分别为 `header_session-id` / `header_session_id` |
| 两种写法同时出现取连字符 | 需自补用例（上游未覆盖） |
| 容量溢出不改绑 | 上游 `openai_gateway_service_test.go` 改动移入；断言「A 队列满 → B 服务本次 → 绑定仍为 A → 下次回 A」 |
| 粘性账号不可用时正常改绑 | 既有调度测试 MUST 全绿：`-run 'Sticky|SessionHash|LoadAwareness'` |
| Gemini 剔除 `deprecated` + enum 归一 | 上游 `gemini_messages_compat_service_test.go` +68 行移入 |
| Antigravity token 封顶 | 上游 `antigravity_gateway_compat_test.go` 改动移入 |
| Sonnet 别名净效果 | 上游 7 个测试文件（`constants_test.go` + 6 个 antigravity/model 测试）改动移入；额外自补「显式 `claude-sonnet-4-5` 仍为 4.5」的断言 |
| 图片 prompt 逐字 | 上游 `openai_images_test.go` 改动移入；断言出站 `instructions` 非空且包含逐字指令 |
| Grok 官方 UA | 上游 `internal/pkg/xai/cli_identity_test.go` 与 `internal/repository/http_upstream_test.go` 改动移入；配合 tasks 6.5 的 grep 门禁 |
| item ID 换前缀保后缀 | 上游新增的 `responses_client_tools_item_id_test.go` 与 `responses_client_tools_item_id_helper_test.go` 整文件移入（本仓库无同名文件，直接新增） |
| 流式 ID 一致性 | 同上；断言 delta/done 事件用 `clientItemID`，内部匹配仍按上游 ID |
| OpenCode Go 重置时长 | 上游 `ratelimit_service_openai_test.go` 改动移入；覆盖 `Resets in 2 days` / `1h 30m` / 不可解析 / 溢出四类 |
| 既有 429 类型行为不变 | `-run 'RateLimitReset|OpenAI429'` MUST 全绿 |
| 充值余额刷新 | `cd frontend && pnpm exec vitest run src/views/user/__tests__/PaymentResultView.spec.ts` |
| 邮箱换绑 alias 查重 | 上游 `auth_service_email_bind_test.go` +多用例移入 |
| 换绑并发守卫 | 见 §4，需专门的并发用例 |

## 3. 计费专项（`bc4a9ae4`）

移植前后各记录一次同一请求的账单，留档以便日后解释账单差异：

- 构造一条上游报数矛盾的流式响应（`message_start` 报 `5m=N`，`message_delta` 报 `5m=0, 1h=M`，
  且 `cache_creation_input_tokens = N`），记录移植前后的 `CacheCreationCost`。
- 预期：移植前 `> N × 单价`（超收），移植后 `≤ N × 单价`（封顶）。
- 预期：明细一致的普通请求，移植前后**逐分相同**。

## 4. 认证路径专项（`4ca86c52`，本批唯一动认证的一项）

- [x] 4.1 登录冒烟：正常登录成功，且 `user_allowed_groups` 加载正常（该表缺失会导致登录 503，
      是本仓库已知的高危面）。证据：`TestAuthServiceLogin_LoadsAllowedGroupsFromJoinTable`
      走真实 `AuthService.Login` → `GetByEmail` → `loadAllowedGroups`，断言 token 非空且
      `AllowedGroups` 等于写入 `user_allowed_groups` 的分组 ID
- [x] 4.2 alias 冲突：用户 B 占 `a.b@gmail.com`，用户 A 申请 `ab@gmail.com` ⇒ 发码与提交两个入口都拒绝
- [x] 4.3 自身 alias：用户 A 从 `a.b@gmail.com` 换成 `ab@gmail.com` ⇒ 放行且换绑成功
- [x] 4.4 并发：两个 goroutine 同时提交互为 alias 的换绑 ⇒ 恰好一个成功，另一个得到「邮箱已存在」，
      库中只有一条记录指向该收件箱
- [x] 4.5 无事务：直接调用 `UpdateEmailWithAliasGuard` 且上下文无事务 ⇒ 返回错误，未写库
- [x] 4.6 唯一约束兜底：模拟写入阶段约束冲突 ⇒ 转换为「邮箱已存在」，错误信息不含 SQL 细节
- [x] 4.7 单机不退化：确认写入侧的取锁—复查—写入顺序完整存在（`design.md` 决策 6）

## 5. 调度观察（`8e60d574` + `e55727d4` 上线后）

两条修改都让会话更粘，账号分布会更集中。上线后一个观察期内盯：

- 账号级并发是否触顶、429 是否集中到少数账号
- prompt cache 命中率是否如预期上升
- 若出现异常集中，按 `design.md` 决策 3 分别 revert 两个提交定位

## 6. 回归红线

以下任一项不成立即视为验收失败：

- 新增了 `backend/migrations/*.sql`
- `backend/cmd/server/VERSION` 被改动
- `wire_gen.go` 被改动
- diff 中出现 routed-catalog 簇文件（`composite_route_resolver.go`、`composite_model_route.go`、
  `gateway_service.go` 的 ownership 接线）
- `grep -rn "sub2api-grok/1.0" backend/` 有命中
- 客户端显式请求 `claude-sonnet-4-5` 被改写成 4.6
- 明细一致的普通请求账单发生变化
