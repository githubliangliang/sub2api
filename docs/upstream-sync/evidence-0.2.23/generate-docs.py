"""Generate consistent intake documents from frozen facts and reviewed scope."""
import csv
import json
import re
from pathlib import Path

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]
facts = json.loads((OUT / "baseline.json").read_text())
review = json.loads((OUT / "review.json").read_text())
classification = json.loads((OUT / "classification.json").read_text())
candidates = list(csv.DictReader((OUT / "candidates.tsv").open(), delimiter="\t"))
by_name = {c["candidate"]: c for c in candidates}
states = {r["candidate"]: r for r in csv.DictReader((OUT / "candidate-stats.tsv").open(), delimiter="\t")}
summary = json.loads((OUT / "probe-summary.json").read_text())
per_release = json.loads((OUT / "per-release-stats.json").read_text())

def q(value):
    return chr(96) + str(value) + chr(96)

def pr(name):
    if name.startswith("PR"):
        return f"[#{name[2:]}](https://github.com/Wei-Shaw/sub2api/pull/{name[2:]})"
    return f"[{name}](https://github.com/Wei-Shaw/sub2api/commit/{by_name[name]['sha']})"

def links(row):
    return "、".join(pr(n) for n in row["candidates"])

def write(path, text):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text.rstrip()+"\n")

baseline_table = f"""| 项目 | 冻结值 |
|---|---|
| fork HEAD / tree | {q(facts['fork_head'])} / {q(facts['fork_tree'])} |
| VERSION / Go / SQLite / auth snapshot | {q('1.1.25 / 1.27.3 / 229 / 21')} |
| private index tree | {q(facts['snapshot_tree'])} |
| v0.2.21 commit | {q(facts['refs']['v0.2.21'])} |
| v0.2.22 commit / tag object | {q(facts['refs']['v0.2.22'])} / {q(facts['tag_objects']['v0.2.22'])} |
| v0.2.23 commit / tag object | {q(facts['refs']['v0.2.23'])} / {q(facts['tag_objects']['v0.2.23'])} |
| upstream/main（观察，不作来源） | {q(facts['refs']['upstream/main'])} |
| 非merge / 候选 / 净diff | 128 / 78 PR + 4直接提交 / 399文件 +25947 -2093 |
| 候选×文件四态 | 524：ALREADY0 / CLEAN269 / CONFLICT199 / NOBASE56 |
"""

