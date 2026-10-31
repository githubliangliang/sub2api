# Source Feature Map

| 编号 | 来源 | 行为与保留边界 | 未来验收 |
|---|---|---|---|
| C01 | e2d9ce0ca + fbc9ee626 | 坏普通工具参数、回放调用/输出过滤；不把有效length或custom输入误判 | HTTP/SSE/历史、合法/非法/空参数、部分usage及dedup |
| C02 | 31d5b67ba | 唯一custom别名还原；显式function、真实namespace和歧义保护 | buffer/stream added/done/terminal身份及碰撞反例 |
| C03 | 7a09a2eaf | 最终声明没有内建tool_search时去除defer标记；保留有效前置 | 通用adapter/Grok实际出站、同名function与真正内建search |
| C04 | PR7129/25c21aedd | 0及合成事件的序号字段；不重置既有正值或有序流 | 首帧、compact、handler错误、WS桥和正值/递增反例 |

历史来源中的邻近主体已经实现。仅上述未合行为进入方案；不导入完整Grok重试、工具新平台、PG-only测试基座或其它schema能力。当前失败探针放在评估证据中，未来verification只记录实际实施结果。
