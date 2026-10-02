# 上游 v0.2.12 移植评估（v0.2.11 → v0.2.12）

评估及实施日期：2026-10-02。**第一档3簇、第二档3簇已开发及验收完成。**
实现提交 `405246789`，分支 `sync/upstream-0.2.12`，实施基线 `5eb58199e`。用户确认两项管理便利功能后，又明确授权“完成第一档和第二档”。

[第一档验收](../../openspec/changes/port-upstream-0.2.12-tier1/verification.md)、[第二档验收](../../openspec/changes/port-upstream-0.2.12-tier2/verification.md)：全量Go unit/build、定向race、SQLite方言/真实用量去重/分组排序，前端typecheck/lint/build及1854通过/2既有跳过。VERSION仍1.1.15、最大迁移226，无依赖变更，未发布或部署。

下文来源、四态与基线测试保持评估语境，source-baseline及evidence-0.2.12冻结不改；当前实施差异另见§10。

## 1. 版本、范围与规模

| 基线 | 值 |
|---|---|
| 本fork main | `5eb58199ee6ef45515ec791b445305651384fc0c`；起始工作树干净；VERSION 1.1.15 |
| 上次已评估v0.2.11 | `96f4c115c9749078f90cbf210a01d39baf3f53b6`；其第一/二档已落96f32d018 |
| 目标v0.2.12 | `5106065716e494204fc0e8db16f68f6e9d576be0`；Release 2026-10-02 14:45:01北京时间 |
| 抓取时upstream/main | `458b92abd4b0b09d123d4c6727dd8d59060bd883`；tag之后的VERSION提交不纳入 |
| 来源规模 | 19条非merge、9个PR、1条直接提交；净151文件/+5779/-311 |
| 运行边界 | SQLite-only、最大迁移226、可关闭Redis/miniredis、simple mode、自有1.1.x |

来源：[Release](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.12)、[发布说明快照](./evidence-0.2.12/releases.json)、[完整非merge列表](./evidence-0.2.12/commits.tsv)。

| 档位 | 来源与簇数 | 决定 |
|---|---|---|
| 第一档 | 3PR/3簇：#7780后端、#7674、#7773 | 已完成，见对应验收 |
| 第二档 | 3PR/3簇：#7676、#7630、#7803 | 已完成，见对应验收；后两项用户确认需要 |
| 第三档 | #7780中的axios依赖升级子项 | 独立依赖维护，未确认收益，不随Grok后端引入 |
| 第四档 | 3PR：#7425、#7802、#7673；1直接提交 | 缺失平台/支付越界/上游版本号，不移植 |

#7780按行为拆分，档位来源计数不可直接相加当作独立PR数。

## 2. 方法与适配边界

执行 `git fetch upstream --tags --prune`。以tag祖先范围枚举非merge，PR使用完整 `git diff <merge>^1 <merge>`，不逐微提交套用。
用独立临时GIT_INDEX_FILE读取冻结fork SHA；**先对全部候选反向check，再正向check**，不更改工作树。均失败且本地路径不存在才记NOBASE。

160条逐候选文件记录（重复文件分别计）：**ALREADY 0 / CLEAN 108 / CONFLICT 42 / NOBASE 10**。
[候选](./evidence-0.2.12/candidates.tsv)、[四态](./evidence-0.2.12/files.tsv)、[分档](./evidence-0.2.12/decisions.tsv)、[基线SHA](./evidence-0.2.12/baseline.json)均冻结。

[新增Go调用清单](./evidence-0.2.12/added-call-inventory.tsv)覆盖所有来源新增行，按本地func定义检索；包含测试、外部包、闭包及注释/正则误命中，不能据同名命中宣称可编译。[调用复核](./evidence-0.2.12/call-review.md)、[关键基座](./evidence-0.2.12/base-symbols.txt)和[前端基座](./evidence-0.2.12/frontend-base-symbols.txt)记录实际定义与消费者。选中范围的新增helper全部随消费者，外部调用沿用stdlib、go-redis、Vue与既有API；未选中平台/支付调用不补基座。

上游新增两个同号241迁移：payment bonus列、TypeSafe平台PG CHECK约束。**本轮均不引入**，不改历史迁移、不占用旧方案预留227；以后schema工作从实施时实际最大号重新分配。#7780的axios 1.20.0与后端修复无依赖，不整取上游lockfile（包含支付依赖）。

