# Source Baseline（冻结）

冻结日期：2026-10-28。实施时不得回填新的HEAD、VERSION或来源SHA；实际起点和实现提交写入verification。

| 项目 | 冻结值 |
|---|---|
| fork HEAD | 49939fb2cf96633688548ca39204d96558e59d79 |
| VERSION / Go | 1.1.20 / 1.27.3 |
| 最大SQLite迁移 | 227 |
| 上游起点v0.2.18 | ac1b8f1271be0d674fa6f074403c1410f30a3a81 |
| 上游终点v0.2.19 | dece26153621959ccd7635814f4b2d08304e681a |
| 整区间规模 | 50非merge / 20PR + 3直接提交 |
| 整区间候选文件态 | 276条，A0/C203/X61/N12 |
| 本change来源候选文件态 | 26条，A0/C24/X2/N0，包含整来源PR的未选定子项 |

| 候选 | merge / direct SHA | first parent | 取法 |
|---|---|---|---|
| PR8377 | 3bd166babc300281f1830e6d7ce24256265a1d20 | 0232cb7137b840926e2dd23c62f1bee1c8557071 | 整merge diff或直接提交diff，按行为摘取 |
| PR8373 | dece26153621959ccd7635814f4b2d08304e681a | 184193f01ea0c681e2ab31c238c067aa9e7f34a6 | 整merge diff或直接提交diff，按行为摘取 |
| PR8354 | 8e231ff5d66afb5789f5a876872d2c8867828e82 | 98bcee7a8934220215268cccb80f15f3cb16091b | 整merge diff或直接提交diff，按行为摘取 |
| PR8326 | bb021d330e2067687d6362dc48991e63b3e56d6c | 3bd166babc300281f1830e6d7ce24256265a1d20 | 整merge diff或直接提交diff，按行为摘取 |
| PR8362 | f178e5ae32659e8901644a9f2378a0439ffa36ae | 449b65a2ed9ae43c8438105b41c500619253090f | 整merge diff或直接提交diff，按行为摘取 |
| PR8303 | 2c9d9760d8d60bc403d50d3d9b56d77c56164adf | af4462ebaef34d68c1e73e8622c4e195648fef75 | 整merge diff或直接提交diff，按行为摘取 |
| PR8352 | b2e7c609430736bd0a4b89c5b27a079ccef820d4 | f178e5ae32659e8901644a9f2378a0439ffa36ae | 整merge diff或直接提交diff，按行为摘取 |

每个PR均取merge第一父到merge的整体diff；source微提交不串行移植。多个簇共用PR时只计算一次来源；未选定子项不因CLEAN成为实现任务。

全部候选先完成反向检查，再完成正向检查；临时index固定在fork SHA，不向工作树或真实index应用补丁。[原始四态](../../../docs/upstream-sync/evidence-0.2.19/files.tsv)、[本轮复查](../../../docs/upstream-sync/evidence-0.2.19/review-file-checks.tsv)、[完整候选/补丁哈希](../../../docs/upstream-sync/evidence-0.2.19/candidates.tsv)和[冻结采集](../../../docs/upstream-sync/evidence-0.2.19/baseline.json)均保留。当前3732个产品文件与原始采集一致；历史status/仅文档提交转换不改写。

证据来自tag终态；原始及复查时upstream/main记录只用于说明，不纳入tag后提交。补丁落点与基座量测仅在[评估](../../../docs/upstream-sync/PORTING-0.2.19.md)，本change只定义行为和未来验收。

