# 前端实施与场景映射

实现范围：F01/F02/F03完整前端；F04规则scope创建/编辑/回显与类型；F05 OpenAI API Key账号三态和渠道配置。后端策略、真实SQLite/miniredis/handler记录由后端与主agent验收。本目录不更改冻结来源、迁移、VERSION或依赖。

2026-10-26完成补记：用户后续选择的完整运维页已恢复，新增路由/权限/standard与simple、侧栏监控开关/隐藏菜单及菜单项覆盖；定向红/绿证据为本目录`ops-restoration-{red,green}.log`。最终前端门禁为**285文件、2089通过、2项既有跳过**，typecheck/lint/build均退出0，见[最新Vitest日志](../completion-2026-10-26/frontend-vitest.log)和[统一验收](../../../../openspec/changes/port-upstream-0.2.18-personal-features/verification.md)。下方2080项结果是入口恢复前的历史记录，保留原值。

## 行为与测试

以下测试文件均在frontend/src内。场景名称与本次spec逐项对应；不将表单测试代替后端隔离/搜索执行证明。

| Spec场景 | 测试文件与用例 |
|---|---|
| F01 使用完整请求耗时 | utils/__tests__/usageOutputRate.spec.ts：uses total duration, supports all text transports and rounds tiny positive rates；components/admin/usage/__tests__/UsageTable.spec.ts：shows total-duration TPS and updates paged rows while honoring the latency column |
| F01 极小但有效的速率 | usageOutputRate.spec.ts：uses total duration…，正数1/100000ms显示0.0 |
| F01 缺少可信数值 | usageOutputRate.spec.ts：does not display invalid output or duration，覆盖0/-1/NaN/Infinity/undefined/null |
| F01 图片与Live | usageOutputRate.spec.ts：excludes image output, image billing and unsupported request types |
| F01 图片输入后的文本输出和历史记录 | UsageTable.spec.ts的TPS用例带image_input_tokens=400仍50.0；usageOutputRate.spec.ts缺少request_type仍可计算 |
| F01 默认显示与用户隐藏 | views/{admin,user}/__tests__/UsageView.spec.ts：keeps latency/TPS visible by default… / keeps the latency/TPS area visible by default…，分别覆盖默认与已保存隐藏偏好 |
| F01 分页或刷新 | UsageTable.spec.ts：shows total-duration TPS and updates paged rows…，更换row request_id/输出/耗时后50.0→40.0，再隐藏latency列 |
| F02 刷新后首个请求采用保存值 | views/admin/__tests__/AccountsView.lite.spec.ts：restores platform and status in the first request… |
| F02 清空与单字段修改 | AccountsView.lite.spec.ts首请求用例保存platform空/status active；utils/__tests__/accountFilters.spec.ts：restores only supported fields and persists clearing |
| F02 非法结构和字段 | accountFilters.spec.ts：ignores malformed structures；validates fields independently against local supported values |
| F02 存储被禁用或空间耗尽 | accountFilters.spec.ts：tolerates storage read and write failures；AccountsView.lite.spec.ts：opens and filters when storage fails specifically for the saved account filters |
| F02 从后页切换筛选 | AccountsView.lite.spec.ts：returns to page one and clears selected accounts when remembered filters change，实测第5页→第1页并清理ID42选择 |
| F02 保留本地平台范围 | accountFilters.spec.ts：validates fields independently…，拒绝cline但保留独立合法status |
| F03 成功和错误记录 | views/admin/ops/components/__tests__/OpsRequestDetailsModal.spec.ts：shows account names with id fallback without blocking details，桌面/移动两种布局，含success/error |
| F03 历史账号已删除或没有归属 | 同上：账号999回退#999，null显示-，正常明细和model保留 |
| F03 名称请求失败 | OpsRequestDetailsModal.spec.ts：refreshes account names on manual refresh and retains details when lookup fails；utils/__tests__/accountNameMap.spec.ts：does not retain failed requests as successful empty caches |
| F03 多页账号列表 | accountNameMap.spec.ts：paginates lite records and shares current in-flight work，断言lite=1、include_scheduler_score=0、AbortSignal以及后续页 |
| F03 账号新增或重命名 | accountNameMap.spec.ts：refreshes renamed accounts on force and expires cached values；组件用例缺项至多额外刷新一次；manual pending替换及replaces an older forced name snapshot on every explicit manual refresh证明每次显式刷新都会更新快照 |
| F03 快速翻页、关闭和重新登录 | OpsRequestDetailsModal.spec.ts：ignores a slower earlier page response after the filter context changes；does not write late names into a closed dialog and shares their pending lookup on reopening；does not apply stale details or names…session changes；cache会话隔离用例 |
| F04 编辑旧规则保持行为（前端部分） | components/account/__tests__/EditAccountModal.spec.ts：preserves legacy rule scope defaults, explicit scope and ordering when saving，覆盖无scope429、非法scope401、显式带空白ACCOUNT |
| F04 新建默认与显式作用范围（前端部分） | CreateAccountModal.spec.ts：creates OpenAI search emulation with model-scoped defaults and explicit account-scoped rules，preset默认model并切换account后真实提交 |
| F05 配置回显与兼容（前端部分） | EditAccountModal.spec.ts：round-trips OpenAI API key web search mode，覆盖enabled/disabled/true/false，保存default删除override且保留其他extra |
| F05 切换账号不串写（前端部分） | EditAccountModal.spec.ts：resets web search mode between accounts and never exposes it to OAuth or SetupToken；CreateAccountModal.spec.ts：clears API key search emulation when switching account categories；does not carry API key search overrides双向平台切换4例，真实提交验证目标无旧override |
| F05 渠道继承可配置（前端部分） | views/admin/__tests__/ChannelsView.webSearch.spec.ts：reads and saves OpenAI independently while retaining the Anthropic flag；creates an OpenAI channel with an enabled search flag |

