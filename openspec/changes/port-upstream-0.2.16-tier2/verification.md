# 第二档实施验收

状态：S01–S16 已实现并完成验收（2026-10-18）。评估期结果见[基线报告](../../../docs/upstream-sync/evidence-0.2.16/baseline-tests.md)，不能填作本次实现验收。

| 实施信息 | 值 |
|---|---|
| 实施起点SHA | `882fe9777`（VERSION 1.1.17，工作区干净） |
| 实现SHA | `847db09e1`（S01–S15）+ `5ff6e5bf9`（S16），已快进合入 main；[代码/测试 SHA256 清单](../../../docs/upstream-sync/implementation-0.2.16-tier2/code-sha256.json)固定验收内容 |
| 实施日期 | 2026-10-17 至 2026-10-18 |
| 分支 | 实施分支 `sync/upstream-0.2.16-tier2`，2026-10-18 已快进合入 `main` |

| 簇 | 目标行为证据 | 反例/范围证据 | 结果 |
|---|---|---|---|
| S01 首次安装管理员校验统一 | 真实 SQLite `TestAdminBootstrapCredentialsSQLite`；Web `TestInstallAdminCredentialsAndExistingUsers`；YAML `TestEnsureConfiguredAdmin*` | 空库随机邮箱、登录校验、8/72 字节/Unicode；已有 admin/user 跳过、锁文件、完成通知；SetupWizardView 5 passed | 通过 |
| S02 旧版函数调用与拒答转换完整性 | `go test ./internal/pkg/apicompat -count=1`；来源 legacy/refusal 用例先失败后通过 | 同名 FIFO、现代 ID 避让、未配对结果、混合现代调用、纯 refusal、混合文本/工具、重复 supplement | 通过 |
| S03 显式工具变更 beta 保留及旧前置 | `go test -tags=unit ./internal/service -run "StructuredOutputs|ToolChangeBetas|ComputeFinal.*AnthropicBeta" -count=1` | 显式/缺省/相似 token/重复/逐项 drop；实际 HTTP 请求正文；count_tokens 原行为 | 通过 |
| S04 WebSocket 逐轮分组价格与图片用量 | 真实 WS 两轮分组改价/利润门；`TestOpenAIWSPassthroughTurnResultCarriesImageUsage`；relay 全包测试 | 同 Key 同组刷新；换组/平台/查询失败回退；用户与 Key 字段保留；交错价格时刻、tool-only 图片回填及重复终态去重；handler/relay 定向 race 通过 | 通过 |
| S05 受约束的上游兼容重试 | `go test -tags=unit ./internal/service -run "HTTPThinkingSignature|StructuredInput|ConvertedResponsesInputString|RawFallbackPreservesResponsesTiming|ChatFallbackPreservesProvidedStartTime|GPT61SolOnlyChatFallback" -count=1` | 签名三段诊断、单次恢复及续轮；完整 input 拒绝回退；重复键/截断/读失败/取消/已写/强制模式/Responses 入站不触发新重试；流与非流计时 | 通过 |
| S06 OAuth 搜索历史回放声明 | `go test -tags=unit ./internal/service -run OAuthWebSearch -count=1`；真实 HTTP transform/passthrough、WS 两轮 | 普通/Lite/compact/API Key；显式选择与声明不重复；大整数、转义类型、工具 map 不共享；既有 Codex/Lite/WS 回归通过 | 通过 |
| S07 组合路由公开模型目录 | `go test ./internal/handler -run "TestGateway.*Models|TestDefaultModelIDs" -count=1` | 精确/禁用/prefix/空别名，模型与 Codex Responses 目录分离；自定义列表与失败回退；组隔离用例由整体 backend unit 通过 | 通过 |
| S08 删除分组同步清理风控范围 | TestGroupDeletion*、TestModerationSaveAndGroupDeleteSerializeSQLite 与 service 保存拒绝用例 | Delete/DeleteCascade 同事务清理；缺失/畸形设置保留；错误回滚；多连接 WAL 保存先/删除先两种交错；race 通过 | 通过 |
| S09 分组倍率/RPM 编辑异步隔离 | 四个 GroupRateMultipliersModal/GroupRPMOverridesModal requests/edit 组件测试共 28 passed | 切组、关闭、卸载、初始打开；晚到加载/保存/清空/用户搜索；RPM 非负整数、倍率有限正数、批量溢出保护；定向 ESLint 通过 | 通过 |
| S10 异步页面、草稿、确认与弹窗隔离 | 全量 Vitest 与浏览器资料草稿核对；来源十组用例及补充 session/保存/卸载用例 | 切换页面/用户/账号/计划、关闭及卸载；旧成功/错误/finally/分页/nextTick 失效；共享 step-up Promise，卸载后迟到请求取消；公告成功标记后关闭 | 通过 |
| S11 GPT-6.1 Sol 上游订阅长上下文统计 | `go test -tags=unit ./internal/service -run "GPT61Sol.*(LongContext|UsesCredential|Exemption)|TestOpenAISubscriptionLongContextExempt|ModelLongContextExemption|CalculateOpenAIRecordUsageCost" -count=1` | OAuth/setup-token、实际父凭据、最终模型/别名、272k 边界；静态及目录价；API Key 长上下文保留；本地绝对区间价保留；账号成本与响应模型重新定价 | 通过 |
| S12 Gemini 既有路径的错误、额度和脱敏修复 | Gemini/Google 分类/redact 定向测试、全量 unit、相关 race；GeminiFlashPromotionExpiresWithoutRestart（含旧命名别名）与目录/override 测试 | 官方额度保留；第三方/Vertex 不套日配额；429 显式延迟或全局冷却；缓存缺失不封号换号；普通403仍按账号策略；促销 2027-01-01 UTC 到期恢复标准默认价，显式覆盖价保留 | 通过 |
| S13 监控配置补充已有平台入口 | MonitorSettingsPanel.spec.ts 6 passed，使用实际 Select 交互并模拟保存/重新读取 | 仅五个本地平台、禁止重复/无效平台、新平台空模型、原配置保留、保存期间禁改；定向 ESLint 通过 | 通过 |
| S14 Responses 无 namespace 时避免重建 | 来源 benchmark + `TestStripOpenAIResponsesInputNamespaces*`；3 次采样 8 MiB tools_only 从 25,190,400 B/3 alloc 降到 8,396,800 B/1 alloc | 原始空白、指数数值、大整数、转义、非对象与保留工具 namespace；畸形 JSON 回退原文 | 通过 |
| S15 Chat 流心跳保活 | `go test -tags=unit ./internal/service -run "TestHandleCCStreamingFromAnthropic|TestStripOpenAIResponsesInputNamespaces" -count=1` | io.Pipe 验证立即 flush；仅 SSE 注释、不设首 token；写失败停止；错误事件/读失败返回非 failover 错误且只发一次错误帧 | 通过 |
| S16 前端依赖安全维护与旧待办归并 | DOMPurify 3.4.17 / axios 1.20.0 / Vue 3.5.44 / source-map-js 1.2.2；pnpm 9 冻结锁检查、真实 DOMPurify 页面净化、API 客户端27项及完整前端回归 | 目标包告警清零；全树 103→70，生产高危剩 xlsx 两条；过期例外移除未续期；原支付SDK版本不变，未新增依赖能力 | 通过（剩余告警另列） |

