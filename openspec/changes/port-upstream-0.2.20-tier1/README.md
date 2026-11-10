# v0.2.20 第一档：既有路径修复与适配

状态：2026-11-10用户授权后，**13簇、44/44任务已完成**；实现随本次提交保存，待归档。[实际验收](verification.md)与[实施证据](../../../docs/upstream-sync/implementation-0.2.20-tier1/README.md)记录本批验证。当前基线为Fast完成后的`326513cdc68cc50d31e25a07ae98ecd6f5a3ca1f` / VERSION1.1.23 / Go1.27.3 / SQLite228。

| 编号 | 范围 | 来源 | 行为 |
|---|---|---|---|
| F01 | Claude setup-token 周用量 | PR8553 | [规格](specs/claude-usage-windows/spec.md) |
| F02 | Codex 强制模式停止多余探测 | PR8445 | [规格](specs/codex-gateway-correctness/spec.md) |
| F03 | WebSocket 每轮首 token 时间 | PR8520 | [规格](specs/codex-gateway-correctness/spec.md) |
| F04 | Codex WS→HTTP 桥接指纹 | PR8464 | [规格](specs/codex-gateway-correctness/spec.md) |
| F05 | 缓冲 Responses reasoning 保真 | PR8563 | [规格](specs/responses-compatibility/spec.md) |
| F06 | Responses 完成消息字段 | PR8429 | [规格](specs/responses-compatibility/spec.md) |
| F07 | 鉴权取消分类 | PR8471 | [规格](specs/gateway-errors/spec.md) |
| F08 | 413 透传规则优先级 | PR8437 | [规格](specs/gateway-errors/spec.md) |
| F09 | Antigravity 直连 SSE 响应 | PR8441 | [规格](specs/antigravity-correctness/spec.md) |
| F10 | Antigravity JSON Schema 参数 | PR8409 | [规格](specs/antigravity-correctness/spec.md) |
| F11 | GPT Image 账号测试参数 | PR8432 | [规格](specs/account-test-correctness/spec.md) |
| F12 | Anthropic兼容 Cloudflare REST URL | PR8420 | [规格](specs/claude-endpoint-compatibility/spec.md) |
| F13 | Bedrock EventStream 分配前边界 | PR8518 | [规格](specs/gateway-resource-safety/spec.md) |

[proposal](proposal.md) / [design](design.md) / [source-baseline](source-baseline.md) / [source-feature-map](source-feature-map.md) / [tasks](tasks.md) / [verification](verification.md)。

[完整评估与patch sites](../../../docs/upstream-sync/PORTING-0.2.20.md) / [评估证据](../../../docs/upstream-sync/evidence-0.2.20/MANIFEST.md)。SQLite、miniredis、simple mode、Fast账号目录/费率/Free Fast、成功usage去重和固定quota窗口必须保留。没有新支付、部署、发布、push或上游版本号替换计划。
