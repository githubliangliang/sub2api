# 六项规格场景验收映射

对照本change的22条需求、50个Scenario逐项核对（含后续选定的完整运维页入口及既有F07兼容delta）。前端用例/参数见[前端映射](./frontend/acceptance.md)，后端用例/调用链见[后端映射](./backend/coverage.md)；最终命令和结果由[verification](../../../openspec/changes/port-upstream-0.2.18-personal-features/verification.md)统一记录。本表不以“整套测试是绿的”替代具体行为证明。

## F01 使用日志TPS（7场景）

| 场景 | 直接证据 |
|---|---|
| 使用完整请求耗时 | usageOutputRate.spec.ts的100/2000=50.0；UsageTable.spec.ts含1500ms首字仍50.0；管理员真实浏览器fixture同值 |
| 极小但有效的速率 | usageOutputRate.spec.ts：1/100000ms显示0.0 tok/s |
| 缺少可信数值 | 同文件参数化0、负数、NaN、Infinity、undefined、null，输出数和耗时均返回不可用 |
| 图片与Live | 同文件image_count/image_output_tokens/image计费/live/unknown排除 |
| 图片输入后的文本输出和历史记录 | UsageTable.spec.ts带image_input_tokens=400仍计算；helper缺少request_type仍有效 |
| 默认显示与用户隐藏 | admin/user两份UsageView.spec.ts分别验证默认latency列与已保存隐藏偏好；中英文标签/提示均存在 |
| 分页或刷新 | UsageTable.spec.ts更换请求行/Token/耗时50.0→40.0再隐藏列；无新增API/数据库字段 |

## F02 账号筛选记忆（6场景）

| 场景 | 直接证据 |
|---|---|
| 刷新后首个请求采用保存值 | AccountsView.lite.spec.ts首个list请求带openai/error；浏览器requests.json真实首请求同样带值 |
| 清空与单字段修改 | accountFilters.spec.ts和页面用例保存platform空/status active，只有两个字段 |
| 非法结构和字段 | helper参数化null/数组/坏JSON/非对象，以及cline平台/非字符串字段独立回落 |
| 存储被禁用或空间耗尽 | helper读取/写入抛错；页面对该偏好键抛错仍可初始化及更改筛选 |
| 从后页切换筛选 | AccountsView.lite.spec.ts第5页选中42后改筛选，回第1页并清理原选择 |
| 保留本地平台范围 | helper拒绝cline、保留独立合法status；未导入新平台清单 |

## F03 运维请求账号名（7场景）

| 场景 | 直接证据 |
|---|---|
| 完整运维页具有实际入口 | feature-access.spec.ts验证完整/admin/ops路由、standard/simple管理员可用及未登录/普通用户拒绝；AppSidebar.collapse.spec.ts验证入口、监控开关和隐藏菜单；menuCatalog包含运维项。OpsAndUsageQueriesRunOnSQLite、OpsUpsertHourlyMetricsUsesSQLiteAggregation、OpsRequestDetailsTTFTSortRunsOnSQLite验证真实SQLite运维查询，最终前端2089项通过 |
| 成功和错误记录 | OpsRequestDetailsModal.spec.ts桌面/移动参数化success/error记录，解析到account-alpha |
| 历史账号已删除或没有归属 | 同用例999回退#999、null为-，主明细与model保留 |
| 名称请求失败 | 组件刷新失败保留明细；accountNameMap.spec.ts失败不永久缓存，允许后续重试 |
| 多页账号列表 | cache测试断言lite=1、include_scheduler_score=0、AbortSignal和后续页；重复/空页终止 |
| 账号新增或重命名 | TTL/强制刷新、缺项一次刷新；普通pending替换和连续手动刷新旧forced失效回归 |
| 快速翻页、关闭和重新登录 | 组件较慢旧页/筛选响应不覆盖、关闭重开共享、resumeState续接和会话变更；cache取消旧会话请求 |

## F04 模型级规则（11场景）

