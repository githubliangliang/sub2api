# 评估阶段基线验证

日期：2026-09-28；固定产品 SHA：`bfbcd79bdc4aa0838616405466750658896e6920`。
本次只评估并写文档；没有应用上游产品补丁。下面的 PASS 是旧基线证据，不代表新需求通过。

| 检查 | 结果 |
|---|---|
| Git 发布/tag | GitHub API 与远端 tag 对照；最新稳定 v0.2.9，发布时间 2026-09-28T03:08:59Z |
| 区间/候选 | 215 non-merge commits；161 PR + 9 first-parent direct commits；867 file patches |
| 四态 | 固定 HEAD 的独立 index；先完成全部反向检查，再正向检查；不落工作树 |
| 前端依赖 | pnpm 10.22.0，frozen lockfile；镜像曾 ECONNRESET，补齐缓存后安装成功；没有保留 packageManager 自动写入 |
| 前端 typecheck | `pnpm run typecheck`，退出码0 |
| 前端定向 Vitest | 8 files / 42 tests passed，退出码0，29.33秒 |
| 后端环境 | PATH 与常用安装路径未找到 Go；未运行 `go test -list`、unit、build、race，不记 PASS |
| 测试来源 | `test-inventory.tsv` 区分实际存在、新增待适配、构建标签；名称清单来自源码，不是 go test -list |
| PG-only 测试 | migrations_schema_integration_test.go 首行 `//go:build integration && postgres`，不得作为 SQLite 验收 |
| 浏览器/真实上游/SQLite 新行为 | 未运行；没有声称新补丁已通过运行验收 |

前端命令（工作目录 `frontend/`）：

```powershell
$env:COREPACK_ENABLE_AUTO_PIN='0'
pnpm run typecheck
pnpm exec vitest run src/utils/__tests__/ccswitchImport.spec.ts src/components/keys/__tests__/UseKeyModal.spec.ts src/components/common/__tests__/Select.spec.ts src/components/common/__tests__/DateRangePicker.spec.ts src/components/common/__tests__/BaseDialog.spec.ts src/components/account/__tests__/UsageProgressBar.spec.ts src/components/user/profile/__tests__/totp-timer-cleanup.spec.ts src/components/auth/__tests__/TotpLoginModal.spec.ts
```

执行日志：同目录 frontend-typecheck.txt、frontend-vitest.txt。OpenSpec strict 与文档自检另见 document-validation.md。
