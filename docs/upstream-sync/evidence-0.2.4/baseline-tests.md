# 评估阶段基线证据

日期：2026-09-12。HEAD `c9d9bebe87926791bfdd716f59abcba413fee20a`，VERSION `1.1.11`。
测试在当前工作树执行，包含用户原有的四个 OpenAI metadata 文件改动；本轮未修改产品代码。
四态补丁检查则使用独立 index 的 HEAD，两个基线口径不可混用。

## 1. 后端

在 `backend/` 执行以下同一 selector 的 `-list` 和 `-run`：

```bash
go test -tags=unit ./internal/service ./internal/handler ./internal/repository \
  ./internal/pkg/apicompat ./internal/pkg/antigravity ./internal/pkg/claude \
  -list 'CalculateOpenAI429ResetTime|OpenAI429FastPath|OpenAICompatPreviousResponse|BillingHeader|ClaudeCodeValidator|ChannelRestriction|CodexDelegation|CodexAutomation|ResponsesEventToChat|BufferedResponse|PricingOverride|OpsSystemLog|OpsCleanup|UpdateProxy|SchedulerCache|ProductionSQLUsesSQLiteDialect|ShouldPreserveOpenAIResponsesNone'

go test -tags=unit ./internal/service ./internal/handler ./internal/repository \
  ./internal/pkg/apicompat ./internal/pkg/antigravity ./internal/pkg/claude \
  -run 'CalculateOpenAI429ResetTime|OpenAI429FastPath|OpenAICompatPreviousResponse|BillingHeader|ClaudeCodeValidator|ChannelRestriction|CodexDelegation|CodexAutomation|ResponsesEventToChat|BufferedResponse|PricingOverride|OpsSystemLog|OpsCleanup|UpdateProxy|SchedulerCache|ProductionSQLUsesSQLiteDialect|ShouldPreserveOpenAIResponsesNone' \
  -count=1
```

`-list` 命中 126 个顶层测试。运行结果：service、handler、repository、apicompat、antigravity 全部 PASS，
其中包含 SQLite SQL 方言审计。claude package 在该 selector 下没有测试，未把该空跑计为验证；另行执行：

```bash
go test -tags=unit ./internal/pkg/claude -list .
go test -tags=unit ./internal/pkg/claude -count=1
```

实际列出 `TestDefaultModelsContainsClaudeFable51`、`TestEffortLevelsForModel`，2 项均 PASS。
这些测试不覆盖尚未移植的 beta/CLI override 行为，新增行为仍需实施阶段的回归测试。

## 2. 前端

在 `frontend/` 执行：

```bash
pnpm exec vitest run \
  src/components/layout/__tests__/AppSidebar.spec.ts \
  src/views/admin/__tests__/AccountsView.selectAllResults.spec.ts \
  src/views/admin/__tests__/AccountsView.bulkEdit.spec.ts \
  src/views/admin/__tests__/ProxiesView.ipv6.spec.ts
```

结果：**4 个文件、30 个测试全部通过**。全选失败场景产生预期的 mocked API 错误日志，suite 未失败。
上游新 `AccountsView.lite.spec.ts` 不存在于当前基线，不能用现有测试的 PASS 替代紧凑列表的验收。

## 3. 适用范围

- 本轮为文档和评估，没有 cherry-pick、产品修复、新迁移、依赖升级、构建、全量 unit 或部署。
- 不把当前 PASS 宣称为新修复已验证；126 是顶层测试数，不是所有子场景数。
- `backend/internal/repository/migrations_schema_integration_test.go` 首行已确认是 `//go:build integration && postgres`。
  本次未运行该文件，也不能将它作为 SQLite 迁移验证依据。
- 原始终端日志留在 `/tmp/sub2api-intake-024-E7mtuM/`，本文件记录可持久保留的命令和结果。
- 三份 TSV 由 Git 元数据和独立 index 检查生成；逐候选总数 114，逐文件总数 1,044，非 merge 提交 151。
