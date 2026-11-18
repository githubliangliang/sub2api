# 上游 v0.2.23 移植评估（v0.2.21 → v0.2.23，含v0.2.22）

评估日期2026-11-18；范围冻结于 2026-11-18T04:29:21.448792+00:00。GitHub API与git fetch确认最新正式版为[v0.2.23](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.23)，发布于11-18 03:51:33 UTC / 11:51:33上海时间；同时覆盖[v0.2.22](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.22)，发布于11-17 02:48:12 UTC / 10:48:12上海时间。

建议**第一档21簇、第二档23簇，选取55个不同PR的既有消费者维护**。当前为**评估完成，两批提案未实施**。优先缓存计量/结算、Claude token轮换、Codex工具/上下文、Grok搜索/错误及首输出换号。新模型、强制/渠道Ultra Fast配置和大规模并发策略仍按需；支付排除。

- [第一档提案](../../openspec/changes/port-upstream-0.2.23-tier1/README.md)
- [第二档提案](../../openspec/changes/port-upstream-0.2.23-tier2/README.md)
- [来源、测量与失败过程](./evidence-0.2.23/MANIFEST.md)

本次只新增评估/提案并更新入口文档。没有产品、VERSION、依赖或schema改动，没有commit/push/release、旧变更归档或业务实例操作。

## 1. 版本、范围与规模

| 项目 | 冻结值 |
|---|---|
| fork HEAD / tree | `7a0b78fe511f17b0f7eb9d26fcf5e4ff083312c7` / `3882c3ce52fb24f6f8d89d782ffc7b48bb1e65e7` |
| VERSION / Go / SQLite / auth snapshot | `1.1.25 / 1.27.3 / 229 / 21` |
| private index tree | `3882c3ce52fb24f6f8d89d782ffc7b48bb1e65e7` |
| v0.2.21 commit | `10ba8f9b4dc2e6023d4a9879ab77b21841babd10` |
| v0.2.22 commit / tag object | `b0fa9f3e6eeab1968a53e1289b69daa4bb40ca2b` / `b11573c9d55d39b6780e1e9f6e193ba2e6ca2321` |
| v0.2.23 commit / tag object | `5af9afb588c4b014b3df9db7d4073af1f8ae01b6` / `ce19a640eab5c4b5ef1827c7c1fcfce49569411a` |
| upstream/main（观察，不作来源） | `2af3f1e15c98c893a89255a037f551046883b8c0` |
| 非merge / 候选 / 净diff | 128 / 78 PR + 4直接提交 / 399文件 +25947 -2093 |
| 候选×文件四态 | 524：ALREADY0 / CLEAN269 / CONFLICT199 / NOBASE56 |


| 范围 | 非merge | PR/直接 | 候选文件态 A/C/X/N | 净diff |
|---|---:|---|---|---|
| v0.2.21..v0.2.22 | 108 | 70/3 | 0/225/157/45 | 340 files changed, 20708 insertions(+), 1652 deletions(-) |
| v0.2.22..v0.2.23 | 20 | 8/1 | 0/44/42/11 | 90 files changed, 5251 insertions(+), 453 deletions(-) |


开始时工作区干净。v0.2.21两档已在v1.1.25；最新fork提交7a0b78fe5恢复用量排行，必须保留。524是候选×文件之和，不是399净文件。相同PR里的新能力只独立分档，不因选取修复就整PR导入。

## 2. 方法、渠道与编号

1. GitHub网页检索工具多次502，改用官方releases API与本地git源码核对。[release元数据](./evidence-0.2.23/releases.json)、[基线](./evidence-0.2.23/baseline.json)保留tag对象、源码commit及main观察值。按tag祖先划范围，不按作者日期漏掉旧分支。
2. 82个候选以first parent→merge整体diff覆盖128个非merge。[candidates.tsv](./evidence-0.2.23/candidates.tsv)、[commits.tsv](./evidence-0.2.23/commits.tsv)、原补丁SHA256保存来源。所有524条先反向、再正向检查同一私有index；未应用产品补丁，真实index不变。
3. [逐文件四态](./evidence-0.2.23/files.tsv)只表达文本可应用性。[5096条新增调用扫描](./evidence-0.2.23/added-call-inventory.tsv)含内置/方法/测试，仅作候选检索；逐簇实际消费者与前置如下。CLEAN不是可编译，NOBASE不是没有缺陷，ALREADY=0不是所有行为都没做过。
4. [193份来源测试记录](./evidence-0.2.23/source-test-inventory.tsv)、[76份本地integration首行](./evidence-0.2.23/integration-build-tags.json)和实际test-list分开。53份integration文件为postgres专用，不冒充SQLite执行。
5. 本地最大迁移229。选定两档不要求新DB列/迁移：5m/1h和账号成本列已有。上游tag实际有两个不同的251文件（PG rollup、force-ultrafast）以及252渠道倍率；不能认为同号只代表一种能力，也不能照抄PG SQL。若另行选择两个Ultra Fast配置，按当时基线重新分配230起的本地编号并验证Ent/auth快照；本轮不生成迁移。
6. 保持SQLite-only/miniredis/simple mode、固定quota、usage dedup、完整Fast/Terra及64MiB/40MiB/120s默认值。评估没有读取真实业务账号，按需项不假定已使用。
7. 发布摘要与最终代码有实质差异：[PR8822来源对照](./evidence-0.2.23/release-source-differences.json)显示b6c21cd5e删除强制缓存计费，8d57f461c在同PR恢复。最终tag仍有ForceCacheBilling；本轮只选稳定前缀哈希，不按摘要改计费。库版本/价格若未独立查官方资料，结论仅是与冻结tag或本地控制一致。

### 全部候选分档