for tier in [1, 2]:
    rows = [r for r in review if r["tier"] == tier]
    name = f"port-upstream-0.2.23-tier{tier}"
    folder = ROOT / "openspec/changes" / name
    label = "第一档" if tier == 1 else "第二档"
    selected = sorted({n for r in rows for n in r["candidates"]})
    source_table = "| PR | merge | first parent | patch SHA256 |\n|---|---|---|---|\n"
    for n in selected:
        c = by_name[n]
        source_table += f"| {pr(n)} | {c['sha']} | {c['parent1']} | {c['patch_sha256']} |\n"
    write(folder/"README.md", f"""# v0.2.21 → v0.2.23 {label}（未实施）

2026-11-18评估完成；本批{len(rows)}簇、选取{len(selected)}个PR中的既有消费者维护。实施任务全部未勾选，verification证据槽为空。

冻结本地基线是v1.1.25加用量排行路由修复 {q(facts['fork_head'])}，Go1.27.3、SQLite229、auth snapshot21。

- [范围、patch sites与最小前置](../../../docs/upstream-sync/PORTING-0.2.23.md)
- [冻结来源](source-baseline.md) / [来源行为映射](source-feature-map.md)
- [方案](proposal.md) / [设计](design.md) / [规格](specs/upstream-0-2-23-tier{tier}-compatibility/spec.md)
- [实施任务](tasks.md) / [实施验收空表](verification.md)
- [评估证据](../../../docs/upstream-sync/evidence-0.2.23/MANIFEST.md)

评估的PASS、预期失败探针和源码适用性判断不是实施验收。新模型/平台、强制Ultra Fast、渠道Ultra Fast倍率、30m/12h schema和大规模并发策略独立按需；支付功能排除。本方案不授权发布、推送、归档或业务实例操作。
""")
    changes = "\n".join(f"- {r['id']} {r['title']}：{r['behaviour']}" for r in rows)
    write(folder/"proposal.md", f"""# Why

新的v0.2.22–v0.2.23合并内容涉及本地已有Claude Code、Codex、Grok及共享网关消费者。本批按实际存在的路径和最小前置选择，保留SQLite单节点用途。

# What Changes

{changes}

# Impact

{label}共{len(rows)}簇。不新增产品schema；第二档的Go1.27.4维护需单独对齐工具和CI，库版本无需随上游升级。保持64MiB请求体预算、40MiB SSE行、120s读写idle和既有8MiB首输出暂存边界。真实实施起点和结果填verification。

# Non-goals

不引入支付、退款、充值/购买/续费及其依赖/迁移；不导入未选模型/平台、30m/12h桶、强制Ultra Fast、新渠道倍率、账务lane或完整COW。发布、推送、归档和业务实例操作不在本方案。
""")
    write(folder/"source-baseline.md", f"""# Source Baseline（冻结）

冻结日期2026-11-18。实施不得改写本文件；真实起点另填verification。

{baseline_table}

{source_table}

整PR first parent→merge为来源；最终tag状态优先于中间提交和发布摘要。全部文件先reverse后forward，未应用产品补丁。[完整来源与评估证据](../../../docs/upstream-sync/evidence-0.2.23/MANIFEST.md)。
""")
    mapping = "| 簇 | 来源 | 所需行为 | 评估依据 |\n|---|---|---|---|\n"
    for r in rows:
        mapping += f"| {r['id']} {r['title']} | {links(r)} | {r['behaviour']} | {r['evidence']} |\n"
    write(folder/"source-feature-map.md", "# 来源行为映射\n\n"+mapping+"""

补丁位置和最小前置只记录在评估报告。本表定义选取的行为；同PR未选分支不构成授权。完整候选和四态见冻结证据。
""")
    design_rows = "\n".join(f"- {r['id']}：{r['boundary']}" for r in rows)
    write(folder/"design.md", f"""# 设计与边界

以现有路由、账号、配置和SQLite写入链为基座，按行为适配。优先完成认证/数据一致性和Claude/Codex/Grok协议问题，再处理相关UI及有测量依据的热点维护。

每次attempt/WS轮次独立观察上游模型、用量和错误；先观察原文再还原公开模型。keepalive不等同于语义输出，但HTTP已提交时错误必须继续为SSE。首输出暂存有界，取消、预算、槽位和会话租约只有一个释放责任。

SQLite事务覆盖一次操作的所有清空、删除和写入；旧TEXT时间与DATETIME扫描保持兼容，固定quota窗口、软删除和usage去重不被重做。不复制PG-only SQL或新增无消费者迁移。

{design_rows}

评估探针只证明冻结基线上的明确断言，静态项仍需实际复现。实施时按簇建立负对照和真实HTTP/WS/SQLite/miniredis证据；前端完成真实浏览器和移动端验收。性能项逐hunk测语义/分配/等待/资源，Go维护完成完整工具链和代码生成兼容检查。不得用本轮编译或旧实施证据替代本批验收。
""")
    tasks = "# 实施任务（全部未实施）\n\n- [ ] 核对实际实施起点与冻结来源，保存保护文件和依赖/生成代码哈希。\n"
    specs = "## ADDED Requirements\n"
    verification = "# 实施验收（空表）\n\n当前仅评估和提案完成；下表全部空白，等待实际实施。\n\n| 项目 | 命令/结果 | 证据 |\n|---|---|---|\n"
    for label2 in ["实际实施起点","实现SHA/产品哈希","VERSION/Go/SQLite/auth snapshot","backend全量unit/embed/lint/race","真实HTTP/WS/SQLite/miniredis","frontend完整门禁/浏览器/移动端","性能/资源/工具链测量","依赖/生成代码/历史保护","遗留与边界"]:
        verification += f"| {label2} | | |\n"
    verification += "\n| 簇 | 实际结果 | 证据 |\n|---|---|---|\n"
    for r in rows:
        tasks += f"\n## {r['id']} {r['title']}\n\n"
        tasks += "- [ ] 在实际基线复现/核实本簇并建立负对照，区分夹具问题与产品行为。\n"
        tasks += f"- [ ] 实现选定行为及最小前置，遵守边界：{r['boundary']}\n"
        tasks += f"- [ ] 验收并保存结果：{r['acceptance']}\n"
        specs += f"\n### Requirement: {r['id']} {r['title']}\n\n{r['behaviour']}\n\n"
        specs += f"#### Scenario: 选定既有路径\n\n- **WHEN** {r['title']}对应的既有请求或操作发生\n- **THEN** 系统满足选定行为，并验证{r['acceptance']}\n\n"
        specs += f"#### Scenario: 保护相邻行为\n\n- **WHEN** 请求或操作触及相邻或未选分支\n- **THEN** {r['boundary']}\n"
        verification += f"| {r['id']} {r['title']} | | |\n"
    tasks += """

## 组合门禁

- [ ] backend全量unit、embed build、lint和相关race通过，保留完整命令/输出。
- [ ] 实际HTTP/WS/SQLite/miniredis验证结算/成功去重、固定quota窗口、旧时间格式和完整Fast保护。
- [ ] 涉及前端完成typecheck/lint/test/build/critical及真实浏览器/移动端；不涉及项记录不适用。
- [ ] 性能/资源/Go项完成对应实际测量；无收益或无法闭合生命周期的优化保留原路径并说明。
- [ ] 核对所有接口/测试mock、依赖、Ent/Wire、默认值和历史文件保护；只在实际需要时生成。
- [ ] 填写实际实现SHA、逐簇结果和遗留，严格校验OpenSpec并完成范围核对。
"""
    write(folder/"tasks.md", tasks)
    write(folder/"verification.md", verification)
    write(folder/f"specs/upstream-0-2-23-tier{tier}-compatibility/spec.md", specs)

