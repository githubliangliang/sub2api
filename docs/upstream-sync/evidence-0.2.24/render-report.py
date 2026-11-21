#!/usr/bin/env python3
"""Render frozen assessment documents; never applies a product patch."""
from pathlib import Path
import collections,csv,json,re,subprocess
OUT=Path(__file__).resolve().parent
ROOT=OUT.parents[2]
facts=json.loads((OUT/'baseline.json').read_text())
review=json.loads((OUT/'review.json').read_text())
classification=json.loads((OUT/'classification.json').read_text())
candidates=list(csv.DictReader((OUT/'candidates.tsv').open(),delimiter='\t'))
stats={c['candidate']:c for c in csv.DictReader((OUT/'candidate-stats.tsv').open(),delimiter='\t')}
assert set(classification)=={c['candidate'] for c in candidates}
selected={c for r in review for c in r['candidates']}

def link(c):
 return f'[#{c[2:]}](https://github.com/Wei-Shaw/sub2api/pull/{c[2:]})' if c.startswith('PR') else f'[{c}](https://github.com/Wei-Shaw/sub2api/commit/{next(x["sha"] for x in candidates if x["candidate"]==c)})'

def table_facts():
 return '\n'.join([
 '| 项目 | 冻结值 |','|---|---|',
 f'| fork HEAD / tree | `{facts["fork_head"]}` / `{facts["fork_tree"]}` |',
 '| VERSION / Go / SQLite / auth snapshot | `1.1.26 / 1.27.4 / 230 / 21` |',
 f'| 私有 index tree | `{facts["snapshot_tree"]}` |',
 f'| v0.2.23 commit / tag object | `{facts["refs"]["v0.2.23"]}` / `{facts["tag_objects"]["v0.2.23"]}` |',
 f'| v0.2.24 commit / tag object | `{facts["refs"]["v0.2.24"]}` / `{facts["tag_objects"]["v0.2.24"]}` |',
 f'| upstream/main（观察） | `{facts["refs"]["upstream/main"]}` |',
 '| 非merge / 完整候选 | 65 / 39 PR + 1普通直接提交 + 1无PR合并簇 |',
 '| 净diff / 候选×文件态 | 190文件 +11853 -610 / 217条 |',
 '| ALREADY / CLEAN / CONFLICT / NOBASE | 0 / 135 / 62 / 20 |',
 ])+'\n'

old=(ROOT/'docs/upstream-sync/PORTING-0.2.23.md').read_text()
leftovers=old[old.index('| v0.2.21 第一/第二档'):old.index('\n\n上一轮#8606')]
leftovers=leftovers.replace('usage compaction/requested effort持久化、峰谷/重置锚点仍按需或需schema','usage compaction、峰谷/重置锚点仍按需或需schema；requested effort已由d2488b110/SQLite230完成')

