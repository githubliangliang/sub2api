# 上游0.2.9第一档实施证据

第一档15簇/39PR已完成开发及验收，工作树未提交。第二档仍未实施。
来源评估基线bfbcd79，实际实施起点50c34ad49；15簇/39PR，分支sync/upstream-0.2.9-tier1。

- [正式验收](../../../openspec/changes/port-upstream-0.2.9-tier1/verification.md)
- [后端逐项报告](./backend-report.md)、[前端逐PR报告](./frontend-report.md)
- [39PR覆盖表](./source-coverage.tsv)、[当前代码内容校验](./code-manifest.tsv)
- 该目录与原始evidence-0.2.9分离，不覆盖历史四态/来源基线。

日志来自实际命令；文本中的NUL等控制字符仅转义为可读的\xNN。前端各轮中间失败按历史保留，以最终全量结果和正式验收为完成依据。
后端基线54个有测试的package通过；最终修改后全量/构建/SQLite/race：PASS。
前端全量269文件/1909通过/2既有支付API跳过，typecheck/lint/build通过。

## 可复跑命令

后端在backend/：

```sh
go test -tags=unit ./...
go build ./...
go test -tags=unit ./internal/repository -run 'Test(ProductionSQLUsesSQLiteDialect|UsageBillingRepositoryApplySQLiteAppliesAllEffects|AccountRepositorySQLiteRemainingPaths|APIKeyRateLimitWindowsSQLite)$' -count=1 -v
go test -race -tags=unit ./internal/repository ./internal/handler/admin ./internal/pkg/apicompat ./internal/service -run 'Test(SchedulerMetadataAccountPreservesRPMPolicy|ApplyOAuthCredentialsPreservesExistingNonAuthCredentials|GetOpenAICodexCanonicalUserAgentOutboundIdentity|AccountUsageService_OpenAIQueriesPreserveRefreshError|AnthropicToResponses_Thinking|AnthropicToChatCompletionsRequest_Thinking|ForwardAsAnthropic_DisabledThinking|StructuredOutputsBetaOAuthMimic|BuildUpstreamRequestStructuredOutputsBeta|OpenAIBuildUpstreamRequestOAuthResponsesPreservesCallerBeta|ChatCompletionsToResponses_MessageTypes|SupplementResponseOutput_|AlphaSearchFallback|OpenAIThresholdCandidate_StaleSnapshotFutureReset|ResolveOpenAIQuotaUtilization_StaleSnapshotFutureReset|ShouldStripOpenAIResponsesInputItemID|SanitizeOpenAIResponsesInputItemIDsRemovesOverlongIDs|GrokQuotaQueryRemainsAvailableWhileSchedulingIsPaused|GrokQuotaBillingStillRejectsUnavailableCredentialsDuringCooldown)' -count=1
```

前端在frontend/（PowerShell执行pnpm时设置COREPACK_ENABLE_AUTO_PIN=0，避免Corepack改写package.json）：

```sh
pnpm typecheck
pnpm lint:check
pnpm exec vitest run --maxWorkers=4 --minWorkers=2
pnpm run build
```

Go1.27.1使用官方go.dev archive并校验SHA256；项目版本要求仍为1.27.0。Windows下载验证Go modules后供WSL共享本地缓存，未关闭sumdb或改依赖。
没有创建版本tag、提交/推送产品代码或部署。
