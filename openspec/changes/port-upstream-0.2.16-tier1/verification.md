# 第一档实施验收

状态：F01–F14已实现。2026-10-14完成本轮验收；来源基线与评估证据保持冻结。本文件记录实际实现验证，不引用评估probe冒充验收。

| 实施信息 | 值 |
|---|---|
| 实施起点SHA | `0fbcab935`（包含冻结评估；产品基线为 `128c5f7bf`） |
| 实现SHA | 尚未提交；[代码/测试SHA256清单](../../../docs/upstream-sync/implementation-0.2.16-tier1/code-sha256.json)固定本次工作区内容 |
| 实施日期 | 2026-10-13至2026-10-14；承接历史会话未完成改动 |
| 分支 | `main` 工作区 |
| 版本/迁移/依赖 | VERSION 1.1.16、最大迁移226；go.mod/go.sum、package.json、pnpm-lock、pnpm-workspace未改 |

## 行为与反例

以下测试均参与后端全量unit或前端全量Vitest；新增的流式签名、图片冷却写入及长凭据脱敏用例另由最终race覆盖。

| 簇 | 目标行为证据 | 反例/范围证据 | 结果 |
|---|---|---|---|
| F01 删除Key后完成请求仍结算 | `TestUsageBillingDeletedKeySQLite` 使用真实SQLite，软删除/物理缺失Key后余额、账号配额与dedup正常提交 | 重放不重复结算；Key自身计数不变；quota/rate触发器数据库错误整体回滚；零BalanceCost不引入simple模式扣款；既有固定窗口/去重回归 | 通过 |
| F02 Chat角色与工具选择 | `TestChatCompletionsToResponses_DeveloperRolePreserved`、具名tool_choice回归 | developer文字/多段内容保持；沿用本地省略type的消息形式；auto/none/required、已是Responses格式、空name及system/user保持 | 通过 |
| F03 缓冲索引与thinking | `TestAnthropicBufferedAggregators_IgnoreNegativeBlockIndex`覆盖两条缓冲桥；`TestStreamingThinkingSignatureWire`及两种内容类型JSON序列化用例 | 负索引丢弃但正常文本保留；空thinking/空signature存在，后续signature_delta保真；普通文本不带thinking字段 | 通过 |
| F04 缓存TTL与Sonnet采样 | `gateway_cache_ttl_order_test.go`覆盖tools→system→messages、顶层断点、拟态和count_tokens；`TestClaudeOAuthSamplingDefaults` | 合法顺序/无断点不变；四块限制保留；Sonnet top_p-only不注入temperature，显式temperature与其它模型默认规则保留；未引入Haiku 5.5 | 通过 |
| F05 Grok空完成与SSO | `TestOpenAIResponsesEmptyCompletedFailsOverGrok`与非空反例；`TestConvertSSOToBuildCompletesDeviceFlow`及`TestSSOConsentToken` | consent按HTML解析并解码实体，非hidden/无字段兼容；origin/referer只在批准首跳，重定向检查无泄漏，consent不入cookie；有输出成功 | 通过 |
| F06 Codex目录 | `TestGPT61SolAPIKeyCatalogContextWindow`与既有catalog测试；`CodexRemoteCatalog.spec.ts` | 缺省/null为500000，显式上游值与用户目录优先；OpenAI HTTP/WS、Grok、Gemini路由配置在macOS/Linux、Windows均仅remote加发现开关；file不加 | 通过 |
| F07 工具图片与冷却 | Gemini工具结果转换测试保持functionResponse→inlineData顺序；`TestSetAntigravityImageRateLimitDoesNotBlockTextFamily`覆盖实际写入及读取 | 远程图片不下载、文本结果不变；图片仅自己的键，文本家族原规则保留；已有映射与同族图片别名仍生效 | 通过 |
| F08 价格/倍率/吞吐 | `pricing.formatScaled`、`formatBytes`、`PricingEntryCard.modelDefaultPrice`与`UsageTable`组件用例 | 科学计数与亚字节正确；0倍率保留，null/undefined回退1；仅空条目首次单模型自动填价，已有/批量/手工价不覆盖；长上下文中英文徽标不再声称统一x2 | 通过 |
| F09 弹窗与选择器 | `BaseDialog.escape`、`disabledSelectors`、`Select.ime`；[浏览器验收](../../../docs/upstream-sync/implementation-0.2.16-tier1/browser.md) | 仅关闭栈顶；disabled关闭并清搜索；IME不拦截按键，普通导航保留；焦点归还及body锁在两层弹窗正确 | 通过 |
| F10 错误与开关 | BulkEditUserModal、useGeminiOAuth.errors、ModelWhitelistSelector、ErrorPassthroughRulesModal.toggle；`apiError.message`8例 | 后端原因优先，非字符串/空文本/仅HTTP状态通用错误回退；规范化错误/标准Error/i18n映射保留；开关按服务端enabled，失败不伪造状态 | 通过 |
| F11 计时与迟到请求 | PendingOAuthCreateAccountForm、ProfileBalanceNotifyCard、UsageProgressBar、BackupView、useBatchImageAccess.retry | 卸载/移除后回包不新建timer；切回可见立即更新时间；权限查询失败可重试、成功缓存；未启用通知/批量生图/后台服务 | 通过 |
| F12 监控与设置边界 | `TestChannelMonitorV2HealthRequiresRequestSamples/MinimumSampleBoundaries/SufficientHighErrorIsCritical`；OpsAlertRulesCard.duration、OpsDashboardHeader.timezone、OpsSettingsDialog.loading、ChannelStatusV1View.details | 样本阈值下所有评分未知，阈值处正常；整数分钟、本地时间；自动/手动同流程刷新当前筛选detail；加载中/失败禁存，成功重试后可存 | 通过 |
| F13 账号测试诊断与分组 | available-models handler测试验证显式请求侧映射名排序；`TestAccountTestFailureLogAttribution`、`TestAccountTestErrorLogRedaction`；分组[截图](../../../docs/upstream-sync/implementation-0.2.16-tier1/groups.png) | 无映射保留默认；HTTP/后台/缺失账号关联account/test/request，先脱敏再截断（长凭据覆盖）；长名无裁切，普通徽标默认省略 | 通过 |
| F14 嵌入前端API别名 | embed标签整包测试；`TestEmbeddedFrontendBypassesBareAPIAliases`及两种中间件实测请求透传 | 仅本地已注册别名；静态SVG fixture与真实dist一致，原两项失败消除；SPA深链保留；未加入不存在的models子路径及contents任务API | 通过 |

