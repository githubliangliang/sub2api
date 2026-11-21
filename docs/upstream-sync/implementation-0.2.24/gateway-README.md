# v0.2.24 gateway 簇实施记录

真实起点 `d2488b1105f091b6fce4445395fec1835306fcf8`，日期 2026-11-21。此文件只说明本子任务 F01/F02/F05/F07/F08/F09/F10/S04/S15；两档总体验收由根任务记录。冻结 intake/source 不回写。

| 簇 | 最小实现与验证 |
|---|---|
| F01 | alpha/search 两个 OAuth/PAT 调用点传入响应体；403 只对明确结构化 account/workspace/org 停用 code 执行既有副作用。文献拒绝/空403不写账号，明确 workspace 停用写一次；401/404/405/429相邻控制保留。普通 Responses 鉴权策略不变。intake 称已有的 `isOpenAIHTTPUpstreamAccessStateError` 实际仅在注释出现，因此把冻结来源的 code 判断局部适配到本 predicate。 |
| F02 | 统一 failover 分类增加 HTTP408，保留现有上下文超长优先拒绝与400控制。图片408验证在写下游之前返回 typed failover；原有换号预算和已输出保护复用。 |
| F05 | Claude 公共 sampling predicate 只覆盖既有 Opus4.7/4.8/5.5家族及现有别名/Bedrock前缀；普通和 Bedrock 测试载荷按型号省略 temperature。未引入 Fable/Mythos 新分支。builder、映射别名、序列化请求及真实本地HTTP抓取有控制。 |
| F07 | 图片 completed builder 对合法 usage 用 json.Compact 后 SetRawBytes。多行/紧凑/空/非法usage、精确token值、图片结果和真实本地TCP单事件解析通过。 |
| F08 | 只对最终 Gemini turn，把最后 functionResponse 后的尾随part保序拆入后续user turn。单/多工具、文本/图片、纯工具/纯文本、非末尾历史控制及既有 parallel工具媒体顺序验证通过；更新两个旧断言以体现新轮次。 |
| F09 | 支持注释的客户端保留注释帧及空行；go/python-genai同时滤上游注释和本地注释心跳。重写data已含分隔时滤重复空行，原样空data/[DONE]仍等待其分隔。LF/CRLF、keepalive、SDK和断连后usage drain通过。 |
| F10 | 两个现有超图数检查返回 typed input error，handler先返回400且保留信息；最大图片数仍为3。同步/异步、OAuth/API Key、JSON/multipart、3图可转发/4图无上游调用、slot释放及既有上游脱敏错误控制通过。 |
| S04 | 保留 raw `server_side_tool_usage_details`，严格读取 web_search_calls/x_search_calls/document_search_calls；明确0覆盖item计数，缺失/非法值沿用fallback，整数溢出亦拒绝。终态成功数锁定后不被晚item事件增加。普通同名function/custom工具及failed item不计服务端搜索。Responses/Chat/Messages、stream/buffered六种路径与原有dedup通过。 |
| S15 | request gin context声明最多一个裸429冷却账号；retry和Claude/Gemini/compat外层共同使用，独立请求与明确details/reset保持原策略。quota fetch使用既有账号转发端点，新增最小单端点client入口并保留原有有界读取、redirect/client/proxy/cancel路径。付费daily/免费prod/环境覆盖实际本地HTTP读数、daily失败不回落prod、原client fallback与取消控制通过；没有引入QuotaSummary。 |

最终定向与同范围 race 均退出0：四包64个顶层测试、193个含子例测试通过。完整命令、离线工具链环境、分包结果和所测文件采样哈希见 [gateway-summary.json](gateway-summary.json)，日志分别为 [gateway-final-targeted.log](gateway-final-targeted.log) / [gateway-final-race.log](gateway-final-race.log)。所有凭据均为合成测试值，本地HTTP仅访问测试服务器。

失败证据分别保留：

- [gateway-before.log](gateway-before.log)：F05/F07/F08/F09四个目标顶层失败；控制分支通过，随后最小修复。
- [gateway-usage-quota-before.log](gateway-usage-quota-before.log)：S04 JSON/SSE/六路径、S15裸429与端点共5个目标顶层失败，另1控制顶层通过。
- [gateway-second-before.log](gateway-second-before.log)：handler的F10目标完整复现为502，3图控制通过；同次service遇到同伴测试夹具的中间编译状态，不能混称完整产品失败。
- [gateway-toolchain-offline-error.log](gateway-toolchain-offline-error.log)：默认go launcher在禁用sumdb时验证工具链失败，改用已缓存Go1.27.4绝对路径；没有开启网络下载。
- [gateway-selected-after.log](gateway-selected-after.log) / [gateway-second-build-in-progress.log](gateway-second-build-in-progress.log)：共享工作区内同伴helper尚未写完的编译失败，不计本簇产品缺陷。首次quota测试构造器也需按本地旧签名适配，最终编译与测试均通过。

这些是共享工作区内局部实现与协议验证，不代表独立提交、全量门禁或1CPU/1GiB实例验收。本子任务没有更改版本、依赖、迁移或资源默认，没有提交、推送、发布、归档或业务实例操作。总任务还须统一记录结算/SQLite/fixed quota及完整门禁；本记录不据此勾选全部组合验收。
