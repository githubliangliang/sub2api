# 来源基线（冻结）

日期2026-10-02。实施时不得改写本文件，新起点记入verification。

| 来源 | SHA |
|---|---|
| 本fork main / VERSION 1.1.15 | `5eb58199ee6ef45515ec791b445305651384fc0c` |
| 上次上游v0.2.11 | `96f4c115c9749078f90cbf210a01d39baf3f53b6` |
| 目标v0.2.12 | `5106065716e494204fc0e8db16f68f6e9d576be0` |
| 抓取时upstream/main（不扩展范围） | `458b92abd4b0b09d123d4c6727dd8d59060bd883` |

全区间19条非merge/9PR/1直接提交，本批3PR/3簇。每PR按merge父1至merge取完整差异，不代表全部hunk采用。最大SQLite迁移226，无新迁移。

| 簇 | PR | Merge SHA |
|---|---|---|
| F01 | #7780 | `d6adebd22de00478cd021119ba755f37bcb94fb5` |
| F02 | #7674 | `43c4c882fd8dd952557ff2dd9ec54eb2cbca2eea` |
| F03 | #7773 | `46f5f5e8ba8d2fbebd824022ccc4b7b5c0173660` |

[候选](../../../docs/upstream-sync/evidence-0.2.12/candidates.tsv) / [四态](../../../docs/upstream-sync/evidence-0.2.12/files.tsv) / [分档](../../../docs/upstream-sync/evidence-0.2.12/decisions.tsv)。文件态不能证明运行正确。
