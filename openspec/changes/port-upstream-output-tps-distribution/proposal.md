# Change: 补齐窗口级输出 TPS 分布

## Why
已有日志输出 TPS 和聚合 Token 吞吐不能回答一个时间窗口里较慢请求和典型请求的输出速度。用户已明确选定历史 PR #7911 的运维统计余项。

## What Changes
- 运维概览提供单请求输出速率的 P5、P10、P50、算术平均值和有效样本数。
- 使用真实 SQLite 窗口查询进行连续分位数插值，所有查询模式保持相同口径。
- 运维页面在普通/全屏模式显示统计，并保留筛选、刷新与缺失数据语义。

## Impact
- Affected specs: ops-output-tps-distribution
- Affected behavior: 管理端运维概览的附加统计与展示。
- 保持 SQLite-only、可选 miniredis、simple mode、用量落库/去重；不修改 schema、依赖、版本、支付或计费写入。
