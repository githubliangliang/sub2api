# S10 依赖与工具链验收

实施查询及复核：2026-10-24（Asia/Shanghai）。本报告记录当前工作树的实际结果；冻结评估、source-baseline 和旧轮次验收不改写。命令记录器 `/tmp/sub2api-0218-tier2/run-check.py` 已检查存在，实际运行命令、退出码及时间写入同名 `.json`，原始输出写入 `.txt`。

## 选定变更与必要兼容前置

| 项目 | 实施前 | 当前 | 说明 |
|---|---|---|---|
| Go | 1.27.0 | 1.27.3 | go.mod、5 处 CI 版本断言、根目录及 deploy Dockerfile 一致 |
| follow-redirects | 1.16.0 | 1.16.1 | package.json / pnpm-workspace.yaml 两份 override 及 lockfile 一致 |
| x/tools | 0.47.0 | 0.50.0 | Go 1.27.3 导出数据格式的最小已发布兼容版本 |

原评估只选定 Go 和 follow-redirects。实施中全量 unit 暴露 Ent schema loader 不能读取 Go 1.27.3 的 export-data V5；[隔离 GOCACHE 复现](./audit-schema-clean-cache.txt)排除了旧缓存因素。当前 Go 1.27.3 的 pkgbits 新增 V5 / PreserveMethodOrder，而旧 x/tools 只支持 V4。对 v0.48.0、v0.49.0、v0.50.0 的已发布源码逐一量测：前两者仍不支持 V5，v0.50.0 首次支持；发布列表在 v0.49.0 与 v0.50.0 之间没有其它 patch release。矩阵、源码 Origin SHA 和模块要求见 [audit-tools-matrix](./audit-tools-matrix.txt)。这是明确量测后增加的必要前置。

[最小 go get](./audit-tools-update.txt)引入下列 8 个 root-pinned 模块变化，没有运行 tidy 或混入其它包升级：

| 模块 | 实施前 | 当前 | 必要依赖边 |
|---|---|---|---|
| golang.org/x/tools | v0.47.0 | v0.50.0 | Ent loader 的 go/packages / importer |
| golang.org/x/mod | v0.37.0 | v0.41.0 | x/tools 要求 |
| golang.org/x/net | v0.56.0 | v0.59.0 | x/tools 要求 |
| golang.org/x/sync | v0.21.0 | v0.23.0 | x/tools 要求 |
| golang.org/x/sys | v0.46.0 | v0.48.0 | x/tools / x/net 要求 |
| golang.org/x/crypto | v0.53.0 | v0.57.0 | x/net 要求 |
| golang.org/x/term | v0.44.0 | v0.46.0 | x/net 要求 |
| golang.org/x/text | v0.39.0 | v0.42.0 | x/net 要求 |

完整模块图另有 x/telemetry 节点从 `v0.0.0-20260625142307-59b4966ccb57` 变为 x/tools 要求的 `v0.0.0-20260908163034-4bcc4b2ee518`，没有新增 root require。因此完整模块图共 9 项版本变化；[before](./audit-tools-before.txt) / [after](./audit-tools-after.txt)可复核。go.mod 共 9 行替换（含 Go header），go.sum 仅新增这 8 个模块的 16 行 checksum，没有删除原 checksum。原失败的 [Ent schema 测试](./audit-schema-tools50.txt)及 [server build](./audit-backend-build-tools50.txt)在 Go 1.27.3 下通过。

前端 lockfile 仅 5 行新增 / 4 行删除，包含 follow-redirects override、版本及 integrity、axios 依赖边和 snapshot 名称。逐字构造预期变更与当前文件一致，package.json、workspace overrides 也只有这一新增 override；Go 配置/依赖差异同样做了精确断言。证据与最终文件 SHA256 见 [dependency-facts.json](./dependency-facts.json)和[核对命令](./audit-dependency-final-facts.txt)。没有复制 PG Docker 阶段、修改现有支付 SDK 或引入新支付依赖。

## pnpm 审计

使用 pnpm 11.21.0。pnpm 11 忽略 package.json.pnpm.overrides 的提示已保留在原始输出；实际生效的 workspace override 与旧版声明一致。

| 扫描 | 记录数 | Critical | High | Moderate | Low | 退出码 |
|---|---:|---:|---:|---:|---:|---:|
| [实施前全树](./s10-audit-before.txt) | 77 | 3 | 35 | 35 | 4 | 1 |
| [当前全树](./audit-frontend-final.txt) | 75 | 3 | 33 | 35 | 4 | 1 |
| [当前生产](./audit-frontend-prod-final.txt) | 17 | 0 | 2 | 13 | 2 | 1 |

仅两个目标 advisory 消失，没有新增 advisory：`GHSA-q3jh-wrc6-jgq8` 和 `GHSA-8r9p-f939-6h3c`。当前 follow-redirects 告警为 0；此次 fresh audit 与首次实施后结果一致。

生产 high 仍是 xlsx 的 `GHSA-4r6h-8v6p-xvw6` / `GHSA-5pgg-2g8v-p4x9`。原六条例外没有续期，`.github/audit-exceptions.yml` 仍为空；[实际例外门禁](./audit-exception-gate-final.txt)按这两条告警退出 1。不能宣称依赖安全门禁全绿。xlsx 发布渠道/导出实现替换及 Vite/Vitest 等其余工具链维护仍属于后续工作，本次没有扩大升级范围。

