# v1.1.21 发布准备验收

2026-10-30（Asia/Shanghai）。用户明确选择新版本 `v1.1.21`，要求详细中文 annotated tag，再 push。产品内容提交为 `1b1c06ac6931947faa7cf3d2b0a215d5aac05c5f`；前一 tag `v1.1.20` 对象为 `e85aa1f78eab6f562aee4af9bee3da60e554833e`，解引用提交为 `2deacc20f4ca6ed0661c589323328bef491986ee`。原 v1.1.20 已正式发布；本轮使用新 tag。

## 产品范围与验收复用

- [详细中文发布说明](../v1.1.21.md)只描述 v1.1.20 之后的实际增量：0.2.0/0.2.5 历史余项、SQLite 个人维护、v0.2.19 第一/第二档、用户选定 Terra 和前端安全依赖维护。七个实现提交均已核对属于产品 HEAD 祖先、尚未包含在上一 tag。
- [release-facts.json](./release-facts.json)核对本轮 backend/frontend/deploy/.github 的唯一产品变更为 `backend/cmd/server/VERSION`，从 1.1.20 升为 1.1.21。Go 1.27.3、最大 SQLite 迁移 227、已有依赖/锁文件及生成代码保持产品提交内容；本轮没有新增 schema 或修改旧迁移。
- 产品提交 backend 与最新 [Terra 验收](../../../openspec/changes/port-upstream-0.2.19-terra/verification.md)内容一致，复用其 54 包 unit、embed build、五包定向 race、SQLite/quota/固定窗口/dedup 及 lint 0 issues。
- 2026-10-30 通过 GitHub API 核对产品提交的 [CI](https://github.com/githubliangliang/sub2api/actions/runs/40980108289)和 [Security Scan](https://github.com/githubliangliang/sub2api/actions/runs/40980108054)均 success，CI 的 test、shell、frontend、golangci-lint 四项工作 success。这些记录属于产品 SHA 1b1c06ac6，不冒充本次发布准备提交的新运行。
- 最新前端安全维护提交及相同 SHA 的 CI 验收保留其 2106 passed / 2 项既有 skipped、typecheck/lint、Node 24/pnpm 9 frozen install、生产审计 high/critical 0 等结果。本轮没有再次安装依赖或运行全量 unit/Vitest/race/审计。

## 本轮实际执行

| 检查 | 实际结果与证据 |
|---|---|
| Release 脚本语法、tag preflight、publish config | `rtk proxy bash -n deploy/resolve-release-tag.sh backend/scripts/resolve-version.sh`、两项 `deploy/tests/release-*-test.sh` 均退出 0；[完整 argv/结果](./release-script-checks.json)、[原输出](./release-script-checks.txt)。 |
| VERSION resolver | 当前未 tag 工作树读取 1.1.21，退出 0；命令及精确输出同上。创建 tag 后还要核对实际 tagged checkout 和提交内 VERSION。 |
| Release 同用 integration | backend：`rtk proxy make test-integration` 退出 0，134.455 秒，包含 Go 正常复用的缓存；[命令记录](./integration.json)、[原输出](./integration.txt)。仅在测试子进程移除代理变量，沿用既有 Aliyun SDK NO_PROXY host:port 问题的已确认处理；`integration && postgres` 文件仍按标签排除。 |
| 当前前端构建 | frontend：`rtk proxy pnpm run build` 退出 0，41.707 秒；本机 Node 24.16.0 / pnpm 11.21.0，使用现有已安装依赖树，未重新安装或改锁文件；[命令记录](./frontend-build.json)、[原输出](./frontend-build.txt)。远端 Release 继续使用 Node 24 / pnpm 9 frozen install。Vite 大 chunk 提示保留在输出，构建成功。 |
| 1.1.21 静态嵌入构建 | CGO_ENABLED=0、linux/amd64，`go build -mod=readonly -tags=embed -trimpath`，注入 Version/Commit/Date/BuildType，退出 0，3.832 秒；实际 `--version` 退出 0并含 `Sub2API 1.1.21`；[完整 argv/结果和 SHA256](./embed-version.json)、[原输出](./embed-version.txt)。该预提交构建注入产品基线提交 1b1c06ac6，最终发布提交尚未生成；二进制仅在 /tmp。 |
| 发布说明和范围 | 上一 tag SHA、版本号、Go、最大迁移、实现提交祖先/上一 tag 差集、说明章节/链接及唯一产品变更核对通过；[最终事实](./release-facts.json)。 |

范围检查的初版临时脚本把已有 `154a` 迁移前缀直接转为整数而失败；[原错误](./scope-check-first.txt)保留。对未改的 154a 文件复现后，改为提取前导数字，并验证 154a/006b/227，最终最大迁移核对为 227。只修正本轮检查方法，迁移及 runner 没有变更。

## 提交、tag 与 push 顺序

1. 提交 VERSION 1.1.21、详细发布说明、本验收目录和 AGENTS 当前版本索引。`docs/releases/` 被 .gitignore 忽略，本轮仅显式 force-add 对应新文件。
2. 创建 annotated tag `v1.1.21`，正文使用完整 `docs/releases/v1.1.21.md`。核对 tag 对象类型、解引用提交、提交内 VERSION 和正文一致。
3. 先 `rtk git push origin main`，成功后 `rtk git push origin v1.1.21`。不推其它本地 tag。
4. 核对远端 main、tag 对象及解引用提交与本地一致，并确认原 v1.1.20 ref 值。读取本次 Release 工作流运行链接和状态。

本文件是提交前的发布准备快照。实际发布提交 SHA、tag 对象、push 返回及工作流状态由后续终端结果和远端 ref 核实，未回写历史实现验收。原始 TXT 保留工具空白和诊断；源码、Markdown、JSON 的 staged 格式检查单独核对，不通过修改原输出制造全绿记录。

完整 staged 格式检查退出 2，仅报告原始 `frontend-build.txt` 第 12/15/18/21/24 行的五处 Vite reporter 尾部空格；[检查记录](./format-check.json)保存实际输出。排除本轮原始 TXT 后，VERSION、AGENTS、Markdown 和 JSON 的 staged 格式检查退出 0，原始构建输出保持原样。

本轮未部署本地/远端业务实例，未操作业务数据库；跨平台打包、容器推送和 GitHub Release 完成状态以远端本次工作流为准。
