# Verification（实施验收空槽）

本批尚未实施。评估基线PASS、预期失败探针与独立来源PoC均不是实施验收，下表保持空白。真正实施后记录真实起点和证据，不回写source-baseline。

| 项目 | 真实值/命令 | 结果/证据 |
|---|---|---|
| 实施起点 HEAD/tree、VERSION/Go/SQLite | | |
| 最终产品SHA与变化范围 | | |
| backend unit/embed/lint/codegen | | |
| 相关race | | |
| 实际HTTP/WS/SQLite/miniredis | | |
| quota/settlement/dedup/成本 | | |
| frontend测试/typecheck/lint/build | | |
| 桌面/移动浏览器（适用时） | | |
| 性能/内存及局限 | | |
| 冻结来源/历史哈希与范围 | | |

| 簇 | 实现证据 | 控制及验收结果 |
|---|---|---|
| F01 Codex 文献搜索 403 的账号副作用 | | |
| F02 OpenAI 408 上游超时换号 | | |
| F03 Codex IDE 与独立上下文 | | |
| F04 Responses 每轮请求推理强度 | | |
| F05 既有 Claude 账号测试 temperature | | |
| F06 兼容转换保留 stop 与原始工具定义 | | |
| F07 图片 SSE usage 单行载荷 | | |
| F08 Gemini 最后工具轮的尾随内容 | | |
| F09 Antigravity 原生 SSE 注释帧 | | |
| F10 Grok 媒体输入错误的 400 分类 | | |