## 检查命令与结果

完整结果位于[实施证据目录](../../../docs/upstream-sync/implementation-0.2.16-tier1/)。Go命令在backend执行，pnpm命令在frontend执行，均通过rtk代理。

| 门禁 | 命令/证据 | 结果 |
|---|---|---|
| backend unit | `go test -tags=unit ./... -count=1`；[输出](../../../docs/upstream-sync/implementation-0.2.16-tier1/backend-unit.txt) | 全包通过 |
| backend build | `go build ./cmd/server`；[输出](../../../docs/upstream-sync/implementation-0.2.16-tier1/backend-build.txt) | 通过 |
| 相关race | 下方相同测试集复测；初始化修复前[真实失败记录](../../../docs/upstream-sync/implementation-0.2.16-tier1/backend-race-initial.txt) | 通过；[最终输出](../../../docs/upstream-sync/implementation-0.2.16-tier1/backend-race.txt) |
| SQLite方言/用量去重/固定quota窗口 | 全量unit及race含真实SQLite结算、删除Key、所有计数效果、固定窗口reset回归 | 通过 |
| embed静态资源与API别名 | `go test -tags=embed,unit ./internal/web -count=1`；[输出](../../../docs/upstream-sync/implementation-0.2.16-tier1/backend-embed.txt) | 通过 |
| embed单文件构建 | `go build -tags=embed -o /tmp/sub2api-tier1/sub2api-embed ./cmd/server`；[输出](../../../docs/upstream-sync/implementation-0.2.16-tier1/backend-embed-build.txt) | 通过 |
| frontend typecheck | `pnpm run typecheck`；[输出](../../../docs/upstream-sync/implementation-0.2.16-tier1/frontend-typecheck.txt)；最终build再次运行vue-tsc -b | 通过 |
| frontend lint/build | `pnpm run lint:check` / `pnpm run build`；[lint](../../../docs/upstream-sync/implementation-0.2.16-tier1/frontend-lint-final.txt)、[build](../../../docs/upstream-sync/implementation-0.2.16-tier1/frontend-build.txt) | 通过 |
| frontend行为/浏览器 | `pnpm run test:run`；[输出摘要](../../../docs/upstream-sync/implementation-0.2.16-tier1/frontend-tests.txt)、[浏览器](../../../docs/upstream-sync/implementation-0.2.16-tier1/browser.md) | 267文件，1934通过/2既有跳过；浏览器通过 |
| dependency audit与剩余项 | `pnpm audit --json`；[原始JSON](../../../docs/upstream-sync/implementation-0.2.16-tier1/dependency-audit.json) | 非零；103条（critical3/high42/moderate50/low8），与评估相同 |
| OpenSpec与文档一致性 | 本地结构/链接/14需求28场景/48任务/冻结SHA与代码清单校验；git diff --check | 通过；[校验记录](../../../docs/upstream-sync/implementation-0.2.16-tier1/consistency.txt) |

