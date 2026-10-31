# Design

## Context

现有转换器、Chat fallback、Grok client tools、compact和HTTP/WS桥均有fork适配；四个旧尾项与这些消费者相交。独立index的CLEAN/CONFLICT只决定适配方式，当前服务/序列化探针证明尚存缺口。

## Decisions

1. C01两条配套终态一起适配。普通arguments只按有效JSON判定，沿用空值默认；不能把finish_reason=length直接当工具参数无效，也不能把custom自由文本当普通arguments。坏历史调用与匹配输出配套处理，保留其它对话内容。
2. C02从请求声明收集显式function与namespace归属，传入实际buffer/stream消费者。只有无歧义custom别名才还原，显式声明与真实namespace归属优先；added/done/最终output一致。
3. C03在最终tools列表上判定内建tool_search是否还存在。普通function同名不算前置；仅修改请求所拥有的副本，遵守现有adapter对请求的修改合同，不改共享账号/catalog数据。
4. C04事件序列化保留0。compact本地事件从0递增，已知前序事件后的合成终止从实际writer/state继续；上游非零值不改写。仅有heartbeat的流没有数字事件，首个终止事件可从0开始。不要照搬“所有失败统一0”覆盖已知序号。
5. 无效流工具参数须阻止成功工具终态；错误结果保留已观测usage，沿现有handler与异步写入政策处理，验证取消、队列边界和SQLite dedup。协议修复不改变成本/配额规则。
6. 四项共同依赖转换文件，按行为顺序适配后一起回归。所有新增helper随消费者落地；不为来源旧测试补完整不存在系统。

## Risks and Verification

主要风险是误拒绝合法工具、别名重写碰撞、改变流序号和丢失部分usage。需正反例、实际出口、race与SQLite账本检查；严格客户端完整连接验收在实施阶段进行。

VERSION、工具链、依赖、schema及生成Ent/Wire保持当前范围；若实施发现必须扩围，先重新评估真实前置，不能携带未选第三档能力。
