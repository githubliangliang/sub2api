"""Standalone selected stream/request checks in a temporary 1 CPU / 1 GiB cgroup."""
import json
import os
import subprocess
import sys
from pathlib import Path

result = subprocess.run([
    "/test", "-test.run=^TestFirstOutputMemoryFallbackEightMiBLifecycle$|^TestS22RequestCandidatesPreserveBytes$|^TestSSEParseGateEscapedMetadata$",
    "-test.bench=^BenchmarkFirstOutputMemoryFallbackEightMiB$|^BenchmarkS22Requests$",
    "-test.benchmem", "-test.benchtime=5x", "-test.v",
])
values = {name: Path("/sys/fs/cgroup", name).read_text().strip()
          for name in ("memory.max", "memory.peak", "cpu.max")}
print(json.dumps({"exit_code": result.returncode, "GOMAXPROCS": os.environ.get("GOMAXPROCS"),
                  "cgroup": values, "scope": "standalone service test binary; 8 MiB staging and 200-message/tool synthetic requests; no business database or vendor network"}), flush=True)
sys.exit(result.returncode)
