# 已移除页面 / 入口清单（个人精简）

## 第一批

| 路径 | 说明 |
|------|------|
| `/admin/redeem` | 管理端兑换码 |
| `/admin/promo-codes` | 管理端优惠码 |
| `/admin/announcements` | 管理端公告 |
| `/admin/ops` | 运维监控页；2026-10-25 按用户要求恢复完整页面、侧栏和菜单管理入口 |
| `/admin/dashboard` | 管理端仪表盘（首页 → `/admin/accounts`） |
| `/monitor` | 用户渠道状态 |
| `/subscriptions` | 用户我的订阅 |
| `/redeem` | 用户兑换码 |

## 第二批（支付 + 邀请）

| 路径 | 说明 |
|------|------|
| `/purchase` | 充值/订阅购买 |
| `/orders` | 用户订单 |
| `/payment/*` | 支付结果/二维码/Stripe 等 |
| `/admin/orders` `/admin/orders/*` | 管理端支付订单/套餐/概览 |
| `/affiliate` | 用户邀请返利 |
| `/admin/affiliates/*` | 管理端邀请/返利/提取 |

## 前端

- 路由删除对应 `path`
- 侧栏删除菜单项
- `menuCatalog` 同步
- 顶栏：无公告铃铛、无订阅进度条
- 用户仪表盘：无兑换快捷入口

## 后端

- 不注册：dashboard / announcements / redeem / promo 管理路由
- ops：`/api/v1/admin/ops/*` 始终保留（用量页错误请求依赖）；2026-10-25 恢复完整前端页面，继续使用这些 API。
- 不注册：用户 announcements / redeem / subscriptions / channel-monitors
- 不注册：`RegisterPaymentRoutes`
- 不注册：`registerAffiliateRoutes` 及用户 `/user/aff*`

handler 代码仍保留，仅下线路由。

恢复的运维页仅管理员可访问，standard/simple 模式均可用；侧栏继续受 `ops_monitoring_enabled` 与 `hidden_menu_keys` 控制。其他已移除页面和 simple 模式限制保持原样，SQLite 聚合/采集缓存默认开关保持原设置。

## 部署配置建议

见 `deploy/config.personal.sqlite.yaml` 与 `deploy/START_NATIVE.md`：

- `run_mode: standard`（推荐；`simple` 会连分组调度一起跳过，理由见根 README「run_mode」一节）
- `default.user_balance: 100000`（standard 会校验余额）
- `redis.enabled: false`
- `ops.enabled: true`（仅错误/系统日志 + 定时清理；`aggregation`、采集缓存、预聚合表仍关）
- `batch_image.enabled: false`
