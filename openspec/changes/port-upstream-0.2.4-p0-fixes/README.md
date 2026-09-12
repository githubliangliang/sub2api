# 上游 0.2.4 第一档

状态：F01–F13 已在本地分支 `sync/upstream-024-p0-RLC1zS` 实施完成。
产品完成 SHA：`0eb88e504244c429ba19b0486b06c8011d78da79`。
目标回归、后端 build、完整 unit 重跑、SQLite 方言审计和用户 metadata 回归均通过。
无迁移、无前端、无依赖升级；未 push 或发布。
范围：PORTING F01-F12（本轮增量）及 F13（旧 API Key instructions 重判）。

- [proposal.md](./proposal.md)：问题与交付边界。
- [source-baseline.md](./source-baseline.md)、[source-feature-map.md](./source-feature-map.md)：冻结来源。
- [design.md](./design.md)：移植约束与风险。
- [specs/gateway-correctness/spec.md](./specs/gateway-correctness/spec.md)：行为契约。
- [tasks.md](./tasks.md)、[verification.md](./verification.md)：已执行任务、red/green、落地 SHA、来源排除与 tier 2 集成要点。
- [PORTING](../../../docs/upstream-sync/PORTING-0.2.4.md)：唯一 patch site 清单。

评估阶段记录与实施证据分开保存；中断且未完成的首轮 unit 不计作 PASS。
coordinator 已集成产品提交，原工作区应用及总 PORTING 状态由 coordinator 负责。
