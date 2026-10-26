# 实施验收证据

当前状态：**六项已实现并完成开发验收（2026-10-26），改动保留在未提交工作树**。2026-10-24用户后续授权实现，2026-10-25补充选择恢复完整运维页；[起点快照](../../../docs/upstream-sync/implementation-0.2.18-personal-features/start-state.json)保留原实施记录。本表使用实际产品测试、真实SQLite/miniredis、浏览器固定数据检查及最终文件哈希，不以[规格基线检查](../../../docs/upstream-sync/evidence-0.2.18-personal-features/baseline-checks.md)或OpenSpec结构校验代替实现验收。

最终产品为[72个文件的SHA-256清单](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/product-files.sha256)，与最终门禁对应；文档收尾复核时72项全部一致。发布、tag、push、部署未执行；change未归档。

## 实施基线

| 项目 | 实际值 |
|---|---|
| 实施开始日期 | 2026-10-24 |
| 实施起点SHA/工作树 | 67b1a07270acd79f5d54432b44ae396c3f7e360c；原spec/评估未提交，AGENTS/上游索引有对应文档修改；相关产品文件起点无修改 |
| 与冻结基线的差异 | c36a19b7a到67b1a0727仅归档文档变化，backend/frontend/deploy/.github无提交差异；冻结SHA/证据哈希保持不变 |
| 产品完成SHA/工作树 | HEAD为f96d00ef8517c7f2d6d6b73017a8bc742136a25f，产品实现尚未提交；使用上述72项文件哈希标识完成工作树，不能将HEAD当作实现提交 |
| 后续基线差异 | 67b1a0727到f96d00ef8仅历史OpenSpec任务核销文档变化；当前产品哈希与最终门禁一致，见[最终状态](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/final-state.json)与[收尾复核](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/closure-check.json) |
| VERSION、Go与最大迁移 | 1.1.19；backend/go.mod和实际构建工具链均为Go1.27.3；最大SQLite迁移226。本批未修改依赖、迁移、Ent/Wire生成文件或VERSION |

## 逐功能证据

逐场景索引见[22条需求、50个Scenario验收映射](../../../docs/upstream-sync/implementation-0.2.18-personal-features/acceptance-matrix.md)，包括完整运维页入口与既有F07兼容delta。具体测试名及参数见[前端映射](../../../docs/upstream-sync/implementation-0.2.18-personal-features/frontend/acceptance.md)和[后端映射](../../../docs/upstream-sync/implementation-0.2.18-personal-features/backend/coverage.md)；测试发现/标签核查见[go test -list](../../../docs/upstream-sync/implementation-0.2.18-personal-features/checks/completion-test-list.json)。下列结果均对应已存在、已执行的用例。

