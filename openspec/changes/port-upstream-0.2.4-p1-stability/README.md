# 上游 0.2.4 第二档

状态：S01-S12 共 12 簇已完成并与第一档集成，最终构建与总检查通过。
目标是改善单机资源占用、请求生命周期、Codex/Claude 兼容性及已有管理页面；无新增迁移。

- [proposal.md](./proposal.md)、[design.md](./design.md)：目标与关键决策。
- [source-baseline.md](./source-baseline.md)、[source-feature-map.md](./source-feature-map.md)：冻结来源。
- [specs/single-node-stability/spec.md](./specs/single-node-stability/spec.md)：行为验收契约。
- [tasks.md](./tasks.md)、[verification.md](./verification.md)：实施待办与空白验收表。
- [PORTING](../../../docs/upstream-sync/PORTING-0.2.4.md)：patch sites、四态证据和依赖判断。

第三、四档未立项；这份 change 不包含整个 v0.2.4，也不包含其后的 upstream/main。

最终集成证据见 [implementation.md](../../../docs/upstream-sync/evidence-0.2.4/implementation.md)。未推送或发布。
