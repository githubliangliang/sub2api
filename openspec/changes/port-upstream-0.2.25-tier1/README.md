# 上游 v0.2.25 第一档

状态：**已实施并验收，随本次提交保存，等待另行归档**。实施日期 2026-11-24；F01共1簇，6/6任务完成。真实起点为v1.1.27 / `39ca66d51d53b98f11349598bbd16cd8ecd0e2e0`。

来源与patch sites见[评估报告](../../../docs/upstream-sync/PORTING-0.2.25.md)；[评估期证据](../../../docs/upstream-sync/evidence-0.2.25/MANIFEST.md)保持原样。[本轮实施证据](../../../docs/upstream-sync/implementation-0.2.25/README.md)记录预修复失败、实际本地HTTP、全量unit/race/build/lint、SQLite保护与最终哈希，和第二档联合验收。

F01已消除既有错误文本的账号凭据与主机名泄漏，保留query净化、Unicode/非法UTF-8安全和非敏感恢复语义。VERSION1.1.27、Go1.27.4、SQLite230/auth21及资源默认保持；组合仅第二档定向维护Gorilla。第三档、支付和业务实例操作未引入。

[行为规格](specs/upstream-0-2-25-tier1-compatibility/spec.md) · [任务](tasks.md) · [实施验收](verification.md) · [冻结来源](source-baseline.md)
