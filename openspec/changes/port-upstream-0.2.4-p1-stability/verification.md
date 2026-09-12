# 实施验收证据

S01-S12 已完成并通过最终集成检查。各簇下文保留当时 red/green 记录；最终结果优先见
[implementation.md](../../../docs/upstream-sync/evidence-0.2.4/implementation.md)。所有后端命令在 backend/，前端命令在 frontend/ 执行。

实际起点：`cf6b42752f88ee4eefe51e1c1c59ca1608ac20f5`，隔离快照包含用户 input metadata 清理和已批准 intake。

第一档完成 SHA：`f55626f99`（产品代码 `0eb88e504`）。已先集成第一档，再合入第二档。

完成产品代码 SHA：集成分支 `3f3f192d9`；后续 `d0667b4cf` 仅校正已有 route-prefetch 测试。

执行日期：2026-09-12。

| 簇 | 触发场景 / 改前证据 | 改后与反例证据 | 落地 SHA / 缺口 |
|---|---|---|---|
| S01 | `go test -tags=unit ./internal/service -run 'TestOpenAIWSReplayStateBuildAllocationBounded\|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_InvalidEncryptedContentLineageStripsNextTurn' -count=1` FAIL: next turn retained stale-cipher; 803631672 bytes allocated exceeds 33554432 bound. | Focused ownership/lineage/metadata unit PASS; allocation 2893688 bytes over 128 turns. Race check running at checkpoint; final result below. | `d905eb76a` |
| S02 | `go test -tags=unit ./internal/service ./internal/handler -run 'TestForwardAsChatCompletions_CancelsUpstreamBeforeClosingBody\|TestForwardOpenAIWSV2_ClientCancellationDrainsWithoutSyntheticFailure\|TestOpenAIEnsureForwardErrorResponse_SkipsCanceledClient' -count=1` FAIL: Body.Close hangs until test unblocks; WS reports context canceled without terminal usage; canceled handler synthesizes response. | Same command PASS: HTTP returns 17 input tokens; WS drains 3 input/5 output tokens and marks client disconnect; no synthetic failure to canceled client. | `b396cc8db + 431704a14` |
| S03 | `go test -tags=unit ./internal/service -run '^TestRuntimeCooldown' -count=1` FAIL on lagging snapshot and failed DB write. CAS mutation removing generation comparison makes `TestRuntimeCooldownCASPreservesConcurrentSameDeadlineGeneration` fail for all concurrent workers. Mutation restored. | `go test -tags=unit ./internal/service -run 'TestRuntime(Block\|Cooldown)\|TestHandleOpenAITransientError_HardDisableStillBlocksWholeAccount' -count=1` PASS. Three active persisted fields retain blocks; model-only cooldown survives account reconciliation; cached hydration lags but DB recheck rejects persisted cooldown. | `467e79bfc` |
| S04 | `go test -tags=unit ./internal/handler -run '^TestGatewayFailedAnthropicRequestReleasesSessionSlot$' -count=1` FAIL: next session rejected after Messages failure. Cancellation-during-hydration variant also fails on occupied miniredis slot before scoped adaptation. | Handler verifies failure/cancel release and success/count_tokens existing-session retention. Redis test verifies idempotent concurrent ZREM, other-session retention, immediate reuse. Focused unit and race commands below. | `37586e485` |
| S05 | Adapted upstream tests fail behaviorally: discovery returns 2 instead of 4 tools, conflict accepted, done arguments empty, heartbeat/history changed=false, OpenCode forwards fixed-account-value. Initial compile-only FunctionToolNames mismatch fixed in test adaptation before counting red. | `go test -tags=unit ./internal/pkg/apicompat ./internal/handler ./internal/service -run 'Test(Responses\|EffectiveResponsesTools\|BufferedResponseAccumulator\|NormalizeCodex\|OpenCodeSession\|ApplyOpenCode)' -count=1` PASS. Covers duplicate JSON/XML rejection, namespace/custom tools, done suffix/no duplicate, official/other origin, API key/OAuth. | `9d23a5fd9 + eabfa89c3` |
| S06 | Source-adapted backend red: Astra aliases missing, max downgraded to xhigh, cache/session keys empty, continuation retries fail, fallback pricing missing, partial metadata not persisted, official stale image modality remains text-only. Frontend red: 4 failures in model selector/OpenCode config/partial warning; additional create-preview test fails to sync ID 42. | Backend focused suite PASS; frontend 4 files/52 tests PASS; typecheck and lint:check PASS. Tool-capability mutation skips copying metadata and fails Ultra effort (high→xhigh) and alias search capability (true→false); mutation restored. Race running at checkpoint. | `8946f18e6` |
| S07 | 实际 support 起点 `64d0d876e`。定时器按 1/3/默认10分钟推进后价格仍为4；none backend 验证/映射失败，前端丢配置且下拉无 none；并发场景 baseline overlay 仍为7而非10。 | Pricing/none unit PASS；真实定时器 + 8 并发读者×20更新 race PASS；真实 Select 选择及配置 round-trip 21 vitest PASS。 | `942a34cd6` |
| S08 | SQLite red: creating A→B writes B→A; omitted proxy updates clear expiry; invalid announcement dates accepted; malformed proxy-list responses resolve. Repeated sweep first exposed PG JSON syntax, then after dialect-only correction returned 0 changed accounts on second expiry instead of 1. Legacy upgrade red: a name-only update of old Ent A→B/B→A(mode=none) failed with PROXY_BACKUP_CYCLE. | Focused repository/service/admin handler tests PASS; legacy upgrade PASS; 26 frontend API/real-view tests PASS; typecheck/lint PASS. Race pending at checkpoint. | `1b707d594` |
| S09 | support red：旧配置 access 默认入库；system retention 错取 error90天；默认 snapshot 关闭已启用cleanup；真实SQLite cron保留应删除的10天system行；无效retention0/-1未拒绝；UI缺opt-in。 | 单独system窗口7天删除老system/audit，error90天保留；runtime opt-in/reset/rollback、旧配置/字段默认、sink backoff、refresh与race PASS；UI保存/重置4 vitest PASS。 | `1915ee622` |
| S10 | 后端 red：lite DTO 仍含 groups；Antigravity token JSON 丢 plan_type。前端 red：lite 参数丢失、分组不显示、未获取详情即打开编辑、菜单越界及相关界面行为失败。 | 后端 `TestAccountHandlerListLite` / `TestAntigravityTokenInfoPreservesPlanTypeJSON` PASS；11 文件126项定向 vitest PASS；补充真实编辑/批量/注册路径42项 PASS。最后修正测试缺失的 settings API mock，并补完整 proxy_id=null fixture。 | `37e3eccc0` |
| S11 | `go test -tags=unit ./internal/service -run 'Test(GetOrCreateFingerprint.*Floor\|GetOrCreateFingerprintFloors\|CreateFingerprintFromHeadersFloors)' -count=1` FAIL: cached 2.1.220, upgraded 2.1.223, new 2.1.75 instead of 2.1.258. Override mutation (`resolvedCLIVersion=resolveCLIVersion("")`) makes `TestFingerprintVersionOverrideIdentityConsistency` fail on all three new/healthy/heal paths (2.1.258 instead of 2.2.0). Mutation restored. | CLI validation + identity + billing focused suite PASS; exact command below. | `d06848c0c` |
| S12 | 新兼容测试在9.17.2上PASS；未声称复现依赖内部nil-context竞态。故意去掉batch上限/放开future score的overlay mutation分别FAIL（剩余1而非4，清理2而非1）。 | 9.22.0 + 两处ZRangeArgs后 miniredis边界/PTTL/批量/池取消恢复PASS；全repository/redissession unit PASS，相关race PASS；go mod verify/build PASS。 | `6620a7c08` |

