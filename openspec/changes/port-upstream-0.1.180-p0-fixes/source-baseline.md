# 源基线固定

复核日期：2026-08-26。基线 `77df363a0`。**实施期间不得改动本文的 SHA。**

## 1. 上游固定点

| 点 | 值 |
|---|---|
| 上游仓库 | `https://github.com/Wei-Shaw/sub2api` |
| 本轮来源版本 | `v0.1.180` = `c40edb4`（2026-08-24 11:58 UTC） |
| 上一版基线 | `v0.1.179` = `75f88be5f` |
| `v0.1.179..v0.1.180` | 162 commits（119 非 merge）/ 481 文件；清单 §5 有 23 个，本 change 交付其中 **21** 个（2 个依赖安全项已推迟） |
| 本仓库基线 | `77df363a0`（= v0.1.179 + 已移植项） |
| 本仓库版本号 | `backend/cmd/server/VERSION` = `1.1.8`（**不同步成 0.1.180**） |
| 本仓库迁移号 | `224`（本批不涉及迁移，**不顺延**） |

取 patch 的两条通道（部分裸克隆对这些较早提交会逐 blob 拉取、很慢，实测走 HTTPS 更快）：

```bash
# 逐 commit 取 patch（23 次请求，非 API，无 60/h 限制）
curl -sSL -o /tmp/pc180/<sha>.patch "https://github.com/Wei-Shaw/sub2api/commit/<sha>.patch"
# 单个文件按需取
curl -sSL "https://raw.githubusercontent.com/Wei-Shaw/sub2api/c40edb4/<path>"
```

## 2. 23 个 commit 的按文件三态

`ok` = 逐字可应用；`NEW` = 上游新增文件（本仓库无同名文件，直接落地）；`NOFILE` / `CONFLICT` 见 §4。

| # | commit | 内容 | 三态 |
|---|---|---|---|
| 1 | `4a1da2950` | dompurify 3.3.1→3.4.14 | `package.json` ok / `pnpm-lock.yaml` **CONFLICT（预期，见 §4.1）**；**已决定推迟** |
| 2 | `b1e60ba45` | 池模式同账号重试（CC + Responses） | 2 ok + 1 NEW |
| 3 | `cd05772e9` | ops 内存不再混用 cgroup/host | 1 ok + 1 NEW |
| 4 | `3445485eb` | 前端 token 刷新锁空转 | 2 ok |
| 5 | `40c26f343` | 空 `openai_capabilities` | 2 ok |
| 6 | `e45490a36` | chat 粘性 hash 不随动态 system 漂移 | 2 ok |
| 7 | `913ec5d74` | OAuth 账号模型同步 | 2 ok，**但缺符号（§4.2）** |
| 8 | `f98a056f7` | Google One 模型目录收紧 | 6 ok |
| 9 | `243921dc0` | 按上报 item 重建流式终端输出 | 1 ok + 1 NEW |
| 10 | `bafd2e293` | 流式 arguments delta 空 tool name | 1 ok + 1 NEW |
| 11 | `e7a3c1202` + `21c07e835` | Antigravity 付费账号走 daily 端点 | 4 ok，**顺序须反转（§4.3）** |
| 12 | `1e1798d90` | Composite 放行视频生成 | 2 ok |
| 13 | `b30651a0a` | Ollama Cloud CC `reasoning_content` | 1 ok + 2 NEW |
| 14 | `86470628d` | Ollama Cloud clamp `max_tokens` | 1 NOFILE + 2 NEW，**依赖 13（§4.4）** |
| 15 | `ee62dfbaf` | 批量代理 `[IPv6]` | 1 ok + 1 NEW |
| 16 | `5dfad32b8` | 用户并发数 0 = 不限 | 3 ok + 1 NEW |
| 17 | `616df479e` | 账号优先级列默认展示 | 1 ok + 1 NEW |
| 18 | `f6aa9dc3c` | `config_loaded` 只在变化时记日志 | 2 ok |
| 19 | `cfecc8d11` + `e4f869e0c` | 运维错误详情返回列表 + 兼容展示 | 10 ok + 1 NEW，顺序无关（§4.5） |
| 20 | `b410c3913` | nanoid 审计例外 | 1 ok；**已决定推迟** |
| 21 | `98c7b0e88` | 文档自引用 URL | 1 ok |

## 3. 复核方法（比 PORTING-0.1.180.md §2 多两道关）

原文档只做了「整 commit `git apply --check`」。本轮补两道：

1. **按文件切开**再逐个 apply，产出三态而不是两态：
   ```bash
   awk '/^diff --git /{n++; f=sprintf("%s/%03d.patch",out,n)} n{print > f}' out=/tmp/split180/<sha> /tmp/pc180/<sha>.patch
   ```
2. **依赖符号 grep**：`apply --check` 只比上下文，不看新代码引用的符号在不在。对每个 patch 的
   `+` 行提取函数调用标识符，逐个 `grep -rE "func (\([^)]*\) )?<id>\b"` 确认已定义（或由同一 patch 定义）。
   这一关抓出了 §4.2。
