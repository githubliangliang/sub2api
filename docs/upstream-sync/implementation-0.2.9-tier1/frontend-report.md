# 第一档前端实施报告：F09–F14

状态：DONE。26 PR 产品与回归均完成；最终32个相关测试文件、193项通过。typecheck、lint:check、git diff --check -- frontend 均通过。未stage/commit/push/切分支。

实施基线：50c34ad49。来源固定 upstream-review/v0.2.9；frontend-port.py prepare 已将26份缓存patch与来源映射merge SHA的 git diff <merge>^1 <merge> -- frontend/逐字规范换行比较，全一致。产品仅按第一档PR序列应用；未整体复制最终tag文件。

## 执行与证据

每簇先应用测试、运行并检查对应旧行为失败，随后串行应用该簇产品patch并运行GREEN。六簇均实际观察到预期产品断言失败，无用测试环境错误替代RED。F10的数字trim未处理异常正是该PR目标缺陷，GREEN无异常。

所有 pnpm 命令工作目录 frontend/，前置 `$env:COREPACK_ENABLE_AUTO_PIN='0'`。F09/F10/F11及F12 RED的命令为 `pnpm exec vitest run <各簇测试路径>`；F12 GREEN起附加 `--maxWorkers=2 --minWorkers=1`，避免与后端验证争抢资源。下文每个PR列出测试路径；同簇一起运行。

| 簇 | RED | GREEN | 原始日志（./） |
|---|---|---|---|
| F09 | 3文件，11失败/9通过 | 3文件，20通过 | frontend-F09-red.txt / frontend-F09-green.txt |
| F10 | 9文件，25失败/15通过，3个预期数字trim异常 | 9文件，40通过 | frontend-F10-red.txt / frontend-F10-green.txt |
| F11 | 3文件，5失败/12通过（含既有proxyExpiry.spec） | 3文件，17通过；最终新增无到期反例后18通过 | frontend-F11-red.txt / frontend-F11-green.txt |
| F12 | 5文件，10失败/12通过 | 5文件，22通过；最终补0.5/2后24通过 | frontend-F12-red.txt / frontend-F12-green.txt |
| F13 | 4文件，14失败/1通过 | 4文件，15通过 | frontend-F13-red.txt / frontend-F13-green.txt |
| F14 | 3文件，11失败/16通过 | 3文件，27通过 | frontend-F14-red.txt / frontend-F14-green.txt |

最终复核命令：`pnpm exec vitest run --maxWorkers=2 --minWorkers=1 <frontend-final-specs.txt内32项>`；结果32 files / 193 tests passed，日志 frontend-final-tests.txt，耗时55.99s。其中26个来源回归文件以及既有 BaseDialog、DateRangePicker、Select、proxyExpiry、useModelWhitelist、AccountUsageCell 的6个文件。

`pnpm typecheck` → 通过（frontend-typecheck.txt）；`pnpm lint:check` → 通过、无自动fix（frontend-lint.txt）；`git diff --check -- frontend` → 通过。Vitest运行有已存在的 caniuse-lite 数据旧提示；未更新依赖。未运行整个前端全量测试，由root汇总门禁统一执行。

## 逐PR产品取舍与回归

### F09

**#7235** 来源 `efd86c63e6724838659cde590a0d4516d045acd0`

- 产品：`frontend/src/components/user/profile/TotpSetupModal.vue`。绑定每格验证码 value；保留输入、粘贴与回退流程。
- 测试：`frontend/src/components/user/profile/__tests__/TotpSetupModal.inputs.spec.ts`。
- RED：2/3 失败：失败后显示值不清空，回到验证码步骤不回显原值。证据 `frontend-F09-red.txt`。
- GREEN：该PR测试文件最终 3 项通过；命令 `pnpm exec vitest run src/components/user/profile/__tests__/TotpSetupModal.inputs.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F09-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：普通输入提交123456；失败后粘贴654321；首格焦点及提交禁用。

**#7184** 来源 `71c0eb51260fd238b5e485a0127770fe33c8d9d1`

- 产品：`frontend/src/components/user/profile/TotpDisableDialog.vue`；`frontend/src/components/user/profile/TotpSetupModal.vue`。两个 TOTP 组件复用 extractApiErrorMessage。
- 测试：`frontend/src/components/user/profile/__tests__/Totp.errors.spec.ts`。
- RED：7/11 失败：规范化 API message 被替换为泛化 fallback。证据 `frontend-F09-red.txt`。
- GREEN：该PR测试文件最终 11 项通过；命令 `pnpm exec vitest run src/components/user/profile/__tests__/Totp.errors.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F09-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：标准 message、旧 response.data.message、无消息 fallback；失败不触发 success。

