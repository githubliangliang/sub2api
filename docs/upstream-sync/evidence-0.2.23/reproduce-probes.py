"""Run desired behaviours against the frozen product with test-only Go overlays."""
import json
import subprocess
import sys
import tempfile
from pathlib import Path

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]
PACKAGES = {
    "apicompat": "internal/pkg/apicompat",
    "service": "internal/service",
    "handler": "internal/handler",
    "antigravity": "internal/pkg/antigravity",
    "repository": "internal/repository",
    "web": "internal/web",
}
with tempfile.TemporaryDirectory(prefix="sub2api-023-probes-") as tmp:
    overlay = Path(tmp) / "overlay.json"
    overlay.write_text(json.dumps({"Replace": {
        str(ROOT / "backend" / package / "intake023_release_test.go"):
        str(OUT / "probes" / (name + ".go.txt"))
        for name, package in PACKAGES.items()
    }}))
    for name, package in PACKAGES.items():
        if len(sys.argv) > 1 and name not in sys.argv[1:]:
            continue
        subprocess.run([
            "rtk", "proxy", "python3", str(OUT / "run-check.py"), name + "-probe",
            "backend", "go", "test", "-tags=" + ("unit,embed" if name == "web" else "unit"), "-overlay=" + str(overlay),
            "-run", "^TestIntake023", "-count=1", "-v", "./" + package,
        ], check=False)
