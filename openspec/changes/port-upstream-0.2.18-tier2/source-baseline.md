# 来源基线（冻结）

2026-10-23。实施时不得编辑本文件；新工作起点另记verification。

| 来源 | SHA |
|---|---|
| fork main / VERSION1.1.18 | `873962ce7f5c07a336479c823419a531529e6f99` |
| v0.2.16 | `01842576b3d5c1e2545b5a99ad555d36e28ed306` |
| v0.2.17 | `060d0160417782cd962a0c7f460f6cd4565d4916` |
| v0.2.18 | `ac1b8f1271be0d674fa6f074403c1410f30a3a81` |
| 抓取时upstream/main（tag后不含） | `5c185c5c5d4f5e409e2139806e0964ae944e03e8` |

本轮56非merge / 35PR / 8直接提交；本批10簇。按merge父1到merge完整diff评估，选定行为不等于整PR移植。

| 来源 | commit SHA |
|---|---|
| PR8106 | `3856a02856ca788cc1714362b1dc8a5a4d06b131` |
| PR8140 | `060d0160417782cd962a0c7f460f6cd4565d4916` |
| f2e5b55f3 | `f2e5b55f3e8069df1c02a9a27af537b5fdefa64e` |
| PR8121 | `43157c478ec29cb765043e2f0cba4f69728dd2f2` |
| PR8305 | `c5b811284aea2c4922561da2f0a78256334b23df` |
| PR8210 | `b7512305123aff307d3ab92840f2077eabc2529f` |
| PR8093 | `857a76d4a0e7b2389926c340555a2f8f5b24a0ae` |
| PR8060 | `3b83e045ed5ddc8246ebe8b3c4045a1c7ff5968b` |
| PR8076 | `abdc82725f1c6f1530d2a5835b0af653d352994e` |
| PR8168 | `0fa02816afcaea3610171dd229dd910c058c47e4` |
| PR8223 | `378984d87f2b68db15626ae8f18b9e923ab82b85` |
| PR8267 | `9a1680042a2315df7fcfa925f9235bfead528ad6` |
| PR8275 | `21cc926933163e015d2b632be07055759c983541` |

额外历史来源（只作最小前置）：#5925完整merge `4033387fd49b15b676d2e5c9fd3833f156050f57`（内部提交 `953028718d590dc64a79899f29f46c7c44f012cd`）；目录解析/流读取helper冻结在上表v0.2.18树，不移植缺失平台。

最大SQLite迁移226；本批不需要新schema。[四态](../../../docs/upstream-sync/evidence-0.2.18/files.tsv) / [逐来源决定](../../../docs/upstream-sync/evidence-0.2.18/decisions.tsv)。
