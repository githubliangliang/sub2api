# Design

按完整PR/安全合并簇理解最终行为，逐个采用本地已有消费者需要的最小hunk。先核对当前HEAD与冻结来源的差异，再构造失败复现和控制。来源新增模块仅在本地消费者确实需要时建立，不携带未选平台与策略。

## Decisions

- 以现有HTTP/WS raw载荷及SQLite事务为边界；调度/计费前检查歧义，稳定保留原始模型、客户端effort与最终实际值。
- 不可逆OAuth轮换完成允许有界保存；并发重新授权由条件更新保护，缓存/outbox只发布持久化状态。不同平台的晚到成功规则分别保持。
- metadata、快照、模型能力取决于可服务账号，聚合需要保守交集；UI异步请求需防止旧结果覆盖新状态。
- 平均模型TPS按累计耗时，已有输出TPS分布仍用逐样本连续分位数与算术mean；同名速率不可混用。
- 性能hunk先做等价/转义/fuzz与单P前后成本，异步字段拥有独立存储。函数级测量不替代1C1G实际实例验收。

## Risks and acceptance

拒绝同值重复字段会改变宽容输入行为，应覆盖大小写、转义、嵌套及freeform合法键。完整入口挂载、后续WS帧、断开结算、CAS并发/锁租期、SQLite取消/错误、缓存旧请求与body生命周期需实际验收。S11需要真实Wire生成；不需要新迁移。参考[任务](tasks.md)与[规格](specs/upstream-0-2-24-tier1-compatibility/spec.md)。

## Preserved constraints

SQLite-only/miniredis/simple mode；用户固定quota与usage去重；已选Fast/Ultra Fast、GPT-6.2、过滤与requested/effective effort；64MiB请求体预算、40MiB SSE、8MiB暂存及120s idle。第三档只保留建议，支付与缺平台不实现。组合验收完成前不得填写实施成功槽或勾选任务。
