# v0.2.21 第二档

2026-11-11开始，2026-11-12（Asia/Shanghai）完成S01–S12，**42/42任务完成**。用户已授权本地commit，实现随本次提交保存，等待归档。实际起点 `4f4c80c3ce48ce4afbc9344536b9cf44ed3fd898` 加已有第一档代码；第一档后由独立执行流程提交为 `63c764704761073e69ff25ba41d64eeab5793cd0`，其28份产品哈希核对一致。

实现WS后续轮次失败关闭、OAuth搜索激活/完整回放、统一TTFT、显式模型冷却、Anthropic兼容错误策略、合成effort策略与校验、refresh token原子消费、Key独立用量/SQLite覆盖索引、账号刷新保留、Codex页面上下文、审计worker唤醒与Claude Code 2.1.321维护。S11实际SQLite验收发现并修正claim到期混合时间比较；不改变事务/claim_version仲裁。

后端全量unit、embed build、lint0、五包定向race及真实本地HTTP/SQLite/miniredis通过；前端2197通过/2既有跳过，critical179/2、typecheck/lint/build及桌面/390px移动浏览器通过。SQLite229迁移在临时库和1GiB/1CPU容器验证查询计划/结果/大小/写入成本，不能外推生产性能。VERSION1.1.24、Go1.27.3、auth snapshot21、依赖、Ent/Wire和64MiB/40MiB/120s默认值保留。

[验收](verification.md) · [任务](tasks.md) · [实施证据](../../../docs/upstream-sync/implementation-0.2.21-tier2/README.md) · [冻结来源](source-baseline.md) · [行为映射](source-feature-map.md) · [设计](design.md) · [原评估](../../../docs/upstream-sync/PORTING-0.2.21.md) · [原始评估证据](../../../docs/upstream-sync/evidence-0.2.21/MANIFEST.md)

原proposal/design/source/delta-spec和历史评估不改；实施前README/tasks/verification在证据目录保存。发布/推送/业务部署、支付、TTL扩展、新模型/平台与自适应预算不纳入本批。
