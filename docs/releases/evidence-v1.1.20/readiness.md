# v1.1.20 发布准备验收

2026-10-26（Asia/Shanghai）。用户要求提交全部改动、创建详细中文 annotated tag `v1.1.20`，再推送。实现已提交为`0e38d06d82557a0a3b372ca9165a24b9b3acbc93`，共225个文件；前一tag v1.1.19为`0c62a424bc261eda9365074a3e92444018b770bc`。VERSION以独立发布提交从1.1.19改为1.1.20，先于tag，不修改历史实现证据。

## 产品与发布前检查

| 检查 | 实际命令与结果 |
|---|---|
| 完成产品核对 | [release-facts.json](./release-facts.json)校验72个产品文件，与实现提交及最终成功门禁哈希逐项一致；实现后产品唯一差异为VERSION1.1.20。冻结source-baseline及原始候选证据哈希保持一致。 |
| 已完成产品门禁 | 复用[原验收报告](../../../openspec/changes/port-upstream-0.2.18-personal-features/verification.md)的54包unit/build、五包定向race、真实SQLite/miniredis，以及285文件2089通过/2项既有跳过、typecheck/lint/build。没有把这些结果写成VERSION1.1.20下重新全量执行。 |
| Release脚本 | root：`rtk proxy bash -n deploy/resolve-release-tag.sh backend/scripts/resolve-version.sh`、`rtk proxy bash deploy/tests/release-tag-preflight-test.sh`、`rtk proxy bash deploy/tests/release-publish-config-test.sh`均通过；[命令记录](./release-script-tests.json)、[输出](./release-script-tests.txt)。 |
| Release同用integration | backend：`rtk proxy make test-integration`退出0，160.665秒，包含Go正常复用的缓存结果；[命令记录](./integration.json)、[输出](./integration.txt)。`integration && postgres`文件按标签排除，不算PG验证。 |
| 1.1.20静态嵌入构建 | CGO_ENABLED=0、linux/amd64，`go build -tags=embed -trimpath`并注入Version/Commit/Date/BuildType；实际二进制输出`Sub2API 1.1.20 (commit: 0e38d06d82557a0a3b372ca9165a24b9b3acbc93, built: 2026-10-26)`，退出0；[完整命令](./embed-version.json)、[输出](./embed-version.txt)。二进制仅生成在/tmp，没有部署。 |
| VERSION解析 | 当前未tag checkout读取1.1.20；隔离Git夹具校验文件fallback、annotated tag优先及下一提交fallback，退出0；[命令](./version-resolver.json)、[输出](./version-resolver.txt)。可在仓库根目录以`rtk proxy python3 docs/releases/evidence-v1.1.20/version-resolver-test.py`复跑[同一脚本](./version-resolver-test.py)。 |
| OpenSpec | `rtk pnpm dlx @fission-ai/openspec@1.14.3 validate port-upstream-0.2.18-personal-features --strict --json --no-interactive`退出0，issues为空；[命令记录](./openspec-strict.json)、[输出](./openspec-strict.txt)。36项任务、22条需求和50个场景已验收；本次不归档change。 |
| 范围 | VERSION1.1.20，Go1.27.3、最大SQLite迁移226；无依赖/锁文件、Ent/Wire、CI变更，没有新平台、支付或运维TPS分位数。当前AGENTS/索引及实施报告补记实际实现提交与独立发布准备，原始状态JSON仍保留提交前快照。 |

## 证据与运行边界

- 原始TSV、测试日志和TXT保留工具产生的空白及失败诊断。全部225个文件纳入实施提交时，原始证据有4194行差异格式诊断；仅排除这些原始TSV/log/TXT后，源码、配置与Markdown的`git diff --cached --check`退出0。没有改写冻结证据，也没有宣称所有原始日志都通过格式检查。
- 本轮发布证据的完整staged格式检查仅报告`embed-version.txt:3: new blank line at EOF`，为二进制版本输出末尾原始空行，保留原样。发布提交的`git diff --cached --check -- . ':(exclude)docs/releases/evidence-v1.1.20/embed-version.txt'`退出0；不将原始输出格式诊断当成源码或产品失败。
- `.gitignore`忽略`docs/releases/`，本次显式force-add v1.1.20说明和此目录，避免发布说明只存在本地。其他被忽略的运行配置、依赖目录和构建产物不纳入。
- 本轮没有重新跑隔离Node20/pnpm9安装构建或依赖安全审计；依赖与锁文件同已发布v1.1.19，前端产品构建见本轮原验收。Release工作流会自行运行固定Node20/pnpm9的frozen install/build；以前的剩余审计记录不因本次发布而清零。
- GoReleaser、跨平台打包、镜像上传与GitHub Release尚需远端工作流实际执行，本地检查不等于这些工作已完成。没有调用真实收费搜索，也没有部署本地或远端业务实例。

## 提交、标签与推送

1. 实现和原始证据已提交为`0e38d06d82557a0a3b372ca9165a24b9b3acbc93`。
2. 单独提交VERSION1.1.20、[详细发布说明](../v1.1.20.md)、此检查目录及当前状态补记。
3. 创建annotated tag `v1.1.20`，使用完整中文说明，核对tag目标的VERSION为1.1.20。
4. 先`rtk git push origin main`，成功后`rtk git push origin v1.1.20`。最终核对远端main、tag对象及解引用commit与本地一致。

本文件记录发布准备。后续实际push/tag结果由终端返回和远端ref核实；不通过回写原始测试结果伪造新版本全量重测。
