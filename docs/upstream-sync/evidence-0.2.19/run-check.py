#!/usr/bin/env python3
"""Capture a baseline command without filtering its evidence or editing source."""
import datetime
import json
from pathlib import Path
import subprocess
import sys
import time

out = Path(__file__).resolve().parent
name, cwd, *command = sys.argv[1:]
started = datetime.datetime.now(datetime.timezone.utc).isoformat()
begin = time.monotonic()
with (out / (name + ".txt")).open("w") as stdout, (out / (name + ".stderr.txt")).open("w") as stderr:
    result = subprocess.run(command, cwd=cwd, stdout=stdout, stderr=stderr)
record = dict(command=command, cwd=str(Path(cwd).resolve()), started_at_utc=started,
    duration_seconds=round(time.monotonic() - begin, 3), exit_code=result.returncode,
    stdout=name + ".txt", stderr=name + ".stderr.txt")
(out / (name + ".json")).write_text(json.dumps(record, ensure_ascii=False, indent=2) + "\n")
print(json.dumps(record, ensure_ascii=False))
lines = (out / (name + ".txt")).read_text(errors="replace").splitlines()
print("\n".join(lines[-12:]))
errors = (out / (name + ".stderr.txt")).read_text(errors="replace").splitlines()
if errors:
    print("\n".join(errors[-12:]))
sys.exit(result.returncode)
