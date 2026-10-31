# 历史已评估版本的未合项复查（2026-10-31）

建议下一批先补 **Codex 工具桥接尾项和 Responses 事件序号**，然后按使用需求选择 **Grok 模型目录诊断**。统计展示、Claude 重置额度、生图扩展和性能缓存继续按实际用途选择。

本次只复查已评估到 v0.2.19 的历史范围；未评估之后的新 release。当前结论覆盖旧报告的完成状态和本次明确复测的行为，不表示重新执行了每个历史 PR 的全部检查。

## 1. 当前基线与已完成项目

| 项目 | 值 |
|---|---|
| fork HEAD / tag | `e44324231ec98b5176fe04c6ddb3f3552a352889` / `v1.1.21` |
| VERSION / Go / 最大 SQLite 迁移 | `1.1.21` / `1.27.3` / `227` |
| 产品范围 | SQLite-only、可选进程内 miniredis、个人单节点、simple mode |
| 历史评估终点 | 上游 v0.2.19 / `dece26153621959ccd7635814f4b2d08304e681a` |
| 既有活动 change | 7 份，276/276 个任务完成，0 个未勾选；仍有待归档记录 |

[当前基线与逐份计数](./unmerged-review-2026-10-31/baseline.json)、[v1.1.21 实际发布内容](../releases/v1.1.21.md)和各批 verification 共同判定完成状态。不能直接沿用旧 PORTING 中的“未实施”。

以下已经完成，**不重复推荐**：历次第一/第二档主体、0.2.0 收尾、0.2.5 平台限额清理与 SQLite227、SQLite 三个后台任务、审计日志清空、延迟账号探测、零分配模型签名、日志输出 TPS、账号筛选记忆、运维详情账号名、模型级冷却、OpenAI 搜索模拟、最小输出 Token、完整运维页、GPT-6.1 Terra，以及已接受的前端安全维护。

0.2.9 第一档的完成依据包含用户明确确认和归档 completion-record，不能从旧冻结验收表推断其仍未合。[记录](../../openspec/changes/archive/2026-10-24-port-upstream-0.2.9-tier1/completion-record.md)。

## 2. 复查方法与证据边界

先读取最新遗留表、后续实现记录和当前代码，再沿现有生产消费者核对缺口。推荐顺序继续优先 Claude Code、Codex、Grok；新增能力保持第三档，支付与不适用架构保持第四档。

对下面 C01–C04 的完整行为来源做独立 index 检查：所有文件先反向，再正向，25 条文件状态为 **A1/C14/X9/N1**。[逐文件状态](./unmerged-review-2026-10-31/source-file-states.tsv)、[原始检查结果](./unmerged-review-2026-10-31/source-file-checks.json)、[来源 SHA 与补丁哈希](./unmerged-review-2026-10-31/source-refs.json)。C01 将两条配套提交合成终态 diff，C04 取 PR merge 第一父到 merge 的完整 diff。

本轮没有向产品文件应用补丁。使用 Go overlay 注入临时测试，分别验证转换器、真实 gateway 的 Chat fallback 出口、Grok 请求工具降级、compact/WS/handler 合成事件；正常基线测试另行执行。探针的预期失败是缺口证据，不是实施验收失败或修复完成证据。

在线 GitHub PR 打开和检索均返回 HTTP 502；本报告的来源依据是可读取的 git 对象和仓库冻结的整 PR 补丁。未据此宣称最新模型可用性、实时价格或 tag 后行为。

## 3. 建议优先补入：四个现存兼容性尾项

四项建议作为一个**第二档**批次，沿现有接口和数据库结构适配。C01–C03 保持旧工具桥接待办的实际范围；C04 从历史“按需/未整体采用”的宽泛标签，明确调整为本次复现的共享客户端协议缺口。调整只覆盖事件字段及相应生产消费者，不带入新平台或完整 Grok/WS 调度机制。

