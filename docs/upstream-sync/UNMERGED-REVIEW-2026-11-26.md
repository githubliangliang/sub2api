# 历史未合内容第二轮复评（2026-11-26）

**上一批已经完成：`89060f532a9000af28a47036f70cc44821dafd55` 包含 Token 统计和 Gemini 带内信号两簇、11项任务。下一批建议第一档3簇、第二档3簇，来自8个旧PR。** 本轮继续检查上轮未逐项行为重放的旧修复，确认 Codex 非GPT身份、S3密钥二次保存、属性表单及异步状态的现有缺口，不重新开发上一批。

[第一档提案](../../openspec/changes/port-upstream-historical-followup-tier1/README.md)13项、[第二档提案](../../openspec/changes/port-upstream-historical-followup-tier2/README.md)14项，**全部未实施、任务未勾选、实施验收槽为空**。[24个整PR逐项判断](./unmerged-review-2026-11-26/decisions.md)和[原始证据](./unmerged-review-2026-11-26/README.md)独立保存；上轮134PR库存和各历史报告保持快照原文。

## 1. 真实基线与本轮范围

| 项目 | 值 |
|---|---|
| fork HEAD / VERSION | `89060f532a9000af28a47036f70cc44821dafd55` / `1.1.28` |
| 当前分支 | main，比本地origin/main多1个提交；本轮没有推送或发布 |
| Go / SQLite / auth snapshot | `1.27.4 / 230 / 21` |
| 旧来源终点 / main观察 | 上轮已经评估至v0.2.25；来源main `0f6671457132425d1b84f3d80536b53e02942424`，本轮未查询新Release |
| 本轮取样 | 上轮134PR库存中的24个第三档PR，全部整PR来源哈希与上轮一致 |
| 候选×文件 A/C/X/N | 61条，0 / 46 / 13 / 2；全体先反向后正向，仅私有index check |
| 调整结果 | 第一档3簇/4PR、第二档3簇/4PR；第三档10PR、第四档6PR |
| 既有活动OpenSpec | 17份、467任务完成、0未完成；上一批两份source-baseline与验收保持原文 |
| 启动工作区 | 9,286份tracked；只有原Fast诊断未跟踪，哈希保护 |

[18份最终实施文件哈希](./unmerged-review-2026-11-26/completion-hash-check.json)全部匹配，确认上一批实现仍在；复用其[独立验收](./implementation-historical-backlog/README.md)，不把本轮定向回归说成重新完成全仓unit/race。2026-11-26是本地日期，前批UTC日志为2026-11-25，两者并不冲突。

## 2. 复核方法

本轮先核对生产消费者，再针对可达行为运行探针。特别复查同名文件/组件、实际router、普通API Key兼容和缺失原生平台的区别。

- 冻结24份 `merge^1..merge` 净补丁、SHA256和61条逐文件reverse/forward结果；549条新增调用索引与当前声明分别记录，CLEAN不等于能编译或应该采用。
- 正常基线先列测试入口，再运行service/repository/admin三包75顶层/121含子例；现有四份前端测试15项通过。没有引用postgres约束文件作为SQLite证据。
- 后端探针仅用Go overlay；前端探针在现有Vitest中短暂加载测试文件，finally删除。真实Vue组件、真实SQLite设置仓库与AES-GCM均不改产品代码，也不读取业务数据。
- 报告中“已复现”只覆盖对应探针；未调用实际Codex/Gemini/DeepSeek等供应商，不把上游注释里的429或封禁原因当本轮验证。
- 第一/第二档均无需DDL、Ent/Wire、依赖、VERSION或资源默认变更。其它按需能力不自动成为前置。

## 3. 第一档：三个窄修复

### F01 非GPT模型的Codex清单身份（#7308）

**旧第三档明确调整为第一档。** 当前 `BuildCodexModelsManifestForGroup` 为Grok、DeepSeek、Claude、Gemini四种非GPT模型输出的 `model_messages.instructions_template` 均包含GPT自我身份。普通GPT模板逐字对照通过。已有真实清单生成与映射后模型选择，不需要旧pinned/live目录系统作为前置；旧“整套目录未选”的标签不能挡住此独立修复。

