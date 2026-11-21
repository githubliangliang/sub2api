import json,os,subprocess,sys
from pathlib import Path
r=subprocess.run(["/test","-test.run=^TestConflictIngressConcurrentBudgetHTTP$|^TestConflictPrereadRetainsNarrowerTextLimitAndStorage$","-test.v"])
print(json.dumps({"exit_code":r.returncode,"GOMAXPROCS":os.environ.get("GOMAXPROCS"),"cgroup":{n:Path("/sys/fs/cgroup",n).read_text().strip() for n in ("memory.max","memory.peak","cpu.max")},"scope":"actual loopback HTTP 64 concurrent 2MiB JSON bodies; middleware/scanner/budget standalone test binary; no business data or external network"}))
sys.exit(r.returncode)
