# v1.1.19 发布链路本地复核

2026-10-24；本次只新增发布证据，没有改生产、VERSION、共享前端依赖或历史实施证据，没有执行 commit/push/tag。构建快照为实现提交 `2b79ac32078e4ac01e037c7ae1fa432d0d3ce1e0` 加工作树 VERSION `1.1.19`；版本准备提交及最终 tag SHA 由发布记录另行记录。原实施门禁使用 VERSION 1.1.18，保持原记录。

## 已完成

| 检查 | 实际结果 |
|---|---|
| [工具链及声明](./s07-toolchain.txt) | 本机 Node 24.16.0 / pnpm 11.21.0；backend Go 1.27.3；Docker 29.6.2。根及 deploy Dockerfile 均声明 Go 1.27.3、Node 24、pnpm 9、embed；保留原有 pnpm minor 和 Alpine 差异。三份 workflow 的 Go 断言一致。 |
| [发布脚本回归](./s07-release-script-tests.txt) | exit 0。shell 语法、release-tag-preflight-test.sh、release-publish-config-test.sh 全过；临时 Git 仓库验证拒绝 branch/SHA 冒充 tag、正确解引用 tag、GoReleaser 固定目标 tag 及重复发布资产配置。 |
| [VERSION resolver](./s07-version-resolver.txt) | exit 0。独立临时仓库验证无仓库/无精确 tag 时取 VERSION、精确 annotated tag 优先、目标 tag 返回 1.1.19。没有创建工作仓库 tag。 |
| [Release integration](./s07-integration.txt) | `go test -tags=integration ./...` exit 0，141.135s，与 Release 的 make test-integration 命令一致；没有启用 `integration && postgres`。 |
| [Node 20 / pnpm 9 实际构建](./s07-node20-pnpm9-build.txt) | exit 0，153.15s。使用已缓存官方 node:20-alpine 镜像，在独立源码副本执行 Node 20.20.2 / pnpm 9.15.9 的 `install --frozen-lockfile` 和 `pnpm run build`。安装 1m28.6s，Vite 构建 31.79s，输出 135 文件。共享 package.json、lockfile、workspace、node_modules 元数据 SHA 均未变，隔离 lockfile 未变。 |
| [前端构建产物比对](./s07-frontend-artifact-compare.txt) | exit 0。隔离 Node 20 产物与当前实际 embed 使用的 dist 逐文件 SHA-256 完全相同，均 135 文件。 |
| [CGO=0 embed 版本构建](./s07-embed-version.txt) | exit 0，60.369s。Go 1.27.3、linux/amd64、CGO_ENABLED=0、`-tags=embed`、trimpath，注入 Version/Commit/Date/BuildType。实际运行 `--version` 为 `Sub2API 1.1.19 (commit: 2b79ac320, built: 2026-10-23T16:57:55Z)`；二进制 SHA-256 为 `dd835f1c17309d12fd8a1da1ed9523abc8d278aa64c3238fa2cd14404304b763`。 |

同名 JSON 保存命令、cwd、开始/结束时间、耗时及真实退出码。执行脚本副本为 [runner](./s07-run-check.py)、[隔离前端构建](./s07-node20-build.py)、[版本构建](./s07-embed-version.py)。完整 54 包 unit 和本地前端 281 文件全量验收沿用第二档已有证据，本次没有重复运行。

## 非门禁诊断与发布触发

- [go mod tidy -diff](./s07-tidy-diff.txt) exit 1 表示存在整理差异：主要删除未使用的 PostgreSQL 测试模块及旧校验和，并补 xxh3 校验和。命令没有改写 go.mod/go.sum；Release 的 GoReleaser hook 使用普通 `go mod tidy -C backend`，并非 `-diff` 检查。
- [SIMPLE_RELEASE 查询](./s07-simple-release-variable.txt) 没有执行远程查询成功：本机没有 gh 命令。需要由远程设置或实际 Actions 确认仓库变量 `SIMPLE_RELEASE`；tag push 时它决定完整多平台资产或仅 x86_64 GHCR 简化发布。
- push `v*` 会自动触发 release.yml：preflight 解引用 tag，后续 checkout 固定 release_sha；前端使用 Node 20 / pnpm 9；后端使用 Go 1.27.3 跑 unit/integration；随后校验 embed 二进制版本，再由 GoReleaser 发布。annotated tag 的首行后正文成为 Release 说明。
- 本地 node:20-alpine 与 GitHub Ubuntu runner 环境不同。本记录验证本地构建及配置，不表示远端 Actions、跨平台打包、GHCR/DockerHub 上传或 GitHub Release 已经成功。Release 不以独立 security-scan workflow 为 needs；现存依赖告警沿用第二档报告。
- `docs/releases/` 默认被仓库忽略，发布准备需有意 force add 新说明及证据。用户已授权发布；具体 main 提交、annotated tag 与 tag push 由 root 执行并另记结果。
