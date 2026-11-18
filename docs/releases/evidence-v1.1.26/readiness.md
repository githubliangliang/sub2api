# v1.1.26 本地 tag 准备与验证

2026-11-18，用户请求「打tag v1.1.26,详细描述」。本次更新 VERSION、整理完整中文增量、提交发布准备并创建本地 annotated tag；没有请求推送，不触发远端 Release/镜像或业务部署。

## 基线与实际增量

- 上一 tag：`v1.1.25`，annotated object `e23c14b5868e57df0c74365c042c04e9d5fe899c`，peeled commit `cb84a2c3c1d3889f4b0f73b368df52f03e0cfa6d`。
- 已验收产品 HEAD：`3652ddb6aa9f39f0cc0ef4e8c7caa381fc784b13`。开工工作区干净，VERSION 为1.1.25，本地没有 v1.1.26。
- 相对上一 tag 只有三项产品提交：`7a0b78fe5` 用量用户排行路由修复；`d8d126da9` 第一档 F01–F21、70/70；`3652ddb6a` 第二档 S01–S23、76/76。完整 SHA 见 [release-facts.json](release-facts.json)。
- 来源按冻结 git 对象的祖先关系分为 v0.2.22/v0.2.23，55 个选定 PR 的归属与中文说明逐项一致；44 个簇全部列出，S22 跨版本子集分别说明。完整归属见 [upstream-pr-releases.json](upstream-pr-releases.json)。
- 只读 `git ls-remote` 确认远端 main 已为上述产品 HEAD、远端 v1.1.26 不存在：[命令](remote-before.json) / [原始结果](remote-before.log)。这次没有执行 git push，远端 main 已有产品提交不代表发布准备或新 tag 已推送。

## 版本与变更范围

- VERSION 1.1.25→1.1.26；Go 相对上一 tag 为1.27.3→1.27.4，来自第二档已完成的工具链维护。
- SQLite 最大迁移229→229，auth snapshot21；相对上一 tag，无迁移、go.sum、前端 manifest/lock 或 Ent/Wire 内容变化。go.mod 只改变 Go patch 声明。
- 保留 body64MiB、SSE行40MiB、stage/queue8MiB、应用读写idle120s、断开drain180s、WS默认容量系数1.0和固定quota/成功去重语义。
- 发布准备只修改已存在的 `backend/cmd/server/VERSION`、AGENTS 当前发布索引，并新增 [详细中文说明](../v1.1.26.md) 和本证据目录。
- [before-hashes.json](before-hashes.json) 记录8163个已有跟踪文件。[scope-check.json](scope-check.json) 确认只有 VERSION 与 AGENTS 按授权改变，其余8161个已有文件逐字节一致，包括冻结来源、历史发布/实施证据与产品源码。没有无归因变更。

## 已有实施验收

本次引用完整产品已完成的实施验收，不重复声称跑过历史 race、浏览器、跨平台/codegen 或性能测量：

- [第一档 verification](../../../openspec/changes/port-upstream-0.2.23-tier1/verification.md)：54包 unit、embed、lint0、六包race、实际HTTP/WS/SQLite/miniredis/Fast/结算去重/固定quota；前端2206/2既有skip、critical179/2和桌面/移动Chrome。
- [第二档 verification](../../../openspec/changes/port-upstream-0.2.23-tier2/verification.md)：54包 unit、embed、Go1.27.4同版本lint重建0问题、六包race；真实临时SQLite229/Ent和HTTP/WS/miniredis，前端2212/2既有skip、critical179/2及8项桌面/移动浏览器操作。
- 第二档 [实施证据](../../upstream-sync/implementation-0.2.23-tier2/README.md) 与 [提交核对](../../upstream-sync/implementation-0.2.23-tier2/commit-preparation.json) 保留3982个产品文件验收身份、55个第一档产品/测试文件一致性、原始失败/来源适配、331文件Ent及Wire一致生成、跨平台编译和局部资源测量。
- 两档临时1CPU/1GiB独立夹具的内存读数、单P200消息/工具基准及100事件TCP flush结果保留原测量范围，不能外推完整进程或生产效果。

