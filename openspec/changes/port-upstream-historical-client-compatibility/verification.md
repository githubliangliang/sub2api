# 实施验收（完成）

2026-10-31 用户授权后完成C01–C04，17/17 tasks。实现已提交本地main，保留活动目录待归档；旧评估来源及失败记录不回写。

| 实施信息 | 值 |
|---|---|
| 实际起点 | `5250907f5b8294eabed3d3101d2c4bf79936e0f7`，保留模型参考价格提交 |
| 实现提交 / 产品哈希 | `becd2d2a31192e13bdc431bc43fbe44436e2c319` / `d472f1d75f0aece88b9930e94cfb52ab232f2d716b521fd2268b60a3e5b72852`；[范围](../../../docs/upstream-sync/implementation-historical-client-compatibility/scope-facts.json)、[提交](../../../docs/upstream-sync/implementation-historical-client-compatibility/implementation-commit.json) |
| 实施时VERSION / Go / 最大迁移 | 1.1.21 / 实际模块工具链1.27.3 / SQLite227 |
| 实际范围 | 33个产品/测试文件，四项兼容性尾项及用户选定Grok诊断；依赖、Ent/Wire、迁移不变 |

## 行为与反例

| 编号 | 实际行为和反例 | 证据 | 结果 |
|---|---|---|---|
| C01 | 非法普通参数不能完成，坏历史及匹配输出过滤；合法JSON/空值/length/custom保留；真实异步usage/dedup | [失败](../../../docs/upstream-sync/implementation-historical-client-compatibility/formal-before.jsonl)、[边界](../../../docs/upstream-sync/implementation-historical-client-compatibility/boundaries-initial.jsonl)、[HTTP/SQLite](../../../docs/upstream-sync/implementation-historical-client-compatibility/strict-http-sqlite-initial.jsonl) | PASS |
| C02 | 唯一custom别名、显式function和真实namespace优先、HTTP/SSE身份一致 | [回归](../../../docs/upstream-sync/implementation-historical-client-compatibility/boundaries-initial.jsonl)、[真实HTTP custom别名](../../../docs/upstream-sync/implementation-historical-client-compatibility/strict-http-sqlite-initial.jsonl) | PASS |
| C03 | 通用/Grok最终tools清理孤儿defer，同名function不充当前置，保留合法内建search和声明源 | [回归](../../../docs/upstream-sync/implementation-historical-client-compatibility/c01-c03.jsonl)、[所有权反例](../../../docs/upstream-sync/implementation-historical-client-compatibility/boundaries-initial.jsonl) | PASS |
| C04 | 第一帧0、已知非零不变、compact/handler/WS事件，真实写出序号续接，整数/延迟错误/未写出反例 | [typed/合成](../../../docs/upstream-sync/implementation-historical-client-compatibility/c01-c04-corrected.jsonl)、[native/passthrough](../../../docs/upstream-sync/implementation-historical-client-compatibility/sequence-native-boundaries.jsonl)、[严格HTTP连接](../../../docs/upstream-sync/implementation-historical-client-compatibility/strict-http-sqlite-initial.jsonl) | PASS |

## 组合检查

| 检查 | 证据 | 结果 |
|---|---|---|
| backend完整unit/build/lint | [54包unit](../../../docs/upstream-sync/implementation-historical-client-compatibility/backend-unit.log)、[CGO=0/embed](../../../docs/upstream-sync/implementation-historical-client-compatibility/backend-build.log)、[0 issues](../../../docs/upstream-sync/implementation-historical-client-compatibility/backend-lint.log) | PASS |
| 定向race与WS相邻保护 | [4包456](../../../docs/upstream-sync/implementation-historical-client-compatibility/backend-race.jsonl)、[2包WS/BareError/compact74](../../../docs/upstream-sync/implementation-historical-client-compatibility/ws-race.jsonl) | PASS |
| 真实SQLite迁移227/usage/dedup/固定窗口 | [468保护](../../../docs/upstream-sync/implementation-historical-client-compatibility/sqlite-protection.jsonl)、[异步结算/取消/只扣一次](../../../docs/upstream-sync/implementation-historical-client-compatibility/strict-http-sqlite-initial.jsonl) | PASS |
| frontend完整门禁 | 2169 passed/2既有skipped；typecheck/lint/build通过，[完整](../../../docs/upstream-sync/implementation-historical-client-compatibility/frontend-full.log)、[关键176/2](../../../docs/upstream-sync/implementation-historical-client-compatibility/frontend-critical.log) | PASS |
| 实际浏览器 | [命令](../../../docs/upstream-sync/implementation-historical-client-compatibility/browser/commands.json)、[移动](../../../docs/upstream-sync/implementation-historical-client-compatibility/browser/diagnostics-mobile.png)；当前源码Vite/Chrome/独立合成API | PASS |
| 源码/依赖/生成文件范围与冻结历史 | [起点2294个历史哈希](../../../docs/upstream-sync/implementation-historical-client-compatibility/baseline.json)、[保护范围](../../../docs/upstream-sync/implementation-historical-client-compatibility/scope-facts.json)、[源SHA冻结](../../../docs/upstream-sync/implementation-historical-client-compatibility/frozen-source-sha256.json)；无依赖更改，本轮无新audit | PASS / audit N/A |
| OpenSpec严格校验 | [兼容](../../../docs/upstream-sync/implementation-historical-client-compatibility/openspec-compatibility.log) | PASS |

严格客户端要求通过实际HTTP连接及必填整数/递增消费者验证；未连接真实Codex/Grok Build业务进程。过程失败、修正及浏览器/数据边界见[实际证据索引](../../../docs/upstream-sync/implementation-historical-client-compatibility/README.md)。评估probe未当作修复完成。

## 完成条件

- [x] 每项规格有修复前后证据与正反例。
- [x] 所有适用门禁通过，N/A与范围已说明。
- [x] 实际实现提交/起点已记录，评估与实施分开保存。
