# 实施验收证据

状态：第一档4簇开发及验收完成。

## 逐场景证据

| 簇 | 失败证据 | 改后场景与正常反例 |
|---|---|---|
| F01 | [red-tool-history.txt](evidence/red-tool-history.txt)：1000条工具历史旧实现分配137,056,945 bytes/op，超出4,505,100上限 | `TestApplyToolNameRewriteToBody_Spans`、`_LongHistoryAllocationBound`及原工具改名测试通过：不同长度/转义名、声明/强制选择/历史引用一致，input、内置和未映射名称保留，最后工具缓存规则保持；见[green-backend.txt](evidence/green-backend.txt)和后端全量 |
| F02 | [red-frontend.txt](evidence/red-frontend.txt)：alias→real被作为identity白名单发出 | `ModelWhitelistSelector.spec.ts` 9项通过：拒绝冲突并提示，允许不同名称、自映射、空目标和规范化空白；创建/编辑/批量三处接线完成 |
| F03 | [red-backend.txt](evidence/red-backend.txt)：两个兼容入口有fallback仍403 | `TestGatewayOpenAICompatibleHandlersClaudeCodeOnlyFallback` 8场景通过：真实HTTP上游SSE经Chat/Responses转换均返回200及正文，只选择fallback账号；无fallback为403，缺失/循环链无上游请求。见[green-price-fallback.txt](../port-upstream-0.2.11-tier2/evidence/green-price-fallback.txt)及race |
| F04 | [red-client-tabs.txt](evidence/red-client-tabs.txt)：受限组仍显示Codex | `UseKeyModal.spec.ts`的`shows only Claude Code for Claude Code-only groups`通过：切换限制时重置标签并去除Codex配置，切回OpenAI恢复默认选项；同文件21项与全量前端均通过 |

F01证据修正：原始上游`_Spans`测试的失败来自本地未引入的deferred-tool缓存策略，不属于本轮修复。原实现通过其余JSON正确性场景；可稳定复现的问题是重复整包改写导致的分配量。现采用一次span组装，保留原有最后工具缓存断点，不宣称已复现或修复deferred缓存问题。

## 执行环境与门禁

2026-09-30，在 `sync/upstream-0.2.11` 分支完成。实施起点 `f13dee245790eb2be058935d0b1aa6265da99474`；开始时只有前一轮评估文档待提交，产品文件干净。开发验收时为未提交工作树，验收产品及测试文件的 SHA-256 清单见 [implementation-files.tsv](../port-upstream-0.2.11-tier2/evidence/implementation-files.tsv)。开发验收时未发布或部署；随后按用户要求完成本地Docker及远端部署，见下方部署记录。

以下命令均使用 `rtk proxy` 执行；Go 命令工作目录为 backend，pnpm 命令工作目录为 frontend。OpenSpec 命令在仓库根目录运行。

| 门禁 | 命令 | 结果 / 证据 |
|---|---|---|
| 后端全量 | `go test -tags=unit ./... -count=1` | 通过，[backend-unit.txt](../port-upstream-0.2.11-tier2/evidence/backend-unit.txt)；service 169.360s，handler 37.181s |
| 后端构建 / DI | `go build ./...`；`go generate ./cmd/server` | 通过，[backend-build.txt](../port-upstream-0.2.11-tier2/evidence/backend-build.txt)；Wire仅增加resolver注入 |
| 并发 / HTTP / WS | `go test -race -tags=unit ./internal/service ./internal/handler -run 'TestAntigravityCompat\|TestAnthropicChatStreamAuthoritativeUsage\|TestHandleResponses.*Usage\|TestOpenAIResponsesWebSocket_Composite\|TestGatewayOpenAICompatibleHandlersClaudeCodeOnlyFallback\|AccountModelRoute' -count=1` | 通过，[backend-race.txt](../port-upstream-0.2.11-tier2/evidence/backend-race.txt) |
| SQLite / 去重 | `go test -tags=unit ./internal/repository -run 'TestProductionSQLUsesSQLiteDialect\|TestUsageBillingRepositoryApplySQLiteAppliesAllEffects' -v -count=1` | 通过，[sqlite.txt](../port-upstream-0.2.11-tier2/evidence/sqlite.txt)；真SQLite重复Apply不再扣费 |
| 前端全量 | `pnpm exec vitest run` | 252文件，1836通过 / 2既有跳过，[frontend-unit.txt](../port-upstream-0.2.11-tier2/evidence/frontend-unit.txt) |
| 前端类型 / lint / build | `pnpm run typecheck`；`pnpm run lint:check`；`pnpm run build` | 通过，见第二档evidence中同名日志；仅既有pnpm/browserslist构建提示 |
| OpenSpec | `pnpm dlx @fission-ai/openspec validate port-upstream-0.2.11-tier1 --strict`；同命令tier2 | 通过，见两档evidence/openspec.txt |
| 范围 / 格式 | `git diff --check`；文件清单与冻结证据哈希比对 | 通过，[scope-check.txt](../port-upstream-0.2.11-tier2/evidence/scope-check.txt) |

全量Go通过后，仅增加GPT目录契约测试并调整两条既有测试的Gin全局模式初始化顺序；对应定向与race再次通过。前端全量通过后仅清理测试的未用import和修正目录提示，最终类型/lint/build及相关组件回归112项通过（[frontend-final.txt](../port-upstream-0.2.11-tier2/evidence/frontend-final.txt)）。未将postgres-only集成测试或“不匹配任何测试”的输出算作行为验收。

## 边界与测试适配

保留SQLite-only、最大迁移226、可关闭Redis/miniredis、simple mode、VERSION 1.1.14；没有schema、依赖、支付、缺失原生平台、Astra Ultrafast或其它第三/四档功能变动。两份source-baseline及20份评估证据的原始哈希保持不变。

旧v0.2.9 F04/#7538只作为本次S02前置完成两桥thinking禁用优先级；旧第一档其它簇仍待处理。旧695ebede7仅移植14行通用usage字段。

新增上游测试中对缺失native Anthropic路径、未引入的allowlist、全套Fast策略的依赖已按本地实际消费者适配；这些编译差异不记作产品缺陷。race发现两条既有Responses测试在`t.Parallel`后写Gin全局模式，已将初始化移至并行边界之前；[原报告](../port-upstream-0.2.11-tier2/evidence/red-race-harness.txt)与最终通过日志均保留。

## 本地Docker部署（2026-09-30）

用户追加要求后，以当前工作树重建`sub2api-new:local`，通过既有SQLite Compose更新`sub2api-new-sqlite`。容器healthy，零重启；健康、首页、公开设置和最新前端资源均HTTP 200。沿用现有数据挂载及环境（此容器原为standard模式），数据库切换前快照与运行库均通过应用SQLite引擎校验。配置备份及旧镜像回退已保留。详见[部署证据](../port-upstream-0.2.11-tier2/evidence/docker-deployment.txt)。部署时尚未创建git提交、tag或Release。

## 远端部署（2026-09-30）

按用户要求执行`deploy/deploy-remote.sh --yes`，本地重新构建后上传至`liang@192.236.223.107:/home/liang/sub2api`，通过配置的自定义命令重启用户级systemd服务。脚本退出0；服务active/running、零重启，本地二进制、远端文件及运行进程的SHA-256完全相同。配置校验值未变，旧二进制保留；健康、首页、公开设置及最新前端资源均HTTP 200，外网健康检查通过。版本仍1.1.14，包含部署时的工作树代码，当时尚未创建git提交或GitHub Release。见[远端部署证据](../port-upstream-0.2.11-tier2/evidence/remote-deployment.txt)。
