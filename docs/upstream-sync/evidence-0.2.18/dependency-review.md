# 2026-10-23 依赖复核

基线873962ce7 / VERSION1.1.18；未修改依赖或锁文件。

## 前端

- `pnpm audit --json`：78条advisory记录，3 critical / 36 high / 35 moderate / 4 low，退出1。
- `pnpm audit --prod --json`：19条，0 critical / 4 high / 13 moderate / 2 low，退出1。
- follow-redirects1.16.0经axios1.20.0引入，命中[GHSA-q3jh-wrc6-jgq8](https://github.com/advisories/GHSA-q3jh-wrc6-jgq8)及[GHSA-8r9p-f939-6h3c](https://github.com/advisories/GHSA-8r9p-f939-6h3c)，修复版本>=1.16.1。进入S10定向维护。
- 浏览器SPA通常使用axios浏览器适配器，npm生产依赖分类不等于已证明后端执行follow-redirects；维护依据是实际依赖命中和兼容修复版本，没有声称当前站点已被利用。
- DOMPurify package/workspace/lock均3.4.17，audit无该包记录；上游828626804不重复实施。Vue3.5.44/source-map-js1.2.2等旧修复保留。
- xlsx0.18.5仍有两条high（GHSA-4r6h-8v6p-xvw6、GHSA-5pgg-2g8v-p4x9）；历史只发现导出路径不等于免除审计，继续单独维护，不续期例外。
- 开发critical为vitest2.1.9一条与tinypool1.1.1两条；需要工具链成组升级验收，第三档独立维护，不随安全补丁整取上游lockfile。
- 生产其它moderate/low以及全量其余记录以两个JSON为准。记录数不是不同漏洞或可利用站点数；不能用旧轮次70替代当前78。

## Go

命令：`go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 -scan=package -json ./...`，在backend执行，退出0。工具配置实际显示go1.27.0，数据库vuln.go.dev最后更新时间2026-10-21T00:27:29Z。没有构建生产二进制或调用外部业务接口。

原始流中共64条finding，28个去重advisory，其中14个有package级命中；同一advisory可能同时有module/package记录。不是符号可达性扫描，退出0也不表示没有风险。OSV数据块还包含历史已修记录，**统计只取finding**，不数所有OSV块。

标准库发现1.27.2修复批次及[GO-2026-6809](https://pkg.go.dev/vuln/GO-2026-6809)（crypto/x509，修复1.27.3），支持S10将Go从1.27.0定向升级到1.27.3，并同步所有CI/Docker断言。这是本轮新实测证据：旧轮次工具链按需维护中的Go安全更新明确接管到第二档，不改变其它工具链项的档位。

其它模块为x/net、x/crypto、x/image、x/text、x/mod、klauspost/compress、otel；具体版本、修复版本及package/module层级见[findings](./backend-vulnerability-findings.tsv)。本轮上游两PR未提供这些依赖的成组迁移，保留为独立维护待办，需要调用可达性/兼容回归后再选定升级范围；没有据本次扫描宣布整个后端安全。

最初两次module扫描命令分别因不接受pattern及backend根无Go文件未完成，随后使用上述package模式得到有效完整结果；未将前两次当通过。

## 实施约束

S10只接Go1.27.3与follow-redirects修复。pnpm11用workspace overrides，Docker/CI pnpm9仍读package overrides，保持两处一致。保留SQLite构建，不导入PG镜像阶段或支付依赖。重跑audit和Go扫描，按目标finding与版本差异验收，剩余记录明确列出，不盲目忽略或全量升级。
