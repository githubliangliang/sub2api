# 实施验收证据

状态：第二档6簇开发及验收完成。

## 逐场景证据

| 簇 | 失败证据 | 改后场景与正常反例 |
|---|---|---|
| S01 | [red-core.txt](evidence/red-core.txt)：Sonnet thinking模式/签名转换丢失 | `TestSonnet55RejectsUnsupportedParametersBeforeMimicry`、`TestSonnet55PreservesSignedHistoryAndAcceptsDefaultParameters`覆盖between_tools/采样/强制工具；`TestClaude55FilterThinkingBlocksRemovesSignedChainAfterInvalidBlock`、`TestClaude55ResponsesSignedThinkingBufferedAndStreamed`覆盖完整与损坏签名；`gateway_sonnet55_toolset_beta_test.go`覆盖现有API/OAuth/Bedrock/Vertex/countTokens；`TestSonnet55AndGPT61PricingSourcesAndOverrides`覆盖别名、5m/1h缓存价及显式零渠道价；目录与effort包回归通过 |
| S02 | [red-core.txt](evidence/red-core.txt)：旧thinking disabled被max覆盖；[red-plan-labels.txt](evidence/red-plan-labels.txt)：套餐标签错误；[red-gpt61-cache-price.txt](evidence/red-gpt61-cache-price.txt)：显式零缓存价在priority产生费用 | `TestGPT61SolRejectsDisabledReasoningBeforeForwarding`、`TestGPT61SolMappedReasoningModeAndSampling`及两桥测试覆盖映射后none/minimal/disabled拒绝、max保留、纯Chat工具拒绝；`TestGPT61SolOfflineCodexCatalog`、`TestGPT61SolAPIKeyCatalogUsesFullResponses`、`TestGPT61SolPublicCatalogPreservesOfficialExtensions`、`TestGPT61SolCatalogPreservesExplicitAccountMetadata`覆盖完整官方字段/nested消息/公开slug/账号覆写；价格sources/gate/priority及显式零测试通过；套餐组件15项、credentialsBuilder56项通过，保留同标签不同SKU和未知值 |
| S03 | [red-core.txt](evidence/red-core.txt)：未请求include_usage时缺usage、权威总量重复计缓存、缺失usage被合成 | `TestAnthropicChatStreamAuthoritativeUsage` 72组合覆盖三种include_usage、显式零/缺失/null、hit/miss、总量/缓存别名、后到缓存、stop/EOF和重复terminal；`TestHandleResponsesStreamingResponse_NormalizesTerminalUsage`验证终态与计费桶一致；既有断开drain测试和真实SQLite去重继续通过 |
| S04 | [red-core.txt](evidence/red-core.txt)：silent/signature-only上游首ping前超时 | `TestAntigravityCompatHandlerPreContentKeepalive`、`TestAntigravityCompatPreContentKeepalive`、`TestAntigravityCompatHandlerRepeatsPreContentKeepalive`验证15s ping与真实首token；`TestAntigravityCompatHandlerPreContentDeadlineWithCommentOnlyStream`验证硬截止；`TestAntigravityCompatHandlerErrorsAfterPreContentPing`、`TestAntigravityCompatEmptyAfterKeepaliveReportsStreamError`验证提交后SSE错误不failover；既有empty/cancel/drain测试与race通过 |
| S05 | [red-core.txt](evidence/red-core.txt)：advanced/legacy优先级、粘性、DB重查选中非owner | 四个`SelectAccountWithScheduler_*AccountModelRoute*`测试通过；`TestOpenAIResponsesWebSocket_CompositeAlias`、`_CompositeChannelBilling`验证OpenAI/Grok、responses/any、dedicated/passthrough、多轮省略/重复模型、公开响应/用量身份与route→channel→account映射；`_CompositeRouteRejections`、`_CompositeModelSwitchRequiresReconnect`、`_CompositeSessionModelSwitchRequiresReconnect`、`_CompositeDetectorFallback`验证拒绝、重连、session更新与已知模型回退；真实HTTP/WS和race通过 |
| S06 | [red-remote-catalog.txt](evidence/red-remote-catalog.txt)：无remote配置/超大目录提示 | `CodexRemoteCatalog.spec.ts`覆盖remote/file互斥、UTF-8超1MiB降级、恰等阈值仍remote、迟到请求隔离和Bearer认证；`api/codex.spec.ts`覆盖根/v1/子路径/尾斜杠、完整下载内容与UTF-8原始字节；`UseKeyModal.spec.ts`覆盖OpenAI HTTP/WS、Grok及Windows旧~/路径、CC Switch回归；252文件前端全量通过 |

定向改后日志：[模型/流初验](evidence/green-model-stream.txt)、[最终目录/价格契约](evidence/green-catalog.txt)、[价格与真实fallback](evidence/green-price-fallback.txt)、[WS](evidence/green-ws.txt)。早期green-remote-catalog/frontend-targeted日志记录过渡状态，不作为最终通过依据；最终以前端全量及收尾组件日志为准。

S02适配中显式零缓存写价在priority的回退由专属GPT-6.1规则修正，不引入未采用的全套Fast价格引擎。S01/S02价格及原有账号长上下文gate均由实际BillingService消费者验证。

## 执行环境与门禁

