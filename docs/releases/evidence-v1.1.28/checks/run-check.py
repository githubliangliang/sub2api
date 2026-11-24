#!/usr/bin/env python3
"""Save release-gate commands and lossless output separately from prior acceptance."""
import collections
import datetime
import gzip
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import time

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[3]
name, relative, *command = sys.argv[1:]
env = dict(os.environ, CI="true", TZ="UTC", GOTOOLCHAIN="go1.27.4")
removed = []
for key in ("HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "http_proxy", "https_proxy", "all_proxy"):
    if key in env:
        removed.append(key)
        del env[key]
started = time.monotonic()
head = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=ROOT).decode().strip()
result = subprocess.run(["rtk", "proxy", *command], cwd=ROOT / relative,
                        env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
raw = result.stdout
events = []
for line in raw.decode(errors="replace").splitlines():
    try:
        item = json.loads(line)
    except ValueError:
        continue
    if isinstance(item, dict) and item.get("Action") in ("pass", "fail", "skip"):
        events.append(item)
tests = [item for item in events if "Test" in item]
packages = [item for item in events if "Test" not in item]
if len(raw) > 500_000:
    output = name + ".log.gz"
    (OUT / output).write_bytes(gzip.compress(raw, compresslevel=9, mtime=0))
else:
    output = name + ".log"
    (OUT / output).write_bytes(raw)
summary = {
    "finished_at": datetime.datetime.now(datetime.timezone.utc).isoformat(),
    "head_at_start": head,
    "command": ["rtk", "proxy", *command], "cwd": str(ROOT / relative),
    "environment": {"CI": "true", "TZ": "UTC", "GOTOOLCHAIN": "go1.27.4"},
    "proxy_environment_removed": removed,
    "exit_code": result.returncode, "seconds": round(time.monotonic() - started, 3),
    "top_level": dict(collections.Counter(item["Action"] for item in tests if "/" not in item["Test"])),
    "including_subcases": dict(collections.Counter(item["Action"] for item in tests)),
    "package_counts": dict(collections.Counter(item["Action"] for item in packages)),
    "packages": [{key: item[key] for key in ("Action", "Package", "Elapsed") if key in item} for item in packages],
    "failures_and_skips": [{key: item[key] for key in ("Action", "Package", "Test") if key in item} for item in tests if item["Action"] != "pass"],
    "output": output, "output_uncompressed_bytes": len(raw),
    "output_uncompressed_sha256": hashlib.sha256(raw).hexdigest(),
    "output_file_sha256": hashlib.sha256((OUT / output).read_bytes()).hexdigest(),
}
(OUT / (name + ".json")).write_text(json.dumps(summary, ensure_ascii=False, indent=2) + "\n")
print(json.dumps({key: value for key, value in summary.items() if key != "packages"}, ensure_ascii=False))
if not events:
    print(raw.decode(errors="replace")[-3000:])
sys.exit(result.returncode)
