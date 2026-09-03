## Why

第一档（[`port-upstream-0.2.0-p0-fixes`](../port-upstream-0.2.0-p0-fixes/)）吃掉了 6 项「小、干净、
无冲突」的修复。本批是评估里的**第二档**：同样是「缺陷已核实或功能收益明确」，但每一项都要手工改
——补基座、拆上游 PR、写 SQLite 迁移、或动本仓库长期自有改动过的前端文件。

四项 P0 余项：

- **bootstrap 请求缺 `call_id` 就被拒。** delegation 与 scheduled-automation 两种引导请求本来
  没有 `call_id`，网关按普通工具调用校验 ⇒ 直接拒。两条上游提交成链（后者 +95 行落在前者新增的
  +232 行之上）。
- **API key 路径丢掉会话缓存身份**，导致本可命中的 prompt cache 落空。本仓库已有
  `internal/service/openai_compat_prompt_cache_key.go`（181 行）作基座。
- **上游在 turn 仍活跃时发来 1000/EOF，被判成 graceful ⇒ 适配层报 `relay_completed`。**
  干净的 WebSocket close 只描述传输层握手；一旦开了 Responses turn，成功还要求一个终止协议事件。
  本仓库缺 `openAIWSRelayActiveTurnID`（上游 11 行纯函数），按 README「基座要顺着符号回溯一层先量
  大小」量过 ⇒ 便宜，随本项一起落。
- **本仓库完全没有 Fable 的 fallback 定价。** `internal/service/billing_service.go` 里 grep `fable`
  **零命中**：Fable 计费只靠 LiteLLM 目录，目录缺条目时 `getFallbackPricing` 的系列匹配
  （opus / sonnet / …）全都不命中。上游 `34b8bf1a6` 同时补齐 `claude-fable-5` 与
  `claude-fable-5-1` 两档，并把 fable 判定放在 opus 判定之前。

四项 P1：

- **`pricing.override_file`** —— 稀疏 JSON 补丁按字段浅合并覆盖官方目录，作为最高优先级数据源。
  对个人部署价值最高的一条：想关掉某个模型的长上下文阶梯只要写
  `long_context_input_token_threshold: 0`，不用自建价格镜像、也不用改 `resources/` 下那份
  199 KB 出厂快照（改了就和上游永久分叉）。基座已在（`pricing.fallback_file` 与 `parsePricingData`）。
- **推理档位两簇** —— 本轮唯一一个「大功能但基座齐」的：`ReasoningEffortMapping` 类型、233 行策略
  实现、ent 的两个字段、前端组件与视图助手全在。PR#6447 加「超限拒绝 or 降级」，PR#6425 加
  「按模型限定映射」。
- **渠道 `cache_write_1h_price`** —— 服务层已有该字段（`internal/service/channel_plaza.go:16`，读
  LiteLLM 的 `cache_creation_input_token_cost_above_1hr`），但渠道自定义定价表没有这一列。
- **分组模型定价弹窗布局** —— 纯样式，与上一项改同一批前端文件，相邻实施省一轮冲突。

外加一项**并入的依赖卫生**（不属于本轮上游，按用户决定并批）：

- **dompurify `3.3.1` → `3.4.14`**（上游 `4a1da2950`，v0.1.180 期的项）。0.1.180 §5.1 当初判「推迟」，
  2026-09-02 重量后改判「建议做」：成本是 **3 个文件 / 22 行 lockfile / 零代码改动**，
  收益是一次清掉 **18 条 sanitizer-bypass advisory**（单包最多）并把两份副本
  （直接 `3.3.1` + mermaid 的 `3.3.3`）合成一份。理由不是「修一个已知可利用的 XSS」——
  原清单唯一引用的 `GHSA-cj63-jhhr-wcxv` 我用四种原型污染形态都没能复现；理由是可验证的那些：
  18 条 advisory、**7 个** `DOMPurify.sanitize` 调用点（原清单只记了 1 个），
  其中 `views/public/LegalDocumentView.vue:158` 走的 `/legal/:documentId` 是
  `requiresAuth: false` 的**公开页**。完整测量见
  [0.1.180 §12](../../../docs/upstream-sync/PORTING-0.1.180.md)。

