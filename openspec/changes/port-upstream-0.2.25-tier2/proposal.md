# Why

既有 OpenAI 透传路径丢失客户端恢复所需的标准错误字段、忽略部分规则结果，compact 已提交后也丢失确定性 code。依赖清单另包含官方已确认受影响的 Gorilla 版本，需要有边界的安全维护。

# What Changes

- S01 透传结构化错误、规则与 compact 终态一致性：OpenAI 透传最终错误 SHALL 应用既有管理员规则，并在允许的确定性 400 或 passthrough_body 规则下保留经过脱敏和限长的标准标量字段；compact 已提交心跳后 SHALL 仅输出有序 response.failed 终态，保留确定性 code 并将容量降载 code 改为可重试值。
- S02 Gorilla WebSocket 定向安全依赖维护：定向维护 SHALL 使用官方修复的 Gorilla mask 随机源并保持现有管理员 WS 鉴权、单一 subprotocol、连接释放及网关互通；最终模块选择与未解决的依赖测试差异 SHALL 可核验。

# Impact

S01 涉及规则 runtime、错误信封与 compact 生命周期；S02 仅允许 Gorilla 精确版本维护及必要 go.sum。临时模块的编译/测试不是实施验收。 保持 SQLite230/auth21、VERSION1.1.27、Go1.27.4 和既有资源默认。

# Non-goals

身份变量、WS pool 身份分桶、上游名称备注同步、include=group、group i18n、新平台/模型、支付、发布/push/归档和业务实例操作均不在本批范围。
