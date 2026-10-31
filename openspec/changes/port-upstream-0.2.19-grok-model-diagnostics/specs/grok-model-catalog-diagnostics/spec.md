## ADDED Requirements

### Requirement: 诊断必须基于已保存的目录观察判断可信状态
系统 SHALL 只读生成Grok目录的unknown/fresh/stale/source_changed状态，保留最近成功和最近尝试，MUST NOT 为读取诊断获取token、发送HTTP或写账号。

#### Scenario: 同源有效目录在TTL内
- **WHEN** 已保存观察有合法成功时间、非空scope和可信目录且仍在TTL内
- **THEN** 诊断 SHALL 为fresh并显示目录和成功/尝试信息

#### Scenario: 过期、来源变化或无可信观察
- **WHEN** TTL已到、来源不同或观察缺失/无效/来自未来
- **THEN** 诊断 SHALL 分别使用stale/source_changed/unknown
- **THEN** 历史目录 MUST NOT 被当作当前模型缺失的证据

#### Scenario: 后续同步失败或配置回退
- **WHEN** 最近尝试失败或使用configured_fallback，且保留了上次成功观察
- **THEN** 诊断 SHALL 分别展示该尝试和最后可信目录
- **THEN** 配置回退 MUST NOT 被宣称为新的上游发现成功

### Requirement: 模型映射比较必须保守保留别名和独立能力边界
系统 SHALL 只在fresh同源文本目录中报告确定的not_listed，MUST 保留通配符/裸动态别名/未知ID和独立媒体能力的不可验证或范围外状态。

#### Scenario: 确定文本映射未列入新鲜目录
- **WHEN** 已保存确定文本模型目标不在fresh同源目录中
- **THEN** 诊断 SHALL 报告not_listed
- **THEN** 系统 MUST NOT 修改映射、路由或账号调度

#### Scenario: 无法确定或非文本能力
- **WHEN** 映射包含通配符、空目标、裸动态别名或独立媒体/语音能力
- **THEN** 诊断 SHALL 使用unverifiable或out_of_scope而不是模型不存在

### Requirement: 管理端详情和同步结果必须保持诊断读取范围
系统 SHALL 仅为Grok管理端账号详情和同步结果提供可选诊断，MUST NOT 将诊断计算扩散到列表、公共mapper或其它平台。

#### Scenario: 成功同步后的详情结果
- **WHEN** 现有同步接口完成并保存观察
- **THEN** 返回诊断 SHALL 依据最新已保存账号计算
- **THEN** 已有metadata和其它响应字段 SHALL 保留

#### Scenario: 非Grok或列表请求
- **WHEN** 请求其它平台、账号列表或公共映射
- **THEN** 系统 SHALL 保持原有响应行为，不增加Grok诊断计算

### Requirement: 编辑页必须保持保存配置、异步生命周期与时间状态
系统 SHALL 在现有账号编辑页展示目录和映射诊断、差异、刷新/同步状态及未保存提示，MUST 保留当前编辑并阻止迟到结果覆盖其它账号或关闭后的页面。

#### Scenario: 表单尚未保存
- **WHEN** 管理员编辑模型映射或相关凭据
- **THEN** 页面 SHALL 提示诊断仍依据已保存配置
- **THEN** 读取或刷新诊断 MUST NOT 覆盖表单修改

#### Scenario: 失败、切换账号或关闭
- **WHEN** 刷新失败或旧请求在账号切换/关闭/卸载后返回
- **THEN** 失败 SHALL 保留已有诊断并显示失败提示
- **THEN** 迟到结果 MUST NOT 覆盖当前账号诊断或编辑

#### Scenario: 目录在页面打开期间过期
- **WHEN** 成功观察在页面打开期间达到有效期
- **THEN** 面板 SHALL 切换到历史状态并停止当前缺失提示
- **THEN** 多面板控制ID与移动端长文本布局 SHALL 保持可用
