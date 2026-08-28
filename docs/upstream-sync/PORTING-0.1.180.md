# 移植清单：上游 v0.1.180

对照日期：2026-08-24。「本仓库现状」结论均在 `f8f257ad7` 上逐条 grep / `git apply --check` 核实。
合完一项就把状态改成「已合」。

通用流程见 [README.md](./README.md)，上一轮清单见 [PORTING-0.1.179.md](./PORTING-0.1.179.md)（P0/P1 已全合，
第 7 节按需项已决策完毕；其第 9 节的长上下文计费门控仍未决，见本文第 9 节）。

移植前先读 README [第 4 节「硬约束」](./README.md#4-硬约束)，尤其 9–12 条；写 SQL 对照
[第 5 节转换速查](./README.md#5-pg--sqlite-转换速查)。

---

## 1. 版本对照

| 点 | 状态 |
|---|---|
| 上游最新正式版 | [`v0.1.180`](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.1.180)（tag commit `c40edb407`，2026-08-24 07:16 UTC） |
| 上一轮对照基线 | `v0.1.179`（tag commit `75f88be5f`，2026-08-20 06:53 UTC） |
| `v0.1.179` → `v0.1.180` | **162 commits（119 个非 merge）/ 481 文件** |
| 本仓库基线 | `f8f257ad7`（= v0.1.179 + 已移植项） |
| 本仓库版本号 | `backend/cmd/server/VERSION` = `1.1.2`（自有编号，不要同步成 0.1.180） |
| 本仓库迁移号 | 已用到 `224`；上游 0.1.179 的 226/227/228 继续跳过，0.1.180 新增 229/230（插件）若做要顺延成 **225/226** |
| Go | 本仓库 `1.26.5`；上游本版 `1.26.6` → **`1.27.0`**，见第 9 节 |

对比链接：<https://github.com/Wei-Shaw/sub2api/compare/v0.1.179...v0.1.180>

**结论：119 个非 merge commit 里 28 个可对当前工作区逐字 `git apply`（⇒ patch site 字节一致、同一缺陷确实存在），
建议本轮吃下约 20 条 P0 + 11 条工具桥接 P1；两个大功能（插件系统、模型广场分时价）不做。**

**2026-08-26 复核后的落地口径**：第 5 节的 21 项 P0 逐条重核（方法见第 2 节，多了按文件切开与
依赖符号两道关），**21 项全部仍然成立**；其中 2 项前端依赖安全项已决定推迟，实际交付 **19 项**。
已固化为 OpenSpec change
[`port-upstream-0.1.180-p0-fixes`](../../openspec/changes/port-upstream-0.1.180-p0-fixes/)。

### 1.1 按簇统计

| 簇 | commits | 可逐字 apply | 文件 | 规模 | 判断 |
|---|---|---|---|---|---|
| 小而独立的 bugfix | 23 | 18 | 63 | +2484 −193 | **P0 全收**（第 5 节）；21 项里 19 项交付、2 项依赖安全项推迟 |
| Responses/Chat 工具桥接修复 | 11 | 2 | 33 | +1922 −108 | **P1**（6.1） |
| Grok 4.6 全套（0.1.179 §11 挂起） | 30 | 4 | 64 | +2840 −1644 | **P1，整套或不动**（6.2） |
| PR #5888 + #5925 兼容大礼包 | 12 | 1 | 163 | +16936 −2498 | P2（7.1） |
| Fast mode `service_tier` + 只降不升计费 | 4 | 0 | 30+ | +2231 −180 | P2（7.2） |
| OpenAI 重置卡按阈值自动使用 | 2 | 0 | 32 | +1913 −81 | P2（7.3） |
| 模型广场阶梯 / 分时价 | 8 | 0 | 38 | +3118 −528 | 大部分 N/A（第 8 节） |
| OAuth 出站传输插件系统 | 5 | 1 | 99 | +7802 −148 | **不合**（第 8 节） |
| Go 1.27.0 + golangci-lint v2.13 | 2 | 0 | 27 | +139 −84 | 独立决策（第 9 节） |
| 国产供应商（Kimi / GLM / DeepSeek） | 10 | 0 | 24 | +1361 −127 | **N/A**（第 8 节） |
| deploy compose / docker | 4 | 0 | 12 | +241 −47 | 基本 N/A（第 8 节） |
| chore（VERSION / sponsors / partners） | 2 | 0 | 6 | +1 −37 | 不合 |

---

## 2. 本轮核查方法（比 0.1.179 §12 更省时，建议后续沿用）

上游 compare API 在这个体量上会 500（Unicorn 页），改用两条通道：

```bash
# 1) 整段 diff（2.5 MB，可离线 grep / 按文件切开）
curl -sSL -o /tmp/up180.diff \
  "https://github.com/Wei-Shaw/sub2api/compare/v0.1.179...v0.1.180.diff"

# 2) 只要提交元数据 + 树的部分裸克隆：12 MB、十几秒，之后 git show 按需惰性取 blob
git clone --bare --filter=blob:none --no-tags --single-branch --branch main \
  https://github.com/Wei-Shaw/sub2api.git /tmp/up180.git
git -C /tmp/up180.git log --format='%H %s' --name-only --no-merges 75f88be..c40edb4
```

比「未认证 API 每小时 60 次逐条拉 patch」快得多，也不动本仓库的 remote 配置。

判定「本仓库有没有这个 bug」的机械化做法——逐 commit 生成 patch 后对工作区试应用：

```bash
git -C /tmp/up180.git show --format='' --binary <sha> > /tmp/pc/<sha>.patch
git apply --check -p1 /tmp/pc/<sha>.patch     # 通过 ⇒ patch site 与上游字节一致
```

**注意三种假信号：**

- 通过不等于「该合」。上游按自己的历史顺序生成，前置 commit 没合时后面那条也可能失败；反过来，
  像 `f7145c750`（Grok 默认模型设置迁移）能干净应用，但它属于一个本仓库尚未拍板的策略决定，见 6.2。
- 失败不等于「不该合」，多半只是同一文件被本轮多条 commit 或本仓库自有改动动过。按文件看冲突归属再决定。
- **通过也不保证能编译**。`apply --check` 只比上下文，不看新代码引用的符号在不在。本节的
  `913ec5d74` 就是实例：目标文件干净，但它调用的 `CodexCanonicalClientVersion()` 本仓库零命中。

### 2.1 复核时补的两道关（2026-08-26，建议后续沿用）

上面那套只做到「整 commit apply」，两个盲区：整 commit 一旦在测试文件上失败就看不出产品代码
对不对得上；以及上面第三种假信号。补两步：

```bash
# 1) 按文件切开，逐个 apply --check，产出三态而不是两态
awk '/^diff --git /{n++; f=sprintf("%s/%03d.patch",out,n)} n{print > f}' \
  out=/tmp/split180/<sha> /tmp/pc180/<sha>.patch
# ⇒ NOFILE（本仓库没这个文件）/ CONFLICT / ok

# 2) 对每个 patch 的 + 行提取函数调用标识符，逐个确认已定义（或由同一 patch 定义）
grep -rnE "func (\([^)]*\) )?<id>\b" --include=*.go backend/
```

⚠️ **部分裸克隆对 0.1.180 这批较早提交会逐 blob 拉取、非常慢**（实测跑不完）。改用逐 commit 的
HTTPS patch，23 次请求几十秒；这条不是 API，没有每小时 60 次限制：

```bash
curl -sSL -o /tmp/pc180/<sha>.patch "https://github.com/Wei-Shaw/sub2api/commit/<sha>.patch"
curl -sSL "https://raw.githubusercontent.com/Wei-Shaw/sub2api/c40edb4/<path>"   # 单文件按需取
```

成对项的依赖方向别靠推断，用临时工作区实测：

```bash
git worktree add -q --detach /tmp/wt180 HEAD
cd /tmp/wt180 && git apply -p1 A.patch && git apply --check -p1 B.patch   # 两种顺序各试一次
```

---

## 3. 基线缺口（延续 0.1.179 第 2 节，本轮复核）

| 底座 | 上游 | 本仓库 | 核实方式 |
|---|---|---|---|
| Kimi / 智谱 GLM / DeepSeek 平台 + `api_protocol` | 有 | **没有** | `internal/domain/constants.go:20-27` 仍只有 anthropic / openai / gemini / antigravity / grok / composite |
| 渠道监控配额模式、渠道倍率 / 分时定价 | 有 | **没有** | `custom_channel_time_pricing.go`、`components/admin/channel/TimePricingSection.vue` 等文件不存在（0.1.179 §8 判为不合） |
| 插件系统 | 0.1.180 新增 | **没有** | `plugin_manager.go` / `pkg/pluginapi` / `sub2api_plugin_*` 表全部零命中 |
| **Ollama Cloud 账号** | 有 | **有** | `account_ollama_cloud_usage_*`、`upstream_billing_probe.go` 等；⇒ 本版两条 Ollama 修复**适用** |
| OpenAI 重置卡（手动） | 有 | **有** | `openai_quota_reset_credits.go`、`openai_quota_service.go`、admin 重置端点、`OpenAIQuotaResetCell.vue`；缺的只是阈值自动触发 |
| OpenAI 按上游服务档位计费 | 有 | **有** | `00a818f` / `6eedaa4`（0.1.178 已合）；缺 `service_tier_billing.go` 与客户端可传 `service_tier` |
| Composite 分组 | 有 | **有** | 含 0.1.179 已合的 Codex 端点（`777f368cd`） |

---

## 4. 建议顺序

**2026-08-28 进度**：① 已全合；② 8 条里已合 4 条（`4d4a0be1a` / `cc894ef57` /
`25da02ddd` + `66808413d`）；③ 已全合（(a)(b) 见 §9.1）；④ 已合 `3fd66a33b` 与 `68653fb2c`，
`d5824f6a5` 判 N/A。剩余：②的另 4 条、6.2(c) 整簇、⑤⑥。

```text
① 第 5 节 P0（19 项，5.1 与 nanoid 已推迟）—— 按 OpenSpec change 的五个阶段走：
   阶段 1 静默失效三件套（5.2 池模式重试 / 40c26f343 空 capabilities / 5.3 ops 内存）
   阶段 2 流式与协议保真（243921dc0 / bafd2e293）
   阶段 3 模型目录与账号端点（先补 CodexCanonicalClientVersion → 913ec5d74 →
          f98a056f7 先做配置检查 → e7a3c1202+21c07e835 同一提交 → 1e1798d90）
   阶段 4 Ollama Cloud 一对（b30651a0a → 86470628d，顺序固定）
   阶段 5 会话种子 / 日志 / token 刷新 / 管理台四项 / 文档
② 6.1 Responses/Chat 工具桥接 11 条（先 4d4a0be1a：PDF 附件被静默丢弃）
③ 6.2 Grok 一套：先 ed4207a16（别名 + grok-3-mini 价卡两个真 bug），
       再决定 39485f2e2 + f7145c750（默认模型 4.5 → 4.6）
④ 6.3 单条小项（3fd66a33b 调度诊断 / 68653fb2c / d5824f6a5）
⑤ 按需：7.2 Fast service_tier / 7.3 重置卡自动使用
⑥ 有整块时间再吃：7.1 PR #5888 大礼包
不做：插件系统、模型广场分时价、CN 供应商、compose 四条
```

第 9 节三个决策点（Grok 默认模型 / Go 1.27 / 0.1.179 遗留的长上下文门控）独立拍板，不排在流程里。

---

## 5. P0 — 本仓库已核实有同一缺陷，patch 逐字对得上

除 5.1 外全部 `git apply --check` 通过。

本节 21 项已固化为 OpenSpec change
[`openspec/changes/port-upstream-0.1.180-p0-fixes/`](../../openspec/changes/port-upstream-0.1.180-p0-fixes/)：
`specs/*/spec.md` 是移植后必须成立的行为（21 条 Requirement / 77 个 Scenario，覆盖交付的 19 项），
`tasks.md` 是按文件的实施清单，`verification.md` 是验收证据矩阵。合完一项后把下面对应条目的
状态改成「已合」。

📌 **其中 2 项前端依赖安全项（5.1 dompurify、5.4 的 `b410c3913` nanoid 审计例外）已决定推迟**，
不在该 change 的交付范围内，理由与重新评估触发条件见其 `design.md` 决策 1。

⚠️ **2026-08-26 复核补充两点**（原清单的 `apply --check` 通道看不出来，详见该 change 的
`source-baseline.md` §4）：

1. **`913ec5d74` 引用了本仓库没有的 `CodexCanonicalClientVersion()`**，`apply --check` 通过但
   **不编译**。上游该函数体只有一行 `return resolveCodexOutboundIdentity("").version`，
   本仓库 `openai_codex_identity.go:108` 有 `resolveCodexOutboundIdentity`，补同名 helper 即可。
   ⇒ 凡新增函数调用都要额外 `grep -rn "func .*<name>"`，别只看 `apply --check`。
2. **`e7a3c1202` 与 `21c07e835` 的落地顺序要与上游相反**。上游先合路由（`e7a3c120`）、后修 URL
   （`21c07e83`），中间那段时间付费账号被打到 `daily-cloudcode-pa.sandbox.googleapis.com` → 401
   「Invalid bearer token」，正是 #3611 / #2962 那个坑。本仓库**必须同一提交落地，或先合 URL 修正**。

另外实测确认：`86470628d` 单独 apply 报 NOFILE（目标文件由 `b30651a0a` 新建），顺序固定；
`cfecc8d11` 与 `e4f869e0c` 两种顺序下都 clean，顺序无关。

### 5.1 dompurify `3.3.1` → `3.4.14`（安全项，**已决定推迟**）

上游 `4a1da2950`。2 文件 +20-64。状态：**已决定推迟**（2026-08-26）

CVE-2026-65913 / GHSA-cj63-jhhr-wcxv：`USE_PROFILES` 打开时 `ALLOWED_ATTR` 被重建成普通数组并用
`ALLOWED_ATTR[lcName]` 查表，被污染的 `Array.prototype` 属性（如 `onclick`）会被当成白名单属性存活。

**本仓库命中同一条路径**：`frontend/src/utils/sanitize.ts:5` 就是
`DOMPurify.sanitize(svg, { USE_PROFILES: { svg: true, svgFilters: true } })`，输出经 `v-html` 落在
`components/common/ImageUpload.vue:14`（SVG 上传预览）和 `components/layout/AppSidebar.vue:95/120/140`
（自定义侧栏图标——本 fork 的隐藏菜单/自定义菜单会走这里）。`frontend/package.json:25` 当前是 `^3.3.1`。

| 文件 | 动作 | 改什么 |
|---|---|---|
| `frontend/package.json` | 改 | `dompurify: ^3.4.14`，并在 `pnpm.overrides` 加 `"dompurify@<3.4.14": ">=3.4.14"`（让 mermaid 传递依赖的那份也去重到同一版本） |
| `frontend/pnpm-lock.yaml` | 重新生成 | ⚠️ **不要抄上游 lockfile**（它是 pnpm 9 产物，本仓库用 pnpm v11）。本地 `pnpm install --lockfile-only` 即可，只重写 lockfile、不装 node_modules |

⚠️ 原清单漏了一处：`frontend/pnpm-workspace.yaml` 里**也有一个 `overrides:` 块**，注释写明
「pnpm v11 reads overrides here; Docker/CI still use pnpm 9 which also honors
package.json.pnpm.overrides. Keep both in sync.」——本机 pnpm 11 读 workspace 那份，CI（pnpm 9）
读 `package.json` 那份。**两处都要加**，只加一处会「本地去重了、CI 没去重」。

**推迟决定（2026-08-26）**：这个绕过需要页面里另有 prototype pollution 原语才可利用；本仓库
`sanitizeSvg` 的全部输入都是管理员自填（`AppSidebar.vue` 的自定义侧栏图标来自菜单设置，
`ImageUpload.vue` 的 `mode='svg'` 只出现在 `SettingsView.vue` / `RiskControlView.vue` /
`AccountTestModal.vue` 三个管理页），单管理员部署下实际可达性低。
**推迟意味着这两个文件一个都不动**——三个 workflow 与 `Dockerfile` 都用
`pnpm install --frozen-lockfile`，只改 `package.json` 会让 CI 与镜像构建直接失败。
重新评估的触发条件（对外发 key / 出现第二个管理员 / 发现 prototype pollution 原语 /
因其他原因顺带升级传递依赖）见 change 的 `design.md` 决策 1。

### 5.2 池模式同账号错误重试丢失（两条 compat 路径）

上游 `b1e60ba45`。3 文件 +88-6。状态：**已合**

本仓库现状：`gateway_forward.go:699` / `:733`（原生 Anthropic 路径）已经在填
`RetryableOnSameAccount: account.IsPoolMode() && account.IsPoolModeRetryableStatus(...)`，但
`gateway_forward_as_chat_completions.go:168-174` 与 `gateway_forward_as_responses.go` 的同一处
只返回 `StatusCode` / `ResponseBody`，**池模式账号在 CC / Responses 上游报错时永远不会同号重试**。

上游顺带用上了 `HandleUpstreamError` 的返回值（`ratelimit_service.go:269` 签名已是
`(shouldDisable bool)`）：刚被判定要禁用的账号不再进入同号重试。

### 5.3 ops 面板混用 cgroup 与 host 内存

上游 `cd05772e9`。2 文件 +160-35。状态：**已合**

本仓库现状：`ops_metrics_collector.go:602` 取 cgroup 三元组，`:622` 就是那句
「If total memory isn't available from cgroup (e.g. memory.max = "max"), fill total from host」。
Docker + cgroup v2 且未设内存上限时，`used` 是容器数、`total` 是宿主机数，面板显示
容器 60 MB / 宿主 23 GB ≈ 0.3%，严重低估。上游改成 `resolveMemoryStats`：只有 cgroup
同时给出 current 和具体 limit 时才整组用 cgroup，否则整组回落 host，两个来源不再混。

### 5.4 其余 P0（逐条已 `apply --check` 通过）

| 上游 commit | 内容 | 规模 | 状态 | 备注 |
|---|---|---|---|---|
| `3445485eb` | 前端 token 刷新锁死循环 | 2 文件 +17-12 | **已合** | 纯删 `api/tokenRefresh.ts` 12 行 + 回归测试；CPU 空转 |
| `40c26f343` | 空 `openai_capabilities` 不再把 OAuth 账号排除出文本调度（#5530） | 2 文件 +76-0 | **已合** | `service/account.go` +15；调度可用性 |
| `e45490a36` | 动态 system message 下 chat 粘性 hash 不再漂移 | 2 文件 +83-11 | **已合** | `openai_content_session_seed.go` |
| `913ec5d74` | OAuth 账号自动同步模型 | 2 文件 +121-0 | **已合** | `upstream_models.go`；⚠️ **先补 `CodexCanonicalClientVersion()`**，否则不编译（见本节开头） |
| `f98a056f7` | 收紧 Google One 模型目录 | 6 文件 +142-1 | **已合** | `pkg/geminicli/models.go` 新增约束。⚠️ 这是**收窄**（3.x / image 模型从 Google One 账号的可见清单里移出）：上线前先查现有 Google One 账号与分组白名单 / `model_pricing` 有没有依赖将被移出的模型 ID，否则会留下孤儿条目 |
| `243921dc0` | 按上报 item 重建流式终端输出 | 2 文件 +206-5 | **已合** | `openai_gateway_response_handling.go` |
| `bafd2e293` | 流式 arguments delta 不再带空 tool name | 2 文件 +47-1 | **已合** | `apicompat/types.go` |
| `e7a3c1202` + `21c07e835` | Antigravity 付费账号改走官方 daily 端点 | 2+2 文件 | **已合** | `antigravity_gateway_retry.go`、`pkg/antigravity/oauth.go`；⚠️ **顺序与上游相反**：URL 修正必须同时或在前（见本节开头） |
| `1e1798d90` | Composite 分组放行视频生成端点 | 2 文件 +16-1 | **已合** | `server/routes/gateway.go` |
| `b30651a0a` | Ollama Cloud CC 思维字段对齐 `reasoning_content` | 3 文件 +445-0 | **已合** | 新建 `openai_gateway_ollama_cloud_cc_reasoning.go`；本 fork 有 Ollama Cloud |
| `86470628d` | Ollama Cloud 账号 clamp `max_tokens` | 3 文件 +245-1 | **已合** | 依赖上一条，先后顺序固定 |
| `ee62dfbaf` | 批量代理解析支持 `[IPv6]` | 2 文件 +62-4 | **已合** | `views/admin/ProxiesView.vue` |
| `5dfad32b8` | 用户并发数 0 = 不限（编辑弹窗） | 4 文件 +109-5 | **已合** | 含中英文案 |
| `616df479e` | 账号优先级列默认展示 | 2 文件 +153-1 | **已合** | |
| `f6aa9dc3c` | `prompt_guard.config_loaded` 只在变化时记日志 | 2 文件 +73-5 | **已合** | `securityaudit/prompt_config_store.go`；1C1G 磁盘友好 |
| `cfecc8d11` + `e4f869e0c` | 运维错误详情「返回列表」+ 保留筛选状态、兼容展示 | 6+5 文件 +228-13 | **已合** | 纯前端，本仓库 ops 页面组件齐全 |
| `b410c3913` | nanoid 审计例外 GHSA-2v37-7h3g-55p8 | 1 文件 +7 | **已决定推迟** | `.github/audit-exceptions.yml`（与 5.1 同批）。另注：该文件里 `lodash` / `lodash-es`（`2026-07-02`）与 `axios`（`2026-07-10`）三条例外**已过期**，而 `tools/check_pnpm_audit_exceptions.py` 对过期条目同样返回 1 ⇒ `security-scan.yml` 很可能已经红了约七周，与 nanoid 无关，需单独处理 |
| `98c7b0e88` | 文档自引用 URL 修正 | 1 文件 | **已合** | 顺手；无 Requirement，只在 change 的 `tasks.md` |

---

## 6. P1 — 值得做，要拆开手工移植

### 6.1 Responses / Chat 工具桥接修复（11 条，33 文件 +1922-108）

只有 `bafd2e293`（已列在 5.4）和另一条 apply 干净，其余因同文件多次改动需按序落地。

**2026-08-28：本簇 8 条里已合 4 条**（`4d4a0be1a` / `cc894ef57` / `25da02ddd` / `66808413d`），
见 OpenSpec change [`port-upstream-p1-tool-bridge-and-composite-dispatch`](../../openspec/changes/port-upstream-p1-tool-bridge-and-composite-dispatch/)。
剩下 4 条都落在 `chatcompletions_responses_bridge.go` / `openai_gateway_responses_chat_fallback.go` /
`responses_client_tools.go` / `openai_gateway_grok.go` 这几个本仓库有自有改动的文件上，需逐 hunk 解。

| 上游 commit | 内容 | 规模 | 状态 / 本仓库现状 |
|---|---|---|---|
| `4d4a0be1a` | `/v1/chat/completions` 的 `type:"file"`（PDF）不再被静默丢弃，转成 Responses `input_file` | 3 文件 +97-2 | **已合**（`ddc63c7ef`）。曾确认有此 bug：`apicompat/types.go:91` 的 content part 分支只有 `"text"` / `"image_url"`，全仓 grep `input_file` 零命中。表现是请求 200、模型照答，但 prompt 里没有那份文件。⚠️ `types.go` 的 hunk 手写落地（本仓库 `x_search` 字段导致上下文漂移，结构体本身与上游一致） |
| `e2d9ce0ca` + `fbc9ee626` | 拒绝非法 tool-call arguments（第二条收窄第一条的范围） | 6+5 文件 | 两条必须一起，否则行为过宽 |
| `cc894ef57` | 剥掉流式 tool_call 的空 `id` / `function.name` | 4 文件 +363-0 | **已合**（`2f091ed66`，4/4 文件逐字 apply）。上游举的例子是 DashScope/DeepSeek，但受害面是**任何**「后续 delta 送空 id」的上游：客户端按 `!== undefined` 合并会覆盖首个 delta 的身份，最后去调一个名为 `""` 的工具。新建 `openai_gateway_cc_tool_call_identity.go` |
| `31d5b67ba` | 恢复带命名空间的自定义工具别名 | 6 文件 +220-23 | 5/6 文件干净，只 `openai_gateway_responses_chat_fallback.go` 冲突——与 `e2d9ce0ca`+`fbc9ee626` 同一文件族，一起排期 |
| `7a09a2eaf` | 清掉孤儿 deferred 工具标记 | 6 文件 +109-0 | 动 `openai_gateway_grok.go`（本仓库该文件有自有改动，注意冲突） |
| `7498d8fdc` | Responses Lite 强制串行工具调用 | 4 文件 +216-10 | 是 [PORTING-0.1.183.md](./PORTING-0.1.183.md) §5.1 那 5 条的基座；要做就整簇做 |
| `25da02ddd` | 避免 HTTP bridge 重复回放 | 5 文件 +240-9 | **已合**（`a3f0c978b`，与下一条同批）。产品代码 4/4 逐字 apply；`openai_ws_http_bridge_test.go` 的两个新增测试因上游锚点不存在而追加到文件末尾 |
| `66808413d` | 丢弃孤儿回放 tool call | 3 文件 +128-2 | **已合**（`a3f0c978b`）。依赖上一条，**顺序不可交换**：先合它会让过滤作用在旧判据上 |

实际在用 Codex 的场景，这一簇价值最高。

### 6.2 Grok 一套（30 条，64 文件 +2840-1644）—— 0.1.179 §11 挂起的那串，现在可以决策

上一轮的判断是「默认模型迁移和计费目录是一套，等上游打 tag 再整体评估」。现在 tag 出来了，
**而且里面有两块可以分开**：

**(a) `ed4207a16` 校正 Grok 模型目录计费与工具出站 —— 两个真 bug，建议单独先合**

5 文件 +62-24。`git apply --check` 只在 `openai_gateway_grok.go:1173` 冲突（本仓库该文件有自有改动），
`xai/models.go` 与 `billing_service.go` 两处干净。

状态：**已合**（2026-08-27），见 [`resolve-pending-decisions-and-p1-fixes`](../../openspec/changes/resolve-pending-decisions-and-p1-fixes/)。

⚠️ **2026-08-27 复核补充**：`openai_gateway_grok.go` 的那个 hunk 不是「冲突」而是**不可移植**——
它调用 `grokSameAccountRetryMetadata`，本仓库 grep 零命中（属下面 (c) 那簇 Grok 429 同号重试）。
⇒ **该文件的 hunk 整个丢弃**，只取 `xai/models.go` 与 `billing_service.go`。
另外 (b) 的 `39485f2e2` 会改 (a) 刚引入的别名收窄条件、并重排 (a) 刚加过 case 的价卡 switch，
⇒ **顺序固定 `ed4207a16` → `39485f2e2` → `f7145c750`**。

1. **别名被无条件重映射**：`ModelMappingWithOptions` 与 `ResolveGrokTextResponsesModelID` 原先对
   *所有*指向 `DefaultTextModel` 常量的别名都替换成运行时默认，于是客户端**显式**要 `grok-4.5`
   也会被改写成运营方配的默认模型。上游收窄成只有 `grok` / `grok-latest` / `grok-build-latest` 跟随默认。
   本仓库 `pkg/xai/models.go:112-113`（`"grok-4.5": DefaultTextModel`）+ `:165` / `:276` 就是改前形态。
2. **`grok-3-mini` 系列超收**：本仓库 `billing_service.go` 没有 `grok-3-mini` / `grok-3-mini-fast`
   价卡（`fallbackPrices` 只有 4.5 / 4.6 / 4.3 / build-0.1），而 `isGrokUnknownTextFamilyModel`
   会把 `grok-3-*` 判成未知 Grok 文本族 → 兜到 `billing_service.go:860` 的 `grok-4.5`（$2 / $6）。
   官方是 $0.30 / $0.50 ⇒ **输入超收 6.7×、输出超收 12×**。而 `grok-3-mini` / `grok-3-mini-fast`
   仍在 `defaultModels` 里对客户端可见（上游本条同时把它们从 `/v1/models` 摘掉、只留别名 + 价卡）。
3. 顺带修 Imagine 模型 ID 反了（`grok-imagine-video-1.5` 与 `-preview` 互换）并加
   `grok-imagine-image-2.0`。

**(b) 默认模型 4.5 → 4.6 —— 策略决定，两条要一起**

| 上游 commit | 内容 | 规模 | 注意 |
|---|---|---|---|
| `39485f2e2` | `DefaultTextModel = "grok-4.6"` + 官方计费目录（新增 `grok-4.20` 独立价卡、`grokUnknownTextFamilyFallback` 改指 4.6）+ `setting_parse.go` / 前端白名单 | 10 文件 +87-38 | 0.1.176 清单当时**明确要求保持 `grok-4.5`** |
| `f7145c750` | 启动时把存量 DB 里 `grok_default_text_model = "grok-4.5"` 的设置重写成 `grok-4.6` | 4 文件 +67-2 | ⚠️ **这条 apply 干净但不要顺手合**：它是上面那个决定的数据迁移。设置未写过时（`ErrSettingNotFound`）为 no-op，只影响显式存了 `grok-4.5` 的库 |

**(c) 其余 Grok 稳定性（Realtime / 429 / 容量 / 媒体）**

Realtime 预握手复用与切号（`61c2f5ad2` `611a7c8ed`）、握手失败账号冷却（`d78e366db`）、
普通 429 有限同号重试（`8db8791a7` `2ab24a1e7` `0b1f79c83`）、compaction 422 重试
（`17c0ee385`，仅 1 行 +1-1，apply 干净但**不可独立移植**，见下）、CC bridge 同号重试（`5ae254f77` `ad87ddee1`）、
stream idle 重试上限作用于主路径（`c628b3eea`）、容量重试与兼容性分类收紧（`953028718` `39aaf2fea`
`0e05c61d3`）、传输超时与握手（`5ade09431`）、媒体超时与内容拒绝计费（`e85348be8` `2e68b10aa`）、
以及 CI 回归修正（`787f875dd` `2ab41b92b` `1bff06ea5` `cca235365` `f7bc1970e` `3243983b7` `3b8177642`）。

⚠️ 两组成对的 revert，**不要只拿前一半**：

- `6c3edc095` feat(429): configurable cooldown and retry strategies ←→ `e62ec2c42` Revert
- `726de3010` 修复 Grok WebSearch SSE action 兼容 ←→ `ab9cb69e7` Revert

⚠️ `16b15e870`「修复 5888 与 5925 的同号重试语义冲突」说明这一簇和 7.1 的大礼包**互相打补丁**，
两者要一起排期，先合哪个都要把这条一并带上。

⚠️ **`17c0ee385` 单独合是 no-op，2026-08-28 实测确认，不要再当独立小项挑出来。**
它改的是 `forwardGrokResponses` 的**外层**守卫（400 → 400/422），而真正决定要不要剥掉加密
reasoning 重试的**内层**判定 `isGrokInvalidEncryptedContentResponse`
（`openai_gateway_grok.go:255`）在本仓库仍是 `if statusCode != http.StatusBadRequest { return false }`
⇒ 422 进了外层也会被内层否掉，净效果只是把响应体多读一遍再放回去。
widen 内层的是本簇的 **`953028718`**（实测：`v0.1.180` 区间里只有 `17c0ee385` 与 `953028718`
两个 commit 含 `StatusUnprocessableEntity`），它同时引入 compaction 错误码
（`invalid_compaction` / `compaction_decode_error`）与 `grokStructuredErrorMessageCandidates`
（本仓库零命中）——也就是说标题里的「compaction」识别能力全在 `953028718` 里。
⇒ **随本簇整体排期。** 详细证据见 OpenSpec change
[`port-upstream-p1-tool-bridge-and-composite-dispatch`](../../openspec/changes/port-upstream-p1-tool-bridge-and-composite-dispatch/) 的 `design.md` 决策 7。

📌 这是 §2「三种假信号」之外的**第四种**：`apply --check` 通过、符号齐全、编译通过，**但行为为空**。
今后判定单行守卫类改动，要连同它守卫的下游判定一起看。

### 6.3 单条小项

| 上游 commit | 内容 | 规模 | 判断 |
|---|---|---|---|
| `3fd66a33b` | 调度「无可用账号」诊断：boolean 门改成返回具体 veto reason（`model_rate_limited` / `quota_auto_pause_<window>` / `platform_mismatch` / …） | 2 文件 +118-16 | **已合**。行为不变、纯可观测性。⚠️ 与 7.3 重置卡功能同改 `openai_gateway_scheduling.go`，后续 7.3 需 rebase。 |
| `68653fb2c` | Composite 分组的 `/v1/messages` 闸门改为尊重分组自己的开关（原先 `sanitizeGroupMessagesDispatchFields` 对 composite 恒置 false） | 7 文件 +70-23 | **已合**（`b21a2df03`）。产品改动只有 2 行 + handler 豁免收窄 + 前端表单；⚠️ 上游 handler hunk 的 `IsCNProvider` 两处分支整段丢弃，且**不要**为对齐上游把 `allowOpenAICompatibleMessagesDispatch` 的 `ctx context.Context` 签名改成 `*gin.Context`。存量 composite 分组的落库值仍是 false，升级不会自动放宽 |
| `d5824f6a5` | 保留原生 `reasoning_effort: max` | 10 文件 +81-14 | **N/A**（2026-08-28 实测定性）。它新增的 `supportsOpenAIReasoningEffortMax` 在 `isOpenAIGPT56Model` 之外只放开 `deepseek-v4` / `glm-` / `kimi-` / `moonshot-` / `k3` 五个前缀，全是本仓库没有的平台；剩下的是把 mappedModel 透进两个 extractor，而本仓库主路径 `openai_gateway_request_body.go:825` 早就在传 `firstNonEmpty(modelCandidates...)`，被补的两条是 CC/Responses → Anthropic 原生上游的路径，那里 mappedModel 是 claude 模型、归一化结果不变 ⇒ **对本仓库空转，不做** |
| `d493ce0bb` + `fa4587041` | Codex 账号身份限定到 OAuth 账号 / auto-review 留在母账号 | 18+7 文件 +1392-68 | **只有在用 spark 影子账号时才需要**。不用就别动，它铺开 25 个文件 |

---

## 7. P2 — 大功能，先确认真的用得上

### 7.1 PR #5888 + #5925 OpenAI / Grok 兼容大礼包 —— 本版最大单项

`cf3577a3c` `acce29af2` `c374ff295` `ccb20ace8` `b2b2adcf8` `1429e8f71` `16b15e870` `7e9af4c10`
（+ CI 收尾 `2ab41b92b` `4eadee107` `1591477a3` `269a40924`）。**163 文件 +16936-2498。**

新增 20 个文件，等于一次进来五六个子系统：

- `openai_apikey_health_breaker.go` + `setting_openai_apikey_health.go`（API key 健康熔断器，含新设置项）
- `openai_compact_fallback.go`（compaction 回落）
- `openai_responses_ingress_compat.go` / `openai_responses_input_compat.go`（ingress / input 兼容层）
- `openai_ws_session_preemption.go`（WS 会话抢占）
- `openai_codex_tool_names.go`、`openai_cyber_transcript.go`、`openai_gateway_grok_model_input.go`、
  `openai_json_decode.go`、`handler/request_body_read_log.go`

对重度 Codex / Responses 用户价值最高，但：

- 冲突最重的是 `handler/ops_error_logger.go`（本簇 +1032-408），而本仓库刚为 SLA 口径改过这个文件
  （`a97c6ff80`、`241eff90c`），必须逐 hunk 解。
- 和 6.2 的 Grok 串交织（`16b15e870`）。
- 至少是多天量级的工作。**本轮不建议开。**

### 7.2 Fast mode `service_tier`

`f06bf181d` + `c0c3e1cb4` + `e457f0fa2`（25+3+4 文件 +1584-134）与 `75faedda9`
「fast/priority 按上游响应实际档位只降不升计费」（25 文件 +647-46）。

本仓库已有「按上游服务层级计费」（`00a818f` / `6eedaa4`），本版增量是：

- `/v1/responses` 与 `/v1/chat/completions` 接受客户端传 `service_tier`
  （`fast|priority` 归一到 priority，另收 `flex|auto|default|scale`；未知 / 空 / 非字符串 → HTTP 400；
  省略与 `null` 保持兼容）
- 全链路透传（JSON / SSE / Responses↔Chat 转换 / 回落路径 / HTTP→上游 WebSocket bridge）
- 计费优先采信**上游终端档位**，出站档位只在上游不带该字段时兜底；上游显式 `default` 即按标准价
- `billing_service.go` 新增 `openAIModelFastPricingRatio` / `enforceOpenAIFastPricingRatio`：
  gpt-5.4 与 gpt-5.6-sol/terra/luna = 2×，gpt-5.5 = 2.5×（本地/远程 LiteLLM 目录可能只带旧口径）
- 新增 `service_tier_billing.go`（本仓库没有此文件）

**判断：只有真在发 fast / priority 请求才划算**，否则是在「没跑 fast 却给计费加代码」的方向上投入。
本仓库 `upstream_response_model.go` 已存在，冲突集中在 `openai_gateway_{passthrough,forward,messages,
chat_completions*}.go` 这几个被本仓库反复改过的文件。

### 7.3 OpenAI 重置卡按用量阈值自动使用

`6f972145b` + `96b160d9a`。32 文件 +1913-81。

底座本仓库**已有**（见第 3 节），缺的是 `openai_quota_auto_reset.go`（+809）、
`openai_quota_auto_reset_config.go`（+148）那套阈值触发 + 前端配置（`EditAccountModal.vue` +82）。

⚠️ 动 `cmd/server/wire.go` 与 `internal/service/wire.go` ⇒ 必须 `go generate ./cmd/server`，
不要手改 `wire_gen.go`（README 硬约束第 7 条）。同时动 `openai_gateway_scheduling.go`，与 6.3 的
`3fd66a33b` 抢同一文件。

**判断：手里真有重置卡才做**；否则是一整套没人触发的后台逻辑。

### 7.4 `847c0c452` 模型列表读取上限可配置 —— 收益很低

14 文件 +240-31，只有 `wire_gen.go` 和 `config_test.go` 两处冲突，技术上很好合。但**本仓库已经硬编码了
8 MB 上限**：`upstream_models.go:104`（`LimitReader(resp.Body, upstreamModelsBodyLimit+1)`）、
`openai_codex_models_service.go:500`、`pkg/antigravity/client.go:696`。这条只是把它变成
`gateway.models_list_read_max_bytes` 配置项，外加 Codex manifest 多读一个哨兵字节以便超限时报明确错误。
个人单机没有调这个值的需求 ⇒ **可跳过**。

### 7.5 `6466978d2` 统一 token 计费路径 + 上下文阶梯单价表查询

6 文件（新增 `billing_context_schedule.go` +470、`billing_token_cost_request.go` +103），
是模型广场那一串里**唯一可以单独摘出来**的一条。不打算做模型广场展示的话没必要。

---

## 8. 不合 / N/A

| 上游条目 | commit | 原因 |
|---|---|---|
| **OAuth 出站传输插件系统** | `40ea3aeba` `26ac0498f` `684d9efb1` `391d69e08` `40aaf7b3a`（99 文件 +7802） | 上游自称实验性。带 `229_plugins.sql` / `230_plugin_artifacts.sql` 两个 **PG 方言**迁移（`BIGSERIAL` / `JSONB` / `TIMESTAMPTZ` / `BYTEA`）要整体重写成 SQLite 并顺延成 225/226；`artifact_data BYTEA` 的存在理由是注释里写明的「供多实例和无状态节点重新复验」——单机用不上。功能上是 gRPC 插件运行时（`pkg/pluginapi/v1/*.pb.go`）+ 本地进程拉起 + 签名校验 + 管理端上传页，在 1C1G 上是纯新增攻击面与常驻内存开销 |
| 模型广场分时 / 渠道定价 | `77e0409f7` `83d4eb6a4` `b07d85c49` `f19095f96` `377d1230f` `ecce0769c` `d9d2854d2` | 依赖 0.1.179 §8 已判「不合」的渠道倍率底座。`service/custom_channel_time_pricing.go`、`components/admin/channel/TimePricingSection.vue` 等 4 个文件本仓库根本不存在。可分离的那一条见 7.5 |
| 国产供应商 10 条 | `695ebede7`（CN Anthropic 用量 token 规范）`cef18b4ad` `b27cd76a8` `b0b2734b0` `f75c4161f` `01a008394`（DeepSeek）`2074fe3ba`（CN 原生 Anthropic reasoning_effort）`a749673de` `2e279c81d`（CN 配额/连接测试）`011745255` | 无 CN 平台、无 `api_protocol` 字段，与 0.1.179 §2 / §8 结论一致 |
| deploy compose 四条 | `9f2f2738f` `10081a812` `e2263d256` `6a1efda0c` | 改的是上游 `docker-compose.yml` / `standalone` / `local` / `dev`，本 fork 走 native systemd + `deploy/docker-compose.sqlite.yml`。**但底层那句「避免默认网关配置被覆盖」值得对 `deploy/config.personal.sqlite.yaml` 自查一遍**，`force_http` 回落的文档说明也值得读一遍 |
| `2bc139ab5` VERSION 同步 0.1.179 | — | 本仓库用自有编号 `1.1.2` |
| `c0e073a79` sponsors、`assets/partners/*`、README star history | — | 上游 README / 素材专属 |
| `354825674` chore: update gitignore | — | 里面是 `/plugins/` 等插件相关忽略项，不做插件就不需要 |

---

## 9. 决策点

### 9.1 Grok 默认模型 4.5 → 4.6

见 6.2(b)。0.1.176 清单当时明确要求保持 `grok-4.5`，`pkg/xai/models.go:54` 至今是
`const DefaultTextModel = "grok-4.5"`。跟或不跟都可以，但要点是：

- **6.2(a) 的两个 bug 与这个决定无关**，可以先合、不影响默认模型。
- 决定跟的话，`39485f2e2` 与 `f7145c750` 必须一起，否则代码默认与存量 DB 设置会不一致。

状态：**已决策并落地（2026-08-27）—— 跟到 `grok-4.6`**。`39485f2e2` + `f7145c750` 成对，
落地顺序在 6.2(a) 之后。见 [`resolve-pending-decisions-and-p1-fixes`](../../openspec/changes/resolve-pending-decisions-and-p1-fixes/)。
⚠️ `f7145c750` 的数据迁移**不可逆**：revert 代码不会把 DB 里已被改写的
`grok_default_text_model` 改回 `grok-4.5`。

### 9.2 Go 1.27.0 + golangci-lint v2.13

上游 `cbe258fd1` + `73aabc861`。不是 bugfix，是要不要跟上游工具链。上游从 `1.26.6` 跳到 `1.27.0`
（本仓库 `backend/go.mod` 是 `1.26.5`，已落后一个 patch）。跟的成本：

- `backend/go.mod` + 三个 workflow 的 `go version | grep -q` 断言 + 三个 Dockerfile 的 golang 镜像 +
  README 徽章 + DEV_GUIDE 同步（CLAUDE.md 也要改：「CI asserts this string」）
- golangci-lint `v2.9` → `v2.13`（v2.9 由 go1.26 构建，会拒绝 go.mod 目标 1.27）；新规则处理见上游
  提交说明（G703/G704 污点分析、`reflect.Ptr` → `reflect.Pointer`、SA4023 / SA1019 加 nolint）
- **`ent/` 要在 Go 1.27 默认 jsonv2 引擎下重新生成**：`json.RawMessage` 字段会生成成
  `jsontext.Value`（`group.model_pricing`、`usage_cleanup_task.filters`）
- `x/net v0.56` 在 go1.27 下包装标准库 HTTP/2：`ConfigureTransports` 经 `RegisterProtocol("http/2")`
  打开 `Protocols.HTTP2` 而不再写 `TLSNextProto`，`ReadIdleTimeout` / `PingTimeout` 映射到
  `HTTP2Config.SendPingTimeout` / `PingTimeout` ⇒ keepalive 相关断言要改

可以先不跟。但拖久了，后续每轮移植都要在两套 ent 生成结果之间手工调和。

状态：**已决策并落地（2026-08-27）—— 跟到 `1.27.0`** + golangci-lint `v2.13`。见 [`resolve-pending-decisions-and-p1-fixes`](../../openspec/changes/resolve-pending-decisions-and-p1-fixes/)。
⚠️ 两处本节原文写少了：版本断言实际是 **5 行**（`backend-ci.yml` 两处、`security-scan.yml` 一处、
`release.yml` 两处），golang 镜像是 **3 个** Dockerfile（根 / `backend/` / `deploy/`）。
`DEV_GUIDE.md:52` 现在写的「三个 workflow ... 两处」也要一并纠正。

### 9.3 遗留：长上下文计费门控 AND → OR（0.1.179 §9）

与 0.1.180 无关。`billing_service.go:1103-1105` 那一行 `applyLongCtx = applyLongCtx && *input.…`。

状态：**已决策并落地（2026-08-27）—— 改成分组为主、账号只能额外开启（区间定价前置条件仍为 AND）**。见 [`resolve-pending-decisions-and-p1-fixes`](../../openspec/changes/resolve-pending-decisions-and-p1-fixes/)。

⚠️ 两条实施要点：

- **不要把 `&&` 直接换成 `||`**。那会让账号开关绕过 `len(resolved.Intervals) == 0` 前置条件
  （区间定价已自含上下文分层）。按上游 v0.1.183 的形状写：先算
  `contextTierPricingEnabled := resolved.longContextPricingEnabled`，账号开关**为真时**置 true，
  再 `applyLongCtx := len(resolved.Intervals) == 0 && contextTierPricingEnabled`。
- **`billing_service.go:1045-1046`（无 Resolver 回退路径）保持原样**，那里拿不到分组开关。

**费用会上升**：分组开关默认开、账号开关默认关 ⇒ 此前不收长上下文倍率的请求之后会收
（2× 输入 / 1.5× 输出）。想保持旧行为的分组把它的 `long_context_pricing_enabled` 关掉。

---

## 10. 不要动

| 项 | 原因 |
|---|---|
| `git merge upstream/main` / 整 PR cherry-pick | 历史无共同祖先 |
| 上游 `migrations/229_plugins.sql` / `230_plugin_artifacts.sql` 原文件 | PG 方言 + 多实例语义；本地要新建 225/226 才行，且插件系统本轮判为不合 |
| 上游 0.1.179 的迁移 226 / 227 / 228 | 沿用 0.1.179 §10 的结论 |
| `f7145c750` 顺手合 | 它是 Grok 默认模型 4.6 的数据迁移，属于 9.1 的决定，不是独立 bugfix |
| 上游 `frontend/pnpm-lock.yaml` | pnpm 9 产物，本仓库 pnpm v11，必须本地重生成（真做 5.1 时用 `pnpm install --lockfile-only`） |
| 推迟期间碰 `frontend/package.json` / `pnpm-lock.yaml` / `pnpm-workspace.yaml` / `.github/audit-exceptions.yml` | 5.1 与 nanoid 已决定推迟，这四个文件一个都不动 |
| 只改 `package.json` 不同步 lockfile | 三个 workflow 与 `Dockerfile` 都用 `pnpm install --frozen-lockfile`，会让 CI 与镜像构建直接失败。没有「只改一半」的中间状态 |
| 改已应用的 `migrations/*.sql` | checksum 不可变 |
| 手改 `wire_gen.go` | 动了 wire 就 `go generate ./cmd/server`（7.3 会触发） |
| 在 `skipSQLiteBackgroundJobs` 里新增服务来「绕过」上游 SQL | README 硬约束第 9 条 |
| `backend/cmd/server/VERSION` 同步成 0.1.180 | 自有编号 |

---

## 11. 自测

沿用 [README 第 3.4 节](./README.md#34-自测再合回-main)：

```bash
cd backend
go build ./...
go test -tags=unit ./... -count=1
golangci-lint run ./...
cd ../frontend && pnpm run typecheck && pnpm run lint:check
cd .. && make test-frontend-critical
```

第 5 节那 19 项的逐条证据矩阵在 OpenSpec change 的
[`verification.md`](../../openspec/changes/port-upstream-0.1.180-p0-fixes/verification.md)，
其中三条**必须先复现失效再验修复**（它们都不抛错，按「代码改了」验收等于没验收）：
5.2 池模式重试、`40c26f343` 空 capabilities、5.3 ops 内存混用。

重点盯：

- 登录 / `user_allowed_groups`（缺表会 503）
- 用量写入（`usage_billing_dedup`）
- 调度冷却与账号 failover（6.3 的 `3fd66a33b`、5.2 的池模式重试都在这条路上）
- `40c26f343` 上线后进观察期：此前被静默排除的 OAuth 账号会重新进入调度，账号池实际容量上升、
  流量分布会变，盯账号级并发与 429 分布
- 前端全量 `pnpm run test:run` 已知有 1 个**与移植无关**的失败文件
  （`src/composables/__tests__/useRoutePrefetch.spec.ts`，5 条），见 PORTING-0.1.179.md §4.6
