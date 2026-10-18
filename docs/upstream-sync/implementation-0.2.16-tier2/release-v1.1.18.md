v1.1.18

相对 v1.1.17（882fe9777）。本仓库使用自有 1.1.x 版本号。
本次新增：上游 v0.2.13–v0.2.16 第二档 16 簇 / 30 个本轮来源 PR 的选定行为，以及逐项量测的必要前置。
功能提交：847db09e1；依赖维护提交：5ff6e5bf9。两者已从 sync/upstream-0.2.16-tier2 快进合入 main。
第一档 14 簇此前已在 7c544c03f 完成并包含于 v1.1.17，不重复计作本次新增。
保持 SQLite-only、可选内置 miniredis、simple mode 与个人单节点定位；无新迁移，最大迁移仍为 226，Go 仍为 1.27.0。

同步上游版本与内容
================

1) 上游 v0.2.13–v0.2.16 第二档现有功能修复（S01–S15，847db09e1）

   协议、流与用量统计：
   - S02：旧版 function_call 按名称 FIFO 配对稳定 ID，避让现代 tool_call ID；Responses 拒答在普通响应、流式 delta 与缓冲汇总中保留，避免重复补入。
   - S03：Claude OAuth mimic 保留调用方显式声明的 structured-outputs、mid-conversation-tool-changes 和 inline-tools beta；逐项 drop 优先，不扩散未知 token，不隐式追加配套能力。
   - S04：WebSocket 后续轮次刷新同 Key、同组、同平台和订阅类型下的价格快照；保留连接身份并隔离交错轮次的计费时刻。图片用量兼容 tool_usage 回填，累计终态并按 response ID 去重。
   - S05：thinking_signature_invalid 的特定诊断复用有界恢复；Auto 模式下已支持 Responses 的 API Key 遇到转换后结构化 input 被明确拒绝时，可回退原始 Chat 请求。取消、已写响应、强制模式、Responses 入站、截断或含输出/用量的错误均不触发该回退，保留原请求计时。
   - S06：OAuth 回放 web_search 历史时补 cached-only 声明，Lite 使用 additional_tools；原本无工具时限制 tool_choice，保留显式选择。覆盖 HTTP 与既有 WS 入口，保持大整数和转义类型，API Key/compact 路径不改写。
   - S07：将启用的精确复合路由公开别名纳入对应模型目录及 Codex Responses 目录；按端点隔离、去重，不展开前缀路由，目录查询失败继续使用账号目录。
   - S11：按真实 OAuth 类凭据与最终 GPT-6.1 Sol 上游模型豁免默认价卡长上下文附加费；用户与账号成本一致，API Key、其它模型与显式本地绝对区间定价保持原规则。
   - S14：Responses input 无 namespace 修改时延迟正文重建并保留原始字节；畸形 JSON 返回原文。8 MiB tools_only 基准由 25,190,400 B / 3 次分配降至 8,396,800 B / 1 次分配。
   - S15：Anthropic ping 转为立即 flush 的 SSE 注释，不计首 token；客户端写失败停止转发，上游 error/读失败明确传播，避免合成成功终态或重放已提交响应。

   SQLite、安装与既有 Gemini 路径：
   - S01：统一 CLI、Web、AUTO_SETUP 和本地 YAML 的管理员创建校验，匹配登录邮箱规则与 bcrypt 8–72 字节密码边界；随机邮箱仅用于确需创建管理员的空库，已有用户/管理员保持原凭据。同步个人部署模板与安装向导。
   - S08：普通/级联删除分组在同一 SQLite 事务中清理风控作用范围；配置保存先取得写事务再校验分组，防止并发保存重新写回已删除组。畸形或缺失配置保留，数据库错误回滚，支持调用方事务。
   - S12：第三方 API Key 与 Vertex 不再套用本地模拟日配额；第三方 429 仅按显式延迟或全局冷却处理。缓存缺失作为请求错误返回，避免错误封号/换号；保留普通认证错误策略。补齐 Google 错误归类与请求归属，脱敏 AQ.Ab 密钥。
   - S12：既有 Gemini Flash 3.6 默认促销价使用每百万 token 输入 0.75 / 输出 3.75 / 缓存读取 0.075，截止 2027-01-01 00:00:00 UTC 后运行时恢复标准默认价，无需重启；现有点号、连字符和提供方前缀别名同样到期，本地显式覆盖价优先。

   前端状态与现有管理入口：
   - S09：分组 RPM/倍率弹窗的加载、保存、清空、搜索按弹窗代次隔离，切组/关闭/卸载清理；RPM 要求非负整数，倍率要求有限正数，批量编辑防溢出。
   - S10：资料草稿不被轮询或保存响应覆盖；身份绑定、Markdown 自定义页、运维分页/详情、监控模板、计划测试、清理弹窗、合规状态与公告排除旧请求回写。二次验证共享待定 Promise，作用域销毁取消；公告成功标记已读后才关闭。
   - S13：监控配置可补充本地已支持的五个平台之一，新项使用空模型列表并保留原配置；禁止重复/无效平台，保存期间禁用编辑。

