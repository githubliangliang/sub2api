# 上游选定个人功能补充评估与规格（v0.2.15–v0.2.18）

**实施状态补记（2026-10-26）：六项已实现并完成开发验收，改动保留在未提交工作树。** 用户在规格后授权实现，并于2026-10-25补充选择恢复完整运维页。实际命令、结果和完成文件哈希见[verification](../../openspec/changes/port-upstream-0.2.18-personal-features/verification.md)及[§10](#10-实施完成补记2026-10-26)。以下§1–§9保留2026-10-24规格评估时点；其中“仅规格/未来验收/未实施”均为该历史阶段状态，不代表当前实现进度。冻结来源和原始证据不改写。

日期：2026-10-24。**本次仅编写规格，未实施产品功能。** 用户明确选择：输出TPS、账号筛选记忆、运维详情账号名称、模型级冷却规则、OpenAI API Key Web Search模拟、最小输出Token配置。

这些项目在原0.2.16/0.2.18评估中属于第三档按需能力；本次建立独立规格，不改称既有缺陷的第一档/第二档，不改写旧冻结报告。用户选择范围仅扩展这六项，其他按需能力和支付边界保持原决定。

入口：[OpenSpec变更](../../openspec/changes/port-upstream-0.2.18-personal-features/README.md)。

## 1. 版本与范围

| 基线 | 值 |
|---|---|
| fork产品基线 | `c36a19b7a1a132e690eaf9ce0d878b301ffffcb4`，VERSION1.1.19；该提交是前一批OpenSpec归档 |
| 来源上限 | 官方v0.2.18 / `ac1b8f1271be0d674fa6f074403c1410f30a3a81` |
| 来源范围 | 6个完整PR；首次进入v0.2.15、v0.2.16、v0.2.17、v0.2.18 |
| 本地边界 | SQLite-only；可选miniredis；simple mode；Go1.27.3；最大迁移226 |
| 工作树 | 用户已有继续归档/索引/主规格修改，未替换或清理；本次相关产品文件无未提交差异 |

| 编号 | 首次版本 / 来源 | 完整PR文件数 | A/C/X/N | 本次选定行为 |
|---|---|---|---|---|
| F01 | v0.2.15 / [#7911](https://github.com/Wei-Shaw/sub2api/pull/7911) | 17 | 0/16/1/0 | 只取使用日志输出TPS，不取同PR运维分位数 |
| F02 | v0.2.18 / [#8142](https://github.com/Wei-Shaw/sub2api/pull/8142) | 1 | 0/1/0/0 | 平台和状态筛选记忆 |
| F03 | v0.2.16 / [#7933](https://github.com/Wei-Shaw/sub2api/pull/7933) | 5 | 0/5/0/0 | 请求详情账号名及ID回退 |
| F04 | v0.2.16 / [#7828](https://github.com/Wei-Shaw/sub2api/pull/7828) | 17 | 0/12/5/0 | 显式规则scope及认证/模型边界 |
| F05 | v0.2.17 / [#8080](https://github.com/Wei-Shaw/sub2api/pull/8080) | 6 | 0/4/2/0 | OpenAI API Key Messages搜索模拟与必要配置 |
| F06 | v0.2.17 / [#8094](https://github.com/Wei-Shaw/sub2api/pull/8094) | 25 | 0/15/8/2 | 已有转换的输出下限配置 |

全体71条候选文件证据：**ALREADY0 / CLEAN53 / CONFLICT16 / NOBASE2**。这些是完整PR的文件态，包含本次不取的运维SQL和缺失平台消费者，不等于计划修改71个文件。多PR涉及同文件不能累计为独立文件数。

来源SHA/first-parent/完整diff哈希见[候选](./evidence-0.2.18-personal-features/candidates.tsv)与[冻结source-baseline](../../openspec/changes/port-upstream-0.2.18-personal-features/source-baseline.md)。

## 2. 方法与适配边界

本次是在已完成release intake上选择六项，不另开新release范围或拉取未来main；沿用本地已抓取的tag并解引用annotated tag。每个来源取完整merge父1到merge的diff，冻结SHA与hash。

独立临时GIT_INDEX_FILE读取固定fork SHA，先对全体候选完成`git apply --cached --check --reverse`，再完成全体正向check；仅检查，不向工作树或真实index应用补丁。原文件在本地缺失记NOBASE；新增文件允许正常前向检查。[files.tsv](./evidence-0.2.18-personal-features/files.tsv)保留双向退出码和原诊断。

新增Go行的570条调用词法记录逐符号对固定基线func声明检索，含标准库、类型转换、SQL和注释误命中；同名声明不证明接线或编译。[调用清单](./evidence-0.2.18-personal-features/added-call-inventory.tsv)和[人工调用复核](./evidence-0.2.18-personal-features/call-review.md)将现有消费者、新helper及排除基座分开。

六项预计均可复用现有JSON/模型限流/使用量字段，无新schema/迁移、Ent/Wire生成、依赖升级或版本变更。不得改226及以前SQL；如果实施测出新schema必要，先修订设计并按届时实际最大号分配。

## 3. 第一档/第二档与按需授权

本补充不新增第一档/第二档修复清单。六项原为第三档，用户现选择**整理spec**；这允许交付行为规格、设计、任务与验收模板，未将产品实现/发布/部署标为已授权或已完成。

F04本地已经对已知模型非401错误使用模型冷却，不能称为“模型隔离完全缺失”。缺口是可选择scope、旧规则等价回显、显式账号级行为及受约束的模型401。F05现有Anthropic搜索服务同样作为基座复用。

## 4. 逐项 patch sites 与基座

以下定位是规格准备证据，实施时按最终起点复核。具体行为以六份spec为准；本节不授权额外平台/入口。

### F01 使用日志输出TPS

- 位置：`frontend/src/components/admin/usage/UsageTable.vue`的`cell-latency`，`frontend/src/utils/latencyHealth.ts`，`frontend/src/i18n/locales/{zh,en}/dashboard.ts`。
- 消费者：`frontend/src/views/{admin,user}/UsageView.vue`共用UsageTable；当前均已有latency列和独立隐藏偏好。已有`UsageLog.output_tokens/duration_ms/image_count/image_output_tokens/billing_mode/request_type`足够，不改后端响应。
- 新符号：`formatUsageOutputRate`在本地缺失，来源为16行格式化/校验函数，只依赖已有UsageLog类型；无后端或数据库前置。
- 验收：100/2000=50.0tok/s、含TTFT、推理不再加、0.0与不可用区分、图片输入文本输出、图片/Live排除、两页面隐藏偏好/刷新。运维`queryOutputTPS`、`percentile_cont`及overview.output_tps明确不取。

### F02 账号筛选记忆

- 位置：`frontend/src/views/admin/AccountsView.vue`的`useTableLoader.initialParams`、`debouncedReload`，以及`frontend/src/components/admin/account/AccountTableFilters.vue`的实际选项。
- 基座：平台anthropic/openai/gemini/antigravity/grok；状态active/inactive/error/rate_limited/temp_unschedulable/unschedulable，外加空值。页面已有排序/列/刷新偏好，不改其持久化逻辑。
- 上游`CONCRETE_PLATFORM_OPTIONS`本地无该清单，不能为一个校验导入新平台画像系统。使用当前选项同源常量或最小抽取；新load/save helper独立于后端。
- 验收：首个请求即恢复；逐字段非法/坏JSON/读写异常；保存清空；回第一页/清理选择；不持久化额外查询条件。

### F03 运维账号名称

- 位置：`frontend/src/views/admin/ops/components/OpsRequestDetailsModal.vue`的`fetchData`、桌面table与移动card；中英文admin/ops文案；新增局部名称映射模块可置于已有ops/utils目录。
- 基座：`frontend/src/api/admin/accounts.ts:list`已有lite/include_scheduler_score参数，后端`backend/internal/handler/admin/account_handler.go`已有lite响应，不返回credentials。`OpsRequestDetail.account_id`已存在。
- 上游新增`getAccountNameMap`约33行，可复用思路；其无限期模块缓存和`while(map.size<total)`不能直接当作本地完整验收。实现须处理并发、缓存失效、重复页/空页终止、会话和迟到结果。
- 验收：成功/错误记录、桌面/移动、名称/#ID/-、分页/TTFT排序，失败不清空主明细。无需后端JOIN或新列。

### F04 模型级冷却规则

- 位置：`backend/internal/service/{account.go,temp_unsched.go,ratelimit_service.go,openai_account_runtime_block_fastpath.go}`；现有gateway错误/重试/count_tokens调用点；`frontend/src/components/account/{CreateAccountModal,EditAccountModal}.vue`、`frontend/src/types/index.ts`及中英文admin/accounts。
- 已有符号：`GetTempUnschedulableRules`、`matchTempUnschedulableRules`、`tempUnschedulableModel`、`triggerTempUnschedulable`、`SetModelRateLimit`、`IsModelRateLimited`；后者沿`extra.model_rate_limits`保存并同步调度快照。本地无需新增scope数据库列。
- 新helper：scope归一、有效范围判定及401凭据判断在来源PR内定义；必须连同消费者适配。当前模型非401路径已不会全局暂停，保留写失败不扩大范围。
- 重点：首匹配规则；真正发送模型与调度键；未知模型账号回退；OAuth/SetupToken/影子及撤销Token保持认证语义；池模式和原错误码筛选的实际生效链。上游Bedrock相关文件本地存在既有消费者，按现有Anthropic账号类型适配，不引入新平台。
- 验收：真实SQLite+miniredis范围、TTL/清理、普通/粘性/WS选择、图片/文本键、写失败、凭据反例、受影响链路race及账本固定窗口回归。
- 主规格`existing-gateway-correctness`的F07旧范围写有“不新增模型级规则UI”，本change提供完整MODIFIED需求，保留原图片场景；原归档文件不改。F12限制运维TPS字段与本次日志派生字段无冲突。

### F05 OpenAI API Key搜索模拟

- 位置：`backend/internal/service/{account.go,gateway_websearch_emulation.go,openai_gateway_messages.go}`；账号创建/编辑表单；`frontend/src/views/admin/ChannelsView.vue`必要的OpenAI开关回显/序列化；真实Messages handler的结果/槽位/用量调用点。
- 已有：`GetWebSearchEmulationMode`仅接受Anthropic API Key；`GatewayService.shouldEmulateWebSearch/handleWebSearchEmulation`、搜索manager、提供商/代理和JSON/SSE写出均已存在；`Channel.IsWebSearchEmulationEnabled(platform)`已支持按平台读取。
- 缺口：`OpenAIGatewayService.ForwardAsAnthropic`在当前早期分支会直接转Chat，模拟需在该分支前判定。来源新增共享helper`shouldEmulateWebSearchFor`与`writeEmulatedWebSearch`可由已有函数抽取，不新增搜索引擎。
- 本地适配：创建/编辑表单和渠道保存目前只支持Anthropic；仅复制上游编辑表单无法让“跟随渠道”在本地UI完整配置。规格因此包含OpenAI账号创建和渠道布尔开关的最小扩展，不能携带渠道支付相关代码。
- 验收：全局硬开关、三态/旧bool、纯搜索单工具、OAuth/SetupToken/影子与混合工具反例，JSON/SSE唯一终态、LLM零调用、代理/取消/错误与真实handler用量归属/释放。使用本地搜索提供商桩，不使用真实收费凭据。

### F06 最小输出Token配置

- 位置：`backend/internal/config/{config.go,config_test.go}`，`backend/internal/pkg/apicompat/{types.go,anthropic_to_responses.go,chatcompletions_to_responses.go,chatcompletions_anthropic_bridge.go}`；已有Gateway/OpenAI/Gemini/Antigravity/Grok兼容转换调用点；完整和个人config示例。
- 已有：三种转换函数与`minMaxOutputTokens=128`；直接Anthropic→Chat和相关测试在本地存在。新WithMinOutputTokens函数与配置helper均由来源PR定义，只需当前config和apicompat，不应使用可变全局变量。
- 本地全体调用点须重查，特别是`openai_gateway_count_tokens.go`已有两次AnthropicToResponses：记录其是否影响构造请求输出字段并保持原计数/probe语义，不漏接真实转换也不扩大count_tokens功能。
- **两处NOBASE**：`openai_gateway_chat_completions_anthropic_native.go`是缺失平台链；`openai_responses_ingress_compat.go`的原基座222行、终态227行，本地没有该legacy入口。后者不能作为“只是漏传Config”的小修套入。
- 本地同名`normalizeOpenAIResponsesWebSocketCompatibilityBody/openAIResponsesCompatibilityOptions`在`openai_oauth_web_search_history.go`，约20行，仅补OAuth搜索历史声明，不执行legacy消息转换。故本规格只配置已有转换；HTTP/WS首帧和后续原生frame列为不变反例，不引入222行新入口。
- 验收：默认128/下限16、YAML/env/非法值、转换三边界/双字段优先级/nil和旧调用/服务实例隔离；原生Responses、compact、WS多轮、OAuth搜索历史、原生Claude探针和Ollama cap反例。

## 5. 不做与历史遗留

- #7911运维百分位卡片/PG查询不取，未来单独选择；本次TPS仅使用日志。
- 账号紧凑筛选布局曾被上游撤回，本地保持平铺；F02只新增记忆。
- Haiku5.5、Argon、Gemini显式缓存/思考摘要、Vertex API Key/配额、生图开关扩围、分组多语言和四项性能方案仍按需；新平台、插件、pinned/live目录、完整Fast/Ultrafast等不随本次加入。
- 更早的Claude原生额度查询/重置、用量费用趋势、风控白名单、API Key创建限额、余额在途预占、批量管理/月备份等仍按既有决定；个人simple模式收益需要实际用途判断。
- 已完成旧批次以当前[归档清单及completion-record](../../openspec/changes/archive/README.md)为准；不把旧冻结评估中的“待实施”再次当现行欠账。两份旧活动变更的剩余项与本次六项独立。
- SQLite后台仍跳过的UsageCleanup/AccountExpiry/ScheduledTestRunner不在本次范围，不为清理冷却引入后台服务。用量记录/配额统计不是支付，继续保留；支付功能始终排除。

## 6. ALREADY、语义已有与不适用

全文件反向可应用数为0，但F04底层模型隔离、F05搜索服务和F06默认floor已有生产行为。文件态不是能力是否存在的判定。

F01同PR的后端SQL虽多处CLEAN仍排除；F06两个NOBASE不能当成必需迁移。其它共享文件CONFLICT需按当前已有协议/结算改动手工对齐，不回滚此前修复。

## 7. 建议实施顺序

1. F01/F02/F03分别完成展示与列表状态；F03缓存单独验收。
2. F04补scope与认证边界，完成SQLite/miniredis/调度回归。
3. F05复用搜索及三态配置，验证转换前分流与生命周期。
4. F06按现有调用图传递配置，联合F05验证Messages两条路径。
5. 最后验证共享账号表单、六项组合与完整后端/前端门禁，填写verification。

本节是后续实施顺序，不是本次执行记录；未使用子代理。

## 8. 基线检查与未来验收

本次已确认相关测试文件与Go测试目录存在并执行定向基线检查：四个Go包通过，前端8文件122项通过。具体命令、现存/尚待新增测试及build tag记录见[baseline-checks](./evidence-0.2.18-personal-features/baseline-checks.md)。

源码定位、来源检查及基线通过均不是新增六项的实现证据。未来需按tasks完成新增场景、真实SQLite/miniredis、受影响race、后端unit/build、前端typecheck/lint/Vitest/build；实施verification保持空白。

## 9. 本次通用教训

1. 用户选定按需功能后建立独立规格，不把历史第三档改写成已证实缺陷或实施授权。
2. 本地已有部分能力时，规格描述真实增量；模型隔离存在不等于规则scope已可配置。
3. 同名兼容helper可以只有很小的不同职责，缺失文件须沿生产调用链量测；不得为配置穿透引入新的legacy入口。
4. 配置继承要同时核对读取、保存和UI设置者；后端能读OpenAI标志不代表本地渠道页面能保存它。
5. 从归档汇总来的主规格也可能带有旧批次“本轮不新增”的限制；新增授权用精确delta解除冲突，不回写冻结来源。

## 10. 实施完成补记（2026-10-26）

六项行为及既有F07兼容delta已完成，完整运维页按后续选择恢复。实际起点为`67b1a0727`及原未提交规格；后续HEAD为`f96d00ef8`，两者之间仅历史归档文档变化。产品实现尚未提交，不能把该HEAD当作实现SHA；以[72项产品文件哈希](./implementation-0.2.18-personal-features/completion-2026-10-26/product-files.sha256)标识验收工作树。[逐场景映射](./implementation-0.2.18-personal-features/acceptance-matrix.md)覆盖22条需求、50个Scenario，[验收报告](../../openspec/changes/port-upstream-0.2.18-personal-features/verification.md)给出可复跑命令和实际来源取舍。

| 门禁 | 最终结果 |
|---|---|
| 后端unit/build/嵌入资源 | 54个有测试的包通过，完整build和embed资源测试通过 |
| 受影响链路race | config/apicompat/service/handler/repository五包通过，在最后凭据修正后执行 |
| SQLite/miniredis | account/model范围、调度快照、运维查询/聚合/TTFT排序、删除Key结算/去重/固定配额窗口通过 |
| 前端typecheck/lint/Vitest/build | 全部通过；285个测试文件、2089通过、2项既有跳过 |
| 页面与配置 | 中文真实Vite固定API桩截图；桌面/移动运维详情、英文文案、权限/隐藏菜单/监控开关由对应组件及路由用例覆盖；无业务实例部署 |
| 范围 | VERSION1.1.19、Go1.27.3、最大迁移226；依赖、Ent/Wire不变；无支付、新平台、运维TPS分位数或缺失legacy入口 |

具体修正的红→绿记录保留于[整合审查](./implementation-0.2.18-personal-features/review.md)：异步使用量先快照Gin请求字段；每个SSE事件Write/Flush边界检查取消；手动刷新名称映射替换旧快照且返回错误详情上下文后续接加载；跨平台明确凭据失败不被模型规则吞掉。基线结构检查与产品验收仍分别记载。

本批不创建发布、tag、push或部署，不归档change；§5其他按需能力和个人无支付边界保持原决定。
