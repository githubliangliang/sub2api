# 实施验收（待填写）

状态：用户选定第三档Terra，提案未实施。所有实施信息、证据和结果槽保持空白。

| 实施信息 | 值 |
|---|---|
| 实施起点HEAD / 工作树 | |
| 实现提交 / 代码SHA256 | |
| 实施VERSION / Go / 迁移 | |
| 实际范围与排除项 | |

## 行为与反例

| 行为 | 必须验收 | 新行为失败证据 | 完成证据 | 结果 |
|---|---|---|---|---|
| 身份/effort | 六档/default medium、none/max、拼写/未知后缀、Sol不回退 | | | |
| 工具协议 | 三个only-Chat出口出站前400、Responses工具及无工具raw正例 | | | |
| 目录/metadata | 账号声明优先、空catalog不增项、原目录/ETag不变 | | | |
| price/usage | 资源/两层fallback、显式0/override/priority/272k边界、异步usage/SQLite dedup | | | |
| 客户端 | 白名单/映射预设/可解析生成配置、六档/context/output、旧默认和模型 | | | |

## 组合检查

| 检查 | 命令 / 证据 | 结果 |
|---|---|---|
| backend完整unit / build / lint | | |
| 协议/usage/race与取消边界 | | |
| SQLite迁移227 / quota / dedup | | |
| frontend typecheck / lint / tests / build / critical | | |
| 浏览器模型选择 / 客户端生成配置 | | |
| OpenSpec严格校验 / source哈希 / 范围 | | |

- [ ] 新功能有前后证据、反例与真实消费者验收。
- [ ] 适用门禁与冻结范围核对完成，实际SHA/起点写入。
- [ ] baseline未冒充实施证据，没有未授权支付、发布或部署。

