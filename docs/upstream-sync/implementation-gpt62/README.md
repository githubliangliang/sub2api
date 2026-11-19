# GPT-6.2 实施证据

用户于 2026-11-18 选定模型、价格、272k 计费与 Ultra Fast 时钟；本地实施验收跨至 2026-11-19。入口 `38990cfa716e337740ae8aa3c316940368a3f9ce`。完整 [范围](../PORTING-gpt62.md)、[冻结来源](../../../openspec/changes/port-upstream-0.2.23-gpt62/source-baseline.md) 和 [验收](../../../openspec/changes/port-upstream-0.2.23-gpt62/verification.md)。

## 失败与恢复

- 首个价格/身份探针编译成功，资源路径将两个模型误归 GPT-5.4，双层 fallback 缺价，默认目录缺模型。
- 静态价卡完成后，促销探针复现 Ultra Fast 仍为通用 2 倍，随后接入按 PricingAt 的 3/6 倍策略。
- 未指定/未知 6.2 型号探针复现 GPT-5.4 猜价，限定 6.2 fallback 后拒绝未知名，显式价格对照继续生效。
- 新目录夹具使用 int 比较既有 int64 字段，改为数值断言；产品字段和窗口数值未修改。
- 首次 pnpm 命令因 shell PATH 缺 Node 工具目录失败，使用已安装 Node 24.16.0/pnpm 运行检查，无依赖安装或升级。
- 官方价格检索遇到工具 502，官方 pricing 页面为 403。价格按冻结 PR8906 实施，在线官网核验未完成。

## 验收记录

已完成12/12任务。[checks.json](./checks.json)记录实际结果，原始Go JSON日志使用gzip保留。全量unit54个测试包通过，四包race44顶层/390子项通过；lint清理后服务race34顶层/120子项通过。Go1.27.4编译的lint2.14.0为0 issues，最终embed构建及1.1.26版本输出通过。前端45定向、293文件/2213测试通过，保留2项既有跳过，typecheck/lint/build通过。

HTTP/WS/SQLite是实际本地transport、Ent/modernc迁移与扣费，共48场景；token、账号及未来截止时刻为合成夹具。独立账号成本另用服务fixture验证分组价格覆盖及显式0，不将夹具价格当作官方现网消费。

全量unit与四包race完成后，lint指出旧私有计费包装已无生产调用。删除包装，既有marker测试改用等价公共入口，并完成后续服务race、lint与embed构建；[原失败](./lint-unused-failure.log)和[后续检查](./after-lint-cleanup.jsonl.gz)保留。最终前端模型顺序保留原OpenCode首项，并重新完成全量测试、lint与构建。

最终[文件哈希](./product-hashes.json)、[范围快照](./scope-validation.json)及[复现命令](./commands.md)可供复核。VERSION1.1.26、SQLite229、Go1.27.4、auth21、依赖、Ent/Wire及资源默认值保持。历史来源、旧活动change和发布证据保留原文。

共享工作区出现发布流程 `.github/workflows/release.yml` 及 `deploy/tests/release-publish-config-test.sh` 的并行修改，本任务未编辑这两份文件，范围快照记录它们。模型实现随本次提交保存，等待归档；未推送/发布/部署。
