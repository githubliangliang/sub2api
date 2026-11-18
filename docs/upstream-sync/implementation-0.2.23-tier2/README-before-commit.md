# v0.2.23 第二档实施证据

2026-11-18，用户选择 `port-upstream-0.2.23-tier2` 的 S01–S23。76/76 任务完成，实现保留工作树，未提交/推送/发布/归档或操作业务实例。

实际起点是 `7a0b78fe511f17b0f7eb9d26fcf5e4ff083312c7` 加已有第一档工作树。期间第一档由独立流程提交为 `d8d126da949d2f7cb12283e594bcdd467adc527d`；其55个产品/测试文件逐个和开工哈希一致。本批增量以 [原工作树快照](start.json)、[产品增量](product-increment.diff)、[变更哈希](changed-files.json) 和 [范围核对](scope-final.json) 为准，不把当前 HEAD 伪称为本批提交。

Go 已按 S23 对齐1.27.4。VERSION1.1.25、SQLite229、auth snapshot21、库版本/go.sum、前端锁文件与 Ent/Wire 保留。默认 body64MiB、SSE行40MiB、读写idle120s、stage8MiB不变。

## 组合门禁

| 检查 | 实际结果 | 完整记录 |
|---|---|---|
| backend全量unit | 54个有测试包通过，unit标签；测试子进程移除环境代理，产品代理行为未改 | [command](backend-unit-final-accepted.json) / [output](backend-unit-final-accepted.log) |
| 相关race | service/handler/repository/apicompat/antigravity/web六包，包含embed，246.6秒通过 | [command](backend-race.json) / [output](backend-race.log) |
| lint | 同版本2.14.0用Go1.27.4重新构建，0 issues | [重建](go1274-lint-rebuild-proxy-adapted.json) / [output](backend-lint-scope-clean.log) |
| embed | CGO=0 Go1.27.4，使用本批实际前端产物 | [output](backend-embed-build.log) |
| HTTP/WS/SQLite | TCP协议夹具+真正SQLite229/Ent；轮换barrier、TTL/模型冲突/别名、工具schema、部分结算、成功去重、完整Fast、旧用量队列路径通过 | [output](backend-http-ledger-final.log) |
| 前端完整门禁 | 293文件，2212通过/2原有跳过；typecheck/lint/build通过，critical179/2 | [full](frontend-full-test-final.log) / [required](frontend-final-checks.log) / [build](frontend-build-final.log) |
| 浏览器 | 真Chrome、实际构建应用，桌面1440与移动390；8项操作，120个合成API请求，无pageerror | [output](browser-menu-adapted.log) / [详细记录及截图](browser/results.json) |
| Go兼容 | 隔离Ent331文件+Wire结果与产品一致，go.mod/go.sum无生成副作用；Darwin/Windows编译通过 | [一致性](go1274-generation-consistency.json) / [Darwin](cross-darwin.log) / [Windows](cross-windows.log) |
| 历史保护 | 3691文件原样；5份第一档说明回写全部匹配独立提交，非本批改写；无未归因变更 | [核对](protected-concurrent-writebacks.json) |

[逐簇结果](cluster-results.md) 和 [来源/夹具适配](source-adaptations.json) 分列实际采用行为、NOBASE、原有政策与runner失败。[命令索引](command-index.json) 保留每次失败和最终通过，不删除失败输出来美化结果。

## 性能与资源

单P、同一200消息/200工具合成请求，原实现和采用路径逐个对照字节/映射/TTL/错误和用量语义。原实现参考仅在测试中保留，产品不引入完整COW或大body懒日志。

