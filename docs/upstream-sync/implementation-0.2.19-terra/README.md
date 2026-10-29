# GPT-6.1 Terra 实施验收（2026-10-30）

用户本轮要求实施已选定的 Terra spec。实际起点为本地 main `b9e51e27366e8eab20ee2c2746f77065c8ee02b6`，产品/测试实现提交为 `a541d0b611f4c2da5cab9811c3c4e02ae8344136`，21 个文件，代码 SHA256 `a97a658cea9ce28d500c61b77f0407734d37d5377ad5b23ce84dfb96810b75b2`。[提交与逐文件核对](./implementation-commit.json) / [范围与哈希](./scope-facts.json)。

[OpenSpec](../../../openspec/changes/port-upstream-0.2.19-terra/verification.md) 14/14 完成，第三档用户选定的新能力单独落地。VERSION 1.1.20、Go 1.27.3、SQLite 迁移227、依赖、Ent/Wire、全局默认模型和菜单/账号平台保留。其余新能力、完整定价/目录大机制及支付仍不在范围内；未发布/tag/push/部署或操作业务实例。上游和评估冻结文档、其它 change 的来源/规格保持原文；[原任务](./tasks-original.md) / [原验收槽](./verification-original.md) 保留。

## 行为验收

- 身份统一为 `gpt-6.1-terra`，六档 none/low/medium/high/xhigh/max，默认 medium；供应商前缀、大小写、空格/下划线与 `-openai-compact` 覆盖，pro/ultra/minimal 等未知后缀不回退到已知家族。Sol 原 none/minimal 拒绝规则通过邻近回归。
- Chat、legacy functions、Messages 和 Responses 在可用账号上经现有 Responses 路径保留工具、模型、effort 与用量；only-Chat 三个出口在出站前 400，none/max 也不绕过，普通文本和既有 raw 兼容路径保留。
- 配置目录默认六档/medium/1050000；显式 text-only/high-only/100k 元数据优先，API Key 官方与自定义目录均保留 raw source、扩展字段、缓存和 ETag。原生空 manifest 仍为空；本地普通模型同步对空 `data` 保留既有“no supported models”错误与空结果，不补入 Terra。
- 冻结资源与两层 fallback 为每百万 input/cache-read/cache-write/output 1/0.05/1.25/5，priority 2x；271999/272000/272001 仅最后一档触发 input 2/output 1.5。显式零、override、渠道覆盖和共享价卡不变性通过。[资源逐项同源与其它205条不变](./resource-scope.json)。
- 真实 HTTP/SSE 转发与 handler 队列后，取消原请求并复用 Gin context，仍保存请求/出站/响应模型、none/max、互斥 token 桶与成本。总输入1000（缓存读200/写300）拆为非缓存500，输出10，priority 实际成本0.00187；SQLite 用户余额/Key quota 一次结算，相同请求身份再次记录不重复扣减。真实 WS handler 六档及映射链另验；HTTP/WS parser 保留既有图/文缓存桶。
- 现有白名单和映射预设同源，配置生成/解析与模型 round-trip 正确；[真实浏览器](./browser/README.md) 完成保存、重开和原生剪贴板写入观测。

16 个新增 Go 顶层行为用例，197 个含子场景的通过项。前后失败/成功分别记录，目录/旧行为保护用例在旧基线上通过亦保留，不将保护项说成新增缺陷。[原始基线失败](./terra-final-baseline-before.txt) / [实现通过](./terra-final-target.txt)；最终 SQLite 夹具对应 [before](./terra-handler-final-before.txt) / [after](./terra-handler-final-after.txt)。前端同一39项为旧产品2失败/37通过、实现全部通过：[before](./terra-baseline-frontend-fixture-fixed-before.txt) / [after](./terra-frontend-after.txt)。

## 组合门禁

| 检查 | 实际证据 | 结果 |
|---|---|---|
| backend 完整 unit | [54个实际测试包](./backend-unit.txt)，SQLite 夹具调整后再跑 [handler 全量](./backend-handler-unit-final.txt) | PASS |
| backend 构建 | [CGO=0、embed server](./backend-build.json) | PASS |
| backend lint | [最终](./backend-lint-final.txt) | PASS，0 issues |
| 相关 race | [五包组合最终输出](./backend-race-final.txt)，另有 [handler 最终](./backend-handler-race-final.txt) | PASS |
| SQLite 迁移227/quota/dedup/固定窗口 | [三包实际检查](./sqlite-protection.txt)及真实 Terra handler SQLite | PASS |
| frontend typecheck/lint | [类型](./frontend-typecheck.txt) / [lint](./frontend-lint.txt) | PASS |
| frontend 全量/critical | [288文件、2106通过/2既有跳过](./frontend-tests.txt) / [176通过/2既有跳过](./frontend-critical.txt) | PASS |
| frontend build | [实际 build](./frontend-build.txt) | PASS |
| 浏览器/清理 | [实际操作](./browser/README.md) / [清理](./browser/cleanup.json) | PASS |
| OpenSpec/source/scope | [strict](./openspec-strict.txt) / [范围检查器](./scope-check.py)及哈希 | PASS |

所有 run-check 记录 command、cwd、UTC 时间、耗时和退出码；原始 stdout/stderr 保留。完整 Go 测试仅在测试子进程移除六项代理变量，沿用前一批已经定位的 SDK localhost 代理问题处理，相关验证码代码/测试未改。

## 必要适配与过程归因

上游通用 version-prefix 表会接受 minimal；Terra 的 base 表及专用规范化已足以覆盖选定身份，未把它放入该宽松表，避免未知能力后缀成为合法别名。没有为 source 的新测试导入缺失的普通模型目录 API。

本地 Responses ingress 与 WS→HTTP 的 none 过滤原来仅看账号类别/base URL；真实转发失败证据定位到这一层，现按映射后的 Terra 身份保留。官方已知 GPT 的强制图片补全对 Terra 单独停止覆盖，以满足账号显式 metadata 优先；其它模型既有策略保持。

新测试第一次 ETag 夹具的 HTTP header map 未用 canonical `Etag`，已更正并在独立旧产品 worktree 重跑。首次 SQLite 去重重放遗漏了原请求上下文身份，定位 `resolveUsageBillingRequestID` 后保留同一身份，未改生产 dedup 代码。首次 race 在12次完整 schema 迁移准备阶段超时，没有 data-race 报告；[原输出](./backend-race.txt) 保留，最终只完整迁移一次并给每个用例独立数据库副本，组合 race 通过。浏览器剪贴板读取权限与合成 load_factor 表单问题详见浏览器记录，未调整产品复制或表单代码。