report = f"""# 上游 v0.2.23 移植评估（v0.2.21 → v0.2.23，含v0.2.22）

评估日期2026-11-18；范围冻结于 {facts['frozen_at_utc']}。GitHub API与git fetch确认最新正式版为[v0.2.23](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.23)，发布于11-18 03:51:33 UTC / 11:51:33上海时间；同时覆盖[v0.2.22](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.22)，发布于11-17 02:48:12 UTC / 10:48:12上海时间。

建议**第一档21簇、第二档23簇，选取55个不同PR的既有消费者维护**。当前为**评估完成，两批提案未实施**。优先缓存计量/结算、Claude token轮换、Codex工具/上下文、Grok搜索/错误及首输出换号。新模型、强制/渠道Ultra Fast配置和大规模并发策略仍按需；支付排除。

- [第一档提案](../../openspec/changes/port-upstream-0.2.23-tier1/README.md)
- [第二档提案](../../openspec/changes/port-upstream-0.2.23-tier2/README.md)
- [来源、测量与失败过程](./evidence-0.2.23/MANIFEST.md)

本次只新增评估/提案并更新入口文档。没有产品、VERSION、依赖或schema改动，没有commit/push/release、旧变更归档或业务实例操作。

## 1. 版本、范围与规模

{baseline_table}

| 范围 | 非merge | PR/直接 | 候选文件态 A/C/X/N | 净diff |
|---|---:|---|---|---|
"""
for r in per_release:
    c = r["states"]
    report += f"| {r['range']} | {r['nonmerge']} | {r['prs']}/{r['direct']} | {c.get('ALREADY',0)}/{c.get('CLEAN',0)}/{c.get('CONFLICT',0)}/{c.get('NOBASE',0)} | {r['net_diff']} |\n"
