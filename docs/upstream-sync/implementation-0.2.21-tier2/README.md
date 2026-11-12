# v0.2.21 第二档实施证据

2026-11-11开始，2026-11-12（Asia/Shanghai）完成 S01–S12。实际起点 HEAD `4f4c80c3ce48ce4afbc9344536b9cf44ed3fd898`，当时已有第一档未提交代码；实施中另一个执行流程将第一档提交为 `63c764704761073e69ff25ba41d64eeab5793cd0`。用户随后授权本地 commit，本批实现随本次提交保存；未 push、发布、归档或业务部署。以工作树产品哈希和 [OpenSpec 验收](../../../openspec/changes/port-upstream-0.2.21-tier2/verification.md) 标识结果。

## 起点与保护

[start.json](start.json) 保存开始时 git status、172 份已有/冻结文件与6953 tracked文件哈希；[tier1-commit-observation.json](tier1-commit-observation.json) 核对第一档独立提交的28份文件全部等于该批最终产品哈希。两批共享 handler、Anthropic转换和SSE文件；本批保留第一档内容。第一档任务/验收和两个测试在其完成流程中更新，不回退这些独立变更；当前差异以其完成提交为基准。[product-hashes.json](product-hashes.json) / [product.diff](product.diff) 给出本批最终产品及测试，包含新增文件。[scope-check.json](scope-check.json) 核对冻结来源、历史文件、依赖、版本、迁移与生成代码。

source-baseline/proposal/design/source-feature-map/delta-spec保留原文，original-README/tasks/verification保存本批实施前状态。只有当前README、tasks、verification及顶层状态入口更新完成状态。VERSION1.1.24 / Go1.27.3 / auth snapshot21、依赖和Ent/Wire不变；新增SQLite迁移229。保留64MiB请求体预算、40MiB SSE行及120s读写idle默认值。

## 验收索引

| 证据 | 内容 |
|---|---|
| baseline-handler/service/frontend/ws、later-overload-before | 本地消费者修复前失败；包括8次并发刷新全部成功、四reader模型规则、兼容错误策略、effort绑定、页面刷新和原生后续轮次回放 |
| core-targeted / protocol-boundaries / ws-model-final | 分阶段本地消费者验收；源测试时钟和URL夹具适配过程原样保留 |
| cooldown-committed-ws / final-platform-guard | 已映射模型、三凭据×四reader×四错误形态、负边界及输出已提交/WS错误副作用 |
| page-http | 真实本地Responses handler→HTTP上游，首轮page输出按原顺序变user文本；严格签名/重复键/真实call_id负对照 |
| sqlite-http-protection | 本地HTTP→真实SQLite排队结算/dedup/Fast、固定配额窗口、迁移、miniredis与模型冷却缓存一致性；PG专项skip另列 |
| backend-unit-exact / backend-build-owned-final / backend-lint-last | 最终全量unit、embed构建、lint；仅测试子进程去代理以沿用既有captcha环境策略 |
| backend-race-complete / backend-race-last-guard / backend-race-exact | 五包定向race；补最终OpenAI平台限制、实际SQLite worker与并发启停 |
| frontend-full-final / frontend-typecheck-final / frontend-lint-final / frontend-build-final-owned / frontend-critical-acceptance | 2197通过/2既有跳过；critical179/2，typecheck/lint/build通过 |
| browser-owned-final / browser-check.mjs / browser/ | 真实构建SPA与Chrome，桌面/390px移动、用量延迟/失败、账号行DOM/额度保留、刷新禁交互、配额/列表失败；API全部合成拦截 |
| sqlite-index-candidate / index-1c1g | actual modernc、228→229、20000合成行、仓储结果一致、覆盖计划、索引大小/写入/查询成本；1GiB/1CPU隔离容器峰值 |
| audit-sqlite-diagnosis / audit-sqlite-acceptance | 混合时间格式导致原比较0条、julianday比较1条；实际SQLite唤醒约1ms及不通知的到期retry轮询 |
| environment / scope-check / openspec-tier2 | 版本/工具/迁移、保护范围、OpenSpec严格验证 |
| run-check.py / 各同名JSON | argv、cwd、UTC结束时间、耗时、真实退出码、日志；中间失败未删除 |

## 适配与过程失败

