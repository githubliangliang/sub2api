## ADDED Requirements

### Requirement: OpenAI OAuth 账号的模型同步必须走 Codex manifest
OAuth 订阅不暴露公共 Platform API 的模型列表端点，系统 SHALL 在为 OpenAI OAuth 账号同步上游模型时
改用 ChatGPT 的 Codex 模型 manifest，并携带该账号的 Codex 出站身份（User-Agent / originator /
版本号同源）。MUST NOT 按 API Key 账号的方式请求 Platform API，否则管理台的同步按钮必然失败。

#### Scenario: 对 OAuth 账号点击同步
- **WHEN** 管理员对 OpenAI OAuth 账号触发上游模型同步
- **THEN** 请求 MUST 指向 Codex 模型 manifest
- **THEN** 请求 MUST 携带与该账号一致的 Codex 客户端版本号
- **THEN** 同步 MUST 能返回模型列表而不是不支持错误

#### Scenario: 对 API Key 账号点击同步
- **WHEN** 管理员对 OpenAI API Key 账号触发同步
- **THEN** 行为 MUST 与移植前一致（走 Platform API 模型列表）

#### Scenario: 既非 OAuth 也非 API Key 的账号类型
- **WHEN** 账号类型不受支持
- **THEN** MUST 返回明确的「不支持该账号类型」错误

#### Scenario: 规范客户端版本号的取值
- **WHEN** 构造 manifest 请求需要当前生效的 Codex 客户端版本号
- **THEN** 该版本号 MUST 与出站规范身份同源（面板配置 → 自动同步值 → 内置常量）
- **THEN** MUST NOT 另立一套版本号来源

### Requirement: Google One OAuth 账号只暴露该渠道能服务的模型
消费级 Google One OAuth 走的是旧 Gemini CLI / Code Assist 通道，系统 SHALL 只向这类账号暴露保守
模型集，MUST NOT 展示该通道无法服务的较新模型。模型白名单 MUST 每次返回独立副本，
调用方 MUST NOT 能改动包级目录。

#### Scenario: Google One OAuth 账号的可用模型
- **WHEN** 管理台查询 Google One OAuth 账号的可用模型
- **THEN** 返回集合 MUST 限于该通道可服务的保守模型
- **THEN** MUST NOT 包含该通道无法服务的较新代际或图像模型

#### Scenario: 其他 Gemini OAuth 账号
- **WHEN** 账号是非 Google One 的 Gemini OAuth
- **THEN** MUST 返回既有默认模型集，行为与移植前一致

#### Scenario: Gemini API Key 账号
- **WHEN** 账号是 Gemini API Key 类型
- **THEN** 行为 MUST 与移植前一致

#### Scenario: 白名单副本隔离
- **WHEN** 调用方取得模型白名单映射并修改它
- **THEN** 包级模型目录 MUST 不受影响
- **THEN** 后续调用 MUST 仍返回完整白名单
