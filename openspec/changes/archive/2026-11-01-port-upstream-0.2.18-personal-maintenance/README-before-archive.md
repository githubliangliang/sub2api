# SQLite 后台补全、操作日志清理与轻量调度优化

状态：**2026-10-28 三项已实现并完成开发验收，16/16任务完成**。实现已提交至本地main为`b554bea554202e3008b99e3faa7a29420467f318`；起点为`5c01d92c7cad41d6605925d076ee571fb420ee41`。见[verification](./verification.md)与[验收代码哈希](../../../docs/upstream-sync/implementation-0.2.18-personal-maintenance/code-sha256.json)。未发布/部署。

| 编号 | 范围 | 行为规格 |
|---|---|---|
| F01 | SQLite 使用记录清理、到期自动暂停、定时账号测试 | [sqlite-background-jobs](./specs/sqlite-background-jobs/spec.md) |
| F02 | 操作日志“全部清理”报错 | [audit-log-cleanup](./specs/audit-log-cleanup/spec.md) |
| F03 | 延迟单账号探测与模型映射签名零分配 | [lightweight-scheduling](./specs/lightweight-scheduling/spec.md) |

F01/F02 是本 fork 的恢复与修复；F03 来自上游 #8249/#8250，仍保留历史第三档来源。用户本次已直接授权实现，无需再次确认。快照解码缓存/COW、PG quota SQL、运维分位数、趋势切换、新模型、支付和部署不在范围内。

- [proposal](./proposal.md)、[design](./design.md)：范围和实现约束。
- [source-baseline](./source-baseline.md)、[source-feature-map](./source-feature-map.md)：冻结来源与行为映射。
- [tasks](./tasks.md)、[verification](./verification.md)：实际完成状态与证据。
- [补充评估与 patch sites](../../../docs/upstream-sync/PORTING-0.2.18-personal-maintenance.md)。

保留 SQLite-only、可选 miniredis、simple mode、固定 quota 窗口、成功用量去重和审计 TOTP。VERSION 1.1.20、Go 1.27.3、最高迁移227及依赖基线不因选择功能而变更。

后端54个测试包、build、golangci-lint 0 issues、四包定向race和真实SQLite/miniredis通过；前端2099项通过/2项既有跳过，typecheck/lint/build及critical通过。三个后台服务恢复，旧时间扫描/比较、取消/停止/重叠与结果存储边界一并验证；清理保留现场TOTP、同步留痕和历史自增序列。
