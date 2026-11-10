# Proposal

补齐已有网关 Fast、Free Fast 和 Ultrafast 的转发、策略、目录、计价与管理配置。防止兼容转换静默删除 ultrafast，保证用户费用与上游成本可区分。保留 SQLite、可选 Redis、simple mode 及现有额度去重。无支付、发布或部署。

## Why

当前兼容转换删除 ultrafast，分组加速策略和计价尚未贯通。

## What Changes

统一档位转发、策略、目录、费用和配置持久化；新增 SQLite228 与 Ent 生成字段。

## Impact

影响已有 OpenAI/Codex/Claude 路径、管理员分组/渠道/策略页；默认分组开关关闭。
