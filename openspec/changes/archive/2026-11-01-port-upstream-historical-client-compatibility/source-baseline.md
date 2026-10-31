# Source Baseline（冻结）

冻结日期：2026-10-31。后续实施起点、产品哈希和实现提交只写verification，不回填本文件。

| 项目 | 冻结值 |
|---|---|
| fork HEAD/tag | `e44324231ec98b5176fe04c6ddb3f3552a352889` / v1.1.21 |
| VERSION / Go / 最大迁移 | 1.1.21 / 1.27.3 / SQLite227 |
| 历史来源tag v0.1.180 | `c40edb4070a9274e8c23f161b4ed552051b14698` |
| 历史来源tag v0.2.5 | `86f93c28ee34cc74b629dafb748bd5ac5ca8c5ea` |
| 当前已评估终点v0.2.19 | `dece26153621959ccd7635814f4b2d08304e681a` |
| 本批源文件态 | 25条 A1/C14/X9/N1，先完成全体reverse再forward，只做独立index检查 |

| 编号 | 冻结来源与完整行为取法 |
|---|---|
| C01 | `e2d9ce0cadadb68c35b08ca434258a4d5a9de0c9`与`fbc9ee626d7298ddc1ba96c95349625f54737543`配套终态；完整diff从`e943f817b118936eb776f20a3a4040a4123dcfff`到后者，5文件 |
| C02 | `31d5b67baa4f21f2203a40b66772ef868e0f4fed`；完整diff从`d45135d87df16d48637f04ccd245727bc955ba54`到该提交，6文件 |
| C03 | `7a09a2eaf7386b299bfa78dfb21df2063f80db9a`；完整diff从`d45135d87df16d48637f04ccd245727bc955ba54`到该提交，6文件 |
| C04 | PR7129 merge `25c21aeddac89eb1b1abd90ae7d90fb5ec14c6e8`；完整diff从第一父`2d9af8592fc97db2ad4b237b32757de8f3af544d`到merge，8文件 |

来源元数据与哈希：[source-refs](../../../docs/upstream-sync/unmerged-review-2026-10-31/source-refs.json)。[原始文件态](../../../docs/upstream-sync/unmerged-review-2026-10-31/source-file-states.tsv)、[当前基线](../../../docs/upstream-sync/unmerged-review-2026-10-31/baseline.json)。不继承原来源版本的价格、默认值、上游main状态或未实施判断为当前事实；未纳入tag后新发布。

25条是候选文件记录，包含重复文件与测试，不能当实际修改文件数。当前新增调用的缺口和真实消费者在[评估](../../../docs/upstream-sync/UNMERGED-REVIEW-2026-10-31.md#patch-sites-与前置量测)量测，文件CLEAN不证明新增测试能直接编译。
