# v0.2.21 → v0.2.23 第一档（已实施，随本次提交保存）

2026-11-18 按用户选择实现本批 21 簇、21 个 PR 的既有消费者维护。backend 全量 unit/embed/lint/race、frontend 完整门禁和桌面/移动浏览器验收通过；实际结果与来源适配见 verification。F20 来源 native 分支本地 NOBASE，按边界 N/A。

冻结本地基线是v1.1.25加用量排行路由修复 `7a0b78fe511f17b0f7eb9d26fcf5e4ff083312c7`，Go1.27.3、SQLite229、auth snapshot21。

- [范围、patch sites与最小前置](../../../docs/upstream-sync/PORTING-0.2.23.md)
- [冻结来源](source-baseline.md) / [来源行为映射](source-feature-map.md)
- [方案](proposal.md) / [设计](design.md) / [规格](specs/upstream-0-2-23-tier1-compatibility/spec.md)
- [实施任务](tasks.md) / [实施验收](verification.md)
- [冻结评估证据](../../../docs/upstream-sync/evidence-0.2.23/MANIFEST.md)
- [本轮实施证据](../../../docs/upstream-sync/implementation-0.2.23-tier1/README.md)

评估的PASS、预期失败探针和源码适用性判断不是实施验收。新模型/平台、强制Ultra Fast、渠道Ultra Fast倍率、30m/12h schema和大规模并发策略独立按需；支付功能排除。本方案不授权发布、推送、归档或业务实例操作。
