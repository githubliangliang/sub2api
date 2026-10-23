# 实施验收

状态：**F01–F06已实现并完成验收**（2026-10-23）。评估证据和source-baseline保持冻结，本文件只记录实施证据。

| 实施信息 | 值 |
|---|---|
| 实施起点 | `4984327590ed29eb4d842bfae8aa3406ef497008`，工作树干净；相比冻结产品基线`873962ce7`仅增加评估文档，见[start.json](../../../docs/upstream-sync/implementation-0.2.18-tier1/start.json) |
| 实现提交 | 当前工作树，未提交；[代码/测试SHA256清单](../../../docs/upstream-sync/implementation-0.2.18-tier1/code-sha256.json)固定本次实现内容 |
| VERSION / 工具链 / 迁移 | VERSION 1.1.18；backend目录实际使用Go1.27.0；最大SQLite迁移226；VERSION、迁移、schema和依赖均未修改 |
| 范围 | 仅第一档6簇；第二档、第三/四档保持原状，无支付、新平台、模型发现能力、发布或部署 |

## 行为与反例

每簇先增加失败回归，再修复。“修复前”是本次实际失败输出，不是评估期断言旧行为的probe。最终定向通过见[backend-targeted.txt](../../../docs/upstream-sync/implementation-0.2.18-tier1/backend-targeted.txt)，F01与F03另有独立最终输出。

| 簇 | 修复前证据 | 最终行为与反例 |
|---|---|---|
| F01 Codex工作区导入 | [失败](../../../docs/upstream-sync/implementation-0.2.18-tier1/f01-before.txt)：同用户两workspace仅创建1个，第二条误去重 | [通过](../../../docs/upstream-sync/implementation-0.2.18-tier1/f01-after.txt)：`TestImportCodexSessionsSameUserDifferentWorkspacesCreatesSeparateAccounts`创建2个并逐workspace重导入，分别更新原ID；同workspace不同成员、accessToken-only指纹/升级/refresh保留、缺user的account兜底、缺account的user键、长度编码分隔符和索引更新/共享候选均覆盖 |
| F02 Codex前导元数据 | [失败](../../../docs/upstream-sync/implementation-0.2.18-tier1/f02-before.txt)、[超时失败](../../../docs/upstream-sync/implementation-0.2.18-tier1/f02-timeout-before.txt)：queued当输出，后续overload失去failover | `TestOpenAIStreamMetadataPreambleAndMessageOnlyOverloadFailOver`覆盖native HTTP SSE与透传；两个`TestProxyOpenAIWSHTTPBridgeTurnStagesMetadata*`覆盖WS-HTTP桥及keepalive后换号；queued/metadata保持私有，真实文本、工具参数和未知事件仍提交输出；已有输出后不重试；`TestOpenAIFirstOutputTimeoutAfterQueuedAndMetadata`及原有首输出/注释写出安全换号测试通过 |
| F03 Grok Fast价卡 | [失败](../../../docs/upstream-sync/implementation-0.2.18-tier1/f03-before.txt)、[来源格式失败](../../../docs/upstream-sync/implementation-0.2.18-tier1/f03-source-fields-before.txt)：Fast落普通价卡、来源长上下文字段未读取 | [最终通过](../../../docs/upstream-sync/implementation-0.2.18-tier1/f03-final.txt)：`TestGrok47FastPricingResourceAndFallback`验证Fast/build-fast、大小写及xai/x-ai/grok前缀；离线与资源在199999使用4/1/12、200000及全缓存输入使用6/1.5/18（每百万input/cache/output）；`TestGrok47FastPricingSourceLongContextFields`直读来源格式；显式渠道价格含0优先且不修改默认卡，普通4.7与其它Grok回归保持 |
| F04 Cloudflare 1010 | [失败](../../../docs/upstream-sync/implementation-0.2.18-tier1/f04-before.txt)：跨平台HTML及结构化1010触发处罚 | `TestHandleUpstreamError_Cloudflare1010DoesNotPenalizeAccount`覆盖OpenAI/Anthropic/Grok/Gemini/Antigravity，重复HTML、文本、JSON、嵌套数组/对象、大小写键与browser_signature_banned均无SetError、临时禁用、计数或调度通知；`TestHandleUpstreamError_NonCloudflare403StillPenalizes`覆盖真实权限403、无Cloudflare标记的1010、非同节点标记、false标记、业务文本、10101与畸形JSON；既有OpenAI HTML例外保留 |
| F05 Gemini非负用量 | [失败](../../../docs/upstream-sync/implementation-0.2.18-tier1/f05-before.txt)：input为-8/-5，其它桶保持正数 | `TestExtractGeminiUsageNonNegativeInputPreservesBuckets`及`TestGeminiUsageMapping_NonNegativeInputPreservesBuckets`覆盖prompt缺失/0/小于/等于/大于cached；服务提取、非流式JSON、流式message_start、后续更新与最终message_delta/返回usage均非负；cached=8、output含thoughts=5、image=2、cache creation=0逐桶断言 |
| F06 Gemini工具Schema | [失败](../../../docs/upstream-sync/implementation-0.2.18-tier1/f06-before.txt)：const/examples仍出站，default数据被改写 | `TestCleanToolSchema_*`覆盖字符串const单值enum与交集（冲突为空、数字与字符串不混同）、非字符串const删除且不推导type、嵌套items/anyOf/oneOf/dependencies等Schema节点；`TestCleanToolSchema_PreservesPropertyNamesAndDataWithoutMutation`保留名为const/examples/$schema的属性、default与自定义数据，检查原Schema及返回副本独立；上一轮integer exclusiveMinimum与enum规范化回归通过 |

