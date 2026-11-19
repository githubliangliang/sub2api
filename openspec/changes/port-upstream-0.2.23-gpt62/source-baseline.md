# 冻结来源

本文件记录实施入口事实，后续验收不回写来源。

- 用户选择日期为 2026-11-18，实施验收跨至 2026-11-19。
- fork 起点为 `38990cfa716e337740ae8aa3c316940368a3f9ce`，入口工作树干净，VERSION 1.1.26、Go 1.27.4、SQLite 229、auth snapshot 21。
- 使用已保存的 v0.2.23 [PR #8906 完整 patch](../../../docs/upstream-sync/evidence-0.2.23/patches/PR8906.patch)，SHA-256 `146eaea850df80d133fcc6caa42bfd0d9a065e2710bbbd1bc117643300298c4e`。
- feature commit `6107cf17487fe5cb0bd179457ede6b1fe8f7271e`，随后套餐修复 `2d20707f16dcd5d2d692a9535370dbbdb9b8b6f0`，见原 [commits 清单](../../../docs/upstream-sync/evidence-0.2.23/commits.tsv)。套餐修复不属于本次用户选定的价格/模型范围。
- Codex 描述来自该 patch 保存的 2026-11-17 models response，client 0.169.0。Sol 原文 SHA-256 `62469f05d46e7108e573a4746490edc6acd8931fb9b259e326c33e61f0c9cf49`，Astra 原文 `d5e1e3e3dcc06e9648bc5eefc3b3d09bd5412046d60e1be3b3f3a45ec3f6b16c`。
- 官方网站读取遇到 web 工具 HTTP 502，直接读取官方 pricing URL 为 HTTP 403。单价依据冻结的上游价格卡，官方在线价格未独立核验；该限制在验收及最终说明中保留。
