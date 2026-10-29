# 依赖维护与复扫（2026-10-29）

只采用 x/net 0.62.0 与实测强制 MVS 变化；实际图与临时最小图一致。Go 1.27.3、VERSION 1.1.20、迁移 227 保留。

| 模块 | 原值 | 实际值 |
|---|---|---|
| golang.org/x/crypto | v0.57.0 | v0.58.0 |
| golang.org/x/net | v0.59.0 | v0.62.0 |
| golang.org/x/sync | v0.23.0 | v0.24.0 |
| golang.org/x/sys | v0.48.0 | v0.49.0 |
| golang.org/x/telemetry | v0.0.0-20260908163034-4bcc4b2ee518 | v0.0.0-20260924152758-ed294f943157 |
| golang.org/x/term | v0.46.0 | v0.47.0 |
| golang.org/x/text | v0.42.0 | v0.43.0 |
| golang.org/x/tools | v0.50.0 | v0.51.0 |

[临时最小图](./s06-minimum-mvs.json) / [实际升级](./s06-upgrade.json) / [图与扫描摘要](./dependency-audit-summary.json)。go.mod 是 7 个已有版本声明变化；telemetry 是第 8 项解析图变化，没有覆盖上游整份 modfile。

GO-2026-6860 未出现在当前 finding 中。独立 HTML 输入 40960→81920 bytes，旧图分配 35.85→141.97 MB（3.960 倍），新图 0.114→0.216 MB（1.902 倍）。[修复前](./s06-html-before.txt) / [修复后](./s06-html-after.txt)。实际消费者为 ollama_cloud_usage_parser.go；这是解析复杂度与既有消费者证据，不宣称匿名 gateway 利用性或完整进程压测。

[Ent/Wire 隔离生成](./generation-compatibility.json)通过，290 个生成文件内容一致，工作树生成代码未更新。

## 剩余 Go finding

下列为模块/包/符号层合并后的唯一 ID，不把模块匹配冒充可达调用。完整 trace 在[原始扫描](./govulncheck.txt)。

