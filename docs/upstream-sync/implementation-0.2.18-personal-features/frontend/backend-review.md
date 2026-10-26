# 后端独立只读交叉审查

审查范围：F05搜索模拟/管理器、Messages handler归属快照、JSON/SSE取消写错及F06配置/实际转换消费者。未修改backend产品或测试文件；唯一执行的额外Go命令使用overlay加入临时测试。

## 已证明的取消终态遗漏

触发：搜索SSE的message_delta写出期间请求context取消，但底层Write暂时仍返回成功。writeSSEMessageEnd在一个helper内写message_delta和message_stop，外层只在进入helper前检查context，导致取消后仍发message_stop并返回非nil成功result。

初次复现的实际输出：`canceled=context canceled err=<nil> result_nil=false stop_after_cancel=true`。原始日志backend-review-cancel-overlay.log保持该失败结果，已交后端实现agent修复；最终修正证据由主verification/后端最终日志记录，不能用本红色日志声称修复完成。

持久化复现材料：backend-review-cancel-test.go.txt为原测试源码，backend-review-cancel-overlay.json为原执行映射。仓库产品文件不被替换；virtual新文件只存在于该命令的Go overlay中。

复跑（仓库根目录准备，Go命令在backend目录）：

```bash
rtk proxy cp docs/upstream-sync/implementation-0.2.18-personal-features/frontend/backend-review-cancel-test.go.txt /tmp/sub2api-frontend-review-cancel-test.go
rtk proxy cp docs/upstream-sync/implementation-0.2.18-personal-features/frontend/backend-review-cancel-overlay.json /tmp/sub2api-frontend-review-overlay.json
rtk proxy go test -overlay /tmp/sub2api-frontend-review-overlay.json -tags=unit ./internal/service -run TestFrontendReview_SearchCanceledDuringEnd -count=1
```

## 已核对且未发现实质遗漏的部分

- SetWebSearchManager保留生产server/http.go对真实*websearch.Manager的装配；getWebSearchManager仍返回具体manager，TestSearch、GetUsage、ResetUsage管理路径未因搜索执行接口抽窄而消失。typed-nil真实manager仍视作不可用。
- OpenAI ForwardAsAnthropic显式限platform=openai；共享判断排除非API Key与影子账号，未把Grok/其他平台新接入模拟。
- 模拟在早期Chat分流前调用；全局开关、账号三态/旧bool、渠道openai继承均有对应调用链。错误经原failover/响应已写出检查处理，JSON路径现在检查真实Write错误。
- handler在队列提交前计算channelUsageFields，值结构只含整数/字符串，无Gin上下文或可变map引用；其他客户端IP、UA、endpoint、session等也在提交前提取。
- min_output_tokens在Load的弱解码前用原值校验，因此显式0/小数/溢出拒绝；程序内零值配置在服务helper安全回落128，Validate允许零值以兼容现有构造。
- 生成输出的既有三种转换消费者均改为实例配置。openai_gateway_count_tokens.go中仍有两处旧包装器，但结果只取Model/Instructions/Input/Tools/ToolChoice，输出上限字段不进入计数请求，属于保留计数语义，不是漏接生成请求。
- 原生Responses/compact/WS规范化未新增legacy入口或可变全局floor。

以上结论来自当前调用链与现有测试结构复核；未重复全量Go门禁，也不作真实外部搜索服务兼容性声明。
