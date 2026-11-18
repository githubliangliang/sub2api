# 实施验收

2026-11-18 按用户选择完成 F01–F21 的既有消费者实现和验收。验收在未提交工作区完成，验收起点/当时 HEAD 与产品 SHA256 分开记录；实现随本次提交保存，冻结 source/proposal/design/spec 保留原内容。

| 项目 | 命令/结果 | 证据 |
|---|---|---|
| 实际实施起点 | HEAD `7a0b78fe511f17b0f7eb9d26fcf5e4ff083312c7` / tree `3882c3ce52fb24f6f8d89d782ffc7b48bb1e65e7`，产品无未提交起点改动 | [start](../../../docs/upstream-sync/implementation-0.2.23-tier1/start.json) |
| 实现SHA/产品哈希 | 验收时 HEAD 同起点，保存 31 个产品源码、24 个测试文件的差异/哈希；实现随本次提交保存，验收记录不回写 | [diff](../../../docs/upstream-sync/implementation-0.2.23-tier1/product.diff) / [hashes](../../../docs/upstream-sync/implementation-0.2.23-tier1/product-hashes.json) |
| VERSION/Go/SQLite/auth snapshot | `1.1.25 / 1.27.3 / 229 / 21` 全部保留 | [scope](../../../docs/upstream-sync/implementation-0.2.23-tier1/scope-final.json) |
| backend全量unit/embed/lint/race | 54 包 unit、含新前端的 embed build、lint 0 issues、六包相关 race 通过 | [完整命令/输出](../../../docs/upstream-sync/implementation-0.2.23-tier1/README.md) |
| 真实HTTP/WS/SQLite/miniredis | Sol/Astra 16 个 HTTP/WS/Fast 结算组合，成功去重、固定 quota、窗口旧格式/新费用、snapshot/hydrate 与 WS sticky 保护通过 | [HTTP/WS/SQLite](../../../docs/upstream-sync/implementation-0.2.23-tier1/real-http-ws-sqlite-csp-accepted.log) / [race](../../../docs/upstream-sync/implementation-0.2.23-tier1/backend-targeted-race.log) |
| frontend完整门禁/浏览器/移动端 | typecheck/lint/build、2206/2 既有 skip、critical 179/2 既有 skip；桌面/移动 Chrome 四组操作、122 次合成 API 请求、0 page errors | [tests](../../../docs/upstream-sync/implementation-0.2.23-tier1/frontend-full-test.log) / [browser](../../../docs/upstream-sync/implementation-0.2.23-tier1/browser/results.json) |
| 性能/资源/工具链测量 | 首输出回退在临时 1 CPU/1 GiB 容器完成 commit/drop/8 MiB 上限/benchmark，memory.peak=70,287,360 B；Go 不升级 | [measurement](../../../docs/upstream-sync/implementation-0.2.23-tier1/resource-1c1g.log) |
| 依赖/生成代码/历史保护 | 3557 个保护文件、669 个静态文件原哈希保留；无 Ent/Wire、依赖、迁移/配置修改 | [scope](../../../docs/upstream-sync/implementation-0.2.23-tier1/scope-final.json) |
| 遗留与边界 | 第一档既有消费者无产品遗留；F20 来源 native 分支 NOBASE，N/A。第二档实现及未选能力不随行；本次提交保存实现与相关评估/提案，未推送/发布/归档/部署 | [adaptation](../../../docs/upstream-sync/implementation-0.2.23-tier1/source-adaptation-facts.json) / [说明](../../../docs/upstream-sync/implementation-0.2.23-tier1/README.md) |

