# 上游 v0.2.24 移植评估（v0.2.23 → v0.2.24）

评估日期 **2026-11-21**。GitHub官方 Releases API与fetch后的tag核对：最新正式版为[v0.2.24](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.24)，发布于 **2026-11-20 09:19:50 UTC / 17:19:50 上海时间**。冻结时刻见[baseline.json](./evidence-0.2.24/baseline.json)；网页工具未返回内容、gh未安装，使用官方API原文和git，不把搜索失败视为无新版本。

**建议第一档10簇、第二档15簇，共选30个PR及1个安全合并簇的既有消费者维护。评估完成，两批未实施。** 第一优先是重复字段造成模型/策略/计费分叉，再处理Codex搜索403、OAuth不可逆轮换和Claude/Codex/Grok协议与结算。第一档/第二档表示成本风险分档，不覆盖安全严重性；复杂的安全修复S01应先做。

- [第一档提案：10簇、37项未勾选任务](../../openspec/changes/port-upstream-0.2.24-tier1/README.md)
- [第二档提案：15簇、52项未勾选任务](../../openspec/changes/port-upstream-0.2.24-tier2/README.md)
- [来源、失败过程、探针与完整核验](./evidence-0.2.24/MANIFEST.md)

本轮新增评估、原始证据和提案，只更新两处入口文档。没有应用产品补丁、改VERSION/依赖/schema/Ent/Wire、提交/推送/发布/归档或业务实例操作。

## 1. 版本、范围与规模

| 项目 | 冻结值 |
|---|---|
| fork HEAD / tree | `d2488b1105f091b6fce4445395fec1835306fcf8` / `17b1b71f61137510e42feeaa3fcf2a1109cb4ce6` |
| VERSION / Go / SQLite / auth snapshot | `1.1.26 / 1.27.4 / 230 / 21` |
| 私有 index tree | `17b1b71f61137510e42feeaa3fcf2a1109cb4ce6` |
| v0.2.23 commit / tag object | `5af9afb588c4b014b3df9db7d4073af1f8ae01b6` / `ce19a640eab5c4b5ef1827c7c1fcfce49569411a` |
| v0.2.24 commit / tag object | `5c2982a81b46e2142d0143d2f83c1ce9c4da1d24` / `7515db6dc56c2a6120dfb7abed70ef089ebc1457` |
| upstream/main（观察） | `328db26f450edbdf9f8d4344a3ea903ac6396842` |
| 非merge / 完整候选 | 65 / 39 PR + 1普通直接提交 + 1无PR合并簇 |
| 净diff / 候选×文件态 | 190文件 +11853 -610 / 217条 |
| ALREADY / CLEAN / CONFLICT / NOBASE | 0 / 135 / 62 / 20 |

启动时工作区干净。当前main已包含v0.2.23两档、GPT-6.2、透传模型过滤与requested/effective effort（SQLite230）；不能沿用旧评估的7a0b78fe5/SQLite229。39个PR必须按整PR first-parent→merge取diff；最终安全合并`5c2982a81`的6个非merge也作为完整簇。65个非merge无遗漏、无重复覆盖。217是候选×文件数，区别于190个净变化文件。

## 2. 方法、最小前置与迁移编号

1. fetch upstream tags/prune，冻结API发布时间、tag对象、源码commit、main及fork。main在tag后只有`328db26f4`的VERSION同步，单独记录，不从main补产品。
2. 所有217条先统一reverse、再统一forward，使用同一私有index且仅`--check`；未真正apply。结果[files.tsv](./evidence-0.2.24/files.tsv)、[完整候选与补丁哈希](./evidence-0.2.24/candidates.tsv)可复核。
3. [2269条新增调用扫描](./evidence-0.2.24/added-call-inventory.tsv)和[本地基座复核](./evidence-0.2.24/local-base-symbols.txt)只辅助定位。CLEAN不证明编译、NOBASE不证明无缺陷、ALREADY=0不等于没有手工等价实现。逐簇最小前置和本地名称差异见下文。
4. 核查85份来源测试文件以及76份本地integration首行，其中53份为postgres专用；另运行九包实际test-list和相关基线。探针采用虚拟test-only Go overlay；前端临时测试结束即删除。来源扫描器只在/tmp隔离编译，未挂产品链。
5. 本地最大迁移230，tag范围内上游迁移与依赖manifest/lock均无变。本轮两档没有新增列需求；OAuth CAS须用SQLite事务/outbox，用户默认配额复用现有表。S11实施时应真实生成Wire；不照抄上游PG或引入shadow/新平台字段。若用户将来另选带schema能力，应从实际基线重新分配231起，不能重复230。
6. S01最小前置须明确：共享字段scanner/中间件本地不存在，但实际路由和parser分叉可复现；旧hasUniqueJSONMembers只保护特定上下文。可新增共享拒绝检查而不引入完整group model_allowlist。来源标题提到的group allowlist绕过不能被记成本地已有该功能。
7. 保留SQLite-only/miniredis/simple、固定quota、usage dedup、用户所选Fast/Ultra Fast和GPT-6.2、64MiB请求体预算/40MiB SSE行/8MiB暂存/120s idle。没有读取真实业务账号来假定新功能使用情况。

### 全部候选分档

