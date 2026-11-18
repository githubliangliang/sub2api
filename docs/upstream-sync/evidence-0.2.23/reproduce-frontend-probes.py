"""Apply only frozen upstream tests to a temporary copy of the fork frontend."""
import csv
import io
import json
import subprocess
import tarfile
import tempfile
from pathlib import Path

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]
SELECTED = {
    "PR8736": "frontend/src/components/account/__tests__/ModelWhitelistSelector.spec.ts",
    "PR8840": "frontend/src/views/admin/__tests__/ChannelsView.bedrockCompat.spec.ts",
}
CANDIDATES = {r["candidate"]: r for r in csv.DictReader((OUT / "candidates.tsv").open(), delimiter="\t")}
patches = []
for name, path in SELECTED.items():
    row = CANDIDATES[name]
    patches.append(subprocess.check_output([
        "rtk", "proxy", "git", "diff", row["parent1"], row["sha"], "--", path,
    ], cwd=ROOT))
patch = b"".join(patches)
(OUT / "probes/frontend-tests.patch").write_bytes(patch)
with tempfile.TemporaryDirectory(prefix="sub2api-023-frontend-") as tmp:
    tmp = Path(tmp)
    archive = subprocess.check_output(["rtk", "proxy", "git", "archive", "HEAD", "frontend"], cwd=ROOT)
    with tarfile.open(fileobj=io.BytesIO(archive)) as tar:
        tar.extractall(tmp, filter="data")
    frontend = tmp / "frontend"
    (frontend / "node_modules").symlink_to(ROOT / "frontend/node_modules", target_is_directory=True)
    whitelist_patch = patches[0].decode()
    added = "\n".join(line[1:] for line in whitelist_patch.splitlines()
        if line.startswith("+") and not line.startswith("+++")) + "\n"
    (OUT / "probes/frontend-whitelist-cases.spec.ts.txt").write_text(added)
    whitelist = frontend / SELECTED["PR8736"].removeprefix("frontend/")
    text = whitelist.read_text()
    anchor = "  it('rejects a custom whitelist model that is already mapped to a different target'"
    assert text.count(anchor) == 1
    whitelist.write_text(text.replace(anchor, added + anchor, 1))
    bedrock = CANDIDATES["PR8840"]
    bedrock_source = subprocess.check_output(["rtk", "proxy", "git", "show",
        bedrock["sha"] + ":" + SELECTED["PR8840"]], cwd=ROOT)
    (OUT / "probes/frontend-bedrock.spec.ts.txt").write_bytes(bedrock_source)
    (frontend / SELECTED["PR8840"].removeprefix("frontend/")).write_bytes(bedrock_source)
    (OUT / "frontend-probe-source.json").write_text(json.dumps({
        "fork_head": json.loads((OUT / "baseline.json").read_text())["fork_head"],
        "selected_test_diffs": SELECTED, "temporary_copy": str(frontend),
        "product_files": "fork HEAD, unchanged; only test diffs applied in the temporary copy",
    }, indent=2)+"\n")
    subprocess.run(["rtk", "proxy", "python3", str(OUT / "run-check.py"), "frontend-probe",
        str(frontend), "node", str(ROOT / "frontend/node_modules/vitest/vitest.mjs"), "run",
        *[path.removeprefix("frontend/") for path in SELECTED.values()]], check=False)