| 编号 | 未合行为与触发 | 当前复测结果 | 收益与适配量 |
|---|---|---|---|
| C01 | 普通 function-call 参数完整性：Chat 兼容上游返回截断 JSON，或 Codex 回放这段历史 | `arguments="{\"path\":"` 仍生成 `status=completed`；真实 gateway 返回 result 且 err=nil。下一次转换仍转发坏调用及对应输出 | 防止工具调用和后续会话持续失败；需手工接入流终态、非流转换、历史过滤，并保留已观测 usage |
| C02 | 自定义工具的无歧义 namespace 别名：声明 custom `exec` 和 namespace `functions/wait`，上游调用 `functions__exec` | 回程仍为 `function_call/functions__exec`，没有还原成 `custom_tool_call/exec` | 恢复 Codex 工具路由；需传递显式 function 声明并验证与真实 namespace 子工具的冲突优先级 |
| C03 | 去除孤儿 `defer_loading`：内建 `tool_search` 降为普通 function 后，其它工具仍保留依赖它的标记 | 通用 adapter 和实际 Grok 出站 body 都保留该标记 | 避免向上游发送失去前置条件的工具声明；只清理最终声明列表中的孤儿标记 |
| C04 | Responses 事件序号：第一帧为 0，或 gateway 合成 compact、失败和 WS HTTP bridge 事件 | 0 被 `omitempty` 丢弃；compact item/terminal、compact failure、WS error、handler failure 均缺少字段 | 补齐严格客户端所需字段；正值与事件顺序必须保留，不能在已编号流里统一重置为 0 |

