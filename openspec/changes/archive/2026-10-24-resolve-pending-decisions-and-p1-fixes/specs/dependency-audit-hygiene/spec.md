## ADDED Requirements

### Requirement: 依赖审计例外清单不得含过期条目
系统 SHALL 保证依赖审计例外清单中每一条的 `expires_on` 均在未来。审计门禁对过期条目与对未登记的
公告一样判为失败，因此过期条目会让门禁**长期常红**，等价于关闭了依赖风险信号。
清单变更 MUST NOT 通过关闭审计、忽略整个包或删除条目而不作说明来「修绿」。

#### Scenario: 存在过期条目
- **WHEN** 任一例外条目的 `expires_on` 早于当日
- **THEN** 审计门禁 MUST 判为失败
- **THEN** 该条目 MUST 被重新核实后续期，或因依赖升级而删除

#### Scenario: 全部条目均在有效期内
- **WHEN** 所有条目的 `expires_on` 均在未来，且当前公告都被条目覆盖
- **THEN** 审计门禁 MUST 通过
- **THEN** 通过 MUST 来自条目覆盖，而不是审计步骤被跳过

#### Scenario: 出现未登记的新公告
- **WHEN** 依赖审计报出清单里没有的高危公告
- **THEN** 门禁 MUST 失败
- **THEN** MUST 通过升级消除或新增带期限的例外条目来处理

### Requirement: 例外条目必须携带可复核的理由与责任归属
每条例外 MUST 包含包名、公告编号、严重级别、「本仓库用法为何不受影响」的理由、缓解措施、
到期日与责任人。续期 MUST 重新核实理由是否仍然成立，MUST NOT 只把日期往后推。

#### Scenario: 续期一条仍不受影响的公告
- **WHEN** 重新核实后确认本仓库用法仍不受该公告影响
- **THEN** 条目 MUST 更新 `expires_on`
- **THEN** `reason` MUST 反映本次核实的结论（含该依赖是直接依赖还是传递依赖）

#### Scenario: 续期时发现理由已不成立
- **WHEN** 重新核实发现本仓库已经用上了受影响的用法
- **THEN** MUST NOT 续期
- **THEN** MUST 改为升级消除，或把该风险升级为待处理项

#### Scenario: 到期日的选取
- **WHEN** 为条目设定新的 `expires_on`
- **THEN** MUST 早于下一次计划性依赖维护
- **THEN** MUST NOT 设成远期日期使其实际上永不复核
