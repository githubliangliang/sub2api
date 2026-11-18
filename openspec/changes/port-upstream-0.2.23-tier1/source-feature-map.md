# 来源行为映射

| 簇 | 来源 | 所需行为 | 评估依据 |
|---|---|---|---|
| F01 Responses 密封调用内容 | [#8682](https://github.com/Wei-Shaw/sub2api/pull/8682) | 函数调用的 delta 回退重建 SHALL 保留 encrypted_content。 | apicompat overlay 丢失 sealed-content；namespace/密封参数对照通过。 |
| F02 对话中途 system 权限 | [#8703](https://github.com/Wei-Shaw/sub2api/pull/8703) | Chat/Anthropic 中途 system 消息转 Responses SHALL 保持 developer 权限。 | MidSystemAuthority overlay 实际 role=user。 |
| F03 JSON Schema 默认 strict | [#8761](https://github.com/Wei-Shaw/sub2api/pull/8761) | Chat json_schema 未设置 strict 时转 Responses SHALL 显式写 false。 | StrictDefault overlay 缺 strict；显式 true 对照通过。 |
| F04 Codex HTTP 外部上下文 | [#8769](https://github.com/Wei-Shaw/sub2api/pull/8769) | 无对应调用的 agent browser_info/goal_objective/untrusted_context 输出 SHALL 转为不可信用户文本。 | handler overlay browser_info 未转换；真实 call_id 对照通过。 |
| F05 Claude per-turn beta | [#8742](https://github.com/Wei-Shaw/sub2api/pull/8742) | 任一已支持的 per-turn/mid-conversation beta SHALL 保留消息级 output_config。 | PerTurnBeta overlay 被净化；原 beta 和无 beta 对照通过。 |
| F06 count_tokens 请求头唯一性 | [#8788](https://github.com/Wei-Shaw/sub2api/pull/8788) | Anthropic count_tokens 透传 SHALL 以原始大小写无关读取/写入避免重复默认头。 | 修正 DI 夹具后 header overlay 复现同名头计数为2。 |
| F07 输出后的兼容流账号副作用 | [#8814](https://github.com/Wei-Shaw/sub2api/pull/8814) | OpenAI Chat/Messages 已输出后的流错误 SHALL 执行既有账号/模型错误副作用。 | 静态：两 reader 在早期 failover 分支之后直接记录错误，缺副作用调用。 |
| F08 既有 Sol Ultra Fast 倍率 | [#8812](https://github.com/Wei-Shaw/sub2api/pull/8812) | 已支持的 GPT-6.1 Sol Ultra Fast SHALL 使用冻结 tag 的6倍模型倍率。 | SolUltrafast overlay 实际缺模型倍率；Astra6倍对照通过。 |
| F09 登录身份回填空指针 | [#8684](https://github.com/Wei-Shaw/sub2api/pull/8684) | 登录邮箱身份回填失败返回 nil 时 SHALL 安全跳过默认设置判断。 | LoginNilIdentity overlay 捕获 panic；历史回填来源对照通过。 |
| F10 5h 滚动保留有效周用量 | [#8723](https://github.com/Wei-Shaw/sub2api/pull/8723) | Claude 5h 窗口滚动 SHALL 仅清除已过期的周读数。 | WeeklySurvivesFiveHourRollover overlay 清空未到期7d读数。 |
| F11 账号5h窗口上下界 | [#8698](https://github.com/Wei-Shaw/sub2api/pull/8698) | 主动读取更新5h窗口 end 时 SHALL 同步 start=end-5h。 | 实际 modernc SQLite/Ent overlay 更新 end 后 start 仍在旧窗口。 |
| F12 首输出临时文件不可用 | [#8808](https://github.com/Wei-Shaw/sub2api/pull/8808) | 首输出暂存无法创建临时文件时 SHALL 在同一容量上限内回退内存。 | SpoolUnavailable overlay 返回错误；stage上限拒绝对照通过。 |
| F13 Claude 默认测试模型维护 | [#8757](https://github.com/Wei-Shaw/sub2api/pull/8757) | 内置 Claude 默认测试模型 SHALL 与冻结 tag 的 claude-sonnet-4-6 一致。 | 静态常量差异；基线相关测试通过，未调用真实 Anthropic 核实退役公告。 |
| F14 Grok 媒体扁平错误 | [#8669](https://github.com/Wei-Shaw/sub2api/pull/8669) | Grok 既有媒体错误 SHALL 保留 xAI 扁平 error/code，保持脱敏与可配置错误规则。 | GrokFlatMediaMessage overlay 丢失字符串；嵌套/未知对象对照通过。 |
| F15 Gemini 非数组 items | [#8759](https://github.com/Wei-Shaw/sub2api/pull/8759) | 既有 Gemini schema 清理 SHALL 删除非 array 的 items，并保持合法数组 schema。 | SchemaNonArrayItems overlay 失败；tuple/default及合法数组对照已经通过。 |
| F16 白名单精确模型输入 | [#8736](https://github.com/Wei-Shaw/sub2api/pull/8736) | 模型白名单添加操作 SHALL 拒绝含星号的条目并保留输入草稿。 | 临时前端4个来源新增测试失败；原11个控件测试通过。 |
| F17 渠道 Bedrock 兼容开关类型 | [#8840](https://github.com/Wei-Shaw/sub2api/pull/8840) | 渠道编辑 SHALL 用 boolean 保存/清除 bedrock_cc_compat，并保留其它 features。 | 临时前端5个来源测试失败，实际保存的是 {anthropic:true/false}。 |
| F18 调度快照端点能力 | [#8837](https://github.com/Wei-Shaw/sub2api/pull/8837) | 调度快照 SHALL 保留已有 openai_capabilities 声明并据此过滤既有端点。 | repository overlay 丢键；实际 Account 能力判断由禁用 embeddings 变成允许。 |
| F19 设置注入 HTML 与 CSP nonce | [#8775](https://github.com/Wei-Shaw/sub2api/pull/8775) | 设置注入 HTML SHALL 禁止共享/浏览器缓存，携带请求 nonce 时不得按旧 ETag 返回304。 | embed overlay 复现nonce仍304、缺no-store；无nonce304对照通过。 |
| F20 Chat 客户端 metadata.user_id | [#8786](https://github.com/Wei-Shaw/sub2api/pull/8786) | API Key 的 Chat转Anthropic SHALL 补回已识别格式的客户端 metadata.user_id。 | 静态：DTO转换丢metadata，两个转发路径没有补回；尚无专用转发探针。 |
| F21 完成转发后的粘性绑定 | [#8726](https://github.com/Wei-Shaw/sub2api/pull/8726) | 完成转发后的粘性写入 SHALL 不随客户端取消，且受5秒截止限制。 | 静态：handler用请求ctx绑定；service原样交给cache，无内部脱离取消。 |


补丁位置和最小前置只记录在评估报告。本表定义选取的行为；同PR未选分支不构成授权。完整候选和四态见冻结证据。
