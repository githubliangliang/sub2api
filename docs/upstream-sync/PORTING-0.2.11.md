# 上游 v0.2.11 移植评估（v0.2.9 → v0.2.11）

评估及实施日期：2026-09-30。**第一档4个PR/4簇、第二档7个PR/6簇开发及验收完成。**
改动位于 `sync/upstream-0.2.11` 分支，起点为 `f13dee245`。包括旧v0.2.9 F04两桥thinking禁用契约及695ebede7的14行通用usage字段前置。第三档6个PR、Astra Ultrafast及第四档仍未引入。

[第一档验收](../../openspec/changes/port-upstream-0.2.11-tier1/verification.md)、[第二档验收](../../openspec/changes/port-upstream-0.2.11-tier2/verification.md)记录逐场景证据：后端全量unit/build、定向race、SQLite方言/真实用量去重通过；前端1836项通过、2项既有跳过，typecheck/lint/build通过。VERSION仍1.1.14，最大迁移226，无依赖变更，已按用户要求完成本地Docker及远端部署并通过健康检查，未发布Release。

下文来源统计、文件态及patch sites保留评估基线；source-baseline和evidence-0.2.11冻结快照没有改写。

## 1. 版本、范围与规模

| 基线 | 值 |
|---|---|
| 本 fork | main，`f13dee245790eb2be058935d0b1aa6265da99474`；开始时工作树干净 |
| 本地版本 | VERSION `1.1.14`；上一版本提交 `047391137`，随后 CI 修正 `f13dee245` |
| 上次已评估上游 | v0.2.9，`4c00df2e0183e2c70b7fa8ba45914205e36aad0c` |
| v0.2.10 | `2f3fed2fdb0787141294cec81487a5df30426f7f`；2026-09-29 16:28:24 北京时间；22 条非 merge、11 个 PR、1 条直接提交 |
| v0.2.11 | `96f4c115c9749078f90cbf210a01d39baf3f53b6`；2026-09-30 15:06:51 北京时间；再增17条非 merge、6个PR、1条直接提交 |
| 抓取时 upstream/main | `42bc7f6cffe24bcb471608e48e66b4a0afa1f882`；本评估止于 v0.2.11，不纳入 tag 之后的提交 |
| 运行约束 | SQLite-only，最大迁移226；Redis可关闭/miniredis；simple mode与个人使用范围保留 |

来源：[v0.2.10 Release](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.10)、[v0.2.11 Release](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.11)。[发布说明快照](./evidence-0.2.11/releases.json)。
合计 **39条非merge、17个PR、2条版本号直接提交**；tag间净差 **183文件 / +8,879 / -602**。逐候选共230条文件态，包含重复文件。

| 分档 | 来源候选 | 簇数 | 处理 |
|---|---:|---:|---|
| 第一档 | 4 PR | 4 | 已完成；验收见tier1 |
| 第二档 | 7 PR | 6 | 已完成；混合PR按行为拆分，验收见tier2 |
| 第三档 | 6 PR | 5 | 用量展示、风控豁免、原生额度查询/重置、Key限额、余额预占，未立实施change |
| 第四档 | 2直接提交 | — | 上游VERSION同步，不覆盖fork自有版本 |

## 2. 方法与适配边界

