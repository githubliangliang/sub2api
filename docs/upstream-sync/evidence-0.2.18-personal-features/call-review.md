# 新增调用与本地基座复核

2026-10-24；固定fork SHA为`c36a19b7a1a132e690eaf9ce0d878b301ffffcb4`。全体新增Go行的逐调用词法结果见added-call-inventory.tsv，共570行；SQL、注释、标准库和局部变量误命中不能作为缺失实现证据。

| 功能 | 实测已有基座 | 新增/缺失与处理 |
|---|---|---|
| F01 | UsageLog类型、输出/耗时/类型、两页面共享UsageTable和latencyHealth | formatUsageOutputRate为新16行前端helper。queryOutputTPS及OpsOutputTPS不选，SQL函数误命中不转SQLite |
| F02 | useTableLoader、当前五平台与六状态选项、localStorage偏好 | CONCRETE_PLATFORM_OPTIONS本地无。load/save为新局部逻辑，不引入平台目录 |
| F03 | accountsAPI.list(lite/include_scheduler_score)、account_id、已有TTFT详情 | getAccountNameMap是新约33行模块思路；需要有限缓存/分页终止/会话等本地边界 |
| F04 | GetTempUnschedulableRules、matchTempUnschedulableRules、triggerTempUnschedulable、tempUnschedulableModel、SetModelRateLimit及调度过滤 | normalizeTempUnschedScope、tempUnschedRulesApply、tempUnschedHitUsesModelScope、tempUnschedRuleUsesModelScope、unauthorizedRequiresCredentialHandling由来源定义；与消费者同批适配 |
| F05 | GatewayService.shouldEmulateWebSearch/handleWebSearchEmulation、GetWebSearchEmulationMode、doWebSearch和JSON/SSE写出 | shouldEmulateWebSearchFor、writeEmulatedWebSearch是抽取现有逻辑；本地ForwardAsAnthropic需在早期Chat分支前调用；创建/渠道setter额外补齐 |
| F06 | 三种apicompat转换、默认128、config loader、调用方各服务cfg | WithMinOutputTokens及configuredMinOutputTokens包装器为来源内新定义；不需要新增依赖 |

F06必须区分：上游新调用normalizeOpenAIResponsesLegacyIngressWithConfig/convertLegacyResponsesMessages依赖222行legacy入口，而本地根本没有对应入口。现有normalizeOpenAIResponsesCompatibilityBodyWithOptions位于openai_oauth_web_search_history.go，只为搜索历史补工具声明；不能凭函数名复制其Config字段和调用。该入口及缺失native Anthropic网关均不选。

F04复用的SetModelRateLimit通过现有账号extra和调度快照写入；F03复用的lite接口不含credentials；两者不需要数据库迁移。Go方法同名/文件CLEAN不证明运行路径已经接线，实施时仍须真实消费者断言。
