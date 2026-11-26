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
| #7500 | `5e4968492f2eb31b672618f0028a30ca699f63e5` / `a81547f244dca300625b818caf4facbde60e5e17` | `18e337a8dcc0ae3f8b1d73a5ee26af8d98da4c28ec6999f67294169ab3da2505` |
| #7373 | `023b38354feb5c1b2371dda51a6973d8bb084033` / `a765b3c1765f50a18e3b845170916d91c21a12cc` | `6731b255fd2a259efabb2ae8e4794aa9486870e982bd25df0a6faae7b76e20fc` |
| #7320 | `7865051fdf93823cb55c099424345ca74ba3a1f1` / `4b0adb68636c4d8bef479e2c1ed541950a031c27` | `e513288323a6b6ab5817951dbd615ae135c26860a528aed96e9d615a805250a9` |
| #7365 | `62a8919a1c82b902a93818eaf2a15ff972d0e610` / `1ea00258d0aa8323c8d67dfc428a4cf03b7cdae2` | `ffdff3f8ebe3de0cf95432762c2755c165ad24fa9df3fc183a30807840eaa927` |

各来源规模/文件四态见[完整来源](../../../docs/upstream-sync/unmerged-review-2026-11-26/sources.json)。整PR只定义审查边界，不授权未选行为；只有check，没有apply，新增调用索引不证明编译。上轮完成的两份变更及旧证据不回写。