- patch sites：`openai_codex_models_service.go` 的 `newConfiguredCodexModelDescriptor` 及模板选择；原指令资源不改写。来源2文件+99/-1，核心约40行规则与两个helper。
- 最小适配：仅移除网关为非GPT条目合成的GPT身份片段，保留Codex工作指令、占位符、格式与剩余正文。按实际映射目标选模板，不能仅凭公开别名判断；用户/上游显式指令不进入全局替换。
- 验收：现有manifest出口、GPT/非GPT/未知别名、gpt-oss、供应商前缀与模型映射；GPT模板原字节保持，提示其它GPT引用不被误删。不宣称修复已经复现的供应商拒绝或账号处罚。

### F02 S3配置二次保存会写回明文密钥（#4867）

**旧第三档明确调整为第一档。** 本轮实际调用现有 `UpdateS3Config`，使用真实Ent设置仓库、modernc SQLite和AES-256-GCM：首次保存加密，下一次只改bucket并留空secret后，设置表中的secret变回原始明文。显式提供新secret的两次替换对照通过。

- patch sites：`backup_service.go` 的 `UpdateS3Config`。`loadS3Config`返回解密结果；当前加密只位于“新填secret”分支，所以“保留旧secret”绕过加密。来源生产改动只需重排该窄分支，2文件+56/-3含回归。
- 适用性：现有 `/api/v1/admin/backups/s3-config` GET/PUT仍有管理员与step-up保护；虽然旧BackupView未注册前端路由，配置API及复用备份S3凭据的图片对象存储链已经存在。无需恢复备份页面、增加S3能力或启动备份任务。
- 验收：新密钥、留空沿用、已有明文兼容、固定加密key缺失、加密失败/写失败、脱敏响应与SQLite实存值。沿用密钥也必须经过既有持久化加密约束；不扩大到密钥轮换、备份/恢复或业务数据操作。

### F03 属性表单的数值类型和显式清空（#7363/#7366）

**旧第三档两项明确调整为第一档。** 数字输入的 `v-model.number` 产生JSON number；既有 `UpdateUserAttributesRequest.Values` 为字符串映射，真实handler返回400，而合法整数的字符串返回200。另一方面，说明/placeholder清空后被 `|| undefined` 从实际JSON省略，后端指针更新只会保留旧值。两项均由真实Vue组件探针复现，不是新增自定义属性功能。

- patch sites：`UserAttributeForm.vue` 的number输入与 `UserAttributesConfigModal.vue` 的提交对象；生产改动各两行左右，保留现有API。
- 合同：以字符串提交数字属性，显式空串表达清空；不改后端整数校验，不新增小数支持。初次HTTP对照误用12.5，被现有 `strconv.Atoi` 拒绝；更正为合法整数12后number400/string200通过，原失败保留。
- 验收：数字/文本/空值、0和负值及已有上下界；清空/省略/非空的区别、实际请求JSON与服务端更新。不得顺带更改属性schema或恢复新多人治理策略。

## 4. 第二档：三个生命周期适配

### S01 用户属性请求和分组保存的归属（#7500/#7373）

**旧第三档明确调整为第二档。** 先打开用户A再切到B，A的迟到属性请求覆盖B当前输入；分组加载失败后，保存按钮与方法仍可发送 `allowed_groups: []`。后者探针只拦截请求，没有修改真实用户授权；风险是把未加载状态当作用户清空意图。

- patch sites：`UserAttributeForm.vue` 的userId watcher/读取/emit；`UserAllowedGroupsModal.vue` 的show/user变化、load、save。
- 已有实际消费者：`UserEditModal`通过v-model接收属性，提交时使用当前用户ID；`UsersView`挂载属性配置和分组对话框，simple路由没有屏蔽 `/admin/users`。
- 手工适配：区分未加载、失败、当前会话加载完成及提交；用户/对话框切换后旧成功、失败和finally不影响新状态，保存只使用本次用户与已加载配置的快照。保留本地公开分组默认语义，不导入来源尚未选定的 `restrictPublicGroups` 策略。
- 验收：A→B乱序、关闭重开、加载失败/重试、保存中切人、旧错误/旧loading、取消；真实handler/SQLite验证未加载时不覆盖原授权和属性。现有组件卸载后不再emit的对照已经通过，不将它另报缺陷。

### S02 共享图片上传的FileReader所有权（#7320）