report += """

开始时工作区干净。v0.2.21两档已在v1.1.25；最新fork提交7a0b78fe5恢复用量排行，必须保留。524是候选×文件之和，不是399净文件。相同PR里的新能力只独立分档，不因选取修复就整PR导入。

## 2. 方法、渠道与编号

1. GitHub网页检索工具多次502，改用官方releases API与本地git源码核对。[release元数据](./evidence-0.2.23/releases.json)、[基线](./evidence-0.2.23/baseline.json)保留tag对象、源码commit及main观察值。按tag祖先划范围，不按作者日期漏掉旧分支。
2. 82个候选以first parent→merge整体diff覆盖128个非merge。[candidates.tsv](./evidence-0.2.23/candidates.tsv)、[commits.tsv](./evidence-0.2.23/commits.tsv)、原补丁SHA256保存来源。所有524条先反向、再正向检查同一私有index；未应用产品补丁，真实index不变。
3. [逐文件四态](./evidence-0.2.23/files.tsv)只表达文本可应用性。[5096条新增调用扫描](./evidence-0.2.23/added-call-inventory.tsv)含内置/方法/测试，仅作候选检索；逐簇实际消费者与前置如下。CLEAN不是可编译，NOBASE不是没有缺陷，ALREADY=0不是所有行为都没做过。
4. [193份来源测试记录](./evidence-0.2.23/source-test-inventory.tsv)、[76份本地integration首行](./evidence-0.2.23/integration-build-tags.json)和实际test-list分开。53份integration文件为postgres专用，不冒充SQLite执行。
5. 本地最大迁移229。选定两档不要求新DB列/迁移：5m/1h和账号成本列已有。上游tag实际有两个不同的251文件（PG rollup、force-ultrafast）以及252渠道倍率；不能认为同号只代表一种能力，也不能照抄PG SQL。若另行选择两个Ultra Fast配置，按当时基线重新分配230起的本地编号并验证Ent/auth快照；本轮不生成迁移。
6. 保持SQLite-only/miniredis/simple mode、固定quota、usage dedup、完整Fast/Terra及64MiB/40MiB/120s默认值。评估没有读取真实业务账号，按需项不假定已使用。
7. 发布摘要与最终代码有实质差异：[PR8822来源对照](./evidence-0.2.23/release-source-differences.json)显示b6c21cd5e删除强制缓存计费，8d57f461c在同PR恢复。最终tag仍有ForceCacheBilling；本轮只选稳定前缀哈希，不按摘要改计费。库版本/价格若未独立查官方资料，结论仅是与冻结tag或本地控制一致。

### 全部候选分档

| 候选 | 非merge | A/C/X/N | 主分档与选取范围 |
|---|---:|---|---|
"""
for c in candidates:
    name = c["candidate"]
    s = states[name]
    cl = classification[name]
    state = "/".join(s[k] for k in ["ALREADY","CLEAN","CONFLICT","NOBASE"])
    judgment = f"第{['','一','二','三','四'][cl['tier']]}档"
    if cl["cluster"]:
        judgment += " "+cl["cluster"]+" "+cl["reason"]
    else:
        judgment += "："+cl["reason"]
    if name == "PR8812": judgment += "；group force-ultrafast为第三档。"
    if name == "PR8904": judgment += "；只选Ollama，DeepSeek缺失消费者不导入。"
    report += f"| {pr(name)} | {c['nonmerge_count']} | {state} | {judgment} |\n"

for tier, section in [(1,"3. 第一档：窄修复和默认值维护"),(2,"4. 第二档：跨路径与SQLite适配")]:
    report += "\n## "+section+"\n"
    for r in review:
        if r["tier"] != tier: continue
        report += f"\n### {r['id']} {r['title']}（{links(r)}）\n\n"
        report += f"- **依据/最小前置**：{r['evidence']} {r['prerequisites']}\n"
        report += f"- **patch sites**：{r['patch_sites']}。\n"
        report += f"- **行为与验收**：{r['behaviour']} 验收：{r['acceptance']}。\n"
        report += f"- **边界**：{r['boundary']}\n"

