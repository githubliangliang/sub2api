# Source Feature Map

| 编号 / 行为 | 来源 | 本地缺口与接受范围 | 验收 |
|---|---|---|---|
| S01 Grok响应模型审计 | #8339 整体PR/直接diff | 原生Responses返回原文审计字段，Chat/Messages缓冲路径保留固定大小前序模型及冲突位。 | JSON/SSE与两条buffered路径、跨事件冲突、终态优先/字段长度有界、非Grok保持；handler异步快照与SQLiteusage字段。 |
| S02 Grok目录与受管快照一致性 | #8339 整体PR/直接diff | 现有手动/后台统一严格解析及成功/失败观测，SQLite单键CAS/有界重读，普通编辑只采用数据库最新受管状态。 | 非法形态/有效空目录/自定义ID、旧格式、失败保留成功、乱序/非法时间/来源变化、CAS竞争、外部事务回滚/无提前缓存、旧表单/缺键/注入/编辑重叠及取消。 |
| S03 图片上传完整性 | #8348 整体PR/直接diff | OpenAI按原20MiB单part边界明确拒绝超限，Grok在现有总body边界内完整解析或明确失败。 | 刚好上限与+1、mask/多part/文件名/类型、Grok审查与实际转发一致、全body上限、取消；不扩大50MiB或下载边界。 |
| S04 渠道搜索平台维度保存 | #8390 整体PR/直接diff | 仅合并现有Web Search特定key的平台维度，现有前端保存者显式写入已支持平台true/false并支持清除。 | 两维/部分旧客户端/缺key、false/空对象/null、未知维度、其它features整块替换、输入map不变、全局隐藏开关和实际UI保存。 |
| S05 既有Gemini模型接口与默认值 | #8395 整体PR/直接diff | 全部账号声明映射时列表/详情与既有/v1/models同源，普通连接测试及已有默认目录更新，按本地custom list规则过滤。 | all-mapped/partially-unmapped、详情404/通配、custom list、路径护栏、force/composite/mixed、Google One/显式旧模型、已有Antigravity下拉去重；不加未选定新模型。 |
| S06 x/net安全维护 | 0a2a29b32 整体PR/直接diff | 以x/net0.62.0和量测的最小MVS图修复已有HTML平方复杂度，不覆盖整份上游mod。 | 独立PoC近线性增长、强制传递版本与Go1.27.3/生成工具兼容、unit/build/lint/相关网络race及漏洞复扫；其它finding单列。 |

[评估](../../../docs/upstream-sync/PORTING-0.2.19.md)保存逐文件落点、前置量测与完整四态；本表定义接受的行为，不是整PR导入授权。新增回归需覆盖真实消费者；源integration仅有标签不代表SQLite可跑。当前实际测试目录见[review-test-list](../../../docs/upstream-sync/evidence-0.2.19/review-test-list.txt)，未来新增用例不能被称为baseline现存测试。

其它新能力与支付保持原档位；评估期原始输出、旧任务完成记录及SHA不改写。