| 门禁 | 命令与结果 |
|---|---|
| 后端 build / unit | 最终集成 build PASS；全量 unit 54个包 PASS |
| SQLite 方言 / miniredis | 方言审计、真实SQLite升级用例与miniredis兼容测试 PASS |
| WS 所有权 / CAS / session race | 最终集成定向 race PASS；原始逐簇结果保留在下文 |
| 前端 typecheck / lint / vitest | vue-tsc生产构建、lint PASS；240文件/1720测试PASS，2原有skipped |
| 依赖与 lockfile | go-redis 9.22.0及必要间接依赖；go mod verify PASS；frontend lockfile未改 |
| 用户 metadata 修复保留 | 原用户文件保持，HTTP/WS metadata回归及集成检查 PASS |
| 来源无关 hunks 排除 | 无PG/新平台/Fast发送/固定账号清单/分组allowlist；详情见各簇适配记录 |
| 无迁移 / 无新增平台 / VERSION 保持 | 无Ent/schema/Wire/迁移变更；VERSION=1.1.11，最大迁移226 |

- [x] 所有交付行为有验收证据；缺口逐条列明。
- [x] 来源冻结未改，实际起点和完成 SHA 可追溯。

## Cluster Commands And Adaptations

### S08

- Reviewed all five whole PRs #6816/#6810/#6811/#6836/#6815. Excluded all Ent schema/generated changes and unrelated go.sum codegen dependencies from #6815. No migration.
- Behavioral red: `go test -tags=unit ./internal/repository ./internal/service -run 'TestProxySQLite|TestAdminProxyPartialUpdate|TestAnnouncementService.*Date' -count=1` FAIL. After fixing only SQLite dialect, `go test -tags=unit ./internal/repository -run 'TestProxySQLiteRepeatedFallback|TestProxySQLiteRejectsBackupCycles' -count=1` FAIL on second expiry returning 0, proving the repeated-fallback behavior independently of SQL syntax.
- Green: `go test -tags=unit ./internal/repository ./internal/service ./internal/handler/admin -run 'TestProxySQLite|TestAdminProxy|TestProxyHandlerUpdatePreservesFieldPresence|TestAnnouncementService.*Date|TestBothProxyUpdate|TestProxyUpdate' -count=1` PASS. SQLite fixtures use actual migrations plus auxiliary tables; assertions cover first-origin retention, direct/backup repeated fallback, outbox account IDs, probe invalidation, revert, directed shared references, self/cycle rejection, and transaction rollback.
- Upgrade red/green: `go test ./internal/repository -run '^TestProxySQLiteLegacySymmetricBacklinkRemainsEditable$' -count=1` FAIL then PASS (0.687s). Fixture uses unchanged legacy Ent builders, reproducing the actual old API sequence. Preserves A expiry/backup on rename, preserves incoming A while editing B, allows C→B, and rejects activating B→A or a missing target. Inactive stored reverse IDs do not participate in forwarding and are preserved.
- Frontend: `pnpm exec vitest run src/views/admin/__tests__/ProxiesView.list.spec.ts src/api/__tests__/admin.proxies.spec.ts` PASS (26 tests). API validation red had 18 failures; added real-view backup-selector red reproduced an unhandled rejection, fixed with local error handling. Existing table rows survive failed refresh. `pnpm run typecheck` and `pnpm run lint:check` PASS.
- Race: `go test -race -tags=unit ./internal/repository ./internal/service ./internal/handler/admin -run 'TestProxySQLite|TestAdminProxy|TestProxyHandlerUpdatePreservesFieldPresence|TestAnnouncementService.*Date|TestBothProxyUpdate|TestProxyUpdate' -count=1` PASS (repository 125.254s, service 2.182s, admin handler 1.155s).
- S06 committed as `8946f18e6e17019c42e8bb8e5cccf11b270eb7ac`.

