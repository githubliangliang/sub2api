# 用户选定：窗口级输出 TPS 分布

2026-11-04 用户授权实现历史 PR #7911 的剩余运维统计能力，保持第三档按需选择的来源身份。使用日志的单条输出 TPS 已完成，本次沿现有运维概览补齐整个窗口的 P5/P10/P50、均值与样本数。

- [proposal](proposal.md)、[design](design.md)、[source-baseline](source-baseline.md)、[source-feature-map](source-feature-map.md)
- [规格](specs/ops-output-tps-distribution/spec.md)、[tasks](tasks.md)、[verification](verification.md)
- [适配与 patch sites](../../../docs/upstream-sync/PORTING-output-tps-distribution.md)、[实施证据](../../../docs/upstream-sync/implementation-output-tps-distribution/README.md)

基于 v1.1.22 / Go1.27.3 / SQLite227，不需要新迁移、依赖或 Ent/Wire 生成。其它按需能力、发布和业务实例部署独立。

已提交`f333aa5f3aa7a6f6ecc4d1620cbfc47a737e16e3`并完成10/10任务，等待归档。详见[验收](verification.md)及实际证据；用户随后授权推送main，不创建新发布。
