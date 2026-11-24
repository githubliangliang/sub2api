# v1.1.28 本地发布门禁（2026-11-24）

本目录由独立检查任务生成，只写发布证据，不修改产品、前端、版本、工作流、标签或业务实例。执行起点为 `9fe9ef3d6329a109923277a27c54fccbaa229f87`；父任务负责 VERSION、详细发布说明及后续提交/tag/push。

`run-check.py` 保存实际命令、cwd、`CI=true` / `TZ=UTC` / `GOTOOLCHAIN=go1.27.4`、去除的代理变量、退出码、计数、耗时及完整输出哈希。较大日志无损 gzip，原始字节数和SHA256记录于JSON中，不以包数代替测试数。

| 检查 | 实际结果 |
|---|---|
| [环境和输入](initial.json) | backend Go1.27.4，Docker29.6.2；本地已经存在redis:8.4-alpine镜像，记录实际image ID，无需拉取。记录4,089个backend/frontend/deploy/workflow输入SHA256。 |
| [完整integration](full-integration.json) / [完整输出](full-integration.log.gz) | `go test -json -tags=integration -count=1 -p 2 ./...`退出0，239.658秒；49个有测试包通过，6,180顶层/12,504含子例通过；13顶层/14含子例跳过，另58无测试包。禁用结果缓存，使用真实本地Redis容器及既有SQLite/miniredis夹具。 |
| [标签预检隔离控制](release-tag-preflight.json) / [log](release-tag-preflight.log) | 退出0；实际执行`deploy/tests/release-tag-preflight-test.sh`，拒绝分支/SHA，接受轻量及annotated标签。这是临时仓库控制，不冒充本版实际tag预检。 |
| [发布配置和下载重试](release-publish-config.json) / [log](release-publish-config.log) | 退出0；实际执行`deploy/tests/release-publish-config-test.sh`，包含Go模块下载重试次数及终止控制。 |
| [工作流actionlint](actionlint.json) / [log](actionlint.log) | Release及backend CI两个工作流退出0，无输出。 |
| [发布shell语法](shell-syntax.json) / [log](shell-syntax.log) | 发布tag解析、preflight/publish测试及backend版本解析脚本的`bash -n`退出0。 |
| [输入及输出保护](source-verification.json) | 上述4,089个输入只允许父任务将VERSION更新为1.1.28，产品/测试/工作流输入保持；完整integration解压后SHA256匹配。最终实际差异与检查时间记录于JSON。 |

完整integration中真实Redis的三个调度缓存控制均实际通过：`TestSchedulerCacheSnapshotUsesSlimMetadataButKeepsFullAccount`、`TestSchedulerCacheRetireAndReopenFencesOldEpochIntegration`、`TestSchedulerCacheGroupLifecycleLeaseOwnerAndTTLIntegration`。Redis容器由已有testcontainers测试创建及清理；没有操作业务实例。`source-verification.json`列出的其他pass名称只是检索索引，名称含Redis不代表每一项都使用真实Redis。

[build-tag清单](integration-build-tags.json)记录72份带integration约束文件，其中56份另需postgres标签。本命令没有启用postgres，不构成PostgreSQL验收；没有实际外部供应商请求或生产容量验收。本次没有重跑历史unit/race/浏览器/性能，其复用边界由父任务readiness中的匹配哈希单独说明。

[初始读取说明](initial-inspection-note.json)保留一次cwd路径误用：从backend目录读取根目录相对docs/deploy路径报不存在；Docker/Go环境读取成功，随后从根目录重读。它未执行门禁或修改产品，以上实际门禁均通过。
