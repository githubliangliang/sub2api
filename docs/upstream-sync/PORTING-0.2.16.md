# 上游 v0.2.16 移植评估（v0.2.12 → v0.2.16）

> 当前实施进展（2026-10-18）：第一档14簇已在 `7c544c03f` 完成（实施起点 `0fbcab935`），验收见[第一档 verification](../../openspec/changes/archive/2026-10-24-port-upstream-0.2.16-tier1/verification.md)。第二档16簇已提交为 `847db09e1`（功能修复）与 `5ff6e5bf9`（依赖维护），并快进合入 main，验收见[第二档 verification](../../openspec/changes/archive/2026-10-24-port-upstream-0.2.16-tier2/verification.md)。实施验收使用 VERSION 1.1.17，当前 VERSION 1.1.18 用于用户要求的发布；未部署业务实例。下文为2026-10-13评估快照，来源SHA/四态/基线证据保持冻结。

评估日期：2026-10-13。**仅评估，尚未实施：第一档14簇、第二档16簇，覆盖69个本轮PR的选定行为。**
用户已选择“暂不增加，优先现有功能修复”。新增平台、Haiku 5.5、Gemini缓存管理/Vertex新认证、模型级冷却配置、分组多语言、TPS和紧凑筛选暂不加入。支付功能始终第四档。

重点建议先处理删除Key后的结算回滚、安装凭据校验、协议转换/重试与现有Gemini额度错误，再处理前端竞态和独立依赖维护。源码、VERSION、迁移、lockfile未改，无提交/发布/部署。

## 1. 版本、范围与规模

| 基线 | 值 |
|---|---|
| 本fork main | `128c5f7bfd4b8ea3f824146ed3b005e9d3e34887`，工作树起始干净，VERSION 1.1.16 |
| 上次评估tag | v0.2.12 / `5106065716e494204fc0e8db16f68f6e9d576be0` |
| 抓取时upstream/main | `4d9ac3e482106d2da3a39fe3e546d8d6801f1cc1`；tag后仅VERSION同步，不扩大范围 |
| 新区间 | 149条非merge / 89个PR / 4条直接VERSION提交；净477文件/+34182/-2129 |
| 运行边界 | SQLite-only，最大迁移226；可选Redis/miniredis；simple mode；保留128c5f7bf固定quota窗口修复 |

| Release | 解引用commit SHA | 发布时间（UTC） | 非merge / PR / 直接提交 |
|---|---|---|---|
| [v0.2.13](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.13) | `3040209f205472038c1ba745a1bedd2edd9053b1` | 2026-10-02T12:16:38Z | 7 / 3 / 1 |
| [v0.2.14](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.14) | `0363b8cdba8cec3e2ba4b2dbd49c4481143fa55d` | 2026-10-07T02:02:43Z | 5 / 3 / 1 |
| [v0.2.15](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.15) | `f2669c8cf62555cd92389b3f55920e9e6e7c6ff2` | 2026-10-09T03:35:49Z | 86 / 59 / 1 |
| [v0.2.16](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.16) | `01842576b3d5c1e2545b5a99ad555d36e28ed306` | 2026-10-13T07:20:24Z | 51 / 24 / 1 |