来源：C01 `e2d9ce0ca` + `fbc9ee626`；C02 `31d5b67ba`；C03 `7a09a2eaf`；C04 [PR #7129](https://github.com/Wei-Shaw/sub2api/pull/7129)。上述用户影响中的客户端解析/执行规则依据冻结上游修复与现有转换合同；本轮执行到服务出口及序列化边界，没有使用真实 Codex/Grok Build 客户端连接验收。

### patch sites 与前置量测

- C01：`backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go` 的历史 items、非流输出和 stream state；`backend/internal/service/openai_gateway_responses_chat_fallback.go` 的 finalize 前检查。当前没有 `ValidateToolCallArguments`，必须与其实际调用者一起适配；配套收窄提交不可遗漏。有效 JSON 即使 `finish_reason=length` 也不能仅凭该字段误拒绝；custom/freeform input 不按普通 JSON arguments 处理。
- C02：同一转换器的 custom 判定和 stream state；fallback 的声明集合传递。新增 `FunctionToolNames` 和别名判定随生产消费者落地，不为了凑符号独立添加 helper。显式 function 和真实 namespace 子工具优先，含歧义的名称不猜测。
- C03：`backend/internal/pkg/apicompat/responses_client_tools.go`、`backend/internal/service/openai_gateway_grok.go` 最终 tools 列表。真正的内建 `tool_search` 仍存在时保留合法 defer 标记，同名普通 function 不能充当该前置。已有 Lite 测试文件为 ALREADY，不重复应用。
- C04：`backend/internal/pkg/apicompat/types.go`、`backend/internal/handler/stream_error_event.go`、`backend/internal/service/openai_compact_stream_bridge.go`、`backend/internal/service/openai_ws_http_bridge.go`。来源缺失文件是一个旧测试文件，当前已有 compact/错误测试入口；按现有入口补回归，不导入缺失测试的全部历史前置。已知流序号从实际 writer/stream state 继续，不照搬固定 0 的终止回退。

这四项没有新依赖、DDL 或 Ent/Wire 来源变更。各来源文件的 CLEAN/ALREADY 不等同于消费者已修，当前探针结果才是继续处理的依据。[建议 OpenSpec](../../openspec/changes/port-upstream-historical-client-compatibility/README.md)定义范围及未来验收；任务未勾选，实施证据为空。

## 4. 仍值得选择的第三档功能

“值得选择”表示使用收益判断，不自动改变档位或启动实施。表中的文件/行数是原完整来源规模，包含测试/生成代码，**不是本轮剩余改动估算**。

| 优先顺序 | 功能 | 何时值得合 | 当前增量与成本 |
|---|---|---|---|
| 1 | **Grok 模型目录诊断**（#8339 未选子项） | 经常配置 Grok 映射、换 Base URL 或排查“目录有模型但不可用” | 当前已有观测/CAS/乱序保护；新增目录新鲜度、来源变化、映射与目录差异展示。诊断服务来源265行，加 DTO/API/编辑面板；中等工作量，可复用既有底座，不重做整44文件 PR |
| 2，取决于端点 | **渠道监控支持 Base URL 路径前缀**（#6917） | 使用 `/proxy`、`/anthropic`、`/v1` 等带前缀的第三方端点 | 当前监控明确仅允许 origin，拼接也未去重；这是扩展受支持配置。来源5文件/+111/-11；需要完整路径段匹配、编码路径和既有 SSRF 边界回归 |
| 3，统计需求明确时 | **整窗口输出 TPS 分布**（#7911 余项） | 希望比较慢请求、供应商和账号的实际输出速度分布 | 日志 TPS 和 GPT 模型平均 `avg_tokens_per_sec` 已存在。余项是整窗口 P5/P10/P50、平均值及样本数，SQLite 需重写 PG percentile SQL，并量查询成本；不平均小时分位数 |
| 4 | **用量/费用趋势切换**（#7646） | 经常看花费曲线，尤其有多个用户时 | 来源改的是 Top12 用户趋势按 tokens/actual_cost 排名与显示，15文件/+118/-20；当前已有费用总览、图表 tooltip 和单节点统计，严格单用户收益较小。应包含 metric 缓存键和乱序响应保护 |
| 5，实际拥有额度时 | **Claude 可用重置额度查询，再选兑换**（#7684/#7726） | Claude OAuth 账号确实拥有可用重置 grant，且需要在网关中管理 | 普通5小时/7天使用率查询已存在；新增的是重置次数、有效期、阻塞/冷却、可兑换性。查询14文件/+691/-13，兑换15文件/+1314/-20；查询需 profile scope，兑换需避免未知结果重试与重复消费 |
| 6，经常批量生图时 | **统一图片实际张数计量开关**（#8323） | 一次生成多张、请求张数与实际返回张数不同，或需要跨现有入口统一计量 | 现有部分入口已有 ImageCount 和按图片计价；未合的是更完整的计数回退/门控及分组开关，41文件/+728/-78。需新的本地 SQLite 迁移、缓存/DTO/UI、token/按次模式和 dedup 回归；不直接采用上游248编号 |
| 按对应账号使用 | **未选新模型与 Gemini 能力**（#7957/#8078/#8114/#8166/#8167/#8395 等） | 实际使用相应 Haiku/Gemini 模型、思考摘要、Vertex service_account 配额或生图入口 | 先核实账号声明及真实工具/流/usage支持；现有映射和 metadata 可用不等于完整新增能力已接受。不捆绑导入模型全集、缓存 CRUD 或未选平台 |

Grok 诊断不能把 stale/来源已变/同步失败的旧目录判为“模型不存在”，也不能把“列在目录中”当可调用性保证。若选择该项，应覆盖保存/重开、未保存编辑提示、失败同步和诊断只读行为。

图片张数与上游账号额度属于个人 gateway 用量管理，不能误判为 payment-only；这也不授权任何购买、充值、付费续订或退款功能。

## 5. 继续暂缓或排除

| 项目 | 当前建议与原因 |
|---|---|
| #8319 受限管理员密钥 | 第三档暂缓。原92文件/+9054/-131，涉及权限、密钥摘要、IP/过期/撤销、schema和UI；单人管理收益有限。需要外部账号自动化或委托管理时再选择 |
| #7681 余额在途预占 | 第三档暂缓。原21文件/+2261/-7，上游在 simple mode 禁用；对目前个人 simple 用途不是默认收益项 |
| #7679 风控白名单、#7752 Key 创建限额、公共分组限制 | 第三档暂缓。属于多人准入和治理策略；不默认新增个人账户限制或豁免 |
| #8251 解码快照缓存/revision/COW、#8269 SQLite 轻量 quota 查询 | 先做实际规模 benchmark，再决定。现有解码路径未证实同类共享对象竞态；缓存新增常驻内存和失效协议。PG JSONB 实现第四档排除，SQLite 思路仍第三档 |
| 完整 pinned/live/model_allowlist glob、新模型目录系统 | 第三档暂缓。已有白名单/映射、目录刷新/ETag、Grok观测/CAS和Terra目录；新授权与目录系统不是本次协议尾项的前置 |
| Fast/Free Fast/Ultrafast、动态长上下文阶梯、effort倍率、usage compaction/requested effort schema | 第三档。需要实际请求策略及模型/账号权利、计价需求；保留现有 Terra/模型价格规则，不为功能齐全改 schema |
| #6960/#7049 exec_scope/抢占/queue-wake、完整 Grok 429/Realtime/WS 容量机制 | 按真实并发/语音用途选择。当前已选容量/退役/断流修复不代表整簇已合；1C1G须验证内存、取消、重放和调度语义 |
| #8287 Gemini 重试可配置 | 低优先级第三档。现有固定5次总尝试与来源默认4次重试+首次一致；仅需实际调参时增加配置 |
| #8348 上传20→50 MiB | 低优先级第三档。当前完整性/明确拒绝已修；只有实际大附件需求才扩大资源边界 |
| 分组 i18n、批量管理、公告/运营展示、月度备份后台 | 按用途选择独立非支付子项；现有原生备份能力不等于需要照搬上游运营后台 |
| 新原生平台/插件/视频引擎 | 缺平台缺陷维持第四档；若用户要新增能力，需重新定义用途和SQLite适配。已有 OpenCode 配置生成不需要 OpenCode 原生账号平台 |
| PostgreSQL、多实例 Redis 租约、平台不存在的修复 | 第四档，架构或消费者不适用；不列作排期中的个人版缺陷 |
| 支付、充值、购买/续费、退款、返利、支付提供商配置/依赖/迁移 | 第四档，个人范围明确排除。旧 mixed PR 只允许独立非支付子集 |
| #7043 默认 WS 系数5.0 | 已独立决定不采纳，保持1.0；不是等待实施的任务 |

## 6. 更早轮次遗留的当前状态

| 轮次 | 已完成与仍未整体采用的边界 |
|---|---|
| 0.1.176/179/183/184 主体 | 主体已完成；Responses Lite、API Key instructions、端点归属和固定 quota 窗口不重开。动态价格/重置锚点、公共分组限制、自动重置卡和读模型上限按需 |
| 0.1.180 工具/Grok/价格 | 已选PDF、tool身份、HTTP回放/孤儿tool、composite闸门完成；工具参数、别名、defer尾项由本报告 C01–C03 重新量测。Grok大机制、Fast和动态阶梯仍按需，跨节点租约架构不需要 |
| 0.2.0/0.2.4 | 主体和0.2.0尾项已完成；旧DOMPurify任务由后续维护完成。Fast/免费Fast和数据驱动长上下文仍按需 |
| 0.2.5 | 主体及第二档33/33、配置限额/0值/软删/SQLite227完成，WS默认1.0为已决策。#7129由C04重测；#6917为按端点选择。exec_scope/抢占、原生Codex图片、DeepSeek峰谷/Ollama重置、目录/compact/pinned、input stripper、Gemini带内信号和管理展示未整体合，仍需实际消费者/用途量测 |
| 0.2.9 | 第一档按用户确认完成并归档，第二档在159d04222完成。OpenCode/Command Code/Kimi原生、插件、完整目录策略、Free Fast/effort/自动用卡、视频及运营后台仍按需或缺平台排除 |
| 0.2.11/0.2.12 | 主体和已选优先级/Key排序完成。重置grant、趋势、风控白名单、Key限制、在途预占与Ultrafast按需；TypeSafe缺平台、支付排除 |
| 0.2.16 | 第一/第二档完成；TPS日志、详情账号名、模型scope冷却已被后续个人功能承接。剩余窗口分布、Haiku新模型、分组i18n、Gemini显式缓存/Vertex API Key仍按需 |
| 0.2.18 | 第一/第二档、六项个人功能及个人维护完成，三个SQLite后台不再跳过。共享快照缓存/COW、轻量SQL思路、Gemini摘要/新增配额/生图扩展等未选子项独立 |
| 0.2.19 | 第一档26/26、第二档23/23、Terra14/14完成。剩余受限管理员密钥、图片计量开关、Grok诊断、重试配置、50MiB和新Gemini目录保持第三档；缺消费者和支付第四档 |

更早细目仍由[最新完整遗留表](./PORTING-0.2.19.md#53-旧轮次状态重述)及各原始评估保留。本轮没有把未复现的 Gemini 带内错误、input stripper、完整 Grok 重试等顺手升级成已确认缺陷。旧 dependency audit 数量不代表当前 finding；本轮未重跑依赖审计，也未重新声明全部依赖无问题。

## 7. 建议落地顺序

1. 建议第二档：固定来源/失败回归 → C01 参数处理与 C02 别名优先级配套适配 → C03 最终 tools 清理 → C04 事件字段与序号 → 组合协议/race/SQLite用量回归。
2. 用户选择新增功能后：Grok 目录诊断优先；带路径端点确有需求时取监控前缀支持。
3. 根据日常统计需求选择窗口 TPS 分布/费用趋势；拥有重置grant或批量生图时再选择对应能力。
4. 缓存/COW/轻量 quota SQL先量收益，Fast、新平台和完整调度机制独立设计。

C01–C03共用转换器和fallback文件，不能因为文件检查 CLEAN 就平行覆盖或逐条直接 cherry-pick。

## 8. 本次实际检查

| 检查 | 结果与证据 |
|---|---|
| 当前版本/提交/7份活动任务 | VERSION1.1.21、Go1.27.3、迁移227、276/276；[baseline](./unmerged-review-2026-10-31/baseline.json) |
| 4簇全文件先反后正 | 25条 A1/C14/X9/N1；仅独立index check，无应用 |
| 既有测试入口确实存在 | `go test -tags=unit -list`列出47个入口；[目录](./unmerged-review-2026-10-31/baseline-test-list.log) |
| 当前正常定向回归 | 3包、47个入口/83个含子用例通过，0失败/跳过；[原始JSONL](./unmerged-review-2026-10-31/baseline-targeted.jsonl) |
| 新隔离探针 | 9个顶层测试按缺口预期失败，合法JSON控制用例通过；[转换器](./unmerged-review-2026-10-31/apicompat-probe.log)、[服务出口](./unmerged-review-2026-10-31/service-probe.log)、[handler](./unmerged-review-2026-10-31/handler-probe.log) |
| 新建议规格/链接/范围 | 结果见[检查索引](./unmerged-review-2026-10-31/README.md)；实施槽保持空白 |

本次没有执行新的完整 unit、race、前端门禁或业务实例验收；这些是未来实现批次的验收任务。使用的测试为当前unit入口，不把PG-only integration作为SQLite证据。

## 9. 本轮结论

- 待归档不等于待实现；也不能从主体完成推断明确排除的尾项已合。
- 最新修复已经覆盖某些邻近路径，旧尾项仍需沿真实生产出口复现；本次找到了四个有直接证据的缺口。
- Claude常规使用率与重置grant是不同能力；模型平均TPS、总token吞吐和单请求窗口分布也不能混称。
- 功能选择按个人用途，价格/模型能力按冻结来源和账号声明分别说明；字段或helper便宜不自动授权整个新系统。
- 现有正常测试通过与新失败探针可以同时成立，不能把前者当作未合缺陷不存在的依据。
