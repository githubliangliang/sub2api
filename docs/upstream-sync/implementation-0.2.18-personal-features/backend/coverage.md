# 后端实现与验收映射

实施起点：`67b1a07270acd79f5d54432b44ae396c3f7e360c`及已生成、未提交的个人功能spec；冻结来源仍使用原source-baseline，不改写。

2026-10-26完成补记：最终54个后端测试包、完整build、embed资源和五包定向race通过，包含最后跨平台凭据标记修正；真实SQLite运维查询/聚合/TTFT排序、模型/账号范围与内嵌Redis、固定配额窗口/去重通过。完整命令与最终日志以[统一验收](../../../../openspec/changes/port-upstream-0.2.18-personal-features/verification.md)和[最终门禁目录](../completion-2026-10-26/)为准，早期红例/诊断不删除、不作最终结果。

## 已执行检查

| 证据 | 内容与结果 |
|---|---|
| red-initial.txt | 实现前运行时失败：显式account scope被丢弃，OpenAI API Key搜索模式仍default，显式非法min_output_tokens被忽略；退出1 |
| targeted-initial.txt | 首轮编译诊断：具体搜索manager管理方法与本地shouldFailover签名不同；已按调用方修正，保留原始失败 |
| targeted-features.txt | 四包配置/转换/规则/搜索/真实SQLite定向通过，退出0 |
| race-features.txt | config/apicompat/service/repository受影响链路race通过，退出0，分别1.169/1.243/4.086/2.025秒 |
| sqlite-ledger-and-scope.jsonl | 新SQLite+内嵌Redis规则范围，以及原SQLite固定配额窗口/全结算/删除Key/去重回归，23个测试节点通过，退出0 |
| red-event-cancellation.txt / green-event-cancellation.txt | 交叉审查发现SSE helper内部取消仍写后续事件/伪成功；新增10事件×Write/Flush取消和JSON写入内取消回归，先失败后通过 |
| green-frontend-review-overlay.txt / race-event-cancellation.txt | 前端agent原独立overlay复现转绿；事件边界/handler定向race通过，退出0 |
| red-cross-platform-credential-markers.txt / green-cross-platform-and-cancellation.txt | Anthropic/Grok/Gemini明确凭据标记曾被显式model规则吞掉；仅解除已知标记的平台限制，跨4个已有平台×3标记+普通模型401控制组转绿，三包定向unit退出0 |

最终全后端unit/build和真实Messages handler搜索检查由主agent统一执行，结果见同批checks目录与verification。未执行真实搜索或上游凭据请求。

## F04 场景映射

| 场景 | 主要证据测试 |
|---|---|
| 显式账号级与模型级/诊断scope | TestHandleTempUnschedulable_ExplicitAccountScopePausesWholeAccountWithKnownModel；TestCheckErrorPolicy_ExplicitModelScopeKeeps401ModelScoped；TestPersonalFeatures_ExplicitAccountScope |
| 旧缺省/非法scope解析 | TestGetTempUnschedulableRules_ParsesScope；TestTempUnschedRuleUsesModelScope；前端表单round-trip由前端agent验收 |
| 首个命中/写失败不换范围 | TestTempUnschedScope_FirstMatchWriteFailureDoesNotFallThrough；TestOpenAIFailoverSideEffects_401FollowsTempUnschedScope中的首条account案例 |
| 映射后实际模型/调度 | TestRateLimitService_ModelTempUnschedulableIsolatesSchedulerByModel；TestCountTokensEntries_ExplicitModelScope401OnlyCoolsModel；TestAnthropicFailoverEntries_ExplicitModelScopeOnlyCoolsModel；新SQLite测试检查DB与实际调度缓存 |
| 图片/文本隔离 | 原TestIsModelRateLimited_AntigravityImageDoesNotInheritGeminiFamilyLimit；生产模型键逻辑未改写 |
| 未知模型/到期 | TestHandleTempUnschedulable_ExplicitModelScopeFallsBackToAccountWhenModelUnknown；TestTempUnschedScopeSQLiteWithEmbeddedRedis的model重置时间过去后恢复 |
| 模型写失败不全局熔断 | 原TestOpenAIModelTempUnschedulable_WriteFailureDoesNotRuntimeBlockWholeAccount；模型错误写路径保留原语义 |
| API Key模型401 | TestCheckErrorPolicy_ExplicitModelScopeKeeps401ModelScoped；TestHandleUpstreamError_ExplicitModelScope401 |
| 凭据无效/重复401 | TestOpenAIFailoverSideEffects_401FollowsTempUnschedScope；TestCheckErrorPolicy_401FollowsTempUnschedScope；TestTempUnschedScope_SetupToken401KeepsCredentialHandling；TestTempUnschedScope_CredentialMarkersAcrossExistingPlatforms确保明确token撤销/失效/detail Unauthorized不被非OpenAI的model规则吞掉 |
| account范围真实运行时阻塞 | TestTempUnschedScope_ExplicitAccountUsesRuntimeBlock，防止known-model被误认为模型级而跳过全局阻塞 |
| SQLite/miniredis/快照/账本 | TestTempUnschedScopeSQLiteWithEmbeddedRedis使用InitRedis(redis.enabled=false)、Ent真实SQLite、NewSchedulerCache、NewTempUnschedCache；sqlite-ledger-and-scope.jsonl同时包含原固定窗口和删除Key结算回归 |
| 关闭/未匹配/原池策略 | 原TestRateLimitService_HandleTempUnschedulable_PoolModeWithoutCustomPolicySkipsState、PoolModeCustomPolicyUsesModelScope，以及匹配/解析既有测试 |

