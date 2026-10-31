# Source Baseline（冻结）

冻结日期2026-10-31；后续实施起点与实现提交写verification，不回写此文件。

| 项目 | 冻结值 |
|---|---|
| fork HEAD | `5250907f5b8294eabed3d3101d2c4bf79936e0f7` |
| VERSION / Go / 最大迁移 | 1.1.21 / 1.27.3 / SQLite227 |
| 已评估上游v0.2.19 | `dece26153621959ccd7635814f4b2d08304e681a` |
| 来源PR8339 merge | `15701ba187af5827d740901d53421bd69ed97717` |
| 来源第一父 | `b332f169f9c6c10722016ff7a09bc9c9658a3152` |
| 原整PR补丁SHA256 | `2e1ca0ff4171d138b54a7e4c9ef8303a5968e300c7dc3a1ac8ff042d64094180` |

来源使用冻结完整PR差异及tag终态。S01模型审计和S02目录一致性已经完成；本次只取诊断/详情/同步结果/前端交互子集，不能重落整44文件PR。

[冻结整PR补丁](../../../docs/upstream-sync/evidence-0.2.19/patches/PR8339.patch)、[原分档](../../../docs/upstream-sync/PORTING-0.2.19.md#51-第三档t01-已选定其它能力继续等待用途选择)、[本次起点及保护文件](../../../docs/upstream-sync/implementation-historical-client-compatibility/baseline.json)。
