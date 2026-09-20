# 上游 0.2.5 第一档

状态：**产品实现已完成，验收证据已回填，工作树未提交**。T01–T23 的代码与门禁已落地；T11/T16 的真实登录态浏览器验收仍为 PARTIAL。
现行范围：22 个来源 PR，拆为 T01–T23；#7052 分为后端 T07 与前端 T08。
原 T01–T16 编号不变，原第二档七项前移为 T17–T23；两档合计仍为原43个 PR，没有新增来源或行为。
无迁移、无 Ent/Wire、无依赖升级；VERSION 保持 1.1.12，最大迁移保持 226。

## 阅读顺序

- [proposal.md](./proposal.md)：问题、范围、交付标准。
- [implementation-map.md](./implementation-map.md)：**实施编号与 patch site 的唯一现行清单**，含冻结来源勘误。
- [design.md](./design.md)：移植决策、兼容边界、风险与回退。
- [specs/single-node-correctness/spec.md](./specs/single-node-correctness/spec.md)：行为契约。
- [tasks.md](./tasks.md)：分阶段实施任务与实际完成状态。
- [verification.md](./verification.md)：文档核对结果、逐项验收矩阵和待执行门禁。
- [source-baseline.md](./source-baseline.md)、[source-feature-map.md](./source-feature-map.md)：原始评估快照，保持冻结，不作为修订后的实施编号表。

原文引用的 `docs/upstream-sync/PORTING-0.2.5.md` 在当前工作树不存在；
本 change 已自包含所有 patch site，不依赖该缺失文档，也不代为创建跨档总报告。

## 实施前必须知道

1. 冻结映射将 #7052 合并为一个 T07，之后编号比实施清单少一位；原任务统一使用 T01–T16；新增实施编号 T17–T23 仅承接原第二档行为。
2. T05 不移植 `account_repo_integration_test.go`：本 fork 为 `integration && postgres`，改用普通 unit。
3. T07 修复 JWT/管理员用户查询。API Key 的查询错误已返回 500，`User == nil` 的 401 是独立分支；保留并回归验证，不能当作同源缺陷直接修改。
4. T15 产品补丁可应用，但来源测试文件本地不存在；需按本地组件/fixture 补测。
5. 原15项的只读整 PR 复核为14 CLEAN、1 CONFLICT；前移七项均为CLEAN，现行22项为21 CLEAN、1 CONFLICT。
   CLEAN 不代表测试能编译或行为已经修复。优先级按 tasks.md 阶段执行，不按编号大小执行。
6. EasyPay 等既有按使用前提生效的项放在后段；本次不把第三/第四档加入实施。

第一档不是“生产事故已复现”声明。实现阶段用针对性回归证明触发条件，按 verification.md 回填真实执行结果。
