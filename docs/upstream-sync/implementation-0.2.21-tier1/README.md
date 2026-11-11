# v0.2.21 第一档实施证据

2026-11-11，用户授权实现 F01–F11。实际起点 `4f4c80c3ce48ce4afbc9344536b9cf44ed3fd898`，VERSION1.1.24 / Go1.27.3 / SQLite228 / auth snapshot21。本批实现随本次提交保存；完成状态以 [verification](../../../openspec/changes/port-upstream-0.2.21-tier1/verification.md) 和 [scope-check](scope-check.json) 为准。

## 验收归属

最终检查期间，共享工作区出现另一个执行流程的 tier2 改动。首次共享 race 编译读到了 `normalizeCodexPageContext` 引用与文件写入之间的瞬态，记录在 `backend-race.log`；不能将此作为 tier1 产品缺陷，也不能将混合工作区检查冒充独立验收。最终全量 unit、embed build、lint、五包定向 race 和 HTTP/SQLite 检查使用 `/tmp/sub2api-021-tier1-acceptance`：起点 HEAD 加本批 28 个产品/测试文件，来源见 [isolated-snapshot.json](isolated-snapshot.json)。共享 Responses handler 仅提取本批 UTF-8 hunk；streaming 文件仅提取 Codex 元数据 hunk。没有回退或覆盖并行 tier2 内容。

[product.diff](product.diff) 包含全部已有及新增文件，可在起点 checkout 直接应用；[product-hashes.json](product-hashes.json) 对应被验收的本批文件。`check-scope.py` 同时验证隔离文件哈希、共享工作区可反向检查整个 tier1 patch、冻结历史及额外并行改动清单。隔离快照的 VERSION、依赖、schema、生成代码与前端都保持基线；共享工作区的其它增量不归入本次完成声明。

| 证据 | 含义 |
|---|---|
| [baseline.json](baseline.json)、[before-hashes.json](before-hashes.json) | 实际起点与已有文件 SHA256，包含尚未提交的 intake/提案 |
| [environment.json](environment.json) | Go1.27.3、lint2.14.0、VERSION1.1.24、SQLite228 |
| [before-regressions.log](before-regressions.log) | 四包实际编译执行，15 个顶层新行为失败、4 个顶层对照通过，覆盖11簇 |
| [before-images-http.log](before-images-http.log) | 两个图片真实 HTTP 顶层测试修复前失败，含 generation/edit 和代理期限 |
| [after-targeted.log](after-targeted.log) | 22 个顶层测试通过，含普通/透传 × 三种重建 × 四种密封标记、工具会话和图片HTTP |
| [hash-case-boundary-before.log](hash-case-boundary-before.log) | hash 大小写变换导致下载次数3而非2的独立修复前复现 |
| [boundaries-final.log](boundaries-final.log) | 13 个顶层测试通过，覆盖槽位/会话释放、价格、热重载、UTF-8实际HTTP与预算 |
| [isolated-unit.log](isolated-unit.log) | 独立 tier1 全量后端 unit；仅测试子进程去除代理变量，沿用现有 captcha 环境策略 |
| [isolated-build.log](isolated-build.log)、[isolated-lint.log](isolated-lint.log) | 独立 embed build 与 lint |
| [isolated-race.log](isolated-race.log) | 五包定向 race，包含定价热重载、转换、图片HTTP、SQLite结算与miniredis |
| [isolated-sqlite-http.log](isolated-sqlite-http.log) | 56 个顶层检查：迁移、方言、配额、固定窗口、Fast、真实HTTP排队结算/dedup及miniredis |
| [openspec-tier1-final.log](openspec-tier1-final.log)、[scope-check.json](scope-check.json) | OpenSpec严格验证、任务与链接、产品/冻结范围核验 |
| [run-check.py](run-check.py)、各同名 JSON | 实际 argv、cwd、UTC完成时间、耗时、退出码及原始日志 |
| original-README.md / original-tasks.md / original-verification.md | 实施前状态完整保存；source/proposal/design/spec不改 |

## 适配与失败记录

- PR8384 的上游 hot-reload 测试夹具与本地不同，未整体套用；新增独立 remote stub 覆盖正文滞后、落盘hash、后续重试、无效JSON、网络失败和hash缺失/大小写。旧有热重载并发测试仍参与 unit/race。
- `product-port.log` 保留初次 patch include/exclude 顺序错误造成的原子拒绝；随后按先 exclude 测试、再 include 产品文件应用。F05 保留本地 `sessionID` 参数与 `ReleaseAccountSession`；F07 和 F11 根据本地上下文手工适配。
- `boundaries.log` 的 Sonnet 覆盖价断言最初错误地把内部每-token价格当作每百万价格，并额外假定默认 Fast 会再次乘已有priority覆盖。查明 `GetModelPricingWithChannel` 的既有契约后只修正测试：覆盖0/非0和显式Fast倍率；产品计价规则未因此改变。`boundaries-final.log` 为修正后验收。
- `model_prices_0211_test.go` 中 Sonnet缓存读1000 token期望由0.0002改为0.0001，并新增远端缺失 fallback 路径。其它模型与费率断言保留。
- 前置Gemini helper仅bool判断；不导入完整thinking策略。F06额外将启动/轮询hash比较改为EqualFold，以通过明确要求的大小写边界。

## 范围与实际环境

真实网络验收为本机 HTTP server、真实 handler/转换/转发及反向代理；供应商内容和鉴权身份是合成夹具，不代表真实账号生成。SQLite使用完整迁移的临时库，miniredis为真实进程内服务；结算/dedup检查走实际 repository。图片原心跳间隔1秒，错误延迟1.2秒；3秒代理头期限保持真实400/成功JSON，1.3秒期限对1.6秒Codex请求产生504，普通客户端仍通过心跳完成。部署方需允许图片操作在首字节前完整等待；本次未更改任何业务代理。

本批没有前端、依赖、索引、迁移、资源预算或后台资源策略变化，前端/browser与1C1G预算压力专项不适用。保持64MiB body budget / 40MiB SSE line / 120s idle；UTF-8早退路径另有预算完整释放验证。无支付、新平台、新模型、TTL扩展、发布、推送或业务部署。
