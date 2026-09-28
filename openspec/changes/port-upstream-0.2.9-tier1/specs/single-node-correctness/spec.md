## ADDED Requirements

### Requirement: F01 调度 RPM 配置投影
系统 MUST 满足：账号 base_rpm、rpm_strategy、rpm_sticky_buffer 经过缓存投影后仍参与候选限流。

#### Scenario: F01 tiered配置经过缓存
- **WHEN** base_rpm=10、rpm_strategy=tiered、buffer=1的账号已用11 RPM，序列化后再解码
- **THEN** 三个配置字段保留，候选被拒绝；sticky_exempt同样负载仅允许已有粘性

### Requirement: F02 OAuth 重授权保留账号设置
系统 MUST 满足：更新认证凭据时保留原 model_mapping 等非认证设置，新认证字段覆盖旧值。

#### Scenario: F02 部分OAuth凭据
- **WHEN** 旧凭据含model_mapping及旧access_token，重授权只返回新access_token
- **THEN** model_mapping保持原值，access_token替换为新值，未返回的配置字段不丢失

### Requirement: F03 账号身份与错误状态
系统 MUST 满足：非法 UA 回退时保留已解析版本；读到用量快照不能抹掉 refresh token 拒绝错误。

#### Scenario: F03 非法UA与指定版本
- **WHEN** 配置已解析的CLI版本，UA含CR/LF或不能配对的指纹
- **THEN** 回退为合法规范UA，使用已解析版本而非丢弃该版本

#### Scenario: F03 快照不证明恢复
- **WHEN** 账号因refresh token拒绝而报错，读取缓存用量快照成功
- **THEN** 不清除账号错误；只有既有真实恢复流程可清除

### Requirement: F04 关闭 thinking 的桥接语义
系统 MUST 满足：Anthropic thinking.type=disabled 优先于 output_config.effort，两条 OpenAI 桥均输出 none。

#### Scenario: F04 显式禁用优先
- **WHEN** thinking.type=disabled且output_config.effort=high，分别走Responses和Chat桥
- **THEN** 两者reasoning effort为none，Responses不请求summary

#### Scenario: F04 未指定thinking
- **WHEN** 未给thinking及effort，或effort=max
- **THEN** 前者保持medium，后者映射xhigh

### Requirement: F05 显式 beta 头兼容
系统 MUST 满足：mimic 路径保留显式 structured-outputs beta；OpenAI 出站保留多智能体 beta 并剔除旧 Responses token。

#### Scenario: F05 结构化输出beta
- **WHEN** mimic请求显式携带structured-outputs-2025-11-13且未被策略drop
- **THEN** 该token保留；未知token不顺带放行；策略drop仍优先

#### Scenario: F05 Responses混合beta
- **WHEN** OpenAI请求携带旧Responses beta及需要保留的多智能体beta
- **THEN** 只删除旧Responses token；空beta不自动注入其它token

### Requirement: F06 Responses 消息与最终正文
系统 MUST 满足：角色消息带 type=message；终止事件空正文时回填已积累的文本。

#### Scenario: F06 角色项类型
- **WHEN** Chat system/user/assistant文本分别转换为Responses input
- **THEN** 每个角色项都含type=message及原角色

#### Scenario: F06 空终态回填
- **WHEN** 已积累delta文本hello，terminal output数组非空但message只有空白文本
- **THEN** 返回hello；若terminal已有world则保留world，工具output不被替换

### Requirement: F07 Alpha search 成功判定
系统 MUST 满足：仅在 response.completed 且状态成功后返回可计费用量。

#### Scenario: F07 成功才返回
- **WHEN** alpha search SSE含response.completed，response.status=completed
- **THEN** 返回解析结果并允许成功用量路径处理

#### Scenario: F07 失败或截断
- **WHEN** SSE只有文本delta/DONE/EOF，或含failed/incomplete/error，或completed缺response
- **THEN** 返回错误，不返回可计费成功结果

### Requirement: F08 未来重置时间前保持配额暂停
系统 MUST 满足：陈旧但有明确未来 reset 的额度快照继续用于暂停，达到 reset 后放行。

#### Scenario: F08 旧快照未来重置
- **WHEN** 额度已到暂停阈值，快照过期，但规范窗口reset仍在未来
- **THEN** 保持暂停并保留已知reset时间

#### Scenario: F08 失去暂停依据
- **WHEN** reset已经到达，或陈旧快照没有已知未来reset
- **THEN** 按既有fail-open规则不再依靠该旧百分比暂停

### Requirement: F09 TOTP 输入与计时器
系统 MUST 满足：验证码输入同步、错误文本可读，弹窗销毁后迟到响应不重启计时器。

#### Scenario: F09 输入与错误
- **WHEN** TOTP设置组件更换输入值并提交，API返回有可读消息的错误
- **THEN** 提交当前验证码，展示可读API错误而非对象字符串

#### Scenario: F09 销毁后迟到响应
- **WHEN** TOTP弹窗请求进行中关闭/卸载，随后返回成功并推进fake timer
- **THEN** 不重新建立倒计时；重新打开的独立会话正常工作

### Requirement: F10 公共表单交互
系统 MUST 满足：分页跳转、对话框标题 ID、复制失败、标签 Tab、IME、下拉键盘与搜索、日期关闭和跨午夜选择保持一致。

#### Scenario: F10 数字跳页
- **WHEN** 分页jumpPage由数值输入产生数字2，提交跳转
- **THEN** 不会调用数字.trim导致异常；跳到第2页

