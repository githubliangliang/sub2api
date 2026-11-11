# v1.1.24 发布准备

日期：2026-11-11。用户授权“打tag v1.1.24，详细描述。push”。本文件为提交前检查快照。

## 基线

- 产品基线a25ce7bedb3167e2b9244117893d6762e1a85a95，起始工作树干净。
- 上一tag v1.1.23对象4be87de3438f6cbd963b21acae844c2e4fc016cd，commit5c135c31dfbd6a1fcb701767f2f78933c4280aea。origin/main仍在该commit，是本地main祖先；远端无v1.1.24。
- 三个实际产品增量：326513cdc Fast完整能力、29681fdf5第一档13簇、a25ce7bed第二档11簇。详见[release-facts](release-facts.json)。
- 相对上一tag新增SQLite迁移228、Fast Ent生成和digest缓存关闭Wire；依赖无变更。本次发布准备只改产品VERSION至1.1.24，增加说明/证据及AGENTS版本索引。

## 已有验收

复用Fast、第一档和第二档各自冻结的unit/race/SQLite/浏览器证据，不说成本轮重跑。第二档最终54包unit、embed、lint0、相关race、2186前端通过/2既有跳过、critical179/2、实际本地HTTP/SQLite/miniredis、浏览器及1GiB读体层压力均已完成，见[第二档验收](../../../openspec/changes/port-upstream-0.2.20-tier2/verification.md)。

远端尚未收到三个产品提交，无该product SHA的远端CI结果。旧v1.1.23发布工作流success；该旧SHA后续计划Security Scan的backend govulncheck步骤failure、frontend-security success。见[product-ci](product-ci.json)和[旧计划扫描](previous-scheduled-security.json)，不将旧结果作为本版成功或精确漏洞诊断。本轮未另做Go漏洞全扫，未擅自升级依赖。

## 本次实际执行

| 检查 | 结果 | 原始证据 |
|---|---|---|
| release-script-checks | 通过，退出0，0.086秒 | [release-script-checks.log](release-script-checks.log) / [release-script-checks.json](release-script-checks.json) |
| integration | 通过，退出0，140.046秒 | [integration.log](integration.log) / [integration.json](integration.json) |
| frontend-build | 通过，退出0，56.012秒 | [frontend-build.log](frontend-build.log) / [frontend-build.json](frontend-build.json) |
| embed-build | 通过，退出0，21.404秒 | [embed-build.log](embed-build.log) / [embed-build.json](embed-build.json) |
| binary-version | 通过，退出0，0.033秒 | [binary-version.log](binary-version.log) / [binary-version.json](binary-version.json) |

integration使用Release同类integration标签与-mod=readonly，允许Go复用已有缓存；只在测试子进程移除代理环境变量，沿用既有captcha测试隔离方式。构建使用现有Node24.16.0/pnpm11，未改依赖/lockfile；远端Release仍按既有pnpm9 frozen install。

静态embed构建CGO_ENABLED=0，注入1.1.24及产品基线SHA；临时二进制实际--version为Sub2API1.1.24。发布提交生成后远端构建注入真正发布SHA。二进制输出/tmp，未部署。已有pnpm/Browserslist/chunk警告保留在日志。

## 内容及发布步骤

[详细中文发布说明](../v1.1.24.md)作为annotated tag正文，包含同步上游（历史Fast及v0.2.20两档）、本fork新增、升级变化、验证边界、明确未合。区分Free Fast用户费用与上游成本；列出227→228、snapshot21、64MiB预算/40MiB行/120秒读写/180秒drain与Terra兼容变化。

先提交VERSION及说明，再创建annotated tag；核对对象类型、tag正文、提交内VERSION和版本解析。然后依次push origin main、v1.1.24，仅推此tag。实际提交/远端引用/Release运行状态在提交后核实，不回写这个提交前快照。无业务数据库操作或业务实例部署。
