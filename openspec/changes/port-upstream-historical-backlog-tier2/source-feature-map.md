# Source Feature Map

| 簇 | 完整来源 | 选定行为与边界 |
|---|---|---|
| S01 Gemini原生2xx响应的带内信号运维归因 | 上游PR#6929，SHA见source-baseline | 识别现有native JSON/SSE/聚合路径的错误信封、请求级策略结果与真实空响应，接入已有Ops日志/SLA/skip合同。保持客户端状态和原始字节、已观察usage及既有计费；不新增重试、账号处罚、quota策略或新平台。 |

patch sites只在[评估报告](../../../docs/upstream-sync/UNMERGED-REVIEW-2026-11-25.md)。原第三档调整的证据与范围见报告，不将其它可选新增功能升档。
