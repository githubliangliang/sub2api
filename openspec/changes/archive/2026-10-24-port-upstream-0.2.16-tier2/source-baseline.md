# 来源基线（冻结）

日期2026-10-13。实施时不得编辑本文件；新的工作起点只记入verification。

| 来源 | commit SHA |
|---|---|
| fork main / VERSION 1.1.16 | `128c5f7bfd4b8ea3f824146ed3b005e9d3e34887` |
| v0.2.12 | `5106065716e494204fc0e8db16f68f6e9d576be0` |
| v0.2.13 | `3040209f205472038c1ba745a1bedd2edd9053b1` |
| v0.2.14 | `0363b8cdba8cec3e2ba4b2dbd49c4481143fa55d` |
| v0.2.15 | `f2669c8cf62555cd92389b3f55920e9e6e7c6ff2` |
| 目标v0.2.16 | `01842576b3d5c1e2545b5a99ad555d36e28ed306` |
| 抓取时upstream/main，tag之后不含 | `4d9ac3e482106d2da3a39fe3e546d8d6801f1cc1` |

新区间149条非merge/89PR/4直接提交，本批16簇/30个本轮PR来源。所有PR取父1到merge完整diff评估；下表不表示整PR都采用。

| PR | merge SHA |
|---|---|
| PR7703 | `2eb7ccb6d258b6e75a7b849edb0119641d9dc024` |
| PR7711 | `6d509c632d1e03351ae3dacc84fb8dc4aee4f709` |
| PR7722 | `9663b5ccf05a0f2efefceef96727170e36042260` |
| PR7724 | `3149d6285178b9f55d5e651f1787f68d6b36afca` |
| PR7732 | `9e46f2d77ec87f5230f04dc01e53a750ce5ce232` |
| PR7738 | `c2885c430fe01d0d88b23d52e062ccbb073d0905` |
| PR7765 | `068239305ede5f7f4548d421485921f8438958c7` |
| PR7779 | `6db4171cfb7592ee6f81c29a82e9fcba0077155e` |
| PR7787 | `8d504668099c6bc1447c713395430a4b479d605e` |
| PR7788 | `9d49d275f34a6859b822f51ce314eb29d74e5273` |
| PR7792 | `ef16f8a3d1ca6fbcb06332c652c91223f4caabb4` |
| PR7793 | `8e2c4fad321fe8ce86d577dd7561c371bf602647` |
| PR7810 | `01842576b3d5c1e2545b5a99ad555d36e28ed306` |
| PR7818 | `b94974467b418a93638e808fa4797ee45ba200c8` |
| PR7827 | `129369204eaccac9a0ee3ea6f4ea0b35be7e4130` |
| PR7833 | `48c647b5d2ef83301de61e4301b1bb12e2f0207b` |
| PR7839 | `f3bcad3c3352eb21b351381131dc3cf1bc3373f2` |
| PR7840 | `2c6bd63de92769d55315d5bd9463abd0d2b1920f` |
| PR7841 | `02aad478f39961619fcbdd8371e33e13f2b020af` |
| PR7854 | `90afaf120e311c8b81f0732999b141457bb81fea` |
| PR7893 | `0363b8cdba8cec3e2ba4b2dbd49c4481143fa55d` |
| PR7900 | `1865f28469e8c66903bfb2c17ccaf31da302b41a` |
| PR7903 | `b126b94525f13da89233a01804d864e28adfd97d` |
| PR7904 | `7fc5332b45e1ad462b1aa0bba9ca7ae595bb26e3` |
| PR7910 | `2afe7d45777723d069b0f74e3cac2ef313539cf8` |
| PR7939 | `f2669c8cf62555cd92389b3f55920e9e6e7c6ff2` |
| PR7972 | `f2b1556bd4136d3d40fb582737fb1298ec37e72f` |
| PR8002 | `e468262294eb962206ce2d6e4b83f2e8bc43ef02` |
| PR8006 | `53c21fed252cec522035cfc2d4dd890a8c47140b` |
| PR8036 | `7b03b1f75a659c4b63f89948a16e04bb00a38180` |

旧来源的必要前置/维护（不计入149条）：

| 簇 | 来源 | SHA | 范围 |
|---|---|---|---|
| S03 | PR7638 | `4c00df2e0183e2c70b7fa8ba45914205e36aad0c` | 0.2.9 F05 structured-output beta必要前置 |
| S16 | PR7780 | `d6adebd22de00478cd021119ba755f37bcb94fb5` | 0.2.12 axios子项重新定级 |
| S16 | 4a1da2950 | `4a1da29509a6cf388ee26c58ec6185bee0cd8676` | 旧DOMPurify来源，目标版本按当前audit更新 |

[候选](../../../docs/upstream-sync/evidence-0.2.16/candidates.tsv) / [四态](../../../docs/upstream-sync/evidence-0.2.16/files.tsv) / [决定](../../../docs/upstream-sync/evidence-0.2.16/decisions.tsv)。最大SQLite迁移226；选定行为无需新schema。
