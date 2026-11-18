#!/usr/bin/env python3
"""Record one local release check without filtering its output."""

import datetime
import json
import os
import pathlib
import subprocess
import sys
import time

out = pathlib.Path(__file__).resolve().parent
root = out.parents[2]
label, cwd, *command = sys.argv[1:]
env = os.environ.copy()
env["PATH"] = "/home/liang/.nvm/versions/node/v24.16.0/bin:" + env["PATH"]
started = datetime.datetime.now(datetime.timezone.utc).isoformat()
clock = time.monotonic()
with (out / f"{label}.log").open("wb") as log:
    result = subprocess.run(command, cwd=root / cwd, env=env, stdout=log, stderr=subprocess.STDOUT)
record = {
    "command": command,
    "cwd": cwd,
    "started_at_utc": started,
    "finished_at_utc": datetime.datetime.now(datetime.timezone.utc).isoformat(),
    "duration_seconds": round(time.monotonic() - clock, 3),
    "exit_code": result.returncode,
    "log": f"{label}.log",
    "node_path": "/home/liang/.nvm/versions/node/v24.16.0/bin",
}
(out / f"{label}.json").write_text(json.dumps(record, ensure_ascii=False, indent=2) + "\n")
print(json.dumps(record, ensure_ascii=False))
sys.exit(result.returncode)
