## ADDED Requirements

### Requirement: 引导类请求缺 `call_id` 时必须被接受
delegation bootstrap 与 scheduled-automation bootstrap 两类引导请求 SHALL 在没有 `call_id` 的情况下
被接受并正常转发。普通工具调用请求对 `call_id` 的既有校验 MUST NOT 放宽。

依据：这两类引导请求本来不携带 `call_id`，网关按普通工具调用校验会直接拒掉，客户端无法完成引导。

#### Scenario: delegation bootstrap 无 call_id
- **WHEN** 收到 delegation bootstrap 请求且其中没有 `call_id`
- **THEN** 请求 MUST 被接受
- **THEN** 转发给上游的请求体 MUST NOT 被补写一个伪造的 `call_id`

#### Scenario: scheduled automation bootstrap 无 call_id
- **WHEN** 收到 scheduled-automation bootstrap 请求且其中没有 `call_id`
- **THEN** 请求 MUST 被接受

#### Scenario: 引导请求带 call_id
- **WHEN** 引导请求确实带了 `call_id`
- **THEN** 该值 MUST 原样透传，MUST NOT 被剥除

#### Scenario: 普通工具调用缺 call_id
- **WHEN** 一个非引导类的工具调用请求缺 `call_id`
- **THEN** 既有校验 MUST 继续生效（本变更 MUST NOT 让这类请求变得可接受）

#### Scenario: 引导判定不得依赖客户端自述
- **WHEN** 请求声称自己是引导请求但形态不符
- **THEN** 判定 MUST 基于请求形态本身，MUST NOT 仅凭客户端提供的标记放宽校验
