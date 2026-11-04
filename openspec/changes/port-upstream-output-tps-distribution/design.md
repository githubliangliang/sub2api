# Design

## Context
现有 usage_logs 具有输出 Token、总耗时、请求类型及图片字段。概览已有 raw/auto/preagg；小时延迟分位数可近似组合，但输出 TPS 需要全窗口明细分布。

## Decisions
1. 每条样本速率为 output_tokens × 1000 / duration_ms，包含首字等待，不重复累加推理 Token。只接受正输出/正耗时、sync/stream/ws_v2/cyber，排除图片计数/图片 Token/图片计费和 Live/unknown。
2. 复用现有 [start,end) 时间、分组及平台筛选；平台优先分组、无分组回退账号，与概览其它字段保持一致。
3. 使用 SQLite ROW_NUMBER/COUNT 排序计算秩，按 (n−1)×p 在相邻样本之间线性插值，与来源 percentile_cont 一致。单次查询返回固定大小结果，不把全部样本读进 Go，也不平均小时分位数。
4. raw/auto/preagg 在基础概览成功后统一查询 retained usage_logs。清理过的明细不会从小时聚合重建；提示说明样本来自保留明细。
5. 查询复用2秒预算；统计子查询自身超时返回 output_tps=null并保留概览，父请求取消/截止和其它数据库错误传播。无样本返回四个null和 sample_count=0。
6. 前端单独显示 P50 主值、P5/P10/平均与样本数，使用 tok/s、保留一位小数。缺失值显示 —，有效小值四舍五入为0.0；复用已有筛选与请求生命周期。

## Verification
真实 SQLite 连续插值、单/空/重复样本、不同耗时、全部排除条件、时间/平台/分组与小时跨界、预聚合及清理后语义；执行中取消/截止和概览降级；实际管理端 HTTP、组件更新/全屏/无样本及浏览器桌面/移动端；查询规模耗时和既有门禁。
