# Proposal

## Why

当前v1.1.21已经完成历史主体与选定个人功能，但四个明确排除的兼容性尾项仍可在现有出口复现：截断工具参数被标为完成、custom别名未还原、降级工具残留defer标记、事件序号字段遗漏。它们影响现有Codex和Grok使用路径，建议优先于新增展示和平台能力。

## What Changes

- C01：普通function-call参数不能以非法JSON完成；回放中的坏调用与其输出配套过滤；合法JSON、既有空参数默认及custom输入保留。
- C02：按显式function、真实namespace归属和无歧义custom别名规则还原工具身份，HTTP/SSE一致。
- C03：依据最终声明中的真实内建tool_search清理孤儿defer_loading。
- C04：Responses第一帧0及本地合成事件写出sequence_number，保留既有数值并保持本地有序事件递增。

## Impact

修改既有协议转换与生产消费者并补相应回归。无需新schema、依赖或平台。usage结果必须沿既有HTTP/SSE/WS写入规则保留并去重，不改变配额和计价政策。

本方案只建议实施C01–C04；评估探针不构成完成依据。Grok诊断、费用/TPS展示、重置grant、生图开关、Fast/Realtime、完整目录/调度系统仍按需，支付和不适用架构排除。
