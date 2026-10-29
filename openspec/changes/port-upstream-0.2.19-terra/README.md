# v0.2.19 用户选定：GPT-6.1 Terra

状态：2026-10-28用户确认实际需要Terra，**第三档新增能力已选定，提案未实施**。本轮原请求为release评估；[选择记录](../../../docs/upstream-sync/evidence-0.2.19/feature-selection.md)不等于整release或发布/部署授权。

| 编号 | 行为 | 规格 |
|---|---|---|
| T01 | 模型身份、effort、工具协议、账号metadata | [gpt61-terra-compatibility](./specs/gpt61-terra-compatibility/spec.md) |
| T01 | 既有usage与价格策略接入 | [gpt61-terra-usage-pricing](./specs/gpt61-terra-usage-pricing/spec.md) |
| T01 | 既有白名单、预设与客户端配置 | [gpt61-terra-client-configuration](./specs/gpt61-terra-client-configuration/spec.md) |

- [proposal](./proposal.md) / [design](./design.md)：选定能力与适配边界。
- [source-baseline](./source-baseline.md) / [source-feature-map](./source-feature-map.md)：冻结来源。
- [tasks](./tasks.md) / [verification](./verification.md)：全部待执行/空实施验收。
- [评估与patch sites](../../../docs/upstream-sync/PORTING-0.2.19.md) / [证据](../../../docs/upstream-sync/evidence-0.2.19/MANIFEST.md)。

保留SQLite、miniredis、simple mode、固定quota窗口、成功usage dedup及既有metadata优先顺序。无新增schema/迁移/依赖或全局默认模型变更；其它新模型、受限管理员密钥、图片计量开关、诊断界面、支付与部署仍独立。

