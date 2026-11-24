# v1.1.28 发布准备（2026-11-24）

用户明确要求创建v1.1.28标签、详细描述并推送。真实起点为已推送的 `9fe9ef3d6329a109923277a27c54fccbaa229f87`；上一版v1.1.27的annotated tag对象为 `9f17304a5f4090a91b0e692ff14a543c5b7920c4`，peeled commit为 `39ca66d51d53b98f11349598bbd16cd8ecd0e2e0`。准备前本地/远端v1.1.28标签均不存在，官方仓库API的该Release返回404。新发布提交号、tag对象与推送结果由实际操作产生，不预写成功。

## 版本、范围和文件保护

- [详细中文正文](../v1.1.28.md)既作为仓库发布说明，也作为annotated tag完整注释。既有Release工作流读取tag正文，不重写发布配置。
- v1.1.27到产品基线只有一个实现提交，实际新增F01/S01/S02三簇、17任务；不重复计入上一版已发布的缓存integration断言修正。F01/S01来自选定#9045行为，S02仅取#9008依赖子项。
- [initial.json](initial.json)保存真实HEAD、上一tag对象/commit、8,895个tracked文件哈希及原先未跟踪Fast诊断哈希。[source-verification.json](source-verification.json)确认只有VERSION和AGENTS索引是有意的既有文件改动；所有其它既有源码、测试、前端、迁移、Ent/Wire、历史证据及活动OpenSpec保持SHA256一致。
- VERSION更新为1.1.28；Go1.27.4、SQLite230/auth21、Gorilla最终精确版本和其它依赖，以及64MiB请求体内存预算/40MiB SSE行/8MiB暂存/120s idle保持。发布准备不引入产品修复、schema变更或额外能力。
- 前端构建生成的dist和检查二进制不纳入提交；原Fast诊断不纳入提交，内容保持。docs/releases按现有仓库策略被忽略，发布说明及本目录仅显式纳入本次提交，不顺带变更.gitignore。

## 复用的实现验收

[v0.2.25实现证据](../../upstream-sync/implementation-0.2.25/README.md)的15份最终后端文件哈希全部匹配当前产品。复用55包unit（10,793顶层/22,386含子例通过，16既有跳过）、三包错误/SQLite race332/669、五包WS race693/1,351、最终定向71/214、SQLite计费/固定窗口/去重69/136、真正本地HTTP/compact/管理员WS/CONNECT/miniredis、lint0等结果。不同定向范围有重叠，不累计为唯一测试总数。

原依赖库两项代理EOF仍保留，wire trace、当前Go源码和仅改为真隧道的临时夹具已确认非Hijack CONNECT关闭根因；不把真实应用控制通过写成依赖库原始测试全量通过，也不声称本轮执行漏洞利用或全仓漏洞扫描。详细局限和失败定位见[S02证据](../../upstream-sync/implementation-0.2.25/s02/README.md)。

## 本次重新执行

| 检查 | 实际结果与证据 |
|---|---|
| Docker/Go/Redis准备 | [checks/initial.json](checks/initial.json)核对Go1.27.4、Docker29.6.2和已存在redis:8.4-alpine镜像；实际测试容器执行，不触及业务实例。 |
| 完整integration | [checks/full-integration.json](checks/full-integration.json)：CI=true、TZ=UTC、Go1.27.4、去除代理变量，`go test -json -tags=integration -count=1 -p 2 ./...`通过；49测试包、6,180顶层/12,504含子例通过，13顶层/14含子例跳过，约239.7秒。完整JSON流[无损gzip](checks/full-integration.log.gz)，哈希及计数可复核。 |
| 标签preflight隔离控制 | [checks/release-tag-preflight.json](checks/release-tag-preflight.json)通过；它是既有脚本测试，创建本版实际tag后另核对实际解析。 |
| Release发布配置控制 | [checks/release-publish-config.json](checks/release-publish-config.json)通过，保持既有下载重试与工作流语义。 |
| 工作流/shell | [checks/actionlint.json](checks/actionlint.json)、[checks/shell-syntax.json](checks/shell-syntax.json)通过。 |
| 前端重建 | [frontend-build.json](frontend-build.json)、[原始输出](frontend-build.log)：`pnpm run build`通过，含vue-tsc与Vite；已有chunk大小提示保留，不修改前端配置。 |
| 本版embed构建 | [embed-build.json](embed-build.json)：使用上述新构建SPA，`CGO_ENABLED=0 go build -tags=embed -trimpath -o /tmp/sub2api-v1.1.28-version-check ./cmd/server`通过，不注入main.Version。 |
| 实际版本 | [version-resolver.json](version-resolver.json)输出1.1.28；[binary-version.json](binary-version.json)/[log](binary-version.log)实际输出Sub2API1.1.28，commit/date为unknown，明确为本地检查二进制，不冒充远端发布产物。 |
| 范围/正文/链接/空白 | [local-checks.json](local-checks.json)确认源码哈希、17任务说明、三簇范围、旧tag/历史保护、说明链接和精确版本。 |

前端原始stdout带有构建器行尾空白；暂存检查只对产品/说明/脚本检查空白，原始日志按原字节保存，不整理成不同输出。若出现其它检查失败，记录原始原因后再处理，不把wrapper成功当作测试成功。

本次integration按现有标签约束执行；56份另需postgres标签的文件未执行，不宣称PostgreSQL验收。构建标签清单、三项真实Redis调度集成控制和4,089个检查输入哈希见[checks/README.md](checks/README.md)及其原始记录。

发布起点SHA的远端CI在首次查询时仍在运行，shell/frontend/lint成功；同SHA的Security Scan为failure。见[远端初始运行](remote-before-runs.json)及[CI jobs](remote-baseline-ci-jobs.json)。该远端扫描状态与本轮本地完整integration分开，不修改扫描规则或声称远端全绿。发布后按真实发布SHA和tag运行检查Release，不把尚未完成的工作流提前写成成功。

## 提交、标签与推送

1. 提交VERSION1.1.28、详细中文说明、当前AGENTS索引及本次发布检查证据。
2. 以 `git tag -a v1.1.28 -F docs/releases/v1.1.28.md` 创建annotated tag，核对tag对象、peeled commit、提交内VERSION与完整正文，并实际执行tag preflight和版本解析。
3. 推送main及新v1.1.28标签，不覆盖已有标签、不强推main。
4. 核对远端main、tag对象和peeled commit，跟踪该tag实际Release工作流；实际对象号/状态由本次执行回报。

不归档OpenSpec，不回写冻结来源/旧验收状态，不执行业务实例部署。
