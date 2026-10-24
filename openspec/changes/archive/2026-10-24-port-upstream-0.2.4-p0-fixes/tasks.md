# 实施任务

- [x] 记录实际起点、工作树已有改动与四态复核，保留 source-baseline.md。
- [x] F01：非耗尽 reset + 关闭 fallback 回归先失败，再修复；保留 shadow 与真正耗尽的反例。
- [x] F02：覆盖 Responses、WS、Chat、Embeddings 映射选号；剔除来源中的外网测试调整。
- [x] F03：覆盖 routed、sticky、普通候选与全受限时的退出行为。
- [x] F04：新增 unavailable continuation 错误恢复场景及无关错误反例。
- [x] F05：透传 none 原样保留，非透传规则不变。
- [x] F06：控制字符 JSON 有效且解码内容等价。
- [x] F07：messages/count_tokens 的最终 UA、版本、fingerprint 后缀一致。
- [x] F08：多模型 max_tokens=1 探测可用，UA 与普通请求验证不降级。
- [x] F09：thinking binding 与最终 beta header 对称。
- [x] F10：Fable credits 仅模型冷却；持久化失败不扩大封锁。
- [x] F11：无工具 reasoning 请求含 toolConfig，混合工具行为不变。
- [x] F12：零指标不落 NULL，nil 指标仍落 NULL。
- [x] F13：API Key 不合成 instructions；OAuth 默认、显式 instructions、用户 metadata 修复均回归。
- [x] 对新增项目函数调用和测试 fixture 逐一查定义；仅测试文件干净时不得先落测试。
- [x] 运行目标回归、go build ./...、go test -tags=unit ./... 与 SQLite 方言审计；首轮中断后完整重跑 exit 0。
- [x] 核对 diff 无 VERSION/迁移/Ent/依赖/前端越界，填写 verification.md。
- [x] 向 coordinator 移交实施 SHA 与验证/集成说明，由 coordinator 更新 PORTING 总状态；本角色不修改其所有文档。
