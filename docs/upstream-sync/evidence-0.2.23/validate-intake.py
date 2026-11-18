"""Validate intake scope, frozen sources, proposal state and local links."""
import ast
import csv
import hashlib
import json
import re
import subprocess
from pathlib import Path

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]

def git(*args):
    return subprocess.check_output(["rtk","proxy","git",*args],cwd=ROOT).decode()

def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

facts = json.loads((OUT/"baseline.json").read_text())
protected = json.loads((OUT/"protected-files.json").read_text())
allowed = {"AGENTS.md","docs/upstream-sync/README.md"}
changed = [name for name,sha in protected.items()
    if not (ROOT/name).is_file() or digest(ROOT/name) != sha]
assert set(changed) == allowed, changed
assert git("rev-parse","HEAD").strip() == facts["fork_head"]
assert git("diff","--cached","--name-only") == ""
assert set(git("diff","--name-only").splitlines()) == allowed
prefixes = (
    "docs/upstream-sync/evidence-0.2.23/",
    "docs/upstream-sync/PORTING-0.2.23.md",
    "openspec/changes/port-upstream-0.2.23-tier1/",
    "openspec/changes/port-upstream-0.2.23-tier2/",
)
untracked = git("ls-files","--others","--exclude-standard").splitlines()
assert all(name.startswith(prefixes) for name in untracked), untracked
candidates = list(csv.DictReader((OUT/"candidates.tsv").open(),delimiter="\t"))
assert len(candidates) == 82 and sum(int(c["nonmerge_count"]) for c in candidates) == 128
for c in candidates:
    assert digest(OUT/"patches"/(c["candidate"]+".patch")) == c["patch_sha256"], c["candidate"]
classification = json.loads((OUT/"classification.json").read_text())
assert set(classification) == {c["candidate"] for c in candidates}
review = json.loads((OUT/"review.json").read_text())
assert {r["id"] for r in review if r["tier"]==1} == {f"F{i:02}" for i in range(1,22)}
assert {r["id"] for r in review if r["tier"]==2} == {f"S{i:02}" for i in range(1,24)}
assert len({c for r in review for c in r["candidates"]}) == 55
summary = json.loads((OUT/"probe-summary.json").read_text())
assert summary["local_failed_top_level"] == 26
assert summary["local_passed_top_level"] == 20
assert all(v["compiled"] and not v["panic"] for v in summary.values() if isinstance(v,dict))
for script in OUT.glob("*.py"):
    ast.parse(script.read_text())

files = [ROOT/"docs/upstream-sync/PORTING-0.2.23.md",OUT/"MANIFEST.md"]
task_counts = {}
for tier, expected in [(1,70),(2,76)]:
    folder = ROOT/"openspec/changes"/f"port-upstream-0.2.23-tier{tier}"
    for name in ["README.md","proposal.md","source-baseline.md","source-feature-map.md","design.md","tasks.md","verification.md"]:
        assert (folder/name).is_file()
    tasks = (folder/"tasks.md").read_text()
    assert not re.search(r"\[[xX]\]",tasks)
    task_counts[str(tier)] = len(re.findall(r"^- \[ \]",tasks,re.M))
    assert task_counts[str(tier)] == expected
    for line in (folder/"verification.md").read_text().splitlines():
        if not line.startswith("|") or line.startswith("|---"):
            continue
        cells = [v.strip() for v in line.split("|")[1:-1]]
        if cells[0] not in ["项目","簇"]:
            assert all(not cell for cell in cells[1:]), line
    source = (folder/"source-baseline.md").read_text()
    assert facts["fork_head"] in source
    assert all(facts["refs"][ref] in source for ref in ["v0.2.21","v0.2.22","v0.2.23"])
    assert json.loads((OUT/f"openspec-tier{tier}.json").read_text())["exit_code"] == 0
    files += list(folder.rglob("*.md"))
broken = []
checked = 0
for file in files:
    for target in re.findall(r"\]\(([^)]+)\)",file.read_text()):
        if target.startswith(("http:","https:","#")):
            continue
        target = target.split("#")[0]
        if not target:
            continue
        checked += 1
        if not (file.parent/target).exists():
            broken.append([str(file.relative_to(ROOT)),target])
assert not broken, broken
result = dict(
    fork_head=facts["fork_head"],protected_files=len(protected),
    allowed_modified_tracked=changed,unexpected_modified=[],
    real_index_unchanged=True,untracked_files=len(untracked),
    candidate_count=82,nonmerge_count=128,patch_hashes_valid=True,
    first_tier_clusters=21,second_tier_clusters=23,selected_prs=55,
    tasks_unchecked=task_counts,verification_evidence_empty=True,
    openspec_strict={"tier1":0,"tier2":0},local_links_checked=checked,
    broken_links=broken,product_changed=False,history_changed=False,
    version_dependency_schema_generated_unchanged=True,
)
(OUT/"document-validation.json").write_text(json.dumps(result,ensure_ascii=False,indent=2)+"\n")
print(json.dumps(result,ensure_ascii=False,indent=2))