## 本次实际执行的发布检查

| 检查 | 结果 | 原始记录 |
|---|---|---|
| 后端 `go test -mod=readonly -tags=unit ./...` | 54个有测试包通过，150.800s；其中51包使用Go测试缓存 | [command](unit.json) / [output](unit.log) |
| Release同类 `go test -mod=readonly -tags=integration ./...` | 48个有测试包通过，169.833s，无cached包 | [command](integration.json) / [output](integration.log) |
| 前端 `pnpm run build` | vue-tsc及Vite通过，76.703s | [command](frontend-build.json) / [output](frontend-build.log) |
| CGO=0、embed、trimpath、readonly构建 | 包含本次实际前端产物，3.642s；本地二进制位于 `/tmp/sub2api-v1.1.26-release-check` | [command](embed-build.json) / [output](embed-build.log) |
| 实际二进制 `--version` | `Sub2API 1.1.26`，Commit注入产品基线3652ddb6a | [command](binary-version.json) / [output](binary-version.log) |
| release-tag-preflight测试 | 通过，拒绝分支/SHA并正确解析annotated及lightweight测试夹具 | [command](release-tag-preflight.json) / [output](release-tag-preflight.log) |
| release-publish-config测试 | 通过，GoReleaser固定RELEASE_TAG、重试资产替换配置有效 | [command](release-publish-config.json) / [output](release-publish-config.log) |
| 发布脚本bash语法 | 四个现有脚本通过 | [command](release-shell-syntax.json) / [output](release-shell-syntax.log) |
| VERSION解析 | 未打tag的产品HEAD正确从文件解析1.1.26 | [command](version-resolver.json) / [output](version-resolver.log) |
| 工具版本 | Go1.27.4、Node24.16.0、pnpm11.21.0 | [command](tool-versions.json) / [output](tool-versions.log) |
| 范围/文档 | 已有8163文件哈希核对，44簇齐全，55PR版本归属正确，VERSION/resolver/binary一致 | [scope-check.json](scope-check.json) |

`capture.py` 保存命令argv、cwd、时间、退出码和完整原始输出。后端两个测试子进程移除环境代理，沿用已确认的本地captcha SDK夹具适配；产品代理行为未改。前端使用现有依赖，未安装或改写锁文件；远端Release仍按其既有pnpm9 frozen-lockfile配置处理。

unit、integration与前端构建独立并行执行，构建缓存和同机负载影响耗时；这里不作性能比较。integration是本fork本地实现的测试标签，不声称运行不存在的PostgreSQL/供应商场景。原始Vite日志、测试输出中的空白不清洗；产品/说明的空白检查单独执行。

## 提交与tag契约

1. 先提交包含VERSION1.1.26的发布准备；完整中文说明使用 [v1.1.26.md](../v1.1.26.md)，同时作为annotated tag正文。
2. 创建 `git tag -a v1.1.26 -F docs/releases/v1.1.26.md`，指向上述发布准备提交；禁止lightweight tag，不移动旧tag。
3. 创建后核对tag对象类型、peeled HEAD、tag内VERSION、正文与文件一致、精确tag版本解析，以及Release preflight解析出的SHA。tag对象和本提交自身的SHA产生于提交之后，在工具结果另行验证，不伪造写入本提交证据。
4. 保持工作区干净；没有推送、远端CI/Release验收、镜像/资产发布、OpenSpec归档或业务数据库操作。

文档/原始记录受现有gitignore覆盖，新增发布文档及证据显式force-add，未修改ignore规则。[SHA256SUMS](SHA256SUMS) 覆盖本目录除校验清单自身外的全部文件。历史intake/implementation中的VERSION、HEAD、未提交状态和当时失败照原记录保留。
