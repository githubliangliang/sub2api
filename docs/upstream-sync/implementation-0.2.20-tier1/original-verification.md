# Verification

状态：未实施。评估测试在[证据目录](../../../docs/upstream-sync/evidence-0.2.20/MANIFEST.md)，不填写为本批通过。

实际实施起点：

实现提交：

| 簇 | 修复前证据 | 实现与回归证据 | 结果 |
|---|---|---|---|
| F01 Claude setup-token 周用量 | | | |
| F02 Codex 强制模式停止多余探测 | | | |
| F03 WebSocket 每轮首 token 时间 | | | |
| F04 Codex WS→HTTP 桥接指纹 | | | |
| F05 缓冲 Responses reasoning 保真 | | | |
| F06 Responses 完成消息字段 | | | |
| F07 鉴权取消分类 | | | |
| F08 413 透传规则优先级 | | | |
| F09 Antigravity 直连 SSE 响应 | | | |
| F10 Antigravity JSON Schema 参数 | | | |
| F11 GPT Image 账号测试参数 | | | |
| F12 Anthropic兼容 Cloudflare REST URL | | | |
| F13 Bedrock EventStream 分配前边界 | | | |

| 门禁 | 命令/日志 | 结果 |
|---|---|---|
| backend unit / embed build / lint | | |
| 定向race | | |
| HTTP / SQLite228 / quotas / dedup / Fast | | |
| 前端与真实浏览器（如适用） | | |
| 1C1G资源/分配测量（如适用） | | |
| 来源、范围、OpenSpec | | |

残余风险及未完成项：