- 执行 `git fetch upstream --tags --prune`，使用解引用后的tag commit SHA。无共同祖先，禁止整仓merge；`HEAD..upstream/main`不是缺失清单。
- 每个PR按 `git diff <merge>^1 <merge>` 取完整结果，保留两个主线直接提交。非merge枚举另存，避免微提交被重复排序。
- 独立临时 `GIT_INDEX_FILE` 读取 `f13dee245`，**全候选先反向** `git apply --cached --reverse --check`，再正向检查。均失败且本地文件不存在记NOBASE。
- 结果：**ALREADY 0 / CLEAN 154 / CONFLICT 58 / NOBASE 18**。CLEAN包括新文件和测试，不是运行证明；NOBASE也可能只是上游专用测试或缺失原生路径。
- [候选](./evidence-0.2.11/candidates.tsv)、[非merge提交](./evidence-0.2.11/commits.tsv)、[逐文件四态](./evidence-0.2.11/files.tsv)、[分档理由](./evidence-0.2.11/decisions.tsv)均固定在上述基线。
- [证据校验值](./evidence-0.2.11/MANIFEST.md)与[文档校验](./evidence-0.2.11/document-validation.md)记录快照和结构检查；新增调用检索中的缺失定义按本节与各簇范围处理。
- [42项关键符号](./evidence-0.2.11/base-symbols.tsv)与[新增调用检索](./evidence-0.2.11/added-call-inventory.tsv)核对定义。后者是词法检索，方法/闭包/外部包或类型转换单列，不能替代类型检查。上游新增helper须随消费者落地，不能据CLEAN整取测试。
- 新增调用中的本地缺失主要为本PR新增helper、S02依赖的 `anthropicReasoningEffort`、已排除的原生Anthropic消费者及Ultrafast目录helper。`configuredCodexServiceTiersForModel`依赖未引入Fast策略，不为模型升级整套搬入。
- **本区间没有迁移、go.mod/go.sum、package.json或pnpm-lock.yaml变更。** 旧226不修改；本次建议无需新迁移。未来其它schema工作从实施时最大号之后编号。
- 支付、充值、退款、购买/续费不引入。上游账号套餐标签和原生额度信息属于账号元数据；它们不因出现“订阅/兑换”字样被误判为支付。

## 3. 第一档：现有缺陷的小范围修复

A/C/X/N分别为ALREADY/CLEAN/CONFLICT/NOBASE，包含来源中应剔除的邻近测试。

| 簇 | 来源 | A/C/X/N | 本地缺口与行为 |
|---|---|---|---|
| F01 多工具改名保持请求体完整 | #7701 | 0/3/0/0 | 反复整包sjson改写造成长历史高分配；已改为收集span一次组装，JSON正确性及缓存原规则回归通过 |
| F02 白名单不覆盖显式模型映射 | #7542 | 0/7/2/0 | ModelWhitelistSelector未接收modelMappings；添加同名白名单前检查from→不同to并提示 |
| F03 Claude Code限制分组的兼容入口降级 | #7579 | 0/3/0/0 | Chat/Responses handler无条件403；已有FallbackGroupID时交现有调度解析降级分组 |
| F04 使用密钥弹窗隐藏不支持客户端 | #7678 | 0/2/1/0 | KeysView未传claude_code_only；受限分组只显示Claude客户端，切换分组重置标签 |

### F01 patch sites / 基座

`backend/internal/service/gateway_tool_rewrite.go` 的 `applyToolNameRewriteToBody`（178）由 `gateway_forward.go`（260）消费；`shouldMimicToolName`、`applyToolsLastCacheBreakpoint`均在位。新增span结构只用现有gjson和标准库。
覆盖tools、tool_choice和历史tool_use，验证转义字符、多个不同长度名称、内置工具不改及缓存断点保留。排除PR内无关 `openai_images_json_keepalive_test.go` 的sleep修正。

### F02 patch sites / 基座

`frontend/src/components/account/ModelWhitelistSelector.vue` 的 `addCustom` 与 CreateAccountModal/EditAccountModal/BulkEditAccountModal 的现有 `modelMappings`；补齐中英文提示。
这是本地既有凭据白名单编辑器，**不依赖**尚未采用的分组 `model_allowlist` glob/schema。自映射、空目标、不同名称可正常添加；已有显式映射不得被静默改成identity。

### F03 patch sites / 基座

