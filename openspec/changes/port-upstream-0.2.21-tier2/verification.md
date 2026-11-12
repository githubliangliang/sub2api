# Verification（实施验收）

状态：**S01–S12实现完成，42/42任务关闭**，2026-11-12（Asia/Shanghai）。实现随本次提交保存，等待归档。所有验收使用本地合成夹具；真实网络为本机HTTP/WS，SQLite和miniredis实际执行，不代表供应商或真实Codex Desktop联机验收。实现证据与冻结intake目录独立。

| 项目 | 实际证据 |
|---|---|
| 实际实施起点SHA | `4f4c80c3ce48ce4afbc9344536b9cf44ed3fd898` 加已有第一档工作树；[start](../../../docs/upstream-sync/implementation-0.2.21-tier2/start.json)。过程中第一档独立提交 `63c764704761073e69ff25ba41d64eeab5793cd0`，28文件哈希一致，见tier1-commit-observation.json |
| 实现提交SHA / 工作树产品哈希 | 本批实现随本次提交保存；提交前验收的[product-hashes](../../../docs/upstream-sync/implementation-0.2.21-tier2/product-hashes.json) / [完整patch](../../../docs/upstream-sync/implementation-0.2.21-tier2/product.diff)，以已完成第一档提交为diff基准 |
| 后端全量unit/build/lint | backend-unit-exact、backend-build-owned-final、backend-lint-last：全量unit、embed构建及lint0通过；测试子进程去代理沿用既有captcha环境策略 |
| 相关race | backend-race-complete五包通过，backend-race-last-guard / backend-race-exact补最终平台限制/SQLite worker/并发启停；无race报告 |
| 真实HTTP/SQLite/miniredis | sqlite-http-protection、page-http、ws-model-final、audit-sqlite-acceptance、cooldown-committed-ws；包含Fast/排队usage结算/dedup、固定quota窗口、SQLite模型状态/缓存、真实refresh原子消费 |
| 前端与浏览器 | frontend-full-final：291文件2197通过/2既有跳过；critical179/2，typecheck/lint/build通过；browser-owned-final真实构建SPA+Chrome桌面/390px移动、慢用量/失败、账号DOM/配额保留、inert与更新，API均合成 |
| 1C1G/索引成本 | index-1c1g：实际modernc完整228→229，20000行结果一致、计划命中covering；868352字节新增、memory.peak=66850816，详见下方成本边界 |
| 保护范围与严格校验 | scope-check.json与openspec-tier2；冻结source/proposal/design/spec、旧迁移、版本/依赖/生成代码保护；第一档共享hunk保留 |

本目录日志均位于 [implementation-0.2.21-tier2](../../../docs/upstream-sync/implementation-0.2.21-tier2/README.md)，每条检查都有同名JSON记录argv、cwd、时间、耗时和退出码。过程失败保留并解释，不能将初次失败或评估探针当成验收通过。

