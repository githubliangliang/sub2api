# 当前依赖审计与旧待办重测

日期2026-10-13；基线128c5f7bf。命令`cd frontend && pnpm audit --json`返回非零，原因是发现告警；[原始响应](./frontend-audit.json)、[stderr](./frontend-audit-stderr.txt)。未执行升级/install，未修改lockfile。

103条registry advisory记录：critical=3、high=42、moderate=50、low=8。记录可包含同GHSA在不同版本范围的重复分支；不等于103个当前可利用漏洞。3条critical均在Vitest/tinypool开发工具链，本轮没有运行或验证攻击PoC。

| 包 | 记录数 | 当前判断 |
|---|---|---|
| DOMPurify | 19 | 7个sanitize调用含公开LegalDocumentView；旧0.2.0阶段9仍未做。本次最低修复版本集合最大到3.4.16，旧3.4.14计划不足。S16接管，生产内容净化与SVG都需回归。 |
| axios | 12 | registry指向>=1.20.0；此前#7780依赖子项无新证据，本轮已明确从第三档提升S16。不能把Node特定advisory自动说成浏览器可利用。 |
| @vue/server-renderer | 1 | registry指向>=3.5.42；项目SPA未证明暴露SSR攻击面，列为现有依赖维护而非已复现漏洞；#7893来源Vue3.5.43为评估参考。 |
| source-map-js | 1 | registry指向>=1.2.2；#7893维护子项随S16，两个pnpm版本配置位置都需对齐。 |
| xlsx | 2 | 仍^0.18.5，本地生产只发现动态导出write；源码未证明read利用路径。旧例外已过期，不自动续至上游2027-01-06；无npm发布修复路径的替换/CDN方案独立设计。 |
| 其它 | 68 | 包含mermaid/工具链等；本轮不假设S16解决全部记录。实施应输出剩余项及真实依赖路径。 |

`.github/audit-exceptions.yml`现有6条例外均2026-10-06到期。审计结果和例外到期是独立问题：不能靠延期让告警消失，也不能因前端typecheck/vitest通过宣称安全扫描通过。

前端pnpm11提示package.json的pnpm.overrides不再读取；本仓库已有pnpm-workspace.yaml，S16必须在该文件配置适用覆盖，并保持CI使用的package.json设置相容。具体可安装版本在实施时再以registry/lock解析核实；当前报告没有改包版本。
