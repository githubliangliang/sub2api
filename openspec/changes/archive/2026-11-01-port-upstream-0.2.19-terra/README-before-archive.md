# v0.2.19 用户选定：GPT-6.1 Terra

状态：2026-10-30用户要求实施选定的 Terra spec，**第三档新增能力已实现，14/14 任务完成，等待归档**。产品/测试实现已提交本地 main 为 `a541d0b611f4c2da5cab9811c3c4e02ae8344136`；[验收](./verification.md) / [实际起点、测试与哈希](../../../docs/upstream-sync/implementation-0.2.19-terra/README.md)。2026-10-28 的[选择记录](../../../docs/upstream-sync/evidence-0.2.19/feature-selection.md)及来源保持原文；该选择不扩展整 release 或发布/部署范围。

| 编号 | 行为 | 规格 |
|---|---|---|
| T01 | 模型身份、effort、工具协议、账号metadata | [gpt61-terra-compatibility](./specs/gpt61-terra-compatibility/spec.md) |
| T01 | 既有usage与价格策略接入 | [gpt61-terra-usage-pricing](./specs/gpt61-terra-usage-pricing/spec.md) |
| T01 | 既有白名单、预设与客户端配置 | [gpt61-terra-client-configuration](./specs/gpt61-terra-client-configuration/spec.md) |

- [proposal](./proposal.md) / [design](./design.md)：选定能力与适配边界。
- [source-baseline](./source-baseline.md) / [source-feature-map](./source-feature-map.md)：冻结来源。
- [tasks](./tasks.md) / [verification](./verification.md)：实施已核销，前后证据与适用门禁齐全。
- [评估与patch sites](../../../docs/upstream-sync/PORTING-0.2.19.md) / [证据](../../../docs/upstream-sync/evidence-0.2.19/MANIFEST.md)。

保留SQLite、miniredis、simple mode、固定quota窗口、成功usage dedup及既有metadata优先顺序。无新增schema/迁移/依赖或全局默认模型变更；其它新模型、受限管理员密钥、图片计量开关、诊断界面、支付与部署仍独立。

