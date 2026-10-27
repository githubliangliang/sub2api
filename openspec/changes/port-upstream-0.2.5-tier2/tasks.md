# 实施任务

2026-10-28 收尾完成：**33/33 项已关闭**。用户要求执行本 change 的剩余任务；S12b 已实现迁移 227 与 SQLite 配额语义，S02 #7043 决策为保留默认 1.0、未采用数值 5.0。实现提交为 `b349c7408325a0e4eba4f32268ad7a3820e2ad4f`；[当前验收](./verification.md)与[收尾报告](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/completion-report.json)保存证据。以下原实施介绍按历史阶段解释，原任务/验收快照另存。

产品代码已实现并完成后端/前端门禁；S02 #7043 默认值调优与 S12b #6954 迁移按规格明确延期。
先落第一档并验证，再实施本批；S12a 可独立落地，S12b 等待迁移批准。
现行归属以design.md为准：S07已迁第一档T19，S11仅余#6654/#6916/#7111；共21个PR/11簇。
七项前移只调整原43个PR的分工，第三/第四档不进入实施。

## 准备
- [x] 记录实施起点 `5803bc1cf` 与工作树状态；source-baseline/source-feature-map 未改写。
- [x] 确认第一档先落并完成本批目标回归；冻结四态文件未改写。
- [x] 核对 CONFLICT/NOBASE 簇并按本地基座适配，排除缺失 raw-input/ingress-compat 基座。
- [x] 新增测试使用 `unit`/本地 SQLite 可执行路径，未移植 PG-only 测试。

## 优先：正文、协议、资源与配额正确性

- [x] S01 #6925：Responses 正文恢复与生命周期回归通过。
- [x] S01 #7094：前导角色归一、中途 user 与 agent_message 顺序回归通过。
- [x] S02 #6965：WS 连接池既有读循环/探活回归通过。
- [x] S02 #7064：容量公式尊重配置系数、硬上限和动态开关。
- [x] S03 #6689：混合工具路径清理旧内置工具注入。
- [x] S05 #6977：冻结 DeepSeek 名单与 `[1m]` 出站清理。
- [x] S06 #6661：媒体槽释放与视频所属账号约束由新增 Grok slot/owner 回归覆盖。
- [x] S08 #6943：客户端 cache_control 保留、count_tokens 断点最多 4 个。
- [x] S08 #6995：最终 beta 与消息级 output_config 语义一致。
- [x] S09 #7074：规范窗口、快照 reset 锚点与评分口径一致。

## 其次：可观察性与其余既有正确性

- [x] S04 #7076：部分刷新 `account + warning` 后端 envelope 与前端更新回归通过。
- [x] S09 #7072：返回耗时与总体粘性命中去重回归通过。
- [x] S09 #6869：旧版/完整 heartbeat 信封与畸形拒绝回归通过。
- [x] S10 #6783：仅适配本地分块正文读取路径，未引入缺失基座。
- [x] S10 #7010：Firefox 隐私池指纹与 `cf-mitigated` 质询识别回归通过。
- [x] S11 #6916：搜索排除已删除用户并保留失效选择规则。
- [x] S11 #7111：代理凭据省略/null/空串/非空及导入语义回归通过。
- [x] S12a #6971：TTFT 字段、`ttft_desc` SQLite 排序与前端展示回归通过。

## 后段：既有体验与默认值调优

- [x] S11 #6654：注册确认密码校验、显隐与禁用状态回归通过。
- [x] S02 #7043：已独立决定保留 OAuth/API Key 默认 1.0。现有 default/legacy/v2/硬上限/保留会话测试通过；没有业务活跃会话与资源余量依据支持扩大 1C1G 默认预算，数值 5.0 未采用，不将其记作已移植。见[决策记录](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/ws-default-decision.json)。

## 原有批准边界：平台限额迁移

- [x] S12b #6954：本次用户“完成未完成任务项”的指令覆盖剩余 S12b 与迁移 227；原文和仓库实现范围解读记录于[start-state.json](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/start-state.json)。没有执行业务库升级。
- [x] S12b：起点最大迁移 226、nullable limit 与历史 TEXT 窗口列已核对；初始化/清空/零限额/新窗口、增量与分批快照更新、前端并集卡片/已保存配置重置语义均实现。SQLite 增量无 FOR UPDATE，读取兼容 TEXT/DATETIME，分批快照共用事务。
- [x] S12b：新增 `227_purge_unlimited_user_platform_quotas.sql`。真实 SQLite 验证只清理三档全 NULL、保留任何非 NULL（含 0）、幂等、不复活软删行、4100 条快照与分批失败回滚、12 个并发累计；shipped runner 的 226→227 升级和重复运行通过。

## 门禁
- [x] 本次后端 `go build ./...`、全量 unit（54 个测试包）、迁移与 quota race 通过；Go lint 0 issues。
- [x] 本次前端全量 Vitest 287 文件 / 2095 passed / 2 既有 skipped，typecheck/lint/build 通过。
- [x] 根目录 `make test-frontend-critical` 通过（176 passed / 2 既有 skipped）。
- [x] SQLite SQL、真实配额与 226→227 升级验证通过，未使用 PG-only 构建标签冒充验收。
- [x] Ent/Wire、依赖、平台枚举和 VERSION 保持起点值；VERSION 1.1.20、Go 1.27.3，当前最大迁移 227。
- [x] verification.md 已回填本次完成 SHA、实际命令、失败修正、S02 未采用 5.0 决策和 S12b 完成状态；历史快照与冻结来源保留。
