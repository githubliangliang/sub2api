# 设计与决策

## 1. 状态与生命周期

S01 采用来源 PR 的不可变回放正文约定；共享 slice/string 的所有写入者必须先查清。
不只补 unsafe helper：lineage digest、TTL、容量上限、历史清理、失败恢复共同验收。
不带入来源文件上下文中本 fork 没有的其它状态分支。现有 input metadata 清理需保留且不原地污染共享正文。

S02 对客户端流退出使用明确的取消上下文，先取消再关闭 Body；保留已有用量采集和终止事件处理。
WS 取消应可识别为客户端生命周期事件，不能随意计入账号故障。

S03 将持久冷却作为判定依据，同时以 generation + deadline CAS 防止删除并发新 block。
上游策略在快照未同步/写库失败时 fail-open，这是一项需要验证的取舍：测试必须覆盖本 fork 的
snapshot/hydrate 调用顺序。若不能保持预期的瞬时错误保护，应记录具体差异后调整局部设计，
不得悄悄把所有本地封锁都清掉或把 S03 标记完成。

S04 在确定失败时释放注册会话，使用脱离已取消请求的上下文；成功或已产生计量的会话沿用原语义。
用 miniredis 验证 ZREM 与并发槽行为；补齐 interface 的全部实现和 mock。

实施适配：本地 CountTokens 的 SelectAccountForModel 不注册 session，故排除来源 PR
的 CountTokens 注销 hunk，避免删除以前成功 Messages 留下的槽。Messages 在选择返回后
立即追踪槽，覆盖等待/准入前退出；newSelectionResult 增加 sessionID 参数，在注册成功但
hydration 因取消等错误失败时直接释放，避免 handler 尚未收到账号而无法回收。

## 2. 协议与模型

S05 按工具身份、arguments 完整性、bootstrap 校验边界分别取终态。done 补齐不能重复发送 delta，
额外工具不能越过命名冲突规则，历史 delegation 不能放过含糊的孤儿 output。
与旧 namespace/custom-tool 待办同文件的冲突按行为解决，不因本轮新增测试而误认为旧项已合。

S06 把模型别名、pricing、metadata snapshot、工具能力、推理模式、指令和缓存作为完整链验证。
能力信息落在已有 JSON Extra，不引入 Ent/schema。保留本地路由目录，不以引入 pinned-account
或整个 service_tier 栈来消除编译冲突。按 v0.2.4 最终行为整合，排除兑换/支付测试污染。
本批的 Ultra reasoning 元数据与未立项的 Ultrafast 服务档位分开验收。

实施适配：S06 保留本地 routed catalogue 的函数参数和筛选逻辑，工具能力经已有 Extra
快照和描述符 JSON 叠加；不引入 pinned-account/额外 search-tool 分组基座。Astra 定价
复用本地 GPT-5.6 固定长上下文和缓存写入策略，不引入动态阶梯或 Fast 转发。
本地没有来源 #6690 所保护的 reasoning.mode 删除路径，保留既有 HTTP/WS 透传语义，
不为添加 guard 而引入非 Astra 的新删除策略。重试回归改用本地已有 encrypted-content
恢复，而非缺失的 input.status rejected-field 扩展。创建账号同步仅补所需的 preview
标记和创建后同步步骤（27 行），避免前端预览的能力在账号创建后丢失。

S11 统一启动时解析的 CLI 版本、出站 UA、billing body 和持久指纹。升内置常量不能替代旧指纹更新。
缓存正常/自愈/新建三条路径都应满足版本下限；更新只涉及版本，不重置其它身份字段。

实施适配：S11 的版本下限取 `claude.CLIVersion()`（包含启动时覆盖值）。来源 #6481
在 floor helper 中直接引用内置常量，会使启用覆盖时老缓存仍低于实际出站默认版本；
本地 subprocess 回归同时验证新建、正常缓存、自愈的 UA/body 一致性。

## 3. 存储、配置和 UI

S07 自定义价格文件按内容变化重载，失败保留有效快照并允许后续恢复；不能发布半更新价格表。
none 映射复用现有配置结构，不引入转发 deny 新策略或迁移。

S08 将 omitted/null/value 明确区分，验证合并后的代理设置；覆盖导入和批量调用点。
有向备份允许多个源共享目标，同时禁止非法自引用/环；多次回退应保留原代理标识。
沿用 SQLite repository，不能复制上游 PG SQL。

实施适配：S08 不修改 Ent schema/生成物。SQLite migration 149 已有非唯一 backup_proxy_id，
但旧 Ent 把关系当作对称 O2O；repository 在原事务内单独写该列，避开反向写入并检查有效
fallback 链。旧 API 会给 fallback_mode=none 的备用 B 写入指向主代理 A 的反向 ID，
这是正常升级数据。循环检查只沿 mode=proxy 的有效转发关系继续，保留未激活反向 ID；
将 B 改为 proxy 时重新检查并拒绝形成有效环。仍拒绝自身引用和不存在/已删除的目标，
不批量重写旧数据。expiry sweep 原有 JSON 问号/减号语法在实际 SQLite 测试中失败，
本簇将所触及 SQL 改为 JSON1 和 CURRENT_TIMESTAMP，保留 probe 清理及 outbox 语义。

S09 保留现有 SQLite cleanup executor 和 sink 写失败退避；默认减少 access log 入库，警告/错误/审计保留。
新配置要兼容缺字段旧配置；7 行缺失 helper 随本簇加入。不能通过关闭 worker 来完成“适配”。

S10 紧凑列表必须保证用户进入编辑时得到完整详情；批量更新不把缺字段误作空值。
保留 simple mode、hidden_menu_keys、Spark shadow 和全选语义。按需导入现有组件接口，跑真实组件用例。

S12 局部升级 go-redis 和必要间接依赖，不同步上游 go.mod；进程内 miniredis 模式也有客户端连接池。
验证 ZRangeArgs 对现有索引扫描范围、批量上限和回收结果的等价性。

## 4. 门禁与回退

每簇独立提交；内存所有权、CAS、会话回收增加有意义的并发/race 验证。
没有新迁移，可按簇回退代码，但涉及配置和已持久化 Extra 时需验证旧版读取兼容。
实际缺口/额外来源必须在 verification 记载，冻结 source-baseline 不变。未经功能决策不扩到第三、四档。
