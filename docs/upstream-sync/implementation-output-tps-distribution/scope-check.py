"""Verify frozen files, changed scope, links and completed acceptance records."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
from urllib.parse import unquote

out = Path(__file__).resolve().parent
root = out.parents[2]
baseline = json.loads((out / "baseline.json").read_text())
git = lambda *args: subprocess.check_output(["git", *args], cwd=root, text=True).strip()
paths = git("ls-tree", "-r", "--name-only", baseline["head"]).splitlines()
protected = [p for p in paths if p.startswith(("openspec/changes/archive/", "openspec/specs/", "docs/releases/", "backend/migrations/", "backend/ent/"))
             or (p.startswith("docs/upstream-sync/") and p != "docs/upstream-sync/README.md")
             or p in ["backend/go.mod", "backend/go.sum", "frontend/package.json", "frontend/pnpm-lock.yaml", "backend/cmd/server/VERSION", "backend/cmd/server/wire_gen.go"]]
digest = hashlib.sha256()
for name in sorted(protected):
    digest.update(name.encode() + b"\0" + hashlib.sha256((root / name).read_bytes()).digest())
assert len(protected) == baseline["protected_file_count"]
assert digest.hexdigest() == baseline["protected_tree_sha256"], "Frozen history/dependencies/schema changed"

change = root / "openspec/changes/port-upstream-output-tps-distribution"
new_files = git("ls-files", "--others", "--exclude-standard").splitlines()
changed = git("diff", "--name-only").splitlines()
expected_product = {
    "backend/internal/service/ops_dashboard_models.go", "backend/internal/repository/ops_repo_dashboard.go",
    "backend/internal/repository/ops_repo_output_tps.go", "backend/internal/repository/ops_repo_output_tps_sqlite_test.go",
    "backend/internal/repository/ops_repo_output_tps_bench_test.go", "backend/internal/repository/ops_repo_dashboard_timeout_test.go",
    "backend/internal/handler/admin/ops_output_tps_sqlite_test.go", "frontend/src/api/admin/ops.ts",
    "frontend/src/views/admin/ops/components/OpsDashboardHeader.vue",
    "frontend/src/views/admin/ops/components/__tests__/OpsDashboardHeader.outputTps.spec.ts",
    "frontend/src/i18n/locales/zh/admin/ops.ts", "frontend/src/i18n/locales/en/admin/ops.ts",
}
all_changes = set(changed + new_files)
assert {p for p in all_changes if p.startswith(("backend/", "frontend/"))} == expected_product
allowed_docs = {"AGENTS.md", "docs/upstream-sync/README.md", "docs/upstream-sync/PORTING-output-tps-distribution.md"}
assert all(p in expected_product | allowed_docs or p.startswith(("openspec/changes/port-upstream-output-tps-distribution/", "docs/upstream-sync/implementation-output-tps-distribution/")) for p in all_changes)

result_path = out / "scope-check.json"
checked_links = 0
documents = list(change.rglob("*.md")) + [out / "README.md", root / "docs/upstream-sync/PORTING-output-tps-distribution.md"]
for path in documents:
    text = path.read_text()
    for target in re.findall(r"\]\(([^)]+)\)", text):
        target = unquote(target.split("#", 1)[0].strip("<>"))
        if not target or "://" in target:
            continue
        resolved = (path.parent / target).resolve()
        assert resolved == result_path or resolved.exists(), (str(path), target)
        checked_links += 1
tasks = (change / "tasks.md").read_text()
assert len(re.findall(r"^- \[x\]", tasks, re.M)) == 10 and "- [ ]" not in tasks
checks = ["backend-targeted-accepted", "backend-http-final", "backend-unit-final", "backend-race", "backend-embed", "backend-lint",
          "backend-query-cost", "frontend-targeted", "frontend-unit", "frontend-typecheck", "frontend-lint", "frontend-build", "frontend-critical", "openspec-strict", "query-plan"]
for name in checks:
    assert json.loads((out / (name + ".json")).read_text())["exit_code"] == 0, name
plan = json.loads((out / "query-plan-details.json").read_text())
assert plan["product_query_sha256"] == hashlib.sha256((root / "backend/internal/repository/ops_repo_output_tps.go").read_bytes()).hexdigest()
assert git("diff", "--check") == ""
record = dict(baseline_head=baseline["head"], current_head=git("rev-parse", "HEAD"),
              committed=False, version=baseline["version"], go=baseline["go"], max_sqlite_migration=baseline["max_sqlite_migration"],
              protected_file_count=len(protected), protected_tree_sha256=digest.hexdigest(),
              relative_links_checked=checked_links, completed_tasks=10, checks_passed=checks,
              source_baseline_sha256=hashlib.sha256((change / "source-baseline.md").read_bytes()).hexdigest(),
              product_sha256={p: hashlib.sha256((root / p).read_bytes()).hexdigest() for p in sorted(expected_product)})
result_path.write_text(json.dumps(record, ensure_ascii=False, indent=2) + "\n")
print(json.dumps(record, ensure_ascii=False, indent=2))
