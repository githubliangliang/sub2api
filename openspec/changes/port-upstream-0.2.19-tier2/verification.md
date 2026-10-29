# 实施验收（2026-10-29）

S01–S06 已实施，23/23 任务与全部适用门禁完成。本文件只引用实施证据，不把冻结评估 PASS 当成实施通过。

| 实施信息 | 值 |
|---|---|
| 实施起点 HEAD / 工作树 | `9b44cab33a04f3a9fe1efd0102b5c562a1c458a7` / 干净本地 main |
| 实现提交 / 代码 SHA256 | 本地 main `cccf931efe201671f8cd27284b04e74114f4f894`；汇总 SHA256 `7c81becbc895fc4117a4ea5cbd4a017c7845b7e9a77d8839d0ae275b3f14936d`；[49 文件范围与哈希](../../../docs/upstream-sync/implementation-0.2.19-tier2/scope-facts.json) |
| 实施 VERSION / Go / 迁移 | 1.1.20 / 1.27.3 / SQLite 227 |
| 实际范围 | 6 簇；SQLite/miniredis/simple mode、成功 usage dedup 与固定 quota 窗口保持；x/net 0.62.0 及 8 项强制 MVS 变化 |
| 排除项 | Terra、新平台/新模型目录/诊断界面、支付、版本/schema/迁移、发布/tag/push/部署、业务实例操作 |

## 行为与反例

各项具体测试、初次夹具失败和最终通过输出在[验收矩阵](../../../docs/upstream-sync/implementation-0.2.19-tier2/acceptance-matrix.md)。

| 编号 | 实际行为 | 修复前 / 修复后证据 | 结果 |
|---|---|---|---|
| S01 | 原生 JSON/SSE、Chat/Messages buffered；终态优先、200 rune 固定前序、非 Grok；真实转发→handler 异步 worker→SQLite 响应模型与计价回退 | [四路径前](../../../docs/upstream-sync/implementation-0.2.19-tier2/s01-before-complete-paths.txt) / [完整链后](../../../docs/upstream-sync/implementation-0.2.19-tier2/grok-async-sqlite-complete-fixtures.txt) | PASS |
| S02 | 严格解析/empty_catalog、自定义 ID 大小写、旧快照、失败保留成功、乱序/时间/来源；SQLite 单键语义 CAS、bounded retries、并发 extra、外部事务/回滚/无缓存、普通编辑最新受管值、软删/平台/取消、legacy JSON null 根 | [服务前](../../../docs/upstream-sync/implementation-0.2.19-tier2/s02-before-service.txt) / [SQLite 前](../../../docs/upstream-sync/implementation-0.2.19-tier2/s02-before-sqlite.txt) / [SQLite 后](../../../docs/upstream-sync/implementation-0.2.19-tier2/grok-catalog-sqlite-fixed-fixtures.txt) / [null 根与最终 gate](../../../docs/upstream-sync/implementation-0.2.19-tier2/backend-unit-complete.txt) | PASS |
| S03 | image/image[n]/mask 刚好 20 MiB 与 +1，字节/文件名/MIME、多 part、拒绝不返回部分请求；Grok 有界完整 body；handler 总 body 413/取消、下载限制与 50 MiB 不变 | [前](../../../docs/upstream-sync/implementation-0.2.19-tier2/s03-s04-before.txt) / [完整 gate](../../../docs/upstream-sync/implementation-0.2.19-tier2/backend-unit-complete.txt) | PASS |
| S04 | 两维/部分/缺 key、false/空/null、未知维度、其它 features replacement、caller maps；UI 显式开关、禁用平台、隐藏保留、对应平台提示 | [UI 前](../../../docs/upstream-sync/implementation-0.2.19-tier2/s04-frontend-before.txt) / [消费套件](../../../docs/upstream-sync/implementation-0.2.19-tier2/s05-active-modal-after.txt) / [浏览器](../../../docs/upstream-sync/implementation-0.2.19-tier2/browser/README.md) | PASS |
| S05 | all-mapped/partially-unmapped、去重/通配/归一化/404、custom display 与原路径/force/composite/mixed；普通 3.5 Flash、Google One/显式旧模型；已有 Antigravity 下拉，无未选定新模型 | [backend 前](../../../docs/upstream-sync/implementation-0.2.19-tier2/s05-backend-before.txt) / [实际管理组件前](../../../docs/upstream-sync/implementation-0.2.19-tier2/s05-active-modal-before.txt) / [实际消费套件后](../../../docs/upstream-sync/implementation-0.2.19-tier2/s05-active-modal-after.txt) / [浏览器](../../../docs/upstream-sync/implementation-0.2.19-tier2/browser/README.md) | PASS |
| S06 | 最小 MVS、独立分配比 3.960→1.902、网络/HTML 消费者、Ent/Wire、GO-2026-6860 与剩余 finding 区分 | [前](../../../docs/upstream-sync/implementation-0.2.19-tier2/s06-html-before.txt) / [后](../../../docs/upstream-sync/implementation-0.2.19-tier2/s06-html-after.txt) / [依赖报告](../../../docs/upstream-sync/implementation-0.2.19-tier2/dependency-report.md) | PASS |

