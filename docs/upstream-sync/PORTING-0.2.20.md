# 上游 v0.2.20 移植评估（v0.2.19 → v0.2.20）

评估日期：2026-11-10。已通过GitHub releases API及git fetch确认最新正式版为v0.2.20，发布于2026-11-07 06:31:24 UTC / 14:31:24上海时间。范围按tag祖先固定，不按作者日期、不把main后续修复混入发布版本。

建议**第一档13簇、第二档11簇**，优先Claude Code、Codex、Grok和共享数据/资源安全。两批均为**未实施提案**；此次没有应用上游产品补丁、修改版本/依赖/schema、提交、push、发布或部署。已存在Fast改动在采集期间由并发操作提交，详情见基线。

- [第一档OpenSpec](../../openspec/changes/port-upstream-0.2.20-tier1/README.md)
- [第二档OpenSpec](../../openspec/changes/port-upstream-0.2.20-tier2/README.md)
- [原始清单、补丁、基线与复现](./evidence-0.2.20/MANIFEST.md)

## 1. 版本、范围与规模

| 项目 | 冻结值 |
|---|---|
| fork HEAD | 326513cdc68cc50d31e25a07ae98ecd6f5a3ca1f |
| VERSION / Go / SQLite最大迁移 | 1.1.23 / 1.27.3 / 228 |
| 会话开始HEAD | 5c135c31dfbd6a1fcb701767f2f78933c4280aea，Fast当时未提交；采集期间提交326513cdc，保护文件哈希未改变 |
| 独立index tree | 77e4e07303db135109fb85c1136eda035889a43a |
| v0.2.19 commit | dece26153621959ccd7635814f4b2d08304e681a |
| v0.2.20 commit | dab829a4e9b7d1cc307189aea8b0319eb29e29d0 |
| v0.2.20 tag对象 | ec41309811cd30af310a0e4b9119b976ce65c9be（annotated tag，不是源码commit） |
| fetch后upstream/main | ee58c0013c76ccd3f36439881bc848d342ec1c9d，只记录，不取码 |
| 非merge / 候选 | 72提交 / 42候选 = 35PR + 7直接提交 |
| 区间净差异 | 459文件，+29482 / -1457 |
| 候选×文件四态 | 562条：ALREADY3 / CLEAN330 / CONFLICT157 / NOBASE72 |

562为每候选文件数之和，同文件可重复；不是459净文件数。#8518的11微提交取整PR终态。第一档13簇覆盖13PR；第二档11簇覆盖12PR，其中#8518与第一档交叉，因此合计24个不同PR，混合PR只取指定子项。

| 档位 | 意义与本轮结论 |
|---|---|
| 第一档 | 已确认本地缺陷、窄修复：F01–F13，生成未实施change |
| 第二档 | 有本地消费者、需要跨路径/手工/前端/资源适配：S01–S11，生成未实施change |
| 第三档 | 新管理MCP/黑名单/图片路由/导出默认策略/配额展示/开关、完整COW优化等，等待用户选定 |
| 第四档 | 无本地缺陷或消费者的Kimi/受限admin-key修补、cyber transcript算法、PG专有hunk、支付、上游宣传与VERSION，不移植 |

## 2. 方法、渠道与迁移边界

1. 执行`git fetch upstream --tags --prune`；GitHub API发布列表与tag互相核对。完整42候选及72非merge覆盖断言、父SHA和补丁SHA256见[candidates.tsv](./evidence-0.2.20/candidates.tsv)与[commits.tsv](./evidence-0.2.20/commits.tsv)。
2. 以实际产品工作树构建临时index，全部562文件先reverse、再forward；不修改真实index，不试落上游代码。[files.tsv](./evidence-0.2.20/files.tsv)记录每次退出码/错误。初始Fast未提交→326513cdc的内容等价另有保护哈希，不借旧HEAD判断遗漏Fast。
3. PR整体diff用于四态，终态tag用于实现设计。文件NOBASE不能直接变成“没缺陷”：Grok上游输入函数缺失，但本地等价消费者存在；反之cyber transcript缺陷的算法本地确实没有。
4. [added-call-inventory.tsv](./evidence-0.2.20/added-call-inventory.tsv)是5795条宽松新增Go调用扫描，含内置函数、测试/注释与方法；不等于5795个前置。[base-symbols.txt](./evidence-0.2.20/base-symbols.txt)及下文逐簇说明实际存在/需新增/不引入的符号。新代码尚未编译，不能由CLEAN推断实现通过。
5. 保留SQLite-only、miniredis、simple mode、迁移228、固定quota窗口、usage dedup和Fast完整能力。上游249_mcp_oauth.sql属于未选MCP能力，不能转成229顺带导入；两批拟议修复不需要新schema。若后续方案改变，重新评估本地编号，不修改历史迁移校验和。
6. **发布说明配置名与源码有差异**：说明写`security.request_body_memory_budget_bytes`等，v0.2.20 Config/Validate/示例均为`gateway.request_body_memory_budget_bytes`、`gateway.request_body_read_timeout_seconds`、`gateway.downstream_write_timeout_seconds`。以实际源码消费链为准。1GiB默认预算、500MiB行上限不适合作为1C1G默认值直接继承；本轮只立适配及压测要求，尚无默认值测量结论。
7. 测试必须区分现存基线、缺陷探针和未来验收。[source-test-inventory.tsv](./evidence-0.2.20/source-test-inventory.tsv)列179来源测试文件；[integration标签](./evidence-0.2.20/integration-build-tags.tsv)单独记录本地PG-only测试。不能通过删除postgres标签伪造SQLite覆盖。

### 全部候选分档

四态依次为ALREADY/CLEAN/CONFLICT/NOBASE；混合项不按整PR提升或降低。

