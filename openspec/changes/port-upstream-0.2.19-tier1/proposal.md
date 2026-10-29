# Proposal

## Why

现有Claude Code指纹用字节索引偏离JS；Codex冷启动cache存在flight窗口；共享cache-token、Gemini fileData、Antigravity工具失败、分页及管理员错误状态都有本地消费者和已复现缺口。

## What Changes

修复七项既有行为，保持协议、成功用量去重、账号权限及输入错误契约。范围限于无新schema的小改动，先把Claude/Codex路径落稳，再组合用量与管理回归。

- F01：按JS UTF-16码元采样后拼接、再按JS UTF-8语义哈希，读取前缀有界。
- F02：miss调用进入flight后遇已填fresh缓存直接复用，使用本地实际缓存消费者。
- F03：OpenAI-compatible顶层prompt_cache_hit_tokens作为fallback，不重复合并已存在的嵌套缓存桶。
- F04：既有图片observer识别inline/file及camel/snake字段，累计chunk保留最大实际张数。
- F05：直接/包装响应和缓冲流中的MALFORMED_FUNCTION_CALL产生既有格式错误，先保留无parts尾帧失败。
- F06：安全解析数字，所有共享offset与offset+limit有界，各入口保留自己的默认/限幅/400策略。
- F07：禁止最后管理员降级及直接disable/delete时使用领域400错误，守卫规则不变。

## Impact

既有网关、usage与相关管理/模型消费者。测试使用本地HTTP夹具、临时SQLite及miniredis；评估probe只证明当前行为，实施前须增加预期新行为的失败回归。

第一/二档不新增schema或迁移；不引入受限管理员密钥、图片新计量开关、诊断大界面、新平台/新模型目录。用户选定Terra单独立项。保持个人范围与既有权限，支付、发布/tag/push/部署另属授权范围。

