# 整合与交叉审查记录

实施分工：backend_features负责F04/F05/F06后端和配置；frontend_features负责F01/F02/F03及F04/F05前端；主agent负责独立handler集成测试、组合门禁、范围复核与验收文档。两名实现agent各自对另一侧作只读审查，不并发修改对方文件。

## 已处理的行为问题

### 搜索使用量异步任务读取复用Gin

主agent的新真实Messages handler测试先阻塞使用量worker，待handler返回后把Gin.Request替换为下一请求，再释放worker。JSON和SSE两条路径均稳定把RequestedModel写成next-request-model；直接同步路径正常。

根因是ChannelUsageFields在worker闭包内读取Gin。后端在入队前快照该字段，继续沿用原用户/Key/账号/分组和零模型Token口径。证据：checks/handler-search-red.log → checks/handler-search-green.log；断言未放宽。该测试同时覆盖成功一次记录、LLM零调用、提供商错误、取消、SSE写错、代理换号和用户/账号槽位一次释放。

### 名称查询手动刷新与返回上下文

前端先后证明并修复两处时序：手动刷新复用普通pending导致仍拿旧快照，以及关闭进入错误详情后resumeState返回没有续接名称查询。保留主明细、分页和筛选上下文，使用请求序号及缓存generation防止迟到结果覆盖。

证据：frontend/red-f03-pending-refresh.log、red-f03-resume.log及其后f03-final.log。

### 独立交叉审查的两项补充

后端agent只读检查真实前端模块后报告：

1. OpenAI API Key新建表单选择enabled/disabled后直接切Anthropic，旧条件watch没有重置三态，旧值会写入新平台账号。
2. 之前已经存在forced名称请求时再次手动刷新，布尔force仍会复用旧forced快照。应区分用户显式刷新入口与普通/同轮缺项查询的合并。

前端agent用双向平台切换和组件级连续手动刷新分别建立失败证据后修正；原始诊断与最终结果见frontend的red-f05-platform-switch.log、red-f03-repeated-manual-refresh.log和cross-review-final.log（最终门禁以verification列出的最新日志为准）。

其余只读检查确认：无合法ID回退、翻页/筛选/关闭/unmount/会话切换的旧结果拦截、失败不清空主明细、有限分页、scope旧值及顺序回显、编辑default删除覆盖值、渠道布尔开关的保存均沿规格边界实现。

### SSE终态内部取消与跨平台凭据标记

前端agent反向只读审查后端，用独立overlay证明在message_delta的Write返回成功时取消请求，旧helper仍写message_stop并返回成功。后端把ctx传到每个私有事件写出边界，在Write前后和Flush后检查，并补JSON写入内取消的返回检查。新增测试遍历10个SSE事件×Write/Flush共20种时序；保留原overlay并复验通过，记录于frontend/backend-review.md与backend的red/green-event-cancellation及race-event-cancellation日志。

主agent最后对照F04明确凭据场景，发现旧判断把token_revoked/token_invalidated/detail Unauthorized限制在OpenAI。通用错误路径对其他现有平台可达，后端以跨平台正反例证明后移除平台限制；仅识别这些已有明确标记，未猜测模糊message或引入平台。普通模型401仍允许显式model，凭据失败保留每个平台既有认证策略。证据：backend/red-cross-platform-credential-markers.txt与green-cross-platform-and-cancellation.txt；最终完整组合门禁在该修正后执行。

## 测试夹具诊断

- checks/handler-search-first.log最初还包含新测试对UsageLog.UserAgent指针字段断言的类型错误；修正测试后，handler-search-red.log只剩真实的Gin复用失败。
- checks/handler-search-race.log报告旧的无锁settings测试桩在清理Set与调度异步GetValue之间读写同一map。主agent只在新handler测试文件内改用带RWMutex的独立仓储，未修改生产设置或放宽race检查。handler-search-race-final.log通过。
- 后端targeted-initial.txt保留接口适配中间态的编译诊断，不把它当作最终测试结果；所有红/诊断原始日志保留。

## 范围

未合并整版上游，不加入运维TPS百分位、缺失legacy入口、新平台/模型、性能缓存或支付；没有依赖/VERSION/迁移变更。后续文档维护只更新当前实施状态和证据，不修改冻结source-baseline或旧评估数据。
