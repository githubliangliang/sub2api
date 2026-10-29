# 行为与反例验收

所有检查使用当前产品入口、本地 HTTP、临时 SQLite 或 miniredis。评估证据只作为冻结来源，不计实施通过。原始 argv、时间、退出码由 run-check.py 保存。

| 簇 | 验收与反例 | 修复前 | 修复后 |
|---|---|---|---|
| S01 | 原生 JSON/SSE、Chat/Messages 缓冲转换，终态优先、固定 200 rune 前序状态、非 Grok 保持；真实转发→handler worker queue→SQLite；冲突回退 0.009、无冲突对照 0.003 | [四路径基线 overlay](./s01-before-complete-paths.txt) 四项均在行为断言失败 | [四路径异步落库](./grok-async-sqlite-complete-fixtures.txt)、最终 unit/race 的 bounded prior 与原有邻近测试 |
| S02 | 手动/后台同一严格解析，自定义大小写、非法/有效空目录、失败保留成功、乱序/非法时间/来源隔离、旧格式；SQLite 语义 CAS、并发 extra、事务内关联读取/回滚/无缓存、软删/错平台/取消、旧表单/缺键/注入/观察重叠、JSON null 根 | [服务基线](./s02-before-service.txt)、[真实 SQLite 旧表单](./s02-before-sqlite.txt)、[null 根探针](./grok-null-extra-before.txt) | [服务矩阵](./grok-async-sqlite.txt)、[SQLite 矩阵](./grok-catalog-sqlite-fixed-fixtures.txt)、[编辑重叠](./grok-async-sqlite-valid-fixtures.txt)、最终 unit/race；前两个过程输出的最终通过项不替代完整门禁 |
| S03 | OpenAI image/image[n]/mask 刚好 20 MiB 与 +1，多个 part、文件名/MIME/字节一致、拒绝返回 nil 请求；Grok 完整解析既有有界 body；真实 handler 总 body 413 和取消不转发，下载/50 MiB 边界未改 | [上传前](./s03-s04-before.txt) | 最终 unit/race 中 TestOpenAIImagesUploadPartBoundaryAndIntegrity、TestGrokMediaUploadParsingRetainsFullBoundedBody、TestImageHandlersPreserveTotalBodyAndCancellationBoundaries |
| S04 | 两维、旧客户端部分/缺键、false/空/null、未知维度与其它 features 整块替换、输入不变；实际 UI 禁用平台、关闭搜索、全局隐藏仍保留；实际支持平台仅 Anthropic/OpenAI，提示文案不再误写 Anthropic-only | [服务前](./s03-s04-before.txt)、[UI 前](./s04-frontend-before.txt) | 最终 unit、[四个前端消费套件](./s05-active-modal-after.txt)、[真实保存状态](./browser-channel-save-state.txt)、[隐藏后状态](./browser-channel-hidden-state.txt)、[最终文案](./browser-final-hint-text.txt) |
| S05 | all-mapped/partially-unmapped、去重、详情通配/归一化/404、custom display 不变成授权、路径/强制/复合/混合路由；前后端普通测试 3.5 Flash、Google One/显式旧模型、已有 Antigravity 合并下拉；未加未选定模型 | [backend 前](./s05-backend-before.txt)、[frontend 前](./s05-frontend-before.txt)、[实际管理入口前](./s05-active-modal-before.txt) | 最终 unit/race、[管理入口后](./s05-active-modal-after.txt)、[浏览器普通模型](./browser-ordinary-options.txt)、[旧模型选择](./browser-legacy-explicit-model.txt)、[合并目录](./browser-combined-models-proof.txt) |
| S06 | 0.62.0 最小 MVS，独立 HTML 倍增分配比 3.960→1.902，已有网络/HTML消费者、生成工具兼容，GO-2026-6860 复扫与其它 finding 区分 | [HTML 前](./s06-html-before.txt) | [HTML 后](./s06-html-after.txt)、[依赖报告](./dependency-report.md)、[Ent/Wire 隔离生成](./generation-compatibility.json)、完整 build/unit/lint 与 network race |

## 夹具与门禁问题归因

S01 初次完整链夹具漏了实际 Grok OAuth provider、数据库凭据和 API Key 上下文；随后把 token 到期设成恰好一小时，又落在既有一小时刷新窗口内。最终用两小时有效 token 和完整认证/缓存上下文，不修改认证生产代码。计价对照按既有“不抬价”准入，采用响应价低于基线的夹具；首次把价设置反向的结果保留。

S02 初次 SQLite 夹具没有 outbox/foreign_keys 前置；HTTP 夹具的 DoWithTLS 误继承空 recorder。补齐夹具后重测，不把这些失败标成生产缺陷。后台目录请求与 quota 请求可同时完成；既有 quota 网络夹具仅对 quota 路径统计，不再让观测请求改变 3/2 次断言。

宽范围 race 暴露既有 grok_free_quota_gate_test.go 直接替换全局 sync.Map，而上一项后台 goroutine 仍读取同一变量。只改六处测试重置为 Clear；[十轮 race](./baseline-race-fixture-repaired.txt)通过，生产 free quota gate 未修改。

Gemini 公共目录退役后，三项旧默认列表断言及 Grok name-only fixture 按新契约调整。两份测试弹窗各自有消费者/历史测试，实际 AccountsView 使用 components/admin/account，四个消费套件均覆盖。浏览器首次等待旧默认模型名超时的过程输出保留，未当作完成证据。

最终 unit 仅在测试子进程移除六个代理变量；既有 Aliyun SDK NO_PROXY host:port 问题的受控归因见第一档证据，captcha 生产/测试代码未修改。

## S01 口径

[用户确认](./s01-acceptance-decision.md)：保持迁移 227；冲突位保留到 usage 写入边界并验证 SQLite 计价回退。数据库 mismatch 仍比较实际发送模型与终态响应模型；没有新冲突列。原始规格快照保留。

最终完整 unit 的 Codex 测试时钟差 1ms 属于既有独立 wall-clock 样本；仅测试对两侧时间进行范围验证并从稳定字段比较中去除该动态值，[30 轮](./codex-clock-fixture-regression.txt)通过，生产指纹文件未修改。
