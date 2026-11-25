# Source Baseline（冻结，实施不得回写）

评估日期2026-11-25。历史范围终点v0.2.25，不是新Release；实际实施起点另填verification。

| 项目 | 冻结值 |
|---|---|
| fork HEAD / tree | `b9d9524c6cce8488edb2eaf899d771176f3b2414` / `37f360f7f8f36afbb392c3fbdfccc94f331cc176` |
| 私有index基线 | `37f360f7f8f36afbb392c3fbdfccc94f331cc176`；启动已有2份文档修改及23份未跟踪文件，均保护 |
| VERSION / Go / SQLite / auth snapshot | `1.1.28 / 1.27.4 / 230 / 21` |
| upstream/main观察 | `0f6671457132425d1b84f3d80536b53e02942424` |
| 完整来源PR | #6929 |
| merge / first parent | `6206ce940ea9bca3708c8b005c219a7bd1d12a07` / `4726bdd08b6201d426a80529b79be123a4008d20` |
| 完整patch SHA256 | `c532c0d9c8ab8f8d409f909cbcb7806b40c33a1f8ce3d87b32004210463b27c0` |
| 来源规模 | 9 files changed, 1451 insertions(+), 34 deletions(-) |
| 当前来源文件四态 | {'CONFLICT': 4, 'CLEAN': 5} |

完整补丁与134个候选来源见[来源清单](../../../docs/upstream-sync/unmerged-review-2026-11-25/sources.json)。所有候选先reverse再forward，仅检查；完整来源不是自动授权整PR，新增调用索引不证明编译。
