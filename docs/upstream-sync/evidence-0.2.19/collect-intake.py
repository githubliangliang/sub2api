#!/usr/bin/env python3
"""Read-only release inventory and reverse-before-forward checks on a frozen index."""
import collections
import csv
import datetime
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]
OLD, NEW = "v0.2.18", "v0.2.19"


def git(*args, env=None):
    return subprocess.check_output(["git", *args], cwd=ROOT, env=env)


def write_tsv(name, fields, rows):
    with (OUT / name).open("w", newline="") as stream:
        writer = csv.DictWriter(stream, fieldnames=fields, delimiter="\t")
        writer.writeheader()
        writer.writerows(rows)


def digest(data):
    return hashlib.sha256(data).hexdigest()


fork = git("rev-parse", "HEAD").decode().strip()
initial_dir = Path("/tmp/sub2api-release-intake-20261028")
initial = json.loads((initial_dir / "initial-facts.json").read_text())
initial["session_start_head"] = initial["fork_head"]
if fork != initial["fork_head"]:
    changed = git("diff", "--name-only", initial["fork_head"], fork).decode().splitlines()
    assert all(p == "AGENTS.md" or p.startswith(("docs/", "openspec/")) for p in changed)
    initial["baseline_transition"] = dict(from_head=initial["fork_head"], to_head=fork,
        changed_files=changed, reason="Existing personal-maintenance documentation was committed during intake; no business source changed.")
initial["fork_head"] = fork
if (OUT / "baseline.json").exists():
    assert json.loads((OUT / "baseline.json").read_text())["fork_head"] == fork
refs = {name: git("rev-parse", name + "^{commit}").decode().strip()
        for name in (OLD, NEW, "upstream/main")}
release_fields = ("id", "tag_name", "name", "created_at", "published_at", "draft",
                  "prerelease", "target_commitish", "html_url", "body")
releases = json.loads((initial_dir / "releases.json").read_text())
(OUT / "releases.json").write_text(json.dumps([
    {k: r.get(k) for k in release_fields} for r in releases
    if r["tag_name"] in (OLD, NEW)], ensure_ascii=False, indent=2) + "\n")
(OUT / "initial-status.txt").write_text(initial["status"])

commits = []
for row in git("log", "--reverse", "--no-merges", "--format=%H%x09%P%x09%aI%x09%s",
               OLD + ".." + NEW).decode().splitlines():
    sha, parents, date, subject = row.split("\t", 3)
    commits.append(dict(sha=sha, parents=parents, author_date=date, subject=subject))
write_tsv("commits.tsv", ["sha", "parents", "author_date", "subject"], commits)
nonmerge_set = {c["sha"] for c in commits}
candidates, patches, covered = [], {}, set()
(OUT / "patches").mkdir(exist_ok=True)
for row in git("log", "--first-parent", "--reverse", "--format=%H%x09%P%x09%s",
               OLD + ".." + NEW).decode().splitlines():
    sha, parents, subject = row.split("\t", 2)
    parent = parents.split()[0]
    match = re.search(r"Merge pull request #(\d+)", subject)
    name = "PR" + match.group(1) if match else sha[:9]
    members = set(git("rev-list", "--no-merges", parent + ".." + sha).decode().splitlines())
    assert not (members - nonmerge_set), (name, members - nonmerge_set)
    assert not (members & covered), (name, members & covered)
    covered |= members
    patch = git("diff", "--binary", "--no-ext-diff", parent, sha)
    patches[name] = patch
    (OUT / "patches" / (name + ".patch")).write_bytes(patch)
    candidates.append(dict(candidate=name, kind="PR" if match else "direct", sha=sha,
        parent1=parent, nonmerge_count=len(members), subject=subject,
        shortstat=git("diff", "--shortstat", parent, sha).decode().strip(),
        patch_sha256=digest(patch)))
assert covered == nonmerge_set
write_tsv("candidates.tsv", list(candidates[0]), candidates)