race命令（不通过降低并发或跳过失败用例规避）：

```bash
go test -race -tags=unit ./internal/pkg/apicompat ./internal/pkg/antigravity ./internal/pkg/xai ./internal/repository ./internal/service ./internal/handler/admin -run 'Test.*(UsageBillingDeletedKey|ProductionSQLUsesSQLiteDialect|UsageBillingRepositoryApplySQLite|UpdateFixedQuotaResetAt|ChatCompletionsToResponses|AnthropicBufferedAggregators|StreamingThinkingSignatureWire|ThinkingAlwaysHasSignature|CacheControl|CacheTTL|ClaudeOAuthSamplingDefaults|SSO|GPT61Sol|EmptyCompleted|AvailableModels|AccountTestFailureLogAttribution|AccountTestErrorLogRedaction|ChannelMonitorV2Health|ModelRateLimit|AntigravityImageRateLimit|ToolResultImage)' -count=1
```

## 实施偏差与问题处理

- 手工适配developer消息以保留本地省略type的合法格式；保留SQLite语法与128c5f7bf固定窗口逻辑，没有照搬PG integration。
- F10既有helper经失败用例证明不满足规格，需要补充优先级和字符串校验。只调整通用错误提取，支付专用解析未变；测量与patch sites记在PORTING F10。
- F14按本地路由裁剪上游补丁；未顺带增加模型、视频/语音/Realtime能力。现有已注册的相关别名仅修复SPA拦截。
- 广泛race测试暴露两个原有并行用例同时写Gin全局模式，移动SetMode到t.Parallel之前；保持并行执行，产品代码无需修复。
- frontend lint曾与Vitest/Vite加载配置并发，扫描到了已经删除的临时timestamp模块，报ENOENT；等待测试/构建结束后独立重跑通过，没有改eslint忽略规则。
- 测试移植时修正本地组件选择器、倍率小数格式与remote配置断言；embed新增请求清单曾误插入函数参数，已修正并重新通过整包embed测试。
- pnpm11仍提示旧overrides配置；build保留大chunk/Browserslist提示。audit告警及过期例外未隐瞒或延期，DOMPurify/axios维护仍在第二档S16；其它历史遗留以PORTING §5为准。
- OpenSpec CLI在当前环境不可用，执行结构与链接校验，不声称运行过CLI strict validation。

## 发布/部署记录

未提交、未打tag、未推送、未发布或部署；未操作现有数据库/凭据/运行实例。第二档16簇及第三/四档仍待定或排除，支付功能不在范围。浏览器临时入口、开发服务与浏览器会话已清理。
