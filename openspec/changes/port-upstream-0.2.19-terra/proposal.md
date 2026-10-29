# Proposal

## Why

用户在评估v0.2.19时明确需要GPT-6.1 Terra。当前代码缺少这个模型的统一身份、专用价卡、Codex声明及已有配置生成；只加一个下拉项不能让none/max、工具协议和实际usage行为一致。

## What Changes

- 在既有OpenAI/Codex兼容路径识别Terra及允许的effort/compact拼写。
- Terra支持none、low、medium、high、xhigh、max，默认medium；显式thinking禁用仍优先。Sol既有none/minimal拒绝规则保持。
- 有工具的Terra使用Responses；only-Chat fallback在出站前返回400。
- 以已有metadata优先和完整目录保留契约接入Codex模型描述，不夸大账号能力。
- 只摘Terra资源/两层fallback及已有缓存/长上下文/priority成本策略，保留显式0和渠道覆盖。
- 为既有白名单、映射预设与客户端配置生成添加Terra。

## Impact

属于用户选定第三档能力，不提升为第一/二档修复。18文件/+292 -17的整PR仅按行为取用；当前fork的协议、pricing与catalog结构不同，需要适配。

不引入新平台、Fast/Free Fast/Ultrafast大机制、支付或新schema。模型默认能力和价格使用冻结来源值；实际账号metadata优先，最终可用性不得由公共默认条目推断。本轮完成提案，实施、发布/部署状态分别记录。

