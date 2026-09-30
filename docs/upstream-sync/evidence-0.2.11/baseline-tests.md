# 本轮评估基线验证

日期：2026-09-30；产品基线 `f13dee245790eb2be058935d0b1aa6265da99474`，VERSION 1.1.14，SQLite最大迁移226。
本轮仅改文档；以下为既有行为的基线结果，不证明新补丁已实现。backend目录自动选择Go 1.27.0，pnpm 11.21.0。

| 检查 | 结果 | 证据 |
|---|---|---|
| `go test -tags=unit -list <pattern>` | 退出0，255个匹配测试名 | [inventory](./backend-test-inventory.txt) |
| service/handler/repository定向unit | 退出0，`-count=1`；完整命令/选择式见日志 | [unit](./backend-unit.txt) |
| apicompat/openai/claude包全量unit | 退出0，`-count=1` | [model packages](./backend-model-packages.txt) |
| `go build ./...` | 退出0 | [build](./backend-build.txt) |
| SQLite | 定向unit包含方言审计、真实SQLite复合平台和用量平台表达式检查 | [unit](./backend-unit.txt) |
| `pnpm run typecheck` | 退出0 | [typecheck](./frontend-typecheck.txt) |
| 4文件Vitest | 97 passed，退出0 | [vitest](./frontend-vitest.txt) |

第一条定向unit表达式在apicompat/openai/claude三个包没有匹配测试，日志明确显示 `[no tests to run]`；因此另跑这三个包的全量unit，不把空匹配当模型验证。

已在 `go test -list` 中确认且由相同表达式执行的代表性用例：

- `TestApplyToolNameRewriteToBody_RenamesToolsAndToolChoice`
- `TestApplyToolNameRewriteToBody_RenamesToolUseInMessages`
- `TestHandleCCStreamingFromAnthropic_PreservesMessageStartCacheUsageAndReasoning`
- `TestHandleResponsesStreamingResponse_PreservesMessageStartCacheUsage`
- `TestAntigravityCompatFirstEventTimeoutTriggersFailover`
- `TestAntigravityCompatClientDisconnectDrainsUsage`
- `TestCompositeRouteResolverExplicitExactRouteRewritesModel`
- `TestCompositeRouteResolverUsesAccountModelOwnershipForUnprefixedAlias`
- `TestBuildCodexModelsManifestForGroupUsesMappedTargetMetadataForCompositeAlias`
- `TestProductionSQLUsesSQLiteDialect`
- `TestChannelMonitorV2ClassifyErrorsRunsOnSQLiteAndResolvesCompositePlatform`

前端使用现有真实组件/工具用例：ModelWhitelistSelector、useModelWhitelist、UseKeyModal、credentialsBuilder。
新增规格要求的多名称偏移、首ping后错误、新模型等场景需在实施时新增或适配，并记录改前失败与改后通过。

[来源测试清单](./source-test-inventory.tsv)记录75条按候选出现的测试文件及本地首行：
`usage_log_repo_integration_test.go`为 `integration && postgres`，不能计SQLite证据；
`api_key_cache_integration_test.go`为 `integration`，本轮unit没有执行它。ABSENT表示本地无文件，不能直接视为已通过。

本轮未跑全仓unit、race、浏览器、生产账号或收费请求；未部署。实施门禁在两份OpenSpec中，全部保持待办。
