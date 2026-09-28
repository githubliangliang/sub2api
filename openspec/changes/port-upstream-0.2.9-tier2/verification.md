# 实施验收证据（尚未实施）

评估证据单列：[基线检查](../../../docs/upstream-sync/evidence-0.2.9/baseline-tests.md)。不作为下表完成证据。
实际实施起点SHA：

完成SHA：

执行日期：

| ID | 失败场景/正常反例 | 改前结果 | 改后结果及退出码 | 用例/证据路径 | 实施SHA |
|---|---|---|---|---|---|
| S01 | 接入 GPT-6 Sol/Luna、Opus 5.5、Grok 4.7 的现有平台目录/能力/价格别名，并将 GPT≥5 按推理模型桥接；兼容 dotted Opus 别名；非 GPT 数字系列和旧模型保持行为；显式用户映射优先 | | | | |
| S02 | 处理根级 union、required:null、tuple/prefixItems，并保留 string const 与 enum 的交集；普通object不变；保留目标协议可表达约束；tuple/root union降格不宣称schema完全等价；多个分流入口均执行清洗 | | | | |
| S03 | arguments.done 等于已发送 delta；仅 block_start 携带参数也保留；内嵌 PDF 转为 document/inlineData；有 delta 时覆盖 seed，不拼接两个 JSON；空 PDF 不发；不自动下载 file_id | | | | |
| S04 | terminal 完整发送后结束；先取消单次请求再关响应体；心跳不算语义输出；错误只发一次；Gemini 传输失败换号；客户端取消归类 499；bare error 后可能 completed 不早退；上游真实错误继续归因；取消不触发重复重放或重复用量；countTokens 保留本地估算 | | | | |
| S05 | 各入口将裸 Gemini 名按 thinking 配置映射；识别特定 SDK 的心跳兼容；MALFORMED_FUNCTION_CALL 空流触发 failover；显式 model_mapping 优先；只有 signature/stop 不算输出；真实文本/工具调用不重试 | | | | |
| S06 | 非高级调度遵守 previous_response 归属；热路径用轻量分组读；API Key 未知模型 401 不误禁账号；暂停 OAuth 仍刷新 token；失效归属释放槽位；真实凭据 401 保留禁用；永久错误账号不进入刷新候选 | | | | |
| S07 | 账号成本是否收取长上下文溢价取决于账号 gate；渠道图片价未填继承目录，显式 0 才免费；保留自定义账号价与 ApplyPricingToAccountStats 优先级；不改变用户售价门控；空价与零价分开 | | | | |
| S08 | Codex config 请求正确 /v1；CC Switch保留配置端点（含显式/v1及子路径）、仅去尾斜杠且不自动追加/v1；usage路径恰好一个/v1；Windows catalog用~/；带/不带/v1、尾斜杠、子路径和不同平台均验证；原生配置与 CC Switch 的 URL 规则分别覆盖 | | | | |
| S09 | 代理恢复改变网络身份时失效探针；过期扫描写入前重新核对快照；禁用代理不可作为回退目标；无真实 proxy 变化不清探针；续期/停用/改回退配置时旧扫描不得覆盖 | | | | |
| S10 | 探测模型不存在时不写入endpoint能力标记（未知或已有值都保留），优先选普通GPT文本模型；真正endpoint404/405保持不支持；不能把任意400当模型不可用或强制清空已有标记 | | | | |
| S11 | Responses 与 Chat 返回请求方的公开模型别名，同时保留真实上游模型用于记录；工具/事件里的其它 model 字段不误改；未映射模型保持原样 | | | | |
| S12 | 关键词检查不丢弃客户端 reminder 内文本，尾随 system 不遮蔽当前用户输入；assistant/tool 结束回合不重复审计；正常语义审核保留既有 reminder 策略 | | | | |
| S13 | 仅在 Antigravity 转换中移除前导 attribution 并中和前导 SDK 身份；正文中提及 Claude 不改；原生 Anthropic attribution 必须保留 | | | | |
| S14 | 嵌套弹窗关闭后滚动锁计数正确，模型标签 IME 不提前提交；其它弹窗仍开时 body 保持锁；非组合输入和原键盘导航不回退 | | | | |
| S15 | window_id 改变时删除旧 previous_response_id 并清空旧续聊推断锚点；同窗口、无窗口标记保持续聊；失败轮次不得提前更新 last window | | | | |
| S16 | 映射到 GPT-5.5 时移除不兼容 Lite 标记但保留历史与工具；turn metadata 序列化保留非 ASCII 转义；其它模型和账户不改；失败后换号仍使用原请求；中文与补充平面字符转义合法 | | | | |

| 最终门禁 | 命令/结果 | 证据 |
|---|---|---|
| 后端unit/build | | |
| go test -list与实际用例存在 | | |
| SQLite方言/真实库 | | |
| 并发与取消race（如涉及） | | |
| 前端typecheck/Vitest | | |
| UI/端点实测（如涉及） | | |
| VERSION/迁移/依赖/平台范围 | | |
| 未通过或未运行项 | | |

未运行、超时、工具缺失不得写PASS；源码有测试名也不代表该测试成功运行。
