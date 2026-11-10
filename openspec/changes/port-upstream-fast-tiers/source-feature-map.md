# Source feature map

固定行为来源为 v0.2.19 `e60cefeaf78f2fbfa95aa076029e28b40d0a5527`，不是最新发布评估。整 PR 文件检查见 [source-file-states.json](../../../docs/upstream-sync/implementation-fast-tiers/source-file-states.json)。

| 上游来源 | 本次行为 | 本地适配 |
|---|---|---|
| #6129 `bd17411d0` | Fast 参数、合法枚举、计价倍率和响应档位 | 复用现有 HTTP/Chat/Responses/WS，补 Messages 显式参数与校验；保留本地流式 usage 和 SQLite dedup |
| #6378 `fdf9751c1` | Codex OAuth default 回报例外 | 转发结果构造时按账号解析；不新增 usage 列 |
| #6379 `6b39d0b45` | 分组强制 priority | 信任已加载分组，执行后仍受全局规则控制；省略/显式档位、HTTP/WS 均覆盖 |
| #6444 `f2804eb2c`、#7613 `dc4874a2f` | Free Fast、缺价格保留零成本用量 | 只替换用户 ActualCost，保留 TotalCost；平台限定 OpenAI/composite 的 OpenAI 账号 |
| #6571 `c42d78e26` | Ultrafast 转发、策略与 Sol 目录 | 新档位不再被兼容转换剥离，路由 hint、日志显示同步 |
| #5833 `a59f0c7fa` | missing -> force priority | all 仍仅匹配显式档位；missing 仅 force_priority 注入 |
| #7730 `327c32218` 内 `e6d191a83` | Astra Ultrafast 能力与 6x 价格 | 官方 API 默认能力，OAuth/自定义上游依赖自身 metadata；显式空列表优先 |
| #7811 `db29b3a77` | service_tiers null 规范化 | 本地目录能力合并中规范 null；原生源字段优先 |

渠道 FastMultiplier 只取所需字段与验证，不携带上游 effort/Flex/时段计价扩展。上游 PG 232/233 分组字段改为新 SQLite228；旧迁移不改，Ent 正式生成。已有本地长上下文策略、Terra 显式零价、Claude speed 日志和 ops 修复保留。
