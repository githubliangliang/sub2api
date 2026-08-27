## ADDED Requirements

### Requirement: 邮箱换绑必须按收件箱身份而不是字面地址查重
系统 SHALL 在发送换绑验证码与提交换绑两个入口都执行收件箱级查重：先按归一化后的字面地址查，
再按 provider alias 规则（Gmail 家族点号、`+suffix` 等）查。任一命中且归属**其他用户**时
MUST 拒绝并返回邮箱已存在的既有错误语义。查重仓储不可用 MUST 按服务不可用处理，
MUST NOT 当作「未被占用」放行。

#### Scenario: alias 变体已被他人占用
- **WHEN** 用户 B 已绑定 `a.b@gmail.com`
- **AND** 用户 A 申请把主邮箱换成 `ab@gmail.com`
- **THEN** 发码入口 MUST 拒绝并返回邮箱已存在
- **THEN** 提交入口 MUST 拒绝并返回邮箱已存在
- **THEN** MUST NOT 出现两条用户记录指向同一个收件箱

#### Scenario: 用户更换自己的 alias 写法
- **WHEN** 用户 A 当前主邮箱为 `a.b@gmail.com`
- **AND** 用户 A 申请换成 `ab@gmail.com`
- **THEN** MUST 放行
- **THEN** 换绑完成后主邮箱 MUST 为新写法

#### Scenario: 字面地址已被他人占用
- **WHEN** 目标地址与另一用户的主邮箱完全相同
- **THEN** MUST 拒绝，错误语义 MUST 与移植前一致

#### Scenario: 查重依赖不可用
- **WHEN** alias 查重查询返回错误
- **THEN** MUST 返回服务不可用
- **THEN** MUST NOT 继续写入

### Requirement: 换绑写入必须在锁内复查归属，关闭并发窗口
系统 SHALL 在调用方事务内完成主邮箱替换：先对「归一化字面地址」与「alias 收件箱身份」两个
维度取仓储级 key 锁，再复查该收件箱是否已归属其他用户，最后才写入新邮箱与新密码哈希。
换绑 MUST 要求存在事务上下文，无事务时 MUST 拒绝执行。

#### Scenario: 两个并发请求争抢同一收件箱
- **WHEN** 两个用户几乎同时提交换绑到互为 alias 的两个地址
- **AND** 两者的前置查重都看到「未被占用」
- **THEN** 最多一个请求 MUST 成功
- **THEN** 另一个 MUST 收到邮箱已存在
- **THEN** 最终数据库中 MUST 只有一条记录指向该收件箱

#### Scenario: 锁内复查发现已被占用
- **WHEN** 取锁后复查发现该收件箱已归属其他用户
- **THEN** MUST 在写入前返回邮箱已存在
- **THEN** MUST NOT 修改任何用户记录

#### Scenario: 缺少事务上下文
- **WHEN** 调用方未提供事务
- **THEN** MUST 返回错误
- **THEN** MUST NOT 退化成无锁的直接写入

#### Scenario: 唯一约束仍然兜底
- **WHEN** 写入阶段仍触发数据库唯一约束冲突
- **THEN** MUST 转换为邮箱已存在的应用层错误
- **THEN** MUST NOT 向客户端暴露底层约束名或 SQL 细节

### Requirement: 单节点部署 MUST NOT 以「跑 SQLite」为由退化守卫
本 fork 的仓储级 key 锁实现为纯进程内锁，单进程语义已足够满足上述并发要求。实现 MUST
保留取锁—复查—写入的完整顺序，MUST NOT 因为不存在跨实例 advisory lock 而删掉锁或只保留前置查重。

#### Scenario: 以单机为由跳过锁
- **WHEN** 实现只保留前置查重、去掉写入侧的锁与复查
- **THEN** 该实现 MUST 被视为不满足本 capability
- **THEN** 并发换绑测试 MUST 失败