lines=['# 上游 v0.2.24 移植评估（v0.2.23 → v0.2.24）','',
 '评估日期 **2026-11-21**。GitHub官方 Releases API与fetch后的tag核对：最新正式版为[v0.2.24](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.24)，发布于 **2026-11-20 09:19:50 UTC / 17:19:50 上海时间**。冻结时刻见[baseline.json](./evidence-0.2.24/baseline.json)；网页工具未返回内容、gh未安装，使用官方API原文和git，不把搜索失败视为无新版本。',
 '',
 '**建议第一档10簇、第二档15簇，共选30个PR及1个安全合并簇的既有消费者维护。评估完成，两批未实施。** 第一优先是重复字段造成模型/策略/计费分叉，再处理Codex搜索403、OAuth不可逆轮换和Claude/Codex/Grok协议与结算。第一档/第二档表示成本风险分档，不覆盖安全严重性；复杂的安全修复S01应先做。',
 '',
 '- [第一档提案：10簇、37项未勾选任务](../../openspec/changes/port-upstream-0.2.24-tier1/README.md)',
 '- [第二档提案：15簇、52项未勾选任务](../../openspec/changes/port-upstream-0.2.24-tier2/README.md)',
 '- [来源、失败过程、探针与完整核验](./evidence-0.2.24/MANIFEST.md)',
 '',
 '本轮新增评估、原始证据和提案，只更新两处入口文档。没有应用产品补丁、改VERSION/依赖/schema/Ent/Wire、提交/推送/发布/归档或业务实例操作。',
 '', '## 1. 版本、范围与规模','',table_facts(),
 '启动时工作区干净。当前main已包含v0.2.23两档、GPT-6.2、透传模型过滤与requested/effective effort（SQLite230）；不能沿用旧评估的7a0b78fe5/SQLite229。39个PR必须按整PR first-parent→merge取diff；最终安全合并`5c2982a81`的6个非merge也作为完整簇。65个非merge无遗漏、无重复覆盖。217是候选×文件数，区别于190个净变化文件。',
 '', '## 2. 方法、最小前置与迁移编号','',
 '1. fetch upstream tags/prune，冻结API发布时间、tag对象、源码commit、main及fork。main在tag后只有`328db26f4`的VERSION同步，单独记录，不从main补产品。',
 '2. 所有217条先统一reverse、再统一forward，使用同一私有index且仅`--check`；未真正apply。结果[files.tsv](./evidence-0.2.24/files.tsv)、[完整候选与补丁哈希](./evidence-0.2.24/candidates.tsv)可复核。',
 '3. [2269条新增调用扫描](./evidence-0.2.24/added-call-inventory.tsv)和[本地基座复核](./evidence-0.2.24/local-base-symbols.txt)只辅助定位。CLEAN不证明编译、NOBASE不证明无缺陷、ALREADY=0不等于没有手工等价实现。逐簇最小前置和本地名称差异见下文。',
 '4. 核查85份来源测试文件以及76份本地integration首行，其中53份为postgres专用；另运行九包实际test-list和相关基线。探针采用虚拟test-only Go overlay；前端临时测试结束即删除。来源扫描器只在/tmp隔离编译，未挂产品链。',
 '5. 本地最大迁移230，tag范围内上游迁移与依赖manifest/lock均无变。本轮两档没有新增列需求；OAuth CAS须用SQLite事务/outbox，用户默认配额复用现有表。S11实施时应真实生成Wire；不照抄上游PG或引入shadow/新平台字段。若用户将来另选带schema能力，应从实际基线重新分配231起，不能重复230。',
 '6. S01最小前置须明确：共享字段scanner/中间件本地不存在，但实际路由和parser分叉可复现；旧hasUniqueJSONMembers只保护特定上下文。可新增共享拒绝检查而不引入完整group model_allowlist。来源标题提到的group allowlist绕过不能被记成本地已有该功能。',
 '7. 保留SQLite-only/miniredis/simple、固定quota、usage dedup、用户所选Fast/Ultra Fast和GPT-6.2、64MiB请求体预算/40MiB SSE行/8MiB暂存/120s idle。没有读取真实业务账号来假定新功能使用情况。',
 '', '### 全部候选分档','',
 '| 候选 | 非merge | A/C/X/N | 判断与选取范围 |','|---|---:|---|---|']
for c in candidates:
 n=c['candidate'];d=classification[n];s=stats[n]
 decision=('第一档' if d['tier']==1 else '第二档' if d['tier']==2 else '第三档' if d['tier']==3 else '第四档')
 if d['clusters']:
  decision+=' '+ '、'.join(d['clusters'])+'：'+'；'.join(r['title'] for r in review if r['id'] in d['clusters'])
 if d.get('reason'):decision+='。'+d['reason']
 lines.append(f'| {link(n)} | {c["nonmerge_count"]} | '+ '/'.join(s[k] for k in ['ALREADY','CLEAN','CONFLICT','NOBASE'])+f' | {decision} |')
for tier,title in [(1,'## 3. 第一档：窄修复'),(2,'## 4. 第二档：需适配、SQLite或前端的修复')]:
 lines+=['',title,'']
 for r in review:
  if r['tier']!=tier:continue
  lines += [f'### {r["id"]} {r["title"]}（'+ '、'.join(link(c) for c in r['candidates'])+'）','',
   f'- **依据**：{r["evidence"]}',f'- **最小前置**：{r["base"]}',f'- **patch sites**：{r["sites"]}',f'- **行为**：{r["requirement"]}',f'- **验收**：{r["acceptance"]}',f'- **边界**：{r["boundary"]}','']
