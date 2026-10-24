# 实施验收

> 后续提交及发布准备（2026-10-24）：两档实现已共同提交为`2b79ac320`；VERSION已另升1.1.19用于用户要求的发布。以下为实施验收时VERSION1.1.18的历史记录，原始命令、哈希与冻结source-baseline不回写。发布检查见[发布说明](../../../../docs/releases/v1.1.19.md)。

状态：**S01–S10已实现并验收完成**（2026-10-24）。评估期[基线证据](../../../../docs/upstream-sync/evidence-0.2.18/baseline-tests.md)与本次实现验收分开。

| 项目 | 证据与结果 |
|---|---|
| 实施起点 | HEAD `4984327590ed29eb4d842bfae8aa3406ef497008`，含同一会话已完成但未提交的第一档；[start.json](../../../../docs/upstream-sync/implementation-0.2.18-tier2/start.json)冻结52个既有改动文件的哈希。 |
| 实现提交 | 当前工作树，尚未提交；未推送、发布或部署。锁文件与依赖声明作为本次实现一并交付。 |
| VERSION / 迁移 | VERSION 1.1.18；最大迁移226；冻结source-baseline、评估证据与SQL/schema保持不变，见范围审查。 |
| S01 混合 OAuth 凭据隔离 | [缺陷复现](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s01-before.txt) → [四包通过](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s01-s09-after.txt)。覆盖HTTP/WS歧义、冷却归属、明确路由、API Key多协议及发送前凭据隔离。 |
| S02 Responses 时间与失败终态 | [最终定向](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s02-final.txt)、[转换器全包](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s02-converter-time-after.txt)。JSON/SSE/WS、SSE转JSON共享时间，透传不改帧，合成失败含时间及序号。 |
| S03 Codex 退休模型一致性 | [最终定向](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s03-final.txt)、[独立审查补强](../../../../docs/upstream-sync/implementation-0.2.18-tier2/audit-catalog-compat.txt)。包含冷却、previous_response_id、双向及compact映射、最新DB复检、remote/file/空目录和ETag。 |
| S04 Codex 身份与容量恢复 | [缺陷复现](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s04-before.txt) → [三包通过](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s04-final.txt)。身份成对归一；空语义不阻止恢复；提示解析含非法/溢出、最多8秒与取消，保留既有指数退避。 |
| S05 Anthropic兼容流 | [契约红测](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s05-contract-before.txt) → [契约通过](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s05-contract-after.txt)；[完整定向](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s05-empty-eof-after.txt)、[race](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s05-empty-eof-race.txt)。涵盖有界排水、真实语义前后错误、部分失败结算、空/ping EOF唯一失败终态。 |
| S06 既有 Antigravity 路径 | [缺陷复现](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s06-before.txt) → [四包通过](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s06-final.txt)。媒体顺序、函数签名真实转发回放、既有路由/档位/价格和映射协议，未引入新目录。 |
| S07 Gemini错误策略/Vertex认证 | [组合通过](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s07-agent-combined-final.txt)、[race](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s07-agent-race-final.txt)。三入口首次/再次401、每次响应单次裁决、容量/配额反例、SA准确识别及取消、耗尽提示和敏感查询值脱敏。 |
| S08 Gemini中断结算 | [取消红测](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s08-before.txt)、[边界红测](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s08-contract-before.txt) → [三协议通过](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s08-after.txt)。[真实队列/SQLite](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s08-agent-sqlite-priced.txt)、[SQLite race](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s08-agent-sqlite-race.txt)覆盖每协议有/无usage，取消后扣费45e-6且重放去重。 |
| S09 图片下载边界 | [缺陷复现](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s09-before.txt) → [四包通过](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s01-s09-after.txt)、[额外非单播边界](../../../../docs/upstream-sync/implementation-0.2.18-tier2/audit-after.txt)。拒绝私网/组播/广播及非图片，保留公网/b64/限额/关闭语义。 |
| S10 工具链/依赖 | Go1.27.3与CI/Docker断言对齐；follow-redirects双override及lockfile均1.16.1；[依赖报告](../../../../docs/upstream-sync/implementation-0.2.18-tier2/dependency-review.md)记录必要的编译兼容前置与新旧扫描，遗留告警未清零。 |
| 后端unit/build | build：[最终handler修复后通过](../../../../docs/upstream-sync/implementation-0.2.18-tier2/backend-build-accepted.txt)；[完整unit](../../../../docs/upstream-sync/implementation-0.2.18-tier2/backend-unit-complete.txt)退出0，54包通过，耗时202.889s（包含缓存命中的包）。 |
| 默认无标签编译 | [修复后通过](../../../../docs/upstream-sync/implementation-0.2.18-tier2/backend-default-compile-complete.txt)；仅编译，不冒充默认测试全量执行。 |
| 受影响链路race | [service/handler/admin/routes/repository](../../../../docs/upstream-sync/implementation-0.2.18-tier2/backend-race-final.txt)、[三个协议转换包全量race](../../../../docs/upstream-sync/implementation-0.2.18-tier2/protocol-race-final.txt)、[最后队列修复后race](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s08-queue-race.txt)均通过。前一命令的apicompat/antigravity筛选无测试，已由第二条全包补足。 |
| SQLite/固定quota窗口 | [真实SQLite回归通过](../../../../docs/upstream-sync/implementation-0.2.18-tier2/sqlite-final.txt)。S05实际handler验证删除Key/缺失Key/取消/失败部分结算及余额、账号配额、dedup；S08三协议实际队列亦已通过。SQLite验证对象为结算账本，usage日志接口使用既有桩，不声称验证Ent日志表插入。 |
| 前端 | [lint](../../../../docs/upstream-sync/implementation-0.2.18-tier2/frontend-lint.txt)、[typecheck](../../../../docs/upstream-sync/implementation-0.2.18-tier2/frontend-typecheck.txt)、[完整Vitest](../../../../docs/upstream-sync/implementation-0.2.18-tier2/frontend-tests.txt)、[build](../../../../docs/upstream-sync/implementation-0.2.18-tier2/frontend-build.txt)通过；281文件，2027通过/2既有跳过。 |
| 冻结/范围/第一档保留 | [最终范围](../../../../docs/upstream-sync/implementation-0.2.18-tier2/scope-facts-final.json)、[135文件代码哈希](../../../../docs/upstream-sync/implementation-0.2.18-tier2/code-sha256.json)、[一致性](../../../../docs/upstream-sync/implementation-0.2.18-tier2/consistency.txt)已记录；[第一档回归](../../../../docs/upstream-sync/implementation-0.2.18-tier2/audit-tier1-preservation.txt)通过。 |
| OpenSpec严格校验/差异格式 | [严格校验通过](../../../../docs/upstream-sync/implementation-0.2.18-tier2/openspec-strict-complete.txt)，最终文档已重验，差异格式通过。 |

