#!/usr/bin/env python3
"""Check the completed change, captured gates and frozen implementation scope."""
import hashlib
import json
import re
import subprocess
from pathlib import Path
from urllib.parse import unquote, urlsplit

out = Path(__file__).resolve().parent
root = out.parents[2]
change = root / "openspec/changes/port-upstream-0.2.19-tier1"
baseline = json.loads((out / "baseline.json").read_text())
scope = json.loads((out / "scope-facts.json").read_text())


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


for name, expected in baseline["frozen_hashes"].items():
    assert digest(root / name) == expected, "frozen source changed: " + name
for name, expected in scope["product_sha256"].items():
    assert digest(root / name) == expected, "tested product changed: " + name

assert scope["production_files"] == 14
assert scope["test_files"] == 17
assert scope["changed_product_files"] == 31
assert (root / "backend/cmd/server/VERSION").read_text().strip() == "1.1.20"
assert re.search(r"^go 1\.27\.3$", (root / "backend/go.mod").read_text(), re.M)
maximum = max(int(re.match(r"\d+", p.name).group()) for p in (root / "backend/migrations").glob("[0-9]*.sql"))
assert maximum == 227
assert scope["implementation_commit"], "implementation commit missing"

head = subprocess.check_output(["rtk", "proxy", "git", "rev-parse", "HEAD"], cwd=root, text=True).strip()
implementation_commit = scope["implementation_commit"]
subprocess.run(["rtk", "proxy", "git", "merge-base", "--is-ancestor", implementation_commit, head], cwd=root, check=True)
for entry in scope["changes"]:
    committed = subprocess.check_output(["rtk", "proxy", "git", "show", implementation_commit + ":" + entry["path"]], cwd=root)
    assert hashlib.sha256(committed).hexdigest() == entry["after_sha256"], entry["path"]

tasks = (change / "tasks.md").read_text()
assert len(re.findall(r"^- \[x\]", tasks, re.M)) == 26
assert not re.search(r"^- \[ \]", tasks, re.M)
verification = (change / "verification.md").read_text()
assert len(re.findall(r"^- \[x\]", verification, re.M)) == 3
assert not re.search(r"\|[ \t]*\|", verification), "empty verification cell"
assert implementation_commit in verification

gates = ["backend-unit-direct", "backend-build", "backend-lint-final", "backend-race", "sqlite-protection", "openspec-strict-final"]
for gate in gates:
    record = json.loads((out / (gate + ".json")).read_text())
    assert record["exit_code"] == 0, gate
    assert (out / record["stdout"]).is_file()
    assert (out / record["stderr"]).is_file()
unit_output = (out / "backend-unit-direct.txt").read_text()
assert len(re.findall(r"^ok\s", unit_output, re.M)) == 54
race_output = (out / "backend-race.txt").read_text()
assert len(re.findall(r"^ok\s", race_output, re.M)) == 8
assert "[no tests to run]" not in race_output
assert "0 issues." in (out / "backend-lint-final.txt").read_text()

checked_links = 0
for file in list(change.rglob("*.md")) + list(out.glob("*.md")):
    for target in re.findall(r"\[[^\]]+\]\(([^)]+)\)", file.read_text()):
        target = target.strip().strip("<>")
        parsed = urlsplit(target)
        if parsed.scheme or not parsed.path:
            continue
        path = unquote(parsed.path)
        assert (file.parent / path).resolve().exists(), str(file) + ": " + target
        checked_links += 1

report = {
    "implementation_commit": implementation_commit,
    "tasks": "26/26",
    "requirements": sum(p.read_text().count("### Requirement:") for p in (change / "specs").rglob("spec.md")),
    "scenarios": sum(p.read_text().count("#### Scenario:") for p in (change / "specs").rglob("spec.md")),
    "frozen_files_unchanged": len(baseline["frozen_hashes"]),
    "protected_product_files_unchanged": scope["protected_product_files"],
    "changed_files": 31,
    "production_files": 14,
    "test_files": 17,
    "backend_unit_packages": 54,
    "race_packages": 8,
    "checked_local_links": checked_links,
    "gates": gates,
}
print(json.dumps(report, ensure_ascii=False, indent=2))
