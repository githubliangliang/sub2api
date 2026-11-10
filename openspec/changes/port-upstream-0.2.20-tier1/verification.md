# Verification

状态：2026-11-10用户授权实施后，**F01–F13及44/44任务完成**。实现随本次提交保存，尚未归档；没有发布、push或业务实例部署。第二档仍未实施。

实际实施起点：`326513cdc68cc50d31e25a07ae98ecd6f5a3ca1f`，VERSION1.1.23 / Go1.27.3 / SQLite228 / auth snapshot20。起点已有本轮评估文档修改，不覆盖或提交这些历史产物。

实现提交：包含本文件本次状态更新的提交（`git log -1 -- openspec/changes/port-upstream-0.2.20-tier1/verification.md`可定位）；[product-hashes.json](../../../docs/upstream-sync/implementation-0.2.20-tier1/product-hashes.json)和[product.diff](../../../docs/upstream-sync/implementation-0.2.20-tier1/product.diff)记录提交前验证的可复核实现。来源SHA、原始评估与第二档提案保持原内容。

## 各簇实际验收

所有日志位于[implementation evidence](../../../docs/upstream-sync/implementation-0.2.20-tier1/README.md)。

| 簇 | 修复前证据 | 实现与回归证据 | 结果 |
|---|---|---|---|
| F01 Claude setup-token 周用量 | 修复前实际运行失败，见before-regressions-compiled.log | SetupTokenWeeklyUsage、WeeklyUsageWithoutFiveHourHeaders、SamplesWeeklyUsage；实际SQLite TestClaudeWeeklyUsageSQLitePersistence验证42%→100%及reset/5h不伪造。 | 通过 |
| F02 Codex 强制模式停止多余探测 | 修复前实际运行失败，见before-regressions-compiled.log | SkipsForcedModes两个强制模式0请求/0写入，auto/非法/非字符串对照；既有Codex探测header回归与全量通过。 | 通过 |
| F03 WebSocket 每轮首 token 时间 | 修复前实际运行失败，见before-regressions-compiled.log | EveryTurnFirstTokenWithoutResponseID、PerTurnFirstTokenNegativeClockAndDuplicateTerminal；第二轮700ms、带ID/负时间/重复终态/Fast tier保持，现有Relay取消与每轮模型回归通过。 | 通过 |
| F04 Codex WS→HTTP 桥接指纹 | 修复前实际运行失败，见before-regressions-compiled.log | StagesCodexFingerprint捕获最终HTTP header/body，复用gin context验证同账号稳定、跨账号隔离、off/API Key清空和显式prompt_cache_key不变；既有Grok桥接/Fast回归通过。 | 通过 |
| F05 缓冲 Responses reasoning 保真 | 修复前实际运行失败，见before-regressions-compiled.log | ResponsesReasoningContentFallback及PerItem覆盖summary优先/签名/空/多item；两条Buffered网关入口和真实HTTP+SQLite Fast结算夹具验证文字到客户端且费用/dedup保持。 | 通过 |
| F06 Responses 完成消息字段 | 修复前实际运行失败，见before-regressions-compiled.log | MessageContentWireConsistent、MessageRefusalPreserved、EmptyMessageContentArray；既有tool_search/function_call、历史序号和转换回归全量通过。 | 通过 |
| F07 鉴权取消分类 | 修复前实际运行失败，见before-regressions-compiled.log | LookupFailureWhenClientGoneReturns499与DeadlineAndBusinessClassification覆盖普通/Google、取消、请求期限、内部期限500、401、503；不新增底层错误日志或PG UTF-8策略。 | 通过 |
| F08 413 透传规则优先级 | 修复前实际运行失败，见before-regressions-compiled.log | 5个来源规则测试及RetryAfterAndStreamMonitoring；匹配/未匹配/默认413、开始SSE、skip-monitoring、Retry-After合法/非法及sequence0；保留本地stream SLA标记。 | 通过 |
| F09 Antigravity 直连 SSE 响应 | 修复前实际运行失败，见before-regressions-compiled.log | ResponseShapesPreserveContent与UsageWithoutContent覆盖direct/wrapped文字、thinking、工具、usage及终止；既有畸形工具终态和取消邻近检查通过。 | 通过 |
| F10 Antigravity JSON Schema 参数 | 修复前实际运行失败，见before-regressions-compiled.log | ResponseJsonSchemaBecomesResponseSchema、SnakeCase、ExistingResponseSchemaWins及NoSchemaLeavesBodyUntouched；源码仅在Antigravity clean入口接入，既有schema cleaner测试全量通过。 | 通过 |
| F11 GPT Image 账号测试参数 | 修复前实际运行失败，见before-regressions-compiled.log | OpenAIImageAPIKeyOmitsUnsupportedResponseFormat覆盖6个既有模型，捕获model/prompt/n及无response_format，默认base64成功；原账户测试全量通过，未引入image size路由。 | 通过 |
| F12 Anthropic兼容 Cloudflare REST URL | 修复前实际运行失败，见before-regressions-compiled.log | BuildAnthropicMessagesURL的25种正反例，EntryPoints贯通普通转发/透传/账号测试；严格模拟Cloudflare拒beta和URL allowlist拦截通过，认证/覆写header保留。 | 通过 |
| F13 Bedrock EventStream 分配前边界 | 修复前实际运行失败，见before-regressions-compiled.log | 修复前仅16字节合法CRC帧复现panic；修复后HeaderBeyondFrame及RejectsUntrustedLengths覆盖最大uint32/128KiB header/16MiB payload和合法边界，既有CRC/截断/合法frame通过。 | 通过 |

