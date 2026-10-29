# 上游 v0.2.19 移植评估（v0.2.18 → v0.2.19）

来源冻结及评估日期：2026-10-28；文档验收于2026-10-29完成。上游 [v0.2.19 release](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.19) 于 2026-10-28 10:16 UTC / 18:16 上海时间发布。Git tag 终态是本轮来源，发布说明用于交叉核对。

结论：建议第一档 7 簇、第二档 6 簇，优先 Claude Code、Codex、Grok 既有路径。用户在本轮明确需要 GPT-6.1 Terra，保留其**第三档新增能力**身份，单独生成 T01 提案；不悄悄提升为修复档。当前完成的是评估及验收规划，三份 change 的实施任务全部未勾选。

- [第一档提案](../../openspec/changes/port-upstream-0.2.19-tier1/README.md)
- [第二档提案](../../openspec/changes/port-upstream-0.2.19-tier2/README.md)
- [用户选定 Terra 提案](../../openspec/changes/port-upstream-0.2.19-terra/README.md)
- [来源、复测与证据目录](./evidence-0.2.19/MANIFEST.md)

## 1. 版本、范围与规模

| 项目 | 冻结值 |
|---|---|
| fork HEAD | 49939fb2cf96633688548ca39204d96558e59d79 |
| fork VERSION / Go | 1.1.20 / 1.27.3 |
| 最大 SQLite 迁移 | 227 |
| 上次已评估 tag | v0.2.18 = ac1b8f1271be0d674fa6f074403c1410f30a3a81 |
| 新 release tag | v0.2.19 = dece26153621959ccd7635814f4b2d08304e681a |
| 原始采集时 upstream/main | a9240d5016c226b341242a8c54a01a8605cbbc72，仅记录，不作为取码来源 |
| 本轮 fetch 后 upstream/main | a3298b46ebe239fb0720b068a1eb3383955bdac2；tag 之后的内容不纳入 |
| 非 merge 增量 / 候选 | 50 提交 / 23 候选 = 20 PR + 3 直接提交 |
| 整区间净差异 | 259 文件，+22597 / -650 |
| 按候选计数的文件态 | 276 条：ALREADY 0 / CLEAN 203 / CONFLICT 61 / NOBASE 12 |

276 是各候选的文件记录之和，同一文件可重复出现；不是 259 个净变化文件的另一种计数。Grok #8339 的 21 个微提交整取为一个 merge diff，不逐个 cherry-pick。旧作者日期不影响 tag 祖先范围，#8003 仍在本轮增量中。

| 档位 | 簇与覆盖 | 当前状态 |
|---|---|---|
| 第一档 | F01–F07，7 簇 / 7 PR | 本地缺陷确认，低成本，无新 schema，提案未实施 |
| 第二档 | S01–S06，6 簇 / 4 PR + 1 直接提交 | Grok、图片、前端和依赖需要手工适配；提案未实施 |
| 第三档 | 7 PR 含混合子项 | T01 Terra 已由用户选定并立独立提案；其它项等待实际用途选择 |
| 第四档 | 5 PR + 2 直接提交，另有混合 PR 的不取实现 | 无本地消费者、无此缺陷、上游资源或不适用实现 |
| 合计拟议范围 | 14 簇 / 12 个不同 PR + 1 直接提交 | 三份提案；不是整 release 移植 |

S01/S02 来自同一 PR，第三档中的诊断界面也在该 PR 内，因此不能把档位的 PR 数直接相加。图片用量、token 成本与账号配额属于个人运行能力；支付、充值、购买订阅、退款仍排除。

## 2. 方法与适配边界

1. 本轮已执行 fetch upstream --tags --prune，并再次通过 GitHub release API 确认 latest 是 v0.2.19；[在线来源记录](./evidence-0.2.19/review-online.json)保留发布时间及 Go 官方 advisory。
2. 以 fork HEAD 创建独立临时 index，对全部 276 个文件先完成反向检查，再完成正向检查；未向工作树或真实 index 应用补丁。[复查脚本](./evidence-0.2.19/review-intake.py)、[文件复查](./evidence-0.2.19/review-file-checks.tsv)与原始四态完全一致。
3. 工作区预存本版本证据；本轮核对 23 份完整 PR/直接补丁、50 提交覆盖和 3732 个产品文件哈希，复跑定向测试及缺陷探针。原始 baseline.json 的早期 status/文档提交转换保留历史；当前起点见 [review-baseline.json](./evidence-0.2.19/review-baseline.json)，不重写旧记录。
4. [新增调用原始清单](./evidence-0.2.19/added-call-inventory.tsv)是宽松扫描，包含测试、内置函数、注释和 SQL；不能把 3637 个 token 当作 3637 个缺函数。拟议范围的 328 条产品调用记录、真正缺的前置及消费者见 [调用审查](./evidence-0.2.19/call-review.md)和[符号检索](./evidence-0.2.19/review-base-symbols.txt)。
5. SQLite-only、miniredis、simple mode、成功 usage 去重、已完成的固定 quota 窗口推进及迁移227保留。第一/二档和 Terra 均不需要新 schema。上游迁移247受限密钥表、248图片计量字段属于未选定新能力，不能直接导入或抢占本地编号；以后选定时从当时本地最大编号重新分配。
6. 本地源文件与上游名称不同，按实际调用链适配。源文件新增而 CLEAN，仍可能含 PG SQL、未采用的权限或目录基座。测试必须看首行和 body：[integration 标签清单](./evidence-0.2.19/review-integration-build-tags.tsv)。本轮源 Grok 两个 integration 文件只有 integration 标签，却调用 PG 夹具及 SQL；验收须改为真实 SQLite 夹具，不能只改标签。

