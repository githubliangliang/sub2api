# 个人版六项功能规格

状态：**六项已实现并完成开发验收（2026-10-26），实现提交为`0e38d06d82557a0a3b372ca9165a24b9b3acbc93`**。2026-10-24 用户先选定六项原第三档功能并要求整理为 spec，随后授权实现；2026-10-25补充选择恢复完整运维页。见[验收报告](./verification.md)及[逐场景映射](../../../docs/upstream-sync/implementation-0.2.18-personal-features/acceptance-matrix.md)。用户后续要求提交并发布v1.1.20，另作VERSION发布准备提交，见[发布说明](../../../docs/releases/v1.1.20.md)及[检查证据](../../../docs/releases/evidence-v1.1.20/readiness.md)；实施验收使用的1.1.19及原始证据不回写。未部署业务实例，change尚未归档。历史分档和冻结来源不改写。

| 编号 | 功能 | 行为规格 |
|---|---|---|
| F01 | 使用日志输出 TPS | [usage-output-tps](./specs/usage-output-tps/spec.md) |
| F02 | 账号平台、状态筛选记忆 | [account-filter-persistence](./specs/account-filter-persistence/spec.md) |
| F03 | 运维请求详情显示账号名称 | [ops-request-account-label](./specs/ops-request-account-label/spec.md) |
| F04 | 临时不可调度规则的账号级、模型级作用范围 | [model-scoped-cooldown](./specs/model-scoped-cooldown/spec.md) |
| F05 | OpenAI API Key 的 Web Search 模拟 | [openai-apikey-websearch-emulation](./specs/openai-apikey-websearch-emulation/spec.md) |
| F06 | 网关最小输出 Token 配置 | [gateway-min-output-tokens](./specs/gateway-min-output-tokens/spec.md) |

TPS 范围为管理员与用户使用日志中的单请求输出速度，沿用延迟列显隐偏好。运维 P5/P10/P50/平均值卡片不在本次选择范围内。

- [proposal](./proposal.md)：动机、范围与能力清单。
- [source-baseline](./source-baseline.md)：冻结的 fork/tag/完整 PR 来源。
- [source-feature-map](./source-feature-map.md)：上游行为、本地差异及验收映射。
- [design](./design.md)：旧数据、配置默认值、错误和生命周期边界。
- [tasks](./tasks.md)：实施清单，只在真实实现并验收后勾选。
- [verification](./verification.md)：实施起点、逐项验收和组合门禁证据。
- [既有规格兼容调整](./specs/existing-gateway-correctness/spec.md)：保留 Gemini 图片隔离，解除旧批次对规则 UI 的范围限制。
- [补充评估及 patch sites](../../../docs/upstream-sync/PORTING-0.2.18-personal-features.md)：本次范围、符号核查、基线测试和历史遗留。

已保持 SQLite-only、可选 miniredis、simple mode、既有使用量记录和固定配额窗口。无 schema/迁移、Ent/Wire 生成、依赖升级或 VERSION 调整；无支付、新平台及其前置。后端54个测试包、定向race/真实SQLite/miniredis、前端2089项（2项既有跳过）及类型检查/lint/构建通过；完整运维页入口恢复，其他页面限制保留。最小输出Token配置为`gateway.min_output_tokens`，默认128、最低16，仅作用于已有协议转换。
