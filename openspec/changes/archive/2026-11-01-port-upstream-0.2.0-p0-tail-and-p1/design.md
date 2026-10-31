# 设计与决策

## 1. 移植策略

按 `docs/upstream-sync/README.md` 第 3 节：**按功能手工移植，不 cherry-pick 整个 merge commit**。
本批与第一档的区别是每一项都要手工改，因此顺序比第一档更重要——**必须在第一档合完之后开始**，
否则第 1 项的 `421a83282` 与第 6 项的 WS/handler 冲突全是「第一档未落」造成的假难度。

四态证据、依赖符号核查、构建标签核查在 `source-baseline.md`；逐条 patch site 在
`docs/upstream-sync/PORTING-0.2.0.md`。本文件只记决策。

## 2. 阶段划分

| 阶段 | 内容 | 迁移 | 提交粒度 |
|---|---|---|---|
| 1 | `1be69e56a` → `421a83282` bootstrap | — | 两条各自单独（顺序不可换） |
| 2 | `504919a05` API key 缓存身份 | — | 单独 |
| 3 | `bfe0a5a87` + 11 行基座 | — | 一起（基座单独提交是死代码） |
| 4 | `34b8bf1a6` 前半 Fable | — | 单独 |
| 5 | `593fc9365` `override_file` | — | 单独 |
| 6 | PR#6447 → PR#6425 推理档位 | **`225`** | 两条各自单独（顺序不可换） |
| 7 | `34b8bf1a6` 后半 渠道 1h 缓存价 | **`226`** | 单独 |
| 8 | `1a33dc8cc` 弹窗布局 | — | 单独，紧随阶段 7 |
| 9 | `4a1da2950` dompurify 升级（并入项） | — | 单独，**必须最后**（决策 11） |

阶段 1–5 无迁移、可先跑通；阶段 6–8 带迁移与大面积前端，风险集中在后半段；阶段 9 只动依赖文件，但要放最后（决策 11）。

## 3. 决策

### 决策 1：阶段 1 与阶段 6 的内部顺序不可交换

- `1be69e56a` → `421a83282`：后者的 +95 行落在前者新增的 +232 行之上。单合后者会把 hunk 打到
  旧上下文上，语义不同。
- PR#6447 → PR#6425：两者改同一批文件（`openai_reasoning_effort_policy.go`、
  `ReasoningEffortPolicyFields.vue`、`groupsReasoningEffort.ts`、i18n），且 PR#6425 是在 PR#6447
  已进 main 的前提下开发的。反序会得到一份既不是上游 A 也不是上游 B 的中间态。

取 PR 净效果用 `git diff <merge>^1 <merge>`，**不要**逐条 cherry-pick 分支上的微提交
（`docs/upstream-sync/PORTING-0.2.0.md` §9.2）。

### 决策 2：阶段 3 的基座与主体同批落，不单独先合

`openAIWSRelayActiveTurnID` 本仓库零命中，上游 11 行。按 README「基座要随用它的簇一起移植」——
单独落一个没人调用的 helper 是死代码，且下一轮容易误以为「基座已补、功能已合」。

**只落本项需要的那一个调用点**。上游同一 helper 在 `passthrough_relay.go:785` 还有第二处调用
（`observed.responseID = openAIWSRelayActiveTurnID(state)`），属另一条改动，不落；落完
`grep -rn 'turnTimingByID'` 一遍，确认本仓库没有别处在手写同样的遍历（README「移植基座之后 grep
一遍谁还在用被它取代的旧写法」）。

### 决策 3：阶段 4 是**新增**而不是修正，`getFallbackPricing` 的判定顺序是硬要求

本仓库 `internal/service/billing_service.go` grep `fable` **零命中** ⇒ 这不是「上游修了个 bug」，
而是本仓库从来没有 Fable 兜底价。因此：

- `initFallbackPricing` 里是**插入**两条新条目，不是替换。
- `getFallbackPricing` 的系列匹配里，fable 判定 MUST 放在 opus 判定**之前**（否则命名里带
  `opus` 的映射会先命中），且 `fable-5-1` MUST 先于 `fable-5`（否则 5.1 被 5 吃掉）。
  这与本仓库既有的 `opus-5` 必须先于裸 `5` 是同一类顺序陷阱，代码里已有注释先例。

