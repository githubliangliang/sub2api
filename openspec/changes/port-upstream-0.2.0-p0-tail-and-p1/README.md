# port-upstream-0.2.0-p0-tail-and-p1

上游 v0.1.185 + v0.2.0 评估（[`docs/upstream-sync/PORTING-0.2.0.md`](../../../docs/upstream-sync/PORTING-0.2.0.md)）
划出的 **第二档**：4 项 P0 余项 + 4 项 P1 + 1 项并入的依赖卫生（第 9 项，来源不是本轮，见下）。**必须在
[`port-upstream-0.2.0-p0-fixes`](../port-upstream-0.2.0-p0-fixes/)（第一档）之后实施**——
本批多处冲突只因第一档未落，排序后会塌掉一大半。

第三档（按需）与第四档（不合）**未立项，等用户确认**，理由见 PORTING §5。

| # | 内容 | 上游 | PORTING | 逐文件 apply 四态 |
|---|---|---|---|---|
| 1 | delegation / scheduled-automation bootstrap 缺 call id 被拒 | `1be69e56a` → `421a83282` | §3.8 | 2/2 `CLEAN` + 链 |
| 2 | API key 会话缓存身份丢失 | `504919a05` | §3.9 | 3 `CLEAN` / 1 `CONFLICT` |
| 3 | 终止事件前收到 close 被判成功（+11 行基座） | `bfe0a5a87` | §3.10 | 2 `CONFLICT` + 缺 `openAIWSRelayActiveTurnID` |
| 4 | Fable 缺 fallback 定价 + Fable 5.1 模型面 | `34b8bf1a6` **前半** | §3.11 | 大部分 `CLEAN` |
| 5 | `pricing.override_file` 价格目录覆盖补丁 | `593fc9365` | §4.2 | 3 `CLEAN` / 1 `CONFLICT` |
| 6 | 推理档位：超限拒绝/降级 + 按模型限定映射（**+迁移 225**） | PR#6447 → PR#6425 | §4.1 | 38/47 + 6/15 `CLEAN` |
| 7 | 渠道 `cache_write_1h_price`（**+迁移 226**） | `34b8bf1a6` **后半** | §4.3 | 需 SQLite 重写 |
| 8 | 分组模型定价弹窗布局 | `1a33dc8cc` | §4.4 | 1 `CLEAN` / 3 `CONFLICT` |
| 9 | dompurify `3.3.1` → `3.4.14`（**并入项，来源 v0.1.180**） | `4a1da2950` | [0.1.180 §12](../../../docs/upstream-sync/PORTING-0.1.180.md) | 不走三态：3 文件手改 + lockfile 重生成 |

⚠️ 六条必须先读的约束：

- **顺序固定**：第 1 项内部 `1be69e56a` → `421a83282`（后者的 `CONFLICT` 只因前者未落）；
  第 6 项内部 **PR#6447 → PR#6425**（后者 9 处冲突基本都是与前者改同一批文件造成的）。
- **第 3 项的基座要跟这一项一起落，不要单独先合。** `openAIWSRelayActiveTurnID` 上游 11 行纯函数，
  本仓库零命中；单独落一个没人调用的 helper 是死代码（README「基座要随用它的簇一起移植」）。
- **第 4 / 7 项来自同一个上游 commit `34b8bf1a6`，但是两件不相关的事，必须拆开。**
  前半是模型面 + 兜底定价（无迁移），后半是渠道 1h 缓存写入价（带迁移）。见 `design.md` 决策 4。
- **两条迁移都要 PG → SQLite 重写**：去掉 `ADD COLUMN ... IF NOT EXISTS`、整段删 `COMMENT ON COLUMN`。
  本仓库号是 `225` / `226`（上游三个文件都叫 `232_*`，自己撞了号）。见 `design.md` 决策 6。
- **落地任何 `_integration_test.go` 之前先 `head -1` 看构建标签。** 本仓库有一批是
  `//go:build integration && postgres`，在 SQLite-only fork 里永远不编译——
  `internal/repository/migrations_schema_integration_test.go` 就是（改它的是第三档的 PR#6443 /
  PR#6444，不在本批）。本批里 `channel_repo_pricing_time_test.go` 是 `NOBASE`，同样先看标签再决定。
  **MUST NOT 为了让这类 hunk 能打上去而动构建标签。**
- **第 9 项不是本轮上游项。** 它是 0.1.180 §5.1 挂了一轮的安全项，2026-09-02 重量后改判「建议做」
  并按用户决定并入本批。**它的三态不适用**（不是打 patch，而是改 3 个依赖文件 + 重生成 lockfile），
  且**必须最后做**——它一改 lockfile，前面每个阶段的前端测试都得在新树上重跑一次才有意义。
  见 `design.md` 决策 11。
- **第 6 / 7 / 8 项动前端，必须真跑 vitest**，`pnpm run typecheck` 抓不到 `is not defined`
  （0.1.184 §11.1）。第 8 项与第 7 项改同一批前端文件 ⇒ 必须相邻实施。

阅读顺序：`proposal.md` → `source-baseline.md` → `source-feature-map.md` → `design.md` →
七个 `specs/*/spec.md` → `tasks.md` → `verification.md`。

逐条 patch site 与上游 diff 说明只在
[`docs/upstream-sync/PORTING-0.2.0.md`](../../../docs/upstream-sync/PORTING-0.2.0.md) §3.8–§3.11 与 §4，
本 change 不复制，只定义**移植后必须成立的行为**与验收证据。
