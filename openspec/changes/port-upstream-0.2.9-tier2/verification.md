# 实施验收证据

状态：16 簇实现与验收完成；2026-09-29 后续按用户要求完成本地 Docker 部署并整理提交，未打 tag。

- 执行日期：2026-09-28 至 2026-09-29。
- 实际实施起点：`50c34ad498fa333e20b4ee46dfe58c4512d88a89`，开始时 `git status --short` 为空。
- 冻结上游终态：`4c00df2e0183e2c70b7fa8ba45914205e36aad0c`；[来源基线](./source-baseline.md) 与 [48 PR 映射](./source-feature-map.md) 未改写。
- 验收版本：实施验收时按设计未自动提交。全部 158 个产品/测试文件由 [product-files.sha256](./evidence/product-files.sha256) 精确标识；清单自身 SHA256：`d355bd9c2cd81dde170184f093923d2e4dd603138d1bd6371eeb3c5d8784d630`。在仓库根目录可运行 `rtk proxy sha256sum -c openspec/changes/port-upstream-0.2.9-tier2/evidence/product-files.sha256`；后续整理提交前复核 158 个文件全部匹配。

## 运行基线与最终门禁

Go 命令的工作目录为 `backend/`，实际使用自动工具链 `go1.27.0 linux/amd64`；前端命令在 `frontend/`。执行通过 RTK 包装，未修改依赖清单。

实施前已运行已有 `go test -tags=unit ./...`、`go build ./...`、相关 package 的 `go test -tags=unit -list .` 及真实 SQLite 基线，均退出 0。部分初始 unit 命中缓存；下表最终 unit/race 均显式 `-count=1`。

| 门禁 | 实际结果 | 持久证据 |
|---|---|---|
| 后端完整 unit | `go test -tags=unit ./... -count=1`，退出 0；service 169.745s，repository 34.401s | [backend-unit.txt](./evidence/backend-unit.txt) |
| 后端完整 build | `go build ./...`，退出 0 | [backend-build.txt](./evidence/backend-build.txt) |
| SQLite 方言与真实库 | 最终 unit 包含 `TestProductionSQLUsesSQLiteDialect`、真实迁移库 OAuth 刷新/代理回退及既有 billing 去重测试；无 postgres-only 用例冒充 SQLite 证据 | [逐场景核对](./completion-audit.md)、[unit](./evidence/backend-unit.txt) |
| 并发/取消/race | service、repository、handler、apicompat 的相关场景 `-race -count=1`，退出 0；包含真实 HTTP/TLS/HTTP2/压缩体、WS 多轮和 SQLite 回退 | [race.txt](./evidence/race.txt)，内含完整选择表达式 |
| 测试清单 | `go test -tags=unit -list . ./internal/repository ./internal/service ./internal/handler ./internal/pkg/...` 退出 0，逐场景文档的 60 个测试引用均匹配真实输出 | [test-inventory.txt](./evidence/test-inventory.txt) |
| 前端 Vitest | 全量 250 files，1808 passed / 2 skipped，退出 0 | [frontend-vitest.txt](./evidence/frontend-vitest.txt) |
| 前端 typecheck | `pnpm run typecheck`，退出 0 | [frontend-typecheck.txt](./evidence/frontend-typecheck.txt) |
| 浏览器 | 实际 Chromium + Vite 组件验证嵌套滚动锁与 IME/Enter；见下文 | [逐场景核对](./completion-audit.md) |
| OpenSpec | `npx --yes @fission-ai/openspec validate port-upstream-0.2.9-tier2 --strict`，退出 0 | [openspec.txt](./evidence/openspec.txt) |
| 版本/迁移/依赖/平台范围 | 无 VERSION、已应用迁移、依赖清单或菜单修改；验收时暂存区为空；158 个产品/测试文件与清单一致 | [scope-check.txt](./evidence/scope-check.txt) |

## 后续本地部署

2026-09-29 使用当前代码重新构建 `sub2api-new:local`，通过 `deploy/docker-compose.sqlite.yml` 启动 `sub2api-new-sqlite`。容器健康检查通过，`http://127.0.0.1:8080/health` 返回 HTTP 200 和 `{"status":"ok"}`，首页及公开设置接口均返回 HTTP 200。沿用现有 SQLite 数据挂载，部署前已完成数据库及配置备份；本地配置与备份不纳入提交。

前端两项 skipped 是 `SettingsView.spec.ts` 中原有的支付 provider API 用例（API 已移除），不在本次范围。目标场景无未通过或未运行项。没有使用真实外部模型账号进行收费调用；协议与生命周期通过本地 HTTP/WS 实测和 unit 验证。

## 分簇结果

所有行的实施版本均为上述清单标识的工作树。改前执行失败的测试名与原始日志哈希见 [regressions.txt](./evidence/regressions.txt)；每个 spec Scenario 的具体用例见 [completion-audit.md](./completion-audit.md)。改后的后端行为均通过最终完整 unit，前端行为通过最终完整 Vitest。

