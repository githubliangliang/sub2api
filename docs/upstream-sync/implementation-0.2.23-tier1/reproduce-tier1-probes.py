"""Run only tier1 intake assertions without changing frozen evidence or code."""
import json
import subprocess
import sys
from pathlib import Path

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]
SOURCE = ROOT / "docs/upstream-sync/evidence-0.2.23/probes"
PACKAGES = {
    "apicompat": "internal/pkg/apicompat",
    "service": "internal/service",
    "handler": "internal/handler",
    "antigravity": "internal/pkg/antigravity",
    "repository": "internal/repository",
    "web": "internal/web",
}
overlay = OUT / "tier1-probe-overlay.json"
overlay.write_text(json.dumps({"Replace": {
    str(ROOT / "backend" / package / "intake023_release_test.go"):
    str(SOURCE / (name + ".go.txt"))
    for name, package in PACKAGES.items()
}}, indent=2) + "\n")
pattern = "^TestIntake023(Sealed|MidSystem|Strict|PerTurn|CountTokensHeader|Login|Spool|Grok.*Media|SolUltra|AstraUltra|Weekly|Codex|SchedulerCap|SQLiteFiveHour|Schema|HTML)"
result = subprocess.run([
    "rtk", "proxy", "python3", str(OUT / "run-check.py"),
    sys.argv[1], "backend", "go", "test", "-tags=unit,embed",
    "-overlay=" + str(overlay), "-run", pattern, "-count=1", "-v",
    *["./" + package for package in PACKAGES.values()],
])
sys.exit(result.returncode)
