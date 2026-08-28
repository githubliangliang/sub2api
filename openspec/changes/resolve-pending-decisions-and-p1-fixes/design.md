## Context

### 位置

两批 P0 已落地（0.1.180 的 19 项提交为 `1b4837098`，0.1.183 的 12 项在工作区，`go build ./...` 通过）。
本 change 是下一批：四项 P1 加上三个从 0.1.176 / 0.1.179 一路拖到现在的决策。

### 为什么把四项 P1 和三个决策放在一起

四项 P1 各自都小（最大的 `3fd66a33b` 是 2 文件 +118-16），单独开 change 的开销大于收益。
三个决策里有两个（9.1 Grok 4.6、9.3 长上下文 OR）**动的就是四项 P1 里第 2 项同一批文件**
（`pkg/xai/models.go`、`billing_service.go`），分开做要解两次冲突。9.2（Go 1.27）与其余无关，
但它是唯一的「工具链」类改动，放在最后一个阶段单独提交即可。

### 这一批的性质与前两批不同

前两批全是「修上游确认过的缺陷」。本批有三项是**主动选择**：跟默认模型、跟工具链、改计费口径。
其中 9.3 会让费用**上升**——这是前两批都没有的方向。因此本 change 的重点不只是「改对」，
还包括「改之前把后果说清楚，改之后能解释账单变化」。

## Goals / Non-Goals

### Goals

- 四项 P1 的缺陷消失；三个决策在代码、配置、文档、存量数据四个面上一致落地。
- 依赖审计门禁恢复成可信信号（这是上一批推迟 dompurify 的前提）。
- 9.3 的计费口径变化有留档证据，事后能解释账单为什么变。
- Go 1.27 的生成代码走 `go generate`，不手改。

### Non-Goals

- 不做 0.1.180 §6.1（Responses/Chat 工具桥接 11 条）、§6.2(c)（Grok 429 / Realtime / 容量串）、
  §6.3 除 `3fd66a33b` 外的三项、§7.x 全部。
- 不做 0.1.183 §4.1（监控 v2 composite 平台归属，需 SQLite 重写）与 §5.x。
- **不解除 dompurify / nanoid 的推迟**。本批只续期已有审计例外，不碰 `package.json` /
  `pnpm-lock.yaml` / `pnpm-workspace.yaml`。
- 不动 `VERSION`，不新增迁移。

## Decisions

### 1. 审计例外只做「续期或标记已解决」，不做依赖升级

三条过期例外（`lodash` / `lodash-es` / `axios`）的处理有两条路：升级消除，或重新核实后续期。
**本批选续期**，理由是升级要改 `package.json` + 重新生成 lockfile，而那正是上一批显式推迟的工作
（`port-upstream-0.1.180-p0-fixes/design.md` 决策 1）；把它塞进来等于绕过那个决定。

续期不是走形式：每条都要重新核实「本仓库用法为什么不受影响」，并写进 `reason`。
现状已查明供填写：`axios` 是直接依赖（`frontend/package.json:23` `^1.18.0` → lock `1.18.1`）；
`lodash` / `lodash-es` 在 `frontend/src/` 里 grep 不到直接 import，是传递依赖。

新的 `expires_on` 建议压在下一次计划性依赖工作之前，别再给一年。

### 2. Grok 三条的落地顺序是硬约束

`ed4207a16` → `39485f2e2` → `f7145c750`，理由在 `source-baseline.md` §2.5：
`39485f2e2` 会改 `ed4207a16` 刚引入的别名收窄条件，并重排 `ed4207a16` 刚加过 case 的价卡 switch。

而且 `ed4207a16` **只能部分移植**：它的 `openai_gateway_grok.go` hunk 调用
`grokSameAccountRetryMetadata`，本仓库零命中（属 §6.2(c) 那簇）⇒ 该文件的 hunk 整个丢弃，
只取 `pkg/xai/models.go` 与 `billing_service.go`。这是又一个「`apply --check` 过了但不编译」的实例。

`f7145c750` 是 `39485f2e2` 的数据迁移，**不能只合一半**：只改代码默认不迁数据，存量库里显式存着
`grok-4.5` 的设置会继续生效，运营方看到的默认模型与代码常量不一致；只迁数据不改代码更荒谬。

