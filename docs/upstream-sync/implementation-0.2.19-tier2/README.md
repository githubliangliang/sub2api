# v0.2.19 第二档实施证据

S01–S06 已实施，23/23 任务完成，代码提交本地 main `cccf931efe201671f8cd27284b04e74114f4f894`。实际起点为本地 main `9b44cab33a04f3a9fe1efd0102b5c562a1c458a7`，工作树干净。[OpenSpec 验收](../../../openspec/changes/port-upstream-0.2.19-tier2/verification.md) / [行为与反例](./acceptance-matrix.md) / [范围与 SHA256](./scope-facts.json)。

Grok 原生和两条转换保留响应原文模型/冲突位，目录严格解析、成功/失败和来源/时间排序通过单键 SQLite CAS；普通编辑保留数据库最新受管状态。OpenAI 上传保留 20 MiB 并明确拒绝超限，Grok 保留总 body 边界内完整数据。渠道保存两维及未知维度，Gemini 声明目录与既有映射/展示同源，两个测试组件采用已有 3.5 Flash 默认，保留 legacy/显式选择。x/net 0.62.0 与 8 项 MVS 变化经量测采用。

VERSION 1.1.20、Go 1.27.3、SQLite 227、Ent/Wire 与前端 package/lockfile 保留。用户确认的 S01 验收修订、[原规格](./s01-spec-original.md)单列；154 个其它冻结来源/规格文件保持原文。没有 Terra、新平台/新模型目录/诊断 UI、支付、发布/tag/push/部署或业务实例操作。

[依赖报告](./dependency-report.md)保留剩余 Go/前端 finding，不把 x/net 目标修复说成全图无漏洞。HTML 独立分配比从 3.960 降至 1.902；实际图与临时最小图一致。Ent/Wire 在临时副本生成，290 个文件与工作树一致。

前端 2105 tests passed、2 既有 skipped，typecheck/lint/build/critical 通过；[真实浏览器](./browser/README.md)验证实际保存与当前管理测试窗口。[原始 gate JSON/TXT](./final-unit.json)由 run-check.py 保留 argv、cwd、时间、耗时和退出码。

## 最终门禁

49 个产品文件（26 个生产/模块文件、23 个测试文件），汇总 SHA256 `7c81becbc895fc4117a4ea5cbd4a017c7845b7e9a77d8839d0ae275b3f14936d`。

| 检查 | 证据 | 结果 |
|---|---|---|
| 完整 backend unit | [54 包](./final-unit.txt) / [命令](./final-unit.json) | PASS |
| server build / lint | [build](./backend-build.json) / [lint](./final-lint.txt) | PASS / 0 issues |
| 相关 race | [service](./backend-service-race-complete.txt)、[handler/repository/securityaudit](./backend-race-final.txt)、[network/curated models](./backend-network-race.txt)；仅保留各包最终成功项，过程合并文件的失败项由后续输出替代 | PASS / 7 包实际测试 |
| SQLite227 / quota / fixed window / dedup | [实际输出](./sqlite-protection.txt) | PASS |
| frontend | [2105 passed、2 skipped](./frontend-tests-complete.txt)、[typecheck](./final-frontend-typecheck.txt)、[lint](./final-frontend-lint.txt)、[build](./frontend-build-complete.txt)、[176 critical](./frontend-critical-final.txt) | PASS |
| OpenSpec / 链接 / 来源 / 完成记录 | [strict](./openspec-strict-final.txt) / [最终范围检查](./implementation-validation.txt) | PASS |

代码提交仅包含产品/测试；证据与完成文档作为后续文档提交。所有失败尝试保留，不作完成结果。
