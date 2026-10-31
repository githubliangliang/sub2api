# 网关最小输出 Token 配置

## ADDED Requirements

### Requirement: 网关转换下限具有明确默认值和配置校验

系统 SHALL 提供 `gateway.min_output_tokens`，默认128，支持既有YAML和`GATEWAY_MIN_OUTPUT_TOKENS`配置链，合法值为至少16的整数。显式非法值 MUST 阻止配置加载并说明字段原因。个人模板和完整示例 SHALL 记录相同默认与下限；这表示转换的请求上限下限，不保证模型实际输出长度。

#### Scenario: 默认与自定义

- **WHEN** 未配置该字段，或通过YAML/环境变量分别设置16、128和256
- **THEN** SHALL 分别得到128和对应配置值，遵守项目既有配置优先级

#### Scenario: 非法配置

- **WHEN** 显式配置15、0、负数、非整数或超出可解析整数范围
- **THEN** SHALL 返回指向gateway.min_output_tokens的配置错误，不静默替换错误值

### Requirement: 已有协议转换一致使用配置下限

Chat→Responses、Anthropic→Responses、Anthropic→Chat等既有相关转换 SHALL 对有效的显式正数输出上限使用 `max(请求值,配置下限)`，保留原字段优先级和未指定/非正数的省略或校验行为。

#### Scenario: 小于、等于和大于下限

- **WHEN** 下限为16，请求的有效上限分别为1、16、100
- **THEN** 转换结果 SHALL 分别为16、16、100

#### Scenario: 缺省和多种字段

- **WHEN** 请求没有输出上限，或同时带max_tokens与max_completion_tokens
- **THEN** 前者 SHALL 保持原省略/默认规则，后者保持原优先级再应用下限
- **THEN** 非法请求值仍按原输入校验处理

### Requirement: 配置传递覆盖已有转换并保留入口兼容边界

所有本地已有消费者需要转换上述字段时 SHALL 使用所属服务的配置，包括已有HTTP/SSE消息桥。新实现 MUST NOT 依赖可变包级下限；旧公共转换调用及nil/未初始化的程序内配置保持128兼容默认。缺失的legacy messages→Responses入口不作为本配置的前置新增。

#### Scenario: 不同已有转发路径一致

- **WHEN** 同一Anthropic或Chat请求分别经本地已有、会执行上述协议转换的转发服务，输出上限低于配置值
- **THEN** SHALL 使用各服务实际配置，不因个别调用点漏接而回退到硬编码128

#### Scenario: 服务实例隔离与旧调用

- **WHEN** 两个服务配置分别为16和256，同时处理请求；旧转换函数或测试夹具未提供配置
- **THEN** 两服务 SHALL 各用自己的值，旧调用安全使用128，不发生共享配置污染

### Requirement: 原生请求与提供商限制保持原语义

系统 SHALL 只在既有兼容转换中应用下限。原生Responses input和显式max_output_tokens、compact、原样透传以及原生Claude探针不得新增统一抬高；Ollama等既有提供商上限仍按其原优先级约束最终请求。

#### Scenario: 原生Responses、compact与多轮WS

- **WHEN** 客户端发送有效原生Responses请求、compact请求，或WS首帧/后续response.create
- **THEN** SHALL 保持原输入与显式输出字段，不仅因为新配置而改写

#### Scenario: 提供商上限和原生探针

- **WHEN** 转换路径还适用既有provider cap，或原生Claude Code发送max_tokens=1探针
- **THEN** SHALL 保持提供商上限及原生探针的既有行为，不为本配置新增平台或改变实际Token记录
