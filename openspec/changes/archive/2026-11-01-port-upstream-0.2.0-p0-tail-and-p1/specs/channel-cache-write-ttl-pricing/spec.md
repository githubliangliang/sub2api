## ADDED Requirements

### Requirement: 渠道自定义定价可分别指定 5m 与 1h 缓存写入价
渠道自定义定价 SHALL 新增 `cache_write_1h_price`，覆盖 1h 缓存写入档。该列 SHALL 同时存在于
按模型定价、定价区间、账号统计按模型定价、账号统计定价区间四张表。配置了该值时，
1h 档 SHALL 按该值计价，且该条定价 SHALL 声明支持缓存明细拆分。

#### Scenario: 同时配置 5m 与 1h
- **WHEN** 渠道同时配了 `cache_write_price` 与 `cache_write_1h_price`
- **THEN** 5m 档 MUST 按前者、1h 档 MUST 按后者计价

#### Scenario: 区间定价
- **WHEN** 某定价区间单独配了 `cache_write_1h_price`
- **THEN** 该区间内 1h 档 MUST 按区间值计价

#### Scenario: 账号统计侧
- **WHEN** 账号统计定价配了该列
- **THEN** 账号统计成本 MUST 按同一规则计算

### Requirement: 未配置 1h 价时必须保持拆分前语义
`cache_write_1h_price` 为 NULL 时，单独的 `cache_write_price` SHALL **继续同时覆盖 5m 与 1h 两档**。
本变更 MUST NOT 让存量只配了 `cache_write_price` 的渠道在 1h 档上回落到官方价。

#### Scenario: 只配了 cache_write_price
- **WHEN** 渠道只配了 `cache_write_price`，`cache_write_1h_price` 为 NULL
- **THEN** 1h 档 MUST 按 `cache_write_price` 计价（与本变更前完全一致）

#### Scenario: 两个都没配
- **WHEN** 两列都为 NULL
- **THEN** 两档 MUST 都回落到官方/兜底价（与本变更前一致）

#### Scenario: 只配了 1h 价
- **WHEN** 只配了 `cache_write_1h_price`
- **THEN** 1h 档 MUST 按该值计价
- **THEN** 5m 档 MUST 回落到官方/兜底价

### Requirement: 迁移必须是 SQLite 方言且不可再改
新增迁移 SHALL 使用本仓库号 `226`，SHALL 只含裸 `ALTER TABLE ... ADD COLUMN`，
MUST NOT 使用列级 `IF NOT EXISTS`，MUST NOT 含 `COMMENT ON`。文件注释中 MUST NOT 出现 PG 语法字面量
（`sqlite_dialect_audit_test` 连注释一起扫）。

#### Scenario: 在空库上跑迁移
- **WHEN** 删除本地 `*.db` 后完整启动
- **THEN** 迁移 `226` MUST 成功执行
- **THEN** 四张表 MUST 都出现新列

#### Scenario: 在已有库上跑迁移
- **WHEN** 在已应用到 `225` 的库上启动
- **THEN** 迁移 `226` MUST 成功执行且不影响既有数据

#### Scenario: 方言审计
- **WHEN** 运行 SQLite 方言审计
- **THEN** MUST 通过（含对注释的扫描）

### Requirement: 管理面与展示面必须暴露新列
admin 渠道 API 的请求与响应、可用渠道接口、前端渠道定价表单与模型广场定价表 SHALL 携带
`cache_write_1h_price`。

#### Scenario: 管理端保存渠道定价
- **WHEN** 在渠道定价表单里填入 1h 缓存写入价并保存
- **THEN** 该值 MUST 落库并在重新读取时回显

#### Scenario: 模型广场展示
- **WHEN** 某模型有 1h 缓存写入价
- **THEN** 定价表 MUST 展示该档

#### Scenario: 留空
- **WHEN** 表单里该字段留空
- **THEN** MUST 落库为 NULL（MUST NOT 写成 0）
