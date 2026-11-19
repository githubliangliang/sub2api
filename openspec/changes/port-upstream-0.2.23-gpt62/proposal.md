# GPT-6.2 Sol / Astra 接入

## Why

用户已选择此前按需保留的 GPT-6.2 能力。当前资源目录会将未知 6.2 型号回退到 GPT-5.4，计费兜底则无法识别；缺少两个模型的 Codex 描述与促销时钟。

## What Changes

补齐两个规范模型 ID、既有 effort/compact 别名、官方 Codex 描述、协议校验、价格资源及双层兜底。按整次输入总量严格大于 272000 应用长上下文，Fast 2 倍；Ultra Fast 在太平洋时间 2027-03-01 00:00 前 3 倍、到点起 6 倍。将既有请求/WS 每轮 PricingAt 贯通用户结算和独立账号统计。补齐白名单、映射预设及 OpenCode 配置。

## Impact

复用现有模型、价格、用量及 SQLite 结算链路。保留运营者显式价格、零价、既有长上下文开关及 Free Fast 的适用边界。订阅套餐补全、强制 Ultra Fast 分组、新渠道 Ultra Fast 字段仍单独按需。版本、schema、依赖、Ent/Wire 与资源默认值不随本次模型接入修改。
