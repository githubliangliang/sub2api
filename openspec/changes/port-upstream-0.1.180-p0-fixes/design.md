## Context

### 为什么这一批拖到现在

`PORTING-0.1.180.md` 的判断是「119 个非 merge commit 里建议吃下约 20 条 P0 + 11 条工具桥接 P1」，
但整版 481 文件的体量让它一直没开工，随后 v0.1.181–v0.1.183 三个 bugfix 版又叠上来。结果是：

- 三个「静默失效」类缺陷一直在跑：池模式账号在 CC / Responses 上永不同号重试、OAuth 账号被空
  capabilities 排除出调度、ops 内存监控读数低估到 0.3%。这三个都不报错，只是功能悄悄不工作——
  正是本仓库 `README.md` §7 那次调度事故的同一类问题。

### 这一批的性质

交付的 19 项全是行为修正：没有依赖版本变更、没有迁移、没有 schema、没有 wire 变更、
没有新全局配置项（清单里那 2 项依赖安全项已推迟，见决策 1）。风险集中在「改漏」和
「上线后流量分布变化」。

### 复核结论

23 个 commit 全部重新核过（三态 + 依赖符号 + 成对顺序实测，见 `source-baseline.md` §3）。
21 项全部仍然成立，并抓出两处原文档漏项：`913ec5d74` 引用了本仓库没有的
`CodexCanonicalClientVersion()`（会编译失败），`e7a3c1202`/`21c07e835` 的顺序必须反转
（否则付费 Antigravity 账号被打到 sandbox 域 401）。

## Goals / Non-Goals

### Goals

- 19 项缺陷在本仓库消失，行为与上游 v0.1.180 一致。
- 每项都有可执行的验收证据；三个「静默失效」项必须有**能观察到失效**的证据，不能只证明代码改了。
- 保持迁移号、VERSION、schema、wire 图、全局配置项不变。

### Non-Goals

- 不做 0.1.180 的 §6（P1：Responses/Chat 工具桥接 11 条、Grok 一套 30 条、§6.3 四项单条小项）。
- 不做 §7（P2：PR #5888 大礼包、Fast `service_tier`、重置卡自动使用、模型列表读取上限、
  统一 token 计费路径）。
- 不做 §8 判定为不合 / N/A 的项（插件系统、模型广场分时价、CN 供应商、compose 四条）。
- 不碰 §9 的三个未决决策（Grok 默认模型 4.5→4.6、Go 1.27 工具链、长上下文计费门控 AND→OR）。
- 不做清单 §5 的 2 项前端依赖安全项（`4a1da2950` dompurify、`b410c3913` nanoid 审计例外），
  见决策 1。
- 不同步 `VERSION`，不合 sponsors / 素材类 chore。

## Decisions

### 1. 两项前端依赖安全项推迟，且**一个文件都不动**

清单 §5 里有两项依赖安全工作：`4a1da2950`（dompurify `3.3.1` → `3.4.14`，
CVE-2026-65913 / GHSA-cj63-jhhr-wcxv）与 `b410c3913`（nanoid GHSA-2v37-7h3g-55p8 审计例外）。
**决定推迟，不在本 change 交付。**

**推迟的依据**（对本仓库的可达性核查）：

- 漏洞形态是 `USE_PROFILES` 打开时 `ALLOWED_ATTR` 被重建成普通数组并按属性名查表，
  被污染的 `Array.prototype` 属性（如 `onclick`）会被当成白名单属性存活。
  **它需要页面里另有一个 prototype pollution 原语才可利用**，本身不是独立可触发的 XSS。
- 本仓库 `sanitizeSvg` 的全部输入都是**管理员自填**的：
  `components/layout/AppSidebar.vue:95/120/140` 的自定义侧栏图标来自菜单设置，
  `components/common/ImageUpload.vue:108`（`mode='svg'`）只出现在 `views/admin/SettingsView.vue`、
  `views/admin/RiskControlView.vue`、`components/admin/account/AccountTestModal.vue` 三个管理页。
- 本 fork 面向单节点、单管理员部署。攻击链需要「存在污染原语」+「管理员亲手粘贴恶意 SVG」两个条件同时成立。

⇒ 实际可达性低到可以先放。这是一个**已记录的决定**，不是遗漏。

