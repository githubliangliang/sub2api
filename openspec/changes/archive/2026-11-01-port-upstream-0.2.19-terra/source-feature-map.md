# Source Feature Map

| 行为 | 来源 | 接受范围 | 验收 |
|---|---|---|---|
| T01 模型身份/effort | #8364整体PR | base及允许的none/low/medium/high/xhigh/max/compact规范，默认medium；Sol规则保持 | 供应商前缀/大小写/下划线、每个effort、未知pro/ultra/minimal拒绝、显式thinking禁用优先 |
| T01 工具协议 | #8364整体PR | Responses可用时既有转换承载工具，only-Chat账号在出站前拒绝工具 | Chat/Responses/Messages的tool与legacy functions、无工具普通文本、HTTP/SSE及既有WS桥 |
| T01 目录/metadata | #8364整体PR | 公共默认描述添加Terra；账号明确metadata覆盖默认，不就地改写上游catalog | 自定义text-only/high/100k声明、空上游列表保持空、raw目录与ETag/缓存保持 |
| T01 usage/price | #8364整体PR | 只摘Terra价卡与既有policy接入，显式0/覆盖优先 | 资源与两层fallback、所有token桶、priority、271999/272000/272001、handler/SQLite dedup |
| T01 现有客户端配置 | #8364整体PR | 白名单、映射预设及已有OpenCode客户端模板包含Terra | 生成配置解析/round-trip、context/output/六档effort、旧默认与其它模型回归 |

源新增Terra测试是待适配/移植的用例，不是baseline现存用例。当前三个后端包及UseKeyModal21项邻近回归通过，见[评估证据](../../../docs/upstream-sync/evidence-0.2.19/MANIFEST.md)；它们不能填写实施verification。

上游基础input/cache-read/cache-write/output每百万为1/0.05/1.25/5，priority2倍；长上下文阈值272000/input2/output1.5，默认context1050000/output128000。这些是冻结来源值，不是实时官方售价或任意账号权限证明；public created占位值不作为发布日期。

其它新增能力及支付保持排除/按需。当前用量成本统计不是充值/购买功能，保留成功usage与配额数据完整性。

