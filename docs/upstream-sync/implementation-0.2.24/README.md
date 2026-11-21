# v0.2.24 两档实现与验收

真实实施起点：`d2488b1105f091b6fce4445395fec1835306fcf8`，tree `17b1b71f61137510e42feeaa3fcf2a1109cb4ce6`。用户于2026-11-21选定两份OpenSpec实现。开始时已有intake文档/两份未跟踪提案；原状态见[initial-state.json](initial-state.json)，原任务/README/verification已分别保存。验收时实现保留在工作区，无新提交；用户后续授权本地提交，实现与相关证据随本次提交保存。

第一档F01–F10、第二档S01–S15均已实现并完成组合验收，37/37与52/52任务已勾选。逐簇最终状态以[第一档verification](../../../openspec/changes/port-upstream-0.2.24-tier1/verification.md)、[第二档verification](../../../openspec/changes/port-upstream-0.2.24-tier2/verification.md)和本目录最终核验为准。

## 实现范围

- [协议及热点说明](protocol-implementation.md)：F03/F04/F06、S03/S08/S12/S13；保留本地更完整的HTTP/WS requested effort、output_config、SQLite230读写和既有协议路径。
- [网关说明](gateway-README.md)：F01/F02/F05/F07/F08/F09/F10、S04/S15。Grok保持3图；成功搜索次数使用真实`server_side_tool_usage_details`字段且严格区分0/缺失/非法值；裸429声明只限同一请求。
- 根任务实现S01/S02/S05/S06/S07/S09/S10/S11/S14。新增最小requestmodel扫描模块；挂HTTP各链和WS每个JSON帧，先于模型改写/调度/计费。预读保留较窄文本限制与原body存储。
- OpenAI已完成轮换使用10秒有界脱离取消上下文做SQLite完整凭据/代理CAS与outbox事务。读取持久化账号后再返回；并发重新授权胜出；非token/代理编辑只合并本次变化，最多3次保存，不重试旧token。锁租期预留保存、快照同步、缓存清理与释放时间。
- OAuth固定Codex出口按实际映射目标排除明确外族模型；API Key自定义base_url/合法映射/未知别名和既有可选透传源模型筛选保留。
- 批量透传与模型过滤拥有独立修改开关，省略与显式false分开。Claude/Gemini快照key参与组件和列表缓存/异步结果归属，旧响应不覆盖新快照，刷新失败保留已知读数。
- Anthropic长上下文展示依据有效映射目标metadata，阈值900000；遵从design的保守交集：可路由候选中有短窗口或未知metadata即不自动追加[1m]。已有显式[1m]/大小写变体保留且不重复新增。与来源“任一候选即可”不同，本地适配及mixed/routed/mapped-target控制在`anthropic_long_context_models_test.go`。
- Composite改写复用本地Sonnet/Opus5.5家族判定适配thinking。后台创建普通用户复用注册的初始默认quota快照，显式0保留、管理员跳过、写失败fail-open；真实Wire生成接入已有service adapter。
- 模型平均TPS只改SQLite有效output总和/有效duration总和；整窗口输出TPS的逐请求算术mean/P5/P10/P50保持独立。

## 测试与证据归因

所有实际HTTP/TCP/WS都访问临时本地测试端点；SQLite为临时数据库并使用本地迁移/Ent/仓储；miniredis为独立临时服务；凭据与用量均为合成夹具。没有业务实例、供应商在线模型或真实OAuth操作。

