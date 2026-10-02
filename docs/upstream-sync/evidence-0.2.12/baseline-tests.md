# v0.2.12评估基线验证

2026-10-02，fork SHA `5eb58199ee6ef45515ec791b445305651384fc0c`，未应用任何上游产品补丁。

| 检查 | 结果 | 证据 |
|---|---|---|
| 相关xai/repository/service/handler unit与SQLite方言 | 4包通过 | backend-tests.txt |
| 同范围go test -list | 成功，确认现有测试存在 | backend-test-inventory.txt |
| 用户确认排序后的现有APIKey repository回归 | 通过 | key-tests.txt、key-test-inventory.txt |
| 现有KeysView | 12通过 | frontend-tests.txt |
| 用户确认优先级后的现有EditAccountModal | 40通过 | account-tests.txt |
| 前端typecheck | 通过 | frontend-typecheck.txt |
| 两项确定性旧邮箱并发缺陷探针 | 均重现旧缺陷 | baseline-race-probe.txt |
| OpenSpec strict | 两批有效 | openspec-tier1.txt、openspec-tier2.txt |

后端在backend目录执行：

```sh
rtk go test -tags=unit ./internal/pkg/xai ./internal/repository ./internal/service ./internal/handler -list '(GrokCLI|CLIVersion|CLIProxy|CLIBilling|Email|VerifyCode|PasswordReset|SanitizeUpstream|ProductionSQLUsesSQLiteDialect)'
rtk go test -tags=unit ./internal/pkg/xai ./internal/repository ./internal/service ./internal/handler -run '(GrokCLI|CLIVersion|CLIProxy|CLIBilling|Email|VerifyCode|PasswordReset|SanitizeUpstream|ProductionSQLUsesSQLiteDialect)' -count=1
rtk go test -tags=unit ./internal/repository -list TestAPIKeyRepository
rtk go test -tags=unit ./internal/repository -run TestAPIKeyRepository -count=1
rtk go test -tags=unit -overlay=../docs/upstream-sync/evidence-0.2.12/probe-overlay.json ./internal/service -run TestIntake012Existing -count=1 -v
```

探针源码为baseline-race-probe.go.txt，overlay映射虚拟的backend/internal/service/intake_012_probe_test.go到该文件。当前probe-overlay.json包含本次工作区绝对路径，换checkout需重新生成映射；不将探针加入产品测试。屏障固定多个读操作同时看到旧值，替身自身加锁避免测试数据竞争；这是可控时序复现，不是外网/真实Redis压测。

前端在frontend目录执行：

```sh
rtk pnpm exec vitest run src/views/user/__tests__/KeysView.spec.ts
rtk pnpm exec vitest run src/components/account/__tests__/EditAccountModal.spec.ts
rtk pnpm run typecheck
```

OpenSpec使用已缓存CLI执行 `validate port-upstream-0.2.12-tier1 --strict`、`validate port-upstream-0.2.12-tier2 --strict`，未安装/更改依赖。typecheck有既有pnpm.overrides提示，未因此改锁文件。

没有执行全量unit/build/race或integration套件；没有新功能的实施验收。email_cache_integration_test.go为integration标签，依赖IntegrationRedisSuite；新来源测试的首行及存在性见source-test-inventory.tsv，缺失测试不计通过。SQLite成功用量去重保留为实施门禁，不冒称本轮已跑。