## 3. 第一档：本地既有问题的小范围修复

A/C/X/N为ALREADY/CLEAN/CONFLICT/NOBASE，按完整来源计。

| 簇 | 来源 | A/C/X/N | 本地缺口 |
|---|---|---|---|
| F01 Grok CLI 1.0.46身份 | #7780 | 0/9/2/0 | CLIClientVersion仍0.2.120、minimum 0.2.93，UA为workspace；上游报告426最低1.0.13 |
| F02 Antigravity客户端错误脱敏 | #7674 | 0/2/1/0 | ForwardGemini非failover错误仍用c.Data返回unwrappedForOps |
| F03 自定义错误码说明 | #7773 | 0/2/0/0 | 中英文仍宣称未选中错误统一返回500，与现有重试/透传语义不符 |

### F01 patch sites / 基座

`backend/internal/pkg/xai/billing.go`的CLIClientVersion/ApplyCLIBillingHeaders，`cli_identity.go`的CLIStableVersion、CLIUserAgent、ApplyCLIProxyHeaders；`backend/internal/repository/http_upstream.go`的applyGrokCLIProxyHeaders及newGrokOfficialAPIFallbackRequest；`backend/internal/service/grok_upstream_headers.go`与`openai_gateway_grok.go`的applyGrokCLIHeaders均存在。

包级校验最低1.0.13，transport仍按preferred pin 1.0.46更严格校验，不能误统一两层规则。UA使用运行平台的Rust命名（linux/x86_64、macos/aarch64）；interactive、grok-pager与authenticate-response头覆盖CLI目标；官方API回退剥离新CLI头，独立api.x.ai调用仍遵守原路径规则。额度探测是账号运营元数据，不是支付。
不动frontend/package.json或pnpm-lock.yaml。未发真实Grok账号请求，426依据为上游说明和测试，本地证据是旧出站版本/身份代码及现有transport回归。

### F02 patch sites / 基座

`backend/internal/service/antigravity_gateway_gemini.go:399`直接输出原错误体。上游新增97行 `antigravity_upstream_error_sanitize.go`复用本地 `sanitizeUpstreamErrorMessage`（gemini_messages_compat_service.go）和 `extractUpstreamErrorMessage`（gateway_upstream_response.go），只依赖标准库。
保留HTTP状态与Gemini code/status/message，message去除项目引用、服务账号邮箱、敏感查询参数，details不返回客户端；非JSON/空正文有安全默认值。ops原有诊断与failover行为保留。需真实ForwardGemini出口测试，不能只测helper；不得宣称正则覆盖所有任意敏感字符串。

### F03 patch sites / 基座

`frontend/src/i18n/locales/{zh,en}/admin/accounts.ts`的customErrorCodesWarning。后端 `Account.ShouldHandleErrorCode` 空列表不筛选，仅决定常规账号错误处理；网关自己的retry/failover/透传规则另外决定请求结果。只改说明，不改调度规则。

规格：[第一档](../../openspec/changes/port-upstream-0.2.12-tier1/README.md)。

## 4. 第二档：缓存并发修复与用户确认的管理便利功能

| 簇 | 来源 | A/C/X/N | 理由 |
|---|---|---|---|
| S01 邮箱验证码与重置令牌原子化 | #7676 | 0/9/0/0 | 接口及全部mock、Redis Lua、缓存格式与并发语义需联合验证 |
| S02 账号优先级快捷调整 | #7630 | 0/5/0/0 | 用户明确需要；新交互有450ms合并保存、失败回退、卸载与键盘行为 |
| S03 密钥按分组名排序 | #7803 | 0/4/0/0 | 用户明确需要；Ent排序必须在SQLite分页前执行，NULL最后 |

### S01 patch sites / 基座

`backend/internal/repository/email_cache.go`目前GET/SET整段JSON；`backend/internal/service/email_service.go`的VerifyCode读取Attempts后自增回写，ConsumePasswordResetToken先验证再删除且删除失败仍成功；`user_service.go`的verifyNotifyCode有同类计数问题。AuthService.ResetPassword实际消费此token。

新增接口 IncrVerificationCodeAttempts、IncrNotifyVerifyCodeAttempts、ConsumePasswordResetToken；补齐所有EmailCache实现与handler/service测试stub，不仅上游列出的几个。新增verifyCodeWithAttempts与hashPasswordResetToken，cache get/set/delete/incr helper一起加入；Redis脚本使用EXISTS/INCR/PTTL/PEXPIRE和cjson比较删除，需在内置miniredis运行验证。