### 3. 默认模型迁移只动「恰好等于旧内置默认值」的那一个值

`MigrateGrokDefaultTextModel` 的判据必须是 `strings.TrimSpace(value) == "grok-4.5"`：

- 设置从未写过（`ErrSettingNotFound`）⇒ no-op，运行时自然取新常量。
- 值是其他任何东西 ⇒ 运营方的显式选择或未来的默认值，**不动**。

这条语义是整个 9.1 里最容易改错的地方——写成「不是 4.6 就改成 4.6」会静默覆盖运营方配置。

### 4. 调度诊断只加 reason，准入行为一个字节不改

`3fd66a33b` 的做法是把 boolean 函数改成返回 `string`（空串 = 通过），原 boolean 函数保留成
`... == ""` 的薄包装。**这个形状要照搬**：它保证准入判定的短路顺序与结果完全不变，
reason 只是把「第一个否决点」的名字带出来。

验收因此不是「代码里有 reason」，而是「同样的账号集合、同样的请求，选中结果与移植前逐个一致」。
以 `c35a482` 那次静默调度事故的经历，这条的价值全在下一次排查时能少走多少弯路。

### 5. ⚠️ 9.3 是本批唯一让费用上升的改动，必须留证据

改成 OR 后：分组开关默认**开**、账号开关默认**关** ⇒ 超过长上下文阈值的请求，
此前因账号开关关着而**不收**倍率，之后会**收** 2× 输入 / 1.5× 输出。上游 v0.1.179 的 release notes
把这条列为 breaking change，原因就是这个。

三件事必须一起做：

1. **按上游形状改，不要把 `&&` 换成 `||`**。权威写法见 `source-baseline.md` §2.7：
   「分组 OR 账号」与「无区间定价」是两件事，后者仍是 AND。直接换 `||` 会让账号开关绕过
   区间定价的前置条件。
2. **第二处不动**。`billing_service.go:1045-1046`（无 Resolver 回退路径）拿不到分组开关，
   账号开关是唯一信号，保持原样。
3. **留一条改前改后的对比样本**（`verification.md` §4），事后能回答「账单为什么涨了」。

想保持旧行为的分组：把该分组的 `long_context_pricing_enabled` 关掉——这是上游给的 opt-out，
不需要改代码。

### 6. Go 1.27 放最后一个阶段，且 ent 必须 `go generate`

9.2 的 diff 会**淹掉**其余六项：ent 重新生成会动 7 个生成文件，golangci-lint 从 v2.9 到 v2.13
会带出一批 lint 修正。放在最后一个阶段、单独提交（或两个：工具链 + lint 修正），
这样前六项出问题时还能廉价 bisect。

`backend/ent/` 在 Go 1.27 默认 jsonv2 引擎下，`json.RawMessage` 字段会生成成 `jsontext.Value`
（`group.model_pricing`、`usage_cleanup_task.filters`）。**必须 `go generate ./ent` 并提交生成结果，
不要手改**（README 硬约束第 7 条）。

另外两处容易漏：

- 版本断言实际有 **5 行**（`backend-ci.yml` 两处、`security-scan.yml` 一处、`release.yml` 两处），
  而 `DEV_GUIDE.md:52` 现在写的是「三个 workflow ... 两处」，改的时候把这句一并纠正。
- golang 镜像有 **3 个** Dockerfile（根 `Dockerfile`、`backend/Dockerfile`、`deploy/Dockerfile`）。

`x/net` 已是 `v0.56.0`，在 go1.27 下它包装标准库 HTTP/2：`ConfigureTransports` 经
`RegisterProtocol("http/2")` 打开 `Protocols.HTTP2` 而不再写 `TLSNextProto`，
`ReadIdleTimeout` / `PingTimeout` 映射到 `HTTP2Config.SendPingTimeout` / `PingTimeout`
⇒ `http_upstream_http2_keepalive_test.go` 的断言要跟着改。

### 7. 六个 capability 按行为域切，Grok 两件事合一个

