# Source Baseline（冻结）

冻结日期：2026-10-28。用户选择单独记录；实施不回填此文件的HEAD/VERSION/SHA。

| 项目 | 冻结值 |
|---|---|
| fork HEAD | 49939fb2cf96633688548ca39204d96558e59d79 |
| VERSION / Go | 1.1.20 / 1.27.3 |
| 最大SQLite迁移 | 227 |
| v0.2.18 | ac1b8f1271be0d674fa6f074403c1410f30a3a81 |
| v0.2.19 | dece26153621959ccd7635814f4b2d08304e681a |
| #8364 merge | b332f169f9c6c10722016ff7a09bc9c9658a3152 |
| #8364 first parent | ba81c7fd758e425d67f856d601147ef0c5263d3f |
| #8364完整diff | 18文件，+292/-17；A0/C11/X7/N0 |
| patch SHA256 | b229b7468006c32019afec39b83068cce8037b36bad8bc62807b39da29205d32 |

取完整merge第一父到merge的diff，不逐微提交挑选；[完整候选](../../../docs/upstream-sync/evidence-0.2.19/candidates.tsv)、[四态](../../../docs/upstream-sync/evidence-0.2.19/files.tsv)和[本轮复查](../../../docs/upstream-sync/evidence-0.2.19/review-file-checks.tsv)已冻结。

本轮全体276文件先反后正；3732产品文件与原采集一致。来源测试CLEAN不代表本地缺失的目录API可用，按现有消费者适配，不能引入未选定完整标准模型系统。

用户在本轮选择需要Terra，保留第三档，见[选择记录](../../../docs/upstream-sync/evidence-0.2.19/feature-selection.md)。此选择未改变过去第三/四档、支付或其它功能边界。patch sites与量测仅在[PORTING评估](../../../docs/upstream-sync/PORTING-0.2.19.md)。