| 簇 | 实际结果 | 证据 |
|---|---|---|
| F01 Responses 密封调用内容 | PASS：三种响应形态/处理模式保留 sealed content、namespace、参数和调用 ID | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/protocol-f01-f03-after.log) |
| F02 对话中途 system 权限 | PASS：developer 正文、首部提示/顺序；工具输出/图片不提升权限 | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/protocol-f01-f03-after.log) |
| F03 JSON Schema 默认 strict | PASS：缺省 false、显式值/null、schema、text/json_object 控制 | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/protocol-f01-f03-after.log) |
| F04 Codex HTTP 外部上下文 | PASS：三个名称、旧 page、空/重复/真实调用/未知名称及实际 HTTP | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/targeted-all-tier1.log) |
| F05 Claude per-turn beta | PASS：精确 beta、API Key/OAuth、最终 policy 与消息/顶层字段 | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/protocol-f04-f07-after.log) |
| F06 count_tokens 请求头唯一性 | PASS：真实 HTTP 大小写/多值、覆写/默认、鉴权/beta 优先级 | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/protocol-f04-f07-after.log) |
| F07 输出后的兼容流账号副作用 | PASS：文本/工具后错误生效、模型作用域、不重放和普通错误/取消控制 | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/backend-targeted-race.log) |
| F08 既有 Sol Ultra Fast 倍率 | PASS：Sol/Astra 6×、别名/长期上下文/0/Fast/Free Fast，实际 HTTP/WS 结算/去重 | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/real-http-ws-sqlite-csp-accepted.log) |
| F09 登录身份回填空指针 | PASS：nil guard、实际 SQLite 身份归属冲突与既有登录/绑定控制 | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/health-f08-f13-after.log) |
| F10 5h 滚动保留有效周用量 | PASS：独立周桶未来/过期/缺 reset、秒/毫秒、仅周响应/sample 时间 | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/health-f08-f13-after.log) |
| F11 账号5h窗口上下界 | PASS：实际 SQLite 229 start/end/outbox、旧窗口费用排除和毫秒精度读取 | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/health-f08-f13-after.log) |
| F12 首输出临时文件不可用 | PASS：实际缺目录、权限失败、commit/drop/取消/清理及 8 MiB 上限 | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/resource-1c1g.log) |
| F13 Claude 默认测试模型维护 | PASS：已有 4-6 默认、显式模型/其他默认保留；无供应商退役联机宣称 | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/health-f08-f13-after.log) |
| F14 Grok 媒体扁平错误 | PASS：嵌套优先/数字 code/脱敏/本地内容策略/规则/未知对象，共享提取器未扩展 | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/targeted-all-tier1.log) |
| F15 Gemini 非数组 items | PASS：三协议 schema、合法 array、既有 tuple/default/enum 保护 | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/backend-targeted-race.log) |
| F16 白名单精确模型输入 | PASS：点击/Enter、草稿/修正、重复/映射提示与桌面/移动保存重开 | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/browser/results.json) |
| F17 渠道 Bedrock 兼容开关类型 | PASS：创建/编辑/重开 true/false、移除区段及其他 features | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/browser/results.json) |
| F18 调度快照端点能力 | PASS：metadata 实际 miniredis、端点过滤、空/未配置、hydrate/WS sticky；无 Seedance | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/backend-targeted-race.log) |
| F19 设置注入 HTML 与 CSP nonce | PASS：生产中间件实际 HTTP 两 nonce、设置/回退、无 nonce 304、hash assets | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/real-http-ws-sqlite-csp-accepted.log) |
| F20 Chat 客户端 metadata.user_id | PASS（既有消费者）；N/A（来源 native 分支 NOBASE）。JSON/legacy、类型/空值、原 token/OAuth 控制 | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/targeted-all-tier1.log) |
| F21 完成转发后的粘性绑定 | PASS：5 秒截止、完成后取消/超时、换号不覆盖/失败不绑定、profit admission | [evidence](../../../docs/upstream-sync/implementation-0.2.23-tier1/targeted-all-tier1.log) |

F20 的来源原生 Anthropic 文件在冻结逐文件表中已为 NOBASE，未引入新消费者；F21 原有脱离取消的 3 秒 helper，仅选定粘性写回对齐 5 秒。实际供应商、Codex Desktop、业务实例与浏览器 bfcache 策略不作为本轮验收对象。失败、负对照、来源/夹具适配和测量边界保留在 [实施证据](../../../docs/upstream-sync/implementation-0.2.23-tier1/README.md)。
