# Design

## Context

现有Claude Code指纹用字节索引偏离JS；Codex冷启动cache存在flight窗口；共享cache-token、Gemini fileData、Antigravity工具失败、分页及管理员错误状态都有本地消费者和已复现缺口。

## Decisions

1. Claude采样只读取覆盖下标20的Unicode前缀，先拼接UTF-16码元再编码，避免一次新全prompt分配。
2. Codex在实际singleflight闭包内重新读取fresh状态，ETag/304与缓存键保持，不新增模型目录系统。
3. cache字段仅作为fallback；图片统计是已有请求内observer的扩展，不加下载或新计量开关。
4. 先保留畸形工具尾帧失败再聚合；HTTP状态沿现有compat出口，不伪造成功。
5. 分页安全保证offset+limit；管理员只修改错误分类，保持禁止动作的数据库效果。

## Acceptance and order

先验Claude/Codex/Grok生产消费者，再做共享用量、网络/数据完整性与UI。源码文件重叠时顺序整合；改签名/窄仓储能力时更新实际测试stub。每个簇先证明预期行为失败，然后对正例、反例、取消/异步快照和SQLite效果验收。

组合门禁包括完整backend unit/build/lint、相关race、真实SQLite迁移227/quota/dedup，适用frontend与依赖检查。评估期PASS不能填写实施verification；完成后另记实现SHA与实际起点。

## Scope

SQLite、miniredis、simple mode、固定窗口与现有计价/权限契约保持。未选定新schema/平台/诊断能力及支付不进入；Terra通过独立提案定义。原始来源冻结，发布与部署不在此提案。

