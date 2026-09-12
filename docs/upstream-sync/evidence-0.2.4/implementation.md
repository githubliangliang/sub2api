# 0.2.4 两档最终验收

日期：2026-09-12。**第一档 F01-F13 与第二档 S01-S12 均已实现并通过最终检查。**
本地集成分支：`sync/upstream-024-integrated-RLC1zS`。验收时未推送、未发布；后续主线交付见第 6 节。

原始 main：`c9d9bebe87926791bfdd716f59abcba413fee20a`。
隔离工作快照：`cf6b42752f88ee4eefe51e1c1c59ca1608ac20f5`，包含用户原有 metadata 修复及评估文档。
最终产品代码基线：`3f3f192d9`；之后的 `d0667b4cf` 只校正已有路由测试，其余收尾为文档。
两份 OpenSpec 的 source-baseline/source-feature-map 与原始三份 intake TSV 均保持不变。

## 1. 最终检查

命令在相应 backend/ 或 frontend/ 执行；全部结果来自集成工作树。

| 检查 | 命令 / 结果 |
|---|---|
| 后端构建 | `GOMAXPROCS=4 go build -p 2 ./...`：PASS |
| 后端全量 unit | `GOMAXPROCS=4 go test -p 2 -tags=unit ./... -count=1`：**54 个包 PASS**，service 155.450s |
| SQLite 方言 | `go test -tags=unit ./internal/repository -run '^TestProductionSQLUsesSQLiteDialect$' -count=1`：PASS |
| 集成 race | service/handler/repository 的回放所有权、冷却 CAS、客户端取消、失败会话释放及 Redis 兼容相关定向测试：PASS |
| 前端生产构建 / 类型检查 | `pnpm run build`（`vue-tsc -b && vite build`）：PASS |
| 前端 lint | `pnpm run lint:check`：PASS |
| 前端全量测试 | `pnpm run test:run`：**240 文件，1720 PASS，2 个原有 skipped** |
| 静态单文件 | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags embed -ldflags='-s -w -X main.Commit=3f3f192d9' ...`：PASS；`file` 确认为 x86-64、statically linked、stripped |
| 格式 / 来源 | 代码与文档 diff 检查通过；Astra instructions 资源逐字保留 v0.2.4，包括上游已有尾随空格，单独校验为字节一致 |
| 边界 | VERSION 仍为 1.1.11，迁移最大号仍为 226；无 Ent/schema/Wire/迁移或前端 lockfile 改动，SQLite/miniredis 保留 |

最终 race 命令：

```bash
GOMAXPROCS=4 go test -p 2 -race -tags=unit \
  ./internal/service ./internal/handler ./internal/repository \
  -run 'TestOpenAIWSReplayConcurrent|TestRuntimeCooldown|TestForwardOpenAIWSV2_ClientCancellation|TestGatewayFailedAnthropicRequestReleasesSessionSlot|TestSessionLimitUnregister|TestRedisUpgrade' \
  -count=1
```

首轮全量前端发现 5 项 route-prefetch 失败。原始工作区单独运行同一文件也复现 5 项失败，
原因是仍假定已移除的 `/admin/dashboard` 存在。测试改为实际保留的 `/admin/accounts`，
**路由产品代码未改**。重跑全量得到上述 1720 PASS。
构建中的大 chunk、旧 Browserslist 数据、pnpm.overrides 位置警告属于既有工具链提示，本批未扩展依赖升级。

## 2. 运行验证

使用独立临时 SQLite 数据库与 miniredis，未访问生产数据库，未调用真实模型。

- 最终静态二进制 `/health` 返回 200；`lite=1` 列表省略 groups/account_groups，凭据被净化；详情接口正常。
- 浏览器从列表打开真实编辑弹窗，修改账号名并保存；API 复核模型映射、base URL、API Key 存在状态及 group_ids 均保留。
- 桌面 1280×720 菜单范围 `(1019,399)-(1227,551)`，手机 390×844 范围 `(39.8,538)-(247.8,690)`，均在视口内；手机页面宽度为 390，无横向页面溢出。截图已人工查看。
- 保留旧 Ent 自动生成的 A→B、B→A 且 B 未启用回退的数据：只改 A/B 名称成功，原字段及引用保持；新增 C→B 共享成功；启用 B 导致有效循环时返回 400，事务回滚。

本地验证地址：`http://127.0.0.1:5317`；API：`http://127.0.0.1:18081`。
运行目录、日志、截图、临时登录凭据与二进制位于 `/tmp/sub2api-port-024-RLC1zS/`。

## 3. 集成修正

