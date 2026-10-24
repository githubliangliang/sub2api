# 验收

原则：验收 `specs/*/spec.md` 的 Requirement 是否成立。三个「静默失效」项（池模式重试、
空 capabilities、ops 内存）**必须先能观察到失效**，再证明修复——它们都不抛错，
按「代码改了」验收等于没验收（`design.md` 决策 2）。

清单里的 2 项前端依赖安全项已决定推迟（`design.md` 决策 1），因此本文**没有依赖审计相关验收**；
对应的是一条反向门禁：那几个文件一个都不能动，见 §6。

## 1. 全量信号（每阶段结束都跑）

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

| Requirement | 证据 |
|---|---|
| token 刷新不采信未过期快照 | 上游 `frontend/src/api/__tests__/tokenRefresh.spec.ts` 改动移入；`pnpm exec vitest run src/api/__tests__/tokenRefresh.spec.ts` |
| 池模式重试（两条） | 见 §3.1；上游新增 `gateway_pool_mode_retry_test.go` 整文件落地 |
| ops 内存自洽（两条） | 见 §3.3；上游新增 `ops_metrics_collector_memory_test.go` 整文件落地 |
| 空能力容器 | 见 §3.2；上游 `openai_images_test.go` 改动移入 |
| 前导 system 前缀 | 上游 `openai_content_session_seed_test.go` 改动移入；断言「中途注入 system 后种子不变」「前导前缀变则种子变」 |
| OAuth 走 Codex manifest | 上游 `upstream_models_test.go` 改动移入；`tasks.md` 6.9 门禁 |
| Google One 保守模型集 | 上游 `pkg/geminicli/models_test.go`、`handler/admin/account_handler_available_models_test.go`、`service/account_wildcard_test.go` 改动移入；另见 §4.2 |
| daily 指向官方域 + 付费账号走 daily | 上游 `pkg/antigravity/oauth_test.go`、`service/antigravity_rate_limit_test.go` 改动移入；`tasks.md` 6.6 门禁；§4.3 |
| 流式按上报 item 重建 | 上游新增 `openai_responses_stream_output_items_test.go` 整文件落地；覆盖「未知 item 类型」「多个同类型 item」「无 done 事件回落」 |
| 空 tool name | 上游新增 `responses_to_chatcompletions_tool_name_test.go` 整文件落地 |
| Ollama 思维字段对齐 | 上游新增 `openai_gateway_ollama_cloud_cc_reasoning_test.go` 整文件落地；断言非 Ollama Cloud 账号请求逐字节不变 |
| Ollama clamp 输出上限 | 上游新增 `openai_gateway_ollama_cloud_max_tokens_test.go` 整文件落地；覆盖默认上限 / 账号覆盖 / 0 禁用 / 字段缺失 |
| Composite 视频端点 | 上游 `server/routes/gateway_test.go` 改动移入 |
| 日志只在变化时输出 | 上游 `securityaudit/prompt_config_test.go` 改动移入；断言「连续重载无变化 ⇒ 0 条」「版本变化 ⇒ 1 条」「失败后恢复 ⇒ 1 条」 |
| IPv6 代理解析 | 上游新增 `views/admin/__tests__/ProxiesView.ipv6.spec.ts` 落地 |
| 并发数 0 = 不限 | 上游新增 `components/admin/user/__tests__/UserEditModal.spec.ts` 落地 |
| 优先级列默认展示 | 上游新增 `views/admin/__tests__/AccountsView.priorityColumn.spec.ts` 落地；额外断言「已保存的列偏好不被覆盖」 |
| 错误详情返回列表 + 诊断分区 | 上游新增 `ops/components/__tests__/OpsErrorDetailModal.spec.ts` 落地；`i18n/__tests__/opsLocaleKeys.spec.ts` 改动移入（防裸 key） |

## 3. 静默失效三件套（阶段 1，必须先复现失效）

### 3.1 池模式同账号重试（`b1e60ba45`）

- [x] 3.1.1 移植前：池模式账号经 Chat Completions 转发触发可重试上游错误，确认
      `UpstreamFailoverError.RetryableOnSameAccount` 为 false（即失效存在）
