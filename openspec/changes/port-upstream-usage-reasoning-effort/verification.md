# 实施验收

2026-11-19 实施通过，6/6 项完成。基线 v1.1.26 / `9c073cb30`；[工作树文件校验](../../../docs/upstream-sync/implementation-passthrough-filter-and-usage-effort/product-files.sha256)。

| 规格 | 验收项目 | 结果 | 证据 |
|---|---|---|---|
| E1 | 显式值与后缀、缺省和未知、none/max、模型与分组规则 | 通过；缺省保持 NULL，max 不受实际模型归一化影响 | [请求值测试](../../../backend/internal/service/openai_reasoning_effort_policy_test.go) |
| E2 | HTTP 降级，WS 逐轮与换号，异步写入 | HTTP xhigh→high，WS 三轮及后续轮重建通过 | [组合 HTTP](../../../backend/internal/handler/passthrough_effort_sqlite_test.go) · [race](../../../docs/upstream-sync/implementation-passthrough-filter-and-usage-effort/backend-retry-race.log) |
| E3 | 229→230、新旧行、单条/批量/best-effort、查询及去重 | 通过；旧值、金额、固定 quota 与已有去重身份保留 | [数据库测试](../../../backend/internal/repository/usage_requested_effort_sqlite_test.go) · [结算](../../../docs/upstream-sync/implementation-passthrough-filter-and-usage-effort/backend-race.log) |
| E4 | 角色 DTO、页面和导出 | 实际数据库回读、两个 DTO、桌面及手机浏览器读取 XLSX/CSV 均通过 | [浏览器](../../../docs/upstream-sync/implementation-passthrough-filter-and-usage-effort/browser.json) · [DTO 回归](../../../backend/internal/handler/dto/mappers_usage_test.go) |
| 集成 | Go、前端、构建及范围 | 54 包 unit，相关 race，前端 2224 通过/2 跳过，lint 及构建通过 | [证据](../../../docs/upstream-sync/implementation-passthrough-filter-and-usage-effort/README.md) |
