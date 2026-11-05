# Fast 服务档位与操作日志清理修复

基线：`667420d77acacde2a88b69b47975d0fab55e099c`，VERSION1.1.22、Go1.27.3、SQLite227。用户于2026-11-05要求修复日志Fast服务档位和操作日志全部清理。本次为当前路径修复，不做新上游release intake，不改历史来源、迁移或依赖。

## 原因与修复

- 原生Anthropic转发结果和日志构造器没有速度字段，客户端转发了`speed=fast`，页面仍因`service_tier`缺失显示Standard。新增转发/用量字段，在native/OAuth和API Key自动透传、流式及非流式路径保留最终发送的请求速度；上游明确返回`usage.speed=standard`时优先记录Standard。流式用量有直接parser和patch合并两条入口，两者均覆盖；未知值和仅beta header不推断Fast。原生请求没有新增Fast开关、价卡或发送行为。
- 操作日志使用异步队列/批次写入，旧记录在数据库清空后仍能重新落库。清理与落库使用同一写锁，队列记录携带代数，成功清理后淘汰旧代队列及本地批次；Record仍非阻塞，新操作保留，失败清理不淘汰。同步清理留痕及现场TOTP要求保留。
- 页面较早发出的列表请求可在清理后晚到，覆盖新结果；加入请求序号，旧结果/错误不再覆盖当前查询。按钮检查状态时显示加载并禁用；未启用TOTP时展示系统设置或个人资料入口；验证码只接受六位数字，清理中禁用输入并阻止重复提交。

Codex/OpenAI的`priority`→Fast展示与上游声明优先规则原本已有，本次正常回归通过；上游明确返回default仍记录Standard。用户未补充客户端和清理报错细节，本次修复以实际复现的原生Claude缺口、异步旧记录恢复和列表乱序为依据。

## 验证

原生Claude native/透传 × SSE/JSON × 上游fast/缺失/standard的12个场景经过实际Forward和日志构造器；单独验证SSE缺失/未知帧不覆盖既有速度。新增测试先复现缺字段，之后通过。

真实SQLite回归先复现清理后旧队列重新出现；修复后验证排队旧数据、清理后新数据、取消失败和同步留痕。独立并发测试验证清理与执行中的批次写入顺序，以及数据库写入阻塞不阻塞Record。既有真实SQLite+miniredis+TOTP的handler场景继续通过。

前端组件先复现旧列表晚到恢复已删行，再验证修复；涵盖状态加载、未启用2FA入口、字母码/Enter、清理失败重输与重复提交。

| 检查 | 实际结果 |
|---|---|
| 后端完整unit | 54包通过；最终运行允许复用已通过包缓存 |
| 定向race | service/repository/admin三包通过 |
| 后端embed构建、lint | 通过，0 issues |
| 前端完整tests | 291文件，2178通过、2既有跳过 |
| 前端typecheck、lint、build | 通过；最终build包含vue-tsc类型检查 |
| 真实浏览器 | 全新临时SQLite+miniredis实例，实际构建二进制/登录/SPA；未配置TOTP引导、实际TOTP绑定与清理、清理后刷新和移动屏幕验收通过 |
| 实际清理后数据库 | 删除3条，等待writer后仅1条`admin.audit_log.clear`同步留痕 |

浏览器账号、用量及管理员gate均为新建测试夹具，现场验证码通过本地新建验证器生成。只清理该临时数据库；自建浏览器/服务已关闭。测试曾使用过于宽松的日志动作断言`audit_log.clear`而失败，核对常量后按实际`admin.audit_log.clear`确认数据库结果；产品动作名未修改。初轮新增流式Standard反例发现speed未经过patch入口，补齐该生产入口后定向/race和完整门禁通过。无业务实例清理、部署、提交或push。

完整命令与检查输出见[检查索引](logging-fixes-2026-11-05/checks.json)。受保护历史/依赖/迁移/Ent/Wire共3134文件保持哈希，最终范围见[scope.json](logging-fixes-2026-11-05/scope.json)。
