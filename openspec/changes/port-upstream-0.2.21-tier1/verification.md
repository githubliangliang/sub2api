# Verification

2026-11-11：F01–F11、39/39任务完成，实现随本次提交保存。实际命令、原始日志和失败过程均保存在 [实施证据](../../../docs/upstream-sync/implementation-0.2.21-tier1/README.md)。评估探针不作为本表的实施验收。

| 项目 | 实际证据 |
|---|---|
| 实际实施起点SHA | `4f4c80c3ce48ce4afbc9344536b9cf44ed3fd898`；[baseline](../../../docs/upstream-sync/implementation-0.2.21-tier1/baseline.json) |
| 实现提交SHA / 产品哈希 | 实现随本次提交保存；[完整产品patch](../../../docs/upstream-sync/implementation-0.2.21-tier1/product.diff)、[28文件SHA256](../../../docs/upstream-sync/implementation-0.2.21-tier1/product-hashes.json) |
| 后端全量unit/build/lint | 独立tier1快照54包unit通过，embed build通过，lint0 issues；[unit](../../../docs/upstream-sync/implementation-0.2.21-tier1/isolated-unit.log)、[build](../../../docs/upstream-sync/implementation-0.2.21-tier1/isolated-build.log)、[lint](../../../docs/upstream-sync/implementation-0.2.21-tier1/isolated-lint.log) |
| 相关race | apicompat、antigravity、service、handler、repository五包定向通过；[race](../../../docs/upstream-sync/implementation-0.2.21-tier1/isolated-race.log) |
| 真实HTTP/SQLite/miniredis | [56个顶层检查](../../../docs/upstream-sync/implementation-0.2.21-tier1/isolated-sqlite-http.log)通过；实际本地HTTP、排队结算/dedup/Fast、完整迁移/228、配额及固定窗口、miniredis冷却范围 |
| 前端与浏览器 / 不适用依据 | 本批无前端文件/行为变化，不要求前端构建、测试或浏览器验收；隔离embed build使用基线既有dist |
| 1C1G/索引成本 / 不适用依据 | 本批无索引/迁移/预算或后台资源策略改变，不重新作1C1G压力验收。维持64MiB body、40MiB SSE行、120s idle；UTF-8早退有预算释放测试 |
| 保护范围与严格校验 | [scope-check](../../../docs/upstream-sync/implementation-0.2.21-tier1/scope-check.json)、[OpenSpec严格校验](../../../docs/upstream-sync/implementation-0.2.21-tier1/openspec-tier1-final.log)；来源与历史冻结 |

## 逐簇结果

共同修复前证据：[before-regressions.log](../../../docs/upstream-sync/implementation-0.2.21-tier1/before-regressions.log)，四包成功编译执行，15顶层失败、4顶层对照通过。首轮修复后 [after-targeted.log](../../../docs/upstream-sync/implementation-0.2.21-tier1/after-targeted.log) 22顶层通过；补充边界 [boundaries-final.log](../../../docs/upstream-sync/implementation-0.2.21-tier1/boundaries-final.log) 13顶层通过。最终独立unit/race覆盖产品最终哈希。

