# 新增调用人工复核

added-call-inventory.tsv是所有候选Go新增行的词法清单；同名不代表同类型，EXTERNAL_OR_DYNAMIC须结合import/调用点解读，不是缺基座的结论。

- #7780选中后端生产新增调用均为本地函数或标准库；runtime只读GOOS/GOARCH，semver既有。包/transport版本限制有意不同。
- #7674的3个helper在PR内新增，复用的sanitizeUpstreamErrorMessage/extractUpstreamErrorMessage本地已有。MustCompile/ReplaceAllString是regexp，b/project为正则字符串误命中。
- #7676 cache helper、verifyCodeWithAttempts/hashPasswordResetToken及三个接口方法随PR加入。NewScript/TxPipeline/Del来自既有go-redis；call/pcall属于Lua/cjson；incr/onSuccess是回调；atomically/token为注释误命中。必须更新fork全部EmailCache mock，不能仅依赖同名检索。脚本在miniredis的执行留给实施验收。
- #7630前端computed/nextTick/onBeforeUnmount/ref/watch来自既有Vue，useI18n来自vue-i18n；updateAccount是现有update导入别名，extractApiErrorMessage和handleAccountUpdated均有定义。step/startEditing/nudgeInput/commitInput/cancelInput/clamp/clearTimer/save/scheduleSave均为新增组件内定义；Math/Number/setTimeout/clearTimeout为运行时。未见跨PR前置。
- #7803 apiKeyListOrder、Ent ByGroupField、SQLite fixture helper均已有；OrderNullsLast为现有Ent SQL依赖，KeysView复用现有排序参数链。尚无新排序行为验收。
- #7773纯字符串修改，无新增调用。
- TypeSafe/payment未选中，其缺失helper与新调用不为了凑编译导入。

所有来源测试中的require/assert、testing、httptest等是外部/标准测试API；新测试里SOURCE_ADDED符号需随选中行为加入。此复核不能替代未来新补丁的类型检查和运行测试。