2026-09-30，在 `sync/upstream-0.2.11` 分支完成。实施起点 `f13dee245790eb2be058935d0b1aa6265da99474`；开始时只有前一轮评估文档待提交，产品文件干净。开发验收时为未提交工作树，验收产品及测试文件的 SHA-256 清单见 [implementation-files.tsv](../port-upstream-0.2.11-tier2/evidence/implementation-files.tsv)。开发验收时未发布或部署；随后按用户要求完成本地Docker及远端部署，见下方部署记录。

以下命令均使用 `rtk proxy` 执行；Go 命令工作目录为 backend，pnpm 命令工作目录为 frontend。OpenSpec 命令在仓库根目录运行。

| 门禁 | 命令 | 结果 / 证据 |
|---|---|---|
| 后端全量 | `go test -tags=unit ./... -count=1` | 通过，[backend-unit.txt](../port-upstream-0.2.11-tier2/evidence/backend-unit.txt)；service 169.360s，handler 37.181s |
| 后端构建 / DI | `go build ./...`；`go generate ./cmd/server` | 通过，[backend-build.txt](../port-upstream-0.2.11-tier2/evidence/backend-build.txt)；Wire仅增加resolver注入 |
| 并发 / HTTP / WS | `go test -race -tags=unit ./internal/service ./internal/handler -run 'TestAntigravityCompat\|TestAnthropicChatStreamAuthoritativeUsage\|TestHandleResponses.*Usage\|TestOpenAIResponsesWebSocket_Composite\|TestGatewayOpenAICompatibleHandlersClaudeCodeOnlyFallback\|AccountModelRoute' -count=1` | 通过，[backend-race.txt](../port-upstream-0.2.11-tier2/evidence/backend-race.txt) |
| SQLite / 去重 | `go test -tags=unit ./internal/repository -run 'TestProductionSQLUsesSQLiteDialect\|TestUsageBillingRepositoryApplySQLiteAppliesAllEffects' -v -count=1` | 通过，[sqlite.txt](../port-upstream-0.2.11-tier2/evidence/sqlite.txt)；真SQLite重复Apply不再扣费 |
| 前端全量 | `pnpm exec vitest run` | 252文件，1836通过 / 2既有跳过，[frontend-unit.txt](../port-upstream-0.2.11-tier2/evidence/frontend-unit.txt) |
| 前端类型 / lint / build | `pnpm run typecheck`；`pnpm run lint:check`；`pnpm run build` | 通过，见第二档evidence中同名日志；仅既有pnpm/browserslist构建提示 |
| OpenSpec | `pnpm dlx @fission-ai/openspec validate port-upstream-0.2.11-tier1 --strict`；同命令tier2 | 通过，见两档evidence/openspec.txt |
| 范围 / 格式 | `git diff --check`；文件清单与冻结证据哈希比对 | 通过，[scope-check.txt](../port-upstream-0.2.11-tier2/evidence/scope-check.txt) |

全量Go通过后，仅增加GPT目录契约测试并调整两条既有测试的Gin全局模式初始化顺序；对应定向与race再次通过。前端全量通过后仅清理测试的未用import和修正目录提示，最终类型/lint/build及相关组件回归112项通过（[frontend-final.txt](../port-upstream-0.2.11-tier2/evidence/frontend-final.txt)）。未将postgres-only集成测试或“不匹配任何测试”的输出算作行为验收。

## 边界与测试适配

保留SQLite-only、最大迁移226、可关闭Redis/miniredis、simple mode、VERSION 1.1.14；没有schema、依赖、支付、缺失原生平台、Astra Ultrafast或其它第三/四档功能变动。两份source-baseline及20份评估证据的原始哈希保持不变。

旧v0.2.9 F04/#7538只作为本次S02前置完成两桥thinking禁用优先级；旧第一档其它簇仍待处理。旧695ebede7仅移植14行通用usage字段。

新增上游测试中对缺失native Anthropic路径、未引入的allowlist、全套Fast策略的依赖已按本地实际消费者适配；这些编译差异不记作产品缺陷。race发现两条既有Responses测试在`t.Parallel`后写Gin全局模式，已将初始化移至并行边界之前；[原报告](../port-upstream-0.2.11-tier2/evidence/red-race-harness.txt)与最终通过日志均保留。

## 本地Docker部署（2026-09-30）

用户追加要求后，以当前工作树重建`sub2api-new:local`，通过既有SQLite Compose更新`sub2api-new-sqlite`。容器healthy，零重启；健康、首页、公开设置和最新前端资源均HTTP 200。沿用现有数据挂载及环境（此容器原为standard模式），数据库切换前快照与运行库均通过应用SQLite引擎校验。配置备份及旧镜像回退已保留。详见[部署证据](../port-upstream-0.2.11-tier2/evidence/docker-deployment.txt)。部署时尚未创建git提交、tag或Release。

## 远端部署（2026-09-30）

按用户要求执行`deploy/deploy-remote.sh --yes`，本地重新构建后上传至`liang@192.236.223.107:/home/liang/sub2api`，通过配置的自定义命令重启用户级systemd服务。脚本退出0；服务active/running、零重启，本地二进制、远端文件及运行进程的SHA-256完全相同。配置校验值未变，旧二进制保留；健康、首页、公开设置及最新前端资源均HTTP 200，外网健康检查通过。版本仍1.1.14，包含部署时的工作树代码，当时尚未创建git提交或GitHub Release。见[远端部署证据](../port-upstream-0.2.11-tier2/evidence/remote-deployment.txt)。