| 候选 | 非merge | A/C/X/N | 判断与选取范围 |
|---|---:|---|---|
| [2af3f1e15](https://github.com/Wei-Shaw/sub2api/commit/2af3f1e15c98c893a89255a037f551046883b8c0) | 1 | 0/0/1/0 | 第四档。上游VERSION同步；fork保持1.1.26，不照搬0.2.23。 |
| [#8982](https://github.com/Wei-Shaw/sub2api/pull/8982) | 1 | 0/0/2/0 | 第一档 F01：Codex 文献搜索 403 的账号副作用 |
| [#8671](https://github.com/Wei-Shaw/sub2api/pull/8671) | 1 | 0/5/2/0 | 第二档 S02：OpenAI 不可逆 OAuth 轮换的 SQLite 保存 |
| [#8794](https://github.com/Wei-Shaw/sub2api/pull/8794) | 1 | 0/4/0/0 | 第三档。compaction请求5m缓存TTL是新的缓存策略；本轮缓存完整性修复无需它，按实际压缩使用选择。 |
| [#8801](https://github.com/Wei-Shaw/sub2api/pull/8801) | 1 | 0/9/0/0 | 第三档。Exa新增搜索服务/凭据/设置；已有搜索修复不需要新的供应商。 |
| [#8803](https://github.com/Wei-Shaw/sub2api/pull/8803) | 2 | 0/5/1/0 | 第三档。web_search来源块是新增响应能力；源含非string URL保护但本地新来源块消费者不存在，不借修复导入。 |
| [#8804](https://github.com/Wei-Shaw/sub2api/pull/8804) | 2 | 0/4/3/1 | 第三档。ToolSearch→原生tool_search新增端到端能力，271行转换器及234行服务测试；不是已有普通工具的必要前置。 |
| [#8820](https://github.com/Wei-Shaw/sub2api/pull/8820) | 7 | 0/7/2/0 | 第二档 S03：缓冲 Responses 的 hosted 输出重建 |
| [#8845](https://github.com/Wei-Shaw/sub2api/pull/8845) | 1 | 0/4/0/0 | 第二档 S13：已有热点的有界窄优化 |
| [#8858](https://github.com/Wei-Shaw/sub2api/pull/8858) | 3 | 0/1/7/0 | 第二档 S07：Claude/Gemini 用量快照刷新 |
| [#8859](https://github.com/Wei-Shaw/sub2api/pull/8859) | 1 | 0/5/1/0 | 第二档 S10：Composite 改写后的 Claude thinking |
| [#8862](https://github.com/Wei-Shaw/sub2api/pull/8862) | 1 | 0/1/0/1 | 第二档 S13：已有热点的有界窄优化 |
| [#8864](https://github.com/Wei-Shaw/sub2api/pull/8864) | 1 | 0/6/0/0 | 第二档 S04：Grok 成功搜索次数结算 |
| [#8866](https://github.com/Wei-Shaw/sub2api/pull/8866) | 3 | 0/5/0/1 | 第二档 S13：已有热点的有界窄优化 |
| [#8868](https://github.com/Wei-Shaw/sub2api/pull/8868) | 4 | 0/2/3/0 | 第二档 S13：已有热点的有界窄优化 |
| [#8870](https://github.com/Wei-Shaw/sub2api/pull/8870) | 2 | 0/2/1/0 | 第二档 S13：已有热点的有界窄优化 |
| [#8872](https://github.com/Wei-Shaw/sub2api/pull/8872) | 1 | 0/6/1/0 | 第二档 S13：已有热点的有界窄优化 |
| [#8875](https://github.com/Wei-Shaw/sub2api/pull/8875) | 1 | 0/9/1/0 | 第二档 S12：异步用量与审计的请求体保活 |
| [#8883](https://github.com/Wei-Shaw/sub2api/pull/8883) | 1 | 0/3/2/0 | 第二档 S09：Anthropic metadata 长上下文目录 |
| [#8893](https://github.com/Wei-Shaw/sub2api/pull/8893) | 1 | 0/3/1/0 | 第二档 S05：OAuth 外族模型与 Messages 映射后调度 |
| [#8899](https://github.com/Wei-Shaw/sub2api/pull/8899) | 1 | 0/1/1/0 | 第二档 S06：批量透传编辑独立修改模型过滤 |
| [#8900](https://github.com/Wei-Shaw/sub2api/pull/8900) | 1 | 0/2/2/0 | 第三档。Claude窗口费用估算新增展示与5h/7d缓存策略；不是使用记录或配额正确性的必要修复。 |
| [#8902](https://github.com/Wei-Shaw/sub2api/pull/8902) | 1 | 0/4/0/0 | 第三档。手动Token刷新频率风险提示是新增交互策略；OAuth持久化修复不依赖它。 |
| [#8911](https://github.com/Wei-Shaw/sub2api/pull/8911) | 1 | 0/4/0/0 | 第一档 F05：既有 Claude 账号测试 temperature |
| [#8918](https://github.com/Wei-Shaw/sub2api/pull/8918) | 1 | 0/3/6/2 | 第二档 S11：后台新建用户的默认平台配额 |
| [#8923](https://github.com/Wei-Shaw/sub2api/pull/8923) | 1 | 0/3/0/0 | 第一档 F08：Gemini 最后工具轮的尾随内容 |
| [#8929](https://github.com/Wei-Shaw/sub2api/pull/8929) | 1 | 0/0/2/0 | 第一档 F03：Codex IDE 与独立上下文 |
| [#8931](https://github.com/Wei-Shaw/sub2api/pull/8931) | 1 | 0/2/4/0 | 第一档 F04：Responses 每轮请求推理强度 |
| [#8932](https://github.com/Wei-Shaw/sub2api/pull/8932) | 1 | 0/1/3/0 | 第二档 S08：Codex 原生音频输入声明 |
| [#8937](https://github.com/Wei-Shaw/sub2api/pull/8937) | 1 | 0/1/1/0 | 第二档 S14：SQLite 模型平均速率按耗时加权 |
| [#8943](https://github.com/Wei-Shaw/sub2api/pull/8943) | 1 | 0/2/0/1 | 第四档。DeepSeek Anthropic-native thinking passback的具体消费者openai_gateway_messages_anthropic_native.go本地缺失；普通映射不等于该原生路径。 |
| [#8946](https://github.com/Wei-Shaw/sub2api/pull/8946) | 1 | 0/1/1/0 | 第一档 F09：Antigravity 原生 SSE 注释帧 |
| [#8962](https://github.com/Wei-Shaw/sub2api/pull/8962) | 1 | 0/5/0/0 | 第一档 F10：Grok 媒体输入错误的 400 分类。只选F10现有3图输入错误分类；3→5图扩展仍为第三档，来源说明不等于用户选择。 |
| [#8963](https://github.com/Wei-Shaw/sub2api/pull/8963) | 1 | 0/5/0/1 | 第一档 F06：兼容转换保留 stop 与原始工具定义 |
| [#8968](https://github.com/Wei-Shaw/sub2api/pull/8968) | 1 | 0/3/1/0 | 第一档 F02：OpenAI 408 上游超时换号 |
| [#8972](https://github.com/Wei-Shaw/sub2api/pull/8972) | 1 | 0/0/0/4 | 第四档。TypeSafe/SystemOne平台及两个核心消费者本地缺失，不为模型映射修复引入平台。 |
| [#8973](https://github.com/Wei-Shaw/sub2api/pull/8973) | 4 | 0/7/0/0 | 第二档 S15：Antigravity 裸 429 范围及额度端点 |
| [#8976](https://github.com/Wei-Shaw/sub2api/pull/8976) | 1 | 0/0/4/0 | 第三档。模型广场复制ID是新增便捷交互，按用途选择。 |
| [#8979](https://github.com/Wei-Shaw/sub2api/pull/8979) | 1 | 0/0/2/0 | 第二档 S15：Antigravity 裸 429 范围及额度端点 |
| [#8886](https://github.com/Wei-Shaw/sub2api/pull/8886) | 1 | 0/1/0/1 | 第一档 F07：图片 SSE usage 单行载荷 |
| [5c2982a81](https://github.com/Wei-Shaw/sub2api/commit/5c2982a81b46e2142d0143d2f83c1ce9c4da1d24) | 6 | 0/9/7/8 | 第二档 S01：所有现有入口的重复字段拒绝与读取一致性 |

## 3. 第一档：窄修复

### F01 Codex 文献搜索 403 的账号副作用（[#8982](https://github.com/Wei-Shaw/sub2api/pull/8982)）

- **依据**：基线 predicate 将所有403送入账号惩罚；隔离目标断言失败，401负对照通过。
- **最小前置**：ForwardAlphaSearch 与现有访问状态错误识别器均存在；只扩展 predicate 的正文参数及两个调用点。
- **patch sites**：backend/internal/service/openai_alpha_search.go：ForwardAlphaSearch、forwardAlphaSearchViaResponsesWebSearch、shouldApplyOpenAIAlphaSearchAccountErrorSideEffects。
- **行为**：alpha/search 工具级403 SHALL 允许既有换号流程且不冷却或永久置错健康账号；明确账号/工作区停用的403继续执行现有账号策略。
- **验收**：OAuth与PAT回退的 literature-search 403、空403、deactivated_workspace、401/404/405和429；记录仓储副作用次数。
- **边界**：不能把所有403全局忽略；仅限定搜索工具路径，不放宽普通Responses鉴权与访问状态错误。

### F02 OpenAI 408 上游超时换号（[#8968](https://github.com/Wei-Shaw/sub2api/pull/8968)）

- **依据**：shouldFailoverUpstreamError(408) 为false；408目标失败，400/429对照通过。
- **最小前置**：现有统一 failover 分类函数和HTTP/图片调用链完整，只增加408分支。
- **patch sites**：backend/internal/service/openai_gateway_upstream_errors.go：shouldFailoverUpstreamError；原生响应与图片消费者。
- **行为**：上游 HTTP408 SHALL 进入既有有界换号流程。
- **验收**：Responses/图片408后下一账号成功、换号上限、父取消、普通400和已输出后不重放。
- **边界**：不新增重试策略；保留一次idle换号预算及已输出边界。

### F03 Codex IDE 与独立上下文（[#8929](https://github.com/Wei-Shaw/sub2api/pull/8929)）

- **依据**：ide_context、空untrusted_content、空白proposed_plan未转换；旧browser_info控制通过。
- **最小前置**：本地同等消费者名为normalizeCodexPageContext；不能照抄来源normalizeCodexUntrustedContext名字。
- **patch sites**：backend/internal/handler/openai_page_context.go：normalizeCodexPageContext及validCodexUntrustedContext。
- **行为**：合法独立 agent 上下文 SHALL 转为不可信 user 文本，保留各类型的空值规则。
- **验收**：IDE非空、untrusted_content允许空串、proposed_plan允许空白非空、null/数组拒绝、真实call_id保持、重复键保护。
- **边界**：不提升到system/developer；未知工具、真实调用及page快照严格校验保持。

### F04 Responses 每轮请求推理强度（[#8931](https://github.com/Wei-Shaw/sub2api/pull/8931)）

- **依据**：configuration_update=high覆盖顶层medium的目标失败；省略、none控制通过。本地已持久化requested/effective且WS从原始轮帧取值。
- **最小前置**：CanonicalRequestedReasoningEffort已存在；只补input中最后一个有效configuration_update读取，不引入来源整套map API。
- **patch sites**：backend/internal/service/openai_gateway_request_body.go：CanonicalRequestedReasoningEffort；openai_ws_forwarder_ingress.go、openai_ws_v2_passthrough_adapter.go现有原帧采集点。
- **行为**：请求推理强度日志 SHALL 优先记录本轮原始input最后一个有效configuration_update的effort，并与最终有效强度独立。
- **验收**：HTTP/Chat/Messages/WS桥接与原生WS，多个update、空/无effort、JSON转义、none/max、映射/封顶、下一轮省略与SQLite230读写。
- **边界**：保留d2488b110的原始载荷取值、output_config支持、缺省NULL及none/max；不重复加入迁移230或撤销本地已修的bridge采集。

### F05 既有 Claude 账号测试 temperature（[#8911](https://github.com/Wei-Shaw/sub2api/pull/8911)）

- **依据**：createTestPayload(claude-opus-4-8)仍含temperature，目标失败；sonnet-4-6控制通过。
- **最小前置**：现有Claude effort family判定完整；可新增21行家族predicate并复用，排除未选Fable/Mythos分支。
- **patch sites**：backend/internal/pkg/claude/sampling.go（拟新增最小predicate）；backend/internal/service/account_test_service.go：createTestPayload与Bedrock测试载荷。
- **行为**：既有拒绝temperature的Claude家族在连接测试中 SHALL 省略该字段。
- **验收**：Opus4.7/4.8/5.5及别名/Bedrock前缀，Sonnet4.6与显式测试模型对照；实际本地HTTP抓取请求。
- **边界**：不扩充模型目录/价卡；未选Fable/Mythos helper不作为前置。

### F06 兼容转换保留 stop 与原始工具定义（[#8963](https://github.com/Wei-Shaw/sub2api/pull/8963)）

- **依据**：ResponsesRequest解码后stop字段消失，转Chat断言失败；无stop控制通过。
- **最小前置**：ResponsesRequest/Chat模型及桥接已存在。原生Anthropic专用消费者文件缺失，不能照搬该hunk。
- **patch sites**：backend/internal/pkg/apicompat/types.go、chatcompletions_responses_bridge.go；backend/internal/service/openai_gateway_chat_completions.go：Responses-shape转Chat字段保留。
- **行为**：既有兼容降级路径 SHALL 保留客户端stop序列与可转发的原始动态工具定义。
- **验收**：stop字符串/数组/省略；Chat→Responses→Chat、原生Responses/OAuth的字段白名单、API Key的动态工具、用户数值和工具顺序。
- **边界**：只选已有通用API Key消费者；不引入原生Kimi平台或缺失的Anthropic-native适配器。

### F07 图片 SSE usage 单行载荷（[#8886](https://github.com/Wei-Shaw/sub2api/pull/8886)）

- **依据**：pretty-printed usage使buildOpenAIImagesStreamCompletedPayload含换行；目标失败且JSON仍合法。
- **最小前置**：现有completed载荷builder完整；新增json.Compact及局部buffer即可。
- **patch sites**：backend/internal/service/openai_images_responses.go：buildOpenAIImagesStreamCompletedPayload。
- **行为**：转发图片的completed事件 SHALL 把usage压为单行JSON后写入SSE data行。
- **验收**：多行/紧凑/空/非法usage，实际TCP SSE事件解析、图像结果与精确token值，断开后结算去重。
- **边界**：不改变图片模型、价格、结果数量或SSE行上限。

### F08 Gemini 最后工具轮的尾随内容（[#8923](https://github.com/Wei-Shaw/sub2api/pull/8923)）

- **依据**：tool_result紧接user文本仍落在同一个最后turn，隔离目标期望两轮而实际一轮。
- **最小前置**：convertClaudeMessagesToGeminiContents完整存在；尾部约28行重排可独立采用。
- **patch sites**：backend/internal/service/gemini_messages_compat_service.go：convertClaudeMessagesToGeminiContents末尾。
- **行为**：最后一个Gemini工具结果turn的尾随非functionResponse内容 SHALL 保序拆成后续user turn。
- **验收**：单/多个工具结果加文本/图片、纯工具结果、纯文本、历史非末尾turn与parallel工具结果媒体顺序。
- **边界**：不导入新Gemini模型、完整thinking策略或改变已完成的工具媒体顺序规则。

### F09 Antigravity 原生 SSE 注释帧（[#8946](https://github.com/Wei-Shaw/sub2api/pull/8946)）

- **依据**：静态核对现有原生Gemini流分支丢弃所有空分隔行；来源修复维护非data帧分隔，已有SDK拒绝注释保护可复用。
- **最小前置**：downstreamRejectsSSEComments与handleGeminiStreamingResponse都存在；局部separator状态适配。
- **patch sites**：backend/internal/service/antigravity_gateway_streaming.go：handleGeminiStreamingResponse。
- **行为**：支持注释的客户端 SHALL 保留上游SSE注释及其分隔；不支持注释的既有SDK继续滤掉注释。
- **验收**：注释/data混排、CRLF、空行、重写data、keepalive、go-genai/python-genai及中断usage。
- **边界**：不通过关闭默认ping回避错误；保持120s idle、40MiB行上限和计费drain。

### F10 Grok 媒体输入错误的 400 分类（[#8962](https://github.com/Wei-Shaw/sub2api/pull/8962)）

- **依据**：本地超出既有3图限制返回普通fmt.Errorf，handler没有输入错误分支；来源typed error可独立于5图能力摘取。
- **最小前置**：prepareGrokMediaForwardBody、normalizeGrokMediaJSONImageRefs和handler均存在；新增小型错误类型及errors.As分支。
- **patch sites**：backend/internal/service/grok_media.go：两处输入校验；backend/internal/handler/grok_media.go：handleGrokMedia错误分类。
- **行为**：既有Grok媒体本地输入校验失败 SHALL 返回400并保留可用校验信息。
- **验收**：JSON/multipart超出当前3图上限、合法输入、内容策略、脱敏及上游可配置错误策略。
- **边界**：本簇保持grokMediaMaxEditSourceImages=3；扩为5张是独立第三档，未获选择不随修复进入。


## 4. 第二档：需适配、SQLite或前端的修复

### S01 所有现有入口的重复字段拒绝与读取一致性（[5c2982a81](https://github.com/Wei-Shaw/sub2api/commit/5c2982a81b46e2142d0143d2f83c1ce9c4da1d24)）

- **依据**：实际Gin路由对重复model/tier/嵌套effort未提前400（到达未装配handler并500）；parser观察到gjson首模型与json末模型分叉。来源扫描器85例通过。
- **最小前置**：本地没有requestmodel共享包、GroupModelAllowlist或完整group allowlist系统；缺共享包不等于缺漏洞。拟采用472行扫描器+约95行入口适配和最小内容识别helper；已有hasUniqueJSONMembers仅用于局部上下文净化。
- **patch sites**：backend/internal/pkg/requestmodel/conflict.go（拟新增）与最小BOM/内容识别helper；backend/internal/server/middleware、server/routes/gateway.go；handler/openai_gateway_handler.go；service/openai_ws_forwarder_ingress.go、openai_ws_v2_passthrough_adapter.go；pkg/httputil/body_read_error.go及现有协议错误writer。
- **行为**：所有现有HTTP网关请求与每轮WS载荷 SHALL 在模型改写、调度和计费之前拒绝语义重复字段，并让内容识别、body重读与协议错误格式一致。
- **验收**：等值/异值/大小写/转义/Unicode键、嵌套/数组/freeform参数、BOM与伪multipart、preamble、标量重复/扩展name与多图合法列表；首帧/后续帧/session.update；无转发/无usage；实际HTTP/WS和1C1G并发预算。
- **边界**：不引入完整group model_allowlist/glob/live、新平台或PG。只挂本地已存在路由；现有账号模型过滤、64MiB预算/40MiB SSE/8MiB暂存/120s idle不变。相同重复也必须拒绝，不能只拒不同值。

### S02 OpenAI 不可逆 OAuth 轮换的 SQLite 保存（[#8671](https://github.com/Wei-Shaw/sub2api/pull/8671)）

- **依据**：模拟刷新返回新access/refresh token但enrichment截止后，本地返回deadline并未保存；目标失败。现有Grok CAS与晚到不保存控制已通过。
- **最小前置**：snapshotOAuthRefreshAccount、Grok success CAS、事务/outbox及本地SQLite JSON持久化基础已存在。源SQL是PG jsonb/CTE，必须改成SQLite事务CAS；不引入credential shadow实体。
- **patch sites**：backend/internal/service/oauth_refresh_api.go、token_refresh_service.go；backend/internal/repository/account_repo.go：既有Grok CAS邻近OpenAI实现与outbox。
- **行为**：已完成的OpenAI轮换 SHALL 在有界脱离取消的上下文保存；CAS失败须读取当前持久化凭据且保留并发重新授权。
- **验收**：请求/后台/旧fallback路径，enrichment超时、重新授权/代理/非token编辑、CAS冲突/写失败/重读失败、缓存失效/outbox、锁租期覆盖保存清理、实际SQLite/miniredis与race。
- **边界**：只对完整OpenAI轮换成功启用例外；Grok/Claude晚到规则、固定quota与取消边界保持。无新列，无旧token重试，不能使用普通Update覆盖并发授权。

### S03 缓冲 Responses 的 hosted 输出重建（[#8820](https://github.com/Wei-Shaw/sub2api/pull/8820)）

- **依据**：web_search_call item.done不能从本地accumulator重建，目标失败；当前已有文本/函数/图片accumulator是可复用基座。
- **最小前置**：来源7个提交需作为完整PR：474行重建helper及必要RawMessage字段；保留本地stream/terminal优先级和Grok server-search过滤。
- **patch sites**：backend/internal/pkg/apicompat/types.go、responses_to_anthropic.go、responses_to_chatcompletions.go；backend/internal/service/openai_gateway_messages.go、openai_gateway_response_handling.go及拟新增buffered helper。
- **行为**：缓冲兼容响应 SHALL 从最终有效终态及已完成item恢复完整hosted内容，输出顺序、工具身份和流式语义一致。
- **验收**：空terminal.output、web/file/computer/code/search、custom工具、函数object input、独立content index、partial/done/terminal冲突、Grok server search不得伪装tool_use、真实HTTP/SSE。
- **边界**：不以本簇启用未选ToolSearch新能力；不吞终态failed、不重复计费、不改变首内容failover/有界spool。

### S04 Grok 成功搜索次数结算（[#8864](https://github.com/Wei-Shaw/sub2api/pull/8864)）

- **依据**：ResponsesUsage解码后丢失server_side_tool_usage.web_search，目标失败；已有搜索事件去重计数完整。
- **最小前置**：apicompat.ResponsesUsage与三条OpenAI兼容用量路径已存在；补显式成功数及presence，保留现有fallback。
- **patch sites**：backend/internal/pkg/apicompat/types.go；backend/internal/service/grok_search_count.go、openai_gateway_chat_completions.go、openai_gateway_messages.go、openai_gateway_response_handling.go。
- **行为**：Grok搜索结算 SHALL 在上游提供成功次数时以其为准，缺省时沿用已有事件计数与去重。
- **验收**：web_search/x_search、显式0/正数/缺省/非法值、item.done+completed、HTTP/SSE/中断、用户与账号独立成本、SQLite settlement/dedup。
- **边界**：不得把失败搜索按item数重复收费；普通函数/客户端工具不进入服务端搜索计费。

### S05 OAuth 外族模型与 Messages 映射后调度（[#8893](https://github.com/Wei-Shaw/sub2api/pull/8893)）

- **依据**：普通/透传OAuth仍支持claude-*，两个目标失败；API Key和自定义别名控制通过。
- **最小前置**：isOpenAIOAuthServableModel、IsModelSupported、现有passthrough filter及channel mapping均存在；需沿Messages的实际调度模型调整。
- **patch sites**：backend/internal/service/account.go、openai_model_mapping.go；backend/internal/handler/openai_gateway_handler.go：Messages渠道映射后routingModel。
- **行为**：固定Codex上游的OAuth类账号 SHALL 跳过明确外族模型，且调度判断使用渠道映射后的实际模型。
- **验收**：OAuth/setup-token/API Key、普通/透传、model_filter开关/白名单、claude别名→GPT、GPT别名→外族、未知别名、HTTP/WS候选池。
- **边界**：保留用户已选#8748，不把API Key自定义base_url限制为OpenAI模型，也不误拦合法映射别名。

### S06 批量透传编辑独立修改模型过滤（[#8899](https://github.com/Wei-Shaw/sub2api/pull/8899)）

- **依据**：临时Vitest两例均发送model_filter=false；无关priority编辑控制通过。本地新功能已选定，不能再标为未选前置。
- **最小前置**：单账号model_filter和bulk字段均已存在，新增独立修改checkbox；当前测试断言false本身需随行为更新。
- **patch sites**：frontend/src/components/account/BulkEditAccountModal.vue及其既有测试。
- **行为**：批量编辑 SHALL 只提交用户明确选择修改的透传模型过滤字段。
- **验收**：OAuth/API Key只改透传、只改过滤、显式false、白名单独立修改、全选筛选目标、重新打开/目标变化、保存重开及移动端。
- **边界**：未勾选修改不等于false；不覆盖既有账号的过滤或凭据。

### S07 Claude/Gemini 用量快照刷新（[#8858](https://github.com/Wei-Shaw/sub2api/pull/8858)）

- **依据**：本地刷新key只覆盖OpenAI，AccountUsageCell及列表缓存未绑定Claude/Gemini快照变化；69项当前测试基线通过但不覆盖新行为。
- **最小前置**：现有buildOpenAIUsageRefreshKey、批量队列/请求token和页面生命周期保护可复用；来源Grok serializer不是必须导入的基座。
- **patch sites**：frontend/src/utils/accountUsageRefresh.ts；components/account/AccountUsageCell.vue；views/admin/AccountsView.vue。
- **行为**：Claude/Gemini快照变化 SHALL 使现有用量缓存失效并刷新，异步旧结果不得覆盖新快照。
- **验收**：5h/7d读数/窗口滚动、相同id更新、非可见行、批量队列、旧请求晚到、离页取消、错误保留旧读数、桌面/移动浏览器。
- **边界**：保留本地已实施的不闪行与OpenAI unknown新鲜度语义；不引入费用估算/强查按钮或来源原生国内平台。

### S08 Codex 原生音频输入声明（[#8932](https://github.com/Wei-Shaw/sub2api/pull/8932)）

- **依据**：normalizeCodexInputModalities删除audio，目标失败；text/image控制通过。
- **最小前置**：现有账号模型metadata、manifest、raw catalog和协议能力判定完整；需保守交集，不能仅改静态目录。
- **patch sites**：backend/internal/service/upstream_models.go：normalizeCodexInputModalities；openai_codex_model_metadata.go；openai_codex_models_service.go：manifest/raw补全。
- **行为**：已有原生Responses音频输入能力 SHALL 在模型声明中正确保留，聚合目录只声明全部可调度候选支持的能力。
- **验收**：GPT6.2 Sol OAuth、API Key明确metadata、映射别名、混合账号交集、raw目录/缺metadata、仅Chat协议剔除音频与ETag稳定性。
- **边界**：不引入Voice/Realtime/新的音频计费或模型；静态订阅目录不能替代API Key实际能力。

### S09 Anthropic metadata 长上下文目录（[#8883](https://github.com/Wei-Shaw/sub2api/pull/8883)）

- **依据**：静态核对默认Anthropic registry host与动态模型列表缺扩展；[1m]解析和GetUpstreamModelMetadata已存在。
- **最小前置**：最小来源47行扩展helper及默认host修正可使用本地账号路由/metadata；完整group allowlist系统缺失而且不是本消费者前置。
- **patch sites**：backend/internal/service/upstream_models.go：upstreamModelRegistryBaseURL/matchModelsDevProviderByKnownHost；拟新增anthropic_models_list.go；handler/gateway_handler.go：Models。
- **行为**：现有Anthropic目录 SHALL 按可服务账号的有效metadata展示长上下文变体，并正确识别默认官方registry host。
- **验收**：ContextWindow阈值900000、别名/显式mapping/路由账号范围、[1m]去重、大小写、自定义base_url、metadata缺失/失败与真实models HTTP。
- **边界**：不凭目录增加新模型/价格，不导入未选group model_allowlist；现有[1m]请求规范化与计费保留。

### S10 Composite 改写后的 Claude thinking（[#8859](https://github.com/Wei-Shaw/sub2api/pull/8859)）

- **依据**：实际Composite middleware把别名改成Sonnet5.5后仍保留enabled；隔离目标应adaptive而失败。
- **最小前置**：本地已有claude.IsSonnet55/IsOpus55、Composite resolver、Bedrock sanitize；来源isClaude55SignedThinkingModel名字缺失不代表无消费者，可用现有家族判定适配27行helper。
- **patch sites**：backend/internal/server/routes/gateway.go：compositeTargetPlatformMiddleware；service/composite_platform.go；bedrock_request.go现有thinking处理。
- **行为**：Composite将请求改写到既有Claude5.5时 SHALL 同步适配该型号接受的thinking模式并保留其它参数。
- **验收**：enabled→adaptive并移除budget、Sonnet disabled→between_tools、Opus/普通Claude、nested model path、Bedrock及实际HTTP请求。
- **边界**：不引入新型号或整套signed-thinking框架；保持用户原模型名的使用记录和最终映射策略。

### S11 后台新建用户的默认平台配额（[#8918](https://github.com/Wei-Shaw/sub2api/pull/8918)）

- **依据**：本地AuthService已快照默认配额，adminServiceImpl.CreateUser只分配默认订阅；SQLite配额控制基线通过。
- **最小前置**：UserPlatformQuotaRepository/BulkInsertInitial、DefaultPlatformQuotaSetting、SQLite227语义已存在；抽出约34行共享helper并给admin注入repo。
- **patch sites**：backend/internal/service/admin_user.go、admin_service.go、auth_service.go及共享defaults helper；backend/cmd/server/wire.go与实际生成wire_gen.go。
- **行为**：后台创建普通用户 SHALL 按当前默认平台配额初始化，显式0与未配置无限额保持区别。
- **验收**：admin创建/注册两路径、普通用户/管理员、0/正数/缺省、写失败fail-open、已有事务隔离、BulkInsertInitial不覆盖已有值、实际SQLite及Wire生成。
- **边界**：不导入支付、购买或新订阅策略；不改变固定窗口重置、已删Key结算与历史用户配额，不新增迁移。

### S12 异步用量与审计的请求体保活（[#8875](https://github.com/Wei-Shaw/sub2api/pull/8875)）

- **依据**：1MiB实际newOpenAIRequestView标量地址仍落在原body内，目标失败；usage UserAgent/IP字段还直接引用整个input。
- **最小前置**：现有零拷贝view和异步审计生命周期均存在；持久标量克隆及局部变量地址可独立修复。
- **patch sites**：backend/internal/service/openai_gateway_request_body.go：newOpenAIRequestView；openai_gateway_usage.go：RecordUsage；openai_gateway_chat_completions.go、openai_gateway_messages.go；securityaudit/prompt_snapshot.go等实际snapshot出口。
- **行为**：离开请求生命周期的用量和审计数据 SHALL 拥有必要标量的独立存储，且不因字段地址或闭包滞留整个请求体。
- **验收**：字段字节一致、alias地址控制、body释放/取消/队列积压、审计快照归属、延迟worker与heap/alloc测量，64MiB预算释放后资源边界。
- **边界**：不全面复制所有body，不引入完整COW；对请求内零拷贝保留所有权，不把函数PoC宣传成生产内存上限。

### S13 已有热点的有界窄优化（[#8845](https://github.com/Wei-Shaw/sub2api/pull/8845)、[#8862](https://github.com/Wei-Shaw/sub2api/pull/8862)、[#8866](https://github.com/Wei-Shaw/sub2api/pull/8866)、[#8868](https://github.com/Wei-Shaw/sub2api/pull/8868)、[#8870](https://github.com/Wei-Shaw/sub2api/pull/8870)、[#8872](https://github.com/Wei-Shaw/sub2api/pull/8872)）

- **依据**：实际消费者均存在；源码核对namespace本地已用parseRawJSONView，但仍有重建拷贝。scheduler投影PoC记录payload；其余hunk实施前仍须逐项测量。
- **最小前置**：filterSchedulerAccountGroups、raw view、现有schema parser/regexp谓词与byte-edit helper均可复用；不导入来源完整大文件或假定CLEAN可编译。
- **patch sites**：backend/internal/repository/scheduler_cache.go：marshalSchedulerCacheAccount；service/gateway_claude_oauth_body.go：TTL次序；image_generation_intent.go；gateway_request.go；openai_responses_namespace.go；openai_responses_tool_schema.go及实际handler调用点。
- **行为**：采用的热点优化 SHALL 保持字节、策略和所有权语义，并通过单P前后成本与转义必要条件验证。
- **验收**：每个hunk独立benchmark/alloc、JSON转义键和值/fuzz corpus、重复字段策略与S01组合、非法UTF8/JSON、body改写后range刷新、schema freeform不误改、快照成员及调用者图不变。
- **边界**：可验证body完全相同才复用validated入口；不降低64MiB预算、不采用1GiB动态最低预算、stale grace/账务lane/完整COW；未达收益或等价性的hunk不落地。

### S14 SQLite 模型平均速率按耗时加权（[#8937](https://github.com/Wei-Shaw/sub2api/pull/8937)）

- **依据**：实际modernc SQLite完整迁移后两条100tokens/1s与100tokens/10s返回55，累计耗时口径目标18.18失败。
- **最小前置**：opsRepository.GetOpenAITokenStats是现有SQLite消费者；只移植聚合公式和NULL/零分母处理，不能照抄PG cast/FILTER SQL。
- **patch sites**：backend/internal/repository/ops_repo_openai_token_stats.go：GetOpenAITokenStats的avg_tokens_per_sec表达式。
- **行为**：模型统计平均Tokens/秒 SHALL 使用有效请求output_tokens总和除以duration_ms总和。
- **验收**：不等耗时、无效/NULL/零token、空集、平台/分组/时间/分页与实际ops HTTP；验证输出TPS分布现有均值与分位数不变。
- **边界**：与用户已选整窗口输出TPS统计口径不同：后者保留逐请求算术mean/P5/P10/P50，不能整站替换为加权值。

### S15 Antigravity 裸 429 范围及额度端点（[#8973](https://github.com/Wei-Shaw/sub2api/pull/8973)、[#8979](https://github.com/Wei-Shaw/sub2api/pull/8979)）

- **依据**：本地retry及外层handleUpstreamError无请求级bare429声明，可能重复冷却；quota fetch仍全局prod→daily而forward依账号选择daily。
- **最小前置**：现有gin请求context、错误分类、resolveAntigravityForwardBaseURL和有界HTTP client完整；来源67行request claim可适配。FetchAvailableModelsFromURL/QuotaSummary本地缺失，只补现有fetch的显式端点参数，不新增quota-summary。
- **patch sites**：backend/internal/service/antigravity_gateway_retry.go及Claude/Gemini/compat外层；antigravity_quota_fetcher.go；backend/internal/pkg/antigravity/client.go。
- **行为**：同一请求的无明确信号裸429 SHALL 最多冷却一个账号；既有额度查询 SHALL 使用该账号实际转发端点。
- **验收**：多账号同请求/独立请求、带details/reset的真实429、retry退出后外层、daily/prod/显式配置、代理/重定向/有界读取、取消和当前quota读数。
- **边界**：不放宽明确限流或模型容量规则，不新增查询按钮/QuotaSummary平台/强查策略；保留本地端点安全与已实施的model cooldown。

## 5. 按需、排除及旧遗留

### 5.1 本轮第三/第四档

第三档为7个PR加#8962的5图扩展：compaction 5m缓存策略、Exa、搜索来源块、ToolSearch原生桥接、Claude窗口费用估算、刷新风险提示、模型ID复制。它们均保留为按需选择；不因服务于Claude/Codex而自动升档。#8804需271行转换器及新增端到端消费者，其普通工具兼容前置并非现有修复所需。

#8962必须拆开：已有3图上限报400属于F10；3→5图是扩展能力，不默默随行。#8794缓存策略也不是本地5m/1h计量完整性修复的前置。#8803 URL防御只服务新来源块，缺消费者时不单独制造缺陷。

第四档含VERSION同步、TypeSafe/SystemOne模型映射（两个核心消费者缺失）及DeepSeek Anthropic-native thinking passback（特定native文件缺失）。不能把支持普通API Key映射误当这些原生平台存在。个人支付、退款、充值/购买/续费及其依赖迁移始终排除；本tag候选中未发现支付专属改动，也不做无关支付代码删除。

### 5.2 历史遗留重新定界

11份旧活动change共350个已完成checkbox、0个未完成，等待归档；本轮保持全部原文。以下是当前状态，历史报告的“未实施/按需”只保留其当时时点，不回写。

| 旧范围 | 2026-11-21复核状态与保留原因 |
|---|---|
| v0.2.23 两档 | d8d126da9/3652ddb6a已完成70/76任务；Go1.27.4和缓存/换号/metadata修复在基线，不重开。 |
| GPT-6.2 Sol/Astra | ad0967903已完成12项；冻结价格/时钟验收保持原证据。Pro套餐放行、分组强制Ultra Fast、渠道Ultra Fast倍率仍未选，本次音频目录维护不是引入这些策略。 |
| #8748 / #6188 | d2488b110完成透传过滤与requested/effective日志12项，SQLite230。S06/F04是其后续缺口，不能再次列为未实施功能或重复迁移。 |
| v0.2.22–23 按需尾项 | #8644/#8779 Fable读数/冷却局部分支可被自定义映射到达，目录/价卡仍未选；Carbon/Hydrogen/Qwen、Vertex us/eu、ping=0按实际用途。30m/12h独立桶仍需schema，未选。 |
| v0.2.22–23 性能尾项 | #8816账务lane/PGbatch、#8833合并加载/last_used/20s stale grace仍需SQLite事务/dedup/内存测量；本轮只量窄函数，未把完整架构升档。#8693 PG rollup缺本地消费者。 |
| v0.2.21 第一/第二档 | 63c764704、83729b130已实现，39/39和42/42，发布v1.1.25；SQLite229覆盖索引/auth21和资源默认保留。 |
| v0.2.20 第一/第二档 | 29681fdf5、a25ce7bed已实现，44/44及38/38，已发布v1.1.24；活动change待归档，不重开旧任务。 |
| Fast/Free Fast/Ultrafast | 326513cdc完成，SQLite228及完整HTTP/Chat/Messages/Responses/WS；v0.2.21对照确认免费输入仍保留已有输出Fast倍率。 |
| v0.2.20 MCP/admin Key、邮箱黑名单 | #8389/#8504新管理授权/注册策略按需；#8476依赖未选#8319受限admin Key，修补不构成引入授权，支付操作排除。 |
| v0.2.20图片/导出/配额策略 | #8462尺寸路由、#8557 Image2.6、#8543/#8521冷却配置、#8522 quota-summary、#8380/#8495导出默认策略仍按需。S09不需要新查询按钮，F10不需要新原生模型。 |
| v0.2.20性能与缺平台项 | #8518完整ojson/COW/compact/orphan仍需收益及所有权测量；必要资源治理与RPM空值缓存已实现。#8503原生Kimi、#8549 transcript阻断算法、Voice/cache CRUD缺消费者，仍不取；显式session ID策略保留。 |
| v0.2.20发布后观察 | 原记录#8610/#8606已进入v0.2.21 tag并重新完整分档，不继续标记为tag之外；其他main观察仅记录本轮tag之后提交。 |
| v0.2.19第一/第二档与Terra | 683354b2c、cccf931ef、a541d0b61完成并归档；Terra在使用范围，v0.2.21 S06维护其none/max和映射。 |
| v0.2.19 Grok诊断/历史兼容 | becd2d2a3完成目录诊断、参数/别名/defer/序号尾项；v0.2.21 F02是delta重建的后续缺口，terminal原文保留已通过，不能说之前整个修复无效。 |
| v0.2.19其它 | #8319受限admin、#8323真实图片张数、#8348 50MiB上传、#8287同号重试配置、#8395更广目录仍按实际用途；上传完整性/fileData/声明目录已完成。 |
| TPS/运维 | f333aa5f3完成整窗口P5/P10/P50/均值/样本数；usage TPS和完整ops入口已完成。新趋势图等按需。 |
| v0.2.18主体/个人能力 | 2b79ac320、0e38d06d8、b554bea55完成两档、日志/筛选/冷却/搜索/min tokens与SQLite后台恢复；1.1.23另修日志清空/写入顺序。不可重新关闭SQLite后台。 |
| v0.2.18性能/平台 | #8251完整共享解码COW、#8269 SQLite轻量quota查询仍需收益测量；Argon/思考摘要/SA quota范围/多平台生图开关/cache CRUD/Vertex API Key仍按需。v0.2.21 F08仅4行bool前置，不把完整Gemini thinking体系带入。 |
| v0.2.16 | 7c544c03f、847db09e1、5ff6e5bf9主体完成；Haiku5.5/group i18n按需；模型冷却/TPS由后续完成。 |
| v0.2.12 / v0.2.11 | 405246789、96f32d018主体及快调/排序完成；TypeSafe缺平台，支付推广/查单排除。Claude原生重置管理、风控白名单、Key限制、在途预占按需；Ultrafast已完成。 |
| v0.2.9 | 第一档用户确认并归档、第二档159d04222；OpenCode/Command Code/Kimi原生、plugins、完整model_allowlist glob/pinned/live、视频路由、多实例备份仍未整体采用。 |
| v0.2.5 quota与WS | b349c7408完成33/33和SQLite227；默认WS倍率1.0为已批准选择，不采用5.0。#6960/#7049 exec_scope、抢占/queue-wake仍需单独并发设计，S01/S02不需要它们。 |
| v0.2.5模型/平台 | #6575/#6890/#6974/#6769 Gemini新能力/原生图片/峰谷/Ollama重置；#6747/#7091/#6691/#7153新平台/批量；#6989/#7057/#6968/#7011/#6945/#6988目录/compact/pinned仍按需。#6986支付排除。 |
| v0.2.5其它协议 | #7129 Grok全流序号、#7085 input stripper、#6929 Gemini带内错误、#6917监控BaseURL及大管理展示未整体采用；新窄修复不代表整簇已完成。 |
| v0.2.4/v0.2.0 | a26b6b55d/c9d9bebe8主体；5b3e1a716完成0.2.0尾项128/128（类型桩、四面浏览器、组合门禁）。 |
| v0.1.180/v0.1.184计价/usage | 5250907f5模型参考价及Fast已完成；动态长上下文阶梯、独立effort倍率、usage compaction、峰谷/重置锚点仍按需或需schema；requested effort已由d2488b110/SQLite230完成。新TTL是另一个独立schema选择。 |
| v0.1.180工具/Grok/WS及更早 | PDF/tool身份/HTTP回放/孤儿工具/composite闸门及历史四尾项已完成；Grok429/Realtime/抢占按需，跨节点Redis租约单节点不需要；#5925最小delay随容量消费者完成但非整簇。0.1.176/179/183/184主体和Responses Lite已验收；128c5f7bf固定窗口与227/228语义保留。 |

旧#8610的动态预算最低1GiB仍不适合1GiB实例，本地保留已测64MiB；#8590强查按钮、#8367浮标、Argon完整quota、新图片/cache CRUD、admin Key/完整COW/抢占等继续按用途。#8269 SQLite轻量quota查询未取得足以替代现状的新收益证据；不能继承“缺基座”字样，现有有界读数路径已复核但仍需实际负载测量。跨节点Redis租约在单节点架构不需要。

### 5.3 依赖、版本与发布后main

tag范围没有go.mod/go.sum、前端package/lock和migrations变更；本轮不升级工具链或依赖，也没有跑新的全量漏洞审计。上一轮12 moderate/2 low是历史时点结果，不能当成本轮当前审计。main只有一个tag后的VERSION同步，证据见[post-release-commits.txt](./evidence-0.2.24/post-release-commits.txt)。

## 6. 已具备与缺基座的精确结论

- #8931的每轮原始帧/bridge取值，本地d2488b110已比来源旧快照前进；仅input configuration_update仍有缺口，不整段替换回来源实现。
- #8899依赖的#8748现已实施，应作为现有功能正确性修复；#8870本地也已用raw view，不能声称整份input拷贝优化全是新增收益。
- 本地hasUniqueJSONMembers是限定场景、精确名字检查；S01还需要所有入口、大小写/转义/freeform/multipart和后续WS帧。
- #8859缺来源helper名字而现有Claude5.5家族判定存在；实际Composite改写复现失败，最小约27行helper值得第二档，不判N/A。
- #8918复用SQLite227配额与现有BulkInsertInitial，#8671复用Grok CAS/outbox边界；前者要Wire，后者不能抄PG JSONB。
- #8979本地没有新QuotaSummary接口；只给现有有界fetch补固定端点，不借它引入#8522。#8883也不要求完整group allowlist。

## 7. 建议实施顺序

1. **先S01**，核对HTTP全部别名与WS首/后续帧、freeform兼容和内存成本，确保拒绝发生在调度/计费前。复杂度属于第二档，严重性最高。
2. F01/F02与S02/S05：搜索403、超时换号、不可逆token保存、映射后选号；确认整号/模型作用域和重试上限。
3. F03/F04/F06/F07、S03/S04/S08：Codex上下文/工具与缓冲、每轮effort、图片SSE、Grok成功数和目录能力，串实际HTTP/WS/SQLite结算。
4. S06/S07与F05/F10：编辑不覆盖和已有账号操作正确性；真实浏览器保存重开和旧请求晚到。
5. S09/S10/S11/S14：metadata/Composite/新用户quota/SQLite模型统计，各自独立公式和事务边界。
6. F08/F09/S15及S12/S13：既有Gemini/Antigravity稳定性和内存热点。S13每hunk等价+收益通过才采用；最后两档组合验收，再记录真实实施SHA。发布、推送、归档或部署另行决定。

## 8. 本轮自检与证据局限

| 检查 | 实际结果 |
|---|---|
| 当前基线 | 九包相关检查：869个顶层测试通过、1原跳过；包含子例为1541通过/1跳过。九包test-list实际执行。 |
| 前端基线 | 3文件69测试通过；日志保留现有预期错误输出与Browserslist提示。 |
| 五包初组行为探针 | 编译后13个顶层目标失败、10控制/观察通过；包括实际SQLite、Gin路由与字段解码，未模拟真实业务。 |
| 追加两包探针 | 4个顶层目标失败：标量body alias、Gemini末工具轮、buffered hosted item、实际Composite thinking。 |
| 前端新行为探针 | 透传批量编辑2目标失败、1无关编辑控制通过；临时文件已删除。 |
| 来源scanner | 冻结472行scanner+原test+单独提取multipart helper，隔离85例通过；单P约95399ns/op、1192B/op、14allocs，约225KB固定夹具。 |
| scheduler投影PoC | 合成40组payload 260839→5060字节，成员/调用者图/模型判断控制通过；不代表生产收益。 |
| 文档/提案 | 严格OpenSpec、候选覆盖/链接/未勾任务/空验收槽与最终产品/历史哈希见document-validation.json。 |

首次Go探针含一个括号拼写错误、图片builder参数与SQLite指针断言夹具不匹配；初始源码/log和修正源码都保留。它们不算产品缺陷或修复验收。后两轮行为探针已全部编译、无panic。共17个目标顶层失败不是完整影响面数量，也不是已修好；F09/F10等部分簇仅有静态可达证据，实施时仍需对应复现。

source scanner单P数值是当前工作站的函数成本，不是1C1G实例峰值、真实HTTP/WS安全接入验收或生产性能承诺。未跑完整backend/frontend门禁、race、browser或部署；这些留作实施验收。保护8311个初始tracked文件，允许改的旧文件仅AGENTS.md及入口README，旧来源/patch/验收/hash均保持。

## 9. 本轮通用教训

- 安全修复可能在无PR的最终合并内；必须枚举first-parent和全部非merge覆盖，不能只看Release列出的PR。
- 本地缺新的共享入口不等于没有解析器分叉；已有局部防重检查也不等于全入口安全。
- 相同值重复也会在模型/策略改写后分叉；扫描应区分协议字段大小写规则与用户自由键空间。
- token轮换不可逆完成与可选enrichment截止是不同边界；CAS保存、重新授权胜出和租期预算应一起验证。
- 已选第三档能力进入基线后，后续缺口应按当前消费者重新分档；不可重复做schema或覆盖更完整本地实现。
- 速率的累计耗时加权与逐请求算术均值各有消费者；修一个指标不能重写另一个用户已选指标。
- 优化要先量现有实现：raw view可能已存在，真正差异是编辑拷贝、异步标量所有权或展开的组图；CLEAN不是收益或可编译证据。
- 来源带来的新接口（group allowlist、QuotaSummary、shadow）要与实际消费修复拆开；用可测的最小helper替代整套新能力。
