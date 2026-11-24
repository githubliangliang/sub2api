#!/usr/bin/env python3
"""Render the explicitly reviewed intake and unimplemented OpenSpec batches."""
import csv
import json
from pathlib import Path

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]
BASE = json.loads((OUT / "baseline.json").read_text())
CANDIDATES = list(csv.DictReader((OUT / "candidates.tsv").open(), delimiter="\t"))

REVIEW = [
    dict(id="F01", tier=1, title="透传上下文错误的账号凭据与主机名脱敏", candidates=["PR9045"],
         evidence="当前 Forward→handleErrorResponsePassthrough 直接保留上下文超限文案；query 参数已净化，但配置的主机名与独立 api_key 仍出现在下游。隔离目标失败。",
         base="sanitizeUpstreamErrorMessage、Account.GetCredential、现有错误信封均在；新增账号感知脱敏和安全大小写替换即可，无新供应商或依赖。",
         sites="backend/internal/service/openai_gateway_passthrough.go：handleErrorResponsePassthrough、writeOpenAIPassthroughErrorEnvelope；复用现有 query 净化，不全局改变其它平台策略。",
         behavior="已有 OpenAI 透传错误向客户端保留上游消息时 SHALL 隐去配置的账号凭据与上游主机名，并保留有助于客户端恢复的非敏感错误语义。",
         acceptance="真实本地 HTTP 上游返回上下文超限；API Key/access_token/refresh_token、主机名大小写、IPv6/无 base_url、Ⱥ/Kelvin/非法 UTF-8、query 凭据、正则元字符、无匹配和普通 403 对照；不得 panic 或多写响应。",
         boundary="只修改面向客户端的错误文本，不改变管理员内部取证、账号惩罚、换号、模型/计费或响应头净化；不引入上游中间提交中以 ToLower 下标切原串的错误实现。"),
    dict(id="S01", tier=2, title="透传结构化错误、规则与 compact 终态一致性", candidates=["PR9045"],
         evidence="400 invalid_encrypted_content/context_length_exceeded 的 type/code/param 被重建丢弃；配置的 422/自定义文案未应用；compact 心跳后 code 变为 upstream_error；关闭详情日志时仅匹配原始 code 的 skip_monitoring 未命中。四个目标失败。",
         base="ErrorPassthroughService.MatchRule、runtime wrapper、compact keepalive/bridge、容量可重试码均在。需完整提取匹配规则、有限标量字段和 compact writer。保留本地 NextOpenAIResponsesSequence，不能照抄来源 sequence_number=0。",
         sites="backend/internal/service/error_passthrough_runtime.go；openai_gateway_passthrough.go：handleErrorResponsePassthrough 及标准字段信封；openai_compact_stream_bridge.go：终态 writer；ops_upstream_context.go 为已有监控控制，非整体重写。",
         behavior="OpenAI 透传最终错误 SHALL 应用既有管理员规则，并在允许的确定性 400 或 passthrough_body 规则下保留经过脱敏和限长的标准标量字段；compact 已提交心跳后 SHALL 仅输出有序 response.failed 终态，保留确定性 code 并将容量降载 code 改为可重试值。",
         acceptance="HTTP/compact 心跳前后，规则优先级/平台/原状态与原 body、passthrough_body、自定义状态/文案、无规则、错误 body 标量类型/256 字节/内嵌 JSON/空白 code；监控/账号副作用、终态序号、取消/停止/无重复写、失败不计成功用量和 SQLite 结算去重控制。",
         boundary="F01 为脱敏前置；保留既有 failover/池重试和成功流语义。状态码匹配的 skip_monitoring 已在事件链生效；容量码重写是新增结构化 compact 路径的配套保护，不把它们整体列成当前缺陷。"),
    dict(id="S02", tier=2, title="Gorilla WebSocket 定向安全依赖维护", candidates=["PR9008"],
         evidence="本地 v1.5.3 命中官方 GO-2026-6278 范围；newMaskKey 使用 math/rand。唯一生产 import 为管理员 ops WS 服务端，弱客户端 mask 机制没有确认的生产调用；不是已复现的 Codex 网关漏洞。",
         base="官方固定修复为 v1.5.4-0.20240701034025-d67f41855da4；现有 Go1.27.4/x/net0.62.0 足够。临时 modfile MVS 仅变化这一模块，但模块内容含 18 文件差异及代理/握手行为改变。",
         sites="backend/go.mod、backend/go.sum；backend/internal/handler/admin/ops_ws_handler.go 及现有 WS 消费者仅验证，不为升级扩展生产调用。",
         behavior="定向维护 SHALL 使用官方修复的 Gorilla mask 随机源并保持现有管理员 WS 鉴权、单一 subprotocol、连接释放及网关互通；最终模块选择与未解决的依赖测试差异 SHALL 可核验。",
         acceptance="核对官方漏洞范围、模块来源/校验和/MVS，复核完整模块差异；实际管理员 WS 鉴权/Origin/限额/控制帧/关闭，HTTP CONNECT 与 Basic-auth 隧道、现有网关 HTTP/WS 及 targeted race、build；保留并定因依赖包历史代理测试失败。",
         boundary="独立摘取混合 PR 的依赖子项，保持 Go/VERSION/schema/前端依赖；不随带名称同步。两项依赖包代理测试在旧/候选版本均 EOF，真实 CONNECT 对照通过仍不能冒称依赖全量绿灯或已验证利用。"),
]