`backend/internal/handler/gateway_handler_chat_completions.go`、`gateway_handler_responses.go` 的ClaudeCodeOnly闸门。
`gateway_scheduling.go` 的 `checkClaudeCodeRestriction`（1000）实际调用 `resolveGatewayGroup`（967），会跟随fallback并检测循环；不是只放宽外层if。
验收两个入口实际选中降级分组账号；无fallback仍403，失效/循环配置仍拒绝，不能仅断言不再403。

### F04 patch sites / 基座

`frontend/src/components/keys/UseKeyModal.vue` 的Props/defaultClientTab/clientTabs/watch，`frontend/src/views/user/KeysView.vue`传值。
本地group类型已有claude_code_only。只改变客户端指引；F03授权路径独立验收，不能让前端标签代替服务端限制。

规格：[port-upstream-0.2.11-tier1](../../openspec/changes/port-upstream-0.2.11-tier1/README.md)。

## 4. 第二档：值得移植，需适配与多路径验证

| 簇 | 来源 | A/C/X/N | 契约 |
|---|---|---|---|
| S01 Sonnet 5.5目录、签名与工具兼容 | #7683 | 0/31/6/3 | 目录/计价/effort一致；between_tools、签名历史、稳定toolset beta按映射后的模型处理 |
| S02 GPT-6.1 Sol与账号套餐标识 | #7730部分 | 0/18/16/4 | 模型目录/官方metadata/价格/请求校验一致，套餐SKU不因同显示名合并；排除Astra Ultrafast子项 |
| S03 Anthropic流式用量转发与归一化 | #7380 | 0/4/1/4 | 接收到的usage不因客户端未给include_usage而丢失；输出与计费桶一致，避免重复扣缓存 |
| S04 Antigravity首内容前保活 | #7643 | 0/2/0/0 | 15秒注释保活、2分钟首内容硬上限；提交响应后错误走SSE，不再换号重放 |
| S05 Composite WS别名与账号归属 | #7378 | 0/4/5/1 | 两种调度仅选拥有公开别名的账号；WS使用responses路由并保留公开模型身份 |
| S06 Codex远程目录与旧客户端回退 | #7680 + #7736 | 0/3/5/0 | 采用v0.2.11远程/文件终态；目录超过1MiB回退本地文件，URL与认证保持正确 |

### S01 patch sites / 基座

- `backend/internal/pkg/claude/{constants,effort_catalog}.go`：新增Sonnet判定，复用Opus 5.5已有规范化；`backend/internal/domain/constants.go`更新现有目录。
- `backend/internal/pkg/apicompat/{types,responses_to_anthropic_request,anthropic_to_responses_response}.go`：映射后的模型控制thinking；保留between_tools与签名响应。
- `backend/internal/service/gateway_request.go`：现有 `validateClaudeOpus55Request`扩为两模型校验；between_tools仅Sonnet支持，low/medium/high；拒绝disabled/enabled、强制工具、不支持的采样设置。签名历史一处无效时清理整段thinking，保留普通正文/工具。
- `gateway_{forward,count_tokens,claude_oauth_body,forward_as_chat_completions,forward_as_responses,upstream_request,anthropic_passthrough,bedrock}.go`、`bedrock_request.go`：helper与本地消费者一起适配；原生、OAuth、API Key、已有Bedrock/countTokens均回归。
- `billing_service.go`、`pricing_service.go`、内置价卡、Codex目录、前端模型白名单/状态/使用指引同步。来源Sonnet输入/输出为$2/$10每百万token，缓存写5m/1h为$2.5/$4，读$0.2；显式渠道价格优先。

本地已有Opus 5.5契约，不能整覆盖。`filterSonnet55ToolsetBeta`只在Sonnet稳定computer/browser toolset时去掉旧fine-grained token，并在账号header override之后再处理；其它beta保持。
排除本地缺失的 `openai_gateway_*_anthropic_native.go`，不为新模型引入原生CN平台。

### S02 patch sites / 基座

