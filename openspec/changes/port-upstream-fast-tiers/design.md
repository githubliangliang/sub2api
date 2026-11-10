# Design

1. 统一 HTTP/Chat/Responses/现有 WS 的档位；fast 规范为 priority，未知枚举明确拒绝。默认不强制加速。
2. 分组 Force Fast 优先写入 priority，再执行全局策略；missing 规则单独控制省略字段。Free Fast 只影响 OpenAI priority 用户成本，保留上游成本与档位，不适用于 ultrafast。
3. 新 SQLite 迁移持久化分组开关及渠道 Fast 倍率，贯通实体、缓存、API 和 UI。平台切换清理不适用配置。
4. 目录保留上游显式能力，OAuth 权限不由套餐推断。
5. 档位计价适配保留当前长上下文规则、显式零价格和覆盖价；不引入 effort/动态阶梯/峰谷等相邻机制。Claude speed 与 OpenAI service_tier 分开处理，保留本地原生日志修复。