| 候选 | 非merge数 | 四态 | 档位/判断 |
|---|---:|---|---|
| [a9240d501](https://github.com/Wei-Shaw/sub2api/commit/a9240d5016c226b341242a8c54a01a8605cbbc72) | 1 | 0/0/1/0 | 第四档：上游VERSION，不改fork版本 |
| [PR7979](https://github.com/Wei-Shaw/sub2api/pull/7979) | 4 | 0/1/2/0 | 第二档 S01：OAuth非流式/终态JSON |
| [PR8389](https://github.com/Wei-Shaw/sub2api/pull/8389) | 7 | 0/58/19/1 | 第三档：管理MCP/OAuth新能力；支付操作子集第四档；249迁移不取 |
| [d851a3847](https://github.com/Wei-Shaw/sub2api/commit/d851a38477597157e78450d30af2d8504764d085) | 1 | 2/1/4/0 | 已满足/N/A：vue-i18n11.4.16、compiler、decode-uri、Node24本地已有；保留更强本地维护 |
| [PR8476](https://github.com/Wei-Shaw/sub2api/pull/8476) | 3 | 0/23/9/33 | 第四档：未引入受限管理员Key框架；该整能力仍第三档，不能为修补导入 |
| [PR8465](https://github.com/Wei-Shaw/sub2api/pull/8465) | 1 | 0/1/1/2 | 第二档 S02：Grok委派消息；量出缺基座后局部适配 |
| [PR8503](https://github.com/Wei-Shaw/sub2api/pull/8503) | 1 | 0/1/0/3 | 第四档：Kimi原生账号/并发消费者不存在 |
| [PR8436](https://github.com/Wei-Shaw/sub2api/pull/8436) | 1 | 0/5/1/0 | 第二档 S09：已有使用记录导出429恢复 |
| [PR8318](https://github.com/Wei-Shaw/sub2api/pull/8318) | 2 | 0/4/1/0 | 第二档 S06：Antigravity身份规范化保留任务 |
| [PR8499](https://github.com/Wei-Shaw/sub2api/pull/8499) | 1 | 0/6/3/0 | 第二档 S01：Codex空instructions与回显还原 |
| [PR8462](https://github.com/Wei-Shaw/sub2api/pull/8462) | 2 | 0/8/13/0 | 第三档：图片1K/2K/4K选号新策略；其sticky溢出修复依附新策略，本地无此缺陷 |
| [PR8504](https://github.com/Wei-Shaw/sub2api/pull/8504) | 1 | 0/31/2/0 | 第三档：注册邮箱精确黑名单/批量禁用新能力 |
| [e353c8322](https://github.com/Wei-Shaw/sub2api/commit/e353c832261c9358949e2071bfb17e99d36aa8cc) | 1 | 1/0/2/1 | 第四档：上游赞助商/宣传资源 |
| [251094436](https://github.com/Wei-Shaw/sub2api/commit/251094436325175c9e0ab06e7b962d05369fd766) | 1 | 0/0/2/1 | 第四档：上游赞助商/宣传资源 |
| [876b96d20](https://github.com/Wei-Shaw/sub2api/commit/876b96d204855ad1aac486b7dd39a95d377d4760) | 1 | 0/2/2/1 | 第四档：上游赞助商/宣传资源 |
| [PR8484](https://github.com/Wei-Shaw/sub2api/pull/8484) | 1 | 0/75/22/7 | 第二档 S04/S05：既有网关资源治理；不存在路由与新平台部分不取 |
| [PR8495](https://github.com/Wei-Shaw/sub2api/pull/8495) | 1 | 0/1/0/0 | 第三档/被后续取代：CC Switch默认5.6-sol由#8380再改6.1-sol，不独立落中间值 |
| [25608d216](https://github.com/Wei-Shaw/sub2api/commit/25608d216bb7be04e4a646830cb065cd09b7d7e9) | 1 | 0/1/1/1 | 第四档：上游赞助商/宣传资源 |
| [PR8549](https://github.com/Wei-Shaw/sub2api/pull/8549) | 2 | 0/0/7/3 | 第四档：本地无transcript阻断算法；保留显式会话ID策略，不顺带升级两项依赖 |
| [PR8543](https://github.com/Wei-Shaw/sub2api/pull/8543) | 1 | 0/2/8/0 | 第三档：关闭OpenAI传输错误冷却的新开关 |
| [PR8518](https://github.com/Wei-Shaw/sub2api/pull/8518) | 11 | 0/31/11/4 | 第一档 F13；第二档 S04/S05/S11；完整ojson/COW/目录外消费者第三档或第四档 |
| [e6c19082f](https://github.com/Wei-Shaw/sub2api/commit/e6c19082f24740a098f6753d0fbfef754098c8b5) | 1 | 0/1/2/1 | 第四档：上游赞助商/宣传资源 |
| [PR8563](https://github.com/Wei-Shaw/sub2api/pull/8563) | 1 | 0/5/0/0 | 第一档 F05：缓冲reasoning完整性 |
| [PR8429](https://github.com/Wei-Shaw/sub2api/pull/8429) | 1 | 0/3/0/0 | 第一档 F06：Responses完成消息字段 |
| [PR8531](https://github.com/Wei-Shaw/sub2api/pull/8531) | 1 | 0/0/2/0 | 第二档 S10：prompt参数链本地缺前置，需手工贯通 |
| [PR8521](https://github.com/Wei-Shaw/sub2api/pull/8521) | 1 | 0/3/0/0 | 第二档 S07图片模型冷却；环境可调惩罚第三档 |
| [PR8437](https://github.com/Wei-Shaw/sub2api/pull/8437) | 1 | 0/1/1/0 | 第一档 F08：413规则优先级 |
| [PR8430](https://github.com/Wei-Shaw/sub2api/pull/8430) | 1 | 0/10/1/1 | 第二档 S08：已有Gemini工具媒体；不新增模型/生图能力 |
| [PR8557](https://github.com/Wei-Shaw/sub2api/pull/8557) | 1 | 0/2/9/5 | 第三档：GPT Image2.6原生链、complexity与编辑扩展；F11无需此基座 |
| [PR8457](https://github.com/Wei-Shaw/sub2api/pull/8457) | 2 | 0/6/17/8 | 第二档 S04：请求体错误分类及预算修复；Grok Voice/缺失route修补第四档 |
| [PR8522](https://github.com/Wei-Shaw/sub2api/pull/8522) | 2 | 0/10/0/0 | 第三档：新的真实quota-summary API和窗口展示；不是SQLite后台修复前置 |
| [PR8553](https://github.com/Wei-Shaw/sub2api/pull/8553) | 1 | 0/4/0/0 | 第一档 F01：setup-token周用量 |
| [PR8441](https://github.com/Wei-Shaw/sub2api/pull/8441) | 1 | 0/2/0/0 | 第一档 F09：直接Gemini SSE解包 |
| [PR8432](https://github.com/Wei-Shaw/sub2api/pull/8432) | 1 | 0/1/1/0 | 第一档 F11：已有GPT Image测试无效参数 |
| [PR8445](https://github.com/Wei-Shaw/sub2api/pull/8445) | 1 | 0/1/1/0 | 第一档 F02：forced模式停止探测 |
| [PR8520](https://github.com/Wei-Shaw/sub2api/pull/8520) | 1 | 0/1/1/0 | 第一档 F03：每轮首token时间，不引入额外计时接口 |
| [PR8464](https://github.com/Wei-Shaw/sub2api/pull/8464) | 1 | 0/0/2/0 | 第一档 F04：WS HTTP桥接指纹 |
| [PR8471](https://github.com/Wei-Shaw/sub2api/pull/8471) | 1 | 0/5/0/0 | 第一档 F07取消分类；PG非法UTF-8故障理由不适用，防御性校验不作为已确认SQLite缺陷 |
| [PR8409](https://github.com/Wei-Shaw/sub2api/pull/8409) | 1 | 0/2/0/0 | 第一档 F10：Antigravity responseJsonSchema |
| [PR8420](https://github.com/Wei-Shaw/sub2api/pull/8420) | 1 | 0/6/0/0 | 第一档 F12：既有Anthropic兼容自定义URL |
| [PR8380](https://github.com/Wei-Shaw/sub2api/pull/8380) | 3 | 0/4/3/0 | 第三档：导出默认模型/API Key mode/live search/native image配置策略；需用户选择用途 |
| [PR8365](https://github.com/Wei-Shaw/sub2api/pull/8365) | 3 | 0/12/6/0 | 第二档 S03：已选Terra后续；目录API/Codex标尺及账号Fast声明分开 |

## 3. 第一档：确认缺陷与窄修复

| 簇 | 来源 | 本地缺口与依据 |
|---|---|---|
| F01 Claude setup-token 周用量 | PR8553 | 单账号GetUsage只走5h估算；UpdateSessionWindow在无5h-status时直接返回，现有getPassiveUsageForAccount已能展示周窗口。隔离探针确认SevenDay为nil。 |
| F02 Codex 强制模式停止多余探测 | PR8445 | ProbeOpenAIAPIKeyResponsesSupport未检查responses模式；模拟transport确认两个强制模式仍发请求并写探测结果，auto三个对照通过。 |
| F03 WebSocket 每轮首 token 时间 | PR8520 | activeTurn计时受连接级firstTokenMs==nil闸门限制；确定性两轮探针第二轮为nil，第一轮700ms正常。 |
| F04 Codex WS→HTTP 桥接指纹 | PR8464 | 桥接proxyOpenAIWSHTTPBridgeTurn缺stageCodexFingerprintIDs调用；HTTP Forward已有实现，上游builder只读暂存ID，导致桥接遗漏账号级缓存身份。 |
| F05 缓冲 Responses reasoning 保真 | PR8563 | 两条转换器只读summary，隔离探针显示有可见reasoning_text时输出仍为空；22行helper可共享。 |
| F06 Responses 完成消息字段 | PR8429 | ResponsesOutput.MarshalJSON仅特判tool_search_call；空message省略content；messageContentWire将refusal写成text。两个探针分别失败。 |
| F07 鉴权取消分类 | PR8471 | 普通及Google鉴权middleware在业务错误之后直接500，未区分请求ctx已取消。非法UTF-8引发PG500的原始理由不适用于SQLite，单列排除。 |
| F08 413 透传规则优先级 | PR8437 | handleFailoverExhausted的body-too-large分支早于MatchRule返回，现有413规则不可达。 |
| F09 Antigravity 直连 SSE 响应 | PR8441 | 直接响应也能成功Unmarshal成V1InternalResponse但字段空，旧代码只有反序列化报错才尝试直接形态；探针仅返回message_start。 |
| F10 Antigravity JSON Schema 参数 | PR8409 | cleanGeminiRequest没有schema映射，隔离探针确认目标字段缺失。 |
| F11 GPT Image 账号测试参数 | PR8432 | testOpenAIImageAPIKey仍构造response_format=b64_json；新GPT Image2.6链路不是修此参数的前提。 |
| F12 Anthropic兼容 Cloudflare REST URL | PR8420 | 普通转发、API Key透传和账号测试都无条件追加?beta=true；是既有兼容URL缺陷，不需新增Cloudflare平台。 |
| F13 Bedrock EventStream 分配前边界 | PR8518 | 仅16字节、合法CRC、headers_length=100的帧已复现slice panic；源码totalLength直接控制make分配，未执行大内存攻击。 |

### F01 patch sites / 前置 / 验收

- service/account_usage_service.go、ratelimit_service.go：getUsageForAccount、UpdateSessionWindow、HandleUpstreamError；supportsAnthropicPassiveUsage/getPassiveUsageForAccount/samplePassiveUsageFromHeaders均已存在。
- 行为：Setup-token 查询 SHALL 使用已有被动采样，成功响应缺少5h头及429早退也保留7d/7d_oi；不要求profile权限或主动调用usage API。
- 验收：无5h、有效/过期5h、显式0、无样本、429以及单查/force/passive/batch一致；SQLite Extra与固定quota窗口不回退。

### F02 patch sites / 前置 / 验收

- service/openai_apikey_responses_probe.go：ProbeOpenAIAPIKeyResponsesSupport；openai_compat.NormalizeResponsesSupportMode与模式常量已有，新增检查约5行，不引入CN平台。
- 行为：系统 SHALL 在force_responses或force_chat_completions模式跳过能力探测；auto及无效模式按现有规则探测。
- 验收：两个强制模式请求数/写Extra数均0；auto/非法/非字符串模式仍探测，凭据覆写和取消正常。

### F03 patch sites / 前置 / 验收

- service/openai_ws_v2/passthrough_relay.go：observeUpstreamMessage；本地relayTurnTiming与activeTurn齐全。上游测试的FirstTurnStartedAt/StartedAt本地不存在，不能顺带引入计时接口改版。
- 行为：系统 SHALL 为每个WS轮次独立记录首token耗时，无response.id的delta归属当前活跃轮次。
- 验收：至少两轮无ID delta、带ID delta、负时差、重复terminal与取消；保留Fast每轮service_tier/模型冲突和usage隔离。

### F04 patch sites / 前置 / 验收

- service/openai_ws_http_bridge.go：proxyOpenAIWSHTTPBridgeTurn；stageCodexFingerprintIDs、resolveCodexFingerprintIDsFromRequest、applyCodexFingerprintClientMetadataRaw已存在，约21行接入。
- 行为：系统 SHALL 在Codex WS HTTP桥接构建上游请求前设置账号级指纹，重试和切换账号时清除旧暂存值。
- 验收：最终HTTP请求header/body同账号稳定、跨账号隔离；无prompt_cache_key/显式客户端key/API Key/Grok负对照；Fast策略不变。

### F05 patch sites / 前置 / 验收

- pkg/apicompat/responses_to_anthropic.go、responses_to_chatcompletions.go；新增responsesReasoningText(22行源文件)，只依赖strings和已有ResponsesOutput。
- 行为：Responses转Messages/Chat SHALL 优先保留非空summary，否则保留content.reasoning_text；encrypted_content签名独立保留。
- 验收：summary优先、不重复拼接、仅reasoning_text、仅签名、普通文本及两条真实缓冲网关路径；usage/dedup不重复。

### F06 patch sites / 前置 / 验收

- pkg/apicompat/types.go、responses_stream_event_wire.go：MarshalJSON/messageContentWire；responsesItemWire/outputTextPartWire已有，不重做历史序号修复。
- 行为：完成响应与流式item SHALL 使用一致message序列化，空content为数组，refusal、annotations及logprobs保持字段语义。
- 验收：空message、output_text、refusal、annotation/logprobs、tool_search/function_call；JSON及SSE终态一致，序号0不被省略。

### F07 patch sites / 前置 / 验收

- server/middleware/api_key_auth.go、api_key_auth_google.go：GetByKey错误出口；context.Err、AbortWithStatus、logger现有；新增小型共用abortIfAPIKeyLookupClientGone。
- 行为：API Key查询因当前请求上下文取消而失败时 SHALL 按客户端离开结束；有效请求的真正仓储错误仍500，业务错误保持原分类，日志不含凭据。
- 验收：取消/截止、真实DB失败、格式错误Key、503过载及401错误；不把所有context错误无条件吞成499，不记录Key。

### F08 patch sites / 前置 / 验收

- handler/openai_gateway_handler.go：handleFailoverExhausted；MatchRule、ExtractUpstreamErrorMessage、stream-aware错误出口已有；提取共用helper约39行。
- 行为：耗尽failover后，已有管理员413透传规则 SHALL 先于内置body-too-large响应生效，账号切换行为保持。
- 验收：有/无匹配规则、自定义状态/文案、skip-monitoring、已开始SSE及Retry-After；不提前终止账号failover。

### F09 patch sites / 前置 / 验收

- pkg/antigravity/stream_transformer.go：StreamingProcessor.ProcessLine；json.RawMessage、GeminiResponse均已有。
- 行为：流转换器 SHALL 按response信封是否存在选择解包形态，直接Gemini响应的文本、工具和用量不能静默丢失。
- 验收：直接/包装响应、空/非法JSON、工具、usage和错误尾帧；普通流与缓冲兼容路径不产生假完成。

### F10 patch sites / 前置 / 验收

- service/antigravity_gateway_gemini.go：cleanGeminiRequest；新normalizeGeminiResponseSchema约31行，DeepCleanUndefined/CleanJSONSchema已有；只作用Antigravity消费者。
- 行为：Antigravity转发 SHALL 将responseJsonSchema转换成上游支持的responseSchema，已有responseSchema优先。
- 验收：camel/snake命名、双字段优先、非object、undefined清理、原body不变正例；不修改原生Gemini协议。

### F11 patch sites / 前置 / 验收

- service/account_test_service.go：testOpenAIImageAPIKey的payload；沿用既有image测试入口、响应parser及大小配置，不拷入OpenAIImageTestSize缺失基座。
- 行为：既有GPT Image API Key账号测试 SHALL 不发送被上游拒绝的response_format，并继续解析默认base64响应。
- 验收：严格模拟上游检查无response_format、模型/prompt/n保留、有效图片解析、上游错误；不新增模型/尺寸路由。

### F12 patch sites / 前置 / 验收

- service/gateway_upstream_request.go、gateway_anthropic_passthrough.go、account_test_service.go；新增buildAnthropicMessagesURL为26行net/url+strings独立文件；仍先经过URL验证。
- 行为：现有Anthropic自定义base URL消费者 SHALL 仅在精确Cloudflare账号REST messages路径省略beta查询，其余上游保持原行为。
- 验收：精确HTTPS host/443/账号路径、相似域名/额外path/query/fragment/rawpath/userinfo负例、Cloudflare native gateway保留beta；三条调用链一致。

### F13 patch sites / 前置 / 验收

- service/bedrock_stream.go：bedrockEventStreamDecoder.Decode；从#8518摘两项常量和约6行边界判断，不依赖#8484队列预算或ojson。
- 行为：EventStream解码 SHALL 在分配和切片前验证headers长度、payload长度及总长关系；合法帧继续校验CRC。
- 验收：header越界、total<16、header>128KiB、payload>16MiB、truncated/CRC、合法frame；大长度仅喂prelude，证实分配前拒绝。

## 4. 第二档：手工适配与跨路径验证

| 簇 | 来源 | 本地缺口与依据 |
|---|---|---|
| S01 Codex Responses非流式与instructions回显 | PR7979, PR8499 | OAuth规范化后覆盖reqStream，SSE转JSON只提取completed；本地empty instructions拒绝探针命中，incomplete原样返回SSE。两PR改同一响应链需合并适配。 |
| S02 Grok 委派消息最小适配 | PR8465 | 实际patchGrokResponsesBodyWithClientTools保留agent_message原样。上游sanitizeGrokResponsesModelInput本地不存在；v0.2.19基座562行、源acce29af2是大混合提交，不能整取。 |
| S03 已使用 Terra 的目录与计费后续 | PR8365 | 本地Terra仍使用1050000窗口/默认medium目录、minimal不拒绝、订阅豁免仅Sol；模型已在a541d0b61选定落地，故本次属第二档维护，不再把它当未选新模型。 |
| S04 请求体预算、读写超时与错误分类 | PR8484, PR8518, PR8457 | 当前只有单请求MaxBytesReader与64MiB解压上限，无实例预算；#8484单取有异步生命周期/重试累积缺陷，#8518修正必需，#8457统一错误分类。 |
| S05 上游流、drain与会话缓存有界 | PR8484, PR8518 | Gemini多处ReadString、Antigravity ForwardUpstream io.ReadAll无界；会话缓存sync.Map只按访问过期，未整体限制数量/字节。断开用量不能靠无期限drain保障。 |
| S06 Antigravity 系统身份保留任务 | PR8318 | 现有Claude身份处理没有Codex开头，Responses→Anthropic转换前未中和，可能丢带任务的前导；Gemini已有身份时提前返回遗漏其它系统块清理。 |
| S07 图片 INTERNAL500 不牵连整账号 | PR8521 | retry loop一律handleInternal500RetryExhausted，图片错误计入整账号永久禁用阶梯；本地已有模型级限流仓储与缓存能力。 |
| S08 Gemini 工具媒体原生承载 | PR8430 | 当前tool_result图片被抽出到另条user消息，native functionResponse关联丢失；不是新增生图或新Gemini平台。 |
| S09 使用记录导出限流恢复 | PR8436 | 用户UsageView导出仍100条逐页、不处理429；API client包装错误丢Retry-After。 |
| S10 Antigravity 测试提示词贯通 | PR8531 | routeAntigravityTest已有prompt却丢给不接收prompt的testAntigravityAccountConnection；底层固定点号和1 token。上游WithPrompt基座本地缺失，不能将两文件补丁判第一档。 |
| S11 RPM空覆盖缓存与既有JSON预检 | PR8518 | snapshotFromAPIKey只存非nil覆盖，checkRPM遇nil逐请求查库；filterThinkingBlocksInternal和normalizeCodexDelegationBootstrap有现存消费者。优化收益需实施benchmark，不能引用上游数字。 |

### S01 patch sites / 前置 / 验收

- service/openai_gateway_passthrough.go、openai_gateway_response_handling.go、openai_gateway_forward.go、openai_gateway_request_body.go与handler/ops_error_logger.go；新terminal helper约15行、instructions echo文件161行，依赖现有gin/gjson/sjson/SSE解析。
- 行为：网关 SHALL 保留客户端原stream选择，将非流式SSE的completed/incomplete/cancelled转换成单JSON；缺Content-Type的OAuth SSE可识别。允许空instructions进入既有规范化，并将网关改写的instructions回显还原为客户端值。
- 验收：HTTP/透传/普通/compact，stream省略/false/true；incomplete/cancelled/canceled、失败终态仍失败；instructions缺失/null/空/多行与failover隔离；回显不泄漏注入内容，usage只落一次。

### S02 patch sites / 前置 / 验收

- service/openai_gateway_grok.go：sanitizeGrokResponsesInput及patchGrokResponsesBodyWithClientTools；pkg/apicompat/chatcompletions_responses_bridge.go已有agentMessageText（重命名导出即可），选择局部raw-item转换，不引入562行回放/图片/call-id重配器。
- 行为：既有Grok Responses入口 SHALL 将agent_message的input_text及明文encrypted_content按原顺序转为user消息，空消息跳过，不尝试解密。
- 验收：封套+正文顺序、空/非文本片段、字符串content、非委派item字节/大整数保真；additional_tools原提升规则、Chat对照、HTTP/Grok usage和模型审计保持。

### S03 patch sites / 前置 / 验收

- pkg/openai/constants.go与新增176行codex_gpt61_terra.json、pkg/apicompat两转换器、service/openai_codex_models_service.go、openai_codex_model_metadata.go、openai_compat_prompt_cache_key.go、各compat校验及openai_gateway_usage.go。复用IsOpenAIOAuthLike和账单策略；保留Fast账号catalog优先，不能照抄上游ultrafast断言覆盖账号声明。
- 行为：已接入Terra SHALL 使用冻结tag中的Codex descriptor作为受限fallback，Codex默认low及500000窗口与max_context_window=872000区分，API兼容默认medium/none仍合法，minimal明确400。OAuth/setup-token长上下文豁免按最终发送模型决定，API Key保留272k严格边界与Fast结算。
- 验收：none/minimal/max/ultra的API与Codex不同标尺；别名/映射后校验，账号raw/metadata/空目录优先、ETag；271999/272000/272001×OAuth/setup-token/API Key×default/priority/ultrafast/Free Fast/显式零价与SQLite去重。

### S04 patch sites / 前置 / 验收

- pkg/httputil/body.go及预算/读写timeout helper、server/http.go、middleware/request_body_limit.go、routes/gateway.go、既有handler读body出口；只取#8518 WithRequestBodyMemoryScope/DetachRequestBodyMemory及释放消费者，跳过缺失异步生图/Voice等路由。
- 行为：所有既有网关入口 SHALL 共享实例级请求体内存预算，读取/解压/修补/异步移交/failover释放按实际所有权计数；超预算503+Retry-After，超时408、断开499、截断/解码失败保持安全分类。超时按单次读写停滞计，不限制正常SSE总时长。
- 验收：1GiB限制环境测RSS/并发吞吐/503恢复；gzip/zstd/lenient/unknown长度、读取/写入停滞、正常长流、WS Hijack、重试100轮/取消无残留；日志无正文/密钥。默认值须量测，不能直接照搬1GiB预算或500MiB行上限。

### S05 patch sites / 前置 / 验收

- service/gemini_messages_compat_service.go、antigravity_gateway_upstream.go、各SSE handler、repository/http_upstream.go、pkg/httputil/upstream_drain.go；openai_messages_continuation/digest_session/turn-state缓存及关闭链。session_memory_cache源244行、sse_transport_limits最终195行；只适配已有消费者，不恢复SQLite后台skip。
- 行为：既有上游流 SHALL 限制SSE队列字节/单行/累计事件，非流式读取有界；客户端离开后的usage drain有总时限，缓存按数量和字节逐出并释放后台任务。
- 验收：持续短行/长行/慢消费/取消均有界；超大单条事件不越行上限；截止前已观测用量仍入SQLite且去重，截止后不虚构完整用量；缓存键含账号/会话隔离，过期/逐出/关闭无goroutine泄漏。

### S06 patch sites / 前置 / 验收

- pkg/antigravity/request_transformer.go、service/antigravity_gateway_compat.go、antigravity_gateway_service.go；导出NeutralizeSystemIdentity并在转换前处理system/developer，仅gjson/sjson与现有cleaner。
- 行为：Antigravity路径 SHALL 仅规范化开头的已知客户端身份和attribution，保留附随任务及中间引用；原生Anthropic不受影响。
- 验收：Claude/Codex多种前导、紧邻逗号/任务、引用/相似字符串、双身份、string/parts、native Anthropic对照；失败与成功请求body归属不混。

### S07 patch sites / 前置 / 验收

- service/antigravity_gateway_retry.go、antigravity_internal500_penalty.go；resolveFinalAntigravityModelKey、setModelRateLimitByModelName、updateAccountModelRateLimitInCache已有；摘handleInternal500ForModel，新增环境可调惩罚另列第三档。
- 行为：现有Antigravity图片模型重试耗尽INTERNAL500 SHALL 只冷却最终映射模型，并同步缓存；文本模型仍用既有账号惩罚。
- 验收：alias映射图片/文本、相同模型计时、其他模型仍可调度、SetModelRateLimit失败、SQLite Extra/调度缓存一致；不增加账号处罚计数。

### S08 patch sites / 前置 / 验收

- pkg/antigravity/gemini_types.go/request_transformer.go、service/gemini_messages_compat_service.go及Chat/Antigravity compat消费者；新增pkg/gemini/SupportsMultimodalFunctionResponses为20行stdlib helper，参数沿最终映射模型传递并更新所有测试调用。
- 行为：Gemini3及以后已使用模型的工具媒体 SHALL 保留在functionResponse.parts；旧模型按已有分离图片回退，关联工具名/顺序保持。
- 验收：Gemini2.5/3/4和别名、base64/URI/text混合、未知工具ID、重试去签名、原输入不变、tool_order及图片用量/dedup不重复。

### S09 patch sites / 前置 / 验收

- frontend/src/api/client.ts、views/user/UsageView.vue及中英dashboard；保留已有AbortController/请求代际和账户切换保护，等待需可取消；不改xlsx或增加支付UI。
- 行为：现有用户导出 SHALL 每页至多1000条，在429时有界等待并重试同一页；重试耗尽或取消不能下载部分结果。
- 验收：数字/缺失/日期/非法Retry-After、上限等待/重试、分页过滤一致、关闭/卸载/换号取消、不重下/漏行；实际浏览器保存导出及失败恢复。

### S10 patch sites / 前置 / 验收

- service/account_test_service.go的route/test调用链及antigravity_gateway_service.go TestConnection/build*TestRequest；用窄适配贯通参数，无需引入Argon/新图片链；现有签名调用者需一起更新。
- 行为：已有账号测试 SHALL 将管理员输入提示词传至Antigravity Claude/Gemini请求，并返回真实回复；空值沿用默认提示词。
- 验收：自定义/空/Unicode/多行prompt、Claude/Gemini模型、响应文本/错误/取消、定时测试旧调用兼容；真实本地HTTP捕获body，外部账号测试不自动执行。

### S11 patch sites / 前置 / 验收

- service/api_key_auth_cache*.go、user.go、billing_cache_service.go；从本地snapshot20递增而不是复制上游27，保留Fast字段。gateway_request.go与handler/openai_gateway_handler.go仅摘既有消费者的stdlib/gjson预检；整ojson批量patch/COW/缺失compact算法不取。
- 行为：鉴权快照 SHALL 区分已加载的nil RPM覆盖与加载失败，并按group隔离；既有thinking/委派预检仅在可证明无修改时跳过解析，不能改变输出语义。
- 验收：nil/0/正值/错误/跨group、失效与并发、miniredis+SQLite查询次数；escaped/重复键/非法JSON/长输入byte等价，benchmark ns/op+B/op+allocs/op前后实测，无收益不落优化hunk。

## 5. 暂不做与历史遗留复查

### 5.1 本轮第三档：需要用户按实际用途选定

| 项目 | 判断依据与边界 |
|---|---|
| #8389 管理MCP / OAuth | 78文件、+7250行及上游249迁移，是新远程管理授权面，不是Claude Code网关修复。未来选用也需单独剔除支付操作、做SQLite schema/撤销/授权边界设计。 |
| #8504 精确邮箱黑名单 | 33文件、+2647行，新注册管理策略及批量禁用；个人自用不默认采用。 |
| #8462 图片尺寸分流 / #8557 GPT Image2.6 | 尺寸账号路由、粘性策略、原生编辑和complexity为新能力。本地旧图片接口继续维护，F11只修既有测试参数。 |
| #8543 / #8521可配置惩罚 | 关闭传输冷却和环境可调处罚是新运维选择。S07只修图片错误扩大到整账号的作用域。 |
| #8522 Antigravity quota-summary | 新上游接口与真实窗口分组展示，包含独立Argon组；缺的是消费者而不是简单改文案，不为当前周窗口F01带入。 |
| #8380 / #8495 导出策略 | 使用Sol作为默认、API Key mode默认、live web search与native image actor marker都改变生成客户端配置；不是当前导出文件无法生成的修复。#8495的5.6中间值被tag内#8380覆盖，若选择只评估终态。 |
| #8518完整JSON性能体系 | 本轮只选资源必需生命周期、Bedrock边界、RPM空值缓存和既有thinking/委派预检。ojson批量patch、RequestBodyRef完整COW、缺失compact/orphan基座及全套隐私快照优化需要测量收益再决定，不以“性能”名义导入整46文件。 |

这些项本轮只做分档和范围说明，未生成实施任务或导入代码；需要用户另行选定。

### 5.2 第四档与不成立的修复

- #8503修改原生Kimi并发403冷却，本地没有对应平台/ratelimit消费者；不是共享OpenAI-compatible的通用403修复。
- #8476修补细粒度管理员Key（含兼容全权模式），依赖未选#8319框架。本地已有普通admin auth不等于这个基座；不导入33个NOBASE文件来“修复”。
- #8549针对transcript哈希前缀阻断。本地`CyberSessionBlockKey`仅接受显式会话ID并按Key隔离，无显式ID就放行；独立对照探针通过。上游旧transcript文件157行、本地不存在；添加它会改变原有策略。PR同时改go-redis/miniredis版本，不是本地修复前置，不带入。
- #8471非法UTF-8导致PostgreSQL参数异常的案例不能声称在SQLite复现；F07只承接实际可达的取消分类。是否追加输入防御需另列明确验证，不引用PG故障当SQLite证据。
- #8457的Grok Voice、cached contents、异步image task等缺失入口及#8484对应消费者不引入。存在通用Grok图片或Gemini接口，不代表Voice/显式缓存CRUD也存在。
- 支付、充值、付费订阅购买/续费、退款及相关UI/依赖/schema始终排除。MCP/admin权限混合PR中的支付操作也不能作为通用权限前置导入。
- 上游VERSION、赞助商文件和PG迁移不合；fork维护自己的1.1.x版本。背景任务有缺陷时按SQLite SQL适配，不能再用skipSQLiteBackgroundJobs关掉已恢复的功能。

### 5.3 旧轮次逐项当前口径

旧报告是冻结历史，下表覆盖当前状态；不改写旧source-baseline、原任务快照或原始验证数据。下列“完成”依据本地实现/归档记录，不宣称本次重跑了每轮全量验收。[本轮旧项符号复查](./evidence-0.2.20/older-scope-search.txt)只辅助判断现有消费者，不能用关键词命中证明整个能力完成。

| 旧轮次 / 项目 | 2026-11-10状态与保留原因 |
|---|---|
| 0.2.19第一/二档 | 683354b2c、cccf931ef完成并已归档；本轮以这些成果为基座，不重复创建旧F/S任务。 |
| 0.2.19 Terra | a541d0b61已完成且用户实际使用；本轮S03是新tag后续，保留API与Codex标尺差异，不能继续写“未选新模型”。 |
| 0.2.19 Grok诊断 / 最老工具桥尾项 | becd2d2a3完成Grok诊断、非法参数、namespace/deferred及Responses序号尾项，2026-11-01归档；新的Grok委派转换S02不等于重开这些已完成任务。 |
| Fast / Free Fast / Ultrafast跨历史轮次 | 326513cdc已完成、SQLite228、auth snapshot20；不再列第三档未实施。动态长上下文阶梯、独立effort倍率、usage compaction/requested effort schema仍未整体采用。 |
| 输出TPS / ops分位数 | f333aa5f3已完成全窗口P5/P10/P50/均值/样本数；usage日志TPS及完整ops入口已完成。新的趋势/额外运维展示仍按用途。 |
| 0.2.18主体 | 2b79ac320完成第一/二档；Go1.27.3与已有凭据路由、退役、断流计量、图片下载边界保留。 |
| 0.2.18个人功能 | 0e38d06d8完成筛选记忆、运维账号名、模型级冷却、OpenAI搜索、min_output_tokens及日志TPS；不是新MCP/图片路由授权。 |
| 0.2.18个人维护 | b554bea55完成SQLite清理、到期、定时测试、操作日志清空、#8249/#8250；1.1.23还有清空/写入顺序与请求生命周期修复，不再把后台禁用当遗留。 |
| 0.2.18 #8251 / #8269 | 完整共享解码快照/COW和SQLite轻量quota查询仍需所有权/benchmark。S11只选明确现有热点，不吸收整簇；PG JSONB实现排除。 |
| 0.2.18 Gemini/Vertex | Argon新模型、思考摘要、service-account quota范围、多平台生图开关、显式cache CRUD、Vertex API Key仍按需；S08维护已有模型工具媒体，不扩平台。 |
| 0.2.19其它第三档 | #8319受限管理员Key、#8323实际图片张数计量、#8348上传20→50MiB、#8287同账号重试配置、#8395更广目录扩展仍按需；现有上传完整性、fileData用量和声明目录已完成。 |
| 0.2.16主体 / 新能力 | 7c544c03f、847db09e1、5ff6e5bf9主体完成；Haiku5.5/group i18n等仍按用途。账号模型冷却与TPS已由后续覆盖。 |
| 0.2.12 | 405246789主体、优先级快调/Key排序完成；TypeSafe缺平台，支付推广/查单第四档，旧axios维护已被后续覆盖。 |
| 0.2.11 | 96f32d018主体完成；Claude原生额度查询/重置、风控白名单、Key创建限制、在途预占仍按需。Ultrafast已完成，不能沿用这轮旧延期标签。 |
| 0.2.9第一/二档 | 第一档用户2026-10-24确认完成并归档；第二档159d04222完成。OpenCode/Command Code/Kimi原生、plugins、完整model_allowlist glob/pinned/live、视频路由、多实例备份运营仍未整体采用。 |
| 0.2.5 quota/WS尾项 | b349c7408完成33/33及SQLite227；默认WS倍率1.0是明确选择，5.0补丁不采纳。S04/S05不得覆盖后续固定窗口或成功去重。 |
| 0.2.5调度 | #6960/#7049 exec_scope、抢占/queue-wake仍需并发设计；本轮强制模式探测与预算不需要这套调度前置。 |
| 0.2.5平台与模型 | #6575/#6890/#6974/#6769的Gemini新能力/原生Codex图片/峰谷/Ollama异步重置；#6747/#7091/#6691/#7153平台/批量能力；#6989/#7057/#6968/#7011/#6945/#6988更广目录/compact/pinned，按实际用途；支付#6986排除。 |
| 0.2.5其它 | Grok全流序号#7129、input stripper#7085、Gemini带内错误#6929、监控BaseURL#6917和管理大展示未整体采用。新完成序号修复不代表这些整簇全部完成；S02不借机导入旧562行Grok sanitizer。 |
| 0.2.4 / 0.2.0 | a26b6b55d / c9d9bebe8主体完成；5b3e1a716完成0.2.0剩余128/128任务，类型桩/四面浏览器/组合门禁已核销。 |
| 0.1.180 / 0.1.184定价和usage | 模型参考价格已由5250907f5完成；Fast已完成。动态长上下文阶梯、usage compaction/requested effort持久化、峰谷/重置锚点仍分别按需或需schema；本轮Terra不引入整机制。 |
| 0.1.180工具桥/Grok/WS | 既有PDF、tool身份、HTTP回放/孤儿工具、composite闸门和四个历史尾项已完成；Grok429/Realtime/抢占整簇按用途。跨节点Redis租约单节点不需要；#5925最小delay已随后续容量消费者落入，不代表整簇完成。 |
| 更早主体 / 固定窗口 | 0.1.176/179/183/184主体、Responses Lite已验收；128c5f7bf固定窗口推进及227/228语义必须保留，不为上游同名优化回退。 |

### 5.4 依赖维护的当前证据

本地已是vue-i18n/intlify compiler11.4.16、decode-uri-component0.5.1覆盖、Node24 CI、follow-redirects1.16.2、xlsx官方0.20.3包；d851a3847主要行为已满足，整文件CONFLICT来自本地更多维护。

此次`pnpm audit --prod --json`为**14条advisory：0 high / 12 moderate / 2 low**，涉及yaml、uuid、mermaid、colord、katex。旧报告中的xlsx两项high不能继续当作今日未解决项。pnpm11的package.json overrides警告保留，但对应覆盖已在pnpm-workspace.yaml镜像；不是本轮覆盖失效证据。[原始输出](./evidence-0.2.20/frontend-production-audit.log)、[摘要](./evidence-0.2.20/dependency-summary.json)。未做本轮全开发依赖或Go漏洞全扫，不用生产14条代替全量安全结论。

### 5.5 发布之后的main观察，单独保留边界

冻结main较v0.2.20还有9个非merge提交，见[列表](./evidence-0.2.20/post-release-commits.txt)。其中：

- #8610自适应请求体预算/cgroup读取，恰好说明1GiB静态默认的适配需求；**不在v0.2.20 tag**，S04不自动取得其代码。实施若选用须另补来源/评估。
- #8606缓存写TTL与零输入价格保留Fast倍率，和本地已完成Fast/计费相关，建议下一轮优先复核；本轮未复现、不宣称已修或已纳入两提案。
- #8590用量刷新按钮、#8603测试隔离、#8607图片格式说明及VERSION同步同样不是此次release差异。

## 6. ALREADY、N/A与最小前置的解释

- 文件态只有3条ALREADY，不能据此说只合过3项；d851a3847有两个CI文件ALREADY，其余Node24/vue-i18n/override语义已经满足。
- #8549虽7个CONFLICT，却是本地没有transcript算法的N/A。补入157行算法并非修复，会改变显式ID策略。
- #8465的NOBASE不意味着Grok委派不存在。现有Chat转换已有agentMessageText，Grok入口缺调用，局部适配足够；562行通用清理器不是必需前置。
- #8531上游仅2文件且看似小，但本地prompt丢失跨AccountTestService→AntigravityGatewayService，故明确列第二档；不能只改固定点号。
- #8518标题是性能，但包含可独立复现的Bedrock越界，列第一档F13；资源生命周期修正随S04/S05，完整COW不随行。
- #8365 Terra模型已经被用户选择并实施，本轮维护自然进入第二档；其Codex descriptor中的service_tiers不能覆盖本地账号显式声明，尤其Fast工作区刚完成的优先级。

## 7. 建议落地顺序

1. F13 Bedrock分配前边界可独立先做；随后F01/F02/F03/F04与F05/F06，优先Claude/Codex及共享数据正确性。
2. S01 Responses终态/instructions、S02 Grok委派、S03 Terra一起安排跨路径测试，但按簇独立提交；与Fast共同文件按顺序手工整合。
3. S04与S05按最终tag组合设计预算、drain、取消/错误语义，纳入#8518生命周期修正后再做1GiB压测；不能先落已知有预算回收问题的中间态。
4. F07/F08共享错误，S06/S07/S08及F09/F10/F11/F12处理既有平台尾项；S10管理员测试提示词与S09导出交互随后。
5. S11测量后做RPM空值缓存与可证明等价的预检；变更auth snapshot必须从本地20递增，保留Force/Free Fast字段，不导入上游25/27的未选schema。
6. 全部实施后按两份verification填实际SHA和输出；归档、发布、推送与部署不是此次评估授权内容。

## 8. 本轮已执行自测与未来验收门槛

| 已执行项目 | 结果与边界 |
|---|---|
| 完整来源与四态 | 42候选覆盖72提交；562文件reverse-before-forward完成；未应用任何补丁 |
| backend测试发现与定向基线 | 10包`go test -tags=unit -list`及对应`-run ... -count=1`通过；[发现](./evidence-0.2.20/baseline-test-list.log)/[执行](./evidence-0.2.20/baseline-targeted.log) |
| 性能邻近基线 | RPM/FilterThinking/DelegationBootstrap/AuthSnapshot/Privacy，service+handler发现/执行；[结果](./evidence-0.2.20/performance-neighbors.log)，不当作性能增益数据 |
| 前端相关基线 | API client、用户UsageView、UseKeyModal、CC Switch，4文件77测试通过；[输出](./evidence-0.2.20/baseline-frontend.log) |
| 隔离缺陷探针 | apicompat3个、service10个、WS1个顶层预期行为失败，另service的cyber无transcript对照通过；forced模式3个auto对照通过。三包均实际编译执行；[复现入口](./evidence-0.2.20/reproduce-probes.py) |
| Bedrock小型畸形帧 | 16字节合法CRC帧复现slice panic，没有运行4GiB分配攻击，也没有据此宣称公网可利用 |
| 当前生产依赖audit | 12moderate/2low；退出1表示仍有advisory，非测试失败 |
| 文档与范围 | [文档核验](./evidence-0.2.20/document-validation.json)记录两份OpenSpec严格验证、链接、任务未勾选、保护文件和原index检查 |

原始WS来源测试曾因FirstTurnStartedAt/StartedAt在本地不存在编译失败，见[记录](./evidence-0.2.20/ws-probe-initial-compile.log)；已改为本地observeUpstreamMessage的确定性两轮探针，确认真正计时缺陷。该编译失败不是缺陷复现，不能算已覆盖接口。

每项命令、工作目录、时间、退出码保存在同名JSON；探针失败是确认旧行为，不是产品已修复。评估没有执行实现后的完整构建/全量测试/race/真实浏览器或1C1G压力验收，未来槽位均为空。

建议实施门禁（backend目录，经rtk执行）：

```sh
rtk go test -tags=unit ./... -count=1
rtk go build -tags embed -o /tmp/sub2api-020/server ./cmd/server
rtk golangci-lint run ./...
rtk go test -race -tags=unit ./internal/service ./internal/service/openai_ws_v2 ./internal/handler ./internal/server/middleware ./internal/repository ./internal/pkg/apicompat ./internal/pkg/httputil -run 'Test.*(Grok|Codex|GPT61|Responses|Fast|Usage|Quota|Bedrock|RequestBody|RPM)' -count=1
rtk go test -tags=unit ./internal/repository ./internal/service -run 'Test.*(Migration|ProductionSQLUsesSQLiteDialect|UsageBilling.*SQLite|PlatformQuota|FixedQuota)' -count=1
```

S09前端需pnpm typecheck、lint:check、test:run、build、critical及真实浏览器导出/429/取消/换号；S04/S05需受限内存下HTTP1/2/WS慢读慢写、压缩/重试/取消、峰值RSS和资源回收。SQLite要实际执行结算、去重、quota和228迁移保护，不能以仓储stub替代。

## 9. 本轮新增通用教训

- 冻结的是源码commit，annotated tag对象SHA与commit SHA要分列；dirty工作树期间并发提交要比较内容而非只报HEAD变化。
- 发布说明也可能写错配置命名空间；必须核对Config、默认值、Validate、示例和真实消费者。
- 一次资源治理PR可能随后被性能PR修回预算所有权；按最终行为组合，而不是只按标题摘取。
- 不能为通过来源测试就引入本地不存在的计时/缓存接口；先把断言重写到本地实际路径，再确认缺陷。
- 最小前置不是“零命中后整文件搬过来”：Grok562行通用sanitizer与已存在的agentMessageText应分别量测。
- 新模型已选落地后，后续目录/校验/计费修正是维护；API、Codex、账号raw catalog标尺不能合成一个全局默认。
- 日志/用量/用户费用/账号成本分别验证；超时drain到期不能用假完整usage掩盖数据缺失。
- 历史依赖高危清单与今日audit分开；旧xlsx结论不能永久复制，0 high也不是全量无风险。
