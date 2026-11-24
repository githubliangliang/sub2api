# Context

基线 v1.1.27 已包含 v0.2.24 两档，SQLite230、auth snapshot21；13 份旧活动 change / 439 项任务已完成，保持原文等待另行归档。

# Decisions

## S01 透传结构化错误、规则与 compact 终态一致性

ErrorPassthroughService.MatchRule、runtime wrapper、compact keepalive/bridge、容量可重试码均在。需完整提取匹配规则、有限标量字段和 compact writer。保留本地 NextOpenAIResponsesSequence，不能照抄来源 sequence_number=0。

F01 为脱敏前置；保留既有 failover/池重试和成功流语义。状态码匹配的 skip_monitoring 已在事件链生效；容量码重写是新增结构化 compact 路径的配套保护，不把它们整体列成当前缺陷。

验收：HTTP/compact 心跳前后，规则优先级/平台/原状态与原 body、passthrough_body、自定义状态/文案、无规则、错误 body 标量类型/256 字节/内嵌 JSON/空白 code；监控/账号副作用、终态序号、取消/停止/无重复写、失败不计成功用量和 SQLite 结算去重控制。

## S02 Gorilla WebSocket 定向安全依赖维护

官方固定修复为 v1.5.4-0.20240701034025-d67f41855da4；现有 Go1.27.4/x/net0.62.0 足够。临时 modfile MVS 仅变化这一模块，但模块内容含 18 文件差异及代理/握手行为改变。

独立摘取混合 PR 的依赖子项，保持 Go/VERSION/schema/前端依赖；不随带名称同步。两项依赖包代理测试在旧/候选版本均 EOF，真实 CONNECT 对照通过仍不能冒称依赖全量绿灯或已验证利用。

验收：核对官方漏洞范围、模块来源/校验和/MVS，复核完整模块差异；实际管理员 WS 鉴权/Origin/限额/控制帧/关闭，HTTP CONNECT 与 Basic-auth 隧道、现有网关 HTTP/WS 及 targeted race、build；保留并定因依赖包历史代理测试失败。

# Resource and Data Boundaries

保持请求体预算64MiB、SSE行40MiB、暂存8MiB、读写idle120s。错误净化发生在错误路径；先限制可输出字段、复用有界错误体，不扩展成功请求的拷贝/缓存。保持固定配额、独立账号成本与用户费用、usage dedup、Fast 和 requested/effective effort。无需新迁移/Ent/Wire。

# Verification

评估证据只证明当前缺口与来源可行性。实施应记录实际 HEAD、最终文件哈希、所有失败/通过和测试范围；不以 source helper、临时 modfile 或旧发布验收冒充新实现。只对实际改动面增加测试。
