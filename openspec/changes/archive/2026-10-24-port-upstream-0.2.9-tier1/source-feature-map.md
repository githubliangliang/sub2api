# 来源到行为映射

| ID | PR / 固定merge SHA | 行为 |
|---|---|---|
| F01 调度 RPM 配置投影 | #7311 `5cc6ca6f5bf33590517099c43590726102a37551` | 账号 base_rpm、rpm_strategy、rpm_sticky_buffer 经过缓存投影后仍参与候选限流 |
| F02 OAuth 重授权保留账号设置 | #7404 `c619ff8474acccd728904a08a6c7daf1b731d10e` | 更新认证凭据时保留原 model_mapping 等非认证设置，新认证字段覆盖旧值 |
| F03 账号身份与错误状态 | #7349 `53b4bbe736323abac37dcf0a5f1ed7cd7893008f`；#7362 `8bb48622a3893c7a968756bb7d4a5d9cfda5c029` | 非法 UA 回退时保留已解析版本；读到用量快照不能抹掉 refresh token 拒绝错误 |
| F04 关闭 thinking 的桥接语义 | #7538 `b5298fdde8e9c5ba4503aca9bac34c1f87774273` | Anthropic thinking.type=disabled 优先于 output_config.effort，两条 OpenAI 桥均输出 none |
| F05 显式 beta 头兼容 | #7638 `4c00df2e0183e2c70b7fa8ba45914205e36aad0c`；#7617 `8616d4e03610e399cba00f62528e16b1141c6b5c` | mimic 路径保留显式 structured-outputs beta；OpenAI 出站保留多智能体 beta 并剔除旧 Responses token |
| F06 Responses 消息与最终正文 | #7635 `be64b553f99cd1f56f9c0ac91aa178197c4b7e4c`；#7569 `8aa7fb8c2d28b5a6b9e17be037eb4bac72f3cde1` | 角色消息带 type=message；终止事件空正文时回填已积累的文本 |
| F07 Alpha search 成功判定 | #7572 `06f70a6232f4d764b367caca71cca4c9dedb24f1` | 仅在 response.completed 且状态成功后返回可计费用量 |
| F08 未来重置时间前保持配额暂停 | #7624 `90ae81e9f8ecbcab633d3df1aa4813a8afffdf97` | 陈旧但有明确未来 reset 的额度快照继续用于暂停，达到 reset 后放行 |
| F09 TOTP 输入与计时器 | #7235 `efd86c63e6724838659cde590a0d4516d045acd0`；#7184 `71c0eb51260fd238b5e485a0127770fe33c8d9d1`；#7496 `9e0e1469844665c4bd3220bbec5836ab48b17537` | 验证码输入同步、错误文本可读，弹窗销毁后迟到响应不重启计时器 |
| F10 公共表单交互 | #7234 `4daabc3d84ddcbaa4df9f74f2f8ac4db6fd157a0`；#7263 `ea68c0e82c1c995a8eb797648af72d2088e2955f`；#7186 `5f75e0d2a7ea7197476be4e77596735c459008f7`；#7238 `57e9f6a4b81db65a9bd1c23c9c6cc7965a8e686b`；#7374 `c96d36ec9567c98a378e0b93b46ee2d9b4242994`；#7142 `ea34a2356ca546cb12cd960efc203baaff5e39c4`；#7109 `d27fbb3b7dfd6a2a5a26e80fbca781fc7bbdc112`；#7108 `974a819ba15359b00b87f4c96ad3b7ae3b4a6f93`；#7499 `0eaa7c3c84c53704438809f52887c44824b14b95` | 分页跳转、对话框标题 ID、复制失败、标签 Tab、IME、下拉键盘与搜索、日期关闭和跨午夜选择保持一致 |
| F11 代理管理小修 | #7183 `53fd5f83512efbe95891e16138ecfd57f9bda8b7`；#7321 `4b0adb68636c4d8bef479e2c1ed541950a031c27` | 同代理重复测试共享进行中的结果，过期边界展示与后台一致 |
| F12 配置表单错误与输入边界 | #7372 `c0b23ff79116baec40fe2d66b33eccb6b74427a2`；#7264 `e7d348868c3891d0a63889807537552bd4c0cadc`；#7364 `9e175bd49161d486b3dd5d87f43724d98a45b07a`；#7376 `ca47fa352c6942907ae01b3c36eb508bc38d01e5`；#7367 `a765b3c1765f50a18e3b845170916d91c21a12cc` | 设置和 Antigravity 映射加载失败可重试，替换分组显示错误，倍率必须正数，RPM override 只接受非负整数 |
| F13 异步弹窗状态隔离 | #7420 `f2e55bf4d704ae5730f207d437f563c41be9573a`；#7422 `5af29a3b93122179568195d92c8f02afe45231fb`；#7493 `24872fda7cf89ad6be5f059751dc01e741da1f9e`；#7497 `89e9976950a8928d0285fb7efd535cdf5c7a5b38` | 错误详情、用户 API Key、临时不可调度状态仅接受当前对象请求，分组弹窗释放监听和待执行搜索 |
| F14 用量与价格显示 | #7446 `5295bd822780c1242d8bd034f817a417367be7b9`；#7418 `5fd346114252ff66f49dc0e70b6f4e9958e0c601`；#7611 `99d86e41acb47d632de27598a727c671ed7fadb1` | CSV 保留缺失推理力度的单独减号占位符；区间输入正确解析科学计数法；空闲窗口有 reset 时显示倒计时 |
| F15 请求 ID 和配额诊断 | #7488 `7be1628a74c0a42f2cb9e04e1a5b1d88ab58a115`；#6920 `20a94fbb567b62208751292ed7786b24a7e7c0fe` | 已识别的 Responses input item ID 超过64字节时移除；Grok 冷却期仍能查询配额 |
