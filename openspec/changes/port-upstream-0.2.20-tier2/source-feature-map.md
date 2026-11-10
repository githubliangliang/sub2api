# Source Feature Map

| 簇 | 整PR来源 | 纳入行为 | 验收边界 |
|---|---|---|---|
| S01 Codex Responses非流式与instructions回显 | PR7979, PR8499 | 网关 SHALL 保留客户端原stream选择，将非流式SSE的completed/incomplete/cancelled转换成单JSON；缺Content-Type的OAuth SSE可识别。允许空instructions进入既有规范化，并将网关改写的instructions回显还原为客户端值。 | HTTP/透传/普通/compact，stream省略/false/true；incomplete/cancelled/canceled、失败终态仍失败；instructions缺失/null/空/多行与failover隔离；回显不泄漏注入内容，usage只落一次。 |
| S02 Grok 委派消息最小适配 | PR8465 | 既有Grok Responses入口 SHALL 将agent_message的input_text及明文encrypted_content按原顺序转为user消息，空消息跳过，不尝试解密。 | 封套+正文顺序、空/非文本片段、字符串content、非委派item字节/大整数保真；additional_tools原提升规则、Chat对照、HTTP/Grok usage和模型审计保持。 |
| S03 已使用 Terra 的目录与计费后续 | PR8365 | 已接入Terra SHALL 使用冻结tag中的Codex descriptor作为受限fallback，Codex默认low及500000窗口与max_context_window=872000区分，API兼容默认medium/none仍合法，minimal明确400。OAuth/setup-token长上下文豁免按最终发送模型决定，API Key保留272k严格边界与Fast结算。 | none/minimal/max/ultra的API与Codex不同标尺；别名/映射后校验，账号raw/metadata/空目录优先、ETag；271999/272000/272001×OAuth/setup-token/API Key×default/priority/ultrafast/Free Fast/显式零价与SQLite去重。 |
| S04 请求体预算、读写超时与错误分类 | PR8484, PR8518, PR8457 | 所有既有网关入口 SHALL 共享实例级请求体内存预算，读取/解压/修补/异步移交/failover释放按实际所有权计数；超预算503+Retry-After，超时408、断开499、截断/解码失败保持安全分类。超时按单次读写停滞计，不限制正常SSE总时长。 | 1GiB限制环境测RSS/并发吞吐/503恢复；gzip/zstd/lenient/unknown长度、读取/写入停滞、正常长流、WS Hijack、重试100轮/取消无残留；日志无正文/密钥。默认值须量测，不能直接照搬1GiB预算或500MiB行上限。 |
| S05 上游流、drain与会话缓存有界 | PR8484, PR8518 | 既有上游流 SHALL 限制SSE队列字节/单行/累计事件，非流式读取有界；客户端离开后的usage drain有总时限，缓存按数量和字节逐出并释放后台任务。 | 持续短行/长行/慢消费/取消均有界；超大单条事件不越行上限；截止前已观测用量仍入SQLite且去重，截止后不虚构完整用量；缓存键含账号/会话隔离，过期/逐出/关闭无goroutine泄漏。 |
| S06 Antigravity 系统身份保留任务 | PR8318 | Antigravity路径 SHALL 仅规范化开头的已知客户端身份和attribution，保留附随任务及中间引用；原生Anthropic不受影响。 | Claude/Codex多种前导、紧邻逗号/任务、引用/相似字符串、双身份、string/parts、native Anthropic对照；失败与成功请求body归属不混。 |
| S07 图片 INTERNAL500 不牵连整账号 | PR8521 | 现有Antigravity图片模型重试耗尽INTERNAL500 SHALL 只冷却最终映射模型，并同步缓存；文本模型仍用既有账号惩罚。 | alias映射图片/文本、相同模型计时、其他模型仍可调度、SetModelRateLimit失败、SQLite Extra/调度缓存一致；不增加账号处罚计数。 |
| S08 Gemini 工具媒体原生承载 | PR8430 | Gemini3及以后已使用模型的工具媒体 SHALL 保留在functionResponse.parts；旧模型按已有分离图片回退，关联工具名/顺序保持。 | Gemini2.5/3/4和别名、base64/URI/text混合、未知工具ID、重试去签名、原输入不变、tool_order及图片用量/dedup不重复。 |
| S09 使用记录导出限流恢复 | PR8436 | 现有用户导出 SHALL 每页至多1000条，在429时有界等待并重试同一页；重试耗尽或取消不能下载部分结果。 | 数字/缺失/日期/非法Retry-After、上限等待/重试、分页过滤一致、关闭/卸载/换号取消、不重下/漏行；实际浏览器保存导出及失败恢复。 |
| S10 Antigravity 测试提示词贯通 | PR8531 | 已有账号测试 SHALL 将管理员输入提示词传至Antigravity Claude/Gemini请求，并返回真实回复；空值沿用默认提示词。 | 自定义/空/Unicode/多行prompt、Claude/Gemini模型、响应文本/错误/取消、定时测试旧调用兼容；真实本地HTTP捕获body，外部账号测试不自动执行。 |
| S11 RPM空覆盖缓存与既有JSON预检 | PR8518 | 鉴权快照 SHALL 区分已加载的nil RPM覆盖与加载失败，并按group隔离；既有thinking/委派预检仅在可证明无修改时跳过解析，不能改变输出语义。 | nil/0/正值/错误/跨group、失效与并发、miniredis+SQLite查询次数；escaped/重复键/非法JSON/长输入byte等价，benchmark ns/op+B/op+allocs/op前后实测，无收益不落优化hunk。 |

来源SHA见source-baseline。PR8518混合资源修复、优化和新依赖，各簇只取明确列出的子集；没有授权整个PR。
