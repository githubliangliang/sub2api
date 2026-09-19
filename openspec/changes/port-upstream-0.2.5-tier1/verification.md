# 验收方案与证据

## 当前状态

**文档已补全，产品实现/构建/测试均未执行。** 下列矩阵为验收计划，不是 PASS 证据。

- 文档核对日期：2026-09-19。
- 核对 HEAD：`fe6bc318800ca86d2b658f0e8f46e83d32626e8b`。
- 核对起点工作树：仅 evidence-0.2.5、tier1、tier2 三个未跟踪目录；无已跟踪产品改动。
- 实施实际起点 SHA/日期：未开始，实施前重新记录。
- 产品代码 SHA：无，本次未实现、未提交。
- 文档静态核对：15 个来源 commit 可读取；整 PR 正向检查 14 CLEAN，#7024 因缺失测试文件不通过；
  SubscriptionsView.vue 单独检查通过。详细勘误及冻结文件哈希见 implementation-map.md。
- OpenSpec CLI：当前 PATH 未找到 `openspec`；不能宣称官方 strict 校验通过。

## 逐项验收矩阵（T01–T23）

每项完成时补充：实际命令及测试名、改前可观察结果、改后/反例结果、日志或截图路径、
执行日期、退出码及实施 SHA（未提交则写明工作树）。测试名在落地后填写，不能预报不存在的用例为通过。

