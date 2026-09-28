# 来源基线（冻结）

评估日期：2026-09-28。实施时不得改写本文件；新的实施起点填写verification。

| 来源 | Commit |
|---|---|
| 本fork main/v1.1.13 | `bfbcd79bdc4aa0838616405466750658896e6920` |
| 上次上游v0.2.5 | `86f93c28ee34cc74b629dafb748bd5ac5ca8c5ea` |
| 上游v0.2.7 | `aea725f2ea644d5592d0bbb1d63b607efa7e200a` |
| 上游v0.2.8 | `fd80b08c90b55edcad5b00171b53f08721d30da1` |
| 目标终态v0.2.9 | `4c00df2e0183e2c70b7fa8ba45914205e36aad0c` |

固定VERSION=1.1.12，最大SQLite迁移226。无共同祖先，不直接merge。
整个来源范围215条非merge、161PR与9直接提交；本批为48PR。
四态来自独立index；文件ALREADY/CLEAN不代替编译与行为检查。
证据：[候选](../../../docs/upstream-sync/evidence-0.2.9/candidates.tsv)、[四态](../../../docs/upstream-sync/evidence-0.2.9/files.tsv)、[分档](../../../docs/upstream-sync/evidence-0.2.9/decisions.tsv)。
