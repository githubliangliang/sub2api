# 实施验收（2026-11-26）

状态：**已实施，F01–F03完成13/13任务**。实现及证据纳入用户授权的本地commit，活动变更保留等待后续归档。

实际实施起点：`89060f532a9000af28a47036f70cc44821dafd55`，2026-11-26，VERSION1.1.28 / Go1.27.4 / SQLite230 / auth21。初始2份已修改文档、112份未跟踪文件；冻结评估来源、source-baseline、上批18份实现文件及17份旧活动变更467完成任务均保留。完整命令、原始失败/通过及范围见[实施证据](../../../docs/upstream-sync/implementation-historical-followup/README.md)。

| 验收项 | 实际证据 |
|---|---|
| 真实起点与范围 | [初始状态 gzip](../../../docs/upstream-sync/implementation-historical-followup/initial-state.json.gz)、[历史/整PR保护](../../../docs/upstream-sync/implementation-historical-followup/protected-source-check.json)、[最终范围](../../../docs/upstream-sync/implementation-historical-followup/validation.json)。 |
| F01 原缺口与修复 | [原探针](../../../docs/upstream-sync/implementation-historical-followup/baseline-replay.json)先重现非GPT错误身份；[修后](../../../docs/upstream-sync/implementation-historical-followup/baseline-replay-after.json)通过。GPT/非GPT/供应商前缀/别名/gpt-oss、正文与占位符原字节、明确源指令、实际HTTP出口及ETag均覆盖；补充代理追加别名和缺失模板两路径的原失败保留，[Codex组合](../../../docs/upstream-sync/implementation-historical-followup/codex-route-paths-after.json)104顶层/211含子例通过。 |
| F02 SQLite/AES与失败边界 | [原SQLite探针](../../../docs/upstream-sync/implementation-historical-followup/baseline-replay.json)重现首次密文/二次留空明文。[HTTP/SQLite](../../../docs/upstream-sync/implementation-historical-followup/http-sqlite.json)、[最终admin](../../../docs/upstream-sync/implementation-historical-followup/backend-admin-final.json)及[最终保护race](../../../docs/upstream-sync/implementation-historical-followup/backend-protection-final.json)验证首次/重复沿用/替换/真实AES legacy plaintext、固定key、加密和写失败、响应脱敏、管理员401/step-up403及数据库实存。unit中store factory零调用；HTTP服务无factory/dumper，没有S3或备份操作。 |
| F03 组件JSON与持久化 | [用户编辑证据](../../../docs/upstream-sync/implementation-historical-followup/user-editors/NOTES.md)保留原数字/清空失败和真实组件回归；[双视口浏览器](../../../docs/upstream-sync/implementation-historical-followup/user-editors/browser/NOTES.md)记录实际JSON。真实handler/service/Ent/完整SQLite迁移验证整数string成功、JSON number/小数/越界拒绝，0/负值/空值/普通文本，以及省略保留/空串清空/非空更新。 |
| backend unit/race与账务保护 | [全量unit](../../../docs/upstream-sync/implementation-historical-followup/backend-unit-final.json)：55包10829顶层/22536含子例通过，16既有跳过；最后仅扩充legacy AES测试后，admin整包445/620再通过。[五包race](../../../docs/upstream-sync/implementation-historical-followup/backend-race-final.json)126/253通过；[三包保护race](../../../docs/upstream-sync/implementation-historical-followup/backend-protection-final.json)16/53通过，包含本地HTTP/SQLite/miniredis、结算/dedup、独立用户与账号成本、固定quota和既有Fast/requested-effort控制。 |
| 前端与构建/lint | [最终全前端](../../../docs/upstream-sync/implementation-historical-followup/frontend-unit-final-summary.json)2329通过/2既有跳过；typecheck/lint/build、backend最终embed build/lint0与版本检查均通过。用户/图片/模板实际桌面及移动Chrome，加最终构建step-up补验；浏览器API是合成拦截，真实HTTP/SQLite单独验收。 |
| 最终哈希/保护/局限 | [21份源码与测试哈希](../../../docs/upstream-sync/implementation-historical-followup/implementation-files.sha256)、[验证脚本](../../../docs/upstream-sync/implementation-historical-followup/verify.py)及[核验](../../../docs/upstream-sync/implementation-historical-followup/validation.json)。9源/12新增测试，无VERSION/Go/依赖/schema/Ent/Wire/资源默认变化；全部历史与Fast诊断保留。原夹具/lint失败和第一次unit143中断分开保存；不当成产品故障。没有供应商、对象存储、业务实例或容量收益结论。 |

多账号冲突别名继续使用既有唯一映射解析的回退；不新增发现系统或改写明确上游/用户instructions。源基线和评估中的“未实施”是冻结快照；本表记录实际实施状态。无push/release/archive。
