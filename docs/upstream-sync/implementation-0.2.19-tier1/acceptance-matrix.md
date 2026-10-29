# v0.2.19 第一档行为验收

实施起点：`49939fb2cf96633688548ca39204d96558e59d79`。评估原始证据、两个冻结来源文件、第二档和 Terra 提案均按 [baseline.json](./baseline.json) 保留。下表的失败输出来自预期新行为的回归；旧评估 probe 的 PASS 不用作实施完成证据。

| 编号 | 行为和实际消费者 | 修复前 | 修复后与反例 |
|---|---|---|---|
| F01 | 首条 user 文本按 UTF-16 下标 4/7/20 采样，拼接后再按 JS UTF-8 哈希；正文 billing、header 版本同步和系统模板一致 | [baseline-claude-codex](./baseline-claude-codex.txt)：中文、emoji、surrogate 采样和正文/header 消费者均失败 | [claude-codex](./claude-codex.txt)、[独立 Node oracle](./claude-js-oracle.txt)：空值、短 ASCII、中文、emoji、采样后孤立/重组 surrogate、NUL、长尾、两种 content 形态、版本同步和幂等。生产采样循环最多保存覆盖下标 20 的 22 个码元；不转换整段 prompt |
| F02 | 本地 Codex manifest 刷新 flight 内复查 fresh 状态 | [baseline-codex-flight](./baseline-codex-flight.txt)：另一 flight 已填缓存后仍 fetch，计数 1→2 | [claude-codex](./claude-codex.txt)：确定性交错的后续 fetch=0；同一缓存指针复用。普通 miss、stale、ETag/304、账号/请求身份隔离、缓存复制、并发和等待者取消均通过 |
| F03 | compatible 顶层 prompt_cache_hit_tokens 回落到 cache-read；HTTP/SSE 到 SQLite 结算 | [baseline-cache-images](./baseline-cache-images.txt)：HTTP、缓冲 SSE、流式均丢失 hit80；[baseline-cache-hit-sqlite](./baseline-cache-hit-sqlite.txt)：实际账本把普通输入记为100而不是20 | [cache-images](./cache-images.txt)、[完整 SQLite 夹具](./cache-hit-sqlite-complete-fixtures.txt)：总输入100/cache80/普通20/output5；嵌套显式0、正值、既有顶层桶优先且不累加。真实本地 HTTP 和 SSE 转发后两次 RecordUsage 仍只有一条 usage、一条 dedup，余额和已配置 Key quota 只扣一次 |
| F04 | inline/file 的 camel/snake 图片 part 判定；完成与断流观察共用同一判定 | [baseline-cache-images](./baseline-cache-images.txt)：自定义别名两张 file 图计0；inline+file 同 part 的断流计数为2，数字 URI 被计费 | [cache-images](./cache-images.txt)：非空字符串 data/URI、图片 MIME、大小写、别名和两张图真实成本；同 part 只算1、非图片/空值/数字不算。累计 SSE 取最大值、usage 尾帧不抹计数、按次重置、旧模型 fallback；三协议断流仅保留已观察到的实际图片 |
| F05 | 选定 candidate 的 MALFORMED_FUNCTION_CALL 在聚合前成为失败 | [baseline-malformed-tools](./baseline-malformed-tools.txt)：JSON、wrapped、无 parts 错误尾帧、已有工具/文字后的错误均变为成功 | [malformed-tools-corrected](./malformed-tools-corrected.txt)、[真实 Forward 活流资源测试](./pagination-admin-and-resource.txt)：Messages、Chat、Responses 非流式均502/malformed_function_call，没有空成功或成功用量。普通文字/工具、非选中 candidate、取消和超时保持。错误终态后上游保持打开时及时返回，并由实际 Forward 关闭 body |
| F06 | 无符号安全解析、offset 与 offset+limit 边界；共享和严格入口各自保持策略 | [baseline-pagination-valid-fixtures](./baseline-pagination-valid-fixtures.txt)：超范围数字回绕、MaxInt offset/slice、严格入口越界均失败 | [pagination-admin-and-resource](./pagination-admin-and-resource.txt)：负数/空值/非法/超长数字、MaxInt、奇数 size、最大合法页、直接参数、Key 和仓储切片；共享入口默认/限幅，Ops/Prompt 严格入口400，合法上界保持200。最初源 Prompt 测试的 nil service 已换成现存 fake，避免把 fixture panic 当作生产缺陷 |
| F07 | 最后管理员降级、直接 disable/delete 返回领域400；保护规则保持 | [baseline-admin-guards-final-fixtures](./baseline-admin-guards-final-fixtures.txt)：六种禁止动作真实 handler500，行未修改；允许降级/基础设施500负对照通过 | [pagination-admin-and-resource](./pagination-admin-and-resource.txt)：三种稳定 reason、六个真实 SQLite 拒绝场景不改 role/status/username/updated_at；另一管理员存在时可降级；数据库执行错误仍500 |

