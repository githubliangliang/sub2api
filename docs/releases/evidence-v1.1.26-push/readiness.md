# v1.1.26 推送准备与补充验证

2026-11-18，用户明确请求「打tag v1.1.26,详细描述，push」。本次提交已完成的二进制发布工具与 CI 修复，补齐中文 tag 说明，再推送 main 和 annotated tag。VERSION 已为1.1.26，最终 tag 提交携带相同文件值。

## 标签与提交范围

- 上一正式 tag 为 `v1.1.25`，commit `cb84a2c3c1d3889f4b0f73b368df52f03e0cfa6d`。
- 开工前 main/远端 main 均为 `4d58d9367d93b432104e5293ddc7cb4e8ab0efff`。本地已有 v1.1.26，object `f19a96b08df4943b59ed3667110ebf73627bd4f8`，指向同一提交；远端 tag 不存在，Release API 返回404。
- 原本地 tag 对象保留在 `refs/backups/v1.1.26-before-push`，不推送该备用引用。其原始正文和发布说明分别保存为 [tag-before-push.txt](tag-before-push.txt)、[release-notes-before-push.md](release-notes-before-push.md)。更新本地 tag 后正常创建远端 v1.1.26，没有覆盖已发布的远端 tag。
- 新增工具提交 `aab5591c700f1045e4452d9b2b626e0930e45a88`，CI 修复提交 `19edd5698d22717eb06cfd547ff40c694a65b446`。最终发布准备只更新当前说明/AGENTS 索引并新增本目录证据，原始一期发布证据保留。
- 原始44簇/146任务的上游产品增量仍由 `d8d126da9`、`3652ddb6a` 完成；当前 backend/frontend/resources 与最初 v1.1.26 tag 逐文件无差异。Go1.27.4、SQLite229、auth21、库版本和 Ent/Wire 保留。

[publication-facts.json](publication-facts.json) 保存完整SHA、范围与日志哈希；[before-hashes.json](before-hashes.json) 保存提交工具/CI修复后的8194个跟踪文件哈希，供最终发布准备检查。原 v1.1.26 [首次准备记录](../evidence-v1.1.26/readiness.md) 和原校验清单保持当时内容。

## CI 修复与验证

用户提供的失败来自 [CI job 134819337361](https://github.com/githubliangliang/sub2api/actions/runs/44309411651/job/134819337361)：`TestRateLimiterSetsTTLAndDoesNotRefresh` 创建 Redis 容器时返回 `No such image: redis:8.4-alpine`。原 job 的 Go版本检查和unit通过，integration失败；日志接口要求仓库权限，失败段由用户提供。

该镜像仍在 [Docker官方镜像清单](https://github.com/docker-library/official-images/blob/master/library/redis) 中，且实际 Docker CLI 拉取成功。当前 testcontainers v0.40.0 仅把拉取流读到 EOF，可能忽略HTTP200流中的JSON错误。[控制实验](pull-stream-probe.go) 同时证明：这类错误可被SDK丢弃，随后创建容器得到 No such image；Docker CLI 对同一错误返回失败。[输出](pull-stream-probe.log) 保留原结果。

CI和Release在integration前复用已有镜像，缺失则显式拉取并检查，最多3次、间隔2秒；准备失败停止该步骤。准备代码内置于workflow，因此按旧tag checkout时也不依赖新tag才有的辅助脚本。

- [5组准备控制](prepare-redis-tests.json)：缓存命中、冷启动成功、首次拉取失败后重试、持续失败、拉取成功但镜像仍缺失分别验证；失败保持非零退出，两处workflow顺序/内容/Redis版本一致。
- actionlint、YAML解析、Bash语法与diff空白检查通过。核心测试和Redis/testcontainers版本未修改。
- [最终integration命令/结果](integration-final.json)：CI=true、TZ=UTC，顶层命令传入readonly/count1/json；48包通过，153.528秒，两个真实Redis TTL测试及限流邻近测试通过。
- [完整JSON输出](integration-final.jsonl.gz) 用gzip无损保存，原始SHA256见publication-facts；这次实际使用Docker Redis与临时SQLite，不以跳过Docker测试冒充通过。
- [验证配置失败](integration-after.json) 及 [对应完整输出](integration-after.jsonl.gz) 保留：初轮误将-json设入GOFLAGS，导致Ent子命令go list -f与-json冲突。纠正为顶层传参后重新完成48包检查，没有修改schema来消除这个验证错误。

开工前旧提交的CI与Security Scan运行均标为failure：[runs.json](runs.json)。本批修复指定CI的Redis准备，不修改安全扫描或依赖；这里只记录安全运行状态，不推断其具体告警或宣称安全扫描通过。推送后按新提交分别检查实际运行状态。

## 二进制工具与先前实例验证

`deploy/deploy-binary-remote.sh` 固定目标 `admin@18.226.163.236:/home/admin/sub2api`，上传仓库根目录现有二进制，检查可执行版本，备份/原子替换、重启sub2api并检查health，失败回滚。默认私钥 `~/.ssh/test.pem`，可用-i/SSH_KEY覆盖；HEALTH_TIMEOUT默认60秒，--dry-run不连接服务器。

该文件的15组隔离上传、权限、输入、重启/健康/回滚控制已通过。随后实际运行：[发布命令与结果](publish-restart.json)，95.534秒通过；[实际实例核对](publish-restart-verification.json) 验证PID改变、运行文件SHA256与上传文件一致、健康200、配置/unit保留。该操作使用1.1.26/4d58d9367，未把先前实例改记为尚未产生的新tag SHA。本次仓库推送不再次操作实例。

## 发布前与推送后契约

1. 完整中文说明以 [v1.1.26.md](../v1.1.26.md) 为准，保留第一/第二档来源、实际提交、146任务、升级变化、验证边界、未合入能力，并增加工具和CI增量。
2. 当前产品源码与原tag一致，复用已完成的54包unit、48包integration、前端2212/2与critical179/2、浏览器/race/codegen等实施验证；没有把它们描述为本次重跑。新工具/CI变化按以上补充检查接受。
3. 发布前核对VERSION/resolver、静态embed构建与实际二进制版本、发布脚本/标签preflight、原始证据哈希和最终改动范围。新增记录随发布准备提交保存。
4. 在包含VERSION1.1.26的最终提交创建annotated tag，正文与完整说明一致；核对tag对象、peeled commit、精确版本解析和preflight输出后，先 `git push origin main`，再 `git push origin v1.1.26`。
5. 新提交/tag对象的SHA产生于提交后，推送结果和远端工作流状态产生于推送后，在工具输出与本地推送记录验证，不预先伪造写入自身提交。对Release/GHCR资产只按实际结果报告。

本目录与完整发布说明显式force-add，沿用现有gitignore；未修改历史intake/implementation或旧发布证据。[SHA256SUMS](SHA256SUMS) 在提交前覆盖本目录除清单自身外的所有记录。
