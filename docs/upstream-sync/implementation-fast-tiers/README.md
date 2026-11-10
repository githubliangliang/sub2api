# Fast / Free Fast / Ultrafast 实施与验收

用户在 2026-11-05 选择完整整合，2026-11-10 继续并完成。起点 `5c135c31dfbd6a1fcb701767f2f78933c4280aea`，固定上游 v0.2.19 `e60cefeaf78f2fbfa95aa076029e28b40d0a5527`。实现位于当前工作区，未创建提交、tag 或发布；未部署业务实例。

## 已交付行为

- OpenAI/Codex 的 fast 归一化为 priority；ultrafast 在原生 Responses、自动透传、Chat、Messages、Chat 后备转换及现有 WS 路径保留。HTTP 显式非法枚举返回 400；WS response.create 使用既有错误事件拒绝，无效非创建事件不改写。
- 全局规则支持 ultrafast 和 missing。missing 只有 force_priority 会注入；旧 all 规则不改变省略档位请求。分组强制 Fast 把该分组的 OpenAI 请求设为 priority，包括原先的 Ultrafast，随后仍执行全局过滤/拦截规则。
- Free Fast 仅对启用分组中的 OpenAI priority 请求，将用户 ActualCost 重算为普通档；保留上游 TotalCost、费用明细和档位。不作用于 ultrafast/Grok，缺价格仍保留零成本 usage。
- OpenAI 公开 API 响应可降档，不因响应声明较高档位加价；Codex OAuth/setup-token 的 default echo 保留实际发送档位。WS 按轮观测，缺失/未知帧不污染已有值，下一轮不继承旧档位。
- Claude 原生 speed=fast 保留现有转发与日志判定，并进入档位计价。Messages 到 OpenAI 的 speed=fast、显式 service_tier 和既有 Fast beta 兼容路径可使用 priority；原生 Claude 仍不单凭 beta header 推断 Fast。
- Codex 目录保留显式账号 metadata 和空列表。官方 Astra API 默认支持 Ultrafast；OAuth 套餐名不授予该能力，自定义上游遵循自身声明。内置 gpt-5.6-sol 目录补齐 Ultrafast。
- 沿用上游固定快照的模型 Fast 比率（如 GPT-5.5 2.5、Astra Fast 2 / Ultrafast 6），支持分组/渠道 Fast 倍率覆盖；Ultrafast 独立。保留当前长上下文、显式零价和缓存写入处理，不搬入 effort、动态阶梯或新平台。

## 使用位置与默认行为

1. 系统设置 → 网关服务 → OpenAI 加速档位策略：选 Ultrafast 或未指定档位，配置透传/过滤/强制/拦截规则。未配置规则时不自动升级请求。
2. 分组管理 → OpenAI/composite 分组：强制 Fast 和 Free Fast 两个独立开关，默认均关闭。想保留客户端 Ultrafast 时不要开启分组强制 Fast。
3. 分组逐模型定价、渠道模型定价 → Fast 价格倍率：填有限正数；清空恢复原有价格策略。只作用于 Fast/priority，不覆盖 Ultrafast。
4. simple mode 的既有分组路由限制保持；需要分组策略时用 standard。此处没有扩大 simple mode 页面权限。

## 数据与范围

SQLite 新增迁移 **228_accelerated_service_tiers.sql**：groups 的两个布尔列默认 false；channel_model_pricing 的 fast_multiplier 可空且必须为正。旧迁移 001–227 不改。Ent 按 schema 生成，鉴权快照版本 19→20；分组复制、读写、平台切换、缓存刷新和管理员 DTO 已接线。

VERSION **1.1.23**、Go **1.27.3**、依赖、Wire、已有配额清理/固定窗口/usage dedup 保持。新的 feature scope 不包含支付、发布、部署、Realtime、新模型或历史动态价格体系。

## 验证索引

| 检查 | 结果 / 日志 |
|---|---|
| 后端全量 unit | 54 包通过；[最终完整运行](backend-unit-complete.log)，允许 Go 缓存 |
| 定向行为 | [backend-targeted.log](backend-targeted.log)；最终新增 Messages 校验亦通过 [messages-validation-after.log](messages-validation-after.log) 和最终 race |
| race | service、openai_ws_v2、handler、repository 四包 Fast/Ultrafast/ServiceTier/WS/SQLite quota 定向通过；[最终日志](backend-race-final-20261110.log) |
| 真实 HTTP 与 SQLite | `TestFastTierAsyncUsageSQLiteSettlementAndDedup`：6 协议 × priority/ultrafast/free-fast，共 18 场景，实际 HTTP 传输、queued settlement、重复请求只入账一次；`TestFastTiersSQLiteUpgradePersistenceAndAuthProjection`：真实 227→228 升级、幂等重启、实体保存、鉴权投影、渠道倍率写入与清除 |
| 后端构建 / lint | embed 构建通过 [backend-build-final.log](backend-build-final.log)；[lint 0 issues](backend-lint-complete.log) |
| 前端全量 / critical | 291 文件 **2179 passed / 2 既有 skipped**，[全量](frontend-unit-final.log)；critical **176 passed / 2 同一既有 skipped**，[日志](frontend-critical.log) |
| 前端类型 / lint / 构建 | [最终构建](frontend-build-final.log) 包含 vue-tsc -b；[lint](frontend-lint-final.log) 通过。保留既有 pnpm overrides、Browserslist 和 bundle 提示 |
| 浏览器 | [命令记录](browser/commands.json)：实际 admin 登录、分组开关与倍率保存/回显/清除、390px 表单无横向溢出、渠道 2.5 倍率保存/回显、Ultrafast 和 missing 策略保存/重新加载；最终二进制 API 对倍率 0 返回 400 |
| OpenSpec | [strict 通过](openspec-strict.log)；[完整 change](../../../openspec/changes/port-upstream-fast-tiers/README.md) |
| 来源 / 范围 | [整 PR 文件四态](source-file-states.json)、[最终文件哈希及保护核对](scope-check.json) |

## 证据边界和过程问题

浏览器使用新建 `/tmp/sub2api-fast-acceptance-data`、127.0.0.1:18923 的本地实例和合成管理员/分组/渠道；真实构建 SPA、API 与 SQLite，未连接业务实例。为界面验收直接初始化了合成管理员 gate 状态，未代表用户提交协议接受操作。临时服务运行曾在仓库产生价格缓存，已移至临时目录。截图和 SQL 查询结果仅包含合成配置。

协议测试中的上游是可控 HTTP/transport fixture，不验证商业账号实际权限或供应商加速延迟。没有声称完成真实供应商收费/吞吐验证。价格来源是选定上游快照，不是实时官方报价。

初轮失败保留在 backend-unit.log / backend-lint.log：Messages 原先丢弃显式档位，修复后 Terra 真实 SQLite 回归通过；旧 Codex OAuth default 断言按已选择的上游计价例外调整；目录原断言只允许一个 Fast 项，改为检查包含 priority，新增 Ultrafast 能力测试。新增测试曾遗漏 unit build tag、DTO 普通用户字段隔离、fixture HTTP/SSE 模式和必填渠道分组，均已修正并验证。Messages 非法档位和倍率错误码的修复前失败分别保留在 messages-validation-before.log / multiplier-validation-before.log。测试命令早期也曾因工作目录、Node PATH、Ent Openai 命名产生错误；未通过删除现有业务测试或削弱断言消除失败。

历史来源、归档、发布及既有实施证据不回写；新增 source-file-states 在结束阶段用独立临时 index 对实施起点补做全量 reverse-before-forward，不能误读为初始执行顺序或整 PR 已直接应用。