| 簇 | 修复前复现 | 实施后结果 | 边界/残余风险 |
|---|---|---|---|
| F01 Codex元数据与failover | 两类Codex元数据×普通/透传使overload失去换号条件；普通metadata对照通过 | 元数据保留为attempt-local；相同测试确认下游未写、换号错误及firstTokenMs为空 | 仅列举两种Codex metadata；文本/工具/终态和未知codex输出不扩大分类；已有首输出/timeout测试通过 |
| F02 密封参数与namespace | delta fallback的8子例丢namespace/密封标记，terminal/item.done通过 | 两模式×三重建形态×缺失/null/[]/非空标记24子例通过；参数字节、call_id、namespace和输入/输出usage保持 | 只补accumulator字段与输出类型，不重构既有原文路径 |
| F03 CC Switch封装 | 合法、坏base64及带空白封装被当作encrypted_content | 三形态均忽略；原生密文、assistant文本、tool_use/result顺序与参数保持 | 只识别`ccswitch-openai-reasoning-v1:`前缀，不泛化删除其它签名 |
| F04 会话种子 | 后续system/developer改变seed并丢粘性账号 | user/assistant/function_call/function_call_output/input_text五种边界、两类提醒通过；普通/高级scheduler及content/header/prompt_cache_key六场景保持账号1 | model、tools、instructions、首system/developer和首user差异仍隔离 |
| F05 释放并发槽 | 补全失败release=0；成功不早释放对照通过 | 取消、仓储错误、账号消失恰好释放一次；成功交给调用方；等待无release不panic | 本地会话释放同时保留，取消后使用独立ctx清理 |
| F06 价格hash | hash先更新使旧正文被标新，第二轮下载次数1；额外大小写测试出现重复下载3次 | 落盘/内存锚点与正文一致，旧正文后再次下载新正文，再次一致跳过；启动与轮询EqualFold | 无hash、hash网络错、正文网络错、非法JSON及热重载并发覆盖；失败不发布新正文/价格；过程见hash-case-boundary-before.log |
| F07 Sonnet5.5缓存读价 | fallback为0.20/百万 | 资源、Pricing缺远端fallback、Billing fallback三路及三个别名为0.10/百万；Fast0.20；覆盖0/非0及显式Fast倍率通过 | 只改Sonnet5.5缓存读字段；其它费率/模型维持原测试；内部覆盖价单位为每token |
| F08 Gemini非流式思考 | Messages/Chat正文含thought=true文字，thought-only响应泄漏 | 两消费者正常文字与thoughtSignature保留；thought-only/MAX_TOKENS、工具调用及thoughts token统计通过 | 最小bool helper，不导入完整thinking策略；媒体既有测试保留 |
| F09 工具图片顺序 | 图片插到连续functionResponse之间 | Claude/Gemini已有模型、Responses兼容入口：三工具结果先连续写出，再两张图片，ID/内容/顺序通过 | native functionResponse.parts、单工具/无图片回归通过 |
| F10 Codex图片状态 | service延迟429被写200；真实HTTP两个顶层测试失败 | generation/edit经真实handler、HTTP上游和代理，1.2秒延迟400/成功均通过，首拍1秒不提前提交 | 普通客户端心跳保留；短代理期限会产生504，部署需要完整首字节等待预算；header不授权、stream不改 |
| F11 UTF-8边界 | 9非法字节子例进入model字段校验 | 流/非流、普通/compact、首/中/尾非法或截断字节的36个真实HTTP请求全部400；有效中文/emoji继续字段校验 | 先鉴权，保留体积和budget处理；无上游重试，早退后1MiB预算完整可再占用；不是PG修复或真实503复现 |

## 并行改动与验证边界

共享工作区在最后检查时出现并行tier2执行。`backend-race.log`记录暂时缺失`normalizeCodexPageContext`的编译失败；共享`backend-unit.log`曾通过，但最终结论使用独立快照的`isolated-*`记录。[isolated-snapshot.json](../../../docs/upstream-sync/implementation-0.2.21-tier1/isolated-snapshot.json)记录起点和产品文件，完整patch反向检查确认tier1仍存在于共享工作区。其它并行修改由scope-check单列，不声称共享工作区全部文件等于本批快照。

实际HTTP提供商、鉴权和数据为本地合成夹具；持久化使用真实SQLite，缓存专项使用miniredis。无真实付费账号、业务实例操作或部署。无前端/依赖/Go/VERSION/schema/Ent/Wire变化由本批引入。其余tier2、第三/四档和支付不属于本次实施结论。原始提案/设计/规格/来源保持原文；旧verification/tasks/README保存在证据目录。
