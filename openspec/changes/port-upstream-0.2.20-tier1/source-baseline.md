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
| PR8553 | 66ea4d311f44158d869892b097da69859d4563d3 | 3c46c128330a8b4ad65025decf90d56155a08cc6 |
| PR8445 | e479d5d24955c1b5bd440727a2f29703ef26184f | 38029f3d2138047263ac2c1fcca9d1eaa7ea0bc6 |
| PR8520 | ec5ee14ca546c71b4c8cdb191a29905c9715c25f | e479d5d24955c1b5bd440727a2f29703ef26184f |
| PR8464 | a2f7c12cddb186ead273d55fbc1813825afd70c8 | ec5ee14ca546c71b4c8cdb191a29905c9715c25f |
| PR8563 | b6283399860c0d854f34c58717597f45a3d1afcc | e6c19082f24740a098f6753d0fbfef754098c8b5 |
| PR8429 | 2ee66a660f2e5ff2a3762677ab070744d53adb22 | b6283399860c0d854f34c58717597f45a3d1afcc |
| PR8471 | c0a3ce60b6f3f816bdf0d3df28d3df1e0d3576e8 | a2f7c12cddb186ead273d55fbc1813825afd70c8 |
| PR8437 | 1209e71e6d29f01494fdc3e7dce04659c385b21d | c86645c3b6d0f1e54654da7d824ac95b2d802533 |
| PR8441 | 52c83021e3c4d0e2021b2c1b2240888976106089 | 66ea4d311f44158d869892b097da69859d4563d3 |
| PR8409 | 2fcd47aa134d689ead51874745114aa9628830e4 | c0a3ce60b6f3f816bdf0d3df28d3df1e0d3576e8 |
| PR8432 | 38029f3d2138047263ac2c1fcca9d1eaa7ea0bc6 | 52c83021e3c4d0e2021b2c1b2240888976106089 |
| PR8420 | 2c93aea9f712c7e9396060acbeb9215d2c1883b0 | 2fcd47aa134d689ead51874745114aa9628830e4 |
| PR8518 | 0eee80ba87aa4f3aff7282bf1695aa6dc0471c36 | c5ca149a3d856d92fa0b61bea276338be718c5db |

所有PR取第一父→merge整体diff，不逐个挑微提交；按tag终态审查相互覆盖，混合PR只取source-feature-map中行为。四态在独立index先完成全部reverse，再全部forward，不向真实index/工作树应用。源补丁、哈希、调用扫描和编译标签见[证据](../../../docs/upstream-sync/evidence-0.2.20/MANIFEST.md)。ALREADY不等于整PR完成，CLEAN不等于可编译。