本变更**只做移植**，不跟进任何功能簇。不做项及理由见 `design.md` 第 4 节与 PORTING §5。

## What Changes

- delegation 与 scheduled-automation 两类 bootstrap 请求在缺 `call_id` 时被接受，其余请求的
  `call_id` 校验不变。
- API key 路径在会话缓存身份上与 OAuth 路径对齐，使 prompt cache 可命中。
- WS v2 透传中继：上游 close 若发生在 turn 仍活跃时，判为**失败**而非 graceful，错误信息前缀标明
  「terminal event 之前关闭」；同时新增 11 行 `openAIWSRelayActiveTurnID` 基座。
- 新增 `claude-fable-5` / `claude-fable-5-1` 的 fallback 定价，`getFallbackPricing` 的 fable 判定
  置于 opus 判定之前且 5.1 先于 5；`claude-fable-5-1` 进入 Claude DefaultModels、Antigravity /
  Bedrock 默认映射与前端模型白名单及三处预设映射。
- 新增 `pricing.override_file` 配置（默认关闭）：补丁挂在 `parsePricingData` 入口，只修补已存在条目
  （同名字段覆盖，值为 `null` 删字段）；目录都没有的模型经独立路径并入；未生效条目打 WARN。
- 分组新增「推理档位超限处置」（拒绝 / 降级）与「按模型限定档位映射」两项策略，贯穿
  admin API / api_key 认证投影 / 网关四条发送路径 / 前端表单（**+迁移 `225`**）。
- 渠道自定义定价新增 `cache_write_1h_price`（四张表各一列，**+迁移 `226`**）；
  `cache_write_1h_price` 为 NULL 时单独的 `cache_write_price` 继续同时覆盖两个 TTL 档。
- 分组模型定价弹窗布局调整（纯样式）。
- `frontend/package.json` 的 `dompurify` 提到 `^3.4.14`，`package.json` 的 `pnpm.overrides` 与
  `pnpm-workspace.yaml` 的 `overrides` **两处**各加 `dompurify@<3.4.14: >=3.4.14`，
  `pnpm-lock.yaml` 由 `pnpm install --lockfile-only` 重新生成；顺手删掉已被官方标注为 stub 的
  `@types/dompurify`。**净化行为零变化。**
- **不动 `wire.go` / `wire_gen.go`；不改 `VERSION`；不引入任何 CN 平台分支；不改
  `.github/audit-exceptions.yml`。**

## Capabilities

### New Capabilities

- `openai-bootstrap-request-admission`: 引导类请求对 `call_id` 的准入契约。
- `openai-apikey-cache-identity`: API key 路径的会话缓存身份保真。
- `openai-ws-relay-terminal-event`: WS 中继「传输层干净关闭」与「协议层成功」的区分。
- `claude-fable-model-pricing`: Fable 系列的兜底定价与模型面登记。
- `pricing-catalog-override`: 价格目录覆盖补丁的合并语义与失败安全。
- `group-reasoning-effort-scope`: 分组推理档位策略的作用域（按模型）与超限处置。
- `channel-cache-write-ttl-pricing`: 渠道自定义定价对 5m / 1h 两个缓存写入档的拆分与向后兼容。
- `frontend-sanitizer-dependency`: 前端 HTML 净化器依赖必须是单一且已打补丁的副本，两处 overrides 同步。

第 8 项（弹窗布局）是纯呈现改动，**不定义 capability**；它的验收只有「前端测试通过 + 不引入横向滚动」。
第 9 项对应 `frontend-sanitizer-dependency`。

### Modified Capabilities

无。`openspec/` 下没有已发布的 capability 基线（只有 `changes/`），故本变更以 ADDED Requirements
形式固化「移植后应当成立的行为」。

## Impact

