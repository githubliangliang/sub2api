# Source Baseline（冻结）

冻结日期2026-11-18。实施不得改写本文件；真实起点另填verification。

| 项目 | 冻结值 |
|---|---|
| fork HEAD / tree | `7a0b78fe511f17b0f7eb9d26fcf5e4ff083312c7` / `3882c3ce52fb24f6f8d89d782ffc7b48bb1e65e7` |
| VERSION / Go / SQLite / auth snapshot | `1.1.25 / 1.27.3 / 229 / 21` |
| private index tree | `3882c3ce52fb24f6f8d89d782ffc7b48bb1e65e7` |
| v0.2.21 commit | `10ba8f9b4dc2e6023d4a9879ab77b21841babd10` |
| v0.2.22 commit / tag object | `b0fa9f3e6eeab1968a53e1289b69daa4bb40ca2b` / `b11573c9d55d39b6780e1e9f6e193ba2e6ca2321` |
| v0.2.23 commit / tag object | `5af9afb588c4b014b3df9db7d4073af1f8ae01b6` / `ce19a640eab5c4b5ef1827c7c1fcfce49569411a` |
| upstream/main（观察，不作来源） | `2af3f1e15c98c893a89255a037f551046883b8c0` |
| 非merge / 候选 / 净diff | 128 / 78 PR + 4直接提交 / 399文件 +25947 -2093 |
| 候选×文件四态 | 524：ALREADY0 / CLEAN269 / CONFLICT199 / NOBASE56 |


