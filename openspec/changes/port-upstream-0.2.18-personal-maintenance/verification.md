# Verification

状态：**开发验收完成（2026-10-28），三项16/16任务完成**。实现已提交至本地main为`b554bea554202e3008b99e3faa7a29420467f318`；所有数据操作均在临时SQLite、本地HTTP和miniredis夹具中执行。

| 项目 | 结果 | 证据 |
|---|---|---|
| 来源与基线 | A0/C9/X2/N0；5个来源测试均非PG-only；当前测试目录已验证 | [来源](./source-baseline.md)、[目录](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/test-inventory.txt) |
| F01 清理周期/续跑/账本 | 周期处理pending/stale任务，旧TEXT与时间比较兼容，取消终态保持，筛选/配额/dedup保留 | [SQLite回归](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/sqlite-background-and-audit.txt)、[行为映射](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/acceptance-matrix.md) |
| F01 到期暂停/快照 | UTC/offset/边界正确；只暂停符合条件账号，立即更新快照与outbox；重复执行幂等 | [SQLite](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/sqlite-background-and-audit.txt) |
| F01 定时测试/结果/生命周期 | SQLite CRUD/旧TEXT/due/裁剪；本地真实HTTP成功/失败；启动/停止/重叠；保存失败不推进/恢复 | [定向](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/targeted-backend.txt)、[原失败](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/scheduled-execution-baseline.txt) |
| F02 SQLite 清理/留痕 | 修复已复现SQL错误，返回数量，保留同步留痕、自增ID和取消错误 | [原失败](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/baseline-audit.txt)、[通过](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/sqlite-background-and-audit.txt) |
| F02 TOTP/页面按钮 | 真实AES/TOTP/miniredis handler7场景；页面确认/验证/刷新/失败重试/Enter防重入4项 | [handler](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/audit-handler.txt)、[页面](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/audit-button.txt) |
| F03 延迟探测 | 三处handler；普通请求额外探测1→0，需要时请求内1次；显式标记/原分组/取消/32并发/503通过 | [原失败](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/baseline-handler-probes.txt)、[通过](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/lightweight-scheduling.txt) |
| F03 零分配/前后 benchmark | 4/32/128映射0 B/0 alloc，中位耗时下降56.17%/65.53%/80.35%；原地同尺寸变更仍失效 | [测量说明](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/README.md)、[结果](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/benchmark-summary.json) |
| 定向回归与 race | service/repository/handler/admin四包通过，含scanner和历史恢复逻辑 | [race](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/targeted-race-final.txt) |
| 后端 unit/build/lint | 全量54个测试包通过，server build通过，golangci-lint2.14.0为0 issues | [unit](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/backend-unit-complete.txt)、[build](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/backend-build.json)、[lint](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/backend-lint-complete.txt) |
| 迁移/quota/dedup | migrations、SQLite结算/去重/固定窗口/平台配额回归通过，最大227未变 | [回归](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/sqlite-migrations-and-quota.txt) |
| 前端 test/typecheck/lint/build/critical | 288文件/2099通过/2既有跳过；typecheck/lint/build通过；critical14文件/176通过/2既有跳过 | [tests](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/frontend-tests.txt)、[组合门禁](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/gate-summary.json) |
| 范围与冻结核对 | 34个代码/测试路径，1106个历史文件未改；VERSION/Go/迁移/依赖/Ent/Wire保持；新规格结构及链接通过 | [范围](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/scope-facts.json)、[代码](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/code-sha256.json)、[文档](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/document-check.json) |

## 实施事实

- 实施起点：`5c01d92c7cad41d6605925d076ee571fb420ee41`。
- 实现提交：`b554bea554202e3008b99e3faa7a29420467f318`（本地main）；提交中的34个代码/测试文件逐项匹配code-sha256验收字节。
- 验收快照：scope-facts和final-integrity记录的是提交前HEAD及未提交状态，原始值保留；当前提交状态以本节为准。
- VERSION / Go / 最大迁移：`1.1.20` / `1.27.3` / `227`；依赖、Ent/schema及生成Wire无改动。
- 未完成项：本次范围无未完成项，change保留活动目录等待归档；其他按需功能及依赖专项独立。
- 发布/部署/业务实例操作：无；仅在临时测试库中清理和验证迁移，使用本地HTTP/miniredis。

## 失败与纠正

原始失败输出保留。初版后台夹具漏了生产连接注册的NOW兼容函数，初版页面夹具缺i18n部分mock；先纠正夹具再复现真实问题。旧auto-pause mock的time.Time期望改为UTC RFC3339。两项新测试lint问题已修正。已有dashboardRepoStub异步计数读取race改为atomic.Int64后四包race和最终54包unit通过；没有更改仪表盘产品实现。[完整记录](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/README.md)。

基准CPU为i9-13900H，比例仅指签名函数，不宣称等同于1C1G VPS整体吞吐或端到端延迟。julianday用于个人规模账号/计划/结果的混合时间比较与排序，没有新增索引或迁移。
