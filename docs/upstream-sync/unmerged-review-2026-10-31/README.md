# 历史未合项复查证据（2026-10-31）

基于v1.1.21 / `e44324231ec98b5176fe04c6ddb3f3552a352889`。本目录记录评估与缺口复现，不是产品实现或发布验收。[建议报告](../UNMERGED-REVIEW-2026-10-31.md)、[未实施OpenSpec](../../../openspec/changes/port-upstream-historical-client-compatibility/README.md)。

| 证据 | 含义 |
|---|---|
| [baseline.json](./baseline.json) | 冻结HEAD/VERSION/Go/SQLite227；既有7份活动change均完成，共276项 |
| [source-refs.json](./source-refs.json) | 旧tag与完整行为来源SHA、父提交、补丁哈希 |
| [source-file-states.tsv](./source-file-states.tsv)、[summary](./source-state-summary.json)、[raw](./source-file-checks.json) | 固定HEAD的独立index，所有reverse完成后再forward；25条 A1/C14/X9/N1，无应用 |
| [baseline-test-list.log](./baseline-test-list.log) | 当前unit实际列出的47个入口 |
| [baseline-targeted.jsonl](./baseline-targeted.jsonl) | 正常定向测试3包、47个入口/83个含子用例通过，0失败/跳过 |
| [apicompat-probe.log](./apicompat-probe.log)、[service-probe.log](./service-probe.log)、[handler-probe.log](./handler-probe.log) | 9个顶层隔离探针按缺口预期失败；合法JSON控制用例通过 |
| [转换探针](./apicompat-probe.go.txt)、[服务探针](./service-probe.go.txt)、[handler探针](./handler-probe.go.txt) | overlay源；实际测试不加入backend产品/测试文件 |
| [openspec-strict.log](./openspec-strict.log)、[checks.json](./checks.json) | 新方案严格校验及实际检查结果 |
| [document-validation.json](./document-validation.json) | 本批相对链接、规格/任务/空验收槽与产品范围核对 |

正常基线使用`go test -tags=unit -list`与`go test -tags=unit -run <pattern> -count=1 -json`；pattern和实际命令保存在checks.json。没有执行PG-only integration，也没有把本次正常定向回归称为全量门禁。

复现：在冻结HEAD运行`rtk proxy python3 docs/upstream-sync/unmerged-review-2026-10-31/reproduce-probes.py`。脚本在临时目录生成overlay和日志，注入三个包的临时测试。当前Go测试预期退出1；输出显示实际gateway成功完成非法工具、别名/孤儿defer和缺失序号。后续实现时需将这些基线探针适配成正式回归，不能把原始接口快照覆盖为修复后版本。

在线GitHub PR打开/检索均返回HTTP502；当前来源由可读取git对象与已冻结整PR补丁支持。新模型可用性、实时价格和v0.2.19之后的发布未在本次确认。
