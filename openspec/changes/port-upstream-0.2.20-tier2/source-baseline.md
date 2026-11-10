# Source Baseline（冻结）

冻结日期：2026-11-10。实施不得改写本文件；实际实施起点另记verification。

| 项目 | 冻结值 |
|---|---|
| fork HEAD | 326513cdc68cc50d31e25a07ae98ecd6f5a3ca1f |
| 最初观察HEAD | 5c135c31dfbd6a1fcb701767f2f78933c4280aea（Fast尚为工作树改动；随后并发操作提交，产品内容哈希一致） |
| private index snapshot tree | 77e4e07303db135109fb85c1136eda035889a43a |
| VERSION / Go / SQLite最大迁移 | 1.1.23 / 1.27.3 / 228 |
| v0.2.19 commit | dece26153621959ccd7635814f4b2d08304e681a |
| v0.2.20 commit | dab829a4e9b7d1cc307189aea8b0319eb29e29d0 |
| v0.2.20 annotated tag object | ec41309811cd30af310a0e4b9119b976ce65c9be |
| upstream/main（只记录） | ee58c0013c76ccd3f36439881bc848d342ec1c9d |
| 整区间 | 72非merge / 35PR + 7直接提交；562候选×文件，A3/C330/X157/N72 |

| 来源 | merge SHA | first parent |
|---|---|---|
| PR7979 | a3298b46ebe239fb0720b068a1eb3383955bdac2 | a9240d5016c226b341242a8c54a01a8605cbbc72 |
| PR8499 | ee74473e038b5c956897cda7aeb9903f3fa3138a | e320a1500be3eb546f85ab29876e38f9841c51df |
| PR8465 | 45b51a8ac4e2096fa440788953b2038226bfc739 | 4a4a83ee5ec86b19d3a2dd9b6471de5128963bec |
| PR8365 | dab829a4e9b7d1cc307189aea8b0319eb29e29d0 | d22e0724d6a28435bcf21280333e89883d069a5e |
| PR8484 | 503617a82fa5b58e4bb1954fb98d4494233d9193 | 876b96d204855ad1aac486b7dd39a95d377d4760 |
| PR8518 | 0eee80ba87aa4f3aff7282bf1695aa6dc0471c36 | c5ca149a3d856d92fa0b61bea276338be718c5db |
| PR8457 | 72c2acdcb92af7bded126cd41db45f7347494054 | 0ff31d7a99c1c6f514e48955fab109705e2c8ed6 |
| PR8318 | e320a1500be3eb546f85ab29876e38f9841c51df | 393167ea9a444f9f22972453686b53e993f90cd4 |
| PR8521 | c86645c3b6d0f1e54654da7d824ac95b2d802533 | c70ef412e33f45a14eff16f5010efa6e8c4546ae |
| PR8430 | 9aaa72cabdf7b3b8866d7b84cc0dd3d36049c119 | 1209e71e6d29f01494fdc3e7dce04659c385b21d |
| PR8436 | 393167ea9a444f9f22972453686b53e993f90cd4 | ccc2c161b5121d16fede588004ed05eb92d96f1b |
| PR8531 | c70ef412e33f45a14eff16f5010efa6e8c4546ae | 2ee66a660f2e5ff2a3762677ab070744d53adb22 |

所有PR取第一父→merge整体diff，不逐个挑微提交；按tag终态审查相互覆盖，混合PR只取source-feature-map中行为。四态在独立index先完成全部reverse，再全部forward，不向真实index/工作树应用。源补丁、哈希、调用扫描和编译标签见[证据](../../../docs/upstream-sync/evidence-0.2.20/MANIFEST.md)。ALREADY不等于整PR完成，CLEAN不等于可编译。
