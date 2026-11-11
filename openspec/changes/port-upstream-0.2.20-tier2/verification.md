# Verification

状态：2026-11-10用户授权后，**S01–S11及38/38任务完成**。行为及验收见下表，最终门禁由[scope-check.json](../../../docs/upstream-sync/implementation-0.2.20-tier2/scope-check.json)汇总；**实现于2026-11-11随本次提交保存，尚未归档**。没有发布、push或业务部署。

实际起点：`29681fdf5968cb3482efa35694bc27cd26ecb802`（第一档已提交），工作区干净。VERSION1.1.23 / Go1.27.3 / SQLite228；运行时auth snapshot从20升至21。Ent、依赖、版本和schema不变；Wire由`go generate ./cmd/server`重新生成，以关闭新增有界共享digest缓存。

实现提交：包含本文件本次状态更新的提交（`git log -1 -- openspec/changes/port-upstream-0.2.20-tier2/verification.md`可定位）。文件哈希与diff见[product-hashes.json](../../../docs/upstream-sync/implementation-0.2.20-tier2/product-hashes.json) / [product.diff](../../../docs/upstream-sync/implementation-0.2.20-tier2/product.diff)。原始source-baseline/proposal/design/source-feature-map/spec和所有历史评估/验收保持原内容。

## 行为与回归

| 簇 | 缺口依据 / 修复前证据 | 实际实现与验收 |
|---|---|---|
| S01 Responses非流式及instructions回显 | s01-s03-before.log的5个非流式失败及空instructions拒绝，来源PR7979/8499；last-terminal-before.log追加6种尾事件反例。 | 普通/透传stream=false、无Content-Type、incomplete/cancelled JSON与instructions缺失/null/空/多行/不变回显；保留客户端原选择，末尾失败先走现有failover，最终终态优先。last-terminal-final.log通过race。 |
| S02 Grok委派消息 | 基线TestTier2GrokAgentMessage保留agent_message而失败。 | 仅替换委派RawMessage，复用AgentMessageText；空委派跳过，其它item字节/大整数保持。真实HTTP上游检查不再含agent_message，随后Grok模型审计/排队SQLite结算/重复请求去重通过。 |
| S03 Terra目录及计费后续 | 基线minimal不拒绝、订阅无豁免，两探针失败；冻结descriptor与本地旧目录差异已量测。 | Codex low、500000/872000及六档目录；API保留medium/none，minimal400。账号raw/metadata/显式空目录优先和ETag保持。271999/272000/272001×OAuth/setup-token/API Key×default/priority成本测试、Fast/Free Fast费用与SQLite去重通过。 |
| S04 入站预算、超时和错误分类 | 原来只有单请求字节上限，无实例预算/读写空闲期限；body_budget、timeout来源回归与实际入口追加测试覆盖新契约。 | gateway实例64MiB预算；64MiB解压上限，读取/解压/lenient/重写作用域、100轮attempt释放及异步图片所有权移交。超预算503+Retry-After，超时408/断开499/截断与解码400；日志仅计数与类别。HTTP1/2超时、长响应不误限、取消及真实HTTP503恢复通过。 |
| S05 流、drain与缓存边界 | 原Gemini ReadString/Antigravity ReadAll、队列按条数和会话缓存无总字节限制；新增资源超限回归实际证明未完整接入的本地路径，修正后通过。 | 既有Anthropic/OpenAI/Grok/Antigravity/Gemini流有行/事件/队列边界，图片未完成事件不输出假完成。transport drain最长180秒；捕获用量仍交usage结算，不虚构完整用量。缓存32MiB/16384条上限，TTL/LRU/键隔离/并发关闭；配额写入在usage worker内最多5秒，Wire真实关闭共享digest缓存。 |
| S06 Antigravity身份规范化 | 原前导只识别Claude，compat转换前无中和；来源身份/任务保留回归和本地原生对照纳入测试。 | 限定系统开头身份与attribution；Codex/Claude、紧邻逗号任务、引用/string/parts及已有Antigravity身份测试通过；不改变原生Anthropic。 |
| S07 图片500模型冷却 | 原图片重试耗尽进入整账号处罚；来源模型作用域测试只取独立子集，不采环境可调处罚。 | 最终映射图片模型短冷却，仓储成功后同步缓存；文本仍走原处罚。图片不递增账号计数，未加入环境变量处罚开关。定向/race和现有模型限流/SQLite保护通过。 |
| S08 Gemini工具媒体 | 原工具媒体被搬到独立user内容，来源functionResponse.parts及既有媒体保真测试形成回归。 | 按最终映射模型传递能力：Gemini3+工具媒体保留functionResponse.parts，旧模型保留原回退；更新全部本地转换调用/测试，不新增模型或图片路由。 |
| S09 使用记录导出恢复 | 原导出100条且无429重试，前端来源回归与旧100条断言明确显示旧契约；新测试覆盖等待和生命周期。 | 1000条分页、数字Retry-After、1–60秒等待、最多3次重试；AbortSignal中止请求/等待，保持初始筛选，卸载/换号不下载部分CSV。真实浏览器2请求成功下载/4请求耗尽/离页取消1请求/移动端通过；页面API为合成拦截。 |
| S10 Antigravity测试提示词 | AccountTestService原prompt在路由链丢弃，底层点号+1token；新增两种模型builder及HTTP回归。 | 默认hi、保留Unicode/多行输入，256输出预算；保留旧TestConnection入口给定时消费者。实际本地HTTP捕获prompt和token预算，并返回真实fixture回复；没有调用供应商账号。 |
| S11 RPM空值缓存与JSON预检 | 旧nil覆盖每请求查SQLite；原解码路径作为预检等价性和benchmark oracle。 | 鉴权snapshot20→21，记录已加载group，nil/0/正值/错误/跨组正确；SQLite+miniredis20次请求不重查，换group重新查询。thinking/委派预检和原解码逐字节等价，重复键/转义/非法JSON测试及7270次fuzz通过；两项benchmark分配均降为0，不引入ojson/COW整套。 |

