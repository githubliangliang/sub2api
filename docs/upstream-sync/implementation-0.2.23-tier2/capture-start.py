"""Freeze the actual shared worktree, including the uncommitted first tier."""
import datetime
import hashlib
import json
import shutil
import subprocess
import tempfile
from pathlib import Path

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]
if (OUT / "start.json").exists():
    raise SystemExit("Entry snapshot already exists; do not overwrite frozen starting evidence")
SOURCE = Path(tempfile.mkdtemp(prefix="sub2api-023-tier2-start-")) / "source"

def git(*args):
    return subprocess.check_output(["rtk", "proxy", "git", *args], cwd=ROOT).decode()

def hashes(paths):
    return {p: hashlib.sha256((ROOT / p).read_bytes()).hexdigest() for p in paths}

files = [p for p in git("ls-files", "--cached", "--others", "--exclude-standard").splitlines()
         if (ROOT / p).is_file() and not p.startswith(str(OUT.relative_to(ROOT)) + "/")]
product = [p for p in files if p.startswith(("backend/", "frontend/"))]
protected = [p for p in files if
             (p.startswith("docs/upstream-sync/") and p != "docs/upstream-sync/README.md")
             or p.startswith("docs/releases/")
             or (p.startswith("openspec/") and p not in {
                 "openspec/changes/port-upstream-0.2.23-tier2/tasks.md",
                 "openspec/changes/port-upstream-0.2.23-tier2/verification.md",
                 "openspec/changes/port-upstream-0.2.23-tier2/README.md"})]
static = [p for p in files if p.startswith(("backend/ent/", "backend/migrations/", "deploy/"))
          or p in {"backend/go.mod", "backend/go.sum", "backend/cmd/server/VERSION",
                   "backend/cmd/server/wire_gen.go", "frontend/package.json",
                   "frontend/pnpm-lock.yaml", "frontend/pnpm-workspace.yaml", "Dockerfile"}
          or p.startswith(".github/workflows/")]
for name, paths in [("product-start-hashes", product), ("protected-files", protected),
                    ("static-start-hashes", static)]:
    (OUT / (name + ".json")).write_text(json.dumps(hashes(paths), indent=2) + "\n")
for p in product + ["AGENTS.md", "docs/upstream-sync/README.md"]:
    target = SOURCE / p
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(ROOT / p, target)
(OUT / "start-worktree.diff").write_text(git("diff", "--binary"))
(OUT / "start.json").write_text(json.dumps(dict(
    started_at_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),
    head=git("rev-parse", "HEAD").strip(), tree=git("rev-parse", "HEAD^{tree}").strip(),
    version=(ROOT / "backend/cmd/server/VERSION").read_text().strip(),
    start_source=str(SOURCE), status=git("status", "--short"), protected_count=len(protected),
    product_count=len(product), static_count=len(static),
), indent=2) + "\n")
print(json.dumps(dict(product=len(product), protected=len(protected), static=len(static))))
