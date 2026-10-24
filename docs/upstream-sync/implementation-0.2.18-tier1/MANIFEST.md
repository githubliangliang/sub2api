# v0.2.18 第一档实施证据

2026-10-23；从`4984327590ed29eb4d842bfae8aa3406ef497008`干净工作树实施，当前未提交。最终状态见[验收报告](../../../openspec/changes/archive/2026-10-24-port-upstream-0.2.18-tier1/verification.md)。

- [start.json](./start.json)记录起点、VERSION与冻结source-baseline哈希；[code-sha256.json](./code-sha256.json)固定17个后端源码/测试文件。
- [checks.json](./checks.json)记录最终unit/build/race/SQLite与Fast来源格式检查的命令及实际退出码；[environment.txt](./environment.txt)记录实际Go工具链。
- [backend-unit.txt](./backend-unit.txt)是最终代码的54包全量unit通过记录；[backend-unit-initial.txt](./backend-unit-initial.txt)为补充来源价卡格式适配前的通过快照。
- [backend-build.txt](./backend-build.txt)是最终构建输出（成功无stdout/stderr），exit0见checks.json；[backend-race.txt](./backend-race.txt)为4包定向race通过。
- [backend-targeted.txt](./backend-targeted.txt)覆盖F02–F06和相邻行为；后续F03来源格式补测与最终通过在[f03-final.txt](./f03-final.txt)，也由最终全量unit/race覆盖。F01通过在[f01-after.txt](./f01-after.txt)。
- [backend-sqlite.txt](./backend-sqlite.txt)覆盖真实SQLite删除Key、去重、全部计数效果、错误回滚、固定quota窗口和SQL方言；[backend-test-inventory.txt](./backend-test-inventory.txt)确认实施测试实际编译存在。
- [openspec-strict.txt](./openspec-strict.txt)是最终文档严格校验；[consistency.txt](./consistency.txt)记录任务/需求/链接/代码清单及范围检查。

`f01/f02/f03/f04/f05/f06-before.txt`、`f02-timeout-before.txt`、`f03-source-fields-before.txt`为缺陷回归失败证据；旧评估probe未重用。`f01-test-setup.txt`仅为测试脚手架编译错误；`f02-intermediate.txt`和`f02-f03-intermediate.txt`保留中间验证，原因及最终解决见验收报告，均不冒充最终通过。

无支付、新平台、schema、依赖、VERSION变更；未连接供应商或业务实例，未提交/发布/部署。
