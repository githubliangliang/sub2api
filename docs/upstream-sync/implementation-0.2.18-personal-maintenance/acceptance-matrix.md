# 行为与验收映射

| 行为 | 真实执行的测试 | 证据 |
|---|---|---|
| 启用后周期领取既有 pending 清理任务，停止后不处理新任务 | TestUsageCleanupWorkerSQLiteResumesPending | sqlite-background-and-audit / targeted-race-final |
| 清理创建/领取/分批/状态/筛选保留、配额与 dedup 不重置 | TestUsageCleanupRepositorySQLiteLifecycle（DATETIME、旧 TEXT） | sqlite-background-and-audit / sqlite-migrations-and-quota |
| 旧 RFC3339 running 任务续跑，取消不被成功/失败回写覆盖 | TestUsageCleanupRepositorySQLiteStaleAndCanceled、TestUsageCleanupRepositorySQLiteCanceledTerminalState、TestUsageCleanupRepositoryEntCanceledStateRemainsTerminal | targeted-race-final / backend-unit-complete |
| 到期暂停只改变满足条件的账号，精确边界与时区被正确读取 | TestAccountExpirySQLiteFormatsAndEligibility | sqlite-background-and-audit / targeted-race-final |
| 暂停后更新账号快照并保留 outbox、重复执行幂等 | TestAccountExpirySQLiteSnapshotRefresh、TestAutoPauseExpiredAccountsEnqueuesAffectedAccounts、TestAutoPauseExpiredAccountsSkipsOutboxWithoutChanges | targeted-race-final / backend-unit-complete |
| 后台启动、重复启动、等待取消、重叠 tick | TestAccountExpiryWorkerSQLiteStarts、TestAccountExpiryWorkerStartsOnce、TestScheduledTestWorkerSQLiteStarts、TestScheduledTestWorkerStopCancelsDelayedTick、TestScheduledTestWorkerSkipsOverlappingTick | worker-lifecycle / targeted-race-final |
| 定时计划/结果 SQLite CRUD、旧 TEXT、带时区 due、结果保留及级联 | TestScheduledTestRepositorySQLiteLifecycle | sqlite-background-and-audit / targeted-race-final |
| 本地真实 HTTP 账号测试，保存结果/裁剪/推进，失败不恢复 | TestScheduledTestWorkerExecutionAndRecovery | targeted-backend / targeted-race-final |
| 结果存储失败后不恢复账号、不推进计划 | TestScheduledTestWorkerExecutionAndRecovery/result_save_failure | scheduled-execution-baseline → targeted-backend / targeted-race-final |
| 操作日志清理返回行数，保留同步留痕与历史 ID 序列 | TestAuditLogClearAllSQLite、TestAuditLogClearAllSQLiteCanceled | baseline-audit → sqlite-background-and-audit |
| 真实 AES/TOTP/miniredis/handler，未验证、错误、关闭、机器凭证及次数限制 | TestAuditLogClearHandlerSQLiteAndTotp（7子场景） | audit-handler / targeted-race-final |
| 页面确认/TOTP/成功刷新/失败重试/Enter 防重入 | AuditLogView.clear.spec.ts（4项） | baseline-audit-button-corrected → audit-button / frontend-tests |
| 普通 Gemini 原生/Messages、Claude Messages 不额外读取 Antigravity 桶 | TestGeminiGroupRequestsDoNotReadAntigravityBucket、TestMessagesAnthropicGroupDoesNotReadAntigravityBucket | baseline-handler-probes → lightweight-scheduling |
| lazy probe 显式 true/false 优先、取消、metadata 更新、原分组快照及503/限流行为 | single_account_retry_probe_test.go 各场景 | lightweight-scheduling / targeted-race-final |
| 32并发消费者只查询一次，取消后的 probe 仍有5秒边界 | TestSingleAccountRetryProbeConcurrentConsumers、TestSingleAccountRetryProbeCanceledRequestUsesBoundedContext | targeted-backend / targeted-race-final |
| 签名零分配、顺序/交换/非字符串/同尺寸原地变更 | TestModelMappingSignatureZeroAllocations、TestModelMappingSignatureContentChanges、TestModelMappingCacheInvalidatesSameSizeMutation | baseline-signature → lightweight-scheduling / targeted-race-final |
| 签名前后性能 | BenchmarkModelMappingSignature（三种规模，每种3次） | benchmark-before / benchmark-after / benchmark-summary.json |
| 新时间 scanner 的 time.Time/string/[]byte/null/错误边界 | TestSQLiteTimeScannerLegacyAndNative | targeted-race-final / backend-unit-complete |

保留原始失败输出。初版两个夹具问题、旧 mock 的参数期望以及已有 dashboardRepoStub 的计数读写 race 均在证据 README 和最终 verification 中说明；不归因给产品，也不以跳过用例处理。