CLASSIFICATION = {
    "328db26f4": dict(tier="第四档", decision="上游 VERSION 同步，fork 使用独立 1.1.x 编号；不移植。"),
    "PR9008": dict(tier="第二档 S02 + 第三档", decision="只建议独立的 Gorilla 安全维护；include=group、手动/自动名称备注同步及其文字校验留按需。文字入库修复服务新增同步路径，本地未写入这些字段。"),
    "PR9045": dict(tier="第一档 F01 + 第二档 S01", decision="选择现有 OpenAI 透传错误的脱敏、标准字段、规则与 compact 配套维护；不把来源新 helper 的 Unicode panic 当成本地已有 panic。"),
    "PR9052": dict(tier="第三档", decision="身份变量、编辑器、错误详情、认证 scope 和 WS pool 身份分桶是新增完整能力；审计修复随新能力评估，不作为现有静态头覆写的安全修复自动引入。"),
}

(OUT / "classification.json").write_text(json.dumps(CLASSIFICATION, ensure_ascii=False, indent=2) + "\n")
(OUT / "review.json").write_text(json.dumps(REVIEW, ensure_ascii=False, indent=2) + "\n")

def write(path, text):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text.rstrip() + "\n")

def baseline_table():
    rows = [
        ("fork HEAD / tree", f"`{BASE['fork_head']}` / `{BASE['fork_tree']}`"),
        ("VERSION / Go / SQLite / auth snapshot", "`1.1.27 / 1.27.4 / 230 / 21`"),
        ("私有 index tree", f"`{BASE['snapshot_tree']}`"),
        ("v0.2.24 commit / tag object", f"`{BASE['refs']['v0.2.24']}` / `{BASE['tag_objects']['v0.2.24']}`"),
        ("v0.2.25 commit / tag object", f"`{BASE['refs']['v0.2.25']}` / `{BASE['tag_objects']['v0.2.25']}`"),
        ("upstream/main（仅观察）", f"`{BASE['refs']['upstream/main']}`"),
        ("非 merge / 整体候选", "12 / 3 整 PR + 1 VERSION 直接提交"),
        ("净 diff / 候选×文件状态", "91 文件 +3486 -319 / 98 条"),
        ("ALREADY / CLEAN / CONFLICT / NOBASE", "0 / 56 / 32 / 10"),
    ]
    return "| 项目 | 冻结值 |\n|---|---|\n" + "\n".join(f"| {a} | {b} |" for a, b in rows)