- `backend/internal/pkg/openai/constants.go`与新增 `codex_gpt61_sol.json`：IsGPT61SolModelSpelling/ValidateGPT61SolReasoningEffort随消费者加入；保留官方metadata未知字段及nested model_messages，同时公开slug不被覆盖。
- `backend/internal/pkg/apicompat/{anthropic_to_responses,chatcompletions_to_responses}.go`、`backend/internal/service/openai_{compat_model,model_alias,codex_transform,codex_models_service,codex_model_metadata,compat_prompt_cache_key,gateway_request_body,gateway_messages,gateway_chat_completions,gateway_chat_completions_raw,gateway_messages_chat_fallback,gateway_responses_chat_fallback}.go`：映射后、转换前校验none/minimal/disabled；保留max；仅Chat能力账号带工具明确报错，不能静默删工具。
- `billing_service.go`、`pricing_service.go`和价卡：普通输入/输出$2/$10每百万token、缓存写$2.5/读$0.1；保留本地priority与账号长上下文gate，不搬新版service_tier_billing接口。
- `frontend/src/utils/planType.ts`、`credentialsBuilder.ts`、`PlatformTypeBadge.vue`与模型配置指引：区分canonical SKU与显示标签，未知值保留。此处为上游账号元数据，不涉及购买/续费。

**前置已完成：0.2.9 F04/#7538两桥thinking禁用契约随本轮S02实施。** 评估时： `anthropicReasoningEffort`在本地缺失，上游helper为10行；先按旧F04规格完成两条桥的thinking禁用优先级，再复用helper实现新模型校验。不得仅加空helper或把旧F04整批记成完成。
本地 `configuredCodexServiceTiersForModel`和 `service_tier_billing.go`缺失，但已有priority目录构造；保留本地构造，不导入Ultrafast发送策略。
**#7730中的 `e6d191a83` Astra Ultrafast能力/6倍价格仍属第三档**，没有随模型/套餐部分提升。API支持与OAuth账号manifest授权不能由套餐标签推断。

### S03 patch sites / 基座

`backend/internal/service/gateway_forward_as_{chat_completions,responses}.go` 的流处理、 `mergeAnthropicUsage`；`backend/internal/pkg/apicompat/types.go` 的AnthropicUsage。
本地只有input/output/cache_read/cache_creation四字段。旧 `695ebede70e0bed4c8fd4c87b5a426448a08ea4c`（0.1.180第四档）中的**14行类型字段基座**可独立补入；结合v0.2.11终态merge/helper，不需原生供应商平台。
**这是明确的范围重判：只将通用类型字段与现有桥接消费者提升为第二档前置，旧CN整簇仍不做。** 不搬其缺失native消费者或其它透传路径。
有权威prompt total或hit/miss时拆互斥桶；只有后到cache字段不能反推扣减早前input。显式零usage与缺失/null区分；Chat仅转发实际接收usage，Responses终态与计费一致。继续保留本地取消/drain、终态结束和用量去重。

### S04 patch sites / 基座

`backend/internal/service/antigravity_gateway_compat_stream.go` 的session、select loop、finish/read-error；writer、adapter、scanner均在位。
虽然2文件CLEAN，**会改变响应提交时机，列第二档**。首ping后HTTP 200已提交；随后空流/超时/读取错误发送一次SSE错误，不再按“未输出”换号。comment/signature不算语义内容，2分钟硬上限不得被持续注释延长。需要真实httptest慢流、客户端断开和race验证。

### S05 patch sites / 基座

`backend/internal/{handler/openai_gateway_handler.go,handler/wire.go,service/openai_account_scheduler.go,service/openai_gateway_scheduling.go,service/openai_ws_forwarder_ingress.go}`；Wire改输入后重新生成 `backend/cmd/server/wire_gen.go`。
本地 `CompositeRouteResolver.Resolve`、WithCompositeRouteDecision、RequestedPublicModelFromContext、CompositeRouteSourceFromContext、explicitModelMappingClaims均存在；Gateway普通调度已检查ownership，OpenAI两条路径缺失。
WS首帧用responses/any路由，只允许既有OpenAI/Grok目标；原始公开名参与准入，映射值参与出站。后续同名/省略名可续聊，更换公开名要求重连；BeforeRequest使用rawForHash而非改写后的payload。不能带入未采用model_allowlist或插件DI。

