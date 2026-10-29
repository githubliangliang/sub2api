# 实施任务（26/26 完成）

状态：2026-10-29 用户授权实施后，F01–F07 与组合门禁已完成。实际前后证据、实施提交及保护哈希见 [verification](./verification.md)；原始来源与评估证据保持冻结。

- [x] 记录实际实施起点，核对冻结产品差异及个人范围，保留source-baseline。

## F01 Claude Code UTF-16指纹

- [x] 先建立预期新行为的失败回归，区分生产缺陷和fixture/缺前置错误。
- [x] 实现选定行为：按JS UTF-16码元采样后拼接、再按JS UTF-8语义哈希，读取前缀有界。
- [x] 验证正例、反例和范围：ASCII、中文、emoji、采样surrogate重组及不足下标；版本更新/重复调用，所有既有Claude正文与header消费者一致。

## F02 Codex模型缓存flight内复查

- [x] 先建立预期新行为的失败回归，区分生产缺陷和fixture/缺前置错误。
- [x] 实现选定行为：miss调用进入flight后遇已填fresh缓存直接复用，使用本地实际缓存消费者。
- [x] 验证正例、反例和范围：确定性交错fetch=0、普通miss/stale仍获取、ETag/304、账号隔离、并发与取消。

## F03 共享缓存命中用量

- [x] 先建立预期新行为的失败回归，区分生产缺陷和fixture/缺前置错误。
- [x] 实现选定行为：OpenAI-compatible顶层prompt_cache_hit_tokens作为fallback，不重复合并已存在的嵌套缓存桶。
- [x] 验证正例、反例和范围：嵌套显式0/正值优先、prompt100/hit80/miss20、HTTP/SSE同桶及SQLite成功usage去重，无新DeepSeek平台。

## F04 Gemini fileData图片统计

- [x] 先建立预期新行为的失败回归，区分生产缺陷和fixture/缺前置错误。
- [x] 实现选定行为：既有图片observer识别inline/file及camel/snake字段，累计chunk保留最大实际张数。
- [x] 验证正例、反例和范围：有效MIME与非空URI/data、同part至多一次、累计SSE/别名/旧fallback、取消已观测用量；不下载URI或扩计量模式。

## F05 Antigravity畸形工具终态

- [x] 先建立预期新行为的失败回归，区分生产缺陷和fixture/缺前置错误。
- [x] 实现选定行为：直接/包装响应和缓冲流中的MALFORMED_FUNCTION_CALL产生既有格式错误，先保留无parts尾帧失败。
- [x] 验证正例、反例和范围：JSON/wrapped/SSE错误尾帧均不返回空成功，普通文字/工具与取消/错误透传保持，失败用量规则不回退。

## F06 分页数值与偏移量边界

- [x] 先建立预期新行为的失败回归，区分生产缺陷和fixture/缺前置错误。
- [x] 实现选定行为：安全解析数字，所有共享offset与offset+limit有界，各入口保留自己的默认/限幅/400策略。
- [x] 验证正例、反例和范围：超长数字、MaxInt/负数/空值/非法值、最大合法页、直接构造params、Key内存切片及HTTP入口。

## F07 管理员保护返回400

- [x] 先建立预期新行为的失败回归，区分生产缺陷和fixture/缺前置错误。
- [x] 实现选定行为：禁止最后管理员降级及直接disable/delete时使用领域400错误，守卫规则不变。
- [x] 验证正例、反例和范围：真实handler400/稳定码、被拒绝动作不写库、其它管理员存在时允许降级、数据库故障仍500。

## 组合门禁

- [x] 完成backend完整unit/build/lint及受影响路径race，记录原始输出和实现哈希。
- [x] 完成真实SQLite迁移227、已配置quota/0值/软删不复活、固定窗口与成功usage去重回归。
- [x] 如实际变更触及前端/依赖，完成适用pnpm与audit检查；未触及时据实记录不适用。
- [x] 核对OpenSpec/链接/冻结来源/支付与新能力边界，在verification填写实际证据。