| ID | 改前实际缺口 | 改后与正常反例 |
|---|---|---|
| S01 | GPT-6 采样/价格缺失、Opus adaptive 与 dotted 别名不识别、Grok 4.7 不可用 | 目录/能力/价格别名接通；显式映射、旧模型和非数字 GPT 行为保留 |
| S02 | root union、required:null、tuple/const 处理不正确 | required 交并集和同名 allOf 约束保留；原生/兼容入口清洗，业务实例数据不改 |
| S03 | done 参数为空、inline input 丢失、PDF 被丢弃 | delta 优先于 seed；PDF 全链保真，空数据和 file_id-only 不下载 |
| S04 | terminal 等 EOF、提前 Close 阻塞、心跳误算输出、错误协议/状态/取消不正确 | terminal 完整发送后结束；单 attempt 取消；真实错误归因、499、countTokens 估算、用量 drain 与不重放；bare error 等后续终态，EOF 只补一次 failed，显式透传规则优先 |
| S05 | 裸 Gemini 直接发上游、signature/stop 误判内容 | thinking 变体映射与显式配置优先；特定 SDK 免注释心跳；有真实内容不重试 |
| S06 | 非高级调度漏响应归属、401 误禁账号、暂停 OAuth 漏刷新 | 正确归属/决策与释放槽位；Lite 分组读；模型冷却与认证禁用分开；真实 SQLite NULL/冷却边界通过 |
| S07 | 账号成本不遵守账号 gate，空图片价变成免费 | 成本 gate 与售价独立，优先级保留；nil 继承图价、显式 0 免费，区间不污染共享价卡 |
| S08 | 原生 URL 缺 /v1、尾斜杠/usage 重复 /v1、Windows catalog 路径错误 | 原生与 CC Switch 分别处理；root、/v1、子路径、不同平台和 Windows HTTP/WS 均验证 |
| S09 | 回退恢复 probe 陈旧、旧扫描覆盖新配置、禁用目标被选中 | SQLite 条件写入核对源快照和目标状态；只在实际代理变化时删除 probe |
| S10 | 模型不可用被写成端点不支持 | 400/404 模型错误不改未知/true/false；真正 404/405 保持不支持，普通 400 不误判 |
| S11 | 上游另一个别名泄漏给客户端 | Responses/Chat 流与非流返回公开名，真实上游模型仍记录，工具内 model 不改 |
| S12 | reminder 绕过关键词，尾随 system 遮蔽用户输入 | 关键词读完整用户文本；语义审核旧过滤不扩大，assistant/tool 回合不重审 |
| S13 | 前导 attribution/SDK 身份导致 Antigravity 不兼容 | 仅 Antigravity 清理前导身份；正文普通提及及原生 Anthropic attribution 保留 |
| S14 | 嵌套弹窗提前解锁、组合输入提前提交 | 最后一个弹窗关闭才解锁；IME、空 Tab、Delete/Backspace 与普通 Enter 验证 |
| S15 | 换 window 后仍沿用旧响应 ID，失败窗口提前生效 | 新窗口断链，同窗口/无标记继续；真实三轮 WS 验证失败不更新 last window |
| S16 | GPT-5.5 仍带 Lite 标记，metadata 非 ASCII 不适合头部 | HTTP、WS→HTTP、直接 WS 的最终请求处理；保留原请求与工具/历史；中文及代理对合法转义 |

## 来源取舍与本地适配

- 按 [design.md](./design.md) 的消费者与行为移植，未整仓覆盖、未引入上游新平台、支付、插件宿主或迁移。
- S01 只新增四个目标模型价卡；保留本地 Grok 函数与长上下文策略，未带入新版 effort/Fast 计价子系统。
- S02 扩展本地 offset sanitizer，未搬上游 parser 基座；补上原生 Anthropic 入口和同名 allOf 属性的约束交集。
- S04 保留本地 Ops 结构和失败归因，未添加上游代理日志字段。补齐 bare-error/terminal 依赖行为；认证、限流等可立即换号的错误及显式错误透传规则保持优先。早期 race 曾与后续源码编辑重叠而编译失败，未计 PASS；最终稳定源码复跑通过。
- S05 使用本 fork 已有 Gemini 3.6 默认目录；显式配置的 3.8 变体仍可用，未顺带扩默认目录。
- S06 保留传输不兼容时抢槽前拒绝的本地优化，已取得槽位均释放；SQLite 刷新候选增加 NULL 安全条件。
- S07 复用 `calculateCostWithServiceTierPolicy`；未搬入新版峰谷/effort resolver。更新了与新 nil 图片价语义冲突的旧断言，共享价卡不受污染的断言保留。
- S08 采用 #7622 修订后的 CC Switch URL 规则。冻结基线 usage 脚本与当前脚本实际求值分别产生重复与单个 /v1，证据收录在 regressions。
- S09 用 SQLite `IS`/`IS NOT` 与 `json_remove`，额外验证扫描后回退目标被停用的交错，写入时再次核对。
- S14 只补必需的 F10 前置（共享 dialog ID counter、空 Tab 导航）；没有扩为整批 Tier 1。
- S15 只在成功 completed/done 后更新 last window；S16 覆盖本地多种 WS 最终出站消费者，未引入缺失的 account_identity 子系统。

## 浏览器实测

通过 agent-browser 在本地 Vite 加载实际 BaseDialog/ModelTagInput 组件。两个 dialog 的标题 ID 不同；body overflow 从 hidden，在关闭子弹窗后仍为 hidden，最后关闭父弹窗后恢复 visible。

输入“中文模型”后，组合 Enter 的 `defaultPrevented=false` 且不产生标签；普通 Enter 后仅生成一个“中文模型”标签。浏览器与本次 Vite 服务均已停止。
