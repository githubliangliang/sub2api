## ADDED Requirements

### Requirement: 普通工具参数完成与历史回放必须保持JSON完整性
系统 SHALL 在现有Chat兼容转换中阻止非法JSON的普通function-call被标记为完成；回放中的坏普通调用及其匹配输出 SHALL 配套处理。系统 MUST 保留合法参数、既有空参数默认和custom自由文本语义。

#### Scenario: 上游返回截断普通工具参数
- **WHEN** 非流响应或流终态的普通function-call参数是截断JSON
- **THEN** 客户端 MUST NOT 收到该工具的成功完成项
- **THEN** 已观测usage SHALL 按既有写入政策保留并去重

#### Scenario: 历史包含坏调用及对应输出
- **WHEN** 客户端回放非法普通function-call及匹配call_id输出
- **THEN** 出站Chat请求 MUST NOT 继续转发这对坏历史项
- **THEN** 其它合法对话与工具项 SHALL 保留

#### Scenario: 合法参数与custom输入
- **WHEN** 普通参数是有效JSON但finish_reason为length，或工具属于custom/freeform
- **THEN** 系统 MUST NOT 仅以length或custom输入不是JSON而拒绝该工具
- **THEN** 整体响应的既有incomplete和工具类型语义 SHALL 保留

### Requirement: 自定义工具别名还原必须尊重显式声明与命名空间归属
系统 SHALL 在既有Chat到Responses的buffer与SSE转换中还原无歧义namespace前缀custom别名；显式function与真实namespace子工具 MUST 优先，歧义名称 MUST NOT 通过猜测重写。

#### Scenario: 唯一custom前缀别名
- **WHEN** 已声明custom exec、namespace functions下的wait，且上游返回functions__exec
- **THEN** 输出 SHALL 为custom_tool_call、名称exec及原始custom input
- **THEN** 流added/done与最终output SHALL 使用一致的身份

#### Scenario: 别名与真实声明冲突
- **WHEN** 返回名称对应显式function、真实namespace子工具或多个可能custom归属
- **THEN** 系统 MUST 保留确定的显式/namespace归属或采用既有明确拒绝策略
- **THEN** 系统 MUST NOT 将其擅自还原为另一个custom工具

### Requirement: 最终工具声明必须清理失去前置的defer标记
系统 SHALL 依据最终声明列表中的真实内建tool_search清理孤儿defer_loading；同名普通function MUST NOT 被视为内建前置。该行为 SHALL 覆盖现有通用adapter与Grok请求出口。

#### Scenario: 内建tool_search已降级或移除
- **WHEN** 最终列表只剩普通function等工具且仍有defer_loading字段
- **THEN** 出站工具 SHALL 不再含孤儿defer_loading
- **THEN** 工具schema、选择与历史的其余既有语义 SHALL 保留

#### Scenario: 合法内建tool_search仍存在
- **WHEN** 最终声明仍包含真实内建tool_search
- **THEN** 合法defer_loading SHALL 保留
- **THEN** 系统 MUST NOT 改写共享账号元数据或目录缓存

### Requirement: Responses事件必须序列化零序号并保留事件顺序
系统 SHALL 在现有Responses序列化及本地合成事件中始终写出sequence_number，包括0。已知上游序号 SHALL 保留，本地有序事件 SHALL 递增。

#### Scenario: 第一帧或仅有heartbeat之后的失败
- **WHEN** 第一条数字事件的序号为0，或仅有heartbeat后合成失败
- **THEN** 事件 SHALL 显式包含sequence_number为0
- **THEN** 既有合法终止类型与错误字段 SHALL 保留

#### Scenario: compact和桥接合成事件
- **WHEN** compact生成item/terminal事件，或handler/WS桥合成错误或失败
- **THEN** 所有合成事件 SHALL 携带序号
- **THEN** 已编号流后续合成事件 SHALL 从实际writer/state继续，不统一重置为0

#### Scenario: 既有正值序号
- **WHEN** 事件已有正值序号或流已写出有序事件
- **THEN** 序列化 SHALL 保留现有值和顺序
- **THEN** 未改变的native透传语义与未知字段 SHALL 保留
