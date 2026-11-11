# v0.2.20 第二档实施证据

2026-11-10，用户授权后S01–S11完成，38/38任务关闭。起点`29681fdf5968cb3482efa35694bc27cd26ecb802`。详细[验收与范围](../../../openspec/changes/port-upstream-0.2.20-tier2/verification.md)。实现于2026-11-11随本次提交保存，等待归档。

| 证据 | 内容 |
|---|---|
| baseline.json / before-hashes.json | 干净起点及6744既有文件哈希 |
| original-README/tasks/verification.md | 实施前状态，原source/proposal/design/spec不改 |
| s01-s03-before.log / last-terminal-before.log | 基线9个顶层旧行为失败；S01追加6个末尾终态子场景失败 |
| acceptance-unit.log/json | 最终完整backend unit |
| acceptance-build / acceptance-lint | embed构建与lint |
| backend-race / final-additions-race / cache-final-race / last-terminal-final | 主路径、资源、存储和最后修正的race证据 |
| settlement-current / final-boundaries / sqlite-protection | 本地HTTP+SQLite结算去重/Fast/Grok、RPM真实SQLite+miniredis、迁移与配额保护 |
| prompt-http / patch-budget-check / quota-worker | 管理员prompt实际HTTP、100次预算释放及过载不转发、配额任务边界 |
| frontend-typecheck / frontend-lint-final / frontend-full-final / frontend-critical / frontend-build | 前端完整门禁与既有警告 |
| browser-check.mjs / browser/ | 实际构建UI与Chrome，拦截的合成API；CSV/截图/请求清单；工具只安装于/tmp |
| memory-1g-stress / body-stress-build | 1GiB/1CPU无外网容器运行静态测试二进制，64MiB预算过载/恢复及memory.peak |
| thinking-benchmark / delegation-benchmark / thinking-fuzz | 实测优化及等价性，非供应商速度宣称 |
| wire-generate-final / server-cleanup-final | Wire由源码生成及关闭流程测试 |
| environment.json / product-hashes.json / product.diff | 实际版本/默认值/提交前验证的代码哈希；diff不含新文件内容，新文件由哈希清单定位 |
| checks-summary.json / scope-check.json / openspec-tier2 | 命令结果、冻结来源/范围和严格规格验证 |
| run-check.py / 各同名JSON | argv、cwd、UTC时间、耗时、退出码，失败过程一并保留 |

配置：gateway请求体预算64MiB、单次读写停滞120秒、SSE默认行40MiB/队列8MiB/事件4096行；断开drain180秒。预算不是总进程heap上限，压力结果仅对应本地读体层。auth snapshot21；VERSION1.1.23/Go1.27.3/SQLite228、依赖/Ent不变；Wire只增加共享缓存清理接线。上游MCP、支付、新平台、新图片链路和tag后提交未纳入。
