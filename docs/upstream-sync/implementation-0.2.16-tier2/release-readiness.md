# GitHub Release 补充复核

2026-10-18，承接用户“github release，也没问题吗”的发布链路检查。以下保留复核时快照：当时未提交、打 tag、触发 Actions 或发布镜像。

后续发布准备：用户随后要求提交并发布 v1.1.18。实现已提交为 `847db09e1` + `5ff6e5bf9` 并快进合入 main，VERSION 更新为 1.1.18；[详细发布说明](./release-v1.1.18.md)。下文环境与测试输出继续记录实际验收时的 1.1.17，远端流水线结果以 GitHub 为准。

## 已验证

- release-tag-preflight-test.sh、release-publish-config-test.sh 通过：发布引用解析、GoReleaser tag 固定与重复上传配置正确。
- Release 的 `make test-integration` 通过。[输出](./release-integration.txt)
- 在隔离源码副本和官方 `node:20-alpine` 容器中，以 **Node 20.20.2 / pnpm 9.15.9** 执行 `install --frozen-lockfile` 与完整 `pnpm run build` 通过。[输出](./release-node20-pnpm9-build.txt) Actions 使用 Ubuntu，未将该容器验证声称为实际 Actions 运行。
- `CGO_ENABLED=0 go build -tags=embed -ldflags='-s -w -X main.Version=1.1.17 -X main.BuildType=release' -trimpath` 通过；`--version` 返回 `Sub2API 1.1.17 (commit: unknown, built: unknown)`，版本注入有效。
- 原后端 unit/build/race/SQLite 和前端全量验收仍见 verification。

## 发布边界

- 当前 Release 的 needs 只有 preflight、build-frontend、verify-backend，**不依赖 security-scan 工作流**。xlsx 两条 high 会使安全扫描失败，但当前 release.yml 未将它设置为发布依赖；仓库规则对合入分支的约束未作远程验证。
- 当前改动仍在工作区，现有 `v1.1.17` 标签指向旧提交。发布本次代码需要先提交/合入 main，按仓库顺序更新并提交 VERSION（下一版本可用 1.1.18），创建 annotated tag，再推送 main 和 tag。
- 未执行完整 GoReleaser 五目标跨平台打包、Docker/GHCR 推送、GitHub Release 上传；令牌权限、网络与远程运行结果仍须实际流水线确认。
- `go mod tidy -diff` 提示已有依赖清单整理差异（主要是不用的 PostgreSQL 测试模块及校验和），[差异](./release-tidy-diff.txt)保留。这个只读命令的非零表示“存在差异”，不是当前工作流的失败门禁；Release hook 使用普通 `go mod tidy`。本次没有改写 go.mod/go.sum。

## 复核时修复的测试残留

原内存 SQLite 测试 DSN 直接使用 t.Name()，空名称子测试的 `#00` 可被当成 URI fragment，丢掉 `mode=memory` 并留下无扩展名数据库文件。测试辅助函数现对名称做 url.PathEscape，相关删除/事务测试经 race 通过，残留文件已删除并从代码清单排除。[回归输出](./release-sqlite-test-fixture.txt)

隔离构建副本与一次性容器用于本地验证，不操作部署数据或业务实例。
