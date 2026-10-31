# 用户选定兼容性与Grok诊断实施证据

2026-10-31 完成用户要求的工具参数/别名/defer、Responses事件序号和Grok模型目录诊断。实施起点`5250907f5b8294eabed3d3101d2c4bf79936e0f7`，保留之前的模型参考价格提交。版本1.1.21、Go1.27.3、最大SQLite迁移227。

当前33个产品/测试文件汇总SHA256为`d472f1d75f0aece88b9930e94cfb52ab232f2d716b521fd2268b60a3e5b72852`。[起点和保护哈希](./baseline.json)、[逐文件范围与哈希](./scope-facts.json)、[门禁索引](./gate-results.json)。实现提交在后续完成记录中填写。

| 行为 | 实际证据 |
|---|---|
| C01普通参数/坏历史/保留合法length与custom | [正式失败回归](./formal-before.jsonl)、[修复后定向](./c01-c03.jsonl)、[边界](./boundaries-initial.jsonl) |
| C02唯一custom别名与显式/namespace优先 | [同组回归](./boundaries-initial.jsonl)、[实际HTTP custom调用](./strict-http-sqlite-initial.jsonl) |
| C03孤儿defer及声明源所有权 | [服务/adapter失败](./service-before.log)、[修复后](./c01-c03.jsonl)、[合法内建search与源map](./boundaries-initial.jsonl) |
| C04首帧0/本地事件/真实序号续接 | [旧事件失败](./handler-before.log)、[修复后](./c01-c04-corrected.jsonl)、[实际native/passthrough与整数边界](./sequence-native-boundaries.jsonl)、[严格HTTP连接](./strict-http-sqlite-initial.jsonl) |
| 无效工具后的异步SQLite用量/取消/去重 | [真实HTTP、队列和SQLite](./strict-http-sqlite-initial.jsonl)：1000输入/10输出、余额/Key quota、同请求只结算一次 |
| D01只读状态/TTL/保守映射/缓存不写 | [诊断语义](./grok-diagnostics-initial.jsonl)、[并发读取缓存不写](./boundaries-initial.jsonl)、[race](./backend-race.jsonl) |
| D02详情/同步结果、无网络/写库、最新SQLite | [管理端HTTP/SQLite](./grok-admin-sqlite-initial.jsonl)：写保护trigger、上游请求计数、普通mapper/非Grok不扩展、同步后新目录 |
| D03已保存编辑、时间、关闭/账号切换/乱序 | [面板和编辑器](./frontend-diagnostics-lifecycle-corrected.log)、[真实浏览器命令](./browser/commands.json)及[截图](./browser/diagnostics-mobile.png) |

诊断计算使用现有纯映射解析，避免调用会填写账号缓存的GetModelMapping；并未引入共享快照/COW系统。事件观察只保留整数，不另建SSE payload缓冲。变更仅覆盖既有消费者和所选诊断。

## 完整检查

- 后端54个有测试包完整unit、CGO=0/embed构建、golangci-lint 0 issues：[unit](./backend-unit.log)、[build](./backend-build.log)、[lint](./backend-lint.log)。完整unit沿用当前环境直接通过，没有修改验证码或代理代码。
- 四包456个含子用例定向race通过：[race](./backend-race.jsonl)。WS/BareError/compact相邻路径另行核对：[入口](./ws-test-list.log)、[WS race](./ws-race.jsonl)。
- SQLite迁移/方言、配额、成功用量去重与固定窗口保护468个用例通过：[保护](./sqlite-protection.jsonl)。测试数据库均为临时独立副本，不操作业务实例。
- 前端290个测试文件，2169 passed/2项既有skipped，typecheck/lint/build通过；critical176 passed/2项既有skipped：[完整](./frontend-full.log)、[构建](./frontend-build.log)、[关键](./frontend-critical.log)。
- 两份OpenSpec严格校验通过：[兼容](./openspec-compatibility.log)、[Grok](./openspec-grok.log)。原有2294个冻结历史文件和Ent/Wire、迁移、依赖/VERSION保护范围不变。

## 浏览器范围

使用本轮源码Vite与真实Chrome连接独立合成API：[夹具](./browser-fixture.py)、[请求](./browser/requests.json)、[命令](./browser/commands.json)。桌面看到目录与差异；同步不覆盖未保存模型输入；更新后重开保留新映射且dirty提示消失；失败保留最近目录并显示错误；390px视口与页面宽度均390、面板340，无横向溢出。已检查浏览器无未捕获异常。

[初始桌面](./browser/diagnostics-desktop.png)、[未保存与稳定同步](./browser/diagnostics-edited-stable.png)、[失败](./browser/diagnostics-sync-failure.png)、[移动](./browser/diagnostics-mobile.png)。本轮没有连接真实Codex/Grok Build进程或业务账号；严格事件消费通过实际HTTP连接验证，管理端数据库行为另由真实SQLite测试验证。

## 过程失败与修正

- 原工具/事件探针与正式基线按预期失败，保留[转换](./apicompat-before.log)、[服务](./service-before.log)、[handler](./handler-before.log)，未冒充修复完成。
- C04旧handler测试明确断言字段不存在；按用户选定的新合同更新为显式0，其它断言保留。
- 本地compact尝试曾在未写出payload时消耗序号，新的[失败回归](./compact-unwritten-sequence-before.log)定位后改为仅查看待用序号。
- 同步失败后读取详情成功曾清除失败提示，[失败用例](./frontend-sync-failure-before.log)定位后保留同步错误。[最终用例](./frontend-diagnostics-lifecycle-corrected.log)通过。
- 新测试使用了Vitest2没有的matcher，改为现有times/with断言；原[日志](./frontend-diagnostics-lifecycle.log)保留，不反推产品错误。
- 原Grok孤儿工具控制函数不在fork中，采用当前sanitizer和tool-choice消费者；一次patch预检查失败没有修改产品文件。来源converter只取选定文件，[来源patch](./c02-source-converter.patch)和源诊断文件快照保留。
- 提交前整体whitespace检查命中原始patch上下文、日志和CRLF TSV；源码/Markdown单独检查通过，原始证据保留字节，不为清理告警改写冻结记录。

后端命令在backend目录按go.mod选择实际Go1.27.3：[工具链](./backend-toolchain.log)。外层go启动器为1.26.5，未用于宣称本轮编译版本；pnpm11.21.0运行现有安装，不改变锁文件。

源/proposal/design/spec SHA冻结在[来源保护](./frozen-source-sha256.json)，原任务与verification另存original快照。当前行为和完成SHA只写实际verification，不回写历史评估或来源。
