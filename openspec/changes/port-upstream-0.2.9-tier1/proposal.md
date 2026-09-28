# Why

上游v0.2.5之后修正了本fork仍存在的网关/账号/界面问题。按第一档范围移植可改善个人单节点可靠性；
完整上游含PostgreSQL、新平台和支付能力，与本fork不一致。

## What Changes

本change包含15簇 / 39个PR，逐项见source-feature-map及spec。
只实现对应行为，采用v0.2.9终态，保留SQLite、miniredis、现有菜单和1.1.x版本体系。
不修改已应用迁移，不引入上游238b/239/240、支付/商业返利、插件宿主或第三/四档能力。

## Impact

本change已完成指定服务、转换器与界面修复；详细位置及评估取舍见[评估文档](../../../docs/upstream-sync/PORTING-0.2.9.md)。
Go1.27.1环境已补齐，54个后端测试包、build、目标race和SQLite检查通过；前端全量1909项通过（2项既有支付API测试跳过），typecheck/lint/build通过。
实际适配、审查修复及未提交状态见[验收记录](./verification.md)。
