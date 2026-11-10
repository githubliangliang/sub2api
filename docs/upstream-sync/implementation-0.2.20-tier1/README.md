# v0.2.20 第一档实施证据

2026-11-10，用户授权实现第一档F01–F13，44/44任务完成。实际起点`326513cdc68cc50d31e25a07ae98ecd6f5a3ca1f`。实现随本次提交保存；第二档仍未实施。详细行为与验收见[verification](../../../openspec/changes/port-upstream-0.2.20-tier1/verification.md)。

| 证据 | 内容 |
|---|---|
| [baseline.json](baseline.json)、[before-hashes.json](before-hashes.json) | 起点与6692个已有文件哈希，含评估/来源/第二档/历史证据 |
| [environment.json](environment.json) | backend实际Go1.27.3、lint2.14.0、VERSION1.1.23/SQLite228 |
| [before-regressions.log](before-regressions.log) | 初次来源fixture缺失的编译问题，不当作产品缺陷 |
| [before-regressions-compiled.log](before-regressions-compiled.log) | 6包实际执行：25顶层失败、8对照通过，覆盖全部13簇 |
| [after-targeted-final.log](after-targeted-final.log) | 修复后6包91顶层通过 |
| [additional-boundaries.log](additional-boundaries.log) | 15顶层通过，含实际HTTP/SQLite排队结算/dedup/Fast及Claude周用量持久化 |
| [backend-unit.log](backend-unit.log) | 54包全量unit通过；仅子进程移除代理变量，保护既有captcha测试环境 |
| [backend-build.log](backend-build.log) | embed构建通过；临时文件输出未部署 |
| [backend-lint.log](backend-lint.log)、[backend-lint-final.log](backend-lint-final.log) | 初次WS测试helper标签错误及修正后0 issues |
| [backend-race.log](backend-race.log) | 6包定向race通过，含HTTP/SQLite和WS取消/多轮 |
| [sqlite-protection.log](sqlite-protection.log) | 迁移、方言、配额/固定窗口及dedup保护通过 |
| [product-hashes.json](product-hashes.json)、[product.diff](product.diff) | 未提交实现全文件哈希；diff只含已有跟踪文件，新文件由哈希索引定位 |
| [scope-check.json](scope-check.json)、[openspec-tier1.log](openspec-tier1.log) | 冻结来源、范围、任务/链接与OpenSpec严格验证 |
| [run-check.py](run-check.py)、同名JSON | 实际命令、cwd、时间、退出码与日志 |
| original-README.md / original-tasks.md / original-verification.md | 实施前状态原样保存；其它source/proposal/design/spec文件不改 |

实际验证使用本地HTTP模拟供应商与临时SQLite库；未访问真实付费账号或部署业务实例。前端、依赖、VERSION、Go、迁移228、Ent/Wire未变；无前端/浏览器及1C1G预算压力验收需求。Fast实际费用/每轮档位及历史保护保持。