| 候选 | 非merge | A/C/X/N | 主分档与选取范围 |
|---|---:|---|---|
| [6dd5a045b](https://github.com/Wei-Shaw/sub2api/commit/6dd5a045b62e27e5b5b139adef91a171097f639b) | 1 | 0/0/1/0 | 第四档：上游VERSION同步；fork保持自己的1.1.x版本。 |
| [530c5e3a4](https://github.com/Wei-Shaw/sub2api/commit/530c5e3a4f9c73ca2b0881191d0ef8e0d82c5484) | 1 | 0/1/2/1 | 第四档：赞助商/品牌文档，无个人网关移植内容；已在上一轮main观察，仍完整记账。 |
| [#8767](https://github.com/Wei-Shaw/sub2api/pull/8767) | 1 | 0/3/5/0 | 第二档 S23 Go工具链与时钟测试维护 |
| [#8644](https://github.com/Wei-Shaw/sub2api/pull/8644) | 1 | 0/3/0/0 | 第三档：Fable limits[]周额度；本地有被动读数分支但未选Fable目录/价格，实际使用时连同已有消费者重评。 |
| [#8649](https://github.com/Wei-Shaw/sub2api/pull/8649) | 1 | 0/0/2/0 | 第四档：支付宝退款request-id；支付功能个人用途排除，不能因修复小而采用。 |
| [#8651](https://github.com/Wei-Shaw/sub2api/pull/8651) | 1 | 0/1/1/0 | 第四档：语义已具备：a25ce7bed按groupID隔离RPM，跨组/本组overlay对照通过；保留legacy groupID=0适配。 |
| [#8653](https://github.com/Wei-Shaw/sub2api/pull/8653) | 1 | 0/5/3/2 | 第二档 S22 既有流与请求热点窄优化 |
| [#8659](https://github.com/Wei-Shaw/sub2api/pull/8659) | 1 | 0/1/1/0 | 第二档 S03 SQLite 分组倍率/RPM 原子同步 |
| [#8662](https://github.com/Wei-Shaw/sub2api/pull/8662) | 1 | 0/2/1/0 | 第二档 S13 Gemini Chat 结构化响应格式 |
| [#8664](https://github.com/Wei-Shaw/sub2api/pull/8664) | 1 | 0/0/7/2 | 第四档：N/A：openai_codex_account_identity.go及该namespace改写本地不存在，不引入整套改写来修父缓存key。 |
| [#8667](https://github.com/Wei-Shaw/sub2api/pull/8667) | 2 | 0/9/2/0 | 第二档 S22 既有流与请求热点窄优化 |
| [#8673](https://github.com/Wei-Shaw/sub2api/pull/8673) | 1 | 0/4/0/0 | 第二档 S14 Antigravity 兼容思考/等待/空流 |
| [#8675](https://github.com/Wei-Shaw/sub2api/pull/8675) | 1 | 0/4/2/0 | 第二档 S11 Bedrock 错误语义与部分用量 |
| [#8678](https://github.com/Wei-Shaw/sub2api/pull/8678) | 1 | 0/8/0/0 | 第三档：concurrency.ping_interval=0是新控制能力；SSE错误修复不依赖关闭ping，默认不改。 |
| [#8682](https://github.com/Wei-Shaw/sub2api/pull/8682) | 1 | 0/1/1/0 | 第一档 F01 Responses 密封调用内容 |
| [#8684](https://github.com/Wei-Shaw/sub2api/pull/8684) | 1 | 0/2/0/0 | 第一档 F09 登录身份回填空指针 |
| [#8693](https://github.com/Wei-Shaw/sub2api/pull/8693) | 1 | 0/2/5/7 | 第四档：PG分组rollup新一代build/锁/迁移251；本地两核心helper文件缺失，SQLite现有汇总不走该维护链。 |
| [#8698](https://github.com/Wei-Shaw/sub2api/pull/8698) | 1 | 0/2/0/0 | 第一档 F11 账号5h窗口上下界 |
| [#8703](https://github.com/Wei-Shaw/sub2api/pull/8703) | 1 | 0/2/0/0 | 第一档 F02 对话中途 system 权限 |
| [#8723](https://github.com/Wei-Shaw/sub2api/pull/8723) | 1 | 0/2/0/0 | 第一档 F10 5h 滚动保留有效周用量 |
| [#8726](https://github.com/Wei-Shaw/sub2api/pull/8726) | 1 | 0/1/1/0 | 第一档 F21 完成转发后的粘性绑定 |
| [#8727](https://github.com/Wei-Shaw/sub2api/pull/8727) | 1 | 0/1/0/1 | 第二档 S08 原文模型和真实端点观测 |
| [#8728](https://github.com/Wei-Shaw/sub2api/pull/8728) | 1 | 0/1/2/0 | 第二档 S08 原文模型和真实端点观测 |
| [#8729](https://github.com/Wei-Shaw/sub2api/pull/8729) | 2 | 0/10/1/1 | 第二档 S16 冷却账号快照成员 |
| [#8735](https://github.com/Wei-Shaw/sub2api/pull/8735) | 2 | 0/0/1/1 | 第四档：N/A：完整raw OAuth输入metadata清理消费者本地不存在；parseRawJSONView存在不代表该缺陷存在。 |
| [#8736](https://github.com/Wei-Shaw/sub2api/pull/8736) | 4 | 0/3/1/0 | 第一档 F16 白名单精确模型输入 |
| [#8737](https://github.com/Wei-Shaw/sub2api/pull/8737) | 1 | 0/2/2/0 | 第二档 S02 Claude 并发 token 轮换恢复 |
| [#8738](https://github.com/Wei-Shaw/sub2api/pull/8738) | 2 | 0/0/1/2 | 第四档：国内原生厂商quota cell/查询平台缺失，不为前端刷新修复导入平台。 |
| [#8739](https://github.com/Wei-Shaw/sub2api/pull/8739) | 1 | 0/1/2/0 | 第二档 S10 Anthropic 兼容入口 effort 策略 |
| [#8740](https://github.com/Wei-Shaw/sub2api/pull/8740) | 1 | 0/0/1/0 | 第四档：语义已具备：本地Gemini分支已有loading && !usageInfo；上一轮共享刷新留行维护保留。 |
| [#8741](https://github.com/Wei-Shaw/sub2api/pull/8741) | 1 | 0/1/0/3 | 第二档 S08 原文模型和真实端点观测 |
| [#8742](https://github.com/Wei-Shaw/sub2api/pull/8742) | 1 | 0/3/0/0 | 第一档 F05 Claude per-turn beta |
| [#8743](https://github.com/Wei-Shaw/sub2api/pull/8743) | 1 | 0/2/0/0 | 第二档 S23 Go工具链与时钟测试维护 |
| [#8745](https://github.com/Wei-Shaw/sub2api/pull/8745) | 4 | 0/1/2/0 | 第二档 S08 原文模型和真实端点观测 |
| [#8754](https://github.com/Wei-Shaw/sub2api/pull/8754) | 2 | 0/1/2/0 | 第二档 S06 Codex additional_tools 声明 ID |
| [#8757](https://github.com/Wei-Shaw/sub2api/pull/8757) | 1 | 0/1/0/0 | 第一档 F13 Claude 默认测试模型维护 |
| [#8759](https://github.com/Wei-Shaw/sub2api/pull/8759) | 1 | 0/2/0/0 | 第一档 F15 Gemini 非数组 items |
| [#8761](https://github.com/Wei-Shaw/sub2api/pull/8761) | 1 | 0/2/0/0 | 第一档 F03 JSON Schema 默认 strict |
| [#8763](https://github.com/Wei-Shaw/sub2api/pull/8763) | 2 | 0/0/1/2 | 第四档：原生Kimi月度quota parser平台缺失；普通OpenAI API Key转发不消费此模块。 |
| [#8769](https://github.com/Wei-Shaw/sub2api/pull/8769) | 1 | 0/3/0/0 | 第一档 F04 Codex HTTP 外部上下文 |
| [#8775](https://github.com/Wei-Shaw/sub2api/pull/8775) | 1 | 0/4/0/0 | 第一档 F19 设置注入 HTML 与 CSP nonce |
| [#8779](https://github.com/Wei-Shaw/sub2api/pull/8779) | 1 | 0/14/1/0 | 第三档：Fable7d_oi恢复；现有model冷却分支可被自定义映射到达，但未确认使用，不能自动引入Fable新能力。 |
| [#8786](https://github.com/Wei-Shaw/sub2api/pull/8786) | 1 | 0/3/0/1 | 第一档 F20 Chat 客户端 metadata.user_id |
| [#8788](https://github.com/Wei-Shaw/sub2api/pull/8788) | 1 | 0/2/1/0 | 第一档 F06 count_tokens 请求头唯一性 |
| [#8790](https://github.com/Wei-Shaw/sub2api/pull/8790) | 1 | 0/1/1/0 | 第四档：只扩大到Zhipu/CommandCode/OpenCodeGo/MiniMax原生平台；现有Anthropic/OpenAI模式已生效，缺平台分支不导入。 |
| [#8796](https://github.com/Wei-Shaw/sub2api/pull/8796) | 2 | 0/3/3/0 | 第二档 S09 Composite 公开模型名还原 |
| [#8798](https://github.com/Wei-Shaw/sub2api/pull/8798) | 1 | 0/1/1/0 | 第二档 S19 价格启动不阻塞远端哈希 |
| [#8800](https://github.com/Wei-Shaw/sub2api/pull/8800) | 2 | 0/6/0/0 | 第二档 S20 上游 TCP 死连接检测 |
| [#8806](https://github.com/Wei-Shaw/sub2api/pull/8806) | 1 | 0/8/3/0 | 第二档 S18 实际扣除与账号成本统计 |
| [#8808](https://github.com/Wei-Shaw/sub2api/pull/8808) | 1 | 0/2/0/0 | 第一档 F12 首输出临时文件不可用 |
| [#8811](https://github.com/Wei-Shaw/sub2api/pull/8811) | 2 | 0/4/3/1 | 第二档 S04 首内容前换号与 keepalive 错误 |
| [#8813](https://github.com/Wei-Shaw/sub2api/pull/8813) | 1 | 0/2/2/0 | 第二档 S04 首内容前换号与 keepalive 错误 |
| [#8814](https://github.com/Wei-Shaw/sub2api/pull/8814) | 1 | 0/3/0/0 | 第一档 F07 输出后的兼容流账号副作用 |
| [#8825](https://github.com/Wei-Shaw/sub2api/pull/8825) | 1 | 0/0/3/1 | 第四档：N/A/对照通过：本地没有强制DeepSeek峰谷覆盖策略；现有override_file的0/正数/缓存读价均保留，不补源强制策略。 |
| [#8826](https://github.com/Wei-Shaw/sub2api/pull/8826) | 2 | 0/3/2/0 | 第二档 S14 Antigravity 兼容思考/等待/空流 |
| [#8828](https://github.com/Wei-Shaw/sub2api/pull/8828) | 4 | 0/2/2/0 | 第二档 S12 Gemini 并行与分片工具 |
| [#8837](https://github.com/Wei-Shaw/sub2api/pull/8837) | 1 | 0/1/1/0 | 第一档 F18 调度快照端点能力 |
| [#8840](https://github.com/Wei-Shaw/sub2api/pull/8840) | 1 | 0/2/0/0 | 第一档 F17 渠道 Bedrock 兼容开关类型 |
| [#8746](https://github.com/Wei-Shaw/sub2api/pull/8746) | 1 | 0/1/2/1 | 第三档：Gemini4 Hydrogen兜底价是新模型覆盖；通用映射可能到达，按实际使用选择并核价。 |
| [#8827](https://github.com/Wei-Shaw/sub2api/pull/8827) | 3 | 0/3/12/5 | 第三档：Vertex us/eu多区域和1.1倍率是新端点/计价策略，需按实际账户与区域验证。 |
| [#8822](https://github.com/Wei-Shaw/sub2api/pull/8822) | 2 | 0/2/1/0 | 第二档 S15 Gemini 原生稳定前缀会话 |
| [#8816](https://github.com/Wei-Shaw/sub2api/pull/8816) | 3 | 0/15/9/0 | 第三档：按用户lane批量账务与广泛零拷贝；源新增544行PGbatch repo+311行combiner，须SQLite事务/dedup/取消/1C1G独立测量。 |
| [#8833](https://github.com/Wei-Shaw/sub2api/pull/8833) | 1 | 0/3/3/1 | 第三档：大bucket合并加载+200ms last_used缓存+20s stale grace为新性能策略；S15快照正确性不需要它，延迟/内存先量测。 |
| [#8661](https://github.com/Wei-Shaw/sub2api/pull/8661) | 1 | 0/5/13/3 | 第三档：Gemini4 Carbon及完整思考策略新能力；现有Gemini修复不依赖整批目录/价格。 |
| [#8834](https://github.com/Wei-Shaw/sub2api/pull/8834) | 1 | 0/8/3/0 | 第二档 S07 Grok 服务端搜索调用 |
| [#8748](https://github.com/Wei-Shaw/sub2api/pull/8748) | 2 | 0/8/6/3 | 第三档：透传模型白名单过滤是新开关/schema外配置；F16输入校验不自动激活此策略。 |
| [#8724](https://github.com/Wei-Shaw/sub2api/pull/8724) | 1 | 0/2/0/0 | 第二档 S04 首内容前换号与 keepalive 错误 |
| [#8772](https://github.com/Wei-Shaw/sub2api/pull/8772) | 4 | 0/0/3/0 | 第二档 S21 OpenAI 工具 schema 正则适配 |
| [#8669](https://github.com/Wei-Shaw/sub2api/pull/8669) | 2 | 0/6/0/0 | 第一档 F14 Grok 媒体扁平错误 |
| [#8747](https://github.com/Wei-Shaw/sub2api/pull/8747) | 3 | 0/4/5/2 | 第二档 S17 OpenAI 用量观测新鲜度 |
| [9cd1a1eb2](https://github.com/Wei-Shaw/sub2api/commit/9cd1a1eb21e84707685e1998dcb5a3d1ceafae49) | 1 | 0/1/2/1 | 第四档：赞助商/品牌文档，不同步fork产品说明。 |
| [#8812](https://github.com/Wei-Shaw/sub2api/pull/8812) | 4 | 0/17/25/3 | 第一档 F08 既有 Sol Ultra Fast 倍率；group force-ultrafast为第三档。 |
| [#7540](https://github.com/Wei-Shaw/sub2api/pull/7540) | 2 | 0/2/0/1 | 第三档：Qwen3 embedding/reranker新兜底目录；包含CNY折算假设，按实际模型/站点核价后选择。 |
| [15871f904](https://github.com/Wei-Shaw/sub2api/commit/15871f904056f548dd874a1d069ffc6cdf777b78) | 1 | 0/0/1/0 | 第四档：上游VERSION同步，不改变fork版本。 |
| [#8906](https://github.com/Wei-Shaw/sub2api/pull/8906) | 2 | 0/11/14/0 | 第三档：GPT-6.2 Sol/Astra、272k计价、Pro套餐/Ultrafast促销时钟为新能力，旧Sol倍率修复独立。 |
| [#8904](https://github.com/Wei-Shaw/sub2api/pull/8904) | 1 | 0/3/1/0 | 第二档 S22 既有流与请求热点窄优化；只选Ollama，DeepSeek缺失消费者不导入。 |
| [#8903](https://github.com/Wei-Shaw/sub2api/pull/8903) | 1 | 0/2/1/0 | 第二档 S22 既有流与请求热点窄优化 |
| [#8598](https://github.com/Wei-Shaw/sub2api/pull/8598) | 1 | 0/0/0/1 | 第四档：N/A：billing_inflight_reservation_test.go及其在途预占基座本地不存在，不为heap测试引入新能力。 |
| [#8914](https://github.com/Wei-Shaw/sub2api/pull/8914) | 1 | 0/5/13/4 | 第三档：新增渠道Ultra Fast倍率及PG252迁移，是新配置能力；现有Fast倍率不受本评估扩展。 |
| [#8577](https://github.com/Wei-Shaw/sub2api/pull/8577) | 1 | 0/1/1/0 | 第二档 S05 空闲超时与一次换号预算 |
| [#8174](https://github.com/Wei-Shaw/sub2api/pull/8174) | 11 | 0/22/10/5 | 第二档 S01 5m/1h 缓存明细与互斥计费 |
| [#8920](https://github.com/Wei-Shaw/sub2api/pull/8920) | 1 | 0/0/1/1 | 第二档 S05 空闲超时与一次换号预算 |

## 3. 第一档：窄修复和默认值维护

### F01 Responses 密封调用内容（[#8682](https://github.com/Wei-Shaw/sub2api/pull/8682)）

- **依据/最小前置**：apicompat overlay 丢失 sealed-content；namespace/密封参数对照通过。 本地 accumulator 完整存在；只缺一个字段及两个赋值。
- **patch sites**：backend/internal/pkg/apicompat/responses_to_chatcompletions.go: bufferedFuncCall、ProcessEvent、BuildOutput。
- **行为与验收**：函数调用的 delta 回退重建 SHALL 保留 encrypted_content。 验收：terminal/item.done/delta 三路比对调用 ID、密封内容和原参数。。
- **边界**：保留上一轮已实现的 namespace、encrypted_function_args 与参数完成事件语义。

### F02 对话中途 system 权限（[#8703](https://github.com/Wei-Shaw/sub2api/pull/8703)）

- **依据/最小前置**：MidSystemAuthority overlay 实际 role=user。 既有 user/assistant 转换存在；新增窄 system 分支。
- **patch sites**：backend/internal/pkg/apicompat/anthropic_to_responses.go: anthropicMsgToResponsesItems。
- **行为与验收**：Chat/Anthropic 中途 system 消息转 Responses SHALL 保持 developer 权限。 验收：首部系统提示及中途 hook 文本、工具历史的权限与顺序。。
- **边界**：不把用户文本或工具结果提升为系统权限。

### F03 JSON Schema 默认 strict（[#8761](https://github.com/Wei-Shaw/sub2api/pull/8761)）

- **依据/最小前置**：StrictDefault overlay 缺 strict；显式 true 对照通过。 既有格式转换函数存在；五行默认值维护。
- **patch sites**：backend/internal/pkg/apicompat/response_format.go: chatResponseFormatToResponsesTextFormat。
- **行为与验收**：Chat json_schema 未设置 strict 时转 Responses SHALL 显式写 false。 验收：缺省/true/false/null、schema 原值和普通 text/json_object 对照。。
- **边界**：不覆盖客户端显式值，不新增结构化输出能力。

### F04 Codex HTTP 外部上下文（[#8769](https://github.com/Wei-Shaw/sub2api/pull/8769)）

- **依据/最小前置**：handler overlay browser_info 未转换；真实 call_id 对照通过。 上一轮 page normalizer 已存在；扩大名称集合，不导入账号命名空间。
- **patch sites**：backend/internal/handler/openai_page_context.go、openai_gateway_handler.go: normalizeCodexPageContext/Responses。
- **行为与验收**：无对应调用的 agent browser_info/goal_objective/untrusted_context 输出 SHALL 转为不可信用户文本。 验收：三种上下文及已有 page 快照、空输出、重复成员、真实调用、未知名称。。
- **边界**：保留 page 快照严格校验和 user 权限，不能吞正常 function_call_output。

### F05 Claude per-turn beta（[#8742](https://github.com/Wei-Shaw/sub2api/pull/8742)）

- **依据/最小前置**：PerTurnBeta overlay 被净化；原 beta 和无 beta 对照通过。 现有 beta 净化函数存在；新增常量及一次精确 token 判断。
- **patch sites**：backend/internal/pkg/claude/constants.go；backend/internal/service/gateway_request.go: stripAnthropicMessageOutputConfigUnlessBeta。
- **行为与验收**：任一已支持的 per-turn/mid-conversation beta SHALL 保留消息级 output_config。 验收：API Key/OAuth、空 system 控制消息、有正文消息和顶层 output_config。。
- **边界**：无相关 beta 仍执行既有净化，不能放开全部消息字段。

### F06 count_tokens 请求头唯一性（[#8788](https://github.com/Wei-Shaw/sub2api/pull/8788)）

- **依据/最小前置**：修正 DI 夹具后 header overlay 复现同名头计数为2。 getHeaderRaw/setHeaderRaw 已存在；不能改回 net/http 的大小写假设。
- **patch sites**：backend/internal/service/gateway_count_tokens.go: buildCountTokensRequestAnthropicAPIKeyPassthrough；gateway_service.go 调试头读取。
- **行为与验收**：Anthropic count_tokens 透传 SHALL 以原始大小写无关读取/写入避免重复默认头。 验收：真实 HTTP 线格式大小写/多值、覆写和缺省头各只有预期值。。
- **边界**：保持账号鉴权、beta 净化与覆写的最终优先级。

### F07 输出后的兼容流账号副作用（[#8814](https://github.com/Wei-Shaw/sub2api/pull/8814)）

- **依据/最小前置**：静态：两 reader 在早期 failover 分支之后直接记录错误，缺副作用调用。 handleOpenAIStreamTerminalAccountSideEffects 和模型规则已有消费者。
- **patch sites**：backend/internal/service/openai_gateway_chat_completions.go: handleChatStreamingResponse；openai_gateway_messages.go: handleAnthropicStreamingResponse。
- **行为与验收**：OpenAI Chat/Messages 已输出后的流错误 SHALL 执行既有账号/模型错误副作用。 验收：文本/工具输出后 401/429/overload、模型规则、普通请求错误与取消。。
- **边界**：已输出后不跨账号重放；沿用模型作用域，不扩大整号冷却。

### F08 既有 Sol Ultra Fast 倍率（[#8812](https://github.com/Wei-Shaw/sub2api/pull/8812)）

- **依据/最小前置**：SolUltrafast overlay 实际缺模型倍率；Astra6倍对照通过。 本地完整 Fast 已实现；Sol 型号 helper 已有，源新增支持谓词可窄适配。
- **patch sites**：backend/internal/service/billing_service.go: applyModelSpecificPricingPolicy；openai_model_alias.go 既有型号判断。
- **行为与验收**：已支持的 GPT-6.1 Sol Ultra Fast SHALL 使用冻结 tag 的6倍模型倍率。 验收：Sol/Astra、别名、HTTP/WS结算、Fast2倍、Free Fast、自定义0和长期上下文。。
- **边界**：只摘既有型号的倍率修复；强制 Ultra Fast 分组、新渠道倍率和 GPT-6.2 不随行。

### F09 登录身份回填空指针（[#8684](https://github.com/Wei-Shaw/sub2api/pull/8684)）

- **依据/最小前置**：LoginNilIdentity overlay 捕获 panic；历史回填来源对照通过。 函数在 nil 检查前读 Metadata；三行 guard 即可。
- **patch sites**：backend/internal/service/auth_service.go: shouldApplyEmailFirstBindDefaults。
- **行为与验收**：登录邮箱身份回填失败返回 nil 时 SHALL 安全跳过默认设置判断。 验收：身份创建/查询失败、既有身份、回填/首次绑定来源与正常登录。。
- **边界**：不引入新的注册、赠额或支付策略。

### F10 5h 滚动保留有效周用量（[#8723](https://github.com/Wei-Shaw/sub2api/pull/8723)）

- **依据/最小前置**：WeeklySurvivesFiveHourRollover overlay 清空未到期7d读数。 既有通用7d被动采样存在；源 rollover helper30行可独立适配。
- **patch sites**：backend/internal/service/ratelimit_service.go: UpdateSessionWindow。
- **行为与验收**：Claude 5h 窗口滚动 SHALL 仅清除已过期的周读数。 验收：未来/过期/缺 reset、仅7d响应、秒/毫秒重置和样本时间。。
- **边界**：只清5h和过期周值；Fable 模型/新额度接口仍独立按需。

### F11 账号5h窗口上下界（[#8698](https://github.com/Wei-Shaw/sub2api/pull/8698)）

- **依据/最小前置**：实际 modernc SQLite/Ent overlay 更新 end 后 start 仍在旧窗口。 本地 Ent更新和 outbox 存在；无 schema 前置。
- **patch sites**：backend/internal/repository/account_repo.go: UpdateSessionWindowEnd。
- **行为与验收**：主动读取更新5h窗口 end 时 SHALL 同步 start=end-5h。 验收：正常滚动、毫秒格式读入、outbox传播和新窗口费用查询。。
- **边界**：账号窗口修复不改变用户平台固定 quota 窗口、软删除或重置策略。

### F12 首输出临时文件不可用（[#8808](https://github.com/Wei-Shaw/sub2api/pull/8808)）

- **依据/最小前置**：SpoolUnavailable overlay 返回错误；stage上限拒绝对照通过。 既有 unlink 失败回退已存在；创建失败缺同样处理。
- **patch sites**：backend/internal/service/openai_first_output_timeout.go: openAIFirstOutputStage.prepareWrite。
- **行为与验收**：首输出暂存无法创建临时文件时 SHALL 在同一容量上限内回退内存。 验收：无权限/目录不可用、commit/drop/取消、上限拒绝及清理。。
- **边界**：保留8MiB stage上限、无具名明文残留和预算释放。

### F13 Claude 默认测试模型维护（[#8757](https://github.com/Wei-Shaw/sub2api/pull/8757)）

- **依据/最小前置**：静态常量差异；基线相关测试通过，未调用真实 Anthropic 核实退役公告。 本地4-6已有兼容和价格消费者；仅改常量及相关期望。
- **patch sites**：backend/internal/pkg/claude/constants.go: DefaultTestModel。
- **行为与验收**：内置 Claude 默认测试模型 SHALL 与冻结 tag 的 claude-sonnet-4-6 一致。 验收：默认账号测试使用已有4-6；显式模型和其它平台默认值不被覆盖。。
- **边界**：属于默认值维护，不新增模型目录/定价，也不宣称独立市场验证。

### F14 Grok 媒体扁平错误（[#8669](https://github.com/Wei-Shaw/sub2api/pull/8669)）

- **依据/最小前置**：GrokFlatMediaMessage overlay 丢失字符串；嵌套/未知对象对照通过。 源两个窄提取器和带提取器的规则入口需新增；既有媒体路由已存在。
- **patch sites**：backend/internal/service/grok_upstream_errors.go、grok_media.go、error_passthrough_runtime.go。
- **行为与验收**：Grok 既有媒体错误 SHALL 保留 xAI 扁平 error/code，保持脱敏与可配置错误规则。 验收：嵌套优先、扁平字符串/数字code、内容策略、自定义规则和未知对象。。
- **边界**：仅 Grok 扩展提取器；不改变其它平台的模型不存在/停号判断。

### F15 Gemini 非数组 items（[#8759](https://github.com/Wei-Shaw/sub2api/pull/8759)）

- **依据/最小前置**：SchemaNonArrayItems overlay 失败；tuple/default及合法数组对照已经通过。 清理器完整存在；窄类型条件修复。
- **patch sites**：backend/internal/pkg/antigravity/schema_cleaner.go: cleanJSONSchemaRecursive。
- **行为与验收**：既有 Gemini schema 清理 SHALL 删除非 array 的 items，并保持合法数组 schema。 验收：嵌套对象/数组、非法items、工具schema及实例 default/enum 的保护。。
- **边界**：不把本地已经处理的 tuple 情形描述成新缺陷，不引入新 Gemini 模型。

### F16 白名单精确模型输入（[#8736](https://github.com/Wei-Shaw/sub2api/pull/8736)）

- **依据/最小前置**：临时前端4个来源新增测试失败；原11个控件测试通过。 共享控件和提示API已有；来源测试需按本地位置插入。
- **patch sites**：frontend/src/components/account/ModelWhitelistSelector.vue；zh/en admin/accounts 文案。
- **行为与验收**：模型白名单添加操作 SHALL 拒绝含星号的条目并保留输入草稿。 验收：点击/Enter、首尾空白、纠正后添加、重复/映射冲突提示。。
- **边界**：不引入 glob 或透传白名单调度新开关。

### F17 渠道 Bedrock 兼容开关类型（[#8840](https://github.com/Wei-Shaw/sub2api/pull/8840)）

- **依据/最小前置**：临时前端5个来源测试失败，实际保存的是 {anthropic:true/false}。 后端读 boolean、本地前端写对象；现有开关消费者已核实。
- **patch sites**：frontend/src/views/admin/ChannelsView.vue: formToAPI。
- **行为与验收**：渠道编辑 SHALL 用 boolean 保存/清除 bedrock_cc_compat，并保留其它 features。 验收：创建/编辑/重开/显式false/移除Anthropic区段及其它features。。
- **边界**：保留本地 run_mode/菜单限制，不新增平台或商业配置。

### F18 调度快照端点能力（[#8837](https://github.com/Wei-Shaw/sub2api/pull/8837)）

- **依据/最小前置**：repository overlay 丢键；实际 Account 能力判断由禁用 embeddings 变成允许。 键读取/普通端点过滤存在；只需补投影白名单。
- **patch sites**：backend/internal/repository/scheduler_cache.go: filterSchedulerCredentials；既有 Account/调度器能力消费者。
- **行为与验收**：调度快照 SHALL 保留已有 openai_capabilities 声明并据此过滤既有端点。 验收：chat/embeddings/Responses生图意图、未配置、空配置、hydrate及WS读取。。
- **边界**：来源标题提及 Seedance，但本地现有能力消费者独立适用；不引入 Seedance。

### F19 设置注入 HTML 与 CSP nonce（[#8775](https://github.com/Wei-Shaw/sub2api/pull/8775)）

- **依据/最小前置**：embed overlay 复现nonce仍304、缺no-store；无nonce304对照通过。 现有HTML cache及注入存在；18行头helper可独立采用。
- **patch sites**：backend/internal/web/embed_on.go: serveIndexHTML；新增窄HTML头 helper。
- **行为与验收**：设置注入 HTML SHALL 禁止共享/浏览器缓存，携带请求 nonce 时不得按旧 ETag 返回304。 验收：两请求不同nonce、设置变更/回退、无nonce条件请求、静态hash资源。。
- **边界**：不承诺所有浏览器 bfcache 行为；仅约定响应头和 nonce/HTML 一致性。

### F20 Chat 客户端 metadata.user_id（[#8786](https://github.com/Wei-Shaw/sub2api/pull/8786)）

- **依据/最小前置**：静态：DTO转换丢metadata，两个转发路径没有补回；尚无专用转发探针。 ParseMetadataUserID存在；源helper32行，不需账号命名空间基座。
- **patch sites**：backend/internal/service/gateway_forward_as_chat_completions.go、openai_gateway_chat_completions_anthropic_native.go；窄metadata helper。
- **行为与验收**：API Key 的 Chat转Anthropic SHALL 补回已识别格式的客户端 metadata.user_id。 验收：合法JSON/legacy会话ID、空值/错误类型、两路径和OAuth负对照。。
- **边界**：只保留识别到的user_id字符串原字节，OAuth仍使用账号拟态身份。

### F21 完成转发后的粘性绑定（[#8726](https://github.com/Wei-Shaw/sub2api/pull/8726)）

- **依据/最小前置**：静态：handler用请求ctx绑定；service原样交给cache，无内部脱离取消。 WithoutCancel/WithTimeout和既有绑定helper均已有；无需新调度架构。
- **patch sites**：backend/internal/handler/gateway_handler.go: Messages 完成后的绑定点。
- **行为与验收**：完成转发后的粘性写入 SHALL 不随客户端取消，且受5秒截止限制。 验收：完成后断开仍写入、超时返回、换号不覆盖原绑定和正常请求。。
- **边界**：保留本地 profit admission/既有绑定条件；只脱离这一收尾写入。

## 4. 第二档：跨路径与SQLite适配

### S01 5m/1h 缓存明细与互斥计费（[#8174](https://github.com/Wei-Shaw/sub2api/pull/8174)）

- **依据/最小前置**：DTO往返/presence及缓存差额overlay失败；完整明细/显式0对照通过。 5m/1h UsageTokens/SQLite列已有；OpenAIUsage及Anthropic DTO/流合并需手工贯通。
- **patch sites**：backend/internal/pkg/apicompat/types.go及三协议桥；backend/internal/service/*native*、gateway_*usage*、billing_service.go、account_stats_pricing.go。
- **行为与验收**：既有转换、读流及结算 SHALL 保留5m/1h明细和聚合presence，差额按5m计价，输入桶互斥。 验收：HTTP/普通与透传SSE/缓冲/WS、缓存字段单独更新、矛盾明细、1h费用、SQLite结算+去重。。
- **边界**：只用既有5m/1h列；不导入30m/12h字段、迁移或未选价格策略；自定义0/Fast保留。

### S02 Claude 并发 token 轮换恢复（[#8737](https://github.com/Wei-Shaw/sub2api/pull/8737)）

- **依据/最小前置**：静态追踪三转发路径；上一轮原子refresh仍存在，未跑新401并发传输探针。 仓储GetByID/token cache已有；31行恢复helper及三路径重发适配。
- **patch sites**：backend/internal/service/gateway_forward.go、gateway_forward_as_chat_completions.go、gateway_forward_as_responses.go。
- **行为与验收**：Claude OAuth 的旧token401 SHALL 仅在数据库证明有新有效token时同号重试一次。 验收：barrier控制旧token在途/新token落库、cache旧值、一次重试、同token/错误类型/取消/预算。。
- **边界**：比对实际发送token；不在401路径刷新/持久化凭据，不绕过重试上限或正常401策略。

### S03 SQLite 分组倍率/RPM 原子同步（[#8659](https://github.com/Wei-Shaw/sub2api/pull/8659)）

- **依据/最小前置**：两份真实modernc SQLite trigger故障探针复现旧行丢失；零倍率成功对照通过。 本地已有SQLite SQL；增加事务与去重并保持本地SQL构造，无schema前置。
- **patch sites**：backend/internal/repository/user_group_rate_repo.go: SyncGroupRateMultipliers/SyncGroupRPMOverrides。
- **行为与验收**：单次分组倍率或RPM同步 SHALL 在一个事务内清空/删除/写入，失败保留旧配置。 验收：事务失败/取消回滚、重复用户last-wins、另一列保留、bind-safe大批次、提交后cache。。
- **边界**：不得复制PG ANY/ALL/unnest/强制类型转换；不更改历史迁移或用户固定quota。

### S04 首内容前换号与 keepalive 错误（[#8811](https://github.com/Wei-Shaw/sub2api/pull/8811)、[#8813](https://github.com/Wei-Shaw/sub2api/pull/8813)、[#8724](https://github.com/Wei-Shaw/sub2api/pull/8724)）

- **依据/最小前置**：SSEErrorAfterPing overlay仍写裸JSON；源码确认Messages前奏立即写出和keepalive被stage扣住。 本地keepalive调整大小、首输出预算/规则/暂存已有；hold与耗尽规则需要适配。
- **patch sites**：backend/internal/service/openai_gateway_messages.go、openai_gateway_response_handling.go、gateway_upstream_response.go；backend/internal/handler/openai_gateway_handler.go。
- **行为与验收**：首内容前仅keepalive输出 SHALL 不阻断安全换号；已提交SSE后的错误必须为SSE事件。 验收：created/item-added后失败→成功账号、keepalive前后耗尽、错误规则、取消/终态/槽位释放。。
- **边界**：前奏需沿用8MiB有界暂存，不能原样引入无界strings.Builder；私有账号头不得提前泄露。

### S05 空闲超时与一次换号预算（[#8577](https://github.com/Wei-Shaw/sub2api/pull/8577)、[#8920](https://github.com/Wei-Shaw/sub2api/pull/8920)）

- **依据/最小前置**：两PR最终代码联合静态审查；未执行新超时/两账号HTTP探针。 SafeToFailoverAfterWrite及一次预算已有；不可只取#8577漏#8920。
- **patch sites**：backend/internal/service/openai_gateway_response_handling.go: interval计时/timeout；既有handler首输出预算。
- **行为与验收**：OpenAI 首输出前空闲超时 SHALL 进入最多一次换号预算，计时按最后上游读取截止。 验收：持续数据/注释、无内容超时、keepalive已写、两号都超时、已有内容与Grok负对照。。
- **边界**：保留配置的idle默认及既有模型冷却；不要把惩罚扩散到整个池。

### S06 Codex additional_tools 声明 ID（[#8754](https://github.com/Wei-Shaw/sub2api/pull/8754)）

- **依据/最小前置**：AdditionalToolsDeclarationID overlay 实际id缺失。 decoded过滤器已有；raw局部输入重建helper/实际调用点需补，不覆盖整来源文件。
- **patch sites**：backend/internal/service/openai_codex_transform.go: filterCodexInputWithOptions；现有HTTP/透传/WS兼容调用点。
- **行为与验收**：Codex additional_tools 的声明ID SHALL 与引用一起保留，仅规范已识别UUID形态。 验收：UUID声明/引用、opaque ID、store=false、完整工具回放、普通callID/UTF8。。
- **边界**：不剥离声明ID、不改普通工具身份；不导入完整账号命名空间或ojson基座。

### S07 Grok 服务端搜索调用（[#8834](https://github.com/Wei-Shaw/sub2api/pull/8834)）

- **依据/最小前置**：GrokServerSearchCount overlay 实际0期望2；普通custom/web_search对照通过。 源24行predicate可复用；需贯通state/accumulator/provider标记，不只补计数。
- **patch sites**：backend/internal/pkg/apicompat/responses_to_anthropic.go、responses_to_chatcompletions.go；backend/internal/service/grok_search_count.go及流/缓冲消费者。
- **行为与验收**：Grok xs_/as_ 服务端 custom 搜索 SHALL 不下发为客户端调用，并计入原生搜索次数。 验收：Grok Chat/Messages/Responses、终态与delta回退、搜索去重、普通custom工具和非Grok前缀。。
- **边界**：仅Grok启用prefix规则；不引入通用工具桥或新搜索平台。

### S08 原文模型和真实端点观测（[#8727](https://github.com/Wei-Shaw/sub2api/pull/8727)、[#8728](https://github.com/Wei-Shaw/sub2api/pull/8728)、[#8745](https://github.com/Wei-Shaw/sub2api/pull/8745)、[#8741](https://github.com/Wei-Shaw/sub2api/pull/8741)）

- **依据/最小前置**：ChatFallbackResponseModel overlay丢模型；现有observer/response_model计费消费者已定位。 本地observer已有约束；源部分ServiceTier观察接口缺失，按本地结果字段适配。
- **patch sites**：backend/internal/service/openai_gateway_*anthropic_native.go、*chat_fallback.go、gateway_forward_as_*.go；既有upstream_response_model.go。
- **行为与验收**：原生Anthropic和Chat/Responses桥 SHALL 把原文模型及冲突带到用量边界，并记录真实/messages端点。 验收：流/非流/中断、多模型冲突、一次attempt隔离、端点错误日志、SQLite计费回退。。
- **边界**：先观察原文再还原公开名；保留本地价格来源准入及已解析Fast tier，不为模型修复重做tier基座。

### S09 Composite 公开模型名还原（[#8796](https://github.com/Wei-Shaw/sub2api/pull/8796)）

- **依据/最小前置**：静态核对本地公开/已解析模型ctx；当前Anthropic响应仍使用转发模型，未跑新别名HTTP探针。 本地RequestedPublicModel/ResolvedUpstreamModel上下文已有；源窄helper需适配。
- **patch sites**：backend/internal/service/composite_platform.go、gateway_anthropic_passthrough.go、gateway_upstream_response.go、openai_gateway_messages*.go。
- **行为与验收**：Composite 响应 SHALL 还原客户端公开模型名，同时保留原文上游模型用于审计/计费。 验收：原生/拟态/Chat/Responses的流与非流、仅message_start.model、模型冲突及SQLite费用。。
- **边界**：先观测后改名；不得替换正文/工具参数中的同名字符串，不影响非Composite响应。

### S10 Anthropic 兼容入口 effort 策略（[#8739](https://github.com/Wei-Shaw/sub2api/pull/8739)）

- **依据/最小前置**：静态：通用OpenAI入口有policy，两个Anthropic兼容handler没有调用。 applyOpenAIReasoningEffortPolicyForRequest已有；本地handler次序需手工适配。
- **patch sites**：backend/internal/handler/gateway_handler_chat_completions.go、gateway_handler_responses.go。
- **行为与验收**：Anthropic组的Chat/Responses兼容入口 SHALL 执行现有group effort策略并一致返回错误。 验收：deny/map/ceiling、显式非法值、禁用/省略、转换后最终effort和费用、count_tokens邻近。。
- **边界**：保留上一轮合成none/最终校验和GLM通用兼容；不复制缺失原生平台闸门。

### S11 Bedrock 错误语义与部分用量（[#8675](https://github.com/Wei-Shaw/sub2api/pull/8675)）

- **依据/最小前置**：静态核对既有Bedrock路由、部分流用量helper和模型限流仓储；未跑新完整错误探针。 110行语义helper及调度键ctx需适配；不导入新Bedrock平台/模型。
- **patch sites**：backend/internal/service/gateway_bedrock.go、gateway_upstream_response.go、ratelimit_service.go；窄ratelimit_bedrock helper；Messages调度模型ctx。
- **行为与验收**：Bedrock SHALL 区分封号/缺用例/请求过滤/模型429，并在中断时返回已观测用量。 验收：精确400/404、反射字段负对照、429多模型、渠道映射键、5xx过滤、截断SQLite结算。。
- **边界**：请求内容错误不换号；只模型冷却，沿用现有显式规则及SQLite单键更新。

### S12 Gemini 并行与分片工具（[#8828](https://github.com/Wei-Shaw/sub2api/pull/8828)）

- **依据/最小前置**：静态：两个现有reader只以name判断边界；65行来源helper已量测。 现有delta累计器存在；两reader添加统一boundary，不需要新模型。
- **patch sites**：backend/internal/service/gemini_messages_compat_service.go、gemini_chat_completions_compat_service.go。
- **行为与验收**：Gemini流 SHALL 按调用身份而非仅名称划分工具块，保留字符串分片和终态快照。 验收：同名/同参并行、不同/相同ID、完整首片字符串、重传/快照、无ID歧义及结束事件。。
- **边界**：无ID的字符串快照与新相同调用不可区分时采用来源保守策略，不猜造调用。

### S13 Gemini Chat 结构化响应格式（[#8662](https://github.com/Wei-Shaw/sub2api/pull/8662)）

- **依据/最小前置**：静态：当前Chat→Anthropic转换后丢response_format；真实消费者存在。 源helper50行；CLI wrapper需原文嵌入而非整请求map重编码。
- **patch sites**：backend/internal/service/gemini_chat_completions_compat_service.go；窄gemini_compat_response_format helper。
- **行为与验收**：Gemini Chat兼容 SHALL 传递已声明json_object/json_schema格式，并保留schema字面量。 验收：两个既有认证入口、属性顺序/大数、工具/图片互斥、文本/缺省/无效格式。。
- **边界**：采用responseJsonSchema，不扩展Gemini新模型或cache CRUD；保留资源限制。

### S14 Antigravity 兼容思考/等待/空流（[#8673](https://github.com/Wei-Shaw/sub2api/pull/8673)、[#8826](https://github.com/Wei-Shaw/sub2api/pull/8826)）

- **依据/最小前置**：静态核对现有adapter/keepalive；专用时间锚和sticky释放尚未执行。 adapter和thinking意图parser已有；源16行effort helper可摘，缺配置helper需窄补。
- **patch sites**：backend/internal/service/antigravity_gateway_compat*.go、antigravity_gemini_thinking_variant.go；两兼容handler错误收尾。
- **行为与验收**：既有Antigravity兼容 SHALL 对齐思考配置、按响应头到达锚定等待，空流仅释放对应粘性会话。 验收：slow headers/有效thought/连续ping/empty timeout、无内容已提交、绑定已改变、取消/用量。。
- **边界**：保留既有timeout默认，不导入Carbon等模型和完整新thinking策略；绑定检查需防并发误删。

### S15 Gemini 原生稳定前缀会话（[#8822](https://github.com/Wei-Shaw/sub2api/pull/8822)）

- **依据/最小前置**：静态：本地仍通用ParseGatewayRequest回退；完整PR最后恢复ForceCacheBilling。 内容种子归一和xxhash已有相邻消费者；缺helper需适配，不能套来源未有的完整raw基座。
- **patch sites**：backend/internal/handler/gemini_v1beta_handler.go；backend/internal/service/gemini_session.go。
- **行为与验收**：Gemini原生会话回退 SHALL 按API Key、模型、tools/system和首用户parts生成稳定哈希。 验收：首用户/省略role、后续tool回合、模型/Key隔离、空parts、显式客户端ID优先。。
- **边界**：不按发布摘要删除ForceCacheBilling；两个微提交必须以最终PR状态解释。

### S16 冷却账号快照成员（[#8729](https://github.com/Wei-Shaw/sub2api/pull/8729)）

- **依据/最小前置**：静态：本地loadAccountsFromDB仍用ListSchedulable*；完整候选仓储接口已存在。 ListModelAvailabilityCandidates已有；差异是缓存成员与读取过滤。
- **patch sites**：backend/internal/service/scheduler_snapshot_service.go: loadAccountsFromDB/ListSchedulableAccounts。
- **行为与验收**：快照 SHALL 保留配置启用的完整账号池，并在每次读时判断临时冷却/expiry。 验收：快照构建时冷却→自然到期无outbox、真正禁用、simple/混排、取消和未就绪回源。。
- **边界**：不随行#8833的20s stale grace或合并回源；保留revision/retire/单节点缓存契约。

### S17 OpenAI 用量观测新鲜度（[#8747](https://github.com/Wei-Shaw/sub2api/pull/8747)）

- **依据/最小前置**：静态：当前读操作生成UpdatedAt；现有用量/监控消费者都有可复用接口。 源8行观测predicate及unknown字段需贯通已有UI；保留上一轮已实现的刷新留行。
- **patch sites**：backend/internal/service/account_usage_service.go、channel_monitor_quota_fetcher.go；frontend UsageProgressBar/AccountUsageCell/types。
- **行为与验收**：额度查询 SHALL 区分真实上游观测、陈旧读数和本地未知统计，刷新失败不得伪造新时间。 验收：主动失败/无quota/显式0、legacy无时间、被动缓存、unknown窗口、UI不误停账号。。
- **边界**：不引入新quota-summary或恢复监控v2整页；缺失Spark专用前置不作为依赖导入。

### S18 实际扣除与账号成本统计（[#8806](https://github.com/Wei-Shaw/sub2api/pull/8806)）

- **依据/最小前置**：静态核实本地条件覆盖actualCost；SQLite查询/字段和Gemini聚合均已有。 源PG ARRAY查询不能覆盖本地SQLite SQL；需同时改查询、consumer和redaction。
- **patch sites**：backend/internal/repository/usage_log_repo_stats.go、usage_log_repo_trend.go；gemini_quota.go、usage_handler.go；EndpointDistributionChart/types/user UsageView。
- **行为与验收**：按account过滤的模型/端点统计 SHALL 分列actual_cost和account_cost，Gemini账号配额使用账号成本。 验收：合成不同标准/实际/账号价、所有筛选/端点来源、管理员/用户redaction与Gemini批量一致。。
- **边界**：只增DTO字段，无新DB列；user API/UI不能泄露账号成本；保持model_source等维度。

### S19 价格启动不阻塞远端哈希（[#8798](https://github.com/Wei-Shaw/sub2api/pull/8798)）

- **依据/最小前置**：静态：Initialize同步fetchRemoteHash；本地正文实际hash修复已存在。 现有价格生命周期和远端client存在；需绑定worker所有权/截止/停止。
- **patch sites**：backend/internal/service/pricing_service.go: Initialize/checkAndUpdatePricing/Stop。
- **行为与验收**：已加载本地有效价格后 SHALL 不阻塞启动等待远端hash，并保持可停止的后台同步。 验收：慢/失败hash不挡启动，下载hash一致、并发同步、Stop/取消、fallback/override。。
- **边界**：不能只加无管理goroutine；不改变已有目录实际正文hash和零覆盖优先级。

### S20 上游 TCP 死连接检测（[#8800](https://github.com/Wei-Shaw/sub2api/pull/8800)）

- **依据/最小前置**：静态socket实现审查；未做丢包/死peer真实计时实验。 Go1.27.3支持该标准API；不依赖Go1.27.4或新x/sys版本。
- **patch sites**：backend/internal/repository/http_upstream.go: newUpstreamDialer；平台专用keepalive helper。
- **行为与验收**：既有标准上游dialer SHALL 使用15s idle/5s interval/3 probes，Linux user-timeout设置失败仍拨号。 验收：Linux实际socket选项、错误降级、非Linux编译、proxy/TLS配置和取消。。
- **边界**：30秒是来源参数推算，不是本次实测断线耗时；应用层120s idle保留。

### S21 OpenAI 工具 schema 正则适配（[#8772](https://github.com/Wei-Shaw/sub2api/pull/8772)）

- **依据/最小前置**：静态：本地只有null-type/required清理，没有来源完整regexp parser；未向远端发送新样例。 基础工具消费者在此；新正则能力不能凭CLEAN判编译，jsonschema-go0.4.2已在本地图中。
- **patch sites**：backend/internal/service/openai_responses_tool_schema.go；HTTP/透传/WS实际schema规范化点。
- **行为与验收**：现有OpenAI工具schema SHALL 窄处理来源不支持的regexp，并保留实例值与仍有效的schema约束。 验收：pattern/patternProperties、不支持转义/组、嵌套编辑、实例default/enum、有效Python形态及非OpenAI。。
- **边界**：不能把来源约899行parser整体覆盖本地窄sanitizer；需测实际适配规模和编辑重叠。

### S22 既有流与请求热点窄优化（[#8653](https://github.com/Wei-Shaw/sub2api/pull/8653)、[#8667](https://github.com/Wei-Shaw/sub2api/pull/8667)、[#8903](https://github.com/Wei-Shaw/sub2api/pull/8903)、[#8904](https://github.com/Wei-Shaw/sub2api/pull/8904)）

- **依据/最小前置**：Ollama200消息PoC语义对照通过，7.92ms→0.34ms、12.56MB→0.26MB分配；来源SSE escaped-usage predicate失败。 raw view已存在；字符串reverse/helper88行/flush-reader16行可窄适配，完整根文件不能覆盖。
- **patch sites**：backend/internal/service/gateway_tool_rewrite.go、gateway_upstream_response.go、gateway_claude_oauth_body.go、gateway_billing_header.go、openai_gateway_chat_completions_raw.go、openai_gateway_ollama_cloud_cc_reasoning.go。
- **行为与验收**：选定热点优化 SHALL 保持原字节、工具顺序、错误/用量语义和有界生命周期。 验收：每个采用hunk做单P成本/字节对照，escaped type/model/usage、read前flush/EOF/断开、cache TTL次序。。
- **边界**：补转义回退；不导入DeepSeek缺失placeholder消费者、无管理的懒日志全body保活、#8816账务lane或完整COW。

### S23 Go工具链与时钟测试维护（[#8767](https://github.com/Wei-Shaw/sub2api/pull/8767)、[#8743](https://github.com/Wei-Shaw/sub2api/pull/8743)）

- **依据/最小前置**：只读Go1.27.4 embed构建及config/apicompat/claude三包通过；首次下载网络超时另存。 范围内go.sum/前端依赖无变；jsonschema-go只是同版本direct标记，非新依赖前置。
- **patch sites**：backend/go.mod、本地CI/Docker/开发版本断言；两个现有*_record_usage_test.go。
- **行为与验收**：Go维护 SHALL 对齐1.27.4声明/CI/工具，并让已有峰谷价格测试使用固定PricingAt。 验收：全量unit/lint/race/embed、同版本lint重建、隔离Ent/Wire兼容及固定23:59边界。。
- **边界**：不升级库版本、上游PG依赖、fork VERSION或历史记录；codegen/CI尚不是本轮已验收项。


## 5. 按需项、排除项与旧遗留

### 5.1 本轮第三/第四档

全部理由在候选表。重点选择边界：

- GPT-6.2 Sol/Astra、Carbon/Hydrogen、Qwen新兜底、Fable quota恢复和Vertex多区域需实际模型/套餐/端点选择。Fable读数/冷却分支本地已有，但目录/价格未选；有自定义映射实际使用时重评，不说绝对不可达。
- 既有Sol Ultra Fast倍率修复可独立摘取；强制Ultra Fast和渠道独立倍率是新增策略/字段，仍按需。GPT-6.2促销时钟/套餐放行不借道进入。
- #8816新增544行PG batch repo和311行lane combiner；#8833追加stale grace/合并加载/last_used缓存。单用户高并发和大bucket收益需SQLite/miniredis、事务/dedup/取消/内存独立测量。S15成员正确性、S22窄热点不需要两套完整架构。
- #8678关闭ping是新控制；协议错误兼容不能靠关闭默认ping逃避。#8693的PG rollup/锁/迁移不在本地消费链。
- #8649支付退款整项排除。个人用量/账号额度/上游订阅元数据的维护仍保留；不进行无关支付代码删除。

### 5.2 旧轮次当前口径

以下依据提交/归档和当前任务清单，不宣称本轮重新执行了旧批全部验收。六份旧活动change共180个完成checkbox、零未完成，等待归档；本轮不改它们。新增两份提案的146任务全部未实施。

| 旧范围 | 2026-11-18当前状态及保留原因 |
|---|---|
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
| v0.1.180/v0.1.184计价/usage | 5250907f5模型参考价及Fast已完成；动态长上下文阶梯、独立effort倍率、usage compaction/requested effort持久化、峰谷/重置锚点仍按需或需schema。新TTL是另一个独立schema选择。 |
| v0.1.180工具/Grok/WS及更早 | PDF/tool身份/HTTP回放/孤儿工具/composite闸门及历史四尾项已完成；Grok429/Realtime/抢占按需，跨节点Redis租约单节点不需要；#5925最小delay随容量消费者完成但非整簇。0.1.176/179/183/184主体和Responses Lite已验收；128c5f7bf固定窗口与227/228语义保留。 |

上一轮#8606 30m/12h独立usage/价格/schema仍按需，与本轮5m/1h完整性修复不同。#8610动态预算最小1GiB不适合1GiB实例，保留已测64MiB；#8590新强查按钮、#8367保存浮标、Argon quota/新图片/cache CRUD及历史admin Key/完整COW/平台/抢占策略继续按实际用途。Fable相关修复本轮明确重述已存在的局部分支，未自动升档。WS默认1.0、固定quota和SQLite后台恢复的用户选择继续保留。

### 5.3 依赖与tag后main

上游范围只有Go1.27.3→1.27.4和jsonschema-go0.4.2由indirect改direct；go.sum/前端manifest和lock未变。本地也已含该库版本，不能以修schema测试为由升级整个依赖图。当前生产审计12 moderate/2 low/0 high/0 critical，退出1代表仍有advisory；[本轮实际记录](./evidence-0.2.23/dependency-summary.json)不替代Go和开发依赖全扫。

[tag后main](./evidence-0.2.23/post-release-commits.txt)仅2af3f1e15的VERSION同步，单独观察，不纳入产品范围。发布、推送和部署另行处理。

## 6. 已具备行为、N/A和前置结论

- #8651跨组RPM与#8740 Gemini刷新留行已在本地手工实现；#8825的强制DeepSeek峰谷策略本地不存在，当前0/正数/缓存读override控制通过。原补丁反向不通过不能重开这些问题。
- #8664账号namespace、#8735完整raw OAuth metadata清理、#8598在途预占测试基座缺失；原缺陷由其特定改写/架构产生，不能引入整套功能来修。
- #8837虽以Seedance为标题，本地chat/embeddings/Responses能力消费者真实存在，丢投影键可独立复现，明确列F18。
- #8772与本地null-type sanitizer不是同一基座；已有工具请求消费者但需新增窄regexp处理，明确列第二档，不能把来源约899行文件全覆盖。#8754同样要手工补raw局部重建调用点。
- F14两个Grok提取器、S07搜索predicate、S11工具boundary、S12格式helper都可独立适配，不导入原生新平台/模型。源码符号扫描见完整inventory，实际前置以逐簇依据为准。

## 7. 建议实施顺序

1. S03 SQLite原子配置、F09登录guard、S02token恢复及S01缓存/结算先处理共享数据和认证正确性。
2. F01–F06/F21与S04–S08优先Claude/Codex/Grok：工具内容、权限、上下文、首输出/idle预算、声明ID、Grok搜索、原文模型。
3. S09–S11/S21处理Composite/effort/Bedrock/schema边界；保留真实模型观察和已映射调度键，验证普通4xx/工具负对照。
4. F08/F10/F11/F14/F18及S16–S20处理既有倍率、quota/快照/成本/启动/传输；不改变固定窗口或带入新价卡。
5. Gemini/Antigravity兼容与相关UI按同文件依赖顺序落地；F16/F17/F19完成保存重开/失败/nonce验收。默认模型维护独立验证。
6. S22每个性能hunk先做语义和资源测量，补转义保护；S23工具链单独统一。两批组合门禁及实际SHA完成后再考虑归档/发布。

## 8. 已执行检查与实施门槛

| 检查 | 结果/实际限制 |
|---|---|
| 来源/四态 | 82候选覆盖128非merge，524文件全部先反后正；真实index未变。 |
| backend发现/定向 | 5611个匹配测试被列出，10包命令退出0；web在该正则无测试，另做embed专用探针。不是新产品全量实施验收。 |
| frontend基线 | 7文件89测试通过。 |
| 冻结产品Go overlay | 6包实际编译，26个顶层目标行为失败、20个顶层对照通过；含真实modernc SQLite/Ent和embed HTML handler。 |
| 临时前端目标测试 | 2文件9失败/11通过；是保存对象类型和星号输入断言，不是浏览器验收。 |
| 性能源码PoC | Ollama9组字节对照通过；200条assistant、GOMAXPROCS=1，7.924855ms/12560039B/2319alloc→0.341573ms/261461B/627alloc。仅函数夹具，不宣称1GiB实例总堆/生产收益。来源escaped-usage谓词目标断言失败，需人工保护。 |
| Go1.27.4可行性 | 当前产品只读embed构建、config/apicompat/claude三包通过；当前go.mod仍1.27.3，未实施新CI/lint/codegen。 |
| 生产依赖审计 | 12 moderate/2 low；无依赖变更。 |
| 文档/范围 | 见document-validation.json：两提案严格校验、任务未勾选/验收空槽、所有历史和产品哈希核对。 |

过程失败与最终结果分开保存：服务首轮count_tokens夹具遗漏生产DI的Config而panic，修正后才统计产品断言；直接embed handler夹具补WriteHeaderNow才观察到Gin缓冲的304状态；来源whitelist测试位置不同需移植测试；pnpm在临时链接目录尝试依赖核验被其安全guard拒绝，CLI负选项也未接受，最终直接用node运行已由pnpm安装的Vitest，无依赖安装/改写；Go官方下载首次网络超时，镜像下载后真实编译通过。上述不计入产品缺陷数。

静态项明确标注，尚无新token并发401、双账号idle/WS、Bedrock错误全集、Grok跨协议搜索、schema真实远端或实际浏览器/移动验收。本轮未触及真实业务实例。

实际移植后必须执行backend全量unit/embed/lint/相关race，真实HTTP/WS/SQLite/miniredis结算+成功去重、固定quota/旧时间格式/Fast/预算/释放检查；前端完整门禁和浏览器保存重开/快速切换/失败/移动端；S22测原文等价和边界，S23做Go/CI/lint/Ent/Wire统一兼容。评估结果不填入实施verification。

## 9. 本轮通用教训

- 发布摘要可能被同PR后续提交推翻；完整PR最终diff与tag优先，Gemini强制缓存策略是实证。
- 原补丁ALREADY=0也可能已经手工修过语义；用对照和实际predicate确认RPM/Gemini/DeepSeek。
- 缺完整改写基座有时意味着没有原缺陷；有小helper缺失但真实消费者存在时则需窄适配。
- 缓存presence、5m/1h明细、聚合差额和互斥输入桶需联合验证，不能只加JSON字段。
- SSE性能predicate的字节匹配会漏JSON转义字段；源码PoC通过语义和资源检查后才可采用。
- 首内容hold需沿用有界stage；keepalive提交HTTP后仍可安全换号的前提是错误保持SSE、账号头仍私有。
- SQLite配置同步必须原子覆盖删除和写入；成功路径通过不代表失败后旧数据可恢复。
- 临时前端的依赖工具可能主动安装；测试隔离也要确认runner，过程失败不能冒充产品断言。
- Go新patch能编译当前产品不等于完成新工具链/代码生成/CI验收；未来验收必须独立记录。
