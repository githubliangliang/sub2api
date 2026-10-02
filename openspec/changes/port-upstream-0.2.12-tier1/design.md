# 设计与适配决策

按v0.2.12完整PR终态手工适配，保留本fork已落0.2.11实现。

- F01保持包级minimum与transport preferred pin区别；既有语义版本库可复用，标准库runtime生成平台UA。最终HTTP transport和官方API回退是必须覆盖的消费者，额度探测不涉及支付。
- F02采用来源97行独立helper并接到现有出口；只回传Gemini三个字段，源状态与已知错误文字保持可诊断，details移除。使用实际handler/service出口断言，不能只测正则。
- F03中英文与现有ShouldHandleErrorCode契约对齐，不增加新错误处理策略。

## 数据与范围边界

不改schema、历史迁移、锁文件、VERSION或菜单结构。支付两PR、TypeSafe原生平台与axios独立升级不引入。测试首先证明目标缺口，再验证正常路径与反例；后端全量unit/build、相关race与SQLite使用记录去重，前端typecheck/lint及真实组件测试在实施验收中记录。

## 实施适配（2026-10-02）

F02按本fork已有重试路径仅替换最终错误输出，未覆盖整个上游文件；真实ForwardGemini出口用例与现有成功/failover回归通过。F03为面向人的说明，逐语言核对现有ShouldHandleErrorCode语义，不增加字符串快照测试。
