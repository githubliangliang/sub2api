# Context

上一批F01/S01于89060f532完成11任务，当前旧17份活动变更467项完成；本批从该提交实施并完成14/14任务，实际证据见verification。

# Decisions

S01针对show/userId和读取/保存会话使用独立归属，切换、关闭或重试使旧回调失效；保存绑定当前已加载对象与快照，不引入来源restrictPublicGroups字段。既有角色提升的step-up错误、验证和重试也属于该会话；实际回归要求旧会话在prompt/重试前取消，编辑器按会话挂载TOTP子对话框，子对话框卸载后忽略晚成功/失败，不改composable API。S02取消并隔离旧FileReader，确保旧load/error回调和删除/卸载不恢复旧值；保留SVG净化及文件大小/类型边界。S03按templateId/会话隔离关联列表及选中ID，提交后旧完成不能关闭新会话；来源只覆盖读取的部分需按本地提交链适配。

# Resource and Data Boundaries

保持SQLite-only/miniredis/simple、固定quota、usage dedup、独立用户费用/账号成本、Fast及requested/effective日志。请求体64MiB、SSE行40MiB、暂存8MiB、idle120s及共享图像原大小限制保留。

# Verification

当前证据只证明对应基线、失败和正常对照，源代码检查不证明未来实现通过；实施另记录真实起点、完整失败与通过、实际HTTP/SQLite或浏览器、最终哈希和范围。完成后才勾任务。
