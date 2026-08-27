## ADDED Requirements

本 capability 覆盖四项管理台修复。共同失效模式：**功能在后端可用，但管理台不让你用对**。

### Requirement: 批量代理解析必须支持带方括号的 IPv6 主机
系统 SHALL 在批量代理 URL 解析中接受 `[2001:db8::1]:8080` 形式的 IPv6 字面量，
并在提交前去掉方括号（后端会自行重新加括号拼接 host:port）。
用户名密码、域名与 IPv4 形式 MUST 继续可用。

#### Scenario: 带方括号的 IPv6 代理
- **WHEN** 批量输入 `socks5://[2001:db8::1]:1080`
- **THEN** MUST 解析成功
- **THEN** 提交的 host MUST 为不带方括号的地址，端口 MUST 为 `1080`

#### Scenario: 带认证的 IPv6 代理
- **WHEN** 批量输入含用户名密码且主机为方括号 IPv6 的代理 URL
- **THEN** MUST 正确拆出协议、用户名、密码、主机与端口

#### Scenario: 域名与 IPv4
- **WHEN** 批量输入域名或 IPv4 形式的代理 URL
- **THEN** 行为 MUST 与移植前一致

#### Scenario: 非法输入
- **WHEN** 输入缺少端口或协议不受支持
- **THEN** MUST 按既有方式提示解析失败

### Requirement: 用户并发数 0 必须表示不限制
管理台的用户编辑表单 SHALL 接受 `0` 作为「不限制并发」，与网关侧「上限小于等于 0 即不限制」
以及批量改限额的语义一致。校验 MUST 只拒绝负数与非整数，MUST NOT 要求至少为 1。
界面 MUST 通过占位符与提示文案说明 `0 = 不限制`，中英文案 MUST 同步。

#### Scenario: 填 0 保存
- **WHEN** 管理员把用户并发数填为 `0` 并保存
- **THEN** MUST 保存成功
- **THEN** 该用户 MUST 不受并发限制

#### Scenario: 填负数
- **WHEN** 管理员填入负数
- **THEN** MUST 提示并发数不能为负并说明 0 表示不限制
- **THEN** MUST NOT 提交

#### Scenario: 填非整数
- **WHEN** 管理员填入非整数
- **THEN** MUST 拒绝并提示

#### Scenario: 填正整数
- **WHEN** 管理员填入正整数
- **THEN** 行为 MUST 与移植前一致

### Requirement: 账号列表默认展示优先级列
账号列表的默认隐藏列集合 MUST NOT 包含优先级列——优先级直接决定调度顺序，
默认隐藏会让管理员在排查调度问题时看不到关键字段。

#### Scenario: 首次打开账号列表
- **WHEN** 管理员在没有个人列偏好的情况下打开账号列表
- **THEN** 优先级列 MUST 可见

#### Scenario: 管理员已保存列偏好
- **WHEN** 管理员此前显式隐藏了优先级列
- **THEN** MUST 尊重该偏好
- **THEN** MUST NOT 因默认值变化而覆盖用户选择

### Requirement: 运维错误详情必须能返回列表并保留筛选状态
从错误列表进入详情后，系统 SHALL 提供「返回列表」入口，返回时 MUST 保留进入前的筛选条件与
列表上下文。详情页 MUST 分区展示诊断载荷（客户端响应、上游消息、上游详情、上游事件）
与上游状态码、根因等字段，且 MUST 只展示有内容的分区。中英文案 MUST 同步。

#### Scenario: 从列表进入详情再返回
- **WHEN** 管理员从筛选后的错误列表进入某条错误详情，再点击返回列表
- **THEN** MUST 回到列表
- **THEN** 原筛选条件 MUST 保持不变

#### Scenario: 直接打开详情
- **WHEN** 详情不是从列表进入的（没有返回目标）
- **THEN** MUST NOT 展示「返回列表」入口

#### Scenario: 诊断载荷分区
- **WHEN** 错误记录携带上游消息、上游详情或上游事件
- **THEN** 详情 MUST 按分区分别展示
- **THEN** 内容为空的分区 MUST NOT 占位展示

#### Scenario: 文案完整性
- **WHEN** 界面语言切换为中文或英文
- **THEN** 新增的所有文案键 MUST 均有对应翻译，MUST NOT 出现裸 key
