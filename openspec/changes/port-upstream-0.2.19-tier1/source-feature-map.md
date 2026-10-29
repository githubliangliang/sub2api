# Source Feature Map

| 编号 / 行为 | 来源 | 本地缺口与接受范围 | 验收 |
|---|---|---|---|
| F01 Claude Code UTF-16指纹 | #8377 整体PR/直接diff | 按JS UTF-16码元采样后拼接、再按JS UTF-8语义哈希，读取前缀有界。 | ASCII、中文、emoji、采样surrogate重组及不足下标；版本更新/重复调用，所有既有Claude正文与header消费者一致。 |
| F02 Codex模型缓存flight内复查 | #8373 整体PR/直接diff | miss调用进入flight后遇已填fresh缓存直接复用，使用本地实际缓存消费者。 | 确定性交错fetch=0、普通miss/stale仍获取、ETag/304、账号隔离、并发与取消。 |
| F03 共享缓存命中用量 | #8354 整体PR/直接diff | OpenAI-compatible顶层prompt_cache_hit_tokens作为fallback，不重复合并已存在的嵌套缓存桶。 | 嵌套显式0/正值优先、prompt100/hit80/miss20、HTTP/SSE同桶及SQLite成功usage去重，无新DeepSeek平台。 |
| F04 Gemini fileData图片统计 | #8326 整体PR/直接diff | 既有图片observer识别inline/file及camel/snake字段，累计chunk保留最大实际张数。 | 有效MIME与非空URI/data、同part至多一次、累计SSE/别名/旧fallback、取消已观测用量；不下载URI或扩计量模式。 |
| F05 Antigravity畸形工具终态 | #8362 整体PR/直接diff | 直接/包装响应和缓冲流中的MALFORMED_FUNCTION_CALL产生既有格式错误，先保留无parts尾帧失败。 | JSON/wrapped/SSE错误尾帧均不返回空成功，普通文字/工具与取消/错误透传保持，失败用量规则不回退。 |
| F06 分页数值与偏移量边界 | #8303 整体PR/直接diff | 安全解析数字，所有共享offset与offset+limit有界，各入口保留自己的默认/限幅/400策略。 | 超长数字、MaxInt/负数/空值/非法值、最大合法页、直接构造params、Key内存切片及HTTP入口。 |
| F07 管理员保护返回400 | #8352 整体PR/直接diff | 禁止最后管理员降级及直接disable/delete时使用领域400错误，守卫规则不变。 | 真实handler400/稳定码、被拒绝动作不写库、其它管理员存在时允许降级、数据库故障仍500。 |

[评估](../../../docs/upstream-sync/PORTING-0.2.19.md)保存逐文件落点、前置量测与完整四态；本表定义接受的行为，不是整PR导入授权。新增回归需覆盖真实消费者；源integration仅有标签不代表SQLite可跑。当前实际测试目录见[review-test-list](../../../docs/upstream-sync/evidence-0.2.19/review-test-list.txt)，未来新增用例不能被称为baseline现存测试。

其它新能力与支付保持原档位；评估期原始输出、旧任务完成记录及SHA不改写。