完整输出和剩余项见[实施证据目录](../../../docs/upstream-sync/implementation-0.2.16-tier2/)。Go 在 backend 执行（自动使用 go1.27.0），pnpm 在 frontend 执行；[完整命令](../../../docs/upstream-sync/implementation-0.2.16-tier2/commands.json)，[测试清单](../../../docs/upstream-sync/implementation-0.2.16-tier2/test-inventory.md)记录构建标签。

| 门禁 | 命令/证据 | 结果 |
|---|---|---|
| backend unit/build | `go test -tags=unit ./... -count=1` / `go build -o /tmp/sub2api-tier2-check ./cmd/server`；[unit](../../../docs/upstream-sync/implementation-0.2.16-tier2/backend-unit.txt)、[build](../../../docs/upstream-sync/implementation-0.2.16-tier2/backend-build.txt) | 通过 |
| 相关race | S04/S05/S06/S08/S11/S12/S15 与相关解析；[输出](../../../docs/upstream-sync/implementation-0.2.16-tier2/backend-race.txt)；[补充别名 race](../../../docs/upstream-sync/implementation-0.2.16-tier2/flash-alias-race.txt) | 通过 |
| SQLite方言/用量去重/固定quota窗口 | `go test -race -tags=unit ./internal/repository -run 'Test.*(ProductionSQLUsesSQLiteDialect|UsageBillingRepositoryApplySQLite|UsageBillingDeletedKey|UpdateFixedQuotaResetAt)' -count=1`；[输出](../../../docs/upstream-sync/implementation-0.2.16-tier2/backend-sqlite.txt) | 通过 |
| embed静态资源与API别名 | `go test -tags=embed,unit ./internal/web -count=1` 与 embed 二进制构建 | 通过；[测试](../../../docs/upstream-sync/implementation-0.2.16-tier2/backend-embed.txt) / [构建](../../../docs/upstream-sync/implementation-0.2.16-tier2/backend-embed-build.txt) |
| frontend typecheck/lint/build | `pnpm run typecheck` / `pnpm run lint:check` / `pnpm run build` | 通过；[typecheck](../../../docs/upstream-sync/implementation-0.2.16-tier2/frontend-typecheck.txt) / [lint](../../../docs/upstream-sync/implementation-0.2.16-tier2/frontend-lint.txt) / [build](../../../docs/upstream-sync/implementation-0.2.16-tier2/frontend-build.txt) |
| frontend行为/浏览器验证 | `pnpm run test:run`；[输出](../../../docs/upstream-sync/implementation-0.2.16-tier2/frontend-tests.txt)、[浏览器](../../../docs/upstream-sync/implementation-0.2.16-tier2/browser.md) | 281 文件，2027 passed / 2 既有 skipped；浏览器通过 |
| dependency audit与剩余项 | [依赖报告](../../../docs/upstream-sync/implementation-0.2.16-tier2/dependency-review.md)、全树及生产 audit JSON、安全门禁结果、pnpm9 冻结锁验证 | 目标修复通过；全树仍70条，生产 high仍xlsx两条，安全扫描非零 |
| OpenSpec与文档一致性 | 16要求/32场景/54任务、来源冻结、链接、代码清单、git diff --check | 通过；[一致性记录](../../../docs/upstream-sync/implementation-0.2.16-tier2/consistency.txt) |