## 检查与命令

Go命令在backend执行，所有命令通过rtk代理。完整输出位于[实施证据目录](../../../docs/upstream-sync/implementation-0.2.18-tier1/)。

| 门禁 | 命令/证据 | 结果 |
|---|---|---|
| backend unit | `go test -tags=unit ./... -count=1`；[最终输出](../../../docs/upstream-sync/implementation-0.2.18-tier1/backend-unit.txt) | 全包通过；初轮通过后补齐来源价格格式适配，最终代码再次全量通过 |
| backend build | `go build -o /tmp/sub2api-0218-tier1/server ./cmd/server`；[输出](../../../docs/upstream-sync/implementation-0.2.18-tier1/backend-build.txt) | 通过 |
| race | 下方命令；[输出](../../../docs/upstream-sync/implementation-0.2.18-tier1/backend-race.txt) | 4包通过，包含全部6簇及SQLite/固定窗口回归 |
| SQLite结算/去重/固定quota窗口 | repository/service，`-tags=unit -run 'Test(UsageBillingDeletedKeySQLite\|UsageBillingRepositoryApplySQLite\|UpdateFixedQuotaResetAt\|ProductionSQLUsesSQLiteDialect\|SplitSQLStatements\|NormalizeFixedQuotaWindows)' -count=1 -v`；[输出](../../../docs/upstream-sync/implementation-0.2.18-tier1/backend-sqlite.txt) | 通过：真实SQLite删除Key、去重、所有写入效果、数据库错误回滚与固定窗口推进；未执行PG专用integration |
| 前端/依赖 | 未修改frontend、go.mod/go.sum或锁文件 | 不适用；无pnpm门禁或重新audit要求，依赖维护仍在第二档S10 |
| OpenSpec | `pnpm dlx @fission-ai/openspec validate port-upstream-0.2.18-tier1 --strict`；[输出](../../../docs/upstream-sync/implementation-0.2.18-tier1/openspec-strict.txt) | 通过 |
| 范围/冻结基线/格式 | 代码清单、受保护路径、source-baseline哈希、文档链接及`git diff --check` | 通过；见[一致性检查](../../../docs/upstream-sync/implementation-0.2.18-tier1/consistency.txt) |

```bash
go test -race -tags=unit ./internal/handler/admin ./internal/service ./internal/pkg/antigravity ./internal/repository -run 'Test(Codex|ImportCodex|NormalizeCodex|Grok47Fast|GetModelPricing_.*Grok|CalculateCost_.*Grok|GetModelPricingWithChannel|OpenAIStream|OpenAIFirstOutput|OpenAINativeFirstOutput|ProxyOpenAIWSHTTPBridgeTurn|OpenAIStreamingPassthrough|OpenAIResponseFlush|HandleUpstreamError_.*403|HandleUpstreamError_Cloudflare|HandleUpstreamError_NonCloudflare|ExtractGeminiUsage|CleanToolSchema|ConvertClaudeToolsToGeminiTools|GeminiUsageMapping|UsageBillingDeletedKeySQLite|UsageBillingRepositoryApplySQLite|UpdateFixedQuotaResetAt|ProductionSQLUsesSQLiteDialect|SplitSQLStatements|NormalizeFixedQuotaWindows)' -count=1
```

## 本地适配与中间验证

- F03来源采用above_200k绝对价格，本地原先只读显式threshold/multiplier。新增Fast专属换算和别名查找，渠道显式价格继续优先，不改普通Grok策略。只取来源Fast条目，未带入Argon或模型发现目录。新增patch sites记在PORTING实施补记。
- F04的HTML/文本同时要求Cloudflare标记与完整1010代码，避免业务1010或10101被豁免；JSON要求明确标记或browser_signature_banned，不扫描message文字冒充结构化信号。
- F02超时验收针对本地原有native SSE首输出守卫；透传及WS-HTTP桥此前没有该计时守卫，本次验证其元数据/keepalive/重试，不新增计时功能。[中间输出](../../../docs/upstream-sync/implementation-0.2.18-tier1/f02-intermediate.txt)保留了最初误向透传断言该守卫的失败；修正适用路径后，native超时与三条重试路径均通过。
- F03的[中间输出](../../../docs/upstream-sync/implementation-0.2.18-tier1/f02-f03-intermediate.txt)发现离线别名未被HasIdentifiedTokenPricing认出，已补齐确定性Fast识别，避免实际消费者丢失价卡。后续来源格式失败另行修复，未用本地扩展字段掩盖问题。
- F01最初测试脚手架误将CreateAccountInput当成带ID的账号，产生[编译错误](../../../docs/upstream-sync/implementation-0.2.18-tier1/f01-test-setup.txt)；改为读取内存账号ID后再取得真实缺陷失败证据。编译错误不计作缺陷复现。

## 发布/部署

未提交、推送、打tag、发布或部署，未连接真实供应商或修改业务数据库/凭据。以上为本地协议、价格与SQLite自动化验收；冻结来源和评估期证据未改。

- [x] 每簇证据完整并满足spec。
- [x] 适用检查完成，未执行项说明原因。
