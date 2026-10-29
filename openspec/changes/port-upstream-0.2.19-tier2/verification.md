# 实施验收（待填写）

状态：提案未实施。评估期的baseline测试/缺陷probe PASS不属于本文件实施证据，所有结果与证据槽保持空白。

| 实施信息 | 值 |
|---|---|
| 实施起点HEAD / 工作树 | |
| 实现提交 / 代码SHA256 | |
| 实施时VERSION / Go / 迁移 | |
| 实际范围与排除项 | |

## 行为、反例与结果

| 编号 | 必须验收的行为与反例 | 修复前证据 | 修复后证据 | 结果 |
|---|---|---|---|---|
| S01 | JSON/SSE与两条buffered路径、跨事件冲突、终态优先/字段长度有界、非Grok保持；handler异步快照与SQLiteusage字段。 | | | |
| S02 | 非法形态/有效空目录/自定义ID、旧格式、失败保留成功、乱序/非法时间/来源变化、CAS竞争、外部事务回滚/无提前缓存、旧表单/缺键/注入/编辑重叠及取消。 | | | |
| S03 | 刚好上限与+1、mask/多part/文件名/类型、Grok审查与实际转发一致、全body上限、取消；不扩大50MiB或下载边界。 | | | |
| S04 | 两维/部分旧客户端/缺key、false/空对象/null、未知维度、其它features整块替换、输入map不变、全局隐藏开关和实际UI保存。 | | | |
| S05 | all-mapped/partially-unmapped、详情404/通配、custom list、路径护栏、force/composite/mixed、Google One/显式旧模型、已有Antigravity下拉去重；不加未选定新模型。 | | | |
| S06 | 独立PoC近线性增长、强制传递版本与Go1.27.3/生成工具兼容、unit/build/lint/相关网络race及漏洞复扫；其它finding单列。 | | | |

## 组合检查

| 检查 | 命令 / 证据 | 结果 |
|---|---|---|
| backend完整unit / build / lint | | |
| 受影响路径race及取消/资源释放 | | |
| 真实SQLite迁移227 / quota / usage dedup | | |
| 前端适用typecheck / lint / tests / build / 浏览器 | | |
| 依赖适用audit / Go扫描 / 未解决finding | | |
| OpenSpec严格校验 / 链接 / 范围 / 冻结来源哈希 | | |

## 完成条件

- [ ] 每项规格有实际前后证据及反例，基线probe未冒充修复完成。
- [ ] 所有适用检查通过，未执行项写明实际原因，保护范围与冻结SHA核对。
- [ ] 实现提交与实际起点已记录，没有未授权支付、发布或部署。

