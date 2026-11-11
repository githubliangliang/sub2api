# Design

## Decisions

按行为适配，不merge上游main。基线已实现完整Fast、Terra、v0.2.20两档、SQLite后台服务及固定配额窗口，不重复导入旧基座。真实业务数据不参与评估与验收夹具。

### S01 原生 WS 后续轮次失败关闭

源码复核：本地已有CurrentTurnFailover包装及handler处理，但主函数返回匿名error，缺后续轮次兜底；需适配返回路径。未在本次跑两轮WS新行为探针。

保留每轮Fast/模型/usage、lease与槽位释放；只有当前轮完整历史足够时才允许已有安全重试。

### S02 OAuth WS 搜索激活与完整回放

AllItems、两个search检测器和当前轮回放builder本地已存在；新增来源helper60行及连接能力状态，未运行新能力专用WS探针。

未知previous_response_id或不完整工具历史要明确关闭，不能猜测回放；写失败不能标记初始化，状态受write锁保护，连接关闭/预算回收。

### S03 WS 与 SSE 首 token 口径一致

本地仍硬编码isTokenEvent；openAIStreamDataStartsTTFT存在可复用。基线WS包通过，新增模式行为本次仅静态核实。

本地上一轮已经修每轮计时；保留本地起点接口，不引入来源FirstTurnStartedAt/TakeNextTurnStartedAt整基座只为编译。

### S04 显式模型冷却覆盖容量重试

HTTP及四reader×三凭据×两错误形态overlay失败。来源WithModel helper本地缺，不能直接复制测试；需贯通canonical model和terminal副作用，保留本地非流式failover修复。

未配、account范围、无模型、排除码及其它平台保留请求级默认；output/metadata关键词不能触发；写失败不扩大整账号冻结。

### S05 Anthropic 兼容入口账号错误策略

compat-policy overlay复现选定规则被默认failover状态列表屏蔽，未配/默认500等负对照通过。

pool模式重试与shouldDisable结果一致；不要将全部客户端4xx升级为failover，不扩商业支付策略。

### S06 合成 effort 的映射与最终校验

effort-binding overlay两平台6子例失败，其余边界通过。需raw显式校验→转换→策略→最终校验，不能简单删除前置校验。

普通省略不能被medium映射/上限意外改变；现有GLM通用兼容降级保留；不引入Kimi/Zhipu等原生平台或requested_effort schema。

### S07 Refresh Token 原子轮换

refresh-race overlay八请求实际全部成功；不需要上游支付或多实例平台前置。

新增接口需更新所有stub；保持预校验顺序，失败不泄露token；单节点miniredis真实执行，不能仅用mock证明原子性。

### S08 Key 列表独立加载用量与 SQLite 覆盖索引

本地await统计阻塞loading；真实modernc SQLite完整228迁移+20000合成记录，新增临时覆盖索引计划命中且仓储结果前后一致；非生产延迟收益证明。

SQLite拟议229而非上游249；(api_key_id,created_at,actual_cost)先验证写入/大小代价与查询计划，禁止CONCURRENTLY/INCLUDE及修改旧迁移。

### S09 账号刷新保留行与已知配额

本地loading整表骨架及quota分支仍存在；59项前端邻近基线通过，新交互本轮未跑浏览器。

只保留同账号数据；旧请求不可覆盖新账号；不依赖未选Antigravity新查询按钮或Ollama基座。

### S10 Codex Desktop 页面上下文兼容

两项前置本地已有，来源新helper123行；本地仅automation/delegation正常化。静态确认缺转换，未宣称实际Desktop联机复现。

有真实call_id、错误namespace/name、重复JSON键、坏page_id/分片头均不转换；不提升system权限；日志不含页面内容。

### S11 既有安全审计 worker 入队唤醒

本地500ms ticker才claim，无wake接口；属于已有功能延迟优化，基线Worker通过，收益需实施时实测。

保持SQLite claim原子性和1C1G资源界限；不新增审计规则、远端扫描策略或实例运营能力，不把来源平均250ms当本地实测。

### S12 Claude Code 内置版本基线

本地2.1.258，CLIVersion覆盖基座完整；这是已有身份维护，列第二档，不声称复现版本导致拒绝。

仅constants的版本literal和必要期望测试；不引入Fable/Mythos5.2、目录/定价/白名单/UI等同PR新能力。

## Validation and rollout

先复现各簇旧行为，再独立实现并执行真实消费者验收；相关错误、取消、重试、usage和SQLite去重一起验证。最终全量unit、构建、lint和定向race。前端变更需typecheck/lint/test/build/critical及本地浏览器；没有前端变更则记录不适用。资源或索引变更需测峰值/查询/写入成本。发布与部署需独立指令。
