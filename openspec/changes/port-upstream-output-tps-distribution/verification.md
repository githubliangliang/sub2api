# Verification

实施起点与冻结来源见source-baseline.md；本文件仅记录实际验收，未执行的项目不记为通过。

| 检查 | 实际结果 | 证据 |
|---|---|---|
| SQLite分布、完整窗口、模式、筛选 | 真实迁移SQLite，1/10/100/1000插值及单/偶/奇/重复/整数秩、全部排除、时间边界、平台归属、raw/auto/preagg通过 | [定向日志](../../../docs/upstream-sync/implementation-output-tps-distribution/backend-targeted-accepted.log) |
| 空样本/清理、取消/截止/失败策略 | 空值序列化/保留聚合时明细清理、真实SQLite执行中断、2秒自身预算、父截止/取消及数据库错误通过 | [定向日志](../../../docs/upstream-sync/implementation-output-tps-distribution/backend-targeted-accepted.log) |
| 管理HTTP、组件、浏览器 | 实际overview+snapshot两接口×3模式通过；组件更新/缺失/小值/全屏通过；新建本地实例实际登录/SPA/SQLite，OpenAI55、Claude7、Grok42、Gemini空、刷新84及390px布局通过 | [HTTP](../../../docs/upstream-sync/implementation-output-tps-distribution/backend-http-final.log)、[组件](../../../docs/upstream-sync/implementation-output-tps-distribution/frontend-targeted.log)、[浏览器命令](../../../docs/upstream-sync/implementation-output-tps-distribution/browser/commands.json) |
| 查询规模/计划 | Go/modernc实际1000/10000/100000样本约3.4/33.6/344.6ms，Go结果分配约4KB；补充Python SQLite计划确认时间/请求类型或分组索引与排序 | [benchmark](../../../docs/upstream-sync/implementation-output-tps-distribution/backend-query-cost.log)、[计划日志](../../../docs/upstream-sync/implementation-output-tps-distribution/query-plan.log) |
| 后端unit/build/lint/race/SQLite保护 | 54包unit、embed构建、lint0 issues；repository/admin/service定向race及SQLite227/quota/固定窗口/dedup/方言保护通过 | [完整unit](../../../docs/upstream-sync/implementation-output-tps-distribution/backend-unit-final.log)、[build](../../../docs/upstream-sync/implementation-output-tps-distribution/backend-embed.log)、[lint](../../../docs/upstream-sync/implementation-output-tps-distribution/backend-lint.log)、[race](../../../docs/upstream-sync/implementation-output-tps-distribution/backend-race.log) |
| 前端tests/typecheck/lint/build/critical | 291文件2174通过/2既有跳过，typecheck/lint/build通过；critical176通过/2同一既有跳过 | [全量tests](../../../docs/upstream-sync/implementation-output-tps-distribution/frontend-unit.log)、[typecheck](../../../docs/upstream-sync/implementation-output-tps-distribution/frontend-typecheck.log)、[lint](../../../docs/upstream-sync/implementation-output-tps-distribution/frontend-lint.log)、[build](../../../docs/upstream-sync/implementation-output-tps-distribution/frontend-build.log)、[critical](../../../docs/upstream-sync/implementation-output-tps-distribution/frontend-critical.log) |
| OpenSpec/链接/冻结范围 | strict通过；链接和受保护3077文件核对结果见检查索引 | [OpenSpec](../../../docs/upstream-sync/implementation-output-tps-distribution/openspec-strict.log)、[最终范围](../../../docs/upstream-sync/implementation-output-tps-distribution/scope-check.json) |

当前成果位于工作树，实施起点`ed2340cc5a922ced3e2e408465237f72e8a93579`，尚无实现提交SHA。VERSION1.1.22、Go1.27.3、SQLite227、依赖及Ent/Wire保持；无业务实例部署。完整后端最终运行允许Go缓存：此前首次完整运行中仅新增HTTP夹具的默认组ID冲突失败，修正后该包与repository重跑通过，其余包复用已通过结果。最终产品文件哈希见scope-check.json。

## 证据边界与过程失败

初轮新增夹具曾复用空request_id而命中真实去重索引；超时递归视图的固定数字时间被SQLite提前过滤，改为行相关合法时间后真正触发中断；随后发现取消丢弃连接导致`:memory:`视图丢失，改用临时文件确认重连。HTTP夹具改用9001避免迁移默认group1。原失败日志保留，均未修改生产去重/迁移/连接代码。

浏览器使用新建临时实例及合成账号/用量，实际二进制/登录/API/SQLite/构建SPA；测试初始化合成管理员gate状态，不提交协议接受操作。首次导航早于启动就绪、首次运维请求遇到gate423；初始化夹具后实际成功。一次全屏点击与已有分组URL防抖冲突，等待URL稳定后复测正常；未扩展修改现有路由防抖。移动截图与DOM断言确认新卡片无溢出。

成本测量来自本机i9-13900H，非VPS承诺；补充EXPLAIN使用PythonSQLite，Go/modernc才是真实查询性能和取消证据。统计只反映仍保留的用量明细，大窗口自身超时显示不可用并保留概览。