# Both passes use the same private index. No patch is applied to any tree or index.
files = []
with tempfile.TemporaryDirectory(prefix="sub2api-intake-index-") as temporary:
    env = dict(os.environ, GIT_INDEX_FILE=str(Path(temporary) / "index"))
    git("read-tree", fork, env=env)
    for candidate in candidates:
        names = git("diff", "--name-only", "-z", candidate["parent1"], candidate["sha"]).decode().split("\0")
        for name in filter(None, names):
            patch = git("diff", "--binary", "--no-ext-diff", candidate["parent1"], candidate["sha"], "--", name)
            exists = bool(git("ls-tree", "--name-only", fork, "--", name))
            status = git("diff", "--name-status", candidate["parent1"], candidate["sha"], "--", name).decode().split("\t")[0]
            files.append(dict(candidate=candidate["candidate"], path=name, change=status,
                base_exists=exists, patch=patch))
    for reverse in (True, False):
        prefix = "reverse" if reverse else "forward"
        for item in files:
            args = ["git", "apply", "--cached", "--check", "--whitespace=nowarn"]
            if reverse:
                args.append("--reverse")
            result = subprocess.run(args, input=item["patch"], env=env, cwd=ROOT,
                                    stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            item[prefix + "_exit"] = result.returncode
            item[prefix + "_error"] = result.stderr.decode(errors="replace").strip().replace("\n", " | ")
for item in files:
    item["state"] = ("ALREADY" if item["reverse_exit"] == 0 else
                     "CLEAN" if item["forward_exit"] == 0 else
                     "CONFLICT" if item["base_exists"] else "NOBASE")
    del item["patch"]
write_tsv("files.tsv", ["candidate", "path", "change", "base_exists", "state",
    "reverse_exit", "forward_exit", "reverse_error", "forward_error"], files)
stats = []
for candidate in candidates:
    counts = collections.Counter(f["state"] for f in files if f["candidate"] == candidate["candidate"])
    stats.append(dict(candidate=candidate["candidate"], **{k: counts[k] for k in
        ("ALREADY", "CLEAN", "CONFLICT", "NOBASE")}))
write_tsv("candidate-stats.tsv", list(stats[0]), stats)

decl_pattern = re.compile(r"^\s*func\s+(?:\([^)]*\)\s*)?(\w+)\s*(?:\[[^]]*\])?\(", re.M)
local_decls = collections.defaultdict(list)
tracked_go = git("ls-files", "-z", "backend").decode().split("\0")
for name in tracked_go:
    path = ROOT / name
    if name.endswith(".go") and path.is_file():
        content = path.read_text()
        for match in decl_pattern.finditer(content):
            line = content.count("\n", 0, match.start()) + 1
            local_decls[match.group(1)].append(f"{name}:{line}")
calls, tests, hunks = [], [], []
for candidate in candidates:
    patch = patches[candidate["candidate"]].decode(errors="replace")
    added_text = "\n".join(line[1:] for line in patch.splitlines()
                            if line.startswith("+") and not line.startswith("+++"))
    added_decls = set(decl_pattern.findall(added_text))
    path, context = "", ""
    seen = set()
    for line in patch.splitlines():
        if line.startswith("+++ b/"):
            path = line[6:]
        elif line.startswith("@@"):
            context = line
            hunks.append(dict(candidate=candidate["candidate"], path=path, hunk=context))
        elif line.startswith("+") and not line.startswith("+++") and path.endswith(".go"):
            for symbol in re.findall(r"\b([A-Za-z_]\w*)\s*\(", line[1:]):
                key = (path, symbol)
                if key in seen or symbol in {"if", "for", "switch", "func"}:
                    continue
                seen.add(key)
                calls.append(dict(candidate=candidate["candidate"], path=path, symbol=symbol,
                    local_declarations=" | ".join(local_decls[symbol][:4]),
                    added_by_candidate=symbol in added_decls, added_line=line[1:].strip()))
    for item in files:
        name = item["path"]
        if item["candidate"] != candidate["candidate"] or not (name.endswith("_test.go") or
                name.endswith(".spec.ts")):
            continue
        source = git("show", candidate["sha"] + ":" + name).decode(errors="replace")
        local = (ROOT / name).read_text() if (ROOT / name).is_file() else ""
        names = re.findall(r"^func (Test\w+)\(", source, re.M) if name.endswith(".go") else []
        tests.append(dict(candidate=candidate["candidate"], path=name,
            source_first_line=source.splitlines()[0] if source else "",
            local_exists=bool(local), local_first_line=local.splitlines()[0] if local else "",
            source_tests=" | ".join(names), local_tests=" | ".join(
                re.findall(r"^func (Test\w+)\(", local, re.M))))
write_tsv("added-call-inventory.tsv", ["candidate", "path", "symbol", "local_declarations",
    "added_by_candidate", "added_line"], calls)
write_tsv("hunks.tsv", ["candidate", "path", "hunk"], hunks)
write_tsv("source-test-inventory.tsv", ["candidate", "path", "source_first_line", "local_exists",
    "local_first_line", "source_tests", "local_tests"], tests)
initial["refs"] = refs
initial["frozen_at_utc"] = datetime.datetime.now(datetime.timezone.utc).isoformat()
initial["release_api_url"] = "https://api.github.com/repos/Wei-Shaw/sub2api/releases?per_page=15"
initial["counts"] = dict(nonmerge=len(commits), candidates=len(candidates),
    prs=sum(c["kind"] == "PR" for c in candidates), files=len(files),
    states=dict(collections.Counter(f["state"] for f in files)),
    net_shortstat=git("diff", "--shortstat", OLD, NEW).decode().strip(),
    calls=len(calls), source_test_files=len(tests))
(OUT / "baseline.json").write_text(json.dumps(initial, ensure_ascii=False, indent=2) + "\n")
print(json.dumps({"refs": refs, "counts": initial["counts"], "candidate_states": stats},
                 ensure_ascii=False, indent=2))
