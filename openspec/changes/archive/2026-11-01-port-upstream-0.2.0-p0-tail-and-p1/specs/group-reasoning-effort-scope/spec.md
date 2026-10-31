## ADDED Requirements

### Requirement: 推理档位映射可按模型限定作用域
分组的推理档位映射 SHALL 支持限定生效模型；未限定模型的映射 SHALL 对该分组的全部模型生效
（与本变更前等价）。映射 MUST 先按模型匹配、再应用上限，顺序不变。

#### Scenario: 映射限定了模型且请求命中
- **WHEN** 某条映射限定模型 M，请求的模型是 M
- **THEN** 该映射 MUST 生效

#### Scenario: 映射限定了模型但请求未命中
- **WHEN** 某条映射限定模型 M，请求的模型不是 M
- **THEN** 该映射 MUST NOT 生效

#### Scenario: 未限定模型的映射
- **WHEN** 某条映射没有限定模型
- **THEN** 对该分组任意模型 MUST 生效（存量配置的语义 MUST NOT 变化）

#### Scenario: 多条映射同时可命中
- **WHEN** 一条限定模型的映射与一条未限定的映射都可命中
- **THEN** 命中优先级 MUST 是确定的且有用例固定（限定的优先）

### Requirement: 超限档位的处置可按分组配置为拒绝或降级
分组 SHALL 新增「推理档位超过上限时的处置」配置，取值为**拒绝**或**降级**。降级 SHALL 等价于
本变更前的静默钳制行为。

#### Scenario: 配置为降级
- **WHEN** 客户端请求的档位高于分组上限，分组配置为降级
- **THEN** 请求 MUST 被钳制到上限并正常转发
- **THEN** 行为 MUST 与本变更前完全一致

#### Scenario: 配置为拒绝
- **WHEN** 客户端请求的档位高于分组上限，分组配置为拒绝
- **THEN** 请求 MUST 被拒绝
- **THEN** 错误信息 MUST 指明请求档位与分组上限

#### Scenario: 请求档位未超限
- **WHEN** 请求档位不高于上限
- **THEN** 两种配置下的行为 MUST 相同且与本变更前一致

### Requirement: 存量分组的行为不得因升级而改变
新列的默认值 SHALL 落在与本变更前等价的那一档（降级）。迁移后未经管理员显式修改的分组 MUST
表现出与升级前一致的行为。

依据：0.1.179 的长上下文门控 AND → OR 就是因为默认值使既有部署行为改变而成为 breaking change。
本批不得重复。

#### Scenario: 迁移后未改配置的分组
- **WHEN** 分组在迁移后未被修改，客户端请求档位超限
- **THEN** MUST 走降级，MUST NOT 返回 4xx

#### Scenario: 分组复制
- **WHEN** 复制一个已配置该策略的分组
- **THEN** 新分组 MUST 继承同样的取值

### Requirement: 策略必须在全部发送路径上一致生效
该策略 SHALL 在 Responses、Chat Completions 回退桥、WS forwarder 与 WS v2 透传适配四条路径上
产生一致结果。任一路径 MUST NOT 绕过策略。

#### Scenario: 四条路径同一请求
- **WHEN** 同一个超限请求分别经四条路径发出
- **THEN** 处置结果 MUST 一致（同为降级或同为拒绝）

#### Scenario: 认证投影携带策略
- **WHEN** 请求经 api_key 认证缓存进入
- **THEN** 投影 MUST 携带该策略字段（MUST NOT 因投影裁字段而失效）

### Requirement: 管理面必须可读可写该策略
admin API 的分组请求与响应 SHALL 携带这两项配置；前端分组表单 SHALL 可编辑。

#### Scenario: 读取分组
- **WHEN** 管理端读取分组详情
- **THEN** 响应 MUST 含超限处置与按模型映射两项

#### Scenario: 写入非法取值
- **WHEN** 提交一个不在允许集合内的超限处置取值
- **THEN** MUST 被拒绝并返回可读的错误
