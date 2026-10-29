# v0.2.19 评估证据索引

基线：fork 49939fb2cf96633688548ca39204d96558e59d79，VERSION1.1.20，Go1.27.3，最大SQLite迁移227；来源v0.2.18→v0.2.19。当前为评估，没有业务代码移植。

## 原始来源（保留）

- [baseline.json](./baseline.json)、[releases.json](./releases.json)：原始采集记录，早期status/文档提交转换保持历史。
- [commits.tsv](./commits.tsv)、[candidates.tsv](./candidates.tsv)：50非merge、20PR+3直接提交；父SHA、完整diff哈希。
- [files.tsv](./files.tsv)、[candidate-stats.tsv](./candidate-stats.tsv)、[hunks.tsv](./hunks.tsv)、patches/：276文件先反后正四态及全部整PR补丁。
- [added-call-inventory.tsv](./added-call-inventory.tsv)、[source-test-inventory.tsv](./source-test-inventory.tsv)：宽松新调用token及98个source测试文件。
- [business-file-sha256.json](./business-file-sha256.json)：3732产品文件哈希。
- 原始backend/frontend输出、probe源码、HTML PoC、MVS、production audit和Go advisory保持原文件。

## 本轮复查与选择

- [review-baseline.json](./review-baseline.json)、[review-intake.py](./review-intake.py)、[review-file-checks.tsv](./review-file-checks.tsv)：重验全部patch/覆盖/四态/产品哈希，不写真实index。
- [review-original-file-hashes.json](./review-original-file-hashes.json)：复查前70个原始文件哈希；后续只新增review和说明。
- [review-online.json](./review-online.json)：GitHub latest与Go官方advisory复核；post-tag upstream/main不取。
- [classification.tsv](./classification.tsv)：23候选全覆盖，档位/选定子项独立。
- [call-review.md](./call-review.md)、[review-base-symbols.txt](./review-base-symbols.txt)、[review-selected-call-inventory.tsv](./review-selected-call-inventory.tsv)：328产品调用token与真正缺前置/消费者审查。
- [review-integration-build-tags.tsv](./review-integration-build-tags.tsv)：77个fork/source integration首行，不能按名称推断SQLite可跑。
- [feature-selection.md](./feature-selection.md)：用户需要Terra，仍第三档，当前仅提案。

## 已运行的基线检查

每个review-检查旁的同名JSON记录完整argv/cwd/UTC开始时间/耗时/退出码，stderr单独保存，run-check.py可复跑。

- [review-test-list.txt](./review-test-list.txt)：10包实际测试目录编译。
- [review-targeted.txt](./review-targeted.txt)：10包既有定向回归。
- [review-baseline-probes.txt](./review-baseline-probes.txt)：12个旧行为probe，含已满足图片定价负对照；PASS代表确认现状。
- [review-claude-js-oracle.txt](./review-claude-js-oracle.txt)：独立Node/JS UTF-16与UTF-8哈希核对三Unicode样本及ASCII负对照。
- [review-grok-audit-probes.txt](./review-grok-audit-probes.txt)：额外2个probe覆盖原生JSON/SSE与两个buffered转换。
- [review-frontend.txt](./review-frontend.txt)：4文件40个既有前端测试。
- [review-terra-neighbors.txt](./review-terra-neighbors.txt)、[review-terra-frontend.txt](./review-terra-frontend.txt)：既有模型/价卡/协议与UseKeyModal回归。
- [x-net-minimum-mvs.json](./x-net-minimum-mvs.json)、[html-complexity-current.txt](./html-complexity-current.txt)、[html-complexity-fixed-version.txt](./html-complexity-fixed-version.txt)、[review-dependency-feasibility.txt](./review-dependency-feasibility.txt)：临时modfile可行性，不是repo升级完成。
- [frontend-production-audit.txt](./frontend-production-audit.txt)：17条production advisory，high两条xlsx；未跑当前全量dev audit。
- [document-validation.md](./document-validation.md)：提案严格校验、链接/覆盖/空实施槽/保护哈希检查。

[主评估](../PORTING-0.2.19.md)记录patch sites、档位与旧遗留；三份OpenSpec只定义行为与未来实施验收。原始/复查输出均不能填入尚空的实施verification。