## 实施适配记录

- 冻结来源仍为 `128c5f7bf` / 上游 `01842576b3`；当前起点已包含 tier1 `7c544c03f`。保持 VERSION 1.1.17、迁移 226、SQLite/miniredis/simple mode；不发布部署。
- S01 已确认 createAdminUser 没有邮箱/弱密码校验，Web 在查库前校验；本地 YAML ensureConfiguredAdmin 也使用不兼容登录的邮箱规则与 128 字节上限，随本簇统一。
- S01–S16 已实现；未引入第三/四档、支付、缓存管理、Vertex 新认证或新模型能力。
- S06 实测冻结基线与当前工作区均缺少来源 `normalizeOpenAIResponsesWebSocketCompatibilityBody`。采用仅承担 web_search 历史声明的本地 helper，接入现有 HTTP、WS ingress、v2 adapter 和 HTTP bridge；不引入来源其它未选定归一化行为。
- S05 来源测试中的 5 个本地未实现平台路由场景排除；其余重试边界完整保留。

- S06 `getOpenAIRequestBodyMap` 改用已有严格 `decodeOpenAIJSONUseNumber`，避免搜索历史中的大整数损失；raw helper 兼容转义后的类型值。
- S14 来源 patch 会改写畸形 JSON；新增真实失败用例后在首次修改前验证 JSON，失败返回原正文。
- S15 原 Anthropic→Responses 转换器完全忽略 `error`，因此在 Chat 流出口显式处理 error 和读取失败，标记响应已提交，避免成功终态合成。
- S04 本地 passthrough 不调用 BeforeTurn，利用已有 BeforeRequest 处理后续轮次，按轮次隔离计费时刻与分组。

