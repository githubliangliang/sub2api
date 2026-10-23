# v0.2.18 第二档实施证据

实施起点HEAD `4984327590ed29eb4d842bfae8aa3406ef497008`；含同一会话已完成的第一档工作树。以[第二档验收](../../../openspec/changes/port-upstream-0.2.18-tier2/verification.md)为当前状态；评估期`evidence-0.2.18`及两批source-baseline保持冻结。

每次runner执行保留`<name>.txt`原始输出与同名JSON（参数、cwd、开始时间、耗时、真实退出码）。JSON不存在代表尚未完成，历史非零退出码不会被删改成通过。最终接受门禁另汇总在`checks.json`。

## 已接受的证据入口

| 类别 | 证据 |
|---|---|
| 起点与既有改动 | [start.json](./start.json)；最终`scope-facts-final.json`与`code-sha256.json`记录范围、当前代码哈希，第一档起点副本另用SHA核对。 |
| S01 / S09初始修复 | [s01-s09-after](./s01-s09-after.txt)，各自`before`为实际缺陷红测；早期缺包装的编译失败只是适配记录。 |
| S02 | [s02-final](./s02-final.txt)、[转换器全包](./s02-converter-time-after.txt)；`s02-sse-json-time-before`和`s02-converter-time-before`为额外真实红测。 |
| S03 | [s03-final](./s03-final.txt)、[独立审查补强](./audit-catalog-compat.txt)；`audit-before/after`补冷却目录与非单播图片边界。 |
| S04 | [s04-final](./s04-final.txt)；`s04-before`保留身份、容量与浮点封顶缺陷。 |
| S05 | [完整定向](./s05-empty-eof-after.txt)、[race](./s05-empty-eof-race.txt)、[真实SQLite](./sqlite-final.txt)。`s05-contract-before/after`是额外契约红绿；`s05-empty-eof-before`是空/ping EOF红测。 |
| S06 | [s06-final](./s06-final.txt)；`s06-before`为选定既有行为红测。 |
| S07 | [组合](./s07-agent-combined-final.txt)、[race](./s07-agent-race-final.txt)；`s07-agent-exhaustion-before/after`与`secret-before/after`补三协议耗尽和脱敏。 |
| S08 | [三协议](./s08-after.txt)、[真实队列与SQLite](./s08-agent-sqlite-priced.txt)、[queue-before](./s08-queue-before.txt)、[queue-after](./s08-queue-after.txt)、[queue-race](./s08-queue-race.txt)。后两条包括三协议全部12个真实SQLite/队列子例。 |
| 后端全量unit | [backend-unit-complete](./backend-unit-complete.txt)完整命令exit0，54包通过；含部分cache命中，各包输出可复核。 |
| 后端构建 | [最终handler修复后build](./backend-build-accepted.txt)；初期`backend-build-final`不覆盖最后队列改动。 |
| 默认编译 | [backend-default-compile-complete](./backend-default-compile-complete.txt)仅编译无标签测试，证明unit辅助函数/tag适配，没有冒称默认完整测试执行。 |
| race | [受影响链路](./backend-race-final.txt)、[协议转换包全量](./protocol-race-final.txt)、最后的[S08队列race](./s08-queue-race.txt)。第一条中两个协议包无匹配测试，由第二条全包补足。 |
| 前端 | [lint](./frontend-lint.txt)、[typecheck](./frontend-typecheck.txt)、[全量Vitest](./frontend-tests.txt)、[build](./frontend-build.txt)：2027通过/2既有跳过。 |
| 依赖 | [dependency-review](./dependency-review.md)、[dependency-facts](./dependency-facts.json)、[当前Go扫描](./audit-govulncheck-final.txt)、[当前pnpm全量](./audit-frontend-final.txt)、[当前pnpm生产](./audit-frontend-prod-final.txt)。audit有非零退出码及已知遗留，不当作门禁全绿。 |
| 独立审查与第一档保留 | [audit-review](./audit-review.md)、[四包第一档回归](./audit-tier1-preservation.txt)。 |
| OpenSpec | [严格校验](./openspec-strict-complete.txt)，最终文档状态已重新校验。 |

## 诊断记录的含义

- `before`只有断言确实复现产品缺陷才计红测；import/tag/缺少桩依赖、缺少动态价格的失败只用于说明夹具适配。
- S07无`-final`的早期组合/race含人工SIGQUIT夹具挂起诊断，不能作为成功记录；最终报告只引用退出0的`-final`。
- S05 overlay负对照验证新增耗尽终态回归能发现缺接线，不代表完整原工作树；原始`before`另存。
- `audit-backend-unit`在S08和必要x/tools前置未完成时运行，不是接受门禁。`backend-unit-final`和`backend-unit-fixed-final`分别暴露旧价格夹具、心跳计时及分配计数稳定性问题；最终全量结果必须另有退出0记录。
- 既有夹具稳定性：[重复心跳红测](./audit-keepalive-before.txt) → [8轮定向通过](./audit-fixture-after.txt)/[3轮race](./audit-fixture-race.txt)；[分配线性退化负对照](./audit-alloc-linear-negative-control.txt)证明原性能门槛仍生效。最终完整unit另覆盖这两处新夹具。
- 最后S08独立审查发现Messages/native的drop队列丢任务与复用Gin串元数据，`s08-queue-before`保留4个真实失败（Chat控制组通过）。补强前的全量结果不替代补强后的unit/race。
- SQLite真实验证对象为UsageBillingRepository账本、用户余额、账号quota和dedup；usage日志接口沿用桩，不声称插入Ent日志表或验证PG集成。

这些检查没有使用真实Google/Vertex凭据或上游模型权限探测，没有发布、部署或修改运行中的业务实例。剩余依赖告警与代理远端DNS限制见依赖报告及verification。
