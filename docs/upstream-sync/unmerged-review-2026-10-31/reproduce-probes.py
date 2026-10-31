"""Reproduce the frozen intake probes without modifying backend files."""

import json
from pathlib import Path
import subprocess
import sys
import tempfile


def main():
    evidence = Path(__file__).resolve().parent
    root = evidence.parents[2]
    work = Path(tempfile.mkdtemp(prefix="sub2api-unmerged-reproduce-"))
    packages = [
        ("apicompat", "pkg/apicompat", "apicompat-probe.go.txt"),
        ("service", "service", "service-probe.go.txt"),
        ("handler", "handler", "handler-probe.go.txt"),
    ]
    overlay = work / "overlay.json"
    overlay.write_text(json.dumps({"Replace": {
        str(root / "backend/internal" / package / "unmerged_review_probe_test.go"):
        str(evidence / filename)
        for _, package, filename in packages
    }}, indent=2) + "\n")
    results = []
    for label, package, _ in packages:
        command = [
            "rtk", "proxy", "go", "test", "-tags=unit", f"-overlay={overlay}",
            "-run", "^TestUnmergedReview", "-count=1", "-v", f"./internal/{package}",
        ]
        result = subprocess.run(command, cwd=root / "backend", capture_output=True, text=True)
        log = work / f"{label}-probe.log"
        log.write_text(result.stdout + result.stderr)
        results.append({"package": package, "exit_code": result.returncode, "log": str(log)})
    print(json.dumps({"work": str(work), "results": results}, indent=2))
    return max(result["exit_code"] for result in results)


if __name__ == "__main__":
    sys.exit(main())
