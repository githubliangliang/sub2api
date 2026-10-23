# 2026-10-23 基线验证

基线main `873962ce7f5c07a336479c823419a531529e6f99`，VERSION1.1.18，迁移226。以下命令均在冻结源码上执行；Go在backend模块内自动选用1.27.0。没有把上游测试假称为本地已有测试，没有实现任何提案。

## 已执行

| 检查 | 命令/范围 | 结果与证据 |
|---|---|---|
| Go测试存在性 | backend：`go test -tags=unit ./internal/pkg/apicompat ./internal/pkg/openai ./internal/pkg/antigravity ./internal/service ./internal/repository ./internal/handler ./internal/handler/admin -list .` | exit0，[inventory](./backend-test-inventory.txt) |
| 协议包 | `go test -tags=unit ./internal/pkg/apicompat ./internal/pkg/openai ./internal/pkg/antigravity` | 3包通过，[protocol](./backend-protocol-tests.txt) |
| 重点回归 | service/handler/admin/repository；run表达式见下方 | 4包通过，[targeted](./backend-targeted-tests.txt) |
| 协议相邻回归 | service/handler；`Test(HandleCCStreamingFromAnthropic\|HandleResponsesStreamingResponse\|.*Gemini.*\|.*RateLimit.*403\|ImageResultUploader\|.*Codex.*\|.*Antigravity.*)` | 2包通过，[neighbor](./backend-neighbor-tests.txt) |
| 403/认证 | service；`TestHandleUpstreamError_.*403\|TestCheckErrorPolicy_GeminiAccounts\|TestShouldFailoverGeminiUpstreamError` | 通过，[auth](./backend-auth-tests.txt) |
| 真实SQLite | repository；`TestUsageBillingDeletedKeySQLite\|TestUsageBillingRepositoryApplySQLite\|Test.*SQLite.*Dialect\|Test.*SplitSQL` | 通过，[SQLite](./backend-sqlite-tests.txt) |
| 混合路由 | server/routes：先`-list Composite`再`-run Composite -count=1` | 通过，[inventory](./backend-route-inventory.txt)/[routes](./backend-route-tests.txt) |
| 前端相邻回归 | frontend：`pnpm exec vitest run src/views/admin/__tests__/AccountsView.lite.spec.ts src/views/admin/__tests__/AccountsView.priorityColumn.spec.ts` | 2文件11测试通过，[vitest](./frontend-neighbor-tests.txt) |
| 缺陷探针 | Go overlay增加2个临时测试文件，service/admin，`-tags=unit -run TestIntake0218 -v -count=1` | 9个service子场景+1个admin用例确认旧行为，[probe](./baseline-probes.txt) |
| 前端审计 | `pnpm audit --json`及`pnpm audit --prod --json` | 均exit1，有现存记录；见[依赖复核](./dependency-review.md) |
| Go审计 | `go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 -scan=package -json ./...` | exit0且有finding，package层级非可利用性证明，见[复核](./dependency-review.md) |
| OpenSpec | `pnpm dlx @fission-ai/openspec validate port-upstream-0.2.18-tier1 --strict`，tier2同样 | 两份通过，[tier1](./openspec-tier1.txt)/[tier2](./openspec-tier2.txt) |

重点回归的完整run表达式：

```text
Test(CodexIdentityKeys|CodexAccountIndex|ImportCodex|OpenAIStream|OpenAIResponseFlush|.*Composite|.*Grok.*Pricing|.*Grok.*Cost|.*Gemini.*Usage|.*CleanToolSchema|.*ImageResultUploader|.*QuotaWindows|.*FixedQuota|.*DeletedAPIKey|.*DeletedKey|.*Gateway.*Stream|.*Antigravity.*Model|.*ModelMapping)
```

以上定向执行均使用`-count=1`（协议包常规go test除外）。测试名称可对照inventory验证；来源新增测试见[source-test-inventory.tsv](./source-test-inventory.tsv)，其中absent是未来移植任务，未在本地冒充执行。

## 探针复现方法与解释

将[service-probe.go.txt](./service-probe.go.txt)和[admin-probe.go.txt](./admin-probe.go.txt)通过Go overlay分别映射为`backend/internal/service/intake_0218_probe_test.go`和`backend/internal/handler/admin/intake_0218_probe_test.go`；目标文件并未落入生产树。overlay的Replace字典使用绝对路径；然后在backend运行：

```bash
go test -overlay /tmp/sub2api-intake-0218/overlay.json -tags=unit ./internal/service ./internal/handler/admin -run TestIntake0218 -v -count=1
```

这些断言明确期望旧缺陷行为：queued/空delta被当输出、合成失败缺时间、仅缓存产生负input、Grok Fast落普通价卡、const/examples未清理、OAuth仍支持退役模型、外平台OAuth可构造Anthropic URL、缺message_stop仍成功、loopback JSON能被图片下载器接收，以及跨workspace用户键碰撞。没有向任何真实供应商发送凭据或请求；下载只用本机临时httptest服务。

初版探针误假设“const同名属性会被删”，被当前cleaner的结果否定；[初次输出](./baseline-probes-initial.txt)保留。检查实际代码后改为验证真正缺口“const/examples仍出站”，再运行通过。未修生产代码以通过探针。探针PASS意味着缺陷复现，**不意味着修复验收PASS**。

## 集成测试边界

[integration-build-tags.txt](./integration-build-tags.txt)记录全部当前`*_integration_test.go`首行。`usage_billing_repo_integration_test.go`等是`//go:build integration && postgres`，不在SQLite-only验收中执行。当前真实SQLite测试包括`TestUsageBillingDeletedKeySQLite`、`TestUsageBillingRepositoryApplySQLiteQuotaWindows`、`TestUsageBillingRepositoryApplySQLiteAppliesAllEffects`，名字在inventory中且实际定向执行通过。未将PG测试改标签后强跑。

## 尚未执行（属于实施验收）

未应用上游产品补丁，未跑拟实现版本的全量unit/build/race或前端全量typecheck/lint/build；未连接真实业务实例验证上游权限/风控；未发布、部署或创建tag。16簇OpenSpec的任务与证据保持未完成状态。