新增标签和说明同时写入中英文locale；保留原有组件布局和菜单限制。F04诊断字段scope/model作为可选字段接收，兼容旧响应。

## 修正证据

- red-f01-f02.log：新增TPS格式化及筛选持久化测试在缺少实现时失败。
- red-f03.log：新名称缓存行为测试在模块不存在时失败。
- red-f03-pending-refresh.log：证明手动刷新与旧pending共享会只调用一次API而拿不到重命名后快照；修复为force加载替换旧普通请求，普通并发及已开始force并发仍共享。
- red-f03-resume.log：证明从错误详情返回resumeState时，旧请求序号已失效且没有续接，名称停留#ID；修复为保留明细上下文并续接名称查询。
- f03-final.log：上述缓存与生命周期相关17项通过。
- red-f05-platform-switch.log：交叉审查发现OpenAI→Anthropic同为API Key时旧enabled/disabled跨平台保留，两例实际失败；修复平台变化时无条件重置此共享表单字段，双向4例均通过。
- red-f03-repeated-manual-refresh.log：交叉审查发现第二次手动刷新会复用未结束的旧forced快照；组件级测试预期3次API、实际2次失败。显式刷新入口现先失效缓存/请求generation，普通及同轮缺项仍共享。
- cross-review-final.log：交叉审查修正后的缓存、详情和创建表单43项通过。

最终门禁结果以本目录原始日志及主verification为准。真实用户界面的桌面/移动展示由组件渲染测试覆盖；未在真实业务实例登录部署或执行收费搜索。

## 最终门禁结果

| 命令（frontend目录） | 结果 | 原始日志 |
|---|---|---|
| `rtk pnpm run typecheck` | 退出0 | typecheck.log |
| `rtk pnpm run lint:check` | 最终退出0 | lint-cross-review.log |
| `rtk pnpm run test:run` | 285文件，2080通过、2个原有跳过 | vitest-cross-review.log |
| `rtk pnpm run build` | vue-tsc/Vite成功，Vite阶段30.38秒 | build-cross-review.log |
| `rtk git diff --check -- frontend` | 退出0 | 主agent组合记录 |

相对基线新增53个测试通过（2027→2080）。第一轮全量2074通过后，自审补充resumeState回归用例并修复得到2075通过；交叉审查再补平台切换和重复手动刷新，最终2080通过，日志跨2026-10-24/25午夜。构建有既有Browserslist过期、pnpm.overrides位置和大chunk提示，未修改依赖/锁文件；打包产物位于gitignored backend/internal/web/dist。
