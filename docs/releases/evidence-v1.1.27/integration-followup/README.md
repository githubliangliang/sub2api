# v1.1.27 集成门禁跟进（2026-11-21）

首次准备提交为`8bf50be352c2a7f2ec57b140738dcbad5c189d3c`，tag对象`ba1fbede506606a3b7a15fc1e97c40af439fe652`。首轮[Release 44894814385](https://github.com/githubliangliang/sub2api/actions/runs/44894814385)通过标签预检、前端构建和后端unit，integration失败，发布job跳过，Release API仍404；[状态](./initial-publication.json)记录实际查询及匿名日志下载403边界。未从公开的退出码推断具体远端断言；根因核验使用下面的独立本地复现。

## 根因和唯一测试变化

本地使用CI=true、TZ=UTC、取消代理变量、Go1.27.4、`go test -json -tags=integration -count=1 -p 2 ./...`完整重现：只有`TestSchedulerCacheSnapshotUsesSlimMetadataButKeepsFullAccount`失败，其旧断言要求`full.AccountGroups[0].Group`非空。v0.2.24 S13已明确只向缓存写入成员标量，展开关系由调用者持有；单元测试覆盖该合同，但integration旧断言漏改。[完整运行统计](./integration-before.json)与[原样失败输出JSON摘录](./integration-before-excerpt.json)保留。全量原始JSON流留在本机临时发布目录，摘录不冒充全量日志。

只修改`backend/internal/repository/scheduler_cache_integration_test.go`：保留原有完整凭据和大字段检查，增加GroupIDs、AccountID、GroupID、Priority保持的断言，要求缓存Group为nil，同时确认调用者Group仍非nil。没有修改产品投影，也没有删除测试或放宽配置/成员保留要求。

## 修正后验证

- [真实Redis三项定向控制](../scheduler-cache-integration.json)及[日志](../scheduler-cache-integration.log)通过，覆盖完整缓存/快照、退休/重开代际隔离及锁owner/TTL。
- [全量integration](./integration-after.json)退出0，约206.601秒：49个有测试包，6163顶层/12395含子例通过；13顶层/14含子例测试跳过，另58个无测试包。命令禁用结果缓存；统计从原始JSON流单独解析，摘要保留原流SHA256，不把包数计入测试数。
- [最终4032文件核验](./source-verification.json)只有两项有意差异：VERSION1.1.27，以及该integration测试断言。业务源码、unit构建输入、前端和Wire没有新增漂移；原unit/race/浏览器/性能和发布构建检查按相同输入复用。本次不为纯integration测试变更重跑这些无变化门禁。
- gofmt差异为空；提交前再核对Git空白、完整中文说明链接、VERSION、既有证据保护及暂存范围。

## 标签更正边界

用户本轮授权发布v1.1.27并push；这里继续完成同一版的发布门禁。初轮尚未生成Release或资产，旧tag对象已保存到本地`refs/backups/v1.1.27-before-integration-fix-20261121`，不推送备份引用。正常提交/推送main后，以`--force-with-lease=refs/tags/v1.1.27:ba1fbede506606a3b7a15fc1e97c40af439fe652`更正本轮新建标签，并以完整更新说明创建annotated对象。远端引用若已被其他操作修改，lease会拒绝覆盖。

先前v1.1.26标签、冻结来源和实施证据不改。最终发布提交/tag对象及新工作流结果在执行后核对；本目录不预填未来SHA或成功状态，不部署或操作业务实例。
