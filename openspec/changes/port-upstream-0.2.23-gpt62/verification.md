# 验收

2026-11-19 完成，12/12 任务。实现随本次提交保存，等待归档。[结构化结果](../../../docs/upstream-sync/implementation-gpt62/checks.json) / [证据索引](../../../docs/upstream-sync/implementation-gpt62/README.md)。

| 项目 | 结果 | 证据 |
|---|---|---|
| 模型、协议、Codex metadata | 两个模型、effort/compact、采样过滤、映射后的三种 HTTP 兼容校验、API full Responses 和显式账号目录通过 | [后续服务 race](../../../docs/upstream-sync/implementation-gpt62/after-lint-cleanup.jsonl.gz) |
| 价格、272k、促销、延迟结算 | 资源/双层 fallback 一致；271999/272000/272001、缓存桶、Fast2、促销纳秒边界与时区、resolver/预解析并发复用、独立账号成本和自定义0通过 | [模型 race](../../../docs/upstream-sync/implementation-gpt62/backend-race.jsonl.gz) |
| HTTP/WS/SQLite 结算和去重 | 48个实际本地 HTTP/WS+SQLite 场景通过，覆盖新旧模型、截止前/截止时、Fast/Free Fast/Ultra Fast、余额、quota和同请求去重 | [HTTP/WS/SQLite](../../../docs/upstream-sync/implementation-gpt62/http-ws-sqlite.log) |
| 后端 unit/race/lint/build | 全量54测试包通过；四包race44顶层/390子项通过；清理后服务race34顶层/120子项通过；Go1.27.4/lint2.14.0为0 issues；embed构建/版本通过 | [unit](../../../docs/upstream-sync/implementation-gpt62/backend-unit.jsonl.gz)、[lint](../../../docs/upstream-sync/implementation-gpt62/backend-lint.log)、[构建](../../../docs/upstream-sync/implementation-gpt62/checks.json) |
| 前端配置及适用检查 | 45定向、293文件/2213测试通过，2项既有跳过；typecheck/lint/build通过；真实Vue组件导出及白名单往返通过 | [全量](../../../docs/upstream-sync/implementation-gpt62/frontend-full.log)、[定向](../../../docs/upstream-sync/implementation-gpt62/frontend-targeted.log) |
| 冻结范围与 OpenSpec | 新change严格校验通过；最终链接、来源hash和保护范围见快照 | [OpenSpec](../../../docs/upstream-sync/implementation-gpt62/openspec.json)、[范围快照](../../../docs/upstream-sync/implementation-gpt62/scope-validation.json) |

全量unit和四包race完成后，lint发现原私有计费包装已无生产调用。移除包装，既有零倍率长上下文marker测试改用等价公共入口；随后完成服务定向race、lint和最终embed构建。日志与最终产品哈希分别保留，未将清理前全量日志改写为清理后执行记录。

价格使用冻结PR8906的Sol 2/10/2.5/0.10、Astra 8/40/10/0.40 USD/百万输入/输出/缓存写入/缓存读取token。官网在线核验因web502/官方pricing页面403未完成。实际本地HTTP/WS/SQLite使用合成账号、token及未来时刻，未进行现网OpenAI或业务实例操作。

共享工作区另外出现 `.github/workflows/release.yml` 与 `deploy/tests/release-publish-config-test.sh` 的并行修改。本次未编辑这两份文件，范围快照单独记录。