1. `eabfa89c3`：工具 done 参数补齐优先使用明确 call_id，避免最终输出重排时串参数；冲突 ID 不按位置强行匹配。
2. `431704a14`：取消时外层 Forward 保留部分结果；提前返回的 WS 结果也携带已完成图片计数，避免用量丢失。
3. `f7768cb6a`：CLI override 测试按第一档重算后的 billing 指纹断言，预期值独立由 SHA256 推导。
4. S06 instructions 移到模型映射之后，同时保留 F13 的 `UsesOpenAICodexProtocol()` 守卫。
5. S08 在当前 SQLite 非唯一标量列上实现有向引用，保留本地 Ent schema；循环验证按实际启用的回退链处理旧版无效反向引用。
6. S07 拒绝整个文件为 null；S09 个人模板显式保留 system-log 7 天，更新热重载说明。

S03 的已批准取舍仍存在：DB 写失败或快照滞后时，账号级陈旧本地 block 可 fail-open；
模型级 block 保留，成功持久化的冷却仍受最终 DB 复查保护。该行为有专门回归，不能解读为完全 fail-closed。
go-redis 内部 nil-context 竞态未在本批复现；版本升级、客户端取消恢复和 miniredis 索引范围/批量语义已验证。

## 4. 落地索引

| 范围 | 实施分支原始提交 |
|---|---|
| F01-F13 | `80bc5baf0`、`25daf1281`、`fd0080650`、`0eb88e504`；第一档验收 `f55626f99` |
| S01 / S02 / S03 | `d905eb76a` / `b396cc8db` / `467e79bfc` |
| S04 / S05 / S06 | `37586e485` / `9d23a5fd9` / `8946f18e6` |
| S07 / S08 / S09 | `942a34cd6` / `1b707d594` / `1915ee622` |
| S10 / S11 / S12 | `37e3eccc0` / `d06848c0c` / `6620a7c08` |

第三、四档与旧 DOMPurify 独立待办未实施；各簇详细 red/green 证据与剔除项仍保留在 OpenSpec verification 中。

## 5. 后续修复：Responses 调用丢失 namespace

2026-09-12，针对 `Missing namespace for function_call 'spawn_agent'` 报错，在原始工作区追加修复。
本地复现的是 API Key HTTP 转发保留 `tools` 中的 namespace 声明，却删除历史调用的
`input[].namespace`；未获取该线上 trace 的请求体，不能据此断言线上请求必经此路径。

- API Key 请求含命名空间工具声明、动态工具发现或续聊上下文时，保留工具调用 namespace。
- 普通消息残留字段、compact、OAuth 摊平开关和无上述上下文的旧兼容请求继续按原规则处理。
- 新增转发回归覆盖普通/透传、previous_response_id、conversation 字符串及对象、动态工具、原始大整数、compact 和明确拒绝字段后的重试。
- 修复前六个转发子用例复现 namespace 丢失；修复后下列定向回归通过，service 0.149s、apicompat 0.012s（不含编译时间）。本次未重跑前述全量 unit / 前端测试。

```bash
GOMAXPROCS=4 go test -p 2 -tags=unit ./internal/service ./internal/pkg/apicompat \
  -run 'Namespace|IndexedNamespace|RejectedField|ContentItemKinds' -count=1
GOMAXPROCS=4 go build -p 2 ./...
```

后端全包构建、前端 `pnpm run build`、格式及差异检查通过；前端构建仍有上述既有提示。
静态产物由当前工作区与重新生成的前端资源构建，原有本地集成分支未改写：

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOMAXPROCS=4 go build -p 2 \
  -tags embed -ldflags='-s -w -X main.Commit=c9d9bebe8-dirty' \
  -o /tmp/sub2api-port-024-RLC1zS/sub2api-namespace-fixed ./cmd/server
```

`file` 确认为 x86-64 静态可执行文件，`-version` 正常退出并报告 1.1.11；
`go version -m` 确认 `embed`、`CGO_ENABLED=0`、Linux/amd64、原始 HEAD 与 dirty 标记。
SHA256：`863ca25faa3f201a4d13c7c945b91ff7b778110afe712158d1099b374a5bfa61`。
测试和前端构建日志分别保留在 `/tmp/sub2api-namespace-fix-tests.log`、
`/tmp/sub2api-namespace-fix-frontend-build.log`。本节验收时未推送、未发布，未更新正在运行的服务。

## 6. 主线交付

2026-09-12，用户授权提交并推送到 `origin/main`（`githubliangliang/sub2api`）。
主线交付包括第一档、第二档、原工作区的 Responses input metadata 修复及第 5 节 namespace 修复。
本地实施分支和集成分支保留；本次没有创建新 tag 或 Release，也未部署正在运行的服务。
