#!/usr/bin/env python3
"""Verify the selected implementation and preserve all other initial files."""
import hashlib
import json
from pathlib import Path
import re
import subprocess

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]
initial = json.loads((OUT / "initial.json").read_text())
final = json.loads((OUT / "final-backend-hashes.json").read_text())
entries = json.loads((OUT / "entry-additions.json").read_text())


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


selected_docs = {
    f"openspec/changes/port-upstream-0.2.25-tier{tier}/{name}.md"
    for tier in (1, 2) for name in ("README", "tasks", "verification")
}
allowed = set(final["hashes"]) | selected_docs | set(entries)
missing = [name for name in initial["hashes"] if not (ROOT / name).is_file()]
changed = {name: digest(ROOT / name) for name, old in initial["hashes"].items()
           if (ROOT / name).is_file() and digest(ROOT / name) != old}
unexpected = sorted(set(changed) - allowed)
assert not missing, missing
assert not unexpected, unexpected
for name, expected in final["hashes"].items():
    assert digest(ROOT / name) == expected, name
for name, entry in entries.items():
    content = (ROOT / name).read_text()
    assert content.count(entry) == 1, name
    restored = content.replace(entry, "", 1).encode()
    assert hashlib.sha256(restored).hexdigest() == initial["hashes"][name], name

current = set(subprocess.check_output([
    "git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"
], cwd=ROOT).decode().rstrip("\0").split("\0"))
new = current - set(initial["hashes"])
unexpected_new = sorted(name for name in new if name not in allowed and
                        not name.startswith("docs/upstream-sync/implementation-0.2.25/"))
assert not unexpected_new, unexpected_new

old_changes = json.loads((ROOT / "docs/upstream-sync/evidence-0.2.25/older-active-changes.json").read_text())
for change in old_changes:
    tasks = (ROOT / "openspec/changes" / change["change"] / "tasks.md").read_text()
    assert len(re.findall(r"^- \[x\]", tasks, re.M)) == change["completed"], change
    assert not re.search(r"^- \[ \]", tasks, re.M), change
selected_tasks = {}
for tier, expected in ((1, 6), (2, 11)):
    path = ROOT / f"openspec/changes/port-upstream-0.2.25-tier{tier}/tasks.md"
    text = path.read_text()
    completed = len(re.findall(r"^- \[x\]", text, re.M))
    assert completed == expected and not re.search(r"^- \[ \]", text, re.M)
    selected_tasks[f"tier{tier}"] = completed

# Only new implementation documentation and the selected mutable change docs.
broken_links = []
docs = list(OUT.rglob("*.md")) + [ROOT / name for name in selected_docs]
for doc in docs:
    for target in re.findall(r"\[[^\]]*\]\(([^)]+)\)", doc.read_text()):
        if "://" in target or target.startswith("#"):
            continue
        path = target.split("#", 1)[0]
        if (doc.parent / path).resolve() == OUT / "scope.json":
            continue  # This report is written after the checks below succeed.
        if path and not (doc.parent / path).exists():
            broken_links.append({"file": str(doc.relative_to(ROOT)), "target": target})
assert not broken_links, broken_links

for name in ("backend-unit", "backend-embed", "backend-lint-final", "backend-error-race",
             "final-targeted", "intake-probes-after", "sqlite-billing-controls",
             "openspec-tier1", "openspec-tier2"):
    result = json.loads((OUT / (name + ".json")).read_text())
    assert result["exit_code"] == 0, name
assert json.loads((OUT / "s02/websocket-race-final.json").read_text())["exit_code"] == 0
assert subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=ROOT).decode().strip() == initial["head"]
subprocess.run(["git", "diff", "--check"], cwd=ROOT, check=True)

summary = {
    "head": initial["head"], "initial_files": len(initial["hashes"]),
    "unchanged_initial_files": len(initial["hashes"]) - len(changed),
    "changed_initial_files": changed,
    "unexpected_changes": unexpected, "unexpected_new_files": unexpected_new,
    "entry_original_content_hashes_restored": True,
    "previous_active_changes": len(old_changes),
    "previous_completed_tasks": sum(change["completed"] for change in old_changes),
    "selected_completed_tasks": selected_tasks, "broken_links": broken_links,
    "final_backend_hashes_match": True,
    "version": (ROOT / "backend/cmd/server/VERSION").read_text().strip(),
    "go": re.search(r"^go (.+)$", (ROOT / "backend/go.mod").read_text(), re.M).group(1),
    "protection": "All initial files outside explicit implementation/docs allowlist remain SHA256-identical, including frozen intake/source/older active changes, Fast diagnostics, frontend, migrations, Ent/Wire and resource defaults.",
}
(OUT / "scope.json").write_text(json.dumps(summary, ensure_ascii=False, indent=2) + "\n")
print(json.dumps(summary, ensure_ascii=False, indent=2))