lines += ['## 5. 按需、排除及旧遗留','',
 '### 5.1 本轮第三/第四档','',
 '第三档为7个PR加#8962的5图扩展：compaction 5m缓存策略、Exa、搜索来源块、ToolSearch原生桥接、Claude窗口费用估算、刷新风险提示、模型ID复制。它们均保留为按需选择；不因服务于Claude/Codex而自动升档。#8804需271行转换器及新增端到端消费者，其普通工具兼容前置并非现有修复所需。',
 '',
 '#8962必须拆开：已有3图上限报400属于F10；3→5图是扩展能力，不默默随行。#8794缓存策略也不是本地5m/1h计量完整性修复的前置。#8803 URL防御只服务新来源块，缺消费者时不单独制造缺陷。',
 '',
 '第四档含VERSION同步、TypeSafe/SystemOne模型映射（两个核心消费者缺失）及DeepSeek Anthropic-native thinking passback（特定native文件缺失）。不能把支持普通API Key映射误当这些原生平台存在。个人支付、退款、充值/购买/续费及其依赖迁移始终排除；本tag候选中未发现支付专属改动，也不做无关支付代码删除。',
 '', '### 5.2 历史遗留重新定界','',
 '11份旧活动change共350个已完成checkbox、0个未完成，等待归档；本轮保持全部原文。以下是当前状态，历史报告的“未实施/按需”只保留其当时时点，不回写。',
 '', '| 旧范围 | 2026-11-21复核状态与保留原因 |','|---|---|',
 '| v0.2.23 两档 | d8d126da9/3652ddb6a已完成70/76任务；Go1.27.4和缓存/换号/metadata修复在基线，不重开。 |',
 '| GPT-6.2 Sol/Astra | ad0967903已完成12项；冻结价格/时钟验收保持原证据。Pro套餐放行、分组强制Ultra Fast、渠道Ultra Fast倍率仍未选，本次音频目录维护不是引入这些策略。 |',
 '| #8748 / #6188 | d2488b110完成透传过滤与requested/effective日志12项，SQLite230。S06/F04是其后续缺口，不能再次列为未实施功能或重复迁移。 |',
 '| v0.2.22–23 按需尾项 | #8644/#8779 Fable读数/冷却局部分支可被自定义映射到达，目录/价卡仍未选；Carbon/Hydrogen/Qwen、Vertex us/eu、ping=0按实际用途。30m/12h独立桶仍需schema，未选。 |',
 '| v0.2.22–23 性能尾项 | #8816账务lane/PGbatch、#8833合并加载/last_used/20s stale grace仍需SQLite事务/dedup/内存测量；本轮只量窄函数，未把完整架构升档。#8693 PG rollup缺本地消费者。 |',
 leftovers,
 '',
 '旧#8610的动态预算最低1GiB仍不适合1GiB实例，本地保留已测64MiB；#8590强查按钮、#8367浮标、Argon完整quota、新图片/cache CRUD、admin Key/完整COW/抢占等继续按用途。#8269 SQLite轻量quota查询未取得足以替代现状的新收益证据；不能继承“缺基座”字样，现有有界读数路径已复核但仍需实际负载测量。跨节点Redis租约在单节点架构不需要。',
 '', '### 5.3 依赖、版本与发布后main','',
 'tag范围没有go.mod/go.sum、前端package/lock和migrations变更；本轮不升级工具链或依赖，也没有跑新的全量漏洞审计。上一轮12 moderate/2 low是历史时点结果，不能当成本轮当前审计。main只有一个tag后的VERSION同步，证据见[post-release-commits.txt](./evidence-0.2.24/post-release-commits.txt)。',
 '', '## 6. 已具备与缺基座的精确结论','',
 '- #8931的每轮原始帧/bridge取值，本地d2488b110已比来源旧快照前进；仅input configuration_update仍有缺口，不整段替换回来源实现。',
 '- #8899依赖的#8748现已实施，应作为现有功能正确性修复；#8870本地也已用raw view，不能声称整份input拷贝优化全是新增收益。',
 '- 本地hasUniqueJSONMembers是限定场景、精确名字检查；S01还需要所有入口、大小写/转义/freeform/multipart和后续WS帧。',
 '- #8859缺来源helper名字而现有Claude5.5家族判定存在；实际Composite改写复现失败，最小约27行helper值得第二档，不判N/A。',
 '- #8918复用SQLite227配额与现有BulkInsertInitial，#8671复用Grok CAS/outbox边界；前者要Wire，后者不能抄PG JSONB。',
 '- #8979本地没有新QuotaSummary接口；只给现有有界fetch补固定端点，不借它引入#8522。#8883也不要求完整group allowlist。',
 '', '## 7. 建议实施顺序','',
 '1. **先S01**，核对HTTP全部别名与WS首/后续帧、freeform兼容和内存成本，确保拒绝发生在调度/计费前。复杂度属于第二档，严重性最高。',
 '2. F01/F02与S02/S05：搜索403、超时换号、不可逆token保存、映射后选号；确认整号/模型作用域和重试上限。',
 '3. F03/F04/F06/F07、S03/S04/S08：Codex上下文/工具与缓冲、每轮effort、图片SSE、Grok成功数和目录能力，串实际HTTP/WS/SQLite结算。',
 '4. S06/S07与F05/F10：编辑不覆盖和已有账号操作正确性；真实浏览器保存重开和旧请求晚到。',
 '5. S09/S10/S11/S14：metadata/Composite/新用户quota/SQLite模型统计，各自独立公式和事务边界。',
 '6. F08/F09/S15及S12/S13：既有Gemini/Antigravity稳定性和内存热点。S13每hunk等价+收益通过才采用；最后两档组合验收，再记录真实实施SHA。发布、推送、归档或部署另行决定。',
 '', '## 8. 本轮自检与证据局限','',
 '| 检查 | 实际结果 |','|---|---|',
 '| 当前基线 | 九包相关检查：869个顶层测试通过、1原跳过；包含子例为1541通过/1跳过。九包test-list实际执行。 |',
 '| 前端基线 | 3文件69测试通过；日志保留现有预期错误输出与Browserslist提示。 |',
 '| 五包初组行为探针 | 编译后13个顶层目标失败、10控制/观察通过；包括实际SQLite、Gin路由与字段解码，未模拟真实业务。 |',
 '| 追加两包探针 | 4个顶层目标失败：标量body alias、Gemini末工具轮、buffered hosted item、实际Composite thinking。 |',
 '| 前端新行为探针 | 透传批量编辑2目标失败、1无关编辑控制通过；临时文件已删除。 |',
 '| 来源scanner | 冻结472行scanner+原test+单独提取multipart helper，隔离85例通过；单P约95399ns/op、1192B/op、14allocs，约225KB固定夹具。 |',
 '| scheduler投影PoC | 合成40组payload 260839→5060字节，成员/调用者图/模型判断控制通过；不代表生产收益。 |',
 '| 文档/提案 | 严格OpenSpec、候选覆盖/链接/未勾任务/空验收槽与最终产品/历史哈希见document-validation.json。 |',
 '',
 '首次Go探针含一个括号拼写错误、图片builder参数与SQLite指针断言夹具不匹配；初始源码/log和修正源码都保留。它们不算产品缺陷或修复验收。后两轮行为探针已全部编译、无panic。共17个目标顶层失败不是完整影响面数量，也不是已修好；F09/F10等部分簇仅有静态可达证据，实施时仍需对应复现。',
 '',
 'source scanner单P数值是当前工作站的函数成本，不是1C1G实例峰值、真实HTTP/WS安全接入验收或生产性能承诺。未跑完整backend/frontend门禁、race、browser或部署；这些留作实施验收。保护8311个初始tracked文件，允许改的旧文件仅AGENTS.md及入口README，旧来源/patch/验收/hash均保持。',
 '', '## 9. 本轮通用教训','',
 '- 安全修复可能在无PR的最终合并内；必须枚举first-parent和全部非merge覆盖，不能只看Release列出的PR。',
 '- 本地缺新的共享入口不等于没有解析器分叉；已有局部防重检查也不等于全入口安全。',
 '- 相同值重复也会在模型/策略改写后分叉；扫描应区分协议字段大小写规则与用户自由键空间。',
 '- token轮换不可逆完成与可选enrichment截止是不同边界；CAS保存、重新授权胜出和租期预算应一起验证。',
 '- 已选第三档能力进入基线后，后续缺口应按当前消费者重新分档；不可重复做schema或覆盖更完整本地实现。',
 '- 速率的累计耗时加权与逐请求算术均值各有消费者；修一个指标不能重写另一个用户已选指标。',
 '- 优化要先量现有实现：raw view可能已存在，真正差异是编辑拷贝、异步标量所有权或展开的组图；CLEAN不是收益或可编译证据。',
 '- 来源带来的新接口（group allowlist、QuotaSummary、shadow）要与实际消费修复拆开；用可测的最小helper替代整套新能力。',
 ]
