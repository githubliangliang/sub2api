# 2026-11-25 新增合并复核证据

对应[复核结论](../RECHECK-2026-11-25.md)。API 响应时间为 2026-11-25 14:16 UTC；结论记录于 14:20 UTC。没有新增产品候选，本目录不是新的实施验收。

| 证据 | 内容与边界 |
|---|---|
| [facts.json](./facts.json) | 实际 fork/tag/main、版本/迁移、fetch 结果、新增提交/合并均为0，以及网页/gh 工具限制。 |
| [main.json](./main.json) / [响应头](./main.json.headers) | 官方 main 分支 API，SHA 与 Git 一致。 |
| [latest-release.json](./latest-release.json) / [响应头](./latest-release.json.headers) | 最新正式版 v0.2.25 原文。 |
| [releases.json](./releases.json) / [响应头](./releases.json.headers) | 最近10个 Release，排除仅凭旧 tag 推断当前最新。 |
| [merged-since.json](./merged-since.json) / [响应头](./merged-since.json.headers) | 官方搜索 `repo:Wei-Shaw/sub2api is:pr is:merged merged:>2026-11-24T03:21:26Z`，total_count0 / incomplete_results=false。 |
| [pull-summary.json](./pull-summary.json) / [原始响应 gzip](./closed-pulls.json.gz) / [响应头](./closed-pulls.json.headers) | 最近更新100个关闭 PR；只作佐证，完整新合并判断另有搜索和 main SHA。gzip 解压后 SHA256 可核对。 |
| [git-checks.json](./git-checks.json) | 实际命令、退出码、原始 stdout/stderr；主线空增量、tag后 VERSION 唯一 diff、4个 CLA 提交及 cla.json 唯一文件。 |
| [active-changes.json](./active-changes.json) | 当前15份活动变更、456个完成任务、0个未完成任务；没有新建 OpenSpec。 |
| [implementation-hash-check.json](./implementation-hash-check.json) | 当前15份实现源码/测试/模块与既有冻结哈希一致；不是重新执行测试。 |
| [protected-summary.json](./protected-summary.json) / [启动全量哈希 gzip](./baseline-hashes.json.gz) | 启动HEAD/index、8,928份tracked和初始未跟踪Fast诊断SHA256。 |
| [collection-note.json](./collection-note.json) | 采集脚本的154a编号解析失败、原因与最小控制；没有产品修复。 |
| [validation.json](./validation.json) | 最终链接、HEAD/index/tag、所有既有文件与原诊断保护，以及新增文件边界。 |
| [artifacts.json](./artifacts.json) | 本目录所有其它最终证据的 SHA256。 |

未重跑产品测试、构建或安全扫描；没有原始失败被改写为通过。原有评估/实施/发布证据保持其历史内容。