| 编号 | 必须验收的行为 | 实现前证据 | 实现后证据 | 反例/边界证据 | 完成SHA/缺口 |
|---|---|---|---|---|---|
| F01 | 两页面TPS、有效性、TTFT口径、列偏好/翻页 | [展示/筛选红例](../../../docs/upstream-sync/implementation-0.2.18-personal-features/frontend/red-f01-f02.log)：缺少格式化实现 | usageOutputRate、UsageTable及管理员/用户UsageView用例通过；100/2000且TTFT1500仍为50.0 tok/s；换行数据更新为40.0；见最终Vitest与浏览器截图 | 0.0与不可用区分；null/缺失/非有限/非正数；图片输出/图片计费/Live/unknown排除，图片输入后的文本仍有效；两页面保存的latency隐藏偏好保留 | 完成工作树；无新增API/数据库TPS字段 |
| F02 | 首请求恢复、清空/非法值/存储失败、分页选择 | 同上红例：无筛选存储helper | accountFilters、AccountsView.lite用例通过；首请求已含openai/error；第5页勾选后切筛选回第1页并清理选择；浏览器请求记录一致 | 坏JSON/null/数组、不支持平台及字段独立回落，读写抛错仍可用；清空保存；仅平台/状态两字段 | 完成工作树；未扩大平台或持久化范围 |
| F03 | 完整运维页、桌面/移动、名称/ID、缓存/分页/失败/会话/迟到响应 | [名称缓存红例](../../../docs/upstream-sync/implementation-0.2.18-personal-features/frontend/red-f03.log)；[入口恢复红例](../../../docs/upstream-sync/implementation-0.2.18-personal-features/frontend/ops-restoration-red.log) | accountNameMap、OpsRequestDetailsModal、feature-access与AppSidebar用例通过；恢复完整/admin/ops、侧栏及菜单项；standard/simple管理员可用；[入口修正定向结果](../../../docs/upstream-sync/implementation-0.2.18-personal-features/frontend/ops-restoration-green.log)及真实SQLite运维查询通过 | lite=1且不聚合调度分数；TTL60秒、分页有界、空页/重复页退出、缺项每轮至多额外刷新一次；#ID/-回退；名称失败保留明细；手动刷新、快速翻页/关闭/恢复上下文/会话切换及迟到响应；监控关闭/隐藏菜单、普通用户及未登录反例 | 完成工作树；仅恢复运维页，其他页面限制与指标口径不变 |
| F04 | scope兼容、401例外、实际模型、写失败、SQLite/miniredis/调度 | [初始红例](../../../docs/upstream-sync/implementation-0.2.18-personal-features/backend/red-initial.txt)：account scope被丢弃；[跨平台凭据红例](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/credential-markers-red.log) | 规则解析/旧规则回显/首命中及真实错误消费者通过；真实SQLite+内嵌Redis验证account/model持久状态和调度快照；alias-a映射model-a后仅对应组合排除；[凭据修正结果](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/credential-markers-green.log)及最终race通过 | 已知API Key模型401允许显式model；OAuth/SetupToken/影子及明确撤销/失效/Unauthorized保留鉴权处理；模型未知回退、写失败不升级账号阻塞、到期恢复；图片/文本键、池策略、粘性/WS、固定配额窗口及去重保留 | 完成工作树；复用JSON与现有模型限流，无迁移 |
| F05 | 配置三态、纯搜索、JSON/SSE、OAuth反例、资源/用量归属 | 初始红例：OpenAI搜索模式仍default；[真实handler红例](../../../docs/upstream-sync/implementation-0.2.18-personal-features/checks/handler-search-red.log)：异步任务读取复用Gin导致RequestedModel串请求 | 创建/编辑/渠道配置用例与TestOpenAIWebSearchMessagesHandler_*通过；JSON/SSE工具引用配对、唯一终态、最后用户查询及LLM零调用；使用量入队前快照修正后归属与一次记录正确；取消/写错/换号及槽位释放通过 | 全局/账号三态/渠道组合、旧bool、nil及typed nil；OAuth/SetupToken/影子/混合工具不拦截；10个SSE事件的Write/Flush取消和JSON写入内取消；空查询/代理/提供商错误；跨账号/类别/平台不串写 | 完成工作树；全部搜索使用本地桩，无真实收费服务调用 |
| F06 | 配置默认/边界、多消费者、原生/WS反例、实例隔离/provider cap | 初始红例：显式非法min_output_tokens未拒绝 | 配置加载、三种转换与实际转发请求通过；默认128、YAML/env16/128/256及优先级；三下限×六请求值×三转换；两实例16/256并行及race隔离；所属服务配置已贯穿现有消费者 | 0/负数/15/非整数/溢出报字段错误；缺省/非正请求与双字段优先级保留；nil/旧包装器128；原生Responses/compact/WS多轮、OAuth搜索历史、原生Claude探针和Ollama cap保持 | 完成工作树；未新增缺失legacy入口或可变全局下限 |

## 合并门禁

