# Why

上游v0.2.10–v0.2.11涉及本fork现有请求路径与客户端指引。第二档按6簇移植能改善个人网关正确性与兼容性。

## What Changes

- S01 Sonnet 5.5多入口兼容：Sonnet 5.5目录、定价、effort、thinking签名与工具beta在本地已有入口中一致。
- S02 GPT-6.1 Sol及账号套餐元数据：现有网关路径识别GPT-6.1 Sol并一致执行请求能力、模型目录与价格规则，账号套餐显示保留SKU身份。
- S03 流式用量输出与计费一致：现有Anthropic到Responses/Chat桥在收到真实usage时转发，并将权威prompt总量与缓存桶统一到计费口径。
- S04 Antigravity首内容前保活边界：兼容流等待首内容时15秒发送注释保活，并在2分钟无语义内容时终止；已提交响应后的失败通过单次SSE错误报告。
- S05 Composite模型归属与多轮WS路由：OpenAI两种调度执行公开alias归属检查，WS按responses/any解析路由并隔离公开模型与出站映射。
- S06 Codex远程目录与文件兼容：提供Codex 0.156.0+远程目录和旧客户端本地文件选择，采用v0.2.11最终行为。

## Impact

详细patch sites只放在[PORTING](../../../docs/upstream-sync/PORTING-0.2.11.md)；本change定义行为与验收。保留SQLite/miniredis/simple mode、现有菜单和自有1.1.x版本。无需新迁移或依赖升级。支付/充值/购买/续费及第三/四档功能不引入。

本次只交付规格；基线unit/build和前端typecheck/97项测试已通过，新增行为尚无实施验收。