`grok-model-catalog-billing` 一个 spec 同时收「目录/价卡/别名修复」（P1 第 2 项）和
「默认模型 4.6」（决策 9.1），因为它们改同一批文件、有硬顺序依赖，拆开会让顺序约束落在两个 spec
之间无人负责。但在 spec 内部它们是**独立的 Requirement**：一个是修缺陷、一个是策略选择。

## Risks / Trade-offs

| 风险 | 处理 |
|---|---|
| **9.3 让费用上升，用户以为是 bug** | 决策 5：留改前改后对比样本；`proposal.md` Impact 写明 opt-out 是关分组开关 |
| 9.3 被写成简单的 `\|\|` 替换，账号开关绕过区间定价前置条件 | `source-baseline.md` §2.7 给出上游权威写法；spec 用「区间定价存在时不叠倍率」做断言 |
| 默认模型迁移误覆盖运营方显式配置 | 决策 3：判据必须是 `== "grok-4.5"`；spec 有对应 Scenario |
| Grok 三条顺序反了 ⇒ 手工解冲突甚至语义错 | 决策 2；`tasks.md` 按顺序编号并在门禁里检查别名收窄的最终形态 |
| `ed4207a16` 顺手把 `openai_gateway_grok.go` 一起合 ⇒ 不编译 | 决策 2 点名；`tasks.md` 门禁 grep `grokSameAccountRetryMetadata` 必须零命中 |
| Go 1.27 的 diff 淹掉其余六项，出问题无法 bisect | 决策 6：放最后一个阶段、单独提交 |
| ent 生成物被手改 | 决策 6；门禁检查 `go generate ./ent` 后工作区干净 |
| golangci-lint v2.13 带出一批新告警，修不完 | 上游 `73aabc861` 已给出处理方式（G703/G704、`reflect.Ptr`→`reflect.Pointer`、SA4023/SA1019 nolint）；照搬，别自己发明 |
| 审计例外续期变成走形式，下次又过期 | 决策 1：新 `expires_on` 压在下一次计划性依赖工作之前；`reason` 必须重新核实 |
| 七项塞进一个 PR 无法评审 | 按 Migration Plan 的五个阶段拆 |

## Migration Plan

### 阶段 1：恢复审计门禁（最先，几分钟）

三条过期例外重新核实 + 续期。独立提交，与后续任何一项都无耦合。
做完就能知道「除了这三条，依赖审计还有没有别的红」——这是后面所有工作的背景信号。

### 阶段 2：Grok 目录与计费（顺序固定）

`ed4207a16`（**裁掉 `openai_gateway_grok.go`**）→ `39485f2e2` → `f7145c750`。
可以三个提交，但必须按这个顺序；`39485f2e2` 与 `f7145c750` 不得分开上线。

### 阶段 3：可观测性两项

`3fd66a33b` 调度诊断 → `4795650d` 真实上游端点（裁掉 `IsCNProvider`）。
两条同向、互不依赖，可合一个提交。

### 阶段 4：长上下文门控 OR

按 `source-baseline.md` §2.7 的上游形状改。单独提交——它是唯一改计费口径的一项，
要能单独 revert，也要能单独指给用户看「就是这次改的」。

### 阶段 5：Go 1.27 + golangci-lint v2.13（最后，单独提交）

`cbe258fd1` → `73aabc861`。ent 走 `go generate ./ent`。
建议拆两个提交：工具链版本 + 生成物，然后 lint 规则修正。

### 回滚

- 阶段 1：revert 即恢复红门禁（本来就是红的），无副作用。
- 阶段 2：三个提交要一起 revert（`f7145c750` 单独留下会让存量设置停在 4.6 而代码默认回到 4.5）。
  ⚠️ 数据迁移**不可逆**：revert 代码不会把 DB 里已被改写的 `grok_default_text_model` 改回
  `grok-4.5`。真要回退需手动把该设置写回，或在管理端重新选择。
- 阶段 3：纯可观测性，随时 revert。
- 阶段 4：revert 即回到 AND 口径；已按 OR 记账的历史记录不会回滚，也不需要回滚。
- 阶段 5：revert 要连 ent 生成物一起，且 golangci-lint 版本要同时退回 v2.9（v2.13 的配置可能
  不被 v2.9 接受）。