**#7496** 来源 `9e0e1469844665c4bd3220bbec5836ab48b17537`

- 产品：`frontend/src/components/user/profile/TotpDisableDialog.vue`；`frontend/src/components/user/profile/TotpSetupModal.vue`。卸载设置 disposed；发送验证码成功返回后先检查 disposed。
- 测试：`frontend/src/components/user/profile/__tests__/totp-timer-cleanup.spec.ts`。
- RED：2/6 失败：卸载后的响应重新调用 setInterval 并显示成功 toast。证据 `frontend-F09-red.txt`。
- GREEN：该PR测试文件最终 6 项通过；命令 `pnpm exec vitest run src/components/user/profile/__tests__/totp-timer-cleanup.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F09-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：两个组件均覆盖迟到响应；既有卸载 clearInterval 和正常发送保留。


### F10

**#7234** 来源 `4daabc3d84ddcbaa4df9f74f2f8ac4db6fd157a0`

- 产品：`frontend/src/components/common/Pagination.vue`。跳页值先 String 再 trim，适配数值 input。
- 测试：`frontend/src/components/common/__tests__/Pagination.jump.spec.ts`。
- RED：3/4 失败：数字 trim 抛 TypeError，click/Enter 不触发跳页；另有3个预期产品异常。证据 `frontend-F10-red.txt`。
- GREEN：该PR测试文件最终 4 项通过；命令 `pnpm exec vitest run src/components/common/__tests__/Pagination.jump.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F10-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：click、Enter、超范围夹到末页、空输入不提交。

**#7263** 来源 `ea68c0e82c1c995a8eb797648af72d2088e2955f`

- 产品：`frontend/src/components/common/BaseDialog.vue`。只把 dialogIdCounter 提升到组件模块作用域。
- 测试：`frontend/src/components/common/__tests__/BaseDialog.ids.spec.ts`。
- RED：1/2 失败：第二个对话框 aria-labelledby 指向第一个标题。证据 `frontend-F10-red.txt`。
- GREEN：该PR测试文件最终 2 项通过；命令 `pnpm exec vitest run src/components/common/__tests__/BaseDialog.ids.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F10-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：两个同时打开的标题关联独立；同组件重开 ID 稳定；不取第二档 scroll-lock。

**#7186** 来源 `5f75e0d2a7ea7197476be4e77596735c459008f7`

- 产品：`frontend/src/composables/useClipboard.ts`。execCommand fallback catch 返回 false，finally 仍删除 textarea。
- 测试：`frontend/src/composables/__tests__/useClipboard.spec.ts`。
- RED：2/10 失败：两种安全上下文中 fallback 异常使 Promise reject。证据 `frontend-F10-red.txt`。
- GREEN：该PR测试文件最终 10 项通过；命令 `pnpm exec vitest run src/composables/__tests__/useClipboard.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F10-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：安全/非安全上下文、复制失败 toast、无成功 toast、textarea 清理；既有成功与空值。

**#7238** 来源 `57e9f6a4b81db65a9bd1c23c9c6cc7965a8e686b`

- 产品：`frontend/src/components/admin/channel/ModelTagInput.vue`。空/空白标签 Tab 不 preventDefault；非空 Tab 添加后清空。
- 测试：`frontend/src/components/admin/channel/__tests__/ModelTagInput.keyboard.spec.ts`。
- RED：4/4 失败：空输入、Shift+Tab、空白和添加后第二次 Tab 被拦截。证据 `frontend-F10-red.txt`。
- GREEN：该PR测试文件最终 4 项通过；命令 `pnpm exec vitest run src/components/admin/channel/__tests__/ModelTagInput.keyboard.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F10-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：非空添加一次，下一次 Tab 放行；不取第二档 ModelTagInput IME 补丁。

**#7374** 来源 `c96d36ec9567c98a378e0b93b46ee2d9b4242994`

- 产品：`frontend/src/components/common/SearchInput.vue`。SearchInput 使用 computed v-model，复用 Vue 原生 composition 语义。
- 测试：`frontend/src/components/common/__tests__/SearchInput.spec.ts`。
- RED：1/3 失败：IME 中间拼音触发 modelValue 更新。证据 `frontend-F10-red.txt`。
- GREEN：该PR测试文件最终 3 项通过；命令 `pnpm exec vitest run src/components/common/__tests__/SearchInput.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F10-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：组合中无中间更新/搜索；中文完成后防抖；普通文本防抖；父值更新不搜索。

