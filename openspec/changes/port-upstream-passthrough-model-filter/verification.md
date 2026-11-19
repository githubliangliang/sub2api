# 实施验收

2026-11-19 实施通过，6/6 项完成。基线 v1.1.26 / `9c073cb30`；[工作树文件校验](../../../docs/upstream-sync/implementation-passthrough-filter-and-usage-effort/product-files.sha256)。

| 规格 | 验收项目 | 结果 | 证据 |
|---|---|---|---|
| P1 | OAuth/API Key、精确键、通配键、目标值、关闭、空映射及非透传对照 | 通过，真实 HTTP 选中有资格的第二账号 | [组合测试](../../../backend/internal/handler/passthrough_effort_sqlite_test.go) |
| P2 | 候选重选、调度快照及透传正文 | legacy/load-batch、排除账号、miniredis 更新、正文通过 | [调度与转发](../../../backend/internal/service/openai_passthrough_model_filter_test.go) · [快照](../../../backend/internal/repository/scheduler_cache_unit_test.go) |
| P3 | 普通和 Codex 模型目录、混合账号、权威空结果、元数据、缓存及 ETag | 通过，规则变更返回新正文 | [目录回归](../../../backend/internal/service/openai_passthrough_model_filter_test.go) |
| P4 | 创建、保存重开、切换账号和批量编辑 | 组件测试、桌面和手机 Chromium 通过 | [浏览器](../../../docs/upstream-sync/implementation-passthrough-filter-and-usage-effort/browser.json) |
| 集成 | Go、前端、构建及范围 | 54 包 unit，相关 race，前端 2224 通过/2 跳过，lint 及构建通过 | [证据](../../../docs/upstream-sync/implementation-passthrough-filter-and-usage-effort/README.md) |
