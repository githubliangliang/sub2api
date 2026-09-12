# 提案

长 Codex 会话当前重复复制回放正文；客户端取消、失败会话槽和持久冷却与本地 block 的关系仍有缺口。
这些问题对 1C1G SQLite 部署尤其值得处理。同时，Astra 能力同步、工具结果转换、Claude 版本身份，
以及代理局部更新和账号列表存在需要多文件配合的修复。

本 change 覆盖 PORTING 的 S01-S12：WS 回放、取消、冷却、会话槽、工具/bootstrap、Astra、
价格热重载/none 映射、代理、日志、管理界面、Claude CLI 身份和 go-redis 客户端。
每簇独立实施和验收，S01/S03 不凭 apply 干净宣称低风险。

成功标准是对应行为在本 fork 的实际入口可达，SQLite/miniredis 和已有用户 metadata 修复保持可用，
测试覆盖并发/取消/失败路径，前端交互和依赖变动有相应门禁。
本 change 不新增平台、Fast 发送策略、分组 allowlist、固定账号清单、支付功能或 PG/multi-instance 架构。
