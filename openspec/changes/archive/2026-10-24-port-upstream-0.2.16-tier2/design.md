# 设计与适配决策

以v0.2.16完整来源及本地128c5f7bf为准，按行为手工摘取，不能覆盖整个上游文件。

## S01 首次安装管理员校验统一

只有空库真正创建管理员时，缺省邮箱随机生成，显式邮箱可登录，密码8–72字节；CLI/Web/AUTO_SETUP 共用校验。

已有管理员或已有普通用户时先跳过创建，不用遗留无效配置阻断启动，不重置密码；保留安装互斥和锁文件。

## S02 旧版函数调用与拒答转换完整性

旧函数调用按同名队列配对，稳定生成且避开已有 call ID；refusal 在 Chat message/delta/缓冲终态全部保留。

混合现代 tool_calls、同名多次调用、不配对结果和纯拒答均有测试；不双计文本或工具调用。

## S03 显式工具变更 beta 保留及旧前置

只保留客户端显式请求的 structured-outputs、mid-conversation-tool-changes、inline-tools 兼容 token。

effectiveDropSet 始终优先，未知 beta 不通过，不推断 companion token；count_tokens 原有不同规则保持。

## S04 WebSocket 逐轮分组价格与图片用量

第二轮起获取同 Key 同组的新价格快照并绑定本轮；输入/输出图片 tokens 从 usage 或 tool_usage 回填并准确累计。

不更新连接用户/订阅/Key 限额身份；换组/换平台/查询失败回退建连快照；交错轮次不串价、不重复累计。

## S05 受约束的上游兼容重试

特定 thinking_signature_invalid 三段诊断匹配既有一次性恢复；auto 且已知支持 Responses 的 API Key 对完整、无歧义 input 类型拒绝回退原始 Chat，并沿用起始时间。

Responses 形状入站、已输出/已取消、截断或读失败、重复字段、含用量/输出及其它错误不得触发危险重试；不降级强制模式。

## S06 OAuth 搜索历史回放声明

OAuth 回放搜索历史时补 cached-only web_search，Lite 放 additional_tools；原本无工具时保持 tool_choice none，compaction_trigger 仍最后。

已有工具声明不重复；保留显式强制选择；API Key 及 /responses/compact 不套此重写；map/raw/WS 一致且保留大整数。

## S07 组合路由公开模型目录

启用精确路由的公开别名加入对应 /models 与 Codex Responses 目录，去重并保持 endpoint 区分。

禁用/glob 路由不凭空展开；读取失败回退现有目录；原自定义列表与组隔离不放宽。

## S08 删除分组同步清理风控范围

组删除与审核范围更新在同一事务；并发配置保存不可重新写入已删除组。

缺失/畸形设置不阻断删除也不覆盖原值；真正 DB 错误回滚；保留顺序与其它配置字段。

## S09 分组倍率/RPM 编辑异步隔离

打开/切组/关闭/卸载使旧加载失效并清空旧行；RPM 为非负整数，倍率为有限正数。

迟到 success/error/finally 均不改当前组或 loading；不把不合法输入保存，不误合旧 F12 未完成的其它表单。

## S10 异步页面、草稿、确认与弹窗隔离

每个视图按对象/请求代次接收响应，保留未提交草稿；同实例二次验证共享待定 Promise；公告仅成功标记后关闭。

错误/finally/nextTick/分页回包也受失效保护；换用户时清理敏感草稿；不启用仍被 SQLite 跳过的清理或定时测试 worker。

## S11 GPT-6.1 Sol 上游订阅长上下文统计

实际凭据为 OpenAI OAuth-like 且最终上游模型为 GPT-6.1 Sol 时豁免模型价卡长上下文附加费，用户与账号成本口径一致。

API Key、其它模型仍原价卡；显式自定义区间价格不豁免；不引入付费订阅购买/续费或新价格体系。

## S12 Gemini 既有路径的错误、额度和脱敏修复

第三方/Vertex 不套本地模拟日配额，第三方429仅显式重试时间或原全局fallback；缺失缓存在错误策略前识别并返回请求错误，不封号换号；本地Google错误正确归类；AQ凭据文本脱敏。

官方AI Studio/OAuth原额度保持；非缓存鉴权错误仍正常处理；不实现缓存创建/续期/删除。Flash促销价仅按来源2026-12-31有效窗口核对，过期需恢复标准兜底，不默许永久优惠价。

## S13 监控配置补充已有平台入口

用户可加入一个当前缺失且本地后端支持的平台，初始无模型，沿原保存接口持久化。

不重复添加、不加入未实现平台，saving 时禁止修改；不是新增供应商平台。

## S14 Responses 无 namespace 时避免重建

无修改时返回原正文；需要剥离时仅首次分配并复制未变字节段，减少单节点长历史开销。

保留工具调用 namespace 的既有规则不变；空白/数字/转义/非对象和错误回退保真，需来源 benchmark 或 AllocsPerRun 支撑收益。

## S15 Chat 流心跳保活

仅 ping 到来时向 Chat 客户端写标准 SSE 注释保活；后续语义事件正常转换。

不泄露 event:ping，写失败停止；错误事件不得吞掉；心跳后失败按已经写出响应的规则处理，不重复响应或危险重放。

## S16 前端依赖安全维护与旧待办归并

定向更新已有DOMPurify、axios、Vue及source-map-js受影响解析，保留pnpm锁文件一致性；sanitize调用、API错误处理与前端构建回归通过；审计重新记录目标包和剩余项。

不宣称103条全部已解决；xlsx无npm修复路径及Vite/Vitest/tinypool工具链大版本另列旧维护范围。例外只能按实际不可利用条件和负责人复核，不自动续期、不扩大支付功能。

## 数据与验收边界

SQLite事务和DATETIME/DSN规则保持，历史迁移不修改。现有成功用量去重与固定quota窗口推进必须保留。账户上游订阅元数据和token/cost统计可修；支付购买、充值、续费、退款及相关配置依赖不引入。

共享文件按簇串行适配。每簇先记录可触发的缺口，再验收正常路径和反例；纯文案/视觉改动人工核对或浏览器检查，不机械增加字符串快照测试。最终backend unit/build、受影响race、SQLite真实事务、frontend typecheck/lint/build及组件测试分别记录。

评估探针断言旧行为，不能直接作为实施门禁；embed基线失败和audit告警在PORTING已有单独记录，不在此伪装为全绿。
