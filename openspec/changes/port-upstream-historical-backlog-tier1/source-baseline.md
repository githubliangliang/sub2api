# Source Baseline（冻结，实施不得回写）

评估日期2026-11-25。历史范围终点v0.2.25，不是新Release；实际实施起点另填verification。

| 项目 | 冻结值 |
|---|---|
| fork HEAD / tree | `b9d9524c6cce8488edb2eaf899d771176f3b2414` / `37f360f7f8f36afbb392c3fbdfccc94f331cc176` |
| 私有index基线 | `37f360f7f8f36afbb392c3fbdfccc94f331cc176`；启动已有2份文档修改及23份未跟踪文件，均保护 |
| VERSION / Go / SQLite / auth snapshot | `1.1.28 / 1.27.4 / 230 / 21` |
| upstream/main观察 | `0f6671457132425d1b84f3d80536b53e02942424` |
| 完整来源PR | #7001 |
| merge / first parent | `c7ed614240e9f2b2ac14f0bc4a2f48dfa1a86e98` / `a59f0c7fac0e653e9135605068c100d75ed5dc2c` |
| 完整patch SHA256 | `1f0ee5fe7e4430f7af91b3ca5b26ec1544e04a255a9b8739569d5651afe66776` |
| 来源规模 | 7 files changed, 169 insertions(+), 13 deletions(-) |
| 当前来源文件四态 | {'CLEAN': 7} |

完整补丁与134个候选来源见[来源清单](../../../docs/upstream-sync/unmerged-review-2026-11-25/sources.json)。所有候选先reverse再forward，仅检查；完整来源不是自动授权整PR，新增调用索引不证明编译。
