# S16 定向依赖维护

实施查询：2026-10-17；验收截至 2026-10-18。当前 registry `pnpm view <package> version`、pnpm 安装和 audit 原始响应见本目录，冻结评估报告保持不变。

| 包 | 实施前 | 实施后 | 控制方式 |
|---|---|---|---|
| DOMPurify | 3.3.1 / 3.3.3 | 3.4.17 | 直接依赖和全部传递解析统一 override |
| axios | 1.18.1 | 1.20.0 | 直接依赖与 `<1.20.0` override |
| Vue | 3.5.26 | 3.5.44 | 直接依赖、peer 解析；server-renderer 固定同版本 |
| source-map-js | 1.2.1 | 1.2.2 | `<1.2.2` override |

pnpm-workspace.yaml 与 package.json.pnpm.overrides 保持一致。锁文件由 pnpm 11.21.0 生成；隔离目录中用 pnpm 9.15.9 `install --lockfile-only --frozen-lockfile --ignore-scripts` 通过，锁文件 SHA256 相同。首次直接在工作区运行 pnpm 9 只出现重装提示，没有执行验证，故不作为验收证据。Vue 更新引入对应编译器/Babel 依赖变化；原支付 SDK（Airwallex 1.30.2 / Stripe 9.0.1）未更新，也未添加新支付依赖。没有整取上游 lockfile。

## 审计结果

- 全树 [before](./dependency-audit-before.json)：103 条记录，critical 3 / high 42 / moderate 50 / low 8。
- 全树 [after](./dependency-audit.json)：70 条记录，critical 3 / high 33 / moderate 31 / low 3。DOMPurify、axios、@vue/server-renderer、source-map-js 的目标告警均为 0，共减少 33 条。
- [生产 high/critical 审计](./dependency-audit-prod.json)仍有 **xlsx 2 条 high**；元数据另计生产 moderate 13 / low 2。扫描非零，不能宣称安全审计全绿。
- 六条旧例外已于 2026-10-06 过期，本轮全部移除，没有续期或新增豁免。[安全门禁](./dependency-audit-gate.txt)按实际剩余 xlsx 告警返回 1。

## 明确未完成的旧维护范围

- xlsx 0.18.5：npm 路径无修复版本，GHSA-4r6h-8v6p-xvw6 / GHSA-5pgg-2g8v-p4x9 仍在。替换导出实现或改用其它发布渠道属于单独维护；动态 import 和管理端用途不等于漏洞已修复，也不自动构成豁免。
- Vite/Vitest/tinypool 等工具链及其余传递依赖：另行维护，不为本批同步整取大版本。保留下表所有剩余项，不将其描述为不可利用；重新设置例外须有实际条件与负责人复核。
- 接管的旧 0.2.0 DOMPurify 及 0.2.12 axios 待办在本次选定范围内完成；旧报告与旧任务框保留历史状态，当前结论由本报告与 tier2 verification 记录。

| 剩余包 | 全树记录数 |
|---|---|
| `@vitest/mocker` | 1 |
| `ajv` | 1 |
| `baseline-browser-mapping` | 1 |
| `brace-expansion` | 14 |
| `braces` | 1 |
| `browserslist` | 2 |
| `colord` | 1 |
| `decode-uri-component` | 1 |
| `esbuild` | 1 |
| `flatted` | 2 |
| `js-yaml` | 4 |
| `katex` | 1 |
| `lodash` | 3 |
| `mermaid` | 9 |
| `minimatch` | 6 |
| `nanoid` | 1 |
| `picomatch` | 4 |
| `postcss-selector-parser` | 2 |
| `rollup` | 1 |
| `tinypool` | 2 |
| `uuid` | 2 |
| `vite` | 3 |
| `vitest` | 2 |
| `ws` | 2 |
| `xlsx` | 2 |
| `yaml` | 1 |

## 回归证据

最终前端全量 2027 passed / 2 既有 skipped；包含真实 DOMPurify 的 Markdown 页面净化用例（script/事件属性/javascript URL 移除，正常图片、链接和既有 iframe 配置保留），以及 API client 的27个错误/响应处理用例。typecheck、lint 与 build 通过，详见本目录最终输出。