### 决策 4：`34b8bf1a6` 必须拆成阶段 4 与阶段 7

上游这个 commit 标题只说 "support Claude Fable 5.1"，实际 44 文件里**一半是渠道
`cache_write_1h_price` 拆分并带一条迁移**。按 PR 整合会把「无迁移的模型面改动」与「需要 SQLite
重写的定价改动」绑死，回滚粒度也会变粗。

判据记录在 `docs/upstream-sync/PORTING-0.2.0.md` §9.5：**看到 PR 里出现迁移文件而标题没提，就先拆。**

### 决策 5：阶段 6 的超限处置默认值必须与现状等价

推理档位超限此前是「静默钳制到上限」。新增的「拒绝 or 降级」是分组级配置 ⇒ **迁移里的
`DEFAULT` 必须落在与现状等价的那一档**（即降级/钳制），否则升级即改变所有既有分组的行为，
把此前能用的请求变成 4xx。

这是 0.1.179 长上下文门控 AND → OR 那次 breaking change 的同类风险，本批 MUST NOT 重复。
验收要求见 `specs/group-reasoning-effort-scope/spec.md` 的「存量分组」Scenario。

### 决策 6：两条迁移的 PG → SQLite 转写规则

| 上游写法 | 本仓库写法 |
|---|---|
| `ALTER TABLE t ADD COLUMN IF NOT EXISTS c NUMERIC(20,12);` | `ALTER TABLE t ADD COLUMN c NUMERIC(20,12);`（SQLite 不支持列级 `IF NOT EXISTS`） |
| `COMMENT ON COLUMN t.c IS '...';` | **整段删除**，说明改写成文件头部的 SQL 注释 |

格式照 `migrations/178_channel_image_input_price.sql`（带 `[sqlite-converted]` 头注）与
`migrations/223_group_model_pricing.sql`（裸 `ALTER TABLE ... ADD COLUMN`）。

⚠️ 两条硬约束：

- **注释里 MUST NOT 出现 PG 语法字面量**（哪怕当反例）。`sqlite_dialect_audit_test` 连注释一起扫，
  0.1.184 §25.2 被抓过一次。
- **迁移文件一旦在任何库上跑过就不能再改**（checksum 不可变，README 第 4 节）。落库前把两条
  都 review 完；本地可删 `*.db` 重来，线上不行。

上游三个新迁移都叫 `232_*`（自己撞号），本仓库按落地顺序自排 `225` / `226`。

### 决策 7：渠道 1h 缓存价的向后兼容条款必须逐字保留

`applyChannelTokenPriceOverrides` 里，`channelPricing.CacheWrite1hPrice == nil` 时
`cache_write_price` **继续同时覆盖 5m 与 1h 两档**（上游注释 "Preserve the pre-split behavior for
existing configurations"）。MUST NOT 简化成无条件赋值——那会让所有存量只配了 `cache_write_price`
的渠道在 1h 档上突然回落到官方价。

### 决策 8：阶段 8 只有呈现契约，不定义 capability

`1a33dc8cc` 是纯样式。它与阶段 7 改同一批前端文件（`IntervalRow.vue`、`PricingEntryCard.vue`、
`GroupsView.vue`）⇒ **必须相邻实施**，否则同一处冲突要解两遍。验收只有「前端测试通过 + 不引入
横向滚动」。

### 决策 9：所有 `_integration_test.go` 落地前先看构建标签

`internal/repository/migrations_schema_integration_test.go` 在本仓库是
`//go:build integration && postgres`，**永远不编译**（改它的是第三档的 PR#6443 / PR#6444，不在本批）。
本批里 `channel_repo_pricing_time_test.go` 是 `NOBASE`，同样先看标签再决定新建还是跳过。
这类 hunk 无论三态报什么都直接丢，**MUST NOT 为了让它能打上去动构建标签**。判定成本是 `head -1`。

记录在 `docs/upstream-sync/PORTING-0.2.0.md` §9.7。

### 决策 10：前端必须真跑 vitest，不能只 typecheck

0.1.184 §11.1：`ed12ea716` 六文件全干净却直接 `escapeTomlBasicString is not defined`，
而 `pnpm run typecheck` 抓不到。本批阶段 6/7/8 都动前端 ⇒ 每个阶段收尾都要
`pnpm run test:run` + `make test-frontend-critical`。

