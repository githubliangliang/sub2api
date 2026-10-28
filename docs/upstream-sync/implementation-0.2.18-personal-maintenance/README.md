# 个人维护实施证据

2026-10-28，实施起点 `5c01d92c7cad41d6605925d076ee571fb420ee41`，实现已提交至本地main为 `b554bea554202e3008b99e3faa7a29420467f318`。产品实现、组合门禁和完成状态统一见 [verification](../../../openspec/changes/port-upstream-0.2.18-personal-maintenance/verification.md)。每个运行项目有同名 `.json` 命令/退出码/耗时记录和 `.txt` 原始输出。

## 来源与失败回归

- [文件四态](./source-file-states.tsv)：独立基线 index，全体先反向后正向；A0/C9/X2/N0。
- [来源测试首行](./source-test-inventory.tsv)、[当前测试目录](./test-inventory.txt)：避免将不存在或 PG-only 用例当作已执行验收。
- [审计日志清理基线](./baseline-audit.txt)：真实 SQLite 报 `near "TRUNCATE": syntax error`；[签名基线](./baseline-signature.txt)每次1次分配。
- [初版后台夹具](./baseline-background.txt)漏了生产连接注册的 NOW 兼容函数；[纠正夹具后的基线](./baseline-background-corrected.txt)仍独立复现旧 TEXT 扫描、到期/续跑比较、取消状态覆写、SQLite 启动与快照传播缺口。未把初版的 NOW 错误当作生产缺陷。
- [后台生命周期基线](./baseline-worker-lifecycle.txt)：到期服务重复 Start、SQLite 未启动与停止不能取消10秒延迟。
- [handler 探测基线](./baseline-handler-probes.txt)：Gemini/Anthropic 请求额外读取 Antigravity 桶。
- [初版页面夹具](./baseline-audit-button.txt)漏了 i18n 部分 mock；[纠正夹具后的页面基线](./baseline-audit-button-corrected.txt)复现清理中按 Enter 产生两次提交。
- [定时测试保存失败基线](./scheduled-execution-baseline.txt)：保存失败后仍恢复账号和推进计划；修复后保留原待执行计划。

## 性能测量

同一主机、同一夹具、每种规模3次、`-benchtime=200ms -benchmem`，使用中位数。CPU 为 i9-13900H；这是函数基准，不代表 VPS 整体吞吐或端到端延迟变化。[before](./benchmark-before.txt)、[after](./benchmark-after.txt)、[结构化结果](./benchmark-summary.json)。

| 映射数 | 修改前 ns/op | 修改后 ns/op | 耗时下降 | 修改前临时内存 | 修改后 |
|---|---:|---:|---:|---:|---:|
| 4 | 225.4 | 98.79 | 56.17% | 64 B / 1 alloc | 0 B / 0 alloc |
| 32 | 1973 | 680.1 | 65.53% | 512 B / 1 alloc | 0 B / 0 alloc |
| 128 | 13333 | 2620 | 80.35% | 2304 B / 1 alloc | 0 B / 0 alloc |

正常请求原本额外执行1次分组探测，修改后只在 Antigravity 的503/模型限流判断需要时执行，且请求内最多1次；显式 true/false 优先，原分组 ID 快照和取消边界均有用例。

## 数据与范围

所有 SQLite 数据、审计清理和账号测试均使用临时库、本地 HTTP 夹具、miniredis。未访问业务实例，未发布/部署。VERSION1.1.20、Go1.27.3、最高迁移227、Ent/Wire、依赖和历史冻结文件保留；具体范围与代码哈希在最终事实记录中核对。

[scope-facts.json](./scope-facts.json)核对34个代码/测试路径、1106个历史来源/证据文件；[code-sha256.json](./code-sha256.json)固定验收实现字节，逐项匹配上述实现提交。scope-facts与[final-integrity.json](./final-integrity.json)为提交前冻结快照，保留原HEAD和implementation_committed值；当前提交状态见verification。没有修改 VERSION、Ent/schema、生成 Wire、迁移或依赖。

## 组合检查中发现的夹具问题

- 首轮 backend-unit 的两个 account_repo_auto_pause_test mock 仍传入 time.Time 作为期望；生产已改为 UTC RFC3339 参数以正确做 SQLite 时间比较。真实 SQLite 用例已通过，更新 mock 期望后 backend-unit-final 通过。
- 首轮 backend-lint 的两项为新用例中未检查的 Rows.Close 和类型断言；已明确检查，backend-lint-complete 为0 issues。
- 首轮 targeted-race 发现已有 dashboardRepoStub.recomputeCalls 被异步重算写入、Eventually 读取；改为 atomic.Int64 后 targeted-race-final 四包通过。只修正测试夹具，未改变仪表盘产品实现。