模型错误仍使用已归一的真实上游调度键，不二次映射；Bedrock账号类型使用调度器读的映射模型名而不是AWS wire ID。已有Antigravity/Grok模型键系统、真实配额窗口和错误策略顺序未改写。

## F05 场景映射

| 场景 | 主要证据测试/实现 |
|---|---|
| 全局/账号三态/渠道继承/旧bool | TestShouldEmulateWebSearchFor_OpenAIEligibility；原TestShouldEmulateWebSearch_*及TestGetWebSearchEmulationMode_* |
| nil manager/服务 | 原TestShouldEmulateWebSearch_NilManager；TestSetWebSearchManager_TypedNilRemainsUnavailable；eligibility的nil setting/account |
| OAuth/SetupToken/影子/混合工具 | eligibility含8类账号与混合工具反例；ForwardAsAnthropic仅对OpenAI平台尝试新增拦截，Grok保持原路径 |
| JSON/SSE/唯一终态/最近查询/零LLM | 主agent的TestOpenAIWebSearchMessagesHandler_UsageAndSlots，以本地窄接口提供商桩验证实际handler路径 |
| 取消/提供商/写失败/代理换号/释放 | TestOpenAIWebSearchMessagesHandler_FailuresReleaseSlots与ProviderFailoverReleasesAttempts；TestWebSearchStream_CancellationAtEveryEventBoundary对10个事件的Write/Flush分别取消，禁止观察到取消后继续写/flush；TestWebSearchJSON_CancellationDuringWriteDoesNotReportSuccess；主agent和backend子目录保存红/绿日志 |
| simple mode/归属与延迟队列 | handler测试显式simple mode，成功结果为零Token，延迟worker确定性复用Gin后仍记录本次RequestedModel及账号/用户/Key；本批在入队前快照ChannelUsageFields修复实际红例 |
| Anthropic既有行为 | shared shouldEmulate wrapper保留原接口，原TestShouldEmulateWebSearch_*通过；响应输出改为真实零模型用量，错误/取消不再伪成功 |

搜索执行依赖通过现有SetWebSearchManager接收仅SearchWithBestProvider接口，原生产*websearch.Manager调用和typed nil兼容；私有getWebSearchManager仍返回具体管理器以保留TestSearch/GetUsage/ResetUsage管理功能。没有新公开测试setter或URL配置。创建、编辑与渠道保存由前端agent提供证据。

## F06 场景映射

| 场景 | 主要证据测试 |
|---|---|
| 默认/YAML/env/优先级 | TestLoadMinOutputTokensDefaultsAndConfiguration |
| 显式0/负数/15/非整数/溢出 | TestLoadMinOutputTokensRejectsExplicitInvalidValue与invalid environment子场景；加载时在弱解码前校验原值，避免浮点截断 |
| 三种转换小于/等于/大于下限 | TestConfiguredMinOutputTokens_AllConverters，3个floor×6个请求值×3种转换 |
| 缺省/负值/双字段优先级/旧接口 | AllConverters与TestConfiguredMinOutputTokens_FieldPrecedenceAndLegacyDefault |
| 多消费者配置/实际请求 | TestMinOutputTokens_ForwardAsAnthropicRawChatUsesInstanceConfig捕获真实出站max_completion_tokens且不改usage；其余现有转换调用已按所属s.cfg/settingService.cfg接线 |
| nil/未初始化配置/实例隔离 | TestConfiguredMinOutputTokens_DefaultAndInstanceIsolation、两实例真实转发并行测试，经race通过 |
| 原生HTTP/WS首帧后续/compact | TestMinOutputTokens_NativeResponsesAndWebSocketRemainNative；现有OAuth搜索历史与WS传输逻辑不引入legacy转换 |
| provider cap/Claude原生探针 | 原TestOllamaCloudMaxTokensClamp/Cap及原生Claude Code max_tokens=1探针回归由统一后端unit覆盖 |

两个count_tokens中的旧AnthropicToResponses调用最终只提取instructions/input/tools/tool_choice，没有输出上限字段；它们保持原接口，不将这项配置扩大到输入Token估算。缺失的legacy入口和native Anthropic平台链不引入。配置模板均保留默认128。
