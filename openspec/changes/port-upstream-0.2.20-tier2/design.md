# Context

以当前实际fork调用链为准，区别可达缺陷、同语义不同名称、缺失新能力及已满足行为。评估使用隔离Go overlay，未修改产品代码。

# Decisions

1. 优先Claude/Codex/Grok；共享内存与数据完整性同样优先。Bedrock帧边界可独立先落，不等待大型预算方案。
2. 保留Fast工作树已落地行为；特别是WS每轮档位、账号catalog声明、OAuth default echo、Free Fast用户费用与账号成本分开，以及显式零价格。
3. 沿本地SQLite228、usage billing dedup和固定quota窗口验收；不导入PG schema、MCP迁移249或支付依赖。
4. Grok使用既有输入处理器做窄转换；Terra是已选能力后续，API与Codex标尺分开，官方fallback不覆盖账号自己的声明。
5. 资源治理按tag最终组合适配；预算和超时实际配置域是gateway，默认值必须满足1C1G测量。保留正常长流，断开drain到期只记录真实观测用量。未来若需要tag后自适应预算，应先另行纳入来源。
6. 只摘独立的RPM/JSON预检优化，实施需量测查询次数及分配；不为优化引入ojson/COW整体系。快照版本从本地20演进，不能复制上游27。
7. 修改接口时更新全部实际调用/测试stub；不存在的接口不能通过空stub或导入新平台来满足。

# Validation

每簇先建立预期行为失败回归，再实现并覆盖正反例。完整backend unit/embed build/lint、相关race、真实HTTP/SQLite settlement与dedup、迁移228/quotas；S09涉及前端时增加pnpm全套与真实浏览器导出/取消。资源步骤增加受限内存压力、HTTP1/2/WS、重试/取消与泄漏验证。所有来源测试需核查构建标签，PG-only测试不能冒充SQLite证据。

# Rollout

本提案为评估产物，实施任务未执行。实施提交、默认值实测、测试输出及残余风险填写verification；不会因此发布、推送或部署业务实例。
