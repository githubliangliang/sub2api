#!/usr/bin/env python3
"""Generate Ent/Wire in a disposable copy of the current backend."""
from pathlib import Path
import hashlib
import json
import shutil
import subprocess
import tempfile

root = Path.cwd()
out = Path(__file__).resolve().parent
temporary = Path(tempfile.mkdtemp(prefix="sub2api-tier2-generation-"))
files = subprocess.check_output(["git", "ls-files", "-co", "--exclude-standard", "backend"], text=True).splitlines()
generated = [p for p in files if p.startswith("backend/ent/") and "/schema/" not in p]
generated.append("backend/cmd/server/wire_gen.go")
before = {p: hashlib.sha256((root / p).read_bytes()).hexdigest() for p in generated}
record = {"steps": [], "generated_files": len(generated), "changes": []}
try:
    for name in files:
        source = root / name
        if source.is_file():
            target = temporary / name
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(source, target)
    for target in ["./ent", "./cmd/server"]:
        command = ["go", "generate", target]
        result = subprocess.run(command, cwd=temporary / "backend", text=True, capture_output=True)
        print(json.dumps({"command": command, "exit_code": result.returncode}))
        print(result.stdout)
        print(result.stderr)
        record["steps"].append({"command": command, "exit_code": result.returncode})
        if result.returncode:
            raise RuntimeError("generator failed")
    for name, digest in before.items():
        if hashlib.sha256((temporary / name).read_bytes()).hexdigest() != digest:
            record["changes"].append(name)
    print(json.dumps(record, indent=2))
    if record["changes"]:
        raise RuntimeError("generated output changed")
finally:
    shutil.rmtree(temporary)
    record["temporary_copy_removed"] = True
    (out / "generation-compatibility.json").write_text(json.dumps(record, indent=2) + "\n")