### 全部候选分档

四态顺序均为 ALREADY / CLEAN / CONFLICT / NOBASE。每行使用整 PR 状态，子项选择不改变冻结数值。

| 候选 | 非 merge 数 | 四态 | 档位 | 判断与范围 |
|---|---:|---|---|---|
| [5c185c5c5](https://github.com/Wei-Shaw/sub2api/commit/5c185c5c5d4f5e409e2139806e0964ae944e03e8) | 1 | 0/0/1/0 | 第四档 | 上游 VERSION 同步；fork 保持 1.1.20 |
| [#8319](https://github.com/Wei-Shaw/sub2api/pull/8319) | 1 | 0/86/6/0 | 第三档 | 受限管理员密钥新能力；未选定，含新 schema/授权/前后端 |
| [3f1c8f36c](https://github.com/Wei-Shaw/sub2api/commit/3f1c8f36ce8c2cd378015830a90e89970b9a9ea7) | 1 | 0/6/3/0 | 第四档 | 上游赞助商资源与宣传 |
| [#8323](https://github.com/Wei-Shaw/sub2api/pull/8323) | 5 | 0/28/13/0 | 第三档 / 已满足子项 | 按实际图片张数计量未选定；渠道按次图片档位保留 token 已由本地实现满足 |
| [#8364](https://github.com/Wei-Shaw/sub2api/pull/8364) | 1 | 0/11/7/0 | 第三档（用户已选定） | T01 GPT-6.1 Terra，独立提案，未实施 |
| [#8339](https://github.com/Wei-Shaw/sub2api/pull/8339) | 21 | 0/31/12/1 | 第二档 / 第三档 | S01 响应模型审计、S02 目录一致性；新增诊断接口与界面仍按需 |
| [#8286](https://github.com/Wei-Shaw/sub2api/pull/8286) | 1 | 0/1/0/1 | 第四档 | 缺 ratelimit_cn_providers.go 及 CN Coding Plan 消费者 |
| [#8377](https://github.com/Wei-Shaw/sub2api/pull/8377) | 1 | 0/2/1/0 | 第一档 | F01 Claude Code UTF-16 指纹 |
| [#8326](https://github.com/Wei-Shaw/sub2api/pull/8326) | 1 | 0/2/0/0 | 第一档 | F04 Gemini fileData 图片用量 |
| [#8003](https://github.com/Wei-Shaw/sub2api/pull/8003) | 1 | 0/0/1/1 | 第四档 / 本地无此缺陷 | 本地没有 Pro 按日期切换为 Flash 的实现 |
| [#8366](https://github.com/Wei-Shaw/sub2api/pull/8366) | 1 | 0/0/0/2 | 第四档 | 缺 model_plaza_service.go；现有账号模型展示不是这个分组定价消费者 |
| [#8354](https://github.com/Wei-Shaw/sub2api/pull/8354) | 1 | 0/2/0/0 | 第一档 | F03 共享 OpenAI usage 缓存命中字段 |
| [#8376](https://github.com/Wei-Shaw/sub2api/pull/8376) | 1 | 0/0/0/2 | 第四档 | OpenCode 原生账号与 strict-role 消费链未引入 |
| [#8303](https://github.com/Wei-Shaw/sub2api/pull/8303) | 2 | 0/11/0/0 | 第一档 | F06 分页数值与偏移量溢出 |
| [#8348](https://github.com/Wei-Shaw/sub2api/pull/8348) | 1 | 0/2/2/0 | 第二档 / 第三档 | S03 上传完整性；20→50 MiB 上限扩张仍按需 |
| [#8362](https://github.com/Wei-Shaw/sub2api/pull/8362) | 1 | 0/5/0/0 | 第一档 | F05 Antigravity 非流式畸形工具调用 |
| [#8352](https://github.com/Wei-Shaw/sub2api/pull/8352) | 1 | 0/2/0/0 | 第一档 | F07 管理员保护规则返回 400 |
| [0a2a29b32](https://github.com/Wei-Shaw/sub2api/commit/0a2a29b32e6724714c69822296ccd05a6b86bfdd) | 1 | 0/0/2/0 | 第二档 | S06 x/net 修复与最小依赖图 |
| [#8390](https://github.com/Wei-Shaw/sub2api/pull/8390) | 2 | 0/3/3/0 | 第二档 | S04 渠道 Web Search 平台维度保存 |
| [#8395](https://github.com/Wei-Shaw/sub2api/pull/8395) | 2 | 0/8/4/2 | 第二档 / 第三档 | S05 既有 Gemini 列表/详情/测试默认值；新模型目录扩展仍按需 |
| [#8287](https://github.com/Wei-Shaw/sub2api/pull/8287) | 1 | 0/2/4/1 | 第三档 | Gemini 同账号重试配置未选定；当前总尝试数已为 5 |
| [#8369](https://github.com/Wei-Shaw/sub2api/pull/8369) | 1 | 0/1/1/2 | 第四档 | 缺显式缓存绑定与 pool retry 消费者 |
| [#8373](https://github.com/Wei-Shaw/sub2api/pull/8373) | 1 | 0/0/1/0 | 第一档 | F02 Codex 模型缓存 flight 内复查 |

[完整候选及父 SHA](./evidence-0.2.19/candidates.tsv)、[微提交清单](./evidence-0.2.19/commits.tsv)、[hunk 落点](./evidence-0.2.19/hunks.tsv)保留全部内容。

## 3. 第一档：既有缺陷的小范围修复

| 簇 | 来源 | 整 PR 四态 | 本地证据与受益路径 |
|---|---|---|---|
| F01 Claude Code 指纹 | #8377 | 0/2/1/0 | 中文/emoji 用 Go 字节索引而不是 JS UTF-16 码元；独立 JS oracle 的三个样本均不同。Claude OAuth 正文与 billing header 同时受益。 |
| F02 Codex 模型缓存 | #8373 | 0/0/1/0 | 调用方 miss 后另一次 flight 已填缓存，旧调用进入新 flight 仍拉上游；确定性交错探针复现一次多余请求。 |
| F03 共享缓存命中用量 | #8354 | 0/2/0/0 | OpenAI-compatible usage 丢 prompt_cache_hit_tokens=80；无需引入 DeepSeek 平台，现有 OpenAI 自定义 base URL 消费者已可达。 |
| F04 Gemini fileData 图片 | #8326 | 0/2/0/0 | 自定义别名下两张实际 fileData 图片被记为0；既有 Gemini 消息、Chat、native 及 Antigravity 计数器受益。 |
| F05 Antigravity 畸形工具终态 | #8362 | 0/5/0/0 | MALFORMED_FUNCTION_CALL 转成空 content 的成功 Claude message；非流式转换和缓冲兼容调用可达。 |
| F06 分页溢出 | #8303 | 0/11/0/0 | MaxInt 页码算出 offset=-40；共享管理/Key 分页、ops 与 prompt audit 都有消费者。 |
| F07 管理员保护状态码 | #8352 | 0/2/0/0 | 最后管理员降级错误被归类为500；disable/delete 同样用普通 error。应保留保护规则并返回400。 |

复现见 [基线探针](./evidence-0.2.19/review-baseline-probes.txt)；其 PASS 表示旧缺陷被稳定观测，不表示已修好。

### F01 patch sites / 前置

- backend/internal/service/gateway_billing_block.go：computeClaudeCodeFingerprint；gateway_billing_header.go 与 gateway_claude_oauth_body.go 已调用同一 helper，不能只验一个入站分支。
- 只补 UTF-16 采样及拼接后 UTF-8 编码。unicode/utf16.AppendRune/Decode 属于当前 Go 标准库；最多读取覆盖索引20的前缀，不把整条长 prompt 转成数组。
- 验收 ASCII、中文、emoji、孤立 surrogate 及采样后重新组成 pair；不足下标仍补0，版本更新与重复调用幂等。

### F02 patch sites / 前置

- backend/internal/service/openai_codex_models_service.go：本地函数是 refreshCachedAPIKeyCodexModelsManifest，而上游修改 refreshCachedOpenAIModels。CONFLICT 来自不同缓存演进，实际 TOCTOU 缺陷相同。
- 本地 codexModelsManifestCache 的 get、fresh/stale 状态和 singleflight 均存在；只在 flight 内重新读取 fresh 状态，不引入完整标准模型目录。
- 验收“先 miss、后别的 flight 填入、再进入 refresh”的 fetch=0；stale、ETag/304、账户隔离、并发和取消规则仍按现有行为。

### F03 patch sites / 前置

- backend/internal/service/openai_gateway_response_handling.go：openAICacheReadTokensFromUsage；openAIUsageFromGJSON 和 HTTP/SSE 解析已共享此入口。
- 新字段作为顶层 fallback；嵌套 input_tokens_details.cached_tokens / prompt_tokens_details.cached_tokens（包括显式0）继续优先，避免重复累计。
- 验收 prompt=100、hit=80、miss=20，各桶保持语义；同一请求的 streaming/nonstream usage 与 SQLite dedup 只落一次。它不修 #8003 的日期定价，也不增加新平台。

### F04 patch sites / 前置

- backend/internal/service/gemini_image_output_accounting.go：countGeminiInlineImageOutputs、part 判定及现有 observer；扩成 inline/file 两种图片引用。
- isGeminiInlineImageMIMEType 已存在；camelCase/snake_case、非空 data/URI、同一个 part 至多一张、累计 SSE 取最大值一起验收。
- 保留已观测输出及旧模型名 fallback；不下载 fileUri、不扩大图片 endpoint/计费模式或生图开关，不导入 #8323/#8167。

### F05 patch sites / 前置

- backend/internal/pkg/antigravity/response_transformer.go、service/antigravity_gateway_streaming.go、antigravity_gateway_compat.go：新增明确错误标识，在聚合丢失无 parts 的尾帧之前检查。
- extractGeminiFinishReason、errors.Is 与既有 compat 错误出口均存在；新增 sentinel 随消费者一起接入。
- 验收直接 JSON、wrapped JSON、SSE 无 parts 错误尾帧、普通 text/tool 正例和取消；客户端收到既有格式的502错误，不收到空成功或虚构完成，用量与失败策略保持现有契约。

### F06 patch sites / 前置

- backend/internal/pkg/pagination/pagination.go、pkg/response/response.go、service/api_key_service.go、handler/admin/ops_dashboard_handler.go、securityaudit/prompt_handler.go。
- strconv.ParseUint 与 math.MaxInt 为标准库；共享 offset 必须保证 offset+limit 也不溢出。各入口维持原有“默认/限幅”或“400”策略，不统一改成同一种错误。
- 验收超长十进制、MaxInt、负数/非法/空值、最大合法页和空结果；直接构造 PaginationParams 与 HTTP query 都覆盖。

### F07 patch sites / 前置

- backend/internal/service/admin_user.go：ensureNotLastAdmin、UpdateUser、DeleteUser；infraerrors.BadRequest 已导入且存在。
- 只改领域错误分类及稳定错误码；数据库错误仍500，最后管理员及禁止直接 disable/delete 的规则不放宽。
- 验收真实 handler 的400、禁止动作无数据修改、另一个管理员存在时正常降级。

## 4. 第二档：手工适配与跨路径验收

| 簇 | 来源 | 整候选四态 | 必须实际完成的适配 |
|---|---|---|---|
| S01 Grok 响应模型审计 | #8339 | 0/31/12/1 | 原生 Responses 返回审计字段；Chat/Messages 缓冲路径保留前序模型及冲突位。 |
| S02 Grok 目录一致性 | #8339 | 同上，不重复累计 | 手动/后台统一解析，成功快照保留，SQLite 单键 CAS、事务及普通编辑受管字段。 |
| S03 图片上传完整性 | #8348 | 0/2/2/0 | OpenAI 拒绝超限而非截断；Grok 已有整体 body 边界内保持解析完整。上限扩大另列第三档。 |
| S04 渠道搜索维度 | #8390 | 0/3/3/0 | 现有 OpenAI 搜索已实现，backend 旧客户端合并与 frontend 真实保存链共同修复。 |
| S05 既有 Gemini 模型接口 | #8395 | 0/8/4/2 | 声明映射列表/详情及测试默认值、已有目录展示；只摘必需 helper，不引入整 Vertex/allowlist 系统。 |
| S06 x/net 维护 | 0a2a29b32 | 0/0/2/0 | HTML 复杂度 PoC 与最小 MVS 图确认，需要 backend 编译/工具链/网络回归。 |

### S01 patch sites / 前置

- backend/internal/service/openai_gateway_grok.go：forwardGrokResponses 的返回结果缺 UpstreamResponseModel/Conflict；上下文已观测到模型，但返回结构丢字段。
- openai_gateway_messages.go：readOpenAICompatBufferedTerminal / handleAnthropicBufferedStreamingResponse；openai_gateway_chat_completions.go：handleChatBufferedStreamingResponse。上游增加“固定大小前序审计”，只针对 Grok。
- 本地 upstreamResponseModelObserver、观测字段与 usage 消费者都存在。新增前序状态随 buffered reader 接入，修改签名时更新实际调用点和 test stub；不带诊断 API/UI。
- [额外 Grok 探针](./evidence-0.2.19/review-grok-audit-probes.txt)验证原生 JSON/SSE 返回字段空，以及两个缓冲转换把4.6→4.7冲突丢掉。验收须继续穿过 handler 异步任务和 SQLite usage 落库；只测 helper 不足。
- 请求模型、映射模型、响应原文模型各自保留；终态模型优先、前序不同模型只保留冲突位，状态不随事件数增长，非 Grok 行为保持。

### S02 patch sites / 前置

- backend/internal/service/grok_observed_models.go、upstream_models.go：本地两套解析不同，错误对象能成为模型ID。合并时只接受有实际协议 ID 的有效形态，保留未知自定义 ID 大小写，去重排序。
- 新 grok_model_catalog.go 为310行、grok_model_catalog_store.go 为112行；包含解析、来源范围、时间排序、固定观测 CAS 重试。grok_model_diagnostics.go 的265行及新增 endpoint/panel 不是正确性修复前置。
- backend/internal/repository/account_repo.go 的 lockAndMergeAccountProbeExtra 与 service/admin_account.go：最新数据库的 grok_observed_models / upstream_model_metadata 属于受管状态，不接受旧表单、缺键或注入覆盖。真实 SQLite 探针已复现旧表单覆盖新目录及 metadata。
- 上游 account_repo_grok_models.go 新文件虽 CLEAN，却用 jsonb_set、类型转换与 PG 比较。需要 SQLite 单键原子更新、语义等价 expected 比较和有界重读重试；保留其它 extra、软删/非Grok拒绝及取消。
- 缺 accountTransactionScope 的来源是 #8319 的 custom_admin_key_transaction.go：函数共11行。它只是事务客户端和缓存边界；本地已有 clientFromContext/withRepositoryTransaction。只实现需要的事务效果，不以此引入受限管理员密钥。
- 缺 resolveModelsListReadLimit 是6行 helper，服务的是未采用的新读取配置。本地 upstreamModelsBodyLimit=8 MiB 且手动同步已做 limit+1 拒绝；后台同步复用这个边界，不扩配置或照旧静默读取1 MiB。
- 验收旧快照读取、失败不抹成功、乱序请求不能倒退、非法时间、空目录记empty_catalog失败并保留成功，与非法形态区别、CAS竞争、外部事务回滚无提前缓存发布、普通编辑与观察重叠、无敏感错误日志。

### S03 patch sites / 前置

- backend/internal/service/openai_images.go：parseOpenAIImagesMultipartRequest 已对20 MiB截断后当成功；改用 limit+1 和明确拒绝，保留每个 part 字节完整性。
- grok_media.go：parseGrokMediaMultipartRequest 的 LimitReader 也会把解析/审查输入截短。源补丁移除此限幅的依据是原请求整体 body 已受限；必须沿真实 handler 验证总大小限制仍在，不能给裸 multipart reader 无限信任。
- openai_images_responses.go 的上传日志是 MIME/大小元数据；不是修复必需，不能记录正文/凭据。上下游 alias、重试、mask、filename、下载20 MiB限制分别验证。
- 20→50 MiB 是附带的限制扩张，用户未选择；S03 不以其为前置。在当前20 MiB边界验证刚好允许、+1拒绝，Grok 在总 body 上限内完整解析或明确失败，不静默截断。

### S04 patch sites / 前置

- backend/internal/service/channel_service.go：applyUpdateInput 整块替换 FeaturesConfig，部分平台更新删掉此前 OpenAI 维度。源新增两个合并 helper 函数共36行（含说明的增量约60行），零 schema。
- frontend/src/views/admin/ChannelsView.vue 与现有 i18n：核对编辑→平台 section→formToAPI→API 保存全过程。用户已在0.2.18选定 OpenAI Web Search，所以此处是现有能力的缺陷。
- 本地只采用实际已有消费者的平台白名单；不复制上游 command_code/opencode_go/minimax/zhipu 等缺平台条目。显式 false、空对象或 null 的清除语义，与缺键继承分开。
- 验收 OpenAI/Anthropic 同时存在、旧客户端只知一维、隐藏全局开关、禁用平台、显式清空、未知维度保留、其它 feature 原整块替换语义和输入 map 不被修改。simple mode 的现有页面限制保持。

### S05 patch sites / 前置

- backend/internal/handler/gemini_v1beta_handler.go：列表/详情只选一个账号透传；/v1/models 已从 model_mapping 求可用模型。全部可调度 Gemini 账号都声明映射时，两个协议入口应一致，未映射账号仍透传。
- 源 gemini_models_endpoints.go 的完整终态344行含未采用的模型端点/Vertex扩展，不能整取。新增 DeclaredModelIDs 与 DeclaredModelsCover 的函数共34行，含注释/空行的增量41行；所需 listSchedulableAccountsOnce、GetModelMapping、normalizeRequestedModelForLookup、matchWildcard 在本地都存在。
- 源 AntigravityGeminiModelIDs 为23行，属于现有 mixed consumers 的目录收集；按本 fork 既有 force/composite/mixed 行为摘取。缺 ModelAllowlistEnabled/FilterForListing 时使用已有 CustomModelsListEnabled/ModelsListConfig 规则，不引入旧按需 allowlist schema。
- backend/internal/pkg/gemini/models.go、pkg/geminicli/models.go、service/account_test_service.go、frontend/src/composables/useModelWhitelist.ts：去除公共默认列表中的源已弃用2.x/3-pro-preview，普通连接测试改用已有3.5-flash；Google One保守映射、显式指定旧模型与自定义别名保留。
- 综合选择列表可包括当前已支持的 Antigravity 条目并去重；不顺带添加 Gemini4 Argon、3.6/3.7/3.8或新的 flash-lite-image 能力/价卡。
- CustomModelsList是既有展示规则，不把它改成新的请求授权ModelAllowlist；详情仍尊重实际账号映射和原授权边界。
- 验收 all-mapped 与 partially-unmapped、列表/详情404及前缀通配、原路径安全校验、custom list、强制Antigravity/composite边界、Google One/显式模型、前后端默认目录一致。

### S06 patch sites / 前置

- backend/go.mod/go.sum：当前 x/net=v0.59.0；[Go 官方 GO-2026-6860](https://pkg.go.dev/vuln/GO-2026-6860)指出相邻 HTML text 合并产生平方复杂度，修复版本0.62.0。当前 Go1.27.3 本身不修第三方 x/net。
- 实际消费者为 backend/internal/service/ollama_cloud_usage_parser.go 的 html.Parse，经既有上游 usage/probe 获取HTML；这不是匿名 gateway 可利用性的证明，但调用点和1C1G内存影响可验证。
- [隔离 PoC](./evidence-0.2.19/html-complexity-probe.go.txt)：输入从约40 KiB倍增至80 KiB，旧版分配35.85→141.96 MB；0.62版约0.114→0.216 MB。[旧版](./evidence-0.2.19/html-complexity-current.txt)、[隔离新版](./evidence-0.2.19/html-complexity-fixed-version.txt)证明的是解析复杂度，不是完整进程压测。
- [最小 MVS](./evidence-0.2.19/x-net-minimum-mvs.json)在临时 modfile 只指定 x/net0.62后，强制变化为 x/crypto0.58、x/net0.62、x/sync0.24、x/sys0.49、x/telemetry ed294f943157、x/term0.47、x/text0.43、x/tools0.51。不能把源整份go.mod覆盖本地，也不能声称只改一个模块。
- [隔离编译/相关测试](./evidence-0.2.19/review-dependency-feasibility.txt)通过；httpclient 包该筛选下无匹配测试，只证明编译。实际实施需 full unit/build/lint、Ent工具兼容、相关HTTP/2/race及漏洞复扫。Go、VERSION及现有生成代码不因提案改变。

## 5. 按需、不做与历史遗留

### 5.1 第三档：T01 已选定，其它能力继续等待用途选择

| 子项 | 判断与成本 | 当前选择 |
|---|---|---|
| T01 #8364 GPT-6.1 Terra | 18文件 / +292 -17，四态 0/11/7/0；模型识别、effort、工具协议、两层价卡、资源、Codex目录及现有客户端配置，属于新能力 | 用户明确需要，已立[独立提案](../../openspec/changes/port-upstream-0.2.19-terra/README.md)，未实施 |
| #8319 受限管理员密钥 | 92文件 / +9054 -131；授权范围、不可逆摘要、IP/过期/撤销审计、schema和UI，需要SQLite与个人权限设计 | 尚未选定，不作为S02前置 |
| #8323 图片按实际张数计量 | 41文件 / +728 -78；分组新列和默认关闭开关，真实返回张数、endpoint与定价模式门控 | 尚未选定；它是usage能力，不能误判payment-only |
| #8339 Grok目录诊断界面 | 诊断服务265行、DTO/endpoint/API、诊断panel和编辑状态交互 | 尚未选定；S01/S02只取既有行为正确性 |
| #8287 Gemini重试配置 | 7文件 / +150 -15；默认4次重试+首次请求，本地已固定5次总尝试 | 尚未选定；不是必须新增设置才能修现有问题 |
| #8348 上传上限20→50 MiB | 额外资源边界/内存及上游兼容选择 | 尚未选定；S03保留20 MiB并解决静默截断 |
| #8395 新Gemini目录项 | Gemini4 Argon、3.6/3.7/3.8及新flash-lite-image | 尚未选定；已有目录清理与接口一致性可独立做 |

### T01 patch sites / 前置与验收

- backend/internal/pkg/openai/constants.go：IsGPT61TerraModelSpelling 及 RequiresResponsesForToolCalls 两个函数共19行；CanonicalizeOpenAIModelAliasSpelling 和 Sol规则已在本地。只接受 base、none/low/medium/high/xhigh/max/compact 及已有供应商前缀/拼写规范；pro/ultra/minimal不能被当作Terra合法 effort。
- service/openai_codex_transform.go、openai_model_alias.go、openai_compat_model.go、openai_codex_models_service.go 与 pkg/apicompat/anthropic_to_responses.go：同一个 canonical ID贯穿映射、effort和manifest。Terra允许none，max保留，默认medium；Sol仍拒绝none/minimal。
- service/openai_gateway_chat_completions_raw.go、openai_gateway_messages_chat_fallback.go、openai_gateway_responses_chat_fallback.go：有工具的Terra需要Responses；only-Chat账号在出站前返回400，不能只加模型下拉就声称工具可用。
- service/billing_service.go、pricing_service.go、resources/model-pricing/model_prices_and_context_window.json：只摘Terra条目和现有策略接入。源基础 input/cache-read/cache-write/output 每百万为1/0.05/1.25/5；priority为2倍；长上下文阈值272000、input倍率2/output倍率1.5。以上是冻结上游补丁值，不能冒充当前官方售价或对任意账号的可用性保证。
- 资源的1,050,000 context/128,000 output是缺省信息；账号明确metadata优先且上游目录不能被就地改写。公共catalog源created=1704067200是占位值，不当作实际发布时间。
- frontend/src/composables/useModelWhitelist.ts 与 components/keys/UseKeyModal.vue：已有白名单/映射预设和OpenCode客户端配置加Terra。这里的OpenCode配置生成是现有客户端能力，不是引入缺失的OpenCode账号平台。
- 源新增Terra测试CLEAN但引用本地不存在的 FetchOpenAIModelsList、FetchOpenAIAccountModels、ordinaryModelsUpstreamResponse。实际本地消费者是 FetchCodexModelsManifest / FetchUpstreamSupportedModels；按这些路径适配测试，不为测试引入未选定完整标准目录系统。
- 不把Terra设成全局默认，不扩大Fast/Free Fast/Ultrafast机制或账号权限。验证官方/自定义上游metadata、HTTP/SSE/现有WS桥、三个only-Chat工具拒绝路径、价格显式0及override、usage/dedup、旧模型及UI生成配置；细节见Terra change。

### 5.2 第四档与本地无此缺陷

- #8286：生产 ratelimit_cn_providers.go 缺失。新test CLEAN只是新文件可加入，不证明本地已有CN Coding Plan频率窗口缺陷。
- #8003：本地没有 deepseekProRoutesToFlashAt / deepseekProBilledAsFlash / 峰谷强制覆盖链；既有Pro/Flash价卡分离。保留现有DeepSeek-compatible token成本，不为了移除上游旧日期分支引入整套峰谷定价。
- #8366：本地没有 model_plaza_service.go；现有 account_model_plaza.go 负责另一种账号模型展示，不能用相似名称替代分组定价消费者证据。
- #8376：openai_chat_roles.go 与OpenCode原生平台缺失；这是扩大已有strict-role逻辑给OpenCode/DeepSeek的修复，不是F03缓存字段的前置。
- #8369：本地有 cached-content-not-found 错误兼容，但没有显式cache绑定/上游池模式pinning；小判断函数也不能制造当前消费者，不引入旧显式缓存CRUD基座。
- 5c185c5c5及3f1c8f36c：上游VERSION同步/赞助商宣传不复制。
- 混合PR中的PG SQL、受限密钥/支付相关授权路由、新平台或整份生成Ent不取。支付、checkout、充值、付费订阅购买/续费、退款仍为第四档，用户选Terra不改变边界。

### 5.3 旧轮次状态重述

此表以当前commit/完成记录为依据，不回写历史PORTING、源SHA、原checkbox或旧audit数量。直接相关前置本轮已重新量测；其余是状态复核，未声称重跑每个历史PR。

| 轮次 / 旧项 | 2026-10-28 当前状态与原因 |
|---|---|
| 0.2.18 第一/二档 | 6+10簇已在2b79ac320完成，Go1.27.3/既有协议、退役、断流和图片下载基座保留；本轮不重复立项。 |
| 0.2.18 用户个人功能 / 0.2.16同名遗留 | 0e38d06d8已完成TPS、筛选记忆、运维账号名、模型级冷却、OpenAI搜索及min_output_tokens，完整ops入口已恢复。运维分位数/趋势和其它新模型仍独立按需。 |
| 0.2.18 个人维护 / 最老SQLite后台项 | b554bea554已恢复清理/到期暂停/定时测试，修audit清空及Enter防重入，并完成#8249/#8250。后台已不再被skipSQLiteBackgroundJobs跳过，旧“尚未SQL核查”标签核销。 |
| 0.2.18 #8251 / #8269 | 共享解码快照缓存/COW需实际共享所有权与内存收益测量；SQLite轻量quota查询需benchmark；PG JSONB具体实现排除。 |
| 0.2.18 Gemini/Vertex按需 | #8078 Argon、#8114思考摘要、#8166 service_account quota范围、#8167多平台生图开关/上游246、显式cacheCRUD/Vertex API Key保持按需；不为S05/S02带入。 |
| 0.2.16主体 / 新功能 | 7c544c03f、847db09e1、5ff6e5bf9主体完成；Haiku5.5、group i18n及其它新能力按用途。模型级冷却/TPS已经由个人功能承接。 |
| 0.2.12 | 405246789主体及用户选定优先级/Key排序完成；TypeSafe缺平台，支付推广/查单排除，旧axios已由0.2.16维护。 |
| 0.2.11 | 96f32d018主体完成；Claude原生额度查询/重置、风控白名单、Key创建限制、在途预占、Astra Ultrafast仍需用途/设计。账号上游订阅metadata继续保留。 |
| 0.2.9 第一档 | 用户于2026-10-24明确确认完成，已归档；[completion-record](../../openspec/changes/archive/2026-10-24-port-upstream-0.2.9-tier1/completion-record.md)为完成依据，不沿用0.2.18历史表的未合标签，不推断新SHA。 |
| 0.2.9 第二档 | 159d04222完成48PR/16簇；工具Schema、PDF、目录、取消和用量路径是当前基座。 |
| 0.2.9按需/排除 | OpenCode/Command Code/Kimi原生、plugins、完整model_allowlist glob/pinned/live、Free Fast/effort倍率、视频新路由、TypeSafe及多实例/备份运营能力仍未整体采用；架构与实际用途分别判断。 |
| 0.2.5第二档收尾 | b349c7408已完成33/33、S12b配置限额/0值/软删不复活/SQLite227。WS默认1.0是独立决策，5.0数值补丁明确不采纳；不再列为未实施。 |
| 0.2.5调度 | #6960/#7049完整exec_scope/抢占/queue-wake未整体采用；需要并发语义与调度设计，非本轮cache复查前置。 |
| 0.2.5模型/平台 | #6575/#6890/#6974/#6769新Gemini/原生Codex图片/DeepSeek峰谷/Ollama异步重置；#6747/#7091/#6691/#7153平台/批量能力；#6989/#7057/#6968/#7011/#6945/#6988目录/compact/pinned仍按使用方式；支付#6986排除。 |
| 0.2.5其它 | Grok全流序号#7129、input stripper#7085、Gemini带内错误#6929、监控BaseURL#6917及管理大展示未整体采用；现有失败序号/ops恢复不等于这些整簇完成。 |
| 0.2.4 / 0.2.0主体 | a26b6b55d / c9d9bebe8完成主体；0.2.0尾项于5b3e1a716完成128/128，类型桩、真实浏览器与组合门禁均核销。旧DOMPurify已由0.2.16完成。 |
| 0.1.180 / 0.1.184价格与usage | 动态长上下文阶梯、完整effort/Fast/Free Fast/Ultrafast、usage compaction/requested effort schema、峰谷/重置锚点等仍独立按需；Terra只接入既有计价策略，不授权旧大机制。 |
| 0.1.180工具桥 / Grok / WS | 已选PDF、tool身份、HTTP回放/孤儿tool及composite闸门已完成；非法参数、namespace/deferred尾项未整体合。Grok冷却/Realtime/抢占按用途；跨节点Redis租约单节点架构不需要。旧#5925最小delay已随0.2.18 Codex容量消费者落入，整簇不因此完成。 |
| 更早主体 / 固定窗口 | 0.1.176/179/183/184已验收主体保留；128c5f7bf固定quota窗口推进和后续227语义不可被PG优化覆盖，Responses Lite/已摘通用usage字段不重复扩schema。 |

历史活动change现在均已完成但仍待归档；本轮新增三份是新的未实施提案，不把“待归档”算作老代码待办。

### 5.4 依赖遗留的当前口径

[本版本生产pnpm audit](./evidence-0.2.19/frontend-production-audit.txt)为17条advisory：2 high / 13 moderate / 2 low，high仍是xlsx原型污染与ReDoS；follow-redirects1.16.1已由0.2.18维护，不能重复列为本轮未修。yaml、uuid、mermaid、decode-uri-component、colord、katex等记录仍需按实际暴露面处理。本轮没有全量开发依赖audit，不能用生产17替代旧全量数量。

S06只处理GO-2026-6860及强制传递版本；其它Go findings与xlsx/开发工具链专项继续独立维护，旧例外不续期。MVS强制x/tools0.51是本次量测结果，不是撤销上次0.50最小前置的历史事实。

## 6. ALREADY、N/A与本地替代实现

- 文件反向检查没有ALREADY；**语义已满足**仍存在。#8323的“渠道按次图片档位保留token”当前真实成本探针5000输入/100输出、2张图已得0.3×2=0.6；源依赖不同，不能因它在大新功能PR中就重复修改。
- #8003本地不含日期切换，不是“没DeepSeek”；现有OpenAI-compatible官方base URL与Pro/Flash成本继续保留。
- #8287默认4次重试加首次=5次，本地固定5次总尝试已匹配；新增配置属于按需。
- #8373修改的上游函数名不在本地，但等价缓存消费者存在且有缺陷，因此按实际函数移植，不判NOBASE/N/A。
- #8366相似“model plaza”文件不代表同一行为。#8369有cached-content错误识别也不代表绑定/pool消费者存在。
- Terra上下文/effort、S05模型列表与S02CAS均使用本地已有metadata、custom list及事务边界；不要引入未选定大功能当基座。

## 7. 建议落地顺序

1. F01 Claude指纹、F02 Codex缓存先落；并安排S01/S02 Grok审计与目录数据完整性，符合三个主路径优先。
2. S06依赖维护和S03上传完整性作为单独可回归步骤；先固定模块图再做最终组合门禁。
3. 用户选定的T01 Terra独立落，覆盖与F02、S01共用的Codex/compat文件。源码落在同一文件不代表可以直接并行套补丁。
4. F03共享cache usage、F04图片计数、F05畸形工具终态，然后S04渠道搜索、S05既有Gemini目录；保持成功与中断用量去重。
5. F06/F07管理输入修复最后组合；完整unit/build、定向race、SQLite迁移227/配额/dedup、适用前端门禁及依赖复扫后，填写各change自己的verification。
6. 本轮没有发布、tag、push或部署计划；后续实施起点单独记录，不回写这次冻结来源。

## 8. 基线自测与实施门槛

| 本轮已执行 | 证据 | 结论及边界 |
|---|---|---|
| 23候选/276文件/3732产品哈希复查 | review-intake.py、review-baseline.json | 与原始来源/四态一致；未应用patch |
| backend测试实际目录与编译 | [review-test-list.txt](./evidence-0.2.19/review-test-list.txt) | 10包 go test -list通过；来源新用例不冒充本地现存测试 |
| backend既有定向回归 | [review-targeted.txt](./evidence-0.2.19/review-targeted.txt) | 10包通过；现存用例没有覆盖全部新边界 |
| 旧行为缺陷探针 | [review-baseline-probes.txt](./evidence-0.2.19/review-baseline-probes.txt) | 12个probe通过，包括1个已满足负对照；通过表示复现 |
| Grok追加审计探针 | [review-grok-audit-probes.txt](./evidence-0.2.19/review-grok-audit-probes.txt) | 2个probe覆盖4条JSON/SSE/缓冲路径，确认字段/冲突丢失 |
| 前端渠道/目录/编辑回归 | [review-frontend.txt](./evidence-0.2.19/review-frontend.txt) | 4文件40测试通过；既有pnpm/Browserslist警告保留 |
| Terra邻近后端/前端 | [后端](./evidence-0.2.19/review-terra-neighbors.txt)、[客户端配置](./evidence-0.2.19/review-terra-frontend.txt) | 验证现有GPT/effort/价卡/生成配置基座，非Terra实施验收 |
| x/net隔离PoC及MVS编译 | [MVS](./evidence-0.2.19/x-net-minimum-mvs.json)、[可行性](./evidence-0.2.19/review-dependency-feasibility.txt) | 当前与修复版复杂度有差异；隔离依赖可编译，未改repo modfile |
| 生产依赖audit | [audit](./evidence-0.2.19/frontend-production-audit.txt) | 17条未处理记录；不是全量开发依赖扫描 |
| 文档/OpenSpec/范围检查 | [文档检查](./evidence-0.2.19/document-validation.md) | 三提案链接、来源、分档、空实施槽及保护哈希另行核对 |

复跑命令由各review-*.json保存完整argv、cwd、时间和退出码。来源测试名字见 [source-test-inventory.tsv](./evidence-0.2.19/source-test-inventory.tsv)；新回归在实施时加入，必须先证明预期行为失败，再修复后通过。

建议实施门禁（backend目录，均通过rtk代理）：

    go test -tags=unit ./... -count=1
    go build -o /tmp/sub2api-0219/server ./cmd/server
    golangci-lint run ./...
    go test -race -tags=unit ./internal/service ./internal/handler ./internal/repository ./internal/pkg/apicompat -run 'Test.*(Grok|CodexModels|GPT61|Billing|Buffered|WebSearch|Images|Gemini)' -count=1
    go test -tags=unit ./internal/repository ./internal/service -run 'Test.*(UsageBilling.*SQLite|ProductionSQLUsesSQLiteDialect|Migration|PlatformQuota|FixedQuota|SplitSQL)' -count=1

前端有改动的S04/S05/T01需pnpm typecheck、lint:check、test:run、build及make test-frontend-critical；浏览器验收覆盖实际保存、模型列表和客户端配置，不把已通过的旧test当作新功能证据。依赖步骤需重跑适用audit/Go扫描，并保留未解决finding。

## 9. 本轮新增通用教训

- “同样的目录功能”可能有多个缓存演进：上游function名缺失不能直接判不适用，要找本地等价生产消费者及确定性交错。
- 新增CLEAN文件仍可含PG-only CAS，原子单键更新还需要与普通表单更新一起验证，SQLite单节点也有并发写入。
- 缺helper可能来自未选定的新授权PR。11行事务wrapper可以独立实现效果，不是导入9000行受限密钥功能的理由。
- JSON语法正确不代表模型目录形态有效；失败、空目录、乱序时间、来源失配与旧表单是不同边界。
- 上下文observer存在不代表usage结果带出；Grok必须验证返回结构、异步任务快照和SQLite账本。
- “拒绝截断”与“放大上传上限”可拆开；下载上限、单part上限、总body上限各自验证。
- 新模型的none/max、工具endpoint、metadata优先与价格显示需贯通；加一个下拉/资源条目不等于支持Terra。占位created值不能当发布日。
- 用户选择第三档时保留原档位及选择记录，独立立项；“需要模型”不等于授权其它目录、支付或整release。
- 依赖安全修复先量最小MVS图；PoC成功、package可达和匿名可利用性分别说明，历史audit与当前扫描分开。
- 旧任务核销以当前完成记录为准，原始source-baseline/TSV/验证快照保留；每轮不能重复把已恢复后台或已完成227列为遗留。

