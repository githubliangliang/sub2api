# 历史未合内容 F01 / S01 实施证据（2026-11-26 本地 / 2026-11-25 UTC）

验收日志均为 UTC；本地 Asia/Shanghai 于 2026-11-25 开始、2026-11-26 完成。用户选择实现并本地 commit 两份 OpenSpec。真实起点为 `b9d9524c6cce8488edb2eaf899d771176f3b2414`（v1.1.28，tree `37f360f7f8f36afbb392c3fbdfccc94f331cc176`）。本批完成第一档 F01 的 5 项任务及第二档 S01 的 6 项任务。未发布、push、归档或操作业务实例。

[第一档验收](../../../openspec/changes/port-upstream-historical-backlog-tier1/verification.md) · [第二档验收](../../../openspec/changes/port-upstream-historical-backlog-tier2/verification.md) · [冻结评估](../UNMERGED-REVIEW-2026-11-25.md) · [冻结评估证据](../unmerged-review-2026-11-25/README.md)。评估目录和两份 source-baseline 保持原文，旧“未实施”状态是该轮历史快照。

## 实现及适配

- F01 移除公共 WHERE 中无依据的 `gpt%` 限制；保留公共时间/platform/group 过滤、分组优先/无分组回退账号平台、TopN/分页及兼容接口名。TPS 仍为有效输出与时长求和之比。既有卡片标题、空态、失败提示和设置说明改为覆盖实际平台范围。
- S01 对现有 native JSON、SSE 和 OAuth SSE 聚合逐事件观察错误信封、内容策略和真实空结果，按信号优先级只登记一次。STOP、MAX_TOKENS、普通工具停止及未知停止原因不记作上游故障。OAuth 解包/帧格式和聚合取最后带 parts 块的原行为均保留。
- 复用本地单标记 `OpsStreamError`、`CountTowardsSLA` 和上游事件合同，增加请求级/非流式标志，没有移植上游 WS 多标记/代理快照体系。策略结果作为 client/request/400、business-limited 记录，排除 SLA；此前尝试的恢复行独立保留或按原 skip 规则跳过。错误信封/空结果保持下游 200，Ops 使用语义状态归因。
- 非 data 兜底最多 1 MiB，超限放弃判定。保留有界 reader 和读错误/取消所有权。错误事件只保留既有 matcher 实际检查的 8 KiB 前缀、至多 2 KiB 详情/消息和有界原因字符串，避免切片继续持有整帧；该前缀只用于 skip 匹配，不入队。JSON 转义字段名正常识别，不使用会遗漏转义键的词法跳过判断。
- 真实 Ops 队列关停 race 暴露 worker 读取全局通道与 stop 置 nil 的竞争。必要修正仅让 worker 持有创建时的通道；队列容量、批处理、超时和过滤规则均保留。

## 可重复行为证据

| 记录 | 结果与边界 |
|---|---|
| [原缺陷重放](baseline-replay.json) / [修复后重放](baseline-replay-fixed.json) | 同一冻结探针的选定子集：最初 2 顶层失败/1 正常对照通过，修复后 3 顶层/9 含子例通过。只对内部 helper 参数作调用适配；未执行其它可选能力探针。 |
| [首轮定向](targeted-first.json) | 四包 59 顶层/172 含子例通过，含真实迁移 SQLite 的平台/group 交叉、无分组及空平台回退、半开边界、TopN/分页/空窗/总数、加权 TPS 和费用不变；HTTP 管理接口覆盖非 gpt 模型。 |
| [真实 HTTP/SQLite](http-sqlite-first.json) | 16 场景/17 含父例通过。真实本地上下游 TCP、native 服务、Ops 中间件及实际队列落库；真实 SQLite 230/Ent、miniredis OAuth token 缓存。覆盖 JSON/SSE/数组、OAuth 解包/聚合、优先级、策略/正常停止/空结果和原始正文 skip。 |
| [关停修正后三轮 race](ops-worker-shutdown-fixed.json) | 3 次完整 16 场景，3 顶层/51 含子例通过。每个响应只调上游一次，Ops 实际落库/SLA 查询；usage 重放两次，用户倍率 2 与账号倍率 0.5 独立，扣款/quota/dedup、simple 仅记录保持原行为。 |
| [最终四包 race](backend-race-complete.json) | 318 顶层/717 含子例通过；含 native 错误残片/活跃上游读失败/带 usage 断连、写失败所有权、有界 JSON/SSE reader、固定 quota reset、SQLite dedup 和已有账务控制。 |
| [前端组合](frontend-combination.json) | 18 文件/108 测试通过，含 9 项统计卡片测试；Claude/Gemini/Grok 平台和 group 切换重置分页、全平台模型显示与既有交互。初次 [定向](frontend-targeted.json) 实际匹配 1 文件/8 项，另两个给定名称没有对应测试文件，未计作覆盖。 |

## 组合门禁与最终文件

