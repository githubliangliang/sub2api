# 逐场景验收核对

对应 [gateway-stability/spec.md](./specs/gateway-stability/spec.md)。以下用例均为实际执行的行为测试；最终全量结果与命令见 [verification.md](./verification.md)。源码定位用于补充调用链和不变量检查，不用文件存在性代替运行结果。

| 场景 | 行为证据 |
|---|---|
| S01 新模型识别 | `TestTier2ModelBillingAliases`、`TestTier2Grok47BridgeAndMapping`、`TestGPT6MappedCompatibilityBridgesKeepReasoningAndTools`；前端模型白名单与 UseKeyModal 测试。目录价卡只新增四个目标模型。 |
| S01 跨代及别名 | `TestTier2ReasoningGenerations`、`TestTier2DottedOpusEffort`、`TestTier2OpusAdaptiveAndSignedHistory`；旧模型与非数字 GPT 系列保留采样规则。 |
| S02 root union | `TestResponsesToAnthropic_ObjectUnionRootMergesBranches`、`TestResponsesToAnthropic_AllOfRootKeepsUnionRequired`、`TestNormalizeAnthropicInputSchema_AllOfKeepsPropertyIntersection`、nested/root properties 反例；降格不宣称完整 oneOf 等价。 |
| S02 null 字段 | `TestSanitizeOpenAIResponsesToolParameterTypes_*`、`TestTier2ToolSchemaBeforeDispatch`、`TestTier2ToolSchemaNativeAnthropicDispatch`；Responses/Chat/Messages 到现有 OpenAI 与原生 Anthropic 入口均清洗，实例 default/const/enum 数据保留。 |
| S02 tuple 与 const | `TestCleanJSONSchema_ArrayPrefixItems`、`TestCleanJSONSchema_ArrayExistingItemsPreserved`、`TestBuildToolsPreservesStringConst`；包括冲突 enum 的空交集。 |
| S03 工具完成参数 | `TestAnthropicEventToResponses_ToolCallArgumentsDoneMatchesDeltas`，完成参数与 delta 逐字一致。 |
| S03 inline 参数 | `TestAnthropicEventToResponses_ToolInputOnContentBlockStart`、`TestAnthropicEventToResponses_ToolInputSeedNotDuplicatedByDeltas`、client tool restore 链测试；有 delta 时丢弃 seed。 |
| S03 内嵌 PDF | `TestTier2PDFResponsesAnthropicGeminiChain` 覆盖 document→inlineData、空数据和 file_id-only，不进行下载。 |
| S04 terminal 不等 EOF | `TestOpenAIResponseFlush_TerminalEventEndsStreamWithoutEOF`；`TestTier2BareErrorWaitsForAuthoritativeTerminal` 覆盖原生 SSE、透传 SSE、WS→HTTP 的 bare error 后成功或 EOF。 |
| S04 取消后关体 | `TestHTTPUpstreamConcurrentEarlyCloseDoesNotPoisonNextResponse`、`TestHTTPUpstreamCompressedEarlyCloseStopsReader`、`TestHTTPUpstreamHTTP2CloseDoesNotCancelOtherStream`、keep-alive/取消/父上下文隔离测试，包含 race。 |
| S04 心跳与失败 | `TestOpenAIStreamKeepaliveOnlyFailureDoesNotRecordFirstToken`、`TestProxyOpenAIWSHTTPBridgeTurnStagesMetadataAndRelaysKeepaliveBeforeCapacityFailover`、写失败/断开后不重放测试；Anthropic ping 在 CC 转换前过滤。 |
| S04 单个协议错误 | `TestOpenAIStreamingReadErrorAfterOutputUsesResponsesErrorSchema`、Timeout/TooLong、`TestOpenAIStreamSemanticStatusHonorsErrorStatusAlias`；bare error 无后续终态时只生成一次 failed，`TestTier2BareErrorExplicitPassthroughRuleWins` 保留显式规则优先级。 |
| S04 Gemini 传输故障 | `TestGeminiForwardNative_TransientTransportErrorFailsOverWithoutRetry`、PersistentTransportError、CountTokensTransportErrorFallsBackToEstimate、Claude/CC 兼容入口测试。 |
| S04 客户端 499 | `TestTransportErrorHandlers_ClientDisconnectRecordsNoOpsEvent`、`TestGeminiV1BetaModels_ClientCancelBeforeUpstreamResponseMarks499`、`TestOpsErrorLoggerMiddleware_*ClientClosed*`、已开始流取消不追加错误测试。已有上游失败仍归因上游。 |
| S05 裸名映射 | `TestResolveGeminiThinkingVariant`、`TestGeminiThinkingLevel_NativeAndCompatAgree`、`TestGetMappedModelResolvesBareGeminiModelForAllEntrypoints`；显式映射、预算、变体回退和已有默认目录分别验证。 |
| S05 SDK 心跳 | `TestGeminiClientRejectsSSEComments`、`TestDownstreamRejectsSSECommentsReadsBothHeaders`、真实流的 OrdinaryClients/GoGenai 对照。 |
| S05 空流重试 | `TestTier2AntigravityMalformedStreamOnlyRetriesWithoutContent`、`TestTier2AntigravityControlEventsAreNotContent`；signature/stop 不算正文，文本/工具有内容时不重复请求。 |
| S06 previous_response 路由 | `TestLegacySchedulerDecision_PreviousResponseRouting` 和 StickySessionLayer；归属、排除、渠道限制、代理隔离、不可用传输与槽位释放。 |
| S06 热路径轻量读 | `gateway_multiplatform_test.go` 的 GetByIDLite mock 与调度测试；`gateway_scheduling.go`、`openai_account_scheduler.go` 使用 Lite 分组接口，归属分支同样检查 privacy。 |
| S06 模型 401 与认证 401 | `TestRateLimitService_HandleUpstreamError_APIKeyModel401UsesModelRateLimit` 与 NonOAuthModel401StillDisablesCredentials；只冷却未知模型，认证错误仍禁用。 |
| S06 暂停 OAuth 刷新 | `TestTier2PausedOAuthRefreshCandidatesSQLite`，真实迁移库验证健康、暂停、永久 error、刷新重试冷却；查询补充 NULL 安全条件。 |
| S07 账号 gate 优先 | `TestTier2AccountStatsLongContextGate` 经真实 RecordUsage，账号成本 0.78/1.545 与分组售价 1.545 独立；既有 custom rule / ApplyPricingToAccountStats / service tier 优先级测试保持通过。 |
| S07 图片未填和零 | `TestTier2ChannelImagePriceInheritsCatalogUnlessExplicit`、隐式图片价 fallback、显式零价与区间不污染共享价卡测试；`intervalToModelPricing` 继承基础图价后再应用显式配置。 |
| S08 CC Switch 配置保真 | `ccswitchImport.spec.ts` 覆盖 root、/v1、尾斜杠、/x/、/x/v1/；Antigravity 两类客户端另测平台路径连接。 |
| S08 原生与 usage 路径 | UseKeyModal 原生 Codex/Claude URL 对照矩阵；CC Switch usage script 实际求值。`evidence/regressions.txt` 记录冻结基线脚本产生重复 /v1、当前脚本修复的结果。 |
| S08 Windows 模型清单 | UseKeyModal 的 OpenAI Codex、Codex WS、Grok Windows 测试，配置使用 `~/.codex/codex-models.json`。 |
| S09 恢复代理失效探针 | `TestTier2RevertProxyInvalidatesOnlyChangedIdentitySQLite`；真实库验证仅网络身份变化删除 probe，其它 extra 保留。 |
| S09 扫描竞争 | `TestTier2ProxySweepRejectsStaleSnapshotSQLite`：读取快照后续期、停用、修改 fallback/backup 均不覆盖；未变化快照正常执行。 |
| S09 不可用目标 | `TestResolveFallbackSkipsInactiveBackup`；SQLite 增加快照读取后目标被停用的交错测试，写入时再次确认目标可用。 |
| S10 模型不可用不写能力 | `TestResponsesModelFailurePreservesEveryCapabilityState` 对 400/404 × 未知/true/false 验证不触发 UpdateExtra。 |
| S10 真正端点不支持 | `TestResponsesProbeModelUnavailableIsInconclusive` 的普通 404/405 对照、`TestSelectResponsesProbeModelPrefersGeneralTextModel` 与普通 400 反例。 |
| S11 别名回显 | `TestMappedResponseModelPreservesOtherData`、`TestMappedResponseModelForwarding`、raw Chat 与 replaceModel 系列；涵盖流/非流、透传、上游另一个别名、工具内 model 保留和真实上游模型记录。 |
| S12 reminder 关键词 | `TestContentModerationCheck_ReminderKeywordModes`、ReminderPolicyPreserved、LongReminder、ResponsesInputForms；实际 Check 与本地 HTTP 审核桩验证关键词命中、语义审核输入不扩大。 |
| S12 尾随 system | `TestExtractContentModerationInput_AnthropicTrailingSystemExtractsLatestUser`、keyword boundaries；assistant/tool 结束不重审历史。 |
| S13 限定前导身份 | `attribution_test.go` 的完整转换、身份中和、CR/LF/正文提及对照；原生 Anthropic passthrough 与 OAuth mimic body 测试仍保留 attribution，清洗函数只在 Antigravity 包内调用。 |
| S14 嵌套滚动锁 | BaseDialog scrollLock 四项与实际 Chromium 组件挂载/点击；最后一个关闭才解除 body 锁，原 overflow 恢复。 |
| S14 标签 IME | ModelTagInput keyboard 九项及浏览器组合 Enter/普通 Enter 对照；空 Tab 导航、Delete/Backspace 语义保留。 |
| S15 切换上下文窗口 | `TestTier2CodexWebSocketBoundaries` 的 new window/failed window；真实客户端 WS 多轮验证旧 ID 删除，失败终态不更新 last window；归一化 helper 与 `expectedPrev` 清空分支同时核对。 |
| S15 同窗口或缺标记 | 同一真实 WS 测试的 same window/missing window；嵌入 turn metadata 回退由 `TestNormalizeOpenAIWSContextWindowBoundary` 验证。 |
| S16 Lite 映射 | `TestMappedGPT55LiteCompatibility`、BuildersPreserveIngressForFailover、HTTPBridgePreservesRequest、真实 WS mapped Lite、`TestTier2MappedLiteCreatePayloadDoesNotMutateInput`；出站修改不污染原请求，API Key/其它模型不改。 |
| S16 非 ASCII metadata | `TestCodexTurnMetadataRewritePreservesHeaderSafeJSON` 覆盖 header/map/raw 三个本地 fingerprint 消费者，中文、emoji、控制字符、字面转义、ASCII 与解码还原。未引入缺失的 account_identity 子系统。 |
| Preserve personal SQLite deployment | 无 VERSION、迁移、依赖清单、菜单修改；真实 SQLite 与方言审计；HTTP 断开后计费用量、既有真实库 usage billing 去重测试；simple mode 分支保持。 |

## 交付与边界

- 起点 `50c34ad498fa333e20b4ee46dfe58c4512d88a89`，开始工作树干净；冻结上游 `4c00df2e0183e2c70b7fa8ba45914205e36aad0c` 未改写。
- 实施验收阶段未提交、推送或部署；2026-09-29 后续按用户要求完成本地 Docker 部署并整理提交。产品和测试文件的准确版本由 [product-files.sha256](./evidence/product-files.sha256) 标识。
- Go 命令在 `backend/` 执行，自动工具链为 Go 1.27.0；前端命令在 `frontend/` 执行。
- 前端两项既有跳过测试位于 SettingsView，原因是已移除的支付 provider API；支付明确不在本次范围。目标功能没有跳过验收。
