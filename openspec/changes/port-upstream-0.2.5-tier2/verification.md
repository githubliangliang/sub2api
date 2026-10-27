# 验收证据

## 四项余项完成（2026-10-28）

当前状态 **33/33 已关闭**，实现提交为 **`b349c7408325a0e4eba4f32268ad7a3820e2ad4f`**。本次起点 `bdd77dca1` / VERSION 1.1.20，Go 1.27.3；迁移从 226 升至 227，Ent/Wire、依赖和平台枚举未改。原 source-* 与 evidence TSV 保留原 SHA 和内容。

| 余项 | 实施与证据 |
|---|---|
| S02 #7043 | [容量决策](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/ws-default-decision.json)：保留默认 1.0，不采用数值 5.0。配置默认、legacy/v2 容量、硬上限、零并发与保留会话获取测试通过；业务活跃会话、账号并发和资源余量未采集，不能据此扩大默认连接预算。 |
| S12b 批准记录 | 用户要求完成本 change 未完成任务，范围包含迁移 227；指令原文、范围解读和起点状态见[start-state.json](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/start-state.json)。只实施仓库代码、以临时库验证，未执行业务升级。 |
| S12b 配额语义 | 仅保存至少一档非 NULL，0 也保留；空默认不建行、初始化只补 NULL、清空软删、重新配置新行/新窗口、更新限额保留用量。增量/快照不建空行或复活软删行；旧 TEXT/DATETIME 窗口读取、1000 行参数安全批次、整批事务与 12 请求并发累计验证通过；[SQLite 回归](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/quota-sqlite-complete.log)。 |
| S12b 迁移/UI | 本地 227 仅清理三档全 NULL，保留 0/非 NULL 及其用量；[真实 226→227 升级](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/quota-migration-upgrade.log)与重复执行通过。卡片并集、计数排除其他、未保存配置不能重置与已保存 0 可重置的 16 条组件用例通过；[前端定向](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/quota-frontend-green.log)。 |

| 本次门禁 | 实际结果 |
|---|---|
| 后端 build / 全量 unit | 全部 exit 0，54 个有测试包通过；[build](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/backend-build.json)、[unit](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/backend-unit.log)。 |
| Go lint / 迁移 | lint **0 issues**、迁移 exit 0；[lint](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/quota-lint-final.log)、[迁移](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/backend-migrations.json)。 |
| quota 定向 race | repository/service/admin 三包 exit 0；[race](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/quota-race.log)。 |
| 前端全量 / typecheck / lint / build | 287 文件，**2095 passed / 2 既有 skipped / 0 未处理异常**；各项 exit 0；[Vitest](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/frontend-full.log)、[类型](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/frontend-typecheck.json)、[lint](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/frontend-lint.json)、[build](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/frontend-build.json)。 |
| `make test-frontend-critical` | 14 文件 / 176 passed / 2 既有 skipped，exit 0；[关键套件](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/frontend-critical.log)。 |
| OpenSpec strict / 冻结与完成检查 | 固定 CLI 1.14.3 严格校验通过；33 项全关闭，新增证据链接存在，产品/日志/冻结来源哈希验证通过；[strict](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/openspec-strict.json)、[收尾检查](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/closure-check.json)。 |

[收尾报告](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/completion-report.json)保存完成 SHA、所有命令与日志哈希；[18 个产品文件哈希](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/product-files.sha256)逐项对照该提交。[实现说明](../../../docs/upstream-sync/completion-0.2.5-tier2-2026-10-28/implementation-notes.md)记录 SQLite 必要适配、失败阶段与修正，不把 PG-only 测试或未采用 5.0 记成移植。未推送、发布、归档或部署业务实例。

## 历史实施验收（2026-09-19 至 2026-09-20）

现行簇已实现并完成定向与全量门禁。本文件记录可复跑命令和未批准/延期项。
后端命令在 backend/，前端命令在 frontend/，make 命令在仓库根目录执行。

实际实施起点 SHA：`5803bc1cf`
完成产品代码 SHA：工作树未提交
执行日期：`2026-09-19` 至 `2026-09-20`
第一档完成与验证证据：沿用第一档现有验收；本批共享文件已通过全量 unit/前端关键套件
S12b 用户批准记录（含迁移编号）：未批准，迁移 227 未实施

OpenSpec CLI 当前 PATH 不可用，未宣称 strict 校验通过；冻结 source-* 与 evidence TSV 未改写。

现行21项；已前移七项由第一档T17–T23验收，本文件不再保留重复产品证据槽。
source-*与原TSV保留原分档，现行归属见design.md。

## 逐来源行为 red / green

每行记录可复跑命令、明确触发输入、改前失败与改后通过证据、正常路径反例及落地 SHA。
共享测试可复用，但须指出覆盖哪项行为；来源未移植不等于验收通过。

