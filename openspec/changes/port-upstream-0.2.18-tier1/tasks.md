# 实施任务

所有任务待实施；评估期基线测试不勾选实施验收。

- [ ] 记录实施起点并核对冻结基线差异、工作树及个人版范围。

## F01 Codex 工作区导入身份隔离

- [ ] 先建立缺陷回归证据，再实现选定行为：以 workspace/account 与 user 的组合区分完整 OAuth 身份；同一工作区重复导入更新原账号，不覆盖其它工作区。
- [ ] 验证正例、反例和边界：两个工作区同一用户各自创建；再次导入只更新对应账号；只有 token、缺 user/account、共享 account 的成员身份分别覆盖。
- [ ] 保留范围约束：保留 accessToken-only 升级和 account 兜底；长度编码防拼接碰撞，不改变凭据存储格式。

## F02 Codex 前导元数据不阻止重试

- [ ] 先建立缺陷回归证据，再实现选定行为：排队和审查元数据保持首个语义输出之前的状态，后续容量错误仍可安全重试，首输出超时仍有效。
- [ ] 验证正例、反例和边界：HTTP SSE、透传和 WS-HTTP 桥在 queued/metadata 后遇到 overload 可重试；已有有效输出后不重试。
- [ ] 保留范围约束：真实文本、工具参数和未知事件仍按现有输出语义；保留本地 SSE 注释写出后的安全换号规则。

## F03 Grok Fast 用量成本价卡

- [ ] 先建立缺陷回归证据，再实现选定行为：既有 Grok 模型映射/透传请求按 Fast 独立价卡计算；短上下文 input/cache/output 每百万为 4/1/12，达到 200k 后为 6/1.5/18（来源补丁值）。
- [ ] 验证正例、反例和边界：fast/build-fast/provider 前缀、199999/200000 边界、缓存桶、动态资源与离线 fallback 一致；普通 4.7 不受影响。
- [ ] 保留范围约束：不扩展模型目录、平台或购买功能；普通 Grok 与显式渠道价格覆盖不变；历史账本不回写。

## F04 Cloudflare 1010 不处罚账号

- [ ] 先建立缺陷回归证据，再实现选定行为：识别明确的 Cloudflare 1010 或 browser_signature_banned，所有既有平台跳过账号处罚，保留请求错误和换号处理。
- [ ] 验证正例、反例和边界：五个已有平台的 HTML/JSON/嵌套 Cloudflare 样本不调用 SetError/限流计数；真实403继续原有策略。
- [ ] 保留范围约束：普通供应商 403、任意数字 1010、与 Cloudflare 无关业务错误不能被豁免；不引入 OpenCode/Command Code。

## F05 Gemini 缓存命中用量非负

- [ ] 先建立缺陷回归证据，再实现选定行为：仅缓存计数或缓存大于 prompt 时非缓存 input_tokens 下限为0，同时保留缓存、输出和图片用量。
- [ ] 验证正例、反例和边界：流式 message_start/更新与非流式在 prompt缺失、0、小于/等于/大于cached 时保持非负且桶语义一致。
- [ ] 保留范围约束：正常 prompt>=cached 的计数不变；不抹掉缓存命中，不引入显式缓存CRUD或存储计费。

## F06 Gemini 工具 Schema 常量和数据边界

- [ ] 先建立缺陷回归证据，再实现选定行为：字符串const转为单值enum，与已有enum求交；移除不支持的非字符串const和注解，区分Schema节点、属性名和数据值。
- [ ] 验证正例、反例和边界：字符串/数值const、enum交集、组合Schema、properties同名键、数据对象、exclusiveMinimum同时验收；原schema不被修改。
- [ ] 保留范围约束：保留上一轮 integer exclusiveMinimum 修复；不得误删名为const/examples/$schema的合法属性，也不得改写default等数据内部值。

## 合并验收

- [ ] 后端unit/build、受影响链路race与真实SQLite结算/dedup/固定窗口回归。
- [ ] 若改变前端或依赖，完成pnpm typecheck/lint/build/关键vitest及重新audit。
- [ ] 验证无支付/新平台/新schema混入、VERSION保持fork版本规则，并填写verification。