来源：[GitHub Releases](https://github.com/Wei-Shaw/sub2api/releases)、[release快照](./evidence-0.2.16/releases.json)、[149条非merge列表](./evidence-0.2.16/commits.tsv)、[93个完整候选](./evidence-0.2.16/candidates.tsv)。统计按tag祖先关系，非作者时间；annotated tag使用`^{commit}`解引用。

| 档位 | 簇数 / 本轮来源 | 决定 |
|---|---|---|
| 第一档 | 14簇 / 40PR | 已证实本地代码缺口的小修，生成OpenSpec，未实施 |
| 第二档 | 16簇 / 30PR | 有价值但需适配/跨路径验收；另有3条旧来源前置/维护，不算进149条 |
| 第三档 | 8PR含混合子项 | 按用户选择暂缓新能力；xlsx/工具链专项独立维护 |
| 第四档 | 18PR含混合子项 + 4直接提交 | 支付/缺平台/本地无此缺陷/上游专用内容，不引入 |

PR会跨档（尤其#7893/#7957/#8006）；档位来源数不能相加当独立PR数。[逐来源决定](./evidence-0.2.16/decisions.tsv)覆盖全部93候选；30簇的实际范围以下文和spec为准。

## 2. 方法与适配边界

执行`git fetch upstream --tags --prune`。按first-parent merge取`git diff <merge>^1 <merge>`完整PR，验证各候选覆盖的非merge集合恰好等于149条；禁止逐微提交移植。独立临时GIT_INDEX_FILE读取冻结fork SHA，先完成全部候选文件反向check，再做正向check，未触碰工作树。

575条逐候选文件记录：**ALREADY 0 / CLEAN 367 / CONFLICT 168 / NOBASE 40**。[四态](./evidence-0.2.16/files.tsv)保留两向结果和错误。状态按完整PR统计，含不采用的hunk和测试；零ALREADY不是“全部功能都缺”。

按源新增Go行采集9947条调用/词法记录并检索本地func及同PR定义；该清单含测试、类型转换、外部包、SQL和注释误命中，不能据同名命中证明可编译。[调用清单](./evidence-0.2.16/added-call-inventory.tsv)、[调用复核](./evidence-0.2.16/call-review.md)、[关键基座](./evidence-0.2.16/base-symbols.txt)、[hunk定位](./evidence-0.2.16/hunk-base-inventory.tsv)、[前端import](./evidence-0.2.16/frontend-imports.tsv)记录适用性。

上游新增5个SQL文件：两份242（平台约束移除、group i18n），243–245（缓存表/渠道计费/修改租约）。本轮选定修复**不需要新迁移**；不移植PG JSONB/锁语句，不改226及以前文件，不占用旧预留227。S08采用现有表的SQLite事务；将来若启用新能力，按实施时实际最大号重新分配。

源码基线保持1.1.16；Go按backend/go.mod自动选择1.27.0。S16是后续依赖维护计划，本次未更新依赖。上游完整lockfile包含无关及支付依赖，禁止整取。

## 3. 第一档：本地既有缺陷的小范围修复

A/C/X/N=ALREADY/CLEAN/CONFLICT/NOBASE，统计按来源整PR；重复来源会重复列数。


| 簇 | 来源 | A/C/X/N | 本地缺口 |
|---|---|---|---|
| F01 删除 Key 后完成请求仍结算 | [#7816](https://github.com/Wei-Shaw/sub2api/pull/7816) | 0/2/1/0 | API Key 计数返回 ErrAPIKeyNotFound 会回滚整笔 Apply；内存 SQLite 探针已复现。 |
| F02 Chat 请求角色与具名工具选择 | [#7785](https://github.com/Wei-Shaw/sub2api/pull/7785)、[#7786](https://github.com/Wei-Shaw/sub2api/pull/7786) | 0/3/1/0 | 具名 tool_choice 仍原样传递嵌套 function；developer 落入普通消息转换。 |
| F03 Anthropic 缓冲索引与 thinking 输出 | [#7789](https://github.com/Wei-Shaw/sub2api/pull/7789)、[#7869](https://github.com/Wei-Shaw/sub2api/pull/7869) | 0/9/0/2 | 本地两条缓冲路径只判断 idx<len，负数可越界；thinking 起始块省略 signature。 |
| F04 Claude 拟态缓存 TTL 与 Sonnet 采样参数 | [#7844](https://github.com/Wei-Shaw/sub2api/pull/7844)、[#7957](https://github.com/Wei-Shaw/sub2api/pull/7957) | 0/18/5/0 | mimic 注入的前置 5m 与后置 1h 断点顺序冲突；Sonnet 5.5 仅传 top_p 时仍补 temperature。 |
| F05 Grok 空完成事件与 SSO consent | [#7775](https://github.com/Wei-Shaw/sub2api/pull/7775)、[#7887](https://github.com/Wei-Shaw/sub2api/pull/7887) | 0/4/0/0 | 共享流处理的空完成保护仅识别 OpenAI；SSO approval 未传 consent token/origin。 |
| F06 Codex 已有模型元数据与远程发现配置 | [#7851](https://github.com/Wei-Shaw/sub2api/pull/7851)、[#8033](https://github.com/Wei-Shaw/sub2api/pull/8033) | 0/3/0/0 | 远程模式生成配置缺 api_key_model_discovery；GPT-6.1 Sol 生成目录默认 context 未更新至来源终态 500k。 |
| F07 Gemini 工具图片与 Antigravity 图片冷却隔离 | [#8038](https://github.com/Wei-Shaw/sub2api/pull/8038)、[#8046](https://github.com/Wei-Shaw/sub2api/pull/8046) | 0/4/1/0 | tool_result 只提取文本；图片 Gemini 与文本 Gemini 共用家族限流键。 |
| F08 价格、倍率与吞吐显示 | [#7716](https://github.com/Wei-Shaw/sub2api/pull/7716)、[#7721](https://github.com/Wei-Shaw/sub2api/pull/7721)、[#7720](https://github.com/Wei-Shaw/sub2api/pull/7720)、[#7897](https://github.com/Wei-Shaw/sub2api/pull/7897)、[#7800](https://github.com/Wei-Shaw/sub2api/pull/7800) | 0/11/2/0 | 价格字符串清零会截断指数；小于1字节单位越界；0倍率被 ||1 替换；长上下文硬写 x2；追加模型可能覆盖空价格条目。 |
| F09 共享弹窗和选择器输入边界 | [#7731](https://github.com/Wei-Shaw/sub2api/pull/7731)、[#7725](https://github.com/Wei-Shaw/sub2api/pull/7725)、[#7970](https://github.com/Wei-Shaw/sub2api/pull/7970) | 0/7/0/0 | Esc 会触发所有显示弹窗；选择器运行中 disabled 不关闭；IME 组字仍参与键盘导航。 |
| F10 管理端错误文本与开关响应 | [#7718](https://github.com/Wei-Shaw/sub2api/pull/7718)、[#7974](https://github.com/Wei-Shaw/sub2api/pull/7974)、[#7902](https://github.com/Wei-Shaw/sub2api/pull/7902)、[#8006](https://github.com/Wei-Shaw/sub2api/pull/8006) | 0/66/23/3 | 批量用户/Gemini OAuth/同步模型错误绕过统一错误解析；透传开关用本地反转覆盖返回状态。 |
| F11 计时器、访问重试与备份卸载 | [#7728](https://github.com/Wei-Shaw/sub2api/pull/7728)、[#7719](https://github.com/Wei-Shaw/sub2api/pull/7719)、[#7696](https://github.com/Wei-Shaw/sub2api/pull/7696)、[#7729](https://github.com/Wei-Shaw/sub2api/pull/7729)、[#7723](https://github.com/Wei-Shaw/sub2api/pull/7723) | 0/9/1/0 | 卸载后迟到请求可重开邮箱/OAuth/备份计时器；标签切回未立即刷新用量时钟；生图权限查询失败被永久记为 loaded。 |
| F12 现有监控健康分与时间/设置边界 | [#7916](https://github.com/Wei-Shaw/sub2api/pull/7916)、[#7843](https://github.com/Wei-Shaw/sub2api/pull/7843)、[#7969](https://github.com/Wei-Shaw/sub2api/pull/7969)、[#7973](https://github.com/Wei-Shaw/sub2api/pull/7973)、[#7842](https://github.com/Wei-Shaw/sub2api/pull/7842) | 0/8/2/0 | 无请求样本时仍可能凭 TTFT/cache 算健康分；分钟允许小数；本地时间输入用 UTC；自动刷新未更新 detail；设置未加载可保存。 |
| F13 账号测试诊断与分组名显示 | [#7837](https://github.com/Wei-Shaw/sub2api/pull/7837)、[#7878](https://github.com/Wei-Shaw/sub2api/pull/7878)、[#7983](https://github.com/Wei-Shaw/sub2api/pull/7983) | 0/6/1/0 | Antigravity 测试忽略显式映射；测试错误日志缺归属；分组名被裁切。 |
| F14 内嵌前端放行已有 API 别名 | [#7822](https://github.com/Wei-Shaw/sub2api/pull/7822)、[#7982](https://github.com/Wei-Shaw/sub2api/pull/7982) | 0/3/0/0 | 已注册的裸 /chat/completions、/embeddings 等可被 SPA 吞掉；来源测试使用过时资源路径。 |

### F01 patch sites / 基座

backend/internal/repository/usage_billing_repo.go:177 的 applyUsageBillingEffects、incrementUsageBillingAPIKeyQuota/RateLimit 均存在；只忽略 Key 已删除错误，保留本地 CURRENT_TIMESTAMP、用量 dedup 与 128c5f7bf 固定窗口修复。上游 PG integration 用例改写为真实 SQLite 验收。

应达到：请求执行中 Key 被软删除，余额/账号配额及 dedup 按既有模式正常落库，已删除 Key 自身计数跳过。

边界：非 Key-not-found 的数据库错误仍失败回滚；重放相同请求不重复计数；simple mode 原本未启用的余额扣减不被开启。

### F02 patch sites / 基座

backend/internal/pkg/apicompat/chatcompletions_to_responses.go 的 ChatCompletionsToResponses、chatMessageToResponsesItems、chatSystemToResponses；新增 convertChatToolChoiceToResponses 仅用 json/strings。

应达到：具名 function choice 转为 Responses 的顶层 name，developer 保持 developer。

边界：auto/none/required 及已经是 Responses 格式的 choice 不变；空名称和普通 system/user 消息保持原语义。

### F03 patch sites / 基座

backend/internal/service/gateway_forward_as_chat_completions.go 的 handleCCBufferedFromAnthropic、gateway_forward_as_responses.go 的 handleResponsesBufferedStreamingResponse；backend/internal/pkg/apicompat/types.go 的 AnthropicContentBlock.MarshalJSON；backend/internal/pkg/antigravity/{claude_types.go,stream_transformer.go}。#7789 的两条 CN native 文件本地不存在，不引入。

应达到：负数内容块索引被忽略；thinking 块即使内容/签名为空也有 thinking、signature 字段。

边界：普通文本/工具块不增加无关字段；正索引输出及后续 signature_delta 保真。

### F04 patch sites / 基座

backend/internal/service/gateway_claude_oauth_body.go 的 enforceCacheControlLimit、collectCacheControlPaths、normalizeClaudeOAuthRequestBody；#7957 仅取已有 Sonnet 5.5 的采样修复，不带 IsHaiku55 或新模型。新增 normalizeCacheControlTTLOrder 复用 gjson/sjson。

应达到：缓存块数量限制后将较早 ephemeral 断点升至后续 1h；Sonnet 仅有 top_p 时不注入 temperature。

边界：已合法 TTL 顺序/无断点请求不变；Opus/其它模型原规则和四块上限保留，不顺带开放 Haiku 5.5。

### F05 patch sites / 基座

backend/internal/service/openai_gateway_response_handling.go 的 handleStreamingResponseWithReasoning；backend/internal/pkg/xai/sso_device.go 的 convert/do。新增 ssoConsentToken 使用已存在的 golang.org/x/net/html，头只加首跳；不改 CLI 1.0.46 身份。

应达到：Grok 空 response.completed 按现有空流失败规则处理；SSO 同一 flow 从合法 consent 页读取 hidden token 并提交 origin/referer。

边界：有输出完成事件保持成功；旧页面没有 consent 字段仍兼容；重定向不泄露首跳额外头。

### F06 patch sites / 基座

frontend/src/components/keys/UseKeyModal.vue 的四种 Codex 配置生成分支；backend/internal/service/openai_codex_models_service.go 的 newConfiguredCodexModelDescriptor。escapeTomlBasicString、IsGPT61SolModelSpelling 已存在。

应达到：remote 模式的生成配置包含发现开关；生成的 GPT-6.1 Sol 默认上下文为来源值 500000。

边界：file 模式不加远程发现开关；显式上游 context_window 和用户目录值仍优先；不引入新模型/Ultrafast。

### F07 patch sites / 基座

backend/internal/service/gemini_messages_compat_service.go 的 convertClaudeMessagesToGeminiContents/extractClaudeContentText；backend/internal/service/model_rate_limit.go 的 modelRateLimitKeysForRequest/antigravityModelRateLimitKeys。新增 extractClaudeToolResultImageParts、isAntigravityGeminiImageModel 不依赖新配置 scope。

应达到：base64 工具结果图片转为 inlineData，工具响应顺序保持；图片模型冷却只写/查对应图片键。

边界：普通文本工具结果及 Gemini 文本家族冷却不变；不下载远程图片、不新增模型级规则 UI。

### F08 patch sites / 基座

frontend/src/utils/{pricing.ts,format.ts}；frontend/src/components/admin/usage/UsageTable.vue；frontend/src/components/admin/channel/PricingEntryCard.vue 的 onModelsUpdate；中英文 admin/resources.ts。#7720 需按 fork 当前 tooltip 块对齐。

应达到：科学计数价格与亚字节速率正确显示，零倍率保留，长上下文徽标不承诺统一倍数；只在空条目首次加入单模型时自动填价。

边界：已设价格、已有模型追加、一次多模型加入及缺失倍率默认1保持边界；不改变账本计算规则。

### F09 patch sites / 基座

frontend/src/components/common/{BaseDialog.vue,Select.vue,ProxySelector.vue}；已有共享 openDialogs/dialogId，0.2.9 S14 已补共享计数基座。

应达到：Esc 只关栈顶；disabled 收起并清理搜索；isComposing 时不处理导航键。

边界：保留原焦点归还/多弹窗 body 锁和非 IME 键盘操作；不重复移植旧 dialog ID 前置。

### F10 patch sites / 基座

frontend/src/components/admin/user/BulkEditUserModal.vue、frontend/src/composables/useGeminiOAuth.ts、frontend/src/components/admin/ErrorPassthroughRulesModal.vue；#8006 仅 frontend/src/components/account/ModelWhitelistSelector.vue 的 catch。extractApiErrorMessage 已存在，不引入 platformCatalog import。

实施补充（2026-10-14）：实测 `frontend/src/utils/apiError.ts:extractApiErrorMessage` 优先返回 Axios 通用 message 且未验证值为字符串，单独替换 catch 不满足F10。小范围适配该既有helper：后端 detail/message 优先，跳过结构化对象、空文本及仅含HTTP状态的通用错误，保留规范化错误和i18n映射；没有扩展支付专用解析逻辑或引入平台目录。

应达到：使用统一解析显示后端原因和合适 fallback；开关状态取服务端返回 enabled。

边界：不显示对象字符串或仅 Axios 通用文本；失败不伪造状态；不改接口和平台目录。

### F11 patch sites / 基座

frontend/src/components/auth/PendingOAuthCreateAccountForm.vue、components/user/profile/ProfileBalanceNotifyCard.vue、components/account/UsageProgressBar.vue、views/admin/BackupView.vue、composables/useBatchImageAccess.ts。只修已存在交互，通知组件按本地结构适配。

应达到：卸载后回包不重启计时器，回到可见标签立即更新倒计时；暂时查询失败后可再次检查访问权限。

边界：成功权限结果仍缓存，重复轮询不增生；不启用后台服务、余额通知或批量生图功能。

### F12 patch sites / 基座

backend/internal/service/channel_monitor_v2.go 的 ChannelMonitorV2HealthForWithThresholds；frontend/src/views/admin/ops/components/{OpsAlertRulesCard.vue,OpsDashboardHeader.vue,OpsSettingsDialog.vue}；frontend/src/views/user/ChannelStatusV1View.vue。formatDateTimeLocalInput 与 detail 刷新函数已存在。

应达到：请求样本不足保持未知；规则分钟为整数；时间按本地初始化；自动刷新与手动刷新使用同一 detail 流程；未完成设置加载禁止保存。

边界：达样本阈值正常评分；刷新保持用户当前筛选；不启用隐藏 ops 菜单或增加 TPS 字段。

### F13 patch sites / 基座

backend/internal/handler/admin/account_handler.go 的 GetAvailableModels；backend/internal/service/account_test_service.go 的 TestAccountConnection/sendError/RunTestBackground 与新增 account_test_logging.go；frontend/src/components/account/AccountGroupsCell.vue、components/common/GroupBadge.vue。logger.FromContext/IntoContext、logredact.RedactText、normalizeAccountTestMode 均在本地。

应达到：测试列表列显式请求侧映射名并排序；HTTP/后台失败日志有账号、测试及请求关联且先脱敏后截断；账号表完整显示长组名。

边界：无映射返回默认列表；普通 GroupBadge 默认仍截断；不通过日志泄露凭据，不更改测试执行计划。

### F14 patch sites / 基座

backend/internal/web/embed_on.go 的 shouldBypassEmbeddedFrontend；backend/internal/web/embed_test.go。仅核对并覆盖本地 gateway.go 已注册别名；#7982 同步来源测试资源选择，embed build 标签单独运行。

应达到：本地已注册 API 别名到达网关；测试使用当前构建实际静态资源。

边界：正常 SPA 深链接和静态资源保持；不注册上游缺失的视频/语音/Realtime API。

规格：[第一档](../../openspec/changes/archive/2026-10-24-port-upstream-0.2.16-tier1/README.md)。

## 4. 第二档：适配、状态隔离与维护

| 簇 | 来源 | A/C/X/N | 本地缺口 |
|---|---|---|---|
| S01 首次安装管理员校验统一 | [#7893](https://github.com/Wei-Shaw/sub2api/pull/7893) | 0/11/8/0 | 真实 SQLite 新安装允许三字节密码；Web/CLI 邮箱校验与登录不一致，校验上限128超过 bcrypt72。 |
| S02 旧版函数调用与拒答转换完整性 | [#7787](https://github.com/Wei-Shaw/sub2api/pull/7787)、[#7788](https://github.com/Wei-Shaw/sub2api/pull/7788) | 0/5/0/0 | legacy function_call 没有与结果稳定配对；Responses refusal 在流、非流和缓冲转换丢失。 |
| S03 显式工具变更 beta 保留及旧前置 | [#7827](https://github.com/Wei-Shaw/sub2api/pull/7827) | 0/1/2/0 | mimic 当前将 clientBeta 全部丢弃；上游补丁以已保留 structured-outputs 的分支为前置。 |
| S04 WebSocket 逐轮分组价格与图片用量 | [#7779](https://github.com/Wei-Shaw/sub2api/pull/7779)、[#7792](https://github.com/Wei-Shaw/sub2api/pull/7792) | 0/4/3/0 | 长连接 AfterTurn 一直用建连 Key 的 Group；relay 没有输入图片 tokens 字段及托管工具回填。 |
| S05 受约束的上游兼容重试 | [#7703](https://github.com/Wei-Shaw/sub2api/pull/7703)、[#7810](https://github.com/Wei-Shaw/sub2api/pull/7810) | 0/8/1/0 | 仅认识 invalid_encrypted_content；结构化 Responses input 被明确拒绝时，原始 Chat 没有安全回退。 |
| S06 OAuth 搜索历史回放声明 | [#7939](https://github.com/Wei-Shaw/sub2api/pull/7939) | 0/4/3/0 | map 与 raw/WS 的归一化未补历史 web_search_call 所需工具声明。 |
| S07 组合路由公开模型目录 | [#7793](https://github.com/Wei-Shaw/sub2api/pull/7793) | 0/3/2/0 | 组合目录仅聚合账号模型，精确路由公开别名遗漏。 |
| S08 删除分组同步清理风控范围 | [#7711](https://github.com/Wei-Shaw/sub2api/pull/7711) | 0/3/1/0 | 删除组不会清理内容审核 group_ids；配置保存与删除可写回失效 ID。 |
| S09 分组倍率/RPM 编辑异步隔离 | [#7904](https://github.com/Wei-Shaw/sub2api/pull/7904)、[#7900](https://github.com/Wei-Shaw/sub2api/pull/7900)、[#8002](https://github.com/Wei-Shaw/sub2api/pull/8002)、[#7972](https://github.com/Wei-Shaw/sub2api/pull/7972) | 0/6/2/0 | 切换组时旧结果可覆盖；仅 watch show；RPM parseInt 截断；倍率允许零/负/无穷。 |
| S10 异步页面、草稿、确认与弹窗隔离 | [#7903](https://github.com/Wei-Shaw/sub2api/pull/7903)、[#7818](https://github.com/Wei-Shaw/sub2api/pull/7818)、[#7833](https://github.com/Wei-Shaw/sub2api/pull/7833)、[#7841](https://github.com/Wei-Shaw/sub2api/pull/7841)、[#7840](https://github.com/Wei-Shaw/sub2api/pull/7840)、[#7732](https://github.com/Wei-Shaw/sub2api/pull/7732)、[#7765](https://github.com/Wei-Shaw/sub2api/pull/7765)、[#7839](https://github.com/Wei-Shaw/sub2api/pull/7839)、[#7724](https://github.com/Wei-Shaw/sub2api/pull/7724)、[#7722](https://github.com/Wei-Shaw/sub2api/pull/7722) | 0/21/1/0 | 页面/用户/计划切换后迟到回包覆盖当前状态；并发二次验证覆盖 resolver；公告失败仍关闭。 |
| S11 GPT-6.1 Sol 上游订阅长上下文统计 | [#8036](https://github.com/Wei-Shaw/sub2api/pull/8036) | 0/1/4/2 | 已有 OAuth GPT-6.1 Sol 仍套模型长上下文倍率，影响个人 token/cost 与账号配额统计。 |
| S12 Gemini 既有路径的错误、额度和脱敏修复 | [#8006](https://github.com/Wei-Shaw/sub2api/pull/8006) | 0/60/23/3 | 第三方 Gemini/Vertex 被映射 AI Studio 免费等级；第三方429套官方日限；不存在缓存响应会进入账号错误策略；Google格式丢 ops 分类；AQ密钥未脱敏；Flash兜底价为来源旧值。 |
| S13 监控配置补充已有平台入口 | [#7738](https://github.com/Wei-Shaw/sub2api/pull/7738) | 0/4/0/0 | 监控设置只能编辑已有平台行，缺失平台没有加入入口；来源导入的 platforms.ts 在本地缺失。 |
| S14 Responses 无 namespace 时避免重建 | [#7910](https://github.com/Wei-Shaw/sub2api/pull/7910) | 0/3/0/0 | 现有 stripOpenAIResponsesInputNamespaces 即使没有修改也遍历并构造临时数组，长历史增加分配。 |
| S15 Chat 流心跳保活 | [#7854](https://github.com/Wei-Shaw/sub2api/pull/7854) | 0/2/0/0 | Anthropic ping 在 Chat 转换路径被直接丢弃，长时间无语义输出时下游看不到保活。 |
| S16 前端依赖安全维护与旧待办归并 | [#7893](https://github.com/Wei-Shaw/sub2api/pull/7893) | 0/11/8/0 | 本次 pnpm audit 103条：DOMPurify19、axios12，并报告 Vue renderer/source-map-js；本地6条例外已过期。旧axios不再是无收益的纯升级。 |

### S01 patch sites / 基座

backend/internal/setup/{setup.go,handler.go,cli.go} 的 createAdminUser/validateEmail/validatePassword；新增 prepareAdminCredentials/generateAdminEmail/bootstrapAdminUser。适配 fork SQLite DSN、时间写入、初始并发及 install 通知；排查 deploy/config.personal.sqlite.yaml、SQLite compose、native 文档固定默认值，不整取 PG deploy 文件。

应达到：只有空库真正创建管理员时，缺省邮箱随机生成，显式邮箱可登录，密码8–72字节；CLI/Web/AUTO_SETUP 共用校验。

边界：已有管理员或已有普通用户时先跳过创建，不用遗留无效配置阻断启动，不重置密码；保留安装互斥和锁文件。

### S02 patch sites / 基座

backend/internal/pkg/apicompat/{chatcompletions_to_responses.go,responses_to_chatcompletions.go,types.go}；新增 legacyFunctionCallIDs 与 refusal 累积；连同所有协议消费者验收，接在 F02/F03 后。

应达到：旧函数调用按同名队列配对，稳定生成且避开已有 call ID；refusal 在 Chat message/delta/缓冲终态全部保留。

边界：混合现代 tool_calls、同名多次调用、不配对结果和纯拒答均有测试；不双计文本或工具调用。

### S03 patch sites / 基座

backend/internal/service/gateway_upstream_request.go 的 computeFinalAnthropicBeta、backend/internal/pkg/claude/constants.go；旧 v0.2.9 F05/#7638 中 structured-output 窄白名单前置随本簇补入；containsBetaToken、mergeAnthropicBetaDropping 已有。

应达到：只保留客户端显式请求的 structured-outputs、mid-conversation-tool-changes、inline-tools 兼容 token。

边界：effectiveDropSet 始终优先，未知 beta 不通过，不推断 companion token；count_tokens 原有不同规则保持。

### S04 patch sites / 基座

backend/internal/handler/openai_gateway_handler.go 的 ResponsesWebSocket BeforeTurn/AfterTurn、WithOpenAITurnPricingContext；backend/internal/service/openai_ws_v2/passthrough_relay.go 的 parseUsageAndAccumulate/mergeRelayUsageNonZero/finalizeRelayTurnUsage 与 passthrough adapter。APIKeyService.GetByKey 已存在；新增 openAIWSTurnBillingAPIKeys/refreshOpenAIWSTurnBillingAPIKey 同消费者加入。

应达到：第二轮起获取同 Key 同组的新价格快照并绑定本轮；输入/输出图片 tokens 从 usage 或 tool_usage 回填并准确累计。

边界：不更新连接用户/订阅/Key 限额身份；换组/换平台/查询失败回退建连快照；交错轮次不串价、不重复累计。

### S05 patch sites / 基座

backend/internal/service/openai_gateway_forward.go、openai_gateway_{cc_pipeline,chat_completions,chat_completions_raw,responses_chat_fallback}.go；现有 readOpenAIUpstreamError、forwardAsRawChatCompletions/forwardResponsesViaRawChatCompletions 与 openai_compat 支持模式已存在。移植完整错误体识别、新 startTime 签名及所有调用，不引入 provider_profile。

应达到：特定 thinking_signature_invalid 三段诊断匹配既有一次性恢复；auto 且已知支持 Responses 的 API Key 对完整、无歧义 input 类型拒绝回退原始 Chat，并沿用起始时间。

边界：Responses 形状入站、已输出/已取消、截断或读失败、重复字段、含用量/输出及其它错误不得触发危险重试；不降级强制模式。

### S06 patch sites / 基座

backend/internal/service/openai_codex_transform.go、openai_gateway_{forward,passthrough,request_body}.go 与新 openai_oauth_web_search_history.go（264行）；firstNonEmptyString、marshalOpenAIUpstreamJSON、isOpenAIResponsesCompactPath 已有。

应达到：OAuth 回放搜索历史时补 cached-only web_search，Lite 放 additional_tools；原本无工具时保持 tool_choice none，compaction_trigger 仍最后。

边界：已有工具声明不重复；保留显式强制选择；API Key 及 /responses/compact 不套此重写；map/raw/WS 一致且保留大整数。

### S07 patch sites / 基座

backend/internal/service/composite_route_resolver.go 的 CompositeRouteResolver/repo.ListByGroup；gateway_service.go 新 GetCompositeRouteModels；backend/internal/handler/gateway_handler.go 的 compositeAvailableModels/codexModelIDsForGroup。只取启用 exact 路由列表和 endpoint 筛选，沿用 Group.CustomModelsListEnabled；不引入 TypeSafe、ModelAllowlist 或 pinned 目录。

应达到：启用精确路由的公开别名加入对应 /models 与 Codex Responses 目录，去重并保持 endpoint 区分。

边界：禁用/glob 路由不凭空展开；读取失败回退现有目录；原自定义列表与组隔离不放宽。

### S08 patch sites / 基座

backend/internal/repository/{group_repo.go,setting_repo.go} 的 deleteCascade/Set，backend/internal/service/content_moderation.go 的 UpdateConfig。上游 lockLiveGroups 本地缺失且 PG FOR UPDATE/jsonb 聚合不可移植；以 SQLite 写事务内校验有效组并 JSON 读改写，保留畸形历史值，不新增 schema。

应达到：组删除与审核范围更新在同一事务；并发配置保存不可重新写入已删除组。

边界：缺失/畸形设置不阻断删除也不覆盖原值；真正 DB 错误回滚；保留顺序与其它配置字段。

### S09 patch sites / 基座

frontend/src/components/admin/group/{GroupRateMultipliersModal.vue,GroupRPMOverridesModal.vue}；本地既有分页/快捷编辑和 API 保留；按目标 tag 终态整合两个 loadVersion 实现，不取 #7618 平台目录改造。

应达到：打开/切组/关闭/卸载使旧加载失效并清空旧行；RPM 为非负整数，倍率为有限正数。

边界：迟到 success/error/finally 均不改当前组或 loading；不把不合法输入保存，不误合旧 F12 未完成的其它表单。

### S10 patch sites / 基座

frontend/src/components/user/profile/{ProfileEditForm.vue,ProfileIdentityBindingsSection.vue}；views/user/CustomPageView.vue；views/admin/ops/components/OpsAlertEventsCard.vue；components/admin/{monitor/MonitorTemplateManagerDialog.vue,account/ScheduledTestsPanel.vue,usage/UsageCleanupDialog.vue}；composables/useStepUp.ts；stores/{adminCompliance.ts,announcements.ts}；components/common/AnnouncementBell.vue。

应达到：每个视图按对象/请求代次接收响应，保留未提交草稿；同实例二次验证共享待定 Promise；公告仅成功标记后关闭。

边界：错误/finally/nextTick/分页回包也受失效保护；换用户时清理敏感草稿；不启用仍被 SQLite 跳过的清理或定时测试 worker。

### S11 patch sites / 基座

backend/internal/service/openai_gateway_usage.go 的 RecordUsage/calculateOpenAIRecordUsageCost/TokenCost；billing_service.go 的 CostInput/CalculateCostUnified/calculateTokenCost。IsOpenAIOAuthLike、openai.IsGPT61SolModelSpelling 已有；按本地较短调用签名适配，不搬 absent CN/image 测试基座。

应达到：实际凭据为 OpenAI OAuth-like 且最终上游模型为 GPT-6.1 Sol 时豁免模型价卡长上下文附加费，用户与账号成本口径一致。

边界：API Key、其它模型仍原价卡；显式自定义区间价格不豁免；不引入付费订阅购买/续费或新价格体系。

### S12 patch sites / 基座

#8006 仅选 backend/internal/service/{account.go,gemini_quota.go,gemini_messages_compat_service.go,ops_upstream_context.go,billing_service.go}、backend/internal/handler/{gemini_v1beta_handler.go,ops_error_logger.go,security_audit_errors.go}、backend/internal/util/logredact/redact.go、模型价格资源及既有 Key 文案。isOfficialGeminiAPIBaseURL 在新缓存文件里但仅16行/stdlib，连 host 常量抽到已有 Gemini 公共文件；缓存不存在固定响应为1常量，错误判定19行。不引入缓存CRUD、binding、租约、存储计费、inflight或Wire变动。

应达到：第三方/Vertex 不套本地模拟日配额，第三方429仅显式重试时间或原全局fallback；缺失缓存在错误策略前识别并返回请求错误，不封号换号；本地Google错误正确归类；AQ凭据文本脱敏。

边界：官方AI Studio/OAuth原额度保持；非缓存鉴权错误仍正常处理；不实现缓存创建/续期/删除。Flash促销价仅按来源2026-12-31有效窗口核对，过期需恢复标准兜底，不默许永久优惠价。

### S13 patch sites / 基座

frontend/src/features/channel-monitor-v2/MonitorSettingsPanel.vue；复用本地已支持 Anthropic/Gemini/OpenAI/Antigravity/Grok 五个平台值和现有 Select，或抽同值小常量；不为 CONCRETE_PLATFORM_OPTIONS 引入 #7618 平台清单。

应达到：用户可加入一个当前缺失且本地后端支持的平台，初始无模型，沿原保存接口持久化。

边界：不重复添加、不加入未实现平台，saving 时禁止修改；不是新增供应商平台。

### S14 patch sites / 基座

backend/internal/service/openai_responses_namespace.go 的 stripOpenAIResponsesInputNamespaces；gjson.Result.Index 与 input.Raw 的相对偏移按来源完整实现；不改 namespace 策略。

应达到：无修改时返回原正文；需要剥离时仅首次分配并复制未变字节段，减少单节点长历史开销。

边界：保留工具调用 namespace 的既有规则不变；空白/数字/转义/非对象和错误回退保真，需来源 benchmark 或 AllocsPerRun 支撑收益。

### S15 patch sites / 基座

backend/internal/service/gateway_forward_as_chat_completions.go 的 handleCCStreamingFromAnthropic；将上游 ping 转为 SSE 注释并 Flush。这里会写出 HTTP 响应，必须联合既有终态/失败/客户端断开规则验证，不能因 CLEAN 判低风险。

应达到：仅 ping 到来时向 Chat 客户端写标准 SSE 注释保活；后续语义事件正常转换。

边界：不泄露 event:ping，写失败停止；错误事件不得吞掉；心跳后失败按已经写出响应的规则处理，不重复响应或危险重放。

### S16 patch sites / 基座

frontend/package.json、frontend/pnpm-workspace.yaml、frontend/pnpm-lock.yaml 和 .github/audit-exceptions.yml；#7893 仅Vue/source-map-js维护，追加冻结旧#7780的axios子项及4a1da2950的DOMPurify来源。pnpm11读取workspace overrides，不能只改package.json。当前registry要求DOMPurify至少3.4.16，不能沿用旧3.4.14目标；axios至少1.20.0，Vue至少3.5.42/source-map-js至少1.2.2，实施时复核。

应达到：定向更新已有DOMPurify、axios、Vue及source-map-js受影响解析，保留pnpm锁文件一致性；sanitize调用、API错误处理与前端构建回归通过；审计重新记录目标包和剩余项。

边界：不宣称103条全部已解决；xlsx无npm修复路径及Vite/Vitest/tinypool工具链大版本另列旧维护范围。例外只能按实际不可利用条件和负责人复核，不自动续期、不扩大支付功能。

规格：[第二档](../../openspec/changes/archive/2026-10-24-port-upstream-0.2.16-tier2/README.md)。

## 5. 按需、不做与历史遗留

### 5.1 本轮第三/四档

**第三档**

- PR7893：上游 xlsx 审计例外延期不直接采用；本轮实际audit后Vue/source-map-js归入S16，xlsx与工具链专项仍待单独评估。
- PR7853：账号筛选紧凑布局：新增便利交互；用户本轮选择暂不增加。
- PR7911：输出 TPS 列与运维分位数：新增统计展示，含需 SQLite 重写的 PG percentile SQL；用户暂不增加。
- PR7933：运维请求详情账号名称列：新增展示；用户暂不增加。
- PR7960：分组多语言：46文件、Ent/schema/JSONB/DTO/UI；用户暂不增加。支付 handler/套餐部分始终第四档。
- PR7957：Haiku 5.5 新模型/桥接/价卡：21文件完整来源；用户暂不增加；仅已有 Sonnet 采样修复进入F04。
- PR8006：显式缓存CRUD/存储费/绑定/租约及 Vertex API Key/模型列表新能力：86文件混合来源，用户暂不增加；仅现有路径修复进入F10/S12。
- PR7828：新增临时不可调度规则 scope/模型401行为；本地没有 scope 配置，不能把该新增能力的修复当现有漏洞。用户暂不增加。

**第四档**

- PR7814：上游 SECURITY policy/举报渠道/维护版本声明是上游项目政策，不直接复制为 fork 承诺。
- PR7813：TypeSafe 平台缺失，修复的默认探测拒绝路径不存在。
- PR7888：EasyPay 支付回调防伪：payment-only，个人范围明确排除，即使全CLEAN也不合。
- PR7811：本地能力字段不含 service_tiers，mergeMissingCodexModelFields 已补 null 为 []；探针确认该路径无此缺陷。
- PR7891：OpenCode Go usage/403/Retry-After 基座不存在。
- PR7889：OpenCode Zen/Go 原生路由与模型拒绝基座不存在。
- PR7877：本地 providerAdapters 未实现智谱；不是现有监控URL缺陷。
- PR7782：本地 OpenAI Codex handler 对显式模型映射先生成本地目录，已保留映射策略；pinned/投影链缺失，来源补丁不适用。
- PR7901：Airwallex 支付导航回调清理：payment-only。
- PR7819：Stripe 支付卸载回调清理：payment-only。
- PR7764：PluginsView 和插件宿主不存在；不引入平台以修会话竞态。
- PR7618：134文件/+13053/-1052的平台清单、Cline/Command Code/CN多协议大改；新增catalog/profile、model-protocol cache及native Anthropic链均缺失。独立trim模型修复的目标native消费者也不存在，不导入整套作为前置。
- PR7717：支付二维码/状态倒计时：payment-only；不因通用timer思路携带。
- PR7649：所修非OpenAI CN原生Responses平台链缺失；Grok在桥接前独立分流，WS resolver仅OpenAI。不能因helper能插一行就宣称修现有bug。
- PR7991：Kimi国际Coding quota GetCodingPlanProvider与CN probe消费者缺失；不为了端点白名单新增平台。
- PR7960：支付handler/套餐i18n即使以后需要分组多语言也不引入。
- PR7789：两条CN native Anthropic缓冲路径本地缺失，只修本地两条既有消费者。
- PR7822：不新增本地未注册的媒体/Realtime路由。
- 458b92abd：上游VERSION同步，保留fork 1.1.16。
- b8dece900：上游VERSION同步，保留fork 1.1.16。
- 3f1a2ea0a：上游VERSION同步，保留fork 1.1.16。
- 3a6fd1c9d：上游VERSION同步，保留fork 1.1.16。

### 5.2 旧轮次未合项与本轮量测

旧冻结报告不改写。本轮仍是**评估**：以下“归入”表示计划接管，不能算已经落地。

| 轮次 | 当前未合/已合边界与原因 |
|---|---|
| 0.2.12 | 第一档3簇/第二档3簇在405246789完成；优先级快捷调整、Key组排序及原子邮箱缓存保留。TypeSafe缺平台、支付排除。**axios原第三档因本次audit的新证据调整为S16第二档**，不是因补丁便宜顺带升级。 |
| 0.2.11 | 第一档4簇/第二档6簇在96f32d018完成。Claude原生额度查询/重置、风控白名单、趋势切换、Key创建限制、余额在途预占、Astra Ultrafast未实施；取决于用途，用户本轮暂不加能力。原生上游额度元数据不等于支付。 |
| 0.2.9 第一档 | 原39PR/15簇只有F04 thinking-disabled已在0.2.11完成，dialog共享ID/空Tab前置此前完成。其余RPM投影、OAuth重授权、账号身份/错误状态、beta、Responses消息/终态、Alpha search、未来reset暂停、TOTP、公共表单、代理、配置输入、异步弹窗、价格显示、ID/配额诊断仍未整批完成。**本轮S03接管F05中#7638；F09/S09只与部分旧公共表单交叉，不把旧簇全算完成。** |
| 0.2.9 第二档 | 48PR/16簇已在159d04222完成。保留现有目录、工具schema/PDF、流结束取消、调度、图片统计、Codex配置等适配；本轮以当前实现核对重叠。 |
| 0.2.9 按需/排除 | OpenCode/Command Code/Kimi原生、插件、model_allowlist glob/pinned/live、Free Fast/effort倍率/自动用卡、视频图像新路由、TypeSafe/月备份/部署运营增强未合。支付/返利排除；当前白名单编辑器和remote/file配置不等于完整pinned目录。 |
| 0.2.5 主体 | 已落bfbcd79；WS系数#7043仍默认1.0而非5.0，#6954/S12b平台配额清理及旧预留227未实施；#6960/#7049完整exec_scope/抢占/queue-wake未整体合。 |
| 0.2.5 平台/目录 | #6575/#6890/#6974/#6769 Gemini3.7/3.8、原生Codex Images、DeepSeek峰谷价、Ollama异步重置；#6747/#7091/#6691/#7153平台/订阅批量/Key批量过滤；#6989/#7057/#6968/#7011/#6945/#6988目录/上下文/视觉/compact/pinned仍按用途。收费开关#6986排除。 |
| 0.2.5 其它 | #7129/#7085/#6929/#6917 Grok序号、input stripper、Gemini带内错误、监控BaseURL；#6847/#7001/#6845/#6913/#6915管理展示仍独立未合。#7126/#7110/#7148的已有namespace语义/撤回/PG不重做。 |
| 0.2.4/0.2.0主体 | 已落a26b6b55d/c9d9bebe8。**0.2.0第二档阶段9 DOMPurify未实施，本轮S16接管维护，旧checkbox不动**。旧3.4.14目标已不足本次audit，不能照搬历史18条数量。 |
| 0.1.180/0.1.184长上下文/Fast | 动态阶梯、effort、Fast/Free Fast/Ultrafast完整方案未合。当前GPT-6.1已有静态价卡，S11只豁免该上游OAuth订阅附加费，不引入旧大价格体系。 |
| 0.1.180工具桥/Grok/WS | 非法参数、namespace别名、deferred尾项仍未整体合；本轮S02/S06/S14不等于完成旧工具桥。Grok冷却/Realtime/抢占与跨节点Redis租约按用途/单节点架构排除或暂缓。 |
| 0.1.184其它 | usage compaction/requested effort迁移、公共分组限制、订阅重置锚点、峰谷价/图像冷却仍独立；历史schema不改。自动重置卡/模型读取上限未采用；本地128c5f7bf固定账号quota窗口推进是另一个已完成修复。 |
| 更早已落 | 0.1.176/179/183主体、0.1.184已验收主体、API Key instructions/routed catalog/Responses Lite按现有实现保留，不重新立项。旧695ebede7通用14行usage字段和0.2.9 F04 helper已在0.2.11补入。 |
| SQLite后台 | UsageCleanupService、AccountExpiryService、ScheduledTestRunnerService仍被skipSQLiteBackgroundJobs控制；必须逐SQL审计后再启用。S10只修前端过期结果，不把启用worker当隐含任务。 |

本轮实际复测：7处DOMPurify调用（含公开LegalDocumentView）；xlsx仍只发现懒加载XLSX.write导出，无生产XLSX.read命中，但不据此替用户续期例外。实际audit记录103条（3 critical/42 high/50 moderate/8 low；包含同advisory的不同版本分支，不等于103个可利用漏洞）；DOMPurify19条、axios12条，另有Vue renderer/source-map-js等。6条例外均在2026-10-06到期，前端测试通过不能替代audit通过。[审计总结](./evidence-0.2.16/dependency-review.md)。

### 5.3 缺基座重新量测

- #7782完整前置投影函数约百行，但本地已有“显式映射→本地生成目录”的有效路径；不是因行数大而放弃，不为此引入pinned/model_allowlist。
- #7827需要的旧#7638只有2行beta常量和11增/6删的请求文件（其中有注释），来源测试74行；本轮明确与S03一起补，旧F05其它子项仍待做。
- #8006中官方Gemini URL判定仅16行，加host常量即可供第三方额度判定；不存在缓存的判定19行/固定响应1常量。**不以这些小前置为由引入约8450新增行的缓存系统**。
- #8036缺的accountStatsLongContextPricingEnabled仅3行，等价于本地现成`gate == nil || *gate`表达式，直接接豁免条件即可；不搬整份账号统计重构。
- #7711的lockLiveGroups为PG的30行函数，含ANY/pq.Array/FOR SHARE；SQLite需要事务内存在性检查与JSON更新，不能“补同名helper”后照跑。
- #7738缺平台常量可以用本地五个已实现值解决，不需要134文件的#7618清单重构。

除这些明确重测点，其余历史项依据落地提交和当前符号/配置复核；本轮没有逐条重跑所有历史PR。

## 6. ALREADY、N/A与本地替代实现

- #7811不进入实施：本地codexToolCapabilityFields不收service_tiers；生成默认[]，API Key元数据补全会将null替换为[]，探针确认。保留原始上游manifest的路径仍按既有契约，不误称全站所有外部正文都会强制清洗。
- #7782不进入实施：openai_codex_models_handler.go先BuildGroupConfiguredCodexModelsManifest，显式映射走本地生成目录；来源pinned/remote投影链不是本地实现。CONFLICT/NOBASE在这里不等于缺陷待修。
- #7649非OpenAI CN原生链缺失；Grok在Forward桥接前return，WS resolver拒绝非OpenAI。单行守卫可套不代表有实际消费者收益。
- #7789只采用本地两条缓冲消费者；缺失的两条native路径、#7877智谱监控、TypeSafe/OpenCode/Cline/Command Code/插件不补平台。
- 本地128c5f7bf已推进固定配额窗口；它与F01删除Key的事务问题不同，后续必须保留并复跑SQLiteQuotaWindows/FixedQuotaResetAtWeeklyDefault。

## 7. 建议落地顺序

1. 开实施分支，记录实际HEAD并复核冻结基线。F01与S01优先，保留当前SQLite事务/DSN与管理员跳过创建语义。
2. F02/F03→S02，F04→S03；旧#7638只作S03行为前置，新增helper随消费者落。
3. F05/F07/S05/S06/S15按HTTP/WS/流式出口逐簇验收；共享网关文件串行适配，不凭CLEAN批量套patch。
4. S04/S11联合用量统计和SQLite dedup回归；S08真实SQLite删除/保存交错验证；S07保留本地目录策略。
5. S12逐一摘取现有Gemini错误/额度/脱敏/价格行为，禁止缓存管理、inflight和Wire作为前置；第三方429与官方日限分开验证。
6. F06/F08–F14及S09/S10/S13处理前端，复核复用行/卸载/旧响应归属；S14用性能数据证明分配收益。
7. S16独立依赖维护提交，重新audit并记录剩余风险；不能直接把103条告警转成一份全量上游lockfile，也不能自动延长例外。
8. 实施完成再运行backend unit/build、相关race、SQLite方言/真实usage dedup、frontend typecheck/lint/build/相关vitest，回填两个verification。本次不实施、不改VERSION、不提交、不发布或部署。

## 8. 基线自测与实施门禁

[基线报告](./evidence-0.2.16/baseline-tests.md)、[Go测试枚举](./evidence-0.2.16/backend-test-inventory.txt)、[197条来源测试文件记录](./evidence-0.2.16/source-test-inventory.tsv)。所有下列通过只证明当前基线；新增来源用例没有落地。

- apicompat/antigravity/claude/xai/setup包unit通过；repository/service/handler/admin受影响定向用例通过。relay最初宽泛regex没有命中，另跑完整openai_ws_v2包通过，没有把“no tests to run”当覆盖。
- 真实SQLite现有quota-window/all-effects/dedup与方言检查通过；overlay复现删除Key回滚、安装校验缺口、Gemini第三方误套免费等级/图像家族冷却，并证明Codex null补齐。探针断言旧行为，不能作实施后通过门禁。
- 前端22个不同测试文件，**242项通过**（13文件185项 + 相邻9文件57项），typecheck通过；未运行全量frontend、lint/build或全量backend/race。
- **embed基线2个静态资源子用例失败**：请求/logo.png但构建资源只有/logo.svg。只在临时overlay按#7982改测试路径/类型后整包通过，证明是过期fixture；真实文件仍未修。
- `pnpm audit`非零：103条告警，详细日期/受影响路径见dependency-review；这是已知未通过门禁，不能写“全部检查通过”。
- `_integration_test.go`先看首行：PG-only的usage_billing/group等来源不作为SQLite门禁、不改标签骗过编译。Redis-only integration也不是postgres-only；本轮未启动外部集成服务。

OpenSpec全部实施checkbox保持未勾选、verification证据槽留空；已完成的评估证据只放本目录。

## 9. 新增通用教训

1. 功能名相同也可能调用链不同：本地“映射生成目录”与上游“远程投影目录”不能直接对照文件补丁。
2. 新功能大PR内的小修要量最小独立基座；16行URL判定不需要整个缓存资源/租约/存储计费系统。
3. 修复范围按实际消费者收敛：缺平台的native索引/图像桥接问题不能因helper可应用列成现存缺陷。
4. 保活注释会写出HTTP响应；即使补丁CLEAN，也要验证心跳后失败和重试的语义边界。
5. 基线测试失败可对应过期fixture；临时overlay可验证原因，但不应把诊断性通过写成代码已修或基线全绿。
6. 每轮依赖audit必须重跑并更新目标版本；过期例外不能照搬延期，pnpm11的workspace overrides也必须纳入计划。
7. 用户“暂不增加功能”不排除现有调用路径的可靠性修复；将混合PR按行为拆开并明示分档变化。
