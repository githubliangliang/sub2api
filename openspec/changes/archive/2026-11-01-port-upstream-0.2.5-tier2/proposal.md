# 提案

本 fork 当前存在一批「已定位到本地调用点、但需要真实工作量」的行为缺口：Responses 流只在
done/terminal 事件里带的正文会整段丢失；chat 桥把 Codex 的 instructions 与 developer 项
展开成两条前导 system 并在会话中途再插 system，触发上游 400；WS 连接池容量不吃
`max_conns_factor`（本仓库默认仍为 1.0）且没有常驻读循环应答上游 ping；Antigravity 在
内置工具与客户端函数工具混用时被 v1internal 拒绝；DeepSeek 空映射放过未知模型，换来上游
404 与 30 分钟的 per-(account,model) 冷却；客户端 `system` 上的 `cache_control` 断点被剥离，
削弱提示缓存命中；Codex 配额重置锚在 `now` 而不是快照时刻，导致重置时间每轮评分向后滑动。

本 change 按 S01–S12 中现行11簇修正这些行为（S07已迁第一档T19）。
原28个PR中7个前移后，本批剩21个；两档仍为原43个PR，第三/第四档不扩入。所有新测试须有明确触发条件和正常路径反例；调度、计费口径、
SQLite 用量写入路径与既有 namespace/Responses Lite 处理保持既有契约。

成功标准：现行簇（不含已迁出的S07）与 S12a 规格满足，目标回归、后端 build/lint、`go test -tags=unit ./...`、SQLite 方言审计、
前端 typecheck/lint 与关键 vitest 通过，来源与落地 SHA 可追溯。S12b 获批后单独验收；
未获批时明确标记待实施，不将其计为通过，也不宣称现行21项全部完成。

不包含：第三档的新平台与新模型（OpenCode、MiniMax、Gemini 3.7/3.8、DeepSeek V4.1-Flash 计费）、
第四档全部条目、Fast 发送侧语义、动态长上下文阶梯、DOMPurify/xlsx 依赖升级。
S12b 在获得用户对迁移 `227` 的明确批准前不实施。

## Capabilities

- `single-node-stability`：保留原S编号的现行11簇行为契约，以本 change 的 delta spec 为准。
  仍限于冻结来源；S07已迁出，S11仅3项，output-config beta归S08。
- 先交付正文/协议/资源/调度正确性，再做展示和体验；#7043默认值调优与#7064容量计算分开验收。