每次比较前原子占用尝试次数，缺失/过期/缓存错误拒绝；普通邮箱与通知邮箱保持各自成功删除时机。计数key TTL对齐主key，重新发码清理旧计数，不延长旧码寿命。只存SHA-256摘要，重发生成新token，原子compare-and-delete只允许一个消费成功；旧明文缓存链接自然失效，需重新申请，不能兼容回退明文。

[确定性并发探针](./evidence-0.2.12/baseline-race-probe.txt)通过Go overlay注入测试、不修改产品文件，使用带屏障的线程安全cache替身重现20并发错误猜测写回Attempts=1，以及20并发消费者接受同一token。它证明服务层读改写时序问题，**不是外网攻击实测，也不是Redis新脚本验收**。探针断言旧缺陷存在，不作为未来通过门禁。

### S02 patch sites / 基座

新增 `frontend/src/components/account/AccountPriorityCell.vue`并接入 `frontend/src/views/admin/AccountsView.vue`的cell-priority；本地 `api/admin/accounts.ts` update、`utils/apiError.ts` extractApiErrorMessage、AccountsView.handleAccountUpdated均存在。中英文priorityQuick说明随组件。

只PATCH priority；1–9999边界，减小数字提高优先级。连续点击450ms合并一次保存，点击数字编辑、Enter/blur提交、Esc取消；保存中禁止重复提交，错误恢复服务器值并提示。卸载时未发送的按钮修改按来源行为提交一次，保证固定账号归属；验收快速翻页/异步回包不改错账号，键盘和触屏可操作。此新组件非现有缺陷补丁，用户确认后明确提升第二档。

### S03 patch sites / 基座

`backend/internal/repository/api_key_repo.go:627` apiKeyListOrder新增group分支，已有Ent apikey.ByGroupField/group.FieldName与ListByUserID过滤分页；`frontend/src/views/user/KeysView.vue`的group列设为sortable。

按分组名称而非ID排序，升降序均未分组最后，以ID同向次排序保证跨页稳定；搜索、状态、用户隔离、软删除过滤及group预加载保持。来源 `api_key_repo_sort_test.go`为无integration标签的SQLite用例，所需newAPIKeyRepoSQLite/mustCreateAPIKeyRepoUser本地存在；正式实施必须运行真实SQLite查询，不能只断言生成SQL或前端列配置。

规格：[第二档](../../openspec/changes/port-upstream-0.2.12-tier2/README.md)。

## 5. 按需、不做与历史遗留

### 5.1 本轮第三/四档

- **第三档：#7780 axios升级子项**。本地^1.18.0解析1.18.1；上游升1.20.0。2文件/依赖维护，与CLI身份无关系。该依赖子项没有新audit或缺陷复现，不将其描述为已证实安全修复；可与旧DOMPurify/xlsx维护单独评估，不自动纳入。
- **第四档：#7425 TypeSafe Jev System One原生平台**。78文件/+2322/-64；本地typesafe/client.go缺失，上游不仅补client，还增加handler/service/测试/内容审核/审计/UI及两个平台约束。缺失平台而非现有网关bug，不为它搬入其他CN/OpenCode枚举或PG DDL；若以后要用需要重新明确需求和SQLite schema设计。
- **第四档：#7802充值促销与#7673支付公开订单验证限流**。均payment-only，个人使用明确排除。#7673虽为安全修复仍不列第一/二档，也不当可顺手携带项。禁止引入支付实现作为非支付功能前置。
- **第四档：42bc7f6cf上游VERSION同步**。继续本fork 1.1.15；目标tag本身的上游VERSION文件仍0.2.11，release与tag范围优先，不能据该文件误判无新版本。tag后458b92abd才同步0.2.12，不扩张本次范围。
- #7630/#7803原按使用方式列第三档，用户已答“需要两项，纳入后续计划”，所以**明确调整为第二档S02/S03**，不是默默提升。未确认其余按需项。

### 5.2 上一轮与更早未合项（当前状态）

