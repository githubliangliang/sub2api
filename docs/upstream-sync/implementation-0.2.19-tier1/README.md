# v0.2.19 第一档实施证据

2026-10-29 用户授权后完成 F01–F07 和26/26任务。实现提交本地 main：`683354b2cdd2461dc581fa14158be09ef51de29b`，父提交 `49939fb2cf96633688548ca39204d96558e59d79`。[OpenSpec验收](../../../openspec/changes/port-upstream-0.2.19-tier1/verification.md) / [行为与反例](./acceptance-matrix.md)。

## 实施范围

31个产品文件：14个生产文件、17个测试文件。提交只包含这些经过验收的代码与测试；预存评估文档和其它提案未混入代码提交。[提交记录](./implementation-commit.json)、[起点](./baseline.json)、[范围与全部产品哈希](./scope-facts.json)。汇总代码 SHA256：`3f97da9098f59ff907f2d5b057fcaa200ab5eca13ca93a6e95d5f1f03e3cf9fd`。

VERSION1.1.20、Go1.27.3、SQLite迁移227、依赖、Ent/Wire及前端保持原值。150个冻结来源/评估/第二档/Terra文件和1416个保护产品文件未改变；旧3732产品中3709个未变，另加8个回归文件。没有新平台/计量模式/schema、支付、发布/tag/push/部署或业务实例数据操作。

## 最终门禁

| 项目 | 实际证据 | 结果 |
|---|---|---|
| 完整 backend unit | [输出](./backend-unit-direct.txt) / [argv、时间和退出码](./backend-unit-direct.json) | PASS：54包 |
| server build | [命令与退出码](./backend-build.json) | PASS |
| golangci-lint 2.14.0 | [输出](./backend-lint-final.txt) / [工具版本](./implementation-tools.txt) | PASS：0 issues |
| 相关 race、取消和活流关闭 | [输出](./backend-race.txt) / [全部包与筛选](./backend-race.json) | PASS：8包实际测试 |
| SQLite迁移/方言/配置quota/0值/软删/固定窗口/usage dedup | [输出](./sqlite-protection.txt) / [HTTP/SSE到实际结算](./cache-hit-sqlite-complete-fixtures.txt) | PASS |
| OpenSpec与完成记录 | [严格校验](./openspec-strict-final.txt) / [完成/链接/提交/冻结范围校验](./implementation-validation-final.txt) | PASS |
| 前端和依赖 | 源码、package/lockfile、go.mod/go.sum未变，见scope-facts | N/A |

## 复现与过程记录

`run-check.py`为每次检查保存完整argv、cwd、UTC时间、耗时、退出码及未过滤stdout/stderr。各项失败与最终通过的对应关系见acceptance-matrix；首次编译/lint或不完整夹具输出不作为完成验收。

最终unit在测试子进程移除六个代理变量，完整命令见backend-unit-direct.json。原Aliyun测试/生产文件没改：SDK对NO_PROXY做含端口精确匹配，[控制本地代理](./captcha-proxy-proof.txt)复现代理HTTP错误被归类为API错误，[移除代理的三轮12项](./captcha-direct-network.txt)即时通过；早先两次完整unit失败及诊断整包的记录仍保留。最初的临时端口推测已被后续环境证据取代。

`claude-js-oracle.js`独立验证12种文本、三个版本的JS码元采样及编码。旧cache-reader和诊断overlay在本目录保存，不向工作树或真实index应用；code/assessment hashes保持可核对。
