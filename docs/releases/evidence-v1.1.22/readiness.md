# v1.1.22 发布准备验收

2026-11-01（Asia/Shanghai）。用户要求打 v1.1.22 tag 并提供详细描述，按仓库 AGENTS.md 的发布顺序提交 VERSION、创建中文 annotated tag，再依次推送 main 和该 tag。产品内容提交为 `41ade9984af76a65af0295abcf3bd15f2357d614`；前一 tag `v1.1.21` 对象为 `e6f17b32d000ed0b08d6e7c5c2bda0441110d6e6`，解引用提交为 `e44324231ec98b5176fe04c6ddb3f3552a352889`。

## 产品范围与验收复用

- [详细中文发布说明](../v1.1.22.md)覆盖 v1.1.21 之后六个已有提交：模型参考价格 `5250907f5`、兼容性与 Grok 诊断 `becd2d2a3`、验收核销 `2fef8d408`、归档 `11260f7c4`、三语 README `5ca7c9b2b` 和 follow-redirects 1.16.2 `41ade9984`。
- 上游选定来源分别为 v0.1.180 的 C01–C03、v0.2.5 的 C04/#7129，以及 v0.2.19 的 Grok 诊断/#8339。通过历史 commit 的包含 tag 和已冻结 feature map 核对，不扩大为其它上游能力。
- 产品提交 backend 与 `becd2d2a3` 的完整验收保持相同内容，复用其 54 包 unit、embed build、四包 456-case race、WS 相邻 74-case race、SQLite 227/quota/固定窗口/dedup 468-case 保护及 lint 0 issues。模型参考价格提交已包含在该验收起点。
- 该实现前端完整验收为 2169 passed / 2 项既有 skipped，另有 typecheck/lint/build/critical 和实际浏览器；前端依赖维护提交另记录 pnpm 9/11 frozen install、8 个重定向用例、关键测试 176 passed / 2 项既有 skipped 和 high/critical 0。上述属于已有实施/维护验收，本轮没有重新全量执行 unit/Vitest/race/审计或浏览器验收。
- 产品提交的远端 [CI](https://github.com/githubliangliang/sub2api/actions/runs/41266577030)和 [Security Scan](https://github.com/githubliangliang/sub2api/actions/runs/41266577044)状态及工作明细保存在 [product-ci.json](./product-ci.json)。已成功的工作与仍在运行的工作分别记录，以该快照的实际值为准。
- [release-facts.json](./release-facts.json)核对本轮产品变更仅 `backend/cmd/server/VERSION`；Go 1.27.3、最大 SQLite 迁移 227、现有依赖/锁文件及 Ent/Wire 内容保持产品提交状态。相对上一 tag 的 Wire 注入差异来自模型参考价格功能。

## 本轮实际执行

| 检查 | 实际结果与证据 |
|---|---|
| 发布脚本语法、tag preflight、publish config | 三项退出 0；[完整命令/结果](./release-script-checks.json)、[原输出](./release-script-checks.txt)。 |
| VERSION resolver | 当前未 tag 工作树精确返回 `1.1.22`，退出 0；命令及输出同上。创建实际 tag 后再核对提交内版本和 tagged checkout resolver。 |
| Release 同用 integration | backend：`rtk proxy make test-integration` 退出 0，142.272 秒，包含 Go 正常复用的缓存；[记录](./integration.json)、[原输出](./integration.txt)。只在测试子进程移除代理变量，沿用已有 Aliyun SDK NO_PROXY host:port 问题的已确认处理；`integration && postgres` 文件仍由标签排除。 |
| 当前前端构建 | frontend：`rtk proxy pnpm run build` 退出 0，65.350 秒；Node 24.16.0 / pnpm 11.21.0，使用现有安装树，实际 follow-redirects 1.16.2，未重新安装或修改锁文件；[记录](./frontend-build.json)、[原输出](./frontend-build.txt)。远端 Release 继续采用 Node 24 / pnpm 9 frozen install；Vite chunk 提示保留。 |
| 1.1.22 静态 embed 构建 | CGO_ENABLED=0、linux/amd64，`go build -mod=readonly -tags=embed -trimpath` 注入 Version/Commit/Date/BuildType，退出 0，18.944 秒；实际 `--version` 返回 `Sub2API 1.1.22`，退出 0；[完整记录和 SHA256](./embed-version.json)、[原输出](./embed-version.txt)。预提交检查注入产品基线 SHA 41ade9984，最终发布提交在检查之后生成；二进制只在 /tmp。 |
| 发布范围与说明 | 上一 tag 对象/提交、六个增量提交、来源 tag、版本/Go/迁移、说明章节/链接和唯一产品变更均核对；[事实快照](./release-facts.json)。 |

源码、Markdown 和 JSON 的 staged 格式检查单独核对；原始 TXT 保留工具输出的字节和诊断，不为消除 reporter 空白而改写日志。

## 提交、tag 与推送

1. 提交 VERSION 1.1.22、完整发布说明、本验收目录和 AGENTS 当前版本索引。`docs/releases/` 被 .gitignore 忽略，只显式 force-add 本版本文件。
2. 创建 annotated tag `v1.1.22`，正文使用完整 `docs/releases/v1.1.22.md`；核对对象类型、解引用提交、提交内 VERSION、正文一致性及 tagged checkout resolver。
3. 先 `rtk git push origin main`，成功后 `rtk git push origin v1.1.22`，只推送本版本 tag。
4. 核对远端 main、tag 对象和解引用提交与本地一致，确认原 v1.1.21 ref 保留；读取本次 Release 工作流和最终发布结果。

本文件是提交前的发布准备快照。实际发布提交 SHA、tag 对象、推送返回及后续工作流状态在提交后通过终端和远端 ref 核实，不回写冻结的历史验收。本轮没有部署业务实例或操作业务数据库；远端跨平台打包、镜像及 GitHub Release 完成状态由本次 Release 工作流核实。
