# Source Feature Map

| 簇 | 完整来源 | 选定行为与边界 |
|---|---|---|
| F01 Codex 文献搜索 403 的账号副作用 | [#8982](https://github.com/Wei-Shaw/sub2api/pull/8982) | alpha/search 工具级403 SHALL 允许既有换号流程且不冷却或永久置错健康账号；明确账号/工作区停用的403继续执行现有账号策略。 不能把所有403全局忽略；仅限定搜索工具路径，不放宽普通Responses鉴权与访问状态错误。 |
| F02 OpenAI 408 上游超时换号 | [#8968](https://github.com/Wei-Shaw/sub2api/pull/8968) | 上游 HTTP408 SHALL 进入既有有界换号流程。 不新增重试策略；保留一次idle换号预算及已输出边界。 |
| F03 Codex IDE 与独立上下文 | [#8929](https://github.com/Wei-Shaw/sub2api/pull/8929) | 合法独立 agent 上下文 SHALL 转为不可信 user 文本，保留各类型的空值规则。 不提升到system/developer；未知工具、真实调用及page快照严格校验保持。 |
| F04 Responses 每轮请求推理强度 | [#8931](https://github.com/Wei-Shaw/sub2api/pull/8931) | 请求推理强度日志 SHALL 优先记录本轮原始input最后一个有效configuration_update的effort，并与最终有效强度独立。 保留d2488b110的原始载荷取值、output_config支持、缺省NULL及none/max；不重复加入迁移230或撤销本地已修的bridge采集。 |
| F05 既有 Claude 账号测试 temperature | [#8911](https://github.com/Wei-Shaw/sub2api/pull/8911) | 既有拒绝temperature的Claude家族在连接测试中 SHALL 省略该字段。 不扩充模型目录/价卡；未选Fable/Mythos helper不作为前置。 |
| F06 兼容转换保留 stop 与原始工具定义 | [#8963](https://github.com/Wei-Shaw/sub2api/pull/8963) | 既有兼容降级路径 SHALL 保留客户端stop序列与可转发的原始动态工具定义。 只选已有通用API Key消费者；不引入原生Kimi平台或缺失的Anthropic-native适配器。 |
| F07 图片 SSE usage 单行载荷 | [#8886](https://github.com/Wei-Shaw/sub2api/pull/8886) | 转发图片的completed事件 SHALL 把usage压为单行JSON后写入SSE data行。 不改变图片模型、价格、结果数量或SSE行上限。 |
| F08 Gemini 最后工具轮的尾随内容 | [#8923](https://github.com/Wei-Shaw/sub2api/pull/8923) | 最后一个Gemini工具结果turn的尾随非functionResponse内容 SHALL 保序拆成后续user turn。 不导入新Gemini模型、完整thinking策略或改变已完成的工具媒体顺序规则。 |
| F09 Antigravity 原生 SSE 注释帧 | [#8946](https://github.com/Wei-Shaw/sub2api/pull/8946) | 支持注释的客户端 SHALL 保留上游SSE注释及其分隔；不支持注释的既有SDK继续滤掉注释。 不通过关闭默认ping回避错误；保持120s idle、40MiB行上限和计费drain。 |
| F10 Grok 媒体输入错误的 400 分类 | [#8962](https://github.com/Wei-Shaw/sub2api/pull/8962) | 既有Grok媒体本地输入校验失败 SHALL 返回400并保留可用校验信息。 本簇保持grokMediaMaxEditSourceImages=3；扩为5张是独立第三档，未获选择不随修复进入。 |

patch sites只保存在[评估报告](../../../docs/upstream-sync/PORTING-0.2.24.md)。新增调用扫描是检索而非编译保证；来源helper与本地consumer区别以该报告的逐簇最小前置为准。
