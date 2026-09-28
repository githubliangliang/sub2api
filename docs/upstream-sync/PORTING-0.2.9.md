# 上游 v0.2.9 移植评估（v0.2.5 → v0.2.9）

评估日期：2026-09-28。**建议分批移植第一档 39 个 PR、第二档 48 个 PR；不做整仓 merge。**
本轮只生成评估与两份 OpenSpec，产品补丁尚未实施。前端基线通过；后端缺少 Go，编译/运行验收待补，不能据此直接发布。

## 1. 版本、范围与规模

| 基线 | 值 |
|---|---|
| 本 fork | main / v1.1.13，`bfbcd79bdc4aa0838616405466750658896e6920` |
| VERSION 文件 | `1.1.12`；与 tag 不同，后续发布前需按自有版本流程核对；本次未改 |
| 上次已评估上游 | v0.2.5，`86f93c28ee34cc74b629dafb748bd5ac5ca8c5ea` |
| 最新稳定发布 | [v0.2.9](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.9)，2026-09-28 11:08:59 北京时间 |
| 现有数据库 | SQLite；最大迁移226；Redis可关闭，miniredis单进程；个人项目不引入支付 |

| 新增上游 tag | Commit SHA | 相对前一表列基线的非 merge 提交 | PR 数 |
|---|---|---:|---:|
| v0.2.7 | `aea725f2ea644d5592d0bbb1d63b607efa7e200a` | 38 | 33 |
| v0.2.8 | `fd80b08c90b55edcad5b00171b53f08721d30da1` | 140 | 97 |
| v0.2.9 | `4c00df2e0183e2c70b7fa8ba45914205e36aad0c` | 37 | 31 |

[v0.2.7](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.7) 发布于09-19；[v0.2.8](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.8) 发布于09-23。
本次远端 `v0.2.*` tag 枚举没有 v0.2.6；覆盖按 Git 祖先区间计算，不依赖 Release 列表连续性。
合计215条非 merge提交、161个PR、9个主线直接提交，tag间净差617文件 / +38,781 / -2,647。
逐PR文件会重复出现，867条文件证据不是867个独立文件。

| 分档 | 候选数 | 处理 |
|---|---:|---|
| 第一档 | 39 | 15簇，确认旧逻辑仍在，改动较小；生成 tier1 规格 |
| 第二档 | 48 | 16簇，有价值但需手工适配/多路径验证；生成 tier2 规格 |
| 第三档 | 48 | 按实际平台/管理/部署用途选择；含1个随功能的直接测试提交 |
| 第四档 | 35 | 不适用、已满足、支付/裁剪入口、上游维护提交；不实施 |

## 2. 评估方法与边界

- 只读抓取上游 tag 到 `refs/remotes/upstream-review/v0.2.*`，保留 origin 配置和本 fork 自有 tag。
- 本地 `git merge-base HEAD upstream-review/v0.2.9` 无共同祖先；`HEAD..upstream` 不能当缺失列表。
- 按主线 PR 合并结果 `git diff <merge>^1 <merge>` 整体取证；直接提交单列，避免漏掉 CI/测试修正。
- 独立 `GIT_INDEX_FILE` 先 `git read-tree bfbcd79`，全候选先反向 `git apply --cached --reverse --check`，再正向 `--check`。两者失败且 HEAD 无文件才记 NOBASE。
- **ALREADY 8 / CLEAN 559 / CONFLICT 228 / NOBASE 72**；都是固定基线的文件态，不是编译结论、功能完成数或可累加工作量。
- 新文件 CLEAN 也可能依赖缺失符号；例 #7571 缺约32行 helper，#7313/#7224 引用了未定义的原生平台常量。缺失测试文件不等于缺产品能力。
- 同一簇取 v0.2.9 终态；跨簇共享 service、handler、schema cleaner、UseKeyModal 和 SettingsView，串行落地后复验。
- 本轮建议的第一/二档不要求新迁移。上游238b（TypeSafe）、239（effort倍率）、240（affiliate）不进入本轮。以后引入 schema 时从当时本地最大号之后新建 SQLite 迁移，不能占用或修改已应用226；旧0.2.5的227只是未实施预留。

原始证据：[commits.tsv](./evidence-0.2.9/commits.tsv)、[candidates.tsv](./evidence-0.2.9/candidates.tsv)、[files.tsv](./evidence-0.2.9/files.tsv)。
当前分档及逐候选理由：[decisions.tsv](./evidence-0.2.9/decisions.tsv)。旧轮次冻结 TSV/source-baseline 未改。
关键消费者/缺失helper的本地定义核对：[base-symbols.tsv](./evidence-0.2.9/base-symbols.tsv)。源码存在性不能替代Go类型检查。

## 3. 第一档：现有缺陷的小范围修复

A/C/X/N 依次表示 ALREADY/CLEAN/CONFLICT/NOBASE，统计含来源 PR 测试和应排除的邻近 hunk。
下列patch sites按来源文件定位；仅核实过的函数标行号。新增文件和缺失平台文件用于评估依赖，不能据清单整文件复制。

