# v1.1.25 发布准备

日期：2026-11-12（Asia/Shanghai）。用户授权“打tag v1.1.25，详细描述。push”。本文件为提交前检查快照，实际发布提交与远端运行结果在推送后另行核实。

## 基线与范围

- 产品基线 `83729b13035c6ad7a26f7dec797ac4a2c0af8e17`，起始工作树干净。
- 上一 tag v1.1.24：对象 `0be13fbc74025302a72dd355685f459e097a4883`，commit `4f4c80c3ce48ce4afbc9344536b9cf44ed3fd898`。远端 origin/main 在 `63c764704761073e69ff25ba41d64eeab5793cd0`，是本地 main 祖先；远端无 v1.1.25。
- 两个实际实现增量：`63c764704` 第一档 F01–F11（39/39），`83729b130` 第二档 S01–S12（42/42）。精确 SHA 及迁移/依赖差异见 [release-facts](release-facts.json)。
- 相对上一 tag，SQLite 最大迁移 228→229，仅新增 API-key cost 覆盖索引。Go1.27.3、auth snapshot21、依赖、Ent/Wire 及 64MiB body / 40MiB SSE 行 / 120s idle 默认值不变。
- 本次发布准备只修改既有产品 VERSION 为 1.1.25，增加说明/证据及 AGENTS 顶部发布索引。历史来源、PORTING、实施证据、OpenSpec、旧迁移和旧 AGENTS 记录原文保留。

## 既有实施验收与远端状态

复用已冻结的 [第一档验收](../../../openspec/changes/port-upstream-0.2.21-tier1/verification.md) 和 [第二档验收](../../../openspec/changes/port-upstream-0.2.21-tier2/verification.md)，不将历史 race、浏览器和索引压力测试称为本轮重跑。

第一档 28 个验收文件与其实现提交逐个哈希一致；第二档 69 个验收文件与当前产品提交一致，见 [scope-check](scope-check.json)。两批共享文件的最终整合使用第二档验收；第一档独立快照的原始结论保持原归属。

发布前 API 查询结果：第一档实现提交的 CI success、Security Scan failure；失败步骤为 backend-security 的 Run govulncheck，frontend-security success。这里只记录执行结果，未据此推断具体漏洞或网络原因，未改依赖或安全例外。当前第二档产品提交尚未推送，无此 SHA 的远端 workflow 结果。旧 v1.1.24 Release 已发布。原始状态见 [tier1-ci](tier1-ci.json)、[安全任务](tier1-security-jobs.json)、[product-ci](product-ci.json) 和 [previous-release](previous-release.json)。

## 本轮实际检查

| 检查 | 结果 | 原始证据 |
|---|---|---|
| 全量 backend unit | 退出0，66.228秒，54个测试包通过 | [unit.log](unit.log) / [unit.json](unit.json) |
| Release 同类 integration | 退出0，144.632秒 | [integration.log](integration.log) / [integration.json](integration.json) |
| 前端 typecheck+build | 退出0，44.344秒 | [frontend-build.log](frontend-build.log) / [frontend-build.json](frontend-build.json) |
| 静态 embed build | 退出0，11.467秒 | [embed-build.log](embed-build.log) / [embed-build.json](embed-build.json) |
| 实际二进制 --version | 退出0，输出 Sub2API 1.1.25 | [binary-version.log](binary-version.log) / [binary-version.json](binary-version.json) |
| tag preflight 测试 | 退出0 | [release-tag-preflight.log](release-tag-preflight.log) / [release-tag-preflight.json](release-tag-preflight.json) |
| 发布配置测试 | 退出0 | [release-publish-config.log](release-publish-config.log) / [release-publish-config.json](release-publish-config.json) |
| shell 解析/版本解析 | 退出0，版本解析为 1.1.25 | [release-shell-syntax.json](release-shell-syntax.json) / [version-resolver.log](version-resolver.log) |
| 工具版本 | Go1.27.3 / Node24.16.0 / pnpm11.21.0 | [tool-versions.log](tool-versions.log) / [tool-versions.json](tool-versions.json) |
| 内容范围与实施哈希 | 无未授权既有文件变化，依赖/生成文件未变 | [scope-check.json](scope-check.json) / [before-hashes.json](before-hashes.json) |

unit/integration 均使用 `-mod=readonly`，允许复用 Go 已有测试缓存；只在测试子进程移除 HTTP_PROXY/HTTPS_PROXY/ALL_PROXY 及小写同名变量，沿用既有 captcha 环境隔离方式。integration 使用既有 `integration` 标签；PG 专用测试的原有标签/排除保持不变，不宣称执行 PostgreSQL 测试。

前端使用现有安装依赖及 Node24/pnpm11 构建，未改 package/lockfile；远端 Release 按既有流程使用 pnpm9 frozen install。既有 pnpm、Browserslist 和 bundle chunk 警告保留在日志。此轮未重跑前端全量/critical/lint/浏览器；第二档已有2197通过/2既有跳过、critical179/2及相关验收。

提交内容的 `git diff --cached --check` 在原始 Vite 日志中报告5行行尾空格；日志保持工具原始输出，不清洗证据。排除本版原始 `*.log` 后的产品/文档/JSON空白检查通过。

embed 使用刚构建的前端、CGO_ENABLED=0、trimpath 和发布版本注入；临时二进制位于 `/tmp/sub2api-v1.1.25-release-check`，实际运行 --version 正确。此预提交构建注入产品基线 SHA；远端发布构建由工作流注入最终 tag 提交 SHA。未部署或操作业务数据库。

## 内容和发布顺序

[详细中文说明](../v1.1.25.md) 同时作为 annotated tag 内容，包含同步上游、相对上一 tag 的本仓库新增、升级行为、验证边界和明确未合；23 簇来源、两个实现 SHA、SQLite229、原子刷新/WS1013/图片代理等待及仍延期的新能力均已说明。

先提交 VERSION、说明、证据和 AGENTS 索引，再 `git tag -a v1.1.25 -F docs/releases/v1.1.25.md`。核验对象类型、tag 正文、提交内 VERSION 和版本解析后，依次 `git push origin main`、`git push origin v1.1.25`，仅推此 tag。仓库 `docs/*` 及 `*.log` 为忽略项，本版说明和证据按既有发布惯例明确 force-add，不改忽略规则。

推送后核对远端 main、tag 对象及 peeled commit，并确认 Release workflow 使用本次 tag/SHA。这个提交前快照不回写运行结果；实际推送/工作流状态在最终回复中报告。未归档活动 OpenSpec，也未部署业务实例。
