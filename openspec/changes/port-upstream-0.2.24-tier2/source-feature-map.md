# Source Feature Map

| 簇 | 完整来源 | 选定行为与边界 |
|---|---|---|
| S01 所有现有入口的重复字段拒绝与读取一致性 | [5c2982a81](https://github.com/Wei-Shaw/sub2api/commit/5c2982a81b46e2142d0143d2f83c1ce9c4da1d24) | 所有现有HTTP网关请求与每轮WS载荷 SHALL 在模型改写、调度和计费之前拒绝语义重复字段，并让内容识别、body重读与协议错误格式一致。 不引入完整group model_allowlist/glob/live、新平台或PG。只挂本地已存在路由；现有账号模型过滤、64MiB预算/40MiB SSE/8MiB暂存/120s idle不变。相同重复也必须拒绝，不能只拒不同值。 |
| S02 OpenAI 不可逆 OAuth 轮换的 SQLite 保存 | [#8671](https://github.com/Wei-Shaw/sub2api/pull/8671) | 已完成的OpenAI轮换 SHALL 在有界脱离取消的上下文保存；CAS失败须读取当前持久化凭据且保留并发重新授权。 只对完整OpenAI轮换成功启用例外；Grok/Claude晚到规则、固定quota与取消边界保持。无新列，无旧token重试，不能使用普通Update覆盖并发授权。 |
| S03 缓冲 Responses 的 hosted 输出重建 | [#8820](https://github.com/Wei-Shaw/sub2api/pull/8820) | 缓冲兼容响应 SHALL 从最终有效终态及已完成item恢复完整hosted内容，输出顺序、工具身份和流式语义一致。 不以本簇启用未选ToolSearch新能力；不吞终态failed、不重复计费、不改变首内容failover/有界spool。 |
| S04 Grok 成功搜索次数结算 | [#8864](https://github.com/Wei-Shaw/sub2api/pull/8864) | Grok搜索结算 SHALL 在上游提供成功次数时以其为准，缺省时沿用已有事件计数与去重。 不得把失败搜索按item数重复收费；普通函数/客户端工具不进入服务端搜索计费。 |
| S05 OAuth 外族模型与 Messages 映射后调度 | [#8893](https://github.com/Wei-Shaw/sub2api/pull/8893) | 固定Codex上游的OAuth类账号 SHALL 跳过明确外族模型，且调度判断使用渠道映射后的实际模型。 保留用户已选#8748，不把API Key自定义base_url限制为OpenAI模型，也不误拦合法映射别名。 |
| S06 批量透传编辑独立修改模型过滤 | [#8899](https://github.com/Wei-Shaw/sub2api/pull/8899) | 批量编辑 SHALL 只提交用户明确选择修改的透传模型过滤字段。 未勾选修改不等于false；不覆盖既有账号的过滤或凭据。 |
| S07 Claude/Gemini 用量快照刷新 | [#8858](https://github.com/Wei-Shaw/sub2api/pull/8858) | Claude/Gemini快照变化 SHALL 使现有用量缓存失效并刷新，异步旧结果不得覆盖新快照。 保留本地已实施的不闪行与OpenAI unknown新鲜度语义；不引入费用估算/强查按钮或来源原生国内平台。 |
| S08 Codex 原生音频输入声明 | [#8932](https://github.com/Wei-Shaw/sub2api/pull/8932) | 已有原生Responses音频输入能力 SHALL 在模型声明中正确保留，聚合目录只声明全部可调度候选支持的能力。 不引入Voice/Realtime/新的音频计费或模型；静态订阅目录不能替代API Key实际能力。 |
| S09 Anthropic metadata 长上下文目录 | [#8883](https://github.com/Wei-Shaw/sub2api/pull/8883) | 现有Anthropic目录 SHALL 按可服务账号的有效metadata展示长上下文变体，并正确识别默认官方registry host。 不凭目录增加新模型/价格，不导入未选group model_allowlist；现有[1m]请求规范化与计费保留。 |
| S10 Composite 改写后的 Claude thinking | [#8859](https://github.com/Wei-Shaw/sub2api/pull/8859) | Composite将请求改写到既有Claude5.5时 SHALL 同步适配该型号接受的thinking模式并保留其它参数。 不引入新型号或整套signed-thinking框架；保持用户原模型名的使用记录和最终映射策略。 |
| S11 后台新建用户的默认平台配额 | [#8918](https://github.com/Wei-Shaw/sub2api/pull/8918) | 后台创建普通用户 SHALL 按当前默认平台配额初始化，显式0与未配置无限额保持区别。 不导入支付、购买或新订阅策略；不改变固定窗口重置、已删Key结算与历史用户配额，不新增迁移。 |
| S12 异步用量与审计的请求体保活 | [#8875](https://github.com/Wei-Shaw/sub2api/pull/8875) | 离开请求生命周期的用量和审计数据 SHALL 拥有必要标量的独立存储，且不因字段地址或闭包滞留整个请求体。 不全面复制所有body，不引入完整COW；对请求内零拷贝保留所有权，不把函数PoC宣传成生产内存上限。 |
| S13 已有热点的有界窄优化 | [#8845](https://github.com/Wei-Shaw/sub2api/pull/8845)、[#8862](https://github.com/Wei-Shaw/sub2api/pull/8862)、[#8866](https://github.com/Wei-Shaw/sub2api/pull/8866)、[#8868](https://github.com/Wei-Shaw/sub2api/pull/8868)、[#8870](https://github.com/Wei-Shaw/sub2api/pull/8870)、[#8872](https://github.com/Wei-Shaw/sub2api/pull/8872) | 采用的热点优化 SHALL 保持字节、策略和所有权语义，并通过单P前后成本与转义必要条件验证。 可验证body完全相同才复用validated入口；不降低64MiB预算、不采用1GiB动态最低预算、stale grace/账务lane/完整COW；未达收益或等价性的hunk不落地。 |
| S14 SQLite 模型平均速率按耗时加权 | [#8937](https://github.com/Wei-Shaw/sub2api/pull/8937) | 模型统计平均Tokens/秒 SHALL 使用有效请求output_tokens总和除以duration_ms总和。 与用户已选整窗口输出TPS统计口径不同：后者保留逐请求算术mean/P5/P10/P50，不能整站替换为加权值。 |
| S15 Antigravity 裸 429 范围及额度端点 | [#8973](https://github.com/Wei-Shaw/sub2api/pull/8973)、[#8979](https://github.com/Wei-Shaw/sub2api/pull/8979) | 同一请求的无明确信号裸429 SHALL 最多冷却一个账号；既有额度查询 SHALL 使用该账号实际转发端点。 不放宽明确限流或模型容量规则，不新增查询按钮/QuotaSummary平台/强查策略；保留本地端点安全与已实施的model cooldown。 |

patch sites只保存在[评估报告](../../../docs/upstream-sync/PORTING-0.2.24.md)。新增调用扫描是检索而非编译保证；来源helper与本地consumer区别以该报告的逐簇最小前置为准。
