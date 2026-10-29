# S01 验收口径确认（2026-10-29）

用户确认：保持迁移 227；冲突位保留到 usage 写入边界，并验证 SQLite 计价回退。

原始场景要求冲突位进入 SQLite record，而 proposal 同时禁止 schema/迁移变更。当前表只有 upstream_response_model 与 upstream_model_mismatch，没有独立冲突列；后者比较实际发出的模型与终态响应模型，不改成跨事件冲突标志。

据此仅修订 S01 该场景的完成条件；[原规格快照](./s01-spec-original.md)保留。实现验证响应模型穿过 handler 异步任务落库，并用同一有冲突响应证明 response_model 定价回退。source-baseline、source-feature-map 与其它规格保留原文。
