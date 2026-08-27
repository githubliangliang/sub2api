# port-upstream-0.1.183-p0-fixes

把上游 v0.1.181 / v0.1.182 / v0.1.183 三个纯 bugfix 版里 **12 项已核实本仓库患同一缺陷**的修复移植进来：两项直接影响计费与缓存命中率（Anthropic cache_creation 重复计费、Codex 粘性会话漂移），其余是上游请求兼容性、账号冷却精度、邮箱换绑并发安全和充值余额可见性。不引入新功能、不新增迁移、不改动数据库 schema。

阅读顺序：`proposal.md` → `source-baseline.md` → `source-feature-map.md` → `design.md` → 六个 `specs/*/spec.md` → `tasks.md` → `verification.md`。

逐条的 patch site、行号和上游 diff 说明在 [`docs/upstream-sync/PORTING-0.1.183.md`](../../../docs/upstream-sync/PORTING-0.1.183.md) 第 3 节，本 change 不复制那份内容，只定义**移植后必须成立的行为**与验收证据。