**#7142** 来源 `ea34a2356ca546cb12cd960efc203baaff5e39c4`

- 产品：`frontend/src/components/common/Select.vue`。打开时 filteredOptions 变化重新定位首个 enabled 选项。
- 测试：`frontend/src/components/common/__tests__/Select.searchHighlight.spec.ts`。
- RED：4/5 失败：过滤、无结果恢复、远端结果及禁用旧索引均不能 Enter 选择。证据 `frontend-F10-red.txt`。
- GREEN：该PR测试文件最终 5 项通过；命令 `pnpm exec vitest run src/components/common/__tests__/Select.searchHighlight.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F10-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：异步结果跳过 disabled；重开恢复原选择。

**#7109** 来源 `d27fbb3b7dfd6a2a5a26e80fbca781fc7bbdc112`

- 产品：`frontend/src/components/common/DateRangePicker.vue`。关闭后以 flush:post 恢复已应用的父日期与预设。
- 测试：`frontend/src/components/common/__tests__/DateRangePicker.dismiss.spec.ts`。
- RED：3/4 失败：outside/Escape/toggle 都保留未应用的7天草稿。证据 `frontend-F10-red.txt`。
- GREEN：该PR测试文件最终 4 项通过；命令 `pnpm exec vitest run src/components/common/__tests__/DateRangePicker.dismiss.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F10-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：三种取消；父组件接受 Apply 后保留新范围。

**#7108** 来源 `974a819ba15359b00b87f4c96ad3b7ae3b4a6f93`

- 产品：`frontend/src/components/common/Select.vue`。非搜索 listbox tabindex=-1，打开后真实聚焦该节点。
- 测试：`frontend/src/components/common/__tests__/Select.keyboard.spec.ts`。
- RED：3/4 失败：无搜索焦点留在 trigger，Escape 无法关闭。证据 `frontend-F10-red.txt`。
- GREEN：该PR测试文件最终 4 项通过；命令 `pnpm exec vitest run src/components/common/__tests__/Select.keyboard.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F10-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：false/auto 键盘跳过禁用选项；Enter 选择及还焦点；搜索输入焦点保留。

**#7499** 来源 `0eaa7c3c84c53704438809f52887c44824b14b95`

- 产品：`frontend/src/components/common/DateRangePicker.vue`。today/tomorrow 改点击/渲染时计算函数，保留时区处理。
- 测试：`frontend/src/components/common/__tests__/DateRangePicker.midnight.spec.ts`。
- RED：4/4 失败：今天、7天、本月的结束日及输入 max 使用前一日缓存。证据 `frontend-F10-red.txt`。
- GREEN：该PR测试文件最终 4 项通过；命令 `pnpm exec vitest run src/components/common/__tests__/DateRangePicker.midnight.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F10-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：本地月底午夜后仍挂载；日期预设与重开 max 都更新。


### F11

**#7183** 来源 `53fd5f83512efbe95891e16138ecfd57f9bda8b7`

- 产品：`frontend/src/components/common/ProxySelector.vue`。批量测试复用 handleTestProxy 与既有 testingProxyIds/testResults。
- 测试：`frontend/src/components/common/__tests__/ProxySelector.testing.spec.ts`。
- RED：1/2 失败：先单测代理1再批量会重复请求代理1。证据 `frontend-F11-red.txt`。
- GREEN：该PR测试文件最终 2 项通过；命令 `pnpm exec vitest run src/components/common/__tests__/ProxySelector.testing.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F11-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：同代理不重复请求；不同代理US/GB隔离；失败后可再次批量。

**#7321** 来源 `4b0adb68636c4d8bef479e2c1ed541950a031c27`

- 产品：`frontend/src/utils/proxyExpiry.ts`。daysUntil 为0时按 expired 展示。
- 测试：`frontend/src/utils/__tests__/proxyExpiry.boundary.spec.ts`。
- RED：4/6 失败：等于截止与未满一天的过期边界仍显示 expiringInDays。证据 `frontend-F11-red.txt`。
- GREEN：该PR测试文件最终 7 项通过；命令 `pnpm exec vitest run src/utils/__tests__/proxyExpiry.boundary.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F11-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：截止前1ms、等于截止、截止后1ms、整天逾期；最终补无 expires_at 反例。


