# #8748 与 #6188 实施验收

实现基于 v1.1.26 的 `9c073cb30bff9dfd9364a059d4e0929b14d8b8ff`，SQLite 从 229 升到 230。已测工作树的 57 个产品及回归测试文件列在[校验清单](./product-files.sha256)。在仓库根目录执行 `sha256sum -c docs/upstream-sync/implementation-passthrough-filter-and-usage-effort/product-files.sha256` 可核对这次验收的内容。

| 检查 | 结果及复核入口 |
|---|---|
| 真实 HTTP 与 SQLite | [组合用例](../../../backend/internal/handler/passthrough_effort_sqlite_test.go) 经过实际网关 handler 和本地 HTTP 上游。两个候选的目标映射均与请求名不同，选中第二个账号；上游收到原模型及 high，SQLite 记录 xhigh/high，两个角色的 DTO 分别返回相应字段。 |
| SQLite 升级与写入 | [升级用例](../../../backend/internal/repository/usage_requested_effort_sqlite_test.go) 在 229 库写入旧行后运行迁移；旧请求值为 NULL，金额和实际强度保留。单条、批量、best-effort 三条路径写入、回读及重复插入通过。Fast 结算的实际余额、账号成本和 quota 去重见 [race](./backend-race.log)。 |
| 候选、快照、目录 | [模型筛选用例](../../../backend/internal/service/openai_passthrough_model_filter_test.go)覆盖普通/批量调度、排除账号、透传原文、混合目录、权威空结果与最终 ETag。[miniredis 快照](../../../backend/internal/repository/scheduler_cache_unit_test.go)覆盖开关往返。 |
| WebSocket | [真实客户端三轮](../../../backend/internal/service/openai_ws_v2_passthrough_lifecycle_test.go)依次发送 xhigh、low、无值，保留所有结果后检查请求与实际值，使用合成上游帧。[后续轮换号](../../../backend/internal/service/openai_ws_http_bridge_resume_test.go)验证重建请求携带原始强度。[专项 race](./backend-retry-race.log)包括组合 HTTP/SQLite 用例。 |
| 后端、前端 | [Go 全量 unit](./backend-unit.log) 54 包通过；四包相关 race 通过；golangci-lint 为 0 issues，CGO_ENABLED=0 的 embed 构建成功。[Vitest](./frontend-tests.log) 294 文件、2224 通过、2 项既有跳过；typecheck、lint:check、build 通过。 |
| 浏览器 | [脚本](./browser-check.mjs)使用 Chromium 真实页面和合成管理 API，覆盖桌面 1440×1000、移动 390×844 的创建/编辑/重开/关闭/批量修改，以及用户与管理员的用量显示。下载并解析实际 CSV/XLSX 文件，分别核对单值与前后两列。[运行结果](./browser.json) |

从仓库根目录运行浏览器脚本。先在 frontend 运行 `pnpm build`、`pnpm exec vite preview --host 127.0.0.1 --port 18927 --strictPort`，然后执行 `node docs/upstream-sync/implementation-passthrough-filter-and-usage-effort/browser-check.mjs`。可通过 `PLAYWRIGHT_MODULE` 指定已安装的 Playwright 模块入口、`CHROMIUM_PATH` 指定浏览器、`APP_BASE_URL` 指定本地预览地址。

实施前，legacy 和 load-batch 调度都在目标模型不匹配时选中了透传账号。实施中，全量回归发现原有 Composite 目录被额外补入默认模型，修正后继续保留旧测试。前端夹具补上 Pinia，并对应真实服务将 load_factor 清除值读回为空。工作树验收使用本地合成账号；未使用生产凭据或业务数据库。
