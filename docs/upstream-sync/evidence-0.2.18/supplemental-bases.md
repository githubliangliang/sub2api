# 旧前置与终态辅助

补充来源不计入本轮56提交/312文件统计。

- #5925完整merge 4033387fd49b15b676d2e5c9fd3833f156050f57，按父1完整diff额外全文件先反向再正向check，见supplemental-files.tsv。只选其历史953028718中的SameAccountRetryDelay字段及handler读取逻辑，绝不逐个微提交cherry-pick。
- modelCatalogEntries：冻结v0.2.18树的backend/internal/service/openai_models_list.go，约17行标准JSON解析。来源完整旧文件还包含pinned/投影功能，本地没有；只给S03提供等价解析行为。
- native line pump：冻结v0.2.18树的backend/internal/service/openai_gateway_anthropic_native_pump.go，162行（含注释/import），S05需要纯读取/取消/定时器行为。来源PR8210只是修改该文件，本地缺文件，不能只套增量hunk；两条CN native消费者仍不引入。
- Cloudflare旧HTML helper：fa79b1ebc中isCloudflareBotBlockResponse仅标准字符串判断；S04之外的F04可在本地handle403前补等价判断，并接本轮PR8255的JSON识别，不引入旧OpenCode/Command Code入口。

这里的量测是实施设计，不是已经补入代码；最终完整编译与回归仍在OpenSpec任务中。