**推迟必须是「两个文件都不动」**，不能只改 `package.json`：
`security-scan.yml` / `backend-ci.yml` / `release.yml` 三个 workflow 与 `Dockerfile` 都用
`pnpm install --frozen-lockfile`，`package.json` 与 `pnpm-lock.yaml` 不一致会让 CI 与镜像构建直接失败。

**重新评估的触发条件**（任一成立就应当立刻做）：

- 开始对外发 key、或出现第二个能进管理页的人（输入不再只由自己掌控）。
- 依赖审计或代码审查发现任何 prototype pollution 原语（攻击链的另一半补齐）。
- 因其他原因升级 mermaid 等传递依赖时顺带把 `dompurify` 抬上去（成本归零，顺手做掉）。

**真要做的时候，两处要注意**（本轮核查所得，先记下来免得下次重新踩）：

1. `pnpm install --lockfile-only` 就够——只重写 lockfile，不装 `node_modules`、不跑 postinstall。
   本机已有 pnpm `11.21.0`，`frontend/pnpm-workspace.yaml` 的 `allowBuilds` 也已配好。
   ⚠️ **不要抄上游 lockfile**（pnpm 9 产物）。
2. **overrides 要写两处**。`frontend/pnpm-workspace.yaml` 里已有一个 `overrides:` 块并注明
   「pnpm v11 reads overrides here; Docker/CI still use pnpm 9 which also honors
   package.json.pnpm.overrides. Keep both in sync.」——本机 pnpm 11 读 workspace 那份，
   CI（pnpm 9）读 `package.json` 那份。只加一处会出现「本地去重了、CI 没去重」。

**另一个独立发现**（不属于本 change，但同一个文件）：`.github/audit-exceptions.yml` 里
`lodash` / `lodash-es`（`2026-07-02`）与 `axios`（`2026-07-10`）三条例外**已过期**，而
`tools/check_pnpm_audit_exceptions.py` 对过期条目同样返回 1 ⇒ `security-scan.yml` 很可能
**已经红了约七周**，与 nanoid 无关。这条需要单独处理（续期或真正解决），不要塞进本 change。

### 2. 三个「静默失效」项的验收必须先复现失效

`b1e60ba45`（池模式重试）、`40c26f343`（空 capabilities）、`cd05772e9`（内存混用）都不抛错。
按「代码改了」验收等于没验收。spec 因此把它们写成可观测的断言：

- 池模式：`RetryableOnSameAccount` 在 CC / Responses 路径上对可重试状态码为 true，且账号刚被判定
  要禁用时为 false。
- 空 capabilities：`{}` / `[]` 与「未配置」返回同一结果（不限制），非空全 false 仍视为已配置。
- 内存：cgroup 无具体上限时 used 与 total **同时**来自 host，不出现容器 used 配宿主 total。

### 3. `913ec5d74` 补一个同名 helper，而不是内联表达式

上游 `CodexCanonicalClientVersion()` 的函数体就是 `resolveCodexOutboundIdentity("").version`，
本仓库有 `resolveCodexOutboundIdentity`。两种做法都能编译，选**补同名 helper**：

- 后续任何引用这个符号的上游提交都能继续逐字应用，不必每次再想一遍怎么替。
- 内联表达式会让「为什么这里不用 helper」变成下一个读代码的人的疑问。

### 4. Antigravity 两条合成一个提交

`source-baseline.md` §4.3：按上游顺序分两次上线会有一个「付费账号打到 sandbox 域」的 401 窗口。
本仓库没有必须复刻上游中间态的理由，合成一个提交，提交信息里写明 URL 修正是前提。

### 5. Ollama Cloud 两条按固定顺序，且视为一个可独立回滚单元

`86470628d` 的 clamp 挂在 `b30651a0a` 新建的文件上（实测单独 apply 报 NOFILE）。
两条按 13 → 14 顺序落地。回滚时也要一起回滚，否则留下引用不存在函数的代码。

### 6. `f98a056f7` 是**收窄**，需要人确认再上

它把 Google One OAuth 账号可见的模型集从 `DefaultModels`（含 3.x 与 image）收到
`GoogleOneModels`（2.5 Flash / 2.5 Pro / 2.0 Flash）。上游理由是这条 consumer 渠道本来就服务不了
新模型，收窄是把「看起来能用但会失败」变成「看不到」。但如果本仓库已有分组白名单或 model_pricing
依赖旧清单里的模型 ID，收窄后那些条目会变成孤儿。**上线前先查一遍现有 Google One 账号与分组配置。**