| 簇 / 来源 | 必须覆盖的行为 | 改前 / 改后 / 反例证据 | 落地 SHA / 缺口 |
|---|---|---|---|
| S01 #6925 | 正文恢复、后缀、去重与生命周期 | apicompat 全量 unit 通过 | 工作树未提交 |
| S01 #7094 | 前导合并、中途 user、原序保留 | apicompat + agent_message 定向回归通过 | 工作树未提交 |
| S02 #6965 | ping、消息顺序、关闭、脏连接与交接竞态 | WS/repository 全量 unit 通过 | 工作树未提交 |
| S02 #7064 | 动态容量、当前默认1.0、显式系数、硬上限、零并发 | req_client_pool 定向/全量 unit 通过 | 工作树未提交 |
| S02 #7043 | 后段独立容量依据；采用5.0时验证默认与显式覆盖 | _待评估_ | _未实施；可单独延期_ |
| S03 #6689 | 三路径混合工具、旧策略清理、仅内置反例 | antigravity 全量 unit 通过 | 工作树未提交 |
| S04 #7076 | 部分刷新 account + warning | `TestAccountRefreshSuccessData` + AccountsView 回归通过 | 工作树未提交 |
| S05 #6977 | 冻结名单、后缀与映射/透传例外 | 官方 `api.deepseek.com` 账号身份的 account/model mapping 回归通过 | 工作树未提交 |
| S06 #6661 | 槽一次释放、所属账号固定、TTL 不覆盖 | Grok media slot/owner handler + service 回归通过 | 工作树未提交 |
| S08 #6943 | cache_control 保留与最多 4 个断点 | gateway/count_tokens 定向回归通过 | 工作树未提交 |
| S08 #6995 | 消息级 output_config 与最终 beta | mid-conversation output_config 定向回归通过 | 工作树未提交 |
| S09 #7074 | 规范窗口、快照锚点与评分一致 | canonical quota/reset 定向回归通过 | 工作树未提交 |
| S09 #7072 | 调度耗时与总体粘性去重 | scheduler metrics 定向回归通过 | 工作树未提交 |
| S09 #6869 | 旧版/完整心跳与畸形拒绝 | automation heartbeat 定向回归通过 | 工作树未提交 |
| S10 #6783 | 固定载荷分配字节前后比较、读取边界 | 69 MiB：legacy `267387104 B/op`，chunked `144712760 B/op`；读取边界回归通过 | 工作树未提交 |
| S10 #7010 | 隐私指纹、质询识别、OAuth 隔离 | privacy challenge 定向回归通过 | 工作树未提交 |
| S11 #6654 | 注册确认密码校验 | RegisterView 定向回归通过 | 工作树未提交 |
| S11 #6916 | 分配搜索排除已删除用户 | RED 为 `users.list` 0 次调用；改后 2 个 SubscriptionsView 用例通过 | 工作树未提交 |
| S11 #7111 | 凭据省略/null/空串及导入适配 | Go handler/service + ProxiesView 定向回归通过 | 工作树未提交 |
| S12a #6971 | TTFT 展示、过滤与 SQLite 排序 | 真实 SQLite 覆盖降序、并列 created_at 与 NULL-last；API/modal 回归通过 | 工作树未提交 |
| S12b #6954 | 获批后限额清理、SQLite 写入与迁移 | _待填_ | _待批准，未实施_ |

## 门禁

| 门禁 | 命令与结果 |
|---|---|
| 后端 build / unit / lint | `go build ./...`、`go test -tags=unit ./...` 通过；golangci-lint 仅报 3 个未改动既有测试告警 |
| 前端 typecheck / lint | `pnpm run typecheck`、`pnpm run lint:check` 通过 |
| 前端定向用例 | tier2 新增 Vitest/后端定向用例通过 |
| 前端关键用例 | `make test-frontend-critical`: 175 passed, 2 skipped |
| SQLite 方言审计 | ops repository unit 覆盖 `IS NULL` 排序，无迁移 SQL |
| 真实 SQLite TTFT 排序 | `go test -tags=unit ./internal/repository -run TestOpsRequestDetailsTTFTSortRunsOnSQLite` 通过 |
| S12b 真实 SQLite 与迁移 | _获批后填写，否则记录未实施_ |
| VERSION / 迁移 / 范围 | VERSION=1.1.12；无 Ent/Wire/依赖/新平台；无 S12b 时最大迁移 226 |

## 来源取舍记录

以下为规格范围约束，尚非实施证据；实施时补充实际 hunk 与理由。
冻结 source-feature-map.md 的编号笔误不改写：其中“第一档 S11 #7024”应按第一档条目查阅；本批现行归属以design.md为准，S07与S11六项已迁第一档。

| 来源 PR | 排除或适配范围 | 理由 / 实施记录 |
|---|---|---|
| #7064 | WS 模式前端/i18n 文案 | 与容量行为无关，本批未纳入；仅移植服务端容量公式与测试 |
| #6783 | 缺失的 raw-input/ingress-compat 基座及其依赖重构 | 仅适配 `httputil/body.go` 的现有读取路径并记录同载荷 benchmark |
| #6954 | 上游迁移编号、PG 专用批量快照 SQL 与 PG-only 测试 | 获批后按本地 227 与 SQLite 改写；_待批准_ |
| #6869 | 冲突的 automation bootstrap 测试 | 按本地 `openai_gateway_handler` heartbeat 校验形态重写并通过 |

## 结论

- [x] 现行各簇（不含迁出的S07）与S12a契约满足；#7043 单列延期，不阻塞 #7064。
- [x] S12b 明确列为待批准/未实施，未将其计入已完成范围。
- [x] 未满足项、排除 hunk 和 lint 缺口已列明。
- [x] 范围门禁与实际迁移状态一致，未新增迁移或本 change 外能力。
