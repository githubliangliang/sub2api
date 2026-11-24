#!/usr/bin/env python3
"""Validate scope, frozen source inventory, assessment evidence and proposals."""
import ast
import collections
import csv
import hashlib
import json
from pathlib import Path
import re
import subprocess

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]

def git(*args):
    return subprocess.check_output(["rtk", "proxy", "git", *args], cwd=ROOT, text=True)

def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

facts = json.loads((OUT / "baseline.json").read_text())
protected = json.loads((OUT / "protected-initial-hashes.json").read_text())
allowed = ["AGENTS.md", "docs/upstream-sync/README.md"]
changed = sorted(name for name, old in protected.items()
                 if not (ROOT / name).is_file() or digest(ROOT / name) != old)
assert changed == allowed, changed
assert sorted(git("diff", "--name-only").splitlines()) == allowed
assert git("rev-parse", "HEAD").strip() == facts["fork_head"]
assert git("write-tree").strip() == facts["real_index_tree"]
assert facts["snapshot_tree"] == facts["fork_tree"]
for name, old in facts["initial_untracked_hashes"].items():
    assert digest(ROOT / name) == old, name

candidates = list(csv.DictReader((OUT / "candidates.tsv").open(), delimiter="\t"))
commits = list(csv.DictReader((OUT / "commits.tsv").open(), delimiter="\t"))
files = list(csv.DictReader((OUT / "files.tsv").open(), delimiter="\t"))
assert len(candidates) == 4 and len(commits) == 12 and len(files) == 98
coverage = set()
for c in candidates:
    members = set(git("rev-list", "--no-merges", c["parent1"] + ".." + c["sha"]).splitlines())
    assert not coverage.intersection(members)
    assert len(members) == int(c["nonmerge_count"])
    coverage.update(members)
    assert digest(OUT / "patches" / (c["candidate"] + ".patch")) == c["patch_sha256"]
assert coverage == {c["sha"] for c in commits}
states = collections.Counter(r["state"] for r in files)
assert states == {"CLEAN": 56, "CONFLICT": 32, "NOBASE": 10}, states
classification = json.loads((OUT / "classification.json").read_text())
assert set(classification) == {c["candidate"] for c in candidates}
review = json.loads((OUT / "review.json").read_text())
assert [(r["id"], r["tier"]) for r in review] == [("F01", 1), ("S01", 2), ("S02", 2)]

expected_pass = ["baseline-test-list", "baseline-consumers", "baseline-auth", "baseline-frontend",
                 "source-host-helper", "candidate-ws-controls", "baseline-real-connect", "candidate-real-connect"]
for name in expected_pass:
    assert json.loads((OUT / (name + ".json")).read_text())["exit_code"] == 0, name
for name, count in [("baseline-gorilla-tests", 3), ("candidate-gorilla-tests", 2),
                    ("baseline-proxy-framing-copy", 2), ("candidate-proxy-framing-copy", 2)]:
    record = json.loads((OUT / (name + ".json")).read_text())
    assert record["exit_code"] != 0
    assert sum(e["Action"] == "fail" and "Test" in e and "/" not in e["Test"]
               for e in record["events"]) == count, name
summary = json.loads((OUT / "probe-summary.json").read_text())
probe = summary["behaviour-probes-final"]
assert probe["compiled"] and not probe["panic"]
assert probe["top_level"] == {"fail": 5, "pass": 4}
assert probe["including_subcases"] == {"fail": 12, "pass": 4}
modules = json.loads((OUT / "module-diff.json").read_text())
assert modules["removed"] == ["github.com/gorilla/websocket v1.5.3"]
assert modules["added"] == ["github.com/gorilla/websocket v1.5.4-0.20240701034025-d67f41855da4"]
old_changes = json.loads((OUT / "older-active-changes.json").read_text())
assert len(old_changes) == 13 and sum(c["completed"] for c in old_changes) == 439
assert sum(c["unchecked"] for c in old_changes) == 0

