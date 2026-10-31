# 来源基线（冻结）

冻结日期：2026-10-24。后续实施不得编辑本文件；实际实施起点和新增来源另记 verification 或独立补充记录。

| 来源 | 解引用 commit SHA |
|---|---|
| fork HEAD / VERSION 1.1.19 | `c36a19b7a1a132e690eaf9ce0d878b301ffffcb4` |
| 官方 v0.2.15 | `f2669c8cf62555cd92389b3f55920e9e6e7c6ff2` |
| 官方 v0.2.16 | `01842576b3d5c1e2545b5a99ad555d36e28ed306` |
| 官方 v0.2.17 | `060d0160417782cd962a0c7f460f6cd4565d4916` |
| 官方 v0.2.18 / 来源终态上限 | `ac1b8f1271be0d674fa6f074403c1410f30a3a81` |
| 本地已抓取 upstream/main，仅记录、不扩大范围 | `5c185c5c5d4f5e409e2139806e0964ae944e03e8` |

| 功能 | PR | 完整 merge SHA | first parent SHA |
|---|---|---|---|
| F01 使用日志 TPS | [#7911](https://github.com/Wei-Shaw/sub2api/pull/7911) | `745ff1d3ca79a7f39520f627f6aa44d778a11580` | `8e2c4fad321fe8ce86d577dd7561c371bf602647` |
| F02 筛选记忆 | [#8142](https://github.com/Wei-Shaw/sub2api/pull/8142) | `d744c22016f15f37191d90fe8850a66ccb730399` | `4942b2afc70be7291df28bce594eeef21b71e336` |
| F03 运维账号名 | [#7933](https://github.com/Wei-Shaw/sub2api/pull/7933) | `59fb22791ea7e5987312750cc2ce8bbd91ec8a65` | `20b0be5b05c22730db78747fe49ce7e898040d0c` |
| F04 规则 scope | [#7828](https://github.com/Wei-Shaw/sub2api/pull/7828) | `49dc6e236380a7198e0658054b0d7c595b597387` | `53c21fed252cec522035cfc2d4dd890a8c47140b` |
| F05 搜索模拟 | [#8080](https://github.com/Wei-Shaw/sub2api/pull/8080) | `b14f316ca2def2fbf379a7413c3f4309cc70acdd` | `43157c478ec29cb765043e2f0cba4f69728dd2f2` |
| F06 输出下限 | [#8094](https://github.com/Wei-Shaw/sub2api/pull/8094) | `58b5e858c82dcf14a637762ee1e6c37ec9f86aad` | `857a76d4a0e7b2389926c340555a2f8f5b24a0ae` |

候选取 `git diff <merge>^1 <merge>`；微提交不单独作为移植单元。F01 仅选日志展示，不选同 PR 的运维分位数。其它来源只选六项及直接依赖；已有平台画像、缺失网关文件和 PG 查询不作为隐含前置。

工作树开始时有用户正在进行的 OpenSpec 归档、主规格及索引变更；产品文件未见本次相关未提交修改。逐文件四态使用独立临时 index 读取上表 fork SHA，不以用户工作树替代冻结基线。已有主规格按当前归档结果复核；未来归档合并时重新核对同名需求。

运行约束：Go 1.27.3，VERSION 1.1.19，最大 SQLite 迁移 226；本次规格不要求新增迁移、依赖或版本变更。

冻结证据：[完整候选与 patch 哈希](../../../docs/upstream-sync/evidence-0.2.18-personal-features/candidates.tsv)、[四态](../../../docs/upstream-sync/evidence-0.2.18-personal-features/files.tsv)、[新增调用词法清单](../../../docs/upstream-sync/evidence-0.2.18-personal-features/added-call-inventory.tsv)。基线通过不等于功能已经实施。
