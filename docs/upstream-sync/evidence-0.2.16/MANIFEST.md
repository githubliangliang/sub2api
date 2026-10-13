# v0.2.16 评估证据

冻结基线、release/tag来源与完整候选四态见baseline.json/releases.json/commits.tsv/candidates.tsv/files.tsv。所有正文文件均在评估时生成；后续实施不得改写本目录为新的基线。

- decisions.tsv：全部93候选的分档（混合PR多行）。
- supplemental-sources.tsv / supplemental-files.tsv：3个旧来源及额外四态，不混入本轮149条/575行统计。
- added-call-inventory.tsv、hunk-base-inventory.tsv、base-symbols.txt、frontend-imports.tsv、call-review.md：词法库存与人工消费者复核。
- source-test-inventory.tsv：197条来源测试路径、存在性、首行标签及Go测试名。
- baseline-tests.md及各日志：通过、未运行、失败分开；overlay源码保存为文本，不会被Go自动编译。
- dependency-review.md / frontend-audit.json：当前审计，替代历史advisory数量；不是已利用漏洞声明。
- leftover-probes.txt：历史未合项的当前符号/调用点/例外日期复核。
- document-validation.md：评估文档与两份OpenSpec一致性检查。

产品代码、VERSION、历史迁移、依赖lockfile均未变；没有合并上游、创建release或部署。
