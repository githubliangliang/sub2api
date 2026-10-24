## ADDED Requirements

### Requirement: daily 端点地址必须指向官方域而不是 sandbox 域
系统 SHALL 把 Antigravity 的 daily base URL 指向官方 daily 域。MUST NOT 指向 `.sandbox.` 域——
带生产 OAuth token 的请求发到 sandbox 域会被上游拒绝，账号被 401「Invalid bearer token」/ 502
打入临时不可调度，且后台「测试连接」用的是生产端点，于是表现为「测试成功但网关 401」。

#### Scenario: 解析 daily 端点
- **WHEN** 需要 daily base URL
- **THEN** 该地址 MUST 为官方 daily 域
- **THEN** MUST NOT 包含 `.sandbox.` 片段

### Requirement: 付费账号默认走 daily 端点，其余账号保持生产端点
未显式配置转发端点时，系统 SHALL 按订阅层级选择 base URL：付费层级（`plan_type` 为 `pro` 或
`ultra`，大小写与空白不敏感）使用 daily 端点，其余账号继续使用生产端点。
显式环境变量配置 MUST 始终优先于层级推断。

#### Scenario: 付费账号且未显式配置端点
- **WHEN** 账号凭据中的 `plan_type` 为 `pro` 或 `ultra`，且未设置转发端点环境变量
- **THEN** 转发 MUST 使用 daily 端点

#### Scenario: 免费账号且未显式配置端点
- **WHEN** 账号没有付费层级标记
- **THEN** 转发 MUST 使用生产端点
- **THEN** MUST NOT 因层级推断把免费账号的 OAuth token 发到 daily 端点造成 401

#### Scenario: 显式配置了端点
- **WHEN** 转发端点环境变量设为 daily 或 sandbox
- **THEN** MUST 使用该配置，无论账号层级
- **THEN** 行为 MUST 与移植前一致

#### Scenario: 凭据缺失或层级字段类型异常
- **WHEN** 账号为空、凭据为空，或 `plan_type` 不是字符串
- **THEN** MUST 判定为非付费
- **THEN** MUST 使用生产端点

#### Scenario: 层级推断与端点地址的落地顺序
- **WHEN** 付费账号路由与 daily 端点地址修正分两次上线
- **THEN** 中间状态 MUST NOT 存在——付费账号会被路由到 sandbox 域并 401
- **THEN** 两项 MUST 同时生效
