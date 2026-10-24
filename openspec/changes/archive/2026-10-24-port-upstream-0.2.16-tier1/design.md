# 设计与适配决策

以v0.2.16完整来源及本地128c5f7bf为准，按行为手工摘取，不能覆盖整个上游文件。

## F01 删除 Key 后完成请求仍结算

请求执行中 Key 被软删除，余额/账号配额及 dedup 按既有模式正常落库，已删除 Key 自身计数跳过。

非 Key-not-found 的数据库错误仍失败回滚；重放相同请求不重复计数；simple mode 原本未启用的余额扣减不被开启。

## F02 Chat 请求角色与具名工具选择

具名 function choice 转为 Responses 的顶层 name，developer 保持 developer。

auto/none/required 及已经是 Responses 格式的 choice 不变；空名称和普通 system/user 消息保持原语义。

## F03 Anthropic 缓冲索引与 thinking 输出

负数内容块索引被忽略；thinking 块即使内容/签名为空也有 thinking、signature 字段。

普通文本/工具块不增加无关字段；正索引输出及后续 signature_delta 保真。

## F04 Claude 拟态缓存 TTL 与 Sonnet 采样参数

缓存块数量限制后将较早 ephemeral 断点升至后续 1h；Sonnet 仅有 top_p 时不注入 temperature。

已合法 TTL 顺序/无断点请求不变；Opus/其它模型原规则和四块上限保留，不顺带开放 Haiku 5.5。

## F05 Grok 空完成事件与 SSO consent

Grok 空 response.completed 按现有空流失败规则处理；SSO 同一 flow 从合法 consent 页读取 hidden token 并提交 origin/referer。

有输出完成事件保持成功；旧页面没有 consent 字段仍兼容；重定向不泄露首跳额外头。

## F06 Codex 已有模型元数据与远程发现配置

remote 模式的生成配置包含发现开关；生成的 GPT-6.1 Sol 默认上下文为来源值 500000。

file 模式不加远程发现开关；显式上游 context_window 和用户目录值仍优先；不引入新模型/Ultrafast。

## F07 Gemini 工具图片与 Antigravity 图片冷却隔离

base64 工具结果图片转为 inlineData，工具响应顺序保持；图片模型冷却只写/查对应图片键。

普通文本工具结果及 Gemini 文本家族冷却不变；不下载远程图片、不新增模型级规则 UI。

## F08 价格、倍率与吞吐显示

科学计数价格与亚字节速率正确显示，零倍率保留，长上下文徽标不承诺统一倍数；只在空条目首次加入单模型时自动填价。

已设价格、已有模型追加、一次多模型加入及缺失倍率默认1保持边界；不改变账本计算规则。

## F09 共享弹窗和选择器输入边界

Esc 只关栈顶；disabled 收起并清理搜索；isComposing 时不处理导航键。

保留原焦点归还/多弹窗 body 锁和非 IME 键盘操作；不重复移植旧 dialog ID 前置。

## F10 管理端错误文本与开关响应

使用统一解析显示后端原因和合适 fallback；开关状态取服务端返回 enabled。

不显示对象字符串或仅 Axios 通用文本；失败不伪造状态；不改接口和平台目录。

## F11 计时器、访问重试与备份卸载

卸载后回包不重启计时器，回到可见标签立即更新倒计时；暂时查询失败后可再次检查访问权限。

成功权限结果仍缓存，重复轮询不增生；不启用后台服务、余额通知或批量生图功能。

## F12 现有监控健康分与时间/设置边界

请求样本不足保持未知；规则分钟为整数；时间按本地初始化；自动刷新与手动刷新使用同一 detail 流程；未完成设置加载禁止保存。

达样本阈值正常评分；刷新保持用户当前筛选；不启用隐藏 ops 菜单或增加 TPS 字段。

## F13 账号测试诊断与分组名显示

测试列表列显式请求侧映射名并排序；HTTP/后台失败日志有账号、测试及请求关联且先脱敏后截断；账号表完整显示长组名。

无映射返回默认列表；普通 GroupBadge 默认仍截断；不通过日志泄露凭据，不更改测试执行计划。

## F14 内嵌前端放行已有 API 别名

本地已注册 API 别名到达网关；测试使用当前构建实际静态资源。

正常 SPA 深链接和静态资源保持；不注册上游缺失的视频/语音/Realtime API。

## 数据与验收边界

SQLite事务和DATETIME/DSN规则保持，历史迁移不修改。现有成功用量去重与固定quota窗口推进必须保留。账户上游订阅元数据和token/cost统计可修；支付购买、充值、续费、退款及相关配置依赖不引入。

共享文件按簇串行适配。每簇先记录可触发的缺口，再验收正常路径和反例；纯文案/视觉改动人工核对或浏览器检查，不机械增加字符串快照测试。最终backend unit/build、受影响race、SQLite真实事务、frontend typecheck/lint/build及组件测试分别记录。

评估探针断言旧行为，不能直接作为实施门禁；embed基线失败和audit告警在PORTING已有单独记录，不在此伪装为全绿。
