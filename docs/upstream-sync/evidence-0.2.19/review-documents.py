#!/usr/bin/env python3
"""Validate intake coverage, proposal status, local links and protected evidence."""
import csv
import datetime
import hashlib
import json
from pathlib import Path
import re
import subprocess
from urllib.parse import unquote

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def links(content):
    return set(re.findall(r"\[[^\]]+\]\(([^)]+)\)", content))


original = json.loads((OUT / "review-original-file-hashes.json").read_text())
for name, expected in original.items():
    assert sha(OUT / name) == expected, ("original evidence changed", name)
products = json.loads((OUT / "business-file-sha256.json").read_text())
for row in products:
    assert sha(ROOT / row["path"]) == row["sha256"], ("product changed", row["path"])

with (OUT / "classification.tsv").open() as source:
    classified = list(csv.DictReader(source, delimiter="\t"))
with (OUT / "candidates.tsv").open() as source:
    candidates = list(csv.DictReader(source, delimiter="\t"))
assert {row["candidate"] for row in classified} == {row["candidate"] for row in candidates}
assert len(classified) == 23 and sum(int(row["nonmerge_count"]) for row in classified) == 50

checks = ("review-claude-js-oracle", "review-test-list", "review-targeted", "review-baseline-probes", "review-grok-audit-probes",
          "review-frontend", "review-terra-neighbors", "review-terra-frontend",
          "review-dependency-feasibility", "review-openspec-tier1", "review-openspec-tier2", "review-openspec-terra")
for name in checks:
    assert json.loads((OUT / (name + ".json")).read_text())["exit_code"] == 0, name

required = {"README.md", "proposal.md", "source-baseline.md", "source-feature-map.md",
            "design.md", "tasks.md", "verification.md"}
proposals = []
documents = [ROOT / "docs/upstream-sync/PORTING-0.2.19.md", OUT / "MANIFEST.md",
             OUT / "call-review.md", OUT / "feature-selection.md"]
for folder in sorted((ROOT / "openspec/changes").glob("port-upstream-0.2.19-*")):
    assert required <= {file.name for file in folder.iterdir()}
    markdown = sorted(folder.rglob("*.md"))
    documents.extend(markdown)
    for file in markdown:
        assert not re.search(r"^\s*[-*]\s+\[[xX]\]", file.read_text(), re.M), file
    tasks = len(re.findall(r"^- \[ \]", (folder / "tasks.md").read_text(), re.M))
    assert tasks > 0
    for line in (folder / "verification.md").read_text().splitlines():
        if not line.startswith("|"):
            continue
        cells = [cell.strip() for cell in line.strip("|").split("|")]
        if all(re.fullmatch(r"[-:]+", cell) for cell in cells) or cells[0] in ("实施信息", "编号", "行为", "检查"):
            continue
        empty_from = 1 if len(cells) == 2 else 2 if len(cells) == 5 else 1
        assert not any(cells[empty_from:]), ("implementation slot filled", folder.name, line)
    specs = sorted((folder / "specs").rglob("spec.md"))
    requirements = sum(file.read_text().count("### Requirement:") for file in specs)
    scenarios = sum(file.read_text().count("#### Scenario:") for file in specs)
    assert specs and requirements and scenarios >= requirements
    baseline = (folder / "source-baseline.md").read_text()
    assert "49939fb2cf96633688548ca39204d96558e59d79" in baseline
    assert "dece26153621959ccd7635814f4b2d08304e681a" in baseline
    assert "227" in baseline and "1.1.20" in baseline and "1.27.3" in baseline
    proposals.append({"change": folder.name, "markdown_files": len(markdown),
                      "specs": len(specs), "requirements": requirements, "scenarios": scenarios,
                      "unchecked_tasks": tasks, "source_baseline_sha256": sha(folder / "source-baseline.md")})
assert len(proposals) == 3

