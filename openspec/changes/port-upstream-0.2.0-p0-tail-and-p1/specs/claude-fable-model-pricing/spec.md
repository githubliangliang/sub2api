## ADDED Requirements

### Requirement: Fable 系列必须有兜底定价
兜底定价表 SHALL 包含 `claude-fable-5` 与 `claude-fable-5-1` 两条条目。两者的输入 / 输出 /
缓存写入（5m 与 1h）单价相同，`claude-fable-5-1` 的缓存读取单价 SHALL 低于 `claude-fable-5`。
两条都 SHALL 声明支持缓存明细拆分。

依据：本仓库此前完全没有 Fable 兜底价（`internal/service/billing_service.go` grep `fable` 零命中），
Fable 计费只靠 LiteLLM 目录；目录缺条目时系列匹配（opus / sonnet / …）全都不命中。

#### Scenario: 目录缺 Fable 条目
- **WHEN** 价格目录里没有请求的 Fable 模型
- **THEN** 计费 MUST 落到兜底价，MUST NOT 按 0 计、MUST NOT 落到其它系列的价格

#### Scenario: 目录有 Fable 条目
- **WHEN** 价格目录里有该模型
- **THEN** MUST 优先使用目录价（兜底只在目录缺失时生效）

#### Scenario: 缓存读取单价差异
- **WHEN** 分别按 `claude-fable-5` 与 `claude-fable-5-1` 计算同样的缓存读取 token
- **THEN** 5.1 的成本 MUST 更低

### Requirement: 兜底系列匹配的判定顺序必须先 fable、再 5.1 先于 5
`getFallbackPricing` 的系列匹配 SHALL 在 opus 判定**之前**判 fable；在 fable 内部，
`fable-5-1` 系列 SHALL 在 `fable-5` 之前判定。

依据：与本仓库既有的「`opus-5` 必须先于裸 `5`」是同一类顺序陷阱。

#### Scenario: `claude-fable-5-1` 及其别名
- **WHEN** 模型名含 `fable-5-1` / `fable-5.1` / `fable5.1` / `fable51` 任一形态
- **THEN** MUST 命中 5.1 的兜底价，MUST NOT 命中 5 的

#### Scenario: `claude-fable-5`
- **WHEN** 模型名含 `fable-5` / `fable5` 但不属于上一个 Scenario 的形态
- **THEN** MUST 命中 5 的兜底价

#### Scenario: 既有系列不受影响
- **WHEN** 模型名是 opus / sonnet / haiku / gemini 任一系列
- **THEN** 命中的兜底价 MUST 与本变更前完全一致

### Requirement: `claude-fable-5-1` 必须完成模型面登记
`claude-fable-5-1` SHALL 出现在 Claude 默认模型列表、Antigravity 默认模型映射、Bedrock 默认模型映射，
以及前端的模型白名单与 anthropic / antigravity / bedrock 三处预设映射中。前后端两侧的清单 MUST 一致。

#### Scenario: Claude Code 客户端拉模型列表
- **WHEN** 客户端请求默认模型列表
- **THEN** 返回结果 MUST 含 `claude-fable-5-1`

#### Scenario: Bedrock 映射
- **WHEN** 请求 `claude-fable-5-1` 且账号是 Bedrock
- **THEN** MUST 映射到对应的 Bedrock 模型标识

#### Scenario: 前后端清单一致
- **WHEN** 比对 `DefaultAntigravityModelMapping` 与前端 `antigravityDefaultMappings`
- **THEN** 两侧 MUST 都含 `claude-fable-5-1`