### 7. 十一个 capability 按行为域切

`admin-console-usability` 一个 spec 收 4 项（IPv6 解析、并发数 0、优先级列、错误详情导航），
因为它们共享同一个失效模式：**功能在后端可用，但管理台不让你用对**。其余按「失效表现」独立成 spec：
认证刷新、重试资格、资源指标、调度可用性、模型目录、流式保真、Ollama 兼容、Antigravity 端点、
Composite 端点、日志量。（原先还有一个 `frontend-dependency-security`，随决策 1 一并移出。）

### 8. `98c7b0e88` 不写 Requirement

它是修一处文档自引用 URL（文件移进 `docs/` 后失效），没有可断言的系统行为。只在 `tasks.md` 里保留
一条，`source-feature-map.md` 显式标注「无 Requirement」，避免为了凑覆盖率编造需求。

### 9. 迁移号、VERSION、wire 明确不动

本批不涉及任何一项。实现中若出现新增 SQL 文件、改 `VERSION` 或改 `wire_gen.go` 的冲动，
说明走偏了（`tasks.md` §7 有对应门禁）。

## Risks / Trade-offs

| 风险 | 处理 |
|---|---|
| `40c26f343` 让此前被静默排除的 OAuth 账号重新进入调度，流量分布突变 | 单独提交；上线后一个观察期盯账号级并发、429 与错误率分布 |
| `f98a056f7` 收窄 Google One 可见模型，可能让既有分组白名单出现孤儿条目 | 决策 6：上线前先查现有 Google One 账号与分组配置；`verification.md` §5 有对应检查项 |
| 推迟 dompurify 后遗忘，直到对外发 key / 多用户时才想起 | 决策 1 写明三条重新评估触发条件；`PORTING-0.1.180.md` §5.1 状态标为「已决定推迟」而不是删掉 |
| `913ec5d74` 漏补 helper ⇒ 编译失败 | 阶段 4 第一步就是补 helper；`tasks.md` 5.1 独立成条 |
| Antigravity 分两次上线 ⇒ 付费账号 401 | 决策 4：合成一个提交 |
| 19 项塞进一个 PR 无法评审 | 按 Migration Plan 的五个阶段拆 |
| `243921dc0` 改流式终端输出重建，影响面是所有 Responses 流式请求 | 单独提交；验收覆盖「上游上报 item 与 delta 累积不一致」的场景 |
| 顺手带入 §6 / §7 的提交 | `tasks.md` §7 有 diff 文件清单门禁 |

## Migration Plan

五个阶段，按「验收独立性 + 风险隔离」划分。

### 阶段 1：静默失效三件套（各自独立提交）

`b1e60ba45` 池模式重试 → `40c26f343` 空 capabilities → `cd05772e9` ops 内存。
三条互不相干，但都要求「先复现失效」的验收（决策 2）。`40c26f343` 上线后进观察期。

### 阶段 2：流式与协议保真

`243921dc0` 终端输出重建 → `bafd2e293` 空 tool name。前者影响面大，单独提交。

### 阶段 3：模型目录与账号端点

`913ec5d74`（**先补 `CodexCanonicalClientVersion()` helper**）→ `f98a056f7`（先做决策 6 的配置检查）
→ `e7a3c1202`+`21c07e835`（合成一个提交）→ `1e1798d90` Composite 视频端点。

### 阶段 4：Ollama Cloud（顺序固定的一对）

`b30651a0a` → `86470628d`。一个提交或两个顺序提交，回滚时一起回滚。

### 阶段 5：会话种子、日志与管理台

`e45490a36` chat 粘性种子 → `f6aa9dc3c` 日志量 → `3445485eb` token 刷新 →
`ee62dfbaf` / `5dfad32b8` / `616df479e` / `cfecc8d11`+`e4f869e0c` 管理台四项 → `98c7b0e88` 文档。

### 回滚

除阶段 4 外，每项都是独立的纯代码修正，`git revert` 即可。

- 阶段 4（Ollama Cloud）：两条必须一起 revert（决策 5），否则留下引用不存在函数的代码。
- 无数据副作用、无 schema、无配置迁移、无依赖版本变更，因此不需要回滚脚本。
