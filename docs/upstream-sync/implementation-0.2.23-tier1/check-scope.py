"""Verify tier1 content, frozen history and completed acceptance records."""
import hashlib
import json
import re
import subprocess
from pathlib import Path

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]

def git(*args):
    return subprocess.check_output(["rtk", "proxy", "git", *args], cwd=ROOT).decode().strip()

def changed_hashes(filename):
    original = json.loads((OUT / filename).read_text())
    changed = [name for name, digest in original.items()
               if not (ROOT / name).is_file()
               or hashlib.sha256((ROOT / name).read_bytes()).hexdigest() != digest]
    return len(original), changed

protected_count, protected_changes = changed_hashes("protected-files.json")
static_count, static_changes = changed_hashes("static-start-hashes.json")
product = json.loads((OUT / "product-hashes.json").read_text())
product_changes = [name for name, digest in product["files"].items()
                   if hashlib.sha256((ROOT / name).read_bytes()).hexdigest() != digest]
tracked = git("diff", "--name-only").splitlines()
new = git("ls-files", "--others", "--exclude-standard", "--", "backend", "frontend", "deploy", ".github").splitlines()
actual = {name for name in tracked + new if name.startswith(("backend/", "frontend/", "deploy/", ".github/"))}
expected = set(product["files"])
change = ROOT / "openspec/changes/port-upstream-0.2.23-tier1"
tasks = (change / "tasks.md").read_text()
spec = (change / "specs/upstream-0-2-23-tier1-compatibility/spec.md").read_text()
tier2 = (ROOT / "openspec/changes/port-upstream-0.2.23-tier2/tasks.md").read_text()
labels = ("backend-unit", "backend-embed-build", "backend-lint-final", "backend-targeted-race",
          "frontend-typecheck", "frontend-lint", "frontend-full-test", "frontend-build",
          "frontend-critical", "browser-final", "resource-1c1g",
          "real-http-ws-sqlite-csp-accepted", "frontend-html-post-build", "openspec-tier1-first", "openspec-tier1")
check_exits = {label: json.loads((OUT / (label + ".json")).read_text())["exit_code"] for label in labels}
result = {
    "head": git("rev-parse", "HEAD"),
    "protected_count": protected_count, "protected_changes": protected_changes,
    "static_count": static_count, "static_changes": static_changes,
    "product_hash_changes": product_changes,
    "source_count": product["source_count"], "test_count": product["test_count"],
    "unexpected_product_files": sorted(actual - expected), "missing_product_files": sorted(expected - actual),
    "tier1_completed_tasks": tasks.count("- [x]"), "tier1_unchecked_tasks": tasks.count("- [ ]"),
    "tier1_requirements": len(re.findall(r"^### Requirement: F\d{2}\b", spec, re.M)),
    "tier2_unchecked_tasks": tier2.count("- [ ]"),
    "VERSION": (ROOT / "backend/cmd/server/VERSION").read_text().strip(),
    "Go": re.search(r"^go (\S+)", (ROOT / "backend/go.mod").read_text(), re.M).group(1),
    "SQLite_max_migration": max(int(p.name.split("_", 1)[0]) for p in (ROOT / "backend/migrations").glob("[0-9][0-9][0-9]_*.sql")),
    "auth_snapshot": int(re.search(r"const apiKeyAuthSnapshotVersion = (\d+)", (ROOT / "backend/internal/service/api_key_auth_cache_impl.go").read_text()).group(1)),
    "acceptance_exit_codes": check_exits,
    "native_Anthropic_source_added": (ROOT / "backend/internal/service/openai_gateway_chat_completions_anthropic_native.go").exists(),
    "new_implementation_commit": git("rev-parse", "HEAD") != json.loads((OUT / "start.json").read_text())["head"],
}
result["passed"] = (
    not protected_changes and not static_changes and not product_changes and actual == expected
    and result["tier1_completed_tasks"] == 70 and result["tier1_unchecked_tasks"] == 0
    and result["tier1_requirements"] == 21 and result["tier2_unchecked_tasks"] == 76
    and result["VERSION"] == "1.1.25" and result["Go"] == "1.27.3"
    and result["SQLite_max_migration"] == 229 and result["auth_snapshot"] == 21
    and all(code == 0 for code in check_exits.values())
    and not result["native_Anthropic_source_added"] and not result["new_implementation_commit"]
)
(OUT / "scope-final.json").write_text(json.dumps(result, indent=2) + "\n")
print(json.dumps(result, indent=2))
if not result["passed"]:
    raise SystemExit(1)
