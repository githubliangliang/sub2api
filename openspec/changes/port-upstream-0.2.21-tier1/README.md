# v0.2.21 第一档

2026-11-11：F01–F11 已实施，39/39 任务完成，实现随本次提交保存。起点 `4f4c80c3ce48ce4afbc9344536b9cf44ed3fd898`；本批保持 VERSION1.1.24 / Go1.27.3 / SQLite228 / auth snapshot21。

[验收](verification.md) · [实施证据](../../../docs/upstream-sync/implementation-0.2.21-tier1/README.md) · [任务](tasks.md) · [冻结来源](source-baseline.md) · [行为映射](source-feature-map.md) · [设计](design.md) · [评估与patch sites](../../../docs/upstream-sync/PORTING-0.2.21.md)

Codex SSE元数据/密封参数/namespace、外来reasoning封装、工具会话稳定性、补全失败释放槽位、价格hash与Sonnet缓存读价、Gemini思考正文过滤、并行工具图片顺序、图片HTTP状态和UTF-8输入边界均完成。全量后端unit、embed build、lint0、五包race、实际HTTP/SQLite/miniredis通过。

最终验收使用独立tier1快照：共享工作区存在并行tier2改动，额外内容不归入本批结果；完整tier1 patch已检查仍在共享工作区。原始评估、来源/proposal/design/spec与历史记录保持冻结，实施前README/tasks/verification保存在实施证据目录。前端/依赖/schema/生成代码未由本批改动；未push/release或业务部署。
