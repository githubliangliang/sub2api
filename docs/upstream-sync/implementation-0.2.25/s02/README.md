# S02 Gorilla 定向安全维护实施证据

实施起点见上层 `initial.json`：`39ca66d51d53b98f11349598bbd16cd8ecd0e2e0`。本目录只保存本轮实际执行结果；不改写评估期证据。`run.py` 使用 `CI=true`、`TZ=UTC`、`GOTOOLCHAIN=go1.27.4`，JSON 中的 `exit_code` 是实际子命令结果，wrapper 自身正常退出不代表测试通过。

## 版本、来源与完整差异

- `fetch.json` 记录 2026-11-24 本轮重新请求官方 Go 漏洞数据库、Gorilla 官方修复提交和官方 Go module proxy 的 URL、时间、HTTP 状态与响应头。原文分别是 `official-vulnerability.json`、`official-fix.json`、`official-module.info`。网页工具调用无正文，随后直接读取这些官方接口成功。
- 官方 GO-2026-6278 的版本事件明确包含重新受影响的 `1.5.3`，固定修复版本为 `1.5.4-0.20240701034025-d67f41855da4`；提交是 `d67f41855da42d7bccd9ef050c49f7e54e783b95`。其 `conn.go` 将 mask 随机源切为 `crypto/rand.Reader`，并通过 `io.ReadFull` 读取四字节。这里不声称执行了漏洞利用。
- 产品只改 `backend/go.mod` 的 Gorilla 一行和 `go.sum` 对应两条校验和。`candidate-download.log` 记录模块校验和 `h1:PYKzliEgITjLJoJqbV90S0YRaG8LNAsICH6fp6MApC0=` 和 go.mod 校验和 `h1:r4w70xmWCQKmi1ONH4KIaBptdivuRPyosB9RmPlGEwA=`；`module-verify.log` 为实际 `go mod verify` 通过。
- `baseline-modules.log` / `candidate-modules.log` / `mvs.diff` / `validation.json` 核验 406 个模块的最终 MVS，仅 Gorilla 版本不同。新模块要求 Go1.20 和 x/net0.26.0，本地现有 Go1.27.4 / x/net0.62.0 足够，无其它依赖变化。
- `full-module-diff.patch` / `full-module-diff.json` 是实际下载内容完整 18 文件差异及各文件旧/新 SHA256，已逐项复核，不能把完整升级描述成仅随机源变化。

| 差异组 | 复核结论及控制 |
|---|---|
| conn.go | mask 改 crypto/rand，另有 ioutil→io 等价替换；prepared 测试改局部可控随机源。 |
| client.go、proxy.go、删除 x_net_proxy.go | DialContext 贯通，旧内嵌 proxy 包改用 x/net/proxy，CONNECT 响应关闭前重置 reader；实际无鉴权/Basic-auth 隧道、库 SOCKS/握手测试和应用互通覆盖。 |
| server.go | subprotocol 从 server 优先改 client 优先；管理员只声明 sub2api-admin，实际带 jwt.<token> 的握手仍只回这一项。缺 Upgrade 的响应从400变426；使用 ResponseController.Hijack 支持包装 writer。 |
| go.mod、go.sum | 模块自身 Go1.20/x/net0.26.0/retract 元数据；本地最终图无额外变化。 |
| 六份库测试文件 | 退出等待、握手/subprotocol/Hijack 新测试、局部随机源与标准库等价替换；原始失败和修正夹具控制分开保存。 |
| CI、README、三份 example 文件 | CI Go 范围、链接分支名、ioutil 等价替换；没有随带 fork 业务改动。 |

生产静态直接 import Gorilla 的位置仅 `internal/handler/admin/ops_ws_handler.go`，其余引用为测试；网关生产 WS 使用既有其它实现。没有执行全仓 callgraph/govulncheck，也不把包版本命中写成确认生产利用。

## 既有库测试失败的定因

