# 实施验收（完成）

状态：2026-10-29 完成用户授权的第一档 F01–F07，26/26 tasks。实现已提交本地 main，change 保持活动状态等待后续归档。评估 probe 的 PASS 没有用作实施完成证据；各项均有预期新行为的失败回归和修复后反例。

| 实施信息 | 值 |
|---|---|
| 实施起点 HEAD / 工作树 | `49939fb2cf96633688548ca39204d96558e59d79`；3732 产品文件与冻结评估一致。预存未提交 AGENTS/README、评估与三份提案，见 [baseline](../../../docs/upstream-sync/implementation-0.2.19-tier1/baseline.json) |
| 实现提交 / 代码 SHA256 | `683354b2cdd2461dc581fa14158be09ef51de29b`；31 文件汇总 SHA256 `3f97da9098f59ff907f2d5b057fcaa200ab5eca13ca93a6e95d5f1f03e3cf9fd`；[提交记录](../../../docs/upstream-sync/implementation-0.2.19-tier1/implementation-commit.json) / [全部文件哈希](../../../docs/upstream-sync/implementation-0.2.19-tier1/scope-facts.json) |
| 实施时 VERSION / Go / 迁移 | 1.1.20 / 1.27.3 / SQLite 227；[工具实际版本](../../../docs/upstream-sync/implementation-0.2.19-tier1/implementation-tools.txt) |
| 实际范围与排除项 | 14 个生产文件、17 个测试文件；仅 F01–F07。第二档、Terra、其它新能力、支付、版本/依赖/schema/生成代码变化、发布/tag/push/部署及业务实例数据操作均未进入；[范围](../../../docs/upstream-sync/implementation-0.2.19-tier1/scope-facts.json) |

## 行为、反例与结果

[完整行为映射、测试入口和失败归因](../../../docs/upstream-sync/implementation-0.2.19-tier1/acceptance-matrix.md)逐项连接实际消费者。

| 编号 | 必须验收的行为与反例 | 修复前证据 | 修复后证据 | 结果 |
|---|---|---|---|---|
| F01 | ASCII、中文、emoji、采样 surrogate 重组与孤立、NUL、短文本/长尾、版本同步/重复调用；正文、header 和模板一致 | [失败](../../../docs/upstream-sync/implementation-0.2.19-tier1/baseline-claude-codex.txt) | [回归](../../../docs/upstream-sync/implementation-0.2.19-tier1/claude-codex.txt)、[独立 Node oracle](../../../docs/upstream-sync/implementation-0.2.19-tier1/claude-js-oracle.txt)；最多保存22个码元 | PASS |
| F02 | 确定性交错后续 fetch=0，普通 miss/stale、ETag/304、账号与请求身份隔离、缓存复制、并发/取消 | [1→2 fetch](../../../docs/upstream-sync/implementation-0.2.19-tier1/baseline-codex-flight.txt) | [回归](../../../docs/upstream-sync/implementation-0.2.19-tier1/claude-codex.txt)、[race](../../../docs/upstream-sync/implementation-0.2.19-tier1/backend-race.txt) | PASS |
| F03 | 嵌套显式0/正值优先，prompt100/hit80/miss20；HTTP/SSE 同桶，SQLite 用量/余额/Key quota/dedup 只结算一次 | [解析失败](../../../docs/upstream-sync/implementation-0.2.19-tier1/baseline-cache-images.txt)、[旧 reader 的真实账本失败](../../../docs/upstream-sync/implementation-0.2.19-tier1/baseline-cache-hit-sqlite.txt) | [解析](../../../docs/upstream-sync/implementation-0.2.19-tier1/cache-images.txt)、[完整 SQLite HTTP/SSE 链](../../../docs/upstream-sync/implementation-0.2.19-tier1/cache-hit-sqlite-complete-fixtures.txt) | PASS |
| F04 | inline/file camel/snake、图片 MIME、非空字符串、同 part 至多一次、累计 SSE 最大值、自定义别名、旧 fallback；取消只保留实际图片 | [失败](../../../docs/upstream-sync/implementation-0.2.19-tier1/baseline-cache-images.txt) | [原生输出/成本/三协议断流](../../../docs/upstream-sync/implementation-0.2.19-tier1/cache-images.txt)、[race](../../../docs/upstream-sync/implementation-0.2.19-tier1/backend-race.txt) | PASS |
| F05 | JSON/wrapped/SSE 无 parts 畸形工具尾帧不产生空成功；正常文字/工具、取消、错误及失败用量；活流关闭 | [失败](../../../docs/upstream-sync/implementation-0.2.19-tier1/baseline-malformed-tools.txt) | [三出口及邻近回归](../../../docs/upstream-sync/implementation-0.2.19-tier1/malformed-tools-corrected.txt)、[实际 Forward 及时返回/关闭 body](../../../docs/upstream-sync/implementation-0.2.19-tier1/pagination-admin-and-resource.txt) | PASS |
| F06 | 超长数字、MaxInt/负数/空值/非法、最大合法页、直接 params、Key/仓储切片与 HTTP 入口；默认/限幅/严格400策略保持 | [有效夹具失败](../../../docs/upstream-sync/implementation-0.2.19-tier1/baseline-pagination-valid-fixtures.txt) | [六包边界与 HTTP](../../../docs/upstream-sync/implementation-0.2.19-tier1/pagination-admin-and-resource.txt)、[race](../../../docs/upstream-sync/implementation-0.2.19-tier1/backend-race.txt) | PASS |
| F07 | 真实 handler400/三个稳定 reason；六种禁止操作不写库；其它管理员存在时允许降级，基础设施故障仍500 | [有效 SQLite 夹具失败](../../../docs/upstream-sync/implementation-0.2.19-tier1/baseline-admin-guards-final-fixtures.txt) | [实际 handler/SQLite 正反例](../../../docs/upstream-sync/implementation-0.2.19-tier1/pagination-admin-and-resource.txt)、[race](../../../docs/upstream-sync/implementation-0.2.19-tier1/backend-race.txt) | PASS |

