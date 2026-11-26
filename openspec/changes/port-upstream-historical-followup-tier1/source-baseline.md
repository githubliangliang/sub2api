# Source Baseline（冻结，实施不得回写）

日期2026-11-26；本轮仅复评历史来源，不纳入上游新Release。实际实施起点另填verification。

| 项目 | 冻结值 |
|---|---|
| fork HEAD / tree | `89060f532a9000af28a47036f70cc44821dafd55` / `e6b78656de377c6f6fac648c67892ed40ed0bca6` |
| private index tree | `e6b78656de377c6f6fac648c67892ed40ed0bca6` |
| VERSION / Go / SQLite / auth snapshot | `1.1.28 / 1.27.4 / 230 / 21` |
| 来源终点 / upstream main观察 | v0.2.25 / `0f6671457132425d1b84f3d80536b53e02942424` |
| 本轮全部整PR / 四态 | 24 PR，61文件A0/C46/X13/N2，全部先反向后正向 |
| 起始既有状态 | 17份活动变更467项完成；仅原Fast诊断未跟踪；9,286份tracked保护 |

| 本批完整PR | merge / first parent | 完整patch SHA256 |
|---|---|---|
| #7308 | `60b9bf75589e71a26f08bbc06ec9e1e64a5ffe7f` / `d6e8b44bcb1247bc4f9730ce93e9ed7558cdb7f8` | `2d4b0d1d7f67324633302a97830b6854513a8d104e96f6f4591c7691a1f9268d` |
| #4867 | `6b71d75bdee26edf44d8fe8f1e934260c7b8c068` / `f53252e292423982e06a6f6ecc8bcf455f0b199b` | `f6735858e0d2c2fc1fe1e9a4b28afd7032311de734c6fac88d5b1ec24dde8201` |
| #7363 | `322cc5e5f1a0049215b73ed12736bfb585bea383` / `23dc073e65cd27656b98c673e19db1413139008e` | `841f9b05d377890da63e529276ac800b2928c67774816dc6dff700f6459f3e5c` |
| #7366 | `e4f97a53ec24a29820d59e53c6c02117e14c6342` / `9e175bd49161d486b3dd5d87f43724d98a45b07a` | `6f5399456a98d816ea421c8dd31de06b7bc52d28b34041809cd1b2521f0c2949` |

各来源规模/文件四态见[完整来源](../../../docs/upstream-sync/unmerged-review-2026-11-26/sources.json)。整PR只定义审查边界，不授权未选行为；只有check，没有apply，新增调用索引不证明编译。上轮完成的两份变更及旧证据不回写。
