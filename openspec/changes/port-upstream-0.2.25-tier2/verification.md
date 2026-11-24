# Verification（2026-11-24 已实施）

S01/S02已在工作区完成，F01前置同时实现。起点为 `39ca66d51d53b98f11349598bbd16cd8ecd0e2e0`；实际结果见[组合验收](../../../docs/upstream-sync/implementation-0.2.25/README.md)、[S01](../../../docs/upstream-sync/implementation-0.2.25/s01/README.md)及[S02](../../../docs/upstream-sync/implementation-0.2.25/s02/README.md)。评估期证据不冒充实施验收。

| 项目 | 实际命令 / 内容哈希 | 结果 / 局限 |
|---|---|---|
| 真实实施起点与最终哈希 | implementation-0.2.25/initial.json、acceptance-start-hashes.json、final-backend-hashes.json、scope.json | 8,774个初始文件保护；8份生产源码/模块哈希与组合验收开始一致，15份最终backend文件哈希可核验。 |
| backend unit / build / lint | backend-unit.json、backend-embed.json、backend-lint-final.json及对应log | 55测试包、10,793顶层/22,386含子例通过，16既有跳过；embed通过，lint0 issues。首次测试代码lint11项问题原样保留，最终仅测试修正后复验。 |
| 相关 race | backend-error-race.json；s01/race.json；s02/websocket-race-final.json | 三包332/669、两包392/800、五包693/1,351分别通过，无race；不累计重叠。 |
| 实际 HTTP / WS / SQLite / miniredis | final-targeted.json、sqlite-billing-controls.json、s01/selected-final.json、s02/admin-connect-final.json | 最终71/214、计费配额69/136、S01定向12/77、真实管理员WS/CONNECT2/14分别通过。既有真实SQLite及miniredis控制执行于全量unit。 |
| 依赖 / 资源及测试局限 | s02/mvs.diff、full-module-diff.patch、validation.json、module-verify.json | MVS406模块仅Gorilla变化，校验和及18文件完整差异已核对，go mod verify通过；原库两项EOF保留并定因。保持Go1.27.4/VERSION1.1.27/SQLite230/auth21、其它依赖、Ent/Wire及资源默认；没有全量integration、漏洞利用或生产容量验收。 |
| 前端（仅实际涉及时） | scope.json | 无前端改动，不新增前端门禁或浏览器验收。 |
| 历史 / 第三档 / 支付边界 | scope.json | 冻结来源/评估/历史、初始Fast诊断、旧13活动change/439完成任务保持。新change维持active；无第三档/支付、提交/推送/发布/归档或业务实例操作。 |
| S01 | s01/before.json、monitoring-priority-before.json、selected-final.json；intake-probes-after.json | 先复现五类原缺口与额外优先级缺陷，原对照保留。规则按原状态/body/平台/优先级执行；有限标量字段、256字节前后界、内嵌/空白/畸形code、F01脱敏、显式文案、头净化、监控/副作用通过。真实compact12组合验证本地序号、单失败终态、确定性code及失败无成功用量结果；独立recorder定向另验证可重试容量code和取消控制，SQLite结算去重单独验证。 |
| S02 | s02/README.md及其中原始命令/完整输出 | 精确版本v1.5.4-0.20240701034025-d67f41855da4。真实管理员鉴权/Origin/单subprotocol/全局-IP限额/控制帧/关闭释放及CONNECT/Basic-auth通过。旧/新库原样代理EOF来自Go1.27.4非Hijack CONNECT关闭，wire trace和临时真隧道夹具双版本各2/2确认；原库不宣称全量绿灯。旧Fast WS设置stub冷缓存panic在旧依赖也复现，只补测试values读取后最终WS race通过。 |