**从旧图片扩展第三档中拆出，明确调整为第二档。** 当前共享 `ImageUpload` 被SettingsView三处使用，不依赖新图片模型或原生Codex Images。可控FileReader探针重现：第二张完成后第一张晚完成覆盖新图；点击删除后晚读又恢复图片。单次读取对照通过。

- patch sites：`components/common/ImageUpload.vue` 的reader、选择、删除、卸载和错误回调。
- 适配与验收：先取消旧读，并约束回调只属于当前选择；新选择、删除、卸载、晚成功/晚错误都不得覆盖新状态。保留SVG净化、类型/大小拒绝及原有文件输入行为，不提高300KiB或其它上传边界。

### S03 管理员监控模板的关联列表（#7365）

**旧第三档子项明确调整为第二档。** 实际 `ChannelMonitorView→MonitorTemplateManagerDialog→MonitorTemplateApplyPickerDialog` 已接通。切换模板后，旧请求晚到覆盖当前列表与默认选中ID，可能令当前模板提交旧列表中的选择；探针复现显示覆盖，不向真实监控提交修改。

- patch sites：`MonitorTemplateApplyPickerDialog.vue` 的watch/fetch/选中状态和提交会话边界。
- 适配与验收：切换templateId、关闭重开、卸载后，旧成功/错误/finally不得影响当前状态；apply绑定本次模板和选中ID，旧提交结果不关闭新会话。来源主要修读取防护，不能据此跳过本地提交生命周期复核。
- 与未注册用户页面区分：#7419/#7492改的是旧用户ChannelStatusV1和MonitorDetailDialog，当前router未注册该view。本轮不恢复该入口，也不把管理员监控的可达性当作用户页可达。

## 5. 其余旧项与明确不取的内容

本轮24项的其余16项均在[逐候选表](./unmerged-review-2026-11-26/decisions.md)记录。以下结论不表示重跑所有历史PR：

| 范围 | 当前判断 |
|---|---|
| #7491 旧模型测试picker | 来源 `FetchOpenAIAccountModels/projectAccountModelsBody`消费者缺失。本地 `GetAvailableModels`直接返回公开mapping键，并有OAuth/APIKey/透传HTTP回归。来源故障不适用，不为它引入完整discovery。与#7308现有manifest身份消费者不同。 |
| #7313/#7224/#7394 | DeepSeek占位reasoning、严格供应商developer角色、Baseten reserve错误仍按实际供应商用途；现有兼容出口可用不代表原生枚举存在，也不等于协议应对所有账号统一改写。未作供应商在线核验，保持第三档。 |
| #7246/#7354 | 来源native `normalizeDeepSeekResponsesRequestBody`不存在，本地不走该修复链；第四档不适用。将来新增独立供应商适配需新合同，不能仅塞入helper宣称已修。 |
| #7057/#7085 | 独立最大上下文仍第三档；仅删metadata子字段与删整个父字段保留上轮已量的不同合同，不视为已等价或自动启用。 |
| #7215/#7636 | 大目录解析优化和setup未消费的rate_limit默认值清理仍第三档；本轮没有新性能收益证据，不把无运行消费者的默认段称为限流故障。 |
| #7442/#7060 | 图片余额结构化code/模型冷却、Gemini通过OpenAI图片端点是独立协议/路由策略，保持第三档；与共享FileReader竞态无必要依赖。 |
| #7375/#7494 | 充值/取款/total_recharged历史混合界面的#7375按个人边界明确排除；普通管理员余额调整错误提取#7494只可考虑独立非支付子项，仍第三档。本轮不修改旧商业界面或业务余额。 |

其余历史未合内容继续沿上轮逐类清单保留：ToolSearch、搜索来源块、可选compaction5m、名称备注同步/身份头；重置grant、30m/12h与effort/schema、模型/区域/图片能力；共享缓存/COW/lane/SQL、动态预算/抢占；MCP/admin Key、注册/多用户、新平台/视频/TypeSafe、自动版本及备份运营。未取得新用途或收益证据的部分不升档。特别保留：