2) 前端依赖维护（S16，5ff6e5bf9）及历史前置

   - DOMPurify 升至 3.4.17、axios 升至 1.20.0、Vue 升至 3.5.44、source-map-js 升至 1.2.2；同步 package.json / pnpm workspace overrides 与生成的锁文件。
   - 完整 audit 记录 103 → 70，四个目标依赖告警清零。删除 6 条已于 2026-10-06 到期的例外，不续期；既有支付 SDK 版本保持原值，没有增加直接依赖能力。
   - 接管旧 v0.2.0 DOMPurify 与 v0.2.12 axios 维护；S03 仅纳入旧 v0.2.9 F05 的 #7638 必要前置。旧轮次其余任务不据此标记完成。

本仓库自 v1.1.17 起另外新增
==========================

- 按本地已有消费者适配来源行为：WS BeforeRequest 逐轮刷新、终态图片用量去重、最小 OAuth 搜索历史 helper、SQLite 串行写事务及既有五平台监控入口。
- 补充真实多连接 WAL 保存/删除交错、跨轮 WS、重试反例、异步 UI 与实际 DOMPurify 净化回归；修正测试中 Gin 全局模式并发写和 SQLite URI 中 #00 导致的意外落盘。
- 回填 OpenSpec 16 要求 / 32 场景 / 54 任务与完整验收输出，归档 138 个实现/配置/测试文件 SHA256 和浏览器截图。
- 更新当前实施进度与剩余维护说明；保留 PORTING/source-baseline 来源 SHA 和历史评估快照。
- VERSION 从 1.1.17 更新为 1.1.18，版本提交先于 annotated tag。实施测试实际使用的 1.1.17 记录保留原值。

验证结果
========

- 后端全量 unit/build、相关 race、SQLite 方言/真实用量去重/删除 Key 结算/固定配额窗口、embed 资源与构建通过。
- Release 的 integration 与发布引用/配置前置脚本通过。
- 前端 281 个文件：2027 passed / 2 个既有 skipped；typecheck、lint、build 通过。
- 浏览器验证分组切换后旧响应隔离及保存、资料草稿保留、实际 Select 添加监控平台。
- 隔离 Node 20.20.2 / pnpm 9.15.9 冻结安装及实际前端构建通过；CGO_ENABLED=0 embed 发布版本注入通过。
- 依赖扫描仍有剩余：全树 70 条（3 critical / 33 high / 31 moderate / 3 low），生产 high 为 xlsx 两条。没有将安全扫描非零记录为通过；Release 工作流的 needs 不包含 security-scan。
- 上述为本地实施/发布前验证。GitHub Actions、跨平台打包与远程上传结果以实际运行记录为准；本次未部署本地或远端业务实例。

明确未合
========

- 支付、充值、购买/续费、退款、返利及 EasyPay/Airwallex/Stripe 支付专属补丁，均不在个人使用范围。
- Haiku 5.5、新平台、插件宿主、TypeSafe、OpenCode/CN 原生协议链、未注册媒体/Realtime 路由；不借修复导入这些平台。
- 显式缓存 CRUD、绑定、租约、存储计费、Vertex API Key 新认证及模型列表新能力。
- 完整 ModelAllowlist/pinned/live 目录、平台清单重构、Free Fast/Ultrafast、动态阶梯/effort 大价格体系。
- TPS/分位数、账号筛选紧凑布局、分组多语言、运维新增展示与临时不可调度规则等第三档能力。
- xlsx 与剩余工具链专项维护仍单列；不采用上游审计例外延期作为修复。
- UsageCleanupService、AccountExpiryService、ScheduledTestRunnerService 仍按现有 SQLite 策略跳过，本次仅修相应前端状态。
- 旧工具桥尾项、Grok 429/Realtime/抢占、跨节点租约、旧 v0.2.9 第一档余项及其余历史遗留仍以最新 PORTING §5 为准；本次局部交叉修复不代表整个旧簇落地。

完整来源与验收
==============

- 评估及旧遗留：docs/upstream-sync/PORTING-0.2.16.md（评估正文保持冻结，顶部记录当前进展）。
- 实施验收：openspec/changes/port-upstream-0.2.16-tier2/verification.md。
- 原始输出/浏览器/依赖报告：docs/upstream-sync/implementation-0.2.16-tier2/。
