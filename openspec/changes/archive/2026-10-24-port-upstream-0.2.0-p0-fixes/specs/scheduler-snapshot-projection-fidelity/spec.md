## ADDED Requirements

### Requirement: 调度快照投影必须携带模型门判定所依赖的全部字段
写入 `sched:meta:<id>` 的账号投影 SHALL 携带 `Extra` 上的 `openai_passthrough` 与
`openai_oauth_passthrough`。凡是候选过滤阶段的判定函数会读取的 `Extra` / `Credentials` 字段，
投影 MUST 一并保留；投影 MUST NOT 出现「保留了判定的一侧输入、裁掉了另一侧」的组合。

依据：投影里 `Credentials.model_mapping` 被完整保留，而透传开关被裁掉，于是同一份数据上
`Account.IsModelSupported` 的透传短路失效、退回按（常已过期的）白名单判定。转发阶段从完整账号
hydrate ⇒ 症状是「单独测账号能通、走网关报 no available accounts / 404」。

#### Scenario: 透传账号请求白名单外的模型
- **WHEN** 某 OpenAI 账号 `Extra.openai_passthrough = true`，且 `Credentials.model_mapping` 是一份
  不含请求模型的残留白名单
- **THEN** 该账号经投影 round-trip 之后 MUST 仍被 `IsModelSupported` 判为支持该模型
- **THEN** 候选过滤 MUST NOT 把它记为 `model_not_supported`

#### Scenario: 历史 OAuth 开关字段
- **WHEN** 账号只有兼容字段 `Extra.openai_oauth_passthrough = true`（无新字段）
- **THEN** 行为 MUST 与上一个 Scenario 一致

#### Scenario: 投影经过 JSON 序列化往返
- **WHEN** 投影被序列化写入 `sched:meta:<id>` 再反序列化读回
- **THEN** 两个透传 key MUST 在读回后仍然存在且保持布尔语义
- **THEN** 验收 MUST 覆盖这一跳，MUST NOT 只断言 `filterSchedulerExtra` 的返回值

#### Scenario: 非透传账号不受影响
- **WHEN** 账号未设置任何透传开关，且 `model_mapping` 不含请求模型
- **THEN** 该账号 MUST 仍被判为不支持该模型（本变更 MUST NOT 放宽白名单语义）

### Requirement: 白名单的其余条目不得在本变更中改动
`filterSchedulerExtra` 的 key 列表 SHALL 只增加上述两个 key。本变更 MUST NOT 增删任何其它 key，
**包括本仓库相对上游缺失的 `codex_fingerprint_mode` / `codex_fingerprint_seed`**。

依据：那两个 key 属指纹收敛功能，投影是否需要它们取决于选号阶段有没有依赖指纹的判定；本批没有
该缺陷证据，补齐等于在无证据的情况下改动调度输入。

#### Scenario: 指纹字段仍被裁掉
- **WHEN** 账号 `Extra` 带 `codex_fingerprint_mode`
- **THEN** 投影中该字段 MUST 仍被裁掉（与本变更前一致）
