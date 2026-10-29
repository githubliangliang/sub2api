#!/usr/bin/env python3
"""Verify the completed change, implementation commit and frozen boundaries."""
from pathlib import Path
import hashlib
import json
import re
import subprocess

root = Path.cwd()
out = Path(__file__).resolve().parent
change = root / "openspec/changes/port-upstream-0.2.19-tier2"
baseline = json.loads((out / "baseline.json").read_text())
scope = json.loads((out / "scope-facts.json").read_text())
failures = []
digest = lambda path: hashlib.sha256(path.read_bytes()).hexdigest()
for name, expected in baseline["frozen_sha256"].items():
    if name == scope["authorized_spec_clarification"]["path"]:
        continue
    if digest(root / name) != expected:
        failures.append("frozen source changed: " + name)
for item in scope["changes"]:
    if digest(root / item["path"]) != item["after_sha256"]:
        failures.append("implementation changed since verification: " + item["path"])
commit = scope["implementation_commit"]
actual = subprocess.check_output(["git", "diff-tree", "--no-commit-id", "--name-only", "-r", commit], text=True).splitlines()
if set(actual) != {item["path"] for item in scope["changes"]}:
    failures.append("implementation commit scope differs")
for filename in ("tasks.md", "verification.md"):
    text = (change / filename).read_text()
    if "- [ ]" in text or "PENDING" in text:
        failures.append("unfinished completion record: " + filename)
if (change / "tasks.md").read_text().count("- [x]") != 23:
    failures.append("task count differs from 23")
if (root / "backend/cmd/server/VERSION").read_text().strip() != "1.1.20":
    failures.append("VERSION changed")
if not re.search(r"^go 1\.27\.3$", (root / "backend/go.mod").read_text(), re.M):
    failures.append("Go version changed")
for filename in ("final-unit", "backend-build", "final-lint", "backend-service-race-complete", "backend-network-race", "sqlite-protection", "frontend-tests-complete", "frontend-build-complete", "final-frontend-typecheck", "final-frontend-lint", "frontend-critical-final", "generation-compatibility", "openspec-strict-final"):
    record = json.loads((out / (filename + ".json")).read_text())
    if record["exit_code"] != 0:
        failures.append("final check failed: " + filename)
race = (out / "backend-race-final.txt").read_text()
for package in ("handler", "repository", "securityaudit"):
    if not re.search(r"^ok\s+github.com/Wei-Shaw/sub2api/internal/" + package + r"\s", race, re.M):
        failures.append("successful race output missing: " + package)
cleanup = json.loads((out / "browser/cleanup.json").read_text())
if not all(value for group in cleanup.values() for value in group.values()):
    failures.append("owned browser server remained running")
for path in list(change.rglob("*.md")) + list(out.rglob("*.md")):
    for target in re.findall(r"\[[^\]]*\]\(([^)]+)\)", path.read_text()):
        target = target.strip("<>").split("#", 1)[0]
        if not target or "://" in target or target.startswith("mailto:"):
            continue
        if not (path.parent / target).resolve().exists():
            failures.append("broken link: " + str(path.relative_to(root)) + " -> " + target)
if failures:
    print("\n".join(failures))
    raise SystemExit(1)
print(json.dumps({"result": "PASS", "tasks": "23/23", "implementation_commit": commit, "product_files": len(actual), "frozen_sources_unchanged": len(baseline["frozen_sha256"]) - 1, "authorized_s01_scenario": True, "version": "1.1.20", "go": "1.27.3", "migration": 227, "final_checks_and_links": "PASS"}, ensure_ascii=False, indent=2))