### F12

**#7372** 来源 `c0b23ff79116baec40fe2d66b33eccb6b74427a2`

- 产品：`frontend/src/stores/adminSettings.ts`。只删除 settings fetch catch 中 loaded=true。
- 测试：`frontend/src/stores/__tests__/adminSettings.retry.spec.ts`。
- RED：2/4 失败：第一次加载失败被永久标记 loaded。证据 `frontend-F12-red.txt`。
- GREEN：该PR测试文件最终 4 项通过；命令 `pnpm exec vitest run src/stores/__tests__/adminSettings.retry.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F12-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：重试成功；成功缓存复用；失败保留缓存值；仅验证本地已有依赖，未改支付逻辑。

**#7264** 来源 `e7d348868c3891d0a63889807537552bd4c0cadc`

- 产品：`frontend/src/composables/useModelWhitelist.ts`。映射失败直接返回[]而不写模块级缓存。
- 测试：`frontend/src/composables/__tests__/antigravityMappings.retry.spec.ts`。
- RED：2/4 失败：失败后不重试，迟到失败覆盖并发成功缓存。证据 `frontend-F12-red.txt`。
- GREEN：该PR测试文件最终 4 项通过；命令 `pnpm exec vitest run src/composables/__tests__/antigravityMappings.retry.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F12-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：成功非空/空映射均缓存；迟到失败不抹掉成功结果。

**#7364** 来源 `9e175bd49161d486b3dd5d87f43724d98a45b07a`

- 产品：`frontend/src/components/admin/user/GroupReplaceModal.vue`。GroupReplace catch 复用 extractApiErrorMessage 显示 toast。
- 测试：`frontend/src/components/admin/user/__tests__/GroupReplaceModal.spec.ts`。
- RED：3/4 失败：错误只有 console 日志，用户看不到API消息。证据 `frontend-F12-red.txt`。
- GREEN：该PR测试文件最终 4 项通过；命令 `pnpm exec vitest run src/components/admin/user/__tests__/GroupReplaceModal.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F12-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：message/detail/fallback；失败不发 success/close且可重试；成功迁移照常关闭。

**#7376** 来源 `ca47fa352c6942907ae01b3c36eb508bc38d01e5`

- 产品：`frontend/src/components/admin/group/GroupRateMultipliersModal.vue`。新增倍率按钮和处理器拒绝 null/<=0。
- 测试：`frontend/src/components/admin/group/__tests__/GroupRateMultipliersModal.spec.ts`。
- RED：1/5 失败：负数倍率被允许添加；0和空值原行为已拒绝。证据 `frontend-F12-red.txt`。
- GREEN：该PR测试文件最终 6 项通过；命令 `pnpm exec vitest run src/components/admin/group/__tests__/GroupRateMultipliersModal.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F12-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：空值、0、-1不添加/不提交；0.25/0.5/1 正常保存；新增0.5反例不伪造RED。

**#7367** 来源 `a765b3c1765f50a18e3b845170916d91c21a12cc`