| 场景 | 直接证据 |
|---|---|
| 显式账号级与模型级 | TestHandleTempUnschedulable_ExplicitAccountScopePausesWholeAccountWithKnownModel；TestCheckErrorPolicy_ExplicitModelScopeKeeps401ModelScoped；reason JSON的scope/model断言 |
| 编辑旧规则保持行为 | TestGetTempUnschedulableRules_ParsesScope；TestTempUnschedRuleUsesModelScope；EditAccountModal旧429/401/显式scope顺序回显和提交 |
| 多条规则同时匹配 | TestTempUnschedScope_FirstMatchWriteFailureDoesNotFallThrough；首account后model的401优先级用例 |
| 别名和独立模型 | 真实SQLite+内嵌Redis测试读持久账号/调度快照，alias-a→model-a被排除、model-b仍可用；原粘性/WS模型调度测试 |
| 保留图片隔离 | 原TestIsModelRateLimited_AntigravityImageDoesNotInheritGeminiFamilyLimit及Gemini图片/文本族键回归；生产键归一未重写 |
| 未知模型与到期 | TestHandleTempUnschedulable_ExplicitModelScopeFallsBackToAccountWhenModelUnknown；SQLite设置过去reset后alias恢复 |
| 模型状态写入失败 | 原TestOpenAIModelTempUnschedulable_WriteFailureDoesNotRuntimeBlockWholeAccount；当前请求换号，不写全局阻塞 |
| API Key模型权限错误 | TestCheckErrorPolicy_ExplicitModelScopeKeeps401ModelScoped；TestHandleUpstreamError_ExplicitModelScope401；count_tokens/Anthropic错误入口用例 |
| 凭据无效但配置model | OAuth/SetupToken/影子及token_revoked/token_invalidated/detail Unauthorized反例；最后跨既有平台明确标记回归保留原鉴权路径 |
| SQLite持久化与miniredis | TestTempUnschedScopeSQLiteWithEmbeddedRedis实际InitRedis(redis.enabled=false)、Ent SQLite、SchedulerCache和TempUnschedCache；固定窗口/删除Key/去重真实SQL回归 |
| 未命中或既有策略跳过 | 规则开关/关键词/错误码匹配测试及原池模式WithoutCustomPolicy/CustomPolicy、Grok现有跳过策略回归 |

## F05 OpenAI API Key搜索模拟（9场景）

| 场景 | 直接证据 |
|---|---|
| 强制开启与渠道继承 | TestShouldEmulateWebSearchFor_OpenAIEligibility的enabled/default+渠道openai标志 |
| 全局、账号或渠道未允许 | 同测试disabled/全局false/nil setting/account与原nil manager/无渠道测试；typed nil管理器回归 |
| 配置回显与兼容 | 创建/编辑enabled/disabled/true/false/default删除override，切账号/类别/双向平台重置；渠道创建/编辑独立保存Anthropic和OpenAI标志 |
| 非流式和流式成功 | TestOpenAIWebSearchMessagesHandler_UsageAndSlots真实Messages入口，强制Chat兼容账号仍只调搜索、LLM零调用、唯一终态 |
| 最近用户消息 | 真实handler多轮请求只搜索latest search query；原文本/数组提取回归 |
| OAuth或混合工具 | eligibility的OAuth/SetupToken/影子/混合工具反例，ForwardAsAnthropic仅OpenAI平台拦截 |
| Anthropic既有路径 | 原TestShouldEmulateWebSearch_*继续通过，共享JSON/SSE写出仍返回Anthropic协议；模型Token统一真实零值 |
| 输出前失败和输出后写失败 | handler提供商失败/取消/SSE写错/代理换号及槽释放；TestWebSearchStream_CancellationAtEveryEventBoundary的10事件×Write/Flush和JSON写入内取消；独立overlay红→绿 |
| simple mode和使用量归属 | 真实handler simple mode，用户/Key/组/账号/模型/UA及零Token；阻塞worker+确定性Gin复用仍保留public-search-alias，且只记录一次 |

## F06 最小输出Token（8场景）

| 场景 | 直接证据 |
|---|---|
| 默认与自定义 | TestLoadMinOutputTokensDefaultsAndConfiguration，默认128、env16/128/256、YAML256及既有env优先级 |
| 非法配置 | TestLoadMinOutputTokensRejectsExplicitInvalidValue及非法env，0/负数/15/分数/非数字/溢出均报字段级错误 |
| 小于、等于和大于下限 | TestConfiguredMinOutputTokens_AllConverters，3个floor×6个输入×3种转换 |
| 缺省和多种字段 | 上述非正值省略与FieldPrecedenceAndLegacyDefault，max_completion_tokens优先，旧接口128 |
| 不同已有转发路径一致 | 全部本地实际生成转换调用使用所属cfg；ForwardAsAnthropic真实raw Chat出站请求16/256；count_tokens只提取input字段保持旧接口 |
| 服务实例隔离与旧调用 | TestConfiguredMinOutputTokens_DefaultAndInstanceIsolation与两配置实际转发并行/race，nil/未初始化默认128 |
| 原生Responses、compact与多轮WS | TestMinOutputTokens_NativeResponsesAndWebSocketRemainNative；原WS/OAuth搜索历史完整unit；不引入缺失legacy入口 |
| 提供商上限和原生探针 | 原Ollama clamp/cap和原生Claude Code max_tokens=1探针测试；最终记录使用实际usage而非配置上限 |

## 既有F07兼容delta（2场景）

| 场景 | 直接证据 |
|---|---|
| 目标行为 | 既有Gemini工具结果图片→inlineData、顺序、图片/文本冷却隔离回归在完整backend unit中通过 |
| 反例与范围边界 | F04新规则UI按既定scope实现；原文本族冷却、工具图片路径及不下载远程图片行为保持，未新增平台或迁移 |