所有命令的同名JSON保留完整参数、开始时间/耗时及退出码。只有已退出且exit0的结果计为通过，历史失败日志不删除。[最终门禁汇总](../../../../docs/upstream-sync/implementation-0.2.18-tier2/checks.json)关联命令与退出码。

- [x] 每簇证据完整并满足spec。
- [x] 适用检查完成，未执行项说明原因。

## 本地适配与边界

- S01 使用本地五平台配置归属，冷却不改变永久协议归属；独立Gemini/Antigravity转发仍先分流。未把API Key多协议兼容扩大为OAuth兼容。
- S02 沿用本地裸错误状态机，补齐原生与转换器两条消费者；转换器缺少时间字段不是可忽略的来源文件差异。保留已有时间及上游usage/工具恢复，不为透传改写事件。
- S03 本地没有来源的全局compact fallback，因此只接既有compact映射和previous_response_id调度。不新增pinned目录或全局替换用户模型；退休结论使用配置候选，避免短暂冷却重现默认模型。
- S04 只有容量错误读取上游提示。有限超大浮点数在转Duration前封顶；非法值仍用既有指数退避。没有引入新截止时间、Grok配额或Realtime。
- S05 按本地spec收紧来源行为：可见输出后的带内错误保留已观察用量；生命周期前导帧暂存，传输注释不等同于语义输出；排水兼有空闲和总时限，停止时关闭body。仅空流/仅ping EOF也必须失败，不能合成成功。两协议耗尽后产生唯一、可解析的失败终态。保留本地native Messages原有403策略。取消后的队列使用已有mandatory能力并在入队前快照请求归属。
- S06 仅使用已有模型/路由与档位。来源Argon、3.8低档及展示新功能没有消费者，测试改为本地既有变体；工具媒体不下载URL，套餐细分ID只用于上游端点选择。显式映射、显式价格优先，不把通配符变成新能力。
- S07 使用现有service_account；补本地缺失的无reset短冷却基座后，只把无配额明细的共享容量从冷却中排除。details缺失/null/[]可继续判定共享容量，非空数组与非null的非数组details均保守拒绝该分类。非认证错误维持本地按请求模型的冷却；401仍账号级首次暂停/再次error。Ops使用现有字段；token端点生产固定可信Google地址，私有注入点只用于本机测试。三协议耗尽保留安全认证提示，保存错误与日志均屏蔽敏感URL查询值。
- S08 `context.Canceled`本身不证明客户端离开；必须检查原请求取消或真实下游写失败。读取错误先于残行解析，写失败后停止后续写入/flush且不读取未来usage；只返回当前已完整观察帧的token/图片。是否收到finishReason单独保留并记日志，ClientDisconnect写入结果。所有断开情况（含收到finishReason）都只用实际inline/file图片数，不按模型名估算；正常完成的旧fallback保留。无usage/无图片仍返回错误，真实上游错误也不伪造成客户端成功。未引入新思考输出。
- S09 新增最小导出包装复用已有IP策略，并仅在图片下载路径额外拒绝非单播；没有顺带改变监控策略。代理分支只能核验本地DNS，不能保证远端代理DNS，限制与来源一致。未开启批量图片能力。
- S10 先升级目标Go版本后，完整unit实测发现Ent加载器不支持Go1.27.3的V5导出格式；不是缓存问题。量测版本矩阵后，x/tools v0.50.0是首个支持版本，只增加其MVS要求的传递依赖；没有无关tidy或完整上游依赖升级。前端lockfile只含follow-redirects九行变化，最初连带Babel快照已恢复。完整差异与遗留告警见依赖报告。

