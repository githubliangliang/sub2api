import datetime
import json
from pathlib import Path
import subprocess
import sys

out = Path(__file__).resolve().parent / "browser"
out.mkdir(exist_ok=True)
file = out / "commands.json"
events = json.loads(file.read_text()) if file.exists() else []
command = ["rtk", "proxy", "agent-browser", "--session", "compatibility-20261031", *sys.argv[1:]]
result = subprocess.run(command, capture_output=True, text=True, timeout=45)
events.append({"at_utc": datetime.datetime.now(datetime.timezone.utc).isoformat(), "command": command,
               "exit_code": result.returncode, "stdout": result.stdout, "stderr": result.stderr})
file.write_text(json.dumps(events, ensure_ascii=False, indent=2) + "\n")
print(result.stdout, end="")
print(result.stderr, end="", file=sys.stderr)
sys.exit(result.returncode)
