# 规格阶段基线检查

日期：2026-10-24；fork SHA：`c36a19b7a1a132e690eaf9ce0d878b301ffffcb4`；VERSION1.1.19，Go1.27.3，最大迁移226。检查相关产品文件未见未提交差异，用户正在进行的文档归档不纳入产品基线。

这是现有测试的可运行性核查，**不是六项新增功能已实现的证据**。不因本次通过而勾选tasks或填写实施verification。

## Go测试目录

在backend目录运行（实际shell命令均以rtk前缀执行）：

```bash
rtk proxy go test -tags=unit ./internal/config ./internal/pkg/apicompat ./internal/service ./internal/repository -list 'Test.*(MinOutput|MaxTokens|TempUnsched|ModelRateLimit|EmulateWebSearch|WebSearchEmulation|HTTPIngressSafetyDefaults|LoadDefaultSchedulingConfig)'
```

退出0，四包均可列目录。确认存在：

- TestAnthropicToResponses_MaxTokensFloor、TestAnthropicToChatCompletionsRequest_MaxTokensFloor、TestChatCompletionsToResponses_MaxTokens。
- TestGetWebSearchEmulationMode_*、TestShouldEmulateWebSearch_*、TestChannel_IsWebSearchEmulationEnabled_*。
- TestOpenAIModelTempUnschedulable_DoesNotRuntimeBlockWholeAccount。
- TestOpenAIModelTempUnschedulable_WriteFailureDoesNotRuntimeBlockWholeAccount、TestOpenAITempUnschedulable_UnknownModelKeepsAccountRuntimeBlock。
- TestRateLimitService_HandleTempUnschedulable_*、TestRateLimitService_TempUnschedulableContextPreservesModelForPoolDependency、TestRateLimitService_ModelTempUnschedulableIsolatesSchedulerByModel。

## Go定向基线

```bash
rtk proxy go test -tags=unit ./internal/config ./internal/pkg/apicompat ./internal/service ./internal/repository -run 'TestLoadHTTPIngressSafetyDefaults$|TestLoadDefaultSchedulingConfig$|TestAnthropicToResponses_MaxTokensFloor$|TestAnthropicToChatCompletionsRequest_MaxTokensFloor$|TestChatCompletionsToResponses_MaxTokens$|TestGetWebSearchEmulationMode_|TestShouldEmulateWebSearch_|TestChannel_IsWebSearchEmulationEnabled_|TestOpenAIModelTempUnschedulable_|TestOpenAITempUnschedulable_UnknownModel|TestRateLimitService_HandleTempUnschedulable_|TestRateLimitService_TempUnschedulableContext|TestRateLimitService_ModelTempUnschedulable|TestAccountRepository_SetTempUnschedulable_' -count=1
```

退出0：config 0.029s、apicompat 0.041s、service 0.813s、repository 0.112s。旧OpenAI搜索模式测试仍断言其返回default，实施F05时需要以新行为和OAuth反例替换，不能保持旧断言宣称新增路径已覆盖。

## 前端定向基线

在frontend目录运行：

```bash
rtk pnpm exec vitest run src/utils/__tests__/latencyHealth.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts src/views/user/__tests__/UsageView.spec.ts src/views/admin/__tests__/UsageView.spec.ts src/views/admin/__tests__/AccountsView.lite.spec.ts src/views/admin/ops/components/__tests__/OpsRequestDetailsModal.spec.ts src/components/account/__tests__/EditAccountModal.spec.ts src/components/account/__tests__/CreateAccountModal.spec.ts
```

退出0，8文件/122项通过：2、25、7、17、8、4、40、19项；用时7.05秒。仅有既有pnpm.overrides位置和Browserslist数据陈旧提示，未修改依赖。

## Build tag与待新增覆盖

- 本地temp_unsched_test.go首行为`//go:build unit`，所以上述命令保留unit tag。
- 本地account_repo_temp_unsched_test.go首行为`package repository`，当前可编译；现有测试主要是SQL mock，不冒充真实SQLite范围验收。
- #7911新增的ops_repo_dashboard_integration_test.go当前不存在且其运维分位数不在范围内；不引用它作为本次可运行测试。
- 后续引用任何integration测试时先核对首行，`integration && postgres`不能算SQLite验收。
- 账号筛选持久化、完整名称缓存生命周期、可配置scope401、OpenAI搜索拦截、MinOutput配置测试尚待实现；不将来源PR里的新测试名称标为本地已执行。

## 文档工具

OpenSpec通过`rtk pnpm dlx @fission-ai/openspec`运行，实际版本1.14.3。裸openspec不在PATH，但pnpm dlx可用；规格严格校验结果单列spec-validation.md。