0.2.11第一档4簇、第二档6簇在96f32d018完成，已包含于本地v1.1.15；旧文件里的部署1.1.14与source-baseline保留历史值。其第三档Claude原生额度查询/重置、风控白名单、趋势切换、Key创建限制、余额在途预占、Astra Ultrafast仍未实施：按实际用途决定，simple mode不使用余额预占，不将账号额度元数据误判支付。


- **0.2.9第二档16簇已在159d04222落地**，v1.1.14已包含；旧PORTING与source-baseline是历史快照。0.2.11 S01/S02/S03/S04/S05/S06分别与已落模型/流/路由/客户端配置相交，应基于当前代码适配。
- **0.2.9第一档原39PR/15簇：F04已作为0.2.11 S02前置完成，其余仍待实施**：RPM投影、OAuth重授权、账号身份/错误状态、beta头、Responses消息/终态正文、Alpha search、未来reset暂停、TOTP、公共表单、代理小修、配置输入、异步弹窗、用量显示、ID/配额诊断。此前dialog ID counter和空Tab前置已随第二档S14落地；0.2.11另外完成F04两桥契约，不能将旧第一档整批算完成。
- **0.2.9第三/四档**：OpenCode Go/Zen/Command Code/Kimi原生/插件、model_allowlist glob/pinned/live混合目录、Free Fast/effort倍率/自动用卡重置、视频/图像新路由、TypeSafe/自动版本/月度备份、运营与部署增强仍待用途决定；支付/返利排除。现有白名单编辑器不等于model_allowlist基座。
- **0.2.5主体已落bfbcd79**；#7043默认WS系数5.0仍未采用（现有默认1.0），#6954/S12b平台额度清理与预留227未实施；#6960/#7049完整exec_scope/抢占/queue-wake仍未整体引入。
- **0.2.5按需目录/平台**：#6575/#6890/#6974/#6769 Gemini3.7/3.8、原生Codex Images、DeepSeek峰谷价、Ollama异步重置；#6747/#7091/#6691/#7153平台/订阅批量/Key批量与过滤；#6989/#7057/#6968/#7011/#6945/#6988目录/上下文/视觉/compact/pinned，不因新模型升级自动完成。收费开关#6986不引入。
- **0.2.5差异与管理**：#7129/#7085/#6929/#6917 Grok序号、input stripper、Gemini带内错误与监控BaseURL继续独立适配；#6847/#7001/#6845/#6913/#6915管理展示待选。#7126/#7110/#7148已有namespace语义/上游撤回/PG等不重复移植。
- **0.2.0 / 0.1.180 / 0.1.184长上下文和Fast**：本地只有fast→priority价格别名、旧gate逻辑；动态阶梯/effort/Fast/免费Fast/Ultrafast完整策略未合。0.2.11 S02排除Ultrafast，不藉模型升级带入新价格体系。
- **0.1.180工具桥尾项与Grok/WS**：非法参数、namespace别名、deferred标记尚未整体合；0.2.11 F01是工具名请求体修复，不替代该旧簇。Grok冷却/Realtime/透传/抢占及Redis跨节点租约按用途/架构保留原取舍。
- **0.1.184其它**：usage compaction/requested effort迁移、公共分组限制、订阅重置锚点、峰谷价/图像工具冷却仍独立；不改已应用schema。0.1.180自动重置卡/模型读取上限未采用；0.2.11 Claude原生重置不是同一子系统。
- **0.2.11已解决的旧基座**：0.1.180第四档695ebede7只有14行通用字段为0.2.11 S03所需，当轮明确提升该子集；0.2.9 F04的10行helper是S02前置。其它“缺基座”不能从字段为零命中直接推导工程量或不适用。
- **依赖与后台**：DOMPurify仍^3.3.1、xlsx仍^0.18.5；本轮仅评估、未改lockfile、未运行新audit，不沿用历史advisory数量。UsageCleanupService、AccountExpiryService、ScheduledTestRunnerService仍受SQLite跳过条件控制，需要各自SQL审计后再启用。
- **已落旧轮次**：0.2.4、0.2.0主体、0.1.183/179/176不重新立项；API Key instructions、routed catalog、Responses Lite等按现有实现保留。


### 5.3 本轮重新量测，避免继承“缺基座”标签