| ID | 改前触发 / 检查 | 改后要求与反例 | 计划验证层级 | 当前执行结果 |
|---|---|---|---|---|
| T01 | A/B 同 project_id、不同 token，先缓存 A 后读 B | B 得自身 token；账号键不同；失效删除账号及旧 project 键；无 project_id 正常 | service unit + cache fake | 未执行 |
| T02 | 三个上游 data 事件各带空分隔行 | 输出恰好三帧，无额外空行粘连；非空注释透传、data 仍正确解包 | service 有限流输出测试 | 未执行 |
| T03 | 客户端 gzip 经实际 transport 转发 | 服务端只见一个 Accept-Encoding；无显式编码时不改变 transport 默认；其它头规则不变 | httptest 本地 HTTP | 未执行 |
| T04 | 候选 priority=10/5/1，全部支持 | 选 1；同优先级选较小 ID；不可用/不支持的候选跳过 | service unit | 未执行 |
| T05 | 6 个字段有实际值，含利用率 0 | 投影原样保留；缺键不凭空补 null；无关敏感键继续过滤 | repository 普通 unit | 未执行 |
| T06 | UA 含 CR/LF、NUL、DEL 或首尾非法控制字节 | 不成为配对身份/出站头；正常 UA、允许的空白与默认回退不变 | pkg/openai + service unit | 未执行 |
| T07 | JWT/管理员 GetByID 返回 DB 错误及包装 not-found | 前者 500 INTERNAL_ERROR，后者 401 USER_NOT_FOUND；有效用户通过；API Key 查询错误500/过载503/缺失401保持 | middleware httptest + 本地 fake | 未执行 |
| T08 | 原请求401后 refresh 无响应/429/500/503 | 四个存储项/当前位置不变；无 auth_expired；TOKEN_REFRESH_UNAVAILABLE 状态保真；401/403/其它原有拒绝仍清理；新会话不受旧请求影响 | client.spec.ts + tokenRefresh.spec.ts | 未执行 |
| T09 | 创建/更新启用峰谷，22:00–02:00 或非法格式 | HTTP400、INVALID_PEAK_RATE_CONFIG，未写库；09:00–18:00 合法；禁用/归一化规则保持 | service + handler 映射 | 未执行 |
| T10 | 浮层 input/output/cache/image/summary 金额为 0.00000012 | 显示 $0.00000012；缺省为 $0.00000000；账号费用也8位；原始计费值/表格不变 | UsageTable.spec.ts | 未执行 |
| T11 | 360×640 视口、至少10个套餐打开续费弹窗 | 标题可见；列表 scrollHeight>clientHeight，滚动到底最后套餐可见可选；1440×900 少量套餐无裁切 | 真实浏览器 + PaymentView 回归 | 未执行 |
| T12 | EasyPay upstreamType=alipay.qr | 前后端通过；空格/斜杠/大写拒绝；原有 alipay_qr/alipay-qr 通过；中英提示含点号 | Go 配置 unit + PaymentProviderDialog.spec.ts | 未执行 |
| T13 | 两待验证邮箱并发；一个请求未完成时重排/移除另一项 | 成功只移除对应身份；无关项保留；失败不删除；手动删除正确 | 新增 Vue 组件异步测试 | 未执行 |
| T14 | 第5页分别改变协议/状态等现有筛选 | 下一请求 page=1、page_size不变；匹配结果显示；分页按钮仍请求指定页 | ProxiesView.list.spec.ts | 未执行 |
| T15 | 选A后改关键词为B，尚未推进防抖计时即提交 | selectedUser/user_id失效且不发旧分配；新选B提交B；trim后仍为A邮箱则保持A | 新增 Vue 组件 + fake timers | 未执行 |
| T16 | 输入366/36500/36501，核对原min/step | 前两者允许，36501因max拒绝；后端MaxValidityDays=36500不变 | 浏览器约束/源定义核验 | 未执行 |
| T17 | 兑换成功、随后 refreshUser 失败 | 保留成功、独立warning、继续后续处理；兑换本身失败仍报错 | src/views/user/__tests__/RedeemView.spec.ts | 未执行 |
| T18 | 跨页导出期间改日期/筛选 | 各页参数及文件名冻结；列格式不变 | src/views/user/__tests__/UsageView.spec.ts | 未执行 |
| T19 | agent_message 信封与正文分离 | input_text/encrypted_content 按序转user；空项跳过，reasoning/工具不污染 | apicompat chat bridge unit | 未执行 |
| T20 | 重置A的配额时选择B | 返回值更新A的quota_used/status，不改B表单；当前仍为A时同步 | src/views/user/__tests__/KeysView.spec.ts | 未执行 |
| T21 | 非默认间隔、连续自动/手动刷新 | 周期无额外一秒、重置仍为所选值；关闭不重启 | 在 composables/views 相邻 __tests__ 适配原第二档用例 | 未执行 |
| T22 | 资料/密码API包装错误 | 显示已提取错误或兜底；成功清理/用户更新不变 | ProfilePasswordForm.spec.ts 及同目录资料表单用例 | 未执行 |
| T23 | 部分导入成功后关闭 | imported只发一次；全失败不发、完整成功不重复、重开清状态 | src/__tests__/integration/proxy-data-import.spec.ts | 未执行 |

前移七项均来自原第二档，所有产品验证仍未执行；上表测试路径省略 frontend/ 时相对前端目录。

API Key 已正确的分支记录改前/改后均通过，不制造虚假的失败证据。
T10/T11/T16 的直接检查可替代专门为格式/类名/max 属性编写的镜像测试，但不能替代真实布局和输入验证。

## 可执行门禁

所有命令均为**后续实施时执行**。退出码非0、超时、未运行均不能标 PASS；
若基线已有失败，记录基线复现与本批影响，不通过修改无关代码掩盖。

