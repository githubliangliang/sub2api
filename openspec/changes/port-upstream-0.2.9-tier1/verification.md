# 实施验收证据（尚未实施）

评估证据单列：[基线检查](../../../docs/upstream-sync/evidence-0.2.9/baseline-tests.md)。不作为下表完成证据。
实际实施起点SHA：

完成SHA：

执行日期：

| ID | 失败场景/正常反例 | 改前结果 | 改后结果及退出码 | 用例/证据路径 | 实施SHA |
|---|---|---|---|---|---|
| F01 | 账号 base_rpm、rpm_strategy、rpm_sticky_buffer 经过缓存投影后仍参与候选限流；tiered 红区拒绝、sticky_exempt 仍仅允许粘性；未配置账号保持现状 | | | | |
| F02 | 更新认证凭据时保留原 model_mapping 等非认证设置，新认证字段覆盖旧值；新旧字段冲突、显式空值遵守已有 MergeCredentials 语义 | | | | |
| F03 | 非法 UA 回退时保留已解析版本；读到用量快照不能抹掉 refresh token 拒绝错误；合法 UA 保持身份；真实恢复通过既有恢复入口清错 | | | | |
| F04 | Anthropic thinking.type=disabled 优先于 output_config.effort，两条 OpenAI 桥均输出 none；未禁用时保持 medium 默认和 max→xhigh；none 不请求 reasoning summary | | | | |
| F05 | mimic 路径保留显式 structured-outputs beta；OpenAI 出站保留多智能体 beta 并剔除旧 Responses token；未知 Anthropic beta 继续受白名单/丢弃策略限制；空请求不自动注入 | | | | |
| F06 | 角色消息带 type=message；终止事件空正文时回填已积累的文本；已有非空最终文本优先，工具结果不得被正文替换 | | | | |
| F07 | 仅在 response.completed 且状态成功后返回可计费用量；error/failed/incomplete、只有 delta 或 DONE、提前 EOF 必须失败 | | | | |
| F08 | 陈旧但有明确未来 reset 的额度快照继续用于暂停，达到 reset 后放行；无未来 reset 的陈旧快照保持 fail-open；兼容规范窗口与旧字段 | | | | |
| F09 | 验证码输入同步、错误文本可读，弹窗销毁后迟到响应不重启计时器；成功流程不变；关闭后返回的网络响应不得重建 timer | | | | |
| F10 | 分页跳转、对话框标题 ID、复制失败、标签 Tab、IME、下拉键盘与搜索、日期关闭和跨午夜选择保持一致；IME 组合期间不提交；多个弹窗不共享 ID；键盘及鼠标选择均可用 | | | | |
| F11 | 同代理重复测试共享进行中的结果，过期边界展示与后台一致；不同代理不串结果；临界时刻及无到期时间正常 | | | | |
| F12 | 设置和 Antigravity 映射加载失败可重试，替换分组显示错误，倍率必须正数，RPM override 只接受非负整数；倍率0/负数拒绝；RPM 0允许、空值和小数拒绝；失败不伪装保存成功 | | | | |
| F13 | 错误详情、用户 API Key、临时不可调度状态仅接受当前对象请求，分组弹窗释放监听和待执行搜索；A→B 快速切换及关闭后迟到响应不覆盖 B 或重开弹窗 | | | | |
| F14 | CSV 保留缺失推理力度的单独减号占位符；区间输入正确解析科学计数法；空闲窗口有 reset 时显示倒计时；其它CSV公式前缀继续转义；不重算历史费用；无 reset 时允许显示可用 | | | | |
| F15 | 已识别的 Responses input item ID 超过64字节时移除；Grok 冷却期仍能查询配额；短合法 ID、未知类型及无 ID 保持本地语义；配额查询仍校验认证 | | | | |

| 最终门禁 | 命令/结果 | 证据 |
|---|---|---|
| 后端unit/build | | |
| go test -list与实际用例存在 | | |
| SQLite方言/真实库 | | |
| 并发与取消race（如涉及） | | |
| 前端typecheck/Vitest | | |
| UI/端点实测（如涉及） | | |
| VERSION/迁移/依赖/平台范围 | | |
| 未通过或未运行项 | | |

未运行、超时、工具缺失不得写PASS；源码有测试名也不代表该测试成功运行。
