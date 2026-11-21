# v1.1.27 发布准备（2026-11-21）

用户明确请求「打tag v1.1.27 ，详细描述。push」。本次将 VERSION 更新为1.1.27并提交，使用[完整中文说明](../v1.1.27.md)创建 annotated tag，依次推送 main 与新标签。GitHub Release 沿用现有 tag 工作流。

## 起点与范围

- 初始工作区干净，本地/远端 main 均为 `c2075a8e98d8d0afda2d6b0b407ad9a3019fc5b0`。
- 上一标签 v1.1.26 的本地/远端对象均为 `34870d4fb390b83ea18d871d850bf578a65a03b5`，peeled commit 为 `9c073cb30bff9dfd9364a059d4e0929b14d8b8ff`，保留其对象、正文及历史说明。
- 查询时远端没有 v1.1.27 tag，Release API 返回404；上一版 Release 已发布，带 checksums.txt 和五个平台二进制压缩包。本版推送后产物以实际工作流为准，不提前记为成功。
- 真实功能增量只有两个提交：`d2488b1105f091b6fce4445395fec1835306fcf8` 的 #8748/#6188（各6任务），以及 `c2075a8e98d8d0afda2d6b0b407ad9a3019fc5b0` 的 v0.2.24 第一档10簇/37任务、第二档15簇/52任务；本版合计27簇/101任务。
- 相对 v1.1.26，SQLite229→230，仅新增可空请求推理强度列。Go1.27.4、auth21、依赖、Ent、资源默认保持；Wire 已在实施提交中为默认平台 quota 重生成。
- 本次发布准备只改 VERSION、新版说明、AGENTS 当前发布索引，并新增本目录证据；不更改业务代码、工作流、依赖、迁移或生成文件。起点与任务数见[release-facts.json](./release-facts.json)。

## 验收复用与本次执行

[source-verification.json](./source-verification.json)逐项核对 v0.2.24 最终清单的2174个产品文件、1857个测试文件及1个生成文件，合计4032个。唯一有意差异为 VERSION 1.1.26→1.1.27；产品、测试和生成源码无意外漂移。原始验收清单及日志不回写。

据此复用[两档最终验收](../../upstream-sync/implementation-0.2.24/README.md)：55测试包、10776顶层/22277含子例通过、16原skip；九包race283顶层/819含子例通过；lint0、Wire、实际本地HTTP/WS/SQLite/miniredis/配额/独立成本/成功去重控制；前端2254通过/2原skip、定向112项、typecheck/lint/build、桌面/移动Chrome6操作。前序[#8748/#6188独立实施证据](../../upstream-sync/implementation-passthrough-filter-and-usage-effort/README.md)保留其54包unit、相关race、2224前端测试及229→230和CSV/XLSX检查，不冒充本次重新执行。

本次重新执行并通过：

| 检查 | 结果与记录 |
|---|---|
| 发布标签preflight隔离控制 | 通过；[命令与结果](./release-tag-preflight.json)、[日志](./release-tag-preflight.log)。尚未创建本版tag，不把隔离控制写成实际tag解析。 |
| Release发布配置/下载重试控制 | 通过；[命令与结果](./release-publish-config.json)、[日志](./release-publish-config.log)。 |
| 工作流actionlint | 通过；[命令与结果](./actionlint.json)，输出为空、退出0。 |
| 发布相关shell语法 | 通过；[命令与结果](./shell-syntax.json)，输出为空、退出0。 |
| 前端构建 | pnpm run build通过，含vue-tsc -b与Vite；[命令与结果](./frontend-build.json)、[日志](./frontend-build.log)。保留已有chunk大小提示，不将提示写成失败。 |
| CGO_ENABLED=0 embed构建 | 使用上述新构建SPA，通过；[命令与结果](./embed-build.json)。构建不注入main.Version，直接检查提交内VERSION的嵌入路径。 |
| 实际二进制版本 | --version输出Sub2API 1.1.27，通过；[命令与结果](./binary-version.json)、[日志](./binary-version.log)。未注入commit/date的本地检查二进制如实显示unknown，不作为发布产物。 |
| 发布说明/链接/范围/空白 | 提交前检查完整三段要求、F01–F10/S01–S15、101任务、全部仓库内说明链接、精确版本和源码漂移；[结果](./local-checks.json)。 |

本次没有重跑历史race、浏览器、fuzz或1CPU/1GiB性能夹具。原夹具为临时合成账号、真实本地传输/SQLite及miniredis，浏览器使用合成管理API；185131008B的middleware峰值不能用作完整生产服务容量结论。

首次暂存空白检查发现新保存的Vite日志中5行reporter提示带行尾空格；定位到构建器原始stdout后，仅移除这些空格，不修改构建结果或消息内容。原始/整理后SHA256和行号记录在frontend-build.json，整理后重新执行暂存范围与空白检查。

起点main的远端CI在查询时shell/lint/frontend成功，integration仍运行；同一SHA的Security Scan为既有failure。本次不修改依赖或扫描规则，也不宣称该扫描通过。远端状态与本地测试分开归因。

## 提交、标签与推送顺序

1. 提交 VERSION1.1.27、完整中文说明、当前AGENTS索引和本目录检查记录。
2. 用 `git tag -a v1.1.27 -F docs/releases/v1.1.27.md` 创建新注释标签；确认对象类型为tag、peeled commit为发布提交、提交内VERSION匹配、正文与说明一致，并执行真实版本解析及preflight。
3. 先推送 `main`，再推送 `v1.1.27`；不覆盖旧标签或强推分支。
4. 推送后核对远端main、tag对象与peeled commit，按实际发布SHA跟踪Release工作流，检查正文及实际产物。后续对象SHA与结果由真实执行产生，不预写到本提交。

历史PORTING、来源、实施/发布证据及已完成活动OpenSpec保留原文和当时状态，不归档或操作业务实例。第三档新能力、5图扩展、完整group allowlist/QuotaSummary及支付不在本次发布准备的新增实现范围。

## 同日推送后集成门禁跟进

上述首次准备与检查对应提交`8bf50be352c2a7f2ec57b140738dcbad5c189d3c`。首轮Release的preflight、前端构建与后端unit成功，integration失败，发布job跳过、Release仍404。匿名下载日志需认证，公开注解只提供退出码；独立本地CI=true/UTC全量integration复现唯一失败为旧缓存Group非空断言。

最终发布范围补充一个integration测试文件：完整缓存保留凭据/配置/成员标量，省略展开关系，调用者关系图不变。业务源码、前端、VERSION与构建流程不变。修正后真实Redis三项控制通过，完整integration49包/6163顶层/12395含子例通过、14项测试跳过，约206.6秒。相对原4032清单的有意差异因此为VERSION和这一个测试文件；初轮source-verification及检查记录保留原时点，最终差异另见[跟进记录](./integration-followup/README.md)。

初轮tag对象`ba1fbede506606a3b7a15fc1e97c40af439fe652`保留在本地`refs/backups/v1.1.27-before-integration-fix-20261121`。先提交本次测试/说明/证据并正常推送main，再用该精确旧对象lease更正尚未发布的v1.1.27；不强推main，不修改v1.1.26。最终SHA、远端refs和新Release结果按真实执行核对。
