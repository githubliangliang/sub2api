#!/usr/bin/env python3
"""Save exact implementation commands, output and test summaries."""
import collections
import datetime
import json
import os
from pathlib import Path
import subprocess
import sys
import time

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]
name, relative, *command = sys.argv[1:]
started = time.monotonic()
env = dict(os.environ, CI="true", TZ="UTC")
result = subprocess.run(["rtk", "proxy", *command], cwd=ROOT / relative,
                        env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
(OUT / (name + ".log")).write_bytes(result.stdout)
events = []
for line in result.stdout.decode(errors="replace").splitlines():
    try:
        event = json.loads(line)
    except ValueError:
        continue
    if isinstance(event, dict) and event.get("Action") in ("pass", "fail", "skip"):
        events.append(event)
tests = [event for event in events if "Test" in event]
packages = [event for event in events if "Test" not in event]
summary = {
    "finished_at": datetime.datetime.now(datetime.timezone.utc).isoformat(),
    "command": ["rtk", "proxy", *command], "cwd": str(ROOT / relative),
    "environment": {"CI": "true", "TZ": "UTC"}, "exit_code": result.returncode,
    "seconds": round(time.monotonic() - started, 3),
    "top_level": dict(collections.Counter(event["Action"] for event in tests if "/" not in event["Test"])),
    "including_subcases": dict(collections.Counter(event["Action"] for event in tests)),
    "packages": [{key: event[key] for key in ("Action", "Package", "Elapsed") if key in event} for event in packages],
    "failures_and_skips": [{key: event[key] for key in ("Action", "Package", "Test") if key in event} for event in tests if event["Action"] != "pass"],
}
(OUT / (name + ".json")).write_text(json.dumps(summary, ensure_ascii=False, indent=2) + "\n")
print(json.dumps(summary, ensure_ascii=False))
if not events:
    print(result.stdout.decode(errors="replace")[-2500:])
sys.exit(result.returncode)