| 簇 | 修复前复现 / 维护基线 | 实施后结果 | 边界 / 实际范围 |
|---|---|---|---|
| S01 原生WS后续轮次失败关闭 | baseline-ws、later-overload-before：后续轮次换号回放首帧或直接透传容量错误 | 实际handler+HTTP/WS：后续429/overload/slow_down及重拨握手失败1013关闭，第二账号无连接/无首帧；首轮429两模式仍换号 | 命名返回+当前轮nil payload包装，槽位/lease释放与Fast/usage旧验证保留；只关闭无法跨账号迁移的原生后续轮次 |
| S02 OAuthWS搜索激活与完整回放 | baseline-ws：无搜索socket继续接收search，未知历史未关闭 | empty prewarm/Lite/工具续接重连，reasoning/message/call/output和用户原文保留，previous_response_id移除 | API Key/同名function/已初始化socket不重连；未知历史/缺call失败关闭；map/raw/bytes与写失败capability检查，状态受write锁保护 |
| S03 WS与SSE首token口径一致 | 原硬编码delta判定；ws-ttft来源计时假设不适配本地，保留过程日志 | semantic/visible两模式、两轮、passthrough及ctx_pool与SSE分类一致；空delta/文本/工具/元数据/终态/取消边界通过 | 保留本地每轮起点，未导入FirstTurnStartedAt/TakeNextTurnStartedAt；HTTP桥共用分类 |
| S04 显式模型冷却覆盖容量重试 | baseline-service：HTTP及四reader×三凭据×两错误形态缺模型冷却 | ws-model-final覆盖四reader×三凭据×四形态含显式503；匹配停止同账号重试；真实HTTP/SQLite/调度cache和WS/已提交输出副作用通过 | canonical model只映射一次；未配/account/无模型/排除码/其它平台不触发；输出/metadata不命中；写失败不扩大账号冻结 |
| S05 Anthropic兼容入口账号错误策略 | baseline-service：410自定义/临时策略被默认failover状态挡住 | Chat/Responses、流/非流自定义410/临时规则生效并返回failover；pool可重试策略与shouldDisable一致 | 未配410/400原响应、默认500及正常路径对照通过；不扩大所有客户端4xx |
| S06 合成effort的映射与最终校验 | baseline-handler：两平台6子例失败 | disabled none一次映射/cap/deny、普通省略不变、显式非法拒绝、Sol/Terra none/max及两兼容出站均通过 | raw显式校验→转换→策略→最终校验；count_tokens不发送effort、保留原始显式验证；既有GLM兼容不变，无新平台/schema |
| S07 Refresh Token原子轮换 | baseline-service：barrier八请求全部成功 | service barrier+真实miniredis八请求仅1成功；直接缓存16并发仅1消费，重用/过期/禁用/version/家族/消费错误/取消失败关闭 | 预校验顺序保留，全部stub更新；Lua GET+DEL真实执行，失败无新token/无敏感token输出 |
| S08 Key独立用量与SQLite覆盖索引 | baseline-frontend慢用量阻塞loading；候选索引先单独测量 | 列表先渲染，用量异步；失败/取消/卸载/空列表/快速分页旧响应测试通过；实际228→229/重启/仓储数值和covering计划一致 | 只新增229(api_key_id,created_at,actual_cost)；旧SQL区间不变，无PG语法/旧迁移改写，成本见下 |
| S09 账号刷新保留行与已知配额 | baseline-frontend刷新行/配额被骨架替换；batch-account-reuse-before跨账号旧batch误显 | 初次骨架、已有DOM/额度保留，busy/inert；成功更新/失败旧值、分页取消/账号请求代次、批量托管同对象隔离、桌面移动通过 | 只保留同账号值；不依赖新Antigravity按钮/Ollama能力；未修改DataTable产品。完整table selection邻近测试参与全量 |
| S10 Codex Desktop页面上下文兼容 | 冻结来源静态缺helper；页面测试旧validator验证缺call_id，不能宣称真实Desktop复现 | page-http真实Responses handler→HTTP捕获首帧，open/selected顺序及页面原文保留为external标记user文本；历史/混合bootstrap通过 | 真实call_id、错namespace/name、重复键、坏page_id/分片头不转换，不提升system，不记页面内容 |
| S11 审计worker入队唤醒 | baseline-audit-cli旧worker通过；audit-sqlite-diagnosis实际raw到期0条、julianday1条 | publish后非阻塞wake、失败不通知、合并唤醒、并发Start/Shutdown/有界超时通过；实际SQLite约1ms认领、到期retry无Notify由轮询拾取 | 只增加小channel/生命周期补偿及claim混合格式比较；事务/claim_version不变，500ms轮询保留，无审计新规则或业务扫描 |
| S12 Claude Code内置版本基线 | baseline-audit-cli原版本2.1.258通过，非版本闸门线上故障复现 | 2.1.321；CLIVersion/User-Agent/缓存指纹/billing身份消费者一致，高版本/低版本/非法覆盖对照通过 | 只有常量和必要期望测试；不引入同PR模型/目录/定价/白名单/UI新能力 |

## 索引 / 资源成本

20000行合成记录，覆盖索引增加868352字节。候选同机7次查询中位7.398ms→3.128ms，5000条写入158.236ms→195.536ms。1GiB/1CPU容器正式迁移构建18.638ms，查询6.433ms→3.002ms，写入153.787ms→143.190ms，memory.peak=66850816字节。写入方向存在样本噪声，不能称为稳定写入提升；不是生产延迟验收或完整实例内存保证。普通gateway预算没有改变。

## 范围 / 遗留

用户随后授权本地commit，实现随本次提交保存；未push/发布/归档/部署或操作业务数据。source/proposal/design/delta-spec及冻结intake/历史不改，迁移229是本spec选定范围。VERSION1.1.24、Go1.27.3、auth snapshot21、依赖、Ent/Wire保持。其它第三档、支付、TTL/new model/platform/adaptive budget未纳入。S11基于实际SQLite缺陷补最小前置；此前后端unit/lint/source-test不适配和浏览器夹具失败详见实施README，未删日志。
