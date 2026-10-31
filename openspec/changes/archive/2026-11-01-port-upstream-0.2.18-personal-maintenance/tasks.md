# Tasks

## 来源与基线
- [x] 冻结 fork/tag/PR 来源，先反向后正向检查全体候选文件与新增调用。
- [x] 保存操作日志清理失败回归、后台缺口与优化前 benchmark。

## F01 SQLite 后台
- [x] 审查清理任务 SQL、时间读取、取消、续跑和用量/账本边界，恢复周期启动。
- [x] 验证到期判断、自动暂停筛选、调度传播和幂等，恢复后台启动。
- [x] 验证定时测试计划/结果 SQL、due 选择、结果裁剪与下次运行，恢复启动。
- [x] 验证启动/停止、取消、执行重叠与旧 TEXT 时间边界。

## F02 操作日志
- [x] 修复 SQLite 全量清理 SQL，验证删除数量、同步留痕、自增 ID 和失败行为。
- [x] 验证管理员/TOTP/API-key 边界及真实页面按钮流程。

## F03 轻量调度
- [x] 接入请求内延迟探测与所有当前消费者，验证显式标记/并发/取消/分组快照/503。
- [x] 实施零分配签名并验证映射失效、内容与顺序边界。
- [x] 保存相同夹具的优化后 benchmark 和正常请求探测次数对照。

## 组合验收
- [x] 定向 service/repository/handler、真实 SQLite/miniredis 与 race 通过。
- [x] 后端全量 unit/build、迁移/固定 quota/dedup 回归及 golangci-lint 通过。
- [x] 前端全量测试、typecheck/lint/build 与 critical 通过。
- [x] 核对 VERSION/Go/迁移/依赖/Ent/Wire/冻结证据及实际工作树范围。
- [x] 填写 verification，更新当前 README/AGENTS 状态与剩余未选范围。