### S06 patch sites / 基座

`frontend/src/api/codex.ts`、`frontend/src/components/keys/UseKeyModal.vue`及中英文dashboard文案。
本地manifest API、normalizeCodexBaseUrl、请求取消/ID隔离与TOML转义已在位。新增buildCodexModelCatalogUrl和responseBytes随消费者落地；终态将OpenAI HTTP/WS目录重新纳入远程/文件选择。
**#7680是被后续改写的中间态，不单列第一档“隐藏目录”。** 远程模式provider下写model_catalog_url；旧客户端文件模式只写model_catalog_json。URL无Key；认证沿用配置；UTF-8原始响应超过1MiB后提示并切文件，不能截断目录。保留旧S08已验证的/v1、子路径、Windows `~/`和CC Switch差异。缺失平台模板不恢复。

规格：[port-upstream-0.2.11-tier2](../../openspec/changes/port-upstream-0.2.11-tier2/README.md)。

## 5. 按需项目与历史遗留

### 5.1 本轮第三/四档

| 来源 | 判断依据与决定边界 |
|---|---|
| #7684 + #7726 Claude原生额度查询/重置 | 查询14文件+691/-13，兑换15文件+1314/-20；新service/handler/UI，后者依赖前者。属于上游账号额度管理，非支付；有实际Claude额度需求时再选，需确认失败/未知兑换结果不盲重试 |
| #7679 风控用户白名单 | 26文件+598/-45；新增豁免设置/执行语义，已有审计可用不等于需要豁免。按实际个人风控需求决定 |
| #7646 用量/费用趋势切换 | 15文件+118/-20，个人统计可用但属展示增强；不恢复已裁剪运营页 |
| #7752 API Key创建限额 | 12文件全CLEAN仍是新默认限制：200有效Key、60次/小时，0不限；Redis计数/接口替换需miniredis与并发验证。个人用途未确认必要性，不默认开启 |
| #7681 余额在途预占 | 21文件+2261/-7；service新文件714行，Lua缓存144行，handler132行。**上游自身在simple mode直接禁用**；普通余额模式才有收益。属于用量准入而非支付，不误排第四档，也不作为个人simple模式必修；以后选用需验证SQLite/miniredis及异步计费释放 |
| #7730内Astra Ultrafast | 旧Fast发送/能力策略未采用，不能仅靠套餐名宣告OAuth权限。保留第三档；模型与标签部分可独立移植 |
| 9a62841fd / a60a29549 | 第四档，上游VERSION同步。fork继续1.1.x |

本次只生成可审阅的第一/二档方案，未启动第三/四档实现；选择按需功能时再明确用途与范围。

### 5.2 上一轮与更早未合项

