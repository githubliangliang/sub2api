# 第一档实施验收

第一档15簇/39PR已完成开发及验收，工作树未提交。第二档仍未实施。

- 实际实施起点：`50c34ad498fa333e20b4ee46dfe58c4512d88a89`，开始时工作区干净。
- 分支：`sync/upstream-0.2.9-tier1`。
- 完成产品SHA：未提交；当前内容固定在[代码校验清单](../../../docs/upstream-sync/implementation-0.2.9-tier1/code-manifest.tsv)。
- 执行日期：2026-09-28。
- Go 1.27.1 / Linux amd64（WSL，GCC可用）；go.mod保持1.27.0，pnpm10.22.0，Vitest2.1.9。
- 冻结source-*和evidence-0.2.9保持不变；[实施证据目录](../../../docs/upstream-sync/implementation-0.2.9-tier1/README.md)单独存放新证据。

## 逐簇验收

| ID | 前后对照与反例 | 证据 |
|---|---|---|
| F01 | TestSchedulerMetadataAccountPreservesRPMPolicy：有效RED断言失败→GREEN通过；正常/边界与源码用例对应 | [RED](../../../docs/upstream-sync/implementation-0.2.9-tier1/backend-red-valid.txt)、[后端报告](../../../docs/upstream-sync/implementation-0.2.9-tier1/backend-report.md) |
| F02 | TestApplyOAuthCredentialsPreservesExistingNonAuthCredentials：有效RED断言失败→GREEN通过；正常/边界与源码用例对应 | [RED](../../../docs/upstream-sync/implementation-0.2.9-tier1/backend-red-valid.txt)、[后端报告](../../../docs/upstream-sync/implementation-0.2.9-tier1/backend-report.md) |
| F03 | TestGetOpenAICodexCanonicalUserAgentOutboundIdentity；TestAccountUsageService_OpenAIQueriesPreserveRefreshError：有效RED断言失败→GREEN通过；正常/边界与源码用例对应 | [RED](../../../docs/upstream-sync/implementation-0.2.9-tier1/backend-red-valid.txt)、[后端报告](../../../docs/upstream-sync/implementation-0.2.9-tier1/backend-report.md) |
| F04 | TestAnthropicToResponses_ThinkingDisabledOverridesOutputEffort；TestAnthropicToChatCompletionsRequest_ThinkingDisabledOverridesOutputEffort；TestForwardAsAnthropic_DisabledThinkingOverridesMaxForResponses：有效RED断言失败→GREEN通过；正常/边界与源码用例对应 | [RED](../../../docs/upstream-sync/implementation-0.2.9-tier1/backend-red-valid.txt)、[后端报告](../../../docs/upstream-sync/implementation-0.2.9-tier1/backend-report.md) |
| F05 | TestStructuredOutputsBetaOAuthMimic；TestBuildUpstreamRequestStructuredOutputsBeta；TestOpenAIBuildUpstreamRequestOAuthResponsesPreservesCallerBeta：有效RED断言失败→GREEN通过；正常/边界与源码用例对应 | [RED](../../../docs/upstream-sync/implementation-0.2.9-tier1/backend-red-valid.txt)、[后端报告](../../../docs/upstream-sync/implementation-0.2.9-tier1/backend-report.md) |
| F06 | TestChatCompletionsToResponses_MessageTypesWithReasoningAndTools；TestSupplementResponseOutput_*：有效RED断言失败→GREEN通过；正常/边界与源码用例对应 | [RED](../../../docs/upstream-sync/implementation-0.2.9-tier1/backend-red-valid.txt)、[后端报告](../../../docs/upstream-sync/implementation-0.2.9-tier1/backend-report.md) |
| F07 | TestAlphaSearchFallbackRequiresSuccessfulCompletion；TestAlphaSearchFallbackCompletionPreservesOutputAndCitations；missing_status反例：有效RED断言失败→GREEN通过；正常/边界与源码用例对应 | [RED](../../../docs/upstream-sync/implementation-0.2.9-tier1/alpha-status-red.txt)、[后端报告](../../../docs/upstream-sync/implementation-0.2.9-tier1/backend-report.md) |
| F08 | TestOpenAIThresholdCandidate_StaleSnapshotFutureReset；TestResolveOpenAIQuotaUtilization_StaleSnapshotFutureReset：有效RED断言失败→GREEN通过；正常/边界与源码用例对应 | [RED](../../../docs/upstream-sync/implementation-0.2.9-tier1/backend-red-valid.txt)、[后端报告](../../../docs/upstream-sync/implementation-0.2.9-tier1/backend-report.md) |
| F09 | 逐PR RED→GREEN；对应真实组件、异步乱序与正常反例通过 | [前端逐PR报告](../../../docs/upstream-sync/implementation-0.2.9-tier1/frontend-report.md)（F09） |
| F10 | 逐PR RED→GREEN；对应真实组件、异步乱序与正常反例通过 | [前端逐PR报告](../../../docs/upstream-sync/implementation-0.2.9-tier1/frontend-report.md)（F10） |
| F11 | 逐PR RED→GREEN；对应真实组件、异步乱序与正常反例通过 | [前端逐PR报告](../../../docs/upstream-sync/implementation-0.2.9-tier1/frontend-report.md)（F11） |
| F12 | 逐PR RED→GREEN；编辑/保存、空值/RPM0及五条科学计数法输入路径补漏后44项通过 | [前端逐PR报告](../../../docs/upstream-sync/implementation-0.2.9-tier1/frontend-report.md)（F12） |
| F13 | 逐PR RED→GREEN；对应真实组件、异步乱序与正常反例通过 | [前端逐PR报告](../../../docs/upstream-sync/implementation-0.2.9-tier1/frontend-report.md)（F13） |
| F14 | 逐PR RED→GREEN；对应真实组件、异步乱序与正常反例通过 | [前端逐PR报告](../../../docs/upstream-sync/implementation-0.2.9-tier1/frontend-report.md)（F14） |
| F15 | TestShouldStripOpenAIResponsesInputItemID_Reasoning；TestSanitizeOpenAIResponsesInputItemIDsRemovesOverlongIDs；TestGrokQuotaQueryRemainsAvailableWhileSchedulingIsPaused；无效认证反例：有效RED断言失败→GREEN通过；正常/边界与源码用例对应 | [RED](../../../docs/upstream-sync/implementation-0.2.9-tier1/backend-red-valid.txt)、[后端报告](../../../docs/upstream-sync/implementation-0.2.9-tier1/backend-report.md) |

