# 用户选定 Fast / Free Fast / Ultrafast 完整整合

## 1. 基线与选择

本轮是用户明确选择的历史第三档能力实施，非新上游 release intake。2026-11-05 授权，2026-11-10 继续完成；fork 起点 `5c135c31dfbd6a1fcb701767f2f78933c4280aea`，VERSION1.1.23 / Go1.27.3 / SQLite227；来源固定 v0.2.19 `e60cefeaf78f2fbfa95aa076029e28b40d0a5527`。新增 SQLite228，保留版本与依赖。

## 2. 方法与来源

按现有路径手工移植所需行为，不合 upstream/main，不整文件替换网关。#6129/#6378/#6379/#6444/#6571/#5833/#7613/#7730/#7811 的整 PR diff 在独立 baseline index 上补做逐文件 reverse-before-forward：164 个 PR×文件状态，ALREADY8 / CLEAN37 / CONFLICT104 / NOBASE15。[原始结果](implementation-fast-tiers/source-file-states.json)仅说明补丁兼容性，不代表所有文件纳入范围。#7730 只接所需档位能力与价格，已有模型工作不重复移植。

## 3. 转发和目录 patch sites

- `backend/internal/service/openai_gateway_request_body.go`、`openai_gateway_forward.go`、`openai_routing_hint.go`：识别 ultrafast、missing 策略、分组强制与全局策略顺序。
- `backend/internal/handler/openai_gateway_handler.go`、`openai_chat_completions.go`、`service/openai_gateway_messages.go`：非法枚举在上游前拒绝。
- `openai_gateway_messages_chat_fallback.go`：补齐省略路径的分组策略和原生 speed/显式 tier 的兼容转换。
- `openai_ws_forwarder_*`、`openai_ws_v2/passthrough_relay.go`：每轮档位观测和结算隔离。
- `openai_codex_model_metadata.go`、`openai_codex_models_service.go`：账号声明优先、null 与空目录、受限官方 Astra fallback。

## 4. 计价、存储和页面 patch sites

- `service_tier_billing.go` 与各结果构造处：公开 API 降档 / Codex default echo 语义。
- `billing_service.go`、`model_pricing_resolver.go`、`gateway_usage_billing.go`：Fast 比率、独立 Ultrafast 比率、Claude 原生 Fast、渠道覆盖、既有长上下文规则。
- `openai_gateway_usage.go`：Free Fast 只重算用户费用，沿用缺价记录和去重。
- `migrations/228_accelerated_service_tiers.sql`、`ent/schema/group.go`、repository、admin_group、auth cache、DTO：新配置从 SQLite 到请求上下文完整贯通，鉴权快照升至20。
- `GroupsView.vue`、`ChannelsView.vue`、`PricingEntryCard.vue`、`SettingsView.vue`、中英语言包和 usageServiceTier：开关、维度无关倍率、策略档位、Ultrafast 显示。

新增核心符号均已编译并进入消费者：`ValidateOpenAIServiceTierField`、`resolvedOpenAIServiceTier`、`configuredServiceTierMultiplier`、`openAIGroupForcesFast`、`groupBillsOpenAIFastAtStandard`、`configuredCodexServiceTiersForModel`。Ent 使用 go generate，非手改生成结构。

## 5. 保留边界及旧项

Fast/Free Fast/Ultrafast 这一已选第三档本次完成。动态长上下文阶梯、effort 倍率、usage compaction/requested effort schema、Realtime、Grok 429/WS 容量、exec_scope/抢占、新平台和其它第三档仍独立按需；既有 Terra 和模型价格规则继续保留。支付仍第四档、未扩围。原 PORTING 和 frozen source/verification 状态原文保留。本次没有新增发布、推送或部署。

## 6. 已有能力

旧 Fast policy、fast→priority 别名、原生 Claude Fast 日志分类、protocol usage/去重、SQLite/miniredis/simple mode 已存在，复用并补齐边界。没有将这些已存在内容重新宣称为全部新实现。

## 7. 落地顺序

非法/丢失档位复现 → 转发/目录 → 分组及渠道持久化 → 费用 → UI → HTTP/SQLite/race/全量/浏览器 → 文档和范围核对。源补丁四态作为最终独立核对记录。

## 8. 验收

见 [实施证据](implementation-fast-tiers/README.md) 和 [OpenSpec verification](../../openspec/changes/port-upstream-fast-tiers/verification.md)。新迁移只在隔离实例执行；实际供应商权限与速度不在合成测试证据范围内。

## 9. 经验

请求参数能穿过某个直通入口不代表兼容桥也保留；日志档位、客户端费用和上游成本应分别验证。Codex OAuth 的 default 回报不能照搬公开 API 结算。分组开关必须覆盖鉴权投影、缓存版本、复制和 WebSocket 多轮，而不能只保存一个表单字段。