- **后端**：`internal/handler/` 的 `openai_gateway_handler.go`、`admin/group_handler.go`、
  `admin/channel_handler.go`、`available_channel_handler.go`、`dto/{types,mappers}.go`、
  `composite_platform.go`、`ops_error_logger.go`；`internal/service/` 的
  `openai_gateway_chat_completions.go`、`openai_compat_prompt_cache_key.go`、
  `openai_ws_v2/passthrough_relay.go`、`openai_ws_forwarder{,_ingress}.go`、
  `openai_ws_v2_passthrough_adapter.go`、`openai_reasoning_effort_policy.go`、
  `openai_gateway_messages{,_chat_fallback}.go`、`admin_group{,_duplicate}.go`、`admin_service.go`、
  `api_key_auth_cache{,_impl}.go`、`group.go`、`billing_service.go`、`channel{,_available,_service}.go`、
  `model_pricing_resolver.go`、`pricing_service.go`、`account_stats_pricing.go`；
  `internal/repository/` 的 `group_repo.go`、`api_key_repo.go`、`channel_repo_pricing.go`、
  `channel_repo_account_stats_pricing.go`；`internal/pkg/claude/constants.go`；
  `internal/pkg/antigravity/`；`internal/domain/{constants,reasoning_effort}.go`；
  `internal/config/config.go`；`ent/schema/group.go`（+ `go generate ./ent`）。
- **前端**：`components/admin/group/ReasoningEffortPolicyFields.vue`、
  `components/admin/channel/{IntervalRow,PricingEntryCard,types}`、
  `components/modelPlaza/PlazaModelPricingTable.vue`、`components/keys/UseKeyModal.vue`、
  `composables/useModelWhitelist.ts`、`views/admin/{GroupsView.vue,ChannelsView.vue,groupsReasoningEffort.ts}`、
  `api/{admin/,}channels.ts`、`types/index.ts`、i18n（zh/en）。
- **数据库**：**新增两条迁移，本仓库号 `225`（分组推理档位超限处置）与 `226`（渠道 1h 缓存写入价，
  四张表各一列）**。迁移号从 `224` 顺延到 `226`。两条都要 PG → SQLite 重写。
- **前端依赖**：`dompurify` `3.3.1` → `3.4.14`（两份副本合一），删 `@types/dompurify`；
  lockfile churn 22 行，全部集中在 dompurify 自己，无其它包版本变化。**这三个文件必须同一提交**
  （5 处 `--frozen-lockfile`）。
- **配置**：新增 `pricing.override_file`（默认空 = 关闭）。无默认值变化。
- **计费**：三处方向性影响 ——（a）Fable 请求在目录缺条目时**从「无兜底」变成有兜底价**；
  （b）渠道配了 `cache_write_1h_price` 后 1h 缓存写入按新列计价；
  （c）`override_file` 一旦启用即为最高优先级数据源，配错会直接改变计费。见 `verification.md` §3。
- **准入 / 路由**：推理档位超限从「静默钳制」变为「按分组配置拒绝或降级」。**默认值必须保持与现状
  等价的那一档**，否则升级即改变既有分组的行为，见 `design.md` 决策 5。
- **兼容性**：`cache_write_1h_price` 为 NULL 时保持拆分前语义（单独的 `cache_write_price` 覆盖两档），
  这是显式的向后兼容条款，MUST NOT 简化成无条件赋值。
- **安全门禁**：**不变**。CI 是 `pnpm audit --prod --audit-level=high`，dompurify 那 18 条全是
  low/moderate ⇒ 从来没进门禁。门禁当前报的 prod+high 是 `xlsx` ×2 + `nanoid` ×1，
  由现有例外条目覆盖（`expires_on: 2026-10-06`），**本批不碰 `.github/audit-exceptions.yml`**。
- **风险面**：本批最高风险是两条迁移 + 推理档位贯穿四条发送路径。迁移一旦落库无法改文件
  （checksum 不可变，README 第 4 节硬约束）。

## Execution References

- `source-baseline.md`：上游 commit / PR merge 固定、按文件 apply 四态证据、依赖符号核查、构建标签核查。
- `design.md`：移植策略、逐项决策、迁移转写规则、阶段划分与回滚。
- `docs/upstream-sync/PORTING-0.2.0.md` §3.8–§3.11 / §4：逐条 patch site 与上游 diff 摘要；§7 落地顺序；§8 自测。
- `docs/upstream-sync/README.md` §4 硬约束 + §5 PG → SQLite 转换速查。
