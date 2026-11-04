# 来源与行为映射

| 来源 | 选定行为 | 本地适配 |
|---|---|---|
| PR #7911 service/models | 可空 output_tps 与四项速率/样本数 | 保持概览 API 的附加字段 |
| PR #7911 repository | 整窗口逐请求连续分位数、筛选、2秒查询预算 | SQLite 排序/窗口函数/线性插值，复用现有筛选 |
| PR #7911 timeout tests | 自身统计超时保留概览，其它错误传播 | 增加 SQLite 执行中断和调用方取消/截止验证 |
| PR #7911 API/header/i18n | 普通/全屏卡片、缺失/刷新、统计口径说明 | 接入当前真实运维页面，中英文文案 |
| PR #7911 integration tests | 样本筛选、完整窗口、raw/auto/preagg | 使用迁移后的真实 SQLite；不依赖 PostgreSQL |
| PR #7911 usage/latency helper | 已有日志输出 TPS | 保留，定向回归验证一致口径 |