3. **成对项顺序实测**：用 `git worktree add --detach /tmp/wt180 HEAD` 做临时工作区，实际按两种顺序
   apply，验证依赖方向。这一关抓出了 §4.3、确认了 §4.4 与 §4.5。

## 4. 原文档漏掉 / 需补充的五点

### 4.1 `pnpm-lock.yaml` CONFLICT 是预期的（该项已推迟，现状留档备查）

`frontend/package.json:25` 当前是 `"dompurify": "^3.3.1"`；`pnpm.overrides` **已存在**且有 3 条
（`js-cookie` / `form-data@<4.0.6` / `postcss@<8.5.18`），追加一条即可。lockfile 里同时存在
`dompurify@3.3.1` 与 `dompurify@3.3.3`（后者是 mermaid 传递依赖）⇒ overrides 那条确实必要。
**不要抄上游 lockfile**（pnpm 9 产物，本仓库 pnpm v11）。

⚠️ 另有一处原文档没提：`frontend/pnpm-workspace.yaml` 里**也有一个 `overrides:` 块**，注释写明
「pnpm v11 reads overrides here; Docker/CI still use pnpm 9 which also honors
package.json.pnpm.overrides. Keep both in sync.」——本机 pnpm 11 读 workspace 那份，CI（pnpm 9）
读 `package.json` 那份。真做这一项时**两处都要加**，只加一处会「本地去重了、CI 没去重」。

本项已按 `design.md` 决策 1 推迟，以上内容留档，供触发重新评估时直接用。

### 4.2 ⚠️ `913ec5d74` 引用了本仓库没有的 `CodexCanonicalClientVersion()`

原文档标它「apply --check 通过」，属实——但它**不编译**。新增的
`buildOpenAIOAuthUpstreamModelsRequest` 里调用：

```go
modelsURL, err := buildCodexModelsManifestURL(chatgptCodexModelsURL, false, CodexCanonicalClientVersion())
```

`grep -rn "CodexCanonicalClientVersion" backend/` **零命中**。上游该函数体（`c40edb4` 的
`internal/service/openai_codex_identity.go:107`）就一行：

```go
func CodexCanonicalClientVersion() string { return resolveCodexOutboundIdentity("").version }
```

本仓库 `internal/service/openai_codex_identity.go:108` 有 `resolveCodexOutboundIdentity`，
所以**补同名 helper 即可**，语义完全等价。其余被引用的符号
（`buildCodexModelsManifestURL`、`chatgptCodexModelsURL`、`resolveCredentialAccount`、
`newUpstreamModelSync*Error`、`buildAgentIdentityAuthenticationHeaders`、`IsOpenAIAgentIdentity`）
均已存在。

### 4.3 ⚠️ `e7a3c1202` 与 `21c07e835` 的顺序要**反过来**

上游拓扑序是 `e7a3c120`（路由付费账号到 daily）**先**、`21c07e83`（把 daily 端点从
`daily-cloudcode-pa.sandbox.googleapis.com` 纠正为 `daily-cloudcode-pa.googleapis.com`）**后**。
只合前一半、或按上游顺序分两次上线，中间那段时间**付费账号会被打到 sandbox 域** ⇒ 401
「Invalid bearer token」，正是 #3611 / #2962 那个「测试连接成功但网关 401」的老坑。

⇒ 本仓库**必须同一提交落地，或先合 `21c07e835` 的 URL 修正**。

### 4.4 `86470628d` 依赖 `b30651a0a`（已实测）

单独 apply 时 `openai_gateway_ollama_cloud_cc_reasoning.go` 报 `NOFILE`（该文件由
`b30651a0a` 新建）；先应用 `b30651a0a` 后 `86470628d` 变为 clean。顺序固定，原文档已说，此处坐实。

### 4.5 `cfecc8d11` 与 `e4f869e0c` 顺序无关（已实测）

两者都改 `OpsErrorDetailModal.vue` 与 `i18n/locales/{en,zh}/admin/ops.ts`，但实测**两种顺序下
另一条都仍然 clean**。原文档把 `cfecc8d11` 写在前面（与上游拓扑序相反），不影响结果。

## 5. 与另一份 P0 清单的关系

`openspec/changes/port-upstream-0.1.183-p0-fixes/` 覆盖 v0.1.181–v0.1.183 的 12 项。两份
**无文件冲突**。需要留意的两点：

- 那份的「不做项」里，Responses Lite 簇（5 条）与 `e440ac48` 缺的四个基座符号来自 0.1.180 的
  **§6.1 / §7.1**，不在本 change（本 change 只做 §5 的 P0）。做完本 change 它们仍然做不了。
- `openai_gateway_scheduling.go` 的抢文件问题涉及 0.1.180 **§6.3** 的 `3fd66a33b`，也不在本 change。
