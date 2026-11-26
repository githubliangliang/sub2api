# Context

上一批F01/S01于89060f532完成11任务，当前旧17份活动变更467项完成；本批从该提交实施并完成13/13任务，实际证据见verification。

# Decisions

F01按实际映射目标选择生成模板，仅处理非GPT自我身份片段；不全局清洗用户instructions，保留其它工作指令和占位符。F02把沿用旧secret也送入既有固定key/加密持久化边界；返回仍脱敏，不新增备份操作。F03让number属性按字符串发送，说明/placeholder清空明确发送空串，缺省更新仍由原后端指针合同解释，不增加小数支持。F03与第二档S01共用属性组件，组合时按最终合同处理。

# Resource and Data Boundaries

保持SQLite-only/miniredis/simple、固定quota、usage dedup、独立用户费用/账号成本、Fast及requested/effective日志。请求体64MiB、SSE行40MiB、暂存8MiB、idle120s及共享图像原大小限制保留。

# Verification

当前证据只证明对应基线、失败和正常对照，源代码检查不证明未来实现通过；实施另记录真实起点、完整失败与通过、实际HTTP/SQLite或浏览器、最终哈希和范围。完成后才勾任务。