## 失败记录与验证限制

- 最早的移植编译失败、未使用import、夹具漏配rate-limit依赖、单测服务器关闭挂住均作为测试适配/诊断记录，不充当生产缺陷红测。S07早期SIGQUIT日志不是通过记录，最后带`-final`的组合及race才是验收。
- `s05-exhausted-comment-negative-control`使用overlay禁用新增接线，验证回归能发现缺陷；它不是完整工作树起点。原始缺陷复现与额外契约红测分别保留。
- 默认无标签编译发现四个新unit夹具缺tag及上一批价格测试引用unit专属helper，已修正测试构造；未因此改变产品或冻结历史报告。
- S08独立审查发现Messages/native仍可能在drop队列饱和时丢任务，且异步读取Gin字段导致请求复用串元数据；[真实红测](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s08-queue-before.txt)准确复现两个入口各两处失败，Chat作为已修控制组通过。已与Chat对齐强制提交及字段快照，[修复后全例通过](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s08-queue-after.txt)，[race通过](../../../../docs/upstream-sync/implementation-0.2.18-tier2/s08-queue-race.txt)，涵盖三协议12个子例。补强前的unit/race不能替代修复后的验收。
- S08队列/SQLite首次夹具缺少Gemini动态价格，三协议均已入队但费用为0；补独立临时价格文件并断言确切45e-6后通过，不修改生产价格目录。
- 没有真实Google/Vertex凭据、上游Codex权限或风控探测；本地桩验证协议/调度/错误规则，不声称证明远端账号能力。没有执行`integration && postgres`测试、发布或部署。
- 完整unit中两处旧夹具稳定性缺口已实证修正：[心跳旧夹具重复红测](../../../../docs/upstream-sync/implementation-0.2.18-tier2/audit-keepalive-before.txt)，修后等待实际心跳；分配计数移到独立子进程，保留原2次采样与+40阈值。[8轮定向](../../../../docs/upstream-sync/implementation-0.2.18-tier2/audit-fixture-after.txt)、[3轮race](../../../../docs/upstream-sync/implementation-0.2.18-tier2/audit-fixture-race.txt)、[默认编译](../../../../docs/upstream-sync/implementation-0.2.18-tier2/audit-fixture-default-compile.txt)通过，[线性退化负对照](../../../../docs/upstream-sync/implementation-0.2.18-tier2/audit-alloc-linear-negative-control.txt)仍以large2016/small10准确失败。没有降低性能门槛或修改这两项生产行为。
- 依赖维护只完成本轮目标。剩余生产high为xlsx两条；Go扫描是包级命中，不等于符号可达或可利用证明。
