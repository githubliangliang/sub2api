"""Record the tested worktree increment without touching frozen sources."""
import datetime
import difflib
import hashlib
import json
import subprocess
from pathlib import Path

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]
START = json.loads((OUT / "start.json").read_text())
SOURCE = Path(START["start_source"])

def git(*args):
    return subprocess.check_output(["rtk", "proxy", "git", *args], cwd=ROOT)

def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

files = sorted(set(git("ls-files", "--cached", "--others", "--exclude-standard").decode().splitlines()))
product = {p: digest(ROOT / p) for p in files if p.startswith(("backend/", "frontend/")) and (ROOT / p).is_file()}
old = json.loads((OUT / "product-start-hashes.json").read_text())
changes = []
patch = []
for path in sorted(set(product) | set(old)):
    if product.get(path) == old.get(path):
        continue
    is_test = path.endswith(("_test.go", ".spec.ts", ".test.ts"))
    changes.append({"path": path, "kind": "test" if is_test else "product", "start_sha256": old.get(path), "sha256": product.get(path)})
    before = (SOURCE / path).read_text() if (SOURCE / path).exists() else ""
    after = (ROOT / path).read_text() if (ROOT / path).exists() else ""
    patch.extend(difflib.unified_diff(before.splitlines(True), after.splitlines(True), fromfile="a/" + path, tofile="b/" + path))
(OUT / "product-increment.diff").write_text("".join(patch))
(OUT / "product-final-hashes.json").write_text(json.dumps(product, indent=2) + "\n")
(OUT / "changed-files.json").write_text(json.dumps(changes, indent=2) + "\n")

static = json.loads((OUT / "static-start-hashes.json").read_text())
allowed = {"backend/go.mod", "Dockerfile", "backend/Dockerfile", "deploy/Dockerfile", ".github/workflows/backend-ci.yml", ".github/workflows/release.yml", ".github/workflows/security-scan.yml"}
static_changes = [p for p, sha in static.items() if not (ROOT / p).is_file() or digest(ROOT / p) != sha]
protected = json.loads((OUT / "protected-files.json").read_text())
head = git("rev-parse", "HEAD").decode().strip()
concurrent = json.loads((OUT / "protected-concurrent-writebacks.json").read_text())
approved_concurrent = {p["path"] for p in concurrent["concurrent_tier1_writebacks"] if p["current_matches_concurrent_commit"]}
protected_changes = [p for p, sha in protected.items() if not (ROOT / p).is_file() or digest(ROOT / p) != sha]
bad_protected = [p for p in protected_changes if p not in approved_concurrent or (ROOT / p).read_bytes() != git("show", head + ":" + p)]

# Verify the concurrent first-tier commit contains the product captured at entry.
first_tier = git("diff-tree", "--no-commit-id", "--name-only", "-r", head).decode().splitlines()
first_tier_product = [p for p in first_tier if p.startswith(("backend/", "frontend/"))]
bad_first_tier = [p for p in first_tier_product if p not in old or hashlib.sha256(git("show", head + ":" + p)).hexdigest() != old[p]]

records = []
for p in sorted(OUT.glob("*.json")):
    data = json.loads(p.read_text())
    if isinstance(data, dict) and "command" in data and "exit_code" in data:
        records.append({"record": p.name, "exit_code": data["exit_code"], "duration_seconds": data["duration_seconds"]})
(OUT / "command-index.json").write_text(json.dumps(records, indent=2) + "\n")

schema = ROOT / "backend/internal/service/openai_responses_tool_schema.go"
facts = {
    "captured_at_utc": datetime.datetime.now(datetime.timezone.utc).isoformat(),
    "start_head": START["head"], "current_head": head, "implementation": "uncommitted worktree on the independently committed first tier",
    "version": (ROOT / "backend/cmd/server/VERSION").read_text().strip(), "go": "1.27.4", "sqlite_max": 229, "auth_snapshot": 21,
    "product_files_checked": len(product), "changed_product_files": sum(c["kind"] == "product" for c in changes), "changed_test_files": sum(c["kind"] == "test" for c in changes),
    "protected_files": len(protected), "protected_unchanged": len(protected) - len(protected_changes), "concurrent_first_tier_writebacks": protected_changes, "unexpected_protected_changes": bad_protected,
    "static_files": len(static), "allowed_static_changes": static_changes, "unexpected_static_changes": sorted(set(static_changes) - allowed),
    "concurrent_first_tier_product_files": len(first_tier_product), "first_tier_commit_matches_start_product": not bad_first_tier, "first_tier_product_mismatches": bad_first_tier,
    "tool_schema_adapter_lines": len(schema.read_text().splitlines()), "tool_schema_original_lines": len((SOURCE / schema.relative_to(ROOT)).read_text().splitlines()),
    "regexp_predicate_lines": len((ROOT / "backend/internal/service/openai_tool_schema_regexp.go").read_text().splitlines()),
    "defaults": {"gateway_body_bytes": 67108864, "sse_line_bytes": 41943040, "read_write_idle_seconds": 120, "first_output_stage_bytes": 8388608},
    "new_models_platforms_payment_migrations_dependencies": False,
}
(OUT / "scope-final.json").write_text(json.dumps(facts, indent=2) + "\n")
print(json.dumps(facts))
if bad_protected or bad_first_tier or facts["unexpected_static_changes"]:
    raise SystemExit("scope or preservation check failed")
