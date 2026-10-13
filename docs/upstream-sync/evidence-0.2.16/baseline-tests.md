# 评估基线测试（非实施验收）

2026-10-13，fork128c5f7bfd4b8ea3f824146ed3b005e9d3e34887。VERSION1.1.16，backend工作目录自动使用Go1.27.0，frontend使用pnpm。未触及用户数据库/配置/账号。

| 检查 | 结果 | 证据 |
|---|---|---|
| Go现有测试枚举：unit下pkg/apicompat、antigravity、claude、xai、setup、repository、service、handler、handler/admin及relay | 通过；确认实际存在 | [inventory](./backend-test-inventory.txt) |
| apicompat/antigravity/claude/xai/setup整包unit | 5包通过 | [log](./backend-core-tests.txt) |
| repository/service/handler/admin定向unit | 4包通过；relay最初regex未命中，另跑整包 | [log](./backend-targeted-tests.txt) |
| openai_ws_v2整包unit | 通过 | [log](./backend-relay-tests.txt) |
| 真实SQLite现有方言/配额窗口/所有结算效果/dedup | 包含在repository定向运行，通过 | 见同上log及inventory的TestProductionSQLUsesSQLiteDialect、TestUsageBillingRepositoryApplySQLiteQuotaWindows、TestUpdateFixedQuotaResetAtWeeklyDefault、TestUsageBillingRepositoryApplySQLiteAppliesAllEffects |
| setup/repository/service overlay探针 | 3项通过（断言旧缺陷/既有保护） | [log](./baseline-probes.txt)，三个`*-probe.go.txt`为完整源码 |
| embed原基线 | **失败**，2个static子用例用/logo.png请求得到SPA；磁盘dist仅/logo.svg | [log](./backend-embed-tests.txt)、[inventory](./backend-embed-inventory.txt) |
| embed测试fixture overlay | 按#7982只改2处请求/2处类型断言，整包通过；真实文件未改 | [log](./backend-embed-fixture-probe.txt) |
| frontend来源中现有测试 | 13文件185通过 | [files](./frontend-test-files.txt)、[log](./frontend-tests.txt) |
| frontend相邻基线回归 | 9文件57通过 | [log](./frontend-neighbor-tests.txt) |
| frontend typecheck | 通过，有pnpm11覆盖配置提示 | [log](./frontend-typecheck.txt) |
| frontend dependency audit | **非零，103条记录**；不能视为通过 | [review](./dependency-review.md) |

核心命令（省略日志重定向，均在backend运行）：

```bash
go test -tags=unit ./internal/pkg/apicompat ./internal/pkg/antigravity ./internal/pkg/claude ./internal/pkg/xai ./internal/setup -count=1
go test -tags=unit ./internal/repository ./internal/service ./internal/service/openai_ws_v2 ./internal/handler ./internal/handler/admin -run 'Test.*(UsageBilling|SQLite|Codex|GPT61|Gemini|Antigravity|Anthropic|CacheControl|Composite|ContentModeration|ChannelMonitorV2Health|OpenAIResponses|ChatCompletions|WebSocket|AccountTest|Grok|ComputeFinal|Claude55|ModelRateLimit)' -count=1
go test -tags=unit ./internal/service/openai_ws_v2 -count=1
go test -tags=embed,unit ./internal/web -list .
go test -tags=embed,unit ./internal/web -count=1
```

首次相同范围加`-list .`枚举；relay额外枚举。frontend运行`pnpm exec vitest run <frontend-test-files.txt列表>`，相邻九文件见log；另`pnpm run typecheck`和`pnpm audit --json`。

探针使用Go overlay映射到临时新增测试文件：setup在临时SQLite创建三字节密码管理员，并对比安装/登录邮箱验证；repository复用真实SQLitefixture，软删除Key后确认Apply错误、余额和dedup回滚；service确认第三方Gemini等级/图片家族键及Codex null归一化。**这些是旧基线证据，不是上游修复已验证，更没有测试在线攻击。**

来源测试197条文件记录（含跨PR重复），首行及存在性保留在source-test-inventory.tsv；PG-only integration不能计入SQLite覆盖。未运行全量Go unit/build/race、frontend全量/lint/build、外部Redis/PG integration。未新增产品测试或改构建标签。
