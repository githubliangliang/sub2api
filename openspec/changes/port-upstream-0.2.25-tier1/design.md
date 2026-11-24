# Context

基线 v1.1.27 已包含 v0.2.24 两档，SQLite230、auth snapshot21；13 份旧活动 change / 439 项任务已完成，保持原文等待另行归档。

# Decisions

## F01 透传上下文错误的账号凭据与主机名脱敏

sanitizeUpstreamErrorMessage、Account.GetCredential、现有错误信封均在；新增账号感知脱敏和安全大小写替换即可，无新供应商或依赖。

只修改面向客户端的错误文本，不改变管理员内部取证、账号惩罚、换号、模型/计费或响应头净化；不引入上游中间提交中以 ToLower 下标切原串的错误实现。

验收：真实本地 HTTP 上游返回上下文超限；API Key/access_token/refresh_token、主机名大小写、IPv6/无 base_url、Ⱥ/Kelvin/非法 UTF-8、query 凭据、正则元字符、无匹配和普通 403 对照；不得 panic 或多写响应。

# Resource and Data Boundaries

保持请求体预算64MiB、SSE行40MiB、暂存8MiB、读写idle120s。错误净化发生在错误路径；先限制可输出字段、复用有界错误体，不扩展成功请求的拷贝/缓存。保持固定配额、独立账号成本与用户费用、usage dedup、Fast 和 requested/effective effort。无需新迁移/Ent/Wire。

# Verification

评估证据只证明当前缺口与来源可行性。实施应记录实际 HEAD、最终文件哈希、所有失败/通过和测试范围；不以 source helper、临时 modfile 或旧发布验收冒充新实现。只对实际改动面增加测试。
