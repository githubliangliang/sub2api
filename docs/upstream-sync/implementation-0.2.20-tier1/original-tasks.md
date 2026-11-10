# 实施任务（全部未开始）

- [ ] 核对实际实施起点、来源哈希与范围，保留冻结source-baseline及Fast行为。

## F01 Claude setup-token 周用量

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：无5h、有效/过期5h、显式0、无样本、429以及单查/force/passive/batch一致；SQLite Extra与固定quota窗口不回退。

## F02 Codex 强制模式停止多余探测

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：两个强制模式请求数/写Extra数均0；auto/非法/非字符串模式仍探测，凭据覆写和取消正常。

## F03 WebSocket 每轮首 token 时间

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：至少两轮无ID delta、带ID delta、负时差、重复terminal与取消；保留Fast每轮service_tier/模型冲突和usage隔离。

## F04 Codex WS→HTTP 桥接指纹

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：最终HTTP请求header/body同账号稳定、跨账号隔离；无prompt_cache_key/显式客户端key/API Key/Grok负对照；Fast策略不变。

## F05 缓冲 Responses reasoning 保真

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：summary优先、不重复拼接、仅reasoning_text、仅签名、普通文本及两条真实缓冲网关路径；usage/dedup不重复。

## F06 Responses 完成消息字段

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：空message、output_text、refusal、annotation/logprobs、tool_search/function_call；JSON及SSE终态一致，序号0不被省略。

## F07 鉴权取消分类

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：取消/截止、真实DB失败、格式错误Key、503过载及401错误；不把所有context错误无条件吞成499，不记录Key。

## F08 413 透传规则优先级

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：有/无匹配规则、自定义状态/文案、skip-monitoring、已开始SSE及Retry-After；不提前终止账号failover。

## F09 Antigravity 直连 SSE 响应

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：直接/包装响应、空/非法JSON、工具、usage和错误尾帧；普通流与缓冲兼容路径不产生假完成。

## F10 Antigravity JSON Schema 参数

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：camel/snake命名、双字段优先、非object、undefined清理、原body不变正例；不修改原生Gemini协议。

## F11 GPT Image 账号测试参数

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：严格模拟上游检查无response_format、模型/prompt/n保留、有效图片解析、上游错误；不新增模型/尺寸路由。

## F12 Anthropic兼容 Cloudflare REST URL

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：精确HTTPS host/443/账号路径、相似域名/额外path/query/fragment/rawpath/userinfo负例、Cloudflare native gateway保留beta；三条调用链一致。

## F13 Bedrock EventStream 分配前边界

- [ ] 建立该行为的修复前失败回归，区分产品缺陷和测试基座缺失。
- [ ] 实现选定行为及必要消费者，排除混合PR未选子项。
- [ ] 验收：header越界、total<16、header>128KiB、payload>16MiB、truncated/CRC、合法frame；大长度仅喂prelude，证实分配前拒绝。

## 组合门禁

- [ ] 完整backend unit、embed build、lint及相关race通过，记录实际命令和实现SHA。
- [ ] 真实HTTP/SQLite迁移228、usage settlement/dedup、固定quota窗口及Fast成本回归通过。
- [ ] 适用的前端typecheck/lint/test/build/critical/真实浏览器、性能与资源压力验证完成。
- [ ] 来源/产品范围/OpenSpec验证完成，verification填入可复核输出及残余风险。