- **0.2.9第二档16簇已在159d04222落地**，v1.1.14已包含；旧PORTING与source-baseline是历史快照。本轮S01/S02/S03/S04/S05/S06分别与已落模型/流/路由/客户端配置相交，应基于当前代码适配。
- **0.2.9第一档原39PR/15簇：F04已作为本轮S02前置完成，其余仍待实施**：RPM投影、OAuth重授权、账号身份/错误状态、beta头、Responses消息/终态正文、Alpha search、未来reset暂停、TOTP、公共表单、代理小修、配置输入、异步弹窗、用量显示、ID/配额诊断。此前dialog ID counter和空Tab前置已随第二档S14落地；本次另外完成F04两桥契约，不能将旧第一档整批算完成。
- **0.2.9第三/四档**：OpenCode Go/Zen/Command Code/Kimi原生/插件、model_allowlist glob/pinned/live混合目录、Free Fast/effort倍率/自动用卡重置、视频/图像新路由、TypeSafe/自动版本/月度备份、运营与部署增强仍待用途决定；支付/返利排除。现有白名单编辑器不等于model_allowlist基座。
- **0.2.5主体已落bfbcd79**；#7043默认WS系数5.0仍未采用（现有默认1.0），#6954/S12b平台额度清理与预留227未实施；#6960/#7049完整exec_scope/抢占/queue-wake仍未整体引入。
- **0.2.5按需目录/平台**：#6575/#6890/#6974/#6769 Gemini3.7/3.8、原生Codex Images、DeepSeek峰谷价、Ollama异步重置；#6747/#7091/#6691/#7153平台/订阅批量/Key批量与过滤；#6989/#7057/#6968/#7011/#6945/#6988目录/上下文/视觉/compact/pinned，不因新模型升级自动完成。收费开关#6986不引入。
- **0.2.5差异与管理**：#7129/#7085/#6929/#6917 Grok序号、input stripper、Gemini带内错误与监控BaseURL继续独立适配；#6847/#7001/#6845/#6913/#6915管理展示待选。#7126/#7110/#7148已有namespace语义/上游撤回/PG等不重复移植。
- **0.2.0 / 0.1.180 / 0.1.184长上下文和Fast**：本地只有fast→priority价格别名、旧gate逻辑；动态阶梯/effort/Fast/免费Fast/Ultrafast完整策略未合。S02排除Ultrafast，不藉模型升级带入新价格体系。
- **0.1.180工具桥尾项与Grok/WS**：非法参数、namespace别名、deferred标记尚未整体合；F01是工具名请求体修复，不替代该旧簇。Grok冷却/Realtime/透传/抢占及Redis跨节点租约按用途/架构保留原取舍。
- **0.1.184其它**：usage compaction/requested effort迁移、公共分组限制、订阅重置锚点、峰谷价/图像工具冷却仍独立；不改已应用schema。0.1.180自动重置卡/模型读取上限未采用；本轮Claude原生重置不是同一子系统。
- **本轮重新量出的旧项**：0.1.180第四档695ebede7只有14行通用字段为S03所需，明确提升该子集；0.2.9 F04的10行helper是S02前置。其它“缺基座”不能从字段为零命中直接推导工程量或不适用。
- **依赖与后台**：DOMPurify仍^3.3.1、xlsx仍^0.18.5；本轮无lockfile变化、未运行新audit，不沿用历史advisory数量。UsageCleanupService、AccountExpiryService、ScheduledTestRunnerService仍受SQLite跳过条件控制，需要各自SQL审计后再启用。
- **已落旧轮次**：0.2.4、0.2.0主体、0.1.183/179/176不重新立项；API Key instructions、routed catalog、Responses Lite等按现有实现保留。