全部剩余前端记录按包列出；完整 advisory、版本范围和路径位于原始 audit 及 [dependency-facts.json](./dependency-facts.json)。生产记录另含 yaml、uuid、mermaid、decode-uri-component、colord 和 katex 的 moderate/low，未将它们当作已修复。

| 剩余包 | 全树记录数 |
|---|---:|
| @vitest/mocker | 2 |
| ajv | 1 |
| baseline-browser-mapping | 1 |
| brace-expansion | 14 |
| browserslist | 2 |
| colord | 1 |
| decode-uri-component | 1 |
| esbuild | 1 |
| flatted | 2 |
| js-yaml | 4 |
| katex | 1 |
| lodash | 3 |
| mermaid | 9 |
| minimatch | 6 |
| nanoid | 1 |
| picomatch | 4 |
| postcss-selector-parser | 2 |
| rollup | 1 |
| tinypool | 2 |
| uuid | 2 |
| vite | 8 |
| vitest | 2 |
| ws | 2 |
| xlsx | 2 |
| yaml | 1 |

## 当前 Go 依赖图扫描

[原始 JSON stream](./audit-govulncheck-final.txt)及[命令记录](./audit-govulncheck-final.json)：`go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 -scan=package -json ./...`。扫描配置确认 Go 1.27.3，数据库 `https://vuln.go.dev` 的修改时间为 `2026-10-21T00:27:29Z`，与[更新必要前置前](./s10-govulncheck.txt)一致。扫描命令退出 0，输出仍有 findings；退出码不能代替漏洞记录检查。

| 口径 | 必要前置前 | 当前 |
|---|---:|---:|
| finding 记录 | 24 | 18 |
| 不同 OSV 编号 | 19 | 13 |
| 含具体 package 的 finding | 5 | 5 |
| stdlib finding | 0 | 0 |

减少的 6 个 OSV 为 `GO-2026-6179`、`GO-2026-6180`、`GO-2026-6303`、`GO-2026-6354`、`GO-2026-6355`、`GO-2026-6629`，分别随必需的 x/mod、x/crypto、x/text 更新消失；没有新增 OSV。剩余项如下：

| 模块 / 当前版本 | 剩余 OSV | 数据库修复版本 |
|---|---|---|
| klauspost/compress v1.18.2 | GO-2026-5841 | v1.18.7 |
| x/crypto v0.57.0 | GO-2026-5932 | 当前记录未提供 |
| otel v1.41.0 | GO-2026-5158 | v1.42.0 |
| x/image v0.41.0 | GO-2026-4961、5061、5062、5066、6222 | 分别为 v0.42.0、v0.43.0、v0.43.0、v0.43.0、v0.45.0 |
| x/net v0.59.0 | GO-2026-6603、6610、6611、6612、6617 | v0.60.0 |

其中 x/net 的 5 个 OSV 各有模块及 `golang.org/x/net/http2` 包两条 finding，所以总记录为 18。此处是 package 级扫描，没有调用链可达性证明；既不据此声称全部可利用，也不将模块级记录视为已消除。CI 的默认 symbol 扫描与这里深度不同，本报告不将 package JSON 的退出 0 等同于完整 CI 安全通过。

## 回归与范围

[frozen install](./s10-install-final.txt)、[typecheck](./frontend-typecheck.txt)、[lint](./frontend-lint.txt)、[build](./frontend-build.txt)通过；[完整 Vitest](./frontend-tests.txt)为 2027 passed / 2 既有 skipped。当前 Go 依赖前置后的 Ent schema 与 server build 通过；[tier1 保留回归](./audit-tier1-preservation.txt)在四包通过，覆盖原六簇及真实 SQLite 去重/删除 Key/固定窗口路径。

[scope-facts-final.json](./scope-facts-final.json)核对实施起点的 52 个既有文件均在，快照 hash 全部匹配：43 个当前字节不变，9 个有增量。增量为 AGENTS / README 的当前状态说明及 PORTING 的第二档实施补记，S06 的 billing/pricing 别名，S07/S08 的 Gemini 共享服务，S02/S04 的 OpenAI passthrough 与附加容量测试，以及默认编译发现后将 unit 专属指针测试 helper 改为局部变量指针的 Grok 测试夹具。PORTING §1–9 评估与 §10 第一档补记逐段核对不变；tier1 历史 OpenSpec 和实施证据 hash 全部保持起点。没有整文件回退或删除原实现，既有 tier1 专属回归已通过。

HEAD 保持 `498432759`；两批 source-baseline、评估证据、Ent/schema、迁移、VERSION 与 audit 例外文件均为 0 diff。VERSION 仍为 1.1.18、最高迁移 226。已检查当前新增及修改路径属于 S01–S10、明确量测的工具链前置、测试夹具确定性修复和实施证据，没有支付/新平台/新 schema 混入。[code-sha256.json](./code-sha256.json)固定最终 135 个 production/tests/config/dependency/workflow 文件（含未改的 VERSION / audit 例外控制文件），不含会自引用的文档或证据。两项全量偶发失败的夹具修复有 8 轮定向、3 轮 race、默认编译和线性复制负对照，详 [audit-review](./audit-review.md)。最终完整 unit/race/build/SQLite 及其它门禁结论统一见 [verification](../../../openspec/changes/archive/2026-10-24-port-upstream-0.2.18-tier2/verification.md)，这里不提前声明尚在执行的 unit-complete 结果。未 commit/push、发布或部署。
