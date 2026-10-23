# 新增调用与最小基座复核

日期2026-10-23，fork873962ce7。`added-call-inventory.tsv`枚举每个完整候选新增Go行里的调用词法项，逐个匹配本地func声明和同PR新增定义；包含类型转换、标准库、注释、局部closure与同名方法，不能替代编译。`base-symbols.txt`是关键符号的实际rg结果，ABSENT明确记录缺口。

## 选定簇

| 簇 | 本地调用与缺口处理 |
|---|---|
| F01 | buildCodexStoredIdentityKeys及索引链存在；codexCompositeIdentityKey由同PR新增，只用fmt.Sprintf。 |
| F02 | openAIStreamEventIsMetadata与HTTP/WS消费者均存在，只扩事件枚举，不缺前置。 |
| F03 | initFallbackPricing/getFallbackPricing、动态资源读取、LongContextThresholdInclusive已有；新价卡无新服务调用。来源测试邻近的Argon用例不携带。 |
| F04 | handle403/handleOpenAI403存在；isCloudflareBotBlockResponse缺失。补旧4行HTML判断及本PRcontainsCloudflareBotBlockCode辅助，标准库encoding/json/strings/slog；不需要原OpenCode/Command Code平台。 |
| F05 | extractGeminiUsage和Antigravity三处相减都存在；max为Go内建，保持缓存/图片桶。 |
| F06 | cleanToolSchema/normalizeGeminiEnum/incrementIntegralSchemaBound存在；cleanToolSchemaNode同PR提供；仅新增递归上下文参数，不引入其它工具Schema框架。 |
| S01 | ListModelAvailabilityCandidates、IsOAuth、DetectModelPlatform、路由归属链存在；domain.ConcretePlatformIDs/LookupPlatform/UsesOpenAIGateway缺失。用本地五平台及现有路由定义表达守卫，禁止为这些调用拉平台catalog。 |
| S02 | helper已迁到openai_bare_error_stream；来源writePendingString为局部closure，不能当缺全局基座。新stamper只需已有gjson/sjson与标准库；本地错误缓冲/失败输出须沿现有路径接入。 |
| S03 | IsOpenAIOAuthLike/IsShadow/resolveOpenAIAccountUpstreamModelForRequest/loadCodexGroupCatalogAccounts均存在；新增retirement helpers同PR定义。modelCatalogEntries缺失但仅约17行标准JSON解析，可随消费者补等价实现；不引入缺失的整个openai_models_list.go、pinned/model_allowlist服务。 |
| S04 | PairCodexClientIdentity/codexUATrailerName、当前身份配对/版本策略、newOpenAIUpstreamFailoverError均存在。缺SameAccountRetryDelay是结构体字段，**func词法扫描发现不了**；需最小字段+handler读取分支，来源#5925完整merge冻结。无需SameAccountRetryDeadline。 |
| S05 | mergeAnthropicUsage、hasObservedTokens、detachStreamUpstreamContext、sseStreamErrorEventError、usage提交队列存在；新compat helper291行。同PR修改的native line pump本地不存在，按v0.2.18完整162行纯读取helper补入。opsUpstreamProxyID/Name及请求effort stamp源辅助缺失，保留本地现有归属/字段，不新增usage schema。两条CN native消费者不加入。 |
| S06 | buildParts、mergeCollectedPartsToResponse、resolveGeminiThinkingVariantForLevel、ensureGeminiFunctionCallThoughtSignatures、既有composite路由/价格/测试链都存在。新增媒体提取、别名价卡、细分套餐ID判断只用已有JSON/strings；mappedModel协议选择无需TestConnectionWithPrompt和新增图像返回类型。目录能力/Argon/hub UA hunk排除。 |
| S07 | IsVertexServiceAccount、exchangeVertexServiceAccountToken和三条Gemini入口已有；finalErrorPolicy/typed token error/429判断同来源提供。opsUpstreamProxyAttribution缺失，改用本地已有ProxyID/Proxy归属；不需要显式缓存或Vertex API Key。 |
| S08 | hasObservedTokens/observedGeminiImageOutputs已有，geminiStreamAbortedWithBillableOutput同PR提供；源hunk中思考摘要来自第三档#8114，不导入，按本地读循环手工接入。 |
| S09 | IsBlockedHost/isBlockedHostname/isPrivateIP/isPrivateOrLoopbackHost/detectedImageContentType已有，新增transport只用标准库。特定URL校验、DNS/代理、重定向和响应体释放须行为测试，CLEAN不是安全证明。 |
| S10 | 无新Go应用调用；必须同时维护Go版本字串与pnpm两个override位置。仅计划依赖目标，不在评估中修改锁文件。 |

## 排除与暂缓的关键反证

- #8079本地没有GetAnthropicProtocolBaseURL/APIProtocolAdaptive，wrapper虽小但没有消费者配置；不提升档位。
- #8087的cn_provider_quota_service.go缺失；#8224本地无aliasDeepSeekResponsesInputImages/normalizeDeepSeekResponsesRequestBody链。
- #8251 GetSnapshot当前调用decodeCachedAccount逐次JSON解码；原地写冷却map确实存在，但没有直接证据证明该PR新增的共享母本契约已经在生产启用。新缓存/COW暂缓，不能单列为已确认现存崩溃。
- #8269有大量PG JSONB表达式；本地SQLite结算是替代实现且已具固定窗口修复，不能以新增函数同名就搬SQL。
- #8093目录扩展、#8114新思考输出、#8166扩配额类型、#8167扩大生图开关都非修复前置，用户明确暂缓。

源码存在性核对不等于新补丁编译通过。本次未编译或应用拟移植产品代码；实际编译运行的是冻结本地的相关测试与独立overlay探针。