| 门禁 | 工作目录 | 命令 | 当前结果 |
|---|---|---|---|
| 相关后端包 | backend/ | `go test -tags=unit ./internal/pkg/openai ./internal/pkg/apicompat ./internal/service ./internal/repository ./internal/server/middleware` | 未执行 |
| 后端 build | backend/ | `go build ./...` | 未执行 |
| 后端 unit | backend/ | `go test -tags=unit ./...`（或根目录 `make -C backend test-unit`） | 未执行 |
| SQLite 方言审计 | backend/ | `go test -tags=unit ./internal/repository -run '^TestProductionSQLUsesSQLiteDialect$' -count=1` | 未执行 |
| 前端 typecheck | frontend/ | `pnpm run typecheck` | 未执行 |
| 前端 lint | frontend/ | `pnpm run lint:check` | 未执行 |
| 前端关键套件 | 仓库根 | `make test-frontend-critical` | 未执行 |
| 已有定向前端用例 | frontend/ | `pnpm exec vitest run src/api/__tests__/client.spec.ts src/api/__tests__/tokenRefresh.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts src/views/user/__tests__/PaymentView.spec.ts src/components/payment/__tests__/PaymentProviderDialog.spec.ts src/views/admin/__tests__/ProxiesView.list.spec.ts src/views/user/__tests__/RedeemView.spec.ts src/views/user/__tests__/UsageView.spec.ts src/views/user/__tests__/KeysView.spec.ts src/components/user/profile/__tests__/ProfilePasswordForm.spec.ts src/__tests__/integration/proxy-data-import.spec.ts` | 未执行 |
| 新增/适配定向用例 | frontend/ | 落地 T13/T15/T21/T22 的缺失覆盖后将实际测试文件路径附加到上一命令并记录；不能只跑现有套件即勾完成 | 未创建、未执行 |
| UI 实视口 | 浏览器 | 按 T11/T16 行记录视口、操作、结果和截图路径 | 未执行 |

已有关键套件不包含所有本批文件，因此必须另跑定向回归。
本次文档补全只作链接/编号/来源路径/冻结哈希核查；不为文档变更运行产品门禁。

## 范围审计与证据要求

实施完成后相对实际起点审查已提交 diff、暂存/未暂存 diff 和未跟踪文件，不能只看 `git diff`：

- `backend/migrations/` 无变更，最大编号仍226。
- `backend/ent/`、Wire 定义及生成文件无变更；VERSION 文件仍1.1.12。
- go.mod/go.sum、package.json、pnpm-lock.yaml 及其它依赖/锁文件无变更。
- T05 postgres 测试 hunk 与 T12 refund 格式 hunk 未带入，构建标签未放宽。
- T07 API Key 无无依据产品变更；T15 没有引入上游无关测试 fixture。
- 每个产品改动能追溯 implementation-map.md 的任务，已有用户改动保持。

| 完成项 | 当前证据 |
|---|---|
| 产品范围审计 | 未实施，待运行上述检查 |
| 全部23项行为证据 | 未实施，矩阵全部待执行 |
| 未通过/未执行门禁登记 | 已明确区分计划与执行；实施后继续回填 |
| 实施交接 | 未提交，无产品 SHA；交付前填实际状态与剩余缺口 |

## 文档核对记录

2026-09-19 首轮补全tier1文档，随后按用户指令同步调整两档任务归属；冻结source-*与原始三个TSV保持原样。
调整前已执行 Node 内联静态检查，退出码0：158项断言通过，覆盖当时 T01–T16 在现行映射/spec/tasks/验收矩阵中完整对应、
现行内部链接可解析、引用的本地产品/测试路径存在、冻结 SHA-256 不变、产品任务未误勾完成、无尾随空白及无已跟踪产品修改。
勘误中明确标注的缺失上游测试文件按“不存在”验证；来源 commit 另经只读 git diff/apply 检查确认可读取。
七项前移后已执行跨档静态检查，退出码0：399项断言通过；第一档22个PR/23项需求、第二档21个PR，
合计仍为原43个PR，归属无重叠或遗漏，第三/第四档未进入实施；任务/验收/路径/链接检查通过。
四份source-*与三个原始TSV的SHA-256均与调整前相同；无已跟踪产品改动。原158项仅保留为历史检查记录。
这项自检不等同于 OpenSpec CLI strict 校验，也不作为任何产品门禁的通过证据。
冻结历史中保留的缺失 PORTING 链接及旧编号由 implementation-map.md 明确勘误，不代表现行依赖。