linked = 0
closing = OUT / "document-validation.md"
for file in documents + [ROOT / "AGENTS.md", ROOT / "docs/upstream-sync/README.md"]:
    content = file.read_text()
    for line_number, line in enumerate(content.splitlines(), 1):
        assert line == line.rstrip(" \t"), ("trailing whitespace", str(file), line_number)
    targets = links(content)
    if file.name == "AGENTS.md" or file == ROOT / "docs/upstream-sync/README.md":
        previous = subprocess.check_output(["git", "show", "HEAD:" + str(file.relative_to(ROOT))], cwd=ROOT, text=True)
        targets -= links(previous)
    for target in targets:
        if re.match(r"^[a-zA-Z][\w+.-]*:", target) or target.startswith("#"):
            continue
        target = unquote(target.split("#", 1)[0].split("?", 1)[0].strip("<>"))
        resolved = (file.parent / target).resolve()
        assert resolved.exists() or resolved == closing, ("broken local link", str(file), target)
        linked += 1

tracked = subprocess.check_output(["git", "diff", "--name-only"], cwd=ROOT, text=True).splitlines()
assert set(tracked) <= {"AGENTS.md", "docs/upstream-sync/README.md"}, tracked
subprocess.run(["git", "diff", "--check"], cwd=ROOT, check=True)
assert not subprocess.check_output(["git", "diff", "--cached", "--name-only"], cwd=ROOT).strip()

readme = ROOT / "docs/upstream-sync/README.md"
content = readme.read_text()
item = "- [ ] 三提案严格校验、链接/来源/空实施槽与最终范围哈希核对。"
completed = item.replace("[ ]", "[x]", 1)
assert item in content or completed in content
if item in content:
    readme.write_text(content.replace(item, completed, 1))

record = {"checked_at_utc": datetime.datetime.now(datetime.timezone.utc).isoformat(),
          "status": "passed", "assessment_only": True, "original_files_checked": len(original),
          "product_files_checked": len(products), "candidates": len(classified), "nonmerge_commits": 50,
          "local_link_checks": linked, "proposals": proposals, "baseline_checks": list(checks),
          "changed_tracked_files": tracked,
          "source_baseline_sha256": {row["change"]: row["source_baseline_sha256"] for row in proposals},
          "document_sha256": {str(file.relative_to(ROOT)): sha(file) for file in documents + [ROOT / "AGENTS.md", readme]}}
(OUT / "review-document-validation.json").write_text(json.dumps(record, ensure_ascii=False, indent=2) + "\n")
closing.write_text("# v0.2.19 文档与范围验收\n\n"
                  "完成日期：2026-10-29（来源冻结于2026-10-28）。结果：通过。当前仍为评估，业务代码未移植。\n\n"
                  "| 检查 | 结果 |\n|---|---|\n"
                  f"| 23候选与50非merge全覆盖 | 通过 |\n"
                  f"| 原始{len(original)}证据及{len(products)}产品文件哈希 | 完全一致 |\n"
                  f"| 新增/修改本地链接 {linked} 项及新文档格式 | 通过 |\n"
                  "| 三份OpenSpec严格校验 | 全部通过 |\n"
                  "| 63项实施任务 | 全未勾选 |\n"
                  "| 实施verification字段/证据/结果 | 全为空 |\n"
                  "| 现有tracked修改 | 仅AGENTS.md和同步README |\n\n"
                  "| 提案 | 规格 | requirements | scenarios | 待执行任务 |\n|---|---:|---:|---:|---:|\n" +
                  "\n".join(f"| {row['change']} | {row['specs']} | {row['requirements']} | {row['scenarios']} | {row['unchecked_tasks']} |" for row in proposals) +
                  "\n\n[检查脚本](./review-documents.py)、[原始结果/文档哈希](./review-document-validation.json)。\n\n"
                  "OpenSpec输出：[第一档](./review-openspec-tier1.txt)、[第二档](./review-openspec-tier2.txt)、[Terra](./review-openspec-terra.txt)。\n\n"
                  "后端10包定向/实际测试目录、Terra邻近三包、12个基线probe及2个额外Grok probe、前端40+21项与隔离MVS可行性通过。"
                  "这些是现状/前置评估证据，不是新行为实施验收。生产audit17条仍未处理，未跑当前全量开发audit。\n\n"
                  "冻结source-baseline、原始TSV/patch/历史验证未改；VERSION1.1.20、Go1.27.3、SQLite227和产品依赖/Ent/Wire保持。未commit/push/tag/部署。\n")
assert closing.exists()
print(json.dumps({key: record[key] for key in ("status", "original_files_checked", "product_files_checked", "local_link_checks")}, ensure_ascii=False))
print(json.dumps(proposals, ensure_ascii=False))