| 门禁 | 记录 |
|---|---|
| 首轮 backend unit | [55 包，10820 顶层/22494 含子例，16 既有 skip](backend-unit.json) |
| 8 KiB 保留修正后全量 unit | [55 包，10821 顶层/22495 含子例，16 既有 skip](backend-unit-final.json) |
| 随后 Ops 通道修正的完整 handler unit | [819 顶层/1973 含子例，1 既有 skip](backend-handler-unit-final.json)；其余最终源文件已由上述全量覆盖；最终四包 race、应用 build/lint 再覆盖队列修正 |
| 最终 backend lint / embed build | [0 issue](backend-lint-publish.json) / [通过](backend-build-publish.json)；产物在临时目录，不运行服务 |
| 前端 typecheck / lint / build | [通过](frontend-typecheck.json) / [通过](frontend-lint.json) / [通过](frontend-build.json) |
| 两份完成状态 strict | [tier1](openspec-tier1-complete.json) / [tier2](openspec-tier2-complete.json) |
| 最终 18 个产品/测试文件 | [SHA256](implementation-files.sha256)；最终 source/dependency/schema/generated 和工作区保护由 [核验](validation.json) 记录 |

未执行新的 PostgreSQL/integration、供应商在线或生产容量验收。所有数据库、HTTP 服务、Redis 和构建产物均为测试临时资源。

## 失败记录与限制

1. [8 KiB 暂存前置失败](signal-retention-before.json) 属于实施适配中的资源边界复核：64 KiB 信封被整块保留（65594 字节），新断言失败。修正复制 matcher 前缀，并在截断字符串转为持久字段时解除大缓冲引用，最终 unit/race 通过。
2. [首轮四包 race](backend-race.json) 和 [已启动的随后 native race](backend-race-final.json) 仍使用通道修正前的 worker；两者都保留真实数据竞争原文，未改成通过。修正后的三轮 HTTP/SQLite race 和最终四包 race 独立记录。
3. 初次 strict [tier1](openspec-tier1.json) / [tier2](openspec-tier2.json) 因 PATH 没有 `openspec` 失败；沿用评估时已安装的 1.15.0 CLI 的固定 node 路径通过，未安装或改依赖。随后 [pinned tier1](openspec-tier1-pinned.json) / [tier2](openspec-tier2-pinned.json) 与完成状态校验分别保留。
4. 前端已有 pnpm 配置、Browserslist、chunk/dynamic-import 提示以及预期接口失败测试日志保留；检查退出码均为 0，没有更新依赖来消除提示。

## 查询范围扩大的成本

[合成夹具](query-cost.go.txt) 在同一连接内存 SQLite 的 20000 行、5 平台、200 模型上比较完整 count+TopN 查询。`-cpu=1 -benchtime=200ms -count=3`，最终两条命令时间窗不重叠；[原始基线](query-cost-baseline-serial.json)、[当前](query-cost-current-serial.json)、[样本中位数](query-cost-summary.json) 分开记录。全范围模型由 100→200，单平台 20→40，返回条数均 20；更大统计范围有相应查询开销，不宣称性能优化或生产 1C1G 容量。前面的有并行编译/测试的探索样本 `query-cost-baseline`、`query-cost-*-initial` 也保留，不代替最后串行样本。

| 范围 | 基线中位数 | 当前中位数 | 时间变化 |
|---|---|---|---|
| 全平台 | 18.610 ms | 29.309 ms | +57.49% |
| OpenAI | 13.595 ms | 17.375 ms | +27.80% |
| Claude | 13.613 ms | 17.702 ms | +30.04% |
| Grok | 13.608 ms | 17.348 ms | +27.49% |

## 工作区与历史保护

实际启动已有 2 份 tracked 文档修改及 228 份未跟踪文件，区别于评估启动时的 23 份；[初始状态](initial-state.json.gz)、[两份原文](initial-AGENTS.md.gz)、[原 README](initial-upstream-README.md.gz) 和 [原文档差异](initial-documentation.patch) 保存真实输入。只对本批 README/design/tasks/verification 回填，其余初始未跟踪资料逐文件保持；先前 recheck 和 Fast diagnostic 均保护。原 15 活动变更/456 完成任务保持，新增两份完成 11 项后共 17 活动变更/467 完成任务。

VERSION 1.1.28、Go 1.27.4、SQLite 230/auth 21、依赖、Ent/Wire 和 64 MiB 请求体/40 MiB SSE 行/8 MiB staging/120s idle 默认保持。没有新增入口、平台、模型、账号处罚、重试、quota 策略、支付或其它按需能力。

`run.py` 保留每条实际命令、时间、退出码及原输出。`pack.py` 只对本目录大日志作可逆 gzip，命令 JSON 的 `stdout_log`/`stderr_log` 和 SHA256 指向可校验原文；`verify.py` 检查通过门禁、文件保护、任务及链接。`artifacts.json` 是本目录文件清单和哈希，不含自身。
