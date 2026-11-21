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
| S01 | 最小requestmodel扫描+HTTP全链及每轮WS检查，预读/错误形状一致 | 152项初组、binary/session.update、真实TCP别名、fuzz及1C1G并发预算；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/README.md)。 |
| S02 | OpenAI完整轮换10s脱离取消SQLite CAS/outbox，重读持久状态 | proxy/非token合并、重授权胜出、写/重读失败、3冲突、实际SQLite/miniredis与并发CAS；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/README.md)。 |
| S03 | 从有效终态/完成item重建buffered hosted/custom输出与独立content index | 实际loopback HTTP/SSE；终态优先、failed/顺序/工具身份/Grok搜索控制；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/protocol-implementation.md)。 |
| S04 | 成功搜索次数严格presence/非负整数；显式0覆盖、缺失或非法fallback | 六种现有转发路径、item去重、终态锁定、普通同名函数/failed item排除；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/gateway-README.md)。 |
| S05 | OAuth-like按实际映射目标排除明确外族；Messages渠道映射后调度 | OAuth/setup-token/API Key、透传开关/映射别名/未知别名及既有候选池控制；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/README.md)。 |
| S06 | 批量透传与过滤独立opt-in，省略与false区分 | OAuth/API Key目标/重开/目标变化；真实桌面移动只改透传与独立false；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/README.md)。 |
| S07 | Claude/Gemini快照key参与缓存/刷新/异步归属 | 同ID两代promise和batch逆序、旧账号/离页、非可见移动行、错误保留读数/桌面移动；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/README.md)。 |
| S08 | 原生Responses audio声明按可服务候选保守交集 | OAuth Sol/API Key metadata/raw目录/映射/混合/Chat剔除/ETag邻域；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/protocol-implementation.md)。 |
| S09 | 有效映射metadata>=900000才自动展示[1m]；官方registry host修正 | 依据design采用可路由候选交集；mixed/unknown/routed/映射目标/大小写去重/实际models HTTP；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/README.md)。 |
| S10 | Composite改写模型时同步本地Claude5.5 thinking适配 | Sonnet enabled实际HTTP、disabled/Opus/nested-path/Bedrock复用既有与新控制；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/README.md)。 |
| S11 | 共享注册/后台新用户默认quota快照，Wire真实生成 | 普通/管理员、0/正数/缺省/失败fail-open、SQLite实际HTTP与BulkInsertInitial不覆盖；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/README.md)。 |
| S12 | 持久用量/审计标量独立存储，局部字段指针不滞留整个input | 1MiB alias字节/地址、延迟与取消/审计快照归属；原body预算生命周期控制；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/protocol-implementation.md)。 |
| S13 | 六个窄热点逐项差分/收益后采纳，handler仅同字节复用Validated | 旧函数差分/转义/freeform/非法JSON/范围刷新/成员图/miniredis/短fuzz/单P前后；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/protocol-implementation.md)。 |
| S14 | SQLite模型平均TPS按有效output总和/有效duration总和 | 1s/10s→18.18、NULL/0/无效/分页/实际ops HTTP；整窗口TPS原算术mean/分位数保留；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/README.md)。 |
| S15 | 同请求裸429最多冷却一个账号；quota使用实际转发端点 | retry+外层共用claim、独立请求/明确details/reset、daily/prod/override/失败不跨端点及原client边界；[详细记录](../../../docs/upstream-sync/implementation-0.2.24/gateway-README.md)。 |

初始复现、来源可行性、测试夹具/标签/并行中间状态与最终产品验收分别保存；不把intake的17个目标失败写成最终结果。原本原始值保护、固定quota/独立用户与账号成本/usage dedup均经当前产品控制测试。第三档能力、5图扩展、完整group allowlist与QuotaSummary不随最小修复引入；支付排除。
