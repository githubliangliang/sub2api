# v1.1.19 发布准备验收

2026-10-24（Asia/Shanghai）。实现提交`2b79ac32078e4ac01e037c7ae1fa432d0d3ce1e0`，承接评估`498432759`；前一tag v1.1.18为`873962ce7`。本轮VERSION另升1.1.19并在tag前提交，历史实施验收的1.1.18和原证据不回写。

## 已完成检查

| 检查 | 结果及原始输出 |
|---|---|
| 已验收实现 | 两档F01–F06/S01–S10均在实现提交中；[原始门禁汇总](../../upstream-sync/implementation-0.2.18-tier2/checks.json)的54包unit、build、定向race、真实SQLite及前端2027通过/2既有跳过均已完成。 |
| Release脚本/语法 | [s07-release-script-tests](./s07-release-script-tests.txt)退出0，覆盖发布引用解析、dispatch约束及配置/前置脚本。 |
| 工具链/配置 | [s07-toolchain](./s07-toolchain.txt)退出0；backend Go1.27.3、两份Dockerfile版本和workflow断言一致。CI前端固定Node20/pnpm9，本地通常使用Node24/pnpm11，因此另做隔离兼容构建。 |
| Integration | [s07-integration](./s07-integration.txt)退出0，执行Release同用`make test-integration`；postgres限定文件依构建标签排除，不冒称验证PG。 |
| CI前端兼容 | [Node20.20.2/pnpm9.15.9](./s07-node20-pnpm9-build.txt)隔离frozen install与实际build退出0，共153.15s；锁文件未改，共享node_modules和依赖声明未改。135个产物文件，index哈希与本地embed所用artifact一致。 |
| 1.1.19静态嵌入构建 | [s07-embed-version](./s07-embed-version.txt)退出0；CGO_ENABLED=0、linux/amd64、embed、实际ldflags注入，二进制输出`Sub2API 1.1.19 (commit: 2b79ac320, ...)`；Go1.27.3。二进制只生成在/tmp，没有部署。 |
| VERSION解析 | [s07-version-resolver](./s07-version-resolver.txt)退出0；在隔离git夹具验证未tag/file fallback、精确annotated tag优先及目标1.1.19，未提前修改本仓库tag。 |
| OpenSpec | 两批strict重新通过；发布只增加状态说明，不回写source-baseline、历史命令或代码哈希。 |
| 范围 | [release-facts](./release-facts.json)核对原135文件哈希，当前唯一预期差异为VERSION1.1.19；go.mod/sum、前端依赖、Ent及迁移均与实现提交相同。 |

## 保留的非零诊断与限制

- [go mod tidy -diff](./s07-tidy-diff.txt)退出1，记录现有依赖整理差异；未运行会改写文件的tidy，也未为该诊断额外升级依赖。GoReleaser before hook使用普通tidy；本地没有声称完整跨平台GoReleaser已执行。
- 原始测试输出包含gotest/testify等工具产生的空白和缩进。最初`git diff --check`只检查已跟踪改动，首次纳入原始日志后会报告这些格式；为保留证据没有改写日志。排除两份implementation目录的原始*.txt后，源码、配置与Markdown的差异格式检查通过；此边界不能写成“所有原始日志无格式告警”。
- 已知安全剩余：pnpm全树75、生产17（xlsx两条high），Go包级18 findings/13 OSV、stdlib0；详[依赖报告](../../upstream-sync/implementation-0.2.18-tier2/dependency-review.md)。不将非零安全检查或未知远端利用性写成通过。
- implementation evidence里的HEAD/VERSION/未提交标记是当时快照。此发布提交不修改冻结来源、评估证据、原始结果或旧hash来伪装新版本重测。
- `docs/releases/`被现有ignore规则排除，本次会显式force-add本轮说明及证据，避免文件只留本地。

## 执行顺序

1. 实现已commit为`2b79ac320`，再commit VERSION1.1.19及本说明/发布证据。
2. `git push origin main`成功后，再创建详细中文annotated tag `v1.1.19`，其目标commit须携带VERSION1.1.19。
3. 单独`git push origin v1.1.19`，核对远端main、tag对象及解引用commit。

以上后续push/tag的实际状态以远端ref和客户端执行记录为准，本文件记录的是提交前的发布准备；tag触发的Actions、跨平台构建、容器上传与GitHub Release结果不由本地检查代替。未部署运行中的本地或远端业务实例。
