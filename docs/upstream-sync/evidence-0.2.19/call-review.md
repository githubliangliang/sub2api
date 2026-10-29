# 新调用、前置与消费者审查

冻结产品基线：49939fb2cf96633688548ca39204d96558e59d79。原始宽松扫描见 [added-call-inventory.tsv](./added-call-inventory.tsv)，拟议第一/二档及用户选定Terra的328条产品调用token见 [review-selected-call-inventory.tsv](./review-selected-call-inventory.tsv)。扫描保留整PR产品文件，含未选定诊断子项；它不是静态类型检查器。

本地func定义已逐符号检索：[review-base-symbols.txt](./review-base-symbols.txt)。整PR新声明在同PR内追踪，Go内置/stdlib、gjson/Gin/Ent方法和SQL/注释误命中与业务缺前置分开。328条不能等同328个独立函数或缺陷。

| 消费者 / 新引用 | 本地状态与取法 | 实施必须证明 |
|---|---|---|
| Claude fingerprint / utf16.AppendRune、Decode | 当前Go标准库已有；共享fingerprint消费者存在 | JS UTF-16 oracle与ASCII对照，不全prompt分配 |
| Codex cache / get及fresh状态 | 本地是codexModelsManifestCache/refreshCachedAPIKeyCodexModelsManifest，上游名字不同 | flight内fresh复查实际减少fetch，ETag和隔离保留 |
| Gemini图片 / countGeminiImageOutputs、geminiPartIsImage | 新命名随计数消费者一起改，已有MIME判定 | 旧测试调用同步更新，累计帧去重，fileUri不下载 |
| Antigravity / ErrMalformedFunctionCall | source同PR新增sentinel；extractGeminiFinishReason、errors.Is及compat出口存在 | 无parts尾帧仍产生错误，调用链不返回空成功 |
| pagination / strconv.ParseUint、math.MaxInt | 标准库；既有Limit/Offset/HTTP入口 | offset+limit不溢出，非法query保持各入口契约 |
| admin / infraerrors.BadRequest | 当前导入和错误基础已存在 | handler400与数据库500分开，守卫数据不变 |
| Grok prior model / openAICompatPriorModelAudit | 新固定状态由同PR添加；本地observer/fields已有 | reader签名所有调用与mock更新，非Grok不累积 |
| Grok parser/store / Parse/Merge/PersistGrokCatalogObservation | source catalog310行+store112行，消费者手动/后台已存在 | 严格形态、旧格式、时间序及CAS有界重试；不引入265行诊断服务 |
| Grok CAS / CompareAndSwapGrokObservedModels | 新窄能力由同PR提供；source SQL是PG，需要SQLite改写 | JSON语义expected比较、单键更新、软删/取消/事务/缓存发布 |
| accountTransactionScope | 本地缺；源custom_admin_key_transaction.go的11行func来自#8319 | 用现有clientFromContext/withRepositoryTransaction落实效果，不引入受限密钥 |
| resolveModelsListReadLimit | 本地缺；源helper6行依赖新ModelsListReadMaxBytes配置 | 复用当前8MiB常量/limit+1边界，不顺带加设置 |
| channel / mergeWebSearchEmulationDimensions、withWebSearchEmulationKey | source同PR两函数36行（含说明增量约60行），applyUpdateInput已存在 | 只合并特定key，明确清空、copy-on-write，UI保存实际覆盖两维 |
| Gemini DeclaredModelIDs/DeclaredModelsCover | 本地缺；两个函数34行，含注释/空行增量41行，底层账号/映射/通配helper已存在 | all-mapped条件、列表详情同源；不整取344行endpoint基座 |
| AntigravityGeminiModelIDs | 本地缺；源23行收集函数，可随既有mixed消费者摘取 | force/composite/mixed边界与原路径护栏 |
| ModelAllowlistEnabled/FilterForListing | 本地缺完整上游系统 | 用本地CustomModelsListEnabled/ModelsListConfig效果，不引入新schema |
| Terra IsGPT61TerraModelSpelling/RequiresResponsesForToolCalls | source同PR两个函数19行；CanonicalizeOpenAIModelAliasSpelling和Sol校验已有 | allowed后缀、none/max、工具路由全贯通，Sol规则保留 |
| Terra新测试 / FetchOpenAIModelsList、FetchOpenAIAccountModels、ordinaryModelsUpstreamResponse | 本地缺这些source目录API/fixture；已有FetchCodexModelsManifest/FetchUpstreamSupportedModels | 适配实际路径证明目录不就地改写，不引入整目录系统 |
| Terra价格/metadata/config | 本地已具fallback、显式0、资源与manifest及UseKeyModal | 只摘Terra条目，账号metadata优先；OpenCode客户端配置不等于平台 |
| x/net | 外部module，0.62与强制MVS版本已隔离验证 | 实施full门禁及漏洞复扫，不能覆盖source整份go.mod |

对未选定能力，缺基座不是先补的任务：scoped key的授权/schema、image_count_billing新列、Grok diagnostics API/UI、Gemini缓存pool绑定、CN/OpenCode平台均不纳入第一/二档。上游价格和模型能力值是冻结来源事实，不是任意供应商/账号的实时保证。

本次baseline测试和overlay probes仅验证现状及必要基座；没有将整份patch编译成修复版本。最终源码/测试接口一致性须由对应change的实施门禁证明。

