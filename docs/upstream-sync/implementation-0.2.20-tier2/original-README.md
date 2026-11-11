# v0.2.20 第二档：既有路径修复与适配

状态：2026-11-10 评估完成，**尚未实施**。共11簇；评估期测试不是实施验收。当前基线为Fast完成后的`326513cdc68cc50d31e25a07ae98ecd6f5a3ca1f` / VERSION1.1.23 / Go1.27.3 / SQLite228。

| 编号 | 范围 | 来源 | 行为 |
|---|---|---|---|
| S01 | Codex Responses非流式与instructions回显 | PR7979, PR8499 | [规格](specs/codex-responses-end-to-end/spec.md) |
| S02 | Grok 委派消息最小适配 | PR8465 | [规格](specs/grok-delegation-input/spec.md) |
| S03 | 已使用 Terra 的目录与计费后续 | PR8365 | [规格](specs/terra-existing-model-followup/spec.md) |
| S04 | 请求体预算、读写超时与错误分类 | PR8484, PR8518, PR8457 | [规格](specs/gateway-ingress-limits/spec.md) |
| S05 | 上游流、drain与会话缓存有界 | PR8484, PR8518 | [规格](specs/gateway-stream-resource-limits/spec.md) |
| S06 | Antigravity 系统身份保留任务 | PR8318 | [规格](specs/antigravity-system-input/spec.md) |
| S07 | 图片 INTERNAL500 不牵连整账号 | PR8521 | [规格](specs/antigravity-model-cooldown/spec.md) |
| S08 | Gemini 工具媒体原生承载 | PR8430 | [规格](specs/gemini-tool-media/spec.md) |
| S09 | 使用记录导出限流恢复 | PR8436 | [规格](specs/usage-export-resilience/spec.md) |
| S10 | Antigravity 测试提示词贯通 | PR8531 | [规格](specs/account-test-correctness/spec.md) |
| S11 | RPM空覆盖缓存与既有JSON预检 | PR8518 | [规格](specs/gateway-hotpath-maintenance/spec.md) |

[proposal](proposal.md) / [design](design.md) / [source-baseline](source-baseline.md) / [source-feature-map](source-feature-map.md) / [tasks](tasks.md) / [verification](verification.md)。

[完整评估与patch sites](../../../docs/upstream-sync/PORTING-0.2.20.md) / [评估证据](../../../docs/upstream-sync/evidence-0.2.20/MANIFEST.md)。SQLite、miniredis、simple mode、Fast账号目录/费率/Free Fast、成功usage去重和固定quota窗口必须保留。没有新支付、部署、发布、push或上游版本号替换计划。
