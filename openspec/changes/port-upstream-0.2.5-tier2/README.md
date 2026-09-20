# 上游 0.2.5 第二档

状态：**现行簇已实现并完成门禁**。S02 #7043 默认值调优与 S12b #6954 迁移按批准边界延期。
现行范围：原28个来源 PR 中7个前移第一档后，剩 **21个 PR / 11簇**。
保留 S01–S12 既有编号，S07迁至第一档T19后退役；S11仅保留#6654/#6916/#7111。
两档合计仍为原43个 PR，没有新增来源或行为。
第一档（[port-upstream-0.2.5-tier1](../port-upstream-0.2.5-tier1/)）先落地，本批在其之上实施。

- [proposal.md](./proposal.md)：问题与交付边界。
- [source-baseline.md](./source-baseline.md)、[source-feature-map.md](./source-feature-map.md)：冻结来源，实施期间不改写。
- [design.md](./design.md)：现行来源归属、移植约束与逐簇决策（优先于冻结历史编号表）。
- [specs/single-node-stability/spec.md](./specs/single-node-stability/spec.md)：行为契约。
- [tasks.md](./tasks.md)、[verification.md](./verification.md)：任务清单与实际验收证据。
- 本地未找到来源文档提及的 `docs/upstream-sync/PORTING-0.2.5.md`；本次按冻结 PR/SHA 与
  [候选表](../../../docs/upstream-sync/evidence-0.2.5/candidates.tsv)、
  [文件四态](../../../docs/upstream-sync/evidence-0.2.5/files.tsv) 核对，不另建 patch site 清单。

## 三个必须先读的边界

1. **S12b（#6954 平台配额清理）带迁移，未获用户批准前不实施。** 它需要本仓库新增
   `227`，且来源含 PG-only 的 `BatchSnapshotUsage` 改写。其余现行簇与 S12a 不需要任何迁移，
   VERSION 保持 1.1.12、最大迁移保持 226。
2. **S03（#6689）会反转本仓库既有决策。** 本仓库当前按旧 #5709 注入
   `IncludeServerSideToolInvocations`，本簇改为「有客户端工具时丢弃内置工具」。必须整 PR 取 diff，
   不能逐 hunk 合。
3. **S10（#7010）是对 chatgpt.com 的真实出站行为变更**（TLS 指纹 Chrome→Firefox）。
   影响面已核实仅限隐私客户端池；`claude_oauth_service.go` 与 `openai_oauth_service.go`
   各自直接调用 `.ImpersonateChrome()`，不受影响。

第二档不是「生产事故已复现」声明；每簇先用失败测试重现触发条件，再提交修复证据。

现行归属以 design.md 的前移清单及本批 spec 为准；source-* 保留原评估记录。
S12a 为 #6971 TTFT，S12b 为 #6954 平台限额；S12b未获批时不得宣称现行21项全部完成。
WS容量修复先落，默认系数调优另列后段；注册确认密码仍保留在既有范围末段，不引入第三/第四档。
