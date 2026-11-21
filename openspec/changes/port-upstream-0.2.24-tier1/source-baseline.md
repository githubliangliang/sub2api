# Source Baseline（冻结，实施不得回写）

冻结日期2026-11-21；源码范围仅v0.2.23..v0.2.24。真实实施起点填verification。

| 项目 | 冻结值 |
|---|---|
| fork HEAD / tree | `d2488b1105f091b6fce4445395fec1835306fcf8` / `17b1b71f61137510e42feeaa3fcf2a1109cb4ce6` |
| VERSION / Go / SQLite / auth snapshot | `1.1.26 / 1.27.4 / 230 / 21` |
| 私有 index tree | `17b1b71f61137510e42feeaa3fcf2a1109cb4ce6` |
| v0.2.23 commit / tag object | `5af9afb588c4b014b3df9db7d4073af1f8ae01b6` / `ce19a640eab5c4b5ef1827c7c1fcfce49569411a` |
| v0.2.24 commit / tag object | `5c2982a81b46e2142d0143d2f83c1ce9c4da1d24` / `7515db6dc56c2a6120dfb7abed70ef089ebc1457` |
| upstream/main（观察） | `328db26f450edbdf9f8d4344a3ea903ac6396842` |
| 非merge / 完整候选 | 65 / 39 PR + 1普通直接提交 + 1无PR合并簇 |
| 净diff / 候选×文件态 | 190文件 +11853 -610 / 217条 |
| ALREADY / CLEAN / CONFLICT / NOBASE | 0 / 135 / 62 / 20 |

| 候选 | merge/direct | first parent | 完整patch SHA256 |
|---|---|---|---|
| [#8982](https://github.com/Wei-Shaw/sub2api/pull/8982) | `16b53b60a8c9b83a6e4e93c8157b9b7756dc2c30` | `2af3f1e15c98c893a89255a037f551046883b8c0` | `e5103717174a8d091befae8ac83999fde32bd76f7d39af249f9c78d2c3ddd005` |
| [#8911](https://github.com/Wei-Shaw/sub2api/pull/8911) | `69237d2ebd166674739b8ee5e6a343d0f3f635b4` | `2f49048cb80a8883df597bcfc4c66787d9662ed2` | `495e69775a67708a234213151ca419e74d7b9c648ebed1529c9007630c38aa04` |
| [#8923](https://github.com/Wei-Shaw/sub2api/pull/8923) | `16b67d9afd0f268ac46ef42a0c7dca15cd36d407` | `a912d13e62514cd4d6233126da0cd77bef901d2b` | `57b110aab82570313e7f4dc426a3eb5973abaa9aebf32ba8fce1b8d2547eac35` |
| [#8929](https://github.com/Wei-Shaw/sub2api/pull/8929) | `a64189740286f936dccccae2184e75674d243d22` | `16b67d9afd0f268ac46ef42a0c7dca15cd36d407` | `fb12a7d1ddcff7bb47e14386744651624775384430917860002e19d142093856` |
| [#8931](https://github.com/Wei-Shaw/sub2api/pull/8931) | `fe1f2b6d76cc1c535ae7bcebdc0e694aac9cb742` | `a64189740286f936dccccae2184e75674d243d22` | `84238f5ae50022b5ce539490121c71e393e42950d1a2ac46ae24d0d1c4917981` |
| [#8946](https://github.com/Wei-Shaw/sub2api/pull/8946) | `55898028ad60b51b56dbd99734e2a372480ae846` | `50dc3141e7dea62f2ed3b92998c1ac4367172f1a` | `ef84d57a5bd92ef3cd8d9c2f348605093778056d347f6a5838f01b8dd8638571` |
| [#8962](https://github.com/Wei-Shaw/sub2api/pull/8962) | `c03fc529b827c46ce6ba1ba04a84f4088d5f0f3b` | `55898028ad60b51b56dbd99734e2a372480ae846` | `bb635c96473b7d46cda38807d3a940e26a763ed1015e00b6b34b71f2e6172c76` |
| [#8963](https://github.com/Wei-Shaw/sub2api/pull/8963) | `ba5ddded9c6fa251d711d9a357881fcd14951c60` | `c03fc529b827c46ce6ba1ba04a84f4088d5f0f3b` | `2c3919b62af7e81ddb8696ca4a1fe731a4f71101a61115818eaf93bf4ce50ddf` |
| [#8968](https://github.com/Wei-Shaw/sub2api/pull/8968) | `58659ec579ad6a84495bf3e96d71c2f1d7dfce63` | `ba5ddded9c6fa251d711d9a357881fcd14951c60` | `1b83d2d4222d08be053ffac483f33f0323acd5f0e7f558269992fb7b63ada1e0` |
| [#8886](https://github.com/Wei-Shaw/sub2api/pull/8886) | `62cebadf54b5b208d6dd9b93f1929d25263e67e8` | `eb9143956f1e74458c6ca56af069c3e8604c3bd0` | `ae233e61d692b26a2fd02f8f19ba57e77ecbf12c7a9751fd14c4dc01cd75af4f` |

最终tag优先于中间提交。全部候选先reverse再forward；来源完整性与评估证据见[MANIFEST](../../../docs/upstream-sync/evidence-0.2.24/MANIFEST.md)。不把来源的PG/缺平台接口当成本地实现。
