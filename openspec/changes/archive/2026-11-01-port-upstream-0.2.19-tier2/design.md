# Design

## Context

Grok原生返回和缓冲转换丢响应模型审计，目录解析及旧表单能覆盖受管成功快照；图片被截断、已启用搜索配置在保存中丢失，Gemini原生目录与映射口径不同。当前x/net存在可测的HTML平方复杂度。

## Decisions

1. Grok前序模型审计为固定大小状态，只在现有Grok转换中生效；返回字段须穿过handler异步usage快照到SQLite。
2. 成功快照与最后尝试是不同状态；根据来源和开始时间处理乱序。严格目录形态与空成功区分，失败不擦掉成功。
3. SQLite的单键CAS必须保留JSON语义比较及其它extra；普通编辑同步保护两项受管值。上下文事务复用、回滚与缓存发布是同一验收边界。
4. 缺的事务/读取helper只落实既有边界，不引入受限管理员密钥或新配置。目录功能只摘现有消费者所需的映射收集/匹配，不扩Vertex/allowlist schema。
5. 上传保留20MiB单part上限；总body、下载和解析完整性分别验证。搜索只合并特定key，清空和缺key语义分开。
6. Gemini普通目录与连接测试采用已有模型，legacy Google One和显式指定保持；新模型目录继续按需。
7. x/net升级按临时modfile量出的强制依赖图；明确近线性PoC、实际调用和利用性证据边界，不宣称其它advisory自动修复。

## Acceptance and order

先验Claude/Codex/Grok生产消费者，再做共享用量、网络/数据完整性与UI。源码文件重叠时顺序整合；改签名/窄仓储能力时更新实际测试stub。每个簇先证明预期行为失败，然后对正例、反例、取消/异步快照和SQLite效果验收。

组合门禁包括完整backend unit/build/lint、相关race、真实SQLite迁移227/quota/dedup，适用frontend与依赖检查。评估期PASS不能填写实施verification；完成后另记实现SHA与实际起点。

## Scope

SQLite、miniredis、simple mode、固定窗口与现有计价/权限契约保持。未选定新schema/平台/诊断能力及支付不进入；Terra通过独立提案定义。原始来源冻结，发布与部署不在此提案。