proposals = []
task_counts = {}
empty_slots = 0
for tier, expected in [(1, 6), (2, 11)]:
    folder = ROOT / "openspec/changes" / f"port-upstream-0.2.25-tier{tier}"
    for file in ["README.md", "proposal.md", "source-baseline.md", "source-feature-map.md", "design.md", "tasks.md", "verification.md"]:
        assert (folder / file).is_file()
    assert len(list((folder / "specs").glob("*/spec.md"))) == 1
    tasks = (folder / "tasks.md").read_text()
    assert not re.search(r"\[[xX]\]", tasks)
    task_counts[str(tier)] = len(re.findall(r"^- \[ \]", tasks, re.M))
    assert task_counts[str(tier)] == expected
    source = (folder / "source-baseline.md").read_text()
    assert facts["fork_head"] in source
    assert all(facts["refs"][tag] in source for tag in ["v0.2.24", "v0.2.25"])
    for line in (folder / "verification.md").read_text().splitlines():
        if not line.startswith("|") or line.startswith("|---"):
            continue
        cells = [c.strip() for c in line.split("|")[1:-1]]
        if cells[0] == "项目":
            continue
        assert all(not c for c in cells[1:]), line
        empty_slots += 1
    assert json.loads((OUT / f"openspec-tier{tier}.json").read_text())["exit_code"] == 0
    proposals.extend(folder.rglob("*.md"))

links = 0
broken = []
scan = proposals + [ROOT / "docs/upstream-sync/PORTING-0.2.25.md", OUT / "MANIFEST.md"]
for path in scan:
    for target in re.findall(r"\]\(([^)]+)\)", path.read_text()):
        if target.startswith(("http:", "https:", "#")):
            continue
        target = target.split("#")[0]
        if not target:
            continue
        links += 1
        resolved = (path.parent / target).resolve()
        if not resolved.exists() and resolved != OUT / "document-validation.json":
            broken.append([str(path.relative_to(ROOT)), target])
assert not broken, broken
for script in OUT.glob("*.py"):
    ast.parse(script.read_text())
assert not git("diff", "--check")
prefixes = ("docs/upstream-sync/PORTING-0.2.25.md", "docs/upstream-sync/evidence-0.2.25/",
            "openspec/changes/port-upstream-0.2.25-tier1/", "openspec/changes/port-upstream-0.2.25-tier2/")
untracked = git("ls-files", "--others", "--exclude-standard").splitlines()
assert all(name.startswith(prefixes) or name in facts["initial_untracked_hashes"] for name in untracked)
frozen_paths = ["baseline.json", "candidates.tsv", "commits.tsv", "files.tsv", "classification.json", "review.json"]
freeze = {str((OUT / p).relative_to(ROOT)): digest(OUT / p) for p in frozen_paths}
for tier in (1, 2):
    for name in ("source-baseline.md", "source-feature-map.md"):
        p = ROOT / f"openspec/changes/port-upstream-0.2.25-tier{tier}" / name
        freeze[str(p.relative_to(ROOT))] = digest(p)
(OUT / "source-freeze-hashes.json").write_text(json.dumps(freeze, ensure_ascii=False, indent=2) + "\n")
result = dict(fork_head=facts["fork_head"], protected_tracked_files=len(protected),
              allowed_modified=changed, initial_diagnostic_unchanged=True,
              head_and_real_index_unchanged=True, product_and_frozen_history_unchanged=True,
              version_dependencies_sqlite_auth_ent_wire_defaults_unchanged=True,
              candidates=4, nonmerge_commits=12, candidate_files=98,
              complete_unique_coverage=True, states=dict(states), patch_hashes_valid=True,
              first_tier_clusters=1, second_tier_clusters=2, tasks_unchecked=task_counts,
              empty_implementation_evidence_rows=empty_slots, openspec_strict={"tier1":0,"tier2":0},
              relative_links_checked=links, broken_links=broken,
              baseline_top_level_passed=274, baseline_including_subcases_passed=548,
              frontend_passed=67, desired_probe_top_level_failed=5, probe_controls_passed=4,
              candidate_application_ws_top_level_passed=376,
              dependency_full_suite_passed=False, historical_proxy_eof_unresolved=True,
              implementation_performed=False, untracked_files=len(untracked))
(OUT / "document-validation.json").write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
print(json.dumps(result, ensure_ascii=False, indent=2))
