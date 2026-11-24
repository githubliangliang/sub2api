# Source Baseline（冻结，实施不得回写）

冻结日期 2026-11-24；完整来源范围仅 v0.2.24..v0.2.25。真实实施起点填 verification。

| 项目 | 冻结值 |
|---|---|
| fork HEAD / tree | `39ca66d51d53b98f11349598bbd16cd8ecd0e2e0` / `cd84cf715cc17aec530c995b2e8fff8b52e46db0` |
| VERSION / Go / SQLite / auth snapshot | `1.1.27 / 1.27.4 / 230 / 21` |
| 私有 index tree | `cd84cf715cc17aec530c995b2e8fff8b52e46db0` |
| v0.2.24 commit / tag object | `5c2982a81b46e2142d0143d2f83c1ce9c4da1d24` / `7515db6dc56c2a6120dfb7abed70ef089ebc1457` |
| v0.2.25 commit / tag object | `348997267d756eab5729bb6b78a275a15d18796a` / `8fab111f92447a3edbcdf67a517e45b6a3cc4d31` |
| upstream/main（仅观察） | `0f6671457132425d1b84f3d80536b53e02942424` |
| 非 merge / 整体候选 | 12 / 3 整 PR + 1 VERSION 直接提交 |
| 净 diff / 候选×文件状态 | 91 文件 +3486 -319 / 98 条 |
| ALREADY / CLEAN / CONFLICT / NOBASE | 0 / 56 / 32 / 10 |

| 完整候选 | merge / first parent | 完整 patch SHA256 |
|---|---|---|
| PR9008 | `4d4433df09d27e7122e58559f77098c8e077b312` / `328db26f450edbdf9f8d4344a3ea903ac6396842` | `e3a9a8d1296df4ac46e5e50e3e1e8e7647f315ddc122e0580a5d356507d2d526` |
| PR9045 | `00c51f07d2d2707d73808fee78c149735478006e` / `4d4433df09d27e7122e58559f77098c8e077b312` | `dfbfbaf2d32aee69bbeedf6b3840a95b98a4b7aedab9773c78a0d2e0d2d2dce9` |

同一整 PR 被按独立行为分档；完整 diff 是审查边界，不代表整 PR 获选。S02 只取 PR9008 的依赖维护，不带名称同步；来源安全修复与本地适用性分别判断。全部 98 文件先统一 reverse 再 forward，仅 check。完整记录见[评估证据](../../../docs/upstream-sync/evidence-0.2.25/MANIFEST.md)。
