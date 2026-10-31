# 历史兼容性尾项：Codex 工具与 Responses 事件

状态：2026-10-31 历史未合项复查后建立的**第二档建议方案，未实施**。用户当前请求是分析值得合并的功能；本方案不把新增第三档能力计入实施范围。

| 编号 | 行为 |
|---|---|
| C01 | 普通工具参数完整性、坏历史调用与匹配输出处理 |
| C02 | 无歧义自定义工具namespace别名还原与声明优先级 |
| C03 | 最终声明失去内建tool_search时清理孤儿defer_loading |
| C04 | Responses事件序号始终序列化且保留正确顺序 |

- [proposal](./proposal.md) / [design](./design.md)：范围与适配决策。
- [source-baseline](./source-baseline.md) / [source-feature-map](./source-feature-map.md)：冻结来源。
- [行为规格](./specs/existing-gateway-correctness/spec.md)、[tasks](./tasks.md)、[verification](./verification.md)：未来实施与验收。
- [评估与patch sites](../../../docs/upstream-sync/UNMERGED-REVIEW-2026-10-31.md)、[实际复查证据](../../../docs/upstream-sync/unmerged-review-2026-10-31/README.md)。

保留SQLite/miniredis/simple mode、现有usage/dedup、固定quota窗口、VERSION1.1.21、Go1.27.3、迁移227和依赖/生成代码。Grok诊断、新模型、付款及其它未选能力独立。
