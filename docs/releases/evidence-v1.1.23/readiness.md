# v1.1.23 发布准备

日期：2026-11-05。用户要求“打tag v1.1.23，详细描述，push”。

## 基线与范围

- 产品基线 `765b4c6b0841fb6b9c6b345543bd06c72f25188b`，准备前本地 main 与 origin/main 一致；远端尚无 v1.1.23。
- 上一 tag v1.1.22 对象 `49337aa2a055128dc8efd271fa7095636a4c7d73`，提交 `ed2340cc5a922ced3e2e408465237f72e8a93579`；四个增量提交见 [release-facts.json](./release-facts.json)。
- 本次产品文件只改 VERSION 至 1.1.23；另外增加发布说明、检查证据和 AGENTS 当前版本索引。Go1.27.3、SQLite227、依赖和 Ent/Wire 无变化。冻结历史不修改。
- 完整中文说明采用“同步上游 / 本仓库自上一 tag 起另外新增 / 明确未合”结构；按实际差异描述 TPS 和日志修复，未将 Fast 日志分类等同新 Fast 功能。

## 已有验收复用

TPS 的后端54包unit、三包race、SQLite/HTTP、2174前端通过和浏览器证据见 [实施索引](../../upstream-sync/implementation-output-tps-distribution/README.md)。后续日志修复已有后端54包unit、三包race、2178前端通过/2既有跳过及真实临时实例TOTP清理验收，见 [修复报告](../../upstream-sync/FIX-logging-2026-11-05.md)。这些是实现阶段证据，未作为本次全量重跑。

产品基线的远端 CI 和 Security Scan 均为 completed/success，见 [product-ci.json](./product-ci.json)。

## 本轮实际执行

| 检查 | 结果 | 证据 |
|---|---|---|
| release-script-checks | 通过，退出 0，0.09 秒 | [release-script-checks.json](./release-script-checks.json)、[完整输出](./release-script-checks.txt) |
| integration | 通过，退出 0，140.56 秒 | [integration.json](./integration.json)、[完整输出](./integration.txt) |
| frontend-build | 通过，退出 0，57.54 秒 | [frontend-build.json](./frontend-build.json)、[完整输出](./frontend-build.txt) |
| embed-build | 通过，退出 0，16.82 秒 | [embed-build.json](./embed-build.json)、[完整输出](./embed-build.txt) |
| binary-version | 通过，退出 0，0.04 秒 | [binary-version.json](./binary-version.json)、[完整输出](./binary-version.txt) |

integration 使用与 Release 相同的 integration build tag，并加 -mod=readonly；允许 Go 复用已通过的缓存。测试子进程移除代理环境变量，沿用已有 captcha 环境问题的隔离方法；不修改全局环境或生产代码。

前端使用已安装依赖、Node24.16.0 / pnpm11.21.0 执行 vue-tsc 与 Vite build，未安装或修改依赖/锁文件；原始输出保留已有 Browserslist、pnpm override 及 chunk 提示。远端 Release 按现有 Node24 / pnpm9 frozen install 配置执行。

根目录 Go launcher 为1.26.5；backend 工作目录依据 go.mod 和 GOTOOLCHAIN=auto 实际运行 Go1.27.3，已单独核对。静态 embed 构建 CGO_ENABLED=0，注入1.1.23与产品基线SHA；实际 --version 返回 Sub2API 1.1.23。二进制在/tmp，未部署。发布提交在检查后创建，远端构建注入最终发布SHA。

## 提交与发布

先提交VERSION与本版文档，创建 annotated tag v1.1.23，正文使用完整发布说明；核对tag对象类型、正文、提交内VERSION、版本解析与发布tag解析脚本。随后依次push origin main和v1.1.23，仅推本版tag；核对远端main、tag对象/解引用SHA和旧tag保持，并检查Release工作流触发结果。

本文件是提交前准备快照。实际发布提交、推送返回与后续工作流状态在提交后核实，不回写本快照。无业务实例部署或业务数据库操作。
