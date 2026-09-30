# Why

上游v0.2.10–v0.2.11涉及本fork现有请求路径与客户端指引。第一档按4簇移植能改善个人网关正确性与兼容性。

## What Changes

- F01 多工具改名保持请求体完整：工具声明、强制工具选择和历史tool_use中的同名工具保持一致，改名不破坏JSON或无关内容。
- F02 白名单编辑保留显式模型映射：在新增白名单项会覆盖from到不同to的映射时提示并拒绝该新增，保留已有映射。
- F03 兼容入口遵守Claude Code降级分组：Claude Code限制分组配置fallback时，Chat Completions和Responses请求沿现有调度选择降级分组账号。
- F04 密钥使用指引匹配客户端限制：受Claude Code限制的分组只显示Claude客户端标签，分组限制改变时重置当前客户端选择。

## Impact

详细patch sites只放在[PORTING](../../../docs/upstream-sync/PORTING-0.2.11.md)；本change定义行为与验收。保留SQLite/miniredis/simple mode、现有菜单和自有1.1.x版本。无需新迁移或依赖升级。支付/充值/购买/续费及第三/四档功能不引入。

本次只交付规格；基线unit/build和前端typecheck/97项测试已通过，新增行为尚无实施验收。
