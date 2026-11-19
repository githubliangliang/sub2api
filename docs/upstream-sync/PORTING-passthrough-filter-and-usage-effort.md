# 透传模型筛选与用量推理强度

2026-11-19 两项均已在 v1.1.26（`9c073cb30`，SQLite 229）基线上实施，共 12/12 项，新增 SQLite 迁移 230。[实施验收与文件校验](./implementation-passthrough-filter-and-usage-effort/README.md)。下文的适配位置与提案基线检查保留实施前的观察。

| 来源 | OpenSpec | 需要改动的数据 |
|---|---|---|
| #8748，v0.2.23 范围 | [OpenAI 透传账号按模型筛选](../../openspec/changes/port-upstream-passthrough-model-filter/README.md) | 既有账号 extra 增加一个可选布尔值 |
| #6188，首次收录于 v0.1.184 | [用量日志的请求与实际推理强度](../../openspec/changes/port-upstream-usage-reasoning-effort/README.md) | usage_logs 增加一个可空文本列 |

### PR8748

- **资格判断与发送。** [账号判定](../../backend/internal/service/account.go) 的 `isModelMappingSupported` 在透传时直接放行；[共享调度](../../backend/internal/service/gateway_scheduling.go) 另有透传短路。[退休规则](../../backend/internal/service/openai_model_retirement.go) 包裹账号匹配。[透传发送](../../backend/internal/service/openai_gateway_passthrough.go) 是核对上游实际正文的位置。
- **快照与列表。** [调度缓存](../../backend/internal/repository/scheduler_cache.go) 精简投影当前包含 `openai_passthrough` 及映射。把新开关沿同一路径传递。[可用模型聚合](../../backend/internal/service/gateway_service.go) 的 `GetAvailableModels` 会收集账号映射键。本地没有来源的 `openai_models_list.go` 或 `group_model_allowlist.go`，应接入[普通目录处理器](../../backend/internal/handler/gateway_handler.go)、[Codex 处理器](../../backend/internal/handler/openai_codex_models_handler.go)及[现有 manifest 服务](../../backend/internal/service/openai_codex_models_service.go)的分组配置、上游发现和最终 ETag。
- **管理入口。** [创建](../../frontend/src/components/account/CreateAccountModal.vue)、[编辑](../../frontend/src/components/account/EditAccountModal.vue)、[批量编辑](../../frontend/src/components/account/BulkEditAccountModal.vue) 已有透传及模型规则表单。沿用现有白名单输入校验，支持已可表达的通配映射来源键。
- [冻结补丁与 Git 对象](../../openspec/changes/port-upstream-passthrough-model-filter/source-baseline.md) · [实施验收矩阵](../../openspec/changes/port-upstream-passthrough-model-filter/verification.md)

### PR6188

- **记录前后值。** [HTTP 策略绑定](../../backend/internal/handler/composite_platform.go)、[请求体解析](../../backend/internal/service/openai_gateway_request_body.go) 和[分组策略](../../backend/internal/service/openai_reasoning_effort_policy.go) 定义采集先后。沿[WS ingress](../../backend/internal/service/openai_ws_forwarder_ingress.go)、[WS v2 转发](../../backend/internal/service/openai_ws_forwarder_v2.go)和[WS 透传](../../backend/internal/service/openai_ws_v2_passthrough_adapter.go)查每轮结果归属。[OpenAI 用量](../../backend/internal/service/openai_gateway_usage.go)及[共享用量](../../backend/internal/service/gateway_usage_billing.go)当前只给日志填写一个强度。
- **SQLite。** [既有强度迁移](../../backend/migrations/046_add_usage_log_reasoning_effort.sql)、[写入参数](../../backend/internal/repository/usage_log_repo_insert.go)及[查询与扫描](../../backend/internal/repository/usage_log_repo_query.go)是本地边界。上游迁移号 231 属于上游序列；本地基线的下一可用号为 230，实施时再次检查。历史行的请求值保留 NULL。
- **可见性。** [DTO 映射](../../backend/internal/handler/dto/mappers.go)分别供[用户列表及详情](../../backend/internal/handler/usage_handler.go)和[管理员列表](../../backend/internal/handler/admin/usage_handler.go)使用。沿[管理表格](../../frontend/src/components/admin/usage/UsageTable.vue)、[管理页与导出](../../frontend/src/views/admin/UsageView.vue)及[用户页与导出](../../frontend/src/views/user/UsageView.vue)验证角色差异。
- [冻结 PR、完整 diff 与校验值](../../openspec/changes/port-upstream-usage-reasoning-effort/source-baseline.md) · [实施验收矩阵](../../openspec/changes/port-upstream-usage-reasoning-effort/verification.md)

### 提案基线检查

在上述 fork 基线上以 `go test -tags=unit -count=1` 运行三个包的六项现有对照，全部通过：

| 包 | 执行的现有测试 |
|---|---|
| service | `TestIsModelSupported_OpenAIOAuthPassthroughIgnoresLeftoverMapping` |
| repository | `TestBuildSchedulerMetadataAccount_KeepsOpenAIPassthroughForModelGate`、`TestUsageLogRepositoryCreateSyncRequestTypeAndLegacyFields` |
| handler | `TestCodexModelsAppliesLocalFiltersBeforeClientETag`、`TestOpenAIReasoningEffortPolicyForCompositeTarget`、`TestOpenAIResponsesWebSocket_PassthroughUsageLogPersistsUserAgentAndReasoningEffort` |

这些测试固定本次移植需要保护的旧行为。仓储的旧字段测试使用 SQL mock；真实 SQLite 229 升级、新列读写、结算去重及用户可见行为都列在对应提案的实施验收中。

其余四个候选继续按实际使用选择：Antigravity 手动刷新额度（#8590）、GPT-6.2 Pro 套餐补全（#8906 剩余部分）、渠道 Ultra Fast 倍率（#8914）、30m/12h 缓存计量（#8606）。