- S04 本地 relay 仅累计终态用量，没有上游 mergeRelayUsageNonZero/finalizeRelayTurnUsage 基座。按本地终态累加接入图片计数、兼容顶层 tool_usage，并按响应 ID 去重重复终态；保留现有解析失败规则。
- S07 排除 TypeSafe/ModelAllowlist 改造，保留本地 CustomModelsListEnabled 与目录聚合策略。
- S11 保留本地短参数签名与绝对区间价格；来源 FreeOpenAIFast 和相对区间倍率能力不引入。远程目录测试夹具明确提供 priority 单价，沿用本地既有 priority 解析。

- S09 本地没有来源的卸载清理基座，连同 listener/timer 清理加入；加载以外的保存、RPM 清空和搜索也按同一弹窗代次隔离。数值约束同时覆盖新增、编辑、批量及提交。
- S13 复用本地 Select 与五平台值，保存期间用原生 disabled fieldset 禁止配置编辑；不引入完整平台清单或未实现平台。

## 补充适配与未纳入项

- S08 用现有 SQLite 写事务与 JSON 读改写，覆盖普通和级联删除；专用设置写入口先取得写事务再校验有效组，复用调用方事务，不引入 FOR UPDATE/jsonb 或迁移。
- S10 在来源加载保护之外补齐本地保存/清空/搜索/用户切换边界；step-up scope 销毁后才收到 STEP_UP_REQUIRED 的请求也会取消。SQLite 当前跳过的 UsageCleanup、AccountExpiry、ScheduledTestRunner 没有启用。
- S12 本地缺少 Google 错误类型记录及部分分类基座，补 permission_error/not_found_error 分类；既有 SLA 用户原因排除仍保持，相关测试只更新正确的类型预期。缓存错误 helper 仅用于现有生成/countTokens，不引入缓存管理接口。
- S12 采用冻结来源的 Flash 促销价格，截止点明确为 `2027-01-01T00:00:00Z`。静态及目录默认价在读取时检查期限，跨年不依赖重启；显式本地价格覆盖优先。兜底过期检查按实际命中的价卡执行，覆盖既有点号/连字符及提供方前缀别名；对应回归先失败后修复。未引入上游完整计费时点、缓存存储费等系统。
- 最终 race 初次暴露既有 Chat 桥测试在 t.Parallel 后写 Gin 全局模式；同文件五处初始化移至 t.Parallel 前，保留并行测试。原始失败与修复后输出分别保留。
- S16 只更新已有目标依赖；Vue 对应编译器/Babel 解析变化属于依赖闭包。pnpm9 验证在隔离目录执行，工作区锁 SHA256 保持一致。xlsx 与工具链剩余项见依赖报告；没有将扫描非零伪装成通过。
- VERSION 1.1.17、迁移最大226、go.mod/go.sum、SQLite/miniredis/simple mode 和固定 quota-window 修复保持；冻结来源和历史评估证据未改。S16 接管的旧维护仅在本轮记完成，不回写旧任务框。

## 发布/部署记录

2026-10-18，用户在实施验收后要求提交、合入 main 并发布 v1.1.18。功能修复 `847db09e1` 与依赖维护 `5ff6e5bf9` 已快进合入 main；后续发布提交将 VERSION 更新为 1.1.18，并归档本报告与完整证据。实施期 VERSION 1.1.17、来源 SHA 与历史输出保持原值。

[详细发布说明](../../../docs/upstream-sync/implementation-0.2.16-tier2/release-v1.1.18.md)用于 annotated tag。此处记录打 tag 前的发布准备快照；远端推送及流水线结果以 GitHub tag / Actions / Release 为准。未部署本地或远端业务实例。

## Release 补充复核

2026-10-18 另外验证 Release 前置脚本、integration、Node20/pnpm9 实际冻结安装与前端构建、CGO_ENABLED=0 embed 版本注入，均通过。[发布复核记录](../../../docs/upstream-sync/implementation-0.2.16-tier2/release-readiness.md)说明未运行的远程发布环节、现有标签边界和测试临时库 URI 修正。代码清单已排除测试意外生成的 SQLite 文件，保留 138 个实现/配置/测试文件。

归档说明：原始 `.txt` 命令输出保留框架产生的尾部空格、混合缩进及捕获 diff 的空白。提交前源码/说明的 `git diff --cached --check` 排除该日志路径后通过；原始日志空白提示不作为实现测试失败，也不改写历史输出。