- [x] 3.1.2 移植后：同场景为 true，且请求确实先在同账号重试
- [x] 3.1.3 Responses 路径重复 4.1.1 / 4.1.2
- [x] 3.1.4 `HandleUpstreamError` 返回「应禁用」时为 false
- [x] 3.1.5 非池模式账号、不可重试状态码：均为 false，与移植前一致

### 3.2 空 `openai_capabilities`（`40c26f343`）

- [x] 3.2.1 移植前：把某 OAuth 账号的能力字段直接写成 `{}`，确认它被文本调度**静默跳过**
      （无错误日志，只是不被选中）
- [x] 3.2.2 移植后：同账号可正常参与文本调度
- [x] 3.2.3 `[]` 同样视为未配置
- [x] 3.2.4 非空但全 false 仍视为「已配置且不含能力」，行为不变
- [x] 3.2.5 正常配置部分能力时行为不变

### 3.3 ops 内存混用（`cd05772e9`）

- [x] 3.3.1 移植前：在 cgroup v2 且未设内存上限的容器里看面板，确认百分比被压到极小
      （容器 used / 宿主 total）
- [x] 3.3.2 移植后：同环境下 used 与 total **同时**来自宿主机，百分比合理
- [x] 3.3.3 设了内存上限的容器：三个读数全部来自 cgroup
- [x] 3.3.4 完全无 cgroup（裸机 / 非 Linux 容器）：三个读数全部来自宿主机
- [x] 3.3.5 CPU 读数不受内存改动影响

## 4. 需要人工确认的三项

### 4.1 `913ec5d74` 的基座补齐

- [x] 4.1.1 `CodexCanonicalClientVersion()` 已补，且函数体为 `resolveCodexOutboundIdentity("").version`
- [x] 4.1.2 `go build ./...` 通过（不补这一步会编译失败，见 `source-baseline.md` §4.2）
- [x] 4.1.3 OAuth 账号同步请求携带的版本号与出站规范身份一致（不是另立来源）

### 4.2 `f98a056f7` 的收窄影响

- [x] 4.2.1 列出现有 Google One OAuth 账号
- [x] 4.2.2 检查分组白名单 / `model_pricing` 是否引用了将被移出清单的模型 ID
- [x] 4.2.3 若有引用，先决定是清理配置还是不合这一条，**再**上线

### 4.3 `e7a3c1202` + `21c07e835` 的顺序

- [x] 4.3.1 两条在**同一个提交**内（或 URL 修正在前）
- [x] 4.3.2 付费账号（`plan_type` = pro / ultra）实际请求打到官方 daily 域，不是 `.sandbox.`
- [x] 4.3.3 免费账号仍走生产端点，未出现 401「Invalid bearer token」

## 5. 上线观察期

`40c26f343` 会让此前被静默排除的 OAuth 账号重新进入文本调度，账号池实际容量上升、流量分布改变；
`e45490a36` 会提高 chat 粘性命中率。上线后一个观察期内盯：

- 账号级并发是否触顶、429 是否集中到少数账号
- 原先承接全部流量的账号负载是否明显下降（说明排除确实解除了）
- prompt cache 命中率变化
- 异常时按阶段 revert：这两条在不同提交里，可分别定位

## 6. 回归红线

任一项不成立即视为验收失败：

- 新增了 `backend/migrations/*.sql`，或 `VERSION` / `wire_gen.go` 被改动
- 新增了全局配置项
- 改动了 `frontend/package.json` / `pnpm-lock.yaml` / `pnpm-workspace.yaml` / `.github/audit-exceptions.yml`（依赖安全两项已推迟）
- `grep -rn "daily-cloudcode-pa.sandbox" backend/` 有命中
- 非 Ollama Cloud 账号的请求体被改动
- 非空但全 false 的 `openai_capabilities` 被放宽成「不限制」
- ops 面板出现容器 used 配宿主 total 的组合
- diff 中出现 0.1.180 §6 / §7 的文件
