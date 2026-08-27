## ADDED Requirements

### Requirement: 跨标签页 token 刷新 MUST NOT 因采信未过期快照而空转
多标签页共享刷新锁时，等待方 SHALL 只在对端确实写入了**新的** token 对时才采信结果。
MUST NOT 因为「本地记录的 token 尚未接近过期」就把**当前快照**当成刷新成功返回——
那会让调用方拿到同一个已被上游拒绝的 token，再次触发刷新，形成锁获取与放弃的空转循环并持续占用 CPU。

#### Scenario: 对端完成刷新
- **WHEN** 等待方发现存储中的 token 对与自己的快照不同
- **THEN** MUST 采信该新 token 对
- **THEN** 调用方 MUST 用新 token 重试原请求

#### Scenario: 对端未刷新且本地 token 尚未接近过期
- **WHEN** 存储中的 token 对与快照一致，且过期时间距现在仍有较长缓冲
- **THEN** MUST 返回「无可用结果」
- **THEN** MUST NOT 把当前快照当作刷新结果返回
- **THEN** MUST NOT 进入反复获取锁又立即放弃的循环

#### Scenario: 触发刷新的请求已明确失败
- **WHEN** 刷新是由一个使用了特定 access token 的 401 触发
- **THEN** 该 token 本身 MUST NOT 被当作有效结果返回
