"""Record exact commands and unfiltered output for second-tier acceptance."""
import datetime
import json
import os
import subprocess
import sys
import time
from pathlib import Path

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]
label, folder, *command = sys.argv[1:]
if (OUT / (label + ".json")).exists():
    raise SystemExit("Use a new label; acceptance records are append-only")
env = dict(os.environ)
env["PATH"] = "/home/liang/.nvm/versions/node/v24.16.0/bin:" + env["PATH"]
started = datetime.datetime.now(datetime.timezone.utc).isoformat()
start = time.monotonic()
with (OUT / (label + ".log")).open("w") as log:
    result = subprocess.run(
        ["rtk", "proxy", *command], cwd=ROOT / folder, env=env,
        stdout=log, stderr=subprocess.STDOUT,
    )
record = dict(
    command=command, cwd=folder, started_at_utc=started,
    finished_at_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),
    duration_seconds=round(time.monotonic() - start, 3),
    exit_code=result.returncode, log=label + ".log",
)
(OUT / (label + ".json")).write_text(json.dumps(record, indent=2) + "\n")
print(json.dumps(record))
sys.exit(result.returncode)
