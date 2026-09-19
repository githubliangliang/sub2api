# 设计与决策

1. 按行为手工移植，PR 整体仅用于理解来源终态，不覆盖整文件。本 fork 在这些文件里有自有逻辑
   （namespace 处理、Responses Lite、SQLite 用量写入），移植时保留。

2. **S02 的连接池是进程内的，单节点同样受影响。** 不要因为「WS 池听起来像多实例设施」就降级：
   `openai_ws_pool.go` 在本进程内维护每账号连接，`max_conns_factor` 默认实测为 `1.0`
   （`config.go:2426-2427`），上游已改为 `5.0`。**先修#7064使已配置系数生效；#7043默认调优单列后段。**
   默认仍为1.0也能正确执行容量公式，因此二者不必绑定。#7043仅在记录活跃会话、账号并发、硬上限及资源余量后决定采用5.0；
   没有依据时保留现有默认并标记调优未实施，不阻塞#7064验收。不得把池容量系数当作在飞请求并发系数。
   #6965 的常驻读循环用于应答上游 ping，是长连接存活的前提，与多实例协调无关。

3. **S03 反转本仓库既有决策。** 当前 `request_transformer.go:161` 注入
   `IncludeServerSideToolInvocations`、`antigravity_gateway_compat.go:317` 置 true，
   这是早先上游 #5709 的做法；#6689 改为「有客户端 function tools 时丢掉内置工具」。
   必须整 PR 移植，不能逐 hunk，否则两种策略混存。移植后 grep 还有谁在用旧写法。

4. **S12b #6954 是本批唯一带迁移的项，且含 PG 专用重写。**
   SQL 迁移在本仓库排 `227`（当前最大 226），不继承上游编号；
   `BatchSnapshotUsage` 的 PG 写法须按 SQLite 重写，禁止照抄。
   迁移文件一旦落地即受 checksum 保护，实施前确认列定义与 `EnsureSQLiteAuxTables` 逐列一致。
   未批准时保持 S12b 待实施，其余现行簇与 S12a 不受影响。

5. **S10 #7010 是对真实上游的行为变更**（TLS 指纹 Chrome→Firefox）。影响面已核实：
   `getSharedReqClient(Impersonate:true)` 仅供隐私客户端，`claude_oauth_service.go:268`
   与 `openai_oauth_service.go` 各自直接调 `.ImpersonateChrome()`，不受影响。
   落地前重新确认没有新的 `Impersonate:true` 消费者。

6. **S10 #6783 只取有基座的部分。** `httputil/body.go` 的分块读取可落；
   `openai_request_raw_input.go`、`openai_responses_ingress_compat.go` 在本仓库不存在，
   相关 service 层重构不整取。四个新 helper 逐个 grep 后再决定取哪些。
   本项对 1C1G 的内存占用有实际意义，但不能以损坏既有 reasoning-replay / call-id 归一为代价。

7. **S09 #7072 需要具名返回值改造**：`openai_account_scheduler.go:377-380` 目前是匿名返回，
   `defer` 写入的 `decision.LatencyMs` 到不了调用方。这是可观测性修复，不改调度决策本身。
   同 PR 的粘性命中重复计数（`StickyHitRatio = prevHit+sessionHit`）一并修。

8. **跨档共享文件须保留先落行为。** S11 #6916 与第一档T15 #7024同改订阅分配；
   来源 `SubscriptionsView.userUsageLink.spec.ts` 本地缺失，需要按本地fixture适配。
   S01 #7094 与已前移T19 #6964同改chat bridge，先落T19，后续角色规范化必须保留agent_message正文。
   固定顺序：先落第一档并验证，再落本批。

9. 前端 apply 干净不等于能编译。S11 剩余3个条目（#7111 同时涉及后端）以及 S04/S12 的前端改动都要跑 `pnpm run typecheck`
   与 `make test-frontend-critical`，不能只看 vitest 单文件。

10. 本轮不改 VERSION、不改 Ent schema、不动 Wire、不升级依赖、不引入新平台。
    S12b 之外无迁移。

## 现行归属与实施顺序

冻结 source-* 记录原28个PR；本次前移七项后剩21个PR/11簇，原S编号不重排。
S04=#7076、S05=#6977、#6995归S08；S12a=#6971，S12b=#6954。
现行契约以本设计/spec/tasks为准，不能再以冻结表要求第二档重复实现已迁出行为。

| 原归属 | PR | 现归属 |
|---|---|---|
| S11 | #7055 | 第一档 T17 |
| S11 | #7054 | 第一档 T18 |
| S07 | #6964 | 第一档 T19 |
| S11 | #7026 | 第一档 T20 |
| S11 | #7112 | 第一档 T21 |
| S11 | #7023 | 第一档 T22 |
| S11 | #7025 | 第一档 T23 |

S07退役；S11仅保留#6916有效用户搜索、#7111代理凭据清空、#6654注册确认密码。
先处理正文/协议/资源生命周期/配额正确性，再做可观察性和界面；注册确认密码与WS默认值调优排后段。
EasyPay仍在第一档后段；第三/第四档不迁入。两档合计仍为原43个PR。

不包含第一档已有的筛选分页，也不纳入#7064夹带的WS模式UI文案。
#6783不新增缺基座文件或全链路重构；#6977冻结名单不扩展成模型目录或计费更新。

验收以spec的WHEN/THEN为准：容量计算与默认值决策分别记录，count_tokens断点上限为4；
延迟用可控耗时验证，Firefox指纹与质询识别用受控响应，不宣称外部Cloudflare必然放行。

每簇可独立交付；S12b历史空配额行清理不能通过代码回退恢复，原有迁移批准边界不变。
第二档不是生产事故复现声明；实施时先以失败测试锁住触发条件。当前仅调整文档，产品任务全部未实施。