| 门禁 | 实际命令 | 结果/证据 |
|---|---|---|
| 后端unit/build | backend：`rtk proxy go test -tags=unit ./... -count=1`；`rtk proxy go build ./...`；`rtk proxy go test -tags=unit,embed ./internal/web -count=1` | 均退出0；54个有测试的包通过，另53个包无测试；[unit日志](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/backend-unit-final.log)、[build记录](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/backend-build-final.json)、[embed记录](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/backend-embed-final.json) |
| 受影响链路race | backend：`rtk proxy go test -tags=unit -race ./internal/config ./internal/pkg/apicompat ./internal/service ./internal/handler ./internal/repository -run '<下方完整表达式>' -count=1` | 五包退出0；[完整命令元数据](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/backend-feature-race-final.json)与[日志](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/backend-feature-race-final.log)；在最后凭据修正后执行 |
| 真实SQLite、miniredis与固定窗口/去重 | backend：`rtk proxy go test -tags=unit ./internal/repository -run 'Test(OpsAndUsageQueriesRunOnSQLite\|OpsUpsertHourlyMetricsUsesSQLiteAggregation\|OpsRequestDetailsTTFTSortRunsOnSQLite\|TempUnschedScopeSQLiteWithEmbeddedRedis\|UsageBilling.*SQLite\|UsageDedup.*SQLite\|AccountQuota.*SQLite\|.*Fixed.*Window.*)' -count=1 -v` | 退出0；实际InitRedis(redis.enabled=false)、Ent SQLite、模型/账号范围及调度快照，运维查询/聚合/TTFT排序、固定窗口、删除Key结算和去重通过；[命令](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/sqlite-ops-and-ledger.json)与[日志](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/sqlite-ops-and-ledger.log)；未将PG限定integration测试计入 |
| 前端typecheck/lint/Vitest/build | frontend：`rtk pnpm run typecheck`、`rtk pnpm run lint:check`、`rtk pnpm run test:run`、`rtk pnpm run build` | 均退出0；285文件、2089通过、2项既有跳过；[类型检查](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/frontend-typecheck.log)、[lint](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/frontend-lint.log)、[Vitest](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/frontend-vitest.log)、[构建](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/frontend-build.log)；既有pnpm.overrides位置与大chunk提示保留，未扩展依赖维护 |
| 桌面/移动及中文/英文 | 真实Vite页面+本地固定API桩；组件的桌面/移动参数化；中英文locale核对 | 管理员桌面/用户390×844移动端TPS和账号表单实际渲染、筛选首请求记录见[浏览器说明与截图](../../../docs/upstream-sync/implementation-0.2.18-personal-features/browser/README.md)；运维详情两布局/错误成功与英文文案由组件测试核对。浏览器桩未冒充真实后端或业务实例验收 |
| 六项共享路径组合回归 | 创建/编辑/渠道组件测试；真实OpenAI Messages handler及实际raw Chat出站请求；完整unit与定向race | scope与搜索三态同时保存且不覆盖模型映射/extra；搜索在转换前分流；非搜索转换使用实例下限；延迟worker复用Gin后仍保留请求归属。已修问题和失败夹具区别见[整合审查](../../../docs/upstream-sync/implementation-0.2.18-personal-features/review.md) |
| OpenSpec strict与差异检查 | root：`rtk pnpm dlx @fission-ai/openspec@1.14.3 validate port-upstream-0.2.18-personal-features --strict --json --no-interactive`；`rtk proxy git diff --check` | 最终文档收尾后退出0，22条需求/50场景；[strict日志](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/closure-openspec-strict.log)、[收尾复核](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/closure-check.json)。旧CLI诊断保留，固定版本通过不改项目依赖 |
| 无新平台/支付/迁移/依赖/VERSION变化 | 对比实际HEAD的依赖清单/锁文件、迁移、VERSION、Ent/Wire与CI；复核全部72个产品文件；校验冻结来源哈希 | 无上述文件变化；VERSION1.1.19/Go1.27.3/迁移226；未加入运维TPS百分位、缺失legacy入口、新平台或支付功能；只恢复已选运维页，不放开其他simple页面；[收尾复核](../../../docs/upstream-sync/implementation-0.2.18-personal-features/completion-2026-10-26/closure-check.json) |

Race使用的完整`-run`表达式：

```text
Test(LoadMinOutputTokens|ConfiguredMinOutputTokens|MinOutputTokens_|TempUnschedScope|OpenAIWebSearchMessagesHandler_|ShouldEmulateWebSearchFor_|WebSearchStream_Cancellation|WebSearchJSON_Cancellation|CountTokensEntries_Explicit|AnthropicFailoverEntries_Explicit|HandleTempUnschedulable_Explicit|CheckErrorPolicy_Explicit|OpenAIFailoverSideEffects_401|OpsAndUsageQueriesRunOnSQLite|OpsRequestDetailsTTFTSortRunsOnSQLite)
```

## 实际来源取舍

| 来源 | 需要记录的适配 | 实际记录 |
|---|---|---|
| #7911 | 仅日志TPS；不取运维分位数/PG查询 | 仅已有UsageLog字段的前端派生值，共用延迟列；没有新后端字段或SQL |
| #8142 | 本地平台/状态选项和首请求流程 | 仅五个既有平台与实际状态，初始化useTableLoader参数时恢复；未引入上游平台画像或紧凑布局 |
| #7933 | lite查询、有限缓存与生命周期 | 复用无凭据lite分页，60秒TTL、并发合并、页数/空页/重复页终止、手动失效与会话generation。后续用户选定的完整运维页恢复单列测试，不更改原始来源diff |
| #7828 | 已有模型隔离、凭据401及当前生产路径 | 在现有JSON/模型限流上增加显式scope；第一条命中决定范围；受约束的模型401与明确凭据失败分开；原图片冷却和F07 delta通过，无新迁移/后台服务 |
| #8080 | 本地创建/渠道配置补齐及真实handler记录 | 共享现有搜索管理器/提供商/代理和Anthropic输出；OpenAI API Key Messages在raw Chat/Responses转换前判断；补本地创建/编辑/渠道三态入口；修异步归属快照与取消伪成功，沿用零模型Token口径 |
| #8094 | 现有转换接线；不引入缺失legacy入口/平台 | 无配置包装器保持128，显式服务配置传入已有转换；count_tokens两处只取input字段的旧调用保留；原生Responses/compact/WS和provider cap不统一抬高，两个NOBASE未引入 |

## 完成条件

- [x] 六项所有必须场景及既有F07兼容delta均有可复核证据。
- [x] 实际命令与结果、来源取舍及完成工作树哈希已填写；产品未提交，不虚构完成SHA。
- [x] 个人单节点约束保持，未扩大到未选能力。
- [x] 无遗留必须实现项。未选功能、PG限定测试、发布与部署不计为本次产品验收；基线检查未被记作功能完成。
