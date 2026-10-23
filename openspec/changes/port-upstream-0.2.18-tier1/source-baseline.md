# 来源基线（冻结）

2026-10-23。实施时不得编辑本文件；新工作起点另记verification。

| 来源 | SHA |
|---|---|
| fork main / VERSION1.1.18 | `873962ce7f5c07a336479c823419a531529e6f99` |
| v0.2.16 | `01842576b3d5c1e2545b5a99ad555d36e28ed306` |
| v0.2.17 | `060d0160417782cd962a0c7f460f6cd4565d4916` |
| v0.2.18 | `ac1b8f1271be0d674fa6f074403c1410f30a3a81` |
| 抓取时upstream/main（tag后不含） | `5c185c5c5d4f5e409e2139806e0964ae944e03e8` |

本轮56非merge / 35PR / 8直接提交；本批6簇。按merge父1到merge完整diff评估，选定行为不等于整PR移植。

| 来源 | commit SHA |
|---|---|
| PR8088 | `698d142468ab15374bc22bd5247759e5b7352006` |
| PR8139 | `4942b2afc70be7291df28bce594eeef21b71e336` |
| PR8170 | `dba770d56a0ee0ae5bad1b3dc5e11614d323163c` |
| PR8255 | `aa48a8bbd80c6ac956dbcbaba651f417b3cf78e3` |
| PR8177 | `8f2cbb53d9d1c40e1865b07ac460571f7d4c48de` |
| PR8188 | `888c6057333e011b8cfeb81a867f445d229d1432` |

最大SQLite迁移226；本批不需要新schema。[四态](../../../docs/upstream-sync/evidence-0.2.18/files.tsv) / [逐来源决定](../../../docs/upstream-sync/evidence-0.2.18/decisions.tsv)。