#### Scenario: F10 对话框唯一ID
- **WHEN** 同时挂载两个BaseDialog
- **THEN** 标题ID不同，各自aria-labelledby指向自身标题

#### Scenario: F10 复制失败
- **WHEN** 旧剪贴板fallback的execCommand抛错
- **THEN** 返回失败状态而不抛未处理异常

#### Scenario: F10 标签Tab
- **WHEN** 空标签输入按Tab；随后非空新标签按Tab
- **THEN** 空值允许焦点移动；非空值添加一次标签并阻止该次Tab默认动作

#### Scenario: F10 搜索IME
- **WHEN** 中文compositionstart到compositionend期间连续输入拼音
- **THEN** 组合中不发中间搜索，完成后按最终文本防抖搜索

#### Scenario: F10 Select筛选焦点
- **WHEN** 下拉已打开后搜索过滤选项，再按Enter；另用键盘打开无搜索下拉
- **THEN** 焦点移动到过滤后首个可用项并选中；无搜索下拉取得键盘焦点

#### Scenario: F10 取消日期草稿
- **WHEN** 改日期范围但按Escape/点击外部关闭，随后重开
- **THEN** 恢复父组件已应用范围，不能复用未应用草稿

#### Scenario: F10 跨午夜日期
- **WHEN** 组件在本地午夜前挂载，午夜后点击今天或最近范围预设
- **THEN** 按点击时日期计算，结束日期/输入max不缓存前一天

### Requirement: F11 代理管理小修
系统 MUST 满足：同代理重复测试共享进行中的结果，过期边界展示与后台一致。

#### Scenario: F11 代理重复测试
- **WHEN** 同一代理已有进行中测试，再发一次；同时测试不同代理
- **THEN** 同一代理不重复请求，不同代理结果彼此隔离

#### Scenario: F11 到期边界
- **WHEN** 代理expires_at等于当前时间
- **THEN** 按已到期展示；没有expires_at的代理不显示虚假过期

### Requirement: F12 配置表单错误与输入边界
系统 MUST 满足：设置和 Antigravity 映射加载失败可重试，替换分组显示错误，倍率必须正数，RPM override 只接受非负整数。

#### Scenario: F12 可重试加载
- **WHEN** 第一次Settings或Antigravity映射加载失败，第二次成功
- **THEN** 失败不写loaded=true或永久缓存空数组，第二次会重新请求

#### Scenario: F12 替换分组失败
- **WHEN** GroupReplace API返回有消息的错误
- **THEN** 用户看到该消息，不能显示成功状态

#### Scenario: F12 倍率与RPM校验
- **WHEN** 分别输入倍率0/-1/0.5，以及RPM空值/0/1.5/2
- **THEN** 倍率只接受0.5；RPM接受0和2，空值/1.5拒绝且不提交

### Requirement: F13 异步弹窗状态隔离
系统 MUST 满足：错误详情、用户 API Key、临时不可调度状态仅接受当前对象请求，分组弹窗释放监听和待执行搜索。

#### Scenario: F13 A到B乱序返回
- **WHEN** 错误详情、UserApiKeys和TempUnsched弹窗从对象A切到B，B先返回A后返回
- **THEN** 最终只显示B的数据/错误，A返回不修改B的loading状态

#### Scenario: F13 卸载清理
- **WHEN** 分组RPM/倍率弹窗已注册document点击监听并有待执行搜索，随后卸载
- **THEN** 清除监听与定时搜索，不再发送卸载后搜索

### Requirement: F14 用量与价格显示
系统 MUST 满足：CSV 保留缺失推理力度的单独减号占位符；区间输入正确解析科学计数法；空闲窗口有 reset 时显示倒计时。

#### Scenario: F14 CSV占位符
- **WHEN** 导出缺推理力度的字段值为单独字符-，并另导出以=或其他公式前缀开头的内容
- **THEN** 单独-原样保留，其它危险公式前缀继续按已有规则转义

#### Scenario: F14 科学计数法
- **WHEN** IntervalRow token边界输入1e5
- **THEN** 解析为100000而不是parseInt得到的1

#### Scenario: F14 空闲倒计时
- **WHEN** utilization=0、showNowWhenIdle=true且resetsAt有未来时间
- **THEN** 显示真实倒计时；未给reset时才显示立即可用

### Requirement: F15 请求 ID 和配额诊断
系统 MUST 满足：已识别的 Responses input item ID 超过64字节时移除；Grok 冷却期仍能查询配额。

#### Scenario: F15 过长ID
- **WHEN** 已识别message/reasoning/tool-call项携带长度65且前缀合法的ID
- **THEN** 移除ID；长度64的合法ID保留，未知类型遵守本地原逻辑

#### Scenario: F15 冷却查询配额
- **WHEN** 有效Grok认证账号处于调度冷却中，管理员发起配额查询
- **THEN** 使用手动测试认证通道完成查询，不受模型调度门阻断；无效凭据仍失败

### Requirement: Preserve personal SQLite deployment
系统 MUST 保留SQLite唯一数据库、可关闭Redis、现有菜单和用量去重；不得引入支付或上游新迁移。

#### Scenario: Scope verification
- **WHEN** 审核相对实施起点的最终diff与真实SQLite运行结果
- **THEN** 无支付/插件/新平台扩展，无已应用迁移修改，成功用量只写入一次