| 簇 | 来源 | A/C/X/N | 要达到的行为 |
|---|---|---|---|
| F01 调度 RPM 配置投影 | [#7311](https://github.com/Wei-Shaw/sub2api/pull/7311) | 0/1/1/0 | 账号 base_rpm、rpm_strategy、rpm_sticky_buffer 经过缓存投影后仍参与候选限流 |
| F02 OAuth 重授权保留账号设置 | [#7404](https://github.com/Wei-Shaw/sub2api/pull/7404) | 0/2/0/0 | 更新认证凭据时保留原 model_mapping 等非认证设置，新认证字段覆盖旧值 |
| F03 账号身份与错误状态 | [#7349](https://github.com/Wei-Shaw/sub2api/pull/7349)、[#7362](https://github.com/Wei-Shaw/sub2api/pull/7362) | 0/4/0/0 | 非法 UA 回退时保留已解析版本；读到用量快照不能抹掉 refresh token 拒绝错误 |
| F04 关闭 thinking 的桥接语义 | [#7538](https://github.com/Wei-Shaw/sub2api/pull/7538) | 0/7/0/0 | Anthropic thinking.type=disabled 优先于 output_config.effort，两条 OpenAI 桥均输出 none |
| F05 显式 beta 头兼容 | [#7638](https://github.com/Wei-Shaw/sub2api/pull/7638)、[#7617](https://github.com/Wei-Shaw/sub2api/pull/7617) | 1/5/0/0 | mimic 路径保留显式 structured-outputs beta；OpenAI 出站保留多智能体 beta 并剔除旧 Responses token |
| F06 Responses 消息与最终正文 | [#7635](https://github.com/Wei-Shaw/sub2api/pull/7635)、[#7569](https://github.com/Wei-Shaw/sub2api/pull/7569) | 0/5/0/0 | 角色消息带 type=message；终止事件空正文时回填已积累的文本 |
| F07 Alpha search 成功判定 | [#7572](https://github.com/Wei-Shaw/sub2api/pull/7572) | 0/3/0/0 | 仅在 response.completed 且状态成功后返回可计费用量 |
| F08 未来重置时间前保持配额暂停 | [#7624](https://github.com/Wei-Shaw/sub2api/pull/7624) | 0/4/0/0 | 陈旧但有明确未来 reset 的额度快照继续用于暂停，达到 reset 后放行 |
| F09 TOTP 输入与计时器 | [#7235](https://github.com/Wei-Shaw/sub2api/pull/7235)、[#7184](https://github.com/Wei-Shaw/sub2api/pull/7184)、[#7496](https://github.com/Wei-Shaw/sub2api/pull/7496) | 0/8/0/0 | 验证码输入同步、错误文本可读，弹窗销毁后迟到响应不重启计时器 |
| F10 公共表单交互 | [#7234](https://github.com/Wei-Shaw/sub2api/pull/7234)、[#7263](https://github.com/Wei-Shaw/sub2api/pull/7263)、[#7186](https://github.com/Wei-Shaw/sub2api/pull/7186)、[#7238](https://github.com/Wei-Shaw/sub2api/pull/7238)、[#7374](https://github.com/Wei-Shaw/sub2api/pull/7374)、[#7142](https://github.com/Wei-Shaw/sub2api/pull/7142)、[#7109](https://github.com/Wei-Shaw/sub2api/pull/7109)、[#7108](https://github.com/Wei-Shaw/sub2api/pull/7108)、[#7499](https://github.com/Wei-Shaw/sub2api/pull/7499) | 0/18/0/0 | 分页跳转、对话框标题 ID、复制失败、标签 Tab、IME、下拉键盘与搜索、日期关闭和跨午夜选择保持一致 |
| F11 代理管理小修 | [#7183](https://github.com/Wei-Shaw/sub2api/pull/7183)、[#7321](https://github.com/Wei-Shaw/sub2api/pull/7321) | 0/4/0/0 | 同代理重复测试共享进行中的结果，过期边界展示与后台一致 |
| F12 配置表单错误与输入边界 | [#7372](https://github.com/Wei-Shaw/sub2api/pull/7372)、[#7264](https://github.com/Wei-Shaw/sub2api/pull/7264)、[#7364](https://github.com/Wei-Shaw/sub2api/pull/7364)、[#7376](https://github.com/Wei-Shaw/sub2api/pull/7376)、[#7367](https://github.com/Wei-Shaw/sub2api/pull/7367) | 0/10/0/0 | 设置和 Antigravity 映射加载失败可重试，替换分组显示错误，倍率必须正数，RPM override 只接受非负整数 |
| F13 异步弹窗状态隔离 | [#7420](https://github.com/Wei-Shaw/sub2api/pull/7420)、[#7422](https://github.com/Wei-Shaw/sub2api/pull/7422)、[#7493](https://github.com/Wei-Shaw/sub2api/pull/7493)、[#7497](https://github.com/Wei-Shaw/sub2api/pull/7497) | 0/9/0/0 | 错误详情、用户 API Key、临时不可调度状态仅接受当前对象请求，分组弹窗释放监听和待执行搜索 |
| F14 用量与价格显示 | [#7446](https://github.com/Wei-Shaw/sub2api/pull/7446)、[#7418](https://github.com/Wei-Shaw/sub2api/pull/7418)、[#7611](https://github.com/Wei-Shaw/sub2api/pull/7611) | 0/6/0/0 | CSV 保留缺失推理力度的单独减号占位符；区间输入正确解析科学计数法；空闲窗口有 reset 时显示倒计时 |
| F15 请求 ID 和配额诊断 | [#7488](https://github.com/Wei-Shaw/sub2api/pull/7488)、[#6920](https://github.com/Wei-Shaw/sub2api/pull/6920) | 0/2/1/1 | 已识别的 Responses input item ID 超过64字节时移除；Grok 冷却期仍能查询配额 |

第一档行为规格：[port-upstream-0.2.9-tier1](../../openspec/changes/port-upstream-0.2.9-tier1/README.md)。

### F01：调度 RPM 配置投影

只补现有 Extra 白名单；适配本地 scheduler_cache_test.go，不整取缺少的上游测试基座。反例/边界：tiered 红区拒绝、sticky_exempt 仍仅允许粘性；未配置账号保持现状。

- `backend/internal/repository/scheduler_cache.go`:980 `filterSchedulerExtra`
### F02：OAuth 重授权保留账号设置

复用 oauth_refresh_api.go 的 MergeCredentials；不新增平台。反例/边界：新旧字段冲突、显式空值遵守已有 MergeCredentials 语义。

- `backend/internal/handler/admin/account_handler.go`:1460 `ApplyOAuthCredentials`
### F03：账号身份与错误状态

沿现有身份配对和账号用量入口移植。反例/边界：合法 UA 保持身份；真实恢复通过既有恢复入口清错。

- `backend/internal/service/account_usage_service.go`
- `backend/internal/service/setting_gateway_runtime.go`
### F04：关闭 thinking 的桥接语义

新增共用 helper 随两条消费者一起落地。反例/边界：未禁用时保持 medium 默认和 max→xhigh；none 不请求 reasoning summary。

- `backend/internal/pkg/apicompat/anthropic_to_responses.go`
- `backend/internal/pkg/apicompat/chatcompletions_anthropic_bridge.go`
- `backend/internal/service/openai_compat_model.go`
### F05：显式 beta 头兼容

OpenAI allowed-header 文件已 ALREADY，仅补缺少的 buildUpstreamRequest 清理调用。反例/边界：未知 Anthropic beta 继续受白名单/丢弃策略限制；空请求不自动注入。

- `backend/internal/pkg/claude/constants.go`
- `backend/internal/service/gateway_upstream_request.go`:490 `computeFinalAnthropicBeta`
- `backend/internal/service/openai_gateway_forward.go`
- `backend/internal/service/openai_gateway_service.go`
### F06：Responses 消息与最终正文

限定 apicompat 现有转换器；不重复移植 0.2.5 的另一条文本恢复路径。反例/边界：已有非空最终文本优先，工具结果不得被正文替换。

- `backend/internal/pkg/apicompat/chatcompletions_to_responses.go`
- `backend/internal/pkg/apicompat/responses_to_chatcompletions.go`
- `backend/internal/pkg/apicompat/types.go`
### F07：Alpha search 成功判定

现有 alpha search 服务可达；修复解析器返回值及调用方。反例/边界：error/failed/incomplete、只有 delta 或 DONE、提前 EOF 必须失败。

- `backend/internal/service/openai_alpha_search.go`:546 `parseOpenAIResponsesSSEForAlphaSearch`
### F08：未来重置时间前保持配额暂停

复用本地 openAICodexWindowResetAt；不引入自动用卡重置服务。反例/边界：无未来 reset 的陈旧快照保持 fail-open；兼容规范窗口与旧字段。

- `backend/internal/service/account_scheduling_threshold_eval.go`
- `backend/internal/service/openai_gateway_scheduling.go`
### F09：TOTP 输入与计时器

现有个人资料与登录 TOTP 组件。反例/边界：成功流程不变；关闭后返回的网络响应不得重建 timer。

- `frontend/src/components/user/profile/TotpDisableDialog.vue`
- `frontend/src/components/user/profile/TotpSetupModal.vue`
### F10：公共表单交互

按各 PR 的独立行为验证；日期/选择器用真实组件测试。反例/边界：IME 组合期间不提交；多个弹窗不共享 ID；键盘及鼠标选择均可用。

- `frontend/src/components/admin/channel/ModelTagInput.vue`
- `frontend/src/components/common/BaseDialog.vue`
- `frontend/src/components/common/DateRangePicker.vue`
- `frontend/src/components/common/Pagination.vue`
- `frontend/src/components/common/SearchInput.vue`
- `frontend/src/components/common/Select.vue`
- `frontend/src/composables/useClipboard.ts`
### F11：代理管理小修

不修改代理回退数据库逻辑；后者归 S09。反例/边界：不同代理不串结果；临界时刻及无到期时间正常。

- `frontend/src/components/common/ProxySelector.vue`
- `frontend/src/utils/proxyExpiry.ts`
### F12：配置表单错误与输入边界

只处理现有非支付表单；共享 SettingsView 不携带支付代码。反例/边界：倍率0/负数拒绝；RPM 0允许、空值和小数拒绝；失败不伪装保存成功。

- `frontend/src/components/admin/group/GroupRPMOverridesModal.vue`
- `frontend/src/components/admin/group/GroupRateMultipliersModal.vue`
- `frontend/src/components/admin/user/GroupReplaceModal.vue`
- `frontend/src/composables/useModelWhitelist.ts`
- `frontend/src/stores/adminSettings.ts`
### F13：异步弹窗状态隔离

保留账号/分组 ID 作为身份；不恢复已移除菜单。反例/边界：A→B 快速切换及关闭后迟到响应不覆盖 B 或重开弹窗。

- `frontend/src/components/account/TempUnschedStatusModal.vue`
- `frontend/src/components/admin/group/GroupRPMOverridesModal.vue`
- `frontend/src/components/admin/group/GroupRateMultipliersModal.vue`
- `frontend/src/components/admin/user/UserApiKeysModal.vue`
- `frontend/src/components/user/UserErrorDetailModal.vue`
### F14：用量与价格显示

IntervalRow 由 PricingEntryCard 实际消费；不是新增推理力度导出列，不引入动态计价新后端。反例/边界：其它CSV公式前缀继续转义；不重算历史费用；无 reset 时允许显示可用。

- `frontend/src/components/account/UsageProgressBar.vue`
- `frontend/src/components/admin/channel/IntervalRow.vue`
- `frontend/src/views/user/UsageView.vue`
### F15：请求 ID 和配额诊断

本地 item ID 前缀策略不同，仅加长度上限；Grok 复用 GetAccessTokenForManualTest。反例/边界：短合法 ID、未知类型及无 ID 保持本地语义；配额查询仍校验认证。

- `backend/internal/service/grok_quota_service.go`
- `backend/internal/service/openai_responses_item_id.go`

## 4. 第二档：适配后移植

| 簇 | 来源 | A/C/X/N | 收益/契约 |
|---|---|---|---|
| S01 模型目录和计费别名 | [#7509](https://github.com/Wei-Shaw/sub2api/pull/7509)、[#7466](https://github.com/Wei-Shaw/sub2api/pull/7466)、[#7597](https://github.com/Wei-Shaw/sub2api/pull/7597)、[#7568](https://github.com/Wei-Shaw/sub2api/pull/7568) | 0/38/23/3 | 接入 GPT-6 Sol/Luna、Opus 5.5、Grok 4.7 的现有平台目录/能力/价格别名，并将 GPT≥5 按推理模型桥接 |
| S02 工具 schema 兼容 | [#7345](https://github.com/Wei-Shaw/sub2api/pull/7345)、[#7489](https://github.com/Wei-Shaw/sub2api/pull/7489)、[#7077](https://github.com/Wei-Shaw/sub2api/pull/7077)、[#7393](https://github.com/Wei-Shaw/sub2api/pull/7393) | 0/7/4/0 | 处理根级 union、required:null、tuple/prefixItems，并保留 string const 与 enum 的交集 |
| S03 工具参数和 PDF 保真 | [#7272](https://github.com/Wei-Shaw/sub2api/pull/7272)、[#7570](https://github.com/Wei-Shaw/sub2api/pull/7570)、[#7585](https://github.com/Wei-Shaw/sub2api/pull/7585) | 0/8/1/0 | arguments.done 等于已发送 delta；仅 block_start 携带参数也保留；内嵌 PDF 转为 document/inlineData |
| S04 流结束、取消与错误协议 | [#6473](https://github.com/Wei-Shaw/sub2api/pull/6473)、[#7303](https://github.com/Wei-Shaw/sub2api/pull/7303)、[#7276](https://github.com/Wei-Shaw/sub2api/pull/7276)、[#7300](https://github.com/Wei-Shaw/sub2api/pull/7300)、[#7339](https://github.com/Wei-Shaw/sub2api/pull/7339)、[#7097](https://github.com/Wei-Shaw/sub2api/pull/7097)、[#7448](https://github.com/Wei-Shaw/sub2api/pull/7448)、[#7609](https://github.com/Wei-Shaw/sub2api/pull/7609) | 0/25/16/2 | terminal 完整发送后结束；先取消单次请求再关响应体；心跳不算语义输出；错误只发一次；Gemini 传输失败换号；客户端取消归类 499 |
| S05 Antigravity 裸模型和空流 | [#7258](https://github.com/Wei-Shaw/sub2api/pull/7258)、[#7432](https://github.com/Wei-Shaw/sub2api/pull/7432)、[#7278](https://github.com/Wei-Shaw/sub2api/pull/7278)、[#7607](https://github.com/Wei-Shaw/sub2api/pull/7607) | 0/11/1/2 | 各入口将裸 Gemini 名按 thinking 配置映射；识别特定 SDK 的心跳兼容；MALFORMED_FUNCTION_CALL 空流触发 failover |
| S06 调度路由和账号可用性 | [#7427](https://github.com/Wei-Shaw/sub2api/pull/7427)、[#7481](https://github.com/Wei-Shaw/sub2api/pull/7481)、[#7139](https://github.com/Wei-Shaw/sub2api/pull/7139)、[#7195](https://github.com/Wei-Shaw/sub2api/pull/7195) | 0/9/7/0 | 非高级调度遵守 previous_response 归属；热路径用轻量分组读；API Key 未知模型 401 不误禁账号；暂停 OAuth 仍刷新 token |
| S07 账号成本与图片价继承 | [#7619](https://github.com/Wei-Shaw/sub2api/pull/7619)、[#7573](https://github.com/Wei-Shaw/sub2api/pull/7573) | 0/2/6/0 | 账号成本是否收取长上下文溢价取决于账号 gate；渠道图片价未填继承目录，显式 0 才免费 |
| S08 Codex 和 CC Switch 配置 | [#7395](https://github.com/Wei-Shaw/sub2api/pull/7395)、[#7322](https://github.com/Wei-Shaw/sub2api/pull/7322)、[#7622](https://github.com/Wei-Shaw/sub2api/pull/7622)、[#7628](https://github.com/Wei-Shaw/sub2api/pull/7628)、[#7549](https://github.com/Wei-Shaw/sub2api/pull/7549) | 0/9/4/0 | Codex config 请求正确 /v1；CC Switch保留配置端点（含显式/v1及子路径）、仅去尾斜杠且不自动追加/v1；usage路径恰好一个/v1；Windows catalog用~/ |
| S09 代理回退一致性 | [#7475](https://github.com/Wei-Shaw/sub2api/pull/7475)、[#7341](https://github.com/Wei-Shaw/sub2api/pull/7341)、[#7342](https://github.com/Wei-Shaw/sub2api/pull/7342) | 0/5/2/0 | 代理恢复改变网络身份时失效探针；过期扫描写入前重新核对快照；禁用代理不可作为回退目标 |
| S10 Responses 探测未知态 | [#7571](https://github.com/Wei-Shaw/sub2api/pull/7571) | 0/2/0/0 | 探测模型不存在时不写入endpoint能力标记（未知或已有值都保留），优先选普通GPT文本模型 |
| S11 公开响应模型别名 | [#7304](https://github.com/Wei-Shaw/sub2api/pull/7304) | 0/6/1/0 | Responses 与 Chat 返回请求方的公开模型别名，同时保留真实上游模型用于记录 |
| S12 现有内容审计输入边界 | [#7314](https://github.com/Wei-Shaw/sub2api/pull/7314)、[#7487](https://github.com/Wei-Shaw/sub2api/pull/7487) | 0/4/1/0 | 关键词检查不丢弃客户端 reminder 内文本，尾随 system 不遮蔽当前用户输入 |
| S13 Antigravity 系统身份兼容 | [#7256](https://github.com/Wei-Shaw/sub2api/pull/7256)、[#7411](https://github.com/Wei-Shaw/sub2api/pull/7411) | 0/4/1/2 | 仅在 Antigravity 转换中移除前导 attribution 并中和前导 SDK 身份 |
| S14 共享弹窗与标签收尾 | [#7319](https://github.com/Wei-Shaw/sub2api/pull/7319)、[#7498](https://github.com/Wei-Shaw/sub2api/pull/7498) | 0/1/2/1 | 嵌套弹窗关闭后滚动锁计数正确，模型标签 IME 不提前提交 |
| S15 Codex WS 上下文切换 | [#7615](https://github.com/Wei-Shaw/sub2api/pull/7615) | 0/3/0/0 | window_id 改变时删除旧 previous_response_id 并清空旧续聊推断锚点 |
| S16 Codex 请求边界兼容 | [#7038](https://github.com/Wei-Shaw/sub2api/pull/7038)、[#7426](https://github.com/Wei-Shaw/sub2api/pull/7426) | 0/7/1/1 | 映射到 GPT-5.5 时移除不兼容 Lite 标记但保留历史与工具；turn metadata 序列化保留非 ASCII 转义 |

第二档行为规格：[port-upstream-0.2.9-tier2](../../openspec/changes/port-upstream-0.2.9-tier2/README.md)。

### S01：模型目录和计费别名

取 v0.2.9 终态；#7568 依赖 #7509 新 spelling helper；剔除 MiniMax/OpenCode 等缺失平台路径。反例/边界：兼容 dotted Opus 别名；非 GPT 数字系列和旧模型保持行为；显式用户映射优先。

- `backend/internal/handler/gateway_handler.go`
- `backend/internal/pkg/apicompat/anthropic_to_responses.go`
- `backend/internal/pkg/apicompat/anthropic_to_responses_response.go`
- `backend/internal/pkg/apicompat/chatcompletions_to_responses.go`
- `backend/internal/pkg/apicompat/responses_to_anthropic_request.go`
- `backend/internal/pkg/apicompat/types.go`
- `backend/internal/pkg/claude/constants.go`
- `backend/internal/pkg/claude/effort_catalog.go`
- `backend/internal/pkg/openai/constants.go`
- `backend/internal/pkg/xai/models.go`
- `backend/internal/service/billing_service.go`
- `backend/internal/service/gateway_claude_oauth_body.go`
- `backend/internal/service/gateway_count_tokens.go`
- `backend/internal/service/gateway_forward.go`
- `backend/internal/service/gateway_forward_as_chat_completions.go`
- `backend/internal/service/gateway_forward_as_responses.go`
- `backend/internal/service/gateway_request.go`
- `backend/internal/service/openai_codex_models_service.go`
- `backend/internal/service/openai_codex_transform.go`
- `backend/internal/service/openai_compat_prompt_cache_key.go`
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_chat_completions_anthropic_native.go`（本地缺文件，不能整取）
- `backend/internal/service/openai_gateway_chat_completions_raw.go`
- `backend/internal/service/openai_gateway_forward.go`
- `backend/internal/service/openai_gateway_grok.go`
- `backend/internal/service/openai_gateway_grok_chat_bridge.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_gateway_request_body.go`
- `backend/internal/service/openai_gateway_responses_anthropic_native.go`（本地缺文件，不能整取）
- `backend/internal/service/openai_model_alias.go`
- `backend/internal/service/opencode_go.go`（本地缺文件，不能整取）
- `backend/internal/service/pricing_service.go`
- `backend/internal/service/upstream_response_model.go`
- `backend/resources/model-pricing/model_prices_and_context_window.json`
- `frontend/src/components/account/AccountStatusIndicator.vue`
- `frontend/src/components/keys/UseKeyModal.vue`
- `frontend/src/composables/useModelWhitelist.ts`
### S02：工具 schema 兼容

量过 #7345 新独立 helper 文件213行；#7489依赖本地缺失的ForPlatform入口/自定义parser helpers，本地现有约139行offset sanitizer，需扩现有清洗器而非照搬上游parser；#7393在#7077后按终态适配。反例/边界：普通object不变；保留目标协议可表达约束；tuple/root union降格不宣称schema完全等价；多个分流入口均执行清洗。

- `backend/internal/pkg/antigravity/schema_cleaner.go`
- `backend/internal/pkg/apicompat/responses_to_anthropic_request.go`
- `backend/internal/pkg/apicompat/responses_to_anthropic_tool_schema.go`（本地缺文件，不能整取）
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_responses_tool_schema.go`:42 `sanitizeOpenAIResponsesToolParameterTypes`
### S03：工具参数和 PDF 保真

#7272 在 #7570 之前或按终态一次移植；PDF 覆盖 Responses→Anthropic→Antigravity 全链。反例/边界：有 delta 时覆盖 seed，不拼接两个 JSON；空 PDF 不发；不自动下载 file_id。

- `backend/internal/pkg/antigravity/request_transformer.go`
- `backend/internal/pkg/apicompat/anthropic_to_responses_response.go`
- `backend/internal/pkg/apicompat/responses_to_anthropic_request.go`
### S04：流结束、取消与错误协议

HTTP/WS/压缩体/handler 共享路径，需分步骤和 race 验证；#7448 在 #7609 前适配；保留 SQLite 用量去重与取消后成功用量落库。反例/边界：bare error 后可能 completed 不早退；上游真实错误继续归因；取消不触发重复重放或重复用量；countTokens 保留本地估算。

- `backend/internal/handler/gateway_handler.go`
- `backend/internal/handler/gateway_handler_chat_completions.go`
- `backend/internal/handler/gateway_handler_responses.go`
- `backend/internal/handler/gemini_v1beta_handler.go`
- `backend/internal/handler/openai_alpha_search.go`
- `backend/internal/handler/ops_error_logger.go`
- `backend/internal/pkg/googleapi/status.go`
- `backend/internal/repository/http_upstream.go`
- `backend/internal/service/antigravity_gateway_claude.go`
- `backend/internal/service/antigravity_gateway_compat.go`
- `backend/internal/service/antigravity_gateway_gemini.go`
- `backend/internal/service/antigravity_gateway_streaming.go`
- `backend/internal/service/gateway_forward_as_chat_completions.go`
- `backend/internal/service/gateway_upstream_transport_error.go`
- `backend/internal/service/gemini_chat_completions_compat_service.go`
- `backend/internal/service/gemini_messages_compat_service.go`
- `backend/internal/service/gemini_upstream_transport_error.go`（本地缺文件，不能整取）
- `backend/internal/service/openai_gateway_passthrough.go`
- `backend/internal/service/openai_gateway_response_handling.go`
- `backend/internal/service/openai_upstream_transport_error.go`
- `backend/internal/service/openai_ws_http_bridge.go`
### S05：Antigravity 裸模型和空流

#7258 的新 helper 137 行，#7432 为后续消费者；不以缺 helper 直接判整簇不可做。反例/边界：显式 model_mapping 优先；只有 signature/stop 不算输出；真实文本/工具调用不重试。

- `backend/internal/pkg/antigravity/stream_transformer.go`
- `backend/internal/service/antigravity_gateway_claude.go`
- `backend/internal/service/antigravity_gateway_compat.go`
- `backend/internal/service/antigravity_gateway_compat_stream.go`
- `backend/internal/service/antigravity_gateway_gemini.go`
- `backend/internal/service/antigravity_gateway_service.go`
- `backend/internal/service/antigravity_gateway_streaming.go`
- `backend/internal/service/antigravity_gemini_thinking_variant.go`（本地缺文件，不能整取）
- `backend/internal/service/gemini_sse_comment_compat.go`（本地缺文件，不能整取）
### S06：调度路由和账号可用性

GetByIDLite/GetGroupByIDLite 和兼容判定在位；#7195 改写 SQLite 手写查询，不能覆盖 account_repo.go。反例/边界：失效归属释放槽位；真实凭据 401 保留禁用；永久错误账号不进入刷新候选。

- `backend/internal/repository/account_repo.go`
- `backend/internal/service/gateway_scheduling.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/openai_account_scheduler.go`
- `backend/internal/service/openai_gateway_scheduling.go`
- `backend/internal/service/openai_gateway_upstream_errors.go`
- `backend/internal/service/ratelimit_service.go`
### S07：账号成本与图片价继承

本地 tryModelFilePricing 只有4参数、使用 CalculateCostWithServiceTier；已有 calculateCostWithServiceTierPolicy 可复用，优先传gate到该现有方法，不引入上游 effort/峰谷/区间定价迁移。反例/边界：保留自定义账号价与 ApplyPricingToAccountStats 优先级；不改变用户售价门控；空价与零价分开。

- `backend/internal/service/account_stats_pricing.go`:68 `tryModelFilePricing`
- `backend/internal/service/billing_service.go`
- `backend/internal/service/gateway_usage_billing.go`
- `backend/internal/service/model_pricing_resolver.go`
- `backend/internal/service/openai_gateway_usage.go`
### S08：Codex 和 CC Switch 配置

#7395 中 CC Switch 的 /v1 改动被 #7622 修正，必须取 v0.2.9 终态；保留本地路由目录和 TOML 转义。反例/边界：带/不带/v1、尾斜杠、子路径和不同平台均验证；原生配置与 CC Switch 的 URL 规则分别覆盖。

- `frontend/src/components/keys/UseKeyModal.vue`
- `frontend/src/utils/ccswitchImport.ts`
- `frontend/src/views/user/KeysView.vue`
### S09：代理回退一致性

两处手写 SQL 必须 SQLite 适配，覆盖真实库和 JSON 删除语义；无新表、无新迁移。反例/边界：无真实 proxy 变化不清探针；续期/停用/改回退配置时旧扫描不得覆盖。

- `backend/internal/repository/account_repo.go`:3532 `RevertProxyFallback`
- `backend/internal/repository/proxy_repo.go`:687 `SweepExpiredProxies`
- `backend/internal/service/proxy_fallback.go`
### S10：Responses 探测未知态

CLEAN 仍缺 isExplicitOpenAIModelAvailabilityMessage：上游 openai_compact_fallback.go 中独立约32行 stdlib helper，随探测一起移植，不引入 compact fallback 整簇。反例/边界：真正endpoint404/405保持不支持；不能把任意400当模型不可用或强制清空已有标记。

- `backend/internal/service/openai_apikey_responses_probe.go`:82 `selectResponsesProbeModel`
### S11：公开响应模型别名

按本地桥接/原生入口核查，不以单条路径通过代替全部入口。反例/边界：工具/事件里的其它 model 字段不误改；未映射模型保持原样。

- `backend/internal/service/openai_gateway_chat_completions_raw.go`
- `backend/internal/service/openai_gateway_passthrough.go`
- `backend/internal/service/openai_gateway_request_body.go`
- `backend/internal/service/openai_gateway_response_handling.go`
### S12：现有内容审计输入边界

只修现有引擎，无 TypeSafe 引擎或 meta 迁移；当前仅静态确认旧路径，实施要提供可运行绕过/反例。反例/边界：assistant/tool 结束回合不重复审计；正常语义审核保留既有 reminder 策略。

- `backend/internal/service/content_moderation.go`
- `backend/internal/service/content_moderation_input.go`
### S13：Antigravity 系统身份兼容

两 PR 按终态一起取；不带历史截图或 .gitignore 变动。反例/边界：正文中提及 Claude 不改；原生 Anthropic attribution 必须保留。

- `backend/internal/pkg/antigravity/request_transformer.go`
### S14：共享弹窗与标签收尾

#7498 在 F10 #7238 后；测试文件缺失只表示要新增本地回归，不是缺产品基座。反例/边界：其它弹窗仍开时 body 保持锁；非组合输入和原键盘导航不回退。

- `frontend/src/components/admin/channel/ModelTagInput.vue`
- `frontend/src/components/common/BaseDialog.vue`
### S15：Codex WS 上下文切换

生产 ingress/payload 均在位；代码小但会改会话状态，独立验证真实多轮行为。反例/边界：同窗口、无窗口标记保持续聊；失败轮次不得提前更新 last window。

- `backend/internal/service/openai_ws_forwarder_ingress.go`
- `backend/internal/service/openai_ws_forwarder_payload.go`
### S16：Codex 请求边界兼容

Lite helper 42行、metadata helper 39行；仅接本地现有 fingerprint 消费者，缺失 account_identity 子系统不引入。反例/边界：其它模型和账户不改；失败后换号仍使用原请求；中文与补充平面字符转义合法。

- `backend/internal/service/openai_codex_account_identity.go`（本地缺文件，不能整取）
- `backend/internal/service/openai_codex_fingerprint.go`
- `backend/internal/service/openai_codex_turn_metadata.go`（本地缺文件，不能整取）
- `backend/internal/service/openai_gateway_forward.go`
- `backend/internal/service/openai_gateway_passthrough.go`
- `backend/internal/service/openai_lite_mapped_gpt55.go`（本地缺文件，不能整取）
- `backend/internal/service/openai_ws_http_bridge.go`

## 5. 按需/不做项目与旧轮次遗留

### 5.1 本轮不进入默认移植的内容

第三/四档完整到候选的判断在 decisions.tsv；以下按决策合并，未创建对应 change。

| 内容 | 处理依据 |
|---|---|
| OpenCode Go/Zen、Command Code、Kimi原生、插件宿主 | 旧平台或宿主基座缺失。OpenCode Go用量PR为41文件+5260行，前置#6747为108文件；投影/CF1010/占位符不能假装独立小修 |
| DeepSeek/Kimi/GLM API Key兼容、Baseten、Vertex | 可经现有兼容账号使用，不因缺专属平台一概排除；使用对应供应商时才取host限定子集 |
| model_allowlist glob、pinned/live模型发现、混合模型目录 | 新授权/目录策略；旧 model_allowlist/schema 基座未采用。#7526/#7595不直接纳入 |
| Free Fast、effort计费倍率、自动用卡重置 | 现有 fast→priority 价格别名及 quota pause 不等于这些功能；#7613/#7555缺的生产子系统仍不存在 |
| Seedance/视频、图像新路由、TypeSafe、自动Claude版本、月度备份 | 新端点、引擎或后台服务；按实际个人用途单独立项，保留现有手动配置/原生备份方式 |
| 运营管理、订阅/兑换/公告、支付/返利 | 已裁剪入口不恢复；支付专用修复也排除。用量落库、token成本与账号额度仍属个人网关核心范围 |
| Redis Docker命令、日志保留、发布矩阵、依赖升级 | 原生单进程部署和上游集群部署不同；#7261中的已有x/*依赖升级需独立审计，不能因gRPC不适用就忽略全部依赖风险 |

### 5.2 历史待办重新核对

| 历史来源 | 当前状态与本轮处理 |
|---|---|
| 0.2.5第一/二档主体 | `bfbcd79` 已包含129文件变更；不能因旧verification仍写“工作树未提交”当作未合。冻结记录保留，当前事实以commit为准 |
| 0.2.5 #7043 | 默认WS系数5.0未采用；本地已有#7064容量逻辑和当前默认1.0，仍需真实容量依据，避免1C1G内存增加 |
| 0.2.5 #6954 / S12b | 平台限额清理及拟定迁移227未实施，最大仍226；与新#7237/#7289按使用情况一并决定 |
| 0.2.5 #6960/#7049 | 完整exec_scope/抢占和queue-wake重构仍未整体引入；现有 reader loop/池容量修复已落。本轮S15只修本地已有window边界，不补整套抢占 |
| 0.2.5 #6575/#6890/#6974/#6769 | Gemini3.7/3.8目录、原生Codex Images、DeepSeek峰谷价、Ollama异步重置仍按需。S05的裸名映射可用现有账号映射，不要求先扩目录 |
| 0.2.5 #6747/#6986/#7091/#6691/#7153 | OpenCode平台、站点收费开关、订阅批量、Key批量/供应商过滤没有全部引入；收费入口现按个人项目边界排除，其余是按需功能 |
| 0.2.5 #6989/#7057/#6968/#7011/#6945/#6988 | 套餐/目录显示、上下文元数据、DeepSeek视觉、compact默认和pinned检索，仍按实际客户端使用；本轮新模型S01只扩现有目录，不默认补pinned子系统 |
| 0.2.5 #7129/#7085/#6929/#6917 | Grok序号全局序列化、本地input stripper、Gemini带内错误/监控BaseURL，需要保留fork差异；不直接覆盖类型或网关文件 |
| 0.2.5 #6847/#7001/#6845/#6913/#6915 | 自定义页按钮、ops统计/列序、批量删用户、审核标签为可选管理/展示项；没有因为版本前进自动提升 |
| 0.2.5第四档 #7126/#7110/#7148 | namespace语义已由fork覆盖；access-cache改动上游已撤回，保持净无变化；其它PG/bootstrap/MiniMax/macOS项仍不适用 |
| 0.2.0 §5.1 / 0.1.180 §7.2 / 0.1.184 §5.2 | Fast发送/免费Fast/Ultrafast整套策略仍未合；不要将已落目录priority与新策略混为一谈 |
| 0.2.0 §5.2 / 0.1.180 §7.5 | 动态长上下文阶梯仍未整体采用。本地 `tryModelFilePricing` 为4参数并调用旧service-tier方法；S07移植gate语义，不带effort/峰谷/新schema |
| 0.1.180 §6.1四条工具桥尾项 | 非法参数、namespace别名、deferred标记未整体合；与S02/S03改同一转换器，实施时回归并明确旧项是否被终态吸收，不能计为自动完成 |
| 0.1.180 §6.2(c)/§7.1 与 0.1.184 §5.4/§5.8 | Grok/WS透传/会话抢占整体仍按需，单节点不引入Redis跨节点租约；本轮S04/S15仅涵盖明确列出的行为 |
| 0.1.184 §5.5/§5.6 与 §11.6 | usage compaction/requested effort迁移、用户公共分组限制、订阅重置锚点、峰谷价和图像工具冷却仍独立待办；支付部分排除 |
| 0.1.180 §7.3/§7.4 | 自动重置卡和模型读取上限仍未采用；#7555不能作为单纯加4个字段处理 |
| DOMPurify/xlsx旧安全待办 | package中DOMPurify ^3.3.1、xlsx ^0.18.5未变；没有运行当前audit，不复用09-02的advisory数、最新版本或PoC结论。需另做依赖核验和lock更新 |
| 0.2.0已落P0/P1、0.1.183/179/176 | 主体已落；API Key instructions、routed catalog、Responses Lite串行等已确认能力不重新立项 |

以上分档是本次审查判断，不代表用户已选择按需项目；未来实施前三/四档需先明确具体用途。本次没有启动其实现。

## 6. ALREADY 与无需移植的语义

- 8个文件级ALREADY里，7个属于上游文档/忽略规则清理，1个是#7617的OpenAI header白名单；没有整个功能PR全文件ALREADY。
- #7402：本地 `http_upstream.go` 已有 `openAIHTTP2ReadIdleTimeout=15s`、`openAIHTTP2PingTimeout=15s`；不受上游10s/5s回退影响。
- #7177：本地 `BindResponseAccount` 已调用 `withOpenAIWSStateStoreRedisWriteTimeout`，后者先WithoutCancel再设3秒超时；上游另修的HTTP owner消费者在本地不存在。上层函数CONFLICT不等于缺陷仍在。
- #7617只需补构建请求时的旧beta清理；不能因为其中一个文件ALREADY就跳过整PR。
- 无Go环境时上述为调用链静态证据，不是新增回归测试PASS。

## 7. 建议落地顺序

1. 固定同步分支基线；补齐Go1.27环境、后端unit/build基线与SQLite方言审计。所有产品任务仍未实施。
2. 第一档先做F01–F08、F15的请求/调度正确性，再做F09–F14界面项；每簇独立验证。
3. 第二档先S01模型→S02工具schema→S03参数/PDF；#7272先于#7570，#7509先于#7568。
4. S04连接生命周期/终止/错误整体按v0.2.9依赖顺序落，之后S06调度、S15/S16多轮与请求边界；不能并行覆盖相同service文件。
5. S05/S13 Antigravity、S10探测、S11公开别名、S12审计分别验收；S12须先跑实际失败场景，不能把静态判断当安全PoC。
6. S07成本与S09代理SQL独立批次，使用真实SQLite；S08配置终态、S14 UI最后收尾（S14在F10之后）。
7. 汇总unit/build/typecheck/Vitest/SQLite/race结果，审阅最终差异后再决定发布；版本仍用fork自己的1.1.x。

不按867个文件态估工时；同一文件和依赖链高度重叠。最大风险集中S04、S06、S07、S09，不能作为“补丁大多CLEAN”的简单升级。

## 8. 自测与验收

[基线验证](./evidence-0.2.9/baseline-tests.md)记录前端42项通过及后端Go缺失；[测试清单](./evidence-0.2.9/test-inventory.tsv)明确哪些引用已存在。
两份OpenSpec任务全部未勾选，verification证据格保留空白；评估检查和产品验收分开。

实施最低门禁：目标缺陷的失败/通过及正常反例，相关Go package unit，`go build ./...`，SQLite方言与真实库用例；
触及前端执行组件Vitest和typecheck，触及共享状态/响应体执行race及可取消的真实httptest流。
`integration && postgres` 的测试不计SQLite证据。检查成功用量只落一次、取消不乱扣费、不会重放已输出请求。

## 9. 本轮可复用的结论

1. 文档里的“最新上游”和旧verification状态可能落后于实际tag/commit；冻结历史不改写，新增现行状态说明。
2. 同一参数上游改坏再改回时，fork可能从未受影响；#7402必须先看本地15s/15s。
3. 反向补丁匹配也包含“未曾导入的文件已不存在”，不能把上游清理文档的ALREADY算成已移植业务。
4. URL补丁也有中间态：#7395与#7622服务不同消费者，最终按原生Codex和CC Switch分别测试。
5. 配额投影小修可能依赖整套后台服务；RPM在位可补，自动用卡重置缺失则不能只加字段。
6. 成本统计修复必须保留本地接口与gate语义；不要为一条修复顺带导入上游多项计价策略和迁移。