## 组合门禁

| 门禁 | 输出与实际范围 | 结果 |
|---|---|---|
| 定向修复前 | before-regressions-compiled.log：6包实际编译，25个顶层失败、8个正/负对照通过 | 缺陷确认 |
| 定向修复后 | after-targeted-final.log：6包91个顶层通过；additional-boundaries.log：4包15个顶层通过 | 通过 |
| 全量backend unit | backend-unit.log：54个有测试包通过；代理环境变量仅在测试子进程移除，沿用既有captcha隔离方式 | 通过 |
| embed build | backend-build.log；输出/tmp/sub2api-020-tier1-server | 通过 |
| lint | backend-lint-final.log：golangci-lint2.14.0，0 issues | 通过 |
| 定向race | backend-race.log：service/handler/middleware/WS/apicompat/antigravity共6包 | 通过 |
| 实际HTTP/SQLite | additional-boundaries及race覆盖本地HTTP上游、完整迁移、排队usage写入/父取消/重复RecordUsage、Fast/Free Fast真实费用；weekly真实SQLite Extra及查询 | 通过 |
| SQLite保护 | sqlite-protection.log：迁移226→227/227→228、dialect、quota/固定窗口/成功去重 | 通过 |
| 前端/浏览器/1C1G压力 | 无前端、预算或性能策略变更；本批不适用。没有借用旧前端通过数作为新验收 | 不适用 |
| 来源/范围/OpenSpec | scope-check.json、openspec-tier1.log、git diff --check | 见对应输出 |

## 实施适配与过程记录

- WS首token修复保留本地Fast每轮观测，未引入上游FirstTurnStartedAt/StartedAt接口。桥接指纹复用本地既有helper，不导入第二档Grok sanitizer。
- 413抽取共用规则判断，但继续调用本地handleStreamingAwareErrorWithCode以保留SLA归类，Retry-After仍经过原过滤。
- 鉴权只检查实际请求ctx取消；内部lookup deadline仍500。未复制上游zap.Error(err)，避免底层错误文本泄漏Key；PG UTF-8修补仍未选。
- 第一次来源测试编译缺Cloudflare上下文/响应fixture，在本地定义真实请求及SSE fixture后重跑失败回归。第一次lint发现无unit标签WS测试引用unit-only helper，改为该测试自建账号后lint/race通过。原失败输出保留。
- Bedrock先以小frame证实slice越界，再加长度guard，最后运行最大长度prelude拒绝测试；未在修复前运行4GiB分配。
- 全量unit后仅修测试helper构建标签依赖并做格式整理，产品行为未再修改；最终lint与六包race覆盖该测试终态。

## 证据边界

HTTP上游为本地httptest服务器或严格transport模拟，不是实际供应商在线验收；SQLite迁移、仓储、结算与去重是真实执行。无生产数据操作、外部账号消耗或供应商权限/速度保证。原始source-baseline、proposal/design/spec及上游评估保持冻结。VERSION、Go、依赖、迁移、Ent/Wire、frontend均未更改。
