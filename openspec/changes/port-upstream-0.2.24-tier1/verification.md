# Verification（实现与组合验收）

实现日期2026-11-21。用户选定本档与另一档，联合实施；真实起点为`d2488b1105f091b6fce4445395fec1835306fcf8` / tree `17b1b71f61137510e42feeaa3fcf2a1109cb4ce6`。开始时已有intake文档和未跟踪提案，保留原始状态/任务/README/空槽副本。冻结source-baseline/source-feature-map及intake不回写。验收时实现位于工作区，最终内容以逐文件SHA256确定；用户后续授权本地提交，实现随本次提交保存，等待归档。

[组合实施与局限](../../../docs/upstream-sync/implementation-0.2.24/README.md) · [原始起点](../../../docs/upstream-sync/implementation-0.2.24/initial-state.json) · [最终文件与范围](../../../docs/upstream-sync/implementation-0.2.24/final-validation.json)

| 项目 | 实际结果与证据 |
|---|---|
| 起点/最终内容 | 验收时HEAD保持上述起点；最终产品/测试/生成文件哈希见final-validation及final-file-hashes，不用基线SHA冒称已提交实现。 |
| backend unit | 全量55个有测试包通过；完整最终命令、顶层/含子例与跳过数见[backend-unit-accepted.json](../../../docs/upstream-sync/implementation-0.2.24/backend-unit-accepted.json)。另保留中间失败及审查前通过日志。 |
| embed/lint/Wire | [embed构建](../../../docs/upstream-sync/implementation-0.2.24/embed-build-accepted.json)、[lint 0](../../../docs/upstream-sync/implementation-0.2.24/backend-lint-final.json)、[实际Wire生成](../../../docs/upstream-sync/implementation-0.2.24/codegen-wire.json)。Ent、依赖、迁移未变。 |
| race | [9包283顶层/819含子例通过](../../../docs/upstream-sync/implementation-0.2.24/race-final.json)，覆盖CAS/取消/缓存/WS/审计/两档交叉路径；单独网关四包race亦通过。 |
| HTTP/WS/SQLite/miniredis | 新的各别名/首后帧/二进制/session.update、取消后轮换/并发CAS、后台用户quota、metadata目录与模型统计；[SQLite成本/固定quota/Fast/去重控制](../../../docs/upstream-sync/implementation-0.2.24/sqlite-cost-quota-controls.json)34含子例通过。所有数据为临时合成夹具，无业务实例/真实供应商操作。 |
| frontend | [全量测试](../../../docs/upstream-sync/implementation-0.2.24/frontend-full-accepted.json)、[定向112项](../../../docs/upstream-sync/implementation-0.2.24/frontend-targeted-accepted.json)、[typecheck](../../../docs/upstream-sync/implementation-0.2.24/frontend-typecheck-final.json)、[lint](../../../docs/upstream-sync/implementation-0.2.24/frontend-lint-final.json)、[build](../../../docs/upstream-sync/implementation-0.2.24/frontend-build.json)通过；2项既有skip。 |
| 桌面/移动浏览器 | 实际构建应用及Chrome1440/390，6操作/144合成API请求/0pageerror；[结果](../../../docs/upstream-sync/implementation-0.2.24/browser/results.json)。合成拦截API与后端SQLite实际保存证据分开。 |
| 资源与热点 | [1CPU/1GiB实际loopback64路2MiB](../../../docs/upstream-sync/implementation-0.2.24/resource-1c1g.json)，memory.peak185131008B、预算完整释放；独立middleware测试二进制。热点逐hunk单P/alloc及差分见协议说明；均不是生产实例容量/收益承诺。 |
| 历史与范围 | [冻结历史核验](../../../docs/upstream-sync/implementation-0.2.24/final-validation.json)；11份旧活动change/350已完成任务保留。VERSION1.1.26、Go1.27.4、SQLite230/auth21、依赖、Ent及64MiB/40MiB/8MiB/120s默认保持，Wire为新依赖注入生成。用户后续授权本地提交；无push/release/archive/业务部署。 |

| 簇 | 最终实现 | 回归/控制及证据 |
|---|---|---|
| F01 | 限定alpha/search工具级403；仅明确账号/工作区停用保留副作用 | OAuth/PAT、副作用计数及401/404/405/429控制；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/gateway-README.md)。 |
| F02 | 408进入已有换号分类，保留输出/取消/重试预算边界 | Responses与图片408、400/上下文超长及原换号邻域；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/gateway-README.md)。 |
| F03 | IDE/独立上下文保留类型空值规则，转换为不可信user文本 | 空串/null/数组/真实call_id/page快照/重复键控制；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/protocol-implementation.md)。 |
| F04 | 原始input最后有效configuration_update优先；requested/effective分开 | HTTP/桥接/原生WS原有采集、none/max/output_config及SQLite230控制；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/protocol-implementation.md)。 |
| F05 | 现有Opus家族连接测试省略temperature | 别名/Bedrock/普通Sonnet控制、实际本地HTTP抓取；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/gateway-README.md)。 |
| F06 | 桥接保留stop和现有API Key可转发动态工具 | 字符串/数组/null、原始数值/顺序、native/OAuth字段白名单；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/protocol-implementation.md)。 |
| F07 | 合法usage压成单行JSON再写completed SSE | 多行/紧凑/空/非法、实际TCP事件、精确token与图像结果；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/gateway-README.md)。 |
| F08 | 最终Gemini工具结果turn尾随part保序拆为user turn | 单/多工具、文本/图片、历史非末尾及parallel媒体控制；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/gateway-README.md)。 |
| F09 | 按SDK能力保留/过滤原生SSE注释及分隔 | LF/CRLF/混排/data重写/空data/DONE/keepalive/断连usage drain；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/gateway-README.md)。 |
| F10 | 现有3图上限校验返回typed 400 | OAuth/API Key×同步/异步×JSON/multipart；合法3图/拒绝4图无转发；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/gateway-README.md)。 |

初始复现、来源可行性、测试夹具/标签/并行中间状态与最终产品验收分别保存；不把intake的17个目标失败写成最终结果。原本原始值保护、固定quota/独立用户与账号成本/usage dedup均经当前产品控制测试。第三档能力、5图扩展、完整group allowlist与QuotaSummary不随最小修复引入；支付排除。
