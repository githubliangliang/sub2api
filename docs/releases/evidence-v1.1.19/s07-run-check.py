import datetime
import json
import pathlib
import subprocess
import sys
import time

name, cwd, *command = sys.argv[1:]
root = pathlib.Path('/home/liang/myself/work_space/idea/github/sub2api')
evidence = root / 'docs/releases/evidence-v1.1.19'
evidence.mkdir(parents=True, exist_ok=True)
output = evidence / (name + '.txt')
metadata = evidence / (name + '.json')
if output.exists() or metadata.exists():
    raise SystemExit('Refusing to overwrite release evidence: ' + name)
record = {
    'name': name,
    'cwd': cwd,
    'command': command,
    'started': datetime.datetime.now(datetime.timezone.utc).isoformat(),
}
t0 = time.monotonic()
with output.open('w') as stream:
    result = subprocess.run(['rtk', 'proxy', *command], cwd=root / cwd, stdout=stream, stderr=subprocess.STDOUT)
record.update(
    exit_code=result.returncode,
    seconds=round(time.monotonic() - t0, 3),
    ended=datetime.datetime.now(datetime.timezone.utc).isoformat(),
)
metadata.write_text(json.dumps(record, indent=2) + '\n')
print(f'{name}: exit {result.returncode} ({record["seconds"]}s)')
sys.exit(result.returncode)