以上“仍待”是来源/当前文件复核，不声称本轮重新验证每个历史功能的运行行为。详细旧项理由保留在[上一轮§5](./PORTING-0.2.9.md#5-按需不做项目与旧轮次遗留)，当前复核见[历史探针](./evidence-0.2.11/leftover-probes.tsv)。

## 6. ALREADY、净终态与无需移植

没有文件反向应用成功，不代表所有缺陷都是新增。#7680与#7736合看才能看到净终态；旧0.2.9第二档以实际commit为准。
NOBASE中的原生Anthropic文件不纳入；源测试缺失时为现有消费者新增本地用例，不能伪称整PR通过。两个VERSION提交不移植。
本区间不需要把PostgreSQL、外部Redis或支付系统恢复进fork。

## 7. 建议实施顺序

1. 使用当前本地基线开同步分支；保留SQLite/miniredis/simple mode和现有更改。先F01/F02/F03，再F04。
2. 先完成旧0.2.9 F04所需的两桥契约，再S01 Sonnet、S02 GPT；共享模型/价格文件串行落地。
3. S03先补14行类型字段，再以v0.2.11终态对齐归一化/流输出；新回归先证明失败，不能整取缺失原生平台测试。
4. S04独立验证“首ping提交→错误不重放”；S05独立验证两种调度和真实多轮WS，Wire重新生成。
5. S06基于F04与新模型后的UseKeyModal终态适配；同时验证旧S08 URL/Windows/CC Switch契约。
6. 完成unit/build、相关race、SQLite真实库和前端typecheck/Vitest；汇总验收后才进入自有1.1.x发布流程。

## 8. 基线自测与实施门禁

[基线报告](./evidence-0.2.11/baseline-tests.md)记录本次命令与结果。Go在backend目录自动选用1.27.0；前端依赖已安装。
后端相关service/handler/repository定向unit、三个模型/转换package全量unit、`go build ./...`通过；包含SQLite方言与现有真实SQLite复合平台检查。
前端typecheck和4文件/97项Vitest通过。`go test -list`确认引用存在；[来源测试构建标签](./evidence-0.2.11/source-test-inventory.tsv)区分缺失文件和integration标签。

上述是评估时的基线结果，不是新增行为验收。当时两份change的checkbox未勾、实施证据空白；现已完成开发，最终unit/race及逐场景证据见两档verification，不用基线结果代替验收。
S04/S05要求race与真实HTTP/WS取消/续聊；S01/S02要求默认/显式模型、价格覆盖与边界；S03要求显式零/缺失usage、权威总量与独立桶、终态计费一致。SQLite未改schema仍需验证成功用量一次落库。前端需真实组件测试，不能仅typecheck。

## 9. 新增通用结论

1. **同一发布区间也会撤换设计**：#7680的隐藏目录不是最终行为，#7736重新提供目录来源选择；先看目标tag再排名。
2. **CLEAN不能代表低风险**：#7643仅2文件干净，却会提前提交HTTP 200并关闭换号机会，必须按生命周期改动验收。
3. **按单节点实际启用条件估收益**：#7681源码明确在simple模式禁用；不能只读“默认开启”就当本fork必修。
4. **混合PR按行为拆**：#7730同时含模型、套餐SKU和Ultrafast。账号元数据不是支付，元数据也不是能力授权。
5. **旧第四档也要量可复用子集**：695ebede7整簇平台不适用不妨碍14行通用字段服务现有桥；明确记录提升范围与冻结SHA。

## 10. 实施补充（2026-09-30）

- F01的原始上游Spans测试混入了未采用的deferred工具缓存规则；该失败不算本轮缺陷。其余正确性用例在旧实现也通过。实际先失败后通过的是1000条工具历史分配量：旧137,056,945 bytes/op，单次span组装通过4,505,100上限；缓存断点仍按本fork既有规则。
- F03使用真实HTTP上游跑通两个兼容协议入口，除选择fallback账号外，验证200及正文；缺失/循环fallback不触达上游。
- S02独立补齐GPT-6.1显式零缓存写价在priority的优先级，未引入全套Fast引擎。官方目录未知字段、nested model_messages、公开slug和账号显式覆写均有实际消费者回归。
- S05的真实WS测试覆盖OpenAI/Grok、responses/any、两种连接模式、多轮模型/身份及重连；未恢复缺失的model_allowlist或native路径。
- race发现两条既有Responses测试在并行阶段写Gin全局模式，调整测试初始化顺序后通过。初始失败与最终通过日志分别保留，未把测试装配错误算作产品修复。
- Codex目录按#7736最终remote/file设计完成，并补齐UTF-8阈值、旧Windows路径和中文/英文说明。

本地Docker更新后，已追加按用户要求执行远端部署脚本；远端用户级systemd服务与外网健康检查通过，运行二进制哈希已核对，配置未变。详见[远端部署证据](../../openspec/changes/port-upstream-0.2.11-tier2/evidence/remote-deployment.txt)。
