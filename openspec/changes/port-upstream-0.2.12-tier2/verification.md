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
| S01 普通验证码并发 | TestEmailCache_ConcurrentWrongCodesCannotExceedAttemptCap | 50并发经真实miniredis/EmailService，只有前4个错误尝试返回invalid，其余到达上限；正确码随后也拒绝 |
| S01 通知邮箱并发 | TestVerifyNotifyCode_ConcurrentGuessesRespectAttemptLimit、TestEmailCache_NotifyAttemptsExpireAndReset | 通知消费路径20并发为4 invalid/16 max；实际cache并发计数不丢失 |
| S01 TTL/重发/缓存错误 | TestEmailCache_AttemptsResetOnNewCodeAndTTLFollowsCode、TestEmailCache_NotifyAttemptsExpireAndReset、TestEmailCache_ExpiredOrBrokenCacheCannotValidate | TTL绑定、删除清理、重发归零、过期与异常拒绝通过 |
| S01 旧验证码兼容 | TestEmailCache_LegacyAttemptsAreNotResetAfterUpgrade | 旧JSON已有4次时，第5次触达上限，不能从0开始 |
| S01 哈希/新旧链接 | TestSendPasswordResetEmail_HashesActualLinkAndInvalidatesPreviousLink、TestConsumePasswordResetToken_ComparesHashNotPlaintext | 本地SMTP真实邮件token与缓存SHA-256对应；重发旧链接及旧部署明文token拒绝 |
| S01 并发单次消费/错误token | TestEmailCache_PasswordResetTokenHashedAndSingleUse、TestEmailCache_ConsumePasswordResetTokenMismatchKeepsToken | 30并发仅一个成功，错误token不删有效token，消费后不可重用 |
| S01 真实改密消费者/接口 | TestAuthService_ResetPasswordConsumesTokenBeforeUpdatingPassword及全量unit/build | bcrypt密码更新一次，第二次消费失败；全部EmailCache实现/mock编译和既有注册/绑定测试通过 |
| S02 连续点击/边界 | AccountPriorityCell.spec.ts：batches rapid、waits 450ms、does not go below 1、bounds typed values | 通过；只发priority字段、450ms合并、上下限与在途禁用 |
| S02 输入/取消/键盘 | 同文件typed value、invalid input参数用例、Escape、keyboard adjustment | 通过；Enter/blur去重、非法输入拒绝、Esc恢复 |
| S02 失败/在途 | reverts and emits an error、waits 450ms | 通过，失败回滚并提示，在途不能重复提交 |
| S02 卸载/行复用 | flushes pending edit、saves pending changes to original account、ignores old response | 通过；保存原账号且仅一次，迟到响应不覆盖新行；原生button/input及aria-label、focus/触屏CSS保留，未另做浏览器实测 |
| S03 SQLite名称/分页/NULL | TestAPIKeyRepositoryListByUserIDSortByGroup ascending/descending | 通过；名称顺序与ID创建顺序相反，跨页仍正确，未分组最后、同组ID次排序 |
| S03 过滤/用户隔离 | 同用例group_filter/ungrouped_filter及搜索、状态、软删除、其他用户夹具 | 通过，总数/预加载保持 |
| S03 前端参数 | KeysView.spec.ts的group asc/desc排序参数、筛选、分页重置用例 | 通过；14项KeysView测试全过 |

## 失败到通过及范围适配

- [邮箱并发初始失败](./evidence/f02-s01-red.txt)：50请求中13次仍返回invalid，超过允许4次；最终[s01-final](./evidence/s01-final.txt)及race通过。
- [旧缓存失败](./evidence/s01-legacy-red.txt)后增加原子继承JSON次数；[缓存测试](./evidence/s01-cache.txt)通过。
- [优先级生命周期失败](./evidence/s02-lifecycle-red.txt)：5项暴露非法输入、错账号提交、旧响应覆盖；修复后[30项定向前端](./evidence/frontend-targeted.txt)通过，包含16项优先级与14项密钥页。
- 最早[s02-s03-red](./evidence/s02-s03-red.txt)中，缺组件的导入错误只说明新组件未存在，不冒称行为失败；selectedKeys断言依赖本地未采用批量功能，剔除该断言后以SQLite真实排序失败及最终前端参数测试验收。
- [SMTP](./evidence/s01-smtp.txt)以及[最终邮箱检查](./evidence/s01-final.txt)通过；[全量unit](./evidence/backend-unit-final.txt)、[race](./evidence/backend-race.txt)、[typecheck](./evidence/frontend-typecheck.txt)、[lint](./evidence/frontend-lint.txt)、[build](./evidence/frontend-build.txt)和[Vitest](./evidence/frontend-vitest.txt)均退出0。

前端构建有既有chunk大小提示，pnpm有既有overrides迁移提示；未为此变更依赖或打包策略。
