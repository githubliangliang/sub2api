# 实施任务（全部未开始）

- [ ] 核对实际实施起点、来源哈希与范围，保留冻结source-baseline及Fast行为。

## S01 Codex Responses非流式与instructions回显

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：HTTP/透传/普通/compact，stream省略/false/true；incomplete/cancelled/canceled、失败终态仍失败；instructions缺失/null/空/多行与failover隔离；回显不泄漏注入内容，usage只落一次。

## S02 Grok 委派消息最小适配

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：封套+正文顺序、空/非文本片段、字符串content、非委派item字节/大整数保真；additional_tools原提升规则、Chat对照、HTTP/Grok usage和模型审计保持。

## S03 已使用 Terra 的目录与计费后续

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：none/minimal/max/ultra的API与Codex不同标尺；别名/映射后校验，账号raw/metadata/空目录优先、ETag；271999/272000/272001×OAuth/setup-token/API Key×default/priority/ultrafast/Free Fast/显式零价与SQLite去重。

## S04 请求体预算、读写超时与错误分类

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：1GiB限制环境测RSS/并发吞吐/503恢复；gzip/zstd/lenient/unknown长度、读取/写入停滞、正常长流、WS Hijack、重试100轮/取消无残留；日志无正文/密钥。默认值须量测，不能直接照搬1GiB预算或500MiB行上限。

## S05 上游流、drain与会话缓存有界

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：持续短行/长行/慢消费/取消均有界；超大单条事件不越行上限；截止前已观测用量仍入SQLite且去重，截止后不虚构完整用量；缓存键含账号/会话隔离，过期/逐出/关闭无goroutine泄漏。

## S06 Antigravity 系统身份保留任务

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：Claude/Codex多种前导、紧邻逗号/任务、引用/相似字符串、双身份、string/parts、native Anthropic对照；失败与成功请求body归属不混。

## S07 图片 INTERNAL500 不牵连整账号

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：alias映射图片/文本、相同模型计时、其他模型仍可调度、SetModelRateLimit失败、SQLite Extra/调度缓存一致；不增加账号处罚计数。

## S08 Gemini 工具媒体原生承载

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：Gemini2.5/3/4和别名、base64/URI/text混合、未知工具ID、重试去签名、原输入不变、tool_order及图片用量/dedup不重复。

## S09 使用记录导出限流恢复

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：数字/缺失/日期/非法Retry-After、上限等待/重试、分页过滤一致、关闭/卸载/换号取消、不重下/漏行；实际浏览器保存导出及失败恢复。

## S10 Antigravity 测试提示词贯通

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：自定义/空/Unicode/多行prompt、Claude/Gemini模型、响应文本/错误/取消、定时测试旧调用兼容；真实本地HTTP捕获body，外部账号测试不自动执行。

## S11 RPM空覆盖缓存与既有JSON预检

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：nil/0/正值/错误/跨group、失效与并发、miniredis+SQLite查询次数；escaped/重复键/非法JSON/长输入byte等价，benchmark ns/op+B/op+allocs/op前后实测，无收益不落优化hunk。

## 组合门禁

- [ ] 完整backend unit、embed build、lint及相关race通过，记录实际命令和实现SHA。
- [ ] 真实HTTP/SQLite迁移228、usage settlement/dedup、固定quota窗口及Fast成本回归通过。
- [ ] 适用的前端typecheck/lint/test/build/critical/真实浏览器、性能与资源压力验证完成。
- [ ] 来源/产品范围/OpenSpec验证完成，verification填入可复核输出及残余风险。