各命令原文、cwd、耗时、UTC时间、退出码与同名日志在[实施目录](../../../docs/upstream-sync/implementation-0.2.20-tier2/README.md)。来源测试发生的编译前置错误不计产品缺陷；尚未存在的新预算/缓存API由源码缺口和行为回归建立，不声称它们在旧代码直接编译执行。

## 最终门禁

| 项目 | 证据 | 范围 |
|---|---|---|
| 全量后端unit | acceptance-unit.log/json | backend所有包；代理变量只从测试子进程移除 |
| embed构建 / lint | acceptance-build、acceptance-lint | 本地临时二进制，未部署；lint需0 issues |
| 定向race | backend-race、final-additions-race、cache-final-race、last-terminal-final | service/handler/repository/httputil/antigravity/apicompat/middleware；server/routes参与编译，原大pattern无匹配用例，不冒充其race覆盖 |
| SQLite/配额/去重 | sqlite-protection、settlement-current、final-boundaries | 实际迁移、固定窗口、Fast真实结算、Grok审计及RPM查询次数 |
| 前端 | frontend-typecheck、frontend-lint-final、frontend-full-final、frontend-critical、frontend-build | 2186通过/2既有跳过；critical179通过/2；构建有既有chunk/Browserslist警告 |
| 真实浏览器 | browser/result.json、desktop.png、mobile.png、usage.csv | 构建产物+Chromium；成功重试下载/耗尽/离页取消/移动端；unit另验换号取消 |
| 1GiB/1CPU | memory-1g-stress.log | 无网络/只读/内存与swap合计1GiB容器，32×4MiB并发HTTP；预算64MiB，峰值175411200字节，波次释放used=0且下一请求成功 |
| 性能等价与测量 | thinking-benchmark、delegation-benchmark、thinking-fuzz | 既有decoder对照；fuzz7270执行，未发现差异 |
| Wire关闭流程 | wire-generate-final、server-cleanup-final | 源文件增加DigestSessionStore cleanup；真实生成，不手改wire_gen |
| OpenSpec与范围 | openspec-tier2、scope-check | 38项任务、冻结来源、代码哈希、版本/依赖/schema和链接核对 |

## 适配决策与边界

- 评估曾将异步图片入口视为缺失，实施核对发现本地已有`AsyncImageHandler`，所以必须实现预算移交；这是已存在消费者的必需修复，未引入新图片能力。`TestAsyncImageBodyBudgetSurvivesIngressRelease`验证父请求释放后仍持有预算，任务结束才归还。
- 本地SSE/时间戳/失败处理比来源拆分不同，按实际函数接入。S01追加“completed之后出现失败/取消/不完整”的6场景失败复现，最终末尾终态优先，保留已完成第一档与既有失败切号。
- 源码实际配置是`gateway.*`。选择64MiB预算、40MiB默认SSE行、8MiB队列、4096事件行；0预算/读写timeout允许禁用，显式较大行上限仍可配置。与旧默认500MiB不同，来源默认断言已按本fork决策调整，并保留64MiB显式配置正例。
- 预算覆盖缓冲入站/解压/归一化及原生Responses补丁作用域，**不是进程总堆上限或所有协议转换分配的精确计量器**。1GiB测量是该层的本地HTTP压力，不是完整业务实例/供应商吞吐基准；模型上下文、数据库缓存、响应和GC仍需要内存。
- S11只取nil RPM标记和两个既有预检；未引入ojson、完整COW/新compact算法或更广隐私快照体系。positive覆盖跨组同样回源；snapshot21保留Fast字段。
- S05配额持久化保留SQLite配置限额/0值/软删不复活及固定窗口契约，不改变计费去重。失败有计数/日志但没有新增持久重试队列。上游email/支付通知整套worker不在本次选择内，原支付与商业通知代码未扩展。
- Wire重新生成只为共享缓存关闭；其9行生成差异和对应源/测试更新单独可查。缓存机会式过期清理不改timer，避免与已触发回调重复调度。

## 测量结果（本机合成数据）

- thinking有效签名长历史：原解码1.46–1.54ms/op、约905KB、10629allocs；预检0.616–0.631ms/op、0B、0alloc。
- 普通委派候选长输入：原唯一键扫描+解码3.43–3.78ms/op、约3.00MB、83allocs；预检18.5–20.9µs/op、0B、0alloc。
- 入站压力波次14个成功/18个503，另1个恢复请求成功（日志accepted=15含恢复），117ms为波次时间附近的本地测量，不当作真实模型吞吐。容器峰值约167.3MiB。

## 过程失败记录

保留compile-*、targeted-*、backend-unit-1/2/final/complete和早期lint日志。失败原因包括：来源测试引用本地缺失接口；新增函数签名的邻近测试未同步；413/499及默认500→40MiB旧断言；图片未完整事件提前写出；instructions原值登记遗漏；Wire cleanup测试参数更新；SQLite测试表`rate_multiplier`非空约束。均按失败原因修正，未隐藏失败日志或把编译错误计为修复前产品复现。

本地HTTP上游、浏览器API均为合成fixture；SQLite/miniredis/HTTP服务器/Chromium真实执行。没有真实供应商账号请求、生产数据修改或业务实例部署。
