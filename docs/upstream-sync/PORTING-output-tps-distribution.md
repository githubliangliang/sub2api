# 用户选定：窗口级输出 TPS 分布

2026-11-04 用户明确授权实现历史PR #7911余项；不新增release intake，不回写历史来源/档位。基线v1.1.22、Go1.27.3、SQLite227，[冻结事实](./implementation-output-tps-distribution/baseline.json)。

## 范围与来源
完整来源为上游merge `745ff1d3ca79a7f39520f627f6aa44d778a11580` 第一父到merge的diff。仅选择运维窗口统计；日志TPS已由个人功能完成。第三档现在获用户选择，不改称既有缺陷。行为与验收见[OpenSpec](../../openspec/changes/port-upstream-output-tps-distribution/README.md)。

## patch sites 与适配
- backend/internal/service/ops_dashboard_models.go：可空output_tps、P5/P10/P50/avg与sample_count。
- backend/internal/repository/ops_repo_dashboard.go：基础概览成功后的统一统计调用，覆盖raw/auto/preagg。
- backend/internal/repository/ops_repo_output_tps.go：复用buildUsageWhere和请求类型常量，用SQLite窗口函数及线性插值替代来源PG percentile_cont；2秒查询预算。
- frontend/src/api/admin/ops.ts、views/admin/ops/components/OpsDashboardHeader.vue及中英文ops语言文件：现有页面卡片接线。日志速率helper已有，不重复移植。
- 使用真实SQLite迁移夹具和管理HTTP验证，不引入来源PG integration夹具；不改变写入/计费路径。

## 验收与边界
全部模式按保留明细计算全窗口分布，不平均小时百分位。复用时间/分组/平台归属；排除无输出/耗时、图片、Live及unknown。普通/全屏/移动、刷新/筛选、空值、2秒超时、父取消/截止和查询规模均需验证。

## 旧遗留状态
历次第一/第二档、历史工具/Responses兼容、Terra和Grok诊断已完成归档；窗口TPS余项由本次承接。Claude重置grant、图片计量开关、费用趋势、监控前缀、受限管理员密钥、缓存/COW及新模型/平台仍按需，详见[原复查](./UNMERGED-REVIEW-2026-10-31.md)。支付及PG-only/缺消费者实现排除。VERSION/Go/SQLite227、依赖与Ent/Wire保持；发布/推送/业务实例部署独立。

## 实施与验收
实现已提交`f333aa5f3aa7a6f6ecc4d1620cbfc47a737e16e3`，起点`ed2340cc5a922ced3e2e408465237f72e8a93579`；[实际验收](../../openspec/changes/port-upstream-output-tps-distribution/verification.md)、[检查索引](./implementation-output-tps-distribution/README.md)。全部54包后端unit、embed构建、lint0、三包定向race和SQLite/adminHTTP通过；前端2174通过/2既有跳过、typecheck/lint/build/critical与真实新建本地实例浏览器通过。1000/10000/100000样本本机现代SQLite查询约3.4/33.6/344.6毫秒，不代表VPS性能保证。历史记录不回写；用户随后授权推送main，不创建新发布或业务实例部署。
