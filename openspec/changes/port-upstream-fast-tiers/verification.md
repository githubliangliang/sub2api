# Verification

基线和来源保持 source-baseline.md 原值。实际实现保留在工作区。

| 项目 | 结果与证据 |
|---|---|
| Fast/Ultrafast 入口 | 原生 Responses、自动透传、Chat、Messages、两种 Chat fallback 的流式/非流式 24 场景检查实际出站字段；HTTP/WS 策略、非法枚举、missing 和分组优先顺序通过 |
| 真实 HTTP/SQLite 结算 | 18 个 protocol×tier 场景实际 queued settlement、SQLite扣减及重复请求去重通过；Free Fast 上游成本保持、用户成本减为普通档 |
| 价格/能力 | Astra6x、其他档位倍率、账号目录空值与 OAuth 边界、显式零价及既有272k边界通过；原生 Claude speed 已进入 token 计价 |
| 持久化/缓存 | 真实227→228升级、重复迁移不重做、Ent存取、APIKey投影、缓存roundtrip/失效、平台切换及管理员DTO隔离通过 |
| 后端 | 54包unit、embed构建、lint0、四包定向race通过；[最终unit](../../../docs/upstream-sync/implementation-fast-tiers/backend-unit-complete.log)、[race](../../../docs/upstream-sync/implementation-fast-tiers/backend-race-final-20261110.log)、[lint](../../../docs/upstream-sync/implementation-fast-tiers/backend-lint-complete.log) |
| 前端 | 2179 passed / 2既有skipped，critical176/2，typecheck/build/lint通过；[全量](../../../docs/upstream-sync/implementation-fast-tiers/frontend-unit-final.log)、[最终构建](../../../docs/upstream-sync/implementation-fast-tiers/frontend-build-final.log) |
| 浏览器 | 实际临时服务器、SPA和SQLite：分组开关/倍率保存重开清除、渠道2.5保存回显、390px布局、Ultrafast与missing策略保存重开、非法倍率API400；[原始命令](../../../docs/upstream-sync/implementation-fast-tiers/browser/commands.json) |
| OpenSpec/范围 | strict通过；旧迁移、版本、依赖、Wire和历史证据保持；[最终范围核对](../../../docs/upstream-sync/implementation-fast-tiers/scope-check.json) |

完整使用说明、过程失败和证据边界见[实施记录](../../../docs/upstream-sync/implementation-fast-tiers/README.md)。上游采用合成响应验证协议，不声称测量真实商业账号权限、延迟或实时官方价格。浏览器初始化的是合成管理员gate，不代表用户接受协议。没有业务实例部署。