| 采用hunk | 原实现 | 采用路径 | 限定 |
|---|---|---|---|
| Ollama 200消息单次offset编辑 | 约12.6ms、29.5MB分配 | 约0.42ms、0.33MB | 完整请求含工具/中文/HTML/转义字符串，sjson fallback保持 |
| cache ASCII预检 | 约270µs、114688B | 约44µs、0B | Unicode转义返回完整原路径，约281µs/131072B，未宣称此类请求收益 |
| 工具缓存断点遍历 | 约112µs、180315B | 约95µs、131562B | 保留本地最后工具/TTL策略 |
| 工具名映射遍历 | 约256µs、143763B | 约239µs、95058B | 同长度reverse项原本无次序契约，映射/长度顺序及原字节对照 |
| 多billing header指纹一次 | 约799µs、794KB | 约300µs、449KB | 无匹配header不计算 |
| string反向无命中 | 约93µs、344064B | 约4.5µs、0B | 字符串调用点，既有byte消费者保留 |
| native SSE map gate | 约480µs、244.6KB、3364 alloc | 约192µs、123.6KB、1042 alloc | escaped type/model/usage强制完整解析，观察/终态/Claude Code noop不漏 |
| raw CC实际TCP flush | 约0.59ms、202次flush | 约0.25ms、1次flush | 100事件，同样帧字节，read前等待/EOF/断开保护；纯Recorder CPU没有收益，记录TCP收益后采用 |

[原候选](s22-candidates-order-adapted.log)、[Go1.27.4 map/ASCII原路径](s22-sse-before-compiled.log)、[采用后](s22-accepted.log)、[TCP前](s22-tcp-flush-before.log) / [后](s22-tcp-flush-after.log)。来源 escaped-usage 判词的失败在冻结评估中保留，本批增加 `\u` 回退并通过真实解析/原文用量验证。

[1CPU/1GiB独立夹具](resource-1c1g.log)：`memory.max=1073741824`、`cpu.max=100000 100000`、`memory.peak=69771264`字节。镜像为已缓存的`sha256:7ce4b6dfe35e55397b7cda544f8a13f191b7ae28dc5aad71fe664dbc9bc2623f`。包含8MiB memory spool生命周期与200消息/工具原实现/采用路径5轮基准；无网络、无业务数据库。它不代表完整服务或生产负载的内存承诺。Linux已实测socket选项，但30秒仅来源参数推算，本次不宣称实测死peer断线耗时。

## 边界与失败留存

未导入新模型/平台、30m/12h列或定价、强制/渠道Ultra Fast、PG SQL/依赖、#8816账务lane、#8833 stale grace/coalescing、缺失DeepSeek placeholder消费者、完整ojson parser和全body保活的懒日志。S08缺失的OpenAI-native Anthropic文件、S17缺失监控quota-fetcher不恢复；仍验证实际现有Gateway/兼容及用量/最小监控消费者。

来源 jsonschema-go 不在实际依赖图。只在隔离modfile/overlay引入v0.4.2验收库，永久测试不用此依赖；首次proxy.golang.org/校验下载超时保留，goproxy.cn适配后通过。schema窄sanitizer185→225行，另155行regexp predicate，未覆盖来源约899行parser。实际兼容转换中四个interface-map解析改用既有UseNumber，避免schema default大整数被舍入。

S14 来源GET+DEL适配为实际Redis Lua compare-delete；S15以PR8822最终状态保留ForceCacheBilling；S10本地没有“mapping destination=deny”枚举，保留现有mapping+ceiling-deny契约。S22不复制来源skip-deferred末工具政策，只采用实测遍历改善。

所有误配夹具/签名、缺失mock、旧DTO私有presence期望、循环回源旧mock、SQLite缺列、代理/工具下载超时、浏览器大小写/菜单/百分比舍入断言保留输出。Grok HTTP夹具明确refresh_token、超过既有1h刷新skew的expiry、API-key-isolated cache identity和有效JSON工具input，保持产品认证/路由策略。一个全套中的分配次数断言受同进程后台工作影响：5次隔离和最终完整unit通过，未放宽它的阈值。

浏览器API是合成拦截数据；真实server/SQLite/Redis传输与结算在独立backend夹具中完成。未访问实际上游账户、生产实例或业务数据库。
