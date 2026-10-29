# v0.2.19 第二档：Grok一致性、上传、搜索、Gemini与依赖维护

状态：2026-10-29 用户授权后已实施，**23/23 任务完成**。代码提交本地 main：`cccf931efe201671f8cd27284b04e74114f4f894`；[完整实施证据](../../../docs/upstream-sync/implementation-0.2.19-tier2/README.md)。2026-10-28 冻结评估来源保持原文，change 仍活动待归档。

| 编号 | 行为 | 规格 |
|---|---|---|
| S01 | Grok响应模型审计 | [grok-response-model-audit](./specs/grok-response-model-audit/spec.md) |
| S02 | Grok目录与受管快照一致性 | [grok-catalog-consistency](./specs/grok-catalog-consistency/spec.md) |
| S03 | 图片上传完整性 | [gateway-image-upload-integrity](./specs/gateway-image-upload-integrity/spec.md) |
| S04 | 渠道搜索平台维度保存 | [channel-search-dimensions](./specs/channel-search-dimensions/spec.md) |
| S05 | 既有Gemini模型接口与默认值 | [gemini-model-list-consistency](./specs/gemini-model-list-consistency/spec.md) |
| S06 | x/net安全维护 | [gateway-dependency-maintenance](./specs/gateway-dependency-maintenance/spec.md) |

- [proposal](./proposal.md) / [design](./design.md)：范围与适配决策。
- [source-baseline](./source-baseline.md) / [source-feature-map](./source-feature-map.md)：冻结SHA与行为来源。
- [tasks](./tasks.md) / [verification](./verification.md)：实际实施任务和验收记录。
- [评估与patch sites](../../../docs/upstream-sync/PORTING-0.2.19.md) / [评估证据](../../../docs/upstream-sync/evidence-0.2.19/MANIFEST.md)。

保留SQLite-only、miniredis、simple mode、成功usage dedup、固定quota窗口及迁移227。VERSION1.1.20/Go1.27.3是冻结评估值；实际起点另记，不改历史证据。Terra是用户选定的独立第三档提案；其它未选定新能力、支付、发布与部署不在此范围。