- S01来源只给defer补丁，本地入口原本匿名返回；增加命名error并保留本地轮次接口。追加真实HTTP/WS探针发现后续轮次overload仍直接下发，故将无下游输出的后续容量错误交给同一1013关闭保护；首轮429换号仍通过。
- S02沿用AllItems、回放builder和连接池所有权；capability只在成功write后记录、由write锁保护。预热后搜索保留reasoning/message/call/output；未知previous_response_id和孤儿工具历史失败关闭。没有复制新的计时基座。
- S03来源测试假定从客户端response.create开始计时，本地passthrough基线从response.created记录每轮起点。两轮测试保留本地起点，只比较semantic/visible事件选择；不引入FirstTurnStartedAt/TakeNextTurnStartedAt。普通池化WS和HTTP桥也使用相同分类。空delta、工具参数、Codex元数据和取消/终态有边界测试。
- S04容量规则只匹配error字段。首次metadata负对照用server_is_overloaded，规则关键词overloaded其实命中了真实error.code；改为slow_down后准确隔离metadata。现有模型写失败策略返回handled=true以停止本次同账号重试，但不写整账号冻结；测试按此既有契约验收。输出已提交时仍冷却模型，不重放已输出流。追加slow_down+短message且无type裸error负边界，原有瞬态分类即通过；short-capacity-before为通过对照，没有为此扩展产品failover规则。
- S06先验证原始显式值，再转换、应用一次none映射/上限、最后验证。最初服务测试使用http base_url而零配置禁止http，只修夹具为本地既有测试配置。count_tokens无effort出站字段，验证原始显式值，允许disabled思考的token估算；不引入新平台。
- S07实际miniredis Lua GET+DEL；8请求barrier保证全部先读到同一旧token，最终仅1成功。过期/重用/禁用/token version/家族吊销/消费失败/取消均验收，失败不返回新token。全部现有stub更新接口。
- S08新增纯SQLite229，无PG CONCURRENTLY/INCLUDE，无旧migration修改；原统计区间SQL不变。先测临时索引，再迁移并重启验证checksum。
- S09已有组件只在初次加载无数据时骨架，失败保留同账号额度；列表保留行DOM并标记aria-busy/inert。补账号切换请求代次、卸载取消、分页旧响应守卫。最后追加批量托管组件复用探针，确认账号ID切换而batch对象仍是旧值时会误显旧配额；修复为拒绝前账号的同一batch对象，单测29项及最终全量2197/2通过。真实浏览器初次行身份检查失败后，节点观察发现fullPage截图临时切换响应式布局（移除目标的父级已是移动卡片分支）；改为视口截图后无resize/移除。另一次点击被既有引导overlay挡住，夹具按AppLayout真实admin_guide key标记引导已完成。保留所有失败日志，不因此改DataTable产品代码。
- S11为避免Start/Shutdown在wg.Add之前完成，先注册完整worker run，并共享done完成信号。原内部wg手工测试改用真实阻塞scanner来证明Shutdown超时。实际SQLite验收另发现NOW()的RFC3339字符串与Go时间绑定格式不能直接排序比较；仅claim的到期/顺序改用julianday，事务与claim_version仲裁不变。测试按真实SQLite初始化先注册既有NOW兼容函数，不新增生产函数。
- S12仅2.1.321常量与必要CLI/身份期望调整。原测试硬编码2.1.258导致全量失败，改为CLIVersion()，保留合法高版本/非法与低版本回退、User-Agent及billing身份一致性。

## 测量边界

候选索引：20000行新增868352字节；同机7次查询中位数7.398ms→3.128ms，5000行写入158.236ms→195.536ms。1GiB/1CPU容器：迁移构建18.638ms，查询6.433ms→3.002ms，写入153.787ms→143.190ms，memory.peak=66850816字节。两次写入测量方向不同，不能声称稳定写入加速，也不能外推生产负载/完整实例内存。后者为静态仓储测试二进制、临时SQLite、无外网，非业务实例。请求体预算未改。

新增审计worker唤醒是非阻塞、每worker最多1个待处理信号，保留500ms到期轮询；实测只对应本地空队列和小任务，不沿用来源平均250ms的宣称。实际外部上游、鉴权、页面/配额均为合成夹具；没有真实Codex Desktop或供应商账号联机，也没有生产库迁移。支付、TTL扩展、新模型/平台及自适应预算保持独立。