- Fast、窗口TPS、Grok诊断、历史工具尾项、requested/effective effort、SQLite后台、固定quota/dedup均已完成；本次旧身份缺口不是重做整个manifest或工具系统。
- #8251/#8833/#8269/#8816/#8518沿用2026-11-25限定的合成测量；本轮未重新测性能或宣称生产收益。#8610的1GiB最低预算不替代64MiB。
- #7684/#7726属于上游权益管理，仍按实际grant选择，不误判支付；购买、充值、退款、付费续订、商业返利和专属依赖迁移始终排除。
- PG专属、跨节点租约、缺原生平台/插件消费者不引入。当前未重跑audit，2026-11-25的12 moderate/2 low只是历史生产审计，不能说成2026-11-26最新或全仓安全结论。

## 6. 已具备和新最小前置

| 项目 | 本轮重新定界 |
|---|---|
| Codex身份 | 当前映射目标元数据/清单生成已在；约40行来源纯规则即可接入，不需要新pinned/live框架。 |
| S3加密 | AES-GCM、固定key检查、真实SQLite设置和受保护API已在；需要修沿用分支，无新备份功能。 |
| 用户表单 | 已有字符串/指针请求合同和管理入口；不需新属性表或public-group restriction。 |
| 异步状态 | 现有父模板管理列表已有请求保护，不等于子“应用模板”选择器也已保护。必须分别沿消费者检查。 |
| 用户监控 | 仅有旧文件不足以证实路由存在；对应用户view不在router，不能凭管理员监控存在而升档。 |

## 7. 建议落地顺序

第一档先做Codex生成身份和S3持久化，再补属性字段提交；F03与S01改同一属性组件，实施时按组合终态适配。第二档先守住用户/授权写入归属，再分别处理共享图片读取和模板选择。每一簇保留失败回归与正常控制后再填写任务。

本轮只生成下一批可审查提案；不提交/推送/发布、归档或操作业务实例。现有main多出的实现提交保持本地状态。

## 8. 本轮实际验证与局限

| 检查 | 结果与边界 |
|---|---|
| 旧实现保护 | 18份最终实现文件哈希匹配；17份活动变更467项完成。 |
| 整PR来源 | 24整PR、61文件A0/C46/X13/N2、549新增调用索引；完整补丁SHA与上轮一致，仅check。 |
| 当前backend | 三包75顶层/121含子例通过；先保存实际测试目录。 |
| 新backend探针 | 2个目标顶层失败/6含父子例失败，2个正常控制顶层通过；实际manifest生成与SQLite/Ent/AES密钥保存，不连接外部供应商/S3。 |
| 现有frontend | 四文件15项通过，覆盖UserEdit、UsersView、模板管理列表及管理员监控旧控制。 |
| 新frontend探针 | 11项中7个期望缺口失败、4个控制或观察通过；真实Vue编译/渲染与可控异步夹具，不是浏览器验收或真实数据写入。 |
| HTTP类型控制 | 实际属性handler：number12返回400且不入库，字符串12返回200。service/repository使用窄stub；此前decimal控制错误原文保留，不改变整数合同。 |
| 临时夹具失败 | 初次TS缺大括号未运行测试；后续email/DOM选择器夹具错误及清理连带错误均独立保存，修正后7个失败均到达目标断言。未改产品消除这些错误。 |
| 最终范围/规格 | 两份未实施OpenSpec严格校验、链接、源hash、空验收槽、旧任务/HEAD/index/ref及初始Fast诊断保护见[validation.json](./unmerged-review-2026-11-26/validation.json)。 |

没有新全量unit/integration/race/build、浏览器、安全全扫、供应商或1C1G容量测试。它们属于未来选定实现的相应门禁；本轮正常基线通过不能掩盖新增失败探针，探针也不能当修复已经完成。

## 9. 本轮通用教训

- 已选新一批完成后，先锁实际提交与哈希；冻结评估的旧“未实施”不重写，也不把它当当前待办。
- 目录大框架未选不代表现有清单生成没有独立缺口；相反，源码方法缺失或路由未注册也不能凭同名文件认作活缺陷。
- “沿用旧密钥”跨解密/保存边界后仍要加密；能读回不证明数据库存的是密文。
- 前端number修饰符、undefined省略与后端string/指针合同需要沿实际JSON验证；数字字符串合同不等于支持小数。
- 旧请求保护要覆盖当前对象和会话，不只按钮loading；父列表通过也不能证明子对话框或FileReader具有同样所有权。