for tier in (1, 2):
    items = [r for r in REVIEW if r["tier"] == tier]
    name = f"port-upstream-0.2.25-tier{tier}"
    folder = ROOT / "openspec/changes" / name
    capability = f"upstream-0-2-25-tier{tier}-compatibility"
    label = "第一档" if tier == 1 else "第二档"
    task_rows = []
    if tier == 1:
        task_rows = [
            ("1. 基线与范围", ["记录真实实施 HEAD、工作区和产品哈希；确认 F01 最小范围、已有诊断与旧冻结证据保留。"]),
            ("2. F01 客户端脱敏", ["复现真实透传上下文错误消息的配置主机名/凭据泄露和已工作的 query 净化对照。", "实现最小账号感知净化与 Unicode/非法 UTF-8 安全替换，保留非敏感错误语义。", "验证实际本地 HTTP、凭据/URL 边界、非匹配与 403 对照、取消/单次写和现有错误处理。"]),
            ("3. 组合验收", ["通过相关 backend unit/race、全量 unit、embed build/lint 和现有 SQLite 用量/结算去重/固定配额控制；仅在实际改动前端时补对应门禁。", "填写实际命令、失败与通过、最终哈希和范围核验；verification 完成后再勾选任务。"]),
        ]
    else:
        task_rows = [
            ("1. 基线与依赖", ["记录真实实施 HEAD、工作区、模块和产品哈希；确认 F01 脱敏前置及本地序号/副作用基座。"]),
            ("2. S01 透传错误一致性", ["复现结构化 400、规则状态/文案、仅原始 body 监控匹配和 compact 心跳后 code 缺口；保留已工作控制。", "实现匹配规则回传、按原状态/原 body 应用规则和有限标准标量信封，接入 F01 脱敏。", "实现 compact 已提交后的有序单终态，保留确定性 code、容量可重试改码及本地 NextOpenAIResponsesSequence。", "验证平台/优先级/自定义文案/无规则、数字/对象/数组/布尔/256 字节、空白与内嵌 code、敏感字段/响应头及非 400 对照。", "验证实际本地 HTTP/compact 心跳前后、停止/取消/无重复写、监控/账号副作用、失败不计成功用量和 SQLite 账本去重。"]),
            ("3. S02 安全依赖维护", ["核对官方 GO-2026-6278、固定修复来源/校验和及完整模块差异；单独处理旧/候选代理夹具 EOF，不隐藏失败。", "只更新 Gorilla 精确修复版本及必要 go.sum，记录最终 MVS；保持 Go、其它依赖、VERSION、schema 与前端不变。", "验证实际管理员 WS 鉴权/Origin/单一 subprotocol/连接限额/控制帧/释放、HTTP CONNECT/Basic-auth 和现有网关互通及 race。"]),
            ("4. 组合验收", ["完成 backend 全量 unit、embed build/lint、相关 race、SQLite 方言/用量结算去重/固定配额控制；若实际涉及前端再运行对应门禁。", "回填 verification 的真实命令、库夹具局限、结果与最终哈希，核对旧冻结文件、第三档/支付边界及无发布/业务实例操作。"]),
        ]
    total = sum(len(rows) for _, rows in task_rows)
    write(folder / "README.md", f"""# 上游 v0.2.25 {label}

状态：**提案未实施**。评估日期 2026-11-24；{len(items)} 簇，{total} 项实施任务均未勾选。当前基线为 v1.1.27 / `{BASE['fork_head']}`。

来源与 patch sites 见[评估报告](../../../docs/upstream-sync/PORTING-0.2.25.md)；只读基线、期望失败探针及依赖可行性见[证据索引](../../../docs/upstream-sync/evidence-0.2.25/MANIFEST.md)。这些记录不填写实施验收槽。

{"F01 先消除既有错误文本的敏感配置泄露，再由另一档扩展结构化错误处理。" if tier == 1 else "S01 依赖 F01；S02 的模块版本维护与错误处理独立评估，实施时仍需完成验收。"} 第三档按需能力、支付及业务实例操作不包含在本提案中。

[行为规格](specs/{capability}/spec.md) · [任务](tasks.md) · [待填写验收](verification.md) · [冻结来源](source-baseline.md)
""")
    write(folder / "proposal.md", "# Why\n\n" + (
        "当前透传上下文错误只净化 query 参数，配置主机名和账号密钥仍会随可操作错误消息交给客户端。需要最小范围修复现有输出。" if tier == 1 else
        "既有 OpenAI 透传路径丢失客户端恢复所需的标准错误字段、忽略部分规则结果，compact 已提交后也丢失确定性 code。依赖清单另包含官方已确认受影响的 Gorilla 版本，需要有边界的安全维护。"
    ) + "\n\n# What Changes\n\n" + "\n".join(f"- {r['id']} {r['title']}：{r['behavior']}" for r in items) + "\n\n# Impact\n\n" + (
        "新增标准库级错误净化 helper，接入既有客户端错误写回。不需要迁移、生成代码、依赖或前端改动。" if tier == 1 else
        "S01 涉及规则 runtime、错误信封与 compact 生命周期；S02 仅允许 Gorilla 精确版本维护及必要 go.sum。临时模块的编译/测试不是实施验收。"
    ) + " 保持 SQLite230/auth21、VERSION1.1.27、Go1.27.4 和既有资源默认。\n\n# Non-goals\n\n身份变量、WS pool 身份分桶、上游名称备注同步、include=group、group i18n、新平台/模型、支付、发布/push/归档和业务实例操作均不在本批范围。")
    selected = {c for r in items for c in r["candidates"]}
    source_rows = [c for c in CANDIDATES if c["candidate"] in selected]
    write(folder / "source-baseline.md", "# Source Baseline（冻结，实施不得回写）\n\n冻结日期 2026-11-24；完整来源范围仅 v0.2.24..v0.2.25。真实实施起点填 verification。\n\n" + baseline_table() + "\n\n| 完整候选 | merge / first parent | 完整 patch SHA256 |\n|---|---|---|\n" + "\n".join(f"| {c['candidate']} | `{c['sha']}` / `{c['parent1']}` | `{c['patch_sha256']}` |" for c in source_rows) + "\n\n同一整 PR 被按独立行为分档；完整 diff 是审查边界，不代表整 PR 获选。S02 只取 PR9008 的依赖维护，不带名称同步；来源安全修复与本地适用性分别判断。全部 98 文件先统一 reverse 再 forward，仅 check。完整记录见[评估证据](../../../docs/upstream-sync/evidence-0.2.25/MANIFEST.md)。")
    write(folder / "source-feature-map.md", "# Source Feature Map\n\n| 簇 | 完整来源 | 行为与边界 |\n|---|---|---|\n" + "\n".join(f"| {r['id']} {r['title']} | " + ", ".join(f"[#{c[2:]}](https://github.com/Wei-Shaw/sub2api/pull/{c[2:]})" for c in r["candidates"]) + f" | {r['behavior']} {r['boundary']} |" for r in items) + "\n\npatch sites 只在[PORTING 报告](../../../docs/upstream-sync/PORTING-0.2.25.md)。新增调用清单是检索索引，不证明编译。来源测试和本地未来验收分开。")
    write(folder / "design.md", "# Context\n\n基线 v1.1.27 已包含 v0.2.24 两档，SQLite230、auth snapshot21；13 份旧活动 change / 439 项任务已完成，保持原文等待另行归档。\n\n# Decisions\n\n" + "\n\n".join(f"## {r['id']} {r['title']}\n\n{r['base']}\n\n{r['boundary']}\n\n验收：{r['acceptance']}" for r in items) + "\n\n# Resource and Data Boundaries\n\n保持请求体预算64MiB、SSE行40MiB、暂存8MiB、读写idle120s。错误净化发生在错误路径；先限制可输出字段、复用有界错误体，不扩展成功请求的拷贝/缓存。保持固定配额、独立账号成本与用户费用、usage dedup、Fast 和 requested/effective effort。无需新迁移/Ent/Wire。\n\n# Verification\n\n评估证据只证明当前缺口与来源可行性。实施应记录实际 HEAD、最终文件哈希、所有失败/通过和测试范围；不以 source helper、临时 modfile 或旧发布验收冒充新实现。只对实际改动面增加测试。")
    write(folder / "tasks.md", "# Tasks（未实施）\n\n" + "\n\n".join("## " + title + "\n\n" + "\n".join("- [ ] " + row for row in rows) for title, rows in task_rows))
    write(folder / "verification.md", "# Verification（实施后填写）\n\n状态：提案未实施。评估基线与探针位于[独立证据](../../../docs/upstream-sync/evidence-0.2.25/MANIFEST.md)，不填入下面的实施证据槽。\n\n| 项目 | 实际命令 / 内容哈希 | 结果 / 局限 |\n|---|---|---|\n" + "\n".join(f"| {row} | | |" for row in ["真实实施起点与最终哈希", "backend unit / build / lint", "相关 race", "实际 HTTP / WS / SQLite / miniredis", "依赖 / 资源及测试局限", "前端（仅实际涉及时）", "历史 / 第三档 / 支付边界", *[r["id"] for r in items]]))
    if tier == 1:
        spec = """# Upstream 0.2.25 Tier 1 Compatibility

## ADDED Requirements

### Requirement: Account-aware client error redaction

Existing OpenAI passthrough error messages exposed to a client SHALL redact configured account credentials and upstream host names while retaining non-sensitive recovery information and the existing query-parameter redaction.

#### Scenario: Context-window message contains account identity
- **WHEN** an upstream context-window error embeds the configured host or account credential
- **THEN** the client SHALL receive the actionable context error without those sensitive values
- **AND** error status, account side effects and strict response-header filtering SHALL retain their established behavior

#### Scenario: Unicode case folding changes byte length
- **WHEN** the message contains case-folding characters, invalid UTF-8, or literal regular-expression characters beside a host
- **THEN** replacement SHALL avoid panic and remove the configured host without slicing by transformed byte offsets
- **AND** a failed host-pattern interpretation SHALL fail closed to a generic message

#### Scenario: Generic forbidden error remains sanitized
- **WHEN** an unrelated forbidden upstream error has no matched disclosure rule
- **THEN** the existing generic sanitized response SHALL remain in effect and SHALL be written once
"""
    else:
        spec = """# Upstream 0.2.25 Tier 2 Compatibility

## ADDED Requirements

### Requirement: Passthrough deterministic error fidelity

Existing OpenAI passthrough final errors SHALL preserve sanitized, bounded standard scalar fields for allowed deterministic 400 errors and administrator passthrough-body rules without forwarding arbitrary upstream fields or sensitive headers.

#### Scenario: Invalid encrypted content or context window error
- **WHEN** an allowed upstream 400 contains type, code, param and message
- **THEN** the client SHALL receive the applicable standard fields with account-aware redaction
- **AND** object, array, boolean and oversized identifier fields SHALL be rejected rather than serialized as text

#### Scenario: Blank code wraps an embedded error
- **WHEN** a missing, null or whitespace code accompanies a supported embedded error message
- **THEN** the outward code SHALL agree with the existing compatible error classification
- **AND** explicit malformed structured code values SHALL not enable arbitrary-field disclosure

### Requirement: Administrator error rule consistency

The passthrough final-error path SHALL apply existing administrator rules using the original upstream status, body and account platform, retaining rule precedence, explicit custom messages and monitoring semantics.

#### Scenario: Custom response or passthrough body is configured
- **WHEN** an applicable rule selects a custom status/message or a passthrough body
- **THEN** the configured result SHALL determine the final client error within the established sanitized envelope
- **AND** unmatched errors and non-400 disclosure boundaries SHALL retain their existing behavior

#### Scenario: Monitoring rule matches a code outside the extracted message
- **WHEN** detail logging is disabled and skip_monitoring matches only the original structured error code
- **THEN** matching SHALL still use that original body and SHALL set the established monitoring marker
- **AND** already-working status-rule suppression SHALL remain effective

### Requirement: Compact failure lifecycle fidelity

After compact keepalive commits HTTP 200, final error handling SHALL emit exactly one ordered response.failed terminal event, preserve deterministic client codes, and rewrite capacity-shed codes to the established retryable client code.

#### Scenario: Encrypted-content error follows heartbeat
- **WHEN** the heartbeat has committed and the upstream returns invalid_encrypted_content
- **THEN** the client SHALL receive a single failed terminal event with that code and sanitized message
- **AND** sequence numbers SHALL advance through the existing request sequence allocator

#### Scenario: Capacity code reaches structured compact output
- **WHEN** structured error forwarding would expose server_is_overloaded or slow_down after heartbeat
- **THEN** the terminal client code SHALL be retryable while internal monitoring retains the original classification
- **AND** stopping, cancellation and failure accounting SHALL preserve existing ownership and deduplication

### Requirement: Bounded Gorilla security maintenance

The selected dependency maintenance SHALL use an official Gorilla fix for GO-2026-6278 and SHALL preserve existing administrator WebSocket security and interoperability while recording the exact module graph and test limitations.

#### Scenario: Exact security candidate is evaluated
- **WHEN** the Gorilla version is changed
- **THEN** its source, checksums, transitive module selection and full behavioral diff SHALL be verified
- **AND** the change SHALL not import upstream group synchronization or unrelated dependencies

#### Scenario: Existing WebSocket consumers remain compatible
- **WHEN** administrator or existing gateway WebSocket traffic executes with the selected module
- **THEN** authentication, Origin, single subprotocol selection, limits, control frames and connection release SHALL retain their established behavior
- **AND** genuine CONNECT and Basic-auth tunnel controls SHALL pass while historical dependency-fixture failures remain explicitly recorded and investigated
"""
    write(folder / "specs" / capability / "spec.md", spec)

print(json.dumps({"reviewed_clusters":len(REVIEW),"first_tier":1,"second_tier":2,"tasks_unchecked":{"tier1":6,"tier2":11}},ensure_ascii=False))