## 回归入口

- F01：`TestClaudeCodeFingerprint_JavaScriptUTF16`、`TestClaudeOAuthSystemPromptTemplate_UnicodeFingerprint`、`TestSyncBillingHeaderVersion_UnicodeSuffix`、`TestBuildOAuthRequest_BillingMatchesWireUserAgent`。
- F02：`TestRefreshCachedAPIKeyCodexModelsManifest_RechecksFreshEntry` 与 `TestFetchCodexModelsManifestAPIKey*`。
- F03：`TestOpenAIUsage_PromptCacheHit*`、`TestUsagePromptCacheHitSQLiteSettlement`；独立 JS oracle 和旧 cache-reader overlay 均保存在此目录。
- F04：`TestGeminiImageOutputs_FileReferences`、`TestHandleNativeNonStreamingResponse_FileDataImageBilling`、`TestHandleNativeStreamingResponse_FileDataCumulativeChunks`，及已有 `TestGemini*Abort*` / `TestGemini*ClientDisconnect*`。
- F05：`TestMalformedFunctionCallTransform*`、`TestAntigravityMalformedFunctionCallNonStreaming*`、`TestAntigravityNonStreamingCollection_PreservesCancellation`。
- F06：`TestPaginationParamsOffset*`、`TestParsePagination*`、`Test_parseInt`、`TestPaginateAPIKeysLargePage`、`TestPaginateSliceLargePage`、Ops 和 Prompt pagination 边界测试。
- F07：`TestAdminService_AdminGuardsAreClientErrors`、`TestUserHandler_AdminGuardsSQLite` 与现存角色保护正反例。

## 实施过程失败的归因

- `malformed-tools.txt` 和早期 `baseline-pagination.txt` 被第一次错误落点的返回值数量编译失败打断；检查已移到正确收集器。它们不是最终验收。
- `baseline-pagination-corrected.txt` 含源新增 Prompt 测试的 nil-service panic；最终前后证据使用现存 fake，明确验证200→400和合法边界200。
- 早期管理员 SQLite 夹具比较了 time.Time 的 monotonic 位，改为比较持久化时间；`RAISE(ABORT)` 按现有契约是409，基础设施负对照改用执行时缺失 SQL 函数。最终输出验证真实500，不修改仓储错误分类。
- 早期账本夹具缺少 gpt-4o 价卡和 APIKeyQuotaUpdater 装配；改用既有 gpt-5.4 fallback 并补齐更新器。完整夹具通过后，对冻结旧 reader 的 overlay 仍失败，证明生产缺陷与夹具缺口分别成立。
- 首轮 lint 的两个 QF1003 位于新测试中，改用 switch 后最终 lint 为0 issues。首轮完整 unit 通过；后续普通和串行重跑均遇到既有 Aliyun 验证码运输错误用例失败。独立重跑3次及仓储诊断整包通过，但诊断里的关闭端口请求等待2秒才超时。
- 最初推测临时端口复用，继续量测后定位为环境代理：安装的 tea v1.3.13 SDK `getHttpProxy` 对 NO_PROXY 与完整 domain（含端口）做精确比较，`127.0.0.1` 不匹配 `127.0.0.1:port`。[控制探针](./captcha-proxy-proof.txt)使用临时本地代理复现“本地请求走代理，代理502成为 API error”；[隔离代理的12项检查](./captcha-direct-network.txt)全部通过，运输错误从2秒超时恢复到即时连接失败。相关生产和现存测试文件均保持原哈希；最终完整 unit 仅在该测试进程移除代理变量，不修改本批范围或运行配置。诊断 overlay 和实际 argv 均保存。
- 文档检查器初版将跨行的表格边界竖线误当成空单元格；改为只匹配行内空白后，[最终检查](./implementation-validation-final.txt)通过。初版输出保留，不作为通过证据。
