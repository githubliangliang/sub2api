"""Measure the bounded first-output fallback in a temporary 1 CPU / 1 GiB cgroup."""
import json
import os
import subprocess
import sys
from pathlib import Path

result = subprocess.run([
    "/test", "-test.run=^TestFirstOutputMemoryFallbackEightMiBLifecycle$",
    "-test.bench=^BenchmarkFirstOutputMemoryFallbackEightMiB$",
    "-test.benchmem", "-test.benchtime=30x", "-test.v",
])
values = {name: Path("/sys/fs/cgroup", name).read_text().strip()
          for name in ("memory.max", "memory.peak", "cpu.max")}
print(json.dumps({"exit_code": result.returncode, "GOMAXPROCS": os.environ.get("GOMAXPROCS"),
                  "cgroup": values, "scope": "standalone service test binary; 8 MiB synthetic bytes; no network or business database"}), flush=True)
sys.exit(result.returncode)