old = (ROOT/"docs/upstream-sync/PORTING-0.2.21.md").read_text()
old_table = old.split("| 旧范围 | 2026-11-11当前状态及保留原因 |",1)[1].split("\n### 5.4",1)[0]
old_rows = "\n".join(line for line in old_table.splitlines() if line.startswith("|") and not line.startswith("|---"))
report += """

## 5. 按需项、排除项与旧遗留

### 5.1 本轮第三/第四档

全部理由在候选表。重点选择边界：

- GPT-6.2 Sol/Astra、Carbon/Hydrogen、Qwen新兜底、Fable quota恢复和Vertex多区域需实际模型/套餐/端点选择。Fable读数/冷却分支本地已有，但目录/价格未选；有自定义映射实际使用时重评，不说绝对不可达。
- 既有Sol Ultra Fast倍率修复可独立摘取；强制Ultra Fast和渠道独立倍率是新增策略/字段，仍按需。GPT-6.2促销时钟/套餐放行不借道进入。
- #8816新增544行PG batch repo和311行lane combiner；#8833追加stale grace/合并加载/last_used缓存。单用户高并发和大bucket收益需SQLite/miniredis、事务/dedup/取消/内存独立测量。S15成员正确性、S22窄热点不需要两套完整架构。
- #8678关闭ping是新控制；协议错误兼容不能靠关闭默认ping逃避。#8693的PG rollup/锁/迁移不在本地消费链。
- #8649支付退款整项排除。个人用量/账号额度/上游订阅元数据的维护仍保留；不进行无关支付代码删除。

### 5.2 旧轮次当前口径

以下依据提交/归档和当前任务清单，不宣称本轮重新执行了旧批全部验收。六份旧活动change共180个完成checkbox、零未完成，等待归档；本轮不改它们。新增两份提案的146任务全部未实施。

| 旧范围 | 2026-11-18当前状态及保留原因 |
|---|---|
| v0.2.21 第一/第二档 | 63c764704、83729b130已实现，39/39和42/42，发布v1.1.25；SQLite229覆盖索引/auth21和资源默认保留。 |
"""
report += old_rows + """

上一轮#8606 30m/12h独立usage/价格/schema仍按需，与本轮5m/1h完整性修复不同。#8610动态预算最小1GiB不适合1GiB实例，保留已测64MiB；#8590新强查按钮、#8367保存浮标、Argon quota/新图片/cache CRUD及历史admin Key/完整COW/平台/抢占策略继续按实际用途。Fable相关修复本轮明确重述已存在的局部分支，未自动升档。WS默认1.0、固定quota和SQLite后台恢复的用户选择继续保留。

### 5.3 依赖与tag后main

上游范围只有Go1.27.3→1.27.4和jsonschema-go0.4.2由indirect改direct；go.sum/前端manifest和lock未变。本地也已含该库版本，不能以修schema测试为由升级整个依赖图。当前生产审计12 moderate/2 low/0 high/0 critical，退出1代表仍有advisory；[本轮实际记录](./evidence-0.2.23/dependency-summary.json)不替代Go和开发依赖全扫。

[tag后main](./evidence-0.2.23/post-release-commits.txt)仅2af3f1e15的VERSION同步，单独观察，不纳入产品范围。发布、推送和部署另行处理。

## 6. 已具备行为、N/A和前置结论

- #8651跨组RPM与#8740 Gemini刷新留行已在本地手工实现；#8825的强制DeepSeek峰谷策略本地不存在，当前0/正数/缓存读override控制通过。原补丁反向不通过不能重开这些问题。
- #8664账号namespace、#8735完整raw OAuth metadata清理、#8598在途预占测试基座缺失；原缺陷由其特定改写/架构产生，不能引入整套功能来修。
- #8837虽以Seedance为标题，本地chat/embeddings/Responses能力消费者真实存在，丢投影键可独立复现，明确列F18。
- #8772与本地null-type sanitizer不是同一基座；已有工具请求消费者但需新增窄regexp处理，明确列第二档，不能把来源约899行文件全覆盖。#8754同样要手工补raw局部重建调用点。
- F14两个Grok提取器、S07搜索predicate、S11工具boundary、S12格式helper都可独立适配，不导入原生新平台/模型。源码符号扫描见完整inventory，实际前置以逐簇依据为准。

## 7. 建议实施顺序

1. S03 SQLite原子配置、F09登录guard、S02token恢复及S01缓存/结算先处理共享数据和认证正确性。
2. F01–F06/F21与S04–S08优先Claude/Codex/Grok：工具内容、权限、上下文、首输出/idle预算、声明ID、Grok搜索、原文模型。
3. S09–S11/S21处理Composite/effort/Bedrock/schema边界；保留真实模型观察和已映射调度键，验证普通4xx/工具负对照。
4. F08/F10/F11/F14/F18及S16–S20处理既有倍率、quota/快照/成本/启动/传输；不改变固定窗口或带入新价卡。
5. Gemini/Antigravity兼容与相关UI按同文件依赖顺序落地；F16/F17/F19完成保存重开/失败/nonce验收。默认模型维护独立验证。
6. S22每个性能hunk先做语义和资源测量，补转义保护；S23工具链单独统一。两批组合门禁及实际SHA完成后再考虑归档/发布。

## 8. 已执行检查与实施门槛

| 检查 | 结果/实际限制 |
|---|---|
| 来源/四态 | 82候选覆盖128非merge，524文件全部先反后正；真实index未变。 |
| backend发现/定向 | 5611个匹配测试被列出，10包命令退出0；web在该正则无测试，另做embed专用探针。不是新产品全量实施验收。 |
| frontend基线 | 7文件89测试通过。 |
| 冻结产品Go overlay | 6包实际编译，26个顶层目标行为失败、20个顶层对照通过；含真实modernc SQLite/Ent和embed HTML handler。 |
| 临时前端目标测试 | 2文件9失败/11通过；是保存对象类型和星号输入断言，不是浏览器验收。 |
| 性能源码PoC | Ollama9组字节对照通过；200条assistant、GOMAXPROCS=1，7.924855ms/12560039B/2319alloc→0.341573ms/261461B/627alloc。仅函数夹具，不宣称1GiB实例总堆/生产收益。来源escaped-usage谓词目标断言失败，需人工保护。 |
| Go1.27.4可行性 | 当前产品只读embed构建、config/apicompat/claude三包通过；当前go.mod仍1.27.3，未实施新CI/lint/codegen。 |
| 生产依赖审计 | 12 moderate/2 low；无依赖变更。 |
| 文档/范围 | 见document-validation.json：两提案严格校验、任务未勾选/验收空槽、所有历史和产品哈希核对。 |

过程失败与最终结果分开保存：服务首轮count_tokens夹具遗漏生产DI的Config而panic，修正后才统计产品断言；来源whitelist测试位置不同需移植测试；pnpm在临时链接目录尝试依赖核验被其安全guard拒绝，CLI负选项也未接受，最终直接用node运行已由pnpm安装的Vitest，无依赖安装/改写；Go官方下载首次网络超时，镜像下载后真实编译通过。上述不计入产品缺陷数。

静态项明确标注，尚无新token并发401、双账号idle/WS、Bedrock错误全集、Grok跨协议搜索、schema真实远端或实际浏览器/移动验收。本轮未触及真实业务实例。

实际移植后必须执行backend全量unit/embed/lint/相关race，真实HTTP/WS/SQLite/miniredis结算+成功去重、固定quota/旧时间格式/Fast/预算/释放检查；前端完整门禁和浏览器保存重开/快速切换/失败/移动端；S22测原文等价和边界，S23做Go/CI/lint/Ent/Wire统一兼容。评估结果不填入实施verification。

## 9. 本轮通用教训

- 发布摘要可能被同PR后续提交推翻；完整PR最终diff与tag优先，Gemini强制缓存策略是实证。
- 原补丁ALREADY=0也可能已经手工修过语义；用对照和实际predicate确认RPM/Gemini/DeepSeek。
- 缺完整改写基座有时意味着没有原缺陷；有小helper缺失但真实消费者存在时则需窄适配。
- 缓存presence、5m/1h明细、聚合差额和互斥输入桶需联合验证，不能只加JSON字段。
- SSE性能predicate的字节匹配会漏JSON转义字段；源码PoC通过语义和资源检查后才可采用。
- 首内容hold需沿用有界stage；keepalive提交HTTP后仍可安全换号的前提是错误保持SSE、账号头仍私有。
- SQLite配置同步必须原子覆盖删除和写入；成功路径通过不代表失败后旧数据可恢复。
- 临时前端的依赖工具可能主动安装；测试隔离也要确认runner，过程失败不能冒充产品断言。
- Go新patch能编译当前产品不等于完成新工具链/代码生成/CI验收；未来验收必须独立记录。
"""
write(ROOT/"docs/upstream-sync/PORTING-0.2.23.md", report)

