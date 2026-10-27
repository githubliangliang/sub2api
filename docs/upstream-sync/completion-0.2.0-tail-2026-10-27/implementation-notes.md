# 收尾实现与必要门禁修复

本次执行剩余 `9.6`、`9.14`、`10.11`，起点为本地 `main` / `2deacc20f`，VERSION 1.1.20、Go 1.27.3、最大迁移 226。[起点记录](./start-state.json)保留此前任务核查留下的文档修改，以及冻结文件哈希。

`@types/dompurify` 类型桩已从 package.json、锁文件和实际安装树移除。锁文件只删除该包相关的 11 行，DOMPurify 本体仍为 3.4.17，两处 overrides 同值。离线 `remove` 的参数试验及首次冻结安装缺少 pnpm 元数据的失败保留在日志中；最终正常联网 `install --frozen-lockfile` 通过供应链校验，没有调整 pnpm 策略或其它包版本。

## lint 工具兼容性

原本机 golangci-lint 2.13.0 内嵌 x/tools v0.49.0，不能导入 Go 1.27.3 的 V5 导出格式。按已发布版本逐项测量：

| 版本 | 内嵌 x/tools | 来源记录 | 判断 |
|---|---|---|---|
| 2.13.0 | v0.49.0 | [原审计](../../../openspec/changes/port-upstream-0.2.0-p0-tail-and-p1/completion-check-2026-10-27.json) | 原标准库导入错误可复现。 |
| 2.13.1 | v0.49.0 | [模块来源](./linter-source-2.13.1.log) | 同样缺少 V5 支持。 |
| 2.13.2 | v0.49.0 | [模块来源](./linter-source-2.13.2.log) | 同样缺少 V5 支持。 |
| 2.14.0 | v0.50.0 | [模块来源](./linter-source-2.14.0.log) | 首个满足此条件的版本；原报错包 lint 通过。 |

本地与 CI 固定到 2.14.0。本地使用 Go 1.27.3 从该版本源码安装，先验证 `internal/pkg/sysutil`，再原子替换原工具；旧二进制备份及新旧 SHA-256 见[工具记录](./linter-tool-state.json)。DEV_GUIDE 的安装命令显式指定 GOTOOLCHAIN，避免在较旧的默认 Go 下重新构建不兼容工具。产品 go.mod/go.sum 和 `.golangci.yml` 规则集未修改。

## 完整门禁暴露的代码问题

首轮前端 2089 条用例虽全部通过，Vitest 仍报告 AccountsView 测试结束后的两条异步异常。原因是共享 `useTableLoader` 的排队防抖请求在 owner 卸载后重新创建 AbortController 并发起请求。新增真实 Vue/VueUse 生命周期回归，原代码在卸载后仍调用 fetch，确定性失败；最小修复为卸载标记与 load 入口保护。原组件、共有 loader 与新回归共 24 条定向用例通过，后续全量没有未处理异常。

兼容工具的完整 lint 首轮报告 28 处诊断，后续遍历还发现嵌套类型断言。处理范围：

- 测试对象/数组断言显式检查类型；关闭连接、管道和事务的清理行为明确处理返回值。原 SQLite 事务提交、取消、断开和测试断言保留。
- 分配复杂度守卫的子进程由 `os.Executable()` 定位当前测试程序，保留原阈值和隔离测量方式。
- OpenAI H2 健康探测直接使用 `http.Protocols` 和 `http.HTTP2Config`；继续启用 HTTP/1.1、HTTP/2，保持 TLS 配置初始化及两项 15 秒 PING 超时。现有真实 TLS 协商、代理和非 H2 模式用例验证行为。
- 独立 H2C 空闲超时依赖 x/net 的配置适配器，标准库 HTTP2Config 没有独立于 HTTP/1 的空闲超时字段，因此保留原行为并对该调用定点说明 SA1019。旧 x/net GOAWAY 类型仍需要匹配，net/http 没有公开替代类型；生产匹配与对应回归同样定点说明。没有全局关闭或移除 linter。
- 删除无消费者的 Anthropic `next()` 包装，保留实际 `nextOrClientGone` 消费者；简化等价布尔返回及无效赋值。

一次 schema 测试辅助函数补丁误匹配了同文件另一段相似测试，导致编译失败；已移入正确测试并保留失败日志与修正后的定向/全量证据。最终验收只引用最后通过的门禁，不把任何失败执行改写为成功。

上述 HTTP/2 适配行为依据当前 x/net v0.59.0 的 `transport_wrap.go` / `server_wrap.go` 和 Go 1.27.3 的 HTTP2Config API 测量；版本、依赖和迁移保持本次起点值。浏览器检查见[四处验收](./browser/README.md)，完整命令、结果和产品文件哈希由收尾记录汇总。

源码与可维护文档的 `git diff --check` 通过。原始命令输出带有 Vite/Vue/Vitest 自带的行尾空白和末尾空行，因此整份含日志差异的 whitespace 检查会报告这些日志；原始字节和 SHA-256 保留，检查范围及结果分别记于 `closure-check.json`。