[当前探针](./evidence-0.2.12/leftover-probes.txt)核对DOMPurify仍^3.3.1，锁文件同时有3.3.1/3.3.3，7处直接sanitize调用含公开LegalDocumentView；不是缺代码基座，而是未完成依赖升级验证。xlsx仍^0.18.5。旧advisory数量不作当前结论，不以本次axios提交替代依赖审计。
TypeSafe前置重新量为上面78文件范围，不沿用旧“缺client所以做不了”的空泛判断；三个SQLite后台worker仍受skipSQLiteBackgroundJobs控制，应逐SQL审计，不能默认永久停用。旧0.2.9 F04 helper与695ebede7通用usage字段已经存在，不再列缺前置。
其余历史项目以上述实际落地提交、旧冻结来源及当前符号复核为依据，未声称本轮逐个重跑全部历史功能。

## 6. ALREADY、N/A与净终态

本轮0条ALREADY是整文件补丁态，不等于所有行均未采用；上一轮upstream/main的42bc7f6cf已经进入本轮tag范围，但它是上游VERSION不适用，不借此覆盖fork版本。
#7425中的不存在平台调用与#7802/#7673支付调用无需补齐。#7780后端和axios分开；新UI均先按目标tag读完整PR终态再设计。已有0.2.11实现不重新建change。

## 7. 建议落地顺序

1. 基于实施时实际SHA开分支并复核本轮冻结基线。先F01 Grok身份，F02出口脱敏，F03文案。
2. S01先补可复现并发失败的验收用例，再接口、缓存脚本、服务及所有mock一并适配；单节点miniredis与真实Redis（若可用）分开记证据。
3. S02/S03按用户需求实施；共享前端locale文件串行修改，排序用SQLite多页/用户隔离验证。
4. 运行受影响unit/build、并发race、SQLite方言/用量去重、前端typecheck/lint及组件测试；回填实施verification。第三/四档保持排除。
5. 只有后续授权发布/部署才进入独立流程；本轮不改VERSION、不打tag。

## 8. 基线自测与实施门禁

[基线测试报告](./evidence-0.2.12/baseline-tests.md)、[go test -list](./evidence-0.2.12/backend-test-inventory.txt)、[来源测试首行/存在性](./evidence-0.2.12/source-test-inventory.tsv)。
相关xai/repository/service/handler基线定向unit及SQLite方言检查通过；KeysView现有12项、EditAccountModal现有40项测试通过；用户确认排序后另跑APIKeyRepository现有回归与测试枚举通过。新增上游测试未落，不能混入“通过”计数。前端typecheck通过，结果见基线报告。
email_cache_integration_test.go首行只有integration，使用IntegrationRedisSuite，未跑，不能说是postgres-only；仓库另有integration && postgres文件，不为本轮修改标签。

评估阶段未运行全量unit/build/race，当时OpenSpec任务未勾选、证据槽留空；现已在独立实施证据中完成上述验收并回填。旧并发探针只记录原基线缺陷。

## 9. 新增通用教训

1. 安全修复也先过个人使用边界：payment-only限流修复不因补丁便宜而进入候选实施。
2. 同一PR夹带无关依赖升级，不能为后端CLI修复整取frontend lockfile。
3. CLEAN的缓存补丁可能改变接口、并发及已发链接有效性，必须按行为风险定档。
4. 包级最低版本与transport preferred pin可以有意不同，勿为了“统一”弱化出站检查。
5. 用户选择按需项后，记录原档位、确认内容和新档位；“纳入后续计划”不等于立即实现。
6. Release/tag时间与仓库VERSION同步提交可能错位，冻结tag SHA并单列tag后提交。

## 10. 实施补充（2026-10-02）

- S01除来源原子化方案外，补齐旧JSON Attempts继承，避免升级让未过期验证码重新获得5次尝试；记录失败/通过证据。实际SMTP邮件链接、缓存哈希、重发失效、AuthService单次改密及miniredis并发/TTL均验证。
- S02补齐严格整数输入与账号ID/请求代次隔离，阻止行复用时错账号写入和旧响应覆盖；16项组件测试包含去抖、键盘、失败恢复、卸载和异步场景。
- S03采用来源SQLite排序测试，KeysView来源测试里selectedKeys批量选择断言不适用，排除而不恢复批量功能；剩余分组参数/筛选/分页测试与SQLite测试通过。
- F02保留本fork的重试与错误处理，只替换最后客户端错误序列化；实际ForwardGemini出口已测试。
- 支付两PR、TypeSafe和axios仍排除；旧0.2.9第一档等历史未合项不因本轮完成而自动标为已合。