## 汇总门禁

| 门禁 | 命令/结果 |
|---|---|
| 基线Go unit/build/list | 独立50c34ad49 archive：完整unit、build与四个相关包-list均PASS |
| 修改后完整后端unit | `go test -p 4 -tags=unit ./...`：PASS |
| 后端build | `go build -p 4 ./...`：PASS |
| SQLite方言/真实库 | ProductionSQLUsesSQLiteDialect、UsageBillingRepositoryApplySQLiteAppliesAllEffects、AccountRepositorySQLiteRemainingPaths、APIKeyRateLimitWindowsSQLite：PASS |
| 后端race | F01–F08/F15目标用例，四个package加`-race`：PASS |
| 前端typecheck/lint | `pnpm typecheck`、`pnpm lint:check`：PASS，最后F12修复后执行 |
| 前端全量 | `pnpm exec vitest run --maxWorkers=4 --minWorkers=2`：269文件，1909通过、2既有跳过 |
| 前端生产构建 | `pnpm run build`：PASS；仅既有Browserslist数据和大chunk提示 |
| 交互/端点行为 | 实际Vue组件输入/焦点/IME/迟到响应测试；Go真实请求构造与ForwardAlphaSearch调用，均PASS。未用线上账号或做生产部署 |
| 独立审查 | 分批+最终全范围Spec/Quality PASS，F05错误ALREADY、F12编辑及输入回写、F07缺状态均已落实 |
| 范围审计 | 无迁移、依赖、VERSION、Ent/Wire、支付/平台/菜单/路由扩展；所有产品路径见code-manifest |

两条跳过均在未修改的SettingsView.spec.ts，测试已移除的支付provider API，与本批范围无关；不计为已运行。
Go `integration && postgres`用例未运行，不冒充SQLite证据。浏览器视觉布局未新增，交互由组件测试验证；未声称有真实供应商端到端流量验证。

## 相对来源补丁的适配与复核修复

1. F05：反向patch匹配了另一份passthrough白名单，普通openaiAllowedHeaders仍缺openai-beta。测试证实后补齐普通白名单和legacy token清理。原始四态是文本证据，保持冻结；其语义解释在当前PORTING中更正。
2. F12：规格要求同一倍率/RPM约束贯穿新增、已有行编辑及最终保存。原PR仅修新增，因此补齐编辑/保存守卫；逐段科学计数法保留字符串草稿至提交，不对输入中间态即时格式化。没有扩大到其它功能。
3. F07：原PR允许completed事件缺status，和已批准规格“且状态成功”不符。本地严格要求status=completed，补两种缺status回归，修正成功fixture；失败不写成功响应、不构造billable结果。
4. F01/F15适配本地测试文件、ID前缀及空ID语义；F05补本地cfg fixture，Grok使用既有snapshot helper。编译/fixture错误不计有效RED。

详细39PR映射：[source-coverage.tsv](../../../docs/upstream-sync/implementation-0.2.9-tier1/source-coverage.tsv)。
