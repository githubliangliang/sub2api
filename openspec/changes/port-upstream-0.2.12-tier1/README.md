# 上游0.2.12 第一档

状态：开发及验收完成（2026-10-02），实现提交 `405246789`，分支 `sync/upstream-0.2.12`。3簇/3来源PR。第二档S02/S03由用户确认需要后从第三档调整，本轮用户已授权实施两批。

- [proposal](./proposal.md)：收益与范围
- [source-baseline](./source-baseline.md)：冻结SHA
- [source-feature-map](./source-feature-map.md)：来源与行为
- [design](./design.md)：适配决策
- [spec](./specs/single-node-correctness/spec.md)：行为契约
- [tasks](./tasks.md)：已完成任务
- [verification](./verification.md)：逐场景验收证据
- [PORTING](../../../docs/upstream-sync/PORTING-0.2.12.md)：完整分档与patch sites

保持SQLite/miniredis/simple mode及VERSION 1.1.15；不引入支付、TypeSafe、axios升级，不发布或部署。