- 产品：`frontend/src/components/admin/group/GroupRPMOverridesModal.vue`。新增 RPM 按钮和处理器补 Number.isInteger。
- 测试：`frontend/src/components/admin/group/__tests__/GroupRPMOverridesModal.spec.ts`。
- RED：2/5 失败：空字符串与1.5可添加。证据 `frontend-F12-red.txt`。
- GREEN：该PR测试文件最终 6 项通过；命令 `pnpm exec vitest run src/components/admin/group/__tests__/GroupRPMOverridesModal.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F12-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：空、1.5、负数拒绝；0、2、100正常保存；新增2反例不伪造RED。


### F13

**#7420** 来源 `f2e55bf4d704ae5730f207d437f563c41be9573a`

- 产品：`frontend/src/components/user/UserErrorDetailModal.vue`。错误详情 watch cleanup/requestVersion 限制当前请求写回。
- 测试：`frontend/src/components/user/__tests__/UserErrorDetailModal.spec.ts`。
- RED：3/4 失败：旧数据覆盖、旧请求结束loading、旧错误覆盖重开状态。证据 `frontend-F13-red.txt`。
- GREEN：该PR测试文件最终 4 项通过；命令 `pnpm exec vitest run src/components/user/__tests__/UserErrorDetailModal.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F13-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：A→B乱序；A先完成不改变B loading；同对象关闭重开；当前失败仍显示。

**#7422** 来源 `5af29a3b93122179568195d92c8f02afe45231fb`

- 产品：`frontend/src/components/admin/user/UserApiKeysModal.vue`。API Key 弹窗监听 user.id、清空旧 keys、按 requestVersion 写回。
- 测试：`frontend/src/components/admin/user/__tests__/UserApiKeysModal.spec.ts`。
- RED：4/4 失败：旧Key泄露、迟到覆盖、旧失败清loading、打开时换用户不加载。证据 `frontend-F13-red.txt`。
- GREEN：该PR测试文件最终 4 项通过；命令 `pnpm exec vitest run src/components/admin/user/__tests__/UserApiKeysModal.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F13-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：切用户失败不能显示前用户Key；过时成功/失败隔离；打开时直接A→B。

**#7493** 来源 `24872fda7cf89ad6be5f059751dc01e741da1f9e`

- 产品：`frontend/src/components/account/TempUnschedStatusModal.vue`。临时不可调度状态清空旧状态、请求版本保护与 watch cleanup。
- 测试：`frontend/src/components/account/__tests__/TempUnschedStatusModal.spec.ts`。
- RED：3/3 失败：旧账号覆盖、加载B时恢复按钮沿用A状态、旧错误清loading。证据 `frontend-F13-red.txt`。
- GREEN：该PR测试文件最终 3 项通过；命令 `pnpm exec vitest run src/components/account/__tests__/TempUnschedStatusModal.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F13-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：A→B成功乱序、恢复按钮状态、关闭重开后旧失败不toast。

**#7497** 来源 `89e9976950a8928d0285fb7efd535cdf5c7a5b38`

- 产品：`frontend/src/components/admin/group/GroupRPMOverridesModal.vue`；`frontend/src/components/admin/group/GroupRateMultipliersModal.vue`。两个分组弹窗卸载 clearTimeout 并移除自身 document click handler。
- 测试：`frontend/src/components/admin/group/__tests__/GroupModal.cleanup.spec.ts`。
- RED：4/4 失败：两组件都残留监听和卸载后搜索。证据 `frontend-F13-red.txt`。
- GREEN：该PR测试文件最终 4 项通过；命令 `pnpm exec vitest run src/components/admin/group/__tests__/GroupModal.cleanup.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F13-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：真实组件注册/移除同一handler；fake timers推进300ms无请求。


### F14

**#7446** 来源 `5295bd822780c1242d8bd034f817a417367be7b9`

- 产品：`frontend/src/views/user/UsageView.vue`。仅单独字符 - 绕过CSV公式前缀转义，不新增导出列。
- 测试：`frontend/src/views/user/__tests__/UsageView.spec.ts`。
- RED：7/13 失败：缺推理力度仍输出带单引号的占位符，六种公式导出断言同时检测该缺口。证据 `frontend-F14-red.txt`。
- GREEN：该PR测试文件最终 13 项通过；命令 `pnpm exec vitest run src/views/user/__tests__/UsageView.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F14-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：真实导出CSV；-1+1、=、+、@、TAB、CR继续转义；列/精度/历史费用保持。

**#7418** 来源 `5fd346114252ff66f49dc0e70b6f4e9958e0c601`

