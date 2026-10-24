# 追踪表：上游 commit → Requirement → 目标代码 → 证据

12 项 P0 与 6 个 capability 的双向覆盖。**每一项都必须在这张表里同时有 Requirement 和证据**，
否则说明 spec 漏了或 verification 漏了。

| 上游 commit | 版本 | capability / Requirement | 目标代码 | 证据（`verification.md`） |
|---|---|---|---|---|
| `8e60d574` | v0.1.183 | `openai-session-stickiness` / 显式会话头双写法 | `service/openai_gateway_scheduling.go:31`、`service/openai_ws_forwarder_logutil.go:68` | §2 双写法两行 + 自补「同时出现取连字符」 |
| `bc4a9ae4` | v0.1.182 | `anthropic-cache-billing-integrity` / 三条 Requirement | `service/gateway_anthropic_passthrough.go:673,676`、`service/gateway_upstream_response.go:1239,1243`、`service/billing_service.go:1223` | §2 前三行 + §3 计费专项 |
| `19da0f24` | v0.1.181 | `upstream-request-compatibility` / `deprecated` 剔除、enum 归一 | `service/gemini_messages_compat_service.go:3532` | §2 Gemini 行 |
| `1a9898a6` | v0.1.183 | `upstream-request-compatibility` / token 上限封顶 | `service/antigravity_gateway_compat.go:152` | §2 Antigravity token 行 |
| `4ca86c52` | v0.1.183 | `email-identity-binding-integrity` / 三条 Requirement | `repository/user_repo.go:1145` 起、`service/auth_email_binding.go:59` | §2 两行 + §4 认证专项 7 条 |
| `329b92ef` | v0.1.182 | `upstream-request-compatibility` / 图片 prompt 逐字 | `service/openai_images.go:42` 常量块、`service/openai_images_responses.go:357` | §2 图片 prompt 行 |
| `e0e5e45c` | v0.1.183 | `upstream-request-compatibility` / item ID 类型一致 | `pkg/apicompat/responses_client_tools.go:220`（4 处调用点） | §2 item ID 两行 |
| `a6b11ccc` | v0.1.182 | `account-cooldown-accuracy` / 单条 Requirement | `service/ratelimit_service.go` | §2 OpenCode 两行 |
| `9fb26043` | v0.1.181 | `upstream-request-compatibility` / Grok 官方 UA | `pkg/xai/billing.go:22`、`service/grok_upstream_headers.go:15`、`service/openai_gateway_chat_completions_raw.go:175`、`service/grok_observed_models.go:95` | §2 Grok 行 + `tasks.md` 6.5 grep 门禁 |
| `eb594eef` | v0.1.182 | `payment-balance-visibility` / 单条 Requirement | `frontend/src/views/user/PaymentResultView.vue` | §2 充值余额行 |
| `e55727d4` | v0.1.183 | `openai-session-stickiness` / 溢出不改绑 | `service/openai_gateway_scheduling.go:932,1029,1185,1224` | §2 溢出两行 + §5 调度观察 |
| `99ec347e` + `71aa6e35` | v0.1.182 | `upstream-request-compatibility` / Sonnet 别名 | `domain/constants.go:105`、`service/account_test_service.go:2272` | §2 Sonnet 行（含「显式 4.5 不变」断言） |

## 反向核对：capability → commit

| capability | Requirement 数 | 来源 commit |
|---|---|---|
| `anthropic-cache-billing-integrity` | 3 | `bc4a9ae4` |
| `openai-session-stickiness` | 2 | `8e60d574`、`e55727d4` |
| `upstream-request-compatibility` | 7 | `19da0f24`、`1a9898a6`、`99ec347e`+`71aa6e35`、`329b92ef`、`9fb26043`、`e0e5e45c` |
| `account-cooldown-accuracy` | 1 | `a6b11ccc` |
| `email-identity-binding-integrity` | 3 | `4ca86c52` |
| `payment-balance-visibility` | 1 | `eb594eef` |

合计 17 条 Requirement / 50 个 Scenario，覆盖 12 项 P0，无遗漏、无多余。

## 明确不在本表内的上游 commit

| commit | 判定 | 理由出处 |
|---|---|---|
| `1563db3f` `53d76ad8` `d6012b0b` `d5e43ef7` `095b5253` | 缺 0.1.180 基座 | `source-baseline.md` §5 |
| `f1aadd48` | 本仓库缺陷不成立 | `source-baseline.md` §5 末段 |
| `e440ac48` | 缺 status/content/cache 类拒绝重试基座 | `PORTING-0.1.183.md` §5.1 |
| `22e1b814` `e471be73` `3e98a5a1` `b16ed03c` `5a2f542a` `e39fce27` `fc589bce` `db01fb98` `5934981e` `195b2197` `2abce650` | 未发布 + 上游返工中 | `PORTING-0.1.183.md` §5.3 |
| `49752060` + `b20f29d1`、`4795650d` | P1，需手工改写，另开 change | `PORTING-0.1.183.md` §4 |
| `3802268e` `4347e555` | CN 平台 N/A | `PORTING-0.1.183.md` §6 |
| `03e8ab41` `e2d9b823` `aa2c4e8d` `7634e3c2` `3b7753a8` `66d664ff` `6ca1e15b` | chore | `PORTING-0.1.183.md` §6 |