manifest = """# v0.2.23 评估证据

本目录是2026-11-18 intake，不是产品实施验收。基线7a0b78fe511f17b0f7eb9d26fcf5e4ff083312c7 / VERSION1.1.25 / Go1.27.3 / SQLite229 / auth21。

- [冻结事实](baseline.json)、[初始事实](initial-facts.json)、[release元数据](releases.json)、[逐release统计](per-release-stats.json)
- [全部候选](candidates.tsv)、[成员提交](commits.tsv)、[逐文件状态](files.tsv)、[原补丁](patches/)、[新增调用扫描](added-call-inventory.tsv)、[hunk](hunks.tsv)
- [评审行为/前置](review.json)、[完整分档](classification.json)、[发布摘要与最终源码差异](release-source-differences.json)
- [测试发现](baseline-test-list.log)、[backend定向](baseline-targeted.json)、[frontend89测试](baseline-frontend.json)、[来源测试](source-test-inventory.tsv)、[本地integration标签](integration-build-tags.json)
- [Go探针摘要](probe-summary.json)、[6包测试源码](probes/)、[Go overlay复现](reproduce-probes.py)
- [前端9失败/11通过](frontend-probe.log)、[复现](reproduce-frontend-probes.py)、[初始位置/runner问题](frontend-probe-initial-apply.json)
- [性能源码来源](performance-source.json)、[语义与转义检查](performance-equivalence.log)、[函数benchmark](performance-benchmark.log)、[复现](reproduce-performance.py)
- [Go1.27.4 embed](go1274-embed-feasibility.json)、[三包unit](go1274-unit-feasibility.json)、[首次网络失败](go1274-embed-initial-network.log)
- [生产audit与依赖范围](dependency-summary.json)、[raw audit](frontend-production-audit.log)、[tag后main](post-release-commits.txt)
- [原始保护哈希](protected-files.json)、[最终范围校验](document-validation.json)、[本目录SHA256](SHA256SUMS)

Go顶层26失败/20通过为旧产品面对目标断言的结果；源码SSE predicate另有1失败，不属于本地产品缺陷计数。临时前端仅修改测试，保存对象/输入断言9失败、既有11通过；没有浏览器验收。性能使用单P函数夹具，非1GiB生产压力。Go1.27.4只读编译不修改当前1.27.3声明，也不完成未来工具链维护。

服务Config夹具错误、前端测试位置/pnpm问题和下载网络问题完整保留；最终期望失败没有被产品patch修成PASS。Go overlays和临时前端复制不会修改真实源码/index。

可重跑collect-intake.py的前提是HEAD仍为冻结基线；它只检查同一私有index。重跑各reproduce脚本会覆盖本目录对应测量记录，应先另存记录。两个OpenSpec的实际实施checkbox/verification仍为空。
"""
write(OUT/"MANIFEST.md", manifest)
print(json.dumps({"report": "docs/upstream-sync/PORTING-0.2.23.md",
    "clusters": {str(t): sum(r["tier"]==t for r in review) for t in [1,2]},
    "tasks_unchecked": {str(t): 7+3*sum(r["tier"]==t for r in review) for t in [1,2]}}))