## 组合检查

| 检查 | 实际证据 | 结果 |
|---|---|---|
| backend 完整 unit | [最终完整 54 包输出](../../../docs/upstream-sync/implementation-0.2.19-tier2/final-unit.txt) / [argv 与退出码](../../../docs/upstream-sync/implementation-0.2.19-tier2/final-unit.json) | PASS |
| backend build / lint | [build](../../../docs/upstream-sync/implementation-0.2.19-tier2/backend-build.json) / [lint](../../../docs/upstream-sync/implementation-0.2.19-tier2/final-lint.txt) | PASS / 0 issues |
| 受影响 race | [service](../../../docs/upstream-sync/implementation-0.2.19-tier2/backend-service-race-complete.txt)、[handler/repository/securityaudit](../../../docs/upstream-sync/implementation-0.2.19-tier2/backend-race-final.txt)、[3 个网络/目录包](../../../docs/upstream-sync/implementation-0.2.19-tier2/backend-network-race.txt)；过程合并文件 service 的失败已由后续独立通过替代 | PASS / 7 个包实际测试 |
| SQLite 迁移 227 / quota / dedup / 固定窗口 | [输出](../../../docs/upstream-sync/implementation-0.2.19-tier2/sqlite-protection.txt) | PASS |
| frontend typecheck / lint / unit / build / critical | [typecheck](../../../docs/upstream-sync/implementation-0.2.19-tier2/final-frontend-typecheck.txt)、[lint](../../../docs/upstream-sync/implementation-0.2.19-tier2/final-frontend-lint.txt)、[2105 passed / 2 既有 skipped](../../../docs/upstream-sync/implementation-0.2.19-tier2/frontend-tests-complete.txt)、[build](../../../docs/upstream-sync/implementation-0.2.19-tier2/frontend-build-complete.txt)、[176 critical / 2 skipped](../../../docs/upstream-sync/implementation-0.2.19-tier2/frontend-critical-final.txt) | PASS |
| 真实浏览器 / 清理 | [验收与截图](../../../docs/upstream-sync/implementation-0.2.19-tier2/browser/README.md) / [端口释放](../../../docs/upstream-sync/implementation-0.2.19-tier2/browser/cleanup.json) | PASS |
| 生成工具 / 依赖复扫 | [290 文件一致](../../../docs/upstream-sync/implementation-0.2.19-tier2/generation-compatibility.json) / [依赖报告](../../../docs/upstream-sync/implementation-0.2.19-tier2/dependency-report.md) | PASS / 剩余 finding 保留 |
| OpenSpec strict / 链接 / 冻结来源 / 范围 | [范围](../../../docs/upstream-sync/implementation-0.2.19-tier2/scope-facts.json) / [strict](../../../docs/upstream-sync/implementation-0.2.19-tier2/openspec-strict-final.txt) / [最终链接与范围](../../../docs/upstream-sync/implementation-0.2.19-tier2/implementation-validation.txt) | PASS |

## S01 用户确认与过程限制

用户确认保持迁移 227；冲突位到 usage 写入边界并验证 SQLite 计价回退。[确认与原规格快照](../../../docs/upstream-sync/implementation-0.2.19-tier2/s01-acceptance-decision.md)。数据库没有专用冲突列，mismatch 保持原“实际发送 vs 响应模型”语义。只据此修改 S01 一条场景；source-baseline/source-feature-map 与其它来源保持。

Aliyun 本地 SDK 测试的代理绕过仅作用于 unit 子进程，captcha 文件未修改。宽 race 中的全局 map 重置、quota/后台目录请求计数及 Codex 两路独立毫秒时钟均归因为既有测试夹具，生产 gate/指纹代码未改；过程失败原始输出保留。

## 完成条件

- [x] 每项规格有实际前后证据及反例，冻结评估未冒充实施。
- [x] 最终组合门禁与 OpenSpec/链接/冻结核对完成。
- [x] 实现提交与实际起点记录，范围和排除项核对。

完整 unit 最后发现 Codex raw/map 测试比较两次独立生成的毫秒时钟，差 1ms 会失败。测试分别验证生成时间在调用范围内，再比较稳定 metadata；[30 轮复测](../../../docs/upstream-sync/implementation-0.2.19-tier2/codex-clock-fixture-regression.txt)通过，生产指纹不变。
