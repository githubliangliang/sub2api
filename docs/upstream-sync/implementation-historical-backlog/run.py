#!/usr/bin/env python3
"""Record implementation checks, preserving the child exit code and raw output."""
from pathlib import Path
import collections
import datetime
import json
import os
import subprocess
import sys
import time

OUT = Path(__file__).resolve().parent
ROOT = OUT.parents[2]
name, relative, *args = sys.argv[1:]
started = time.time()
environment = {'TZ': 'UTC', 'CI': 'true', 'GOMAXPROCS': '2'}
process = subprocess.run(['rtk', 'proxy', *args], cwd=ROOT / relative,
                         env={**os.environ, **environment}, capture_output=True)
(OUT / (name + '.stdout.log')).write_bytes(process.stdout)
(OUT / (name + '.stderr.log')).write_bytes(process.stderr)
events = []
for line in process.stdout.decode(errors='replace').splitlines():
    try:
        event = json.loads(line)
    except ValueError:
        continue
    if isinstance(event, dict) and event.get('Action') in {'pass', 'fail', 'skip'}:
        events.append(event)
summary = {
    'command': ['rtk', 'proxy', *args], 'cwd': str(ROOT / relative),
    'environment': environment,
    'started_at_utc': datetime.datetime.fromtimestamp(started, datetime.timezone.utc).isoformat(),
    'exit_code': process.returncode, 'seconds': round(time.time() - started, 3),
    'top_level': dict(collections.Counter(e['Action'] for e in events if e.get('Test') and '/' not in e['Test'])),
    'including_subtests': dict(collections.Counter(e['Action'] for e in events if e.get('Test'))),
    'packages': dict(collections.Counter(e['Action'] for e in events if not e.get('Test'))),
}
(OUT / (name + '.json')).write_text(json.dumps(summary, ensure_ascii=False, indent=2) + '\n')
print(json.dumps(summary, ensure_ascii=False))
if not events:
    print((process.stdout + process.stderr).decode(errors='replace')[-2200:])
sys.exit(process.returncode)
