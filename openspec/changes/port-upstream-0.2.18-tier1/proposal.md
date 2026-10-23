# 第一档现有功能修复提案

## Why

当前fork保留了上游已修复的既有协议/结算边界。目标是修正真实消费者的行为，优先保障Claude Code、Codex、Grok，并保留个人单节点约束。

## What Changes

- F01 Codex 工作区导入身份隔离：以 workspace/account 与 user 的组合区分完整 OAuth 身份；同一工作区重复导入更新原账号，不覆盖其它工作区。
- F02 Codex 前导元数据不阻止重试：排队和审查元数据保持首个语义输出之前的状态，后续容量错误仍可安全重试，首输出超时仍有效。
- F03 Grok Fast 用量成本价卡：既有 Grok 模型映射/透传请求按 Fast 独立价卡计算；短上下文 input/cache/output 每百万为 4/1/12，达到 200k 后为 6/1.5/18（来源补丁值）。
- F04 Cloudflare 1010 不处罚账号：识别明确的 Cloudflare 1010 或 browser_signature_banned，所有既有平台跳过账号处罚，保留请求错误和换号处理。
- F05 Gemini 缓存命中用量非负：仅缓存计数或缓存大于 prompt 时非缓存 input_tokens 下限为0，同时保留缓存、输出和图片用量。
- F06 Gemini 工具 Schema 常量和数据边界：字符串const转为单值enum，与已有enum求交；移除不支持的非字符串const和注解，区分Schema节点、属性名和数据值。

## Impact

小范围后端行为和价格资源修复，不需要迁移或依赖更新。

## Non-goals

不新增平台、Argon、Gemini思考能力、Web Search模拟、Vertex额度功能、生图开关适用范围或共享调度缓存；不引入任何支付代码。评估不是实施授权，历史source-baseline不改。
