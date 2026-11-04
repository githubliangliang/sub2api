"""Record commands against the isolated local acceptance browser."""
import datetime
import json
from pathlib import Path
import subprocess
import sys

out = Path(__file__).resolve().parent / "browser"
out.mkdir(exist_ok=True)
path = out / "commands.json"
commands = json.loads(path.read_text()) if path.exists() else []
command = ["rtk", "proxy", "agent-browser", "--session", "tps-20261104", *sys.argv[1:]]
result = subprocess.run(command, capture_output=True, text=True, timeout=45)
commands.append(dict(at_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(), command=command,
                     exit_code=result.returncode, stdout=result.stdout, stderr=result.stderr))
path.write_text(json.dumps(commands, ensure_ascii=False, indent=2) + "\n")
print(result.stdout, end="")
print(result.stderr, end="", file=sys.stderr)
sys.exit(result.returncode)
