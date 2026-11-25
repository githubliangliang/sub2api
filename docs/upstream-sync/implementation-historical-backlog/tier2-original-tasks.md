# Tasks（未实施）

- [ ] 记录真实实施HEAD和范围，重放native JSON/SSE错误信封遗漏及正常停止控制，保护旧证据和工作区输入。
- [ ] 适配有界信号解析：错误信封、OAuth包装/数组、policy与空结果、优先级及正常停止排除。
- [ ] 接入native JSON、SSE及SSE聚合出口，保留原状态/字节、已有有界reader、取消与已观察usage。
- [ ] 接入现有Ops归因/日志/SLA与skip_monitoring，验证policy不作账号故障且同响应不重复记录，不新增处罚或重试。
- [ ] 通过实际本地HTTP及Ops落库、相关unit/race/build/lint、SQLite费用/账号成本/固定quota/dedup控制，覆盖读写失败和资源边界。
- [ ] 保存真实失败/通过、命令与最终文件哈希及范围检查，填写实施verification后再勾任务。
