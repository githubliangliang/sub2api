# v1.1.26 同名标签重打准备（2026-11-19）

用户明确请求「重新打tag v1.1.26,详细描述，push」。本次沿用VERSION1.1.26，更新完整中文说明，将已提交的GPT-6.2模型与Release Go模块准备修复纳入同名annotated tag，并推送main及更新后的tag。

## 起点与旧标签

- 本地/远端main为 `9d147ae69f954898453e9748080bc9c2888771c8`，工作区干净。
- 本地/远端v1.1.26均为tag对象 `b88f6896682e4a5cc1bcf448c9f702cfc47d8643`，peeled commit为 `38990cfa716e337740ae8aa3c316940368a3f9ce`。
- 旧对象已保留在本地 `refs/backups/v1.1.26-before-retag-20261119-b88f6896`。原始tag对象和当前发布说明分别冻结为 [tag-before-retag.txt](./tag-before-retag.txt)、[release-notes-before-retag.md](./release-notes-before-retag.md)，原始SHA256见 [before.json](./before.json)。备份引用不推送。
- 相对上一版本v1.1.25（`cb84a2c3c1d3889f4b0f73b368df52f03e0cfa6d`），保留44簇/146任务；本次追加GPT-6.2独立12任务，合计45簇/158任务。
- 旧Release工作流 [44334908913](https://github.com/githubliangliang/sub2api/actions/runs/44334908913) 为failure，Release API仍404。[旧工作流步骤](./previous-release-jobs.json)与[运行状态](./runs-before.json)按实际查询保存。
- 起点Security Scan仍为failure，本次不改安全扫描、依赖或告警规则，也不宣称安全扫描通过。

## 纳入的已提交内容

1. `ad09679034169993f813d39372f564fab533f9c8` 提供GPT-6.2 Sol/Astra规范ID、Codex描述与协议校验、价格资源/双层fallback、严格272k整次计价、Fast2倍、Ultra Fast3→6倍促销时钟、请求/WS每轮PricingAt及独立账号成本，白名单/映射预设/OpenCode配置同步。[12/12验收](../../../openspec/changes/port-upstream-0.2.23-gpt62/verification.md)与[证据](../../upstream-sync/implementation-gpt62/README.md)保留。
2. `9d147ae69f954898453e9748080bc9c2888771c8` 为Release设置Go代理错误时直连回退，并在发布构建前进行最多3次、间隔2秒的模块下载；持续失败终止。对应发布测试覆盖首试成功、失败两次后成功和持续失败。
3. 本次发布准备只修改当前 [v1.1.26详细说明](../v1.1.26.md)、AGENTS当前发布索引，并新增本目录证据。原发布/模型实施资料、版本/schema/依赖/Ent/Wire及资源默认值保持。

## 验证归属

- 复用GPT-6.2已完成的54包unit、四包race44顶层/390子项、清理后服务race34顶层/120子项、实际本地HTTP/WS/SQLite48场景、前端2213通过/2既有跳过与typecheck/lint/build、OpenSpec严格校验。
- 当前28个模型产品/测试文件逐一与已测试 [product-hashes.json](../../upstream-sync/implementation-gpt62/product-hashes.json)匹配；`ad0967903..9d147ae69` 的backend/frontend无差异。
- 本次重新执行发布配置重试控制、标签preflight控制、actionlint、版本/静态embed构建核对、说明范围/链接/原始证据hash及git空白检查。执行结果在提交前写入本目录。
- 价格仍依据冻结PR8906；官网在线核验502/403的既有失败保留，没有新增现网价格或供应商权限结论。
- 前次AWS部署实际使用1.1.26/4d58d9367，仍记录原SHA；本次标签/仓库发布不再次操作该实例。

## 提交与推送顺序

1. 提交完整说明、AGENTS新索引及本目录准备记录，提交内VERSION为1.1.26。
2. 使用完整说明创建annotated v1.1.26，保留旧对象；核对tag正文、peeled commit、精确版本解析与preflight输出。
3. 先推送main，再用明确的旧对象lease更新远端同名tag：`--force-with-lease=refs/tags/v1.1.26:b88f6896682e4a5cc1bcf448c9f702cfc47d8643`。
4. 推送后验证远端main/tag、按新SHA跟踪Release工作流、正文与实际产物。未来提交/tag对象SHA和推送后结果由实际执行产生，不预先写入本提交的准备记录。
