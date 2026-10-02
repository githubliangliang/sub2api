# v0.2.12实施验收

日期：2026-10-02。状态：本批3簇开发及验收完成。

| 项目 | 证据 |
|---|---|
| 实施起点 | `5eb58199ee6ef45515ec791b445305651384fc0c`，产品工作树干净；既有本轮评估文档未提交 |
| 实现提交 | `4052467891779c6f957324053c23814a359cfd25`，`sync/upstream-0.2.12` |
| 实施授权 | 用户“完成第一档和第二档”，包含已确认需要的两个管理操作 |
| 后端全量unit/build | `go test -tags=unit ./...`、`go build ./...`退出0；[unit](../port-upstream-0.2.12-tier2/evidence/backend-unit-final.txt)、[build](../port-upstream-0.2.12-tier2/evidence/backend-build.txt) |
| 定向race | repository/service涉及邮箱并发、SMTP、改密、Grok、脱敏和排序的选中测试通过；[race](../port-upstream-0.2.12-tier2/evidence/backend-race.txt) |
| SQLite | 方言、真实用量效果与去重、分组名跨页排序通过；[SQLite](../port-upstream-0.2.12-tier2/evidence/sqlite.txt) |
| 前端 | typecheck/lint/build通过；253测试文件，1854通过/2既有跳过；[全量Vitest](../port-upstream-0.2.12-tier2/evidence/frontend-vitest.txt) |
| 版本与范围 | VERSION 1.1.15、最大迁移226、依赖锁文件不变；支付/TypeSafe/axios未引入；未发布或部署 |

未使用postgres-only测试充数。内置miniredis验证真实Lua/TTL；未连接外部Redis、真实上游账号或线上SMTP。SMTP验收使用本地协议服务，出口测试使用受控HTTP响应。基线探针与评估证据保持冻结，不替代本页验收。

## 逐场景证据

| 簇/场景 | 测试或核对 | 结果 |
|---|---|---|
| F01 默认请求/旧覆盖 | TestApplyCLIProxyHeadersMeetsUpstreamMinimumVersion、TestHTTPUpstreamDoAppliesGrokCLIIdentityBeforeOAuthRoundTrip、TestApplyGrokCLIProxyHeaders | 通过，固定1.0.46身份，低于各层下限值回退 |
| F01 有效覆盖/平台UA | TestCLIUserAgentMatchesOfficialInteractiveCapture、TestResolveCLIVersionAcceptsValidOverride、TestApplyDefaultGrokUpstreamHeadersHonorsCLIVersionOverride | 通过；平台转换实现按GOOS/GOARCH分支，当前运行linux/amd64 |
| F01 官方API边界 | TestApplyCLIProxyHeadersLeavesAPIHostUnchanged、TestHTTPUpstreamDoFallsBackToOfficialGrokAPIOnCLIAccessDenied | 通过，回退清除新增CLI头 |
| F02 JSON/非JSON/空正文 | TestAntigravityGatewayService_ForwardGemini_SanitizesClientError、TestBuildAntigravityClientErrorBody_ScrubsPoolIdentity、TestBuildAntigravityClientErrorBody_NonJSONBody | 通过；真实出口HTTP状态保持，三字段JSON，不含已知项目/邮箱/敏感查询参数/details |
| F02 正常/failover | 现有ForwardGemini成功、配置fallback、ModelRateLimitTriggersFailover等用例 | 定向及全量unit通过，未改用量与failover策略 |
| F03 中英文说明/配置行为 | 两语言customErrorCodesWarning人工核对Account.ShouldHandleErrorCode；只有locale字符串差异 | 空列表/账号错误处理/独立retry与透传语义一致；未改后端配置逻辑；typecheck/lint/全量测试通过 |

## 失败到通过

- [F01与S03初始失败](./evidence/f01-s03-red.txt)：旧CLI版本/identity和SQLite分组排序断言失败。
- [F02真实出口初始失败](../port-upstream-0.2.12-tier2/evidence/f02-s01-red.txt)：错误正文仍为text/plain原样输出；实施后[定向检查](../port-upstream-0.2.12-tier2/evidence/targeted-initial.txt)与最终全量/race通过。
- F03是人工说明变更，没有增加镜像文案的测试。