- [root-before.json](root-before.json)：真实起点上模型误选、轮换结果丢失、TPS统计与Composite thinking复现；其余before证据分别见两份子说明。
- [root-security-targeted-adapted.json](root-security-targeted-adapted.json)、[root-security-budget.json](root-security-budget.json)、[root-security-binary.json](root-security-binary.json)：scanner、协议错误形状、真实HTTP各别名、首/后续/二进制WS、session.update、无上游调用和usage边界。
- [root-quota-http-adapted.json](root-quota-http-adapted.json)、[root-ops-http.json](root-ops-http.json)、[reviewer-controls.json](reviewer-controls.json)、[final-review-tests.json](final-review-tests.json)：实际SQLite/HTTP默认quota、weighted TPS、mapped metadata/保守交集、OpenAI proxy/non-token/CAS冲突与真实SQLite/miniredis脱离取消保存。
- [frontend-late-controls.json](frontend-late-controls.json)：同ID组件及批量请求逆序完成控制；目标变更、离页和旧账号请求控制沿用既有生命周期测试。
- [browser-selection-adapted.json](browser-selection-adapted.json) / [浏览器明细](browser/results.json)：实际构建应用与Chrome，1440桌面/390移动6操作、144个合成API请求、无pageerror；批量只改透传/独立false/重开，以及Claude/Gemini同ID快照自动刷新。API为浏览器拦截合成响应，不能当真实数据库保存验收；真实后端证据独立。
- [resource-1c1g.json](resource-1c1g.json)：隔离`--network none`、1CPU/1GiB、GOMAXPROCS=1，64路2MiB JSON实际loopback HTTP，40成功/24过载，预算最终完整释放；memory.peak=185131008字节。独立middleware测试二进制含真实scanner/预读/限制，不是完整生产实例容量结论。
- [scanner-fuzz.json](scanner-fuzz.json)：17313次key等价模糊测试；[协议热点短fuzz](protocol-hotpath-fuzz.log)与逐hunk单P/alloc差分另列子说明。工作站函数成本、合成40组scheduler投影、独立1C1G夹具各自独立，不能互相替代生产验收。

## 过程失败与边界

首次全量运行`backend-unit-first`读取了并行工作中的新增预修复目标，且旧stop/重复prev-id夹具需适配新行为。其完整日志保持；不是既有基线失败，更不是最终结果。`backend-unit-final`随后55有测试包/10773顶层/22269含子例通过，16既有测试跳过；独立审查随后收紧Validated前置、目录交集并补控制，最终接受[backend-unit-accepted](backend-unit-accepted.json)55测试包/10776顶层/22277含子例通过、16原skip；[race-final](race-final.json)9包283顶层/819含子例通过；[lint](backend-lint-final.json)0问题、[embed](embed-build-accepted.json)/[Wire](codegen-wire.json)通过。[前端最终全量](frontend-full-accepted.json)2254通过/2原skip、定向112项/typecheck/lint/build通过。

早期root安全组合遇到同伴的metadata resolver中间编译状态；quota HTTP夹具一度错传repository DTO而非已有service adapter；前端view新增测试改变顺序后需恢复settings mock默认响应；browser-first选择器仅匹配组件未使用的属性，后改可见checkbox。默认lint发现新测试引用unit-only helper、未检查Close/断言和未用差分helper；均保留原日志并最小适配。不得把这些夹具/标签/工作中状态算成产品起点缺陷。

独立审查确认Validated解析只能用于同一份已验证字节；两个HTTP handler若effort policy改写body则改走checked parser。对嵌套父字段重复缺子字段的gjson读取曾提出假设，实际离线复现与差分确认旧gjson会继续查下一同名父字段，故保留等价行为；入口S01仍拒绝这些重复请求。

VERSION1.1.26、Go1.27.4、SQLite230/auth21、依赖、Ent及64MiB body/40MiB SSE/8MiB staging/120s idle保持。Wire只为默认quota注入实际重生成。11份旧活动change、350已完成任务及冻结intake/source保留。新增搜索/ToolSearch/compaction/cache-estimate/UI、5图和支付不在本次实现范围；用户后续授权本地提交；无push/release/archive或业务实例操作。

最终范围、声明与保护核验见[final-validation.json](final-validation.json)，最终产品/测试/生成文件哈希见[final-file-hashes.json](final-file-hashes.json)。OpenSpec两档严格校验通过；首次命令路径抄写少字符的MODULE_NOT_FOUND保留在specification-tier1日志，按冻结命令JSON读取精确已安装路径后离线校验通过。

本地提交准备核对见[commit-readiness.json](commit-readiness.json)。final-validation、initial-state及各次测试日志记录验收时点的HEAD/工作区状态，作为冻结证据保留；本次提交不回写这些历史值。
