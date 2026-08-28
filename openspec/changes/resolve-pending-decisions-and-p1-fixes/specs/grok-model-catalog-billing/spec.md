## ADDED Requirements

本 capability 覆盖两件事：**修缺陷**（目录 / 价卡 / 别名，P1 第 2 项）与**策略选择**
（默认文本模型跟到 4.6，决策 9.1）。两者改同一批文件且有硬顺序依赖，见 `design.md` 决策 2。

### Requirement: 每个对客户端可见的 Grok 文本模型都必须有对应价卡
系统 SHALL 保证凡是能被请求到的 Grok 文本模型 ID 都命中它自己的官方价卡，
MUST NOT 落进「未知 Grok 文本族」的兜底价卡。`grok-3-mini` 与 `grok-3-mini-fast` MUST 各有
独立价卡；同时它们 MUST 从公开模型列表中移除（仍可作为别名被请求）。

#### Scenario: 请求 grok-3-mini
- **WHEN** 客户端请求 `grok-3-mini`
- **THEN** 计费 MUST 使用该模型自己的官方价卡（输入 `$0.30/MTok`、输出 `$0.50/MTok`、
  缓存读取 `$0.075/MTok`）
- **THEN** MUST NOT 使用 Grok 4.5 的价卡（`$2` / `$6`），后者相当于输入超收约 6.7×、输出超收 12×

#### Scenario: 请求 grok-3-mini-fast
- **WHEN** 客户端请求 `grok-3-mini-fast`
- **THEN** 计费 MUST 使用输入 `$0.60/MTok`、输出 `$4.00/MTok`、缓存读取 `$0.15/MTok`
- **THEN** MUST NOT 使用 Grok 4.5 的价卡

#### Scenario: 公开模型列表
- **WHEN** 客户端拉取模型列表
- **THEN** `grok-3-mini` 与 `grok-3-mini-fast` MUST NOT 出现在其中
- **THEN** 直接请求这两个 ID MUST 仍然可用并按各自价卡计费

#### Scenario: 真正未知的 Grok 文本模型
- **WHEN** 请求一个既无价卡也无别名的 Grok 文本模型 ID
- **THEN** MUST 仍回落到未知文本族兜底价卡，行为与改动前一致

### Requirement: 只有跟随型别名才被重写为运行时默认模型
系统 SHALL 只把明确表达「给我当前默认」的别名映射到运营方配置的运行时默认文本模型。
指向某个具体版本的别名（如 `grok-4.5` / `grok-4.5-latest`）MUST 原样解析为该版本，
MUST NOT 因其目标恰好等于内置默认常量就被重写——那会让客户端的**显式**选择被静默改掉。

#### Scenario: 客户端显式请求 grok-4.5
- **WHEN** 运营方把默认文本模型配成别的模型
- **AND** 客户端显式请求 `grok-4.5`
- **THEN** 上游模型 MUST 仍为 `grok-4.5`
- **THEN** MUST NOT 被改写为运营方配置的默认模型

#### Scenario: 客户端请求跟随型别名
- **WHEN** 客户端请求 `grok` 或 `grok-latest`
- **THEN** MUST 解析为运营方配置的运行时默认文本模型

#### Scenario: 未配置运行时默认
- **WHEN** 运营方没有配置默认文本模型
- **THEN** 跟随型别名 MUST 解析为内置默认常量

#### Scenario: grok-build-latest
- **WHEN** 客户端请求 `grok-build-latest`
- **THEN** MUST 解析为 `grok-build-0.1`
- **THEN** MUST NOT 跟随文本默认模型（它属于 Build 系，不是文本默认的候选）

### Requirement: Imagine 媒体模型 ID 必须与官方一致
系统 SHALL 让 `grok-imagine-video-1.5` 指向 `grok-imagine-video-1.5`、
`grok-imagine-video-1.5-preview` 指向 preview 变体，两者 MUST NOT 互换。
`grok-imagine-image-2.0` MUST 出现在公开模型列表中。

#### Scenario: 请求 1.5 正式版
- **WHEN** 客户端请求 `grok-imagine-video-1.5`
- **THEN** 出站模型 ID MUST 为 `grok-imagine-video-1.5`

#### Scenario: 请求 1.5 preview
- **WHEN** 客户端请求 `grok-imagine-video-1.5-preview`
- **THEN** 出站模型 ID MUST 为 preview 变体

#### Scenario: image 2.0 可见性
- **WHEN** 客户端拉取模型列表
- **THEN** `grok-imagine-image-2.0` MUST 在列表中

### Requirement: 默认文本模型为 grok-4.6，且计费目录随之对齐
系统 SHALL 把内置默认文本模型定为 `grok-4.6`。计费目录 MUST 同步：
`grok` / `grok-latest` 归到 4.6 的价卡，`grok-4.5` / `grok-4.5-latest` 保留 4.5 的价卡，
`grok-4.20-*` 变体 MUST 使用自己的价卡而不再与 `grok-4.3` 共用。
后台设置解析与前端模型白名单 MUST 与新默认值一致。

#### Scenario: 未配置默认模型时的请求
- **WHEN** 运营方未配置默认文本模型，客户端请求 `grok`
- **THEN** 上游模型 MUST 为 `grok-4.6`
- **THEN** 计费 MUST 使用 4.6 的价卡

#### Scenario: 显式请求 4.5 的计费
- **WHEN** 客户端显式请求 `grok-4.5`
- **THEN** 计费 MUST 使用 4.5 的价卡，MUST NOT 使用 4.6 的

#### Scenario: grok-4.20 变体的计费
- **WHEN** 请求任一 `grok-4.20-*` 变体
- **THEN** MUST 使用 `grok-4.20` 的价卡
- **THEN** MUST NOT 与 `grok-4.3` 共用同一张卡

#### Scenario: 后台可选值与前端白名单
- **WHEN** 管理员在后台选择默认文本模型
- **THEN** 可选值 MUST 包含 `grok-4.6`
- **THEN** 前端模型白名单 MUST 与后端一致，MUST NOT 出现能选但被后端拒绝的值

### Requirement: 存量默认模型设置的迁移只改旧内置默认值
系统 SHALL 在启动时把存量库中**恰好等于旧内置默认值** `grok-4.5` 的
`grok_default_text_model` 设置重写为 `grok-4.6`。设置从未写过时 MUST 为 no-op；
值为任何其他内容时 MUST 保持不变——那是运营方的显式选择或未来的默认值。
迁移失败 MUST NOT 阻止服务启动。

#### Scenario: 存量库显式存着旧默认值
- **WHEN** 设置值为 `grok-4.5`（含前后空白）
- **THEN** MUST 被重写为 `grok-4.6`

#### Scenario: 设置从未写过
- **WHEN** 该设置不存在
- **THEN** MUST 为 no-op，MUST NOT 写入任何值
- **THEN** 运行时 MUST 自然取到新的内置默认值

#### Scenario: 运营方显式选择了别的模型
- **WHEN** 设置值为 `grok-4.3` 或任何非 `grok-4.5` 的值
- **THEN** MUST 保持不变
- **THEN** MUST NOT 被「不是 4.6 就改成 4.6」这类逻辑覆盖

#### Scenario: 迁移过程出错
- **WHEN** 读取或写入设置时数据库报错
- **THEN** MUST 记录告警
- **THEN** MUST NOT 阻止服务启动
