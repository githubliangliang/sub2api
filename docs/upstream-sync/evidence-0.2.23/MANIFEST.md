# v0.2.23 评估证据

本目录是2026-11-18 intake，不是产品实施验收。基线7a0b78fe511f17b0f7eb9d26fcf5e4ff083312c7 / VERSION1.1.25 / Go1.27.3 / SQLite229 / auth21。

- [冻结事实](baseline.json)、[初始事实](initial-facts.json)、[release元数据](releases.json)、[逐release统计](per-release-stats.json)
- [全部候选](candidates.tsv)、[成员提交](commits.tsv)、[逐文件状态](files.tsv)、[原补丁](patches/)、[新增调用扫描](added-call-inventory.tsv)、[hunk](hunks.tsv)
- [评审行为/前置](review.json)、[完整分档](classification.json)、[发布摘要与最终源码差异](release-source-differences.json)
- [测试发现](baseline-test-list.log)、[backend定向](baseline-targeted.json)、[frontend89测试](baseline-frontend.json)、[来源测试](source-test-inventory.tsv)、[本地integration标签](integration-build-tags.json)
- [Go探针摘要](probe-summary.json)、[6包测试源码](probes/)、[Go overlay复现](reproduce-probes.py)
- [前端9失败/11通过](frontend-probe.log)、[复现](reproduce-frontend-probes.py)、[初始位置/runner问题](frontend-probe-initial-apply.json)
- [性能源码来源](performance-source.json)、[语义与转义检查](performance-equivalence.log)、[函数benchmark](performance-benchmark.log)、[复现](reproduce-performance.py)
- [Go1.27.4 embed](go1274-embed-feasibility.json)、[三包unit](go1274-unit-feasibility.json)、[首次网络失败](go1274-embed-initial-network.log)
- [生产audit与依赖范围](dependency-summary.json)、[raw audit](frontend-production-audit.log)、[tag后main](post-release-commits.txt)
- [原始保护哈希](protected-files.json)、[最终范围校验](document-validation.json)、[本目录SHA256](SHA256SUMS)

Go顶层26失败/20通过为旧产品面对目标断言的结果；源码SSE predicate另有1失败，不属于本地产品缺陷计数。临时前端仅修改测试，保存对象/输入断言9失败、既有11通过；没有浏览器验收。性能使用单P函数夹具，非1GiB生产压力。Go1.27.4只读编译不修改当前1.27.3声明，也不完成未来工具链维护。

服务Config夹具错误、前端测试位置/pnpm问题和下载网络问题完整保留；最终期望失败没有被产品patch修成PASS。Go overlays和临时前端复制不会修改真实源码/index。

可重跑collect-intake.py的前提是HEAD仍为冻结基线；它只检查同一私有index。重跑各reproduce脚本会覆盖本目录对应测量记录，应先另存记录。两个OpenSpec的实际实施checkbox/verification仍为空。
