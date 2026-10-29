#!/usr/bin/env python3
"""Recheck the existing intake evidence without replacing its frozen records."""
import collections
import csv
import datetime
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]


def git(*args, env=None):
    return subprocess.check_output(["git", *args], cwd=ROOT, env=env)


def digest(data):
    return hashlib.sha256(data).hexdigest()


def rows(name):
    with (OUT / name).open() as source:
        return list(csv.DictReader(source, delimiter="\t"))


baseline = json.loads((OUT / "baseline.json").read_text())
head = git("rev-parse", "HEAD").decode().strip()
assert head == baseline["fork_head"]
refs = {ref: git("rev-parse", ref + "^{commit}").decode().strip()
        for ref in ("v0.2.18", "v0.2.19", "upstream/main")}
assert all(refs[ref] == baseline["refs"][ref] for ref in ("v0.2.18", "v0.2.19"))

product_hashes = json.loads((OUT / "business-file-sha256.json").read_text())
for item in product_hashes:
    assert digest((ROOT / item["path"]).read_bytes()) == item["sha256"], item["path"]

candidates = rows("candidates.tsv")
commits = set(git("rev-list", "--no-merges", "v0.2.18..v0.2.19").decode().splitlines())
covered = set()
for candidate in candidates:
    patch = git("diff", "--binary", "--no-ext-diff", candidate["parent1"], candidate["sha"])
    assert digest(patch) == candidate["patch_sha256"], candidate["candidate"]
    assert patch == (OUT / "patches" / (candidate["candidate"] + ".patch")).read_bytes()
    members = set(git("rev-list", "--no-merges", candidate["parent1"] + ".." + candidate["sha"]).decode().splitlines())
    assert not (members & covered)
    covered |= members
assert covered == commits

files = rows("files.tsv")
by_candidate = {item["candidate"]: item for item in candidates}
checks = []
with tempfile.TemporaryDirectory(prefix="sub2api-019-review-index-") as temporary:
    env = dict(os.environ, GIT_INDEX_FILE=str(Path(temporary) / "index"))
    git("read-tree", head, env=env)
    patches = {}
    for item in files:
        candidate = by_candidate[item["candidate"]]
        key = (item["candidate"], item["path"])
        patches[key] = git("diff", "--binary", "--no-ext-diff", candidate["parent1"],
                           candidate["sha"], "--", item["path"])
        checks.append({"candidate": item["candidate"], "path": item["path"]})
    # Finish the reverse pass for the entire candidate set before any forward check.
    for reverse in (True, False):
        field = "reverse_exit" if reverse else "forward_exit"
        for item, checked in zip(files, checks):
            command = ["git", "apply", "--cached", "--check", "--whitespace=nowarn"]
            if reverse:
                command.append("--reverse")
            result = subprocess.run(command, cwd=ROOT, env=env,
                                    input=patches[(item["candidate"], item["path"])],
                                    stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            checked[field] = result.returncode
            assert result.returncode == int(item[field]), (field, item["candidate"], item["path"])
    for item, checked in zip(files, checks):
        checked["state"] = ("ALREADY" if checked["reverse_exit"] == 0 else
                            "CLEAN" if checked["forward_exit"] == 0 else
                            "CONFLICT" if item["base_exists"] == "True" else "NOBASE")
        assert checked["state"] == item["state"]

with (OUT / "review-file-checks.tsv").open("w") as output:
    writer = csv.DictWriter(output, fieldnames=list(checks[0]), delimiter="\t")
    writer.writeheader()
    writer.writerows(checks)

record = {
    "checked_at_utc": datetime.datetime.now(datetime.timezone.utc).isoformat(),
    "fork_head": head,
    "refs": refs,
    "version": (ROOT / "backend/cmd/server/VERSION").read_text().strip(),
    "initial_status": git("status", "--short").decode(),
    "index_patch_sha256": digest(git("diff", "--cached", "--binary")),
    "worktree_patch_sha256": digest(git("diff", "--binary")),
    "product_files_checked": len(product_hashes),
    "product_mismatches": [],
    "nonmerge_commits": len(commits),
    "candidates": len(candidates),
    "prs": sum(item["kind"] == "PR" for item in candidates),
    "file_checks": len(checks),
    "states": {state: sum(item["state"] == state for item in checks)
               for state in ("ALREADY", "CLEAN", "CONFLICT", "NOBASE")},
    "all_original_patches_and_states_match": True,
    "frozen_evidence_sha256": {name: digest((OUT / name).read_bytes()) for name in
        ("baseline.json", "commits.tsv", "candidates.tsv", "files.tsv", "candidate-stats.tsv",
         "source-test-inventory.tsv", "added-call-inventory.tsv", "business-file-sha256.json")},
}
(OUT / "review-baseline.json").write_text(json.dumps(record, ensure_ascii=False, indent=2) + "\n")
print(json.dumps({key: record[key] for key in
    ("fork_head", "product_files_checked", "nonmerge_commits", "candidates", "prs", "file_checks", "states")}, ensure_ascii=False))