- 产品：`frontend/src/components/admin/channel/IntervalRow.vue`。token边界以 Number 后 Math.trunc 替代 parseInt。
- 测试：`frontend/src/components/admin/channel/__tests__/IntervalRow.spec.ts`。
- RED：3/5 失败：1e5、1e6、2.72e5被截成1/1/2。证据 `frontend-F14-red.txt`。
- GREEN：该PR测试文件最终 5 项通过；命令 `pnpm exec vitest run src/components/admin/channel/__tests__/IntervalRow.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F14-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：指数、普通整数、空min=0/max=null；fixture使用本地cache_write_1h_price且不带上游倍率字段。

**#7611** 来源 `99d86e41acb47d632de27598a727c671ed7fadb1`

- 产品：`frontend/src/components/account/UsageProgressBar.vue`。只有无 reset 且空闲才显示 resetNow；已有reset按真实倒计时。
- 测试：`frontend/src/components/account/__tests__/UsageProgressBar.spec.ts`。
- RED：1/9 失败：利用率0且未来reset错误显示现在。证据 `frontend-F14-red.txt`。
- GREEN：该PR测试文件最终 9 项通过；命令 `pnpm exec vitest run src/components/account/__tests__/UsageProgressBar.spec.ts`，实际以簇批次执行；初次GREEN日志 `frontend-F14-green.txt`，最后复核 `frontend-final-tests.txt`。
- 反例/边界：无reset空闲、非空闲未来reset、过期仍保留本地resetPending、剩余容量颜色。

## 本地适配与范围说明

- 产品patch全部使用第一档来源hunk，保留本地Select布局、IntervalRow一小时cache价格、UsageProgressBar过期resetPending、UsageView现有CSV列及费用等差异。无新增后端、依赖、VERSION、迁移或支付功能改动。
- #7418上游测试fixture含本地没有的倍率字段，按本地IntervalFormEntry调整为cache_write_1h_price；不补第二档产品类型。#7446测试改由真实导出按钮触发，并增加六种CSV公式前缀反例。#7611补无reset反例及组件自动卸载。
- #7183按源PR采用既有testingProxyIds防重复、共享同一testResults槽。再次发起同一代理时立即跳过；没有新增返回同一Promise或让批量等待先前单测的接口。这满足当前可观察的同代理不重复请求/不同代理不串结果场景。
- #7376/#7367源PR只收紧“新增覆盖”输入。既有行编辑/保存路径保留本地现状，不以第一档名义扩大校验范围；如将规格解释为所有编辑路径也须同样严格，需补一轮编辑/保存回归与最小修复。已告知root。
- #7496使用真实组件与setInterval spy证明卸载后的返回不建timer；原有成功发送/卸载清理同时通过。宿主ProfileTotpCard通过v-if关闭并销毁组件。
- #7321未配置expires_at仍返回原有remainingDays语义，实际UI按既有条件隐藏；新反例确认不会标为过期。
- 纯样式/字面量无额外镜像测试；异步组件只stub接口/展示容器，不stub被测业务逻辑。无已发现的未修复第一档来源回归。

报告生成数据：frontend-manifest.json；全部原始RED/GREEN日志与分离的来源product/tests patches均位于./。

## F12 复核补漏：已有行编辑与最终提交（2026-09-28）

本节取代上文“#7376/#7367只收紧新增覆盖，已有编辑保留原状”的范围说明。root复核确认该限制未满足已批准F12规格，因此本轮在两个已有组件内补齐新增、编辑、最终保存的一致校验。

产品只修改 `frontend/src/components/admin/group/GroupRateMultipliersModal.vue` 与 `frontend/src/components/admin/group/GroupRPMOverridesModal.vue`：

- 每个组件使用单一有效性判定：倍率为finite且>0；RPM为Number.isInteger且>=0（同时排除NaN/Infinity）。新增按钮、行输入状态、保存按钮和handleSave共用该判定。
- 使用number input原生valueAsNumber读取完整数值，移除输入的parseFloat/parseInt和v-model.number部分解析路径；合法1e3变成1000，1.5/-0.5不再截断成1/0。
- 无效草稿保留在本地，并禁用整次保存；handleSave自身同样拒绝无效条目，不能靠程序化点击绕过。其它条目不丢失、不提交旧值假装编辑成功。修正无效值后可正常提交全体条目，合法RPM 0保留。
- 倍率真实清空继续为null，保存时移除该覆盖且保留其它用户覆盖；原生badInput区分“未完成的数字”与用户主动清空。
- 使用空DOM value表示内部NaN，避免在用户尚未输完1e等输入时将NaN写回原生控件打断输入。该行为另有五个DOM setter回归测试。
- 批量倍率调整也使用同一正数有限性判定；乘积出现Infinity时最终保存被整体拒绝。无新依赖或i18n改动。

测试只扩充原来的两个validation spec，均通过真实组件、server entries、输入和保存按钮验证：

- `frontend/src/components/admin/group/__tests__/GroupRateMultipliersModal.spec.ts`：17项。
- `frontend/src/components/admin/group/__tests__/GroupRPMOverridesModal.spec.ts`：18项。
- 复用 `GroupModal.cleanup.spec.ts`：4项，确认生命周期清理不回退。

RED：`frontend-F12-edit-red.txt`，两文件30项中14失败/16通过，复现倍率0/-1/指数溢出/未完成输入、RPM空值/小数/负数/指数溢出/未完成输入以及合法指数被截断。

中间验证：`frontend-F12-edit-green.txt` 为33通过/1失败，发现jsdom把1e309规范为空值却不模拟原生badInput；按浏览器输入状态补正fixture。随后加入五项“不能向DOM写入NaN”测试，`frontend-F12-edit-native-red.txt` 为5失败/30通过；本轮修复该新发现，没有留下已知失败。

GREEN：`pnpm exec vitest run --maxWorkers=2 --minWorkers=1 src/components/admin/group/__tests__/GroupRateMultipliersModal.spec.ts src/components/admin/group/__tests__/GroupRPMOverridesModal.spec.ts src/components/admin/group/__tests__/GroupModal.cleanup.spec.ts` → 3文件39项全通过，exit 0；日志 `frontend-F12-edit-green-final.txt`，14.55s。

`pnpm typecheck` → exit 0，日志 `frontend-F12-edit-typecheck.txt`。`git diff --check`（两组件和两spec）→ exit 0。`pnpm lint:check` → exit 0，日志 `frontend-F12-edit-lint.txt`。所有pnpm命令仍设置COREPACK_ENABLE_AUTO_PIN=0，未stage/commit/push。

## F12 输入过程复核：保留科学计数法原始文本（2026-09-28）

本节取代上一轮“每次input使用valueAsNumber写入本地值”的实现说明。限定复核发现，这种即时数值化会把用户正在编辑的`2e1`回写为`20`，随后追加`2`会错误提交`202`。问题同时影响已有倍率行、新增倍率、batchFactor、已有RPM行和新增RPM五条路径。

新增五项回归先按真实`input`事件逐段输入`2`、未完成的`2e`、`2e1`，再对控件当时实际保留的值追加`2`。RED：`frontend-F12-draft-red.txt`，两文件40项中5失败/35通过；五项都观察到DOM被改为20并最终提交202，期望应为2e12/2000000000000。

最小修复仍限于上述两组件和两validation spec：

- 新增倍率、RPM及批量倍率保留字符串草稿；已有行的组件局部LocalEntry允许字符串，服务端/API类型不变。
- `@input`只保存原始文本。完整数值转换仅用于有效性判断、dirty比较、预览、明确点击添加/批量应用和最终提交，不再每次input向控件写入格式化数值。
- 仍通过Number完整转换与finite正倍率/非负整数RPM检查；不恢复parseFloat/parseInt截断行为。空RPM依然无效，已有倍率主动清空仍为null移除覆盖，badInput仍不能假作清空。
- 原有无效条目阻止整次提交、程序化保存保护、其它条目不丢失、合法RPM0和倍率清空、批量溢出保护等全部回归保持通过。

GREEN：`pnpm exec vitest run --maxWorkers=2 --minWorkers=1 src/components/admin/group/__tests__/GroupRateMultipliersModal.spec.ts src/components/admin/group/__tests__/GroupRPMOverridesModal.spec.ts src/components/admin/group/__tests__/GroupModal.cleanup.spec.ts` → 3文件44项全通过（倍率20、RPM20、cleanup4），exit 0；最终日志 `frontend-F12-draft-green-final.txt`，10.86s。

`pnpm typecheck` → exit 0，`frontend-F12-draft-typecheck.txt`；两组件与两spec `git diff --check` → exit 0。本轮`pnpm lint:check` → exit 0，`frontend-F12-draft-lint.txt`。仍未stage/commit/push。
