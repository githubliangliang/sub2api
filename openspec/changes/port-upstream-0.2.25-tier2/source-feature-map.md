# Source Feature Map

| 簇 | 完整来源 | 行为与边界 |
|---|---|---|
| S01 透传结构化错误、规则与 compact 终态一致性 | [#9045](https://github.com/Wei-Shaw/sub2api/pull/9045) | OpenAI 透传最终错误 SHALL 应用既有管理员规则，并在允许的确定性 400 或 passthrough_body 规则下保留经过脱敏和限长的标准标量字段；compact 已提交心跳后 SHALL 仅输出有序 response.failed 终态，保留确定性 code 并将容量降载 code 改为可重试值。 F01 为脱敏前置；保留既有 failover/池重试和成功流语义。状态码匹配的 skip_monitoring 已在事件链生效；容量码重写是新增结构化 compact 路径的配套保护，不把它们整体列成当前缺陷。 |
| S02 Gorilla WebSocket 定向安全依赖维护 | [#9008](https://github.com/Wei-Shaw/sub2api/pull/9008) | 定向维护 SHALL 使用官方修复的 Gorilla mask 随机源并保持现有管理员 WS 鉴权、单一 subprotocol、连接释放及网关互通；最终模块选择与未解决的依赖测试差异 SHALL 可核验。 独立摘取混合 PR 的依赖子项，保持 Go/VERSION/schema/前端依赖；不随带名称同步。两项依赖包代理测试在旧/候选版本均 EOF，真实 CONNECT 对照通过仍不能冒称依赖全量绿灯或已验证利用。 |

patch sites 只在[PORTING 报告](../../../docs/upstream-sync/PORTING-0.2.25.md)。新增调用清单是检索索引，不证明编译。来源测试和本地未来验收分开。
