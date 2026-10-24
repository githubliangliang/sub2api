# Proposal: Port upstream v0.2.5 第一档 fixes

## Why

本 fork 在 `fe6bc318800ca86d2b658f0e8f46e83d32626e8b`（v1.1.12）仍保留一组上游 v0.2.5 已修复的行为：
共享 project_id 的 Antigravity token 缓存串号、Gemini SSE 重复空行、错误的批量生图优先级、
调度投影字段缺失，以及瞬时认证故障导致前端登出等。其余问题涉及配置错误状态码、UA 校验与管理界面交互。

这是一批有明确触发条件的修复，并不声明所有问题已在生产复现。
本次仅补全文档并调整既有第一/第二档任务顺序；产品实现与验收均待后续执行。
两档总范围仍为原43个 PR，第三/第四档不进入实施。

## What Changes

按 [implementation-map.md](./implementation-map.md) 的 T01–T23 手工移植最小行为补丁：

- T01–T06：账号级 token 缓存与旧键清理、SSE 分隔、规范请求头、升序选号、调度元数据投影、Codex UA 校验。
- T07–T09：JWT/管理员查询失败区分 401 与 500；前端刷新遇到网络/429/5xx 时保留会话；非法峰谷配置返回 400。
- T10–T12：用量成本浮层 8 位精度、续费套餐列表滚动、EasyPay 自定义 upstreamType 允许点号。
- T13–T16：邮箱验证竞态按身份移除、代理筛选重置分页、订阅分配选择及时失效、兑换时长上限 36500。

- T17–T23：承接原第二档的兑换结果保护、CSV 参数快照、agent_message 桥接、密钥状态同步、自动刷新间隔、资料错误提示和部分导入刷新。

22 个来源 PR 中 #7052 拆成 T07/T08。以整 PR diff 理解完整语义，以本地调用链决定实际 hunks；
不覆盖整文件，不机械移植依赖上游 fixture 的测试。

## Impact

- 后端涉及 service、repository、middleware、pkg/openai；前端涉及 API client、管理/用户组件和中英支付配置文案。
- T07 的 API Key 路径仅作兼容性回归：查询错误已有 500、过载已有 503；缺失关联用户仍为 401。
- 无新增 API、配置项、表结构或依赖。HTTP 可观察变化限于区分查询故障和校验失败，响应码详见行为契约。
- 保留本 fork 的 SQLite-only、可选 Redis、认证并发刷新、调度与支付自有逻辑。
- 冻结来源不修改；原始评估与现行代码的差异见 implementation-map.md。

## Non-goals

- 第二档及以后移植、WS/provider 扩展、迁移、Ent/Wire 生成、VERSION 或依赖升级。
- PostgreSQL 专用集成测试、无关 refund 测试格式整理、跨档 PORTING 总状态维护。
- 扩展刷新重试策略或重写 tokenRefresh 协调机制；T08 只新增瞬时错误保留会话分支。
- 改写计费算法或统一全站数字格式；T10 仅对齐成本浮层的 8 位展示。
- 本次文档补全不执行产品修复、部署、提交或发布。

## Success Criteria

文档交付：23 项行为均能追溯到来源、实际产品路径、任务与验收；历史快照保留，差异有勘误；
待执行与已执行证据明确区分，无依赖缺失总报告的悬空引用。

产品交付：每项行为满足 spec 和验收矩阵；后端 build/unit/SQLite 方言审计、前端 typecheck/lint、
关键套件及本批定向回归通过，真实浏览器确认续费列表可达；无迁移、生成代码、版本、依赖或 lockfile 越界。
命令、工作目录与通过标准统一见 [verification.md](./verification.md)。