| 实际检查 | 结果 |
|---|---|
| `baseline-library-project.json` | 原产品模块环境：55 顶层通过 / 3 失败；含子例57通过。失败 ProxyDial、ProxyAuthorizationDial、PreparedMessage。 |
| `candidate-library-project.json` | 实际更新后产品模块：59 顶层通过 / 2 失败；含子例61通过。仅同两项历史代理 EOF，故不宣称依赖包全量绿灯。 |
| `baseline-library-trace.json` / `candidate-library-trace.json` | 临时模块副本加入纯观测 net.Conn 读写记录；两版本均在 CONNECT200 之后发送 GET，再读到 EOF。 |
| `baseline-fixture-tunnel.json` / `candidate-fixture-tunnel.json` | 仅临时副本将两处假 CONNECT handler 改为 Hijack+真正 TCP 隧道，两版本各2/2通过。可重放差异见两个 `*-fixture-tunnel.patch`，未改 module cache 或产品依赖源码。 |
| `baseline-prepared-compat.json` | 仅旧模块控制 `GODEBUG=randseednop=0` 后 PreparedMessage 通过，确认旧测试依赖 rand.Seed 全局兼容行为；没有把此环境变量加入产品。 |

两项 EOF 的确切原因是 Go1.27.4 `net/http/server.go:1433` 对 CONNECT 一律设置 `closeAfterReply`，`:1576` 说明成功 CONNECT 即使没有 `Connection: close` 也会在 handler 不 Hijack 时关闭。库的两个旧夹具只 `WriteHeader(200)` 后返回，期望同一连接继续作为普通 HTTP 服务接收 GET；当前标准库不再允许这一伪隧道。`go1.27.4-connect-lifecycle.txt` 保存实际工具链源码摘录、路径与 SHA256，wire trace 与仅改夹具的双版本通过共同确认原因。评估期增加 Content-Length 失败的假设保持原样；本轮没有再次用它掩盖失败。

`baseline-library.json` / `candidate-library.json` 另外在模块自身 go.mod 下执行全量库测试，均仍有两项 EOF；旧模块的 go1.12 兼容设置使 PreparedMessage 在该上下文通过，和产品 go1.27.4 上下文的原始三失败分列。临时副本目录最初继承只读权限导致创建 trace 文件失败，记录在 `instrumentation-initial-failure.json`；只修正临时目录权限后继续。

## 实际应用控制与测试夹具修复

- `admin-connect-final.json`：新依赖下两个实际 TCP 测试，2 顶层 / 14 含子例通过。`baseline-admin-connect.json` 使用临时旧版本 modfile运行完全相同测试，同样2/14通过；没有为比较切回产品 go.mod。
- `ops_ws_security_test.go` 经实际 Gin 管理员认证中间件、QPS handler、Gorilla upgrader 和 TCP 握手检查：缺失/无效/已撤销 token、非管理员/停用账号、跨域和 strict 缺 Origin、单一 subprotocol 且不回传 JWT、全局连接限额、每 IP 限额失败后的总槽回滚、ping/pong、正常 close、超大帧1009、全局/IP槽释放。只有账号仓库为合成 stub，无 DB/Redis 外部实例。
- `gorilla_security_test.go` 在真正的 TCP CONNECT 隧道上执行无鉴权和 Basic-auth 的 WebSocket 回显，并等待双向转发 goroutine 和隧道释放。
- 初次五包广义 WS/Passthrough race 见 `websocket-race.json`：691顶层通过、2失败，无 race detector 报告。一项为并行 S01 实施中尚未修正的原 body 规则优先级测试，另一项为既有 Fast WS 测试的设置 stub 在冷缓存下 `GetMultiple` panic；两种原始输出都保存。
- `baseline-ws-fixture.json` 在旧 Gorilla 临时 modfile下单独执行原 Fast WS race，同样失败；根因是 `openAIFastPolicyRepoStub.GetMultiple` 原实现直接 panic，而真实 WS 转发会在共享设置缓存冷启动时读取它。只把该**测试 stub**改为按其已有 values map返回请求键，未改产品逻辑；`baseline-ws-fixture-fixed.json` 保存旧依赖下修正后1/1通过。此修正使控制不再依赖其它测试预热进程缓存。
- 最终五包 WS race 见 `websocket-race-final.json`：693顶层 /1351含子例通过，五包全部通过，无 race detector 报告，149.617秒。上层组合验收另包含全量 unit、embed、lint、SQLite/账本/固定配额控制。lint发现的新测试 unchecked Close/type assertion/HTTP数字状态已修正，最终文件哈希见 `validation.json` 和上层最终清单。

VERSION1.1.27、Go1.27.4、SQLite230/auth21、其它依赖、前端、Ent/Wire 与64MiB/40MiB/8MiB/120s默认值均不由 S02 改动。无 commit/push/release/archive 或业务实例操作。
