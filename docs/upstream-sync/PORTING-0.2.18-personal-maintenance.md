# 个人维护补充：SQLite 后台、操作日志清理、轻量调度

2026-10-28 用户选定并直接授权实施三个项目。本补充不评估新 release，不改写旧四档判断或历史来源；F03 的第三档能力由本次明确选择纳入，F01/F02 是 fork 自身维护。[OpenSpec](../../openspec/changes/port-upstream-0.2.18-personal-maintenance/README.md)。

## 1. 来源与范围

fork 起点 `5c01d92c7cad41d6605925d076ee571fb420ee41`，VERSION1.1.20、Go1.27.3、最高迁移227。上游 #8249/#8250 使用完整 merge diff；11文件四态 A0/C9/X2/N0，见[冻结来源](../../openspec/changes/port-upstream-0.2.18-personal-maintenance/source-baseline.md)和[原始文件态](./implementation-0.2.18-personal-maintenance/source-file-states.tsv)。

## 2. 方法

用独立 index 固定来源基线，对全体文件先 reverse check 再 forward check；生产新增调用与真实消费者一起验证。SQLite 后台逐项跟踪 SQL，不依据跳过函数的旧注释认定 PG-only。回归在临时数据中复现，不清空业务实例。

## 3. F01 SQLite 后台 patch sites

- `backend/internal/service/wire.go` 的 ProvideUsageCleanupService、ProvideAccountExpiryService、ProvideScheduledTestRunnerService：三处仍跳过 SQLite 的 Start。
- `backend/internal/repository/usage_cleanup_repo.go`：领取、状态、分批筛选与时间读取；`backend/internal/service/usage_cleanup_service.go`：周期/取消/续跑。
- `backend/internal/repository/account_repo.go` 的 AutoPauseExpiredAccounts：到期比较、outbox 与快照传播；`account_expiry_service.go`：启动/停止。
- `backend/internal/repository/scheduled_test_repo.go`：计划/结果/时间读取；`scheduled_test_runner_service.go`：有界执行、结果保存、恢复与停止。

## 4. F02/F03 patch sites

- 操作日志：`AuditLogView.vue` 的 submitClear → `api/admin/audit.ts` → `audit_log_handler.go` 的 Clear → `audit_log_service.go` 的 ClearAll → `audit_log_repo.go` 的 TruncateAll。基线真实 SQLite 报 `near "TRUNCATE": syntax error`；保留 TOTP 与同步留痕。
- #8249：gateway_handler.go 两处、gemini_v1beta_handler.go 一处改为请求内延迟探测；gateway_scheduling.go、request_metadata.go、antigravity_gateway_retry.go 接入实际消费者。当前 eager 调用/查询方法均存在；新增 probe helper 由来源定义并与消费者一起落地。
- #8250：account.go 的 modelMappingSignature 当前每次排序并分配键列表；GetModelMapping 的两个调用都在。新增 FNV/fmix helper 随函数一起落地，不独立添加无消费者的基座。

## 5. 不做与历史遗留

旧第一/第二档和六项已选个人功能均完成，不依据冻结旧报告的未勾任务重开。运维分位数、费用趋势、Claude 重置额度、新模型、新平台、快照解码缓存/COW、PG quota SQL、pinned/live、Fast/Ultrafast、支付仍不扩围。依赖审计与 xlsx 专项保持独立。本补充的功能不隐含发布、部署或迁移业务数据授权。

## 6. 已有基础

固定 quota 窗口、成功用量 dedup、调度 outbox、可选 miniredis 与审计 TOTP 保留。NOW() 已有 SQLite 兼容函数，不能把所有出现 NOW 的 SQL 作为未适配证据。时间列仍需要验证实际驱动扫描与比较。

## 7. 实施顺序

冻结来源及失败回归/前置 benchmark → 三个后台服务逐项验收 → 操作日志完整链路 → 两项调度优化 → 组合门禁及当前状态回填。

## 8. 验证

[verification](../../openspec/changes/port-upstream-0.2.18-personal-maintenance/verification.md)统一记录实际结果；[实施证据目录](./implementation-0.2.18-personal-maintenance/)保存命令、退出码、回归和前后 benchmark。新用例必须在 SQLite 构建中实际执行，不能引用 PG-only integration 作为通过。

## 9. 本次实际结论

- 旧时间的Scan类型与SQL比较是独立边界，兼容读取还需同时核查RFC3339/offset的时间顺序。
- 后台恢复须验证状态、停止与重叠；迟到回写不能覆盖取消，结果保存失败不能恢复账号或推进计划。
- 按钮disabled不能拦截输入框Enter，提交方法也需进行中保护及事件回归。
- 裸连接漏注册兼容函数、旧mock参数和测试计数race先定位夹具，不反推生产缺陷或扩大产品修改。
- 相同夹具实测签名0 B/0 alloc，三种规模耗时下降56%–80%；lazy probe保留显式标记、原分组及取消/并发语义。

## 10. 实施完成

2026-10-28 三项开发验收完成，后端54包unit/build、四包race、SQLite迁移/quota/dedup及lint0 issues通过；前端2099通过/2既有跳过，typecheck/lint/build/critical通过。实现已提交至本地main为`b554bea554202e3008b99e3faa7a29420467f318`，见[verification](../../openspec/changes/port-upstream-0.2.18-personal-maintenance/verification.md)和[代码/范围快照](./implementation-0.2.18-personal-maintenance/scope-facts.json)。VERSION1.1.20/Go1.27.3/最大迁移227/依赖/生成代码保持，历史1106个来源/证据文件未改。未发布、部署或操作业务实例。
