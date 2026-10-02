## ADDED Requirements

### Requirement: F01 Grok CLI身份匹配官方交互版本

Grok CLI代理请求 SHALL 使用1.0.46默认身份与运行平台UA；包级最低版本1.0.13、最终transport最低preferred pin保持各自约束。额度探测与网关身份一致，回退官方API清除CLI专用头。

#### Scenario: 默认请求与旧覆盖值

- **WHEN** 调用CLI代理且未设版本或配置低于允许下限的值
- **THEN** 发出1.0.46、grok-pager、interactive及authenticate-response；包级和transport分别验证下限

#### Scenario: 有效覆盖与平台UA

- **WHEN** 使用符合相应层版本约束的覆盖值
- **THEN** UA与版本一致；amd64/arm64和darwin按Rust格式输出

#### Scenario: 官方API边界

- **WHEN** 独立访问api.x.ai或CLI访问拒绝后回退官方API
- **THEN** 独立API路径不被CLI-host helper覆盖；回退请求不携带新增CLI mode/authenticate头

### Requirement: F02 Antigravity客户端错误隐藏账号池身份

既有Antigravity Gemini出口 SHALL 将非failover上游错误规范化为JSON，仅保留code/status/message，清理项目引用、服务账号邮箱与敏感查询参数，details不返回客户端。

#### Scenario: JSON身份字段

- **WHEN** ForwardGemini上游错误正文包含projects引用、服务账号邮箱和details.metadata
- **THEN** 实际客户端响应中不含这些身份内容；HTTP状态保持，ops与调度错误处理保留

#### Scenario: 非JSON或空错误

- **WHEN** 上游错误为普通文本、空正文或无法提取message
- **THEN** 客户端仍获得合法Gemini JSON及对应状态；已知身份形态清理，空内容使用通用错误

#### Scenario: 正常响应与failover

- **WHEN** 上游成功或错误符合既有failover条件
- **THEN** 保留成功用量及原有failover语义，不因脱敏增加重放

### Requirement: F03 自定义错误码说明符合执行语义

中英文说明 SHALL 描述自定义错误码筛选常规账号错误处理，不承诺未选中错误统一返回500，也不宣称其决定retry/failover。

#### Scenario: 查看设置说明

- **WHEN** 用户以中文或英文查看自定义错误码设置
- **THEN** 说明空列表不筛选、重试/切号由独立网关逻辑决定、最终状态取决于路径与透传规则

#### Scenario: 保持配置行为

- **WHEN** 修改此次说明并使用原有配置
- **THEN** 不改变错误码默认值、调度与重试实现
