## ADDED Requirements

### Requirement: 流式终端输出必须按上游上报的 item 重建
系统 SHALL 在流式 Responses 转发中记录每个 `response.output_item.done` 事件携带的原始 item
（按 `output_index` 归档、逐字节保留），并在规范化终端输出时优先使用这些 item，
而不是只依赖 delta 累积器。累积器只建模「一个 reasoning、一个 message、N 个 function call」，
无法保留 item 身份、逐 item 的状态/阶段、顺序，以及它不认识的 item 类型。

#### Scenario: 上游上报了累积器无法建模的 item 类型
- **WHEN** 流中出现累积器不认识的 item 类型的 `output_item.done`
- **THEN** 终端输出 MUST 包含该 item
- **THEN** 该 item 的字段 MUST 逐字节保留，厂商扩展字段 MUST NOT 丢失

#### Scenario: 多个同类型 item
- **WHEN** 流中出现多个同类型 item（例如两段 message）
- **THEN** 终端输出 MUST 按 `output_index` 保留全部 item 及其顺序
- **THEN** MUST NOT 被折叠成一个

#### Scenario: 逐 item 的状态与身份
- **WHEN** 上报的 item 携带自身的 id 与 status
- **THEN** 终端输出中的对应 item MUST 保留原 id 与 status

#### Scenario: 上游未上报 done item
- **WHEN** 流中没有任何 `output_item.done` 事件
- **THEN** MUST 回落到既有的 delta 累积重建路径
- **THEN** 行为 MUST 与移植前一致

#### Scenario: 图像生成输出
- **WHEN** 流中包含图像生成输出
- **THEN** 既有的图像输出收集与去重 MUST 继续生效

### Requirement: 流式工具调用的 arguments delta MUST NOT 携带空工具名
系统 SHALL 在序列化流式工具调用增量时省略空的函数名字段。客户端普遍按「字段存在即覆盖」合并
增量，若后续仅含 arguments 的增量带上空名，会覆盖首个增量里的工具名，
最终客户端去调用一个名为空字符串的工具。

#### Scenario: 仅含 arguments 的后续增量
- **WHEN** 首个增量携带工具名，后续增量只携带 arguments 片段
- **THEN** 后续增量 MUST NOT 包含函数名字段
- **THEN** 客户端合并后的工具名 MUST 仍为首个增量给出的名字

#### Scenario: 首个增量
- **WHEN** 增量携带非空工具名
- **THEN** 该字段 MUST 正常序列化输出

#### Scenario: 非流式响应
- **WHEN** 响应为非流式且函数名为空
- **THEN** 省略该字段 MUST NOT 破坏既有客户端解析
