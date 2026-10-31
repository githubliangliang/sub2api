# 用户选定补充：工具兼容、Responses序号与Grok诊断

2026-10-31 用户要求补齐：Codex工具参数/别名/defer、Responses序号、Grok模型目录诊断。前两类覆盖历史第二档C01–C04，后一项为用户选定的第三档。当前实施起点为`5250907f5b8294eabed3d3101d2c4bf79936e0f7`，VERSION1.1.21、Go1.27.3、SQLite227；已有模型参考价格提交保留。

[原冻结复查](./UNMERGED-REVIEW-2026-10-31.md)、[兼容性规格](../../openspec/changes/port-upstream-historical-client-compatibility/README.md)、[Grok诊断规格](../../openspec/changes/port-upstream-0.2.19-grok-model-diagnostics/README.md)。不改写旧来源SHA、评估、原始失败探针或历史验收。

## patch sites 与适配

- C01–C03沿原复查的实际转换器/fallback/Grok出口适配，新增helper随消费者落地；参数收窄和custom别名配套验证。
- C04沿类型序列化、实际SSE写出、compact和WS桥修复，已知流序号从真实写出状态继续，heartbeat不消耗数字序号。
- D01新增`backend/internal/service/grok_model_diagnostics.go`，复用现有scope/观测/URL和模型alias，不引入新CAS、授权或目录系统。
- D02为`handler/admin/account_handler.go`的详情和同步结果提供诊断；DTO与`UpstreamModelCatalog`只加可选字段，同步后读取最新账号。
- D03为实际`frontend/src/components/account/EditAccountModal.vue`、`ModelWhitelistSelector.vue`和诊断面板接线，复用现有detail/sync接口。来源frontend patch中的缺平台字段不移植。

## 验证

[本次起点及保护哈希](./implementation-historical-client-compatibility/baseline.json)。正式回归、完整backend/frontend门禁、SQLite/dedup/固定窗口、定向race、严格协议及真实浏览器结果写入两份verification和独立实施证据，原评估探针不冒充修复验收。

VERSION、Go、迁移、依赖及Ent/Wire保持当前范围。付款、新模型/平台、Fast/Realtime、统计扩展、发布与部署均独立于本次选择。

## 实施完成

用户选定三类行为已在本地main实现`becd2d2a31192e13bdc431bc43fbe44436e2c319`，C01–C04 17/17、D01–D03 12/12验收完成。后端54包unit/build/lint0、四包456及WS相邻74 race、SQLite468保护通过；前端2169/2 skipped、typecheck/lint/build/critical和真实浏览器通过。版本1.1.21、Go1.27.3、迁移227、依赖和生成代码保持；完整结果见两份verification及[独立证据](./implementation-historical-client-compatibility/README.md)。