(ROOT/'docs/upstream-sync/PORTING-0.2.24.md').write_text('\n'.join(lines)+'\n')

for tier in [1,2]:
 rows=[r for r in review if r['tier']==tier];name=f'port-upstream-0.2.24-tier{tier}';folder=ROOT/'openspec/changes'/name;cap=f'upstream-0-2-24-tier{tier}-compatibility';(folder/'specs'/cap).mkdir(parents=True,exist_ok=True)
 tiername='第一档' if tier==1 else '第二档'
 task_count=7+3*len(rows)
 (folder/'README.md').write_text(f'''# v0.2.24 {tiername}（未实施）

评估日期2026-11-21，{len(rows)}簇、{task_count}项实施任务全部未勾选。基线为main `{facts['fork_head']}` / VERSION1.1.26 / Go1.27.4 / SQLite230 / auth21。提案不是实现或发布授权。

- [评估与唯一patch sites](../../../docs/upstream-sync/PORTING-0.2.24.md)
- [冻结来源](source-baseline.md) / [来源行为映射](source-feature-map.md)
- [方案](proposal.md) / [设计](design.md) / [增量规格](specs/{cap}/spec.md)
- [实施任务](tasks.md) / [实施验收空槽](verification.md)
- [仅评估证据](../../../docs/upstream-sync/evidence-0.2.24/MANIFEST.md)

优先已有Claude Code、Codex、Grok及SQLite数据完整性。按需项保持第三档、缺消费者与支付保持第四档。保留本地固定quota/usage dedup及资源预算。后续实施先核对真实起点，不改本提案冻结SHA；填入实际verification。没有推送、Release、归档或业务部署。
''')
 (folder/'proposal.md').write_text('# Why\n\n上游v0.2.24涉及现有网关解析、Claude/Codex/Grok协议与账号数据正确性，需要按SQLite个人部署进行选择性维护。\n\n# What Changes\n\n'+'\n'.join(f'- {r["id"]} {r["title"]}：{r["requirement"]}' for r in rows)+'''\n\n# Impact

不需要新增数据库列。保持VERSION1.1.26、Go1.27.4、SQLite230/auth21和依赖。新增共享入口检查需覆盖所有现有HTTP别名和WS轮帧；OAuth CAS适配SQLite，默认配额消费者实施时重新生成Wire。真实影响面由所选行为决定，不能以文本CLEAN替代编译。

# Non-goals

不导入支付或商业流程、缺失TypeSafe/DeepSeek native平台、新搜索供应商/来源块/ToolSearch能力、5图扩展、费用估算/刷新提示/复制ID、完整group allowlist、QuotaSummary、shadow、PGbatch/lane/COW/stale策略。发布、push、归档和业务实例操作另行处理。
''')
 selection=[c for c in candidates if c['candidate'] in {p for r in rows for p in r['candidates']}]
 source='# Source Baseline（冻结，实施不得回写）\n\n冻结日期2026-11-21；源码范围仅v0.2.23..v0.2.24。真实实施起点填verification。\n\n'+table_facts()+'\n| 候选 | merge/direct | first parent | 完整patch SHA256 |\n|---|---|---|---|\n'
 for c in selection:source+=f'| {link(c["candidate"])} | `{c["sha"]}` | `{c["parent1"]}` | `{c["patch_sha256"]}` |\n'
 source+='\n最终tag优先于中间提交。全部候选先reverse再forward；来源完整性与评估证据见[MANIFEST](../../../docs/upstream-sync/evidence-0.2.24/MANIFEST.md)。不把来源的PG/缺平台接口当成本地实现。\n'
 (folder/'source-baseline.md').write_text(source)
 mapping='# Source Feature Map\n\n| 簇 | 完整来源 | 选定行为与边界 |\n|---|---|---|\n'
 for r in rows:mapping+=f'| {r["id"]} {r["title"]} | '+ '、'.join(link(c) for c in r['candidates'])+f' | {r["requirement"]} {r["boundary"]} |\n'
 mapping+='\npatch sites只保存在[评估报告](../../../docs/upstream-sync/PORTING-0.2.24.md)。新增调用扫描是检索而非编译保证；来源helper与本地consumer区别以该报告的逐簇最小前置为准。\n'
 (folder/'source-feature-map.md').write_text(mapping)
 design='''# Design

按完整PR/安全合并簇理解最终行为，逐个采用本地已有消费者需要的最小hunk。先核对当前HEAD与冻结来源的差异，再构造失败复现和控制。来源新增模块仅在本地消费者确实需要时建立，不携带未选平台与策略。

## Decisions

- 以现有HTTP/WS raw载荷及SQLite事务为边界；调度/计费前检查歧义，稳定保留原始模型、客户端effort与最终实际值。
- 不可逆OAuth轮换完成允许有界保存；并发重新授权由条件更新保护，缓存/outbox只发布持久化状态。不同平台的晚到成功规则分别保持。
- metadata、快照、模型能力取决于可服务账号，聚合需要保守交集；UI异步请求需防止旧结果覆盖新状态。
- 平均模型TPS按累计耗时，已有输出TPS分布仍用逐样本连续分位数与算术mean；同名速率不可混用。
- 性能hunk先做等价/转义/fuzz与单P前后成本，异步字段拥有独立存储。函数级测量不替代1C1G实际实例验收。

## Risks and acceptance

拒绝同值重复字段会改变宽容输入行为，应覆盖大小写、转义、嵌套及freeform合法键。完整入口挂载、后续WS帧、断开结算、CAS并发/锁租期、SQLite取消/错误、缓存旧请求与body生命周期需实际验收。S11需要真实Wire生成；不需要新迁移。参考[任务](tasks.md)与[规格](specs/'''+cap+'''/spec.md)。

## Preserved constraints

SQLite-only/miniredis/simple mode；用户固定quota与usage去重；已选Fast/Ultra Fast、GPT-6.2、过滤与requested/effective effort；64MiB请求体预算、40MiB SSE、8MiB暂存及120s idle。第三档只保留建议，支付与缺平台不实现。组合验收完成前不得填写实施成功槽或勾选任务。
'''
 (folder/'design.md').write_text(design)
 spec='## ADDED Requirements\n\n'
 for r in rows:
  spec+=f'''### Requirement: {r['id']} {r['title']}

{r['requirement']}

#### Scenario: 既有消费者满足目标行为

- **WHEN** {r['acceptance']}
- **THEN** 对应消费者满足上述SHALL行为，并保存实际请求/数据库或UI证据。

#### Scenario: 保持边界与控制行为

- **WHEN** 请求触及相邻或未选路径
- **THEN** {r['boundary']}

'''
 (folder/'specs'/cap/'spec.md').write_text(spec)
 tasks=f'# 实施任务（0/{task_count}，未实施）\n\n- [ ] 核对真实实施起点、冻结来源、产品/历史保护哈希和全部最小前置。\n'
 for r in rows:
  tasks+=f'''\n## {r['id']} {r['title']}

- [ ] 在实际起点建立本簇复现与负对照，区分夹具/来源与产品缺陷。
- [ ] 实现选定行为及最小前置，保留边界：{r['boundary']}
- [ ] 完成验收并保存证据：{r['acceptance']}
'''
 tasks+='''
## 组合验收

- [ ] 全量backend unit、embed build和lint；变化涉及生成代码时真实生成并核对。
- [ ] 相关包race，覆盖并发CAS/缓存/取消/所有权及两档交叉路径。
- [ ] 实际临时HTTP/WS/SQLite/miniredis：迁移230基线、settlement/dedup、固定配额和独立成本。
- [ ] 前端受影响范围的测试/typecheck/lint/build；涉及交互时实际桌面/移动保存重开与旧请求验收。
- [ ] 逐hunk资源测量及1C1G相关边界，完成两档组合控制并记录局限。
- [ ] 记录真实实施SHA、测试命令/结果及最终哈希，核对产品范围和冻结历史，完成verification后再更新任务状态。
'''
 assert len(re.findall(r'^- \[ \]',tasks,re.M))==task_count
 (folder/'tasks.md').write_text(tasks)
 verification='''# Verification（实施验收空槽）

本批尚未实施。评估基线PASS、预期失败探针与独立来源PoC均不是实施验收，下表保持空白。真正实施后记录真实起点和证据，不回写source-baseline。

| 项目 | 真实值/命令 | 结果/证据 |
|---|---|---|
| 实施起点 HEAD/tree、VERSION/Go/SQLite | | |
| 最终产品SHA与变化范围 | | |
| backend unit/embed/lint/codegen | | |
| 相关race | | |
| 实际HTTP/WS/SQLite/miniredis | | |
| quota/settlement/dedup/成本 | | |
| frontend测试/typecheck/lint/build | | |
| 桌面/移动浏览器（适用时） | | |
| 性能/内存及局限 | | |
| 冻结来源/历史哈希与范围 | | |

| 簇 | 实现证据 | 控制及验收结果 |
|---|---|---|
'''
 for r in rows:verification+=f'| {r["id"]} {r["title"]} | | |\n'
 (folder/'verification.md').write_text(verification)
print('Rendered PORTING and two unimplemented OpenSpec batches: 10/15 clusters, 37/52 unchecked tasks.')
