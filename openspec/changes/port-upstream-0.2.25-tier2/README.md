# 上游 v0.2.25 第二档

状态：**已实施并验收，随本次提交保存，等待另行归档**。实施日期2026-11-24；S01/S02共2簇，11/11任务完成。真实起点为v1.1.27 / `39ca66d51d53b98f11349598bbd16cd8ecd0e2e0`。

来源与patch sites见[评估报告](../../../docs/upstream-sync/PORTING-0.2.25.md)，[评估期证据](../../../docs/upstream-sync/evidence-0.2.25/MANIFEST.md)保持原样。[本轮实施证据](../../../docs/upstream-sync/implementation-0.2.25/README.md)分别记录S01错误语义与S02模块维护，未用临时可行性冒充实现验收。

F01前置已同时实现。S01完成原状态/body规则、有限结构化信封和compact单次有序失败终态；S02精确更新Gorilla，MVS仅一项变化，实际管理员WS/CONNECT和网关race通过。库原始两项代理EOF已定因，原失败与真隧道夹具控制分别保留。VERSION1.1.27/Go1.27.4/SQLite230/auth21、其余依赖、前端、生成代码及资源默认保持；第三档/支付及业务实例操作未引入。

[行为规格](specs/upstream-0-2-25-tier2-compatibility/spec.md) · [任务](tasks.md) · [实施验收](verification.md) · [冻结来源](source-baseline.md)
