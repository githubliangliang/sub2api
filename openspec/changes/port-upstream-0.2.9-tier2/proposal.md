# Why

上游v0.2.5之后修正了本fork仍存在的网关/账号/界面问题。按第二档范围移植可改善个人单节点可靠性；
完整上游含PostgreSQL、新平台和支付能力，与本fork不一致。

## What Changes

本change包含16簇 / 48个PR，逐项见source-feature-map及spec。
只实现对应行为，采用v0.2.9终态，保留SQLite、miniredis、现有菜单和1.1.x版本体系。
不修改已应用迁移，不引入上游238b/239/240、支付/商业返利、插件宿主或第三/四档能力。

## Impact

本次交付规格，尚未修改产品。实施涉及现有服务/转换器/界面；详细patch位置仅在[评估文档](../../../docs/upstream-sync/PORTING-0.2.9.md)。
前端现有基线通过；后端工具链缺失，需在实施前补齐unit/build/race与SQLite验证。
