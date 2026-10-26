# OpenAI API Key 搜索模拟

## ADDED Requirements

### Requirement: 搜索模拟受全局开关和账号渠道策略约束

OpenAI API Key 的 `/v1/messages` SHALL 在现有搜索管理器可用、全局启用、请求恰有一个受支持搜索工具且账号/渠道允许时进入模拟。账号支持default/enabled/disabled：enabled覆盖渠道，disabled关闭，default继承当前分组渠道的openai标志。旧true转enabled、false转default；全局关闭始终不启用模拟。

#### Scenario: 强制开启与渠道继承

- **WHEN** 全局可用且纯搜索请求的账号为enabled，或为default且当前渠道openai标志开启
- **THEN** SHALL 使用既有搜索服务模拟该请求

#### Scenario: 全局、账号或渠道未允许

- **WHEN** 管理器/配置服务缺失、全局关闭、账号disabled，或default没有允许的渠道配置
- **THEN** SHALL 沿原网关路径处理，不调用搜索提供商或产生空指针异常

#### Scenario: 配置回显与兼容

- **WHEN** 管理员创建/编辑OpenAI API Key，或编辑渠道OpenAI搜索开关后保存重开
- **THEN** SHALL 正确保存/回显账号三态与渠道布尔值，兼容旧bool
- **THEN** 切换账号或隐藏配置区域不得把另一账号的搜索设置写入当前账号

### Requirement: 模拟请求只调用既有搜索服务并输出Anthropic协议

被拦截请求 SHALL 从最后一条用户消息提取查询，调用已有搜索提供商/代理选择逻辑，并按stream输出完整Anthropic JSON或SSE，包括对应工具结果及终态。模拟 MUST 位于转换为Responses/Chat之前，不再调用所选LLM上游。

#### Scenario: 非流式和流式成功

- **WHEN** 搜索提供商桩返回结果，客户端分别请求stream=false/true
- **THEN** SHALL 返回可解析且有唯一成功终态的Anthropic消息/事件，工具引用成对
- **THEN** LLM上游调用次数 SHALL 为0，响应模型保持请求侧语义

#### Scenario: 最近用户消息

- **WHEN** 请求含多轮消息及最后一条用户的文本内容
- **THEN** SHALL 只用该查询执行搜索，沿用已有文本提取规则

### Requirement: 模拟范围不覆盖原生搜索和混合工具

本新增路径 SHALL 限于OpenAI API Key Messages入口。OAuth、SetupToken/影子账号、其他平台及入口、非搜索和混合工具请求保持各自既有路径；Anthropic原有模拟继续工作。

#### Scenario: OAuth或混合工具

- **WHEN** OpenAI OAuth/SetupToken携带搜索工具，或API Key请求同时声明搜索与其他工具
- **THEN** MUST NOT 被本新增路径拦截，不剥离其他工具或改变原生搜索

#### Scenario: Anthropic既有路径

- **WHEN** 合法Anthropic API Key按原配置请求搜索模拟
- **THEN** SHALL 保持原有账号/渠道三态、提供商和响应语义

### Requirement: 搜索错误和取消保持网关资源与记录一致性

空查询、代理/提供商错误和取消 SHALL 返回符合当前协议的失败或按现有规则换号。输出后不得重放或追加伪成功。所有退出路径 MUST 释放现有并发/串行资源，并沿用搜索模拟的真实用量/提供商额度语义与网关归属记录，不将结果文本长度冒充LLM Token。

#### Scenario: 输出前失败和输出后写失败

- **WHEN** 代理不可用/搜索失败发生在输出前，或客户端在SSE输出中断开
- **THEN** SHALL 分别遵守输出前换号预算与输出后不重放规则
- **THEN** 不泄漏槽位、重复终态、搜索调用或重复用量记录

#### Scenario: simple mode和使用量归属

- **WHEN** SQLite/miniredis/simple mode下配置全局提供商与账号enabled并执行模拟
- **THEN** SHALL 无需启用隐藏渠道/运维菜单即可工作
- **THEN** 使用量记录的用户、API Key、分组、账号和模型 SHALL 属于本次请求，保持既有模拟Token口径及去重语义
