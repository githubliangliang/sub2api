# 实施验收（2026-10-30）

状态：本轮用户要求实施选定第三档 Terra，14/14 任务完成，产品/测试已提交本地 main，等待归档。[完整实施证据](../../../docs/upstream-sync/implementation-0.2.19-terra/README.md)。[原验收槽](../../../docs/upstream-sync/implementation-0.2.19-terra/verification-original.md)保留，source-baseline/source-feature-map/proposal/design/specs 未改。

| 实施信息 | 值 |
|---|---|
| 实施起点HEAD / 工作树 | b9e51e27366e8eab20ee2c2746f77065c8ee02b6，干净本地 main，起点领先 origin/main 2 提交；与冻结评估起点49939fb2c分开记录 |
| 实现提交 / 代码SHA256 | a541d0b611f4c2da5cab9811c3c4e02ae8344136 / a97a658cea9ce28d500c61b77f0407734d37d5377ad5b23ce84dfb96810b75b2；[21文件逐项哈希](../../../docs/upstream-sync/implementation-0.2.19-terra/scope-facts.json) |
| 实施VERSION / Go / 迁移 | 1.1.20 / 1.27.3 / SQLite227，均保持 |
| 实际范围与排除项 | T01身份/effort/工具协议/metadata/既有计价和客户端配置；无新增schema/迁移/依赖/平台/菜单/全局默认，其他新能力及支付排除；未发布/push/部署或操作业务实例 |

## 行为与反例

| 行为 | 必须验收 | 新行为失败证据 | 完成证据 | 结果 |
|---|---|---|---|---|
| 身份/effort | 六档/default medium、none/max、拼写/未知后缀、Sol不回退 | [旧产品独立基线](../../../docs/upstream-sync/implementation-0.2.19-terra/terra-final-baseline-before.txt) Identity/Reasoning/ConfiguredManifest失败 | [实现通过](../../../docs/upstream-sync/implementation-0.2.19-terra/terra-final-target.txt)，Sol邻近/全量/race；未加入接受minimal的通用prefix表 | PASS |
| 工具协议 | 三个only-Chat出口出站前400、Responses工具及无工具raw正例 | 同一基线的OnlyChatRejects/ResponsesTools，未拒绝或max丢失 | Chat/legacy functions/Messages/Responses六档HTTP/SSE实际转发，三工具出口400/零上游调用；三普通文本出口保留raw和扩展；WS handler另验 | PASS |
| 目录/metadata | 账号声明优先、空catalog不增项、原目录/ETag不变 | 新默认descriptor在旧基线失败；已有目录保护场景在旧基线通过，ETag夹具错误另记 | 官方/自定义text-only/high-only/100k优先，raw source/unknown/cache/ETag保留；空原生manifest为空，普通sync保留既有错误和空结果 | PASS |
| price/usage | 资源/两层fallback、显式0/override/priority/272k边界、异步usage/SQLite dedup | 基线专用fallback/资源/阈值/显式0失败；[最终handler before](../../../docs/upstream-sync/implementation-0.2.19-terra/terra-handler-final-before.txt)失败 | [资源同源与其它205条不变](../../../docs/upstream-sync/implementation-0.2.19-terra/resource-scope.json)，SQLite priority实际0.00187一次结算；取消/Gin重用保留模型与token桶，重放不扣第二次；quota/固定窗口保护回归 | PASS |
| 客户端 | 白名单/映射预设/可解析生成配置、六档/context/output、旧默认和模型 | [同一39项旧产品2失败](../../../docs/upstream-sync/implementation-0.2.19-terra/terra-baseline-frontend-fixture-fixed-before.txt) | [39项通过](../../../docs/upstream-sync/implementation-0.2.19-terra/terra-frontend-after.txt)，[真实浏览器保存/重开/复制](../../../docs/upstream-sync/implementation-0.2.19-terra/browser/README.md) | PASS |

## 组合检查

| 检查 | 命令 / 证据 | 结果 |
|---|---|---|
| backend完整unit / build / lint | [54包完整unit](../../../docs/upstream-sync/implementation-0.2.19-terra/backend-unit.json)，夹具调整后[handler全量](../../../docs/upstream-sync/implementation-0.2.19-terra/backend-handler-unit-final.json)，[CGO=0/embed构建](../../../docs/upstream-sync/implementation-0.2.19-terra/backend-build.json)，[最终lint](../../../docs/upstream-sync/implementation-0.2.19-terra/backend-lint-final.txt) | PASS / 0 issues |
| 协议/usage/race与取消边界 | service/handler/repository/openai/apicompat [五包最终race](../../../docs/upstream-sync/implementation-0.2.19-terra/backend-race-final.json)；16顶层/197含子场景专项，最终SQLite夹具before/after单列 | PASS |
| SQLite迁移227 / quota / dedup | [三包命令](../../../docs/upstream-sync/implementation-0.2.19-terra/sqlite-protection.json)，[真实Terra handler 12场景](../../../docs/upstream-sync/implementation-0.2.19-terra/terra-handler-final-after.txt)及既有soft-delete/不复活/固定窗口用例 | PASS |
| frontend typecheck / lint / tests / build / critical | [类型](../../../docs/upstream-sync/implementation-0.2.19-terra/frontend-typecheck.json) / [lint](../../../docs/upstream-sync/implementation-0.2.19-terra/frontend-lint.json) / [2106通过/2既有跳过](../../../docs/upstream-sync/implementation-0.2.19-terra/frontend-tests.json) / [build](../../../docs/upstream-sync/implementation-0.2.19-terra/frontend-build.json) / [176 critical](../../../docs/upstream-sync/implementation-0.2.19-terra/frontend-critical.json) | PASS |
| 浏览器模型选择 / 客户端生成配置 | 当前实际页面加独立API夹具：白名单保存/重开、Terra预设实际PUT、OpenCode解析和原生clipboard写入promise完成；[截图/边界/清理](../../../docs/upstream-sync/implementation-0.2.19-terra/browser/README.md) | PASS |
| OpenSpec严格校验 / source哈希 / 范围 | [strict](../../../docs/upstream-sync/implementation-0.2.19-terra/openspec-strict.json)，[起点/源patch SHA](../../../docs/upstream-sync/implementation-0.2.19-terra/baseline.json)，[scope-check](../../../docs/upstream-sync/implementation-0.2.19-terra/scope-check.py)及[冻结哈希](../../../docs/upstream-sync/implementation-0.2.19-terra/scope-facts.json) | PASS |

- [x] 新功能有前后证据、反例与真实消费者验收。
- [x] 适用门禁与冻结范围核对完成，实际SHA/起点写入。
- [x] baseline未冒充实施证据，没有未授权支付、发布或部署。

## 实施边界与过程归因

公共created为来源占位值、价格/能力为冻结#8364值；未当作实时官方售价、发布日期或任意账号可用性证明。普通空目录按本地消费者保留既有错误合同，未为source测试导入缺失的大目录系统。

真实转发定位native Responses和WS→HTTP的账号URL none过滤，Terra按映射后身份保留；metadata优先需要对Terra避开强制图片覆盖。首次ETag header canonicalization、重放请求身份和浏览器合成load_factor为新夹具适配，生产dedup/表单/clipboard代码保持。首次race超时在重复完整迁移阶段，最终用一次完整迁移模板的独立副本保留隔离并通过；失败过程保留，不作PASS证据。
