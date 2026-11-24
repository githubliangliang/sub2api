# Verification（2026-11-24 已实施）

F01已在工作区完成。起点为 `39ca66d51d53b98f11349598bbd16cd8ecd0e2e0`；与第二档联合验收，实际结果见[实施证据](../../../docs/upstream-sync/implementation-0.2.25/README.md)。[评估证据](../../../docs/upstream-sync/evidence-0.2.25/MANIFEST.md)保持原样，单独表述。

| 项目 | 实际命令 / 内容哈希 | 结果 / 局限 |
|---|---|---|
| 真实实施起点与最终哈希 | implementation-0.2.25/initial.json、acceptance-start-hashes.json、final-backend-hashes.json、scope.json | 8,774个初始文件保护；组合验收期间8份生产源码/模块哈希一致，最终测试专用变更单独复验。 |
| backend unit / build / lint | backend-unit.json、backend-embed.json、backend-lint-final.json；实际命令及原始输出见实施目录 | 55测试包、10,793顶层/22,386含子例通过，16既有跳过；embed通过，Go1.27.4构建的lint2.14.0为0 issues。 |
| 相关 race | backend-error-race.json；s01/race.json；s02/websocket-race-final.json | 三包332/669、两包392/800、五包693/1,351分别通过，重叠范围不相加。 |
| 实际 HTTP / WS / SQLite / miniredis | f01-http-corrected.json、final-targeted.json、sqlite-billing-controls.json；S01/S02子目录 | 上下游实际本地TCP、403/单信封/头净化通过；四包计费配额69/136通过；已有miniredis控制包含于全量unit。无外部供应商或业务实例验收。 |
| 依赖 / 资源及测试局限 | 第二档S02与s02/README.md | F01不引入依赖；组合仅精确维护Gorilla。资源默认、Go/VERSION/SQLite230/auth21与生成代码保持；库历史代理EOF已定因但原始测试仍失败。无全量integration或生产容量验收。 |
| 前端（仅实际涉及时） | scope.json | 无前端修改，不运行新增前端门禁。 |
| 历史 / 第三档 / 支付边界 | scope.json | 旧13活动change/439完成任务、冻结来源、历史与初始Fast诊断保持。第三档/支付不引入；没有提交/推送/发布/归档或业务实例操作。 |
| F01 | f01-before.json、f01-helper.json、f01-http-before.json、f01-http-fixture-note.md、intake-probes-after.json；upstream_error_redaction.go及测试 | 先复现配置泄漏；三类凭据按长值优先脱敏，主机在原字节上大小写替换；query、Unicode/非法UTF-8、IPv6、正则字面量、无匹配、不可解析回退、普通403及内部诊断所有权均通过。首次HTTP夹具403误用context code的失败单独保留。 |