### S06

- Backend red/green command: `go test -tags=unit ./internal/pkg/openai ./internal/service ./internal/handler -run 'Test.*(Astra|GPT6|Gpt53|MaxCapable|SyncUpstreamModelCatalog|MatchModelsDevProvider)|TestCodexBaseInstructionsForModel' -count=1` (red/green output captured in worktree execution logs).
- Mutation red: `go test -tags=unit ./internal/service -run '^TestAstra(UltraCatalogPreservesWorkflowMetadata|CodexToolCapabilitiesFollowAPIKeyAlias)$' -count=1` FAIL after disabling capability copy; restored before green/race.
- Frontend: `pnpm exec vitest run src/components/account/__tests__/CreateAccountModal.spec.ts src/components/account/__tests__/ModelWhitelistSelector.spec.ts src/composables/__tests__/useModelWhitelist.spec.ts src/components/keys/__tests__/UseKeyModal.spec.ts` PASS (52 tests). `pnpm run typecheck` PASS; `pnpm run lint:check` PASS. Frozen dependency installation was completed earlier; no package/lock change.
- Race: `go test -race -tags=unit ./internal/service -run 'TestAstra|TestSyncUpstreamModelCatalog|TestForwardAsAnthropic_Astra' -count=1` PASS (1.908s).
- All seven source PRs reviewed (#6620/#6628/#6572/#6718/#6678/#6743/#6690). Preserved relevant regression tests and existing Grok tests that #6628 removed incidentally. Excluded unrelated payment/redeem tests, requested-effort persistence assertion (absent schema field), missing non-Astra mode-strip behavior, and source-only pinned/routed-catalog signature extensions.
- Tested OAuth/API key distinctions, capability alias/mixed-peer intersections, explicit null/false, Ultra workflow metadata, partial refresh and unavailable-list snapshot retention, official versus compatible image metadata, instructions after mapped model resolution, stable cache identity and unsupported-continuation replay, pricing at 272k boundary and existing priority/flex billing.
- Forward modification is limited to moving default instruction selection after mapped upstream model resolution. Coordinator separately owns integration of review fixes `eabfa89c3`, `431704a14`, and tier-1 F07 fingerprint suffix recalculation.
- S03 committed as `467e79bfcfc5036b308dc5e26276630cd047f33b`.

### S03

- Whole #6320 reviewed. Added persisted-field reconciliation, peek, and generation/deadline CAS. Excluded references to absent `openaiOAuth429RetryStartedAt` (an unrelated retry-state feature), retaining all existing fork runtime state.
- Verified ordering with real `listSchedulableAccounts`, `hydrateSelectedAccount`, and `recheckSelectedOpenAIAccountFromDBBeforeProfit` against controlled snapshot/repository boundaries: stale candidate clears local block; cache hydration can remain stale; final DB recheck rejects a successfully persisted cooldown. On write failure the approved fail-open behavior permits scheduling again.
- No silent policy change: snapshot lag/write failure lose the instantaneous account-level protection; model-level blocks remain. This is the approved upstream tradeoff and is explicitly exercised, not described as fail-closed.
- CAS mutation command: `go test -tags=unit ./internal/service -run '^TestRuntimeCooldownCASPreservesConcurrentSameDeadlineGeneration$' -count=1` FAIL when generation comparison removed, preserving same deadline.
- Race: `go test -race -tags=unit ./internal/service -run 'TestRuntime(Block|Cooldown)|TestHandleOpenAITransientError_HardDisableStillBlocksWholeAccount' -count=1` PASS (1.569s).

S04 committed as `37586e4853567e948fef9ca62e1dc66cc6add5b2`.

### S04

- Unit: `go test -tags=unit ./internal/handler ./internal/service ./internal/repository -run 'TestGatewayFailedAnthropicRequestReleasesSessionSlot|TestReleaseAccountSession|TestSessionLimitUnregister|TestGatewayProfitControlSelectionCarriesGate' -count=1`.
- Race: `go test -race -tags=unit ./internal/handler ./internal/repository -run 'TestGatewayFailedAnthropicRequestReleasesSessionSlot|TestSessionLimitUnregister' -count=1` PASS (handler 1.148s, repository 1.126s).
- Whole #6510 reviewed. Ported Redis interface/repository/service/handler and all concrete cache stubs. Adapted external-Redis suite to real miniredis unit/race tests; kept useful service applicability/error tests.
- Fork adaptations are in design.md: CountTokens does not register and must retain a previously served session; track slots before waiting; release a registered session on hydration failure. New helper parameter updated at all production/test callers.
- Canceled-registration red: `go test -tags=unit ./internal/handler -run '^TestGatewayFailedAnthropicRequestReleasesSessionSlot/messages_canceled$' -count=1` FAIL: next session still rejected; same scenario passes after selection-result cleanup.

S02 committed as `b396cc8dbcf54c38ee3f739ebfae74ac3e3bbb66`; cancellation race command PASS (1.748s).

### S02

- Whole #6754 and #6434 diffs reviewed and ported. HTTP streaming cancellation is invoked before Body.Close. WS uses a detached context with the existing read timeout as total drain budget after client cancellation.
- Reused fork `recordUsage()` closure in handler; excluded source-only service-tier observer and requested-effort fields/helpers absent in fork. Existing usage billing and model metadata retained.
- Race: `go test -race -tags=unit ./internal/service -run 'TestForwardOpenAIWSV2_ClientCancellationDrainsWithoutSyntheticFailure|TestForwardAsChatCompletions_CancelsUpstreamBeforeClosingBody' -count=1` running at checkpoint.

### S01

- Green: `go test -tags=unit ./internal/service -run 'Test(OpenAIWSReplay|OpenAIWSStateStore|CollectOpenAIEncrypted|StripOpenAIInvalid|CombineOpenAIWSReplay|OpenAIGatewayService_ProxyResponsesWebSocketFromClient_InvalidEncrypted|StripOpenAIResponsesInputContent)' -count=1`.
- Allocation/metadata/expiry: `go test -tags=unit ./internal/service -run 'Test(OpenAIWSReplayStateBuildAllocationBounded|OpenAIWSReplaySequenceSharesBodies|OpenAIWSLineageCleanup|StripOpenAIResponsesInputContent)' -count=1 -v` PASS.
- Race: `go test -race -tags=unit ./internal/service -run 'Test(OpenAIWSReplayConcurrent|OpenAIWSLineageCleanup|OpenAIWSStateStoreInvalidEncrypted|OpenAIGatewayService_ProxyResponsesWebSocketFromClient_InvalidEncrypted)' -count=1` PASS (1.138s).
- S01 committed as `d905eb76adb3f858c588aa4f9f6197e8e0c73948`.
- Whole #6397 reviewed. Retained fork state store without absent HTTP response-owner fields, and existing rate-limit call signature. Adapted HTTP bridge lineage marking to its current error handling; tier 1 owns rejected-field retry integration.
- Ownership audit: ingress uses replacement byte slices; collector creates independent RawMessage bodies and now clones headers only; replay merge creates header arrays; sanitation decodes/re-encodes or sjson allocates a new body. Concurrent test shares the same payload across readers and runs lineage plus user metadata clearing without changing the original bytes.
- No input metadata source file changed. No new unrelated state branch introduced.

S05 committed as `9d23a5fd9bea01881c593eb02548216af7337b5b`.

### S05

- Whole PRs #6593/#6629/#6553/#6539/#6581 reviewed. All five behaviors ported.
- Kept existing fork tool bridge signature; excluded `FunctionToolNames` test argument tied to old deferred tool validation. Kept fork ServiceTier field placement and existing proxy resolution.
- Preserved user's input metadata sanitizer and all three existing upstream builders. No older deferred namespace/custom-tool policy promoted into scope.

S11 committed as `d06848c0cce1a654e48a4f00359617283aaa1893`.

### S11

- Green: `go test -tags=unit ./internal/pkg/claude ./internal/service -run 'Test(IsSupportedCLIVersion|ResolveCLIVersion|DefaultHeadersUserAgent|CLIVersion|GetOrCreateFingerprint|CreateFingerprintFromHeaders|SyncBillingHeaderVersion|ExpandClaudeOAuth|FingerprintVersionOverrideIdentityConsistency)' -count=1`.
- Whole PRs #6606 and #6481 reviewed. Preserved validation regression tests; omitted source-only constant gate tests in favor of outbound identity behavior.
- Adaptation: fingerprint version floor uses resolved `claude.CLIVersion()` instead of the upstream built-in constant, so explicit override also updates cached fingerprints. New/healthy/healed paths preserve ClientID, OS and UA suffix; newer valid client versions still win.
- Tier 1 billing sanitization remains separately owned; this cluster only changes template version reads and fingerprint resolution.

### S07 (support worktree)

- 实际起点 `64d0d876e`，仅在 `sync/upstream-024-support-RLC1zS` 实施 S07/S09/S12。完整 PR #6535/#6626 diff 已审阅。
- Red/green backend：`GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/service -run 'TestPricingCustomFilesScheduler|TestNormalizeReasoningEffortMappings|TestApplyOpenAIReasoningEffortPolicy' -count=1`，先 exit 1 后通过。Green 扩为 `-run 'TestPricing|TestNormalizeReasoningEffortMappings|TestApplyOpenAIReasoningEffortPolicy'`，exit 0。
- Concurrent red：仅将 pricing_service.go overlay 为 `git show 64d0d876e:backend/internal/service/pricing_service.go`，运行 `-run '^TestPricingCustomFilesConcurrentReadersSeeWholeSnapshots$'`，exit 1（expected10/actual7）。Race green：`GOMAXPROCS=4 go test -p 2 -race -tags=unit ./internal/service -run '^TestPricingCustomFiles' -count=1`，exit 0。
- 定时器通过 Go `testing/synctest` 推进真实配置间隔；无新增产品 test-only 方法。覆盖未到周期不更新、无 remote URL 仍重载、invalid JSON/array/null 不发布部分层、删除后撤销该层、重建后恢复，保留目录同步 hash/时间和已发布价格对象。
- Frontend red/green：`pnpm exec vitest run src/components/admin/group/__tests__/ReasoningEffortPolicyFields.spec.ts src/views/admin/__tests__/groupsReasoningEffort.spec.ts`，先2失败，后21通过。使用真实 Select 点击、选值、发出 mappings；等待过渡菜单消失后验证 target/ceiling 不出现 none，不把过渡 DOM 算产品故障。
- 排除 absent `warnDroppedLongContextLadders`，未加 no-op，不引入动态 context ladder。保留 fork override 的字段浅合并、null删除字段和原定价查找；仅 JSON null 整文件被判无效，避免误清有效价格。None 仅作为 mapping source；不新增 deny target 策略，保留原 ceiling 行为。

### S09 (support worktree)

- 完整 #6424 diff 已审阅。保持 SQLite executor、现有日志写失败退避、worker生命周期和 miniredis/simple-mode。使用已有 `config.OpsCleanupConfig` 新字段保存 effective retention，没有引入上游冗余的嵌入 wrapper。七行 `defaultOpsAdvancedSettingsForConfig` 随行为带入。
- Red（exit1）：`GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/service -run '^TestOps(AccessPersistenceRuntimeLifecycle|RuntimeRetentionDoesNotInheritErrorLogWindow|AdvancedDefaultDoesNotDisableConfiguredCleanup|ScheduledCleanupUsesIndependentSystemRetentionOnSQLite)$' -count=1`。分别观察到默认access入库、90而非30、cleanup被关闭、SQLite cron旧行未删。
- 配置 red（exit1）：`GOMAXPROCS=4 go test -p 2 ./internal/config -run '^TestLoadRejectsInvalidSystemLogRetention$' -count=1`，0/-1返回nil error。
- Green：`GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/config ./internal/service -run 'Test.*(Ops|RuntimeLogConfig|RuntimeRetention|SystemLogRetention|ComputeEffective)' -count=1`，exit0。旧配置缺字段使用30天；部署显式disabled仍关闭；partial/invalid settings各自fallback；update/reset同步sink与cleanup，写库失败回滚。
- Race：`GOMAXPROCS=4 go test -p 2 -race -tags=unit ./internal/service -run 'TestOpsAccessPersistence|TestOpsRuntimeSettings|TestOpsScheduledCleanup|TestOpsSystemLogSink|Test(Update|Reset|Apply).*RuntimeLogConfig' -count=1`，exit0；含原有抑制重试/恢复/成功清除backoff用例。
- UI red/green：`pnpm exec vitest run src/views/admin/ops/components/__tests__/OpsSystemLogTable.spec.ts`，先缺checkbox失败，后4测试通过。验证初始off、提交true到API、使用返回值、reset恢复off。
- 仅更新通用 deploy/config.example.yaml 说明与默认30天；没有修改 deploy/config.personal.sqlite.yaml，个人7天策略由 coordinator 维护。原 PostgreSQL sink 注释改成 database；无生产SQL/迁移变化。

### S12 (support worktree)

- 完整 #6814 diff 以及所选 go-redis 模块 pool 修复已审阅。使用 `go get github.com/redis/go-redis/v9@v9.22.0`，仅升级 go-redis、其要求的 cpuid/v2 2.2.10 和 atomic 1.11.0；`go mod why -m github.com/dgryski/go-rendezvous` 确认不再需要后移除。没有照搬上游 go.mod；既有 SQLite/miniredis、Go 1.27.0、VERSION 均保留，无新增PG依赖。
- Baseline/green：`GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/repository -run '^TestRedisUpgrade' -count=1`，旧/新版本都PASS，属于兼容性证据。真实 miniredis 验证两个活跃索引每次最多1000个、now边界包含/future排除、队列maxCount/default、PTTL续租/无TTL删除；通过 `InitRedis(redis.enabled=false)` 持有唯一连接，超时等待者退出后16工作者×8次锁操作与最终PING成功。
- Red mutation：在临时Go overlay中将旧API的 active Count改0、queue Max改+inf，运行 `GOMAXPROCS=4 go test -p 2 -tags=unit -overlay=../.support-evidence/s12-range-mutation.json ./internal/repository -run '^TestRedisUpgrade(ActiveIndex|QueueIndex)' -count=1`，exit1，两个行为断言均失败；工作树未写入mutation。两处正式实现保留ByScore/上界/Count语义。
- 不把旧版客户端PASS描述成nil-context问题已复现；该竞态的修复来自冻结目标依赖，本地证据为取消/恢复及race兼容验证。无新增产品test-only方法。
- 完整repository与redissession：`GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/repository ./internal/pkg/redissession -count=1`，exit0（25.167s/0.010s）。首个带统一过滤器的试跑在redissession未匹配测试，因此改为完整包测试，不把no-tests-to-run计作验证。
- Race：`GOMAXPROCS=4 go test -p 2 -race -tags=unit ./internal/repository -run 'TestRedisUpgrade|TestSessionLimitUnregister|TestLiveLease' -count=1`，exit0（1.949s）。
- 依赖：`go mod verify` PASS；`go list -m github.com/redis/go-redis/v9 github.com/klauspost/cpuid/v2 go.uber.org/atomic modernc.org/sqlite github.com/alicebob/miniredis/v2` 分别为9.22.0/2.2.10/1.11.0/1.44.3/2.38.0。

### S07 / S09 / S12 support final evidence

本次 support 起点：`64d0d876eb4bf4572ff756ab4c903eb9239830e6`。产品提交：

- S07：`942a34cd620f04c70d0af5640b12f7e13ad0be20`
- S09：`1915ee6223e3b10b25bcc7061ca04371cfbb5fd1`
- S12：`6620a7c0898efdf7932ca6bfe530810b052510c7`

2026-09-12，support 三簇最终代码对应 `6620a7c08`，以下门禁 exit0：

```sh
# backend/
GOMAXPROCS=4 go build -p 2 ./...
GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/config -count=1
GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/service -run 'TestPricing|TestNormalizeReasoningEffortMappings|TestApplyOpenAIReasoningEffortPolicy|Test.*(Ops|RuntimeLogConfig|RuntimeRetention|ComputeEffective)' -count=1
GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/repository ./internal/pkg/redissession -count=1
GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/repository -run '^TestProductionSQLUsesSQLiteDialect$' -count=1 -v
go mod verify
# frontend/
pnpm install --frozen-lockfile
pnpm run typecheck
pnpm run lint:check
```

各簇 race/vitest 命令与结果见对应段落。安装与检查没有改变 pnpm-lock.yaml；既有 pnpm.overrides 和 Browserslist 陈旧提示未扩范围处理。
实际 red/green 输出汇总在 [support-verification.txt](./support-verification.txt)，仅规范化显示空白并保留原始输出SHA256，包括 S12 基线PASS与故意mutation失败的区别。
全仓组合 unit 与 S06/S08/S10 合并后的联合门禁仍由 coordinator 执行，本记录只宣称本次实际完成的范围。

独立 index 以 `64d0d876e` 做反向优先的逐文件 source diff 检查：31项记录 A/C/X/N=0/28/3/0；S07=0/7/2/0，S09=0/18/0/0，S12=0/3/1/0。
冻结文件 SHA256 与起点相同：source-baseline `8c38c1c837b1b6b3af10ffac62d8e8a183de65e919e0ed30e39003a3810ac753`；source-feature-map `176bfa0067b162bbeb1be4440231f33d2839659abfa544be65f7316c1f71a84f`。
`git diff --check 64d0d876e` PASS。无迁移/schema/Ent/Wire/VERSION/个人配置/CLAUDE/PORTING/README 修改，无新增平台/PG依赖。
S01–S06/S08/S10/S11 的任务和验证行保持继承状态；原第一档分支仍为 `f55626f997661a2c5078d2dcf4f58d701613e2a2`。

集成注意：S06 的 Astra pricing hunks 应与 S07 的 scheduler/build/load helpers 一起保留；none source helper/前端 source options 继续复用基础 effort 列表，使 S06 新能力可自然并入。
S09 没有复制/改写 SQLite SQL，也没有关闭清理worker；coordinator 合个人配置时保留 system_log_retention_days=7。
S12 不需要上游其它依赖或新 Redis 服务；嵌入式 miniredis 继续通过相同客户端连接池运行。未 push 或发布。