## 组合检查

| 检查 | 命令 / 证据 | 结果 |
|---|---|---|
| backend 完整 unit / build / lint | `go test -tags=unit ./... -count=1`（最终仅为该测试进程移除代理变量）；[54包](../../../docs/upstream-sync/implementation-0.2.19-tier1/backend-unit-direct.txt) / [完整 argv](../../../docs/upstream-sync/implementation-0.2.19-tier1/backend-unit-direct.json)。[build](../../../docs/upstream-sync/implementation-0.2.19-tier1/backend-build.json)、[golangci-lint 2.14.0](../../../docs/upstream-sync/implementation-0.2.19-tier1/backend-lint-final.txt) | PASS：54包、build、0 issues |
| 受影响路径 race 及取消/资源释放 | `go test -race -tags=unit`，完整包与筛选见 [argv](../../../docs/upstream-sync/implementation-0.2.19-tier1/backend-race.json) / [输出](../../../docs/upstream-sync/implementation-0.2.19-tier1/backend-race.txt) | PASS：8包实际测试，无空筛选 |
| 真实 SQLite 迁移227 / quota / usage dedup | [命令](../../../docs/upstream-sync/implementation-0.2.19-tier1/sqlite-protection.json) / [输出](../../../docs/upstream-sync/implementation-0.2.19-tier1/sqlite-protection.txt)：完整迁移、拆句/方言、配置限额/0值/软删不复活、legacy TEXT、固定窗口与成功结算 | PASS |
| 前端适用 typecheck / lint / tests / build / 浏览器 | [1416 个保护产品文件哈希](../../../docs/upstream-sync/implementation-0.2.19-tier1/scope-facts.json)：前端源码、package/lockfile 未变，本批没有前端实施 | N/A：无前端变更 |
| 依赖适用 audit / Go扫描 / 未解决 finding | go.mod/go.sum、前端依赖及工具链/生成文件未变；评估 [既有依赖项](../../../docs/upstream-sync/PORTING-0.2.19.md#54-依赖遗留的当前口径)保持原记录，本批未把其作为修复结果 | N/A：无依赖变更 |
| OpenSpec 严格校验 / 链接 / 范围 / 冻结来源哈希 | [严格校验](../../../docs/upstream-sync/implementation-0.2.19-tier1/openspec-strict-final.txt)、[完成与链接校验](../../../docs/upstream-sync/implementation-0.2.19-tier1/implementation-validation-final.txt)、[保护范围](../../../docs/upstream-sync/implementation-0.2.19-tier1/scope-facts.json) | PASS：150 个冻结文件及1416个保护产品文件未变 |

## 验收边界与过程失败

早期编译落点、nil-service、time.Time monotonic、数据库 constraint 分类和缺失价卡/配额更新器的夹具问题分别保留在 [失败归因](../../../docs/upstream-sync/implementation-0.2.19-tier1/acceptance-matrix.md)。它们没有冒充生产缺陷或最终通过。

全量 unit 曾命中既有 Aliyun 运输错误测试：安装 SDK 对 NO_PROXY 按含端口 domain 精确匹配，导致本地请求经过环境代理。相关生产/测试文件未改。[控制代理探针](../../../docs/upstream-sync/implementation-0.2.19-tier1/captcha-proxy-proof.txt)复现原因；[代理隔离后三轮12项](../../../docs/upstream-sync/implementation-0.2.19-tier1/captcha-direct-network.txt)及最终54包全部通过。仅调整测试子进程的代理环境，未修改产品配置或 SDK。

## 完成条件

- [x] 每项规格有实际前后证据及反例，基线 probe 未冒充修复完成。
- [x] 所有适用检查通过；前端/依赖 N/A 原因、保护范围与冻结 SHA 已核对。
- [x] 实现提交与实际起点已记录，没有未授权支付、发布或部署。
