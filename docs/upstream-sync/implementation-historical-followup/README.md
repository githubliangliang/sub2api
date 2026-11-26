# 历史未合第二轮：第一、第二档实施验收

2026-11-26 实际起点为 `89060f532a9000af28a47036f70cc44821dafd55` / VERSION1.1.28；起始有2份已修改文档和112份未跟踪文件，含本次评估、两份提案及原Fast诊断。上一批18份实现哈希、旧17份活动变更467项完成任务均受保护。实施覆盖第一档F01–F03、第二档S01–S03，任务关闭及最终范围以[第一档](../../../openspec/changes/port-upstream-historical-followup-tier1/verification.md)、[第二档](../../../openspec/changes/port-upstream-historical-followup-tier2/verification.md)和[最终核验](validation.json)为准。

## 实现范围

- F01：只调整生成模板的开头GPT身份，沿已存在的有效映射目标选择；直接生成、代理清单追加别名和补齐缺失模板均覆盖。GPT原模板、其它工作正文、占位符及明确上游/用户指令保留；不引入发现系统。多账号冲突别名继续沿用既有唯一目标解析的保守回退，不新增路由政策。
- F02：新填及沿用的非空S3 secret都经过既有固定key检查、加密和持久化边界，响应继续脱敏。真实AES-GCM、完整SQLite迁移/Ent设置仓库、管理员认证及step-up门禁控制；未构建对象存储客户端或启动备份。
- F03：number控件提交原始字符串；description/placeholder明确空串表示清空。后端整数/必填/上下界及省略语义保持，没有小数或schema扩展。
- S01：属性读取、用户及分组读取/保存按对象和会话隔离；未成功加载分组时按钮/方法拒绝替换授权；保存捕获本次对象与值。交叉审查补出既有step-up消费者的迟到403/验证结果问题：旧会话在prompt/重试前取消，编辑器按会话挂载TOTP子对话框，子组件卸载后忽略旧成功/失败，保留正常验证重试。这是S01必要的共享组件最小前置，不修改composable API。
- S02：新选择、拒绝文件、删除和卸载取消并隔离旧FileReader，保留300KiB默认、类型限制及SVG预览净化。
- S03：模板/对话框/列表请求各自拥有状态，apply绑定模板与选中ID快照；旧完成、错误和finally不影响新会话，保留防重复提交。

## 证据入口

| 证据 | 内容 |
|---|---|
| initial-state.json.gz / initial-AGENTS.md.gz / initial-upstream-README.md.gz | 实际HEAD/index/refs、9286份tracked及112份untracked哈希、旧活动任务和原文。 |
| [protected-source-check.json](protected-source-check.json) | 上批18份实现哈希及24个完整PR补丁SHA256全部匹配；冻结评估来源原文不回写。 |
| [baseline-replay.json](baseline-replay.json) / [修复后](baseline-replay-after.json) / replay-*.go.txt | 原评估探针在实际起点2个目标顶层失败、3个控制顶层通过（含HTTP类型控制）；修后5顶层/11含子例通过。 |
| backend-regression-before-corrected.json / backend-regression-after.json | 扩充原缺口：2目标顶层失败、1控制通过；实际保存边界、固定key、legacy plaintext、加密/写失败及不创建store。首轮夹具编译错误另保留。 |
| codex-route-paths-before*.json / [补路径验收](codex-route-paths-after.json) | 代理追加别名及补齐缺失模板的映射身份回归，保留原失败；Codex组合104顶层/211含子例通过。 |
| http-sqlite*.json / [最终admin包](backend-admin-final.json) / [最终保护](backend-protection-final.json) | 本地TCP HTTP、完整SQLite迁移/Ent/AES、属性string/number/0/负值/空值/上下界/普通文本、显式清空与省略、原用户快照/其他用户保留、S3重复保存/legacy/认证和写失败；数据库故障为合成注入。最终admin445/620；三包race16/53含结算/dedup/独立成本/固定quota/既有Fast/requested-effort/miniredis控制。 |
| [backend-unit-final.json](backend-unit-final.json) / backend-race-final.json | 全量55个测试包10829顶层/22536含子例通过，16个既有跳过。最终五包race126顶层/253含子例通过。补充legacy AES控制后仅测试内容变化，由最终admin及保护race重新覆盖。 |
| frontend-unit-final.json / [摘要](frontend-unit-final-summary.json) | 最终前端2329通过、2个既有跳过。完整typecheck/lint/build和backend embed/lint命令在同名JSON记录。 |
| [user-editors/NOTES.md](user-editors/NOTES.md) / [浏览器](user-editors/browser/NOTES.md) | F03/S01原失败、最终用户/分组与真实step-up组件测试、桌面/移动实际页面及实际JSON请求。旧权限错误和旧TOTP验证单独复现/修复；以该目录最终补充为准。 |
| [image-template/notes.md](image-template/notes.md) / [浏览器结果](image-template/browser/result.json) | S02/S03原失败、30新增+5既有测试、Settings控制，实际Settings和管理员模板管理链的桌面/移动读取与apply会话。 |
| [implementation-files.sha256](implementation-files.sha256) / [verify.py](verify.py) / [validation.json](validation.json) | 9份产品源码、12份新增回归测试，最终哈希、旧任务/历史/依赖/schema/生成代码/资源默认及原Fast诊断保护。 |

根目录每个执行记录保存实际命令、cwd、环境、UTC时间、退出码和原始stdout/stderr；压缩日志解压后内容不变。浏览器使用本地Chrome及真实Vue页面，API为合成HTTP拦截，图片先强制乱序回调再用原生FileReader；浏览器不冒称同一会话直接写入真实数据库。真实HTTP/SQLite验收单独记录。

## 保留的失败与边界

`backend-regression-before` 首轮新测试曾与既有测试文件重名而临时遮住旧helper，并有指针receiver夹具错误；在产品修复前恢复原文件、改为独立测试文件后重现目标失败，原文件最终哈希受保护。`http-sqlite-first` 将不存在的设置误当成空值，修正测试以符合既有SETTING_NOT_FOUND合同。第一次全量unit执行在缓冲runner写出前收到143退出，原因未确认，见[记录](backend-unit-interrupted.json)；改用流式日志后的完整运行通过，不将中断记作产品失败或验收。

第一次lint要求新测试对JSON类型断言检查ok，已修正并通过；原错误保留。首轮范围检查在证据压缩前验证了尚未生成的gzip链接，原输出见scope-before-pack；压缩后重做最终核验。用户/图片/模板/TOTP测试的首轮mock、DOM选择器、带timezone的URL、sr-only控件、移动菜单滚动和异步定位等夹具问题均保留原文并在子目录说明，不通过修改产品迎合夹具。既有pnpm字段提示、Browserslist资料和Vite chunk/import提示保留，没有依赖升级。暂存审查中产品、测试和可审阅文档的空白检查通过；冻结TSV的CRLF及原命令/浏览器文本尾部空白按原字节保留，明细见staged-diff-check.json，不为清理格式改写历史证据。

VERSION1.1.28、Go1.27.4、SQLite230/auth21、依赖、Ent/Wire、64MiB请求体/40MiB SSE行/8MiB暂存/120s idle和图片既有边界保持。用户仅要求本地实现并commit；没有push、release、archive或业务实例操作。其它第三/四档、支付、新平台及容量收益不在本批范围。
