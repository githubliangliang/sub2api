# Source Feature Map

| 行为 | 来源与当前缺口 | 接受边界 | 验收 |
|---|---|---|---|
| F01 清理任务周期运行 | fork DI 对 SQLite 跳过 Start；已有请求可创建任务 | 配置禁用仍禁用，重启能领取 pending/stale running；筛选与账本语义保持 | 真实 SQLite 的任务创建/领取/状态/分批删除/续跑与停止 |
| F01 到期自动暂停 | fork DI 跳过 AccountExpiryService | 只暂停启用 auto_pause_on_expired 的到期可调度账号；快照传播与幂等 | SQLite 的不同时间格式、未来/未配置/已暂停/软删账号及 outbox |
| F01 定时账号测试 | fork DI 跳过 ScheduledTestRunnerService | 已启用且到期的计划执行；保存/裁剪结果/推进下次时间；成功才恢复可恢复状态 | SQLite CRUD/时间读取/结果裁剪，服务执行与停止/重叠保护 |
| F02 操作日志全部清理 | AuditLogView → audit API → handler → AuditLogService → auditLogRepository | SQLite 清空成功，保留新留痕和单调 ID；TOTP 与权限保持 | 已有 SQL 错误的失败回归、真实 handler、按钮流程 |
| F03a 延迟单账号识别 | #8249 | 普通成功/其他平台请求不额外探测；显式标记优先；需要时请求内一次 | handler 调用计数、503 语义、并发/取消/分组 ID 快照 |
| F03b 映射签名 | #8250 | 不分配临时键列表；内容变化仍使既有映射缓存失效 | 顺序/交换/原地修改/非字符串与零分配，前后 benchmark |

不导入 PR 中只为上游接口差异服务的测试 stub 修改。来源 integration 文件需核对首行；不能将 `integration && postgres` 用例当成本 fork 的验收。
