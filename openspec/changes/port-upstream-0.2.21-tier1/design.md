# Design

## Decisions

按行为适配，不merge上游main。基线已实现完整Fast、Terra、v0.2.20两档、SQLite后台服务及固定配额窗口，不重复导入旧基座。真实业务数据不参与评估与验收夹具。

### F01 Codex 元数据与 failover

metadata overlay 已复现4个子例失败，普通 metadata 两个负对照通过。

普通 response.metadata 对照通过；真实文本、工具输出及终态按既有语义发送，不把所有 codex.* 任意当作可丢弃内容。

### F02 Responses 密封参数与 namespace

sealed overlay 24子例中delta fallback 8失败，其余16通过；namespace单独也丢失。

现有 terminal 与 item.done 原文路径已经通过；只补 delta accumulator 缺口，不重复重构已实现的终态保真。

### F03 CC Switch reasoning 封装

foreign reasoning overlay 失败，原生签名负对照通过。

文本、tool_use、tool_result、签名顺序不变；不把未知所有签名一概删除。

### F04 Responses 工具回合会话种子

本地后续developer提示被拼入种子；SessionPrefix overlay 确认字符串变化。

首个系统/开发者提示、模型、工具、instructions、首用户差异仍隔离；显式 session-id/prompt_cache_key 优先。

### F05 账号补全失败释放并发槽

失败overlay实际release=0，成功对照通过；本地签名适配仅在探针中。

本地 newSelectionResult 多一个 sessionID 参数且已有 ReleaseAccountSession；不可用上游函数整体覆盖。

### F06 价格目录 hash 与正文一致

PricingHash overlay下载次数实际1而期望2；remote stub不接触外部价格服务。

远端hash缺失/大小写、无效JSON、网络失败、fallback/override及热重载并发保留，失败不发布半成品。

### F07 Sonnet 5.5 缓存读兜底价

本地实际0.20/百万；SonnetPrice overlay失败。本结论是与冻结tag一致性，非独立市场价格审计。

显式自定义0与覆盖价优先；其它模型和输出/缓存创建费率不改。

### F08 Gemini 非流式正文排除思考摘要

GeminiThought overlay复现正文含private planning；前置已逐行量测。

本地缺 isGeminiThoughtPart，但来源helper仅4行布尔判断；只补这一最小前置，不导入320行完整thinking策略。

### F09 Antigravity 并行工具图片顺序

ParallelToolImages sideband失败、native对照通过。

支持原生 functionResponse.parts 的模型继续嵌入媒体；单工具及无图片顺序不回退。

### F10 Codex 图片请求保留错误状态码

ImageStatus overlay期望400实际200；是现有心跳消费者缺陷，不依赖新模型。

其它客户端保留既有心跳，流式不改；该header不是认证。只维护现有图片入口，不引入2.6或新原生路由。

### F11 Responses UTF-8 输入边界

9个非法UTF8诊断子例失败，有效Unicode对照通过；较上一轮PG论据改为独立协议输入边界。

本次探针仅证明非法字节进入model字段检查，不证明曾产生真实上游503或PG故障；保留大小/预算释放和鉴权次序。

## Validation and rollout

先复现各簇旧行为，再独立实现并执行真实消费者验收；相关错误、取消、重试、usage和SQLite去重一起验证。最终全量unit、构建、lint和定向race。前端变更需typecheck/lint/test/build/critical及本地浏览器；没有前端变更则记录不适用。资源或索引变更需测峰值/查询/写入成本。发布与部署需独立指令。
