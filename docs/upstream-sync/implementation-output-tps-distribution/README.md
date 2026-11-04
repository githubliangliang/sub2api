# 窗口级输出 TPS 分布实施证据

2026-11-04 用户选定，范围见[补充报告](../PORTING-output-tps-distribution.md)、[OpenSpec](../../../openspec/changes/port-upstream-output-tps-distribution/README.md)。

[baseline.json](baseline.json)记录原始HEAD、完整PR来源哈希、版本和受保护历史树摘要。后续检查记录实际命令、结果与局限；合成浏览器/API夹具与真实SQLite/管理HTTP结果分别说明。

当前工作树已完成：[verification](../../../openspec/changes/port-upstream-output-tps-distribution/verification.md)汇总行为、门禁、过程失败和实际局限；[scope-check.json](scope-check.json)记录最终产品哈希、历史保护与链接检查。尚未创建实现提交、发布或业务实例部署。

| 证据 | 结果 |
|---|---|
| [SQLite定向](backend-targeted-accepted.log) | 连续分位数、样本/筛选/模式、清理与实际执行中断、2秒自身预算/父取消截止 |
| [HTTP](backend-http-final.log) | 真实service/repository/迁移SQLite，overview及snapshot-v2×三模式；handler测试不替代认证验证 |
| [全量backend](backend-unit-final.log)、[race](backend-race.log)、[build](backend-embed.log)、[lint](backend-lint.log) | 54包unit通过；三包定向race含SQLite227/quota/固定窗口/dedup；embed构建通过、0 issues |
| [全量frontend](frontend-unit.log)、[typecheck](frontend-typecheck.log)、[lint](frontend-lint.log)、[build](frontend-build.log)、[critical](frontend-critical.log) | 291文件2174通过/2既有跳过，其它门禁通过；critical176通过/2同一跳过 |
| [浏览器命令](browser/commands.json)、[夹具说明](browser/fixture.json) | 新建临时SQLite实例，真实构建二进制、登录、管理API和SPA；全部账号/用量为合成数据；桌面/全屏/390px、OpenAI/Claude/Grok/空样本、分组和刷新均验证 |
| [桌面](browser/desktop.png)、[全屏](browser/fullscreen.png)、[移动](browser/mobile.png) | 实际页面截图，新卡片在390px视口内换行，无水平溢出 |
| [Go/modernc性能](backend-query-cost.log)、[补充计划](query-plan-details.json) | 1000/10000/100000样本约3.4/33.6/344.6ms；Python SQLite额外EXPLAIN确认现有索引及临时排序 |
| [OpenSpec](openspec-strict.log)、[范围](scope-check.json) | strict通过；完整PR/source-baseline/受保护3077文件及依赖/迁移/版本/生成代码保持 |

每个run-check结果附同名JSON，保留完整命令、执行时间与退出码。backend-targeted.log、backend-targeted-final.log、backend-http.log、backend-unit.log记录新增测试夹具失败，不作为最终验收；修正原因见verification。最终完整Go运行允许复用通过的包缓存，admin和repository重新执行。首次浏览器导航/gate初始化与URL防抖交错也保留命令记录；稳定状态复测通过。

性能来自本机i9-13900H，不承诺低配VPS延迟。只统计保留明细，大窗口超过统计查询自身2秒预算时显示不可用；调用方取消/截止仍传播。测试结束关闭自建浏览器和临时实例，业务实例未操作。
