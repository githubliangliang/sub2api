# v0.2.18 评估证据索引

基线/时间：873962ce7，VERSION1.1.18，2026-10-23。正文为[PORTING-0.2.18](../PORTING-0.2.18.md)。这些是评估证据，不是实施验收。

| 文件 | 用途 |
|---|---|
| baseline.json / releases.json / fetch-final.txt | 冻结fork/tag/main SHA、发布说明与最终串行fetch确认 |
| commits.tsv / candidates.tsv | 56非merge及35PR+8直接提交完整候选 |
| files.tsv / candidate-stats.tsv | 312条文件的先反向后正向四态及汇总 |
| decisions.tsv / clusters.json | 全候选分档及16簇选定行为；混合PR不得整取 |
| added-call-inventory.tsv / hunks.tsv | 新增调用词法清单、对应本地声明、同PR新增声明和hunk上下文 |
| base-symbols.txt / call-review.md | 关键基座实测命中/缺失与人工消费者复核 |
| supplemental-files.tsv / supplemental-bases.md | #5925完整旧PR四态及小helper前置量测；不计入本轮总数 |
| source-test-inventory.tsv / integration-build-tags.txt | 来源测试在本地的存在性和集成测试编译条件 |
| baseline-tests.md / backend-*.txt / frontend-neighbor-tests.txt | 本轮实际基线命令、测试目录与执行输出 |
| service-probe.go.txt / admin-probe.go.txt / baseline-probes*.txt | 临时overlay探针源码和输出；PASS表示旧缺陷存在，初次错误假设保留 |
| frontend-audit*.json / frontend-production-audit.json / *audit-stderr.txt | pnpm当前审计快照和诊断 |
| backend-vulnerability-scan.json / backend-vulnerability-findings.tsv / backend-vulnerability-stderr.txt | Go package扫描原始JSON流、只统计finding的摘要与诊断 |
| dependency-review.md | 依赖目标、暴露面证据边界和未处理记录 |
| openspec-tier1.txt / openspec-tier2.txt / document-validation.md | 官方OpenSpec严格校验、链接/计数/范围一致性检查 |

复现文件态：用baseline.json中的fork SHA建立独立临时GIT_INDEX_FILE，候选patch统一来自`git diff parent1 sha`；对全部候选文件先`git apply --cached --reverse --check`，再正向check。反向成功为ALREADY；否则正向成功为CLEAN；双向失败且冻结fork无文件为NOBASE，其余CONFLICT。直接提交使用其唯一父；整PR使用merge父1，禁止拆微提交评估。

源码/迁移/依赖均未更改。历史PORTING与OpenSpec冻结证据没有重写。