| PR | merge | first parent | patch SHA256 |
|---|---|---|---|
| [#8669](https://github.com/Wei-Shaw/sub2api/pull/8669) | 7e4c614bcfb7fc93c955f6da3227c4c0cdfab8cb | 3e8e3254d1a5c9f4e2bf66d05ca95ba57bdf6297 | 878aece0675c0ea62af7a2ec07f1641854cf20db26ad50eae97c4332a78fb294 |
| [#8682](https://github.com/Wei-Shaw/sub2api/pull/8682) | 89fce5446d0a449c8d0ca1dd650245747bf3cee5 | bd3a878ef49baa32f741041755e7b9b6041ddc1e | ac71481c4994d74d87bef4f6514065e3091b460f666cb5095aa57e637556d5dd |
| [#8684](https://github.com/Wei-Shaw/sub2api/pull/8684) | 910d4d86b2c0d389f3201415123c386c98c6be2c | 89fce5446d0a449c8d0ca1dd650245747bf3cee5 | 2e1eabfba3b036e8b49892eb931353bc6558e552149f20cb2e941bb6f0bfec20 |
| [#8698](https://github.com/Wei-Shaw/sub2api/pull/8698) | 2df3dbcfbb2973927dfe585bdaaa092e09fee7f7 | fde504e343b2897c738c1fea7014cac7fc731371 | aef7f0a4c4e3d302da3b1035b58a03b37f3905930d251a375b86cd28b900135b |
| [#8703](https://github.com/Wei-Shaw/sub2api/pull/8703) | 0b473ad60b442828dfbecde8a9653789b862b492 | 2df3dbcfbb2973927dfe585bdaaa092e09fee7f7 | bb86344386cf13f67296f932f4d26e248fcc6f7b89b8bcc1c4d4280c71434c28 |
| [#8723](https://github.com/Wei-Shaw/sub2api/pull/8723) | 90eb1744d42adeb7c78896ab094627dcd676f7c3 | 0b473ad60b442828dfbecde8a9653789b862b492 | 18abc68ef5532c87858a1e875398c37be144833c0746dbda90fd8428b3296c13 |
| [#8726](https://github.com/Wei-Shaw/sub2api/pull/8726) | 51e76a46c181ae93697f3a06702bc74ab4a45931 | 90eb1744d42adeb7c78896ab094627dcd676f7c3 | 16a4ffe3fcd41f6ac589b299e7876f1e86de82718b0aad77357a1a400afc1358 |
| [#8736](https://github.com/Wei-Shaw/sub2api/pull/8736) | 472ea5752598d7e4581ba195aaa72989c165640b | eb25511c7166b3e85e911ebb969bb1c9fffba215 | b36becca0b67ab10af30ab137a50cbb641a1fb626948fe35f7c2e6d6fcd99a44 |
| [#8742](https://github.com/Wei-Shaw/sub2api/pull/8742) | 6fd905f30d890c18152c17050cba4232658a6c22 | e54a186548a0ca1aac6211a7d8c0ed6f8b939d8f | 9e4647a62e5fb9d5b866b42c275efcb7fba0eab71b152944193bcb6588da2797 |
| [#8757](https://github.com/Wei-Shaw/sub2api/pull/8757) | 3b5310c5d54ff9fcb047e0b0bb59b19691873b0c | e16a349ebb102caef1ec3bf5804f0c7058251636 | 4e3fa267883ad670e3c95e09ce27f82ef141d035c57e486d24fb0b5206ab7fed |
| [#8759](https://github.com/Wei-Shaw/sub2api/pull/8759) | 202740441d6352c2218b1c2abc8479a94539d1a3 | 3b5310c5d54ff9fcb047e0b0bb59b19691873b0c | 34700a31abbbf83611d0c51d311c263fd7bc3a0c28cf3df8a0878dbf8887d4fc |
| [#8761](https://github.com/Wei-Shaw/sub2api/pull/8761) | 4763569af25273c38476dfcd9f5dbf1965428751 | 202740441d6352c2218b1c2abc8479a94539d1a3 | 61e17b52395f4cc12d741a89f00b645eb691ca16b2e14e241e33630e7980a1dc |
| [#8769](https://github.com/Wei-Shaw/sub2api/pull/8769) | d31a87a4e2024de151ac517ef7eeb16cb30085c0 | 56848016c5b57573e92cf3d5a0f9dcc2caad663e | 1cba5c1e875a1c28294b38be3f8dee4e0d97a99642e32421e692dfb28d849ad3 |
| [#8775](https://github.com/Wei-Shaw/sub2api/pull/8775) | c7df3b693f075945da9ba2d3f03bb12e5a541bb3 | d31a87a4e2024de151ac517ef7eeb16cb30085c0 | 89bcc8e7063f4b865f6255c3f7e53bc52b1bc70d7d5c7ae40bad4be39efc386e |
| [#8786](https://github.com/Wei-Shaw/sub2api/pull/8786) | 1d585c94fd793e98d8e531b5eb1634405e323fec | 4cebc098c32940939dc1e850f23d757cfc742d24 | fffa3e22b18f77974796eba78bfafbd17658e9c2e68b29bbc2ef87cc4638e3c4 |
| [#8788](https://github.com/Wei-Shaw/sub2api/pull/8788) | 63751ea4de56e5204f2619d66e202efe441b7269 | 1d585c94fd793e98d8e531b5eb1634405e323fec | 018771da6f04f879d57cd46a57fee3def3ed139310ce77b1cbbfaad8368a5b8d |
| [#8808](https://github.com/Wei-Shaw/sub2api/pull/8808) | 9750f8e039a74bdb76cead0d1b23624db3f49ddb | 568db32f22c1e7b315367a2f5219565e0a9b776a | 38a56a5aa9509e44b3c52e845988b4c55ac5053702af23e8e115e5e243543822 |
| [#8812](https://github.com/Wei-Shaw/sub2api/pull/8812) | 8b13851ac830885eceaf509273418dc47eb90a62 | 9cd1a1eb21e84707685e1998dcb5a3d1ceafae49 | 6759dde5a065ddecc057209829f1fa0e6f4ad2a6a0805f931bca72ad4caca60c |
| [#8814](https://github.com/Wei-Shaw/sub2api/pull/8814) | 992ac633082d33e8dd0ed519014c98018ed7258e | 101c2462ddb748e3ceb5e5a2e999fce9cf92467c | 108b21ef88fd440bf8e79761e0a0dc521d8886f23a98bcd349d3928853251ee5 |
| [#8837](https://github.com/Wei-Shaw/sub2api/pull/8837) | 6b7cbc63bed7defa0a5b83c21f5a315a1ab02beb | 46b879265896850f0ee3cb4971c5bdf3dde7df2f | 4f993f547097ec81af3146e4e495dcd3d5a0faebe2922ffaec3d945251bb30d4 |
| [#8840](https://github.com/Wei-Shaw/sub2api/pull/8840) | d70676756fd3a35b854d41aa46c07935bbef283f | 6b7cbc63bed7defa0a5b83c21f5a315a1ab02beb | c41ce148442a6518b967f2b0215dab471ee008492d94aa7f646222d689e98e92 |


整PR first parent→merge为来源；最终tag状态优先于中间提交和发布摘要。全部文件先reverse后forward，未应用产品补丁。[完整来源与评估证据](../../../docs/upstream-sync/evidence-0.2.23/MANIFEST.md)。
