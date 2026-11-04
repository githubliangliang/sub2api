#!/usr/bin/env python3
"""Record a check's exact command, raw output, exit status and duration."""
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
with (out / (name + ".log")).open("w") as log:
    result = subprocess.run(command, cwd=cwd, stdout=log, stderr=subprocess.STDOUT)
record = dict(command=command, cwd=str(Path(cwd).resolve()), started_at_utc=started,
              duration_seconds=round(time.monotonic() - begin, 3), exit_code=result.returncode,
              log=name + ".log")
(out / (name + ".json")).write_text(json.dumps(record, ensure_ascii=False, indent=2) + "\n")
print(json.dumps(record, ensure_ascii=False))
lines = (out / (name + ".log")).read_text(errors="replace").splitlines()
print("\n".join(lines[-12:]))
sys.exit(result.returncode)
