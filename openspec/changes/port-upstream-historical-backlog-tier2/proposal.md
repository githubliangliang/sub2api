# Why

当前native JSON/SSE对HTTP200内的Google503错误信封原样写回，却没有Ops错误标记；现有中间件无法发现该语义失败。当前真实消费者及后续已具备的Ops归因合同使旧第三档明确调整为第二档。

# What Changes

识别现有native JSON/SSE/聚合路径的错误信封、请求级策略结果与真实空响应，接入已有Ops日志/SLA/skip合同。保持客户端状态和原始字节、已观察usage及既有计费；不新增重试、账号处罚、quota策略或新平台。

# Impact

Gemini原生2xx响应的带内信号运维归因，仅改现有消费者。保持VERSION1.1.28、Go1.27.4、SQLite230/auth21、依赖与资源默认。

# Non-goals

其它按需功能、支付/购买/续费/充值/退款、发布/push/归档和业务实例操作不在本批范围。