- [GO-2026-4961](https://pkg.go.dev/vuln/GO-2026-4961)
- [GO-2026-5061](https://pkg.go.dev/vuln/GO-2026-5061)
- [GO-2026-5062](https://pkg.go.dev/vuln/GO-2026-5062)
- [GO-2026-5066](https://pkg.go.dev/vuln/GO-2026-5066)
- [GO-2026-5158](https://pkg.go.dev/vuln/GO-2026-5158)
- [GO-2026-5841](https://pkg.go.dev/vuln/GO-2026-5841)
- [GO-2026-5932](https://pkg.go.dev/vuln/GO-2026-5932)
- [GO-2026-6222](https://pkg.go.dev/vuln/GO-2026-6222)

## 前端复扫

package、workspace 与 lockfile 未变。Registry 结果与旧冻结评估可能有时间差；旧例外不续期，其它升级保持独立维护。

### production

计数：{"info": 0, "low": 4, "moderate": 13, "high": 5, "critical": 0}。扫描退出码 1 为已报告 finding，不记成无漏洞通过。

| 模块 | 严重度 | Advisory |
|---|---|---|
| xlsx | high | [1108110](https://github.com/advisories/GHSA-4r6h-8v6p-xvw6) Prototype Pollution in sheetJS |
| xlsx | high | [1108111](https://github.com/advisories/GHSA-5pgg-2g8v-p4x9) SheetJS Regular Expression Denial of Service (ReDoS) |
| @intlify/shared | high | [1243637](https://github.com/advisories/GHSA-83rm-6c8r-g728) Vue I18n: escapeParameter does not prevent XSS through srcdoc in translation messages |
| @intlify/core-base | high | [1243649](https://github.com/advisories/GHSA-83rm-6c8r-g728) Vue I18n: escapeParameter does not prevent XSS through srcdoc in translation messages |
| vue-i18n | high | [1243657](https://github.com/advisories/GHSA-83rm-6c8r-g728) Vue I18n: escapeParameter does not prevent XSS through srcdoc in translation messages |

### full

计数：{"info": 0, "low": 6, "moderate": 36, "high": 35, "critical": 3}。扫描退出码 1 为已报告 finding，不记成无漏洞通过。

| 模块 | 严重度 | Advisory |
|---|---|---|
| xlsx | high | [1108110](https://github.com/advisories/GHSA-4r6h-8v6p-xvw6) Prototype Pollution in sheetJS |
| xlsx | high | [1108111](https://github.com/advisories/GHSA-5pgg-2g8v-p4x9) SheetJS Regular Expression Denial of Service (ReDoS) |
| minimatch | high | [1113459](https://github.com/advisories/GHSA-3ppc-4f35-3m26) minimatch has a ReDoS via repeated wildcards with non-matching literal in pattern |
| minimatch | high | [1113465](https://github.com/advisories/GHSA-3ppc-4f35-3m26) minimatch has a ReDoS via repeated wildcards with non-matching literal in pattern |
| rollup | high | [1113515](https://github.com/advisories/GHSA-mw96-cpmx-2vgc) Rollup 4 has Arbitrary File Write via Path Traversal |
| minimatch | high | [1113538](https://github.com/advisories/GHSA-7r86-cg39-jmmj) minimatch has ReDoS: matchOne() combinatorial backtracking via multiple non-adjacent GLOBSTAR segments |
| minimatch | high | [1113544](https://github.com/advisories/GHSA-7r86-cg39-jmmj) minimatch has ReDoS: matchOne() combinatorial backtracking via multiple non-adjacent GLOBSTAR segments |
| minimatch | high | [1113546](https://github.com/advisories/GHSA-23c5-xmqv-rm74) minimatch ReDoS: nested *() extglobs generate catastrophically backtracking regular expressions |
| minimatch | high | [1113552](https://github.com/advisories/GHSA-23c5-xmqv-rm74) minimatch ReDoS: nested *() extglobs generate catastrophically backtracking regular expressions |
| flatted | high | [1114526](https://github.com/advisories/GHSA-25h7-pfq9-p65f) flatted vulnerable to unbounded recursion DoS in parse() revive phase |
| picomatch | high | [1115552](https://github.com/advisories/GHSA-c2c7-rcm5-vvqj) Picomatch has a ReDoS vulnerability via extglob quantifiers |
| picomatch | high | [1115554](https://github.com/advisories/GHSA-c2c7-rcm5-vvqj) Picomatch has a ReDoS vulnerability via extglob quantifiers |
| lodash | high | [1115806](https://github.com/advisories/GHSA-r5fr-rjxr-66jc) lodash vulnerable to Code Injection via `_.template` imports key names |
| ws | high | [1123259](https://github.com/advisories/GHSA-96hv-2xvq-fx4p) ws: Memory exhaustion DoS from tiny fragments and data chunks |
| brace-expansion | high | [1123896](https://github.com/advisories/GHSA-3jxr-9vmj-r5cp) brace-expansion: DoS via exponential-time expansion of consecutive non-expanding {} groups |
| brace-expansion | high | [1123897](https://github.com/advisories/GHSA-3jxr-9vmj-r5cp) brace-expansion: DoS via exponential-time expansion of consecutive non-expanding {} groups |
| js-yaml | high | [1123911](https://github.com/advisories/GHSA-52cp-r559-cp3m) js-yaml: YAML merge-key chains can force quadratic CPU consumption |
| brace-expansion | high | [1130588](https://github.com/advisories/GHSA-mh99-v99m-4gvg) brace-expansion: DoS via unbounded expansion length causing an out-of-memory process crash |
| brace-expansion | high | [1130589](https://github.com/advisories/GHSA-mh99-v99m-4gvg) brace-expansion: DoS via unbounded expansion length causing an out-of-memory process crash |
| brace-expansion | high | [1130736](https://github.com/advisories/GHSA-rgw5-rvv9-x895) brace-expansion: DoS via unbounded intermediate arrays, bypassing the CVE-2026-14257 mitigation |
| brace-expansion | high | [1130737](https://github.com/advisories/GHSA-rgw5-rvv9-x895) brace-expansion: DoS via unbounded intermediate arrays, bypassing the CVE-2026-14257 mitigation |
| js-yaml | high | [1138115](https://github.com/advisories/GHSA-5p4m-2wfm-xmqj) JS-YAML: Quadratic CPU consumption in !!omap resolution (3.x and 4.x) — CVE-2026-59870 fix not backported |
| nanoid | high | [1139427](https://github.com/advisories/GHSA-2v37-7h3g-55p8) nanoid: custom generators can loop indefinitely when size is zero |
| vitest | critical | [1139528](https://github.com/advisories/GHSA-5xrq-8626-4rwp) When Vitest UI server is listening, arbitrary file can be read and executed |
| browserslist | high | [1153171](https://github.com/advisories/GHSA-c83g-rgw3-j3cx) Browserslist: Unbounded memory growth (no cache eviction) via distinct query results, leading to eventual OOM |
| browserslist | high | [1153172](https://github.com/advisories/GHSA-73wf-gq98-2v4g) Browserslist: Uncaught crash / prototype write via untrusted browserslist-stats.json custom stats (normalizeStats) |
| js-yaml | high | [1193727](https://github.com/advisories/GHSA-2883-xcg3-v3hh) js-yaml: maxTotalMergeKeys does not limit CPU use for empty merge sources |
| brace-expansion | high | [1240104](https://github.com/advisories/GHSA-qhr7-859c-m2p7) brace-expansion: DoS via uncontrolled recursion on nested brace groups causing stack exhaustion |
| brace-expansion | high | [1240105](https://github.com/advisories/GHSA-qhr7-859c-m2p7) brace-expansion: DoS via uncontrolled recursion on nested brace groups causing stack exhaustion |
| brace-expansion | high | [1240108](https://github.com/advisories/GHSA-6j4f-fj2g-mc7p) brace-expansion: DoS via uncontrolled recursion in parseCommaParts causing stack exhaustion |
| brace-expansion | high | [1240109](https://github.com/advisories/GHSA-6j4f-fj2g-mc7p) brace-expansion: DoS via uncontrolled recursion in parseCommaParts causing stack exhaustion |
| flatted | high | [1241196](https://github.com/advisories/GHSA-rf6f-7fwh-wjgh) Prototype Pollution via parse() in NodeJS flatted |
| tinypool | critical | [1241260](https://github.com/advisories/GHSA-5gmw-xhrv-c9v3) Tinypool: Prototype Pollution gadget in worker options leads to Remote Code Execution |
| tinypool | critical | [1241261](https://github.com/advisories/GHSA-85c8-ppgw-ccpr) Tinypool: Prototype Pollution Gadget to RCE in run() options |
| vite | high | [1242658](https://github.com/advisories/GHSA-587h-7wmj-h5pm) launch-editor: `fs.existsSync()` existence check can be bypassed on Windows |
| @intlify/shared | high | [1243637](https://github.com/advisories/GHSA-83rm-6c8r-g728) Vue I18n: escapeParameter does not prevent XSS through srcdoc in translation messages |
| @intlify/core-base | high | [1243649](https://github.com/advisories/GHSA-83rm-6c8r-g728) Vue I18n: escapeParameter does not prevent XSS through srcdoc in translation messages |
| vue-i18n | high | [1243657](https://github.com/advisories/GHSA-83rm-6c8r-g728) Vue I18n: escapeParameter does not prevent XSS through srcdoc in translation messages |
