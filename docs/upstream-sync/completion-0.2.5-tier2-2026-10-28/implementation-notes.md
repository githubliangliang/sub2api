# 0.2.5 第二档余项实现

用户要求完成 `port-upstream-0.2.5-tier2` 的未完成任务，本次按该指令执行其四项余项：容量默认值决策、S12b 批准记录、配额兼容实现、迁移 227。指令原文和范围解读见[start-state.json](./start-state.json)。迁移仅写入仓库并在临时 SQLite 验证，没有操作业务数据库。

实现提交为 `b349c7408325a0e4eba4f32268ad7a3820e2ad4f`，VERSION 为 1.1.20、Go 为 1.27.3，最大迁移升为 227。原 source-baseline/source-feature-map、候选和四态 TSV 逐字保留；实施前[任务](./tasks-before.md)与[验收](./verification-before.md)另存快照。

## S02 默认容量决策

保留 OAuth/API Key 默认系数 1.0，不采用 #7043 的数值 5.0。没有访问业务实例，活跃 ctx_pool 会话数、实际账号并发和资源余量未测量；这些未知值不能作为扩大 1C1G 默认连接预算的依据。现有 config/legacy/v2 容量、硬上限、零并发和保留会话获取测试通过。并发 3、硬上限 20 下，1.0→3、显式 1.5→5、5.0→15；账号在飞请求槽没有放大。[决策记录](./ws-default-decision.json)完成了决策任务，不将数值 5.0 方案记作已移植。

## S12b 配额记录

- 只有至少一档非 NULL 才保存，0 也是配置。注册默认值和管理员输入分别过滤空记录；初始化仅补 NULL 档，不覆盖已有个性化值。
- 清空限额使活跃行软删除；之后配置新建行，用量为 0、窗口为空。修改仍配置的行保留其用量与窗口；审计仍记录移除，并区分提交数与保存数。
- 增量和快照更新不插入或复活缺失/已删记录。SQLite 增量先取得写锁再读取，避免多连接读取同一用量快照；没有使用不支持的 FOR UPDATE。快照使用 WITH/VALUES/UPDATE FROM，每条最多 1000 行/8001 绑定参数，所有分批共用一个事务；6000 条 flusher 输入仍可原子处理。
- 历史迁移 142 的日/周窗口是 TEXT，现代驱动不会自动扫描为 time.Time。本次复用既有类型无关时间解析器，兼容 TEXT 和 DATETIME；没有编辑历史迁移或 Ent schema。
- 仪表盘卡片取“有用量 ∪ 有限额配置”的并集，忽略空配额占位，计数排除“其他”。未保存配置的平台不能重置；已保存零限额仍可重置，编辑中的未保存值不改变该判断。

真实 SQLite 回归覆盖空默认值、零限额、只补 NULL、清空/再次配置、缺失行写入、4100 条快照、后段失败整批回滚、历史 TEXT 窗口和 12 个并发累计。一次测试使用 Ent 默认查询统计软删除行，因默认过滤得到 0，修正为 SkipSoftDelete 后检查真实历史行数量；没有放松“不复活”断言。前端原先允许空配置重置的测试，改为保存有限额的明确夹具并补充缺失/未保存/零限额的反例。

## 迁移 227

`227_purge_unlimited_user_platform_quotas.sql` 仅 DELETE 日/周/月 limit 全为 NULL 的行；含 0 或其它非 NULL 限额的活跃与历史软删行保留。没有调整账单、token/cost 记录或平台枚举。

迁移包的实际 SQL 测试覆盖三种零限额、有限额、历史软删、用量保留与幂等。repository 测试通过全部旧迁移创建真实 226 数据库，插入配置/空行，再由 shipped ApplyMigrations 升到 227，核对 schema_migrations 并重复执行。[升级记录](./quota-migration-upgrade.log)与[配额 SQLite 回归](./quota-sqlite-complete.log)均通过。

所指向的历史 `_integration_test.go` / `_upsert_test.go` 首行是 `integration && postgres`；保持其构建标签，没有用这些不会在本 fork 编译的文件冒充验收。新增的是正常可运行的 SQLite 路径，既有用户配额 service/handler/cache 回归也通过。

## 最终门禁

后端 54 个有测试包 unit/build、迁移、repository/service/admin quota race、Go lint 0 issues 通过。前端 287 文件/2095 passed/2 既有 skipped，typecheck/lint/build 和关键套件 176 passed/2 skipped 通过。测试源补齐后另执行迁移升级与最终 lint；已实现产品文件哈希与提交逐项一致。[收尾报告](./completion-report.json)保存命令、最终输出和哈希，失败阶段日志保留原结果。

这是现有个人配额和默认值决策，未引入支付功能、第三/第四档平台能力、#7237/#7289 的额外范围或依赖升级。原本已经落地的其它簇不重复移植，历史 VERSION 1.1.12/最大迁移 226 的实施证据仍按当时值解释。