### 决策 11：dompurify 并入本批，但排在最后一个阶段

来源与本批其余八项**不同**：它是上游 `4a1da2950`（v0.1.180 期），本仓库 0.1.180 §5.1 当初判「推迟」，
2026-09-02 重量后改判「建议做」并按用户决定并入本批。

**排最后的理由**：它一改 `pnpm-lock.yaml`，阶段 6 / 7 / 8 的前端测试就都是在旧依赖树上跑的。
放最后 ⇒ 收尾那次全量前端测试同时覆盖「前面三个阶段的前端改动」与「新依赖树」，只跑一遍。
反过来放最前，则前面每个阶段都要在新树上重跑一次才有意义。

**它不是「移植」**：上游那两个文件都不能照搬（上游 lockfile 是 pnpm 9 产物，本仓库本地 pnpm 11，
且本仓库多一个上游没有的 `pnpm-workspace.yaml`）⇒ 按结论手改，三态不适用。

**理由要写对**：不要按「修一个已知可利用的 XSS」来记账——原清单唯一引用的
`GHSA-cj63-jhhr-wcxv` 用四种原型污染形态都没能复现（0.1.180 §12.5）。可验证的理由是：
18 条 advisory 一次清掉、两份副本合成一份、7 个净化调用点（其中 `/legal/*` 是公开页）。

**也不要按「让审计门禁转绿」来记账**：门禁是 `pnpm audit --prod --audit-level=high`，
dompurify 那 18 条全是 low/moderate，从来没进门禁。升级前后门禁结果必须**相同**
（`xlsx` ×2 + `nanoid` ×1），且 `.github/audit-exceptions.yml` **一个字节都不改**。

## 4. 明确不做（本批边界）

- **不碰 PORTING §5 六簇** —— 第三档（PR#6443/#6444 Fast 组策略、`530fb20f2`+`e2cfaa46e` 长上下文
  阶梯）与第四档（Kimi 整支、`e21b849a9`、数据库启动重试、404-vs-429）**等用户确认后再决定是否立项**。
- **不落 `openAIWSRelayActiveTurnID` 的第二个调用点**（决策 2）。
- **不动 `wire.go` / `wire_gen.go`**（本批不涉及 DI）。
- **不改 `VERSION`**（保持 `1.1.11`；发版另走 CLAUDE.md「Releases / tags」流程）。
- **不改 `.github/audit-exceptions.yml`**（决策 11）；也不顺手升 `xlsx` / `nanoid` / `axios`
  ——`xlsx` 是运行时唯一的 high 且没有干净升级路径（SheetJS 已不在 npm 发新版），单独立项判。
- **不改任何 `.ts` / `.vue`**（阶段 9 只动 3 个依赖文件，7 个净化调用点一个都不改）。
- **不引入 `service.IsCNProvider` / `nativeDeepSeekResponses` 或任何 CN 平台分支**（零命中）。
- **不打任何 `ent/` 生成物补丁**（改 schema 后 `go generate ./ent`）。
- **不顺延到迁移 `227` / `228`**（那是第三档 Fast 组策略的号）。

## 5. 回滚

阶段 1–5、8 各自独立提交 ⇒ 单条 `git revert` 即可。带迁移的两阶段不同：

- **阶段 6 / 7 的迁移不可 revert。** SQLite 上 `ADD COLUMN` 已落库后回退代码不回退列；
  且改迁移文件会触发 checksum 失败。回滚方式是**只 revert 代码、保留列**（多出一个不被读写的列是无害的）。
- 阶段 6 若因决策 5 的默认值判断失误导致既有分组开始被拒：**先改分组配置**（数据面），
  再决定是否 revert 代码。revert 会把「按分组配置」退回「静默钳制」，对已经依赖新行为的分组是二次变更。
- 阶段 9 是干净的单提交 revert（3 个依赖文件一起回退）。⚠️ **revert 后必须重新
  `pnpm install --frozen-lockfile` 验证**，且回退等于把 18 条 advisory 请回来 ⇒
  只在确认某处渲染真的坏了时才回退，先查是哪个调用点。
